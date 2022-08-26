package server

import (
	"context"
	"fmt"
	"github.com/hylent/sf/logger"
	"github.com/soheilhy/cmux"
	"net"
)

type Server interface {
	Match(cm cmux.CMux) net.Listener
	Serve(ctx context.Context, listener net.Listener) error
}

type Default struct {
	Port int64 `yaml:"port"`

	Server
}

func (x *Default) Run(ctx context.Context) {
	listener, listenerErr := net.Listen("tcp", fmt.Sprintf(":%d", x.Port))
	if listenerErr != nil {
		logger.Warn("default_server_listen_fail", logger.M{
			"err": listenerErr.Error(),
		})
		return
	}

	if err := x.Server.Serve(ctx, listener); err != nil {
		logger.Warn("default_server_serve_fail", logger.M{
			"err": err.Error(),
		})
		return
	}
}
