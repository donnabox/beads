package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/steveyegge/beads/cmd/bd/protocol"
)

// capturedValue is a per-workspace random value a step captured. Only token
// and id captures are registered; a stable capture is compared as it is.
type capturedValue struct {
	Kind  captureKind
	Value string
}

// normalizer turns one workspace's raw transcript into something two
// workspaces can be compared on, without hiding what differs:
//
//   - the workspace root becomes <WS>;
//   - JSON stdout is canonicalized (sorted keys, id-sorted object arrays,
//     RFC3339 -> <TS>, commit dropped) by protocol.CanonicalizeJSON;
//   - every per-workspace random value (version/revision tokens, captured ids)
//     becomes an ordinal by first appearance: <TOKEN#1>, <TOKEN#2>, <ID#1>.
//     Equal values keep equal placeholders and distinct values stay distinct.
//     A blanket <VERSION> would hide the bug where two versions collapse into
//     one, which is exactly what a scenario driver is for.
//
// A normalizer is stateful: ordinals run across the whole transcript, so the
// steps of one workspace must go through one normalizer, in order.
type normalizer struct {
	roots   []string
	ordinal map[string]string // registered value -> placeholder
	order   []string          // registered values, in registration order
	next    map[captureKind]int
}

func newNormalizer(roots ...string) *normalizer {
	n := &normalizer{ordinal: map[string]string{}, next: map[captureKind]int{}}
	seen := map[string]bool{}
	for _, root := range roots {
		if root == "" {
			continue
		}
		spellings := []string{root}
		if resolved, err := filepath.EvalSymlinks(root); err == nil {
			spellings = append(spellings, resolved)
		}
		for _, s := range spellings {
			if !seen[s] {
				seen[s] = true
				n.roots = append(n.roots, s)
			}
		}
	}
	return n
}

func (n *normalizer) register(kind captureKind, value string) {
	if value == "" {
		return // "" would match everywhere
	}
	if _, ok := n.ordinal[value]; ok {
		return
	}
	label := "TOKEN"
	if kind == kindID {
		label = "ID"
	}
	n.next[kind]++
	n.ordinal[value] = fmt.Sprintf("<%s#%d>", label, n.next[kind])
	n.order = append(n.order, value)
}

// replacer masks the workspace root and every registered value in one pass, so
// a placeholder is never rescanned as if it were data. Longer matches win.
func (n *normalizer) replacer() *strings.Replacer {
	pairs := make([]string, 0, 2*(len(n.roots)+len(n.order)))
	roots := append([]string(nil), n.roots...)
	sort.SliceStable(roots, func(i, j int) bool { return len(roots[i]) > len(roots[j]) })
	for _, r := range roots {
		pairs = append(pairs, r, "<WS>")
	}
	values := append([]string(nil), n.order...)
	sort.SliceStable(values, func(i, j int) bool { return len(values[i]) > len(values[j]) })
	for _, v := range values {
		pairs = append(pairs, v, n.ordinal[v])
	}
	return strings.NewReplacer(pairs...)
}

// normalize returns the comparable form of a step. captured are the values the
// step captured, in capture-name order; they are registered after the tokens
// found in stdout so the ordinal order is a function of the transcript alone.
func (n *normalizer) normalize(r *stepResult, captured []capturedValue) *stepResult {
	stdout := r.Stdout
	if canon, tree, ok := canonicalObject(r.Stdout); ok {
		stdout = canon
		n.registerTokens(tree)
	}
	for _, c := range captured {
		if c.Kind == kindToken || c.Kind == kindID {
			n.register(c.Kind, c.Value)
		}
	}
	rep := n.replacer()
	out := &stepResult{Name: r.Name, Exit: r.Exit}
	if r.Argv != nil {
		out.Argv = make([]string, len(r.Argv))
		for i, a := range r.Argv {
			out.Argv[i] = rep.Replace(a)
		}
	}
	out.Stdout = []byte(rep.Replace(string(stdout)))
	out.Stderr = []byte(rep.Replace(string(r.Stderr)))
	return out
}

var (
	tokenKeys = map[string]bool{"version": true, "revision": true, "previousVersion": true}
	tokenRE   = regexp.MustCompile(`^[0-9a-f]{32}$`)
)

// registerTokens registers the version/revision tokens of a canonical tree in
// canonical order (sorted keys, id-sorted arrays), so first-appearance is the
// same for any two workspaces that hold the same structure.
func (n *normalizer) registerTokens(v any) {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if s, ok := t[k].(string); ok && tokenKeys[k] && tokenRE.MatchString(s) {
				n.register(kindToken, s)
			}
			n.registerTokens(t[k])
		}
	case []any:
		for _, c := range t {
			n.registerTokens(c)
		}
	}
}

// shieldPrefix cannot appear in a key bd prints.
const shieldPrefix = "\x00"

