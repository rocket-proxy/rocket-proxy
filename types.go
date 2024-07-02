package rocket

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy/net"
)

//// Server Type

type ServerType uint8

const (
	ServerTypeTCP ServerType = iota
	ServerTypeUDP
	ServerTypeSOCKS
	ServerTypeHTTPS
)

func (t ServerType) String() string {
	switch t {
	case ServerTypeTCP:
		return "tcp"
	case ServerTypeUDP:
		return "udp"
	case ServerTypeSOCKS:
		return "socks"
	case ServerTypeHTTPS:
		return "https"
	}
	return "unknown"
}

//// Authenticate Types

const (
	AuthenticateSource = "Source"
	AuthenticateBasic  = "Basic"
	AuthenticateBearer = "Bearer"
	AuthenticateToken  = "Token"
)

//// ListenerHandlerAdapter

var (
	_ ListenerHandler = (*ListenerHandlerAdapter)(nil)
)

type ListenerHandlerAdapter struct {
	Authenticator Authenticator
	Dispatcher    DispatchFunc
}

func (l *ListenerHandlerAdapter) Dispatch(ctx context.Context, conn net.Connection) error {
	return l.Dispatcher(ctx, conn)
}

func (l *ListenerHandlerAdapter) Authenticate(ctx context.Context, conn net.Connection, auth Authentication) error {
	return l.Authenticator.Authenticate(ctx, conn, auth)
}

//// AuthenticatorFunc

var (
	_ Authenticator = (AuthenticatorFunc)(nil)
)

type AuthenticatorFunc func(ctx context.Context, conn net.Connection, auth Authentication) error

func (f AuthenticatorFunc) Authenticate(ctx context.Context, connection net.Connection, authentication Authentication) error {
	return f(ctx, connection, authentication)
}
