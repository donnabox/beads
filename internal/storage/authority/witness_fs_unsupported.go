//go:build !unix

package authority

import (
	"context"
	"os"
)

func newWitnessManager(context.Context, string) (*Manager, error) { return nil, ErrWitnessUnsupported }
func openWitnessFile(string, bool) (*os.File, error)              { return nil, ErrWitnessUnsupported }
func checkWitnessFile(string, *os.File) error                     { return ErrWitnessUnsupported }
func readWitnessFile(string) (Snapshot, error)                    { return Snapshot{}, ErrWitnessUnsupported }
func waitWitnessLock(context.Context, *os.File, func(*os.File) error) error {
	return ErrWitnessUnsupported
}
func syncWitnessAncestors(context.Context, string) error { return ErrWitnessUnsupported }
