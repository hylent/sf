package demo

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hylent/sf/config"
	"github.com/hylent/sf/demo/api"
	"github.com/hylent/sf/demo/bs"
	"github.com/hylent/sf/demo/proto"
	"github.com/hylent/sf/logger"
	"github.com/hylent/sf/restful"
	"github.com/hylent/sf/server"
	"github.com/hylent/sf/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
)

func Main() {
	// deal flags
	clientType := flag.String("t", "", "client type: g")
	configFile := flag.String("c", "config.yaml", "config file. file format .yaml")
	help := flag.Bool("h", false, "show usage")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	switch *clientType {
	case "gs":
		listen, err := net.Listen("tcp", "127.0.0.1:9900")
		if err != nil {
			logger.Fatal("listen_fail", logger.M{
				"err": err.Error(),
			})
		}
		s := grpc.NewServer()
		proto.RegisterFooServer(s, new(FooServiceImpl))
		_ = s.Serve(listen)
	case "g":
		conn, err := grpc.Dial(
			"127.0.0.1:9900",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			logger.Fatal("dial_fail", logger.M{
				"err": err.Error(),
			})
		}
		defer conn.Close()
		logger.Info("connected")
		cli := proto.NewFooClient(conn)
		req := &proto.FooRequest{What: "wtf"}
		resp, respErr := cli.Get(context.TODO(), req)
		if respErr != nil {
			logger.Fatal("rpc_fail", logger.M{
				"err": respErr.Error(),
			})
		}
		logger.Info("rpc_ret", logger.M{
			"resp": fmt.Sprintf("%#v", resp),
		})
		return
	}

	// read configs
	conf, confErr := config.ParseFromYamlFile(*configFile)
	if confErr != nil {
		logger.Fatal("config_fail", logger.M{
			"err": confErr.Error(),
		})
	}

	// prepare server
	s := &server.Default{
		Port: 9900,
		Server: &server.Mixed{
			ServerList: []server.Server{
				&server.Grpc{
					Setup: func(s *grpc.Server) {
						proto.RegisterFooServer(s, new(FooServiceImpl))
					},
				},
				&server.Http{
					Setup: func(s *http.Server) {
						rg := &restful.RouterConfig{
							Middlewares: []gin.HandlerFunc{
								restful.LogPerRequest(),
							},
							Handlers: map[string]interface{}{
								"/api/v1/foo": new(FooHandlerImpl),
							},
						}
						s.Handler = rg.NewGinHandler()
					},
				},
			},
		},
	}

	// setup server
	{
		if err := conf.Get("server", s); err != nil {
			logger.Fatal("server_conf_fail", logger.M{
				"err": err.Error(),
			})
		}
	}

	<-util.Terminated(context.TODO(), s.Run)

	logger.Info("bye")
}

type FooServiceImpl struct {
	proto.UnimplementedFooServer
}

func (x *FooServiceImpl) Get(ctx context.Context, req *proto.FooRequest) (*proto.FooResponse, error) {
	in := new(api.FooIn)
	out := new(api.FooOut)
	resp := new(proto.FooResponse)

	in.What = req.What
	err := bs.FooImplInstance.Get(ctx, in, out)
	resp.What = out.What

	return resp, err
}

type FooHandlerImpl struct {
}

func (x *FooHandlerImpl) HandleGet(ctx *gin.Context) {
	restful.Handle(ctx, bs.FooImplInstance.Get)
}
