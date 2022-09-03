package server

import (
	"context"
	"fmt"
	"github.com/hylent/sf/logger"
	"github.com/soheilhy/cmux"
	"net"
)

var log = logger.NewLogger(nil, "github.com/hylent/sf/server")

type Server interface {
	Match(cm cmux.CMux) net.Listener
	Serve(ctx context.Context, listener net.Listener) error
}

type Default struct {
	Address string `yaml:"address"`
	Port    int64  `yaml:"port"`

	Server
}

func (x *Default) Run(ctx context.Context) {
	listener, listenerErr := net.Listen("tcp", fmt.Sprintf("%s:%d", x.Address, x.Port))
	if listenerErr != nil {
		log.Warn("default_server_listen_fail", logger.M{
			"err": listenerErr.Error(),
		})
		return
	}

	if err := x.Server.Serve(ctx, listener); err != nil {
		log.Warn("default_server_serve_fail", logger.M{
			"err": err.Error(),
		})
		return
	}
}
