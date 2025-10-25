package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sunr3d/simple-http-calendar/internal/config"
)

func New(cfg config.LoggerConfig) (*zap.Logger, error) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		lvl = zapcore.InfoLevel
	}

	stdout := zapcore.AddSync(os.Stdout)
	// Для пет-проекта использует stdout и на фоллбэк и на основнй аутпут writer
	asyncWriter := NewAsyncWriter(stdout, stdout, cfg.ChanSize)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encoderCfg)
	core := zapcore.NewCore(encoder, asyncWriter, lvl)

	logger := zap.New(core)
	if logger == nil {
		return nil, fmt.Errorf("не удалось создать логгер")
	}

	return logger, nil
}
