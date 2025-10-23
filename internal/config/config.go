package config

import "time"

type Config struct {
	HTTPPort    string        `default:"8080" envconfig:"HTTP_PORT"`
	HTTPTimeout time.Duration `default:"20s"  envconfig:"HTTP_TIMEOUT"`
	LogLevel    string        `default:"info" envconfig:"LOG_LEVEL"`

	ReminderChanSize int           `default:"100" envconfig:"REMINDER_CHAN_SIZE"`
	ReminderInterval time.Duration `default:"2s" envconfig:"REMINDER_INTERVAL"`
}
