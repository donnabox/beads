package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// Human projection never changes public identity. In particular, literal
// beads/, links/ and types/ segments survive Scope-relative display.
func graphHumanID(scope, id string) string {
	if local, ok := graphPreviewPublicResourceID(scope, id); ok {
		return graphMemoryDisplayText(local)
	}
	return graphMemoryDisplayText(id)
}

func graphHumanType(scope, typ string) string {
	if _, name, ok := graphstore.NativeTypeClassification(scope, typ); ok {
		return graphMemoryDisplayText(name)
	}
	return graphHumanID(scope, typ)
}

func graphHumanEndpoint(scope, id, token string) string {
	out := graphHumanID(scope, id)
	if token != "" {
		out += "@=" + graphMemoryDisplayText(token)
	}
	return out
}

// One row contract serves both successful creation and collection output.
// Endpoint tokens are exact stored pins; absence denotes a floating reference.
func graphHumanRow(scope string, value any) string {
	var id, typ, title, extra string
	switch r := value.(type) {
	case graphstore.Record:
		id, typ, title = r.ID, r.Type, r.Properties.Title
	case graphstore.IssueRecord:
		id, typ = r.ID, r.Type
		if r.Properties != nil {
			title = r.Properties.Title
			extra = fmt.Sprintf(" %s P%d", graphMemoryDisplayText(string(r.Properties.Status)), r.Properties.Priority)
		}
	case graphstore.LinkRecord:
		id, typ = r.ID, r.Type
		extra = " " + graphHumanEndpoint(scope, r.Source, r.SourceRevision) + " → " + graphHumanEndpoint(scope, r.Target, r.TargetRevision)
		title, _ = r.Properties["title"].(string)
	}
	row := fmt.Sprintf("%s [%s]%s", graphHumanID(scope, id), graphHumanType(scope, typ), extra)
	if title != "" {
		row += fmt.Sprintf(" %q", title)
	}
	return row
}

// Ordinals are optional presentation aliases, never a prerequisite for a
// successful read or applied write. Give their bounded lookup its own deadline
// and fall back to the exact token for missing/over-budget History.
func graphHumanRevision(ctx context.Context, store *graphstore.Store, id, token string) string {
	if ctx.Err() != nil || store == nil {
		return "@=" + graphMemoryDisplayText(token)
	}
	aliasCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	identity, err := store.ResolveCanonicalPublicResourceIdentity(aliasCtx, id)
	if err == nil {
		row, lookupErr := store.LocalRevisionForToken(aliasCtx, identity.PrivatePath, token)
		if lookupErr == nil && row.Ordinal > 0 {
			return fmt.Sprintf("@%d", row.Ordinal)
		}
	}
	return "@=" + graphMemoryDisplayText(token)
}

