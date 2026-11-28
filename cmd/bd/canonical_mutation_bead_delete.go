package main

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/spf13/cobra"
    graph "github.com/steveyegge/beads/graphops"
    "github.com/steveyegge/beads/internal/clirevision"
    "github.com/steveyegge/beads/internal/clioutput"
    "github.com/steveyegge/beads/internal/httpapi/bdpwire"
    "github.com/steveyegge/beads/internal/httpapi/graphread"
    "github.com/steveyegge/beads/internal/storage/graphstore"
)

// This is an internal checked-operation carrier, never a BDP mutation DTO.
// Preview contains an ordinary current Resource; Native is the actual writer's
// final result. No result is populated by a subsequent current-state reread.
type canonicalBeadDeleteOutcome struct {
    Preview bool
    Resource bdpwire.BeadRecord
    Native any
}

type canonicalBeadDeleteIntent struct {
    Identity string
    Apply bool
    Guarded bool
    Guard clirevision.Selector
    Actor string
}

func prepareCanonicalBeadDelete(cmd *cobra.Command, args []string) (canonicalBeadDeleteIntent, error) {
    if len(args) != 1 {
        return canonicalBeadDeleteIntent{}, canonicalMutationLeafInvalid("delete requires one unpinned Bead Reference")
    }
    identity, err := graphPreviewWritableIdentity(args[0])
    if err != nil { return canonicalBeadDeleteIntent{}, canonicalMutationLeafInvalid(err.Error()) }
    apply, err := cmd.Flags().GetBool("force")
    if err != nil { return canonicalBeadDeleteIntent{}, canonicalMutationLeafInvalid(err.Error()) }
    intent := canonicalBeadDeleteIntent{Identity: identity, Apply: apply, Guarded: cmd.Flags().Changed("if-revision")}
    if intent.Guarded {
        if !intent.Apply { return canonicalBeadDeleteIntent{}, canonicalMutationLeafInvalid("--if-revision requires delete --force; preview is an observational read") }
        raw, err := cmd.Flags().GetString("if-revision")
        if err != nil { return canonicalBeadDeleteIntent{}, canonicalMutationLeafInvalid(err.Error()) }
        intent.Guard, err = graphPreviewRevisionSelector("if-revision", raw)
        if err != nil { return canonicalBeadDeleteIntent{}, canonicalMutationLeafInvalid(err.Error()) }
    }
    return intent, nil
}

// Select exact lifetime allocation identity; a canonical local ID must never
// inherit the compatibility writer's abbreviation/private-path interpretation.
// The backing kind is immutable. Each writer rechecks its live state, guard,
// Type/ownership/fence rules and incident Links inside its own ordinary TX.
func executeCanonicalBeadDelete(ctx context.Context, store *graphstore.Store, intent canonicalBeadDeleteIntent) (canonicalBeadDeleteOutcome, error) {
    if err := ctx.Err(); err != nil { return canonicalBeadDeleteOutcome{}, err }
    if store == nil { return canonicalBeadDeleteOutcome{}, canonicalMutationLeafUnavailable("canonical delete requires its checked graph store") }
    var identity graphstore.AllocatedResourceIdentity
    var err error
    if strings.Contains(intent.Identity, "://") {
        identity, err = store.ResolveCanonicalPublicResourceIdentity(ctx, intent.Identity)
    } else {
        identity, err = store.ResolveAllocatedResourceIdentity(ctx, intent.Identity)
    }
    if err != nil { return canonicalBeadDeleteOutcome{}, err }
    if identity.Kind != graph.KindBead {
        return canonicalBeadDeleteOutcome{}, canonicalMutationLeafInvalid("delete requires a Bead; use link delete for a Link")
    }
    if !intent.Apply {
        // GET only: linked Beads remain previewable. Do not run mutation
        // validation or reserve a revision, allocation, journal or execution key.
        resource, err := graphread.New(store).Resource(ctx, identity.PrivatePath)
        if err != nil { return canonicalBeadDeleteOutcome{}, err }
        bead, ok := resource.(bdpwire.BeadRecord)
        if !ok { return canonicalBeadDeleteOutcome{}, canonicalMutationLeafUnavailable("selected Bead has no checked Resource projection") }
        return canonicalBeadDeleteOutcome{Preview: true, Resource: bead}, nil
    }
    revision := intent.Guard.Token
    if intent.Guarded && intent.Guard.Ordinal != 0 {
        row, err := store.ResolveLocalRevision(ctx, identity.PrivatePath, intent.Guard.Ordinal)
        if err != nil { return canonicalBeadDeleteOutcome{}, err }
        revision = row.Version
    }
    if intent.Guarded && revision == "" {
        return canonicalBeadDeleteOutcome{}, canonicalMutationLeafInvalid("delete guard did not resolve to an exact revision")
    }
    // This read only selects immutable backing. Never use it as an afterimage,
    // authorize an incident cascade or replace the writer's checked guard.
    current, err := store.Read(ctx, identity.PrivatePath)
    if err != nil { return canonicalBeadDeleteOutcome{}, err }
    switch current.(type) {
    case graphstore.Record:
        result, err := store.DeleteMemory(ctx, graphstore.MemoryDeleteRequest{Path: identity.PrivatePath, Actor: intent.Actor, ExpectedRevision: revision, Unconditional: !intent.Guarded})
        if err != nil { return canonicalBeadDeleteOutcome{}, err }
        return canonicalBeadDeleteOutcome{Native: result}, nil
    case graphstore.IssueRecord:
        // Deliberately NOT DeleteIssueWithIncidentLinks: canonical deletion has
        // no implicit cascade. Leave that separate compatibility route intact.
        result, err := store.DeleteIssue(ctx, graphstore.IssueDeleteRequest{Path: identity.PrivatePath, Actor: intent.Actor, ExpectedRevision: revision, Unconditional: !intent.Guarded})
        if err != nil { return canonicalBeadDeleteOutcome{}, err }
        return canonicalBeadDeleteOutcome{Native: result}, nil
    default:
        return canonicalBeadDeleteOutcome{}, canonicalMutationLeafUnavailable("this Bead backing has no canonical checked deletion writer")
    }
}

