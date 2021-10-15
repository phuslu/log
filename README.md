# phuslog - Structured Logging Made Easy

[![godoc][godoc-img]][godoc]
[![goreport][report-img]][report]
[![build][build-img]][build]
[![coverage][cov-img]][cov]
![stability-stable][stability-img]

## Features

* Dependency Free
* Simple and Clean Interface
* Consistent Writer
    - `IOWriter`, *io.Writer wrapper*
    - `ConsoleWriter`, *colorful & formatting*
    - `FileWriter`, *rotating & effective*
    - `MultiLevelWriter`, *multiple level dispatch*
    - `SyslogWriter`, *syslog server logging*
    - `JournalWriter`, *linux systemd logging*
    - `EventlogWriter`, *windows system event*
    - `AsyncWriter`, *asynchronously writing*
* Third-party Logger Interceptor
    - `Logger.Std`, *(std)log*
    - `Logger.Grpc`, *grpclog.LoggerV2*
    - `Logger.Logr`, *logr.Logger*
* Useful utility function
    - `Goid()`, *the goroutine id matches stack trace*
    - `NewXID()`, *create a tracing id*
    - `Fastrandn(n uint32)`, *fast pseudorandom uint32 in [0,n)*
    - `IsTerminal(fd uintptr)`, *isatty for golang*
    - `Printf(fmt string, a ...interface{})`, *printf logging*
* High Performance
    - [Significantly faster][high-performance] than all other json loggers.

## Interfaces

### Logger
```go
// DefaultLogger is the global logger.
var DefaultLogger = Logger{
	Level:      DebugLevel,
	Caller:     0,
	TimeField:  "",
	TimeFormat: "",
	Writer:     &IOWriter{os.Stderr},
}

// A Logger represents an active logging object that generates lines of JSON output to an io.Writer.
type Logger struct {
	// Level defines log levels.
	Level Level

	// Caller determines if adds the file:line of the "caller" key.
	Caller int

	// TimeField defines the time filed name in output.  It uses "time" in if empty.
	TimeField string

	// TimeFormat specifies the time format in output. It uses RFC3339 with millisecond if empty.
	// If set with `TimeFormatUnix`, `TimeFormatUnixMs`, times are formated as UNIX timestamp.
	TimeFormat string

	// Writer specifies the writer of output. It uses a wrapped os.Stderr Writer in if empty.
	Writer Writer
}
```

### ConsoleWriter & FileWriter
```go
// ConsoleWriter parses the JSON input and writes it in a colorized, human-friendly format to Writer.
// IMPORTANT: Don't use ConsoleWriter on critical path of a high concurrency and low latency application.
//
// Default output format:
//     {Time} {Level} {Goid} {Caller} > {Message} {Key}={Value} {Key}={Value}
type ConsoleWriter struct {
	// ColorOutput determines if used colorized output.
	ColorOutput bool

	// QuoteString determines if quoting string values.
	QuoteString bool

	// EndWithMessage determines if output message in the end of line.
	EndWithMessage bool

	// Writer is the output destination. using os.Stderr if empty.
	Writer io.Writer

	// Formatter specifies an optional text formatter for creating a customized output,
	// If it is set, ColorOutput, QuoteString and EndWithMessage will be ignored.
	Formatter func(w io.Writer, args *FormatterArgs) (n int, err error)
}


// FileWriter is an Writer that writes to the specified filename.
type FileWriter struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.
	Filename string

	// FileMode represents the file's mode and permission bits.  The default
	// mode is 0644
	FileMode os.FileMode

	// MaxSize is the maximum size in bytes of the log file before it gets rotated.
	MaxSize int64

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files
	MaxBackups int

	// TimeFormat specifies the time format of filename, uses `2006-01-02T15-04-05` as default format.
	// If set with `TimeFormatUnix`, `TimeFormatUnixMs`, times are formated as UNIX timestamp.
	TimeFormat string

	// LocalTime determines if the time used for formatting the timestamps in
	// log files is the computer's local time.  The default is to use UTC time.
	LocalTime bool

	// HostName determines if the hostname used for formatting in log files.
	HostName bool

	// ProcessID determines if the pid used for formatting in log files.
	ProcessID bool

	// EnsureFolder ensures the file directory creation before writing.
	EnsureFolder bool

	// Header specifies an optional header function of log file after rotation,
	Header func(fileinfo os.FileInfo) []byte

	// Cleaner specifies an optional cleanup function of log backups after rotation,
	// if not set, the default behavior is to delete more than MaxBackups log files.
	Cleaner func(filename string, maxBackups int, matches []os.FileInfo)
}
```
> Note: FileWriter implements log.Writer and io.Writer interfaces both, it is a drop-in replacement of [lumberjack][lumberjack].
> FileWriter also creates a symlink to the current logging file, it requires administrator privileges on Windows.

