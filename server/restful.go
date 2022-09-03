package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hylent/sf/logger"
	"net/http"
	"strings"
	"time"
)

type Response struct {
	HttpStatusCode int         `json:"-"`
	Code           int         `json:"code"`
	Message        string      `json:"message"`
	Data           interface{} `json:"data"`
}

type E int

func (x E) Error() string {
	return fmt.Sprintf("E%08d", int(x))
}

func (x E) Code() int {
	return int(x)
}

func LogPerRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTp := time.Now()

		ctx.Next()

		log.Info("log_per_request", logger.M{
			"method": ctx.Request.Method,
			"path":   ctx.Request.URL.Path,
			"cost":   time.Now().Sub(startTp).Milliseconds(),
			"status": ctx.Writer.Status(),
		})
	}
}

type RouterConfig struct {
	Middlewares []gin.HandlerFunc
	Handlers    map[string]map[string]gin.HandlerFunc
	Groups      map[string]RouterConfig
}

func (x *RouterConfig) NewGinHandler() http.Handler {
	engine := gin.New()
	engine.Use(gin.Recovery())

	g := engine.Group("/")
	x.registerTo(g)

	return engine
}

func (x *RouterConfig) registerTo(g *gin.RouterGroup) {
	// register middlewares
	if len(x.Middlewares) > 0 {
		g.Use(x.Middlewares...)
	}

	for uri, handlers := range x.Handlers {
		for method, handler := range handlers {
			g.Handle(method, uri, handler)
		}
	}

	// register groups
	for uri, group := range x.Groups {
		g2 := g.Group(uri)
		group.registerTo(g2)
	}
}

func WrapAsGin[IN any, OUT any](h func(context.Context, *IN) (*OUT, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		callAsGin(ctx, h)
	}
}

func callAsGin[IN any, OUT any](ctx *gin.Context, h func(context.Context, *IN) (*OUT, error)) {
	resp := &Response{}
	hCtx := context.TODO()
	in := new(IN)

	for {
		// parse request
		bindFunc := ctx.ShouldBindJSON
		if !strings.HasPrefix(ctx.ContentType(), "application/json") {
			bindFunc = ctx.ShouldBind
		}
		if err := bindFunc(in); err != nil {
			resp.HttpStatusCode = http.StatusBadRequest
			resp.Code = -1
			resp.Message = err.Error()
			break
		}

		// call handler
		out, hErr := h(hCtx, in)

		// check error
		if hErr != nil {
			if duck, duckOk := hErr.(E); duckOk {
				resp.HttpStatusCode = http.StatusOK
				resp.Code = duck.Code()
				resp.Message = duck.Error()
			} else {
				log.Warn("internal_error", logger.M{
					"err": hErr.Error(),
				})
				resp.HttpStatusCode = http.StatusInternalServerError
				resp.Code = -2
			}
			break
		}

		// ok
		resp.HttpStatusCode = http.StatusOK
		resp.Code = 0
		resp.Message = "OK"
		resp.Data = out
		break
	}

	ctx.JSON(resp.HttpStatusCode, resp)
}
