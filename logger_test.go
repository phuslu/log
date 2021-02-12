package log

import (
	"bytes"
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

	Trace().Str("foo", "bar").Msg("hello from Trace")
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
		Dur("1_sec", time.Second+2*time.Millisecond+30*time.Microsecond+400*time.Nanosecond).
		Dur("1_sec", -time.Second+2*time.Millisecond+30*time.Microsecond+400*time.Nanosecond).
		Durs("hour_minute_second", []time.Duration{time.Hour, time.Minute, time.Second, -time.Second}).
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
		Uint64("uint64", 1234567890).
		Uint32("uint32", 123).
		Uint16("uint16", 123).
		Uint8("uint8", 123).
		Int64("int64", 1234567890).
		Int32("int32", 123).
		Int16("int16", 123).
		Int8("int8", 123).
		Int("int", 123).
		Uints64("uints64", []uint64{1234567890, 1234567890}).
		Uints32("uints32", []uint32{123, 123}).
		Uints16("uints16", []uint16{123, 123}).
		Uints8("uints8", []uint8{123, 123}).
		Uints("uints", []uint{123, 123}).
		Ints64("ints64", []int64{1234567890, 1234567890}).
		Ints32("ints32", []int32{123, 123}).
		Ints16("ints16", []int16{123, 123}).
		Ints8("ints8", []int8{123, 123}).
		Ints("ints", []int{123, 123}).
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
		Times("now_2", []time.Time{timeNow(), timeNow()}).
		TimeFormat("now_3", time.RFC3339, timeNow()).
		TimeFormat("now_3_1", TimeFormatUnix, timeNow()).
		TimeFormat("now_3_2", TimeFormatUnixMs, timeNow()).
		TimesFormat("now_4", time.RFC3339, []time.Time{timeNow(), timeNow()}).
		TimeDiff("time_diff_1", timeNow().Add(time.Second), timeNow()).
		TimeDiff("time_diff_2", time.Time{}, timeNow()).
		Stringer("ip_str", ipv4Addr).
		GoStringer("big_edian", binary.BigEndian).
		IPAddr("ip6", net.ParseIP("2001:4860:4860::8888")).
		IPAddr("ip4", ipv4Addr).
		IPPrefix("ip_prefix", *ipv4Net).
		MACAddr("mac", net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01}).
		Xid("xid", NewXID()).
		Errs("errors", []error{errors.New("error1"), nil, errors.New("error3")}).
		Interface("console_writer", ConsoleWriter{ColorOutput: true}).
		Interface("time.Time", timeNow()).
		KeysAndValues("foo", "bar", "number", 42).
		Msgf("this is a \"%s\"", "test")
}

