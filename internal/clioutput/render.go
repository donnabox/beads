package clioutput

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// SchemaVersion versions the CLI envelope and CLI-derived result shapes.
const SchemaVersion = 2

// ErrInvalidProjection distinguishes a user mask refusal from invalid provider
// data or an adapter descriptor bug. It may be discovered after payload read.
var ErrInvalidProjection = errors.New("invalid projection")

// Descriptor identifies the payload contract actually supplied by the adapter.
// It must not name a newer BDP contract while emitting an older DTO. Protection
// contains only semantic context; complete Resource validation belongs upstream.
type Descriptor struct {
	BDPSpec    string
	Protection *Protection
	// ResultKind identifies the CLI payload shape governed by schema_version.
	// Empty means resource. BDPSpec independently identifies contained BDP values;
	// it does not claim that a CLI-derived comparison is a BDP operation result.
	ResultKind string
	// Expandable lists JSON Pointer subtrees that admit nested field masks.
	// Whole root members are addressable by default; mandatory context still
	// cannot be removed. These are descriptor paths, not user wildcards.
	Expandable []string
}

// Options preserves occurrences so contradictory scalar flags can be rejected
// before any read. An empty Formats list means detail. This initial foundation
// admits only default, detail and json; other result kinds/formats need adapters.
type Options struct {
	Formats []string
	JSON    bool
	Include []string
	Omit    []string
}

// Renderer is immutable after New. It returns complete byte slices and never
// writes stdout/stderr, sets an exit code, or performs a read. Command adapters
// retain responsibility for legacy-mode dispatch, provider errors and IO errors.
type Renderer struct {
	bdpSpec    string
	resultKind string
	protection *Protection
	format     string
	raw        bool
	include    *mask
	omit       *mask
	projection *projectionMetadata
}

type projectionMetadata struct {
	Include []string `json:"include,omitempty"`
	Omit    []string `json:"omit,omitempty"`
}

type envelope struct {
	SchemaVersion int                 `json:"schema_version"`
	BDPSpec       string              `json:"bdp_spec"`
	ResultKind    string              `json:"result_kind"`
	Outcome       string              `json:"outcome"`
	Data          json.RawMessage     `json:"data,omitempty"`
	Projection    *projectionMetadata `json:"projection,omitempty"`
	Error         *problemWrapper     `json:"error,omitempty"`
}

type problemWrapper struct {
	Origin  string          `json:"origin"`
	Problem json.RawMessage `json:"problem"`
}

// New validates output options and snapshots the descriptor. Call it before
// invoking any operation, including when the operation may fail without data.
func New(descriptor Descriptor, options Options) (*Renderer, error) {
	if strings.TrimSpace(descriptor.BDPSpec) == "" {
		return nil, fmt.Errorf("an exact BDP payload contract identifier is required")
	}
	if descriptor.Protection == nil || len(descriptor.Protection.Required) == 0 {
		return nil, fmt.Errorf("a Resource protection descriptor is required")
	}
	kind := descriptor.ResultKind
	if kind == "" {
		kind = "resource"
	}
	switch kind {
	case "resource", "resource_collection", "version_collection", "history",
		"comparison", "graph", "rows", "aggregate", "mutation":
	default:
		return nil, fmt.Errorf("unsupported result kind %q", kind)
	}
	format, err := resolveFormat(options.Formats, options.JSON)
	if err != nil {
		return nil, err
	}
	omitPaths := make([]string, 0, len(options.Omit))
	raw := false
	for _, path := range options.Omit {
		if path == "envelope" {
			raw = true
		} else {
			omitPaths = append(omitPaths, path)
		}
	}
	if raw && format != "json" {
		return nil, fmt.Errorf("omit envelope requires JSON output")
	}
	include, err := parseMask(options.Include)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProjection, err)
	}
	omit, err := parseMask(omitPaths)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProjection, err)
	}
	if omit.terminal {
		return nil, fmt.Errorf("%w: cannot omit the Resource root", ErrInvalidProjection)
	}
	for _, key := range descriptor.Protection.Required {
		if omit.child(key).selectsAll() {
			return nil, fmt.Errorf("%w: cannot omit mandatory context %q", ErrInvalidProjection, pointerChild("", key))
		}
	}
	expandable, err := parseMask(descriptor.Expandable)
	if err != nil {
		return nil, err
	}
	if err := validateDepth(options.Include, expandable); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProjection, err)
	}
	if err := validateDepth(omitPaths, expandable); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProjection, err)
	}
	protection, err := cloneProtection(descriptor.Protection, make(map[*Protection]bool))
	if err != nil {
		return nil, err
	}
	var projection *projectionMetadata
	if len(options.Include) != 0 || len(omitPaths) != 0 {
		projection = &projectionMetadata{
			Include: canonicalPaths(options.Include),
			Omit:    canonicalPaths(omitPaths),
		}
	}
	return &Renderer{
		bdpSpec: descriptor.BDPSpec, resultKind: kind, protection: protection, format: format,
		raw: raw, include: include, omit: omit, projection: projection,
	}, nil
}

