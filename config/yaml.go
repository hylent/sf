package config

import (
	"fmt"
	"github.com/hylent/sf/clients"
	"github.com/hylent/sf/logger"
	"github.com/hylent/sf/reloadable"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type yamlConfig struct {
	segmentConfigs map[string]yaml.Node
}

func (x *yamlConfig) Get(segment string, target any) error {
	node, nodeFound := x.segmentConfigs[segment]
	if !nodeFound {
		return fmt.Errorf("config_segment_missing: segment=%s", segment)
	}
	if err := node.Decode(target); err != nil {
		return fmt.Errorf("config_segment_decode_fail: segment=%s", segment)
	}
	return nil
}

func FromEnvYamlFile(envKey string, file string) (Config, error) {
	data, dataErr := os.ReadFile(file)
	if dataErr != nil {
		return nil, fmt.Errorf("config_file_read_fail: file=%s err=%v", file, dataErr)
	}
	return FromEnvYaml(envKey, data)
}

func FromEnvYaml(envKey string, data []byte) (Config, error) {
	configs := struct {
		CurrentEnv string                          `yaml:"current_env"`
		Envs       map[string]map[string]yaml.Node `yaml:"envs"`
		Defaults   map[string]yaml.Node            `yaml:"defaults"`
	}{}

	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("config_decode_fail: err=%v", err)
	}

	currentEnv := os.Getenv(envKey)
	if len(currentEnv) < 1 {
		currentEnv = configs.CurrentEnv
	}

	envConfig, envConfigFound := configs.Envs[currentEnv]
	if !envConfigFound {
		return nil, fmt.Errorf("config_env_missing: env=%s", currentEnv)
	}

	segmentConfigs := configs.Defaults
	if segmentConfigs == nil {
		segmentConfigs = map[string]yaml.Node{}
	}
	for k, v := range envConfig {
		segmentConfigs[k] = v
	}

	x := &yamlConfig{
		segmentConfigs: segmentConfigs,
	}

	return x, nil
}

type reloadableYamlConfig struct {
	reloadable reloadable.Reloadable[yamlConfig]
}

func FromNacos(initTimeout time.Duration, nacosCli *clients.NacosClient, dataId string) (Config, error) {
	contentCh, contentChErr := nacosCli.Get(dataId)
	if contentChErr != nil {
		return nil, fmt.Errorf("config_nacos_fail: err=%v", contentChErr)
	}

	parseFunc := func(content string) (*yamlConfig, error) {
		configs := map[string]yaml.Node{}
		if err := yaml.Unmarshal([]byte(content), &configs); err != nil {
			return nil, fmt.Errorf("config_decode_fail: err=%v", err)
		}
		data := &yamlConfig{
			segmentConfigs: configs,
		}
		return data, nil
	}

	f := func(ch chan<- *yamlConfig) {
		for content := range contentCh {
			data, dataErr := parseFunc(content)
			if dataErr != nil {
				log.Warn("config_parse_fail", logger.M{
					"err": dataErr.Error(),
				})
				time.Sleep(time.Second)
				continue
			}
			ch <- data
		}
	}

	r, rErr := reloadable.New(initTimeout, f)
	if rErr != nil {
		return nil, fmt.Errorf("config_reload_init_fail: err=%v", rErr)
	}

	x := &reloadableYamlConfig{
		reloadable: r,
	}
	return x, nil
}

func (x *reloadableYamlConfig) Get(segment string, target interface{}) error {
	return x.reloadable.Get().Get(segment, target)
}
