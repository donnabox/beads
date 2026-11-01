package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/steveyegge/beads/internal/clioutput"
)

const canonicalRetainedBatchCapability = "https://github.com/gastownhall/bdp/capabilities/atomic-batch"

// Select the existing native batch1 lane explicitly. Experimental2/offline
// authoring keeps its existing compiler and meaning outside this selection.
func canonicalRetainedBatchSelected(cmd *cobra.Command) bool {
	version, _ := cmd.Flags().GetString("batch-version")
	mode, _ := cmd.Flags().GetString("dry-run-mode")
	return version == "1" || mode == "bdp" || cmd.Flags().Changed("retry-file")
}
func validateCanonicalRetainedBatchSyntax(cmd *cobra.Command, argv []string) error {
	refuse := func(message string) error { return graphFailure("invalid_selector", message, 2) }
	if err := canonicalScalarOccurrences(cmd, argv, "scope", "input-format", "batch-version", "file", "dry-run-mode", "idempotency-key", "retry-file"); err != nil {
		return refuse(err.Error())
	}
	var rejected string
	allowed := map[string]bool{}
	for _, name := range []string{"cli-profile", "scope", "input-format", "batch-version", "file", "dry-run-mode", "idempotency-key", "retry-file", "output-format", "include", "omit", "json", "directory", "quiet", "no-color"} {
		allowed[name] = true
	}
	visit := func(flag *pflag.Flag) {
		if !allowed[flag.Name] && rejected == "" {
			rejected = flag.Name
		}
	}
	cmd.Flags().Visit(visit)
	cmd.InheritedFlags().Visit(visit)
	if rejected != "" {
		return refuse("native batch1 does not accept --" + rejected)
	}
	if cmd.Flags().Changed("retry-file") {
		file, _ := cmd.Flags().GetString("retry-file")
		if file == "" {
			return refuse("batch recovery requires a nonempty saved file")
		}
		for _, name := range []string{"file", "input-format", "batch-version", "dry-run-mode", "idempotency-key"} {
			if cmd.Flags().Changed(name) {
				return refuse("batch recovery does not accept fresh --" + name)
			}
		}
		return nil
	}
	format, _ := cmd.Flags().GetString("input-format")
	version, _ := cmd.Flags().GetString("batch-version")
	file, _ := cmd.Flags().GetString("file")
	mode, _ := cmd.Flags().GetString("dry-run-mode")
	if format != "json" || version != "1" || file == "" {
		return refuse("native batch1 requires --input-format json --batch-version 1 --file FILE (or -)")
	}
	if mode != "" && mode != "bdp" {
		return refuse("native batch1 preview requires bdp mode")
	}
	if cmd.Flags().Changed("idempotency-key") {
		key, _ := cmd.Flags().GetString("idempotency-key")
		if !canonicalRetainedKey(key) {
			return refuse("explicit retained idempotency key must match [A-Za-z0-9_-]{1,256}")
		}
	}
	return nil
}