func TestLoggerNil(t *testing.T) {
	e := Info()
	e.Caller(1).Str("foo", "bar").Int("num", 42).Msgf("this is a nil entry test")

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
		Uint64("uint64", 1234567890).
		Uint32("uint32", 123).
		Uint16("uint16", 123).
		Uint8("uint8", 123).
		Int64("int64", 1234567890).
		Int32("int32", 123).
		Int16("int16", 123).
		Int8("int8", 123).
		Int("int", 123).
		Uints64("uints64", []uint64{1234567890, 1234567890}).
		Uints32("uints32", []uint32{123, 123}).
		Uints16("uints16", []uint16{123, 123}).
		Uints8("uints8", []uint8{123, 123}).
		Uints("uints", []uint{123, 123}).
		Ints64("ints64", []int64{1234567890, 1234567890}).
		Ints32("ints32", []int32{123, 123}).
		Ints16("ints16", []int16{123, 123}).
		Ints8("ints8", []int8{123, 123}).
		Ints("ints", []int{123, 123}).
		RawJSON("raw_json", []byte("{\"a\":1,\"b\":2}")).
		RawJSONStr("raw_json", "{\"c\":1,\"d\":2}").
		Hex("hex", []byte("\"<>?'")).
		Bytes("bytes1", []byte("bytes1")).
		Bytes("bytes2", []byte("\"<>?'")).
		BytesOrNil("bytes3", []byte("\"<>?'")).
		BytesOrNil("bytes4", nil).
		Byte("zero", 0).
		Str("foobar", "\"\\\t\r\n\f\b\x00<>?'").
		Strs("strings", []string{"a", "b", "\"<>?'"}).
		Stringer("stringer", nil).
		Stringer("stringer", ipv4Addr).
		GoStringer("gostringer", nil).
		GoStringer("gostringer", binary.BigEndian).
		Time("now_1", timeNow()).
		Times("now_2", []time.Time{timeNow(), timeNow()}).
		TimeFormat("now_3", time.RFC3339, timeNow()).
		TimesFormat("now_4", time.RFC3339, []time.Time{timeNow(), timeNow()}).
		TimeDiff("time_diff_1", timeNow().Add(time.Second), timeNow()).
		TimeDiff("time_diff_2", time.Time{}, timeNow()).
		IPAddr("ip6", net.ParseIP("2001:4860:4860::8888")).
		IPAddr("ip4", ipv4Addr).
		IPPrefix("ip_prefix", *ipv4Net).
		MACAddr("mac", net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01}).
		Xid("xid", NewXID()).
		Errs("errors", []error{errors.New("error1"), nil, errors.New("error3")}).
		Interface("console_writer", ConsoleWriter{ColorOutput: true}).
		Interface("time.Time", timeNow()).
		KeysAndValues("foo", "bar", "number", 42).
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

type testMarshalObject struct {
	I int
	N string
}

func (o *testMarshalObject) MarshalObject(e *Entry) {
	e.Int("id", o.I).Str("name", o.N)
}

type nullMarshalObject struct {
	I int
	N string
}

func (o *nullMarshalObject) MarshalObject(e *Entry) {
}

func TestLoggerObject(t *testing.T) {
	logger := Logger{
		Level: ParseLevel("debug"),
	}

	logger.Info().Object("test_object", &testMarshalObject{1, "foo"}).Msg("this is a object test")
	logger.Info().EmbedObject(&testMarshalObject{1, "foo"}).Msg("this is a object test")
	logger.Info().Object("empty_object", nil).Msg("this is a empty_object test")
	logger.Info().EmbedObject(nil).Msg("this is a empty_object test")

	logger.Info().Object("null_object", &nullMarshalObject{3, "xxx"}).Msg("this is a empty_object test")
	logger.Info().EmbedObject(&nullMarshalObject{3, "xxx"}).Msg("this is a empty_object test")
}

func TestLoggerLog(t *testing.T) {
	logger := Logger{
		Level: ParseLevel("debug"),
	}

	logger.Log().Msgf("this is a no level log")

	logger.Caller = 1
	logger.Log().Msgf("this is a no level log with caller")
}

func TestLoggerByte(t *testing.T) {
	logger := Logger{
		Level: ParseLevel("debug"),
	}

	logger.Info().Byte("gender", 'm').Msg("")
	logger.Info().Byte("quote", '"').Msg("")
	logger.Info().Byte("reverse", '\\').Msg("")
	logger.Info().Byte("cf", '\n').Msg("")
	logger.Info().Byte("cr", '\r').Msg("")
	logger.Info().Byte("tab", '\t').Msg("")
	logger.Info().Byte("forward", '\f').Msg("")
	logger.Info().Byte("back", '\b').Msg("")
	logger.Info().Byte("less", '<').Msg("")
	logger.Info().Byte("singlequote", '\'').Msg("")
	logger.Info().Byte("zerobyte", '\x00').Msg("")
}

func TestLoggerSetLevel(t *testing.T) {
	DefaultLogger.SetLevel(InfoLevel)
	Warn().Msg("1. i am a warn log")
	Info().Msg("2. i am a info log")
	Debug().Msg("3. i am a debug log")
	Trace().Msg("3. i am a trace log")
	DefaultLogger.SetLevel(TraceLevel)
	Info().Msg("4. i am a info log")
	Debug().Msg("5. i am a debug log")
	Trace().Msg("5. i am a trace log")
}

