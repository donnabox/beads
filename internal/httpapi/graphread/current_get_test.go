package graphread

import (
	"context"
	"errors"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

func TestCurrentSelectorUsesLiteralPublicIdentity(t *testing.T) {
	const scope = "https://example.test/get/"
	// A colon is admitted identity data, including URN-looking local IDs.
	// The public role does not reinterpret these literal IDs as foreign URLs.
	for _, id := range []string{"task-1", "team/task-1", "links/literal", "beads/literal", "urn:example:task"} {
		path, local, err := currentSelector(scope, resourceget.Request{ID: id})
		if err != nil || path != "beads/"+id || local != id {
			t.Fatalf("local %q: %q %q %v", id, path, local, err)
		}
		path, local, err = currentSelector(scope, resourceget.Request{ID: scope + id, AssertKind: resourceget.Link})
		if err != nil || path != scope+id || local != "" {
			t.Fatalf("URL %q: %q %q %v", id, path, local, err)
		}
	}
	for _, id := range []string{"", "https://foreign.test/get/task-1", scope + "task-1?q=x", scope + "task-1#part", "task-1?x", "task-1#part", "../task-1"} {
		if _, _, err := currentSelector(scope, resourceget.Request{ID: id}); !errors.Is(err, resourceget.ErrInvalidRequest) {
			t.Fatalf("accepted selector %q: %v", id, err)
		}
	}
	if _, _, err := currentSelector(scope, resourceget.Request{ID: "task-1", AssertKind: resourceget.Kind(255)}); !errors.Is(err, resourceget.ErrInvalidRequest) {
		t.Fatalf("accepted unknown kind: %v", err)
	}
}

func TestCurrentGetInvalidSourceAndCancellation(t *testing.T) {
	for _, reader := range []*Reader{nil, New(nil)} {
		if _, err := reader.Get(context.Background(), resourceget.Request{ID: "task-1"}); !errors.Is(err, resourceget.ErrInvalidSource) || !errors.Is(err, graphstore.ErrInvalidStore) {
			t.Fatalf("source error lost identity: %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := reader.Get(ctx, resourceget.Request{ID: "task-1"}); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation lost: %v", err)
		}
	}
	for native, public := range map[error]error{graphstore.ErrNotFound: resourceget.ErrNotFound, graphstore.ErrGone: resourceget.ErrGone, graphstore.ErrInvalidStore: resourceget.ErrInvalidSource} {
		if err := currentSourceError(native); !errors.Is(err, native) || !errors.Is(err, public) {
			t.Fatalf("native/public error identity lost: %v", err)
		}
	}
}
