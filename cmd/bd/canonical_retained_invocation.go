package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/config"
	"io"
	"sort"
)

const canonicalRetainedContract = "https://github.com/gastownhall/bdp/schemas/bdp-v0.1.schema.json"
const canonicalRetainedMedia = "application/vnd.bdp.v0.1+json"

type canonicalRetainedContextKey uint8

const (
	canonicalRetainedArgvKey canonicalRetainedContextKey = iota
	canonicalRetainedBindingsKey
	canonicalRetainedScopeKey
	canonicalRetainedSignalKey
)

type canonicalRetainedOperation string

const (
	retainedAtomicBatch canonicalRetainedOperation = "atomicBatch"
	retainedPointRead   canonicalRetainedOperation = "pointRead"
	retainedCreateBead  canonicalRetainedOperation = "createBead"
	retainedUpdateBead  canonicalRetainedOperation = "updateBead"
	retainedDeleteBead  canonicalRetainedOperation = "deleteBead"
	retainedCreateLink  canonicalRetainedOperation = "createLink"
	retainedUpdateLink  canonicalRetainedOperation = "updateLink"
	retainedDeleteLink  canonicalRetainedOperation = "deleteLink"
)

func canonicalRetainedOperationFor(cmd *cobra.Command) canonicalRetainedOperation {
	switch strings.Join(canonicalReadCommandPath(cmd, cmd.Root()), " ") {
	case "batch":
		return retainedAtomicBatch
	case "show":
		return retainedPointRead
	case "create":
		return retainedCreateBead
	case "update":
		return retainedUpdateBead
	case "delete":
		return retainedDeleteBead
	case "link create":
		return retainedCreateLink
	case "link update":
		return retainedUpdateLink
	case "link delete":
		return retainedDeleteLink
	}
	return ""
}

func canonicalRetainedSelected(cmd *cobra.Command) bool {
	if !canonicalCLIProfileSelected(cmd) || canonicalRetainedOperationFor(cmd) == "" {
		return false
	}
	if canonicalRetainedOperationFor(cmd) == retainedAtomicBatch {
		return canonicalRetainedBatchSelected(cmd)
	}
	if cmd.Flags().Lookup("dry-run-mode") != nil {
		mode, _ := cmd.Flags().GetString("dry-run-mode")
		if mode == "batch" {
			return false
		}
	}
	return true
}

// Extend existing spellings only on the adopting leaves. No backend,
// transport, command tree, credential or process-global provider is registered.
func init() {
	// These three selected Link update inputs are absent at the fixed producer.
	for _, name := range []string{"type", "from", "to"} {
		if linkUpdateCmd.Flags().Lookup(name) == nil {
			linkUpdateCmd.Flags().String(name, "", "Canonical retained Link update input")
		}
	}
	if linkDeleteCmd.Flags().Lookup("revision-provenance") == nil {
		linkDeleteCmd.Flags().String("revision-provenance", "", "Canonical retained deletion provenance")
	}
	for _, command := range []*cobra.Command{showCmd, createCmd, updateCmd, deleteCmd, linkCreateCmd, linkUpdateCmd, linkDeleteCmd, batchCmd} {
		cmd := command
		if cmd.Flags().Lookup("scope") == nil {
			cmd.Flags().String("scope", "", "Selected Scope for the canonical retained caller")
		}
		if cmd != showCmd {
			if cmd.Flags().Lookup("idempotency-key") == nil {
				cmd.Flags().String("idempotency-key", "", "Exact retained singleton key; preview does not reserve it")
			}
			if cmd.Flags().Lookup("retry-file") == nil {
				cmd.Flags().String("retry-file", "", "Recover one previously saved retained singleton invocation")
			}
		}
		if cmd != showCmd && cmd.Flags().Lookup("output-format") == nil {
			registerCanonicalOutputFlags(cmd)
		}
		previous := cmd.RunE
		cmd.RunE = func(actual *cobra.Command, args []string) error {
			if canonicalRetainedSelected(actual) {
				if canonicalRetainedOperationFor(actual) == retainedPointRead {
					return runCanonicalShow(actual, args, false)
				}
				return runCanonicalRetained(actual, args)
			}
			return previous(actual, args)
		}
	}
}

