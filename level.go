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
