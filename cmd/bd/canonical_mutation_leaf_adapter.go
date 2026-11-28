package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
)

// Capture the SAME checked, prefix-stripped argv that ExecuteC will receive.
// Reuse the existing occurrence checker: parsing into its private values never
// calls an original setter or consumes JSON/stdin. Array edits remain repeatable.
func canonicalMutationLeafArgs(cmd *cobra.Command, path, argv []string) func(*cobra.Command, []string) error {
	selectedPath, checkedArgv := append([]string(nil), path...), append([]string(nil), argv...)
	return func(actual *cobra.Command, args []string) error {
		if actual != cmd {
			return canonicalMutationLeafUnavailable("canonical mutation command changed after preparation")
		}
		if err := admitCanonicalMutationLeaf(actual, selectedPath, args); err != nil {
			return err
		}
		if err := canonicalScalarOccurrences(actual, checkedArgv,
			"directory", "graph-mode", "readonly", "quiet", "no-color", "actor", "json",
			"scope", "dry-run-mode", "type", "id", "id-root", "if-revision", "batch-bind", "force",
			"properties", "metadata", "replace-properties", "patch-properties", "replace-metadata", "patch-metadata", "revision-provenance",
			"input-format", "batch-version", "file", "idempotency-key", "retry-file"); err != nil {
			return canonicalMutationLeafInvalid(err.Error())
		}
		d, _ := lookupCanonicalMutationLeaf(selectedPath)
		if d.Link && !d.Delete {
			if err := canonicalExactOccurrences(actual, checkedArgv, "from", "to"); err != nil {
				return canonicalMutationLeafInvalid(err.Error())
			}
		}
		return nil
	}
}

// Call from the canonical router's Args wrapper, after its prohibited-Value
// guard and before original compatibility Args or any pre-run/provider hook.
// It performs syntax/presence checks only: no file/stdin/config/store reads.
// Do not run compatibility Args afterwards: create's title inference and
// update's last-touched policy are different operations.
func admitCanonicalMutationLeaf(cmd *cobra.Command, path []string, args []string) error {
	d, ok := lookupCanonicalMutationLeaf(path)
	if !ok || cmd == nil {
		return canonicalMutationLeafUnavailable("unrecognized canonical mutation operation")
	}
	if err := d.admitPositionals(len(args)); err != nil {
		return err
	}
	if called := cmd.CalledAs(); called != "" && called != cmd.Name() {
		return canonicalMutationLeafInvalid("use the canonical command path rather than a compatibility alias")
	}
	allowed := make(map[string]bool, len(d.Flags))
	for _, name := range d.Flags {
		allowed[name] = true
	}
	var unsupported string
	visit := func(f *pflag.Flag) {
		if f.Changed && !allowed[f.Name] && unsupported == "" {
			unsupported = f.Name
		}
	}
	cmd.Flags().VisitAll(visit)
	cmd.InheritedFlags().VisitAll(visit)
	if unsupported != "" {
		return canonicalMutationLeafInvalid("--" + unsupported + " is not in this canonical operation's surface")
	}
	if cmd.Flags().Changed("dry-run-mode") {
		mode, _ := cmd.Flags().GetString("dry-run-mode")
		if mode != "batch" && mode != "bdp" {
			return canonicalMutationLeafInvalid("--dry-run-mode must be batch or bdp")
		}
	}
	if d.Batch {
		return admitCanonicalMutationBatch(cmd)
	}
	if d.Create {
		typ, _ := cmd.Flags().GetString("type")
		if !cmd.Flags().Changed("type") || typ == "" {
			return canonicalMutationLeafInvalid("create requires an explicit nonempty --type")
		}
		if cmd.Flags().Changed("id") && cmd.Flags().Changed("id-root") {
			return canonicalMutationLeafInvalid("choose --id or --id-root, not both")
		}
		if d.Link {
			from, to, err := graphPreviewAuthoringEndpoints(cmd, args)
			if err != nil {
				return canonicalMutationLeafInvalid(err.Error())
			}
			for _, endpoint := range []string{from, to} {
				if strings.HasPrefix(endpoint, "batch-ref::") {
					if !canonicalEmissionSelected(cmd) {
						return canonicalMutationLeafInvalid("a batch Reference requires member emission")
					}
					continue
				}
				ref, err := graphPreviewParseReference(endpoint)
				if err != nil {
					return canonicalMutationLeafInvalid(err.Error())
				}
				if ref.currentSelection() {
					return canonicalMutationLeafInvalid("an endpoint pin requires an explicit revision; omit the suffix for an unpinned endpoint")
				}
			}
		}
	} else {
		if strings.HasPrefix(args[0], "batch-ref::") {
			if !canonicalEmissionSelected(cmd) {
				return canonicalMutationLeafInvalid("a batch Reference requires member emission")
			}
		} else if _, err := graphPreviewWritableIdentity(args[0]); err != nil {
			return canonicalMutationLeafInvalid(err.Error())
		}
		if d.Link && !d.Delete {
			for _, name := range []string{"from", "to"} {
				if !cmd.Flags().Changed(name) {
					continue
				}
				value, _ := cmd.Flags().GetString(name)
				if strings.HasPrefix(value, "batch-ref::") {
					if !canonicalEmissionSelected(cmd) {
						return canonicalMutationLeafInvalid("a batch Reference requires member emission")
					}
					continue
				}
				ref, err := graphPreviewParseReference(value)
				if err != nil {
					return canonicalMutationLeafInvalid(err.Error())
				}
				if ref.currentSelection() {
					return canonicalMutationLeafInvalid("an endpoint pin requires an explicit revision; omit the suffix for an unpinned endpoint")
				}
			}
		}
	}
	if cmd.Flags().Changed("batch-bind") {
		mode, _ := cmd.Flags().GetString("dry-run-mode")
		if !canonicalEmissionSelected(cmd) || mode != "batch" {
			return canonicalMutationLeafInvalid("--batch-bind requires --dry-run-mode batch")
		}
	}
	for _, plane := range []string{"properties", "metadata"} {
		singular := "property"
		if plane == "metadata" {
			singular = "metadata"
		}
		modes := 0
		for _, flag := range []string{plane, "replace-" + plane, "patch-" + plane} {
			if cmd.Flags().Changed(flag) {
				modes++
			}
		}
		if cmd.Flags().Changed("set-"+singular) || cmd.Flags().Changed("unset-"+singular) {
			modes++
		}
		if modes > 1 {
			return canonicalMutationLeafInvalid(fmt.Sprintf("conflicting %s edit modes", plane))
		}
	}
	return nil
}

