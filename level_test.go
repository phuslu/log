package log

import (
	"testing"
)

func TestLevelParse(t *testing.T) {
	cases := []struct {
		Level  Level
		String string
	}{
		{TraceLevel, "trace"},
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{FatalLevel, "fatal"},
		{PanicLevel, "panic"},
		{noLevel, "????"},
	}

	for _, c := range cases {
		if v := ParseLevel(c.String); v != c.Level {
			t.Errorf("ParseLevel(%#v) must return %#v, not %#v", c.String, c.Level, v)
		}
		if v := c.Level.String(); v != c.String {
			t.Errorf("%T.String() must return %#v, not %#v", c.Level, c.String, v)
		}
	}
}
