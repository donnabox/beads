package graphstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

func (run *atomicBatchRun) publish() error {
	unscope := issueops.ScopeVersionedHistoryTransaction(run.tx, true)
	defer unscope()
	for id, value := range run.initial {
		run.final[id] = value
	}
	// New Beads must have allocations and final payloads before final Link
	// records can resolve their endpoints. Their one retained version follows
	// final Links, so its capture contains their actual published revisions.
	for _, o := range run.plan.Outcomes {
		if o.After == nil || o.Kind != graph.KindBead || !o.Changed {
			continue
		}
		if run.backings[o.ID] == "issue" {
			if err := run.writeNativeIssue(o); err != nil {
				return err
			}
			continue
		}
		r := *o.After
		var properties Properties
		if err := json.Unmarshal(r.Properties, &properties); err != nil {
			return err
		}
		if o.Before == nil {
			path, publicID, err := run.newPath(r.ID, graph.KindBead)
			if err != nil {
				return err
			}
			revision, err := freshToken()
			if err != nil {
				return err
			}
			if _, err := run.tx.ExecContext(run.ctx, `INSERT INTO graph_preview_catalog(path,resource_kind,type_url,revision,allocation_state,backing,public_id) VALUES(?,'bead',?,?,'live','generic',?)`, path, r.Type.URI, revision, publicID); err != nil {
				return err
			}
			if _, err := run.tx.ExecContext(run.ctx, `INSERT INTO graph_preview_payloads(path,properties,metadata) VALUES(?,?,?)`, path, r.Properties, r.Metadata); err != nil {
				return err
			}
			attribution := writeAttribution(run.actor)
			attribution.Inception = "create"
			run.final[r.ID] = Record{PublicID: publicID, ID: r.ID, Type: r.Type.URI, Revision: revision, Version: revision, Properties: properties, Metadata: r.Metadata, Owned: []json.RawMessage{}, Attribution: attribution}
		} else {
			before, ok := run.initial[r.ID].(Record)
			if !ok {
				return ErrCapabilityUnavailable
			}
			before.Type = r.Type.URI
			before.Properties = properties
			before.Metadata = r.Metadata
			run.final[r.ID] = before
			if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_catalog SET type_url=? WHERE path=?`, r.Type.URI, run.paths[r.ID]); err != nil {
				return err
			}
			if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_payloads SET properties=?,metadata=? WHERE path=?`, r.Properties, r.Metadata, run.paths[r.ID]); err != nil {
				return err
			}
		}
	}
	if err := run.applyNativeDependencies(); err != nil {
		return err
	}
	// Remove Links before Beads, without publishing an intermediate owner.
	for _, o := range run.plan.Outcomes {
		if o.Outcome == "deleted" && o.Kind == graph.KindLink {
			if err := run.deleteLink(o); err != nil {
				return err
			}
			delete(run.final, o.ID)
		}
	}
	for _, o := range run.plan.Outcomes {
		if o.Outcome == "deleted" && o.Kind == graph.KindBead {
			if err := run.deleteBead(o); err != nil {
				return err
			}
			delete(run.final, o.ID)
		}
	}
	if err := run.retainNativeDependencies(); err != nil {
		return err
	}
	// Persist all final Link rows before retaining them. Correcting one source
	// never versions an intermediate source or re-pins a literal endpoint.
	for _, o := range run.plan.Outcomes {
		if o.After == nil || o.Kind != graph.KindLink || !o.Changed {
			continue
		}
		if run.backings[o.ID] == "dependency" {
			continue
		}
		if err := run.writeLink(o); err != nil {
			return err
		}
	}
	for _, o := range run.plan.Outcomes {
		if o.After == nil || o.Kind != graph.KindLink || !o.Changed || run.backings[o.ID] == "dependency" {
			continue
		}
		link, err := run.store.currentInformationalLinkInTx(run.ctx, run.tx, run.paths[o.ID])
		if err != nil {
			return err
		}
		snapshot, err := canonicalJSON(link)
		if err != nil {
			return err
		}
		if err := insertPreviewVersionInTx(run.ctx, run.tx, run.paths[o.ID], link.Revision, snapshot, run.actor); err != nil {
			return err
		}
		run.final[o.ID] = link
	}
	ids := make([]string, 0, len(run.ownerCandidates))
	for id := range run.ownerCandidates {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return graph.CompareCodeUnits(ids[i], ids[j]) < 0 })
	for _, id := range ids {
		if _, survives := run.plan.Final[id]; !survives {
			continue
		}
		if run.backings[id] == "issue" {
			if err := run.finishNativeIssue(id); err != nil {
				return err
			}
		} else if err := run.finishBead(id); err != nil {
			return err
		}
	}
	// Allocations are independent from live/retained Resource versions. An
	// allocated-only identity has no catalog row, body, affiliation or revision.
	for _, id := range run.plan.ConsumedIDs {
		local, err := atomicBatchLocalID(run.store.ScopeURL(), id)
		if err != nil {
			return err
		}
		var kind graph.ResourceKind
		for _, o := range run.plan.Outcomes {
			if o.ID == id {
				kind = o.Kind
				break
			}
		}
		if !kind.Valid() {
			return ErrInvalidStore
		}
		if _, err := run.tx.ExecContext(run.ctx, `INSERT INTO graph_preview_batch_allocations(public_id,resource_kind) VALUES(?,?)`, local, kind); err != nil {
			return err
		}
	}
	if err := run.emitNativeFacts(); err != nil {
		return err
	}
	return run.store.afterStage("batch-final-state")
}