## Getting Started

### Simple Logging Example

An out of box example. [![playground][play-simple-img]][play-simple]
```go
package main

import (
	"github.com/phuslu/log"
)

func main() {
	log.Info().Str("foo", "bar").Int("number", 42).Msg("hi, phuslog")
	log.Info().Msgf("foo=%s number=%d error=%+v", "bar", 42, "an error")
}

// Output:
//   {"time":"2020-03-22T09:58:41.828Z","level":"info","foo":"bar","number":42,"message":"hi, phuslog"}
//   {"time":"2020-03-22T09:58:41.828Z","level":"info","message":"foo=bar number=42 error=an error"}
```
> Note: By default log writes to `os.Stderr`

### Customize the configuration and formatting:

To customize logger filed name and format. [![playground][play-customize-img]][play-customize]
```go
package main

import (
	"github.com/phuslu/log"
)

func main() {
	log.DefaultLogger = log.Logger{
		Level:      log.InfoLevel,
		Caller:     1,
		TimeField:  "date",
		TimeFormat: "2006-01-02",
		Writer:     &log.IOWriter{os.Stdout},
	}

	log.Info().Str("foo", "bar").Msgf("hello %s", "world")
}

// Output: {"date":"2019-07-04","level":"info","caller":"prog.go:16","foo":"bar","message":"hello world"}
```

### Pretty Console Writer

To log a human-friendly, colorized output, use `ConsoleWriter`. [![playground][play-pretty-img]][play-pretty]

```go
if log.IsTerminal(os.Stderr.Fd()) {
	log.DefaultLogger = log.Logger{
		TimeFormat: "15:04:05",
		Caller:     1,
		Writer: &log.ConsoleWriter{
			ColorOutput:    true,
			QuoteString:    true,
			EndWithMessage: true,
		},
	}
}

log.Debug().Int("everything", 42).Str("foo", "bar").Msg("hello world")
log.Info().Int("everything", 42).Str("foo", "bar").Msg("hello world")
log.Warn().Int("everything", 42).Str("foo", "bar").Msg("hello world")
log.Error().Err(errors.New("an error")).Msg("hello world")
```
![Pretty logging][pretty-img]
> Note: pretty logging also works on windows console

### Formatting Console Writer

To log with user-defined format(e.g. glog), using `ConsoleWriter.Formatter`. [![playground][play-formatting-img]][play-formatting]

```go
log.DefaultLogger = log.Logger{
	Level:      log.InfoLevel,
	Caller:     1,
	TimeFormat: "0102 15:04:05.999999",
	Writer: &log.ConsoleWriter{
		Formatter: func (w io.Writer, a *log.FormatterArgs) (int, error) {
			return fmt.Fprintf(w, "%c%s %s %s] %s\n%s", strings.ToUpper(a.Level)[0],
				a.Time, a.Goid, a.Caller, a.Message, a.Stack)
		},
	},
}

log.Info().Msgf("hello glog %s", "Info")
log.Warn().Msgf("hello glog %s", "Warn")
log.Error().Msgf("hello glog %s", "Error")

// Output:
// I0725 09:59:57.503246 19 console_test.go:183] hello glog Info
// W0725 09:59:57.504247 19 console_test.go:184] hello glog Warn
// E0725 09:59:57.504247 19 console_test.go:185] hello glog Error
```

