package log

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestLoggerSugar(t *testing.T) {
	ipv4Addr, ipv4Net, err := net.ParseCIDR("192.0.2.1/24")
	if err != nil {
		t.Fatalf("net.ParseCIDR error: %+v", err)
	}

	logger := Logger{
		Level:  ParseLevel("info"),
		Caller: 1,
		Writer: &ConsoleWriter{ColorOutput: true},
	}

	sugar := logger.Sugar(DebugLevel, NewContext().Str("tag", "hi sugar").Value())
	logger.Level = InfoLevel
	sugar.Print("hello from sugar Print")
	sugar.Println("hello from sugar Println")
	sugar.Printf("hello from sugar %s", "Printf")
	sugar.Log("foo", "bar")

	sugar = logger.Sugar(InfoLevel, NewContext().Str("tag", "hi sugar").Value())
	sugar.Print("hello from sugar Print")
	sugar.Println("hello from sugar Println")
	sugar.Printf("hello from sugar %s", "Printf")
	sugar.Log(
		"bool", true,
		"bools", []bool{false},
		"bools", []bool{true, false},
		"1_hour", time.Hour,
		"hour_minute_second", []time.Duration{time.Hour, time.Minute, time.Second},
		"error", errors.New("test error"),
		"an_error", fmt.Errorf("an %w", errors.New("test error")),
		"an_nil_error", nil,
		"dict", NewContext().Str("foo", "bar").Int("no", 1).Value(),
		"float32", float32(1.111),
		"float32", []float32{1.111},
		"float32", []float32{1.111, 2.222},
		"float64", float64(1.111),
		"float64", []float64{1.111, 2.222},
		"int64", int64(1234567890),
		"int32", int32(123),
		"int16", int16(123),
		"int8", int8(123),
		"int", int(123),
		"uint64", uint64(1234567890),
		"uint32", uint32(123),
		"uint16", uint16(123),
		"uint8", uint8(123),
		"uint", uint(123),
		"raw_json", []byte("{\"a\":1,\"b\":2}"),
		"hex", []byte("\"<>?'"),
		"bytes1", []byte("bytes1"),
		"bytes2", []byte("\"<>?'"),
		"foobar", "\"\\\t\r\n\f\b\x00<>?'",
		"strings", []string{"a", "b", "\"<>?'"},
		"stringer_1", nil,
		"stringer_2", ipv4Addr,
		"gostringer_1", nil,
		"gostringer_2", binary.BigEndian,
		"now_1", timeNow(),
		"ip_str", ipv4Addr,
		"big_edian", binary.BigEndian,
		"ip6", net.ParseIP("2001:4860:4860::8888"),
		"ip4", ipv4Addr,
		"ip_prefix", *ipv4Net,
		"mac", net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01},
		"errors", []error{errors.New("error1"), nil, errors.New("error3")},
		"console_writer", ConsoleWriter{ColorOutput: true},
		"time.Time", timeNow(),
		"buffer", bytes.NewBuffer([]byte("a_bytes_buffer")),
		"message", "this is a test")
}
