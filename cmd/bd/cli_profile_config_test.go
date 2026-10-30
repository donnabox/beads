package main

import (
	"errors"
	"testing"
)

func TestCLIProfileStrictReaderBoundary(t *testing.T) {
	wantErr := errors.New("unreadable saved configuration")
	read := func() (string, bool, error) { return "", false, wantErr }
	if _, err := selectCLIProfileFromReader(nil, read); !errors.Is(err, wantErr) {
		t.Fatalf("reader error = %v; want original error", err)
	}
	argument := "compat"
	got, err := selectCLIProfileFromReader(&argument, func() (string, bool, error) {
		t.Fatal("explicit override must not require saved configuration")
		return "", false, wantErr
	})
	if err != nil || got.Profile != cliSurfaceCompat || got.Source != cliProfileArgument {
		t.Fatalf("override = %#v, %v", got, err)
	}
	if _, err := selectCLIProfileFromReader(nil, func() (string, bool, error) {
		return "unknown", true, nil
	}); err == nil {
		t.Fatal("invalid saved profile silently became compat")
	}
	got, err = selectCLIProfileFromReader(nil, func() (string, bool, error) {
		return "", false, nil
	})
	if err != nil || got.Profile != cliSurfaceCompat || got.Source != cliProfileBuiltin {
		t.Fatalf("absent = %#v, %v", got, err)
	}
}
