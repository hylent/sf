package restful

import (
	"context"
	"fmt"
	"github.com/hylent/sf/logger"
	"net/http"
	"time"
)

type Server struct {
	Port                int   `yaml:"port"`
	ShutdownWaitSeconds int64 `yaml:"shutdown_wait_seconds"`

	Handler http.Handler
}

func (x *Server) Run(ctx context.Context) {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", x.Port),
		Handler: x.Handler,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancelFunc := context.WithTimeout(
			context.Background(),
			time.Duration(x.ShutdownWaitSeconds)*time.Second,
		)
		defer cancelFunc()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Warn("server_shutdown_fail", logger.M{
				"err": shutdownErr.Error(),
			})
		}
	}()

	serverErr := server.ListenAndServe()
	if serverErr != nil && serverErr != http.ErrServerClosed {
		logger.Warn("server_fail", logger.M{
			"err": serverErr.Error(),
		})
	}
}
