package issueops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// CanonicalIssueContent contains final, storage-ready values for an existing
// durable Issue. Omitted fields are preserved; set nullable fields may clear.
// It deliberately cannot express classification, lifecycle, assignment, notes,
// labels, leases, persistence, identity, or creation facts. It is a final-row
// projection, not another patch interpreter. The caller must separately admit
// every authored operation, including operations later canceled in the batch.
type CanonicalIssueContent struct {
	Title, Description, Design, AcceptanceCriteria publicops.Field[string]
	SpecID                                         publicops.Field[string]
	Priority                                       publicops.Field[int]
	EstimatedMinutes                               publicops.Field[*int]
	ExternalRef                                    publicops.Field[*string]
	DueAt                                          publicops.Field[*time.Time]
	Metadata                                       publicops.Field[json.RawMessage]
}

func (content CanonicalIssueContent) patch() publicops.IssuePatch {
	return publicops.IssuePatch{
		Title: content.Title, Description: content.Description, Design: content.Design,
		AcceptanceCriteria: content.AcceptanceCriteria, SpecID: content.SpecID,
		Priority: content.Priority, EstimatedMinutes: content.EstimatedMinutes,
		ExternalRef: content.ExternalRef, DueAt: content.DueAt,
		Metadata: publicops.MetadataPatch{Replace: content.Metadata},
	}
}

// ValidateCanonicalIssueContent checks the finite final-row vocabulary without
// SQL. Dates must already fit DATETIME(0); it never rounds the staged intent.
// Configured metadata-schema validation still runs inside the composite.
func ValidateCanonicalIssueContent(content CanonicalIssueContent) error {
	if err := ValidateUpdateRequest(publicops.UpdateRequest{Patch: content.patch()}); err != nil {
		return err
	}
	for _, field := range []struct {
		name  string
		value publicops.Field[string]
	}{
		{"title", content.Title}, {"description", content.Description},
		{"design", content.Design}, {"acceptance_criteria", content.AcceptanceCriteria},
		{"spec_id", content.SpecID},
	} {
		if field.value.Set && !utf8.ValidString(field.value.Value) {
			return fmt.Errorf("%w: Issue %s must be UTF-8", storage.ErrValidation, field.name)
		}
	}
	if content.SpecID.Set && utf8.RuneCountInString(content.SpecID.Value) > 1024 {
		return fmt.Errorf("%w: Issue spec_id exceeds its 1024-character column", storage.ErrValidation)
	}
	if content.ExternalRef.Set && content.ExternalRef.Value != nil {
		if !utf8.ValidString(*content.ExternalRef.Value) {
			return fmt.Errorf("%w: Issue external_ref must be UTF-8", storage.ErrValidation)
		}
		if err := types.CheckFieldLen("external_ref", *content.ExternalRef.Value); err != nil {
			return fmt.Errorf("%w: %w", storage.ErrValidation, err)
		}
	}
	if content.EstimatedMinutes.Set && content.EstimatedMinutes.Value != nil && *content.EstimatedMinutes.Value > math.MaxInt32 {
		return fmt.Errorf("%w: Issue estimated_minutes exceeds its signed INT column", storage.ErrValidation)
	}
	if content.DueAt.Set && content.DueAt.Value != nil {
		value := content.DueAt.Value.UTC()
		if value.Year() < 1 || value.Year() > 9999 || value.Nanosecond() != 0 {
			return fmt.Errorf("%w: Issue due_at must be a representable whole-second date", storage.ErrValidation)
		}
	}
	if content.Metadata.Set {
		return validateCanonicalMetadata(content.Metadata.Value)
	}
	return nil
}

func validateCanonicalMetadata(value json.RawMessage) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(value, &object); err != nil || object == nil {
		return fmt.Errorf("%w: canonical metadata must be a JSON object", storage.ErrValidation)
	}
	_, err := storage.NormalizeMetadataValue(value)
	return err
}

// CanonicalFinalFacts carries deferred native facts from one final row edit.
// Its contents are private so callers cannot fabricate audit/journal payloads.
// It belongs to one transaction and must be emitted exactly once after the
// caller has published final native versions and graph owner snapshots.
type CanonicalFinalFacts struct {
	tx                 *sql.Tx
	issueID, actor     string
	changed, emitted   bool
	oldValue, newValue string
	dependency         *types.Dependency
}

// Changed reports whether the native row changed; a no-op needs no version.
func (facts *CanonicalFinalFacts) Changed() bool { return facts != nil && facts.changed }

// IssueID identifies the native owner whose final version the caller publishes.
func (facts *CanonicalFinalFacts) IssueID() string {
	if facts == nil {
		return ""
	}
	return facts.issueID
}

func validateCanonicalIdentity(id, actor string) error {
	if id == "" || actor == "" || !utf8.ValidString(id) || !utf8.ValidString(actor) {
		return fmt.Errorf("%w: canonical native edit requires UTF-8 identity and actor", storage.ErrValidation)
	}
	return nil
}

