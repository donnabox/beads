package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// driverVersion identifies the driver's own behavior in the receipts. Bump it
// when normalization or the transcript layout changes.
const driverVersion = "1"

type sourceInfo struct {
	Commit string `json:"commit"`
	Dirty  *bool  `json:"dirty"` // nil when the source is not a git checkout we could read
}

type bdInfo struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Version string `json:"version"`
}

type scenarioReceipt struct {
	ID                string            `json:"id"`
	State             string            `json:"state"`
	Outcome           string            `json:"outcome"`
	Engine            string            `json:"engine"`
	Steps             int               `json:"steps"`
	File              string            `json:"file"`
	FileSHA256        string            `json:"file_sha256"`
	TranscriptSHA256A string            `json:"transcript_sha256_A"`
	TranscriptSHA256B string            `json:"transcript_sha256_B"`
	Equal             bool              `json:"equal"`
	DurationS         float64           `json:"duration_s"`
	Finding           string            `json:"finding,omitempty"`
	Gap               string            `json:"gap,omitempty"`
	Captures          map[string]string `json:"captures,omitempty"`
}

// receiptsDoc is receipts.json. discovered counts the scenario files in the
// directory whatever was selected; executed counts the scenarios processed.
// With no --scenario filter they must be equal, so a scenario that silently
// went missing cannot be mistaken for a green run.
type receiptsDoc struct {
	DriverVersion string            `json:"driver_version"`
	Source        sourceInfo        `json:"source"`
	BD            bdInfo            `json:"bd"`
	Env           map[string]string `json:"env"`
	ScenariosDir  string            `json:"scenarios_dir"`
	Engine        string            `json:"engine"`
	Discovered    int               `json:"discovered"`
	Executed      int               `json:"executed"`
	Scenarios     []scenarioReceipt `json:"scenarios"`
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // G304: the bd binary the caller named with --bd
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// bdVersion asks bd for its version from a throwaway workspace.
func (d *driver) bdVersion(ctx context.Context) (string, error) {
	w, err := newWorkspace()
	if err != nil {
		return "", harnessf("cannot create a workspace: %v", err)
	}
	defer func() { _ = w.close() }()
	res, err := w.exec(ctx, d.bd, []string{"version"}, nil)
	if err != nil {
		return "", err
	}
	if res.Exit != 0 {
		return "", harnessf("bd version failed (exit %d): %s", res.Exit, strings.TrimSpace(string(res.Stderr)))
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(res.Stdout)), "\n")
	return line, nil
}

func (d *driver) bdInfo(ctx context.Context) (bdInfo, error) {
	sum, err := fileSHA256(d.bd)
	if err != nil {
		return bdInfo{}, harnessf("cannot hash bd: %v", err)
	}
	version, err := d.bdVersion(ctx)
	if err != nil {
		return bdInfo{}, err
	}
	return bdInfo{Path: d.bd, SHA256: sum, Version: version}, nil
}

// gitProvenance records which source the scenarios came from. Anything it
// cannot read is left empty rather than guessed.
func gitProvenance(dir string) sourceInfo {
	git, err := exec.LookPath("git")
	if err != nil {
		return sourceInfo{}
	}
	read := func(args ...string) (string, bool) {
		cmd := exec.Command(git, append([]string{"-C", dir}, args...)...) //nolint:gosec // G204: git is the resolved system binary and the arguments are fixed strings
		cmd.Env = []string{"PATH=" + hermeticPath, "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0"}
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err == nil
	}
	commit, ok := read("rev-parse", "HEAD")
	if !ok {
		return sourceInfo{}
	}
	status, ok := read("status", "--porcelain")
	if !ok {
		return sourceInfo{Commit: commit}
	}
	dirty := status != ""
	return sourceInfo{Commit: commit, Dirty: &dirty}
}

// receiptEnv is the environment whitelist with the workspace root masked, so
// the receipts show what bd was given without depending on where it ran.
func receiptEnv() map[string]string {
	return hermeticVars(&workspace{Root: "<WS>", Work: "<WS>/work", Home: "<WS>/home"})
}

func newReceipt(engine string, r *scenarioResult) scenarioReceipt {
	rc := scenarioReceipt{
		ID:         r.Scenario.ID,
		State:      r.Scenario.State,
		Outcome:    string(r.Outcome),
		Engine:     engine,
		Steps:      r.Steps,
		File:       filepath.Base(r.Scenario.File),
		FileSHA256: r.Scenario.SHA256,
		Equal:      r.Equal,
		DurationS:  r.Duration.Seconds(),
		Finding:    r.Scenario.Finding,
		Gap:        r.Scenario.Gap,
		Captures:   r.Captures,
	}
	if r.Outcome != outcomeSkip {
		rc.TranscriptSHA256A = transcriptDigest(r.A)
		rc.TranscriptSHA256B = transcriptDigest(r.B)
	}
	return rc
}

// writeTranscripts writes transcripts/<scenario>/<A|B>/NN-<step>/{argv,exit,
// stdout,stderr,sha256} for the normalized steps of one workspace.
func writeTranscripts(out, id, ws string, steps []*stepResult) error {
	for i, s := range steps {
		dir := filepath.Join(out, "transcripts", id, ws, fmt.Sprintf("%02d-%s", i+1, s.Name))
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
		argv := s.Argv
		if argv == nil {
			argv = []string{}
		}
		encoded, err := json.Marshal(argv)
		if err != nil {
			return err
		}
		files := []struct {
			name string
			data []byte
		}{
			{"argv", append(encoded, '\n')},
			{"exit", []byte(strconv.Itoa(s.Exit) + "\n")},
			{"stdout", s.Stdout},
			{"stderr", s.Stderr},
			{"sha256", []byte(stepDigest(s) + "\n")},
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(dir, f.name), f.data, 0o600); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeReceipts(out string, doc *receiptsDoc) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "receipts.json"), append(data, '\n'), 0o600)
}