func runCanonicalRetainedBatch(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return graphFailure("invalid_selector", "native batch1 accepts no positional operands", 2)
	}
	selected, ok := cmd.Context().Value(canonicalRetainedScopeKey).(canonicalScopeContext)
	if !ok {
		return graphFailure("capability_unavailable", "batch1 root admission is missing", 5)
	}
	options, ok := cmd.Context().Value(canonicalRetainedBindingsKey).(canonicalRetainedBindings)
	if !ok {
		var err error
		options, err = newCanonicalRetainedRootBindings(cmd.Context(), true)
		if err != nil {
			return graphFailure("capability_unavailable", err.Error(), 5)
		}
	}
	if err := options.validate(selected.URL); err != nil {
		return graphFailure("capability_unavailable", err.Error(), 5)
	}
	output, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	completion, err := newCanonicalMutationOutput(canonicalMutationOutputRetainedBatch, clioutput.MutationDescriptor{BDPSpec: canonicalRetainedContract, Required: []string{"capabilityVersion", "changed", "results", "resources"}}, output, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	ctx, cancel := context.WithDeadline(cmd.Context(), options.Deadline)
	defer cancel()
	retry, _ := cmd.Flags().GetString("retry-file")
	var reply canonicalRetainedReply
	var recovery string
	if retry != "" {
		// Load original carrier before navigation. Fresh capability/subset/limits
		// cannot prohibit a retained result after execution support is withdrawn.
		saved, loadErr := loadCanonicalRetainedInvocation(options.JournalRoot, retry)
		if loadErr != nil {
			return graphFailure("invalid_selector", loadErr.Error(), 2)
		}
		original := saved.record.Request
		if original.Operation != retainedAtomicBatch || original.Scope != selected.URL {
			return graphFailure("invalid_selector", "saved invocation is not batch1 in this selected Scope", 2)
		}
		if _, err := discoverCanonicalRetainedBatch(ctx, selected.URL, options, false); err != nil {
			return graphFailure("capability_unavailable", "original batch Scope/contract cannot be verified", 5)
		}
		binding := canonicalRetainedWriteBinding{options: options, scope: selected.URL, batchTarget: original.Target}
		saved, err = saveCanonicalRetainedInvocation(options.JournalRoot, saved.record)
		if err != nil {
			return graphFailure("capability_unavailable", err.Error(), 5)
		}
		request, provider, requestErr := binding.request(ctx, original)
		if requestErr != nil {
			reply = canonicalRetainedReply{OperationErr: requestErr}
		} else {
			reply = binding.submitSaved(ctx, canonicalRetainedSubmission{saved: saved, request: request, provider: provider})
		}
		recovery = retry
	} else {
		file, _ := cmd.Flags().GetString("file")
		input := &canonicalRetainedInputs{stdin: cmd.InOrStdin()}
		raw, readErr := input.read("@" + file)
		if readErr != nil {
			return graphFailure("invalid_selector", readErr.Error(), 2)
		}
		members, syntaxErr := canonicalRetainedBatchMembers(raw)
		if syntaxErr != nil {
			return graphFailure("invalid_selector", syntaxErr.Error(), 2)
		}
		binding, discoverErr := discoverCanonicalRetainedBatch(ctx, selected.URL, options, true)
		if discoverErr != nil {
			return graphFailure("capability_unavailable", discoverErr.Error(), 5)
		}
		if err := binding.admit(members); err != nil {
			return graphFailure("capability_unavailable", err.Error(), 5)
		}
		// Match the journal RawMessage encoding before the FIRST submission. Only
		// insignificant JSON whitespace is removed; number/string/pin tokens remain.
		var wire bytes.Buffer
		if err := json.Compact(&wire, raw); err != nil {
			return err
		}
		raw = bytes.Clone(wire.Bytes())
		key, _ := cmd.Flags().GetString("idempotency-key")
		prepared := canonicalRetainedPrepared{Scope: selected.URL, Contract: canonicalRetainedContract, Operation: retainedAtomicBatch, Method: http.MethodPost, Target: binding.write.batchTarget, Key: key, Body: bytes.Clone(raw)}
		mode, _ := cmd.Flags().GetString("dry-run-mode")
		if mode == "bdp" {
			// Exact native request preview; no key generation, private save or POST.
			preview, err := canonicalMutationMarshalNative(struct {
				Method      string          `json:"method"`
				Target      string          `json:"target"`
				ContentType string          `json:"contentType"`
				Accept      string          `json:"accept"`
				Key         string          `json:"idempotencyKey,omitempty"`
				Body        json.RawMessage `json:"body"`
			}{prepared.Method, prepared.Target, canonicalRetainedMedia, canonicalRetainedMedia, prepared.Key, prepared.Body})
			if err != nil {
				return err
			}
			return writeCanonicalDocument(cmd.OutOrStdout(), append(preview, '\n'))
		}
		reply, recovery = binding.write.apply(ctx, prepared)
		// Fresh advertised whole-result bound is checked without truncation. Explicit
		// recovery uses the original protocol ceiling, never a fresh smaller bound.
		canonicalRetainedBatchResultBound(&reply, binding.maximumResultBytes)
	}
	if recovery != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "Retained recovery file: %s\n", recovery)
	}
	result := completion.Complete(reply, "", reply.OperationErr, reply.CloseErr)
	if result.Err() != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), result.Err())
		return &exitError{Code: 1}
	}
	return nil
}

type canonicalRetainedBatchBinding struct {
	write              canonicalRetainedWriteBinding
	operations         map[canonicalRetainedOperation]bool
	maximumOperations  *big.Rat
	maximumResultBytes *big.Rat
}