// ApplyCanonicalIssueContentInTx writes an existing durable Issue once, using
// the native scalar/metadata validators, row fence and no-op behavior. It does
// not begin, commit, retry, mint a version, or emit intermediate facts. Any
// error requires the caller to roll back the entire enclosing transaction.
func ApplyCanonicalIssueContentInTx(ctx context.Context, tx *sql.Tx, id string, content CanonicalIssueContent, actor string) (*CanonicalFinalFacts, error) {
	if tx == nil {
		return nil, fmt.Errorf("%w: canonical edit requires a transaction", storage.ErrValidation)
	}
	if err := validateCanonicalIdentity(id, actor); err != nil {
		return nil, err
	}
	if err := ValidateCanonicalIssueContent(content); err != nil {
		return nil, err
	}
	before, err := getIssueFromTableInTx(ctx, tx, "issues", "labels", id)
	if err != nil {
		return nil, err
	}
	if before.Ephemeral || before.NoHistory || IsActiveWispInTx(ctx, tx, id) {
		return nil, fmt.Errorf("%w: canonical final edit requires a durable Issue", storage.ErrValidation)
	}
	if err := AuthorizeTemplateUpdate(before, publicops.UpdateRequest{}); err != nil {
		return nil, err
	}
	updates := UpdateFields(content.patch())
	metadata, changed, err := ApplyMetadataPatch(before.Metadata, content.patch().Metadata)
	if err != nil {
		return nil, err
	}
	if changed {
		updates["metadata"] = metadata
	}
	updates, err = DiscardNoopIssueUpdates(before, updates)
	if err != nil {
		return nil, err
	}
	facts := &CanonicalFinalFacts{tx: tx, issueID: id, actor: actor}
	if len(updates) == 0 {
		return facts, nil
	}
	oldValue, err := json.Marshal(before)
	if err != nil {
		return nil, err
	}
	newValue, err := json.Marshal(updates)
	if err != nil {
		return nil, err
	}
	// This context override is private to this composite and this real tx.
	// The transaction's journal activation remains intact for final emission.
	deferred := context.WithValue(ctx, journalContextKey{}, false)
	result, err := updateIssueInTx(deferred, tx, id, updates, actor, false, false)
	if err != nil {
		return nil, err
	}
	if !result.Changed || result.IsWisp {
		return nil, fmt.Errorf("%w: final native Issue edit changed its admitted plane or intent", storage.ErrValidation)
	}
	after, err := getIssueFromTableInTx(ctx, tx, "issues", "labels", id)
	if err != nil {
		return nil, err
	}
	if err := checkCanonicalStoredContent(after, updates); err != nil {
		return nil, err
	}
	facts.changed, facts.oldValue, facts.newValue = true, string(oldValue), string(newValue)
	return facts, nil
}

func checkCanonicalStoredContent(after *types.Issue, updates map[string]interface{}) error {
	for key, value := range updates {
		if key == "metadata" {
			wanted, err := storage.NormalizeMetadataValue(value)
			if err != nil {
				return err
			}
			stored, err := normalizedMetadata(after.Metadata)
			if err != nil {
				return err
			}
			if DependencyMetadataEqual(stored, wanted) {
				continue
			}
		} else if matches, err := issueFieldMatches(after, key, value); err != nil {
			return err
		} else if matches {
			continue
		}
		return fmt.Errorf("%w: native storage did not retain exact canonical %s intent", storage.ErrValidation, key)
	}
	return nil
}

// CanonicalDependencyMetadata identifies an existing durable local Issue pair.
// Type is a checked expected native classification, never a retype request.
type CanonicalDependencyMetadata struct {
	IssueID, DependsOnID string
	Type                 types.DependencyType
	Metadata             json.RawMessage
}

// ValidateCanonicalDependencyMetadata limits this first seam to positively
// recognized native relationships whose metadata cannot affect readiness.
func ValidateCanonicalDependencyMetadata(request CanonicalDependencyMetadata, actor string) error {
	if err := validateCanonicalIdentity(request.IssueID, actor); err != nil {
		return err
	}
	if !utf8.ValidString(request.DependsOnID) {
		return fmt.Errorf("%w: canonical Dependency target must be UTF-8", storage.ErrValidation)
	}
	if err := ValidateAddDependenciesRequest(publicops.AddDependenciesRequest{Actor: actor, Edges: []publicops.DependencyEdge{{IssueID: request.IssueID, DependsOnID: request.DependsOnID, Type: request.Type}}}); err != nil {
		return err
	}
	if !request.Type.IsWellKnown() || request.Type.AffectsReadyWork() {
		return fmt.Errorf("%w: canonical Dependency metadata requires a recognized non-readiness relationship", storage.ErrValidation)
	}
	return validateCanonicalMetadata(request.Metadata)
}