func TestLoggerStack(t *testing.T) {
	Info().Stack().Msg("this is single stack log entry")
}

func TestLoggerEnabled(t *testing.T) {
	DefaultLogger.SetLevel(InfoLevel)
	Debug().Stack().Msgf("hello %s", "world")
	if Debug().Enabled() {
		t.Fatal("debug level should enabled")
	}
}

func TestLoggerDiscard(t *testing.T) {
	Info().Stack().Str("foo", "bar").Discard()
	DefaultLogger.SetLevel(InfoLevel)
	Debug().Stack().Str("foo", "bar").Discard()
}

func TestLoggerWithLevel(t *testing.T) {
	DefaultLogger.WithLevel(InfoLevel).Msg("this is with level log entry")
	DefaultLogger.Caller = 1
	DefaultLogger.WithLevel(InfoLevel).Msg("this is with level caller log entry")
}

func TestLoggerCaller(t *testing.T) {
	notTest = false

	DefaultLogger.Caller = 1
	DefaultLogger.SetLevel(TraceLevel)
	Trace().Str("foo", "bar").Msg("hello from Trace")
	Debug().Str("foo", "bar").Msg("hello from Debug")
	Info().Str("foo", "bar").Msg("hello from Info")
	Warn().Str("foo", "bar").Msg("hello from Warn")
	Error().Str("foo", "bar").Msg("hello from Error")
	Fatal().Str("foo", "bar").Msg("hello from Fatal")
	Panic().Str("foo", "bar").Msg("hello from Panic")
	Printf("hello from %s", "Printf")

	logger := Logger{
		Level:  ParseLevel("trace"),
		Caller: 1,
	}
	logger.Trace().Str("foo", "bar").Msg("hello from Trace")
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
	logger.Info().Int64("timestamp_ms", timeNow().UnixNano()/1000000).Msg("this is unix time log entry")

	logger.TimeFormat = TimeFormatUnixMs
	logger.Info().Int64("timestamp_ms", timeNow().UnixNano()/1000000).Msg("this is unix_ms time log entry")

	logger.TimeFormat = time.RFC3339Nano
	logger.Info().Int64("timestamp_ms", timeNow().UnixNano()/1000000).Msg("this is rfc3339 time log entry")
}

func TestLoggerTimeOffset(t *testing.T) {
	logger := Logger{}

	timeOffset = -7 * 3600
	timeZone = "-07:00"

	logger.Info().Msg("this is -7:00 timezone time log entry")
}

func TestLoggerContext(t *testing.T) {
	ctx := NewContext(nil).Bool("ctx_bool", true).Str("ctx_str", "ctx str").Value()

	logger := Logger{Level: InfoLevel}
	logger.Trace().Context(ctx).Int("no0", 0).Msg("this is zero context log entry")
	logger.Debug().Context(ctx).Int("no0", 0).Msg("this is zero context log entry")
	logger.Info().Context(ctx).Int("no1", 1).Msg("this is first context log entry")
	logger.Info().Context(ctx).Int("no2", 2).Msg("this is second context log entry")
}

func TestLoggerContextDict(t *testing.T) {
	ctx := NewContext(nil).Bool("ctx_bool", true).Str("ctx_str", "ctx str").Value()

	logger := Logger{Level: InfoLevel, Writer: &ConsoleWriter{ColorOutput: true}}
	logger.Trace().Dict("akey", ctx).Int("no0", 0).Msg("this is zero dict log entry")
	logger.Debug().Dict("akey", ctx).Int("no0", 0).Msg("this is zero dict log entry")
	logger.Info().Dict("akey", ctx).Int("no1", 1).Msg("this is first dict log entry")
	logger.Info().
		Dict("a", NewContext(nil).
			Bool("b", true).
			Dict("c", NewContext(nil).
				Bool("d", true).
				Str("e", "a str").
				Value()).
			Value()).
		Msg("")

	ctx = NewContext(nil).Value()
	logger.Trace().Dict("akey", ctx).Int("no0", 0).Msg("this is zero dict log entry")
	logger.Debug().Dict("akey", ctx).Int("no0", 0).Msg("this is zero dict log entry")
	logger.Info().Dict("akey", ctx).Int("no1", 1).Msg("this is first dict log entry")
}

