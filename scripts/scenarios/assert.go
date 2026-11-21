package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// stepResult is one bd invocation as observed: what was run and what came back.
type stepResult struct {
	Name   string
	Argv   []string
	Exit   int
	Stdout []byte
	Stderr []byte
}

// evalEnv is what an assertion may consult besides the step it sits on.
type evalEnv struct {
	Steps    map[string]*stepResult // by step name; the current step is included
	Captures map[string]string      // captured values, per workspace
	Inputs   map[string][]byte      // declared stdin payloads, by name
}

// ref names one step's field or JSON-path value, for compare.
type ref struct {
	Step  string `json:"step"`
	Field string `json:"field,omitempty"`
	Path  string `json:"path,omitempty"`
	jp    *jsonPath
}

// assertion is one entry of the closed assertion vocabulary. A scenario file
// can say only these things about a step; anything else is a manifest error,
// so a scenario is data and never carries code.
type assertion struct {
	Op string

	ExitEquals *int // exit
	Nonzero    bool // exit

	CodeEquals    string // stderr_code
	CodeNotEquals string // stderr_code

	EqualsInput   string // stdout
	EqualsCapture string // stdout, json_path
	Empty         bool   // stdout

	In     string // contains, not_contains: "stdout"
	Path   string // contains, not_contains, json_path, id_set
	Marker string // contains, not_contains

	IDs []string // id_set

	Left, Right ref    // compare
	Relation    string // compare: equal | differ

	Capture string // stderr_contains

	jp *jsonPath
}

type assertionWire struct {
	Op            string          `json:"op"`
	Equals        json.RawMessage `json:"equals"`
	Nonzero       *bool           `json:"nonzero"`
	NotEquals     *string         `json:"not_equals"`
	EqualsInput   *string         `json:"equals_input"`
	EqualsCapture *string         `json:"equals_capture"`
	Empty         *bool           `json:"empty"`
	In            *string         `json:"in"`
	Path          *string         `json:"path"`
	Marker        *string         `json:"marker"`
	Left          *ref            `json:"left"`
	Right         *ref            `json:"right"`
	Relation      *string         `json:"relation"`
	Capture       *string         `json:"capture"`
}

// opFields is the vocabulary: each op and the only fields it takes.
var opFields = map[string][]string{
	"exit":            {"equals", "nonzero"},
	"stderr_code":     {"equals", "not_equals"},
	"stdout":          {"equals_input", "equals_capture", "empty"},
	"contains":        {"in", "path", "marker"},
	"not_contains":    {"in", "path", "marker"},
	"json_path":       {"path", "equals_capture"},
	"id_set":          {"path", "equals"},
	"compare":         {"left", "right", "relation"},
	"stderr_contains": {"capture"},
}

var compareFields = []string{"stdout", "stderr", "stderr_code", "exit"}

func rawPresent(raw json.RawMessage) bool {
	return len(raw) > 0 && string(raw) != "null"
}

