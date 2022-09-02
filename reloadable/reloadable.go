package reloadable

import (
	"fmt"
	"github.com/hylent/sf/clients"
	"github.com/hylent/sf/config"
	"sync/atomic"
	"unsafe"
)

type Reload[T any] struct {
	ptr unsafe.Pointer
}

func (x *Reload[T]) Set(data *T) {
	atomic.StorePointer(&x.ptr, unsafe.Pointer(data))
}

func (x *Reload[T]) Get() (data *T) {
	return (*T)(atomic.LoadPointer(&x.ptr))
}

type FooConfig struct {
	reload Reload[config.YamlConfig]
}

func (x *FooConfig) Init(nacosCli *clients.NacosClient, dataId string) error {
	return nacosCli.GetThenListen(dataId, func(content string) error {
		cfg, cfgErr := config.ParseFromYaml([]byte(content))
		if cfgErr != nil {
			return fmt.Errorf("foo_config_load_fail: err=%v", cfgErr)
		}
		x.reload.Set(cfg)
		return nil
	})
}

func (x *FooConfig) GetCurrentEnv() string {
	return x.reload.Get().GetCurrentEnv()
}

func (x *FooConfig) Get(segment string, target interface{}) error {
	return x.reload.Get().Get(segment, target)
}

var (
	_ config.Config = &FooConfig{}
)