### Rotating File Writer

To log to a daily-rotating file, use `FileWriter`. [![playground][play-file-img]][play-file]
```go
package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/phuslu/log"
	"github.com/robfig/cron/v3"
)

func main() {
	logger := log.Logger{
		Level: log.ParseLevel("info"),
		Writer: &log.FileWriter{
			Filename:     "logs/main.log",
			FileMode:     0600,
			MaxSize:      100 * 1024 * 1024,
			MaxBackups:   7,
			EnsureFolder: true,
			LocalTime:    true,
		},
	}

	runner := cron.New(cron.WithLocation(time.Local))
	runner.AddFunc("0 0 * * *", func() { logger.Writer.(*log.FileWriter).Rotate() })
	go runner.Run()

	for {
		time.Sleep(time.Second)
		logger.Info().Msg("hello world")
	}
}
```

### Rotating File Writer within a total size

To rotating log file hourly and keep in a total size, use `FileWriter.Cleaner`.
```go
package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/phuslu/log"
	"github.com/robfig/cron/v3"
)

func main() {
	logger := log.Logger{
		Level: log.ParseLevel("info"),
		Writer: &log.FileWriter{
			Filename: "main.log",
			MaxSize:  500 * 1024 * 1024,
			Cleaner:  func(filename string, maxBackups int, matches []os.FileInfo) {
				var dir = filepath.Dir(filename)
				var total int64
				for i := len(matches) - 1; i >= 0; i-- {
					total += matches[i].Size()
					if total > 5*1024*1024*1024 {
						os.Remove(filepath.Join(dir, matches[i].Name()))
					}
				}
			},
		},
	}

	runner := cron.New(cron.WithLocation(time.UTC))
	runner.AddFunc("0 * * * *", func() { logger.Writer.(*log.FileWriter).Rotate() })
	go runner.Run()

	for {
		time.Sleep(time.Second)
		logger.Info().Msg("hello world")
	}
}
```

### Random Sample Logger:

To logging only 5% logs, use below idiom.
```go
if log.Fastrandn(100) < 5 {
	log.Log().Msg("hello world")
}
```

### Multiple Dispatching Writer

To log to different writers by different levels, use `MultiLevelWriter`.

```go
log.DefaultLogger.Writer = &log.MultiLevelWriter{
	InfoWriter:    &log.FileWriter{Filename: "main.INFO", MaxSize: 100<<20},
	WarnWriter:    &log.FileWriter{Filename: "main.WARNING", MaxSize: 100<<20},
	ErrorWriter:   &log.FileWriter{Filename: "main.ERROR", MaxSize: 100<<20},
	ConsoleWriter: &log.ConsoleWriter{ColorOutput: true},
	ConsoleLevel:  log.ErrorLevel,
}

log.Info().Int("number", 42).Str("foo", "bar").Msg("a info log")
log.Warn().Int("number", 42).Str("foo", "bar").Msg("a warn log")
log.Error().Int("number", 42).Str("foo", "bar").Msg("a error log")
```

### Multiple Entry Writer
To log to different writers, use `MultiEntryWriter`.

```go
log.DefaultLogger.Writer = &log.MultiEntryWriter{
	&log.ConsoleWriter{ColorOutput: true},
	&log.FileWriter{Filename: "main.log", MaxSize: 100<<20},
	&log.EventlogWriter{Source: ".NET Runtime", ID: 1000},
}

log.Info().Int("number", 42).Str("foo", "bar").Msg("a info log")
```

### Multiple IO Writer

To log to multiple io writers like `io.MultiWriter`, use below idiom. [![playground][play-multiio-img]][play-multiio]

```go
log.DefaultLogger.Writer = &log.MultiIOWriter{
	os.Stdout,
	&log.FileWriter{Filename: "main.log", MaxSize: 100<<20},
}

log.Info().Int("number", 42).Str("foo", "bar").Msg("a info log")
```

### Multiple Combined Logger:

