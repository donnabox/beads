//go:build windows

package doltcli

import (
	"errors"

	"golang.org/x/sys/windows"
)

// processAlive reports whether a process with the given pid is running. Opening
// a process is not enough: an exited process stays openable while another handle
// refers to it, so a zero-timeout wait tells a signaled process from a running
// one. A process owned by another account refuses the open with access denied,
// which still means it exists, as EPERM does on the other platforms.
func processAlive(pid int) bool {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return errors.Is(err, windows.ERROR_ACCESS_DENIED)
	}
	defer func() { _ = windows.CloseHandle(h) }()

	status, err := windows.WaitForSingleObject(h, 0)
	return err == nil && status == uint32(windows.WAIT_TIMEOUT)
}
