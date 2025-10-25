package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	server          *http.Server
	logger          *zap.Logger
	shutdownTimeout time.Duration
}

func New(port string, handler http.Handler, timeout time.Duration, logger *zap.Logger) *Server {
	return &Server{
		server: &http.Server{
			Addr:              ":" + port,
			Handler:           handler,
			ReadTimeout:       timeout,
			WriteTimeout:      timeout,
			IdleTimeout:       timeout,
			ReadHeaderTimeout: timeout,
		},
		logger:          logger,
		shutdownTimeout: timeout,
	}
}

func (s *Server) Start(ctx context.Context) error {
	logger := s.logger.With(
		zap.String("service", "server"),
		zap.String("op", "Start"),
	)

	logger.Info("запуск HTTP сервера",
		zap.String("address", s.server.Addr),
	)

	serverErr := make(chan error, 1)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("ошибка HTTP сервера", zap.Error(err))
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("получен сигнал завершения")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			logger.Error("ошибка при завершении работы сервера", zap.Error(err))
			return fmt.Errorf("ошибка при завершении работы сервера: %w", err)
		}

		logger.Info("сервер остановлен")
		return nil
	case err := <-serverErr:
		return fmt.Errorf("ошибка HTTP сервера: %w", err)
	}
}
