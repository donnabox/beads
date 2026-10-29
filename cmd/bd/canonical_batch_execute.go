package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

const canonicalBatchContract = "urn:beads:experimental:canonical-batch:layout-15:encoder-2"

type canonicalBatchInvocation struct {
	Format    string          `json:"format"`
	Workspace string          `json:"workspace"`
	Scope     string          `json:"scope"`
	Authority string          `json:"authority"`
	Database  string          `json:"database"`
	Schema    int             `json:"schema"`
	Key       string          `json:"key"`
	Actor     string          `json:"actor"`
	Request   json.RawMessage `json:"request"`
}

// The CLI executes one Cobra command at a time. This frozen value only selects
// the original workspace during admission; the actual store binding is checked.
var canonicalBatchRetry *canonicalBatchInvocation

func init() {
	batchCmd.Flags().String("idempotency-key", "", "Key for one durable canonical batch decision; omission generates and saves one")
	batchCmd.Flags().String("retry-file", "", "Recover a frozen canonical batch invocation without rereading its source inputs")
	registerCanonicalOutputFlags(batchCmd)
	canonicalFailureFactories[batchCmd] = func(opts clioutput.Options) (canonicalFailureRenderer, error) {
		return canonicalBatchRenderer(opts)
	}
}

func canonicalBatchRenderer(opts clioutput.Options) (*clioutput.MutationRenderer, error) {
	return clioutput.NewMutation(clioutput.MutationDescriptor{BDPSpec: canonicalBatchContract, Required: []string{"capabilityVersion", "changed", "results", "resources"}}, opts)
}

func readCanonicalBatchInvocation(file string) (*canonicalBatchInvocation, error) {
	f, err := os.Open(file) // #nosec G304 -- explicit recovery file
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, canonicalbatch.MaxInputBytes+32769))
	if err != nil {
		return nil, err
	}
	if len(raw) > canonicalbatch.MaxInputBytes+32768 {
		return nil, fmt.Errorf("recovery file exceeds input limit")
	}
	raw, err = graph.CanonicalizeJSON(raw)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var invocation canonicalBatchInvocation
	if err := decoder.Decode(&invocation); err != nil {
		return nil, err
	}
	if invocation.Format != "canonical-batch-invocation-1" || invocation.Schema != graphstore.AtomicBatchSchemaVersion || invocation.Key == "" || !filepath.IsAbs(invocation.Workspace) || filepath.Clean(invocation.Workspace) != invocation.Workspace || invocation.Authority == "" || invocation.Database == "" {
		return nil, fmt.Errorf("invalid canonical batch recovery binding")
	}
	if err := graph.ValidateScopeURL(invocation.Scope); err != nil {
		return nil, err
	}
	if _, err := canonicalbatch.Compile(invocation.Request, canonicalbatch.Options{Format: "json", Scope: invocation.Scope}); err != nil {
		return nil, err
	}
	return &invocation, nil
}

func saveCanonicalBatchInvocation(invocation canonicalBatchInvocation) (string, error) {
	dir := filepath.Join(invocation.Workspace, "batch-invocations")
	if err := os.Mkdir(dir, 0o700); err != nil && !os.IsExist(err) {
		return "", err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("invalid private batch invocation directory")
	}
	raw, err := json.Marshal(invocation)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	file := filepath.Join(dir, hex.EncodeToString(digest[:])+".json")
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- digest filename under admitted workspace
	if os.IsExist(err) {
		info, statErr := os.Lstat(file)
		if statErr != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("invalid retained invocation file")
		}
		prior, readErr := os.ReadFile(file) // #nosec G304 -- checked private invocation
		if readErr != nil || !bytes.Equal(prior, raw) {
			return "", fmt.Errorf("existing invocation file differs")
		}
		retained, openErr := os.Open(file) // #nosec G304 -- verified private invocation
		if openErr != nil {
			return "", openErr
		}
		if err := errors.Join(retained.Sync(), retained.Close()); err != nil {
			return "", err
		}
		return file, syncCanonicalInvocationDirectories(dir, invocation.Workspace)
	}
	if err != nil {
		return "", err
	}
	writeErr := writeCanonicalDocument(f, raw)
	syncErr := f.Sync()
	closeErr := f.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return "", err
	}
	return file, syncCanonicalInvocationDirectories(dir, invocation.Workspace)
}

func syncCanonicalInvocationDirectories(dir, workspace string) error {
	for _, path := range []string{dir, workspace} {
		d, err := os.Open(path) // #nosec G304 -- admitted private invocation directory and its owning workspace
		if err != nil {
			return err
		}
		if err := errors.Join(d.Sync(), d.Close()); err != nil {
			return err
		}
	}
	return nil
}

