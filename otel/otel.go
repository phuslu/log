package otel

import (
	"context"
	"math"
	"strconv"
	"sync"

	"github.com/phuslu/log"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// FieldNames controls the phuslog fields used for OpenTelemetry metadata.
type FieldNames struct {
	Timestamp       string
	ObservedTime    string
	EventName       string
	SeverityText    string
	TraceID         string
	SpanID          string
	TraceFlags      string
	ScopeName       string
	ScopeVersion    string
	ScopeSchemaURL  string
	ScopeAttributes string
}

// Logger bridges OpenTelemetry log records into a phuslog Logger.
type Logger struct {
	embedded.Logger

	// Logger specifies the output logger.
	Log log.Logger

	// FieldNames overrides non-empty OpenTelemetry metadata field names.
	FieldNames FieldNames

	// DisableTraceContext disables trace_id, span_id, and trace_flags extraction.
	DisableTraceContext bool

	// DisableScopeFields disables instrumentation scope fields from LoggerProvider.
	DisableScopeFields bool

	scopeName    string
	scopeVersion string
	scopeSchema  string
	scopeAttrs   []attribute.KeyValue
}

// LoggerProvider creates OpenTelemetry loggers backed by a phuslog Logger.
type LoggerProvider struct {
	embedded.LoggerProvider

	// Log specifies the output logger.
	Log log.Logger

	// FieldNames overrides non-empty OpenTelemetry metadata field names.
	FieldNames FieldNames

	// DisableTraceContext disables trace_id, span_id, and trace_flags extraction.
	DisableTraceContext bool

	// DisableScopeFields disables instrumentation scope fields.
	DisableScopeFields bool
}

var defaultFieldNames = FieldNames{
	Timestamp:       "timestamp",
	ObservedTime:    "observed_time",
	EventName:       "event_name",
	SeverityText:    "severity_text",
	TraceID:         "trace_id",
	SpanID:          "span_id",
	TraceFlags:      "trace_flags",
	ScopeName:       "scope_name",
	ScopeVersion:    "scope_version",
	ScopeSchemaURL:  "scope_schema_url",
	ScopeAttributes: "scope_attributes",
}

func fieldNames(names FieldNames) FieldNames {
	fields := defaultFieldNames
	if names.Timestamp != "" {
		fields.Timestamp = names.Timestamp
	}
	if names.ObservedTime != "" {
		fields.ObservedTime = names.ObservedTime
	}
	if names.EventName != "" {
		fields.EventName = names.EventName
	}
	if names.SeverityText != "" {
		fields.SeverityText = names.SeverityText
	}
	if names.TraceID != "" {
		fields.TraceID = names.TraceID
	}
	if names.SpanID != "" {
		fields.SpanID = names.SpanID
	}
	if names.TraceFlags != "" {
		fields.TraceFlags = names.TraceFlags
	}
	if names.ScopeName != "" {
		fields.ScopeName = names.ScopeName
	}
	if names.ScopeVersion != "" {
		fields.ScopeVersion = names.ScopeVersion
	}
	if names.ScopeSchemaURL != "" {
		fields.ScopeSchemaURL = names.ScopeSchemaURL
	}
	if names.ScopeAttributes != "" {
		fields.ScopeAttributes = names.ScopeAttributes
	}
	return fields
}

func (l Logger) log() log.Logger {
	logger := l.Log
	// OpenTelemetry records do not carry a program counter. Keeping Caller on
	// would report this adapter instead of the application call site.
	logger.Caller = 0
	return logger
}

// Logger returns a scoped OpenTelemetry logger.
func (p LoggerProvider) Logger(name string, options ...otellog.LoggerOption) otellog.Logger {
	cfg := otellog.NewLoggerConfig(options...)
	attrs := cfg.InstrumentationAttributes()
	return Logger{
		Log:                 p.Log,
		FieldNames:          p.FieldNames,
		DisableTraceContext: p.DisableTraceContext,
		DisableScopeFields:  p.DisableScopeFields,
		scopeName:           name,
		scopeVersion:        cfg.InstrumentationVersion(),
		scopeSchema:         cfg.SchemaURL(),
		scopeAttrs:          attrs.ToSlice(),
	}
}

func logLevel(severity otellog.Severity) (log.Level, bool) {
	switch {
	case severity >= otellog.SeverityFatal1:
		return log.FatalLevel, true
	case severity >= otellog.SeverityError1:
		return log.ErrorLevel, true
	case severity >= otellog.SeverityWarn1:
		return log.WarnLevel, true
	case severity >= otellog.SeverityInfo1:
		return log.InfoLevel, true
	case severity >= otellog.SeverityDebug1:
		return log.DebugLevel, true
	case severity >= otellog.SeverityTrace1:
		return log.TraceLevel, true
	default:
		return 0, false
	}
}

func levelString(level log.Level) string {
	switch level {
	case log.TraceLevel:
		return log.TraceLevelString
	case log.DebugLevel:
		return log.DebugLevelString
	case log.InfoLevel:
		return log.InfoLevelString
	case log.WarnLevel:
		return log.WarnLevelString
	case log.ErrorLevel:
		return log.ErrorLevelString
	case log.FatalLevel:
		return log.FatalLevelString
	case log.PanicLevel:
		return log.PanicLevelString
	default:
		return ""
	}
}

func (l Logger) enabled(level log.Level, ok bool) bool {
	return !ok || level >= l.Log.Level
}

func (l Logger) newEntry(level log.Level, ok bool) *log.Entry {
	logger := l.log()
	if !ok {
		return logger.Log()
	}
	if !l.enabled(level, ok) {
		return nil
	}
	if level == log.FatalLevel || level == log.PanicLevel {
		return logger.Log().Str(log.LevelKey, levelString(level))
	}
	return logger.WithLevel(level)
}

// Enabled reports whether the logger would emit a record with param.
func (l Logger) Enabled(_ context.Context, param otellog.EnabledParameters) bool {
	level, ok := logLevel(param.Severity)
	return l.enabled(level, ok)
}

// Emit emits an OpenTelemetry log record to log.
func (l Logger) Emit(ctx context.Context, record otellog.Record) {
	level, ok := logLevel(record.Severity())
	e := l.newEntry(level, ok)
	if e == nil {
		return
	}

	fields := fieldNames(l.FieldNames)
	if !ok {
		if text := record.SeverityText(); text != "" {
			e = e.Str(log.LevelKey, text)
		}
	} else if text := record.SeverityText(); text != "" && fields.SeverityText != "" {
		e = e.Str(fields.SeverityText, text)
	}

	if timestamp := record.Timestamp(); !timestamp.IsZero() && fields.Timestamp != "" {
		e = e.Time(fields.Timestamp, timestamp)
	}
	if observed := record.ObservedTimestamp(); !observed.IsZero() && fields.ObservedTime != "" {
		e = e.Time(fields.ObservedTime, observed)
	}
	if eventName := record.EventName(); eventName != "" && fields.EventName != "" {
		e = e.Str(fields.EventName, eventName)
	}
	if err := record.Err(); err != nil {
		e = e.Err(err)
	}

	if !l.DisableScopeFields {
		if l.scopeName != "" && fields.ScopeName != "" {
			e = e.Str(fields.ScopeName, l.scopeName)
		}
		if l.scopeVersion != "" && fields.ScopeVersion != "" {
			e = e.Str(fields.ScopeVersion, l.scopeVersion)
		}
		if l.scopeSchema != "" && fields.ScopeSchemaURL != "" {
			e = e.Str(fields.ScopeSchemaURL, l.scopeSchema)
		}
		if len(l.scopeAttrs) != 0 && fields.ScopeAttributes != "" {
			e = e.Object(fields.ScopeAttributes, attributeMap(l.scopeAttrs))
		}
	}

	if !l.DisableTraceContext {
		if sc := oteltrace.SpanContextFromContext(ctx); sc.IsValid() {
			if fields.TraceID != "" {
				e = e.Str(fields.TraceID, sc.TraceID().String())
			}
			if fields.SpanID != "" {
				e = e.Str(fields.SpanID, sc.SpanID().String())
			}
			if fields.TraceFlags != "" {
				e = e.Str(fields.TraceFlags, sc.TraceFlags().String())
			}
		}
	}

	if body := record.Body(); !body.Empty() {
		e = appendValue(e, log.MessageKey, body)
	}
	record.WalkAttributes(func(kv otellog.KeyValue) bool {
		e = appendValue(e, kv.Key, kv.Value)
		return true
	})
	e.Msg("")
}

func appendValue(e *log.Entry, key string, value otellog.Value) *log.Entry {
	switch value.Kind() {
	case otellog.KindBool:
		return e.Bool(key, value.AsBool())
	case otellog.KindFloat64:
		return e.Float64(key, value.AsFloat64())
	case otellog.KindInt64:
		return e.Int64(key, value.AsInt64())
	case otellog.KindString:
		return e.Str(key, value.AsString())
	case otellog.KindBytes:
		return e.Bytes(key, value.AsBytes())
	case otellog.KindSlice:
		return appendSlice(e, key, value.AsSlice())
	case otellog.KindMap:
		return e.Object(key, mapValue(value.AsMap()))
	case otellog.KindEmpty:
		fallthrough
	default:
		return e.Any(key, nil)
	}
}

type valueBuffer struct {
	B []byte
}

var valueBufferPool = sync.Pool{
	New: func() any {
		return &valueBuffer{B: make([]byte, 0, 1024)}
	},
}

const valueBufferCap = 1 << 16

func appendSlice(e *log.Entry, key string, values []otellog.Value) *log.Entry {
	if e == nil {
		return nil
	}
	if len(values) == 0 {
		return e.RawJSONStr(key, "[]")
	}

	b := valueBufferPool.Get().(*valueBuffer)
	b.B = appendJSONSlice(b.B[:0], values)
	e = e.RawJSON(key, b.B)
	if cap(b.B) <= valueBufferCap {
		valueBufferPool.Put(b)
	}
	return e
}

func appendJSONValue(b []byte, value otellog.Value) []byte {
	switch value.Kind() {
	case otellog.KindBool:
		return strconv.AppendBool(b, value.AsBool())
	case otellog.KindFloat64:
		return appendJSONFloat(b, value.AsFloat64(), 64)
	case otellog.KindInt64:
		return strconv.AppendInt(b, value.AsInt64(), 10)
	case otellog.KindString:
		return appendJSONString(b, value.AsString())
	case otellog.KindBytes:
		return appendJSONBytes(b, value.AsBytes())
	case otellog.KindSlice:
		return appendJSONSlice(b, value.AsSlice())
	case otellog.KindMap:
		return appendJSONMap(b, value.AsMap())
	default:
		return append(b, "null"...)
	}
}

func appendJSONSlice(b []byte, values []otellog.Value) []byte {
	b = append(b, '[')
	for i, value := range values {
		if i != 0 {
			b = append(b, ',')
		}
		b = appendJSONValue(b, value)
	}
	return append(b, ']')
}

func appendJSONMap(b []byte, values []otellog.KeyValue) []byte {
	b = append(b, '{')
	for i, kv := range values {
		if i != 0 {
			b = append(b, ',')
		}
		b = appendJSONString(b, kv.Key)
		b = append(b, ':')
		b = appendJSONValue(b, kv.Value)
	}
	return append(b, '}')
}

func appendJSONFloat(b []byte, f float64, bits int) []byte {
	abs := math.Abs(f)
	fmt := byte('f')
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}
	b = strconv.AppendFloat(b, f, fmt, -1, bits)
	if fmt == 'e' {
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	return b
}

func appendJSONString(b []byte, s string) []byte {
	b = append(b, '"')
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 0x20 && c != '"' && c != '\\' {
			continue
		}
		b = append(b, s[start:i]...)
		b = appendJSONEscape(b, c)
		start = i + 1
	}
	b = append(b, s[start:]...)
	b = append(b, '"')
	return b
}

