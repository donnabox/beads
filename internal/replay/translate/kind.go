package translate

import "fmt"

// String is the name a result row records for the kind of an action. A kind this
// package does not name is still named, by its number, never reported as nothing.
func (k ActionKind) String() string {
	switch k {
	case KindCreate:
		return "create"
	case KindUpdate:
		return "update"
	case KindDelete:
		return "delete"
	case KindDepAdd:
		return "dep_add"
	case KindDepRemove:
		return "dep_remove"
	case KindClose:
		return "close"
	case KindNoop:
		return "noop"
	}
	return fmt.Sprintf("kind(%d)", int(k))
}