func executeCanonicalBatch(cmd *cobra.Command, batch *canonicalbatch.Batch) error {
	if !graphPreviewActive || graphPreviewConfig == nil || graphPreviewConfig.GraphSchemaVersion != graphstore.AtomicBatchSchemaVersion {
		return graphFailure("capability_unavailable", "canonical batch execution requires an explicitly initialized --atomic-batch workspace", 5)
	}
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	opts, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if jsonOutput {
		opts.JSON = true
	}
	renderer, err := canonicalBatchRenderer(opts)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	cfg := graphPreviewConfig
	if canonicalBatchRetry == nil {
		scope, err := canonicalAuthoringScope(cmd)
		if err != nil {
			return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
		}
		if scope != cfg.GraphScopeURL {
			return canonicalFailure(cmd, renderer, "not_authority", "selected Scope differs from the actual workspace", 5)
		}
	}
	var invocation canonicalBatchInvocation
	if canonicalBatchRetry != nil {
		invocation = *canonicalBatchRetry
		if invocation.Workspace != cfg.GraphWorkspace || invocation.Scope != cfg.GraphScopeURL || invocation.Authority != cfg.GraphAuthorityID || invocation.Database != cfg.DoltDatabase || invocation.Schema != cfg.GraphSchemaVersion {
			return canonicalFailure(cmd, renderer, "not_authority", "recovery binding differs from the actual workspace", 5)
		}
	} else {
		key, _ := cmd.Flags().GetString("idempotency-key")
		if key == "" {
			key = rand.Text()
		}
		invocation = canonicalBatchInvocation{Format: "canonical-batch-invocation-1", Workspace: cfg.GraphWorkspace, Scope: cfg.GraphScopeURL, Authority: cfg.GraphAuthorityID, Database: cfg.DoltDatabase, Schema: cfg.GraphSchemaVersion, Key: key, Actor: getActor(), Request: batch.NativeJSON()}
	}
	journal, err := saveCanonicalBatchInvocation(invocation)
	if err != nil {
		return canonicalFailure(cmd, renderer, "invocation_not_saved", err.Error(), 1)
	}
	ctx, cancel := context.WithTimeout(getRootContext(), 2*time.Minute)
	defer cancel()
	store, err := openSelectedGraphStore(ctx, graphOptions(cfg))
	if err != nil {
		return canonicalFailure(cmd, renderer, "graph_not_initialized", err.Error(), 5)
	}
	result, applyErr := store.ApplyCanonicalBatch(ctx, batch, invocation.Key, invocation.Actor)
	closeErr := store.Close()
	if applyErr != nil || closeErr != nil {
		err = errors.Join(applyErr, closeErr)
		if closeErr != nil && applyErr == nil {
			err = errors.Join(graphstore.ErrOutcomeUnknown, err)
		}
		if errors.Is(err, graphstore.ErrOutcomeUnknown) {
			problem, _ := json.Marshal(map[string]string{"code": "outcome_unknown", "message": "batch outcome is unknown; recover the saved invocation with the same key"})
			raw, renderErr := renderer.RenderUnknown("cli", problem)
			if renderErr == nil && len(raw) > 0 {
				renderErr = writeCanonicalDocument(cmd.OutOrStdout(), raw)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Batch recovery file: %s\n", journal)
			if renderErr != nil {
				return renderErr
			}
			return &exitError{Code: 1}
		}
		code, status := "batch_failed", 1
		switch {
		case errors.Is(err, canonicalbatchstate.ErrRevision):
			code, status = "revision_mismatch", 4
		case errors.Is(err, canonicalbatchstate.ErrNotFound), errors.Is(err, graphstore.ErrNotFound):
			code, status = "not_found", 3
		case errors.Is(err, graphstore.ErrCapabilityUnavailable), errors.Is(err, canonicalbatchstate.ErrUnsupported):
			code, status = "capability_unavailable", 5
		case errors.Is(err, graphstore.ErrBatchKeyConflict):
			code = "idempotency_conflict"
		}
		return canonicalFailure(cmd, renderer, code, err.Error(), status)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	raw, err = renderer.RenderSuccess(raw)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Batch committed; recover its retained result with --retry-file %s\n", journal)
		return err
	}
	if err := writeCanonicalDocument(cmd.OutOrStdout(), raw); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Batch committed; recover its retained result with --retry-file %s\n", journal)
		return err
	}
	return nil
}
