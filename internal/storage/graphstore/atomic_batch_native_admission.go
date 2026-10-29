package graphstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	publicops "github.com/steveyegge/beads/issueops"
)

// Native policy is checked per authored mutation, before final no-op collapse.
// Final diffs alone cannot authorize a canceled protected field or template edit.
func (run *atomicBatchRun) admitNativeMutation(_ int, member canonicalbatch.Member, before *canonicalbatchstate.Resource) error {
	if before == nil {
		return nil // Native creations are statically refused by the receiver.
	}
	refuse := func() error {
		return fmt.Errorf("%w: authored native mutation requires an unavailable final publisher", ErrCapabilityUnavailable)
	}
	switch original := run.initial[before.ID].(type) {
	case IssueRecord:
		// Ordinary native write admission requires an actor even when the
		// authored changes later cancel and publication has no row to write.
		if run.actor == "" {
			return fmt.Errorf("%w: Issue update requires a nonempty UTF-8 actor", storage.ErrValidation)
		}
		if member.Operation != "updateBead" || member.Properties != nil || member.Type.URI != "" && member.Type.URI != original.Type {
			return refuse()
		}
		if original.Properties == nil || original.Properties.Ephemeral || original.Properties.NoHistory {
			return refuse()
		}
		if err := issueops.AuthorizeTemplateUpdate(original.Properties, publicops.UpdateRequest{}); err != nil {
			return err
		}
		var operations []struct {
			Path  string
			Value json.RawMessage
		}
		if member.PropertiesChange != nil {
			if err := json.Unmarshal(member.PropertiesChange, &operations); err != nil {
				return err
			}
		}
		for _, operation := range operations {
			if !canonicalNativeContentPath(operation.Path) {
				return refuse()
			}
			if err := canonicalNativeAuthoredValue(operation.Path, operation.Value); err != nil {
				return err
			}
		}
	case LinkRecord:
		if run.backings[before.ID] != "dependency" {
			if member.Type.URI != "" && !IsInformationalTypeURL(run.store.ScopeURL(), member.Type.URI) {
				return refuse()
			}
			return nil
		}
		if run.actor == "" {
			return fmt.Errorf("%w: Dependency update requires a nonempty UTF-8 actor", storage.ErrValidation)
		}
		if member.Operation != "updateLink" || member.Properties != nil || member.Source.Present() || member.Target.Present() || member.Type.URI != "" && member.Type.URI != original.Type {
			return refuse()
		}
		var operations []json.RawMessage
		if member.PropertiesChange != nil && json.Unmarshal(member.PropertiesChange, &operations) != nil || len(operations) != 0 {
			return refuse()
		}
		kind, err := nativeDependencyProperties(before.Properties)
		if err != nil {
			return err
		}
		if !kind.IsWellKnown() || kind.AffectsReadyWork() {
			return refuse()
		}
		for _, id := range []string{original.Source, original.Target} {
			endpoint, ok := run.initial[id].(IssueRecord)
			if !ok || endpoint.Properties == nil || endpoint.Properties.Ephemeral || endpoint.Properties.NoHistory {
				return refuse()
			}
			if id == original.Source {
				if err := issueops.AuthorizeTemplateUpdate(endpoint.Properties, publicops.UpdateRequest{}); err != nil {
					return err
				}
			}
		}
	case Record:
		if member.Type.URI != "" && !run.store.isMemoryType(member.Type.URI) {
			return refuse()
		}
	case nil:
		// Earlier creations have no committed baseline. They still cannot
		// temporarily acquire an unsupported native backing and cancel it.
		if member.Type.URI != "" && (before.Kind == graph.KindBead && !run.store.isMemoryType(member.Type.URI) || before.Kind == graph.KindLink && !IsInformationalTypeURL(run.store.ScopeURL(), member.Type.URI)) {
			return refuse()
		}
	}
	return nil
}

func canonicalNativeAuthoredValue(path string, value json.RawMessage) error {
	if value == nil {
		return nil // Removal is the explicit spelling for an optional clear.
	}
	refuse := func() error {
		return fmt.Errorf("%w: native property presence/value cannot be retained; use removal for an optional clear", ErrCapabilityUnavailable)
	}
	switch path {
	case "/estimated_minutes", "/external_ref", "/due_at":
		if bytes.Equal(value, []byte("null")) {
			return refuse()
		}
	}
	switch path {
	case "/description", "/design", "/acceptance_criteria", "/spec_id", "/external_ref":
		if bytes.Equal(value, []byte(`""`)) {
			return refuse()
		}
	case "/due_at":
		var date time.Time
		if err := json.Unmarshal(value, &date); err != nil {
			return refuse()
		}
		canonical, err := json.Marshal(date.UTC())
		if err != nil || date.Nanosecond() != 0 || !bytes.Equal(canonical, value) {
			return refuse()
		}
	}
	return nil
}

