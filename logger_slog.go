//go:build go1.21
// +build go1.21

package log

import (
	"context"
	"log/slog"
)

func slogAttrEval(e *Entry, a slog.Attr) *Entry {
	value := a.Value.Resolve()
	switch value.Kind() {
	case slog.KindGroup:
		return e.Dict(a.Key, slogAttrEval(NewContext(nil), value.Group()[0]).Value())
	case slog.KindBool:
		return e.Bool(a.Key, value.Bool())
	case slog.KindDuration:
		return e.Dur(a.Key, value.Duration())
	case slog.KindFloat64:
		return e.Float64(a.Key, value.Float64())
	case slog.KindInt64:
		return e.Int64(a.Key, value.Int64())
	case slog.KindString:
		return e.Str(a.Key, value.String())
	case slog.KindTime:
		return e.Time(a.Key, value.Time())
	case slog.KindUint64:
		return e.Uint64(a.Key, value.Uint64())
	case slog.KindAny:
		fallthrough
	default:
		return e.Any(a.Key, value.Any())
	}
}

type slogGroup struct {
	name  string
	attrs []any // slog.Attr or *slogGroup
}

func (group *slogGroup) lastChild() *slogGroup {
	if len(group.attrs) == 0 {
		return nil
	}
	child, _ := group.attrs[len(group.attrs)-1].(*slogGroup)
	return child
}

func (group *slogGroup) WithAttrs(attrs []slog.Attr) {
	if child := group.lastChild(); child == nil {
		for _, attr := range attrs {
			group.attrs = append(group.attrs, attr)
		}
	} else {
		child.WithAttrs(attrs)
	}
}

func (group *slogGroup) WithGroup(name string) {
	if child := group.lastChild(); child == nil {
		group.attrs = append(group.attrs, &slogGroup{name: name})
	} else {
		child.WithGroup(name)
	}
}

func (group *slogGroup) Eval(e *Entry) *Entry {
	for _, v := range group.attrs {
		switch v := v.(type) {
		case slog.Attr:
			e = slogAttrEval(e, v)
		case *slogGroup:
			e = e.Dict(v.name, v.Eval(NewContext(nil)).Value())
		}
	}
	return e
}

type slogHandler struct {
	logger  Logger
	group   slogGroup
	context Context
}

func (h slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.group.WithAttrs(attrs)
	h.context = h.group.Eval(NewContext(nil)).Value()
	return &h
}

func (h slogHandler) WithGroup(name string) slog.Handler {
	h.group.WithGroup(name)
	h.context = h.group.Eval(NewContext(nil)).Value()
	return &h
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	switch level {
	case slog.LevelDebug:
		return h.logger.Level <= DebugLevel
	case slog.LevelInfo:
		return h.logger.Level <= InfoLevel
	case slog.LevelWarn:
		return h.logger.Level <= WarnLevel
	case slog.LevelError:
		return h.logger.Level <= ErrorLevel
	}
	return false
}

func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	var e *Entry
	switch r.Level {
	case slog.LevelDebug:
		e = h.logger.Debug()
	case slog.LevelInfo:
		e = h.logger.Info()
	case slog.LevelWarn:
		e = h.logger.Warn()
	case slog.LevelError:
		e = h.logger.Error()
	default:
		e = h.logger.Log()
	}

	if len(h.group.attrs) != 0 {
		e = e.Context(h.context)
	}

	r.Attrs(func(attr slog.Attr) bool {
		e = slogAttrEval(e, attr)
		return true
	})
	e.Msg(r.Message)
	return nil
}

// Slog wraps the Logger to provide *slog.Logger
func (l *Logger) Slog() *slog.Logger {
	logger := *l
	switch {
	case logger.Caller > 0:
		logger.Caller += 3
	case logger.Caller < 0:
		logger.Caller -= 3
	}
	return slog.New(&slogHandler{logger: logger})
}
