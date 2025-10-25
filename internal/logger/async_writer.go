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
	wg        sync.WaitGroup
	once      sync.Once
}

// NewAsyncWriter - конструктор асинхронного writer для логгера zap.
func NewAsyncWriter(output, fallback zapcore.WriteSyncer, chanSize int) zapcore.WriteSyncer {
	aw := &asyncWriter{
		writeChan: make(chan []byte, chanSize),
		output:    output,
		fallback:  fallback,
		wg:        sync.WaitGroup{},
	}
	aw.start()
	return aw
}

// Write - пишет данные в канал асинхронного логгера с выводом на фоллбэк на случай переполнения канала.
// (реализация io.Writer)
func (w *asyncWriter) Write(p []byte) (n int, err error) {
	buf := make([]byte, len(p))
	copy(buf, p)

	defer func() {
		if r := recover(); r != nil {
			_, _ = w.fallback.Write(p)
		}
	}()

	select {
	case w.writeChan <- buf:
		return len(p), nil
	default:
		return w.fallback.Write(p)
	}
}

// Sync - функция синхронизации асинхронного логгера при graceful shutdown.
// Закрывает канал и выводит оставшиеся данные из канала в фоллбэк.
// (реализация zapcore.WriteSyncer)
func (w *asyncWriter) Sync() error {
	w.once.Do(func() {
		close(w.writeChan)
	})
	w.wg.Wait()
	return w.fallback.Sync()
}

// start - читает данные из канал и записывает в output writer.
func (w *asyncWriter) start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for data := range w.writeChan {
			_, _ = w.output.Write(data)
		}
	}()
}
