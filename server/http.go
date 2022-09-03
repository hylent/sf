package server

import (
	"context"
	"fmt"
	"github.com/hylent/sf/logger"
	"github.com/soheilhy/cmux"
	"net"
	"net/http"
)

type Http struct {
	Setup func(s *http.Server)
}

func (x *Http) Match(cm cmux.CMux) net.Listener {
	return cm.Match(cmux.HTTP1Fast())
}

func (x *Http) Serve(ctx context.Context, listener net.Listener) error {
	s := new(http.Server)

	x.Setup(s)

	go func() {
		<-ctx.Done()
		if err := s.Shutdown(context.Background()); err != nil {
			log.Warn("http_server_shutdown_fail", logger.M{
				"err": err.Error(),
			})
		}
	}()

	log.Debug("http_server_starting")

	if err := s.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http_server_fail: err=%v", err)
	}

	log.Debug("http_server_stopped")
	return nil
}
