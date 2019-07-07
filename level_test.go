package log

import (
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		Level  Level
		String string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{FatalLevel, "fatal"},
	}

	for _, c := range cases {
		if v := ParseLevel(c.String); v != c.Level {
			t.Errorf("ParseLevel(%#v) must return %#v, not %#v", c.String, c.Level, v)
		}
	}
}