// Pure captured-argv admission runs at the very beginning of the actual root
// pre-run, before telemetry, metrics defaults, signals or configuration writes.
func validateCanonicalRetainedSyntax(cmd *cobra.Command) error {
	if !canonicalRetainedSelected(cmd) {
		return nil
	}
	argv, ok := cmd.Context().Value(canonicalRetainedArgvKey).([]string)
	if !ok {
		return graphFailure("capability_unavailable", "retained caller lacks the bound root invocation", 5)
	}
	if err := canonicalScalarOccurrences(cmd, argv, "scope", "type", "id", "if-revision", "dry-run-mode", "properties", "metadata", "replace-properties", "replace-metadata", "patch-properties", "patch-metadata", "revision-provenance", "idempotency-key", "retry-file", "from", "to"); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if canonicalRetainedOperationFor(cmd) == retainedAtomicBatch {
		return validateCanonicalRetainedBatchSyntax(cmd, argv)
	}
	if canonicalRetainedOperationFor(cmd) == retainedPointRead {
		if cmd.Flags().Changed("scope") {
			return graphFailure("invalid_selector", "direct absolute point read does not consume --scope", 2)
		}
		return nil
	}
	if cmd.Flags().Changed("idempotency-key") {
		key, _ := cmd.Flags().GetString("idempotency-key")
		if !canonicalRetainedKey(key) {
			return graphFailure("invalid_selector", "explicit retained idempotency key must match [A-Za-z0-9_-]{1,256}", 2)
		}
	}
	return nil
}

// The actual root attaches its existing signal context after signal setup.
// This records cancellation authority without replacing the leaf caller.
func bindCanonicalRetainedRootSignal(cmd *cobra.Command, signal context.Context) {
	if canonicalRetainedSelected(cmd) {
		cmd.SetContext(context.WithValue(cmd.Context(), canonicalRetainedSignalKey, signal))
	}
}
func canonicalRetainedExecutionContext(caller context.Context) (context.Context, func()) {
	signal, _ := caller.Value(canonicalRetainedSignalKey).(context.Context)
	if signal == nil {
		return caller, func() {}
	}
	joined, cancel := context.WithCancel(caller)
	var stop func() bool
	if signal.Err() != nil {
		cancel()
	} else {
		stop = context.AfterFunc(signal, cancel)
	}
	var deadlineCancel context.CancelFunc
	if deadline, ok := signal.Deadline(); ok {
		joined, deadlineCancel = context.WithDeadline(joined, deadline)
	}
	return joined, func() {
		if stop != nil {
			stop()
		}
		if deadlineCancel != nil {
			deadlineCancel()
		}
		cancel()
	}
}

// Called at the actual root pre-provider junction. Read explicit presence,
// including empty env/config, once. The detached selector itself is unchanged.
func admitCanonicalRetained(cmd *cobra.Command) (bool, error) {
	if !canonicalRetainedSelected(cmd) {
		// Newly registered retained-only spellings remain refusals in compatibility
		// execution; they must never be silently ignored by a legacy mutation.
		if canonicalRetainedOperationFor(cmd) != "" && canonicalRetainedOperationFor(cmd) != retainedAtomicBatch {
			names := []string{"idempotency-key", "retry-file"}
			if cmd == linkUpdateCmd {
				names = append(names, "type", "from", "to")
			}
			if cmd == linkDeleteCmd {
				names = append(names, "revision-provenance")
			}
			for _, name := range names {
				if cmd.Flags().Changed(name) {
					return true, graphFailure("capability_unavailable", "--"+name+" requires the canonical retained caller", 5)
				}
			}
		}
		return false, nil
	}
	if err := validateCanonicalRetainedSyntax(cmd); err != nil {
		return true, err
	}
	// Direct reads do not select or read mutation Scope/configuration.
	if canonicalRetainedOperationFor(cmd) == retainedPointRead {
		return true, nil
	}
	var input canonicalScopeContextInput
	if cmd.Flags().Changed("scope") {
		value, _ := cmd.Flags().GetString("scope")
		input.Explicit = &value
	}
	if value, present := os.LookupEnv("BDP_SCOPE_URL"); present {
		input.Environment = &value
	}
	if input.Explicit == nil && input.Environment == nil {
		value, present, err := config.UserYamlValueStrict("scope.url")
		if err != nil {
			return true, graphFailure("invalid_selector", "cannot read selected Scope configuration", 2)
		}
		if present {
			input.Configuration = &value
		}
	}
	selected, err := selectCanonicalScopeContext(input)
	if err != nil {
		return true, graphFailure("invalid_selector", err.Error(), 2)
	}
	if !selected.Present() {
		return true, graphFailure("capability_unavailable", "retained caller requires an explicitly selected Scope context", 5)
	}
	cmd.SetContext(context.WithValue(cmd.Context(), canonicalRetainedScopeKey, selected))
	return true, nil
}

