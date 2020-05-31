package log

import (
	"errors"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

func TestDefaultLogger(t *testing.T) {
	osExit = func(int) {}

	Debug().Str("foo", "bar").Msg("hello from Debug")
	Info().Str("foo", "bar").Msg("hello from Info")
	Warn().Str("foo", "bar").Msg("hello from Warn")
	Error().Str("foo", "bar").Msg("hello from Error")
	Fatal().Str("foo", "bar").Msg("hello from Fatal")
	Print("hello from Print")
	Printf("hello from %s", "Printf")
}

func TestLogger(t *testing.T) {
	ipv4Addr, ipv4Net, err := net.ParseCIDR("192.0.2.1/24")
	if err != nil {
		t.Fatalf("net.ParseCIDR error: %+v", err)
	}

	logger := Logger{
		Level: ParseLevel("debug"),
	}
	logger.Info().
		Caller(1).
		Bool("bool", true).
		Bools("bools", []bool{false}).
		Bools("bools", []bool{true, false}).
		Dur("1_hour", time.Hour).
		Durs("hour_minute_second", []time.Duration{time.Hour, time.Minute, time.Second}).
		Err(errors.New("test error")).
		Err(nil).
		Float32("float32", 1.111).
		Floats32("float32", []float32{1.111}).
		Floats32("float32", []float32{1.111, 2.222}).
		Float64("float64", 1.111).
		Floats64("float64", []float64{1.111, 2.222}).
		Uint64("int64", 1234567890).
		Uint32("int32", 123).
		Uint16("int16", 123).
		Uint8("int16", 123).
		Int64("int64", 1234567890).
		Int32("int32", 123).
		Int16("int16", 123).
		Int8("int16", 123).
		Int("int", 123).
		RawJSON("raw_json", []byte("{\"a\":1,\"b\":2}")).
		Hex("hex", []byte("\"<>?'")).
		Bytes("bytes1", []byte("bytes1")).
		Bytes("bytes2", []byte("\"<>?'")).
		Str("foobar", "\"\\\t\r\n\f\b\x00<>?'").
		Strs("strings", []string{"a", "b", "\"<>?'"}).
		Time("now_1", timeNow()).
		TimeFormat("now_2", time.RFC3339, timeNow()).
		TimeDiff("time_diff_1", timeNow().Add(time.Second), timeNow()).
		TimeDiff("time_diff_2", time.Time{}, timeNow()).
		IPAddr("ip6", net.ParseIP("2001:4860:4860::8888")).
		IPAddr("ip4", ipv4Addr).
		IPPrefix("ip_prefix", *ipv4Net).
		MACAddr("mac", net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01}).
		Errs("errors", []error{errors.New("error1"), nil, errors.New("error3")}).
		Interface("console_writer", ConsoleWriter{ANSIColor: true}).
		Interface("time.Time", timeNow()).
		Msgf("this is a \"%s\"", "test")
}

func TestLoggerNil(t *testing.T) {
	e := Info()
	e.buf = nil
	e.Caller(1).Str("foo", "bar").Int("num", 42).Msgf("this is a nil event test")

	ipv4Addr, ipv4Net, err := net.ParseCIDR("192.0.2.1/24")
	if err != nil {
		t.Fatalf("net.ParseCIDR error: %+v", err)
	}

	logger := Logger{
		Level: ParseLevel("info"),
	}
	logger.Debug().
		Caller(1).
		Bool("bool", true).
		Bools("bools", []bool{true, false}).
		Dur("1_hour", time.Hour).
		Durs("hour_minute_second", []time.Duration{time.Hour, time.Minute, time.Second}).
		Err(errors.New("test error")).
		Err(nil).
		Float32("float32", 1.111).
		Floats32("float32", []float32{1.111}).
		Float64("float64", 1.111).
		Floats64("float64", []float64{1.111}).
		Floats64("float64", []float64{1.111}).
		Uint64("int64", 1234567890).
		Uint32("int32", 123).
		Uint16("int16", 123).
		Uint8("int16", 123).
		Int64("int64", 1234567890).
		Int32("int32", 123).
		Int16("int16", 123).
		Int8("int16", 123).
		Int("int", 123).
		RawJSON("raw_json", []byte("{\"a\":1,\"b\":2}")).
		Hex("hex", []byte("\"<>?'")).
		Bytes("bytes1", []byte("bytes1")).
		Bytes("bytes2", []byte("\"<>?'")).
		Str("foobar", "\"\\\t\r\n\f\b\x00<>?'").
		Strs("strings", []string{"a", "b", "\"<>?'"}).
		Time("now_1", timeNow()).
		TimeFormat("now_2", time.RFC3339, timeNow()).
		TimeDiff("time_diff_1", timeNow().Add(time.Second), timeNow()).
		TimeDiff("time_diff_2", time.Time{}, timeNow()).
		IPAddr("ip6", net.ParseIP("2001:4860:4860::8888")).
		IPAddr("ip4", ipv4Addr).
		IPPrefix("ip_prefix", *ipv4Net).
		MACAddr("mac", net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01}).
		Errs("errors", []error{errors.New("error1"), nil, errors.New("error3")}).
		Interface("console_writer", ConsoleWriter{ANSIColor: true}).
		Interface("time.Time", timeNow()).
		Msgf("this is a \"%s\"", "test")
}

