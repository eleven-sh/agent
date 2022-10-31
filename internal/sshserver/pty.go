package sshserver

import (
	"os"
	"syscall"
	"unsafe"
)

func setWindowSize(f *os.File, width, height int) {
	syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct {
			h, w, x, y uint16
		}{
			uint16(height), uint16(width), 0, 0,
		})),
	)
}
