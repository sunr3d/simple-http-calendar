package entrypoint

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/config"
	httphandlers "github.com/sunr3d/simple-http-calendar/internal/handlers/http"
	"github.com/sunr3d/simple-http-calendar/internal/infra/inmembroker"
	"github.com/sunr3d/simple-http-calendar/internal/infra/inmemdb"
	"github.com/sunr3d/simple-http-calendar/internal/middleware"
	"github.com/sunr3d/simple-http-calendar/internal/server"
	"github.com/sunr3d/simple-http-calendar/internal/services/calendarsvc"
	"github.com/sunr3d/simple-http-calendar/internal/services/remindersvc"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	logger.Info("запуск приложения...")

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	/// Инфра слой
	repo := inmemdb.New(logger)
	broker := inmembroker.New(cfg.ReminderChanSize, logger)

	/// Сервисный слой
	calSvc := calendarsvc.New(repo, broker, logger)
	remSvc := remindersvc.New(repo, broker, logger)

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

	/// TODO: HTTP сервер
	srv := server.New(cfg.HTTPPort, handler, cfg.HTTPTimeout, logger)

	go remSvc.Start(appCtx, cfg.ReminderInterval)
	return srv.Start(appCtx)
}
