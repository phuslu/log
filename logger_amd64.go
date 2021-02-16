// +build amd64

package log

func (l *Logger) shouldnot(level Level) bool {
	return level < l.Level
}
