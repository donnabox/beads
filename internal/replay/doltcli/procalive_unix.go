//go:build !windows

package doltcli

import (
	"errors"
	"os"
	"syscall"
)

// processAlive reports whether a process with the given pid exists. Signal 0
// delivers nothing and only checks; a process owned by another user answers
// EPERM, which still means it exists.
func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}
