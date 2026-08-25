//go:build windows

package collaboration

import "syscall"

func setBroadcast(fd uintptr) {
	syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
}
