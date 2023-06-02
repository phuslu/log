//go:build go1.21
// +build go1.21

package log

import (
	"context"
	"log/slog"
)

type slogHandler struct {
	Logger
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	switch level {
	case slog.LevelDebug:
		return h.Logger.Level <= DebugLevel
	case slog.LevelInfo:
		return h.Logger.Level <= InfoLevel
	case slog.LevelWarn:
		return h.Logger.Level <= WarnLevel
	case slog.LevelError:
		return h.Logger.Level <= ErrorLevel
	}
	return false
}

func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	e := h.Logger.Log()
	r.Attrs(func(attr slog.Attr) bool {
		e = e.Any(attr.Key, attr.Value)
		return true
	})
	e.Msg(r.Message)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	e := NewContext(nil)
	for _, attr := range attrs {
		e = e.Any(attr.Key, attr.Value)
	}

	l := h.Logger
	l.Context = e.Value()

	return &slogHandler{l}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return h
}

// Slog wraps the Logger to provide *slog.Logger
func (l *Logger) Slog() *slog.Logger {
	return slog.New(&slogHandler{*l})
}