func canonicalRetainedBatchOwnedTarget(scope, target string) bool {
	owner, err := canonicalRetainedURL(scope)
	if err != nil {
		return false
	}
	operation, err := canonicalRetainedURL(target)
	if err != nil {
		return false
	}
	// URL path boundaries prevent /scope-other from being mistaken for /scope.
	prefix := owner.EscapedPath()
	if prefix == "" {
		prefix = "/"
	}
	for _, segment := range strings.Split(operation.Path, "/") {
		if segment == "." || segment == ".." {
			return false
		}
	}
	within := operation.EscapedPath() == prefix || strings.HasSuffix(prefix, "/") && strings.HasPrefix(operation.EscapedPath(), prefix) || !strings.HasSuffix(prefix, "/") && strings.HasPrefix(operation.EscapedPath(), prefix+"/")
	return owner.Scheme == operation.Scheme && owner.Host == operation.Host && within
}
func discoverCanonicalRetainedBatch(ctx context.Context, scope string, options canonicalRetainedBindings, fresh bool) (canonicalRetainedBatchBinding, error) {
	landing := options.get(ctx, scope)
	if err := canonicalRetainedScopeLanding(landing); err != nil {
		return canonicalRetainedBatchBinding{}, err
	}
	target, err := canonicalRetainedServiceDescription(landing.Headers)
	if err != nil {
		return canonicalRetainedBatchBinding{}, err
	}
	description, err := options.description(ctx, target)
	if err != nil {
		return canonicalRetainedBatchBinding{}, err
	}
	if description.Scope != scope {
		return canonicalRetainedBatchBinding{}, fmt.Errorf("batch1 owning Scope changed")
	}
	b := canonicalRetainedBatchBinding{write: canonicalRetainedWriteBinding{options: options, scope: scope}}
	if !fresh {
		return b, nil
	}
	if description.Profile != "write" {
		return b, fmt.Errorf("Scope does not advertise selected batch1 execution")
	}
	capabilities, err := canonicalRetainedObject(description.Capabilities)
	if err != nil {
		return b, err
	}
	versions, err := canonicalRetainedObject(capabilities[canonicalRetainedBatchCapability])
	if err != nil {
		return b, fmt.Errorf("Scope does not advertise atomic batch1")
	}
	fields, err := canonicalRetainedObject(versions["1"], "operation", "operations", "maximumOperations", "maximumResultBytes")
	if err != nil {
		return b, err
	}
	operation, err := canonicalRetainedString(fields["operation"])
	if err != nil || !canonicalRetainedBatchOwnedTarget(scope, operation) {
		return b, fmt.Errorf("batch1 target is outside owning Scope")
	}
	b.write.batchTarget = operation
	operations, err := canonicalRetainedArray(fields["operations"])
	if err != nil || len(operations) == 0 {
		return b, fmt.Errorf("empty batch1 operation support")
	}
	b.operations = map[canonicalRetainedOperation]bool{}
	for _, raw := range operations {
		name, err := canonicalRetainedString(raw)
		op := canonicalRetainedOperation(name)
		if err != nil || !canonicalRetainedBatchOperation(op) || b.operations[op] {
			return b, fmt.Errorf("invalid batch1 operation support")
		}
		b.operations[op] = true
	}
	for _, name := range []string{"maximumOperations", "maximumResultBytes"} {
		if raw, ok := fields[name]; ok {
			if err := canonicalRetainedInteger(raw, 1); err != nil {
				return b, err
			}
			value, _ := new(big.Rat).SetString(strings.TrimSpace(string(raw)))
			if name == "maximumOperations" {
				b.maximumOperations = value
			} else {
				b.maximumResultBytes = value
			}
		}
	}
	return b, nil
}
func (b canonicalRetainedBatchBinding) admit(members []canonicalRetainedBatchMember) error {
	if b.maximumOperations != nil && b.maximumOperations.Cmp(big.NewRat(int64(len(members)), 1)) < 0 {
		return fmt.Errorf("batch1 exceeds advertised operation bound")
	}
	for index, member := range members {
		if !b.operations[member.operation] {
			return fmt.Errorf("batch1 operation %d is not advertised", index)
		}
	}
	return nil
}
func canonicalRetainedBatchOperation(op canonicalRetainedOperation) bool {
	switch op {
	case retainedCreateBead, retainedUpdateBead, retainedDeleteBead, retainedCreateLink, retainedUpdateLink, retainedDeleteLink:
		return true
	}
	return false
}

type canonicalRetainedBatchMember struct {
	operation canonicalRetainedOperation
	name      string
	kind      string
	identity  string
}

func canonicalRetainedBatchName(value string) bool {
	if len(value) == 0 {
		return false
	}
	for i, c := range []byte(value) {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || i > 0 && (c >= '0' && c <= '9' || c == '_' || c == '-')) {
			return false
		}
	}
	return true
}

