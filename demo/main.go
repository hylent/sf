package demo

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hylent/sf/config"
	"github.com/hylent/sf/demo/bs"
	"github.com/hylent/sf/demo/proto"
	"github.com/hylent/sf/logger"
	"github.com/hylent/sf/restful"
	"github.com/hylent/sf/server"
	"github.com/hylent/sf/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"time"
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
		req := &proto.FooIn{What: "wtf"}
		if time.Now().Unix()%2 == 0 {
			req.What = "fuck"
		}
		resp, respErr := cli.Get(context.TODO(), req)
		if respErr != nil {
			logger.Fatal("rpc_fail", logger.M{
				"err": fmt.Sprintf("[%T]%+v", respErr, respErr),
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
						proto.RegisterFooServer(s, bs.FooInstance)
					},
				},
				&server.Http{
					Setup: func(s *http.Server) {
						rg := &restful.RouterConfig{
							Middlewares: []gin.HandlerFunc{
								restful.LogPerRequest(),
							},
							Handlers: map[string]interface{}{
								"/api/v1/foo": bs.FooInstance,
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
