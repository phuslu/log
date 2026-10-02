package log

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestSlogJsonHandler(t *testing.T) {
	logger := slog.New(SlogNewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: false}))

	logger1 := logger.WithGroup("g").With("1", "2").With("3", "4")
	logger1.Info("hello from group slog 1", "number", 42)
	logger1.Info("hello from group slog 2")

	logger2 := logger1.WithGroup("g1").With("a", "b").With("c", "d").
		WithGroup("g2").With("foo", "bar").With("bar", "foo").
		WithGroup("g3").With("x", 1).With("y", 2).With("z", 3)
	logger2.Info("hello from group slog 3", "number", 42)
	logger2.Info("hello from group slog 4")

	logger1.Info("hello from group slog 1", "number", 42)
	logger1.WithGroup("group").Info("hello from group slog 2", "number", 42)
}

func TestSlogJsonHandlerClosed(t *testing.T) {
	logger := slog.New(SlogNewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: false}))

	logger1 := logger.WithGroup("g").With("number", 42).WithGroup("g1")
	logger1.Info("hello from group slog 1", "a", 1, "b", 2)
	logger1.With("x", "1", "y", "2").Info("hello from group slog 2", "a", 1, "b", 2)
	logger1.Info("hello from group slog 3")
}

func TestSlogJsonHandlerGroups(t *testing.T) {
	logger := slog.New(SlogNewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	logger.WithGroup("group1").WithGroup("group2").Info("hello from slog groups", slog.Group("subGroup", "a", 1, "b", 2))
	logger.WithGroup("group1").WithGroup("group2").Info("hello from slog groups", slog.Group("subGroup"))
	logger.WithGroup("group1").WithGroup("group2").Info("hello from slog groups")
}

func TestSlogJsonHandlerAny(t *testing.T) {
	logger := slog.New(SlogNewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	var obj = struct {
		Rate string
		Low  int
		High float32
	}{"15", 16, 123.2}

	logger.Info("hello from slog any", "good object", obj)

	logger.Info("hello from slog any", "bad object", logger.Info)
}

// TestSlogHandlerDerivedAliasing makes sure that handlers derived from a
// shared parent do not write into the same attribute buffer: the second With
// must not overwrite the attributes of the first one.
func TestSlogHandlerDerivedAliasing(t *testing.T) {
	handlers := []struct {
		name string
		new  func(w io.Writer) slog.Handler
	}{
		{"SlogNewJSONHandler", func(w io.Writer) slog.Handler {
			return SlogNewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
		}},
		{"LoggerSlog", func(w io.Writer) slog.Handler {
			return (&Logger{Level: InfoLevel, Writer: IOWriter{w}}).Slog().Handler()
		}},
	}

	for _, handler := range handlers {
		t.Run(handler.name, func(t *testing.T) {
			for _, c := range []struct {
				name         string
				first, other string
			}{
				{"same length", "AAA", "BBB"},
				{"different length", "A", "BBBBB"},
			} {
				t.Run(c.name, func(t *testing.T) {
					var buf bytes.Buffer
					// "pad" is sized so that the parent buffer keeps spare
					// capacity, which is what a shallow copy would share.
					parent := handler.new(&buf).WithAttrs([]slog.Attr{slog.String("pad", "pppppppp")})
					first := parent.WithAttrs([]slog.Attr{slog.String("who", c.first)})
					_ = parent.WithAttrs([]slog.Attr{slog.String("who", c.other)})

					if err := first.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)); err != nil {
						t.Fatal(err)
					}

					var m map[string]any
					if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m); err != nil {
						t.Fatalf("derived handler wrote invalid json: %v, got %q", err, buf.Bytes())
					}
					if m["who"] != c.first {
						t.Fatalf("derived handler got \"who\": %v, want %q, got %q", m["who"], c.first, buf.Bytes())
					}
				})
			}
		})
	}
}

// TestSlogHandlerDerivedConcurrent derives loggers from one shared parent
// logger concurrently, which must be safe (run with -race).
func TestSlogHandlerDerivedConcurrent(t *testing.T) {
	base := slog.New(SlogNewJSONHandler(io.Discard, nil)).With("pad", "pppppppp")

	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := range 100 {
				base.With("worker", i).WithGroup("g").Info("hello from derived slog", "j", j)
			}
		}(i)
	}
	wg.Wait()
}

