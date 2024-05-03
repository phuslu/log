package bench

import (
	"testing"

	"github.com/madkins23/go-slog/bench/tests"

	"go_slog/creator"
)

// BenchmarkPhusluSlog runs benchmarks for the phuslu/slog handler.
func BenchmarkPhusluSlog(b *testing.B) {
	slogSuite := tests.NewSlogBenchmarkSuite(creator.PhusluSlog())
	tests.Run(b, slogSuite)
}
