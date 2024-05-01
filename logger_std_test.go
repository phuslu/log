package log

import (
	"fmt"
	stdLog "log"
	"os"
	"testing"
)

func TestStdWriter(t *testing.T) {
	w := &stdLogWriter{
		Logger: Logger{
			Level:  InfoLevel,
			Writer: IOWriter{os.Stderr},
		},
	}

	fmt.Fprint(w, "hello from stdLog debug Print")
	fmt.Fprintln(w, "hello from stdLog debug Println")
	fmt.Fprintf(w, "hello from stdLog debug %s", "Printf")
}

func TestStdLogger(t *testing.T) {
	logger := Logger{
		Level:   DebugLevel,
		Caller:  -1,
		Context: NewContext(nil).Str("tag", "std_log").Value(),
		Writer:  &ConsoleWriter{ColorOutput: true, EndWithMessage: true},
	}

	stdLog := logger.Std("", stdLog.LstdFlags)
	stdLog.Print("hello from stdLog Print")
	stdLog.Println("hello from stdLog Println")
	stdLog.Printf("hello from stdLog %s", "Printf")

	stdLog = logger.Std("", 0)
	stdLog.Print("hello from stdLog Print")
	stdLog.Println("hello from stdLog Println")
	stdLog.Printf("hello from stdLog %s", "Printf")
}