// Grammar-only validation preserves the original bytes. Handle interpretation
// is confined to prior kind-correct subject/endpoint slots, never literal data.
func canonicalRetainedBatchMembers(raw []byte) ([]canonicalRetainedBatchMember, error) {
	if len(raw) > 65536 {
		return nil, fmt.Errorf("batch1 request exceeds 65536 bytes")
	}
	root, err := canonicalRetainedObject(raw, "capabilityVersion", "operations")
	if err != nil {
		return nil, err
	}
	version, err := canonicalRetainedString(root["capabilityVersion"])
	if err != nil || version != "1" {
		return nil, fmt.Errorf("native batch1 requires capabilityVersion 1")
	}
	operations, err := canonicalRetainedArray(root["operations"])
	if err != nil || len(operations) == 0 {
		return nil, fmt.Errorf("batch1 requires nonempty operations")
	}
	names := map[string]string{}
	members := make([]canonicalRetainedBatchMember, 0, len(operations))
	for index, raw := range operations {
		fields, err := canonicalRetainedObject(raw)
		if err != nil {
			return nil, err
		}
		operation, err := canonicalRetainedString(fields["operation"])
		op := canonicalRetainedOperation(operation)
		if err != nil || !canonicalRetainedBatchOperation(op) {
			return nil, fmt.Errorf("unsupported native member at %d", index)
		}
		create := strings.HasPrefix(operation, "create")
		deleted := strings.HasPrefix(operation, "delete")
		kind := "bead"
		if strings.HasSuffix(operation, "Link") {
			kind = "link"
		}
		allowed := []string{"operation"}
		if create {
			allowed = append(allowed, "name", "id", "type", "properties", "metadata", "revisionProvenance")
		} else {
			allowed = append(allowed, kind, "expectedRevision")
			if !deleted {
				allowed = append(allowed, "type", "properties", "metadata", "propertiesChange", "metadataChange", "revisionProvenance")
			} else if kind == "link" {
				allowed = append(allowed, "revisionProvenance")
			}
		}
		if kind == "link" && !deleted {
			allowed = append(allowed, "source", "target")
		}
		if _, err := canonicalRetainedObject(raw, allowed...); err != nil {
			return nil, err
		}
		name := ""
		if value, ok := fields["name"]; ok {
			name, err = canonicalRetainedString(value)
			if err != nil || !canonicalRetainedBatchName(name) || names[name] != "" {
				return nil, fmt.Errorf("invalid or duplicate batch1 creator name")
			}
		}
		if create {
			if err := canonicalRetainedReference(fields["type"], true); err != nil {
				return nil, err
			}
			if value, ok := fields["id"]; ok {
				id, err := canonicalRetainedString(value)
				if err != nil || strings.HasPrefix(id, "@") {
					return nil, fmt.Errorf("invalid durable create id")
				}
			}
		} else {
			if err := canonicalRetainedBatchSlot(fields[kind], kind, false, names); err != nil {
				return nil, err
			}
		}
		for _, slot := range []string{"source", "target"} {
			if value, ok := fields[slot]; ok || create && kind == "link" {
				if err := canonicalRetainedBatchSlot(value, "bead", true, names); err != nil {
					return nil, err
				}
			}
		}
		if !create && fields["type"] != nil {
			if err := canonicalRetainedReference(fields["type"], true); err != nil {
				return nil, err
			}
		}
		for _, plane := range []string{"properties", "metadata"} {
			if fields[plane] != nil && fields[plane+"Change"] != nil {
				return nil, fmt.Errorf("batch1 replacement and Change conflict")
			}
			if value, ok := fields[plane]; ok {
				if _, err := canonicalRetainedObject(value); err != nil {
					return nil, err
				}
			}
			if value, ok := fields[plane+"Change"]; ok {
				if err := canonicalRetainedValidateChanges(value); err != nil {
					return nil, err
				}
			}
		}
		if value, ok := fields["revisionProvenance"]; ok {
			if err := canonicalRetainedProvenanceInput(value); err != nil {
				return nil, err
			}
		}
		if value, ok := fields["expectedRevision"]; ok {
			if _, err := canonicalRetainedString(value); err != nil {
				return nil, err
			}
		}
		if !create && !deleted {
			meaningful := false
			for _, key := range []string{"type", "source", "target", "properties", "metadata", "propertiesChange", "metadataChange"} {
				meaningful = meaningful || fields[key] != nil
			}
			if !meaningful {
				return nil, fmt.Errorf("batch1 update requires a meaningful input")
			}
		}
		identity := ""
		if create {
			identity, _ = canonicalRetainedString(fields["id"])
		} else {
			identity, _ = canonicalRetainedString(fields[kind])
		}
		members = append(members, canonicalRetainedBatchMember{op, name, kind, identity})
		if name != "" {
			names[name] = kind
		}
	}
	return members, nil
}
func canonicalRetainedBatchSlot(raw []byte, kind string, pinned bool, names map[string]string) error {
	value, err := canonicalRetainedString(raw)
	if err != nil {
		if !pinned {
			return fmt.Errorf("batch1 mutation subject must be an unpinned Reference")
		}
		fields, err := canonicalRetainedObject(raw, "uri", "revision")
		if err != nil {
			return err
		}
		value, err = canonicalRetainedString(fields["uri"])
		if err != nil {
			return err
		}
		if _, err := canonicalRetainedString(fields["revision"]); err != nil {
			return err
		}
	}
	if strings.HasPrefix(value, "@") {
		name := strings.TrimPrefix(value, "@")
		if !canonicalRetainedBatchName(name) || names[name] != kind {
			return fmt.Errorf("unknown, forward or wrong-kind batch1 handle")
		}
	}
	return nil
}

