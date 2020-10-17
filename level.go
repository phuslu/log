package log

// Level defines log levels.
type Level int32

const (
	// TraceLevel defines trace log level.
	TraceLevel Level = -1
	// DebugLevel defines debug log level.
	DebugLevel Level = 0
	// InfoLevel defines info log level.
	InfoLevel Level = 1
	// WarnLevel defines warn log level.
	WarnLevel Level = 2
	// ErrorLevel defines error log level.
	ErrorLevel Level = 3
	// FatalLevel defines fatal log level.
	FatalLevel Level = 4
	// PanicLevel defines panic log level.
	PanicLevel Level = 5
	// NoLevel defines an absent log level.
	noLevel Level = 6
)

// String return lowe case string of Level
func (l Level) String() (s string) {
	switch l {
	case TraceLevel:
		s = "trace"
	case DebugLevel:
		s = "debug"
	case InfoLevel:
		s = "info"
	case WarnLevel:
		s = "warn"
	case ErrorLevel:
		s = "error"
	case FatalLevel:
		s = "fatal"
	case PanicLevel:
		s = "panic"
	default:
		s = "????"
	}
	return
}

// ParseLevel converts a level string into a log Level value.
func ParseLevel(s string) (level Level) {
	switch s {
	case "trace", "Trace", "TRACE", "TRC":
		level = TraceLevel
	case "debug", "Debug", "DEBUG", "DBG":
		level = DebugLevel
	case "info", "Info", "INFO", "INF":
		level = InfoLevel
	case "warn", "Warn", "WARN", "warning", "Warning", "WARNING", "WRN":
		level = WarnLevel
	case "error", "Error", "ERROR", "ERR":
		level = ErrorLevel
	case "fatal", "Fatal", "FATAL", "FTL":
		level = FatalLevel
	case "panic", "Panic", "PANIC", "PNC":
		level = PanicLevel
	default:
		level = noLevel
	}
	return
}
