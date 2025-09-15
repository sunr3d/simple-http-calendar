package entrypoint

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/config"
	"github.com/sunr3d/simple-http-calendar/internal/middleware"
	"github.com/sunr3d/simple-http-calendar/internal/server"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	logger.Info("запуск приложения...")

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	/// TODO: Инфра слой - инмем БД (мапа)

	/// TODO: Сервисный слой

	/// TODO: HTTP слой
	controller := http_handlers.New(svc, logger)
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

	return srv.Start(appCtx)
}
