package stream

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/net"
	"time"
)

var (
	_ rocket.Connector = (*TcpConnector)(nil)
)

type TcpConnector struct {
	opts net.TcpOptions
}

func NewTcpConnector() *TcpConnector {
	return &TcpConnector{
		opts: net.DefaultTcpOptions(),
	}
}

func (c *TcpConnector) DialServe(srcConnCtx context.Context, link *net.Connection) error {
	assert.MustTrue(link.Destination.Network == net.NetworkTCP, "dest network is not tcp, was: %s", link.Destination.Network)
	assert.MustTrue(link.Destination.Address.Family().IsIP(), "dest addr is not an ip, was: %s", link.Destination.Address.String())
	srcConn := link.TCPConn()
	dialer := &net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Duration(0),
	}
	dstConn, err := dialer.DialContext(srcConnCtx, "tcp", link.Destination.NetAddr())
	if err != nil {
		return fmt.Errorf("tcp-dial: %w", err)
	}
	defer helper.Close(dstConn)
	if err := net.SetTcpConnOptions(dstConn, c.opts); err != nil {
		return fmt.Errorf("tcp-dial: set options: %w", err)
	}
	dstCtx, dstCancel := context.WithCancel(srcConnCtx)
	defer dstCancel()
	// Hook: dail
	if hook := rocket.LookupHookFunc(srcConnCtx, rocket.CtxHookFuncOnDialer); hook != nil {
		if err := hook(srcConnCtx, link); err != nil {
			return err
		}
	}
	errors := make(chan error, 2)
	copier := func(_ context.Context, name string, from, to net.Conn) {
		errors <- helper.Copier(from, to)
	}
	go copier(dstCtx, "src-to-dest", srcConn, dstConn)
	go copier(dstCtx, "dest-to-src", dstConn, srcConn)
	return <-errors
}
