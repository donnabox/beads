//go:build !cgo

package graphstore

import "fmt"

// A pure binary cannot authenticate native ignore semantics. Never simulate
// the law or emit a checked Types profile without its actual implementation.
func checkIntrinsicIgnorePolicy(_ []intrinsicExpectedObject, _ [][]any) error {
	return fmt.Errorf("%w: Types profile requires native ignore-policy support", ErrCapabilityUnavailable)
}
