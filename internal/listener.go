package internal

import (
	"fluxway/proxy"
	"runtime/debug"
)

import (
	"context"
	"fluxway/net"
	"fmt"
	"github.com/sirupsen/logrus"
	stdnet "net"
)

type TcpListener struct {
	tag      string
	options  proxy.ListenerOptions
	listener *stdnet.Listener
	defaults net.TcpOptions
}

func NewTcpListener(tag string, options net.TcpOptions) *TcpListener {
	return &TcpListener{
		tag:      tag,
		defaults: options,
	}
}

func (t *TcpListener) ProxyType() proxy.ProxyType {
	return proxy.ProxyType_RAWTCP
}

func (t *TcpListener) Tag() string {
	return t.tag
}

func (t *TcpListener) Network() net.Network {
	return net.Network_TCP
}

func (t *TcpListener) Init(options proxy.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *TcpListener) Serve(serveCtx context.Context, handler proxy.ListenerHandler) error {
	addr := fmt.Sprintf("%s:%d", t.options.Address, t.options.Port)
	logrus.Infof("%s serve, address: %s", t.Tag(), addr)
	listener, err := stdnet.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen tcp address %s %w", addr, err)
	}
	t.listener = &listener
	defer func() {
		logrus.Infof("%s terminated, address: %s", t.Tag(), addr)
		_ = listener.Close()
	}()
	for {
		select {
		case <-serveCtx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				logrus.Errorf("%s accept error: %s", t.Tag(), err)
				return fmt.Errorf("%s accept: %w", t.Tag(), err)
			}
			go func(tcpConn *stdnet.TCPConn) {
				defer func() {
					if err := recover(); err != nil {
						logrus.Errorf("%s handler err: %s, trace: %s", t.Tag(), err, string(debug.Stack()))
					}
				}()
				defer func() {
					logrus.Infof("%s connection terminated, address: %s", t.Tag(), tcpConn.RemoteAddr())
					net.Close(tcpConn)
				}()
				if err := net.SetTcpOptions(tcpConn, t.defaults); err != nil {
					logrus.Errorf("%s handler set local option: %s, trace: %s", t.Tag(), err, string(debug.Stack()))
				} else {
					connCtx, connCancel := context.WithCancel(serveCtx)
					defer connCancel()
					handler(connCtx, net.Connection{
						Network:     t.Network(),
						Address:     net.IPAddress((conn.RemoteAddr().(*stdnet.TCPAddr)).IP),
						TCPConn:     tcpConn,
						Destination: net.DestinationNotset,
						ReadWriter:  conn,
					})
				}
			}(conn.(*stdnet.TCPConn))
		}
	}
}
