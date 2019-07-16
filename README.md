# Structured Logging for Humans

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/phuslu/log) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/phuslu/log/master/LICENSE) [![goreport](https://goreportcard.com/badge/github.com/phuslu/log)](https://goreportcard.com/report/github.com/phuslu/log)  [![gocover](http://gocover.io/_badge/github.com/phuslu/log)](http://gocover.io/github.com/phuslu/log)

## Features

* Simple, Minimalist interface
* Effective, Outperforms [zerolog](https://github.com/rs/zerolog) and [zap](https://github.com/uber-go/zap)
* JSON and TSV/CSV Loggers
* Rotating/Buffering/Pretty Writers
* Adaptation for glog/grpc

## Getting Started

### Simple Logging Example

```go
package main

import (
	"github.com/phuslu/log"
)

func main() {
	log.Debug().Str("foo", "bar").Msg("hello world")
}

// Output: {"time":"2019-07-10T16:00:19.092Z","level":"debug","foo":"bar","message":"hello world"}
```
> Note: By default log writes to `os.Stderr`

### Pretty logging

To log a human-friendly, colorized output, use `log.ConsoleWriter`:

```go
log.DefaultLogger.Writer = &log.ConsoleWriter{ANSIColor: true}
log.DefaultLogger.Caller = true

log.Info().Str("foo", "bar").Msg("hello world")

// Output: 2019-07-11T16:41:43.256Z INF pretty.go:10 > hello world foo=bar
```
![](https://user-images.githubusercontent.com/195836/61068992-ec8af200-a43d-11e9-891f-c6987b402f21.png)
> Note: pretty logging also works on windows console

### Customize the configuration and formatting:

```go
log.DefaultLogger := log.Logger{
	Level:      log.DebugLevel,
	Caller:     true,
	TimeField:  "_time",
	TimeFormat: "0102 15:04:05",
	Writer:     &log.Writer{},
}
log.Info().Msg("hello world")

// Output: {"_time":"0717 01:07:18","level":"info","caller":"test.go:42","message":"hello world"}
```

### Rotating log files hourly

```go
package main

import (
	"time"

	"github.com/phuslu/log"
	"github.com/robfig/cron"
)

func main() {
	var localtime bool = true

	logger := log.Logger{
		Level:      log.ParseLevel("info"),
		Writer:     &log.Writer{
			Filename:   "main.log",
			MaxSize:    50*1024*1024,
			MaxBackups: 7,
			LocalTime:  localtime,
		},
	}

	var runner *cron.Cron
	if localtime {
		runner = cron.New()
	} else {
		runner = cron.NewWithLocation(time.UTC)
	}
	runner.AddFunc("0 0 * * * *", func() { logger.Writer.(*log.Writer).Rotate() })
	runner.ErrorLog = log.DefaultLogger
	go runner.Run()

	for {
		time.Sleep(time.Second)
		logger.Info().Msg("hello world")
	}
}
```
