package _go_slog

import (
	"io"
	"log/slog"
	"testing"

	benchtests "github.com/madkins23/go-slog/bench/tests"
	"github.com/madkins23/go-slog/infra"
	"github.com/madkins23/go-slog/infra/warning"
	verifytests "github.com/madkins23/go-slog/verify/tests"
	"github.com/phuslu/log"
	"github.com/stretchr/testify/suite"
)

// BenchmarkSlogJSON runs benchmarks for the slog/JSONHandler JSON handler.
func BenchmarkSlogJSON(b *testing.B) {
	slogNewJSONHandler := func(w io.Writer, options *slog.HandlerOptions) slog.Handler {
		return slog.NewJSONHandler(w, options)
	}
	creator := infra.NewCreator("slog/JSONHandler", slogNewJSONHandler, nil,
		`^slog/JSONHandler^ is the JSON handler provided with the ^slog^ library.
		It is fast and as a part of the Go distribution it is used
		along with published documentation as a model for ^slog.Handler^ behavior.`,
		map[string]string{
			"slog/JSONHandler": "https://pkg.go.dev/log/slog#JSONHandler",
		})
	slogSuite := benchtests.NewSlogBenchmarkSuite(creator)
	benchtests.Run(b, slogSuite)
}

// BenchmarkPhusluSlog runs benchmarks for the phuslu/slog handler.
func BenchmarkPhusluSlog(b *testing.B) {
	creator := infra.NewCreator("phuslu/slog", log.SlogNewJSONHandler, nil,
		`^phuslu/slog^ is a wrapper around the pre-existing ^phuslu/log^ logging library.`,
		map[string]string{
			"phuslu/log": "https://github.com/phuslu/log",
		})
	slogSuite := benchtests.NewSlogBenchmarkSuite(creator)
	benchtests.Run(b, slogSuite)
}

// TestVerifyPhusluSlog runs tests for the phuslu/slog handler.
func TestVerifyPhusluSlog(t *testing.T) {
	creator := infra.NewCreator("phuslu/slog", log.SlogNewJSONHandler, nil,
		`^phuslu/slog^ is a wrapper around the pre-existing ^phuslu/log^ logging library.`,
		map[string]string{
			"phuslu/log": "https://github.com/phuslu/log",
		})
	slogSuite := verifytests.NewSlogTestSuite(creator)
	slogSuite.WarnOnly(warning.Duplicates)
	suite.Run(t, slogSuite)
}

func TestMain(m *testing.M) {
	warning.WithWarnings(m)
}
