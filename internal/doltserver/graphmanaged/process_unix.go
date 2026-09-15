//go:build darwin || linux

package graphmanaged

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func platformSupported() bool { return true }
func ownedFile(info os.FileInfo) bool {
	st, ok := info.Sys().(*syscall.Stat_t)
	return ok && int(st.Uid) == os.Geteuid()
}
func trustedDirectory(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0022 != 0 || !ownedFile(info) {
		return nil, errInput
	}
	return info, nil
}
func openKind(path string, directory bool) (*os.File, error) {
	flags := syscall.O_RDONLY | syscall.O_CLOEXEC | syscall.O_NOFOLLOW | syscall.O_NONBLOCK
	if directory {
		flags |= syscall.O_DIRECTORY
	}
	fd, err := syscall.Open(path, flags, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	info, err := f.Stat()
	if err != nil || directory && !info.IsDir() || !directory && !info.Mode().IsRegular() {
		return nil, errors.Join(err, errInput, f.Close())
	}
	return f, nil
}
func openRegular(path string) (*os.File, error)   { return openKind(path, false) }
func openDirectory(path string) (*os.File, error) { return openKind(path, true) }
func signalTerm(process *os.Process) error        { return process.Signal(syscall.SIGTERM) }

func spawn(a admitted, args []string) (cmd *exec.Cmd, p processPipes, result error) {
	owned := []*os.File{}
	defer func() {
		for _, f := range owned {
			result = errors.Join(result, f.Close())
		}
		if result != nil && p.listener != nil {
			result = errors.Join(result, p.listener.Close())
		}
	}()
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		return nil, p, err
	}
	p.listener = listener
	listenerFile, err := listener.File()
	if err != nil {
		return nil, p, err
	}
	owned = append(owned, listenerFile)
	null, err := os.Open(os.DevNull)
	if err != nil {
		return nil, p, err
	}
	owned = append(owned, null)
	controlRead, controlWrite, err := os.Pipe()
	if err != nil {
		return nil, p, err
	}
	owned = append(owned, controlRead, controlWrite)
	reportRead, reportWrite, err := os.Pipe()
	if err != nil {
		return nil, p, err
	}
	owned = append(owned, reportRead, reportWrite)
	outRead, outWrite, err := os.Pipe()
	if err != nil {
		return nil, p, err
	}
	owned = append(owned, outRead, outWrite)
	errRead, errWrite, err := os.Pipe()
	if err != nil {
		return nil, p, err
	}
	owned = append(owned, errRead, errWrite)
	if err = controlWrite.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		return nil, p, err
	}
	if err = reportRead.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return nil, p, err
	}
	cmd = exec.Command(a.executable, args...)
	cmd.Dir = a.cwd
	cmd.Env = append([]string{}, a.environment...)
	cmd.Stdin = null
	cmd.Stdout = outWrite
	cmd.Stderr = errWrite
	cmd.ExtraFiles = []*os.File{listenerFile, controlRead, reportWrite}
	if err = cmd.Start(); err != nil {
		return nil, p, err
	}
	p.control = controlWrite
	p.report = reportRead
	p.out = outRead
	p.stderr = errRead
	owned = nil
	for _, f := range []*os.File{listenerFile, null, controlRead, reportWrite, outWrite, errWrite} {
		p.startupError = errors.Join(p.startupError, f.Close())
	}
	// If closing child-only duplicates fails after Start, return the live owner
	// rather than orphaning a child. os.File Close normally cannot fail here;
	// retain any failure in the ordinary lifecycle's result via startupError.
	return cmd, p, nil
}