func lookupCanonicalRetainedLeaf(path []string) (canonicalReadLeafDescriptor, bool) {
	name := strings.Join(path, " ")
	switch name {
	case "show", "create", "update", "delete", "link create", "link update", "link delete", "batch":
	default:
		return canonicalReadLeafDescriptor{}, false
	}
	flags := []string{"directory", "quiet", "no-color", "json", "output-format", "include", "omit"}
	max := 1
	if name == "batch" {
		return canonicalReadLeafDescriptor{Operation: canonicalReadLeafOperation("retained-singleton"), Path: append([]string(nil), path...), Flags: append(flags, "scope", "idempotency-key", "retry-file", "dry-run-mode", "input-format", "batch-version", "file"), MaximumPositionals: 0}, true
	}
	if name != "show" {
		flags = append(flags, "scope", "idempotency-key", "retry-file", "dry-run-mode")
		if name == "create" || name == "link create" {
			flags = append(flags, "id", "type", "properties", "metadata", "revision-provenance")
			max = 0
			if name == "link create" {
				max = 2
				flags = append(flags, "from", "to")
			}
		} else {
			flags = append(flags, "if-revision")
			if name == "delete" || name == "link delete" {
				flags = append(flags, "force")
				if name == "link delete" {
					flags = append(flags, "revision-provenance")
				}
			} else {
				flags = append(flags, "type", "properties", "metadata", "replace-properties", "replace-metadata", "patch-properties", "patch-metadata", "set-property", "unset-property", "set-metadata", "unset-metadata", "revision-provenance")
			}
		}
	}
	if name == "link update" {
		flags = append(flags, "from", "to")
	}
	return canonicalReadLeafDescriptor{Operation: canonicalReadLeafOperation("retained-singleton"), Path: append([]string(nil), path...), Flags: flags, MaximumPositionals: max}, true
}

type canonicalRetainedPrepared struct {
	Scope     string                     `json:"scope"`
	Contract  string                     `json:"contract"`
	Operation canonicalRetainedOperation `json:"operation"`
	Method    string                     `json:"method"`
	Target    string                     `json:"target"`
	Key       string                     `json:"key,omitempty"`
	Body      json.RawMessage            `json:"body,omitempty"`
}

