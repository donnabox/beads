package main

import (
	"bytes"
	"errors"
	"testing"
)

type graphFailingWriter struct {
	err   error
	calls int
}

func (w *graphFailingWriter) Write([]byte) (int, error) { w.calls++; return 0, w.err }

func TestGraphPreviewOutputFailures(t *testing.T) {
	failure := errors.New("output device failed")
	for _, structured := range []bool{false, true} {
		w := &graphFailingWriter{err: failure}
		if err := graphPrintTo(w, map[string]any{"changes": []any{}}, "complete comparison", false, structured); !errors.Is(err, failure) || w.calls == 0 {
			t.Fatalf("structured=%v output failure lost: %v", structured, err)
		}
	}
	w := &graphFailingWriter{err: failure}
	if err := graphPrintTo(w, nil, "quiet", true, false); err != nil || w.calls != 0 {
		t.Fatalf("quiet human wrote: %v", err)
	}
	if err := graphPrintTo(w, nil, "quiet", true, true); !errors.Is(err, failure) {
		t.Fatalf("quiet JSON hid output failure: %v", err)
	}
	var out bytes.Buffer
	if err := graphPrintTo(&out, nil, "whole result", false, false); err != nil || out.String() != "whole result\n" {
		t.Fatalf("human framing changed: %q %v", out.String(), err)
	}
}
