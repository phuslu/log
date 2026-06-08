package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	phuslog "github.com/phuslu/log"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
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
	logger := NewLogger(&phuslog.Logger{
		Level: phuslog.DebugLevel,
		Context: phuslog.NewContext(nil).
			Str("service", "checkout").
			Value(),
		Writer: phuslog.IOWriter{Writer: &b},
	})

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

	if got[phuslog.LevelKey] != phuslog.WarnLevelString {
		t.Fatalf("level = %v", got[phuslog.LevelKey])
	}
	if got["severity_text"] != "WARN2" {
		t.Fatalf("severity_text = %v", got["severity_text"])
	}
	if got[phuslog.MessageKey] != "hello from otel" {
		t.Fatalf("message = %v", got[phuslog.MessageKey])
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
}

func TestLoggerStructuredBody(t *testing.T) {
	var b bytes.Buffer
	logger := NewLogger(&phuslog.Logger{
		Level:  phuslog.InfoLevel,
		Writer: phuslog.IOWriter{Writer: &b},
	})

	var record otellog.Record
	record.SetSeverity(otellog.SeverityInfo)
	record.SetBody(otellog.MapValue(
		otellog.String("event", "cache_miss"),
		otellog.Int("attempt", 2),
	))

	logger.Emit(context.Background(), record)
	got := decodeLog(t, &b)

	body, ok := got[phuslog.MessageKey].(map[string]any)
	if !ok {
		t.Fatalf("structured body = %#v", got[phuslog.MessageKey])
	}
	if body["event"] != "cache_miss" || body["attempt"] != float64(2) {
		t.Fatalf("structured body fields = %#v", body)
	}
}

func TestLoggerEnabledAndFatalNoExit(t *testing.T) {
	var b bytes.Buffer
	logger := NewLogger(&phuslog.Logger{
		Level:  phuslog.ErrorLevel,
		Writer: phuslog.IOWriter{Writer: &b},
	})

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
	if got[phuslog.LevelKey] != phuslog.FatalLevelString {
		t.Fatalf("fatal level = %v", got[phuslog.LevelKey])
	}
}

func TestLoggerProviderScope(t *testing.T) {
	var b bytes.Buffer
	provider := NewLoggerProvider(&phuslog.Logger{
		Level:  phuslog.InfoLevel,
		Writer: phuslog.IOWriter{Writer: &b},
	})
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
