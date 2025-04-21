package log

import (
	"fmt"
	"io"
	"testing"
)

func TestFormatterParse(t *testing.T) {
	var jsons = []string{
		`{"time":"2019-07-10T05:35:54.277Z","level":"debug","level":"error","caller":"pretty.go:42","error":"这是一个🌐哦\n","foo":"bar","n":42,"t":true,"f":false,"o":null,"a":[1,2,3],"obj":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer\t123"}`,
		`{"ts":1234567890,"level":"info","caller":"pretty.go:42","foo":"bad value","foo":"haha","message":"hello self-define time field\t\n"}`,
	}

	for _, s := range jsons {
		var args FormatterArgs
		parseFormatterArgs([]byte(s), &args)
		t.Logf("%+v", args)
		t.Logf("foo=%v", args.Get("foo"))
	}
}

func TestFormatterArgsParse(t *testing.T) {
	timestamp := "2019-07-10T05:35:54.277Z"
	level := "debug"
	msg := "hello json console color writer\t123"
	category := "cat1"
	var json = `{"time":"` + timestamp + `","level":"` + level + `","category":"` + category + `","message":"` + msg + `"}`

	var args FormatterArgs
	parseFormatterArgs([]byte(json), &args)
	if args.Time != timestamp {
		t.Fatalf("Failed to parse timestamp: %s != %s", args.Time, timestamp)
	}
	if args.Level != level {
		t.Fatalf("Failed to parse level: %s != %s", args.Level, level)
	}
	if args.Message != msg {
		t.Fatalf("Failed to parse messae: %s != %s", args.Message, msg)
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
