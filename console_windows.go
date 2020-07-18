// +build windows

package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

func isTerminal(fd uintptr, _, _ string) bool {
	var mode uint32
	err := syscall.GetConsoleMode(syscall.Handle(fd), &mode)
	if err != nil {
		return false
	}

	return true
}

var (
	vtInited  uint32
	vtEnabled bool
	muConsole sync.Mutex
)

func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	// try init windows 10 virtual terminal
	if atomic.LoadUint32(&vtInited) == 0 {
		muConsole.Lock()
		if atomic.LoadUint32(&vtInited) == 0 {
			if tryEnableVirtualTerminalProcessing() == nil {
				vtEnabled = true
			}
			atomic.StoreUint32(&vtInited, 1)
		}
		muConsole.Unlock()
	}
	// write
	if vtEnabled {
		n, err = w.write(p)
	} else {
		n, err = w.writeWindows(p)
	}
	return
}

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode          = kernel32.NewProc("SetConsoleMode").Call
	setConsoleTextAttribute = kernel32.NewProc("SetConsoleTextAttribute").Call
)

func tryEnableVirtualTerminalProcessing() error {
	var h syscall.Handle
	var b [64]uint16
	var n uint32

	// open registry
	err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, syscall.StringToUTF16Ptr(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`), 0, syscall.KEY_READ, &h)
	if err != nil {
		return err
	}
	defer syscall.RegCloseKey(h)

	// read windows build number
	n = uint32(len(b))
	err = syscall.RegQueryValueEx(h, syscall.StringToUTF16Ptr(`CurrentBuild`), nil, nil, (*byte)(unsafe.Pointer(&b[0])), &n)
	if err != nil {
		return err
	}
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			break
		}
		n = n*10 + uint32(b[i]-'0')
	}

	// return if lower than windows 10 16257
	if n < 16257 {
		return errors.New("not implemented")
	}

	// get console mode
	err = syscall.GetConsoleMode(syscall.Stderr, &n)
	if err != nil {
		return err
	}

	// enable ENABLE_VIRTUAL_TERMINAL_PROCESSING
	ret, _, err := setConsoleMode(uintptr(syscall.Stderr), uintptr(n|0x4))
	if ret == 0 {
		return err
	}

	return nil
}

func (w *ConsoleWriter) writeWindows(p []byte) (n int, err error) {
	muConsole.Lock()
	defer muConsole.Unlock()

	const (
		Blue   = 1
		Green  = 2
		Aqua   = 3
		Red    = 4
		Purple = 5
		Yellow = 6
		White  = 7
		Gray   = 8
	)

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = os.Stderr.Write(p)
		return
	}

	var cprintf = func(color uintptr, format string, args ...interface{}) {
		if color != White {
			setConsoleTextAttribute(uintptr(syscall.Stderr), color)
			defer setConsoleTextAttribute(uintptr(syscall.Stderr), White)
		}
		var i int
		i, err = fmt.Fprintf(os.Stderr, format, args...)
		n += i
	}

	var timeField = w.TimeField
	if timeField == "" {
		timeField = "time"
	}
	if v, ok := m[timeField]; ok {
		if w.ColorOutput || w.ANSIColor {
			cprintf(Gray, "%s ", v)
		} else {
			cprintf(White, "%s ", v)
		}
	}

	if v, ok := m["level"]; ok {
		var s string
		var c uintptr
		switch s, _ = v.(string); ParseLevel(s) {
		case DebugLevel:
			c, s = Yellow, "DBG"
		case InfoLevel:
			c, s = Green, "INF"
		case WarnLevel:
			c, s = Red, "WRN"
		case ErrorLevel:
			c, s = Red, "ERR"
		case FatalLevel:
			c, s = Red, "FTL"
		default:
			c, s = Red, "???"
		}
		if w.ColorOutput || w.ANSIColor {
			cprintf(c, "%s ", s)
		} else {
			cprintf(White, "%s ", s)
		}
	}

	if v, ok := m["caller"]; ok {
		cprintf(White, "%s ", v)
	}

	var msgField = "message"
	if _, ok := m[msgField]; !ok {
		if _, ok := m["msg"]; ok {
			msgField = "msg"
		}
	}

	if v, ok := m[msgField]; ok && !w.EndWithMessage {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ColorOutput || w.ANSIColor {
			cprintf(Aqua, ">")
		} else {
			cprintf(White, ">")
		}
		cprintf(White, " %s", v)
	} else {
		if w.ColorOutput || w.ANSIColor {
			cprintf(Aqua, ">")
		} else {
			cprintf(White, ">")
		}
	}

	for _, k := range jsonKeys(p) {
		switch k {
		case timeField, msgField, "level", "caller", "stack":
			continue
		}
		v := m[k]
		if w.QuoteString {
			if s, ok := v.(string); ok {
				v = strconv.Quote(s)
			}
		}
		if w.ColorOutput || w.ANSIColor {
			if k == "error" && v != nil {
				cprintf(Red, " %s=%v", k, v)
			} else {
				cprintf(Aqua, " %s=", k)
				cprintf(Gray, "%v", v)
			}
		} else {
			cprintf(White, " %s=%v", k, v)
		}
	}

	if w.EndWithMessage {
		if v, ok := m[msgField]; ok {
			if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
				v = s[:len(s)-1]
			}
			cprintf(White, " %s", v)
		}
	}

	if v, ok := m["stack"]; ok {
		if s, ok := v.(string); ok {
			cprintf(White, "\n%s", s)
		} else {
			data, _ := json.MarshalIndent(v, "", "  ")
			cprintf(White, "\n%s", data)
		}
	}

	cprintf(White, " \n")

	return n, err
}
