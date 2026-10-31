//go:build !cgo

package graphstore

import (
	"errors"
	"testing"
)

func TestIntrinsicPolicyUnavailableWithoutCGO(t *testing.T) {
	m, err := loadIntrinsicExpected()
	if err != nil {
		t.Fatal(err)
	}
	var selectedRows [][]any
	for _, object := range m.Objects {
		if object.Ignore == "Ignore" {
			selectedRows = append(selectedRows, []any{object.Name, int64(1)})
		}
	}
	for _, rows := range [][][]any{nil, selectedRows, {{"*", true}}} {
		if err := checkIntrinsicIgnorePolicy(m.Objects, rows); !errors.Is(err, ErrCapabilityUnavailable) {
			t.Fatalf("pure build invented native policy support: %v", err)
		}
	}
}
