package entrypoint

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/config"
	httphandlers "github.com/sunr3d/simple-http-calendar/internal/handlers/http"
	"github.com/sunr3d/simple-http-calendar/internal/infra/inmembroker"
	"github.com/sunr3d/simple-http-calendar/internal/infra/inmemdb"
	"github.com/sunr3d/simple-http-calendar/internal/middleware"
	"github.com/sunr3d/simple-http-calendar/internal/server"
	"github.com/sunr3d/simple-http-calendar/internal/services/archiversvc"
	"github.com/sunr3d/simple-http-calendar/internal/services/calendarsvc"
	"github.com/sunr3d/simple-http-calendar/internal/services/remindersvc"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	logger.Info("запуск приложения...")

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer func() {
		if err := logger.Sync(); err != nil {
			if !strings.Contains(err.Error(), "invalid argument") {
				log.Printf("ошибка при синхронизации и остановке логгера: %v\n", err)
			}
		}
	}()

	/// Инфра слой
	repo := inmemdb.New(logger)
	broker := inmembroker.New(cfg.ReminderCfg.ChanSize, logger)

	/// Сервисный слой
	calSvc := calendarsvc.New(repo, broker, logger)
	remSvc := remindersvc.New(repo, broker, logger)
	archSvc := archiversvc.New(repo, logger, cfg.ArchiveCfg)

	/// HTTP слой
	controller := httphandlers.New(calSvc, logger)
	mux := http.NewServeMux()
	controller.RegisterCalendarHandlers(mux)

	// Middleware
	handler := middleware.Recovery(logger)(
		middleware.ReqLogger(logger)(
			middleware.JSONValidator(logger)(mux),
		),
	)

	// HTTP сервер
	srv := server.New(cfg.HTTPPort, handler, cfg.HTTPTimeout, logger)

	// Запуск сервисов и сервера
	go func() {
		if err := remSvc.Start(appCtx, cfg.ReminderCfg.Interval); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("ошибка в работе сервиса напоминаний: %v\n", err)
			}
		}
	}()
	go func() {
		if err := archSvc.Start(appCtx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("ошибка в работе сервиса архивации: %v\n", err)
			}
		}
	}()

	return srv.Start(appCtx)
}
