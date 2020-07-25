package log

import (
	"testing"
)

func TestLevel(t *testing.T) {
	cases := []struct {
		Level Level
		Lower string
		Upper string
		Title string
		Three string
		One   string
	}{
		{DebugLevel, "debug", "DEBUG", "Debug", "DBG", "D"},
		{InfoLevel, "info", "INFO", "Info", "INF", "I"},
		{WarnLevel, "warn", "WARN", "Warn", "WRN", "W"},
		{ErrorLevel, "error", "ERROR", "Error", "ERR", "E"},
		{FatalLevel, "fatal", "FATAL", "Fatal", "FTL", "F"},
		{PanicLevel, "panic", "PANIC", "Panic", "PNC", "P"},
		{noLevel, "????", "????", "????", "???", "?"},
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
		if v := c.Level.One(); v != c.One {
			t.Errorf("%T.One() must return %#v, not %#v", c.Level, c.One, v)
		}
	}
}
