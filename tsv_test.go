package log

import (
	"testing"
)

func TestTSVLogger(t *testing.T) {
	log := TSVLogger{
		Separator: ',',
		Writer:    &Writer{},
	}
	log.New().
		Timestamp().
		Bool(true).
		Float64(1.111).
		Float64(1.111).
		Int64(1234567890).
		Int32(123456).
		Str("\"<,\t>?'").
		Send()
}
