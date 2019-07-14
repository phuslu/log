package log

type Level uint8

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
	NoLevel
	// Disabled disables the logger.
	Disabled
)

func ParseLevel(s string) (level Level) {
	switch s {
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
		level = NoLevel
	}
	return
}
