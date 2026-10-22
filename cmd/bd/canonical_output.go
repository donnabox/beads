package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Register only on adopting commands. StringArray retains every scalar occurrence
// so contradictory formats are refused rather than silently taking the last one.
func registerCanonicalOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringArray("output-format", nil, "Canonical output: default, detail or json (cannot combine with legacy --json or --format)")
	cmd.Flags().StringArray("include", nil, "Include a JSON Pointer; repeat to union selections")
	cmd.Flags().StringArray("omit", nil, "Omit an optional JSON Pointer, or envelope for raw JSON; repeatable")
}

// Adapters register the same result kind for failures that happen before reads.
// The CLI command tree is constructed once and executes one command at a time.
var canonicalFailureFactories = map[*cobra.Command]func(clioutput.Options) (canonicalFailureRenderer, error){}

func canonicalFailureRendererFor(cmd *cobra.Command) canonicalFailureRenderer {
	opts, selected, err := canonicalOutputOptions(cmd)
	if err != nil || !selected {
		return nil
	}
	// Invalid masks cannot suppress a failure document. Only the valid raw-output
	// request survives; formats must still resolve without ambiguity.
	opts.Include = nil
	var omit []string
	for _, path := range opts.Omit {
		if path == "envelope" {
			omit = append(omit, path)
		}
	}
	opts.Omit = omit
	if factory := canonicalFailureFactories[cmd]; factory != nil {
		r, err := factory(opts)
		if err != nil {
			return nil
		}
		return r
	}
	r, err := clioutput.New(canonicalResourceDescriptor(), opts)
	if err != nil {
		return nil
	}
	return r
}

func admitCanonicalGraphPreview(cmd *cobra.Command) (bool, error) {
	if !canonicalOutputSelected(cmd) {
		return admitGraphPreview(cmd)
	}
	previous := graphFailureReporter
	graphFailureReporter = func(code, message string, status int) error {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), code, message, status)
	}
	defer func() { graphFailureReporter = previous }()
	return admitGraphPreview(cmd)
}

const canonicalDefaultAnnotation = "bd:canonical_output_default"

func canonicalOutputDefault(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Annotations[canonicalDefaultAnnotation] == "1" {
			return true
		}
	}
	return false
}

func canonicalOutputSelected(cmd *cobra.Command) bool {
	return canonicalOutputDefault(cmd) || cmd.Flags().Changed("output-format")
}

// Legacy switches remain legacy even when their spelling mentions JSON. The
// explicit opt-in is also required for projection on a legacy command path.
func canonicalOutputOptions(cmd *cobra.Command) (clioutput.Options, bool, error) {
	var opts clioutput.Options
	if cmd.Flags().Lookup("output-format") == nil {
		return opts, false, nil
	}
	selected := canonicalOutputSelected(cmd)
	if !selected {
		var canonicalOnly string
		cmd.Flags().Visit(func(f *pflag.Flag) {
			if len(f.Annotations["canonical-output-only"]) != 0 && canonicalOnly == "" {
				canonicalOnly = f.Name
			}
		})
		if canonicalOnly != "" {
			return opts, false, fmt.Errorf("--%s requires --output-format on this legacy command path", canonicalOnly)
		}
		if cmd.Flags().Changed("include") || cmd.Flags().Changed("omit") {
			return opts, false, fmt.Errorf("--include and --omit require --output-format on this legacy command path")
		}
		return opts, false, nil
	}
	for _, flag := range []string{"json", "format"} {
		if flag == "json" && canonicalOutputDefault(cmd) {
			continue
		}
		if cmd.Flags().Changed(flag) || cmd.Root().PersistentFlags().Changed(flag) {
			return opts, true, fmt.Errorf("--output-format cannot be combined with legacy --%s; choose one output contract", flag)
		}
	}
	if canonicalOutputDefault(cmd) {
		if cmd.Flags().Lookup("json") != nil {
			opts.JSON, _ = cmd.Flags().GetBool("json")
		} else if cmd.Root().PersistentFlags().Lookup("json") != nil {
			opts.JSON, _ = cmd.Root().PersistentFlags().GetBool("json")
		}
	}
	opts.Formats, _ = cmd.Flags().GetStringArray("output-format")
	opts.Include, _ = cmd.Flags().GetStringArray("include")
	opts.Omit, _ = cmd.Flags().GetStringArray("omit")
	return opts, true, nil
}

