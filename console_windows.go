// +build windows

package log

import (
	"errors"
	"fmt"
	"io"
	"os"
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
	n, err = w.writeWindows(os.Stderr, p)
	// if vtEnabled {
	// 	n, err = w.writeTo(os.Stderr, p)
	// } else {
	// 	n, err = w.writeWindows(os.Stderr, p)
	// }
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

func (w *ConsoleWriter) writeWindows(out io.Writer, p []byte) (n int, err error) {
	muConsole.Lock()
	defer muConsole.Unlock()

	b := bbpool.Get().(*bb)
	b.Reset()
	defer bbpool.Put(b)

	n, err = w.writeTo(b, p)

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
	var _ = func(color uintptr, format string, args ...interface{}) {
		if color != White {
			setConsoleTextAttribute(uintptr(syscall.Stderr), color)
			defer setConsoleTextAttribute(uintptr(syscall.Stderr), White)
		}
		var i int
		i, err = fmt.Fprintf(os.Stderr, format, args...)
		n += i
	}

	n, err = out.Write(b.B)

	return n, err
}