// The selected record is the sole content source for both current and retained
// documents. Current relationships/comments are not hydrated into a past view.
func graphHumanDocument(scope string, value any, revision string, retained, long bool) string {
	var token, body string
	var properties any
	var metadata json.RawMessage
	var attribution graphstore.Attribution
	var owned []json.RawMessage
	var endpoints *graphstore.LinkRecord
	switch r := value.(type) {
	case graphstore.Record:
		token, body, properties, metadata, attribution, owned = r.Revision, r.Properties.Body, r.Properties, r.Metadata, r.Attribution, r.Owned
	case graphstore.IssueRecord:
		token, properties, metadata, attribution, owned = r.Revision, r.Properties, r.Metadata, r.Attribution, r.Owned
		if r.Properties != nil {
			body = r.Properties.Description
		}
	case graphstore.LinkRecord:
		token, properties, metadata, attribution, endpoints = r.Revision, r.Properties, r.Metadata, r.Attribution, &r
		body, _ = r.Properties["body"].(string)
	}
	var out strings.Builder
	out.WriteString(graphHumanRow(scope, value))
	state := "current"
	if retained {
		state = "retained snapshot"
	}
	if revision == "" {
		revision = "@=" + graphMemoryDisplayText(token)
	}
	fmt.Fprintf(&out, "\nRevision: %s (%s)\n", revision, state)
	if endpoints != nil {
		fmt.Fprintf(&out, "From: %s\nTo: %s\n", graphHumanEndpoint(scope, endpoints.Source, endpoints.SourceRevision), graphHumanEndpoint(scope, endpoints.Target, endpoints.TargetRevision))
	}
	section := func(name, text string) {
		if text != "" {
			fmt.Fprintf(&out, "\n%s\n%s\n", name, text)
		}
	}
	section("Body", body)
	// Rendering a checked in-memory record cannot introduce a storage dependency.
	raw, _ := json.Marshal(properties)
	var props map[string]json.RawMessage
	_ = json.Unmarshal(raw, &props)
	switch r := value.(type) {
	case graphstore.Record:
		// Only these two fields are protected Memory strings. Domain extras
		// named description, metadata, revision, etc. are still user content.
		delete(props, "title")
		delete(props, "body")
	case graphstore.LinkRecord:
		for _, name := range []string{"title", "body"} {
			if text, ok := r.Properties[name].(string); ok && text != "" {
				delete(props, name)
			}
		}
	case graphstore.IssueRecord:
		for _, name := range []string{"title", "description", "metadata"} {
			delete(props, name)
		}
		if !long {
			for _, name := range []string{"created_at", "created_by", "updated_at", "closed_at", "revision", "id"} {
				delete(props, name)
			}
		}
	}
	if len(props) != 0 {
		data, _ := json.MarshalIndent(props, "", "  ")
		section("Properties", string(data))
	}
	if len(metadata) != 0 && string(metadata) != "null" && string(metadata) != "{}" {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, metadata, "", "  "); err == nil {
			section("Metadata", pretty.String())
		} else {
			section("Metadata", string(metadata))
		}
	}
	if len(owned) != 0 {
		data, _ := json.MarshalIndent(owned, "", "  ")
		section("Owned Links (selected revision)", string(data))
	}
	if long {
		typ := ""
		switch r := value.(type) {
		case graphstore.Record:
			typ = r.Type
		case graphstore.IssueRecord:
			typ = r.Type
		case graphstore.LinkRecord:
			typ = r.Type
		}
		section("Type identity", graphMemoryDisplayText(typ))
		section("Exact revision token", graphMemoryDisplayText(token))
		if attribution.Actor != "" || attribution.RecordedAt != "" {
			data, _ := json.MarshalIndent(attribution, "", "  ")
			section("Attribution", string(data))
		}
	}
	return strings.TrimSuffix(out.String(), "\n")
}

func graphHumanRecordIdentity(value any) (string, string) {
	switch r := value.(type) {
	case graphstore.Record:
		return r.ID, r.Revision
	case graphstore.IssueRecord:
		return r.ID, r.Revision
	case graphstore.LinkRecord:
		return r.ID, r.Revision
	}
	return "", ""
}

// These adjuncts are explicitly current, and only a current-state reader may
// supply them. Retained documents never call this function.
func graphHumanCurrentIssueDetails(scope string, details *types.IssueDetails, opts graphIssueDetailOptions) string {
	if details == nil {
		return ""
	}
	var out strings.Builder
	for _, group := range []struct {
		name string
		rows []*types.IssueWithDependencyMetadata
	}{{"Current dependencies", details.Dependencies}, {"Current dependents", details.Dependents}} {
		if len(group.rows) == 0 {
			continue
		}
		fmt.Fprintf(&out, "\n\n%s\n", group.name)
		for _, issue := range group.rows {
			fmt.Fprintf(&out, "%s [%s] %s P%d %q\n", graphHumanID(scope, issue.ID), graphMemoryDisplayText(string(issue.IssueType)), graphMemoryDisplayText(string(issue.Status)), issue.Priority, issue.Title)
		}
	}
	comments := details.Comments
	if opts.CommentsTail > 0 && len(comments) > opts.CommentsTail {
		comments = comments[len(comments)-opts.CommentsTail:]
	}
	if len(comments) != 0 {
		out.WriteString("\n\nCurrent comments\n")
		for _, comment := range comments {
			fmt.Fprintf(&out, "%s: %s\n", graphMemoryDisplayText(comment.Author), graphMemoryDisplayText(comment.Text))
		}
	}
	return strings.TrimSuffix(out.String(), "\n")
}
