package server

import (
	"context"
	"fmt"
	"github.com/hylent/sf/logger"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"net"
)

type Grpc struct {
	Setup func(s *grpc.Server)
}

func (x *Grpc) Match(cm cmux.CMux) net.Listener {
	return cm.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
}

func (x *Grpc) Serve(ctx context.Context, listener net.Listener) error {
	s := grpc.NewServer()

	x.Setup(s)

	go func() {
		<-ctx.Done()
		s.GracefulStop()
	}()

	logger.Debug("grpc_server_starting")

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("grpc_server_fail: err=%v", err)
	}

	logger.Debug("grpc_server_stopped")

	return nil
}
