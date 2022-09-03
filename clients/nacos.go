package clients

import (
	"fmt"
	"github.com/hylent/sf/logger"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type NacosClient struct {
	Addr         string `yaml:"addr"`
	Port         uint64 `yaml:"port"`
	Namespace    string `yaml:"namespace"`
	AccessKey    string `yaml:"access_key"`
	SecretKey    string `yaml:"secret_key"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Group        string `yaml:"group"`
	TimeoutMilli uint64 `yaml:"timeout_milli"`

	client config_client.IConfigClient
}

type nacosLogger struct{}

func (x *nacosLogger) Debug(args ...interface{}) {
}

func (x *nacosLogger) Info(args ...interface{}) {
}

func (x *nacosLogger) Warn(args ...interface{}) {
	log.Warn("nacos_warn", logger.M{
		"args": fmt.Sprintf("%+v", args),
	})
}

func (x *nacosLogger) Error(args ...interface{}) {
	log.Warn("nacos_error", logger.M{
		"args": fmt.Sprintf("%+v", args),
	})
}

func (x *nacosLogger) Debugf(fmt string, args ...interface{}) {
}

func (x *nacosLogger) Infof(fmt string, args ...interface{}) {
}

func (x *nacosLogger) Warnf(format string, args ...interface{}) {
	log.Warn("nacos_warn", logger.M{
		"msg": fmt.Sprintf(format, args),
	})
}

func (x *nacosLogger) Errorf(format string, args ...interface{}) {
	log.Warn("nacos_error", logger.M{
		"msg": fmt.Sprintf(format, args),
	})
}

func (x *NacosClient) Init() error {
	param := vo.NacosClientParam{
		ClientConfig: &constant.ClientConfig{
			NamespaceId:         x.Namespace,
			AccessKey:           x.AccessKey,
			SecretKey:           x.SecretKey,
			Username:            x.Username,
			Password:            x.Password,
			TimeoutMs:           x.TimeoutMilli,
			NotLoadCacheAtStart: true,
			CacheDir:            "tmp",
			CustomLogger:        new(nacosLogger),
		},
		ServerConfigs: []constant.ServerConfig{
			{
				IpAddr: x.Addr,
				Port:   x.Port,
			},
		},
	}

	client, clientErr := clients.NewConfigClient(param)
	if clientErr != nil {
		return fmt.Errorf("nacos_client_fail: err=%v", clientErr)
	}

	x.client = client
	return nil
}

func (x *NacosClient) Get(dataId string) (<-chan string, error) {
	content, contentErr := x.client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  x.Group,
	})
	if contentErr != nil {
		return nil, fmt.Errorf("nacos_init_fail: err=%v", contentErr)
	}

	ch := make(chan string, 1)
	ch <- content

	listenErr := x.client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  x.Group,
		OnChange: func(namespace, group, dataId, content string) {
			ch <- content
		},
	})
	if listenErr != nil {
		return nil, fmt.Errorf("nacos_listen_fail: err=%v", listenErr)
	}

	return ch, nil
}
