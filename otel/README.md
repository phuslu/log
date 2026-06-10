# phuslog OpenTelemetry Logs Adapter

This submodule is for applications that want OpenTelemetry Logs API
compatibility while keeping phuslog JSON as the log output format. It bridges
OpenTelemetry Logs API records to `github.com/phuslu/log`, preserving common
OpenTelemetry metadata such as severity, trace context, scope, and attributes as
JSON fields.

It is not an OTLP exporter and does not implement the OpenTelemetry SDK
pipeline. To send logs to an OpenTelemetry Collector, write phuslog JSON to
stdout or a file and configure a Collector receiver, or use an OpenTelemetry SDK
exporter separately.

```go
package main

import (
	"context"

	phuslog "github.com/phuslu/log"
	phuslogotel "github.com/phuslu/log/otel"
	otellog "go.opentelemetry.io/otel/log"
)

func main() {
	var logger otellog.Logger = phuslogotel.Logger{
		Log: phuslog.Logger{
			Level: phuslog.InfoLevel,
		},
	}

	var record otellog.Record
	record.SetSeverity(otellog.SeverityInfo)
	record.SetBody(otellog.StringValue("hello from otel"))
	record.AddAttributes(otellog.String("component", "worker"))

	logger.Emit(context.Background(), record)
}
```

Use `LoggerProvider` when integrating with OpenTelemetry code that expects a `log.LoggerProvider`.

## Performance

Benchmarks compare this adapter with the official OpenTelemetry `stdoutlog`
exporter, both writing equivalent nested log records as JSON to `io.Discard`.
The `stdoutlog` benchmark is an exporter-only lower bound; it skips the SDK
logger and processor path.
In this nested-record benchmark, the direct adapter path runs with `0 B/op` and
`0 allocs/op`.

Measured with Go 1.26.2 on linux/arm64:

```sh
go test -run=^$ -bench='Benchmark(PhusluLoggerEmitNestedValues|OTelStdoutExporterNestedValues)$' -benchmem -count=10
```

Representative result:

```text
BenchmarkPhusluLoggerEmitNestedValues-4         1000000    1008 ns/op       0 B/op   0 allocs/op
BenchmarkOTelStdoutExporterNestedValues-4         67358   18098 ns/op    2178 B/op  41 allocs/op
```

Across 10 runs:

| Benchmark | Path | Result range |
| --- | --- | --- |
| `BenchmarkPhusluLoggerEmitNestedValues` | `otellog.Record` to phuslog JSON | `1000-1015 ns/op`, `0 B/op`, `0 allocs/op` |
| `BenchmarkOTelStdoutExporterNestedValues` | `sdklog.Record` through `stdoutlog` | `17964-18289 ns/op`, `2177-2178 B/op`, `41 allocs/op` |