// The endpoint and credential tuple is NOT derived from user tokens, the Issue
// http sidecar or an old fixture Host rewrite. Missing real bindings refuse.
func runCanonicalRetained(cmd *cobra.Command, args []string) error {
	originalContext := cmd.Context()
	executionContext, release := canonicalRetainedExecutionContext(originalContext)
	cmd.SetContext(executionContext)
	defer func() { release(); cmd.SetContext(originalContext) }()

	op := canonicalRetainedOperationFor(cmd)
	if op == retainedAtomicBatch {
		return runCanonicalRetainedBatch(cmd, args)
	}
	scope := ""
	if op != retainedPointRead {
		selected, ok := cmd.Context().Value(canonicalRetainedScopeKey).(canonicalScopeContext)
		if !ok {
			return graphFailure("capability_unavailable", "retained mutation root admission is missing", 5)
		}
		scope = selected.URL
	}
	// Production constructs existing per-target credential/configuration and
	// service bounds; the context override is only a supplied caller/test seam.
	// Authentication establishes server principal identity, not this local value.
	bindings, ok := cmd.Context().Value(canonicalRetainedBindingsKey).(canonicalRetainedBindings)
	if !ok {
		var bindErr error
		bindings, bindErr = newCanonicalRetainedRootBindings(cmd.Context(), op != retainedPointRead)
		if bindErr != nil {
			return graphFailure("capability_unavailable", bindErr.Error(), 5)
		}
	}
	if err := bindings.validateRead(); err != nil {
		return graphFailure("capability_unavailable", err.Error(), 5)
	}
	if op != retainedPointRead {
		if err := bindings.validate(scope); err != nil {
			return graphFailure("capability_unavailable", err.Error(), 5)
		}
	}
	opts, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if op != retainedPointRead {
		mode, _ := cmd.Flags().GetString("dry-run-mode")
		if mode != "" && mode != "bdp" {
			return graphFailure("invalid_selector", "retained singleton preview requires bdp mode", 2)
		}
	}
	retryPath := ""
	if op != retainedPointRead {
		retryPath, _ = cmd.Flags().GetString("retry-file")
	}
	var body json.RawMessage
	var subject string
	if op != retainedPointRead && cmd.Flags().Changed("retry-file") && retryPath == "" {
		return graphFailure("invalid_selector", "recovery requires an explicit nonempty file", 2)
	}
	if retryPath != "" {
		if len(args) != 0 {
			return graphFailure("invalid_selector", "recovery does not accept fresh operands", 2)
		}
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			switch flag.Name {
			case "retry-file", "scope", "output-format", "include", "omit", "json", "directory", "quiet", "no-color":
			default:
				err = fmt.Errorf("recovery does not accept fresh --%s", flag.Name)
			}
		})
	} else {
		body, subject, err = canonicalRetainedNativeInput(cmd, args, op, scope)
	}
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	ctx, cancel := context.WithDeadline(cmd.Context(), bindings.Deadline)
	defer cancel()
	if op == retainedPointRead {
		// readResource verifies explicit owning service-desc/contract/kind; Core-only
		// discovery is sufficient and no Write directory/defaultScope is requested.
		binding := canonicalRetainedReadBinding{options: bindings}
		reply := binding.readResource(ctx, subject)
		return completeCanonicalRetainedRead(cmd, opts, reply)
	}
	binding, err := discoverCanonicalRetainedWrite(ctx, scope, bindings)
	if err != nil {
		return graphFailure("capability_unavailable", "retained mutation discovery or navigation refused", 5)
	}
	request := canonicalRetainedPrepared{Scope: scope, Contract: canonicalRetainedContract, Operation: op, Method: http.MethodPost, Target: binding.targets[string(op)], Body: body}
	if op != retainedPointRead {
		request.Key, _ = cmd.Flags().GetString("idempotency-key")
		mode, _ := cmd.Flags().GetString("dry-run-mode")
		if mode == "bdp" {
			if retryPath != "" {
				return graphFailure("invalid_selector", "recovery cannot be combined with preview", 2)
			}
			if request.Key != "" && !canonicalRetainedKey(request.Key) {
				return graphFailure("invalid_selector", "invalid retained idempotency key", 2)
			}
			// This is a request preview, not a native operation result. No credentials,
			// generated/reserved key, invocation file or POST is produced.
			raw, marshalErr := canonicalMutationMarshalNative(struct {
				Method      string          `json:"method"`
				Target      string          `json:"target"`
				ContentType string          `json:"contentType"`
				Accept      string          `json:"accept"`
				Key         string          `json:"idempotencyKey,omitempty"`
				Body        json.RawMessage `json:"body"`
			}{request.Method, request.Target, canonicalRetainedMedia, canonicalRetainedMedia, request.Key, bytes.Clone(request.Body)})
			if marshalErr != nil {
				return marshalErr
			}
			return writeCanonicalDocument(cmd.OutOrStdout(), append(raw, '\n'))
		}

	}
	var completion *canonicalMutationOutput
	if op != retainedPointRead {
		completion, err = newCanonicalMutationOutput(canonicalMutationOutputRetained, clioutput.MutationDescriptor{BDPSpec: canonicalRetainedContract, Required: []string{"outcome", "changed"}}, opts, cmd.OutOrStdout())
		if err != nil {
			return err
		}
	}
	// One private orchestrator owns validate/key/save/fsync/submit. A leaf cannot
	// submit an unsaved or preview value. The persisted handle is private.
	var reply canonicalRetainedReply
	var recovery string
	if retryPath != "" {
		reply, recovery = binding.retry(ctx, retryPath, op)
	} else {
		reply, recovery = binding.apply(ctx, request)
	}
	if recovery != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "Retained recovery file: %s\n", recovery)
	}
	reply.validateMutation(op)
	result := completion.Complete(reply, "", reply.OperationErr, reply.CloseErr)
	if result.Err() != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), result.Err())
		return &exitError{Code: 1}
	}
	return nil
}

