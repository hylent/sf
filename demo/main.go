package demo

import (
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/hylent/sf/clients"
	"github.com/hylent/sf/config"
	"github.com/hylent/sf/db"
	"github.com/hylent/sf/demo/api"
	"github.com/hylent/sf/demo/ds"
	"github.com/hylent/sf/logger"
	"github.com/hylent/sf/restful"
	"github.com/hylent/sf/util"
)

func Main() {
	// deal flags
	configFile := flag.String("c", "config.yaml", "config file. file format .yaml")
	help := flag.Bool("h", false, "show usage")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	// read configs
	conf, confErr := config.ParseFromYamlFile(*configFile)
	if confErr != nil {
		logger.Fatal("config_fail", logger.M{
			"err": confErr.Error(),
		})
	}

	// setup db
	if true {
		ds.Db = &db.AdapterMysql{}
		if err := conf.Get("mysql", ds.Db); err != nil {
			logger.Fatal("mysql_conf_fail", logger.M{
				"err": err.Error(),
			})
		}
		if err := ds.Db.Init(); err != nil {
			logger.Fatal("mysql_init_fail", logger.M{
				"err": err.Error(),
			})
		}
	}

	// setup es
	if true {
		ds.Es = &clients.EsClient{}
		if err := conf.Get("es", ds.Es); err != nil {
			logger.Fatal("es_conf_fail", logger.M{
				"err": err.Error(),
			})
		}
		if err := ds.Es.Init(context.TODO()); err != nil {
			logger.Fatal("es_init_fail", logger.M{
				"err": err.Error(),
			})
		}
	}

	// prepare http server
	serv := &restful.Server{
		Port:                9900,
		ShutdownWaitSeconds: 5,
	}
	if err := conf.Get("restful_server", serv); err != nil {
		logger.Fatal("restful_server_conf_fail", logger.M{
			"err": err.Error(),
		})
	}
	serv.Handler = newRouterConfig().NewGinHandler()

	// serv start
	logger.Info("serv_starting")
	<-util.Terminated(context.TODO(), serv.Run)

	// serv stopped
	logger.Info("serv_stopped")
}

func newRouterConfig() *restful.RouterConfig {
	return &restful.RouterConfig{
		Middlewares: []gin.HandlerFunc{
			restful.LogPerRequest(),
		},
		Handlers: map[string]interface{}{
			"/api/echo": new(api.Echo),
		},
		Groups: map[string]restful.RouterConfig{
			"/api/v1": {
				Middlewares: nil,
				Handlers: map[string]interface{}{
					"/exp": new(api.Exp),
				},
				Groups: nil,
			},
		},
	}
}