To logging to different logger as you want, use below idiom. [![playground][play-combined-img]][play-combined]
```go
package main

import (
	"github.com/phuslu/log"
)

var logger = struct {
	Console log.Logger
	Access  log.Logger
	Data    log.Logger
}{
	Console: log.Logger{
		TimeFormat: "15:04:05",
		Caller:     1,
		Writer: &log.ConsoleWriter{
			ColorOutput:    true,
			EndWithMessage: true,
		},
	},
	Access: log.Logger{
		Level: log.InfoLevel,
		Writer: &log.FileWriter{
			Filename:   "access.log",
			MaxSize:    50 * 1024 * 1024,
			MaxBackups: 7,
			LocalTime:  false,
		},
	},
	Data: log.Logger{
		Level: log.InfoLevel,
		Writer: &log.FileWriter{
			Filename:   "data.log",
			MaxSize:    50 * 1024 * 1024,
			MaxBackups: 7,
			LocalTime:  false,
		},
	},
}

func main() {
	logger.Console.Info().Msgf("hello world")
	logger.Access.Log().Msgf("handle request")
	logger.Data.Log().Msgf("some data")
}
```

### SyslogWriter & JournalWriter & EventlogWriter

To log to a syslog server, using `SyslogWriter`.

```go
log.DefaultLogger.Writer = &log.SyslogWriter{
	// Network : "unixgram",
	// Address : "/run/systemd/journal/syslog",
	Network : "tcp",
	Address : "192.168.0.2:601",
	Tag     : "",
	Marker  : "@cee:",
	Dial    : net.Dial,
}

log.Info().Msg("hi")

// Output: <6>Oct 5 16:25:38 [237]: @cee:{"time":"2020-10-05T16:25:38.026Z","level":"info","message":"hi"}
```

To log to linux systemd journald, using `JournalWriter`.

```go
log.DefaultLogger.Writer = &log.JournalWriter{}

log.Info().Int("number", 42).Str("foo", "bar").Msg("hello world")
```

To log to windows system event, using `EventlogWriter`.

```go
log.DefaultLogger.Writer = &log.EventlogWriter{
	Source: ".NET Runtime",
	ID:     1000,
}

log.Info().Int("number", 42).Str("foo", "bar").Msg("hello world")
```

### AsyncWriter

To logging asynchronously for performance stability, use `AsyncWriter`.

```go
logger := log.Logger{
	Level:  log.InfoLevel,
	Writer: &log.AsyncWriter{
		ChannelSize: 100,
		Writer:      &log.FileWriter{
			Filename:   "main.log",
			FileMode:   0600,
			MaxSize:    50*1024*1024,
			MaxBackups: 7,
			LocalTime:  false,
		},
	},
}

logger.Info().Int("number", 42).Str("foo", "bar").Msg("a async info log")
logger.Warn().Int("number", 42).Str("foo", "bar").Msg("a async warn log")
logger.Writer.(io.Closer).Close()
```

> Note: To flush data and quit safely, call `AsyncWriter.Close()` explicitly.

### StdLog & Logr & Grpc Interceptor

Using wrapped loggers for stdlog/grpc/logr. [![playground][play-interceptor-img]][play-interceptor]

```go
package main

import (
	stdLog "log"
	"github.com/go-logr/logr"
	"github.com/phuslu/log"
	"google.golang.org/grpc/grpclog"
)

func main() {
	ctx := log.NewContext(nil).Str("tag", "hi log").Value()

	var stdlog *stdLog.Logger = log.DefaultLogger.Std(log.InfoLevel, ctx, "prefix ", stdLog.LstdFlags)
	stdlog.Print("hello from stdlog Print")
	stdlog.Println("hello from stdlog Println")
	stdlog.Printf("hello from stdlog %s", "Printf")

	var grpclog grpclog.LoggerV2 = log.DefaultLogger.Grpc(ctx)
	grpclog.Infof("hello %s", "grpclog Infof message")
	grpclog.Errorf("hello %s", "grpclog Errorf message")

	var logrLog logr.Logger = log.DefaultLogger.Logr(ctx)
	logrLog = logrLog.WithName("a_named_logger").WithValues("a_key", "a_value")
	logrLog.Info("hello", "foo", "bar", "number", 42)
	logrLog.Error(errors.New("this is a error"), "hello", "foo", "bar", "number", 42)
}
```

