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
			LogDir:              "tmp",
			CacheDir:            "tmp",
			LogLevel:            "warn",
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

func (x *NacosClient) GetThenListen(dataId string, contentFunc func(content string) error) error {
	content, contentErr := x.client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  x.Group,
	})
	if contentErr != nil {
		return fmt.Errorf("nacos_init_fail: err=%v", contentErr)
	}

	if err := contentFunc(content); err != nil {
		return fmt.Errorf("nacos_init_cb_fail: err=%v", err)
	}

	listenErr := x.client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  x.Group,
		OnChange: func(namespace, group, dataId, content string) {
			if err := contentFunc(content); err != nil {
				logger.Warn("nacos_cb_fail", logger.M{
					"err":       err.Error(),
					"namespace": namespace,
					"group":     group,
					"dataId":    dataId,
					"content":   content,
				})
			}
		},
	})

	if listenErr != nil {
		return fmt.Errorf("nacos_listen_fail: err=%v", listenErr)
	}

	return nil
}
