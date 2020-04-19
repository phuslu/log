# Structured Logging for Humans

[![go.dev][pkg-img]][pkg] [![goreport][report-img]][report] [![coverage][cov-img]][cov]

## Features

* No Dependencies
* Intuitive Interfaces
* JSON/TSV/Printf Loggers
* Rotating File Writer
* Pretty Console Writer
* Dynamic Log Level
* High Performance

## Interfaces

### Logger
```go
// DefaultLogger is the global logger.
var DefaultLogger = Logger{
	Level:      DebugLevel,
	Caller:     0,
	TimeField:  "",
	TimeFormat: "",
	Timestamp:  false,
	HostField:  "",
	Writer:     os.Stderr,
}

// A Logger represents an active logging object that generates lines of JSON output to an io.Writer.
type Logger struct {
	// Level defines log levels.
	Level Level

	// Caller determines if adds the file:line of the "caller" key.
	Caller int

	// TimeField defines the time filed name in output.  It uses "time" in if empty.
	TimeField string

	// TimeFormat specifies the time format in output. It uses time.RFC3389 in if empty.
	TimeFormat string

	// Timestamp determines if time is formatted as an UNIX timestamp as integer.
	// If set, the value of TimeField and TimeFormat will be ignored.
	Timestamp bool

	// HostField specifies the key for hostname in output if not empty
	HostField string

	// Writer specifies the writer of output. It uses os.Stderr in if empty.
	Writer io.Writer
}
```

### FileWriter & ConsoleWriter
```go
// FileWriter is an io.WriteCloser that writes to the specified filename.
type FileWriter struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.
	Filename string

	// FileMode represents the file's mode and permission bits.  The default
	// mode is 0644
	FileMode os.FileMode

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated.
	MaxSize int64

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files
	MaxBackups int

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool

	// HostName determines if the hostname used for formatting in backup files.
	HostName bool
}

// ConsoleWriter parses the JSON input and writes it in an
// (optionally) colorized, human-friendly format to os.Stderr
type ConsoleWriter struct {
	// ANSIColor determines if used colorized output.
	ANSIColor bool
}
```

## Getting Started

### Simple Logging Example

A out of box example. [![playground][play-simple-img]][play-simple]
```go
package main

import (
	"github.com/phuslu/log"
)

func main() {
	log.Printf("Hello, %s", "世界")
	log.Info().Str("foo", "bar").Int("number", 42).Msg("a structured logger")
}

// Output:
//   {"time":"2020-03-22T09:58:41.828Z","level":"debug","message":"Hello, 世界"}
//   {"time":"2020-03-22T09:58:41.828Z","level":"info","foo":"bar","number":42,"message":"a structure logger"}
```
> Note: By default log writes to `os.Stderr`

### Customize the configuration and formatting:

To customize logger filed name and format. [![playground][play-customize-img]][play-customize]
```go
log.DefaultLogger = log.Logger{
	Level:      log.InfoLevel,
	Caller:     1,
	TimeField:  "date",
	TimeFormat: "2006-01-02",
	HostField:  "host",
	Writer:     os.Stderr,
}
log.Info().Str("foo", "bar").Msgf("hello %s", "world")

// Output: {"date":"2019-07-04","level":"info","host":"hk","caller":"test.go:42","foo":"bar","message":"hello world"}
```

### Rotating File Writer

```go
package main

import (
	"time"

	"github.com/phuslu/log"
	"github.com/robfig/cron/v3"
)

func main() {
	logger := log.Logger{
		Level:      log.ParseLevel("info"),
		Writer:     &log.FileWriter{
			Filename:   "main.log",
			FileMode:   0600,
			MaxSize:    50*1024*1024,
			MaxBackups: 7,
			LocalTime:  false,
		},
	}

	runner := cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC))
	runner.AddFunc("0 0 * * * *", func() { logger.Writer.(*log.FileWriter).Rotate() })
	go runner.Run()

	for {
		time.Sleep(time.Second)
		logger.Info().Msg("hello world")
	}
}
```

### Pretty Console Writer

To log a human-friendly, colorized output, use `log.ConsoleWriter`. [![playground][play-pretty-img]][play-pretty]

```go
if log.IsTerminal(os.Stderr.Fd()) {
	log.DefaultLogger = log.Logger{
		Caller: 1,
		Writer: &log.ConsoleWriter{ANSIColor: true},
	}
}

log.Printf("a printf style line")
log.Info().Err(errors.New("an error")).Int("everything", 42).Str("foo", "bar").Msg("hello world")
```
![Pretty logging][pretty-logging-img]
> Note: pretty logging also works on windows console

### Dynamic log Level