func (w *assertionWire) present() []string {
	var p []string
	add := func(ok bool, name string) {
		if ok {
			p = append(p, name)
		}
	}
	add(rawPresent(w.Equals), "equals")
	add(w.Nonzero != nil, "nonzero")
	add(w.NotEquals != nil, "not_equals")
	add(w.EqualsInput != nil, "equals_input")
	add(w.EqualsCapture != nil, "equals_capture")
	add(w.Empty != nil, "empty")
	add(w.In != nil, "in")
	add(w.Path != nil, "path")
	add(w.Marker != nil, "marker")
	add(w.Left != nil, "left")
	add(w.Right != nil, "right")
	add(w.Relation != nil, "relation")
	add(w.Capture != nil, "capture")
	return p
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func knownOps() string {
	ops := make([]string, 0, len(opFields))
	for op := range opFields {
		ops = append(ops, op)
	}
	sort.Strings(ops)
	return strings.Join(ops, ", ")
}

func parseAssertion(raw []byte) (*assertion, error) {
	// A scenario that invents a primitive is best told so before anything about
	// the fields it invented for it, so the op is checked first. Decoding errors
	// are left to the strict decode below.
	var head struct {
		Op string `json:"op"`
	}
	_ = json.Unmarshal(raw, &head)
	if _, ok := opFields[head.Op]; !ok && head.Op != "" {
		return nil, fmt.Errorf("unknown op %q (the vocabulary is closed: %s)", head.Op, knownOps())
	}

	var w assertionWire
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&w); err != nil {
		return nil, fmt.Errorf("assertion: %w", err)
	}
	if w.Op == "" {
		return nil, errors.New(`assertion needs an "op"`)
	}
	allowed := opFields[w.Op]
	for _, f := range w.present() {
		if !containsString(allowed, f) {
			return nil, fmt.Errorf("op %q does not take the field %q", w.Op, f)
		}
	}
	a := &assertion{Op: w.Op}
	var err error
	switch w.Op {
	case "exit":
		err = a.parseExit(&w)
	case "stderr_code":
		err = a.parseStderrCode(&w)
	case "stdout":
		err = a.parseStdout(&w)
	case "contains", "not_contains":
		err = a.parseContains(&w)
	case "json_path":
		err = a.parseJSONPathOp(&w)
	case "id_set":
		err = a.parseIDSet(&w)
	case "compare":
		err = a.parseCompare(&w)
	case "stderr_contains":
		if w.Capture == nil || *w.Capture == "" {
			err = errors.New(`stderr_contains needs a "capture" name`)
		} else {
			a.Capture = *w.Capture
		}
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

// UnmarshalJSON makes a scenario file's expect list strict by construction.
func (a *assertion) UnmarshalJSON(b []byte) error {
	parsed, err := parseAssertion(b)
	if err != nil {
		return err
	}
	*a = *parsed
	return nil
}

func (a *assertion) parseExit(w *assertionWire) error {
	hasEquals, hasNonzero := rawPresent(w.Equals), w.Nonzero != nil
	if hasEquals == hasNonzero {
		return errors.New(`exit needs exactly one of "equals" or "nonzero"`)
	}
	if hasNonzero {
		if !*w.Nonzero {
			return errors.New(`exit "nonzero" must be true`)
		}
		a.Nonzero = true
		return nil
	}
	var n int
	if err := json.Unmarshal(w.Equals, &n); err != nil {
		return fmt.Errorf("exit equals must be an integer: %w", err)
	}
	if n < 0 || n > 255 {
		return fmt.Errorf("exit code %d is out of range 0..255", n)
	}
	a.ExitEquals = &n
	return nil
}

func (a *assertion) parseStderrCode(w *assertionWire) error {
	hasEquals, hasNot := rawPresent(w.Equals), w.NotEquals != nil
	if hasEquals == hasNot {
		return errors.New(`stderr_code needs exactly one of "equals" or "not_equals"`)
	}
	if hasNot {
		if *w.NotEquals == "" {
			return errors.New(`stderr_code "not_equals" must not be empty`)
		}
		a.CodeNotEquals = *w.NotEquals
		return nil
	}
	var s string
	if err := json.Unmarshal(w.Equals, &s); err != nil {
		return fmt.Errorf("stderr_code equals must be a string: %w", err)
	}
	if s == "" {
		return errors.New(`stderr_code "equals" must not be empty`)
	}
	a.CodeEquals = s
	return nil
}

func (a *assertion) parseStdout(w *assertionWire) error {
	n := 0
	if w.EqualsInput != nil {
		n++
		a.EqualsInput = *w.EqualsInput
	}
	if w.EqualsCapture != nil {
		n++
		a.EqualsCapture = *w.EqualsCapture
	}
	if w.Empty != nil {
		n++
		if !*w.Empty {
			return errors.New(`stdout "empty" must be true`)
		}
		a.Empty = true
	}
	if n != 1 {
		return errors.New(`stdout needs exactly one of "equals_input", "equals_capture" or "empty"`)
	}
	if (w.EqualsInput != nil && a.EqualsInput == "") || (w.EqualsCapture != nil && a.EqualsCapture == "") {
		return errors.New(`stdout: the input or capture name must not be empty`)
	}
	return nil
}

// singlePath parses a JSON path that must select exactly one value.
func singlePath(op, path string) (*jsonPath, error) {
	jp, err := parseJSONPath(path)
	if err != nil {
		return nil, err
	}
	if jp.hasWildcard() {
		return nil, fmt.Errorf("%s: path %q must select one value, not a [*] wildcard", op, path)
	}
	return jp, nil
}

func (a *assertion) parseContains(w *assertionWire) error {
	if w.Marker == nil || *w.Marker == "" {
		return fmt.Errorf(`%s needs a non-empty "marker"`, a.Op)
	}
	a.Marker = *w.Marker
	if (w.In != nil) == (w.Path != nil) {
		return fmt.Errorf(`%s needs exactly one of "in" or "path"`, a.Op)
	}
	if w.In != nil {
		if *w.In != "stdout" {
			return fmt.Errorf(`%s: "in" must be "stdout", got %q`, a.Op, *w.In)
		}
		a.In = "stdout"
		return nil
	}
	jp, err := singlePath(a.Op, *w.Path)
	if err != nil {
		return err
	}
	a.Path, a.jp = *w.Path, jp
	return nil
}

func (a *assertion) parseJSONPathOp(w *assertionWire) error {
	if w.Path == nil {
		return errors.New(`json_path needs a "path"`)
	}
	jp, err := singlePath(a.Op, *w.Path)
	if err != nil {
		return err
	}
	a.Path, a.jp = *w.Path, jp
	if w.EqualsCapture == nil || *w.EqualsCapture == "" {
		return errors.New(`json_path needs "equals_capture", the captured value it must equal`)
	}
	a.EqualsCapture = *w.EqualsCapture
	return nil
}

func (a *assertion) parseIDSet(w *assertionWire) error {
	if w.Path == nil {
		return errors.New(`id_set needs a "path"`)
	}
	jp, err := parseJSONPath(*w.Path)
	if err != nil {
		return err
	}
	a.Path, a.jp = *w.Path, jp
	if !rawPresent(w.Equals) {
		return errors.New(`id_set needs the declared set in "equals"`)
	}
	if err := json.Unmarshal(w.Equals, &a.IDs); err != nil {
		return fmt.Errorf("id_set equals must be an array of strings: %w", err)
	}
	seen := map[string]bool{}
	for _, id := range a.IDs {
		if seen[id] {
			return fmt.Errorf("id_set declares a duplicate id %q", id)
		}
		seen[id] = true
	}
	return nil
}

func (a *assertion) parseCompare(w *assertionWire) error {
	if w.Left == nil || w.Right == nil {
		return errors.New(`compare needs "left" and "right"`)
	}
	for side, r := range map[string]*ref{"left": w.Left, "right": w.Right} {
		if r.Step == "" {
			return fmt.Errorf("compare %s needs a step", side)
		}
		if (r.Field == "") == (r.Path == "") {
			return fmt.Errorf(`compare %s needs exactly one of "field" or "path"`, side)
		}
		if r.Field != "" && !containsString(compareFields, r.Field) {
			return fmt.Errorf("compare field %q is not one of %s", r.Field, strings.Join(compareFields, ", "))
		}
		if r.Path != "" {
			jp, err := singlePath("compare", r.Path)
			if err != nil {
				return err
			}
			r.jp = jp
		}
	}
	if w.Relation == nil || (*w.Relation != "equal" && *w.Relation != "differ") {
		return errors.New(`compare "relation" must be "equal" or "differ"`)
	}
	a.Left, a.Right, a.Relation = *w.Left, *w.Right, *w.Relation
	return nil
}

var stderrCodeRE = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// stderrCode is the typed code a refusal carries: the "code" of a JSON
// {code,message,retryable} object, or the leading "code:" token of the first
// line of the CLI's plain-text form. recall refuses --json, so its refusals
// only ever come in the plain-text form. "" means no typed code.
func stderrCode(stderr []byte) string {
	trimmed := bytes.TrimSpace(stderr)
	if len(trimmed) == 0 {
		return ""
	}
	if trimmed[0] == '{' {
		var obj map[string]any
		if err := json.NewDecoder(bytes.NewReader(trimmed)).Decode(&obj); err != nil {
			return ""
		}
		code, _ := obj["code"].(string)
		return code
	}
	line, _, _ := bytes.Cut(trimmed, []byte("\n"))
	code, _, found := strings.Cut(string(line), ": ")
	if found && stderrCodeRE.MatchString(code) {
		return code
	}
	return ""
}

// decodeJSON reads exactly one JSON document, keeping numbers exact.
func decodeJSON(b []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing data after the JSON document")
	}
	return v, nil
}

// scalarString renders a JSON scalar the way a scenario captures and compares it.
func scalarString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case json.Number:
		return t.String(), true
	case bool:
		return strconv.FormatBool(t), true
	}
	return "", false
}

