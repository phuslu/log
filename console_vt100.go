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
	var req uintptr
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "ppc64", "ppc64le":
			req = 0x402c7413
		case "mips", "mipsle", "mips64", "mips64le":
			req = 0x540d
		default:
			req = 0x5401
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64", "arm64":
			req = 0x40487413
		default:
			req = 0x402c7413
		}
	default:
		req = 0x402c7413
	}

	var termios [256]byte
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, file.Fd(), req, uintptr(unsafe.Pointer(&termios[0])), 0, 0, 0)
	return err == 0
}