// A distinct result carrier is checked against the ORIGINAL submitted members.
// No singleton wrapper, fresh Resource GET, label rewrite or retry supplies it.
func (r *canonicalRetainedReply) validateBatch(request []byte, scope string) {
	if r.Checked || r.ParseErr != nil || r.OperationErr != nil || r.acquisitionFailed() {
		return
	}
	if r.Status != http.StatusOK {
		media, _, err := mime.ParseMediaType(r.Headers.Get("Content-Type"))
		if err != nil || media != "application/problem+json" || canonicalRetainedBatchProblem(r.NativeBody, r.Status, request) != nil {
			r.ParseErr = fmt.Errorf("unresolved batch1 operation response")
			return
		}
		r.Problem = bytes.Clone(r.NativeBody)
		r.OperationErr = fmt.Errorf("batch1 operation returned HTTP %d", r.Status)
		r.Checked = true
		return
	}
	check := *r
	check.CloseErr = nil
	if err := canonicalRetainedJSON(check, http.StatusOK); err != nil {
		r.ParseErr = err
		return
	}
	if err := canonicalRetainedBatchResult(r.NativeBody, request, scope); err != nil {
		r.ParseErr = err
		return
	}
	r.Mutation = bytes.Clone(r.NativeBody)
	r.Checked = true
}
func canonicalRetainedBatchProblem(raw []byte, status int, request ...[]byte) error {
	fields, err := canonicalRetainedObject(raw)
	if err != nil {
		return err
	}
	code, _ := canonicalRetainedString(fields["code"])
	if code != "capability-not-supported" && code != "operation-not-supported" {
		return canonicalRetainedProblem(raw, status, canonicalRetainedWriteProblem)
	}
	allowed := []string{"type", "title", "status", "detail", "instance", "code", "retry", "pointer"}
	if code == "capability-not-supported" {
		allowed = append(allowed, "capability", "capabilityVersion")
	} else {
		allowed = append(allowed, "operation")
	}
	if _, err := canonicalRetainedObject(raw, allowed...); err != nil {
		return err
	}
	typ, _ := canonicalRetainedString(fields["type"])
	retry, _ := canonicalRetainedString(fields["retry"])
	if typ != "https://github.com/gastownhall/bdp/problems/request" || status != 422 || retry != "after-state-change" {
		return fmt.Errorf("batch1 feature Problem binding mismatch")
	}
	if err := canonicalRetainedInteger(fields["status"], 422); err != nil {
		return err
	}
	number, _ := new(big.Rat).SetString(strings.TrimSpace(string(fields["status"])))
	if number.Cmp(big.NewRat(422, 1)) != 0 {
		return fmt.Errorf("batch1 Problem status mismatch")
	}
	for _, key := range []string{"title", "detail"} {
		if value, ok := fields[key]; ok {
			if _, err := canonicalRetainedJSONString(value); err != nil {
				return err
			}
		}
	}
	if value, ok := fields["instance"]; ok {
		if err := canonicalRetainedAbsoluteURI(value); err != nil {
			return err
		}
	}
	pointer, err := canonicalRetainedString(fields["pointer"])
	if err != nil {
		return err
	}
	if code == "capability-not-supported" {
		capability, _ := canonicalRetainedString(fields["capability"])
		if capability != canonicalRetainedBatchCapability || pointer != "/capabilityVersion" {
			return fmt.Errorf("batch1 capability Problem mismatch")
		}
		version, err := canonicalRetainedString(fields["capabilityVersion"])
		if err != nil {
			return err
		}
		if len(request) != 0 {
			input, err := canonicalRetainedObject(request[0], "capabilityVersion", "operations")
			if err != nil {
				return err
			}
			selected, err := canonicalRetainedString(input["capabilityVersion"])
			if err != nil || selected != version {
				return fmt.Errorf("batch1 capability Problem describes another request")
			}
		}
		return nil
	}
	operation, _ := canonicalRetainedString(fields["operation"])
	if !canonicalRetainedBatchOperation(canonicalRetainedOperation(operation)) {
		return fmt.Errorf("batch1 operation Problem mismatch")
	}
	index := strings.TrimSuffix(strings.TrimPrefix(pointer, "/operations/"), "/operation")
	if len(index) == 0 || len(index) > 1 && index[0] == '0' || pointer != "/operations/"+index+"/operation" {
		return fmt.Errorf("batch1 operation Problem pointer mismatch")
	}
	for _, digit := range []byte(index) {
		if digit < '0' || digit > '9' {
			return fmt.Errorf("batch1 operation Problem pointer mismatch")
		}
	}
	if len(request) != 0 {
		members, err := canonicalRetainedBatchMembers(request[0])
		if err != nil {
			return err
		}
		match := false
		for position, member := range members {
			if pointer == fmt.Sprintf("/operations/%d/operation", position) && operation == string(member.operation) {
				match = true
				break
			}
		}
		if !match {
			return fmt.Errorf("batch1 operation Problem describes another request")
		}
	}
	return nil
}

