package tcp

import (
	"context"
	"fluxway/internal"
	"fluxway/net"
	"fluxway/proxy"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	_ proxy.Connector = (*Connector)(nil)
)

type Connector struct {
	opts net.TcpOptions
}

func NewTcpConnector() *Connector {
	return &Connector{
		opts: net.TcpOptions{
			ReadTimeout:  time.Second * 30,
			WriteTimeout: time.Second * 10,
			ReadBuffer:   1024,
			WriteBuffer:  1024,
			NoDelay:      true,
			KeepAlive:    time.Second * 10,
		},
	}
}

func (d *Connector) DailServe(inctx context.Context, link *net.Connection) error {
	logrus.Infof("tcp-connector: connect %s to %s", link.Address, link.Destination)
	return internal.TcpConnect(inctx, d.opts, link)
}
