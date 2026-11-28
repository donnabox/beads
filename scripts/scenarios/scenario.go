package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// manifestError is a problem with the scenario files themselves. run maps it
// to exit 3, before any bd is started.
type manifestError struct{ Problems []string }

func (e *manifestError) Error() string {
	return "manifest error:\n  - " + strings.Join(e.Problems, "\n  - ")
}

// Scenario is one JSON data file: what to run against bd and what the cited
// spec text says must happen. It carries no code; the assertion vocabulary is
// closed.
type Scenario struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Spec    []string `json:"spec"`
	State   string   `json:"state"` // pass | xfail | skip
	Finding string   `json:"finding,omitempty"`
	Gap     string   `json:"gap,omitempty"`
	Steps   []Step   `json:"steps"`

	File   string `json:"-"` // path of the scenario file
	SHA256 string `json:"-"` // hex sha256 of the file's bytes
}

// Step is one bd invocation. argv may reference ${NAME} captures of earlier steps.
type Step struct {
	Name    string              `json:"name"`
	Argv    []string            `json:"argv"`
	Actor   string              `json:"actor,omitempty"`
	Stdin   *Stdin              `json:"stdin,omitempty"`
	Expect  []*assertion        `json:"expect"`
	Capture map[string]*Capture `json:"capture,omitempty"`
}

// Stdin is a step's input bytes: literal text, base64, or a deterministic
// generated body for sizes no one wants to paste into a file. A name makes the
// bytes referable from stdout equals_input.
type Stdin struct {
	Name     string    `json:"name,omitempty"`
	Text     *string   `json:"text,omitempty"`
	B64      *string   `json:"b64,omitempty"`
	Generate *Generate `json:"generate,omitempty"`
}

// Generate is repeat cycled and cut at exactly Bytes bytes. No randomness.
type Generate struct {
	Repeat string `json:"repeat"`
	Bytes  int    `json:"bytes"`
}

// Capture names a value of a step's result for later steps and assertions.
// Kind marks a value that is random per workspace (a token, a generated id) so
// the normalizer can give it a stable ordinal instead of hiding it.
type Capture struct {
	Path       string `json:"path,omitempty"`
	Stdout     bool   `json:"stdout,omitempty"`
	StderrCode bool   `json:"stderr_code,omitempty"`
	Kind       string `json:"kind,omitempty"`
	jp         *jsonPath
}

type captureKind string

const (
	kindStable captureKind = ""
	kindToken  captureKind = "token"
	kindID     captureKind = "id"
)

// maxGenerated caps stdin.generate so a typo cannot ask for gigabytes.
const maxGenerated = 64 << 20

var (
	idRE        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	stepNameRE  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,47}$`)
	actorRE     = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	findingRE   = regexp.MustCompile(`^[a-z][a-z0-9]*-[a-z0-9]+(\.[a-z0-9]+)*$`)
	gapRE       = regexp.MustCompile(`^be-tatfn#[1-9][0-9]*$`)
	identifier  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	scenarioExt = ".json"
)

// loadScenarios reads every top-level *.json file of dir at run time. There is
// no registry: a scenario exists because its file does. A missing or empty
// directory is a manifest error, never a silent zero.
func loadScenarios(dir string) ([]*Scenario, error) {
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return nil, &manifestError{[]string{fmt.Sprintf("scenarios directory %s is missing or not a directory", dir)}}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, &manifestError{[]string{fmt.Sprintf("scenarios directory %s: %v", dir, err)}}
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), scenarioExt) {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return nil, &manifestError{[]string{fmt.Sprintf("no scenario files (*%s) in %s", scenarioExt, dir)}}
	}
	sort.Strings(names)

	var (
		scs      []*Scenario
		problems []string
		byLower  = map[string]string{}
	)
	for _, name := range names {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path) //nolint:gosec // G304: a scenario file in the caller's --scenarios directory
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		sc, probs := parseScenario(strings.TrimSuffix(name, scenarioExt), data)
		for _, p := range probs {
			problems = append(problems, name+": "+p)
		}
		if sc == nil {
			continue
		}
		sc.File = path
		if other, dup := byLower[strings.ToLower(sc.ID)]; dup {
			problems = append(problems, fmt.Sprintf("%s: duplicate scenario id %q (also in %s); ids are compared case-insensitively because transcript directories are named after them", name, sc.ID, other))
		}
		byLower[strings.ToLower(sc.ID)] = name
		scs = append(scs, sc)
	}
	if len(problems) > 0 {
		return nil, &manifestError{problems}
	}
	sort.Slice(scs, func(i, j int) bool { return scs[i].ID < scs[j].ID })
	return scs, nil
}