// Reuse only selected CLI syntax/edit admission; do not call the experimental
// member compiler or remove a discriminator from its version-2 result.
func canonicalRetainedNativeInput(cmd *cobra.Command, args []string, op canonicalRetainedOperation, scope string) (json.RawMessage, string, error) {
	ref := func(value string, pinned bool) (json.RawMessage, error) {
		r, err := canonicalbatch.ParseReference(value, scope)
		if err != nil {
			return nil, err
		}
		if !pinned && r.Revision != "" {
			return nil, fmt.Errorf("mutation subject must be unpinned; use --if-revision")
		}
		return json.Marshal(r)
	}
	if op == retainedPointRead {
		if len(args) != 1 {
			return nil, "", fmt.Errorf("point show requires one exact Resource Reference")
		}
		r, err := canonicalbatch.ParseReference(args[0], scope)
		if err != nil {
			return nil, "", err
		}
		if r.Revision != "" {
			return nil, "", fmt.Errorf("retained historical point reads are not implemented in this slice")
		}
		return nil, r.URI, nil
	}
	create := op == retainedCreateBead || op == retainedCreateLink
	link := op == retainedCreateLink || op == retainedUpdateLink || op == retainedDeleteLink
	deleted := op == retainedDeleteBead || op == retainedDeleteLink
	fields := make(map[string]json.RawMessage)
	if create {
		if !cmd.Flags().Changed("type") {
			return nil, "", fmt.Errorf("creation requires explicit --type")
		}
		if cmd.Flags().Changed("id") {
			id, _ := cmd.Flags().GetString("id")
			raw, err := ref(id, false)
			if err != nil {
				return nil, "", err
			}
			_ = raw
			fields["id"], _ = json.Marshal(id)
		}
		if link {
			source, target, err := graphPreviewAuthoringEndpoints(cmd, args)
			if err != nil {
				return nil, "", err
			}
			fields["source"], err = ref(source, true)
			if err != nil {
				return nil, "", err
			}
			fields["target"], err = ref(target, true)
			if err != nil {
				return nil, "", err
			}
		} else if len(args) != 0 {
			return nil, "", fmt.Errorf("canonical create accepts no positional title")
		}
	} else {
		if len(args) != 1 {
			return nil, "", fmt.Errorf("mutation requires one Resource Reference")
		}
		raw, err := ref(args[0], false)
		if err != nil {
			return nil, "", err
		}
		name := "bead"
		if link {
			name = "link"
		}
		fields[name] = raw
	}
	if op == retainedUpdateLink {
		for _, pair := range [][2]string{{"from", "source"}, {"to", "target"}} {
			if cmd.Flags().Changed(pair[0]) {
				value, _ := cmd.Flags().GetString(pair[0])
				raw, err := ref(value, true)
				if err != nil {
					return nil, "", err
				}
				fields[pair[1]] = raw
			}
		}
	}
	if deleted {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return nil, "", fmt.Errorf("delete requires --force apply intent")
		}
	} else {
		if cmd.Flags().Changed("type") {
			value, _ := cmd.Flags().GetString("type")
			raw, err := ref(value, true)
			if err != nil {
				return nil, "", err
			}
			fields["type"] = raw
		}
		input := &canonicalRetainedInputs{stdin: cmd.InOrStdin()}
		for _, plane := range []string{"properties", "metadata"} {
			name, raw, err := canonicalRetainedObjectEdit(cmd, plane, create, input)
			if err != nil {
				return nil, "", err
			}
			if name != "" {
				fields[name] = raw
			}
		}
		if cmd.Flags().Changed("revision-provenance") {
			value, _ := cmd.Flags().GetString("revision-provenance")
			raw, err := input.read(value)
			if err != nil {
				return nil, "", err
			}
			if err := canonicalRetainedProvenanceInput(raw); err != nil {
				return nil, "", err
			}
			fields["revisionProvenance"] = raw
		}
	}
	if op == retainedDeleteLink && cmd.Flags().Changed("revision-provenance") {
		value, _ := cmd.Flags().GetString("revision-provenance")
		input := &canonicalRetainedInputs{stdin: cmd.InOrStdin()}
		raw, err := input.read(value)
		if err != nil {
			return nil, "", err
		}
		if err := canonicalRetainedProvenanceInput(raw); err != nil {
			return nil, "", err
		}
		fields["revisionProvenance"] = raw
	}
	if cmd.Flags().Lookup("if-revision") != nil && cmd.Flags().Changed("if-revision") {
		value, _ := cmd.Flags().GetString("if-revision")
		if !strings.HasPrefix(value, "=") || len(value) == 1 {
			return nil, "", fmt.Errorf("HTTP guards require =TOKEN; no local ordinal lookup")
		}
		fields["expectedRevision"], _ = json.Marshal(strings.TrimPrefix(value, "="))
	}
	if !create && !deleted {
		meaningful := false
		for _, key := range []string{"type", "properties", "metadata", "propertiesChange", "metadataChange", "source", "target"} {
			meaningful = meaningful || fields[key] != nil
		}
		if !meaningful {
			return nil, "", fmt.Errorf("update requires a meaningful selected input")
		}
	}
	raw, err := canonicalMutationMarshalNative(fields)
	if err != nil {
		return nil, "", err
	}
	if err := canonicalRetainedStrictJSON(raw); err != nil {
		return nil, "", err
	}
	return raw, "", nil
}

