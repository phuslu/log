package log

// Level defines log levels.
type Level uint32

const (
	_ Level = iota
	// DebugLevel defines debug log level.
	DebugLevel
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel

	noLevel
)

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
	default:
		s = "????"
	}
	return
}

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
	default:
		s = "????"
	}
	return
}

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
	default:
		s = "????"
	}
	return
}

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
	default:
		s = "????"
	}
	return
}

func (l Level) One() (s string) {
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
	}
	return
}
