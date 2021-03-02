package log

import (
	"fmt"
	"io"
	"testing"
)

func TestFormatterParse(t *testing.T) {
	var jsons = []string{
		`{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"这是一个🌐哦\n","foo":"bar","n":42,"t":true,"f":false,"o":null,"a":[1,2,3],"obj":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer\t123"}`,
		`{"ts":1234567890,"level":"info","caller":"pretty.go:42","foo":"haha","message":"hello self-define time field\t\n"}`,
	}

	for _, s := range jsons {
		var args FormatterArgs
		parseFormatterArgs([]byte(s), &args)
		t.Logf("%#v", args)
		t.Logf("foo=%v", args.Get("foo"))
	}
}

func TestFormatterDefault(t *testing.T) {
	DefaultLogger.Writer = &ConsoleWriter{
		Formatter: func(w io.Writer, a *FormatterArgs) (int, error) {
			return fmt.Fprintf(w, "%s\n", a.Message)
		},
	}

	Info().Msg("aaaa 'b' cccc")
}
