package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/phuslu/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func decodeLog(t *testing.T, b *bytes.Buffer) map[string]any {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(b.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal log: %v\n%s", err, b.String())
	}
	return got
}

func TestLoggerEmit(t *testing.T) {
	var b bytes.Buffer
	logger := Logger{
		Log: log.Logger{
			Level: log.DebugLevel,
			Context: log.NewContext(nil).
				Str("service", "checkout").
				Value(),
			Writer: log.IOWriter{Writer: &b},
		},
	}

	var record otellog.Record
	record.SetTimestamp(time.Date(2026, 6, 8, 10, 11, 12, 123000000, time.UTC))
	record.SetObservedTimestamp(time.Date(2026, 6, 8, 10, 11, 13, 456000000, time.UTC))
	record.SetSeverity(otellog.SeverityWarn2)
	record.SetSeverityText("WARN2")
	record.SetEventName("checkout.warning")
	record.SetErr(errors.New("card declined"))
	record.SetBody(otellog.StringValue("hello from otel"))
	record.AddAttributes(
		otellog.String("component", "payment"),
		otellog.Int64("answer", 42),
		otellog.Bool("ok", true),
		otellog.Map("nested",
			otellog.String("name", "alice"),
			otellog.Slice("scores", otellog.Int64Value(1), otellog.Float64Value(2.5)),
		),
		otellog.Slice("mixed",
			otellog.StringValue("item"),
			otellog.MapValue(otellog.Bool("enabled", true)),
			otellog.SliceValue(otellog.Int64Value(7), otellog.StringValue("inner")),
		),
	)

	traceID := oteltrace.TraceID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
	spanID := oteltrace.SpanID{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17}
	ctx := oteltrace.ContextWithSpanContext(context.Background(), oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: oteltrace.FlagsSampled,
	}))

	logger.Emit(ctx, record)
	got := decodeLog(t, &b)

	if got[log.LevelKey] != log.WarnLevelString {
		t.Fatalf("level = %v", got[log.LevelKey])
	}
	if got["severity_text"] != "WARN2" {
		t.Fatalf("severity_text = %v", got["severity_text"])
	}
	if got["severity_number"] != float64(14) {
		t.Fatalf("severity_number = %v", got["severity_number"])
	}
	if got[log.MessageKey] != "hello from otel" {
		t.Fatalf("message = %v", got[log.MessageKey])
	}
	if got["service"] != "checkout" || got["component"] != "payment" {
		t.Fatalf("context/attribute fields missing: %#v", got)
	}
	if got["answer"] != float64(42) || got["ok"] != true {
		t.Fatalf("scalar attributes missing: %#v", got)
	}
	if got["trace_id"] != traceID.String() || got["span_id"] != spanID.String() || got["trace_flags"] != "01" {
		t.Fatalf("trace fields missing: %#v", got)
	}
	if got["timestamp"] == nil || got["observed_time"] == nil {
		t.Fatalf("timestamp fields missing: %#v", got)
	}
	if got["event_name"] != "checkout.warning" || got["error"] != "card declined" {
		t.Fatalf("event/error fields missing: %#v", got)
	}

	nested, ok := got["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested attribute = %#v", got["nested"])
	}
	if nested["name"] != "alice" {
		t.Fatalf("nested.name = %#v", nested["name"])
	}
	scores, ok := nested["scores"].([]any)
	if !ok || len(scores) != 2 || scores[0] != float64(1) || scores[1] != 2.5 {
		t.Fatalf("nested.scores = %#v", nested["scores"])
	}
	mixed, ok := got["mixed"].([]any)
	if !ok || len(mixed) != 3 {
		t.Fatalf("mixed = %#v", got["mixed"])
	}
	if mixed[0] != "item" {
		t.Fatalf("mixed[0] = %#v", mixed[0])
	}
	mixedMap, ok := mixed[1].(map[string]any)
	if !ok || mixedMap["enabled"] != true {
		t.Fatalf("mixed[1] = %#v", mixed[1])
	}
	mixedSlice, ok := mixed[2].([]any)
	if !ok || len(mixedSlice) != 2 || mixedSlice[0] != float64(7) || mixedSlice[1] != "inner" {
		t.Fatalf("mixed[2] = %#v", mixed[2])
	}
}

func TestLoggerEmitSeverityNumberWithoutText(t *testing.T) {
	var b bytes.Buffer
	logger := Logger{
		Log: log.Logger{
			Level:  log.DebugLevel,
			Writer: log.IOWriter{Writer: &b},
		},
	}

	var record otellog.Record
	record.SetSeverity(otellog.SeverityWarn3)
	record.SetBody(otellog.StringValue("warn without text"))

	logger.Emit(context.Background(), record)
	got := decodeLog(t, &b)

	if got[log.LevelKey] != log.WarnLevelString {
		t.Fatalf("level = %v", got[log.LevelKey])
	}
	if got["severity_number"] != float64(15) {
		t.Fatalf("severity_number = %v", got["severity_number"])
	}
	if _, ok := got["severity_text"]; ok {
		t.Fatalf("unexpected severity_text = %v", got["severity_text"])
	}
}

func TestLoggerStructuredBody(t *testing.T) {
	var b bytes.Buffer
	logger := Logger{
		Log: log.Logger{
			Level:  log.InfoLevel,
			Writer: log.IOWriter{Writer: &b},
		},
	}

	var record otellog.Record
	record.SetSeverity(otellog.SeverityInfo)
	record.SetBody(otellog.MapValue(
		otellog.String("event", "cache_miss"),
		otellog.Int("attempt", 2),
	))

	logger.Emit(context.Background(), record)
	got := decodeLog(t, &b)

	body, ok := got[log.MessageKey].(map[string]any)
	if !ok {
		t.Fatalf("structured body = %#v", got[log.MessageKey])
	}
	if body["event"] != "cache_miss" || body["attempt"] != float64(2) {
		t.Fatalf("structured body fields = %#v", body)
	}
}