### GrcpGateway Interceptor

Using wrapped loggers for grcp-gateway in go-grpc-middleware.

```go
package main

import (
	"context"
	"testing"

	grpcphuslog "github.com/grpc-ecosystem/go-grpc-middleware/providers/phuslog/v2"
	"github.com/phuslu/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

func Example_initialization() {
	// Logger is used, allowing pre-definition of certain fields by the user.
	logger := log.DefaultLogger.GrcpGateway()
	// You can also add grpc LoggerV2 logger wrapper
	grpclog.SetLoggerV2(log.DefaultLogger.Grpc(nil))
	// Create a server, make sure we put the tags context before everything else.
	_ = grpc.NewServer(
		middleware.WithUnaryServerChain(
			tags.UnaryServerInterceptor(),
			logging.UnaryServerInterceptor(grpcphuslog.InterceptorLogger(logger)),
		),
		middleware.WithStreamServerChain(
			tags.StreamServerInterceptor(),
			logging.StreamServerInterceptor(grpcphuslog.InterceptorLogger(logger)),
		),
	)
}
```

### User-defined Data Structure

To log with user-defined struct effectively, implements `MarshalObject`. [![playground][play-marshal-img]][play-marshal]

```go
package main

import (
	"github.com/phuslu/log"
)

type User struct {
	ID   int
	Name string
	Pass string
}

func (u *User) MarshalObject(e *log.Entry) {
	e.Int("id", u.ID).Str("name", u.Name).Str("password", "***")
}

func main() {
	log.Info().Object("user", &User{1, "neo", "123456"}).Msg("")
	log.Info().EmbedObject(&User{2, "john", "abc"}).Msg("")
}

// Output:
//   {"time":"2020-07-12T05:03:43.949Z","level":"info","user":{"id":1,"name":"neo","password":"***"}}
//   {"time":"2020-07-12T05:03:43.949Z","level":"info","id":2,"name":"john","password":"***"}
```

### Contextual Fields

To add preserved `key:value` pairs to each entry, use `NewContext`. [![playground][play-context-img]][play-context]

```go
logger := log.Logger{
	Level:   log.InfoLevel,
	Context: log.NewContext(nil).Str("ctx", "some_ctx").Value(),
}

logger.Debug().Int("no0", 0).Msg("zero")
logger.Info().Int("no1", 1).Msg("first")
logger.Info().Int("no2", 2).Msg("second")

// Output:
//   {"time":"2020-07-12T05:03:43.949Z","level":"info","ctx":"some_ctx","no1":1,"message":"first"}
//   {"time":"2020-07-12T05:03:43.949Z","level":"info","ctx":"some_ctx","no2":2,"message":"second"}
```

You can make a copy of log and add contextual fields. [![playground][play-context-add-img]][play-context-add]

```go
package main

import (
	"github.com/phuslu/log"
)

func main() {
	sublogger := log.DefaultLogger
	sublogger.Level = log.InfoLevel
	sublogger.Context = log.NewContext(nil).Str("ctx", "some_ctx").Value()

	sublogger.Debug().Int("no0", 0).Msg("zero")
	sublogger.Info().Int("no1", 1).Msg("first")
	sublogger.Info().Int("no2", 2).Msg("second")
	log.Debug().Int("no3", 3).Msg("no context")
}

// Output:
//   {"time":"2021-06-14T06:36:42.904+02:00","level":"info","ctx":"some_ctx","no1":1,"message":"first"}
//   {"time":"2021-06-14T06:36:42.905+02:00","level":"info","ctx":"some_ctx","no2":2,"message":"second"}
//   {"time":"2021-06-14T06:36:42.906+02:00","level":"debug","no3":3,"message":"no context"}
```

### High Performance

The most common benchmarks(disable/normal/interface/printf/caller) with zap/zerolog, which runs on [github actions][benchmark]:

