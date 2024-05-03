//go:build go1.21
// +build go1.21

package log

import (
	"log/slog"
	"os"
	"testing"
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
