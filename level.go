package log

// Level defines log levels.
type Level uint32

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	noLevel
)

// Lower return lowe case string of Level
func (l Level) Lower() (s string) {
	switch l {
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
// returns an error if the input string does not match known values.
func ParseLevel(s string) (level Level) {
	switch s {
	case "debug", "Debug", "DEBUG", "D", "DBG", "DEBU":
		level = DebugLevel
	case "info", "Info", "INFO", "I", "INF":
		level = InfoLevel
	case "warn", "Warn", "WARN", "warning", "Warning", "WARNING", "W", "WRN":
		level = WarnLevel
	case "error", "Error", "ERROR", "E", "ERR", "ERRO":
		level = ErrorLevel
	case "fatal", "Fatal", "FATAL", "F", "FTL", "FATA":
		level = FatalLevel
	case "panic", "Panic", "PANIC", "P", "PNC", "PANI":
		level = PanicLevel
	default:
		level = noLevel
	}
	return
}