func TestLoggerFields(t *testing.T) {
	ipv4Addr, ipv4Net, err := net.ParseCIDR("192.0.2.1/24")
	if err != nil {
		t.Fatalf("net.ParseCIDR error: %+v", err)
	}

	logger := Logger{
		Level:  InfoLevel,
		Caller: 1,
		Writer: &ConsoleWriter{ColorOutput: true, EndWithMessage: true},
	}

	logger.Info().Fields(map[string]interface{}{
		"bool":               true,
		"bools":              []bool{false},
		"bools_2":            []bool{true, false},
		"1_hour":             time.Hour,
		"hour_minute_second": []time.Duration{time.Hour, time.Minute, time.Second},
		"error":              errors.New("test error"),
		"an_error":           fmt.Errorf("an %w", errors.New("test error")),
		"an_nil_error":       nil,
		"dict":               NewContext(nil).Str("foo", "bar").Int("no", 1).Value(),
		"float32":            float32(1.111),
		"float32_2":          []float32{1.111},
		"float32_3":          []float32{1.111, 2.222},
		"float64":            float64(1.111),
		"float64_2":          []float64{1.111, 2.222},
		"int64":              int64(1234567890),
		"int32":              int32(123),
		"int16":              int16(123),
		"int8":               int8(123),
		"int":                int(123),
		"uint64":             uint64(1234567890),
		"uint32":             uint32(123),
		"uint16":             uint16(123),
		"uint8":              uint8(123),
		"uint":               uint(123),
		"raw_json":           []byte("{\"a\":1,\"b\":2}"),
		"hex":                []byte("\"<>?'"),
		"bytes1":             []byte("bytes1"),
		"bytes2":             []byte("\"<>?'"),
		"foobar":             "\"\\\t\r\n\f\b\x00<>?'",
		"strings":            []string{"a", "b", "\"<>?'"},
		"stringer_1":         nil,
		"stringer_2":         ipv4Addr,
		"gostringer_1":       nil,
		"gostringer_2":       binary.BigEndian,
		"now_1":              timeNow(),
		"ip_str":             ipv4Addr,
		"big_edian":          binary.BigEndian,
		"ip6":                net.ParseIP("2001:4860:4860::8888"),
		"ip4":                ipv4Addr,
		"ip_prefix":          *ipv4Net,
		"mac":                net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01},
		"errors":             []error{errors.New("error1"), nil, errors.New("error3")},
		"console_writer":     ConsoleWriter{ColorOutput: true},
		"time.Time":          timeNow(),
		"buffer":             bytes.NewBuffer([]byte("a_bytes_buffer")),
	}).Msg("this is a fields test")
}

type errno uint

func (e errno) Error() string {
	return fmt.Sprintf("errno: %d", e)
}

func (e errno) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if e == 0 {
			fmt.Fprintf(s, "%d", e)
			return
		}
		fmt.Fprintf(s, "stack layer: %v\t", e-1)
	case 's':
		fmt.Fprintf(s, "errno: %d", e)
	}
}

func TestLoggerErrorStack(t *testing.T) {
	logger := Logger{Level: TraceLevel, Writer: &ConsoleWriter{ColorOutput: true}}
	logger.Info().Err(errno(0)).Msg("log errno(0) here")
	logger.Info().Err(errno(1)).Msg("log errno(1) here")
	logger.Info().Err(errno(2)).Msg("log errno(2) here")
	logger.Info().Err(errno(3)).Msg("log errno(3) here")
}

func BenchmarkLogger(b *testing.B) {
	logger := Logger{
		TimeFormat: TimeFormatUnix,
		Level:      DebugLevel,
		Writer:     IOWriter{ioutil.Discard},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("foo", "bar").Msgf("hello %s", "world")
	}
}
