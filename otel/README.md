# phuslog OpenTelemetry Logs Adapter

This submodule bridges OpenTelemetry Logs API records to `github.com/phuslu/log`.

```go
package main

import (
	"context"

	phuslog "github.com/phuslu/log"
	phuslogotel "github.com/phuslu/log/otel"
	otellog "go.opentelemetry.io/otel/log"
)

func main() {
	logger := phuslogotel.Logger{
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

Nested OpenTelemetry values are encoded without building `[]any` or
`map[string]any` intermediates. This keeps direct `otellog.Record` slice
values, including maps and nested slices, on the zero-allocation path.

The slog comparison uses the official
`go.opentelemetry.io/contrib/bridges/otelslog` handler with the official
OpenTelemetry Logs SDK `LoggerProvider`, `SimpleProcessor`, and `stdoutlog`
JSON exporter writing to `io.Discard`. This keeps both paths on synchronous JSON
encoding and discard-output paths.

Measured with Go 1.26.2 on linux/arm64:

```sh
go test -run=^$ -bench='Benchmark(LoggerEmitNestedValues|OTelSlogHandlerNestedValues)$' -benchmem -count=10
```

Representative result:

```text
BenchmarkLoggerEmitNestedValues-4        1256209    951.7 ns/op     0 B/op   0 allocs/op
BenchmarkOTelSlogHandlerNestedValues-4     38079  31641 ns/op    3896 B/op  69 allocs/op
```

Across 10 runs:

| Benchmark | Path | Result range |
| --- | --- | --- |
| `BenchmarkLoggerEmitNestedValues` | prebuilt `otellog.Record` to phuslog JSON writer | `951.7-958.7 ns/op`, `0 B/op`, `0 allocs/op` |
| `BenchmarkOTelSlogHandlerNestedValues` | prebuilt `slog.Record` through `otelslog`, SDK provider, `SimpleProcessor`, and `stdoutlog` JSON exporter | `31641-33516 ns/op`, `3895-3897 B/op`, `69 allocs/op` |
