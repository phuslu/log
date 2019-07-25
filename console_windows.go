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
		n, err = w.vtWrite(p)
	} else {
		n, err = w.leacyWrite(p)
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
		n, err = os.Stderr.Write(p)
		return
	}

	var printf = func(color uintptr, format string, args ...interface{}) {
		if color != winColorWhite {
			setConsoleTextAttribute(uintptr(syscall.Stderr), color)
		}
		var i int
		i, err = fmt.Fprintf(os.Stderr, format, args...)
		n += i
		if color != winColorWhite {
			setConsoleTextAttribute(uintptr(syscall.Stderr), winColorWhite)
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