// pathScalar evaluates a single-valued path against a stdout document.
func pathScalar(stdout []byte, jp *jsonPath) (string, error) {
	doc, err := decodeJSON(stdout)
	if err != nil {
		return "", fmt.Errorf("stdout is not a JSON document: %w", err)
	}
	vals, err := jp.eval(doc)
	if err != nil {
		return "", err
	}
	if len(vals) != 1 {
		return "", fmt.Errorf("json path %s selected %d values, want 1", jp.raw, len(vals))
	}
	s, ok := scalarString(vals[0])
	if !ok {
		return "", fmt.Errorf("json path %s is %s, not a string, number or boolean", jp.raw, describeJSON(vals[0]))
	}
	return s, nil
}

func (r ref) value(env *evalEnv) (string, error) {
	st, ok := env.Steps[r.Step]
	if !ok {
		return "", fmt.Errorf("unknown step %q", r.Step)
	}
	if r.jp != nil {
		return pathScalar(st.Stdout, r.jp)
	}
	switch r.Field {
	case "stdout":
		return string(st.Stdout), nil
	case "stderr":
		return string(st.Stderr), nil
	case "stderr_code":
		return stderrCode(st.Stderr), nil
	case "exit":
		return strconv.Itoa(st.Exit), nil
	}
	return "", fmt.Errorf("unknown field %q", r.Field)
}