func TestLoggerEnabledAndFatalNoExit(t *testing.T) {
	var b bytes.Buffer
	logger := Logger{
		Log: log.Logger{
			Level:  log.ErrorLevel,
			Writer: log.IOWriter{Writer: &b},
		},
	}

	var info otellog.Record
	info.SetSeverity(otellog.SeverityInfo)
	info.SetBody(otellog.StringValue("filtered"))
	if logger.Enabled(context.Background(), otellog.EnabledParameters{Severity: info.Severity()}) {
		t.Fatal("info record enabled for error logger")
	}
	logger.Emit(context.Background(), info)
	if b.Len() != 0 {
		t.Fatalf("filtered record was emitted: %s", b.String())
	}

	var fatal otellog.Record
	fatal.SetSeverity(otellog.SeverityFatal)
	fatal.SetBody(otellog.StringValue("fatal but no os.Exit"))
	if !logger.Enabled(context.Background(), otellog.EnabledParameters{Severity: fatal.Severity()}) {
		t.Fatal("fatal record disabled for error logger")
	}
	logger.Emit(context.Background(), fatal)
	got := decodeLog(t, &b)
	if got[log.LevelKey] != log.FatalLevelString {
		t.Fatalf("fatal level = %v", got[log.LevelKey])
	}
}

func TestLoggerProviderScope(t *testing.T) {
	var b bytes.Buffer
	provider := LoggerProvider{
		Log: log.Logger{
			Level:  log.InfoLevel,
			Writer: log.IOWriter{Writer: &b},
		},
	}
	logger := provider.Logger(
		"github.com/example/instrumentation",
		otellog.WithInstrumentationVersion("v1.2.3"),
		otellog.WithInstrumentationAttributes(attribute.String("library.language", "go")),
		otellog.WithSchemaURL("https://example.com/schema"),
	)

	var record otellog.Record
	record.SetSeverity(otellog.SeverityInfo)
	record.SetBody(otellog.StringValue("scoped"))
	logger.Emit(context.Background(), record)

	got := decodeLog(t, &b)
	if got["scope_name"] != "github.com/example/instrumentation" ||
		got["scope_version"] != "v1.2.3" ||
		got["scope_schema_url"] != "https://example.com/schema" {
		t.Fatalf("scope fields missing: %#v", got)
	}
	scopeAttrs, ok := got["scope_attributes"].(map[string]any)
	if !ok || scopeAttrs["library.language"] != "go" {
		t.Fatalf("scope attributes missing: %#v", got["scope_attributes"])
	}
}

func nestedBenchmarkRecord() otellog.Record {
	var record otellog.Record
	record.SetTimestamp(time.Date(2026, 6, 8, 10, 11, 12, 123000000, time.UTC))
	record.SetSeverity(otellog.SeverityInfo)
	record.SetSeverityText("INFO")
	record.SetBody(otellog.StringValue("bench"))
	record.AddAttributes(
		otellog.String("component", "payment"),
		otellog.Int64("answer", 42),
		otellog.Slice("mixed",
			otellog.StringValue("item"),
			otellog.Int64Value(7),
			otellog.Float64Value(2.5),
			otellog.MapValue(
				otellog.String("name", "alice"),
				otellog.Slice("scores", otellog.Int64Value(1), otellog.Int64Value(2)),
			),
		),
	)
	return record
}

func nestedBenchmarkSDKRecord() sdklog.Record {
	var record sdklog.Record
	record.SetTimestamp(time.Date(2026, 6, 8, 10, 11, 12, 123000000, time.UTC))
	record.SetSeverity(otellog.SeverityInfo)
	record.SetSeverityText("INFO")
	record.SetBody(otellog.StringValue("bench"))
	record.SetAttributes(
		otellog.String("component", "payment"),
		otellog.Int64("answer", 42),
		otellog.Slice("mixed",
			otellog.StringValue("item"),
			otellog.Int64Value(7),
			otellog.Float64Value(2.5),
			otellog.MapValue(
				otellog.String("name", "alice"),
				otellog.Slice("scores", otellog.Int64Value(1), otellog.Int64Value(2)),
			),
		),
	)
	return record
}

func BenchmarkPhusluLoggerEmitNestedValues(b *testing.B) {
	logger := Logger{
		Log: log.Logger{
			Level:  log.InfoLevel,
			Writer: log.IOWriter{Writer: io.Discard},
		},
	}
	record := nestedBenchmarkRecord()

	logger.Emit(context.Background(), record)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Emit(context.Background(), record)
	}
}

func BenchmarkOTelStdoutExporterNestedValues(b *testing.B) {
	exporter, err := stdoutlog.New(stdoutlog.WithWriter(io.Discard))
	if err != nil {
		b.Fatalf("create stdoutlog exporter: %v", err)
	}
	b.Cleanup(func() {
		if err := exporter.Shutdown(context.Background()); err != nil {
			b.Fatalf("shutdown stdoutlog exporter: %v", err)
		}
	})
	records := []sdklog.Record{nestedBenchmarkSDKRecord()}

	if err := exporter.Export(context.Background(), records); err != nil {
		b.Fatalf("export warmup record: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := exporter.Export(context.Background(), records); err != nil {
			b.Fatalf("export record: %v", err)
		}
	}
}
