package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicaledit"
	"github.com/steveyegge/beads/internal/configfile"
)

func init() {
	for _, cmd := range []*cobra.Command{createCmd, updateCmd, deleteCmd, linkCreateCmd, linkUpdateCmd, linkDeleteCmd} {
		cmd.Flags().String("dry-run-mode", "", "Emit a canonical batch member without execution: batch (bdp requires an available native write route)")
		cmd.Flags().String("scope", "", "Scope for canonical offline authoring; defaults to BD_SCOPE or the selected graph workspace")
		if cmd == createCmd || cmd == linkCreateCmd {
			cmd.Flags().String("batch-bind", "", "Name this creation when emitting a batch member")
		}
		if cmd != deleteCmd && cmd != linkDeleteCmd {
			cmd.Flags().String("revision-provenance", "", "Canonical revision provenance input for member emission")
		}
		previousArgs, previousRun := cmd.Args, cmd.RunE
		cmd.Args = func(cmd *cobra.Command, args []string) error {
			if canonicalEmissionSelected(cmd) {
				return nil // The compiler validates the selected canonical signature.
			}
			if previousArgs != nil {
				return previousArgs(cmd, args)
			}
			return nil
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			if canonicalEmissionSelected(cmd) {
				return runCanonicalMemberEmission(cmd, args)
			}
			return previousRun(cmd, args)
		}
	}
	batchCmd.Flags().String("input-format", "", "Canonical batch input: jsonl or native json; omitted retains the legacy command-text runner outside graph mode")
	batchCmd.Flags().String("batch-version", "", "Explicit canonical JSONL encoder version (currently experimental 2)")
	batchCmd.Flags().StringArray("batch-arg", nil, "Supply an external canonical batch Reference argument NAME=VALUE")
	batchCmd.Flags().String("scope", "", "Scope for canonical authoring; defaults to BD_SCOPE or the selected graph workspace")
	batchCmd.Flags().String("dry-run-mode", "", "Emit the compiled canonical batch request without executing it")
	previous := batchCmd.RunE
	batchCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if canonicalBatchSelected(cmd) {
			return runCanonicalBatch(cmd)
		}
		return previous(cmd, args)
	}
}

func canonicalEmissionSelected(cmd *cobra.Command) bool {
	return cmd.Flags().Lookup("dry-run-mode") != nil && cmd.Flags().Changed("dry-run-mode")
}

func canonicalBatchSelected(cmd *cobra.Command) bool {
	return cmd == batchCmd && (graphPreviewActive || canonicalEmissionSelected(cmd) || cmd.Flags().Changed("scope") || cmd.Flags().Changed("input-format") || cmd.Flags().Changed("batch-version") || cmd.Flags().Changed("batch-arg") || cmd.Flags().Changed("retry-file") || cmd.Flags().Changed("idempotency-key"))
}