func canonicalRetainedKey(key string) bool {
	if len(key) < 1 || len(key) > 256 {
		return false
	}
	for _, c := range []byte(key) {
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func completeCanonicalRetainedRead(cmd *cobra.Command, opts clioutput.Options, reply canonicalRetainedReply) error {
	result := canonicalMutationOutputResult{Value: reply, OperationErr: reply.OperationErr, ExchangeErr: reply.ExchangeErr, ReadErr: reply.ReadErr, LimitErr: reply.LimitErr, CloseErr: reply.CloseErr, PayloadErr: reply.ParseErr, Disposition: canonicalMutationOutputReturned}
	descriptor := canonicalResourceDescriptor()
	descriptor.BDPSpec = canonicalRetainedContract
	if reply.Headers.Get("BDP-Resource-Kind") == "link" {
		descriptor.Protection = canonicalLinkProtection()
	}
	renderer, err := clioutput.New(descriptor, opts)
	if err != nil {
		result.RenderErr = err
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return &exitError{Code: 1}
	}
	var raw []byte
	switch {
	case len(reply.Problem) != 0:
		result.Disposition = canonicalMutationOutputFailed
		raw, result.RenderErr = renderer.RenderFailure("bdp", reply.Problem)
	case reply.OperationErr != nil || reply.acquisitionFailed() || reply.ParseErr != nil || !reply.Checked:
		result.Disposition = canonicalMutationOutputFailed
		raw, result.RenderErr = renderer.RenderFailure("cli", canonicalMutationOutputProblem("read_unavailable", "the original retained read could not be verified"))
	default:
		// A Close error does not erase or replace the checked original Resource.
		raw, result.RenderErr = renderer.Render(reply.Resource)
	}
	if result.RenderErr == nil {
		result.OutputErr = writeCanonicalDocument(cmd.OutOrStdout(), raw)
	}
	if result.Err() != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), result.Err())
		return &exitError{Code: 1}
	}
	return nil
}

type canonicalRetainedInputs struct {
	stdin io.Reader
	used  bool
}

func (input *canonicalRetainedInputs) read(value string) (json.RawMessage, error) {
	var raw []byte
	if strings.HasPrefix(value, "@") {
		var reader io.Reader
		var file *os.File
		if value == "@-" {
			if input.used {
				return nil, fmt.Errorf("only one value may consume stdin")
			}
			input.used = true
			reader = input.stdin
		} else {
			var err error
			file, err = os.Open(strings.TrimPrefix(value, "@")) // #nosec G304 -- explicit CLI value-file input
			if err != nil {
				return nil, err
			}
			reader = file
		}
		var err error
		raw, err = io.ReadAll(io.LimitReader(reader, 65537))
		var closeErr error
		if file != nil {
			closeErr = file.Close()
		}
		if err != nil || closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
	} else {
		raw = []byte(value)
	}
	if len(raw) > 65536 {
		return nil, fmt.Errorf("retained value input exceeds request bound")
	}
	if err := canonicalRetainedStrictJSON(raw); err != nil {
		return nil, err
	}
	return bytes.Clone(raw), nil
}

type canonicalRetainedEdit struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value,omitempty"`
}

