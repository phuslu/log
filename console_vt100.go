// +build !windows

package log

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

func (w *ConsoleWriter) Write(p []byte) (int, error) {
	return w.write(p)
}

func IsTerminal(file *os.File) bool {
	var control uintptr
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "ppc64", "ppc64le":
			control = 0x402c7413
		case "mips", "mipsle", "mips64", "mips64le":
			control = 0x540d
		default:
			control = 0x5401
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64", "arm64":
			control = 0x40487413
		default:
			control = 0x402c7413
		}
	default:
		control = 0x402c7413
	}

	var termios [256]byte
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, file.Fd(), control, uintptr(unsafe.Pointer(&termios[0])), 0, 0, 0)
	return err == 0
}
