package log

import (
	"testing"
)

func TestRandomLogger(t *testing.T) {
	logger := RandomLogger{
		Logger: DefaultLogger,
		N:      10,
	}

	for i := 0; i < 100; i++ {
		logger.Info().Int("i", i).Msg("hello from random logger")
	}
}
