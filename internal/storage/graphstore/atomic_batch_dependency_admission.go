package graphstore

import (
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	publicops "github.com/steveyegge/beads/issueops"
)

func (run *atomicBatchRun) admitAuthoredNativeDependency(member canonicalbatch.Member, before *canonicalbatchstate.Resource) error {
	if run.actor == "" {
		return fmt.Errorf("%w: Dependency mutation requires an actor", storage.ErrValidation)
	}
	if member.Operation != "updateLink" && member.Operation != "deleteLink" || member.Properties != nil || member.Source.Present() || member.Target.Present() || member.Type.URI != "" && member.Type != before.Type {
		return fmt.Errorf("%w: native Dependency retype, endpoint or property edits are unavailable", ErrCapabilityUnavailable)
	}
	var changes []json.RawMessage
	if member.PropertiesChange != nil && json.Unmarshal(member.PropertiesChange, &changes) != nil || len(changes) != 0 {
		return fmt.Errorf("%w: native Dependency properties cannot change", ErrCapabilityUnavailable)
	}
	_, err := run.checkedNativeDependency(*before, member.Operation == "deleteLink")
	return err
}

// Admission sees every resolved authored state, even if a later member deletes
// the new identity or restores old values. It never mutates the staging input.
func (run *atomicBatchRun) admitResolvedNativeMutation(_ int, _ canonicalbatch.Member, before, after *canonicalbatchstate.Resource) error {
	resource := after
	if resource == nil {
		resource = before
	}
	if resource == nil || resource.Kind != graph.KindLink || !isConcreteNativeDependency(run.store.ScopeURL(), resource.Type.URI, run.nativeDescriptors...) {
		return nil
	}
	if _, err := run.checkedNativeDependency(*resource, after == nil); err != nil {
		return err
	}
	if run.backings == nil {
		run.backings = map[string]string{}
	}
	run.backings[resource.ID] = "dependency"
	run.nativeDependencyAuthored = true
	return nil
}

func (run *atomicBatchRun) checkedNativeDependency(resource canonicalbatchstate.Resource, removing bool) (string, error) {
	var pair issueops.CanonicalDependencyPair
	if run.actor == "" {
		return "", fmt.Errorf("%w: Dependency mutation requires an actor", storage.ErrValidation)
	}
	if resource.Source.Revision != "" || resource.Target.Revision != "" {
		return "", fmt.Errorf("%w: native Dependency endpoints must float", ErrCapabilityUnavailable)
	}
	kind, err := nativeDependencyProperties(resource.Properties)
	if err != nil {
		return "", err
	}
	typeURI, err := NativeDependencyTypeURL(run.store.ScopeURL(), string(kind))
	if err != nil || !kind.IsWellKnown() || resource.Type.URI != typeURI || !isConcreteNativeDependency(run.store.ScopeURL(), resource.Type.URI, run.nativeDescriptors...) {
		return "", fmt.Errorf("%w: native Dependency classification must match its concrete Type", ErrCapabilityUnavailable)
	}
	source, sourceOK := run.initial[resource.Source.URI].(IssueRecord)
	target, targetOK := run.initial[resource.Target.URI].(IssueRecord)
	if !sourceOK || !targetOK || source.Properties == nil || target.Properties == nil || source.Properties.Ephemeral || source.Properties.NoHistory || target.Properties.Ephemeral || target.Properties.NoHistory {
		return "", fmt.Errorf("%w: native Dependency requires existing durable local Issues", ErrCapabilityUnavailable)
	}
	if err := issueops.AuthorizeTemplateUpdate(source.Properties, publicops.UpdateRequest{}); err != nil {
		return "", err
	}
	pair = issueops.CanonicalDependencyPair{SourceID: source.Properties.ID, TargetID: target.Properties.ID}
	if removing {
		err = issueops.ValidateRemoveDependencyRequest(publicops.RemoveDependencyRequest{Actor: run.actor, IssueID: pair.SourceID, DependsOnID: pair.TargetID})
	} else {
		err = issueops.ValidateAddDependenciesRequest(publicops.AddDependenciesRequest{Actor: run.actor, Edges: []publicops.DependencyEdge{{IssueID: pair.SourceID, DependsOnID: pair.TargetID, Type: kind}}})
	}
	if err != nil {
		return "", err
	}
	if _, err := commonMetadata(resource.Metadata); err != nil {
		return "", err
	}
	return source.ID, nil
}

func (run *atomicBatchRun) admitNativeDependencyOutcome(o canonicalbatchstate.Outcome) (bool, error) {
	if o.Outcome == "allocated" {
		return true, nil
	}
	resource := o.After
	if resource == nil {
		resource = o.Before
	}
	if resource == nil {
		return true, ErrInvalidStore
	}
	source, err := run.checkedNativeDependency(*resource, o.After == nil)
	if err != nil {
		return true, err
	}
	if o.Before != nil && o.After != nil && (o.Before.Type != o.After.Type || o.Before.Source != o.After.Source || o.Before.Target != o.After.Target || string(o.Before.Properties) != string(o.After.Properties)) {
		return true, fmt.Errorf("%w: native Dependency identity/classification is immutable in this batch candidate", ErrCapabilityUnavailable)
	}
	run.nativeDependencyAuthored = true
	run.ownerIDs[o.ID] = []string{source}
	run.ownerCandidates[source] = true
	return true, nil
}
