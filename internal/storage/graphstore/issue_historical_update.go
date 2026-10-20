package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/steveyegge/beads/internal/storage/issueops"
	publicops "github.com/steveyegge/beads/issueops"
)

// These are the existing graph UpdateIssue setters, not the complete native
// Issue JSON shape. Unmapped editable members must agree with current state;
// protected initialization and service/lifecycle facts always remain current.
var historicalIssueEditableMembers = map[string]bool{
	"title": true, "description": true, "design": true, "acceptance_criteria": true,
	"priority": true, "estimated_minutes": true, "assignee": true, "notes": true,
	"external_ref": true, "spec_id": true, "due_at": true,
}
var historicalIssueCurrentMembers = map[string]bool{
	"id": true, "created_at": true, "created_by": true, "updated_at": true,
	"started_at": true, "closed_at": true, "close_reason": true, "closed_by_session": true,
	"lease_expires_at": true, "heartbeat_at": true, "lease_granted_node": true,
	"is_blocked": true, "source_system": true, "storage_class": true,
	"compaction_level": true, "compacted_at": true, "compacted_at_commit": true, "original_size": true,
	"source_formula": true, "source_location": true,
}

func historicalIssueField[T any](source T, edit publicops.Field[T]) publicops.Field[T] {
	if edit.Set {
		return edit
	}
	return publicops.Field[T]{Set: true, Value: source}
}

func composeHistoricalIssuePatch(current, source IssueRecord, edit publicops.IssuePatch) (publicops.IssuePatch, error) {
	if current.Properties == nil || source.Properties == nil || current.ID != source.ID || current.Type != source.Type {
		return publicops.IssuePatch{}, fmt.Errorf("%w: historical Issue source requires the same native subject", ErrCapabilityUnavailable)
	}
	objects := make([]map[string]json.RawMessage, 2)
	for i, issue := range []*publicops.Issue{current.Properties, source.Properties} {
		raw, err := canonicalJSON(issue)
		if err != nil {
			return publicops.IssuePatch{}, err
		}
		if err := json.Unmarshal(raw, &objects[i]); err != nil {
			return publicops.IssuePatch{}, err
		}
	}
	for _, name := range comparisonKeys(objects[0], objects[1]) {
		if historicalIssueEditableMembers[name] || historicalIssueCurrentMembers[name] {
			continue
		}
		a, haveA := objects[0][name]
		b, haveB := objects[1][name]
		equal, err := comparisonJSONEqual(a, b)
		if haveA != haveB || err != nil || !equal {
			return publicops.IssuePatch{}, fmt.Errorf("%w: historical Issue member %s requires its native domain mapping", ErrCapabilityUnavailable, name)
		}
	}
	s := source.Properties
	patch := publicops.IssuePatch{
		Title: historicalIssueField(s.Title, edit.Title), Description: historicalIssueField(s.Description, edit.Description),
		Design: historicalIssueField(s.Design, edit.Design), AcceptanceCriteria: historicalIssueField(s.AcceptanceCriteria, edit.AcceptanceCriteria),
		Priority: historicalIssueField(s.Priority, edit.Priority), EstimatedMinutes: historicalIssueField(s.EstimatedMinutes, edit.EstimatedMinutes),
		Assignee: historicalIssueField(s.Assignee, edit.Assignee), Notes: historicalIssueField(s.Notes, edit.Notes),
		ExternalRef: historicalIssueField(s.ExternalRef, edit.ExternalRef), SpecID: historicalIssueField(s.SpecID, edit.SpecID), DueAt: historicalIssueField(s.DueAt, edit.DueAt),
	}
	if edit.AppendNotes.Set {
		if s.Notes == current.Properties.Notes {
			// Preserve ordinary native append intent and its overwrite fence.
			patch.Notes = publicops.Field[string]{}
			patch.AppendNotes = edit.AppendNotes
		} else {
			resolved, err := issueops.ResolveMergeOps(s, issueops.UpdateFields(publicops.IssuePatch{AppendNotes: edit.AppendNotes}))
			if err != nil {
				return publicops.IssuePatch{}, err
			}
			notes, ok := resolved["notes"].(string)
			if !ok {
				return publicops.IssuePatch{}, fmt.Errorf("%w: native notes composition lost its value", ErrInvalidStore)
			}
			patch.Notes = publicops.Field[string]{Set: true, Value: notes}
		}
	}
	metadata, _, err := issueops.ApplyMetadataPatch(source.Metadata, edit.Metadata)
	if err != nil {
		return publicops.IssuePatch{}, err
	}
	metadata, err = commonMetadata(metadata)
	if err != nil {
		return publicops.IssuePatch{}, err
	}
	patch.Metadata = publicops.MetadataPatch{Replace: publicops.Field[json.RawMessage]{Set: true, Value: metadata}}
	return patch, nil
}

func (s *Store) historicalIssuePatchInTx(ctx context.Context, tx *sql.Tx, path, sourceRevision string, current IssueRecord, edit publicops.IssuePatch) (publicops.IssuePatch, error) {
	value, err := s.readVersionInTx(ctx, tx, path, sourceRevision)
	if err != nil {
		return publicops.IssuePatch{}, err
	}
	source, ok := value.(IssueRecord)
	if !ok {
		return publicops.IssuePatch{}, fmt.Errorf("%w: historical Issue source requires a live native Issue snapshot", ErrCapabilityUnavailable)
	}
	return composeHistoricalIssuePatch(current, source, edit)
}
