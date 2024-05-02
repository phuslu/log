package verify

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/madkins23/go-slog/infra/warning"
	"github.com/madkins23/go-slog/verify/tests"

	"github.com/phuslu/log/go_slog/creator"
)

// TestVerifyPhusluSlog runs tests for the phuslu/slog handler.
func TestVerifyPhusluSlog(t *testing.T) {
	slogSuite := tests.NewSlogTestSuite(creator.PhusluSlog())
	slogSuite.WarnOnly(warning.Duplicates)
	slogSuite.WarnOnly(warning.DurationMillis)
	slogSuite.WarnOnly(warning.GroupAttrMsgTop)
	slogSuite.WarnOnly(warning.LevelVar)
	slogSuite.WarnOnly(warning.TimeMillis)
	slogSuite.WarnOnly(warning.ZeroPC)
	suite.Run(t, slogSuite)
}

func TestMain(m *testing.M) {
	warning.WithWarnings(m)
}
