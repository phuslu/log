// +build windows

package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

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
		windowsColorBlue   = 1
		windowsColorGreen  = 2
		windowsColorAqua   = 3
		windowsColorRed    = 4
		windowsColorPurple = 5
		windowsColorYellow = 6
		windowsColorWhite  = 7
		windowsColorGray   = 8
	)

	var m map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(p))
	decoder.UseNumber()
	err = decoder.Decode(&m)
	if err != nil {
		n, err = os.Stderr.Write(p)
		return
	}

	var printf = func(color uintptr, format string, args ...interface{}) {
		if color != windowsColorWhite {
			setConsoleTextAttribute(uintptr(syscall.Stderr), color)
		}
		var i int
		i, err = fmt.Fprintf(os.Stderr, format, args...)
		n += i
		if color != windowsColorWhite {
			setConsoleTextAttribute(uintptr(syscall.Stderr), windowsColorWhite)
		}
	}

	if v, ok := m["time"]; ok {
		if w.ANSIColor {
			printf(windowsColorGray, "%s ", v)
		} else {
			printf(windowsColorWhite, "%s ", v)
		}
	}

	if v, ok := m["level"]; ok {
		var s string
		var c uintptr
		switch s, _ = v.(string); ParseLevel(s) {
		case DebugLevel:
			c, s = windowsColorYellow, "DBG"
		case InfoLevel:
			c, s = windowsColorGreen, "INF"
		case WarnLevel:
			c, s = windowsColorRed, "WRN"
		case ErrorLevel:
			c, s = windowsColorRed, "ERR"
		case FatalLevel:
			c, s = windowsColorRed, "FTL"
		case PanicLevel:
			c, s = windowsColorRed, "PNC"
		default:
			c, s = windowsColorRed, "???"
		}
		if w.ANSIColor {
			printf(c, "%s ", s)
		} else {
			printf(windowsColorWhite, "%s ", s)
		}
	}

	if v, ok := m["goid"]; ok {
		printf(windowsColorWhite, "%s ", v)
	}

	if v, ok := m["caller"]; ok {
		printf(windowsColorWhite, "%s ", v)
	}

	if v, ok := m["message"]; ok {
		if s, _ := v.(string); s != "" && s[len(s)-1] == '\n' {
			v = s[:len(s)-1]
		}
		if w.ANSIColor {
			printf(windowsColorAqua, ">")
		} else {
			printf(windowsColorWhite, ">")
		}
		printf(windowsColorWhite, " %s", v)
	}

	for k, v := range m {
		switch k {
		case "time", "level", "goid", "caller", "message":
			continue
		}
		if w.ANSIColor {
			if k == "error" && v != nil {
				printf(windowsColorRed, " %s=%v", k, v)
			} else {
				printf(windowsColorAqua, " %s=", k)
				printf(windowsColorGray, "%v", v)
			}
		} else {
			printf(windowsColorWhite, " %s=%v", k, v)
		}
	}

	printf(windowsColorWhite, " \n")

	return n, err
}

func IsTerminal(file *os.File) bool {
	var mode uint32
	err := syscall.GetConsoleMode(syscall.Handle(file.Fd()), &mode)
	if err != nil {
		return false
	}

	return true
}
