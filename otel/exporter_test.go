package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func stdoutExporterRecord() sdklog.Record {
	var record sdklog.Record
	record.SetTimestamp(time.Date(2026, 6, 8, 10, 11, 12, 123000000, time.UTC))
	record.SetObservedTimestamp(time.Date(2026, 6, 8, 10, 11, 13, 456000000, time.UTC))
	record.SetEventName("testing.event")
	record.SetSeverity(otellog.SeverityWarn2)
	record.SetSeverityText("WARN2")
	record.SetBody(otellog.StringValue("test"))
	record.SetTraceID(oteltrace.TraceID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f})
	record.SetSpanID(oteltrace.SpanID{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17})
	record.SetTraceFlags(oteltrace.FlagsSampled)
	record.SetAttributes(
		otellog.String("string", "value"),
		otellog.Int64("int", 42),
		otellog.Float64("float", 2.5),
		otellog.Bool("bool", true),
		otellog.Bytes("bytes", []byte{1, 2, 3}),
		otellog.Map("map",
			otellog.String("name", "alice"),
			otellog.Slice("scores", otellog.Int64Value(1), otellog.Float64Value(2.5)),
		),
		otellog.Slice("slice", otellog.StringValue("item"), otellog.Int64Value(7)),
		otellog.KeyValue{Key: "empty"},
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

func decodeJSONStream(t *testing.T, data []byte) []any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	var values []any
	for {
		var value any
		if err := dec.Decode(&value); err != nil {
			if errors.Is(err, io.EOF) {
				return values
			}
			t.Fatalf("decode JSON stream: %v\n%s", err, string(data))
		}
		values = append(values, value)
	}
}

func TestStdoutExporterMatchesOfficial(t *testing.T) {
	records := []sdklog.Record{stdoutExporterRecord(), nestedBenchmarkSDKRecord()}
	tests := []struct {
		name            string
		exporter        StdoutExporter
		officialOptions []stdoutlog.Option
	}{
		{name: "default"},
		{
			name:            "without timestamps",
			exporter:        StdoutExporter{DisableTimestamps: true},
			officialOptions: []stdoutlog.Option{stdoutlog.WithoutTimestamps()},
		},
		{
			name:            "pretty",
			exporter:        StdoutExporter{PrettyPrint: true},
			officialOptions: []stdoutlog.Option{stdoutlog.WithPrettyPrint()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got, want bytes.Buffer

			exporter := tt.exporter
			exporter.Writer = &got
			official, err := stdoutlog.New(append(tt.officialOptions, stdoutlog.WithWriter(&want))...)
			if err != nil {
				t.Fatalf("create official exporter: %v", err)
			}

			if err := exporter.Export(context.Background(), records); err != nil {
				t.Fatalf("export: %v", err)
			}
			if err := official.Export(context.Background(), records); err != nil {
				t.Fatalf("official export: %v", err)
			}

			if got.String() != want.String() {
				gotJSON := decodeJSONStream(t, got.Bytes())
				wantJSON := decodeJSONStream(t, want.Bytes())
				if !reflect.DeepEqual(gotJSON, wantJSON) {
					t.Fatalf("exported JSON mismatch\ngot:  %s\nwant: %s", got.String(), want.String())
				}
			}
		})
	}
}

func TestStdoutExporterShutdownAndContext(t *testing.T) {
	var b bytes.Buffer
	exporter := StdoutExporter{Writer: &b}
	if err := exporter.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}
	if err := exporter.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if err := exporter.Export(context.Background(), []sdklog.Record{stdoutExporterRecord()}); err != nil {
		t.Fatalf("export after shutdown: %v", err)
	}
	if b.Len() != 0 {
		t.Fatalf("export after shutdown wrote %q", b.String())
	}

	exporter = StdoutExporter{Writer: &b}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := exporter.Export(ctx, []sdklog.Record{stdoutExporterRecord()}); !errors.Is(err, context.Canceled) {
		t.Fatalf("export canceled context error = %v", err)
	}
	if b.Len() != 0 {
		t.Fatalf("canceled export wrote %q", b.String())
	}
}

func BenchmarkPhusluStdoutExporterNestedValues(b *testing.B) {
	exporter := StdoutExporter{Writer: io.Discard}
	b.Cleanup(func() {
		if err := exporter.Shutdown(context.Background()); err != nil {
			b.Fatalf("shutdown stdout exporter: %v", err)
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
