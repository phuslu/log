package _go_slog

import (
	"testing"

	"github.com/madkins23/go-slog/bench/tests"
)

// BenchmarkSlogJSON runs benchmarks for the slog/JSONHandler JSON handler.
func BenchmarkSlogJSON(b *testing.B) {
	slogSuite := tests.NewSlogBenchmarkSuite(creatorSlogJson())
	tests.Run(b, slogSuite)
}

// BenchmarkPhusluSlog runs benchmarks for the phuslu/slog handler.
func BenchmarkPhusluSlog(b *testing.B) {
	slogSuite := tests.NewSlogBenchmarkSuite(creatorPhusluSlog())
	tests.Run(b, slogSuite)
}
