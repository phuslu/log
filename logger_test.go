package log

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

func TestLoggerDefault(t *testing.T) {
	notTest = false

	Debug().Str("foo", "bar").Msg("hello from Debug")
	Info().Str("foo", "bar").Msg("hello from Info")
	Warn().Str("foo", "bar").Msg("hello from Warn")
	Error().Str("foo", "bar").Msg("hello from Error")
	Fatal().Str("foo", "bar").Msg("hello from Fatal")
	Panic().Str("foo", "bar").Msg("hello from Panic")
	Printf("hello from %s", "Printf")
}

func TestLoggerInfo(t *testing.T) {
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
		AnErr("an_error", fmt.Errorf("an %w", errors.New("test error"))).
		AnErr("an_error", nil).
		Int64("goid", Goid()).
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
		RawJSONStr("raw_json", "{\"c\":1,\"d\":2}").
		Hex("hex", []byte("\"<>?'")).
		Bytes("bytes1", []byte("bytes1")).
		Bytes("bytes2", []byte("\"<>?'")).
		BytesOrNil("bytes3", []byte("\"<>?'")).
		Bytes("nil_bytes_1", nil).
		BytesOrNil("nil_bytes_2", nil).
		Str("foobar", "\"\\\t\r\n\f\b\x00<>?'").
		Strs("strings", []string{"a", "b", "\"<>?'"}).
		Stringer("stringer", nil).
		Stringer("stringer", ipv4Addr).
		GoStringer("gostringer", nil).
		GoStringer("gostringer", binary.BigEndian).
		Time("now_1", timeNow()).
		TimeFormat("now_2", time.RFC3339, timeNow()).
		TimeDiff("time_diff_1", timeNow().Add(time.Second), timeNow()).
		TimeDiff("time_diff_2", time.Time{}, timeNow()).
		Stringer("ip_str", ipv4Addr).
		GoStringer("big_edian", binary.BigEndian).
		IPAddr("ip6", net.ParseIP("2001:4860:4860::8888")).
		IPAddr("ip4", ipv4Addr).
		IPPrefix("ip_prefix", *ipv4Net).
		MACAddr("mac", net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01}).
		Xid("xid", [12]byte{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}).
		Errs("errors", []error{errors.New("error1"), nil, errors.New("error3")}).
		Interface("console_writer", ConsoleWriter{ColorOutput: true}).
		Interface("time.Time", timeNow()).
		kvs("foo", "bar", "number", 42).
		Msgf("this is a \"%s\"", "test")
}

func TestLoggerNil(t *testing.T) {
	e := Info()
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
		AnErr("an_error", fmt.Errorf("an %w", errors.New("test error"))).
		AnErr("an_error", nil).
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
		RawJSONStr("raw_json", "{\"c\":1,\"d\":2}").
		Hex("hex", []byte("\"<>?'")).
		Bytes("bytes1", []byte("bytes1")).
		Bytes("bytes2", []byte("\"<>?'")).
		BytesOrNil("bytes3", []byte("\"<>?'")).
		BytesOrNil("bytes4", nil).
		Str("foobar", "\"\\\t\r\n\f\b\x00<>?'").
		Strs("strings", []string{"a", "b", "\"<>?'"}).
		Stringer("stringer", nil).
		Stringer("stringer", ipv4Addr).
		GoStringer("gostringer", nil).
		GoStringer("gostringer", binary.BigEndian).
		Time("now_1", timeNow()).
		TimeFormat("now_2", time.RFC3339, timeNow()).
		TimeDiff("time_diff_1", timeNow().Add(time.Second), timeNow()).
		TimeDiff("time_diff_2", time.Time{}, timeNow()).
		IPAddr("ip6", net.ParseIP("2001:4860:4860::8888")).
		IPAddr("ip4", ipv4Addr).
		IPPrefix("ip_prefix", *ipv4Net).
		MACAddr("mac", net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01}).
		Xid("xid", [12]byte{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}).
		Errs("errors", []error{errors.New("error1"), nil, errors.New("error3")}).
		Interface("console_writer", ConsoleWriter{ColorOutput: true}).
		Interface("time.Time", timeNow()).
		kvs("foo", "bar", "number", 42).
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
	notTest = false

	DefaultLogger.Caller = 1
	DefaultLogger.SetLevel(DebugLevel)
	Debug().Str("foo", "bar").Msg("hello from Debug")
	Info().Str("foo", "bar").Msg("hello from Info")
	Warn().Str("foo", "bar").Msg("hello from Warn")
	Error().Str("foo", "bar").Msg("hello from Error")
	Fatal().Str("foo", "bar").Msg("hello from Fatal")
	Panic().Str("foo", "bar").Msg("hello from Panic")
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
	logger.Panic().Str("foo", "bar").Msg("hello from Panic")
	logger.Printf("hello from %s", "Printf")
}

func TestLoggerTimeField(t *testing.T) {
	logger := Logger{}

	logger.TimeField = "_time"
	logger.Printf("this is no level and _time field log")
}

func TestLoggerTimeFormat(t *testing.T) {
	logger := Logger{}

	logger.TimeFormat = TimeFormatUnix
	logger.Info().Int64("timestamp_ms", timeNow().UnixNano()/1000000).Msg("this is unix time log event")

	logger.TimeFormat = TimeFormatUnixMs
	logger.Info().Int64("timestamp_ms", timeNow().UnixNano()/1000000).Msg("this is unix_ms time log event")

	logger.TimeFormat = time.RFC3339Nano
	logger.Info().Int64("timestamp_ms", timeNow().UnixNano()/1000000).Msg("this is rfc3339 time log event")
}

func TestLoggerTimeOffset(t *testing.T) {
	logger := Logger{}

	timeOffset = -7 * 3600
	timeZone = "-07:00"

	logger.Info().Msg("this is -7:00 timezone time log event")
}

func TestLoggerContext(t *testing.T) {
	ctx := NewContext().Bool("ctx_bool", true).Str("ctx_str", "ctx str").Value()

	logger := Logger{Level: InfoLevel}
	logger.Debug().Context(ctx).Int("no0", 0).Msg("this is zero context log event")
	logger.Info().Context(ctx).Int("no1", 1).Msg("this is first context log event")
	logger.Info().Context(ctx).Int("no2", 2).Msg("this is second context log event")
}

func TestLoggerContextDict(t *testing.T) {
	ctx := NewContext().Bool("ctx_bool", true).Str("ctx_str", "ctx str").Value()

	logger := Logger{Level: InfoLevel, Writer: &ConsoleWriter{ColorOutput: true}}
	logger.Debug().Dict("akey", ctx).Int("no0", 0).Msg("this is zero dict log event")
	logger.Info().Dict("akey", ctx).Int("no1", 1).Msg("this is first dict log event")
	logger.Info().
		Dict("a", NewContext().
			Bool("b", true).
			Dict("c", NewContext().
				Bool("d", true).
				Str("e", "a str").
				Value()).
			Value()).
		Msg("")

	ctx = NewContext().Value()
	logger.Debug().Dict("akey", ctx).Int("no0", 0).Msg("this is zero dict log event")
	logger.Info().Dict("akey", ctx).Int("no1", 1).Msg("this is first dict log event")
}

func BenchmarkLogger(b *testing.B) {
	logger := Logger{
		TimeFormat: TimeFormatUnix,
		Level:      DebugLevel,
		Writer:     ioutil.Discard,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("foo", "bar").Msgf("hello %s", "world")
	}
}
