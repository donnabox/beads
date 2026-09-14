package utils

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCanonicalizeExistingPath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "MixedCase")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalizeExistingPath(path)
	if err != nil || !filepath.IsAbs(canonical) {
		t.Fatalf("canonical: %q %v", canonical, err)
	}
	if bestEffort := CanonicalizePath(path); bestEffort != canonical {
		t.Fatalf("existing canonicalizers disagree: %q != %q", bestEffort, canonical)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(path, alias); err != nil {
		t.Fatal(err)
	}
	got, err := CanonicalizeExistingPath(alias)
	if err != nil || got != canonical {
		t.Fatalf("alias: %q %v", got, err)
	}
	if bestEffort := CanonicalizePath(alias); bestEffort != canonical {
		t.Fatalf("symlink canonicalizers disagree: %q != %q", bestEffort, canonical)
	}
	if _, err := CanonicalizeExistingPath(filepath.Join(root, "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing path fallback: %v", err)
	}
	if runtime.GOOS == "darwin" {
		lower := filepath.Join(root, strings.ToLower(filepath.Base(path)))
		info, err := os.Stat(lower)
		if errors.Is(err, os.ErrNotExist) {
			t.Log("fixture filesystem is case-sensitive; no case alias exists")
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		original, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !os.SameFile(info, original) {
			t.Fatal("unexpected different lowercase fixture")
		}
		got, err := CanonicalizeExistingPath(lower)
		if err != nil || got != canonical {
			t.Fatalf("case alias: %q, want %q: %v", got, canonical, err)
		}
		if bestEffort := CanonicalizePath(lower); bestEffort != canonical {
			t.Fatalf("case canonicalizers disagree: %q != %q", bestEffort, canonical)
		}
	}
}
