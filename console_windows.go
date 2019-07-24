// +build windows

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

var (
	muConsole               sync.Mutex
	onceConsole             sync.Once
	vtEnabled               = false
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode          = kernel32.NewProc("SetConsoleMode").Call
	setConsoleTextAttribute = kernel32.NewProc("SetConsoleTextAttribute").Call
)

func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	onceConsole.Do(tryEnableVirtualTerminalProcessing)
	if vtEnabled {
		n, err = w.ansiWrite(p)
	} else {
		n, err = w.leacyWrite(p)
	}
	return
}

func (w *ConsoleWriter) leacyWrite(p []byte) (n int, err error) {
	muConsole.Lock()
	defer muConsole.Unlock()

	const (
		winColorBlue   = 1
		winColorGreen  = 2
		winColorAqua   = 3
		winColorRed    = 4
		winColorPurple = 5
		winColorYellow = 6
		winColorWhite  = 7
		winColorGray   = 8
	)

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = os.Stdout.Write(p)
		return
	}

	var printf = func(color uintptr, format string, args ...interface{}) {
		if color != winColorWhite {
			setConsoleTextAttribute(uintptr(syscall.Stdout), color)
		}
		var i int
		i, err = fmt.Fprintf(os.Stdout, format, args...)
		n += i
		if color != winColorWhite {
			setConsoleTextAttribute(uintptr(syscall.Stdout), winColorWhite)
		}
	}

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			printf(winColorGray, "%s ", v)
		} else {
			printf(winColorWhite, "%s ", v)
		}
	}

	if v, ok := m["level"]; ok {
		var s string
		var c uintptr
		switch s, _ = v.(string); ParseLevel(s) {
		case DebugLevel:
			c, s = winColorYellow, "DBG"
		case InfoLevel:
			c, s = winColorGreen, "INF"
		case WarnLevel:
			c, s = winColorRed, "WRN"
		case ErrorLevel:
			c, s = winColorRed, "ERR"
		case FatalLevel:
			c, s = winColorRed, "FTL"
		case PanicLevel:
			c, s = winColorRed, "PNC"
		default:
			c, s = winColorRed, "???"
		}
		if w.ANSIColor {
			printf(c, "%s ", s)
		} else {
			printf(winColorWhite, "%s ", s)
		}
	}

	if v, ok := m["goid"]; ok {
		printf(winColorWhite, "%s ", v)
	}

	if v, ok := m["caller"]; ok {
		printf(winColorWhite, "%s ", v)
	}

	if v, ok := m["message"]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ANSIColor {
			printf(winColorAqua, ">")
		} else {
			printf(winColorWhite, ">")
		}
		printf(winColorWhite, " %s", v)
	}

	for k, v := range m {
		switch k {
		case "time", "level", "goid", "caller", "message":
			continue
		}
		if w.ANSIColor {
			if k == "error" && v != nil {
				printf(winColorRed, " %s=%v", k, v)
			} else {
				printf(winColorAqua, " %s=", k)
				printf(winColorGray, "%v", v)
			}
		} else {
			printf(winColorWhite, " %s=%v", k, v)
		}
	}

	printf(winColorWhite, " \n")

	return n, err
}

func tryEnableVirtualTerminalProcessing() {
	var handle syscall.Handle

	err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, syscall.StringToUTF16Ptr(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`), 0, syscall.KEY_READ, &handle)
	if err != nil {
		return
	}
	defer syscall.RegCloseKey(handle)

	var t, n uint32
	var b [64]uint16

	n = uint32(len(b))
	err = syscall.RegQueryValueEx(handle, syscall.StringToUTF16Ptr(`CurrentBuild`), nil, &t, (*byte)(unsafe.Pointer(&b[0])), &n)
	if err != nil {
		return
	}

	var ver uint32
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			break
		}
		ver = ver*10 + uint32(b[i]-'0')
	}

	if ver < 16257 {
		return
	}

	const ENABLE_VIRTUAL_TERMINAL_PROCESSING uint32 = 0x4

	var mode uint32
	err = syscall.GetConsoleMode(syscall.Stdout, &mode)
	if err != nil {
		return
	}
	mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING
	ret, _, err := setConsoleMode(uintptr(syscall.Stdout), uintptr(mode))
	if ret == 0 {
		return
	}

	vtEnabled = true
	return
}
