//go:build go1.21
// +build go1.21

package log

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

type slogJSONHandler struct {
	writer   io.Writer
	level    slog.Leveler
	options  *slog.HandlerOptions
	fallback slog.Handler

	group    slogGroup
	grouping bool
	once     sync.Once
	context  Context
}

func (h *slogJSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.level.Level() <= level
}

// nolint:govet // disable copylocks lint
func (h slogJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.options != nil && h.options.ReplaceAttr != nil {
		h.fallback = h.fallback.WithAttrs(attrs)
	}
	h.group.WithAttrs(attrs)
	h.grouping = false
	h.once = sync.Once{}
	return &h
}

// nolint:govet // disable copylocks lint
func (h slogJSONHandler) WithGroup(name string) slog.Handler {
	if h.options != nil && h.options.ReplaceAttr != nil {
		h.fallback = h.fallback.WithGroup(name)
	}
	if name != "" {
		h.group.WithGroup(name)
		h.grouping = true
	}
	return &h
}

func (h *slogJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.options != nil && h.options.ReplaceAttr != nil {
		return h.fallback.Handle(ctx, r)
	}
	return h.handle(ctx, r)
}

func (h *slogJSONHandler) addSource(e *Entry, pc uintptr) *Entry {
	name, file, line := pcNameFileLine(pc)
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, slog.SourceKey...)
	e.buf = append(e.buf, `":{"function":"`...)
	e.buf = append(e.buf, name...)
	e.buf = append(e.buf, `","file":"`...)
	e.buf = append(e.buf, file...)
	e.buf = append(e.buf, `","line":`...)
	e.buf = strconv.AppendInt(e.buf, int64(line), 10)
	e.buf = append(e.buf, '}')
	return e
}

func (h *slogJSONHandler) handle(_ context.Context, r slog.Record) error {
	e := epool.Get().(*Entry)
	e.buf = e.buf[:0]

	e.buf = append(e.buf, '{')

	// time
	if !r.Time.IsZero() {
		e.buf = append(e.buf, '"')
		e.buf = append(e.buf, slog.TimeKey...)
		e.buf = append(e.buf, `":"`...)
		e.buf = r.Time.AppendFormat(e.buf, time.RFC3339Nano)
		e.buf = append(e.buf, `",`...)
	}

	// level
	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, slog.LevelKey...)
	switch r.Level {
	case slog.LevelDebug:
		e.buf = append(e.buf, `":"DEBUG"`...)
	case slog.LevelInfo:
		e.buf = append(e.buf, `":"INFO"`...)
	case slog.LevelWarn:
		e.buf = append(e.buf, `":"WARN"`...)
	case slog.LevelError:
		e.buf = append(e.buf, `":"ERROR"`...)
	default:
		e.buf = append(e.buf, `":"`...)
		e.buf = append(e.buf, r.Level.String()...)
		e.buf = append(e.buf, '"')
	}

	// source
	if h.options != nil && h.options.AddSource {
		e = h.addSource(e, r.PC)
	}

	// msg
	e = e.Str(slog.MessageKey, r.Message)

	if h.grouping {
		// attrs
		var attrs []slog.Attr
		r.Attrs(func(attr slog.Attr) bool {
			attrs = append(attrs, attr)
			return true
		})
		h.group.WithAttrs(attrs)
		h.context = h.group.Eval(NewContext(nil)).Value()

		// with
		if len(h.group.attrs) != 0 {
			e = e.Context(h.context)
		}
	} else {
		// with
		if len(h.group.attrs) != 0 {
			h.once.Do(func() { h.context = h.group.Eval(NewContext(nil)).Value() })
			e = e.Context(h.context)
		}

		// attrs
		r.Attrs(func(attr slog.Attr) bool {
			e = slogAttrEval(e, attr)
			return true
		})
	}

	e.buf = append(e.buf, '}', '\n')

	_, err := h.writer.Write(e.buf)

	if cap(e.buf) <= bbcap {
		epool.Put(e)
	}

	return err
}

func SlogNewJSONHandler(writer io.Writer, options *slog.HandlerOptions) slog.Handler {
	h := &slogJSONHandler{
		writer:   writer,
		level:    slog.LevelInfo,
		options:  options,
		fallback: slog.NewJSONHandler(writer, options),
	}
	if h.options != nil && h.options.Level != nil {
		h.level = h.options.Level
	}
	return h
}