func canonicalNativeContentPath(path string) bool {
	switch path {
	case "/title", "/description", "/design", "/acceptance_criteria", "/spec_id", "/priority", "/estimated_minutes", "/external_ref", "/due_at":
		return true
	default:
		return false
	}
}

// Translate the admitted final content into the existing finite native setters.
// This is not a second patch evaluator or a full-DTO replacement contract.
func canonicalNativeIssueContent(before IssueRecord, after canonicalbatchstate.Resource) (issueops.CanonicalIssueContent, error) {
	var result issueops.CanonicalIssueContent
	oldRaw, err := canonicalJSON(before.Properties)
	if err != nil {
		return result, err
	}
	var oldFields, fields map[string]json.RawMessage
	if err := json.Unmarshal(oldRaw, &oldFields); err != nil {
		return result, err
	}
	if err := json.Unmarshal(after.Properties, &fields); err != nil {
		return result, err
	}
	var projected publicops.Issue
	if err := json.Unmarshal(after.Properties, &projected); err != nil {
		return result, fmt.Errorf("%w: native property type: %v", graph.ErrValidation, err)
	}
	if projected.DueAt != nil {
		utc := projected.DueAt.UTC()
		projected.DueAt = &utc
	}
	representable, err := canonicalJSON(&projected)
	if err != nil {
		return result, err
	}
	if !bytes.Equal(representable, after.Properties) {
		return result, fmt.Errorf("%w: staged native properties do not have an exact native projection", ErrCapabilityUnavailable)
	}
	changed := func(key string) bool { return !bytes.Equal(oldFields[key], fields[key]) }
	for key := range oldFields {
		if changed(key) && !canonicalNativeContentPath("/"+key) {
			return result, fmt.Errorf("%w: native protected property changed", ErrCapabilityUnavailable)
		}
	}
	for key := range fields {
		if changed(key) && !canonicalNativeContentPath("/"+key) {
			return result, fmt.Errorf("%w: native unmapped property changed", ErrCapabilityUnavailable)
		}
	}
	for _, setter := range []struct {
		key   string
		field *publicops.Field[string]
	}{{"title", &result.Title}, {"description", &result.Description}, {"design", &result.Design}, {"acceptance_criteria", &result.AcceptanceCriteria}, {"spec_id", &result.SpecID}} {
		if !changed(setter.key) {
			continue
		}
		value := fields[setter.key]
		if value == nil && setter.key == "title" || value != nil && (len(value) == 0 || value[0] != '"') {
			return result, fmt.Errorf("%w: native %s requires a string", graph.ErrValidation, setter.key)
		}
		setter.field.Set = true
		if value != nil {
			if err := json.Unmarshal(value, &setter.field.Value); err != nil {
				return result, err
			}
		}
	}
	if changed("priority") {
		result.Priority.Set = true
		if value := fields["priority"]; value == nil || bytes.Equal(value, []byte("null")) {
			return result, fmt.Errorf("%w: native priority cannot be removed", graph.ErrValidation)
		} else if err := json.Unmarshal(value, &result.Priority.Value); err != nil {
			return result, err
		}
	}
	if changed("estimated_minutes") {
		result.EstimatedMinutes.Set = true
		if value := fields["estimated_minutes"]; value != nil {
			if err := json.Unmarshal(value, &result.EstimatedMinutes.Value); err != nil {
				return result, err
			}
		}
	}
	if changed("external_ref") {
		result.ExternalRef.Set = true
		if value := fields["external_ref"]; value != nil {
			if err := json.Unmarshal(value, &result.ExternalRef.Value); err != nil {
				return result, err
			}
		}
	}
	if changed("due_at") {
		result.DueAt.Set = true
		if value := fields["due_at"]; value != nil && !bytes.Equal(value, []byte("null")) {
			var date time.Time
			if err := json.Unmarshal(value, &date); err != nil {
				return result, err
			}
			result.DueAt.Value = &date
		}
	}
	result.Metadata = publicops.Field[json.RawMessage]{Set: !bytes.Equal(before.Metadata, after.Metadata), Value: bytes.Clone(after.Metadata)}
	return result, issueops.ValidateCanonicalIssueContent(result)
}
