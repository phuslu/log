package otel

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// StdoutExporter writes OpenTelemetry SDK log records as standard stdout JSON.
type StdoutExporter struct {
	// Writer is the destination. If nil, os.Stdout is used.
	Writer io.Writer

	// PrettyPrint prettifies the emitted output.
	PrettyPrint bool

	// DisableTimestamps omits timestamp fields from the emitted output.
	DisableTimestamps bool

	mu       sync.Mutex
	shutdown atomic.Bool
}

// Export writes records to the configured writer.
func (e *StdoutExporter) Export(ctx context.Context, records []sdklog.Record) error {
	if e.shutdown.Load() {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.shutdown.Load() {
		return nil
	}
	writer := e.Writer
	if writer == nil {
		writer = os.Stdout
	}

	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return err
		}

		b := make([]byte, 0, 1024)
		if err := appendStdoutLogRecord(&b, record, !e.DisableTimestamps); err != nil {
			return err
		}
		if e.PrettyPrint {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, b, "", "\t"); err != nil {
				return err
			}
			pretty.WriteByte('\n')
			if _, err := writer.Write(pretty.Bytes()); err != nil {
				return err
			}
			continue
		}

		b = append(b, '\n')
		if _, err := writer.Write(b); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown shuts down the Exporter.
func (e *StdoutExporter) Shutdown(context.Context) error {
	e.shutdown.Store(true)
	return nil
}

// ForceFlush performs no action.
func (*StdoutExporter) ForceFlush(context.Context) error {
	return nil
}

func appendStdoutLogRecord(b *[]byte, record sdklog.Record, timestamps bool) error {
	*b = append(*b, '{')
	first := true

	if timestamps {
		appendStdoutFieldPrefix(b, &first, "Timestamp")
		timestamp, err := record.Timestamp().MarshalJSON()
		if err != nil {
			return err
		}
		*b = append(*b, timestamp...)

		appendStdoutFieldPrefix(b, &first, "ObservedTimestamp")
		observed, err := record.ObservedTimestamp().MarshalJSON()
		if err != nil {
			return err
		}
		*b = append(*b, observed...)
	}

	if eventName := record.EventName(); eventName != "" {
		appendStdoutFieldPrefix(b, &first, "EventName")
		appendJSONString(b, eventName)
	}

	appendStdoutFieldPrefix(b, &first, "Severity")
	*b = strconv.AppendInt(*b, int64(record.Severity()), 10)

	appendStdoutFieldPrefix(b, &first, "SeverityText")
	appendJSONString(b, record.SeverityText())

	appendStdoutFieldPrefix(b, &first, "Body")
	if err := appendStdoutLogValue(b, record.Body()); err != nil {
		return err
	}

	appendStdoutFieldPrefix(b, &first, "Attributes")
	if err := appendStdoutLogAttributes(b, record); err != nil {
		return err
	}

	appendStdoutFieldPrefix(b, &first, "TraceID")
	appendJSONString(b, record.TraceID().String())

	appendStdoutFieldPrefix(b, &first, "SpanID")
	appendJSONString(b, record.SpanID().String())

	appendStdoutFieldPrefix(b, &first, "TraceFlags")
	appendJSONString(b, record.TraceFlags().String())

	appendStdoutFieldPrefix(b, &first, "Resource")
	if resource := record.Resource(); resource == nil {
		*b = append(*b, "null"...)
	} else {
		resourceJSON, err := json.Marshal(resource)
		if err != nil {
			return err
		}
		*b = append(*b, resourceJSON...)
	}

	appendStdoutFieldPrefix(b, &first, "Scope")
	scopeJSON, err := json.Marshal(record.InstrumentationScope())
	if err != nil {
		return err
	}
	*b = append(*b, scopeJSON...)

	appendStdoutFieldPrefix(b, &first, "DroppedAttributes")
	*b = strconv.AppendInt(*b, int64(record.DroppedAttributes()), 10)

	*b = append(*b, '}')
	return nil
}

func appendStdoutFieldPrefix(b *[]byte, first *bool, name string) {
	if *first {
		*first = false
	} else {
		*b = append(*b, ',')
	}
	appendJSONString(b, name)
	*b = append(*b, ':')
}

func appendStdoutLogAttributes(b *[]byte, record sdklog.Record) error {
	*b = append(*b, '[')
	i := 0
	var err error
	record.WalkAttributes(func(kv otellog.KeyValue) bool {
		if i != 0 {
			*b = append(*b, ',')
		}
		err = appendStdoutLogKeyValue(b, kv)
		i++
		return err == nil
	})
	*b = append(*b, ']')
	return err
}

func appendStdoutLogKeyValue(b *[]byte, kv otellog.KeyValue) error {
	*b = append(*b, `{"Key":`...)
	appendJSONString(b, kv.Key)
	*b = append(*b, `,"Value":`...)
	if err := appendStdoutLogValue(b, kv.Value); err != nil {
		return err
	}
	*b = append(*b, '}')
	return nil
}

func appendStdoutLogValue(b *[]byte, value otellog.Value) error {
	*b = append(*b, '{')
	first := true

	appendStdoutFieldPrefix(b, &first, "Type")
	appendJSONString(b, value.Kind().String())

	appendStdoutFieldPrefix(b, &first, "Value")
	switch value.Kind() {
	case otellog.KindString:
		appendJSONString(b, value.AsString())
	case otellog.KindInt64:
		*b = strconv.AppendInt(*b, value.AsInt64(), 10)
	case otellog.KindFloat64:
		f := value.AsFloat64()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return errors.New("otel: unsupported non-finite float value")
		}
		*b = appendJSONFloat(*b, f, 64)
	case otellog.KindBool:
		*b = strconv.AppendBool(*b, value.AsBool())
	case otellog.KindBytes:
		appendJSONString(b, base64.StdEncoding.EncodeToString(value.AsBytes()))
	case otellog.KindMap:
		if err := appendStdoutLogValueMap(b, value.AsMap()); err != nil {
			return err
		}
	case otellog.KindSlice:
		if err := appendStdoutLogValueSlice(b, value.AsSlice()); err != nil {
			return err
		}
	case otellog.KindEmpty:
		*b = append(*b, "null"...)
	default:
		return errors.New("otel: invalid log value kind")
	}

	*b = append(*b, '}')
	return nil
}

func appendStdoutLogValueMap(b *[]byte, values []otellog.KeyValue) error {
	*b = append(*b, '[')
	for i, kv := range values {
		if i != 0 {
			*b = append(*b, ',')
		}
		if err := appendStdoutLogKeyValue(b, kv); err != nil {
			return err
		}
	}
	*b = append(*b, ']')
	return nil
}

func appendStdoutLogValueSlice(b *[]byte, values []otellog.Value) error {
	*b = append(*b, '[')
	for i, value := range values {
		if i != 0 {
			*b = append(*b, ',')
		}
		if err := appendStdoutLogValue(b, value); err != nil {
			return err
		}
	}
	*b = append(*b, ']')
	return nil
}

var _ sdklog.Exporter = (*StdoutExporter)(nil)
