package log

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
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
	}
	return
}

type color string

const (
	colorReset    color = "\x1b[0m"
	colorRed      color = "\x1b[31m"
	colorGreen    color = "\x1b[32m"
	colorYellow   color = "\x1b[33m"
	colorCyan     color = "\x1b[36m"
	colorDarkGray color = "\x1b[90m"
)