func runCanonicalBeadDelete(cmd *cobra.Command, args []string) error {
    intent, err := prepareCanonicalBeadDelete(cmd, args)
    if err != nil { return canonicalMutationArgumentFailure(err) }
    if graphPreviewConfig == nil || !graphPreviewActive {
        return canonicalMutationArgumentFailure(canonicalMutationLeafUnavailable("canonical delete requires its checked graph workspace"))
    }
    if _, _, err := graphPreviewResourceReadSelectorIn(graphPreviewConfig.GraphScopeURL, intent.Identity); err != nil {
        return graphFailure("invalid_selector", err.Error(), 2)
    }
    if intent.Apply {
        if err := graphPreviewWritePolicy(); err != nil { return err }
        intent.Actor = getActorWithGit()
    }
    if ctx := cmd.Context(); ctx != nil && ctx.Err() != nil { return graphStorageError(ctx.Err()) }
    return withGraphStoreStageErrorMappers(30*time.Second, func(ctx context.Context, store *graphstore.Store) (any, string, error) {
        // Keep the existing bounded store lifetime while also observing an
        // explicitly supplied leaf cancellation; no second store or TX wrapper.
        opCtx, cancel := context.WithCancel(ctx)
        defer cancel()
        if leaf := cmd.Context(); leaf != nil {
            stop := context.AfterFunc(leaf, cancel)
            defer stop()
            if err := leaf.Err(); err != nil { return nil, "", err }
        }
        outcome, err := executeCanonicalBeadDelete(opCtx, store, intent)
        if err != nil { return nil, "", err }
        if outcome.Preview { return outcome, "Would delete "+outcome.Resource.ID, nil }
        return outcome, "Deleted "+intent.Identity, nil
    }, func(value any, human string) error {
        outcome, ok := value.(canonicalBeadDeleteOutcome)
        if !ok { return fmt.Errorf("canonical delete returned an unexpected result carrier") }
        if outcome.Preview && jsonOutput {
            raw, err := canonicalBeadDeletePreviewDocument(outcome.Resource)
            if err != nil { return err }
            return writeCanonicalDocument(cmd.OutOrStdout(), raw)
        }
        if outcome.Preview { return graphPrintTo(cmd.OutOrStdout(), outcome.Resource, human, quietFlag, false) }
        // Native result remains unchanged; complete canonical mutation envelope
        // and durable singleton delivery are separate outstanding adapter work.
        return graphPrintTo(cmd.OutOrStdout(), outcome.Native, human, quietFlag, jsonOutput)
    }, graphStorageError, canonicalBeadDeleteError)
}

func canonicalBeadDeleteError(err error) error {
    var refusal *canonicalMutationLeafRefusal
    if errors.As(err, &refusal) { return canonicalMutationArgumentFailure(err) }
    return graphStorageError(err)
}

// The marker is CLI context on the existing Resource envelope, never a member
// of the BDP Resource data. No public mutation result or preview reservation is
// invented. Shared general output-context plumbing remains a separate seam.
func canonicalBeadDeletePreviewDocument(resource bdpwire.BeadRecord) ([]byte, error) {
    renderer, err := clioutput.New(canonicalResourceDescriptor(), clioutput.Options{JSON: true})
    if err != nil { return nil, err }
    raw, err := json.Marshal(resource)
    if err != nil { return nil, err }
    raw, err = renderer.Render(raw)
    if err != nil { return nil, err }
    var envelope map[string]json.RawMessage
    if err := json.Unmarshal(raw, &envelope); err != nil { return nil, err }
    envelope["context"] = json.RawMessage(`{"preview":"delete"}`)
    raw, err = json.Marshal(envelope)
    if err != nil { return nil, err }
    return append(raw, '\n'), nil
}