// parseScenario decodes strictly (an unknown field anywhere is an error) and
// validates. The returned Scenario is nil only when decoding failed.
func parseScenario(stem string, data []byte) (*Scenario, []string) {
	var sc Scenario
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&sc); err != nil {
		return nil, []string{err.Error()}
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, []string{"unexpected data after the scenario object"}
	}
	sum := sha256.Sum256(data)
	sc.SHA256 = hex.EncodeToString(sum[:])
	return &sc, sc.validate(stem)
}

func (sc *Scenario) validate(stem string) []string {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	if !idRE.MatchString(sc.ID) {
		add("invalid id %q (want %s)", sc.ID, idRE)
	}
	if sc.ID != stem {
		add("id %q must equal the file stem %q", sc.ID, stem)
	}
	if len(sc.Spec) == 0 {
		add("spec must cite the spec text the expectations are declared from")
	}
	for _, s := range sc.Spec {
		if strings.TrimSpace(s) == "" {
			add("spec has an empty citation")
		}
	}
	switch sc.State {
	case "pass":
		if sc.Finding != "" || sc.Gap != "" {
			add("finding and gap are only valid with state xfail and skip")
		}
	case "xfail":
		if !findingRE.MatchString(sc.Finding) {
			add("state xfail needs a finding bead id in finding (got %q)", sc.Finding)
		}
		if sc.Gap != "" {
			add("gap is only valid with state skip")
		}
	case "skip":
		if !gapRE.MatchString(sc.Gap) {
			add("state skip needs an item number from be-tatfn in gap, like be-tatfn#3 (got %q)", sc.Gap)
		}
		if sc.Finding != "" {
			add("finding is only valid with state xfail")
		}
	default:
		add("unknown state %q (want pass, xfail or skip)", sc.State)
	}
	if len(sc.Steps) == 0 {
		add("steps must not be empty")
		return problems
	}
	return append(problems, sc.validateSteps()...)
}

func (sc *Scenario) validateSteps() []string {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	inputs := map[string]bool{}
	for _, st := range sc.Steps {
		if st.Stdin != nil && st.Stdin.Name != "" {
			if !identifier.MatchString(st.Stdin.Name) {
				add("invalid stdin name %q", st.Stdin.Name)
			}
			if inputs[st.Stdin.Name] {
				add("stdin name %q is declared more than once", st.Stdin.Name)
			}
			inputs[st.Stdin.Name] = true
		}
	}

	seen := map[string]bool{}     // step names so far, including the current one
	captured := map[string]bool{} // captures of strictly earlier steps
	for i := range sc.Steps {
		st := &sc.Steps[i]
		where := fmt.Sprintf("step %d (%q)", i+1, st.Name)
		if !stepNameRE.MatchString(st.Name) {
			add("%s: invalid step name (want %s)", where, stepNameRE)
		}
		if seen[st.Name] {
			add("duplicate step name %q", st.Name)
		}
		seen[st.Name] = true
		if len(st.Argv) == 0 {
			add("%s: argv must not be empty", where)
		}
		for _, arg := range st.Argv {
			names, err := varRefs(arg)
			if err != nil {
				add("%s: %v", where, err)
			}
			for _, n := range names {
				if !captured[n] {
					add("%s: undefined variable ${%s} (only captures of earlier steps can be used)", where, n)
				}
			}
		}
		if st.Actor != "" && !actorRE.MatchString(st.Actor) {
			add("%s: invalid actor %q (want %s)", where, st.Actor, actorRE)
		}
		if st.Stdin != nil {
			for _, p := range st.Stdin.validate() {
				add("%s: %s", where, p)
			}
		}

		capNames := make([]string, 0, len(st.Capture))
		for name := range st.Capture {
			capNames = append(capNames, name)
		}
		sort.Strings(capNames)
		here := map[string]bool{}
		for _, name := range capNames {
			c := st.Capture[name]
			if c == nil {
				add("%s: capture %q is null", where, name)
				continue
			}
			if !identifier.MatchString(name) {
				add("%s: invalid capture name %q", where, name)
			}
			if captured[name] || here[name] {
				add("%s: capture %q is defined more than once", where, name)
			}
			here[name] = true
			for _, p := range c.validate(name) {
				add("%s: %s", where, p)
			}
		}
		for name := range here {
			captured[name] = true
		}

		exits := 0
		for _, a := range st.Expect {
			if a == nil {
				add("%s: expect has a null entry", where)
				continue
			}
			if a.Op == "exit" {
				exits++
			}
			for _, p := range a.checkRefs(inputs, captured, seen) {
				add("%s: %s", where, p)
			}
		}
		if exits != 1 {
			add("%s: needs exactly one exit expectation, has %d", where, exits)
		}
	}
	return problems
}

