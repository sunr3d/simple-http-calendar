package config

import "time"

type Config struct {
	HTTPPort    string        `default:"8080" envconfig:"HTTP_PORT"`
	HTTPTimeout time.Duration `default:"20s"  envconfig:"HTTP_TIMEOUT"`
	LogLevel    string        `default:"info" envconfig:"LOG_LEVEL"`
}
