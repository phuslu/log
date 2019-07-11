package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

var timeNow = time.Now

var _ io.WriteCloser = (*Writer)(nil)

var hostname, _ = os.Hostname()

type Writer struct {
	Filename   string
	MaxSize    int64
	MaxBackups int
	LocalTime  bool
	HostName   bool

	mu   sync.Mutex
	size int64
	file *os.File
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		err = w.rotate(false)
		if err != nil {
			return
		}
	}

	n, err = w.file.Write(p)

	w.size += int64(n)
	if w.MaxSize > 0 && w.size > w.MaxSize {
		w.rotate(true)
	}

	return
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.file.Close()
	w.file = nil
	w.size = 0
	return err
}

func (w *Writer) Rotate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.rotate(true)
}

func (w *Writer) rotate(newFile bool) (err error) {
	if w.Filename == "" {
		w.file = os.Stderr
		return nil
	}

	if w.file != nil {
		err = w.file.Close()
		if err != nil {
			return
		}
	}

	now := timeNow()
	if !w.LocalTime {
		now = now.UTC()
	}

	ext := filepath.Ext(w.Filename)
	prefix := w.Filename[0 : len(w.Filename)-len(ext)]
	filename := prefix + now.Format(".2006-01-02T15-04-05")
	if w.HostName {
		filename += "." + hostname + ext
	} else {
		filename += ext
	}
	w.size = 0

	if !newFile {
		if link, err := os.Readlink(w.Filename); err == nil {
			if fi, err := os.Stat(link); err == nil {
				filename = link
				w.size = fi.Size()
			}
		}
	}

	w.file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil && w.file == nil {
		return
	}

	go func(filename string) {
		os.Remove(w.Filename)
		err := os.Symlink(filename, w.Filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "symlink %+v to %+v error: %+v", filename, w.Filename, err)
			return
		}

		switch runtime.GOOS {
		case "linux":
			uid, _ := strconv.Atoi(os.Getenv("SUDO_UID"))
			gid, _ := strconv.Atoi(os.Getenv("SUDO_GID"))
			if uid != 0 && gid != 0 && os.Geteuid() == 0 {
				os.Lchown(w.Filename, uid, gid)
				os.Chown(filename, uid, gid)
			}
		}

		var matches []string
		matches, err = filepath.Glob(prefix + ".20*" + ext)
		if err != nil {
			return
		}

		sort.Strings(matches)
		for i := 0; i < len(matches)-w.MaxBackups-1; i++ {
			os.Remove(matches[i])
		}
	}(filename)

	return
}

const (
	DefaultBufferSize    = 32 * 1024
	DefaultFlushDuration = 5 * time.Second
)

var _ io.WriteCloser = (*BufferWriter)(nil)

type BufferWriter struct {
	BufferSize    int
	FlushDuration time.Duration
	*Writer

	mu sync.Mutex
	bw *bufio.Writer
}

func (w *BufferWriter) Flush() (err error) {
	w.mu.Lock()
	err = w.bw.Flush()
	w.mu.Unlock()
	return
}

func (w *BufferWriter) Close() error {
	w.mu.Lock()
	w.bw.Flush()
	w.mu.Unlock()
	return w.Writer.Close()
}

func (w *BufferWriter) Write(p []byte) (n int, err error) {
	if w.bw == nil {
		w.Writer.mu.Lock()
		if w.BufferSize == 0 {
			w.BufferSize = DefaultBufferSize
		}
		if w.FlushDuration == 0 {
			w.FlushDuration = DefaultFlushDuration
		}
		if w.bw == nil {
			w.bw = bufio.NewWriterSize(w.Writer, w.BufferSize)
		}
		go func() {
			for {
				time.Sleep(w.FlushDuration)
				w.Flush()
			}
		}()
		w.Writer.mu.Unlock()
	}

	return w.bw.Write(p)
}

type ConsoleWriter struct {
	ANSIColor bool
}

func (w *ConsoleWriter) Write(p []byte) (written int, err error) {
	var m map[string]interface{}

	err = json.Unmarshal(p, &m)
	if err != nil {
		return
	}

	var n int

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			n, err = fmt.Fprintf(os.Stderr, "%s%s%s ", colorDarkGray, v, colorReset)
		} else {
			n, err = fmt.Fprintf(os.Stderr, "%s ", v)
		}
		written += n
	}

	if v, ok := m["level"]; ok {
		var s string
		var c color
		switch s, _ = v.(string); ParseLevel(s) {
		case DebugLevel:
			c, s = colorYellow, "DBG"
		case InfoLevel:
			c, s = colorGreen, "INF"
		case WarnLevel:
			c, s = colorRed, "WRN"
		case ErrorLevel:
			c, s = colorRed, "ERR"
		case FatalLevel:
			c, s = colorRed, "FTL"
		default:
			c, s = colorRed, "???"
		}
		if w.ANSIColor {
			n, err = fmt.Fprintf(os.Stderr, "%s%s%s ", c, s, colorReset)
		} else {
			n, err = fmt.Fprintf(os.Stderr, "%s ", s)
		}
		written += n
	}

	if v, ok := m["caller"]; ok {
		n, err = fmt.Fprintf(os.Stderr, "%s ", v)
		written += n
	}

	if v, ok := m["message"]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ANSIColor {
			n, err = fmt.Fprintf(os.Stderr, "%s>%s %s", colorCyan, colorReset, v)
		} else {
			n, err = fmt.Fprintf(os.Stderr, "> %s", v)
		}
		written += n
	}

	for k, v := range m {
		switch k {
		case "time", "level", "caller", "message":
			continue
		}
		if w.ANSIColor {
			switch k {
			case "error":
				n, err = fmt.Fprintf(os.Stderr, " %s%s=%v%s", colorRed, k, v, colorReset)
			default:
				n, err = fmt.Fprintf(os.Stderr, " %s%s=%s%v", colorCyan, k, colorReset, v)
			}
		} else {
			n, err = fmt.Fprintf(os.Stderr, " %s=%v", k, v)
		}
		written += n
	}

	n, err = os.Stderr.Write([]byte{'\n'})
	written += n

	return
}
