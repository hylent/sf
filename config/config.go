package config

const (
	EnvNameOfEnv = "RUN_ENV"
)

type Config interface {
	GetCurrentEnv() string
	Get(segment string, target interface{}) error
}
