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

// Lower return lowe case string of Level
func (l Level) Lower() (s string) {
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

// Upper return upper case string of Level
func (l Level) Upper() (s string) {
	switch l {
	case TraceLevel:
		s = "TRACE"
	case DebugLevel:
		s = "DEBUG"
	case InfoLevel:
		s = "INFO"
	case WarnLevel:
		s = "WARN"
	case ErrorLevel:
		s = "ERROR"
	case FatalLevel:
		s = "FATAL"
	case PanicLevel:
		s = "PANIC"
	default:
		s = "????"
	}
	return
}

// Title return title case string of Level
func (l Level) Title() (s string) {
	switch l {
	case TraceLevel:
		s = "Trace"
	case DebugLevel:
		s = "Debug"
	case InfoLevel:
		s = "Info"
	case WarnLevel:
		s = "Warn"
	case ErrorLevel:
		s = "Error"
	case FatalLevel:
		s = "Fatal"
	case PanicLevel:
		s = "Panic"
	default:
		s = "????"
	}
	return
}

// Three return three letters of Level
func (l Level) Three() (s string) {
	switch l {
	case TraceLevel:
		s = "TRC"
	case DebugLevel:
		s = "DBG"
	case InfoLevel:
		s = "INF"
	case WarnLevel:
		s = "WRN"
	case ErrorLevel:
		s = "ERR"
	case FatalLevel:
		s = "FTL"
	case PanicLevel:
		s = "PNC"
	default:
		s = "???"
	}
	return
}

// First return first upper letter of Level
func (l Level) First() (s string) {
	switch l {
	case TraceLevel:
		s = "T"
	case DebugLevel:
		s = "D"
	case InfoLevel:
		s = "I"
	case WarnLevel:
		s = "W"
	case ErrorLevel:
		s = "E"
	case FatalLevel:
		s = "F"
	case PanicLevel:
		s = "P"
	default:
		s = "?"
	}
	return
}

// ParseLevel converts a level string into a log Level value.
func ParseLevel(s string) (level Level) {
	switch s {
	case "trace", "Trace", "TRACE", "T", "TRC":
		level = TraceLevel
	case "debug", "Debug", "DEBUG", "D", "DBG":
		level = DebugLevel
	case "info", "Info", "INFO", "I", "INF":
		level = InfoLevel
	case "warn", "Warn", "WARN", "warning", "Warning", "WARNING", "W", "WRN":
		level = WarnLevel
	case "error", "Error", "ERROR", "E", "ERR":
		level = ErrorLevel
	case "fatal", "Fatal", "FATAL", "F", "FTL":
		level = FatalLevel
	case "panic", "Panic", "PANIC", "P", "PNC":
		level = PanicLevel
	default:
		level = noLevel
	}
	return
}

// ParseLevelByte converts a level byte into a log Level value.
func ParseLevelByte(b byte) (level Level) {
	switch b {
	case 't', 'T':
		level = TraceLevel
	case 'd', 'D':
		level = DebugLevel
	case 'i', 'I':
		level = InfoLevel
	case 'w', 'W':
		level = WarnLevel
	case 'e', 'E':
		level = ErrorLevel
	case 'f', 'F':
		level = FatalLevel
	case 'p', 'P':
		level = PanicLevel
	default:
		level = noLevel
	}
	return
}