func (s *Stdin) validate() []string {
	var problems []string
	forms := 0
	if s.Text != nil {
		forms++
	}
	if s.B64 != nil {
		forms++
		if _, err := base64.StdEncoding.DecodeString(*s.B64); err != nil {
			problems = append(problems, fmt.Sprintf("stdin b64 is not valid base64: %v", err))
		}
	}
	if s.Generate != nil {
		forms++
		if s.Generate.Repeat == "" || s.Generate.Bytes < 1 || s.Generate.Bytes > maxGenerated {
			problems = append(problems, fmt.Sprintf("stdin generate needs a non-empty repeat and 1..%d bytes", maxGenerated))
		}
	}
	if forms != 1 {
		problems = append(problems, "stdin needs exactly one of text, b64 or generate")
	}
	return problems
}

func (c *Capture) validate(name string) []string {
	var problems []string
	sources := 0
	if c.Path != "" {
		sources++
		jp, err := singlePath("capture "+name, c.Path)
		if err != nil {
			problems = append(problems, err.Error())
		}
		c.jp = jp
	}
	if c.Stdout {
		sources++
	}
	if c.StderrCode {
		sources++
	}
	if sources != 1 {
		problems = append(problems, fmt.Sprintf("capture %q needs exactly one of path, stdout or stderr_code", name))
	}
	switch captureKind(c.Kind) {
	case kindStable, kindToken, kindID:
	default:
		problems = append(problems, fmt.Sprintf("capture %q has unknown kind %q (want token or id, or none)", name, c.Kind))
	}
	return problems
}

// checkRefs verifies the names an assertion refers to exist by the time it runs.
func (a *assertion) checkRefs(inputs, captured, steps map[string]bool) []string {
	var problems []string
	needCapture := func(name string) {
		if !captured[name] {
			problems = append(problems, fmt.Sprintf("%s refers to undefined capture %q", a.Op, name))
		}
	}
	switch a.Op {
	case "stdout":
		if a.EqualsInput != "" && !inputs[a.EqualsInput] {
			problems = append(problems, fmt.Sprintf("stdout equals_input refers to undeclared input %q (declare it with a named stdin)", a.EqualsInput))
		}
		if a.EqualsCapture != "" {
			needCapture(a.EqualsCapture)
		}
	case "json_path":
		needCapture(a.EqualsCapture)
	case "stderr_contains":
		needCapture(a.Capture)
	case "compare":
		for _, r := range []ref{a.Left, a.Right} {
			if !steps[r.Step] {
				problems = append(problems, fmt.Sprintf("compare references unknown step %q (only this step and earlier ones may be referenced)", r.Step))
			}
		}
	}
	return problems
}

// varRefs returns the ${NAME} references in s, or an error for a malformed one.
func varRefs(s string) ([]string, error) {
	var names []string
	for {
		i := strings.Index(s, "${")
		if i < 0 {
			return names, nil
		}
		rest := s[i+2:]
		j := strings.IndexByte(rest, '}')
		if j < 0 {
			return names, fmt.Errorf("unterminated variable reference in %q", s)
		}
		name := rest[:j]
		if !identifier.MatchString(name) {
			return names, fmt.Errorf("invalid variable name %q in a ${...} reference", name)
		}
		names = append(names, name)
		s = rest[j+1:]
	}
}

// expandVars substitutes ${NAME} from vars.
func expandVars(s string, vars map[string]string) (string, error) {
	var b strings.Builder
	for {
		i := strings.Index(s, "${")
		if i < 0 {
			b.WriteString(s)
			return b.String(), nil
		}
		rest := s[i+2:]
		j := strings.IndexByte(rest, '}')
		if j < 0 {
			return "", fmt.Errorf("unterminated variable reference in %q", s)
		}
		val, ok := vars[rest[:j]]
		if !ok {
			return "", fmt.Errorf("undefined variable ${%s}", rest[:j])
		}
		b.WriteString(s[:i])
		b.WriteString(val)
		s = rest[j+1:]
	}
}

// bytes materializes the step's input.
func (s *Stdin) bytes() ([]byte, error) {
	switch {
	case s.Text != nil:
		return []byte(*s.Text), nil
	case s.B64 != nil:
		return base64.StdEncoding.DecodeString(*s.B64)
	case s.Generate != nil:
		out := make([]byte, 0, s.Generate.Bytes+len(s.Generate.Repeat))
		for len(out) < s.Generate.Bytes {
			out = append(out, s.Generate.Repeat...)
		}
		return out[:s.Generate.Bytes], nil
	}
	return nil, errors.New("stdin has no content")
}

// inputs materializes every named stdin of the scenario.
func (sc *Scenario) inputs() (map[string][]byte, error) {
	out := map[string][]byte{}
	for _, st := range sc.Steps {
		if st.Stdin == nil || st.Stdin.Name == "" {
			continue
		}
		b, err := st.Stdin.bytes()
		if err != nil {
			return nil, fmt.Errorf("stdin %q: %w", st.Stdin.Name, err)
		}
		out[st.Stdin.Name] = b
	}
	return out, nil
}
