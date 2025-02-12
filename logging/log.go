package logging

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	logger *slog.Logger
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Trace(msg string, args ...any) {
	l.logger.Log(context.Background(), LevelTrace, msg, args...)
}

type LevelHandler struct {
	level   slog.Leveler
	handler slog.Handler
}

func NewLevelHandler(level slog.Leveler, h slog.Handler) *LevelHandler {
	if lh, ok := h.(*LevelHandler); ok {
		h = lh.Handler()
	}
	return &LevelHandler{level, h}
}

func (h *LevelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *LevelHandler) Handle(ctx context.Context, r slog.Record) error {
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)
			if level < slog.LevelDebug {
				a.Value = slog.StringValue("TRACE")
			}
			return false
		}
		return true
	})
	return h.handler.Handle(ctx, r)
}

func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithAttrs(attrs))
}

func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return NewLevelHandler(h.level, h.handler.WithGroup(name))
}

func (h *LevelHandler) Handler() slog.Handler {
	return h.handler
}

const LevelTrace = slog.Level(-8)

func strToLevel(val string) slog.Level {
	switch val {
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info":
		return slog.LevelInfo
	}
	return slog.LevelWarn
}

func NewDefaultLogger() *Logger {
	level := LevelTrace
	if val, ok := os.LookupEnv("GO_LOG_LEVEL"); ok {
		level = strToLevel(val)
	}
	return &Logger{NewLogger(level)}
}

func NewLogger(level slog.Leveler) *slog.Logger {
	h := slog.Default().Handler()
	return slog.New(NewLevelHandler(level, h))
}
