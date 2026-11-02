package graphstore

import (
	"bytes"
	"database/sql"
	"fmt"
	"slices"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/storage/depid"
	"github.com/steveyegge/beads/internal/storage/issueops"
)

type atomicNativeDependencyLineage struct{ BeforeID, AfterID string }
type atomicNativeDependencyPlan struct {
	request issueops.CanonicalDependencyFinalRequest
	lineage map[issueops.CanonicalDependencyPair]atomicNativeDependencyLineage
}

// Build native pair transitions from complete committed/final public identities,
// never from command order or pair-key equality alone. Allocated-only creates
// intentionally have neither a baseline nor final native pair.
func (run *atomicBatchRun) planNativeDependencies() (*atomicNativeDependencyPlan, error) {
	plan := &atomicNativeDependencyPlan{request: issueops.CanonicalDependencyFinalRequest{Actor: run.actor}, lineage: map[issueops.CanonicalDependencyPair]atomicNativeDependencyLineage{}}
	seenIssues := map[string]bool{}
	for _, record := range run.initial {
		issue, ok := record.(IssueRecord)
		if !ok {
			continue
		}
		if issue.Properties == nil || seenIssues[issue.Properties.ID] {
			return nil, fmt.Errorf("%w: ambiguous native Issue population", ErrInvalidStore)
		}
		seenIssues[issue.Properties.ID] = true
		plan.request.Population = append(plan.request.Population, issue.Properties.ID)
	}
	slices.SortFunc(plan.request.Population, graph.CompareCodeUnits)
	initial := map[issueops.CanonicalDependencyPair]canonicalbatchstate.Resource{}
	final := map[issueops.CanonicalDependencyPair]canonicalbatchstate.Resource{}
	for id, record := range run.initial {
		if run.backings[id] != "dependency" {
			continue
		}
		resource, err := atomicBatchStateRecord(record)
		if err != nil {
			return nil, err
		}
		pair, err := run.nativeDependencyPair(resource)
		if err != nil {
			return nil, err
		}
		if _, exists := initial[pair]; exists {
			return nil, fmt.Errorf("%w: multiple public Links for a native pair", ErrInvalidStore)
		}
		initial[pair] = resource
	}
	for id, resource := range run.plan.Final {
		if resource.Kind != graph.KindLink || run.backings[id] != "dependency" {
			continue
		}
		pair, err := run.nativeDependencyPair(resource)
		if err != nil {
			return nil, err
		}
		if _, exists := final[pair]; exists {
			return nil, fmt.Errorf("%w: final native Dependency pair has multiple public identities", ErrAlreadyExists)
		}
		final[pair] = resource
	}
	pairs := map[issueops.CanonicalDependencyPair]bool{}
	for pair := range initial {
		pairs[pair] = true
	}
	for pair := range final {
		pairs[pair] = true
	}
	for pair := range pairs {
		before, had := initial[pair]
		after, has := final[pair]
		plan.lineage[pair] = atomicNativeDependencyLineage{BeforeID: before.ID, AfterID: after.ID}
		if !had {
			kind, err := nativeDependencyProperties(after.Properties)
			if err != nil {
				return nil, err
			}
			plan.request.Insert = append(plan.request.Insert, issueops.CanonicalDependencyInsertFinal{Pair: pair, Type: kind, Metadata: bytes.Clone(after.Metadata)})
			continue
		}
		kind, err := nativeDependencyProperties(before.Properties)
		if err != nil {
			return nil, err
		}
		existing := issueops.CanonicalExistingDependency{Pair: pair, NativeID: depid.New(pair.SourceID, pair.TargetID), Type: kind}
		if !has {
			plan.request.Delete = append(plan.request.Delete, existing)
			continue
		}
		afterKind, err := nativeDependencyProperties(after.Properties)
		if err != nil {
			return nil, err
		}
		if kind != afterKind {
			return nil, fmt.Errorf("%w: native Dependency reclassification is unavailable", ErrCapabilityUnavailable)
		}
		if before.ID != after.ID {
			if _, stillLive := run.plan.Final[before.ID]; stillLive {
				return nil, ErrInvalidStore
			}
			if _, alreadyAllocated := run.initial[after.ID]; alreadyAllocated {
				return nil, ErrAlreadyExists
			}
			plan.request.Reassert = append(plan.request.Reassert, issueops.CanonicalDependencyReassertFinal{Existing: existing, Metadata: bytes.Clone(after.Metadata)})
		} else if !bytes.Equal(before.Metadata, after.Metadata) {
			plan.request.Metadata = append(plan.request.Metadata, issueops.CanonicalDependencyMetadataFinal{Existing: existing, Metadata: bytes.Clone(after.Metadata)})
		}
	}
	// Normalize before any SQL to validate and bound even canceled/empty decisions.
	if _, err := issueops.NormalizeCanonicalDependencyFinalRequest(plan.request); err != nil {
		return nil, err
	}
	return plan, nil
}

