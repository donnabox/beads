package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
)

func TestCanonicalOutputLegacyBoundary(t *testing.T) {
	for _, args := range [][]string{
		{"--json", "--output-format=json"}, {"--output-format=json", "--json"},
		{"--json=false", "--output-format=json"}, {"--format=json", "--output-format=json"},
		{"--include=/properties"}, {"--omit=envelope"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			c := &cobra.Command{Use: "show"}
			c.Flags().Bool("json", false, "")
			c.Flags().String("format", "", "")
			registerCanonicalOutputFlags(c)
			if err := c.ParseFlags(args); err != nil {
				t.Fatal(err)
			}
			if _, _, err := canonicalOutputOptions(c); err == nil {
				t.Fatal("ambiguous legacy/canonical invocation accepted")
			}
		})
	}
	c := &cobra.Command{Use: "show"}
	registerCanonicalOutputFlags(c)
	if err := c.ParseFlags([]string{"--output-format=json", "--output-format=json", "--include=/properties", "--omit=/metadata"}); err != nil {
		t.Fatal(err)
	}
	opts, selected, err := canonicalOutputOptions(c)
	if err != nil || !selected {
		t.Fatalf("canonical selection: %v", err)
	}
	if _, err := clioutput.New(canonicalResourceDescriptor(), opts); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalFailureOutput(t *testing.T) {
	for _, raw := range []bool{false, true} {
		opts := clioutput.Options{Formats: []string{"json"}}
		if raw {
			opts.Omit = []string{"envelope"}
		}
		renderer, err := clioutput.New(canonicalResourceDescriptor(), opts)
		if err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		cmd := &cobra.Command{}
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		err = canonicalFailure(cmd, renderer, "not_found", "missing", 3)
		var status *exitError
		if !errors.As(err, &status) || status.Code != 3 || !strings.Contains(stderr.String(), "not_found") {
			t.Fatalf("failure: %v %s", err, &stderr)
		}
		if raw {
			if stdout.Len() != 0 {
				t.Fatal("raw failure emitted payload")
			}
			continue
		}
		var frame struct {
			Schema  int    `json:"schema_version"`
			Outcome string `json:"outcome"`
			Error   struct {
				Origin  string `json:"origin"`
				Problem struct {
					Code string `json:"code"`
				} `json:"problem"`
			} `json:"error"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &frame); err != nil {
			t.Fatal(err)
		}
		if frame.Schema != 2 || frame.Outcome != "failure" || frame.Error.Origin != "cli" || frame.Error.Problem.Code != "not_found" {
			t.Fatalf("failure contract: %s", &stdout)
		}
	}
}

type canonicalShortWriter struct{}

func (canonicalShortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestCanonicalOutputShortWrite(t *testing.T) {
	if !errors.Is(writeCanonicalDocument(canonicalShortWriter{}, []byte("payload")), io.ErrShortWrite) {
		t.Fatal("short write accepted")
	}
}

func TestCanonicalInvalidFormatFailureDoesNotPanic(t *testing.T) {
	for _, formats := range [][]string{{"bogus"}, {"detail", "json"}} {
		cmd := &cobra.Command{}
		registerCanonicalOutputFlags(cmd)
		for _, format := range formats {
			if err := cmd.Flags().Set("output-format", format); err != nil {
				t.Fatal(err)
			}
		}
		var out, diagnostic bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&diagnostic)
		renderer := canonicalFailureRendererFor(cmd)
		if renderer != nil {
			t.Fatal("ambiguous format returned renderer")
		}
		if err := canonicalFailure(cmd, renderer, "invalid_selector", "invalid format", 2); err == nil {
			t.Fatal("failure succeeded")
		}
		if out.Len() != 0 || !strings.Contains(diagnostic.String(), "invalid_selector") {
			t.Fatalf("unexpected refusal: %s / %s", &out, &diagnostic)
		}
	}
}

func TestCanonicalOnlyFlagsRequireOptIn(t *testing.T) {
	cmd := &cobra.Command{}
	registerCanonicalOutputFlags(cmd)
	cmd.Flags().Bool("paged", false, "")
	cmd.Flags().Lookup("paged").Annotations = map[string][]string{"canonical-output-only": {"true"}}
	if err := cmd.ParseFlags([]string{"--paged"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := canonicalOutputOptions(cmd); err == nil {
		t.Fatal("canonical paging silently accepted on legacy route")
	}
}

func TestLegacyIncludeIsNotCanonicalProjection(t *testing.T) {
	cmd := &cobra.Command{Use: "legacy"}
	cmd.Flags().String("include", "", "")
	if err := cmd.ParseFlags([]string{"--include=upstream"}); err != nil {
		t.Fatal(err)
	}
	if _, selected, err := canonicalOutputOptions(cmd); err != nil || selected {
		t.Fatalf("legacy flag reinterpreted: %v", err)
	}
}
