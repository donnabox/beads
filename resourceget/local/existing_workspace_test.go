//go:build cgo

package local_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	mysql "github.com/go-sql-driver/mysql"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/testutil/bazeltest"
	publicops "github.com/steveyegge/beads/issueops"
	"github.com/steveyegge/beads/resourceget"
	"github.com/steveyegge/beads/resourceget/local"
)

type existingFixture struct {
	GoneURI                                                      string `json:"-"`
	Options                                                      local.Options
	BeadID, LinkID, BeadURI, LinkURI, BeadRevision, LinkRevision string
	LinkType, TargetURI, TargetRevision                          string
}

func prepareExistingFixture(t *testing.T, backend local.Backend) existingFixture {
	t.Helper()
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const scope = "https://example.test/current-get/"
	o := graphstore.Options{Backend: "embedded", Database: fmt.Sprintf("current_get_%d", time.Now().UnixNano()), DataDir: filepath.Join(workspace, "dolt"), Branch: "main", Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: scope, AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}, TypeGeneration: graphstore.TypeGeneration}
	public := local.Options{Backend: backend, WorkspaceID: workspace, ScopeURL: scope, AuthorityID: o.Binding.AuthorityID, Database: o.Database, DataDir: o.DataDir}
	if backend == local.Server {
		port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
		if port == "" {
			t.Skip("checked local shared-server backend not configured")
		}
		o.ServerPort, err = strconv.Atoi(port)
		if err != nil || o.ServerPort < 1 || o.ServerPort > 65535 {
			t.Fatalf("invalid server port %q", port)
		}
		o.Backend, o.ServerHost, o.ServerUser = "server", "127.0.0.1", "root"
		public.Server = local.ServerOptions{Host: o.ServerHost, Port: o.ServerPort, User: o.ServerUser}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := graphstore.Init(ctx, o); err != nil {
		t.Fatal(err)
	}
	if backend == local.Server {
		// Init uses CREATE DATABASE without IF NOT EXISTS: success establishes that
		// this fixture allocated this exact generated name, never an ambient DB.
		registerCurrentGetServerCleanup(t, o.Database, o.ServerPort)
	}
	store, err := graphstore.OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	bead, err := store.Create(ctx, graphstore.CreateRequest{Path: "beads/source", PublicID: "team/task-1", Title: "Current — 記憶", Body: "", Metadata: json.RawMessage(`{"reviewed":true}`)})
	if err != nil {
		t.Fatal(err)
	}
	target, err := store.Create(ctx, graphstore.CreateRequest{Path: "beads/target", PublicID: "task-2", TypeURL: scope + "constraint", Title: "Target"})
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/edge", PublicID: "team/link-1", TypeURL: scope + "subject-to", SourcePath: "beads/source", TargetPath: "beads/target", TargetPin: graphstore.EndpointPin{Revision: target.Revision}, ExpectedSourceRevision: bead.Revision, Metadata: json.RawMessage(`{"reviewed":true}`), Properties: map[string]any{"explanation": "kept"}})
	if err != nil {
		t.Fatal(err)
	}
	source, ok := created.Source.(graphstore.Record)
	if !ok {
		t.Fatalf("fixture source %T", created.Source)
	}
	gone, err := store.Create(ctx, graphstore.CreateRequest{Path: "beads/gone", PublicID: "gone", Title: "Deleted fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.DeleteMemory(ctx, graphstore.MemoryDeleteRequest{Path: "beads/gone", ExpectedRevision: gone.Revision}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	return existingFixture{GoneURI: gone.ID, Options: public, BeadID: "team/task-1", LinkID: created.Link.ID, BeadURI: bead.ID, LinkURI: created.Link.ID, BeadRevision: source.Revision, LinkRevision: created.Link.Revision, LinkType: created.Link.Type, TargetURI: target.ID, TargetRevision: target.Revision}
}

func TestExistingWorkspaceCurrentGet(t *testing.T) {
	for _, backend := range []local.Backend{local.Embedded, local.Server} {
		t.Run(fmt.Sprint(backend), func(t *testing.T) {
			fixture := prepareExistingFixture(t, backend)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			h, err := local.OpenExisting(ctx, fixture.Options)
			if err != nil {
				t.Fatal(err)
			}
			defer h.Close()
			for _, request := range []resourceget.Request{{ID: fixture.BeadID, AssertKind: resourceget.Bead}, {ID: fixture.BeadURI}, {ID: fixture.LinkID, AssertKind: resourceget.Link}, {ID: "team/link-1", AssertKind: resourceget.Link}} {
				doc, err := h.Get(ctx, request)
				if err != nil {
					t.Fatalf("Get %+v: %v", request, err)
				}
				if doc.Kind() == resourceget.Bead {
					if doc.ID() != fixture.BeadURI || doc.Revision() != fixture.BeadRevision {
						t.Fatal("wrong current Bead")
					}
					var body struct {
						OwnedLinks map[string][]struct {
							ID, Revision string
							Metadata     map[string]json.RawMessage
						}
					}
					if err := json.Unmarshal(doc.NativeJSON(), &body); err != nil {
						t.Fatal(err)
					}
					group := body.OwnedLinks[fixture.LinkType]
					if len(group) != 1 || group[0].ID != fixture.LinkURI || group[0].Revision != fixture.LinkRevision || group[0].Metadata == nil {
						t.Fatal("incomplete owned current state")
					}
				} else if doc.ID() != fixture.LinkURI || doc.Revision() != fixture.LinkRevision {
					t.Fatal("wrong current Link")
				}
			}
			for _, request := range []resourceget.Request{{ID: "https://foreign.test/task-1"}, {ID: fixture.BeadURI + "?view=properties"}, {ID: fixture.LinkURI + "#part"}, {ID: fixture.BeadID, AssertKind: resourceget.Kind(255)}} {
				if _, err := h.Get(ctx, request); !errors.Is(err, resourceget.ErrInvalidRequest) {
					t.Fatalf("admitted %+v: %v", request, err)
				}
			}
			if _, err := h.Get(ctx, resourceget.Request{ID: fixture.BeadID, AssertKind: resourceget.Link}); !errors.Is(err, resourceget.ErrKindMismatch) || !errors.Is(err, graphstore.ErrCapabilityUnavailable) {
				t.Fatalf("kind assertion: %v", err)
			}
			if _, err := h.Get(ctx, resourceget.Request{ID: fixture.Options.ScopeURL + "never-allocated"}); !errors.Is(err, resourceget.ErrNotFound) || !errors.Is(err, graphstore.ErrNotFound) {
				t.Fatalf("missing source: %v", err)
			}
			if _, err := h.Get(ctx, resourceget.Request{ID: fixture.GoneURI}); !errors.Is(err, resourceget.ErrGone) || !errors.Is(err, graphstore.ErrGone) {
				t.Fatalf("gone source: %v", err)
			}
			canceled, stop := context.WithCancel(ctx)
			stop()
			if _, err := h.Get(canceled, resourceget.Request{ID: fixture.BeadID}); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation: %v", err)
			}
			if err := h.Close(); err != nil {
				t.Fatal(err)
			}
			if _, err := h.Get(ctx, resourceget.Request{ID: fixture.BeadID}); !errors.Is(err, resourceget.ErrClosed) {
				t.Fatalf("closed handle: %v", err)
			}
			wrong := fixture.Options
			wrong.AuthorityID = "ffffffffffffffffffffffffffffffff"
			if bad, err := local.OpenExisting(ctx, wrong); !errors.Is(err, resourceget.ErrInvalidSource) {
				if bad != nil {
					bad.Close()
				}
				t.Fatalf("wrong authority admitted: %v", err)
			}
		})
	}
}

// Build testdata/consumer as its own module first, then supply the executable.
// This test never builds a hidden fallback binary and never reads user state.
func TestExternalConsumerExistingWorkspace(t *testing.T) {
	if os.Getenv("BEADS_RESOURCEGET_CONSUMER") == "" {
		t.Skip("standalone public consumer executable not supplied")
	}
	binary, err := bazeltest.RunfileEnv("BEADS_RESOURCEGET_CONSUMER")
	if err != nil {
		t.Fatal(err)
	}
	for _, backend := range []local.Backend{local.Embedded, local.Server} {
		t.Run(fmt.Sprint(backend), func(t *testing.T) {
			fixture := prepareExistingFixture(t, backend)
			input, err := json.Marshal(fixture)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary)
			cmd.Stdin = bytes.NewReader(input)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("external public consumer: %v: %s", err, &stderr)
			}
			var result struct{ Bead, Link json.RawMessage }
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("consumer result: %v: %s", err, &stdout)
			}
			for kind, raw := range map[resourceget.Kind]json.RawMessage{resourceget.Bead: result.Bead, resourceget.Link: result.Link} {
				doc, err := resourceget.ParseLegacyCurrent(fixture.Options.ScopeURL, kind, raw)
				if err != nil {
					t.Fatal(err)
				}
				if kind == resourceget.Bead && doc.ID() != fixture.BeadURI || kind == resourceget.Link && doc.ID() != fixture.LinkURI {
					t.Fatal("consumer returned unrelated Resource")
				}
			}
		})
	}
}