func canonicalGraphFlags(cmd *cobra.Command, renderer canonicalFailureRenderer, extra ...string) error {
	allowed := map[string]bool{"output-format": true, "include": true, "omit": true, "graph-mode": true, "actor": true, "quiet": true, "no-color": true, "directory": true, "readonly": true}
	for _, name := range extra {
		allowed[name] = true
	}
	var unsupported string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if !allowed[f.Name] && unsupported == "" {
			unsupported = f.Name
		}
	})
	if unsupported != "" {
		return canonicalFailure(cmd, renderer, "capability_unavailable", "canonical output does not implement --"+unsupported, 5)
	}
	return nil
}

func canonicalResourceDescriptor() clioutput.Descriptor {
	return clioutput.Descriptor{BDPSpec: bdpwire.Pin,
		Protection: &clioutput.Protection{Required: []string{"id", "type", "revision"}},
		Expandable: []string{"/properties", "/metadata"},
	}
}

// This validation runs before workspace admission and before any provider read.
// Result-specific masks/capabilities are checked again by the actual adapter.
func validateCanonicalOutputInvocation(cmd *cobra.Command) error {
	_, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return canonicalFailure(cmd, nil, "invalid_selector", err.Error(), 2)
	}
	return nil
}

type canonicalFailureRenderer interface {
	RenderFailure(string, json.RawMessage) ([]byte, error)
}

func canonicalFailure(cmd *cobra.Command, renderer canonicalFailureRenderer, code, message string, exit int) error {
	if renderer != nil {
		problem, err := json.Marshal(map[string]any{"code": code, "message": message, "retryable": false})
		if err != nil {
			return err
		}
		data, err := renderer.RenderFailure("cli", problem)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			if err := writeCanonicalDocument(cmd.OutOrStdout(), data); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "output_error: %v\n", err)
				return &exitError{Code: 1}
			}
		}
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "%s: %s\n", code, message) //nolint:gosec // G705: CLI stderr diagnostics are not HTML.
	return &exitError{Code: exit}
}

func writeCanonicalDocument(out io.Writer, data []byte) error {
	n, err := out.Write(data)
	if err == nil && n != len(data) {
		return io.ErrShortWrite
	}
	return err
}

func renderCanonicalResult(cmd *cobra.Command, renderer *clioutput.Renderer, result any) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(result)
	var raw []byte
	if err == nil {
		raw, err = renderer.Render(buf.Bytes())
	}
	if err != nil {
		if errors.Is(err, clioutput.ErrInvalidProjection) {
			return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
		}
		return canonicalFailure(cmd, renderer, "output_error", err.Error(), 1)
	}
	if err := writeCanonicalDocument(cmd.OutOrStdout(), raw); err != nil {
		// A partially written document cannot be replaced by a failure envelope.
		fmt.Fprintf(cmd.ErrOrStderr(), "output_error: %v\n", err)
		return &exitError{Code: 1}
	}
	return nil
}

// Keep read refusal statuses aligned with the existing graph CLI. Local service
// errors are CLI problems, not invented BDP HTTP problems.
func canonicalReadError(cmd *cobra.Command, renderer canonicalFailureRenderer, err error, fallback string) error {
	code, status := fallback, 5
	switch {
	case errors.Is(err, graphstore.ErrVersionUnknown):
		code, status = "revision_unknown", 3
	case errors.Is(err, graphstore.ErrGone):
		code, status = "gone", 3
	case errors.Is(err, graphstore.ErrNotFound):
		code, status = "not_found", 3
	case errors.Is(err, graphstore.ErrInvalidStore):
		code = "invalid_store"
	case errors.Is(err, graphstore.ErrCapabilityUnavailable), errors.Is(err, graphstore.ErrLimitExceeded):
		code = "capability_unavailable"
	}
	return canonicalFailure(cmd, renderer, code, err.Error(), status)
}

// All reads and ordinary cleanup finish before a success document is emitted.
// The callback owns payload projection, but does not print or open another store.
func canonicalReadWithGraphStore(cmd *cobra.Command, renderer canonicalFailureRenderer,
	read func(context.Context, *graphstore.Store) (any, error), output func(any) error) error {
	ctx, cancel := context.WithTimeout(getRootContext(), 30*time.Second)
	defer cancel()
	store, err := graphstore.OpenExisting(ctx, graphOptions(graphPreviewConfig))
	if err != nil {
		return canonicalReadError(cmd, renderer, err, "graph_not_initialized")
	}
	defer func() { _ = store.Close() }()
	result, readErr := read(ctx, store)
	closeErr := store.Close()
	if readErr != nil {
		return canonicalReadError(cmd, renderer, errors.Join(readErr, closeErr), "route_unavailable")
	}
	if closeErr != nil {
		return canonicalReadError(cmd, renderer, closeErr, "route_unavailable")
	}
	return output(result)
}
