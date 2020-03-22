# Structured Logging for Humans

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/phuslu/log) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/phuslu/log/master/LICENSE) [![goreport](https://goreportcard.com/badge/github.com/phuslu/log)](https://goreportcard.com/report/github.com/phuslu/log)  [![gocover](http://gocover.io/_badge/github.com/phuslu/log)](http://gocover.io/github.com/phuslu/log)

## Features

* Simple & Straightforward interfaces
* JSON/GRPC/CSV/TSV/Printf Loggers
* Rotating File Writer
* Pretty Console Writer(with windows 7/8/10 support)
* Effective, Outperforms [zerolog](https://github.com/rs/zerolog) and [zap](https://github.com/uber-go/zap)
* No external dependences

## Getting Started

### Simple Logging Example

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

### Pretty logging

To log a human-friendly, colorized output, use `log.ConsoleWriter`:

```go
package main

import (
	"errors"
	"os"
	"github.com/phuslu/log"
)

func main() {
	if log.IsTerminal(os.Stderr.Fd()) {
		log.DefaultLogger = log.Logger{
			Caller: 1,
			Writer: &log.ConsoleWriter{ANSIColor: true},
		}
	}

	log.Printf("a printf style line")
	log.Info().Err(errors.New("an error")).Int("everything", 42).Str("foo", "bar").Msg("hello world")
}
```
![Pretty logging](https://user-images.githubusercontent.com/195836/77247067-5cf24000-6c68-11ea-9e65-6cdc00d82384.png)
> Note: pretty logging also works on windows console

### Customize the configuration and formatting:

```go
log.DefaultLogger = log.Logger{
	Level:      log.InfoLevel,
	Caller:     1,
	TimeField:  "date",
	TimeFormat: "2006-01-02",
	Writer:     &log.Writer{},
}
log.Info().Msg("hello world")

// Output: {"date":"2019-07-04","level":"info","caller":"test.go:42","message":"hello world"}
```

### Multi Writers:

```go
log.DefaultLogger.Writer = io.MultiWriter(&log.Writer{Filename: "1.log"}, &log.ConsoleWriter{ANSIColor: true})
log.Info().Msg("hello world")
```

### Rotating log files hourly

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
		Writer:     &log.Writer{
			Filename:   "main.log",
			MaxSize:    50*1024*1024,
			MaxBackups: 7,
			LocalTime:  false,
		},
	}

	runner := cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC), cron.WithLogger(cron.PrintfLogger(log.DefaultLogger)))
	runner.AddFunc("0 0 * * * *", func() { logger.Writer.(*log.Writer).Rotate() })
	go runner.Run()

	for {
		time.Sleep(time.Second)
		logger.Info().Msg("hello world")
	}
}
```