func (run *atomicBatchRun) newPath(id string, kind graph.ResourceKind) (string, string, error) {
	local, err := atomicBatchLocalID(run.store.ScopeURL(), id)
	if err != nil {
		return "", "", err
	}
	exists, err := publicResourceIDExistsInTx(run.ctx, run.tx, local)
	if err != nil {
		return "", "", err
	}
	if exists {
		return "", "", publicResourceIDReservedErrorInTx(run.ctx, run.tx, local)
	}
	preferred, err := preferredPrivateResourcePath(kind, local)
	if err != nil {
		return "", "", err
	}
	path, err := choosePrivateResourcePathInTx(run.ctx, run.tx, preferred, true, kind)
	if err != nil {
		return "", "", err
	}
	run.paths[id] = path
	return path, local, nil
}

func (run *atomicBatchRun) writeLink(o canonicalbatchstate.Outcome) error {
	r := *o.After
	path := run.paths[r.ID]
	publicID := ""
	if o.Before == nil {
		var err error
		path, publicID, err = run.newPath(r.ID, graph.KindLink)
		if err != nil {
			return err
		}
	}
	revision, err := freshToken()
	if err != nil {
		return err
	}
	attribution := writeAttribution(run.actor)
	if before, ok := run.initial[r.ID].(LinkRecord); ok {
		attribution.Inception = before.Attribution.Inception
	}
	encodedAttribution, err := canonicalJSON(attribution)
	if err != nil {
		return err
	}
	sourcePath, sourceOK := run.paths[r.Source.URI]
	targetPath, targetOK := run.paths[r.Target.URI]
	if !sourceOK || !targetOK {
		return fmt.Errorf("%w: missing final Link endpoint allocation", ErrInvalidStore)
	}
	if o.Before == nil {
		if _, err := run.tx.ExecContext(run.ctx, `INSERT INTO graph_preview_catalog(path,resource_kind,type_url,revision,allocation_state,backing,public_id) VALUES(?,'link',?,?,'live','informational',?)`, path, r.Type.URI, revision, publicID); err != nil {
			return err
		}
		_, err = run.tx.ExecContext(run.ctx, `INSERT INTO graph_preview_links(path,source_path,target_path,source_revision,target_revision,properties,metadata,attribution) VALUES(?,?,?,?,?,?,?,?)`, path, sourcePath, targetPath, r.Source.Revision, r.Target.Revision, r.Properties, r.Metadata, encodedAttribution)
	} else {
		if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_catalog SET type_url=?,revision=? WHERE path=?`, r.Type.URI, revision, path); err != nil {
			return err
		}
		_, err = run.tx.ExecContext(run.ctx, `UPDATE graph_preview_links SET source_path=?,target_path=?,source_revision=?,target_revision=?,properties=?,metadata=?,attribution=? WHERE path=?`, sourcePath, targetPath, r.Source.Revision, r.Target.Revision, r.Properties, r.Metadata, encodedAttribution, path)
	}
	return err
}

func (run *atomicBatchRun) finishBead(id string) error {
	memory, ok := run.final[id].(Record)
	if !ok {
		return fmt.Errorf("%w: final generic Bead absent", ErrInvalidStore)
	}
	owned, err := run.store.memoryOwnedLinksInTx(run.ctx, run.tx, run.paths[id])
	if err != nil {
		return err
	}
	memory.Owned = owned
	before, existing := run.initial[id].(Record)
	if existing && sameBatchMemoryCovered(before, memory) {
		run.final[id] = before
		return nil
	}
	if err := run.store.requireAssessmentAuthoring(memory.Type); err != nil {
		return err
	}
	if existing {
		memory.Revision, err = freshToken()
		if err != nil {
			return err
		}
		memory.Version = memory.Revision
		attribution := writeAttribution(run.actor)
		attribution.Inception = before.Attribution.Inception
		memory.Attribution = attribution
	}
	if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_catalog SET revision=? WHERE path=?`, memory.Revision, run.paths[id]); err != nil {
		return err
	}
	snapshot, err := canonicalJSON(memory)
	if err != nil {
		return err
	}
	if err := insertPreviewVersionInTx(run.ctx, run.tx, run.paths[id], memory.Revision, snapshot, run.actor); err != nil {
		return err
	}
	run.final[id] = memory
	return nil
}

func sameBatchMemoryCovered(a, b Record) bool {
	return a.ID == b.ID && a.Type == b.Type && a.Properties == b.Properties && bytes.Equal(a.Metadata, b.Metadata) && equalBatchOwned(a.Owned, b.Owned)
}

func equalBatchOwned(a, b []json.RawMessage) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

