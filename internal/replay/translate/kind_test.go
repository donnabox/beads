package translate

import "testing"

// Each kind of action has a stable name, because a result row records the kinds
// of what it replayed. A kind this package does not know is still named, by its
// number, rather than being reported as nothing.
func TestActionKindString(t *testing.T) {
	want := map[ActionKind]string{
		KindCreate:    "create",
		KindUpdate:    "update",
		KindDelete:    "delete",
		KindDepAdd:    "dep_add",
		KindDepRemove: "dep_remove",
		KindClose:     "close",
		KindNoop:      "noop",
	}
	seen := map[string]ActionKind{}
	for kind, name := range want {
		if got := kind.String(); got != name {
			t.Errorf("ActionKind(%d).String() = %q, want %q", int(kind), got, name)
		}
		if other, dup := seen[name]; dup {
			t.Errorf("%d and %d share the name %q", int(other), int(kind), name)
		}
		seen[name] = kind
	}
	if got := ActionKind(99).String(); got != "kind(99)" {
		t.Errorf("an unknown kind is named %q, want kind(99)", got)
	}
}
