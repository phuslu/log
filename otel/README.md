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
`map[string]any` intermediates. This keeps slice values, including maps and
nested slices, on the zero-allocation path.

Measured with Go 1.26.2 on linux/arm64:

```sh
go test -run=^$ -bench=BenchmarkLoggerEmitNestedValues -benchmem -count=10
```

Representative result:

```text
BenchmarkLoggerEmitNestedValues-4  1754367  691.0 ns/op  0 B/op  0 allocs/op
```

Across 10 runs: `682.5-695.2 ns/op`, `0 B/op`, `0 allocs/op`.