func TestLoggerInterface(t *testing.T) {
	logger := Logger{
		Level: ParseLevel("debug"),
	}

	var cyclicStruct struct {
		Value interface{}
	}

	cyclicStruct.Value = &cyclicStruct

	logger.Info().
		Caller(1).
		Interface("a_cyclic_struct", cyclicStruct).
		Msgf("this is a cyclic struct test")
}

func TestLoggerSetLevel(t *testing.T) {
	DefaultLogger.SetLevel(InfoLevel)
	Warn().Msg("1. i am a warn log")
	Info().Msg("2. i am a info log")
	Debug().Msg("3. i am a debug log")
	DefaultLogger.SetLevel(DebugLevel)
	Info().Msg("4. i am a info log")
	Debug().Msg("5. i am a debug log")
}

func TestLoggerStack(t *testing.T) {
	Info().Stack(false).Msg("this is single stack log event")
	Info().Stack(true).Msg("this is full stack log event")
}

func TestLoggerEnabled(t *testing.T) {
	DefaultLogger.SetLevel(InfoLevel)
	Debug().Stack(false).Msgf("hello %s", "world")
	if Debug().Enabled() {
		t.Fatal("debug level should enabled")
	}
}

func TestLoggerDiscard(t *testing.T) {
	Info().Stack(false).Str("foo", "bar").Discard()
	DefaultLogger.SetLevel(InfoLevel)
	Debug().Stack(false).Str("foo", "bar").Discard()
}

func TestLoggerWithLevel(t *testing.T) {
	DefaultLogger.WithLevel(InfoLevel).Msg("this is with level log event")
	DefaultLogger.Caller = 1
	DefaultLogger.WithLevel(InfoLevel).Msg("this is with level caller log event")
}

func TestLoggerCaller(t *testing.T) {
	osExit = func(int) {}

	DefaultLogger.Caller = 1
	DefaultLogger.SetLevel(DebugLevel)
	Debug().Str("foo", "bar").Msg("hello from Debug")
	Info().Str("foo", "bar").Msg("hello from Info")
	Warn().Str("foo", "bar").Msg("hello from Warn")
	Error().Str("foo", "bar").Msg("hello from Error")
	Fatal().Str("foo", "bar").Msg("hello from Fatal")
	Print("hello from Print")
	Printf("hello from %s", "Printf")

	logger := Logger{
		Level:  ParseLevel("debug"),
		Caller: 1,
	}
	logger.Debug().Str("foo", "bar").Msg("hello from Debug")
	logger.Info().Str("foo", "bar").Msg("hello from Info")
	logger.Warn().Str("foo", "bar").Msg("hello from Warn")
	logger.Error().Str("foo", "bar").Msg("hello from Error")
	logger.Fatal().Str("foo", "bar").Msg("hello from Fatal")
	logger.Print("hello from Print")
	logger.Printf("hello from %s", "Printf")
}

func TestLoggerTime(t *testing.T) {
	logger1 := Logger{
		Level:     ParseLevel("debug"),
		TimeField: "_time",
	}
	logger1.Info().Time("now", timeNow()).Msg("this is test time log event")
	logger2 := Logger{
		Level:      ParseLevel("debug"),
		TimeField:  "_time",
		TimeFormat: time.RFC822,
	}
	logger2.Info().Time("now", timeNow()).Msg("this is test time log event")
}

func TestLoggerTimestamp(t *testing.T) {
	logger := Logger{
		Level:     ParseLevel("debug"),
		Timestamp: true,
	}
	logger.Info().Int64("timestamp_ms", timeNow().UnixNano()/1000000).Msg("this is test time log event")
}

func TestLoggerHost(t *testing.T) {
	logger := Logger{
		Level:     ParseLevel("debug"),
		HostField: "host",
	}
	logger.Info().Time("now", timeNow()).Msg("this is test host log event")
}

func BenchmarkLogger(b *testing.B) {
	logger := Logger{
		Timestamp: true,
		Level:     DebugLevel,
		Writer:    ioutil.Discard,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("foo", "bar").Msgf("hello %s", "world")
	}
}
