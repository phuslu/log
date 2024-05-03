package _go_slog

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/madkins23/go-slog/infra/warning"
	"github.com/madkins23/go-slog/verify/tests"
)

// TestVerifyPhusluSlog runs tests for the phuslu/slog handler.
func TestVerifyPhusluSlog(t *testing.T) {
	slogSuite := tests.NewSlogTestSuite(creatorPhusluSlog())
	slogSuite.WarnOnly(warning.Duplicates)
	slogSuite.WarnOnly(warning.StringAny)
	suite.Run(t, slogSuite)
}

func TestMain(m *testing.M) {
	warning.WithWarnings(m)
}