func (run *atomicBatchRun) nativeDependencyPair(resource canonicalbatchstate.Resource) (issueops.CanonicalDependencyPair, error) {
	var pair issueops.CanonicalDependencyPair
	source, sok := run.initial[resource.Source.URI].(IssueRecord)
	target, tok := run.initial[resource.Target.URI].(IssueRecord)
	if !sok || !tok || source.Properties == nil || target.Properties == nil || resource.Source.Revision != "" || resource.Target.Revision != "" {
		return pair, fmt.Errorf("%w: native Dependency requires floating existing local Issues", ErrCapabilityUnavailable)
	}
	pair.SourceID, pair.TargetID = source.Properties.ID, target.Properties.ID
	return pair, nil
}

func (run *atomicBatchRun) prepareNativeDependencies() error {
	if !run.nativeDependencyAuthored {
		return nil
	}
	plan, err := run.planNativeDependencies()
	if err != nil {
		return err
	}
	// CurrentSnapshot already proves each retained head and both mapping
	// directions, including the schema's one_backing uniqueness. Independently
	// compare raw Issue population and require history participation; the native
	// mint deliberately skips legacy NULL-generation rows.
	rows, err := run.tx.QueryContext(run.ctx, `SELECT id, participation_generation, current_revision FROM issues LIMIT 1001`)
	if err != nil {
		return err
	}
	expected := map[string]bool{}
	for _, id := range plan.request.Population {
		expected[id] = true
	}
	seen := map[string]bool{}
	for rows.Next() {
		var id string
		var generation, revision sql.NullInt64
		if err := rows.Scan(&id, &generation, &revision); err != nil {
			_ = rows.Close()
			return err
		}
		if !expected[id] || seen[id] || !generation.Valid || generation.Int64 <= 0 || !revision.Valid || revision.Int64 <= 0 {
			_ = rows.Close()
			return fmt.Errorf("%w: native Dependency batch requires complete participating Issue history", ErrInvalidStore)
		}
		seen[id] = true
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("%w: incomplete native Issue authority", ErrInvalidStore)
	}
	run.nativeDependencyPlan = plan
	return nil
}

func (run *atomicBatchRun) applyNativeDependencies() error {
	if run.nativeDependencyPlan == nil {
		return nil
	}
	rows, err := issueops.ApplyCanonicalDependencyFinalRowsInTx(run.ctx, run.tx, run.nativeDependencyPlan.request)
	if err != nil {
		return err
	}
	if err := run.store.afterStage("batch-native-dependency-rows"); err != nil {
		return err
	}
	final, err := issueops.SettleCanonicalDependencyReadinessInTx(run.ctx, run.tx, rows)
	if err != nil {
		return err
	}
	run.nativeDependencyFinal = final
	return run.store.afterStage("batch-native-dependency-settled")
}