// Offline authoring exits the root provider gate before any database is opened.
// A successful emission is an unexecuted request, never a mutation result.
func admitCanonicalAuthoring(cmd *cobra.Command) (bool, error) {
	canonicalBatchRetry = nil
	if cmd == batchCmd {
		if err := canonicalScalarOccurrences(cmd, os.Args[1:], "scope", "input-format", "batch-version", "file", "dry-run-mode", "idempotency-key", "retry-file"); err != nil {
			return true, graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	if cmd == batchCmd && cmd.Flags().Changed("retry-file") {
		if err := canonicalAuthoringFlags(cmd, "retry-file", "output-format", "omit", "include", "json"); err != nil {
			return true, graphFailure("invalid_selector", err.Error(), 2)
		}
		file, _ := cmd.Flags().GetString("retry-file")
		invocation, err := readCanonicalBatchInvocation(file)
		if err != nil {
			return true, graphFailure("invalid_selector", err.Error(), 2)
		}
		canonicalBatchRetry = invocation
	}
	if cmd == initCmd {
		if enabled, _ := cmd.Flags().GetBool("atomic-batch"); enabled {
			mode, _ := cmd.Flags().GetString("graph-mode")
			if mode != "link" && os.Getenv("BD_GRAPH_MODE") != "link" {
				return true, graphFailure("invalid_selector", "--atomic-batch requires fresh --graph-mode link initialization", 2)
			}
		}
	}
	if canonicalEmissionSelected(cmd) {
		graphPreviewActive, graphPreviewConfig, graphPreviewDir = false, nil, ""
		return true, nil
	}
	for _, name := range []string{"batch-bind", "revision-provenance", "scope"} {
		if cmd != batchCmd && cmd.Flags().Lookup(name) != nil && cmd.Flags().Changed(name) {
			return true, graphFailure("capability_unavailable", "--"+name+" currently requires canonical member emission; ordinary execution has not adopted that input", 5)
		}
	}
	return false, nil
}

func canonicalAuthoringScope(cmd *cobra.Command) (string, error) {
	scope, _ := cmd.Flags().GetString("scope")
	if scope == "" {
		scope = os.Getenv("BD_SCOPE")
	}
	if scope == "" {
		dir, err := graphCandidateDir(cmd)
		if err != nil {
			return "", err
		}
		if dir != "" {
			cfg, err := configfile.LoadForDiscovery(dir)
			if err != nil {
				return "", err
			}
			if cfg != nil {
				scope = cfg.GraphScopeURL
			}
		}
	}
	if scope != "" {
		if err := graph.ValidateScopeURL(scope); err != nil {
			return "", err
		}
	}
	return scope, nil
}

type canonicalInputs struct {
	stdin io.Reader
	used  bool
}

func (i *canonicalInputs) read(value string) (json.RawMessage, error) {
	var raw []byte
	if strings.HasPrefix(value, "@") {
		var source io.Reader
		if value == "@-" {
			if i.used {
				return nil, fmt.Errorf("only one input may consume stdin")
			}
			i.used, source = true, i.stdin
		} else {
			file, err := os.Open(strings.TrimPrefix(value, "@")) // #nosec G304 -- explicit CLI value-file input
			if err != nil {
				return nil, err
			}
			defer file.Close()
			source = file
		}
		var err error
		raw, err = io.ReadAll(io.LimitReader(source, canonicalbatch.MaxInputBytes+1))
		if err != nil {
			return nil, err
		}
	} else {
		raw = []byte(value)
	}
	if len(raw) > canonicalbatch.MaxInputBytes {
		return nil, canonicalbatch.ErrLimit
	}
	return graph.CanonicalizeJSON(raw)
}

func canonicalObjectEdit(cmd *cobra.Command, plane string, create bool, input *canonicalInputs) (string, json.RawMessage, error) {
	singular := "property"
	if plane == "metadata" {
		singular = "metadata"
	}
	merge, replace, patch := cmd.Flags().Changed(plane), cmd.Flags().Changed("replace-"+plane), cmd.Flags().Changed("patch-"+plane)
	keyed := cmd.Flags().Changed("set-"+singular) || cmd.Flags().Changed("unset-"+singular)
	count := 0
	for _, yes := range []bool{merge, replace, patch, keyed} {
		if yes {
			count++
		}
	}
	if count > 1 || create && (replace || patch || keyed) {
		if count == 1 && create {
			return "", nil, fmt.Errorf("creation accepts only initial %s", plane)
		}
		return "", nil, fmt.Errorf("conflicting %s edit modes", plane)
	}
	if count == 0 {
		return "", nil, nil
	}
	if keyed {
		sets, _ := cmd.Flags().GetStringArray("set-" + singular)
		unsets, _ := cmd.Flags().GetStringArray("unset-" + singular)
		raw, err := canonicaledit.KeyedOperations(sets, unsets)
		return plane + "Change", raw, err
	}
	flag := plane
	if replace {
		flag = "replace-" + plane
	}
	if patch {
		flag = "patch-" + plane
	}
	value, _ := cmd.Flags().GetString(flag)
	raw, err := input.read(value)
	if err != nil {
		return "", nil, err
	}
	if merge && !create {
		raw, err = canonicaledit.MergeOperations(raw)
		return plane + "Change", raw, err
	}
	if patch {
		return plane + "Change", raw, nil
	}
	return plane, raw, nil
}

func runCanonicalMemberEmission(cmd *cobra.Command, args []string) error {
	if err := canonicalScalarOccurrences(cmd, os.Args[1:], "scope", "type", "id", "if-revision", "dry-run-mode", "batch-bind", "properties", "metadata", "replace-properties", "replace-metadata", "patch-properties", "patch-metadata", "revision-provenance"); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if cmd == linkCreateCmd {
		if err := canonicalExactOccurrences(cmd, os.Args[1:], "from", "to"); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	mode, _ := cmd.Flags().GetString("dry-run-mode")
	if mode != "batch" {
		if mode != "bdp" {
			return graphFailure("invalid_selector", "--dry-run-mode must be batch or bdp", 2)
		}
		return graphFailure("capability_unavailable", "BDP request preview requires an actual discovered native write route; only batch member emission is available here", 5)
	}
	scope, err := canonicalAuthoringScope(cmd)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	raw, err := canonicalCompileCRUDMember(cmd, args, scope)
	if err != nil {
		if errors.Is(err, canonicalbatch.ErrUnsupported) {
			return graphFailure("capability_unavailable", err.Error(), 5)
		}
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	return writeCanonicalDocument(cmd.OutOrStdout(), raw)
}

func canonicalCompileCRUDMember(cmd *cobra.Command, args []string, scope string) ([]byte, error) {
	member := map[string]any{}
	create := cmd == createCmd || cmd == linkCreateCmd
	link := cmd == linkCreateCmd || cmd == linkUpdateCmd || cmd == linkDeleteCmd
	deleted := cmd == deleteCmd || cmd == linkDeleteCmd
	allowed := []string{"dry-run-mode", "scope"}
	if create {
		allowed = append(allowed, "id", "id-root", "batch-bind")
		if link {
			allowed = append(allowed, "from", "to")
		}
	} else {
		allowed = append(allowed, "if-revision")
	}
	if deleted {
		allowed = append(allowed, "force")
	} else {
		allowed = append(allowed, "type", "properties", "metadata", "revision-provenance")
		if !create {
			allowed = append(allowed, "replace-properties", "replace-metadata", "patch-properties", "patch-metadata", "set-property", "unset-property", "set-metadata", "unset-metadata")
		}
	}
	if err := canonicalAuthoringFlags(cmd, allowed...); err != nil {
		return nil, err
	}
	operation := "updateBead"
	if link {
		operation = "updateLink"
	}
	if create {
		if link {
			operation = "createLink"
		} else {
			operation = "createBead"
		}
	}
	if deleted {
		if link {
			operation = "deleteLink"
		} else {
			operation = "deleteBead"
		}
	}
	member["operation"] = operation
	if create {
		if cmd.Flags().Changed("id-root") {
			return nil, fmt.Errorf("%w: offline emission cannot allocate --id-root; supply --id or omit allocation", canonicalbatch.ErrUnsupported)
		}
		if cmd.Flags().Changed("id") {
			value, _ := cmd.Flags().GetString("id")
			member["id"] = value
		}
		if cmd.Flags().Changed("batch-bind") {
			value, _ := cmd.Flags().GetString("batch-bind")
			member["name"] = value
		}
		if !cmd.Flags().Changed("type") {
			return nil, fmt.Errorf("create requires explicit --type")
		}
		if link {
			source, target, err := graphPreviewAuthoringEndpoints(cmd, args)
			if err != nil {
				return nil, err
			}
			for _, endpoint := range []struct{ key, value string }{{"source", source}, {"target", target}} {
				ref, err := canonicalbatch.ParseOperand(endpoint.value, scope)
				if err != nil {
					return nil, err
				}
				member[endpoint.key] = ref
			}
		} else if len(args) != 0 {
			return nil, fmt.Errorf("canonical create takes properties through flags, not positional arguments")
		}
	} else {
		if len(args) != 1 {
			return nil, fmt.Errorf("canonical mutation requires one Resource Reference")
		}
		ref, err := canonicalbatch.ParseOperand(args[0], scope)
		if err != nil {
			return nil, err
		}
		if ref.Revision != "" {
			return nil, fmt.Errorf("mutation subject must be unpinned; use --if-revision")
		}
		key := "bead"
		if link {
			key = "link"
		}
		member[key] = ref
	}
	if deleted {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return nil, fmt.Errorf("delete member emission requires --force apply intent")
		}
	} else {
		if cmd.Flags().Changed("type") {
			value, _ := cmd.Flags().GetString("type")
			ref, err := canonicalbatch.ParseReference(value, scope)
			if err != nil {
				return nil, err
			}
			member["type"] = ref
		}
		input := &canonicalInputs{stdin: cmd.InOrStdin()}
		for _, plane := range []string{"properties", "metadata"} {
			key, value, err := canonicalObjectEdit(cmd, plane, create, input)
			if err != nil {
				return nil, err
			}
			if key != "" {
				member[key] = value
			}
		}
		if cmd.Flags().Changed("revision-provenance") {
			value, _ := cmd.Flags().GetString("revision-provenance")
			raw, err := input.read(value)
			if err != nil {
				return nil, err
			}
			member["revisionProvenance"] = raw
		}
	}
	if cmd.Flags().Lookup("if-revision") != nil && cmd.Flags().Changed("if-revision") {
		value, _ := cmd.Flags().GetString("if-revision")
		if !strings.HasPrefix(value, "=") || len(value) == 1 {
			return nil, fmt.Errorf("%w: offline guards require =TOKEN; local revision numbers need a checked source", canonicalbatch.ErrUnsupported)
		}
		member["expectedRevision"] = strings.TrimPrefix(value, "=")
	}
	raw, err := json.Marshal(member)
	if err != nil {
		return nil, err
	}
	_, emitted, err := canonicalbatch.CompileMember(raw, scope)
	return emitted, err
}

func runCanonicalBatch(cmd *cobra.Command) error {
	if canonicalBatchRetry != nil {
		batch, err := canonicalbatch.Compile(canonicalBatchRetry.Request, canonicalbatch.Options{Format: "json", Scope: canonicalBatchRetry.Scope})
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		return executeCanonicalBatch(cmd, batch)
	}
	if err := canonicalAuthoringFlags(cmd, "file", "input-format", "batch-version", "batch-arg", "scope", "dry-run-mode", "idempotency-key", "output-format", "include", "omit", "json"); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	mode, _ := cmd.Flags().GetString("dry-run-mode")
	if mode != "" && mode != "batch" {
		if mode != "bdp" {
			return graphFailure("invalid_selector", "--dry-run-mode must be batch or bdp", 2)
		}
		return graphFailure("capability_unavailable", "BDP preview requires a discovered native atomic-batch route", 5)
	}
	if mode != "" && cmd.Flags().Changed("idempotency-key") {
		return graphFailure("invalid_selector", "emission does not reserve an execution key", 2)
	}
	if mode != "" {
		if err := canonicalAuthoringFlags(cmd, "file", "input-format", "batch-version", "batch-arg", "scope", "dry-run-mode"); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	file, _ := cmd.Flags().GetString("file")
	if file == "" {
		return graphFailure("invalid_selector", "canonical batch requires --file (use - for stdin)", 2)
	}
	var source io.Reader = cmd.InOrStdin()
	if file != "-" {
		f, err := os.Open(file) // #nosec G304 -- explicit batch input
		if err != nil {
			return err
		}
		defer f.Close()
		source = f
	}
	raw, err := io.ReadAll(io.LimitReader(source, canonicalbatch.MaxInputBytes+1))
	if err != nil {
		return err
	}
	if len(raw) > canonicalbatch.MaxInputBytes {
		return graphFailure("invalid_selector", canonicalbatch.ErrLimit.Error(), 2)
	}
	scope, err := canonicalAuthoringScope(cmd)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	format, _ := cmd.Flags().GetString("input-format")
	version, _ := cmd.Flags().GetString("batch-version")
	args, _ := cmd.Flags().GetStringArray("batch-arg")
	batch, err := canonicalbatch.Compile(raw, canonicalbatch.Options{Format: format, Version: version, Args: args, Scope: scope})
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if mode == "" {
		return executeCanonicalBatch(cmd, batch)
	}
	return writeCanonicalDocument(cmd.OutOrStdout(), append(batch.NativeJSON(), '\n'))
}

func canonicalAuthoringFlags(cmd *cobra.Command, names ...string) error {
	allowed := make(map[string]bool, len(names))
	for _, name := range names {
		allowed[name] = true
	}
	var unsupported string
	check := func(flag *pflag.Flag) {
		if !allowed[flag.Name] && unsupported == "" {
			unsupported = flag.Name
		}
	}
	cmd.Flags().Visit(check)
	cmd.InheritedFlags().Visit(check)
	if unsupported != "" {
		return fmt.Errorf("--%s is not supported in canonical authoring", unsupported)
	}
	return nil
}
