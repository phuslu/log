# Structured Logging for Humans

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/phuslu/log) [![goreport](https://goreportcard.com/badge/github.com/phuslu/log)](https://goreportcard.com/report/github.com/phuslu/log) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/phuslu/log/master/LICENSE)

## Features

* No external dependences
* Simple & Straightforward interfaces
* JSON/TSV/GRPC/Printf Loggers
* Rotating File Writer
* Pretty Console Writer(with windows 7/8/10 support)
* Dynamic log Level
* High Performance

## Getting Started

### Simple Logging Example

A out of box example. [![playground](https://img.shields.io/badge/playground-600IpaPBF95-29BEB0?style=flat&logo=go)](https://play.golang.org/p/600IpaPBF95)
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

To customize logger filed name and format. [![playground](https://img.shields.io/badge/playground-wXPaGTjBJcX-29BEB0?style=flat&logo=go)](https://play.golang.org/p/wXPaGTjBJcX)
```go
log.DefaultLogger = log.Logger{
	Level:      log.InfoLevel,
	Caller:     1,
	TimeField:  "date",
	TimeFormat: "2006-01-02",
	Writer:     os.Stderr,
}
log.Info().Str("foo", "bar").Msg("hello world")

// Output: {"date":"2019-07-04","level":"info","caller":"test.go:42","foo":"bar","message":"hello world"}
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

To log a human-friendly, colorized output, use `log.ConsoleWriter`. [![playground](https://img.shields.io/badge/playground-62bWGk67apR-29BEB0?style=flat&logo=go)](https://play.golang.org/p/62bWGk67apR)

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
![Pretty logging](https://user-images.githubusercontent.com/195836/77247067-5cf24000-6c68-11ea-9e65-6cdc00d82384.png)
> Note: pretty logging also works on windows console

### Dynamic log Level

To change log level on the fly, use `log.DefaultLogger.SetLevel`. [![playground](https://img.shields.io/badge/playground-0S--JT7h--QXI-29BEB0?style=flat&logo=go)](https://play.golang.org/p/0S-JT7h-QXI)

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

```go
// go test -v -run=none -bench=. log_test.go
package main

import (
	"io/ioutil"
	"testing"

	"github.com/phuslu/log"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BenchmarkZapSugar(b *testing.B) {
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(ioutil.Discard),
		zapcore.DebugLevel,
	)).Sugar()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Infow("hello world","foo", "bar")
	}
}

func BenchmarkZeroLog(b *testing.B) {
	logger := zerolog.New(ioutil.Discard).With().Timestamp().Logger()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("foo", "bar").Msgf("hello %s", "world")
	}
}

func BenchmarkPhusLog(b *testing.B) {
	logger := log.Logger{
		Timestamp: true,
		Level:     log.DebugLevel,
		Writer:    ioutil.Discard,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("foo", "bar").Msgf("hello %s", "world")
	}
}
```
Performance results on my laptop
```
BenchmarkZapSugar-16    	 1850250	       690 ns/op	      32 B/op	       1 allocs/op
BenchmarkZeroLog-16     	 2570610	       471 ns/op	      16 B/op	       1 allocs/op
BenchmarkPhusLog-16     	 5508972	       218 ns/op	       0 B/op	       0 allocs/op
```