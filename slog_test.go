// gccgo's libgo is based on Go 1.18, which has no log/slog package.
//go:build go1.21 && !gccgo

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
			for j := 0; j < 100; j++ {
				base.With("worker", i).WithGroup("g").Info("hello from derived slog", "j", j)
			}
		}(i)
	}
	wg.Wait()
}
