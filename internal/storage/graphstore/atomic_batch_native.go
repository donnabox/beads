package graphstore

import (
	"bytes"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

func (run *atomicBatchRun) admitNativeOutcome(o canonicalbatchstate.Outcome) (bool, error) {
	backing := run.backings[o.ID]
	if backing != "issue" && backing != "dependency" {
		return false, nil
	}
	if o.Before == nil || o.After == nil || o.Before.Type != o.After.Type {
		return true, fmt.Errorf("%w: native creation, deletion and retype require their final publisher", ErrCapabilityUnavailable)
	}
	if _, err := commonMetadata(o.After.Metadata); err != nil {
		return true, err
	}
	if backing == "issue" {
		before, ok := run.initial[o.ID].(IssueRecord)
		if !ok || o.Kind != graph.KindBead {
			return true, ErrInvalidStore
		}
		if o.Changed {
			content, err := canonicalNativeIssueContent(before, *o.After)
			if err != nil {
				return true, err
			}
			run.nativeContents[o.ID] = content
		}
		run.ownerCandidates[o.ID] = true
		return true, nil
	}
	before, ok := run.initial[o.ID].(LinkRecord)
	if !ok || o.Kind != graph.KindLink || o.Before.Source != o.After.Source || o.Before.Target != o.After.Target || !bytes.Equal(o.Before.Properties, o.After.Properties) {
		return true, fmt.Errorf("%w: native Dependency final edit is metadata-only", ErrCapabilityUnavailable)
	}
	run.ownerIDs[o.ID] = []string{before.Source}
	run.ownerCandidates[before.Source] = true
	return true, nil
}

func (run *atomicBatchRun) writeNativeIssue(o canonicalbatchstate.Outcome) error {
	before, ok := run.initial[o.ID].(IssueRecord)
	if !ok || before.Properties == nil {
		return ErrInvalidStore
	}
	content, ok := run.nativeContents[o.ID]
	if !ok {
		return ErrInvalidStore
	}
	facts, err := issueops.ApplyCanonicalIssueContentInTx(run.ctx, run.tx, before.Properties.ID, content, run.actor)
	if err != nil {
		return err
	}
	run.nativeFacts = append(run.nativeFacts, facts)
	run.nativeChanged[o.ID] = run.nativeChanged[o.ID] || facts.Changed()
	return run.store.afterStage("batch-native-issue-row")
}

func (run *atomicBatchRun) writeNativeLink(o canonicalbatchstate.Outcome) error {
	before, ok := run.initial[o.ID].(LinkRecord)
	if !ok || o.After == nil {
		return ErrInvalidStore
	}
	source, sourceOK := run.initial[before.Source].(IssueRecord)
	target, targetOK := run.initial[before.Target].(IssueRecord)
	if !sourceOK || !targetOK || source.Properties == nil || target.Properties == nil {
		return ErrInvalidStore
	}
	kind, err := nativeDependencyProperties(o.After.Properties)
	if err != nil {
		return err
	}
	facts, err := issueops.ApplyCanonicalDependencyMetadataInTx(run.ctx, run.tx, issueops.CanonicalDependencyMetadata{IssueID: source.Properties.ID, DependsOnID: target.Properties.ID, Type: kind, Metadata: o.After.Metadata}, run.actor)
	if err != nil {
		return err
	}
	run.nativeFacts = append(run.nativeFacts, facts)
	if !facts.Changed() {
		return nil
	}
	var stored []byte
	if err := run.tx.QueryRowContext(run.ctx, `SELECT d.metadata FROM dependencies d JOIN graph_preview_catalog c ON c.backing='dependency' AND c.backing_key=d.id WHERE c.path=?`, run.paths[o.ID]).Scan(&stored); err != nil {
		return err
	}
	metadata, err := commonMetadata(stored)
	if err != nil {
		return err
	}
	if !bytes.Equal(metadata, o.After.Metadata) {
		return fmt.Errorf("%w: native Dependency did not retain exact metadata intent", ErrInvalidStore)
	}
	link := before
	link.Metadata = metadata
	link.Revision, err = freshToken()
	if err != nil {
		return err
	}
	link.Version = link.Revision
	link.Attribution = writeAttribution(run.actor)
	link.Attribution.Inception = before.Attribution.Inception
	path := run.paths[o.ID]
	if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_catalog SET revision=? WHERE path=?`, link.Revision, path); err != nil {
		return err
	}
	snapshot, err := canonicalJSON(link)
	if err != nil {
		return err
	}
	if err := insertPreviewVersionInTx(run.ctx, run.tx, path, link.Revision, snapshot, run.actor); err != nil {
		return err
	}
	attribution, err := canonicalJSON(link.Attribution)
	if err != nil {
		return err
	}
	if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_versions SET native_attribution=? WHERE path=? AND version=?`, attribution, path, link.Revision); err != nil {
		return err
	}
	run.final[o.ID], run.nativeChanged[before.Source] = link, true
	return run.store.afterStage("batch-native-link-retained")
}

func (run *atomicBatchRun) finishNativeIssue(id string) error {
	before, ok := run.initial[id].(IssueRecord)
	if !ok || before.Properties == nil {
		return ErrInvalidStore
	}
	owned, err := run.store.ownedLinksInTx(run.ctx, run.tx, before.Properties.ID)
	if err != nil {
		return err
	}
	if !run.nativeChanged[id] && equalBatchOwned(before.Owned, owned) {
		run.final[id] = before
		return nil
	}
	previous, err := currentIssueRevisionContextInTx(run.ctx, run.tx, run.paths[id])
	if err != nil {
		return err
	}
	previous.HistoricalUpdate = nil
	if err := issueops.RecordVersionInTx(run.ctx, run.tx, before.Properties.ID, run.actor); err != nil {
		return err
	}
	var ordinal int64
	if err := run.tx.QueryRowContext(run.ctx, `SELECT current_revision FROM issues WHERE id=?`, before.Properties.ID).Scan(&ordinal); err != nil {
		return err
	}
	if err := run.store.recordIssueMappingSnapshotInTx(run.ctx, run.tx, run.paths[id], before.Properties.ID, ordinal, owned, previous); err != nil {
		return err
	}
	after, err := run.store.showIssueInTx(run.ctx, run.tx, run.paths[id])
	if err != nil {
		return err
	}
	if after.Type != before.Type {
		return fmt.Errorf("%w: native final publisher changed classification", ErrInvalidStore)
	}
	run.final[id] = after
	return nil
}

func (run *atomicBatchRun) emitNativeFacts() error {
	for _, facts := range run.nativeFacts {
		if err := issueops.EmitCanonicalFinalFactsInTx(run.ctx, run.tx, facts); err != nil {
			return err
		}
	}
	return nil
}
