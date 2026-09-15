//go:build !darwin && !linux

package graphmanaged

import (
	"os"
	"os/exec"
)

func platformSupported() bool                      { return false }
func ownedFile(os.FileInfo) bool                   { return false }
func trustedDirectory(string) (os.FileInfo, error) { return nil, errUnsupported }
func openRegular(string) (*os.File, error)         { return nil, errUnsupported }
func openDirectory(string) (*os.File, error)       { return nil, errUnsupported }
func signalTerm(*os.Process) error                 { return errUnsupported }
func spawn(admitted, []string) (*exec.Cmd, processPipes, error) {
	return nil, processPipes{}, errUnsupported
}
