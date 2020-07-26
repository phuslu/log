package log

import (
	"fmt"
	"os"
	"testing"
	"text/template"
)

func TestIsTerminal(t *testing.T) {
	file, _ := os.Open(os.DevNull)

	if IsTerminal(file.Fd()) {
		t.Errorf("test is terminal mode for %s failed", os.DevNull)
	}

	cases := []struct {
		GOOS   string
		GOARCH string
	}{
		{"plan9", "amd64"},
		{"js", "wasm"},
		{"nacl", "amd64"},
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"linux", "mips"},
		{"linux", "mipsle"},
		{"linux", "mips64"},
		{"linux", "mips64le"},
		{"linux", "ppc64"},
		{"linux", "ppc64le"},
		{"linux", "386"},
		{"darwin", "amd64"},
		{"darwin", "386"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
		{"windows", "386"},
		{"windows", "arm64"},
	}

	for _, c := range cases {
		if isTerminal(file.Fd(), c.GOOS, c.GOARCH) {
			t.Errorf("test is terminal mode for %s failed", os.DevNull)
		}
	}

}

func TestConsoleWriter(t *testing.T) {
	w := &ConsoleWriter{}

	for _, level := range []string{"debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json console writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json console writer error: %+v", err)
		}
	}
}

func TestConsoleWriterColor(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	for _, level := range []string{"debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json console color writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json color console writer error: %+v", err)
		}
	}
}

func TestConsoleWriterNewline(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterQuote(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
		QuoteString: false,
	}

	_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	w.QuoteString = true

	_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,"foo"],"obj":{"a":["1"], "b":{"1":"2"}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterMessage(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput:    true,
		EndWithMessage: true,
	}

	_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"msg":"hello json msg color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	w.ColorOutput = false
	w.ColorOutput = false

	_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,"foo"],"obj":{"a":["1"], "b":{"1":"2"}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterStack(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","stack":"stack1\n\tstack2\n\t\tstack3\n","message":"hello console stack writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterTemplate(t *testing.T) {
	w := &ConsoleWriter{
		Template: template.Must(template.New("").Funcs(ColorFuncMap).Parse(ColorTemplate)),
	}

	_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","goid":123,"error":"i am test error","stack":"stack1\n\tstack2\n\t\tstack3\n","message":"hello console stack writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"debug","caller":"pretty.go:42","foo":"bar","n":42,"a":[1,2,3],"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"error","caller":"pretty.go:42","goid":0,"error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"hahaha","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	w.QuoteString = true

	_, err = fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"hahaha","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"msg":"hello json console color writer\n"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = fmt.Fprintf(w, "a long long message not a json format\n")
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterTemplateColor(t *testing.T) {
	w := &ConsoleWriter{
		Template: template.Must(template.New("").Funcs(ColorFuncMap).Parse(`
			black {{black .Message}}
			red {{red .Message}}
			green {{green .Message}}
			yellow {{yellow .Message}}
			blue {{blue .Message}}
			magenta {{magenta .Message}}
			cyan {{cyan .Message}}
			white {{white .Message}}
			gray {{gray .Message}}
		`)),
	}

	_, err := fmt.Fprintf(w, `{"time":"2019-07-10T05:35:54.277Z","level":"info","message":"hello console stack writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterGlog(t *testing.T) {
	notTest = false

	glog := (&Logger{
		Level:      InfoLevel,
		Caller:     1,
		TimeFormat: "0102 15:04:05.999999",
		Writer: &ConsoleWriter{
			Template: template.Must(template.New("").Parse(
				`{{.Level.One}}{{.Time}} {{.Goid}} {{.Caller}}] {{.Message}}`)),
		},
	}).Sugar(nil)

	glog.Infof("hello glog %s", "Info")
	glog.Warnf("hello glog %s", "Earn")
	glog.Errorf("hello glog %s", "Error")
	glog.Fatalf("hello glog %s", "Fatal")
}

func TestConsoleWriterTime(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
		TimeField:   "ts",
	}

	_, err := fmt.Fprintf(w, `{"ts":1594828508,"level":"info","caller":"pretty.go:42","error":"i am test error","message":"hello console time writer\n"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterInvaild(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	_, err := fmt.Fprintf(w, "a long long long long plain text\n")
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}
