package log

import (
	"testing"
)

func TestLevelParse(t *testing.T) {
	cases := []struct {
		Level Level
		Lower string
		Upper string
		Title string
		Three string
	}{
		{TraceLevel, "trace", "TRACE", "Trace", "TRC"},
		{DebugLevel, "debug", "DEBUG", "Debug", "DBG"},
		{InfoLevel, "info", "INFO", "Info", "INF"},
		{WarnLevel, "warn", "WARN", "Warn", "WRN"},
		{ErrorLevel, "error", "ERROR", "Error", "ERR"},
		{FatalLevel, "fatal", "FATAL", "Fatal", "FTL"},
		{PanicLevel, "panic", "PANIC", "Panic", "PNC"},
		{noLevel, "????", "????", "????", "???"},
	}

	for _, c := range cases {
		if v := ParseLevel(c.Lower); v != c.Level {
			t.Errorf("ParseLevel(%#v) must return %#v, not %#v", c.Lower, c.Level, v)
		}
		if v := c.Level.Lower(); v != c.Lower {
			t.Errorf("%T.Lower() must return %#v, not %#v", c.Level, c.Lower, v)
		}
		if v := c.Level.Upper(); v != c.Upper {
			t.Errorf("%T.Upper() must return %#v, not %#v", c.Level, c.Upper, v)
		}
		if v := c.Level.Title(); v != c.Title {
			t.Errorf("%T.Title() must return %#v, not %#v", c.Level, c.Title, v)
		}
		if v := c.Level.Three(); v != c.Three {
			t.Errorf("%T.Three() must return %#v, not %#v", c.Level, c.Three, v)
		}
	}
}

func TestLevelParseByte(t *testing.T) {
	cases := []struct {
		Level Level
		Byte1 byte
		Byte2 byte
	}{
		{TraceLevel, 't', 'T'},
		{DebugLevel, 'd', 'D'},
		{InfoLevel, 'i', 'I'},
		{WarnLevel, 'w', 'W'},
		{ErrorLevel, 'e', 'E'},
		{FatalLevel, 'f', 'F'},
		{PanicLevel, 'p', 'P'},
		{noLevel, 'x', '?'},
	}

	for _, c := range cases {
		if v := ParseLevelByte(c.Byte1); v != c.Level {
			t.Errorf("ParseLevel(%#v) must return %#v, not %#v", c.Byte1, c.Level, v)
		}
		if v := ParseLevelByte(c.Byte2); v != c.Level {
			t.Errorf("ParseLevel(%#v) must return %#v, not %#v", c.Byte2, c.Level, v)
		}
	}
}