```go
// go test -v -cpu=4 -run=none -bench=. -benchtime=10s -benchmem log_test.go
package main

import (
	"io/ioutil"
	"testing"

	"github.com/phuslu/log"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var msg = "The quick brown fox jumps over the lazy dog"
var obj = struct {Rate string; Low int; High float32}{"15", 16, 123.2}

func BenchmarkDisableZap(b *testing.B) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ioutil.Discard),
		zapcore.InfoLevel,
	))
	for i := 0; i < b.N; i++ {
		logger.Debug(msg, zap.String("rate", "15"), zap.Int("low", 16), zap.Float32("high", 123.2))
	}
}

func BenchmarkNormalZap(b *testing.B) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ioutil.Discard),
		zapcore.InfoLevel,
	))
	for i := 0; i < b.N; i++ {
		logger.Info(msg, zap.String("rate", "15"), zap.Int("low", 16), zap.Float32("high", 123.2))
	}
}

func BenchmarkInterfaceZap(b *testing.B) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ioutil.Discard),
		zapcore.InfoLevel,
	)).Sugar()
	for i := 0; i < b.N; i++ {
		logger.Infow(msg, "object", &obj)
	}
}

func BenchmarkPrintfZap(b *testing.B) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ioutil.Discard),
		zapcore.InfoLevel,
	)).Sugar()
	for i := 0; i < b.N; i++ {
		logger.Infof("rate=%s low=%d high=%f msg=%s", "15", 16, 123.2, msg)
	}
}

func BenchmarkCallerZap(b *testing.B) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ioutil.Discard),
		zapcore.InfoLevel),
		zap.AddCaller(),
	)
	for i := 0; i < b.N; i++ {
		logger.Info(msg, zap.String("rate", "15"), zap.Int("low", 16), zap.Float32("high", 123.2))
	}
}

func BenchmarkDisableZeroLog(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(ioutil.Discard).With().Timestamp().Logger()
	for i := 0; i < b.N; i++ {
		logger.Debug().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkNormalZeroLog(b *testing.B) {
	logger := zerolog.New(ioutil.Discard).With().Timestamp().Logger()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkInterfaceZeroLog(b *testing.B) {
	logger := zerolog.New(ioutil.Discard).With().Timestamp().Logger()
	for i := 0; i < b.N; i++ {
		logger.Info().Interface("object", &obj).Msg(msg)
	}
}

func BenchmarkPrintfZeroLog(b *testing.B) {
	logger := zerolog.New(ioutil.Discard).With().Timestamp().Logger()
	for i := 0; i < b.N; i++ {
		logger.Info().Msgf("rate=%s low=%d high=%f msg=%s", "15", 16, 123.2, msg)
	}
}

func BenchmarkCallerZeroLog(b *testing.B) {
	logger := zerolog.New(ioutil.Discard).With().Caller().Timestamp().Logger()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkDisablePhusLog(b *testing.B) {
	logger := log.Logger{Level: log.InfoLevel, Writer: log.IOWriter{ioutil.Discard}}
	for i := 0; i < b.N; i++ {
		logger.Debug().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkNormalPhusLog(b *testing.B) {
	logger := log.Logger{Writer: log.IOWriter{ioutil.Discard}}
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkInterfacePhusLog(b *testing.B) {
	logger := log.Logger{Writer: log.IOWriter{ioutil.Discard}}
	for i := 0; i < b.N; i++ {
		logger.Info().Interface("object", &obj).Msg(msg)
	}
}

func BenchmarkPrintfPhusLog(b *testing.B) {
	logger := log.Logger{Writer: log.IOWriter{ioutil.Discard}}
	for i := 0; i < b.N; i++ {
		logger.Info().Msgf("rate=%s low=%d high=%f msg=%s", "15", 16, 123.2, msg)
	}
}

func BenchmarkCallerPhusLog(b *testing.B) {
	logger := log.Logger{Caller: 1, Writer: log.IOWriter{ioutil.Discard}}
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}
```
A Performance result as below, for daily benchmark results see [github actions][benchmark]
```
BenchmarkDisableZap-4         	100000000	       111 ns/op	     192 B/op	       1 allocs/op
BenchmarkNormalZap-4          	 9192708	      1258 ns/op	     192 B/op	       1 allocs/op
BenchmarkInterfaceZap-4       	 7009218	      1746 ns/op	     208 B/op	       2 allocs/op
BenchmarkPrintfZap-4          	 7389448	      1648 ns/op	      96 B/op	       2 allocs/op
BenchmarkCallerZap-4          	 4112085	      2948 ns/op	     440 B/op	       4 allocs/op

BenchmarkDisableZeroLog-4     	940196673	        12.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkNormalZeroLog-4      	16421583	       717 ns/op	       0 B/op	       0 allocs/op
BenchmarkInterfaceZeroLog-4   	11076124	      1075 ns/op	      48 B/op	       1 allocs/op
BenchmarkPrintfZeroLog-4      	 9741673	      1233 ns/op	      96 B/op	       2 allocs/op
BenchmarkCallerZeroLog-4      	 3882822	      3152 ns/op	     272 B/op	       4 allocs/op

BenchmarkDisablePhusLog-4     	1000000000	        11.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkNormalPhusLog-4      	27455464	       426 ns/op	       0 B/op	       0 allocs/op
BenchmarkInterfacePhusLog-4   	14598555	       756 ns/op	       0 B/op	       0 allocs/op
BenchmarkPrintfPhusLog-4      	15283591	       796 ns/op	      16 B/op	       1 allocs/op
BenchmarkCallerPhusLog-4      	 8991439	      1376 ns/op	     216 B/op	       2 allocs/op
```
This library uses the following special techniques to achieve better performance,
1. handwriting time formatting
1. manual inlining
1. unrolled functions