func resolveFormat(formats []string, shorthand bool) (string, error) {
	selected := ""
	for _, format := range formats {
		if format == "default" {
			format = "detail"
		}
		if format != "json" && format != "detail" {
			return "", fmt.Errorf("unsupported single-Resource output format %q", format)
		}
		if selected != "" && selected != format {
			return "", fmt.Errorf("conflicting output formats")
		}
		selected = format
	}
	if shorthand {
		if selected != "" && selected != "json" {
			return "", fmt.Errorf("JSON shorthand conflicts with output format")
		}
		selected = "json"
	}
	if selected == "" {
		selected = "detail"
	}
	return selected, nil
}

func canonicalPaths(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if !seen[path] {
			out = append(out, path)
			seen[path] = true
		}
	}
	sort.Strings(out)
	return out
}

// Render produces a complete successful document. Invalid input or a prohibited
// mask returns no bytes. The caller can render a separate failure document;
// there is never a partially serialized success prefix to retract.
func (r *Renderer) Render(resource json.RawMessage) ([]byte, error) {
	value := bytes.TrimSpace(resource)
	if !json.Valid(value) || value[0] != '{' {
		return nil, fmt.Errorf("single-Resource output requires a JSON object")
	}
	full := r.projection == nil || len(r.projection.Include) == 0
	projected, _, err := project(value, r.protection, r.include, r.omit, full, true, "")
	if err != nil {
		return nil, err
	}
	if r.format == "detail" {
		return renderDetail(projected)
	}
	if r.raw {
		return document(projected)
	}
	return document(envelope{
		SchemaVersion: SchemaVersion, BDPSpec: r.bdpSpec, ResultKind: r.resultKind,
		Outcome: "success", Data: projected, Projection: r.projection,
	})
}

// RenderFailure frames a read failure using the common error
// wrapper. The adapter supplies the original BDP problem or its chosen CLI
// problem object; this package invents neither codes nor retry/exit semantics.
// Command adapters preserve their established error codes. Human/raw modes
// return no stdout; diagnostics and a
// nonzero status remain the caller's responsibility. Mutation outcomes are not
// admitted by this single-Resource read renderer.
func (r *Renderer) RenderFailure(origin string, problem json.RawMessage) ([]byte, error) {
	if origin != "bdp" && origin != "cli" {
		return nil, fmt.Errorf("problem origin must be bdp or cli")
	}
	value := bytes.TrimSpace(problem)
	if !json.Valid(value) || value[0] != '{' {
		return nil, fmt.Errorf("problem must be a JSON object")
	}
	if origin == "cli" {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(value, &fields); err != nil {
			return nil, err
		}
		for _, name := range []string{"code", "message"} {
			var text string
			if err := json.Unmarshal(fields[name], &text); err != nil || strings.TrimSpace(text) == "" {
				return nil, fmt.Errorf("CLI problem requires nonempty string %s", name)
			}
		}
	}
	if r.raw || r.format != "json" {
		return nil, nil
	}
	return document(envelope{
		SchemaVersion: SchemaVersion, BDPSpec: r.bdpSpec, ResultKind: r.resultKind, Outcome: "failure",
		Projection: r.projection, Error: &problemWrapper{Origin: origin, Problem: value},
	})
}

// marshalValue preserves JSON string values without HTML-specific escaping.
// CLI output is not embedded in an HTML document.
func marshalValue(value any) ([]byte, error) {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(out.Bytes(), []byte("\n")), nil
}

func document(value any) ([]byte, error) {
	encoded, err := marshalValue(value)
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func renderDetail(raw json.RawMessage) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out bytes.Buffer
	for _, key := range keys {
		// JSON string encoding makes arbitrary field names and control characters
		// visible without letting them alter the terminal presentation.
		label, err := marshalValue(key)
		if err != nil {
			return nil, err
		}
		out.Write(label)
		out.WriteString(": ")
		// Protocol values may already contain HTML escapes in RawMessage fields.
		// Decode with UseNumber before human rendering so values stay exact while
		// literal text is readable; control characters remain JSON-escaped.
		decoder := json.NewDecoder(bytes.NewReader(fields[key]))
		decoder.UseNumber()
		var decoded any
		if err := decoder.Decode(&decoded); err != nil {
			return nil, err
		}
		encoded, err := marshalValue(decoded)
		if err != nil {
			return nil, err
		}
		var value bytes.Buffer
		if err := json.Indent(&value, encoded, "  ", "  "); err != nil {
			return nil, err
		}
		out.Write(value.Bytes())
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}
