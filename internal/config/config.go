package config

import "time"

type Config struct {
	HTTPPort    string        `envconfig:"HTTP_PORT" default:"8080"`
	HTTPTimeout time.Duration `envconfig:"HTTP_TIMEOUT" default:"20s"`
	LogLevel    string        `envconfig:"LOG_LEVEL" default:"info"`
}