package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
)

func TestIsTerminal(t *testing.T) {
	file, _ := os.Open(os.DevNull)

	if IsTerminal(file.Fd()) {
		t.Errorf("test %s terminal mode failed", os.DevNull)
	}

	// The SIGSYS signal would be triggered for errors like "function not implemented".
	// The process would crash for SIGSYS on some platforms (eg. Darwin), so we need to
	// ignore the signal to make sure this test runs correctly on all platforms.
	// signal.Ignore(syscall.SIGSYS)

	// Mute "function not implemented" and "undefined: syscall.SIGSYS" for non linux_amd64 platforms
	if !(runtime.GOOS == "linux" && runtime.GOARCH == "amd64") {
		return
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
		t.Logf("isTerminal(%s, %s) return %+v", c.GOOS, c.GOARCH, isTerminal(file.Fd(), c.GOOS, c.GOARCH))
	}

}

func TestConsoleWriter(t *testing.T) {
	w := &ConsoleWriter{}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"test.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json console writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json console writer error: %+v", err)
		}
	}
}

func TestConsoleWriterColor(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	for _, level := range []string{"trace", "debug", "info", "warning", "error", "fatal", "panic", "hahaha"} {
		_, err := wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"%s","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"message":"hello json console color writer"}`+"\n", level)
		if err != nil {
			t.Errorf("test json color console writer error: %+v", err)
		}
	}
}

func TestConsoleWriterNewline(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	_, err := wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterQuote(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
		QuoteString: false,
	}

	_, err := wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	w.QuoteString = true

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,"foo"],"obj":{"a":["1"], "b":{"1":"2"}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterMessage(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput:    true,
		EndWithMessage: true,
	}

	_, err := wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"msg":"hello json msg color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	w.ColorOutput = false
	w.ColorOutput = false

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,"foo"],"obj":{"a":["1"], "b":{"1":"2"}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, "{\"msg\":\"hello world\\n\"}\n")
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterStack(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	_, err := wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","stack":"stack1\n\tstack2\n\t\tstack3\n","message":"hello console stack writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterFormatter(t *testing.T) {
	w := &ConsoleWriter{
		Formatter: func(w io.Writer, a *FormatterArgs) (int, error) {
			n, _ := fmt.Fprintf(w, "%c%s %s %s] %s", a.Level[0]-32, a.Time, a.Goid, a.Caller, a.Message)
			for _, kv := range a.KeyValues {
				i, _ := fmt.Fprintf(w, " %s=%s", kv.Key, kv.Value)
				n += i
			}
			i, err := fmt.Fprintf(w, "\n")
			return n + i, err
		},
	}

	_, err := wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","goid":123,"error":"i am test error","stack":"stack1\n\tstack2\n\t\tstack3\n","message":"hello console stack writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"debug","caller":"pretty.go:42","foo":"bar","n":42,"a":[1,2,3],"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"error","caller":"pretty.go:42","goid":0,"error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"hahaha","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	w.QuoteString = true

	_, err = wlprintf(w, InfoLevel, `{"time":"2019-07-10T05:35:54.277Z","level":"hahaha","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"msg":"hello json console color writer\n"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, `{"ts":1234567890,"level":"hahaha","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"stack":{"a":[1,2], "b":{"c":3}},"msg":"hello json console color writer\n"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}

	_, err = wlprintf(w, InfoLevel, "a long long message not a json format\n")
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
			Formatter: func(w io.Writer, a *FormatterArgs) (int, error) {
				return fmt.Fprintf(w, "%c%s %s %s] %s\n", a.Level[0]-32, a.Time, a.Goid, a.Caller, a.Message)
			},
		},
	})

	glog.Info().Msgf("hello glog %s", "Info")
	glog.Warn().Msgf("hello glog %s", "Earn")
	glog.Error().Msgf("hello glog %s", "Error")
	glog.Fatal().Msgf("hello glog %s", "Fatal")
}

func TestConsoleWriterTime(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	_, err := wlprintf(w, InfoLevel, `{"ts":1594828508,"level":"info","caller":"pretty.go:42","goid":123,"error":"i am test error","message":"hello console time writer\n"}`)
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}

func TestConsoleWriterInvaild(t *testing.T) {
	w := &ConsoleWriter{
		ColorOutput: true,
	}

	_, err := wlprintf(w, InfoLevel, "a long long long long plain text\n")
	if err != nil {
		t.Errorf("test plain text console writer error: %+v", err)
	}
}