func admitCanonicalMutationBatch(cmd *cobra.Command) error {
	if cmd.Flags().Changed("retry-file") {
		allowed := map[string]bool{"retry-file": true, "output-format": true, "include": true, "omit": true, "json": true}
		var conflict string
		check := func(f *pflag.Flag) {
			if f.Changed && !allowed[f.Name] && conflict == "" {
				conflict = f.Name
			}
		}
		cmd.Flags().VisitAll(check)
		cmd.InheritedFlags().VisitAll(check)
		if conflict != "" {
			return canonicalMutationLeafInvalid("--retry-file cannot combine with --" + conflict)
		}
		file, _ := cmd.Flags().GetString("retry-file")
		if file == "" {
			return canonicalMutationLeafInvalid("--retry-file requires a nonempty filename")
		}
		return nil
	}
	if cmd.Flags().Changed("input-format") {
		format, _ := cmd.Flags().GetString("input-format")
		if format != "jsonl" && format != "json" {
			return canonicalMutationLeafInvalid("--input-format must be jsonl or json")
		}
	}
	if cmd.Flags().Changed("batch-version") {
		version, _ := cmd.Flags().GetString("batch-version")
		if version != canonicalbatch.Version {
			return canonicalMutationLeafUnavailable("requested canonical batch encoder is unsupported")
		}
	}
	file, _ := cmd.Flags().GetString("file")
	if file == "" {
		return canonicalMutationLeafInvalid("canonical batch requires --file (use - for stdin)")
	}
	return nil
}

// The root calls this immediately after ordinary graph admission and BEFORE
// its legacy store acquisition branch. Canonical execution never falls back to
// Issue-only CUD or the command-text batch runner on a dependency workspace.
// Offline emission has already exited through admitCanonicalAuthoring.
func requireCanonicalMutationGraphRoute(cmd *cobra.Command, path []string, graphAdmitted bool) error {
	if _, ok := lookupCanonicalMutationLeaf(path); !ok || cmd == nil {
		return canonicalMutationLeafUnavailable("unrecognized canonical mutation operation")
	}
	if canonicalEmissionSelected(cmd) || graphAdmitted {
		return nil
	}
	return canonicalMutationLeafUnavailable("canonical mutation execution requires its checked graph workspace; compatibility fallback is disabled")
}

// After the admission hooks above, execute the existing checked implementation.
// In particular, batch is unconditional: its input must never be reinterpreted
// as legacy command text merely because no optional selector flag was supplied.
// This preserves the existing local result and capability limitations; the
// profile owner must not advertise public Write/output/delivery completeness.
func canonicalMutationLeafExecutor(cmd *cobra.Command, path []string) (func(*cobra.Command, []string) error, error) {
	var expected *cobra.Command
	d, ok := lookupCanonicalMutationLeaf(path)
	if !ok {
		return nil, canonicalMutationLeafUnavailable("unrecognized canonical mutation operation")
	}
	switch d.Operation {
	case "createBead": expected = createCmd
	case "updateBead": expected = updateCmd
	case "deleteBead": expected = deleteCmd
	case "createLink": expected = linkCreateCmd
	case "updateLink": expected = linkUpdateCmd
	case "deleteLink": expected = linkDeleteCmd
	case "batch": expected = batchCmd
	}
	if cmd == nil || cmd != expected || cmd.RunE == nil {
		return nil, canonicalMutationLeafUnavailable("canonical mutation requires the existing registered command")
	}
	// Capture BEFORE the router replaces RunE. Calling cmd.RunE from the
	// wrapper would recurse; resolving it later would also borrow changed state.
	ordinaryRun := cmd.RunE
	return func(actual *cobra.Command, args []string) error {
		if actual != cmd {
			return canonicalMutationLeafUnavailable("canonical mutation command changed after preparation")
		}
		if err := admitCanonicalMutationLeaf(actual, d.Path, args); err != nil {
			return err
		}
		if err := requireCanonicalMutationGraphRoute(actual, d.Path, graphPreviewActive); err != nil {
			return err
		}
		if d.Batch {
			return runCanonicalBatch(actual)
		}
		if canonicalEmissionSelected(actual) {
			return runCanonicalMemberEmission(actual, args)
		}
		if d.Operation == "updateBead" {
			// The compat top-level update also accepts Links. Canonical update
			// cannot inherit that convenience. Backing kind is immutable, so a
			// checked resolution cannot race into a different kind before write.
			selectedPath, err := graphPreviewResolvedWritePath(args[0])
			if err != nil {
				return err
			}
			if err := graph.ValidateBeadPath(selectedPath); err != nil {
				return canonicalMutationLeafInvalid("update requires a Bead; use link update for a Link")
			}
		}
		return ordinaryRun(actual, args)
	}, nil
}