func canonicalRetainedPointer(key string) string {
	return "/" + strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
}
func canonicalRetainedObjectEdit(cmd *cobra.Command, plane string, create bool, input *canonicalRetainedInputs) (string, json.RawMessage, error) {
	singular := "property"
	if plane == "metadata" {
		singular = "metadata"
	}
	merge, replace, patch := cmd.Flags().Changed(plane), cmd.Flags().Changed("replace-"+plane), cmd.Flags().Changed("patch-"+plane)
	keyed := cmd.Flags().Changed("set-"+singular) || cmd.Flags().Changed("unset-"+singular)
	count := 0
	for _, set := range []bool{merge, replace, patch, keyed} {
		if set {
			count++
		}
	}
	if count > 1 || create && (replace || patch || keyed) {
		return "", nil, fmt.Errorf("conflicting %s edit modes", plane)
	}
	if count == 0 {
		return "", nil, nil
	}
	var edits []canonicalRetainedEdit
	if keyed {
		sets, _ := cmd.Flags().GetStringArray("set-" + singular)
		unsets, _ := cmd.Flags().GetStringArray("unset-" + singular)
		values := map[string]json.RawMessage{}
		var order []string
		for _, assignment := range sets {
			key, value, ok := strings.Cut(assignment, "=")
			if !ok || !utf8.ValidString(key) || !utf8.ValidString(value) {
				return "", nil, fmt.Errorf("set requires UTF-8 KEY=VALUE")
			}
			var raw json.RawMessage
			if value == "null" || value == "true" || value == "false" || len(value) > 0 && (value[0] == '-' || value[0] >= '0' && value[0] <= '9') && json.Valid([]byte(value)) {
				raw = []byte(value)
			} else {
				raw, _ = json.Marshal(value)
			}
			if values[key] == nil {
				order = append(order, key)
			}
			values[key] = raw
		}
		for _, key := range order {
			edits = append(edits, canonicalRetainedEdit{Op: "add", Path: canonicalRetainedPointer(key), Value: values[key]})
		}
		seen := map[string]bool{}
		for _, key := range unsets {
			if !utf8.ValidString(key) {
				return "", nil, fmt.Errorf("unset requires UTF-8 key")
			}
			if !seen[key] {
				seen[key] = true
				edits = append(edits, canonicalRetainedEdit{Op: "add", Path: canonicalRetainedPointer(key), Value: json.RawMessage("null")}, canonicalRetainedEdit{Op: "remove", Path: canonicalRetainedPointer(key)})
			}
		}
	} else {
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
		if patch {
			if err := canonicalRetainedValidateChanges(raw); err != nil {
				return "", nil, err
			}
			return plane + "Change", raw, nil
		}
		fields, err := canonicalRetainedObject(raw)
		if err != nil {
			return "", nil, err
		}
		if create || replace {
			return plane, raw, nil
		}
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			edits = append(edits, canonicalRetainedEdit{Op: "add", Path: canonicalRetainedPointer(key), Value: fields[key]})
		}
	}
	if edits == nil {
		edits = []canonicalRetainedEdit{}
	}
	raw, err := canonicalMutationMarshalNative(edits)
	return plane + "Change", raw, err
}
func canonicalRetainedValidateChanges(raw []byte) error {
	var edits []json.RawMessage
	if err := canonicalRetainedStrictJSON(raw); err != nil {
		return err
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &edits) != nil {
		return fmt.Errorf("Change must be an array")
	}
	for _, edit := range edits {
		fields, err := canonicalRetainedObject(edit, "op", "path", "value")
		if err != nil {
			return err
		}
		op, err := canonicalRetainedString(fields["op"])
		if err != nil {
			return err
		}
		var path string
		if fields["path"] == nil || json.Unmarshal(fields["path"], &path) != nil {
			return fmt.Errorf("Change requires a JSON Pointer")
		}
		if path != "" && !strings.HasPrefix(path, "/") {
			return fmt.Errorf("invalid Change JSON Pointer")
		}
		for i := 0; i < len(path); i++ {
			if path[i] == '~' {
				if i+1 >= len(path) || path[i+1] != '0' && path[i+1] != '1' {
					return fmt.Errorf("invalid JSON Pointer escape")
				}
				i++
			}
		}
		switch op {
		case "add", "replace":
			if fields["value"] == nil {
				return fmt.Errorf("Change value missing")
			}
		case "remove":
			if fields["value"] != nil {
				return fmt.Errorf("remove has no value")
			}
		default:
			return fmt.Errorf("unsupported Change operation")
		}
	}
	return nil
}