## A Real World Example

The example starts a geoip http server which supports change log level dynamically
```go
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/phuslu/iploc"
	"github.com/phuslu/log"
)

type Config struct {
	Listen struct {
		Tcp string
	}
	Log struct {
		Level   string
		Maxsize int64
		Backups int
	}
}

type Handler struct {
	Config       *Config
	AccessLogger log.Logger
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	reqID := log.NewXID()
	remoteIP, _, _ := net.SplitHostPort(req.RemoteAddr)
	geo := iploc.Country(net.ParseIP(remoteIP))

	h.AccessLogger.Log().
		Xid("req_id", reqID).
		Str("host", req.Host).
		Bytes("geo", geo).
		Str("remote_ip", remoteIP).
		Str("request_uri", req.RequestURI).
		Str("user_agent", req.UserAgent()).
		Str("referer", req.Referer()).
		Msg("access log")

	switch req.RequestURI {
	case "/debug", "/info", "/warn", "/error":
		log.DefaultLogger.SetLevel(log.ParseLevel(req.RequestURI[1:]))
	default:
		fmt.Fprintf(rw, `{"req_id":"%s","ip":"%s","geo":"%s"}`, reqID, remoteIP, geo)
	}
}

func main() {
	config := new(Config)
	err := json.Unmarshal([]byte(`{
		"listen": {
			"tcp": ":8080"
		},
		"log": {
			"level": "debug",
			"maxsize": 1073741824,
			"backups": 5
		}
	}`), config)
	if err != nil {
		log.Fatal().Msgf("json.Unmarshal error: %+v", err)
	}

	handler := &Handler{
		Config: config,
		AccessLogger: log.Logger{
			Writer: &log.FileWriter{
				Filename:   "access.log",
				MaxSize:    config.Log.Maxsize,
				MaxBackups: config.Log.Backups,
				LocalTime:  true,
			},
		},
	}

	if log.IsTerminal(os.Stderr.Fd()) {
		log.DefaultLogger = log.Logger{
			Level:      log.ParseLevel(config.Log.Level),
			Caller:     1,
			TimeFormat: "15:04:05",
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				EndWithMessage: true,
			},
		}
		handler.AccessLogger = log.DefaultLogger
	} else {
		log.DefaultLogger = log.Logger{
			Level: log.ParseLevel(config.Log.Level),
			Writer: &log.FileWriter{
				Filename:   "main.log",
				MaxSize:    config.Log.Maxsize,
				MaxBackups: config.Log.Backups,
				LocalTime:  true,
			},
		}
	}

	server := &http.Server{
		Addr:     config.Listen.Tcp,
		ErrorLog: log.DefaultLogger.Std(log.ErrorLevel, nil, "", 0),
		Handler:  handler,
	}

	log.Fatal().Err(server.ListenAndServe()).Msg("listen failed")
}
```