func (run *atomicBatchRun) deleteLink(o canonicalbatchstate.Outcome) error {
	before, ok := run.initial[o.ID].(LinkRecord)
	if !ok {
		return ErrCapabilityUnavailable
	}
	revision, err := freshToken()
	if err != nil {
		return err
	}
	attribution := writeAttribution(run.actor)
	attribution.Inception = before.Attribution.Inception
	tombstone := LinkTombstone{ID: before.ID, Type: before.Type, Revision: revision, Version: revision, State: "deleted", PreviousVersion: before.Revision, Attribution: attribution}
	path := run.paths[o.ID]
	if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_catalog SET allocation_state='deleted',revision=? WHERE path=?`, revision, path); err != nil {
		return err
	}
	if run.backings[o.ID] == "dependency" {
		if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_catalog SET backing_key=NULL WHERE path=?`, path); err != nil {
			return err
		}
	}
	if _, err := run.tx.ExecContext(run.ctx, `DELETE FROM graph_preview_links WHERE path=?`, path); err != nil {
		return err
	}
	snapshot, err := canonicalJSON(tombstone)
	if err != nil {
		return err
	}
	return insertPreviewVersionInTx(run.ctx, run.tx, path, revision, snapshot, run.actor)
}

func (run *atomicBatchRun) deleteBead(o canonicalbatchstate.Outcome) error {
	before, ok := run.initial[o.ID].(Record)
	if !ok {
		return ErrCapabilityUnavailable
	}
	revision, err := freshToken()
	if err != nil {
		return err
	}
	attribution := writeAttribution(run.actor)
	attribution.Inception = before.Attribution.Inception
	terminal := MemoryTerminalState{ID: before.ID, Revision: revision, Version: revision, State: "deleted", PreviousVersion: before.Revision, Attribution: attribution}
	path := run.paths[o.ID]
	if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_catalog SET allocation_state='deleted',revision=? WHERE path=?`, revision, path); err != nil {
		return err
	}
	if _, err := run.tx.ExecContext(run.ctx, `DELETE FROM graph_preview_payloads WHERE path=?`, path); err != nil {
		return err
	}
	snapshot, err := canonicalJSON(terminal)
	if err != nil {
		return err
	}
	return insertPreviewVersionInTx(run.ctx, run.tx, path, revision, snapshot, run.actor)
}

func (run *atomicBatchRun) result() (BatchResult, error) {
	result := BatchResult{CapabilityVersion: canonicalbatch.Version, Results: []BatchOperationResult{}, Resources: []BatchResourceOutcome{}}
	var err error
	result.Results, err = atomicBatchOperationResults(run.members, run.plan)
	if err != nil {
		return result, err
	}
	outcomes := map[string]canonicalbatchstate.Outcome{}
	for _, o := range run.plan.Outcomes {
		outcomes[o.ID] = o
	}
	for id := range run.ownerCandidates {
		if _, direct := outcomes[id]; direct {
			continue
		}
		beforeState, err := atomicBatchStateRecord(run.initial[id])
		if err != nil {
			return result, err
		}
		after := beforeState
		outcomes[id] = canonicalbatchstate.Outcome{ID: id, Kind: graph.KindBead, Before: &beforeState, After: &after}
	}
	ids := make([]string, 0, len(outcomes))
	for id := range outcomes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return graph.CompareCodeUnits(ids[i], ids[j]) < 0 })
	for _, id := range ids {
		o := outcomes[id]
		row := BatchResourceOutcome{ID: id, ResourceKind: o.Kind, Outcome: o.Outcome, Changed: o.Changed}
		if o.After != nil {
			value, ok := run.final[id]
			if !ok {
				return result, ErrInvalidStore
			}
			var err error
			row.Resource, err = canonicalJSON(value)
			if err != nil {
				return result, err
			}
			if o.Before != nil {
				old, err := canonicalJSON(run.initial[id])
				if err != nil {
					return result, err
				}
				row.Changed = !bytes.Equal(old, row.Resource)
				if row.Changed {
					row.Outcome = "updated"
				} else {
					row.Outcome = "unchanged"
				}
			}
		} else if o.Outcome == "deleted" {
			row.Deleted = &BatchDeletedIdentity{ID: id, ResourceKind: o.Kind, Type: o.Before.Type.URI, Revision: o.Before.Revision}
		}
		for _, ownerID := range run.ownerIDs[id] {
			value, live := run.final[ownerID]
			if !live {
				// ownerRevisions means the final published live revision:
				// for an owner deleted in this decision, that is its committed
				// baseline. Never expose a staged revision or the private
				// terminal bookkeeping token in this field.
				value = run.initial[ownerID]
			}
			state, err := atomicBatchStateRecord(value)
			if err != nil {
				return result, err
			}
			row.OwnerRevisions = append(row.OwnerRevisions, BatchOwnerRevision{ownerID, state.Revision})
		}
		result.Changed = result.Changed || row.Changed
		result.Resources = append(result.Resources, row)
	}
	if result.Changed {
		var err error
		result.Transaction, err = freshToken()
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
