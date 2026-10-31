package graphstore

import (
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
	if backing == "dependency" {
		return run.admitNativeDependencyOutcome(o)
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
	return true, ErrInvalidStore
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
	// This context describes the new revision. An ordinary batch update is
	// not another historical copy; the predecessor's retained context stays
	// untouched, as with recordIssueMappingWithContextInTx(..., nil).
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
	if run.nativeDependencyFinal != nil {
		if err := issueops.EmitCanonicalDependencyFinalFactsInTx(run.ctx, run.tx, run.nativeDependencyFinal); err != nil {
			return err
		}
	}
	for _, facts := range run.nativeFacts {
		if err := issueops.EmitCanonicalFinalFactsInTx(run.ctx, run.tx, facts); err != nil {
			return err
		}
	}
	return nil
}
