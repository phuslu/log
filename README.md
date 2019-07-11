# Structured Logging for Humans

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/phuslu/log)

## Features

* Effective
* Level Logging
* File Rotating/Buffering
* JSON and TSV Formats
* Pretty Logging for Console
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

To log a human-friendly, colorized output, use `log.JSONConsoleWriter`:

```go
log.DefaultLogger.Writer = &log.JSONConsoleWriter{ANSIColor: true}

log.Info().Str("foo", "bar").Msg("hello world")

// Output: 2019-07-10T05:35:54.277Z INF test.go:42 > hello world foo=bar
```

### Customize the configuration and formatting:

```go
logger := log.JSONLogger{
	Level:      log.DebugLevel,
	Caller:     true,
	EscapeHTML: true,
	TimeField:  "_time",
	TimeFormat: time.RFC850,
	Writer:     &log.Writer{},
}
logger.Info().Msg("hello world")

// Output: {"_time":"11 Jul 19 01:00 CST","level":"info","caller":"test.go:42","message":"hello world"}
```

### Rotating log files hourly

```go
package main

import (
	"github.com/phuslu/log"
	"github.com/robfig/cron"
)

func main() {
	var localtime bool = true

	logger := log.JSONLogger{
		Level:      ParseLevel("info"),
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

	logger.Info().Msg("hello world")
}
```