// TestSlogHandlerEmptyGroups checks that groups whose members are all empty
// (or themselves empty groups) are dropped instead of panicking or emitting an
// empty object, matching slog.NewJSONHandler.
func TestSlogHandlerEmptyGroups(t *testing.T) {
	handlers := []struct {
		name string
		new  func(w io.Writer) *slog.Logger
	}{
		{"SlogNewJSONHandler", func(w io.Writer) *slog.Logger {
			return slog.New(SlogNewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
		}},
		{"LoggerSlog", func(w io.Writer) *slog.Logger {
			return (&Logger{Level: InfoLevel, Writer: IOWriter{w}}).Slog()
		}},
	}

	attr := func(t *testing.T, m map[string]any, path ...string) (any, bool) {
		t.Helper()
		var v any = m
		for _, key := range path {
			obj, ok := v.(map[string]any)
			if !ok {
				return nil, false
			}
			if v, ok = obj[key]; !ok {
				return nil, false
			}
		}
		return v, true
	}
	check := func(want any, path ...string) func(*testing.T, map[string]any) {
		return func(t *testing.T, m map[string]any) {
			t.Helper()
			got, ok := attr(t, m, path...)
			if !ok {
				t.Fatalf("%v not found in %v", path, m)
			}
			if got != want {
				t.Fatalf("%v = %v, want %v in %v", path, got, want, m)
			}
		}
	}
	absent := func(path ...string) func(*testing.T, map[string]any) {
		return func(t *testing.T, m map[string]any) {
			t.Helper()
			if v, ok := attr(t, m, path...); ok {
				t.Fatalf("%v = %v, want it to be absent in %v", path, v, m)
			}
		}
	}

	cases := []struct {
		name  string
		log   func(l *slog.Logger)
		check func(*testing.T, map[string]any)
	}{
		{
			name: "record empty group",
			log: func(l *slog.Logger) {
				l.LogAttrs(context.Background(), slog.LevelInfo, "test", slog.Group("g", slog.Attr{}))
			},
			check: absent("g"),
		},
		{
			name: "record nested empty group",
			log: func(l *slog.Logger) {
				l.LogAttrs(context.Background(), slog.LevelInfo, "test", slog.Group("g", slog.Group("h")))
			},
			check: absent("g"),
		},
		{
			name: "record empty group then value",
			log: func(l *slog.Logger) {
				l.LogAttrs(context.Background(), slog.LevelInfo, "test", slog.Group("g", slog.Attr{}), slog.String("k", "v"))
			},
			check: func(t *testing.T, m map[string]any) {
				absent("g")(t, m)
				check("v", "k")(t, m)
			},
		},
		{
			name: "record group with empty and non-empty members",
			log: func(l *slog.Logger) {
				l.LogAttrs(context.Background(), slog.LevelInfo, "test", slog.Group("g", slog.Attr{}, slog.Int("a", 1)))
			},
			check: check(float64(1), "g", "a"),
		},
		{
			name: "record empty group inside derived group",
			log: func(l *slog.Logger) {
				l.WithGroup("outer").LogAttrs(context.Background(), slog.LevelInfo, "test", slog.Group("g", slog.Attr{}))
			},
			check: absent("outer"),
		},
		{
			name: "With empty attr then record value",
			log: func(l *slog.Logger) {
				l.WithGroup("g").With(slog.Attr{}).Info("test", "k", "v")
			},
			check: check("v", "g", "k"),
		},
		{
			name: "With empty attr and no attrs",
			log: func(l *slog.Logger) {
				l.WithGroup("g").With(slog.Attr{}).Info("test")
			},
			check: absent("g"),
		},
		{
			name: "With empty attr then nested group",
			log: func(l *slog.Logger) {
				l.WithGroup("g").With(slog.Attr{}).WithGroup("h").Info("test", "k", "v")
			},
			check: check("v", "g", "h", "k"),
		},
		{
			name: "With empty attr then With attr",
			log: func(l *slog.Logger) {
				l.WithGroup("g").With(slog.Attr{}).With("b", 2).Info("test", "a", 1)
			},
			check: func(t *testing.T, m map[string]any) {
				check(float64(2), "g", "b")(t, m)
				check(float64(1), "g", "a")(t, m)
			},
		},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) {
					var buf bytes.Buffer
					c.log(h.new(&buf))

					var m map[string]any
					if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m); err != nil {
						t.Fatalf("invalid json: %v, got %q", err, buf.Bytes())
					}
					c.check(t, m)
				})
			}
		})
	}
}
