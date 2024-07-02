package internal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bytepowered/goes"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	"io"
	stdnet "net"
	"runtime/debug"
	"time"
)

var (
	_ rocket.Listener = (*UdpListener)(nil)
)

type UdpListener struct {
	tag     string
	options rocket.ListenerOptions
	udpOpts net.UdpOptions
}

func NewUdpListener(tag string, udpOpts net.UdpOptions) *UdpListener {
	return &UdpListener{
		tag:     tag,
		udpOpts: udpOpts,
	}
}

func (t *UdpListener) Network() net.Network {
	return net.Network_UDP
}

func (t *UdpListener) Init(options rocket.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *UdpListener) Listen(serveCtx context.Context, dispatchHandler rocket.ListenerHandler) error {
	addr := &stdnet.UDPAddr{IP: stdnet.ParseIP(t.options.Address), Port: t.options.Port}
	logrus.Infof("%s: listen start, address: %s", t.tag, addr)
	listener, lErr := stdnet.ListenUDP("udp", addr)
	if lErr != nil {
		return fmt.Errorf("failed to listen udp address %s %w", addr, lErr)
	}
	_ = listener.SetDeadline(time.Time{})
	go func() {
		<-serveCtx.Done()
		_ = listener.Close()
	}()
	for {
		buffer := make([]byte, 1024*32)
		n, srcAddr, aErr := listener.ReadFromUDP(buffer)
		if aErr != nil {
			select {
			case <-serveCtx.Done():
				return serveCtx.Err()
			default:
				return fmt.Errorf("%s listen read: %w", t.tag, aErr)
			}
		}
		goes.Go(func() {
			t.handle(serveCtx, listener, srcAddr, buffer[:n], dispatchHandler)
		})
	}
}

func (t *UdpListener) handle(ctx context.Context, listener *net.UDPConn, srcAddr *net.UDPAddr, data []byte,
	dispatchHandler rocket.ListenerHandler) {
	connCtx := SetupUdpContextLogger(ctx, srcAddr)
	defer func() {
		if rErr := recover(); rErr != nil {
			rocket.Logger(connCtx).Errorf("%s handle conn: %s, trace: %s", t.tag, rErr, string(debug.Stack()))
		}
	}()
	hErr := dispatchHandler(connCtx, net.Connection{
		Network:     t.Network(),
		Address:     net.IPAddress(srcAddr.IP),
		UserContext: context.Background(),
		ReadWriter: &wrapper{
			reader: bytes.NewReader(data),
			writer: func(b []byte) (_ int, _ error) {
				return listener.WriteToUDP(b, srcAddr)
			},
		},
		Destination: net.DestinationNotset,
	})
	if hErr != nil {
		rocket.Logger(connCtx).Errorf("%s conn error: %s", t.tag, hErr)
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
