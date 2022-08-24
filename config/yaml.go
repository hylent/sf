package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type yamlConfig struct {
	currentEnv     string
	segmentConfigs map[string]yaml.Node
}

func ParseFromYamlFile(file string) (Config, error) {
	data, dataErr := ioutil.ReadFile(file)
	if dataErr != nil {
		return nil, fmt.Errorf("config_file_read_fail: file=%s err=%v", file, dataErr)
	}
	return ParseFromYaml(data)
}

func ParseFromYaml(data []byte) (Config, error) {
	configs := struct {
		CurrentEnv string                          `yaml:"current_env"`
		Envs       map[string]map[string]yaml.Node `yaml:"envs"`
		Defaults   map[string]yaml.Node            `yaml:"defaults"`
	}{}

	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("config_decode_fail: err=%v", err)
	}

	currentEnv := os.Getenv(EnvNameOfEnv)
	if len(currentEnv) < 1 {
		currentEnv = configs.CurrentEnv
	}

	envConfig, envConfigFound := configs.Envs[currentEnv]
	if !envConfigFound {
		return nil, fmt.Errorf("config_env_missing: run_env=%s", currentEnv)
	}

	segmentConfigs := configs.Defaults
	if segmentConfigs == nil {
		segmentConfigs = map[string]yaml.Node{}
	}
	for k, v := range envConfig {
		segmentConfigs[k] = v
	}

	x := &yamlConfig{
		currentEnv:     currentEnv,
		segmentConfigs: segmentConfigs,
	}

	return x, nil
}

func (x *yamlConfig) GetCurrentEnv() string {
	return x.currentEnv
}

func (x *yamlConfig) Get(segment string, target interface{}) error {
	node, nodeFound := x.segmentConfigs[segment]
	if !nodeFound {
		return fmt.Errorf("config_segment_missing: segment=%s", segment)
	}
	if err := node.Decode(target); err != nil {
		return fmt.Errorf("config_segment_decode_fail: segment=%s", segment)
	}
	return nil
}
