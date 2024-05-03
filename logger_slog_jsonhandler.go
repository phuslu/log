//go:build go1.21
// +build go1.21

package log

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"time"
)

type slogJSONHandler struct {
	writer   io.Writer
	level    slog.Leveler
	options  *slog.HandlerOptions
	fallback slog.Handler

	grouping bool
	groups   int
	entry    Entry
}

func (h *slogJSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.level.Level() <= level
}

func (h slogJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.options != nil && h.options.ReplaceAttr != nil {
		h.fallback = h.fallback.WithAttrs(attrs)
	}
	if len(attrs) == 0 {
		return &h
	}
	i := len(h.entry.buf)
	for _, attr := range attrs {
		h.entry = *slogAttrEval(&h.entry, attr)
	}
	if h.grouping {
		h.entry.buf[i] = '{'
	}
	h.grouping = false
	return &h
}

func (h slogJSONHandler) WithGroup(name string) slog.Handler {
	if h.options != nil && h.options.ReplaceAttr != nil {
		h.fallback = h.fallback.WithGroup(name)
	}
	if name != "" {
		if h.grouping {
			h.entry.buf = append(h.entry.buf, '{')
		} else {
			h.entry.buf = append(h.entry.buf, ',')
		}
		h.entry.buf = append(h.entry.buf, '"')
		h.entry.buf = append(h.entry.buf, name...)
		h.entry.buf = append(h.entry.buf, '"', ':')
		h.grouping = true
		h.groups++
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

	// with
	if b := h.entry.buf; len(b) != 0 {
		e = e.Context(b)
	}
	i := len(e.buf)

	// attrs
	r.Attrs(func(attr slog.Attr) bool {
		e = slogAttrEval(e, attr)
		return true
	})

	lastindex := func(buf []byte) int {
		for i := len(buf) - 3; i >= 1; i-- {
			if buf[i] == '"' && (buf[i-1] == ',' || buf[i-1] == '{') {
				return i
			}
		}
		return -1
	}

	// group attrs
	if h.grouping {
		if r.NumAttrs() > 0 {
			e.buf[i] = '{'
		} else if i = lastindex(e.buf); i > 0 {
			e.buf = e.buf[:i-1]
			h.groups--
			for e.buf[len(e.buf)-1] == ':' {
				if i = lastindex(e.buf); i > 0 {
					e.buf = e.buf[:i-1]
					h.groups--
				}
			}
		} else {
			e.buf = append(e.buf, '{')
		}
	}

	// brackets closing
	switch h.groups {
	case 0:
		e.buf = append(e.buf, '}', '\n')
	case 1:
		e.buf = append(e.buf, '}', '}', '\n')
	case 2:
		e.buf = append(e.buf, '}', '}', '}', '\n')
	case 3:
		e.buf = append(e.buf, '}', '}', '}', '}', '\n')
	case 4:
		e.buf = append(e.buf, '}', '}', '}', '}', '}', '\n')
	default:
		for i := 0; i <= h.groups; i++ {
			e.buf = append(e.buf, '}')
		}
		e.buf = append(e.buf, '\n')
	}

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
		entry:    *NewContext(nil),
	}
	if h.options != nil && h.options.Level != nil {
		h.level = h.options.Level
	}
	return h
}
