package bdpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/bdp"
	"github.com/steveyegge/beads/internal/bdp/bdpclient"
)

// discoveryServer serves canned GET / and GET /sys/capabilities bodies so the
// wire-DTO decode path runs end-to-end (HTTP + JSON) without a live clean-room.
func discoveryServer(t *testing.T, descriptor, capabilities string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/sys/capabilities", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(capabilities))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(descriptor))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

const descriptorRC2 = `{
  "bdp": "0.2.3",
  "base": "http://example.test",
  "links": [
    {"rel": "service-meta", "href": "/sys/capabilities"},
    {"rel": "service-desc", "href": "/bdp/openapi"},
    {"rel": "tag:bdp.dev,2026:rel/changes", "href": "/sys/changes"}
  ]
}`

// capsArrayForm is the clean-room "shape gap": core is an ARRAY of singleton
// op objects, and put/patch/delete carry the HTTP header-name guard keys.
const capsArrayForm = `{
  "core": [
    {"get": {"selector": false}},
    {"list": {"conditions": ["compare", "isType"], "selector": true, "sort": false}},
    {"create": {"serverAllocatesId": true, "clientProposedId": true}},
    {"put": {"ifMatch": true, "ifNoneMatch": true, "precondition": false}},
    {"patch": {"ifMatch": true, "precondition": false}},
    {"delete": {"ifMatch": true, "precondition": false}},
    {"revisions": {}},
    {"export": {"revisions": true}},
    {"resolve": {"selector": false}}
  ],
  "scope": {
    "storageClasses": ["durable"],
    "revisionModel": "versioned",
    "idModel": "permanent",
    "bodyFormSpill": false
  }
}`

func TestDiscoverDescriptorAndCapabilities(t *testing.T) {
	srv := discoveryServer(t, descriptorRC2, capsArrayForm)
	s := bdpclient.New(srv.URL)

	desc, caps, err := s.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	// Descriptor: the wire "bdp" key maps to Version; the RFC 8288 links array
	// folds into the named-field ServiceLinks.
	if desc.Version != "0.2.3" {
		t.Errorf("Version = %q, want 0.2.3", desc.Version)
	}
	if desc.Links.ServiceMeta != "/sys/capabilities" {
		t.Errorf("ServiceMeta = %q", desc.Links.ServiceMeta)
	}
	if desc.Links.Changes != "/sys/changes" {
		t.Errorf("Changes = %q", desc.Links.Changes)
	}

	// Capabilities: presence drives Offers; a non-Core op (resolve) is ignored.
	for _, p := range []bdp.Primitive{
		bdp.PrimitiveGet, bdp.PrimitiveList, bdp.PrimitiveCreate,
		bdp.PrimitivePut, bdp.PrimitivePatch, bdp.PrimitiveDelete,
		bdp.PrimitiveRevisions, bdp.PrimitiveExport,
	} {
		if !caps.Offers(p) {
			t.Errorf("Offers(%s) = false, want true", p)
		}
	}
	if caps.Offers(bdp.PrimitivePull) {
		t.Error("Offers(pull) = true, but pull was not advertised")
	}
	if caps.Offers(bdp.PrimitiveImport) {
		t.Error("Offers(import) = true, but import was not advertised")
	}

	// Guard-key remap: wire ifMatch/ifNoneMatch → contract IfRevision/IfAbsent.
	if caps.Core.Put == nil || !caps.Core.Put.IfRevision || !caps.Core.Put.IfAbsent {
		t.Errorf("Put caps remap wrong: %+v", caps.Core.Put)
	}

	// Scope-wide traits.
	if caps.Scope.RevisionModel != bdp.RevisionModelVersioned {
		t.Errorf("RevisionModel = %q", caps.Scope.RevisionModel)
	}
	if caps.Scope.IdModel != bdp.IdModelPermanent {
		t.Errorf("IdModel = %q", caps.Scope.IdModel)
	}
	if caps.Unversioned() {
		t.Error("Unversioned() = true for a versioned scope")
	}
	if len(caps.Scope.StorageClasses) != 1 || caps.Scope.StorageClasses[0] != bdp.StorageClassDurable {
		t.Errorf("StorageClasses = %v", caps.Scope.StorageClasses)
	}
}

// The decoder must also accept the ratified RECORD form of `core`, so the client
// works against a spec-faithful server as well as the clean-room.
func TestCapabilitiesRecordForm(t *testing.T) {
	const recordForm = `{
      "core": {
        "get": {"selector": false},
        "put": {"ifMatch": true, "ifNoneMatch": false, "precondition": true}
      },
      "scope": {"storageClasses": ["durable"], "revisionModel": "unversioned", "idModel": "reusable"}
    }`
	srv := discoveryServer(t, descriptorRC2, recordForm)
	s := bdpclient.New(srv.URL)

	caps, err := s.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("Capabilities: %v", err)
	}
	if !caps.Offers(bdp.PrimitiveGet) || !caps.Offers(bdp.PrimitivePut) {
		t.Error("record-form core did not decode get/put")
	}
	if caps.Core.Put == nil || !caps.Core.Put.IfRevision || caps.Core.Put.IfAbsent || !caps.Core.Put.Precondition {
		t.Errorf("record-form Put caps wrong: %+v", caps.Core.Put)
	}
	if !caps.Unversioned() {
		t.Error("Unversioned() = false for an unversioned scope")
	}
	if caps.Scope.IdModel != bdp.IdModelReusable {
		t.Errorf("IdModel = %q, want reusable", caps.Scope.IdModel)
	}
}

// Version negotiation refuses an incompatible minor, and nothing is attached.
func TestDiscoverRejectsIncompatibleVersion(t *testing.T) {
	srv := discoveryServer(t, `{"bdp":"0.3.0","links":[]}`, capsArrayForm)
	s := bdpclient.New(srv.URL)

	_, _, err := s.Discover(context.Background())
	if err == nil {
		t.Fatal("Discover: want version-negotiation error for 0.3.0, got nil")
	}
	if !strings.Contains(err.Error(), "0.3.0") {
		t.Errorf("error should name the incompatible version: %v", err)
	}
}

// A required guarantee tier with an unrecognized value must fault, never be
// guessed into a guarantee the server never made.
func TestCapabilitiesRejectsUnknownIdModel(t *testing.T) {
	const badScope = `{"core": [{"get": {}}], "scope": {"storageClasses": ["durable"], "revisionModel": "versioned", "idModel": "sometimes"}}`
	srv := discoveryServer(t, descriptorRC2, badScope)
	s := bdpclient.New(srv.URL)

	_, err := s.Capabilities(context.Background())
	if err == nil {
		t.Fatal("Capabilities: want error for unknown idModel, got nil")
	}
}
