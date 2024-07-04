package internal

import (
	"context"
	"errors"
	"github.com/hashicorp/go-uuid"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/net"
	"strings"
)

func SetupTcpContextLogger(ctx context.Context, conn net.Conn) context.Context {
	id, _ := uuid.GenerateUUID()
	remoteAddr := conn.RemoteAddr()
	return rocket.SetContextLogID(ctx, id, remoteAddr.String())
}

func SetupUdpContextLogger(ctx context.Context, conn *net.UDPAddr) context.Context {
	id, _ := uuid.GenerateUUID()
	return rocket.SetContextLogID(ctx, id, conn.String())
}

func onTailError(connCtx context.Context, tag string, disErr error) {
	if disErr == nil {
		return
	}
	if !helper.IsCopierError(disErr) && !errors.Is(disErr, context.Canceled) {
		LogTailError(connCtx, tag, disErr)
	}
}

func LogTailError(connCtx context.Context, tag string, disErr error) {
	msg := disErr.Error()
	if strings.Contains(msg, "i/o timeout") {
		return
	}
	if strings.Contains(msg, "connection reset by peer") {
		return
	}
	rocket.Logger(connCtx).Errorf("%s conn error: %s", tag, disErr)
}
