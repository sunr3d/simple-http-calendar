package main

import (
	"log"

	"github.com/sunr3d/simple-http-calendar/internal/config"
	"github.com/sunr3d/simple-http-calendar/internal/entrypoint"
	"github.com/sunr3d/simple-http-calendar/internal/logger"
)

func main() {
	cfg, err := config.GetConfigFromEnv()
	if err != nil {
		log.Fatalf("ошибка при загрузке конфигруации: %v\n", err)
	}

	asyncLogger, err := logger.New(cfg.LoggerCfg)
	if err != nil {
		log.Fatalf("ошибка при создании логгера: %v\n", err)
	}

	if err = entrypoint.Run(cfg, asyncLogger); err != nil {
		log.Fatalf("ошибка при запуске приложения: %v\n", err)
	}
}
