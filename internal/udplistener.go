package internal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	ionet "net"
	"rocket/net"
	"rocket/proxy"
	"runtime/debug"
)

var (
	_ proxy.Listener = (*UdpListener)(nil)
)

type UdpListener struct {
	tag     string
	options proxy.ListenerOptions
	udpOpts net.UdpOptions
}

func NewUdpListener(tag string, udpOpts net.UdpOptions) *UdpListener {
	return &UdpListener{
		tag:     tag,
		udpOpts: udpOpts,
	}
}

func (t *UdpListener) ServerType() proxy.ServerType {
	return proxy.ServerType_RAWUDP
}

func (t *UdpListener) Network() net.Network {
	return net.Network_UDP
}

func (t *UdpListener) Init(options proxy.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *UdpListener) Listen(serveCtx context.Context, handler proxy.ListenerHandler) error {
	addr := &ionet.UDPAddr{IP: ionet.ParseIP(t.options.Address), Port: t.options.Port}
	logrus.Infof("%s: listen start, address: %s", t.tag, addr)
	listener, lErr := ionet.ListenUDP("udp", addr)
	if lErr != nil {
		return fmt.Errorf("failed to listen udp address %s %w", addr, lErr)
	}
	go func() {
		<-serveCtx.Done()
		_ = listener.Close()
	}()
	for {
		var buffer = make([]byte, 1024*32)
		n, srcAddr, aErr := listener.ReadFromUDP(buffer)
		if aErr != nil {
			select {
			case <-serveCtx.Done():
				return serveCtx.Err()
			default:
				return fmt.Errorf("%s listen read: %w", t.tag, aErr)
			}
		}
		go t.handle(serveCtx, listener, srcAddr, bytes.NewReader(buffer[:n]), handler)
	}
}

func (t *UdpListener) handle(ctx context.Context, listener *net.UDPConn, srcAddr *net.UDPAddr, reader io.Reader,
	handler proxy.ListenerHandler) {
	defer func() {
		if rErr := recover(); rErr != nil {
			logrus.Errorf("%s handle conn: %s, trace: %s", t.tag, rErr, string(debug.Stack()))
		}
	}()
	// Next
	connCtx := SetupUdpContextLogger(ctx, srcAddr)
	hErr := handler(connCtx, net.Connection{
		Network:     t.Network(),
		Address:     net.IPAddress(srcAddr.IP),
		UserContext: context.Background(),
		ReadWriter: &wrapper{
			reader: reader,
			writer: func(b []byte) (_ int, _ error) {
				return listener.WriteToUDP(b, srcAddr)
			},
		},
		Destination: net.DestinationNotset,
	})
	if hErr != nil {
		proxy.Logger(connCtx).Errorf("%s conn error: %s", t.tag, hErr)
	}
}

var (
	_ io.ReadWriter = (*wrapper)(nil)
)

type wrapper struct {
	reader io.Reader
	writer func(b []byte) (n int, err error)
}

func (c *wrapper) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

func (c *wrapper) Write(b []byte) (n int, err error) {
	return c.writer(b)
}

func (c *wrapper) Close() error {
	return nil
}