// protocolMaskedKeys are the keys protocol.CanonicalizeJSON blanket-masks by
// bare name whatever they hold (revision -> <REVISION>; version -> <VERSION>).
// Left visible to it, two distinct tokens would collapse into one placeholder
// and this driver could never see a bug that merges two versions. They are
// renamed out of its way for the call and restored afterwards.
var protocolMaskedKeys = []string{"revision", "version"}

func rekey(v any, rename func(string) (string, bool)) {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		for _, k := range keys {
			rekey(t[k], rename)
			if to, ok := rename(k); ok {
				t[to] = t[k]
				delete(t, k)
			}
		}
	case []any:
		for _, c := range t {
			rekey(c, rename)
		}
	}
}

func shieldKey(k string) (string, bool) {
	for _, masked := range protocolMaskedKeys {
		if k == masked {
			return shieldPrefix + k, true
		}
	}
	return "", false
}

func unshieldKey(k string) (string, bool) {
	if strings.HasPrefix(k, shieldPrefix) {
		return strings.TrimPrefix(k, shieldPrefix), true
	}
	return "", false
}

// canonicalObject canonicalizes stdout when it is one JSON object, which every
// bd envelope is. Anything else (recall's raw body bytes, plain text) is left
// untouched and reported as not ok.
func canonicalObject(raw []byte) (canonical []byte, tree any, ok bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, nil, false
	}
	doc, err := decodeJSON(trimmed)
	if err != nil {
		return nil, nil, false
	}
	rekey(doc, shieldKey)
	shielded, err := json.Marshal(doc)
	if err != nil {
		return nil, nil, false
	}
	canon, err := protocol.CanonicalizeJSON(shielded)
	if err != nil {
		return nil, nil, false
	}
	back, err := decodeJSON(canon)
	if err != nil {
		return nil, nil, false
	}
	rekey(back, unshieldKey)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(back); err != nil {
		return nil, nil, false
	}
	return buf.Bytes(), back, true
}

// frame writes b with a length prefix, so where one field ends and the next
// begins is part of what is hashed.
func frame(h hash.Hash, b []byte) {
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], uint64(len(b))) //nolint:gosec // G115: len is never negative
	_, _ = h.Write(n[:])
	_, _ = h.Write(b)
}

// stepDigest is the sha256 of one normalized step: argv, exit, stdout, stderr.
func stepDigest(r *stepResult) string {
	argv := r.Argv
	if argv == nil {
		argv = []string{}
	}
	encoded, _ := json.Marshal(argv)
	h := sha256.New()
	frame(h, encoded)
	frame(h, []byte(strconv.Itoa(r.Exit)))
	frame(h, r.Stdout)
	frame(h, r.Stderr)
	return hex.EncodeToString(h.Sum(nil))
}

// transcriptDigest is the sha256 over a workspace's normalized steps, names
// included. Two workspaces are equal exactly when their digests are.
func transcriptDigest(steps []*stepResult) string {
	h := sha256.New()
	for _, s := range steps {
		frame(h, []byte(s.Name))
		frame(h, []byte(stepDigest(s)))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// checkRecordedAtOrder asserts that the states a workspace wrote were stamped
// in non-decreasing order, even though the timestamps themselves are masked
// for the byte comparison. Only writes that created a new state take part:
// a create (result.attribution) or an update that changed something
// (result.memory.attribution). forget and list report the stamp of an older
// write, and replaced.attribution is the predecessor's, so none of those say
// anything about the order of writes.
func checkRecordedAtOrder(steps []*stepResult) error {
	var (
		prev     time.Time
		prevName string
	)
	for _, s := range steps {
		ts, ok := writtenAt(s.Stdout)
		if !ok {
			continue
		}
		if !prev.IsZero() && ts.Before(prev) {
			return fmt.Errorf("attribution.recordedAt went backwards: step %q wrote %s, before step %q at %s",
				s.Name, ts.Format(time.RFC3339Nano), prevName, prev.Format(time.RFC3339Nano))
		}
		prev, prevName = ts, s.Name
	}
	return nil
}

func writtenAt(stdout []byte) (time.Time, bool) {
	doc, err := decodeJSON(bytes.TrimSpace(stdout))
	if err != nil {
		return time.Time{}, false
	}
	obj, _ := doc.(map[string]any)
	result, _ := obj["result"].(map[string]any)
	var attribution any
	if a, ok := result["attribution"]; ok {
		attribution = a
	} else if changed, _ := result["changed"].(bool); changed {
		memory, _ := result["memory"].(map[string]any)
		attribution = memory["attribution"]
	}
	fields, _ := attribution.(map[string]any)
	stamp, _ := fields["recordedAt"].(string)
	ts, err := time.Parse(time.RFC3339Nano, stamp)
	if err != nil {
		return time.Time{}, false
	}
	return ts, true
}
