package log

import (
	"errors"
	"testing"
	"time"
)

func TestJOSNLogger(t *testing.T) {
	log := JSONLogger{
		Level:      ParseLevel("debug"),
		EscapeHTML: false,
		Writer: &Writer{
			LocalTime: true,
		},
	}
	log.Info().
		Bool("bool", true).
		Dur("1_hour", time.Hour).
		Durs("hour_minute_second", []time.Duration{time.Hour, time.Minute, time.Second}).
		Err(errors.New("test error")).
		Float64("float32", 1.111).
		Float64("float64", 1.111).
		Int64("int64", 1234567890).
		Int32("int32", 123).
		Str("foobar", "\"<>?'").
		Time("now", timeNow()).
		Errs("errors", []error{errors.New("error1"), nil, errors.New("error3")}).
		Msgf("this is a \"%s\"", "test")
}

func TestJOSNLoggerCaller(t *testing.T) {
	log := JSONLogger{
		Level:  ParseLevel("debug"),
		Caller: true,
		Writer: &Writer{},
	}
	log.Info().Msg("this is caller log event")
}

func TestJOSNLoggerTime(t *testing.T) {
	log := JSONLogger{
		Level:      ParseLevel("debug"),
		TimeField:  "_time",
		TimeFormat: time.RFC850,
		Writer:     &Writer{},
	}
	log.Info().Timestamp().Time("now", timeNow()).Msg("this is time log event")
}