To change log level on the fly, use `log.DefaultLogger.SetLevel`. [![playground][play-dynamic-img]][play-dynamic]

```go
log.DefaultLogger.SetLevel(log.InfoLevel)
log.Debug().Msg("debug log")
log.Info().Msg("info log")
log.Warn().Msg("warn log")
log.DefaultLogger.SetLevel(log.DebugLevel)
log.Debug().Msg("debug log")
log.Info().Msg("info log")

// Output:
//   {"time":"2020-03-24T05:06:54.674Z","level":"info","message":"info log"}
//   {"time":"2020-03-24T05:06:54.674Z","level":"warn","message":"warn log"}
//   {"time":"2020-03-24T05:06:54.675Z","level":"debug","message":"debug log"}
//   {"time":"2020-03-24T05:06:54.675Z","level":"info","message":"info log"}
```

### Logging to syslog

```go
package main

import (
	"log/syslog"

	"github.com/phuslu/log"
)

func main() {
	logger, err := syslog.NewLogger(syslog.LOG_INFO, 0)
	if err != nil {
		log.Fatal().Err(err).Msg("new syslog error")
	}

	log.DefaultLogger.Writer = logger.Writer()
	log.Info().Str("foo", "bar").Msg("a syslog message")
}
```

### High Performance

A quick and simple benchmark with zap/zerolog/onelog

```go
// go test -v -run=none -bench=. -benchtime=10s -benchmem log_test.go
package main

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/francoispqt/onelog"
	"github.com/phuslu/log"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var fakeMessage = "Test logging, but use a somewhat realistic message length. "

func BenchmarkZap(b *testing.B) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ioutil.Discard),
		zapcore.InfoLevel,
	))
	for i := 0; i < b.N; i++ {
		logger.Info(fakeMessage, zap.String("foo", "bar"), zap.Int("int", 123))
	}
}

func BenchmarkOneLog(b *testing.B) {
	logger := onelog.New(ioutil.Discard, onelog.INFO)
	logger.Hook(func(e onelog.Entry) { e.Int64("time", time.Now().Unix()) })
	for i := 0; i < b.N; i++ {
		logger.InfoWith(fakeMessage).String("foo", "bar").Int("int", 123).Write()
	}
}

func BenchmarkZeroLog(b *testing.B) {
	logger := zerolog.New(ioutil.Discard).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	for i := 0; i < b.N; i++ {
		logger.Info().Str("foo", "bar").Int("int", 123).Msg(fakeMessage)
	}
}

func BenchmarkPhusLog(b *testing.B) {
	logger := log.Logger{
		Timestamp: true,
		Writer:    ioutil.Discard,
	}
	for i := 0; i < b.N; i++ {
		logger.Info().Str("foo", "bar").Int("int", 123).Msg(fakeMessage)
	}
}
```
Performance results on my laptop:
```
BenchmarkZap-16        	14383234	       828 ns/op	     128 B/op	       1 allocs/op
BenchmarkOneLog-16     	42118165	       285 ns/op	       0 B/op	       0 allocs/op
BenchmarkZeroLog-16    	46848562	       252 ns/op	       0 B/op	       0 allocs/op
BenchmarkPhusLog-16    	78545636	       155 ns/op	       0 B/op	       0 allocs/op
```

[pkg-img]: https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square
[pkg]: https://pkg.go.dev/github.com/phuslu/log?tab=doc
[report-img]: https://goreportcard.com/badge/github.com/phuslu/log
[report]: https://goreportcard.com/report/github.com/phuslu/log
[cov-img]: http://gocover.io/_badge/github.com/phuslu/log
[cov]: https://gocover.io/github.com/phuslu/log
[pretty-logging-img]: https://user-images.githubusercontent.com/195836/77247067-5cf24000-6c68-11ea-9e65-6cdc00d82384.png
[play-simple-img]: https://img.shields.io/badge/playground-NGV25aBKmYH-29BEB0?style=flat&logo=go
[play-simple]: https://play.golang.org/p/NGV25aBKmYH
[play-customize-img]: https://img.shields.io/badge/playground-EaFFre1DUVJ-29BEB0?style=flat&logo=go
[play-customize]: https://play.golang.org/p/EaFFre1DUVJ
[play-pretty-img]: https://img.shields.io/badge/playground-62bWGk67apR-29BEB0?style=flat&logo=go
[play-pretty]: https://play.golang.org/p/62bWGk67apR
[play-dynamic-img]: https://img.shields.io/badge/playground-0S--JT7h--QXI-29BEB0?style=flat&logo=go
[play-dynamic]: https://play.golang.org/p/0S-JT7h-QXI