func (a *assertion) targetString(cur *stepResult) (string, error) {
	if a.jp == nil {
		return string(cur.Stdout), nil
	}
	doc, err := decodeJSON(cur.Stdout)
	if err != nil {
		return "", fmt.Errorf("%s: stdout is not a JSON document: %w", a.Op, err)
	}
	vals, err := a.jp.eval(doc)
	if err != nil {
		return "", fmt.Errorf("%s: %w", a.Op, err)
	}
	s, ok := vals[0].(string)
	if !ok {
		return "", fmt.Errorf("%s: %s is %s, not a string", a.Op, a.Path, describeJSON(vals[0]))
	}
	return s, nil
}

// eval checks the assertion against a step's raw result. nil means it holds.
func (a *assertion) eval(cur *stepResult, env *evalEnv) error {
	switch a.Op {
	case "exit":
		if a.ExitEquals != nil {
			if cur.Exit != *a.ExitEquals {
				return fmt.Errorf("exit: got %d, want %d", cur.Exit, *a.ExitEquals)
			}
		} else if cur.Exit == 0 {
			return errors.New("exit: got 0, want nonzero")
		}
	case "stderr_code":
		got := stderrCode(cur.Stderr)
		if a.CodeEquals != "" && got != a.CodeEquals {
			return fmt.Errorf("stderr_code: got %q, want %q", got, a.CodeEquals)
		}
		if a.CodeNotEquals != "" && got == a.CodeNotEquals {
			return fmt.Errorf("stderr_code: got %q, want anything else", got)
		}
	case "stdout":
		return a.evalStdout(cur, env)
	case "contains", "not_contains":
		target, err := a.targetString(cur)
		if err != nil {
			return err
		}
		has := strings.Contains(target, a.Marker)
		if a.Op == "contains" && !has {
			return fmt.Errorf("contains: %s does not contain %q", a.where(), a.Marker)
		}
		if a.Op == "not_contains" && has {
			return fmt.Errorf("not_contains: %s contains %q", a.where(), a.Marker)
		}
	case "json_path":
		want, ok := env.Captures[a.EqualsCapture]
		if !ok {
			return fmt.Errorf("json_path: capture %q is not defined", a.EqualsCapture)
		}
		got, err := pathScalar(cur.Stdout, a.jp)
		if err != nil {
			return fmt.Errorf("json_path: %w", err)
		}
		if got != want {
			return fmt.Errorf("json_path: %s is %q, want the captured %s %q", a.Path, got, a.EqualsCapture, want)
		}
	case "id_set":
		return a.evalIDSet(cur)
	case "compare":
		return a.evalCompare(env)
	case "stderr_contains":
		token, ok := env.Captures[a.Capture]
		if !ok || token == "" {
			return fmt.Errorf("stderr_contains: capture %q is not defined", a.Capture)
		}
		if !bytes.Contains(cur.Stderr, []byte(token)) {
			return fmt.Errorf("stderr_contains: stderr does not contain the captured %s", a.Capture)
		}
	default:
		return fmt.Errorf("unknown op %q", a.Op)
	}
	return nil
}

