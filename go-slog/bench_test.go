package _go_slog

import (
	"testing"

	"github.com/madkins23/go-slog/bench/tests"
)

// BenchmarkPhusluSlog runs benchmarks for the phuslu/slog handler.
func BenchmarkPhusluSlog(b *testing.B) {
	slogSuite := tests.NewSlogBenchmarkSuite(creatorPhusluSlog())
	tests.Run(b, slogSuite)
}