// A Get edition covers ordinary authored state and complete owned state. This
// regression deliberately alternates checked writer and public reader handles:
// it does not require two embedded connectors to hold the same directory open.
// Every accepted change must appear with the writer's exact new revision.
func TestExistingWorkspaceCurrentGetCompleteEditions(t *testing.T) {
	fixture := prepareExistingFixture(t, local.Embedded)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	o := fixture.Options
	options := graphstore.Options{Backend: "embedded", Database: o.Database, DataDir: o.DataDir, Branch: "main", Binding: graphstore.Binding{WorkspaceID: o.WorkspaceID, ScopeURL: o.ScopeURL, AuthorityID: o.AuthorityID, SchemaVersion: graphstore.SchemaVersion}, TypeGeneration: graphstore.TypeGeneration}
	write := func(apply func(*graphstore.Store) error) {
		t.Helper()
		s, err := graphstore.OpenExisting(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		if err := apply(s); err != nil {
			t.Fatal(err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
	get := func(id string) resourceget.NativeDocument {
		t.Helper()
		h, err := local.OpenExisting(ctx, o)
		if err != nil {
			t.Fatal(err)
		}
		defer h.Close()
		doc, err := h.Get(ctx, resourceget.Request{ID: id})
		if err != nil {
			t.Fatal(err)
		}
		if err := h.Close(); err != nil {
			t.Fatal(err)
		}
		return doc
	}
	edition := func(label string, before resourceget.NativeDocument, expected string) resourceget.NativeDocument {
		t.Helper()
		after := get(before.ID())
		if after.ID() != before.ID() || after.Kind() != before.Kind() || after.Revision() != expected || after.Revision() == before.Revision() {
			t.Fatalf("%s: Get edition %s -> %s, accepted %s", label, before.Revision(), after.Revision(), expected)
		}
		return after
	}
	decode := func(doc resourceget.NativeDocument) currentEditionBody {
		t.Helper()
		var body currentEditionBody
		if err := json.Unmarshal(doc.NativeJSON(), &body); err != nil {
			t.Fatal(err)
		}
		return body
	}
	field := func(label string, fields map[string]json.RawMessage, key, expected string) {
		t.Helper()
		if string(fields[key]) != expected {
			t.Fatalf("%s: %s = %s, want %s", label, key, fields[key], expected)
		}
	}
	owned := func(owner, link resourceget.NativeDocument, count int) {
		t.Helper()
		items := decode(owner).OwnedLinks[fixture.LinkType]
		if len(items) != count {
			t.Fatalf("owned state: got %d Links, want %d", len(items), count)
		}
		if count == 1 && !bytes.Equal(items[0], link.NativeJSON()) {
			t.Fatalf("owner edition and direct Link differ:\n%s\n%s", items[0], link.NativeJSON())
		}
	}
	owner, link := get(fixture.BeadURI), get(fixture.LinkURI)
	owned(owner, link, 1)
	patchMemory := func(label string, request graphstore.MemoryPatchRequest, verify func(currentEditionBody)) {
		t.Helper()
		request.Path, request.ExpectedRevision = "beads/source", owner.Revision()
		var result graphstore.MemoryMutationResult
		write(func(s *graphstore.Store) (err error) { result, err = s.PatchMemory(ctx, request); return err })
		if !result.Changed {
			t.Fatalf("%s: writer reported no change", label)
		}
		owner = edition(label, owner, result.Memory.Revision)
		assertCurrentEditionPayload(t, owner, result.Memory.Properties, result.Memory.Metadata)
		verify(decode(owner))
		owned(owner, link, 1)
	}
	title, body, status := "Edition title", "Edition body — 雪\n", "draft"
	patchMemory("Memory title", graphstore.MemoryPatchRequest{Title: &title}, func(v currentEditionBody) { field("title", v.Properties, "title", `"Edition title"`) })
	patchMemory("Memory body", graphstore.MemoryPatchRequest{Body: &body}, func(v currentEditionBody) { field("body", v.Properties, "body", `"Edition body — 雪\n"`) })
	patchMemory("Memory status", graphstore.MemoryPatchRequest{Status: &status}, func(v currentEditionBody) { field("status", v.Properties, "status", `"draft"`) })
	patchMemory("Memory open property", graphstore.MemoryPatchRequest{PropertiesMerge: json.RawMessage(`{"extension":{"kept":true}}`)}, func(v currentEditionBody) { field("extension", v.Properties, "extension", `{"kept":true}`) })
	patchMemory("Memory metadata", graphstore.MemoryPatchRequest{Metadata: publicops.MetadataPatch{Set: map[string]json.RawMessage{"edition": json.RawMessage(`"second"`)}}}, func(v currentEditionBody) { field("metadata", v.Metadata, "edition", `"second"`) })

	// Owned deletion revises the source, removes the nested record and makes
	// current direct Get report Gone. Adding a fresh identity revises it again.
	deleteOwned := func(label, path string) {
		t.Helper()
		var result graphstore.LinkDeleteResult
		write(func(s *graphstore.Store) (err error) {
			result, err = s.Unlink(ctx, graphstore.LinkDeleteRequest{Path: path, ExpectedRevision: link.Revision(), ExpectedSourceRevision: owner.Revision()})
			return err
		})
		if !result.Changed {
			t.Fatalf("%s: writer reported no change", label)
		}
		source, ok := result.Source.(graphstore.Record)
		if !ok {
			t.Fatalf("source %T", result.Source)
		}
		owner = edition(label, owner, source.Revision)
		owned(owner, link, 0)
		h, err := local.OpenExisting(ctx, o)
		if err != nil {
			t.Fatal(err)
		}
		_, readErr := h.Get(ctx, resourceget.Request{ID: link.ID()})
		if err := h.Close(); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(readErr, resourceget.ErrGone) {
			t.Fatalf("%s: deleted Link Get: %v", label, readErr)
		}
	}
	deleteOwned("owned Link delete", "links/edge")
	var added graphstore.LinkMutationResult
	write(func(s *graphstore.Store) (err error) {
		added, err = s.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/edition", PublicID: "edition-link", TypeURL: fixture.LinkType, SourcePath: "beads/source", TargetPath: "beads/target", ExpectedSourceRevision: owner.Revision(), Properties: map[string]any{"note": "first"}})
		return err
	})
	source, ok := added.Source.(graphstore.Record)
	if !ok {
		t.Fatalf("source %T", added.Source)
	}
	owner = edition("owned Link add", owner, source.Revision)
	link = get(added.Link.ID)
	if link.Revision() != added.Link.Revision {
		t.Fatal("added Link edition differs from accepted write")
	}
	owned(owner, link, 1)
	assertCurrentEditionPayload(t, link, added.Link.Properties, added.Link.Metadata)
	updateOwned := func(label string, request graphstore.LinkUpdateRequest, verify func(currentEditionBody)) {
		t.Helper()
		request.Path, request.ExpectedRevision, request.ExpectedSourceRevision = "links/edition", link.Revision(), owner.Revision()
		var result graphstore.LinkMutationResult
		write(func(s *graphstore.Store) (err error) { result, err = s.UpdateLink(ctx, request); return err })
		if !result.Changed {
			t.Fatalf("%s: writer reported no change", label)
		}
		source, ok := result.Source.(graphstore.Record)
		if !ok {
			t.Fatalf("source %T", result.Source)
		}
		owner = edition(label+" owner", owner, source.Revision)
		link = edition(label+" Link", link, result.Link.Revision)
		assertCurrentEditionPayload(t, link, result.Link.Properties, result.Link.Metadata)
		verify(decode(link))
		owned(owner, link, 1)
	}
	updateOwned("Link properties", graphstore.LinkUpdateRequest{Properties: map[string]any{"note": "second"}}, func(v currentEditionBody) { field("Link properties", v.Properties, "note", `"second"`) })
	updateOwned("Link metadata", graphstore.LinkUpdateRequest{MetadataOnly: true, Metadata: publicops.MetadataPatch{Set: map[string]json.RawMessage{"reviewed": json.RawMessage(`true`)}}}, func(v currentEditionBody) { field("Link metadata", v.Metadata, "reviewed", `true`) })
	updateOwned("Link target pin set", graphstore.LinkUpdateRequest{TargetPin: &graphstore.EndpointPin{Revision: fixture.TargetRevision}}, func(v currentEditionBody) {
		var ref struct{ URI, Revision string }
		if err := json.Unmarshal(v.Target, &ref); err != nil {
			t.Fatal(err)
		}
		if ref.URI != fixture.TargetURI || ref.Revision != fixture.TargetRevision {
			t.Fatal("target pin changed or disappeared")
		}
	})
	updateOwned("Link target pin clear", graphstore.LinkUpdateRequest{TargetPin: &graphstore.EndpointPin{}}, func(v currentEditionBody) {
		var uri string
		if err := json.Unmarshal(v.Target, &uri); err != nil {
			t.Fatal(err)
		}
		if uri != fixture.TargetURI {
			t.Fatal("clearing pin changed endpoint identity")
		}
	})
	deleteOwned("owned Link final delete", "links/edition")

	// The old writer can pin an unowned source. Owned source pins and endpoint
	// identity/Link Type changes are not admitted by this source's writer API.
	var unowned graphstore.LinkMutationResult
	write(func(s *graphstore.Store) (err error) {
		unowned, err = s.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/unowned-edition", SourcePath: "beads/source", TargetPath: "beads/target", SourcePin: graphstore.EndpointPin{Revision: owner.Revision()}})
		return err
	})
	unownedLink := get(unowned.Link.ID)
	var pinned struct{ URI, Revision string }
	if err := json.Unmarshal(decode(unownedLink).Source, &pinned); err != nil {
		t.Fatal(err)
	}
	if pinned.URI != owner.ID() || pinned.Revision != owner.Revision() {
		t.Fatal("unowned source pin was lost")
	}
	var floated graphstore.LinkMutationResult
	write(func(s *graphstore.Store) (err error) {
		floated, err = s.UpdateLink(ctx, graphstore.LinkUpdateRequest{Path: "links/unowned-edition", ExpectedRevision: unownedLink.Revision(), SourcePin: &graphstore.EndpointPin{}})
		return err
	})
	unownedLink = edition("unowned source pin clear", unownedLink, floated.Link.Revision)
	var floating string
	if err := json.Unmarshal(decode(unownedLink).Source, &floating); err != nil || floating != owner.ID() {
		t.Fatalf("source pin clear changed identity: %s %v", floating, err)
	}
	var repinned graphstore.LinkMutationResult
	write(func(s *graphstore.Store) (err error) {
		repinned, err = s.UpdateLink(ctx, graphstore.LinkUpdateRequest{Path: "links/unowned-edition", ExpectedRevision: unownedLink.Revision(), SourcePin: &graphstore.EndpointPin{Revision: owner.Revision()}})
		return err
	})
	unownedLink = edition("unowned source pin set", unownedLink, repinned.Link.Revision)
	if err := json.Unmarshal(decode(unownedLink).Source, &pinned); err != nil || pinned.URI != owner.ID() || pinned.Revision != owner.Revision() {
		t.Fatalf("source repin changed Reference: %+v %v", pinned, err)
	}
	assertCurrentEditionPayload(t, unownedLink, repinned.Link.Properties, repinned.Link.Metadata)
	if current := get(owner.ID()); current.Revision() != owner.Revision() {
		t.Fatal("unowned Link writes changed Bead edition")
	}

	var retyped graphstore.MemoryMutationResult
	write(func(s *graphstore.Store) (err error) {
		retyped, err = s.PatchMemory(ctx, graphstore.MemoryPatchRequest{Path: "beads/source", ExpectedRevision: owner.Revision(), TypeURL: o.ScopeURL + "constraint"})
		return err
	})
	owner = edition("Memory affiliation", owner, retyped.Memory.Revision)
	assertCurrentEditionPayload(t, owner, retyped.Memory.Properties, retyped.Memory.Metadata)
	if decode(owner).Type != o.ScopeURL+"constraint" {
		t.Fatal("Memory affiliation missing from current Get")
	}

	// Native Issue affiliation and lifecycle properties use their domain writer,
	// rather than pretending every Bead is a generic Memory record.
	var issue graphstore.IssueRecord
	write(func(s *graphstore.Store) (err error) {
		issue, err = s.CreateIssue(ctx, "beads/edition-work", publicops.CreateRequest{Issue: &publicops.Issue{ID: "edition-work", Title: "Issue", Status: publicops.StatusOpen, IssueType: "task", Priority: 2}, ForceIDPrefix: true, Actor: "edition-author"})
		return err
	})
	issueDoc := get(issue.ID)
	if issueDoc.Revision() != issue.Revision {
		t.Fatal("Issue initial edition differs")
	}
	issueTitle, description, issueStatus := "Revised Issue", "Issue body", publicops.StatusInProgress
	bugType, err := graphstore.NativeIssueTypeURL(o.ScopeURL, "bug")
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		label, key, want      string
		request               graphstore.UpdateIssueRequest
		metadata, affiliation bool
	}{
		{label: "Issue title", key: "title", want: `"Revised Issue"`, request: graphstore.UpdateIssueRequest{Title: &issueTitle}},
		{label: "Issue body", key: "description", want: `"Issue body"`, request: graphstore.UpdateIssueRequest{Description: &description}},
		{label: "Issue status", key: "status", want: `"in_progress"`, request: graphstore.UpdateIssueRequest{Status: &issueStatus}},
		{label: "Issue metadata", key: "edition", want: `true`, metadata: true, request: graphstore.UpdateIssueRequest{Metadata: publicops.MetadataPatch{Set: map[string]json.RawMessage{"edition": json.RawMessage(`true`)}}}},
		{label: "Issue affiliation", key: "issue_type", want: `"bug"`, affiliation: true, request: graphstore.UpdateIssueRequest{TypeURL: bugType}},
	} {
		request := step.request
		request.Path, request.ExpectedRevision, request.Actor = "beads/edition-work", issueDoc.Revision(), "edition-editor"
		var result graphstore.IssueMutationResult
		write(func(s *graphstore.Store) (err error) { result, err = s.UpdateIssue(ctx, request); return err })
		if !result.Changed {
			t.Fatalf("%s: writer reported no change", step.label)
		}
		issueDoc = edition(step.label, issueDoc, result.Issue.Revision)
		assertCurrentEditionPayload(t, issueDoc, result.Issue.Properties, result.Issue.Metadata)
		body := decode(issueDoc)
		fields := body.Properties
		if step.metadata {
			fields = body.Metadata
		}
		field(step.label, fields, step.key, step.want)
		if step.affiliation && body.Type != bugType {
			t.Fatal("Issue canonical affiliation disagrees with native classification")
		}
	}
}

type currentEditionBody struct {
	Type           string
	Properties     map[string]json.RawMessage
	Metadata       map[string]json.RawMessage
	OwnedLinks     map[string][]json.RawMessage
	Source, Target json.RawMessage
}

// Compare complete values supplied by the normal writer, not merely the field
// just edited. This catches a point projection silently dropping sibling state.
func assertCurrentEditionPayload(t *testing.T, doc resourceget.NativeDocument, properties any, metadata json.RawMessage) {
	t.Helper()
	var got struct{ Properties, Metadata json.RawMessage }
	if err := json.Unmarshal(doc.NativeJSON(), &got); err != nil {
		t.Fatal(err)
	}
	wantProperties, err := json.Marshal(properties)
	if err != nil {
		t.Fatal(err)
	}
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	for name, pair := range map[string][2]json.RawMessage{"properties": {got.Properties, wantProperties}, "metadata": {got.Metadata, metadata}} {
		var actual, expected any
		if err := json.Unmarshal(pair[0], &actual); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(pair[1], &expected); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("complete %s differs from accepted writer state: got %s, want %s", name, pair[0], pair[1])
		}
	}
}

// Only the generated database allocated by prepareExistingFixture is removed.
// The explicit admin connection does not select an ambient/default database or
// stop/reconfigure the caller's configured fresh fixture server.
func registerCurrentGetServerCleanup(t *testing.T, database string, port int) {
	t.Helper()
	const prefix = "current_get_"
	suffix := strings.TrimPrefix(database, prefix)
	if suffix == database || suffix == "" || strings.IndexFunc(suffix, func(r rune) bool { return r < '0' || r > '9' }) >= 0 || port < 1 || port > 65535 {
		t.Fatal("refusing cleanup of an unknown fixture database or server")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cfg := mysql.NewConfig()
		cfg.Net, cfg.Addr, cfg.User = "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), "root"
		cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 10*time.Second, 10*time.Second, 10*time.Second
		connector, err := mysql.NewConnector(cfg)
		if err != nil {
			t.Error(err)
			return
		}
		admin := sql.OpenDB(connector)
		defer func() {
			if err := admin.Close(); err != nil {
				t.Error(err)
			}
		}()
		if _, err := admin.ExecContext(ctx, "DROP DATABASE `"+database+"`"); err != nil {
			t.Errorf("cleanup generated current Get fixture: %v", err)
		}
	})
}