// ApplyCanonicalDependencyMetadataInTx reuses the existing same-type re-add
// writer with publication deferred. It checks that the local typed pair exists
// first, so the writer cannot take its create/retype/recovery paths. Native ID,
// endpoints, creation facts and thread are preserved. The owner is not versioned.
func ApplyCanonicalDependencyMetadataInTx(ctx context.Context, tx *sql.Tx, request CanonicalDependencyMetadata, actor string) (*CanonicalFinalFacts, error) {
	if tx == nil {
		return nil, fmt.Errorf("%w: canonical edit requires a transaction", storage.ErrValidation)
	}
	if err := ValidateCanonicalDependencyMetadata(request, actor); err != nil {
		return nil, err
	}
	for _, id := range []string{request.IssueID, request.DependsOnID} {
		issue, err := getIssueFromTableInTx(ctx, tx, "issues", "labels", id)
		if err != nil {
			return nil, err
		}
		if issue.Ephemeral || issue.NoHistory || IsActiveWispInTx(ctx, tx, id) {
			return nil, fmt.Errorf("%w: canonical Dependency requires durable local Issue endpoints", storage.ErrValidation)
		}
		if id == request.IssueID {
			if err := AuthorizeTemplateUpdate(issue, publicops.UpdateRequest{}); err != nil {
				return nil, err
			}
		}
	}
	const pairPredicate = `issue_id=? AND depends_on_issue_id=? AND depends_on_wisp_id IS NULL AND depends_on_external IS NULL`
	var actualType string
	var metadata sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT type, metadata FROM dependencies WHERE `+pairPredicate, request.IssueID, request.DependsOnID).Scan(&actualType, &metadata); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: canonical Dependency pair", storage.ErrNotFound)
		}
		return nil, err
	}
	if actualType != string(request.Type) {
		return nil, fmt.Errorf("%w: canonical Dependency classification changed", storage.ErrValidation)
	}
	facts := &CanonicalFinalFacts{tx: tx, issueID: request.IssueID, actor: actor}
	current := metadata.String
	if !metadata.Valid {
		current = "{}"
	}
	if DependencyMetadataEqual(current, string(request.Metadata)) {
		return facts, nil
	}
	dep := types.Dependency{IssueID: request.IssueID, DependsOnID: request.DependsOnID, Type: request.Type, Metadata: string(request.Metadata)}
	kind := DepTargetIssue
	deferred := context.WithValue(ctx, journalContextKey{}, false)
	if _, err := addDependencyInTx(deferred, tx, &dep, actor, AddDependencyOpts{SourceTable: "issues", TargetTable: "issues", WriteTable: "dependencies", TargetKind: &kind}, nil, false); err != nil {
		return nil, err
	}
	var stored string
	if err := tx.QueryRowContext(ctx, `SELECT metadata FROM dependencies WHERE `+pairPredicate, request.IssueID, request.DependsOnID).Scan(&stored); err != nil {
		return nil, err
	}
	if !DependencyMetadataEqual(stored, dep.Metadata) {
		return nil, fmt.Errorf("%w: native storage did not retain canonical Dependency metadata", storage.ErrValidation)
	}
	// Use the actual stored JSON spelling, matching ordinary journal behavior.
	dep.Metadata = stored
	facts.changed, facts.dependency = true, &dep
	return facts, nil
}

// EmitCanonicalFinalFactsInTx emits deferred facts after ALL final native rows,
// native versions and graph snapshots have been published in this same tx.
// The ordinary transaction-scoped journal setting is honored. Issue content
// emits its ordinary full audit and update journal; same-type Dependency
// metadata emits dep_add replacement journal only, as its native writer does.
// This function neither mints history nor supplies a new BDP event vocabulary.
func EmitCanonicalFinalFactsInTx(ctx context.Context, tx *sql.Tx, facts *CanonicalFinalFacts) error {
	if facts == nil || tx == nil || facts.tx != tx || facts.emitted {
		return fmt.Errorf("%w: invalid or already emitted canonical final facts", storage.ErrValidation)
	}
	if !facts.changed {
		facts.emitted = true
		return nil
	}
	if dep := facts.dependency; dep != nil {
		if err := RecordDepEventInTx(ctx, tx, EventDepAdd, dep.IssueID, string(dep.Type), dep.DependsOnID, dep.Metadata, facts.actor); err != nil {
			return err
		}
	} else {
		if err := RecordFullEventInTable(ctx, tx, "events", facts.issueID, types.EventUpdated, facts.actor, facts.oldValue, facts.newValue); err != nil {
			return err
		}
		if err := RecordEventInTx(ctx, tx, EventUpdate, facts.issueID, facts.actor); err != nil {
			return err
		}
	}
	facts.emitted = true
	return nil
}
