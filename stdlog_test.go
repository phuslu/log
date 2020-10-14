package log

import (
	"fmt"
	stdLog "log"
	"os"
	"testing"
)

func TestStdWriter(t *testing.T) {
	w := &stdLogWriter{
		logger: Logger{
			Level:  InfoLevel,
			Writer: IOWriter{os.Stderr},
		},
		level: DebugLevel,
	}

	fmt.Fprint(w, "hello from stdLog debug Print")
	fmt.Fprintln(w, "hello from stdLog debug Println")
	fmt.Fprintf(w, "hello from stdLog debug %s", "Printf")

	w.level = InfoLevel
	fmt.Fprint(w, "hello from stdLog info Print")
	fmt.Fprintln(w, "hello from stdLog info Println")
	fmt.Fprintf(w, "hello from stdLog info %s", "Printf")
}

func TestStdLogger(t *testing.T) {
	logger := Logger{
		Level:  DebugLevel,
		Caller: 1,
		Writer: &ConsoleWriter{ColorOutput: true, EndWithMessage: true},
	}

	stdLog := logger.Std(InfoLevel, NewContext(nil).Str("tag", "std_log").Value(), "", stdLog.LstdFlags)
	stdLog.Print("hello from stdLog Print")
	stdLog.Println("hello from stdLog Println")
	stdLog.Printf("hello from stdLog %s", "Printf")

	stdLog = logger.Std(InfoLevel, nil, "", 0)
	stdLog.Print("hello from stdLog Print")
	stdLog.Println("hello from stdLog Println")
	stdLog.Printf("hello from stdLog %s", "Printf")
}
