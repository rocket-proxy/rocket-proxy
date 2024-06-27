package internal

import (
	"context"
	"fluxway/net"
	"fluxway/proxy"
	"fmt"
	"github.com/bytepowered/assert-go"
	stdnet "net"
)

func TcpDialServe(srcConnCtx context.Context, opts net.TcpOptions, link *net.Connection) error {
	assert.MustTrue(link.Destination.Network == net.Network_TCP, "unsupported network: %s", link.Destination.Network)
	assert.MustTrue(link.Destination.Address.Family().IsIP(), "destination must be an ip, was: %s", link.Destination.Address.String())
	srcConn := link.TCPConn
	dstConn, err := stdnet.DialTCP("tcp", nil, &stdnet.TCPAddr{IP: link.Destination.Address.IP(), Port: int(link.Destination.Port)})
	if err != nil {
		return fmt.Errorf("tcp-dial: %w", err)
	}
	defer net.Close(dstConn)
	if err := net.SetTcpOptions(dstConn, opts); err != nil {
		return fmt.Errorf("tcp-dial: set options: %w", err)
	}
	dstCtx, dstCancel := context.WithCancel(srcConnCtx)
	defer dstCancel()
	// Hook: dail
	if hook := proxy.LookupHookDialPhased(srcConnCtx); hook != nil {
		if err := hook(srcConnCtx, link); err != nil {
			return err
		}
	}
	errors := make(chan error, 2)
	copier := func(_ context.Context, name string, from, to net.Conn) {
		errors <- net.Copier(from, to)
	}
	go copier(dstCtx, "src-to-dest", srcConn, dstConn)
	go copier(dstCtx, "dest-to-src", dstConn, srcConn)
	return <-errors
}