func appendJSONBytes(b []byte, values []byte) []byte {
	b = append(b, '"')
	start := 0
	for i, c := range values {
		if c >= 0x20 && c != '"' && c != '\\' {
			continue
		}
		b = append(b, values[start:i]...)
		b = appendJSONEscape(b, c)
		start = i + 1
	}
	b = append(b, values[start:]...)
	b = append(b, '"')
	return b
}

func appendJSONEscape(b []byte, c byte) []byte {
	switch c {
	case '"':
		return append(b, '\\', '"')
	case '\\':
		return append(b, '\\', '\\')
	case '\b':
		return append(b, '\\', 'b')
	case '\f':
		return append(b, '\\', 'f')
	case '\n':
		return append(b, '\\', 'n')
	case '\r':
		return append(b, '\\', 'r')
	case '\t':
		return append(b, '\\', 't')
	default:
		return append(b, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xf])
	}
}

const hex = "0123456789abcdef"

type mapValue []otellog.KeyValue

func (m mapValue) MarshalObject(e *log.Entry) {
	for _, kv := range m {
		e = appendValue(e, kv.Key, kv.Value)
	}
}

type attributeMap []attribute.KeyValue

func (m attributeMap) MarshalObject(e *log.Entry) {
	for _, kv := range m {
		e = e.Any(string(kv.Key), kv.Value.AsInterface())
	}
}

var (
	_ otellog.Logger         = Logger{}
	_ otellog.LoggerProvider = LoggerProvider{}
)
