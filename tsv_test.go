package log

import (
	"io/ioutil"
	"net"
	"testing"
)

func TestTSVLogger(t *testing.T) {
	logger := TSVLogger{}

	logger.New().
		Timestamp().
		TimestampMS().
		Bool(true).
		Bool(false).
		Byte('m').
		Float64(0.618).
		Int64(123).
		Uint64(456).
		Float32(12.0).
		Int(42).
		Int32(42).
		Int16(42).
		Int8(42).
		Uint32(42).
		Uint16(42).
		Uint8(42).
		Bytes([]byte("\"<,\t>?'")).
		Str("\"<,\t>?'").
		IPAddr(net.IP{1, 11, 111, 200}).
		IPAddr(net.ParseIP("2001:4860:4860::8888")).
		Msg()

}

func TestTSVSeparator(t *testing.T) {
	logger := TSVLogger{
		Separator: '¥',
	}

	logger.New().Msg()

	logger.New().
		TimestampMS().
		Bool(true).
		Bool(false).
		Byte('m').
		Float64(0.618).
		Int64(123).
		Uint64(456).
		Float32(12.0).
		Int(42).
		Int32(42).
		Int16(42).
		Int8(42).
		Uint32(42).
		Uint16(42).
		Uint8(42).
		Bytes([]byte("\"<,\t>?'")).
		Str("\"<,\t>?'").
		Msg()
}

func TestTSVDiscard(t *testing.T) {
	logger := TSVLogger{
		Writer: ioutil.Discard,
	}

	logger.New().
		TimestampMS().
		Bool(true).
		Bool(false).
		Float64(0.618).
		Int64(123).
		Uint64(456).
		Float32(12.0).
		Int(42).
		Int32(42).
		Int16(42).
		Int8(42).
		Uint32(42).
		Uint16(42).
		Uint8(42).
		Bytes([]byte("\"<,\t>?'")).
		Str("\"<,\t>?'").
		Msg()
}

func BenchmarkTSVLogger(b *testing.B) {
	logger := TSVLogger{
		Writer: ioutil.Discard,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.New().TimestampMS().Str("a tsv message").Msg()
	}
}
