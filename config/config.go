package config

import "github.com/hylent/sf/logger"

type Config interface {
	Get(segment string, target any) error
}

var log = logger.NewLogger(nil, "github.com/hylent/sf/config")
