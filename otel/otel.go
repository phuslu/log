package otel

import (
	"context"

	phuslog "github.com/phuslu/log"
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

type config struct {
	fields       FieldNames
	traceContext bool
	scopeFields  bool
}

// Option configures the OpenTelemetry-to-phuslog adapter.
type Option func(*config)

// Logger bridges OpenTelemetry log records into a phuslog Logger.
type Logger struct {
	embedded.Logger

	logger       phuslog.Logger
	config       config
	scopeName    string
	scopeVersion string
	scopeSchema  string
	scopeAttrs   []attribute.KeyValue
}

// LoggerProvider creates OpenTelemetry loggers backed by a phuslog Logger.
type LoggerProvider struct {
	embedded.LoggerProvider

	logger phuslog.Logger
	config config
}

func defaultConfig() config {
	return config{
		fields: FieldNames{
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
		},
		traceContext: true,
		scopeFields:  true,
	}
}

func applyOptions(options []Option) config {
	c := defaultConfig()
	for _, option := range options {
		if option != nil {
			option(&c)
		}
	}
	return c
}

func loggerValue(logger *phuslog.Logger) phuslog.Logger {
	if logger == nil {
		return phuslog.DefaultLogger
	}
	l := *logger
	// OpenTelemetry records do not carry a program counter. Keeping Caller on
	// would report this adapter instead of the application call site.
	l.Caller = 0
	return l
}

// WithFieldNames overrides non-empty OpenTelemetry metadata field names.
func WithFieldNames(names FieldNames) Option {
	return func(c *config) {
		if names.Timestamp != "" {
			c.fields.Timestamp = names.Timestamp
		}
		if names.ObservedTime != "" {
			c.fields.ObservedTime = names.ObservedTime
		}
		if names.EventName != "" {
			c.fields.EventName = names.EventName
		}
		if names.SeverityText != "" {
			c.fields.SeverityText = names.SeverityText
		}
		if names.TraceID != "" {
			c.fields.TraceID = names.TraceID
		}
		if names.SpanID != "" {
			c.fields.SpanID = names.SpanID
		}
		if names.TraceFlags != "" {
			c.fields.TraceFlags = names.TraceFlags
		}
		if names.ScopeName != "" {
			c.fields.ScopeName = names.ScopeName
		}
		if names.ScopeVersion != "" {
			c.fields.ScopeVersion = names.ScopeVersion
		}
		if names.ScopeSchemaURL != "" {
			c.fields.ScopeSchemaURL = names.ScopeSchemaURL
		}
		if names.ScopeAttributes != "" {
			c.fields.ScopeAttributes = names.ScopeAttributes
		}
	}
}

// WithoutTraceContext disables trace_id, span_id, and trace_flags extraction.
func WithoutTraceContext() Option {
	return func(c *config) {
		c.traceContext = false
	}
}

// WithoutScopeFields disables instrumentation scope fields from LoggerProvider.
func WithoutScopeFields() Option {
	return func(c *config) {
		c.scopeFields = false
	}
}

// NewLogger returns an OpenTelemetry Logs API logger backed by logger.
func NewLogger(logger *phuslog.Logger, options ...Option) *Logger {
	return &Logger{
		logger: loggerValue(logger),
		config: applyOptions(options),
	}
}

// NewLoggerProvider returns an OpenTelemetry Logs API provider backed by logger.
func NewLoggerProvider(logger *phuslog.Logger, options ...Option) *LoggerProvider {
	return &LoggerProvider{
		logger: loggerValue(logger),
		config: applyOptions(options),
	}
}

// Logger returns a scoped OpenTelemetry logger.
func (p *LoggerProvider) Logger(name string, options ...otellog.LoggerOption) otellog.Logger {
	cfg := otellog.NewLoggerConfig(options...)
	attrs := cfg.InstrumentationAttributes()
	return &Logger{
		logger:       p.logger,
		config:       p.config,
		scopeName:    name,
		scopeVersion: cfg.InstrumentationVersion(),
		scopeSchema:  cfg.SchemaURL(),
		scopeAttrs:   attrs.ToSlice(),
	}
}

func phuslogLevel(severity otellog.Severity) (phuslog.Level, bool) {
	switch {
	case severity >= otellog.SeverityFatal1:
		return phuslog.FatalLevel, true
	case severity >= otellog.SeverityError1:
		return phuslog.ErrorLevel, true
	case severity >= otellog.SeverityWarn1:
		return phuslog.WarnLevel, true
	case severity >= otellog.SeverityInfo1:
		return phuslog.InfoLevel, true
	case severity >= otellog.SeverityDebug1:
		return phuslog.DebugLevel, true
	case severity >= otellog.SeverityTrace1:
		return phuslog.TraceLevel, true
	default:
		return 0, false
	}
}

func levelString(level phuslog.Level) string {
	switch level {
	case phuslog.TraceLevel:
		return phuslog.TraceLevelString
	case phuslog.DebugLevel:
		return phuslog.DebugLevelString
	case phuslog.InfoLevel:
		return phuslog.InfoLevelString
	case phuslog.WarnLevel:
		return phuslog.WarnLevelString
	case phuslog.ErrorLevel:
		return phuslog.ErrorLevelString
	case phuslog.FatalLevel:
		return phuslog.FatalLevelString
	case phuslog.PanicLevel:
		return phuslog.PanicLevelString
	default:
		return ""
	}
}

func (l *Logger) enabled(level phuslog.Level, ok bool) bool {
	return !ok || level >= l.logger.Level
}

func (l *Logger) newEntry(level phuslog.Level, ok bool) *phuslog.Entry {
	if !ok {
		return l.logger.Log()
	}
	if !l.enabled(level, ok) {
		return nil
	}
	if level == phuslog.FatalLevel || level == phuslog.PanicLevel {
		return l.logger.Log().Str(phuslog.LevelKey, levelString(level))
	}
	return l.logger.WithLevel(level)
}

// Enabled reports whether the logger would emit a record with param.
func (l *Logger) Enabled(_ context.Context, param otellog.EnabledParameters) bool {
	level, ok := phuslogLevel(param.Severity)
	return l.enabled(level, ok)
}

// Emit emits an OpenTelemetry log record to phuslog.
func (l *Logger) Emit(ctx context.Context, record otellog.Record) {
	level, ok := phuslogLevel(record.Severity())
	e := l.newEntry(level, ok)
	if e == nil {
		return
	}

	fields := l.config.fields
	if !ok {
		if text := record.SeverityText(); text != "" {
			e = e.Str(phuslog.LevelKey, text)
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

	if l.config.scopeFields {
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

	if l.config.traceContext {
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
		e = appendValue(e, phuslog.MessageKey, body)
	}
	record.WalkAttributes(func(kv otellog.KeyValue) bool {
		e = appendValue(e, kv.Key, kv.Value)
		return true
	})
	e.Msg("")
}

func appendValue(e *phuslog.Entry, key string, value otellog.Value) *phuslog.Entry {
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
		return e.Any(key, sliceValue(value.AsSlice()))
	case otellog.KindMap:
		return e.Object(key, mapValue(value.AsMap()))
	case otellog.KindEmpty:
		fallthrough
	default:
		return e.Any(key, nil)
	}
}

func sliceValue(values []otellog.Value) []any {
	out := make([]any, len(values))
	for i, value := range values {
		out[i] = anyValue(value)
	}
	return out
}

func anyValue(value otellog.Value) any {
	switch value.Kind() {
	case otellog.KindBool:
		return value.AsBool()
	case otellog.KindFloat64:
		return value.AsFloat64()
	case otellog.KindInt64:
		return value.AsInt64()
	case otellog.KindString:
		return value.AsString()
	case otellog.KindBytes:
		return value.AsBytes()
	case otellog.KindSlice:
		return sliceValue(value.AsSlice())
	case otellog.KindMap:
		m := make(map[string]any, len(value.AsMap()))
		for _, kv := range value.AsMap() {
			m[kv.Key] = anyValue(kv.Value)
		}
		return m
	default:
		return nil
	}
}

type mapValue []otellog.KeyValue

func (m mapValue) MarshalObject(e *phuslog.Entry) {
	for _, kv := range m {
		e = appendValue(e, kv.Key, kv.Value)
	}
}

type attributeMap []attribute.KeyValue

func (m attributeMap) MarshalObject(e *phuslog.Entry) {
	for _, kv := range m {
		e = e.Any(string(kv.Key), kv.Value.AsInterface())
	}
}

var (
	_ otellog.Logger         = (*Logger)(nil)
	_ otellog.LoggerProvider = (*LoggerProvider)(nil)
)
