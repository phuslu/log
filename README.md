# Structured Logging for Humans

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/phuslu/log) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/phuslu/log/master/LICENSE) [![goreport](https://goreportcard.com/badge/github.com/phuslu/log)](https://goreportcard.com/report/github.com/phuslu/log)  [![gocover](http://gocover.io/_badge/github.com/phuslu/log)](http://gocover.io/github.com/phuslu/log)

## Features

* Simple & Straightforward interfaces
* Effective, Outperforms [zerolog](https://github.com/rs/zerolog) and [zap](https://github.com/uber-go/zap)
* Rotating/Pretty/Buffering Writers
* No external dependences

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
package main

import (
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

	log.Info().Str("foo", "bar").Msg("hello world")
}

// Output: 2019-07-11T16:41:43.256Z INF 12 pretty.go:10 > hello world foo=bar
```
![](https://user-images.githubusercontent.com/195836/61725177-dff18c80-ada1-11e9-8338-3374a21a625b.png)
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
	"github.com/robfig/cron"
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
