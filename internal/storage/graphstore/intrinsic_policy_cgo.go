//go:build cgo

package graphstore

import (
	"fmt"
	"unicode/utf8"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
)

// The exact pinned native law is available only with the native backend.
func checkIntrinsicIgnorePolicy(objects []intrinsicExpectedObject, rows [][]any) error {
	if len(rows) > 256 {
		return fmt.Errorf("%w: ignore rows", ErrLimitExceeded)
	}
	patterns := doltdb.IgnorePatterns{}
	seen := make(map[string]bool)
	bytes := 0
	for _, row := range rows {
		if len(row) != 2 {
			return fmt.Errorf("%w: ignore row shape", ErrInvalidStore)
		}
		pattern, err := intrinsicCellString(row[0])
		if err != nil {
			return err
		}
		if !utf8.ValidString(pattern) || len(pattern) > 1024 || seen[pattern] {
			return fmt.Errorf("%w: ignore pattern", ErrInvalidStore)
		}
		bytes += len(pattern)
		if bytes > 64<<10 {
			return fmt.Errorf("%w: ignore bytes", ErrLimitExceeded)
		}
		ignored, err := intrinsicIgnoreBoolean(row[1])
		if err != nil {
			return err
		}
		seen[pattern] = true
		patterns = append(patterns, doltdb.NewIgnorePattern(pattern, ignored))
	}
	for _, object := range objects {
		if object.Kind != "BASE TABLE" {
			continue
		}
		if object.Ignore != "Ignore" && object.Ignore != "DontIgnore" {
			return fmt.Errorf("%w: expected ignore decision", ErrInvalidStore)
		}
		result, err := (&patterns).IsTableNameIgnored(doltdb.TableName{Name: object.Name})
		want := doltdb.DontIgnore
		if object.Ignore == "Ignore" {
			want = doltdb.Ignore
		}
		if err != nil || result != want {
			return fmt.Errorf("%w: ignore policy for %s: %v", ErrInvalidStore, object.Name, err)
		}
	}
	return nil
}
