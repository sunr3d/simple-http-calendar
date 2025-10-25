package config

import "time"

type Config struct {
	HTTPPort    string        `default:"8080" envconfig:"HTTP_PORT"`
	HTTPTimeout time.Duration `default:"20s"  envconfig:"HTTP_TIMEOUT"`
	LogLevel    string        `default:"info" envconfig:"LOG_LEVEL"`

	ReminderCfg ReminderConfig `envconfig:"REMINDER"`
	ArchiveCfg  ArchiverConfig `envconfig:"ARCHIVE"`
}

type ReminderConfig struct {
	ChanSize int           `default:"100" envconfig:"CHAN_SIZE"`
	Interval time.Duration `default:"2s" envconfig:"INTERVAL"`
}

type ArchiverConfig struct {
	Interval time.Duration `default:"10m" envconfig:"INTERVAL"`
}
