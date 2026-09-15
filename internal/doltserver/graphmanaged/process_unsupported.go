//go:build !darwin && !linux

package graphmanaged

import (
	"os"
	"os/exec"
)

func platformSupported() bool                      { return false }
func ownedFile(os.FileInfo) bool                   { return false }
func protectedArtifact(os.FileInfo) bool           { return false }
func trustedAncestor(os.FileInfo) bool             { return false }
func trustedDirectory(string) (os.FileInfo, error) { return nil, errUnsupported }
func openRegular(string) (*os.File, error)         { return nil, errUnsupported }
func openDirectory(string) (*os.File, error)       { return nil, errUnsupported }
func signalTerm(processOwner) error                { return errUnsupported }
func spawn(admitted, uint64) (*exec.Cmd, processPipes, error) {
	return nil, processPipes{}, errUnsupported
}