func canonicalRetainedBatchResult(raw, request []byte, scope string) error {
	members, err := canonicalRetainedBatchMembers(request)
	if err != nil {
		return err
	}
	fields, err := canonicalRetainedObject(raw, "capabilityVersion", "changed", "transaction", "results", "resources")
	if err != nil {
		return err
	}
	version, _ := canonicalRetainedString(fields["capabilityVersion"])
	if version != "1" {
		return fmt.Errorf("batch1 result version mismatch")
	}
	changed, err := canonicalRetainedBoolean(fields["changed"])
	if err != nil {
		return err
	}
	if changed {
		if _, err := canonicalRetainedString(fields["transaction"]); err != nil {
			return err
		}
	} else if fields["transaction"] != nil {
		return fmt.Errorf("unchanged batch1 result has a transaction")
	}
	results, err := canonicalRetainedArray(fields["results"])
	if err != nil || len(results) != len(members) {
		return fmt.Errorf("batch1 result mapping cardinality mismatch")
	}
	mapped := map[string]string{}
	namedIDs := map[string]string{}
	firstOperations := map[string]string{}
	lastOperations := map[string]string{}
	for index, raw := range results {
		mapping, err := canonicalRetainedObject(raw, "operationIndex", "operation", "id", "operationName")
		if err != nil {
			return err
		}
		if err := canonicalRetainedInteger(mapping["operationIndex"], 0); err != nil {
			return err
		}
		ordinal, _ := new(big.Rat).SetString(strings.TrimSpace(string(mapping["operationIndex"])))
		if ordinal.Cmp(big.NewRat(int64(index), 1)) != 0 {
			return fmt.Errorf("batch1 mapping order mismatch")
		}
		operation, _ := canonicalRetainedString(mapping["operation"])
		name, _ := canonicalRetainedString(mapping["operationName"])
		if operation != string(members[index].operation) || name != members[index].name || (mapping["operationName"] != nil) != (members[index].name != "") {
			return fmt.Errorf("batch1 mapping operation/name mismatch")
		}
		id, err := canonicalRetainedString(mapping["id"])
		if err != nil {
			return err
		}
		if _, err := canonicalRetainedURL(id); err != nil {
			return err
		}
		identity := members[index].identity
		if strings.HasPrefix(identity, "@") {
			if namedIDs[strings.TrimPrefix(identity, "@")] != id {
				return fmt.Errorf("batch1 handle mapping identity mismatch")
			}
		} else if strings.HasPrefix(string(members[index].operation), "create") && identity != "" {
			expected := identity
			if _, absoluteErr := canonicalRetainedURL(identity); absoluteErr != nil {
				if scope == "" {
					return fmt.Errorf("batch1 caller-assigned identity lacks its saved Scope")
				}
				expected = scope + identity
			}
			if expected != id {
				return fmt.Errorf("batch1 caller-assigned create identity mapping mismatch")
			}
		} else if members[index].kind == "link" && !strings.HasPrefix(string(members[index].operation), "create") {
			// Link subjects have no Bead-alias resolution. Compare their once-decoded
			// local URI spelling without fetching or substituting any current record.
			expected, err := canonicalRetainedBatchLinkIdentity(scope, identity)
			if err != nil || expected != id {
				return fmt.Errorf("batch1 Link subject mapping mismatch")
			}
		}
		if members[index].name != "" {
			namedIDs[members[index].name] = id
		}
		if old := mapped[id]; old != "" && old != members[index].kind {
			return fmt.Errorf("batch1 mapped identity changes kind")
		}
		mapped[id] = members[index].kind
		if firstOperations[id] == "" {
			firstOperations[id] = operation
		}
		lastOperations[id] = operation
	}
	resources, err := canonicalRetainedArray(fields["resources"])
	if err != nil || len(resources) == 0 {
		return fmt.Errorf("batch1 lacks final Resource rows")
	}
	rows := map[string]map[string]json.RawMessage{}
	kinds := map[string]string{}
	revisions := map[string]string{}
	previous := ""
	anyChanged := false
	for _, raw := range resources {
		row, err := canonicalRetainedObject(raw, "id", "resourceKind", "outcome", "changed", "resource", "deleted", "ownerRevisions")
		if err != nil {
			return err
		}
		id, err := canonicalRetainedString(row["id"])
		if err != nil {
			return err
		}
		if _, err := canonicalRetainedURL(id); err != nil {
			return err
		}
		if previous != "" && previous >= id {
			return fmt.Errorf("batch1 rows are not unique canonical-id ordered")
		}
		previous = id
		kind, _ := canonicalRetainedString(row["resourceKind"])
		if kind != "bead" && kind != "link" {
			return fmt.Errorf("batch1 final Resource kind mismatch")
		}
		outcome, _ := canonicalRetainedString(row["outcome"])
		rowChanged, err := canonicalRetainedBoolean(row["changed"])
		if err != nil {
			return err
		}
		if rowChanged != (outcome != "unchanged") {
			return fmt.Errorf("batch1 row changed disposition mismatch")
		}
		anyChanged = anyChanged || rowChanged
		if mapped[id] == "" && kind == "bead" && outcome != "updated" && outcome != "unchanged" {
			return fmt.Errorf("indirect batch1 owner row must preserve existence")
		}
		if mapped[id] != "" {
			created := strings.HasPrefix(firstOperations[id], "create")
			deleted := strings.HasPrefix(lastOperations[id], "delete")
			switch outcome {
			case "allocated":
				if !created || !deleted {
					return fmt.Errorf("batch1 allocation row contradicts member history")
				}
			case "created":
				if !created || deleted {
					return fmt.Errorf("batch1 created row contradicts member history")
				}
			case "deleted":
				if created || !deleted {
					return fmt.Errorf("batch1 deleted row contradicts member history")
				}
			case "updated", "unchanged":
				if created || deleted {
					return fmt.Errorf("batch1 final row contradicts member history")
				}
			}
		}
		identity := map[string]json.RawMessage(nil)
		switch outcome {
		case "created", "updated", "unchanged":
			if row["deleted"] != nil {
				return fmt.Errorf("live batch1 row has a tombstone")
			}
			if err := validateCanonicalRetainedResource(row["resource"], id); err != nil {
				return err
			}
			identity, _ = canonicalRetainedObject(row["resource"])
			if (identity["source"] != nil) != (kind == "link") || identity["links"] != nil {
				return fmt.Errorf("batch1 final Resource kind/context mismatch")
			}
		case "deleted":
			if row["resource"] != nil {
				return fmt.Errorf("deleted batch1 row has a Resource")
			}
			deleted, err := canonicalRetainedObject(row["deleted"], "resourceKind", "resource")
			if err != nil {
				return err
			}
			deletedKind, _ := canonicalRetainedString(deleted["resourceKind"])
			if deletedKind != kind {
				return fmt.Errorf("batch1 tombstone kind mismatch")
			}
			identity, err = canonicalRetainedObject(deleted["resource"], "id", "type", "revision")
			if err != nil {
				return err
			}
			captured, _ := canonicalRetainedString(identity["id"])
			if captured != id {
				return fmt.Errorf("batch1 tombstone identity mismatch")
			}
			if err := canonicalRetainedReference(identity["type"], true); err != nil {
				return err
			}
		case "allocated":
			if row["resource"] != nil || row["deleted"] != nil || row["ownerRevisions"] != nil {
				return fmt.Errorf("allocation-only batch1 row leaks Resource or owner evidence")
			}
		default:
			return fmt.Errorf("unknown batch1 row outcome")
		}
		if identity != nil {
			revision, err := canonicalRetainedString(identity["revision"])
			if err != nil {
				return err
			}
			revisions[id] = revision
		}
		if kind == "bead" && row["ownerRevisions"] != nil {
			return fmt.Errorf("Bead batch1 row has Link owners")
		}
		rows[id] = row
		kinds[id] = kind
	}
	if anyChanged != changed {
		return fmt.Errorf("batch1 aggregate changed mismatch")
	}
	for id, kind := range mapped {
		if kinds[id] != kind {
			return fmt.Errorf("batch1 mapping lacks matching final row")
		}
	}
	// Owner rows can be additional affected Beads. Do not infer missing initial
	// ownership from final Types: this validates declared owner correspondence;
	// service transaction qualification proves completeness of the affected set.
	declaredOwners := map[string]bool{}
	for id, row := range rows {
		if raw, ok := row["ownerRevisions"]; ok {
			owners, err := canonicalRetainedArray(raw)
			if err != nil || len(owners) < 1 || len(owners) > 2 {
				return fmt.Errorf("invalid batch1 owner rows")
			}
			previous := ""
			for _, raw := range owners {
				owner, err := canonicalRetainedObject(raw, "id", "revision")
				if err != nil {
					return err
				}
				ownerID, err := canonicalRetainedString(owner["id"])
				if err != nil {
					return err
				}
				revision, err := canonicalRetainedString(owner["revision"])
				if err != nil {
					return err
				}
				if previous != "" && previous >= ownerID || kinds[ownerID] != "bead" || revisions[ownerID] != revision {
					return fmt.Errorf("batch1 Link owner correspondence mismatch")
				}
				previous = ownerID
				declaredOwners[ownerID] = true
			}
		}
		if mapped[id] == "" && kinds[id] != "bead" {
			return fmt.Errorf("unmapped batch1 final Link row")
		}
	}
	for id := range rows {
		if mapped[id] == "" && !declaredOwners[id] {
			return fmt.Errorf("batch1 final row is neither mapped nor a declared affected owner")
		}
	}
	return nil
}

