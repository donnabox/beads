package graphstore

import (
	"bytes"
	"fmt"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

// Native rows/readiness have settled; deleted old allocations have been
// detached. Retain all final native Link heads before any owner's one mint.
func (run *atomicBatchRun) retainNativeDependencies() error {
	if run.nativeDependencyFinal == nil {
		return nil
	}
	for _, change := range run.nativeDependencyFinal.RowChanges() {
		lineage, ok := run.nativeDependencyPlan.lineage[change.Pair]
		if !ok {
			return ErrInvalidStore
		}
		if change.After == nil {
			continue
		}
		resource, ok := run.plan.Final[lineage.AfterID]
		if !ok || resource.Kind != graph.KindLink {
			return ErrInvalidStore
		}
		metadata, err := commonMetadata([]byte(change.After.Metadata))
		if err != nil {
			return err
		}
		if !bytes.Equal(metadata, resource.Metadata) {
			return fmt.Errorf("%w: native Dependency metadata did not retain staged intent", ErrInvalidStore)
		}
		path := run.paths[resource.ID]
		revision, err := freshToken()
		if err != nil {
			return err
		}
		attribution := writeAttribution(run.actor)
		if change.Transition == issueops.CanonicalDependencyMetadataTransition {
			before, ok := run.initial[lineage.BeforeID].(LinkRecord)
			if !ok {
				return ErrInvalidStore
			}
			attribution.Inception = before.Attribution.Inception
			if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_catalog SET revision=? WHERE path=?`, revision, path); err != nil {
				return err
			}
		} else {
			if err := run.store.ensureNativeTypeInTx(run.ctx, run.tx, resource.Type.URI, graph.KindLink); err != nil {
				return err
			}
			var publicID string
			path, publicID, err = run.newPath(resource.ID, graph.KindLink)
			if err != nil {
				return err
			}
			if _, err := run.tx.ExecContext(run.ctx, `INSERT INTO graph_preview_catalog(path,resource_kind,type_url,revision,allocation_state,backing,backing_key,public_id) VALUES(?,'link',?,?,'live','dependency',?,?)`, path, resource.Type.URI, revision, change.After.ID, publicID); err != nil {
				return err
			}
			// A fresh assertion has actual native creation facts. It does not inherit
			// its predecessor's attribution simply because the deterministic key recurs.
			attribution = Attribution{Actor: change.After.CreatedBy, Status: "claimed", RecordedAt: change.After.CreatedAt.UTC().Format(time.RFC3339Nano)}
		}
		link := LinkRecord{ID: resource.ID, Type: resource.Type.URI, Revision: revision, Version: revision, Source: resource.Source.URI, Target: resource.Target.URI, Properties: map[string]any{"dep_type": string(change.After.Type)}, Metadata: metadata, Attribution: attribution}
		snapshot, err := canonicalJSON(link)
		if err != nil {
			return err
		}
		if err := insertPreviewVersionInTx(run.ctx, run.tx, path, revision, snapshot, run.actor); err != nil {
			return err
		}
		if change.Transition == issueops.CanonicalDependencyMetadataTransition {
			raw, err := canonicalJSON(attribution)
			if err != nil {
				return err
			}
			if _, err := run.tx.ExecContext(run.ctx, `UPDATE graph_preview_versions SET native_attribution=? WHERE path=? AND version=?`, raw, path, revision); err != nil {
				return err
			}
		}
		run.final[resource.ID] = link
		run.nativeChanged[resource.Source.URI] = true
		if err := run.store.afterStage("batch-native-link-retained"); err != nil {
			return err
		}
	}
	// All mappings now exist. Independently compare every surviving changed
	// Link through the ordinary checked reader before retaining its owner.
	for _, change := range run.nativeDependencyFinal.RowChanges() {
		if change.After == nil {
			continue
		}
		id := run.nativeDependencyPlan.lineage[change.Pair].AfterID
		shown, err := run.store.showLinkInTx(run.ctx, run.tx, run.paths[id])
		if err != nil {
			return err
		}
		actual, err := canonicalJSON(shown)
		if err != nil {
			return err
		}
		expected, err := canonicalJSON(run.final[id])
		if err != nil {
			return err
		}
		if !bytes.Equal(actual, expected) {
			return fmt.Errorf("%w: final native Link differs from ordinary projection", ErrInvalidStore)
		}
	}
	return nil
}
