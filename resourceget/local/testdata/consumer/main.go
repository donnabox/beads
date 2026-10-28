// This standalone module deliberately imports no internal package or root beads
// facade. The parent harness supplies an already-existing disposable workspace.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/steveyegge/beads/resourceget"
	"github.com/steveyegge/beads/resourceget/local"
)

type input struct {
	Options                                                      local.Options
	BeadID, LinkID, BeadURI, LinkURI, BeadRevision, LinkRevision string
	LinkType, TargetURI, TargetRevision                          string
}

// A consumer may decorate the supported public role without obtaining storage
// or reconstructing a source. This sample denies reads outside its explicit set.
type limitedReader struct {
	reader  resourceget.Reader
	allowed map[string]bool
	calls   int
}

func (r *limitedReader) Get(ctx context.Context, request resourceget.Request) (resourceget.NativeDocument, error) {
	if !r.allowed[request.ID] {
		return resourceget.NativeDocument{}, resourceget.ErrInvalidRequest
	}
	r.calls++
	return r.reader.Get(ctx, request)
}

func run() error {
	var in input
	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&in); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	handle, err := local.OpenExisting(ctx, in.Options)
	if err != nil {
		return err
	}
	defer handle.Close()
	reader := &limitedReader{reader: handle, allowed: map[string]bool{in.BeadID: true, in.LinkID: true}}
	bead, err := reader.Get(ctx, resourceget.Request{ID: in.BeadID, AssertKind: resourceget.Bead})
	if err != nil {
		return err
	}
	link, err := reader.Get(ctx, resourceget.Request{ID: in.LinkID, AssertKind: resourceget.Link})
	if err != nil {
		return err
	}
	if bead.ID() != in.BeadURI || link.ID() != in.LinkURI || bead.Revision() != in.BeadRevision || link.Revision() != in.LinkRevision {
		return errors.New("read identity or revision changed")
	}
	if _, err := reader.Get(ctx, resourceget.Request{ID: "not-authorized"}); !errors.Is(err, resourceget.ErrInvalidRequest) || reader.calls != 2 {
		return errors.New("decorator was bypassed")
	}
	for _, doc := range []resourceget.NativeDocument{bead, link} {
		contract := doc.Contract()
		if contract.SpecCommit != "bad780f19761e5714873aa3a85adeb75bc885928" || contract.BDPVersion != "0" {
			return errors.New("unexpected shipping contract")
		}
		original := doc.NativeJSON()
		mutated := doc.NativeJSON()
		mutated[0] = '!'
		if !bytes.Equal(original, doc.NativeJSON()) {
			return errors.New("document bytes were mutable")
		}
		if _, err := resourceget.ParseLegacyCurrent(in.Options.ScopeURL, doc.Kind(), original); err != nil {
			return err
		}
	}
	var linkBody struct {
		Source   json.RawMessage
		Target   struct{ URI, Revision string }
		Metadata map[string]json.RawMessage
	}
	if err := json.Unmarshal(link.NativeJSON(), &linkBody); err != nil {
		return err
	}
	wantSource, _ := json.Marshal(in.BeadURI)
	if !bytes.Equal(linkBody.Source, wantSource) || linkBody.Target.URI != in.TargetURI || linkBody.Target.Revision != in.TargetRevision || string(linkBody.Metadata["reviewed"]) != "true" {
		return errors.New("whole endpoint Reference or metadata changed")
	}
	var beadBody struct {
		OwnedLinks map[string][]json.RawMessage
		Properties map[string]json.RawMessage
		Metadata   map[string]json.RawMessage
	}
	if err := json.Unmarshal(bead.NativeJSON(), &beadBody); err != nil {
		return err
	}
	if len(beadBody.OwnedLinks[in.LinkType]) != 1 || string(beadBody.Properties["body"]) != `""` || string(beadBody.Metadata["reviewed"]) != "true" {
		return errors.New("incomplete current owned state/properties/metadata")
	}
	var nested struct {
		ID, Revision string
		Metadata     map[string]json.RawMessage
		Target       struct{ URI, Revision string }
	}
	if err := json.Unmarshal(beadBody.OwnedLinks[in.LinkType][0], &nested); err != nil {
		return err
	}
	if nested.ID != link.ID() || nested.Revision != link.Revision() || nested.Metadata == nil || nested.Target.URI != in.TargetURI || nested.Target.Revision != in.TargetRevision {
		return errors.New("owned Link and direct Get disagree")
	}
	if err := handle.Close(); err != nil {
		return err
	}
	if _, err := handle.Get(ctx, resourceget.Request{ID: in.BeadID}); !errors.Is(err, resourceget.ErrClosed) {
		return errors.New("closed source remained readable")
	}
	return json.NewEncoder(os.Stdout).Encode(struct{ Bead, Link resourceget.NativeDocument }{bead, link})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
