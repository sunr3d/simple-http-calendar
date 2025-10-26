package logger

import (
	"sync"

	"go.uber.org/zap/zapcore"
)

var _ zapcore.WriteSyncer = (*asyncWriter)(nil)

type asyncWriter struct {
	writeChan chan []byte
	output    zapcore.WriteSyncer
	fallback  zapcore.WriteSyncer
	mu        sync.Mutex
	closed    bool
}

// NewAsyncWriter - конструктор асинхронного writer для логгера zap.
func NewAsyncWriter(output, fallback zapcore.WriteSyncer, chanSize int) zapcore.WriteSyncer {
	aw := &asyncWriter{
		writeChan: make(chan []byte, chanSize),
		output:    output,
		fallback:  fallback,
	}
	aw.start()
	return aw
}

// Write - пишет данные в канал асинхронного логгера с выводом на фоллбэк на случай переполнения канала.
// (реализация io.Writer).
func (w *asyncWriter) Write(p []byte) (int, error) {
	buf := make([]byte, len(p))
	copy(buf, p)

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return w.fallback.Write(p)
	}

	select {
	case w.writeChan <- buf:
		return len(p), nil
	default:
		return w.fallback.Write(p)
	}
}

// Sync - функция синхронизации асинхронного логгера при graceful shutdown.
// Закрывает канал и выводит оставшиеся данные из канала в фоллбэк.
// (реализация zapcore.WriteSyncer).
func (w *asyncWriter) Sync() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return w.fallback.Sync()
	}

	w.closed = true
	close(w.writeChan)
	w.mu.Unlock()

	for msg := range w.writeChan {
		_, _ = w.output.Write(msg)
	}

	return w.fallback.Sync()
}

// start - читает данные из канал и записывает в output writer.
func (w *asyncWriter) start() {
	go func() {
		for data := range w.writeChan {
			_, _ = w.output.Write(data)
		}
	}()
}