func canonicalRetainedBatchResultBound(reply *canonicalRetainedReply, maximum *big.Rat) {
	if reply.Status == http.StatusOK && maximum != nil && maximum.Cmp(big.NewRat(int64(len(reply.NativeBody)), 1)) < 0 {
		reply.LimitErr = fmt.Errorf("batch1 response exceeds advertised whole-result bound")
		reply.Checked = false
		reply.Mutation = nil
	}
}

// Match the bound service's canonicalSubject/canonicalLocalReference URI rule
// for Link subjects only. This is pure spelling comparison, never resolution.
func canonicalRetainedBatchLinkIdentity(scope, reference string) (string, error) {
	local := reference
	if _, err := canonicalRetainedURL(reference); err == nil {
		if scope == "" || !strings.HasPrefix(reference, scope) {
			return reference, nil
		}
		local = strings.TrimPrefix(reference, scope)
	}
	if scope == "" || local == "" || strings.ContainsAny(local, "?#\\ \t\r\n\f\v") {
		return "", fmt.Errorf("invalid local Link subject spelling")
	}
	segments := strings.Split(local, "/")
	const hex = "0123456789ABCDEF"
	for index, segment := range segments {
		decoded, err := url.PathUnescape(segment)
		if err != nil || !utf8.ValidString(decoded) {
			return "", fmt.Errorf("invalid local Link URI segment")
		}
		var canonical strings.Builder
		for _, c := range []byte(decoded) {
			if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("._~!$&'()*+,;=:@-", rune(c)) {
				canonical.WriteByte(c)
			} else {
				canonical.WriteByte('%')
				canonical.WriteByte(hex[c>>4])
				canonical.WriteByte(hex[c&15])
			}
		}
		segments[index] = canonical.String()
	}
	return scope + strings.Join(segments, "/"), nil
}