func (a *assertion) where() string {
	if a.jp == nil {
		return "stdout"
	}
	return a.Path
}

func (a *assertion) evalStdout(cur *stepResult, env *evalEnv) error {
	switch {
	case a.Empty:
		if len(cur.Stdout) != 0 {
			return fmt.Errorf("stdout: got %d bytes, want none", len(cur.Stdout))
		}
	case a.EqualsInput != "":
		want, ok := env.Inputs[a.EqualsInput]
		if !ok {
			return fmt.Errorf("stdout: input %q is not declared", a.EqualsInput)
		}
		if !bytes.Equal(cur.Stdout, want) {
			return fmt.Errorf("stdout: got %q, want the bytes of input %s (%q)", clip(cur.Stdout), a.EqualsInput, clip(want))
		}
	case a.EqualsCapture != "":
		want, ok := env.Captures[a.EqualsCapture]
		if !ok {
			return fmt.Errorf("stdout: capture %q is not defined", a.EqualsCapture)
		}
		if string(cur.Stdout) != want {
			return fmt.Errorf("stdout: got %q, want the captured %s (%q)", clip(cur.Stdout), a.EqualsCapture, clip([]byte(want)))
		}
	}
	return nil
}

func (a *assertion) evalIDSet(cur *stepResult) error {
	doc, err := decodeJSON(cur.Stdout)
	if err != nil {
		return fmt.Errorf("id_set: stdout is not a JSON document: %w", err)
	}
	vals, err := a.jp.eval(doc)
	if err != nil {
		return fmt.Errorf("id_set: %w", err)
	}
	if !a.jp.hasWildcard() {
		arr, ok := vals[0].([]any)
		if !ok {
			return fmt.Errorf("id_set: %s is %s, not an array", a.Path, describeJSON(vals[0]))
		}
		vals = arr
	}
	got := make([]string, 0, len(vals))
	for _, v := range vals {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("id_set: %s holds %s, not an id string", a.Path, describeJSON(v))
		}
		got = append(got, s)
	}
	want := append([]string{}, a.IDs...)
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") || len(got) != len(want) {
		return fmt.Errorf("id_set: got %v, want %v (order ignored)", got, want)
	}
	return nil
}

func (a *assertion) evalCompare(env *evalEnv) error {
	left, err := a.Left.value(env)
	if err != nil {
		return fmt.Errorf("compare left: %w", err)
	}
	right, err := a.Right.value(env)
	if err != nil {
		return fmt.Errorf("compare right: %w", err)
	}
	if a.Relation == "equal" && left != right {
		return fmt.Errorf("compare: %s differs from %s (%q vs %q), want equal", a.Left.describe(), a.Right.describe(), clip([]byte(left)), clip([]byte(right)))
	}
	if a.Relation == "differ" && left == right {
		return fmt.Errorf("compare: %s equals %s (%q), want them to differ", a.Left.describe(), a.Right.describe(), clip([]byte(left)))
	}
	return nil
}

func (r ref) describe() string {
	if r.Path != "" {
		return r.Step + " " + r.Path
	}
	return r.Step + " " + r.Field
}

// clip keeps failure messages readable when a value is large.
func clip(b []byte) string {
	const limit = 80
	if len(b) <= limit {
		return string(b)
	}
	return string(b[:limit]) + fmt.Sprintf("...(%d bytes)", len(b))
}