## Acknowledgment
This log is heavily inspired by [zerolog][zerolog], [glog][glog], [gjson][gjson] and [lumberjack][lumberjack].

[godoc-img]: http://img.shields.io/badge/godoc-reference-5272B4.svg
[godoc]: https://godocs.io/github.com/phuslu/log
[report-img]: https://goreportcard.com/badge/github.com/phuslu/log
[report]: https://goreportcard.com/report/github.com/phuslu/log
[build-img]: https://github.com/phuslu/log/workflows/build/badge.svg
[build]: https://github.com/phuslu/log/actions
[cov-img]: http://gocover.io/_badge/github.com/phuslu/log
[cov]: https://gocover.io/github.com/phuslu/log
[stability-img]: https://img.shields.io/badge/stability-stable-green.svg
[high-performance]: https://github.com/phuslu/log#high-performance
[play-simple-img]: https://img.shields.io/badge/playground-NGV25aBKmYH-29BEB0?style=flat&logo=go
[play-simple]: https://play.golang.org/p/NGV25aBKmYH
[play-customize-img]: https://img.shields.io/badge/playground-emTsJJKUGXZ-29BEB0?style=flat&logo=go
[play-customize]: https://play.golang.org/p/emTsJJKUGXZ
[play-multiio-img]: https://img.shields.io/badge/playground-bwo03SW0B3l-29BEB0?style=flat&logo=go
[play-multiio]: https://play.golang.org/p/bwo03SW0B3l
[play-combined-img]: https://img.shields.io/badge/playground-24d4eDIpDeR-29BEB0?style=flat&logo=go
[play-combined]: https://play.golang.org/p/24d4eDIpDeR
[play-file-img]: https://img.shields.io/badge/playground-nS--ILxFyhHM-29BEB0?style=flat&logo=go
[play-file]: https://play.golang.org/p/nS-ILxFyhHM
[play-pretty-img]: https://img.shields.io/badge/playground-SCcXG33esvI-29BEB0?style=flat&logo=go
[play-pretty]: https://play.golang.org/p/SCcXG33esvI
[pretty-img]: https://user-images.githubusercontent.com/195836/101993218-cda82380-3cf3-11eb-9aa2-b8b1c832a72e.png
[play-formatting-img]: https://img.shields.io/badge/playground-UmJmLxYXwRO-29BEB0?style=flat&logo=go
[play-formatting]: https://play.golang.org/p/UmJmLxYXwRO
[play-context-img]: https://img.shields.io/badge/playground-oAVAo302faf-29BEB0?style=flat&logo=go
[play-context]: https://play.golang.org/p/oAVAo302faf
[play-context-add-img]: https://img.shields.io/badge/playground-LuCghJxMPHI-29BEB0?style=flat&logo=go
[play-context-add]: https://play.golang.org/p/LuCghJxMPHI
[play-marshal-img]: https://img.shields.io/badge/playground-SoQdwQOaQR2-29BEB0?style=flat&logo=go
[play-marshal]: https://play.golang.org/p/SoQdwQOaQR2
[play-interceptor]: https://play.golang.org/p/upmVP5cO62Y
[play-interceptor-img]: https://img.shields.io/badge/playground-upmVP5cO62Y-29BEB0?style=flat&logo=go
[benchmark]: https://github.com/phuslu/log/actions?query=workflow%3Abenchmark
[zerolog]: https://github.com/rs/zerolog
[glog]: https://github.com/golang/glog
[gjson]: https://github.com/tidwall/gjson
[lumberjack]: https://github.com/natefinch/lumberjack
