package main

import "fmt"

// This is an initial, explicitly bounded surface policy over the existing
// initialized command tree. It neither constructs nor reparents Cobra commands.
// The candidate flag table still requires owner review; it is not a declaration
// that every canonical command or backend capability is implemented.
type canonicalReadLeafOperation string

const (
	canonicalReadLeafListBeads canonicalReadLeafOperation = "list-beads"
	canonicalReadLeafListLinks canonicalReadLeafOperation = "list-links"
	canonicalReadLeafQuery     canonicalReadLeafOperation = "query"
)

type canonicalReadLeafDescriptor struct {
	Operation          canonicalReadLeafOperation
	Path               []string
	Flags              []string
	MaximumPositionals int
	RequiresAST        bool
}

// Return fresh slices: profile decoration must not mutate shared descriptors.
// Operational inherited flags match the current canonicalGraphFlags boundary.
// --json is common presentation, matching canonicalOutputDefault. Legacy
// --format, old Issue selectors and --conforms-to are not promoted by this proposal. Query projection stays in the AST, so --include is not
// advertised there. The outer control-plane router owns help/config/completion.
func lookupCanonicalReadLeaf(path []string) (canonicalReadLeafDescriptor, bool) {
	common := []string{"directory", "graph-mode", "readonly", "quiet", "no-color", "actor", "json"}
	collection := []string{"output-format", "include", "omit", "type", "exact-type", "ast", "property-field", "has-property-key", "metadata-field", "has-metadata-key", "count-only", "paged", "unpaged", "page-size", "limit", "continue"}
	var result canonicalReadLeafDescriptor
	switch {
	case len(path) == 1 && path[0] == "list":
		result = canonicalReadLeafDescriptor{Operation: canonicalReadLeafListBeads, Path: []string{"list"}, Flags: append(collection, common...)}
	case len(path) == 2 && path[0] == "link" && path[1] == "list":
		flags := append(collection, "from", "to", "incident-to", "direction", "all-revisions")
		result = canonicalReadLeafDescriptor{Operation: canonicalReadLeafListLinks, Path: []string{"link", "list"}, Flags: append(flags, common...), MaximumPositionals: 2}
	case len(path) == 1 && path[0] == "query":
		flags := []string{"ast", "output-format", "omit"}
		result = canonicalReadLeafDescriptor{Operation: canonicalReadLeafQuery, Path: []string{"query"}, Flags: append(flags, common...), RequiresAST: true}
	default:
		return lookupCanonicalRetainedLeaf(path)
	}
	return result, true
}

// Every entry means explicitly present, including false/default values.
// Normalize a shorthand through flag-name metadata without calling Value.Set;
// inherited flags must be included rather than silently dropped. Values are
// deliberately absent: this gate cannot confuse false with "not supplied".
type canonicalReadExplicitFlag struct {
	Name      string
	Shorthand string
	Inherited bool
}

type canonicalReadLeafInvocation struct {
	Profile     cliSurfaceProfile
	Path        []string // exact resolved leaf path, without the executable name
	AliasUsed   bool
	Flags       []canonicalReadExplicitFlag
	Positionals int
}

type canonicalReadLeafAdmission struct {
	Canonical  bool
	Descriptor canonicalReadLeafDescriptor
}

type canonicalReadLeafRefusal struct {
	Code     string
	ExitCode int
	Message  string
}

func (e *canonicalReadLeafRefusal) Error() string { return e.Code + ": " + e.Message }

// The existing-tree adapter blocks prohibited Value.Set calls during normal
// pflag parsing, then runs this policy in its Args wrapper before original Args
// and all command hooks/acquisition. No production Cobra OnInitialize callback
// may acquire externally: Cobra runs those callbacks before Args. Outer routing
// owns help/version/completion, which can bypass Args.
//
// Compat returns no decoration/admission: its exact existing commands, pointer
// identity, flag objects, parser errors and executor routes remain untouched.
// A successful canonical result permits only leaf output-default decoration;
// it is not a store/Type/authority/capability admission. Existing checked read
// executors must still perform their complete normal validation afterward.
func admitCanonicalReadLeaf(input canonicalReadLeafInvocation) (canonicalReadLeafAdmission, error) {
	if input.Profile == cliSurfaceCompat {
		return canonicalReadLeafAdmission{}, nil
	}
	refuse := func(code string, exit int, message string) (canonicalReadLeafAdmission, error) {
		return canonicalReadLeafAdmission{}, &canonicalReadLeafRefusal{Code: code, ExitCode: exit, Message: message}
	}
	if input.Profile != cliSurfaceCanonical {
		return refuse("invalid_selector", 2, "unresolved CLI surface profile")
	}
	descriptor, covered := lookupCanonicalReadLeaf(input.Path)
	if !covered {
		return refuse("capability_unavailable", 5, "canonical operation is outside the selected finite surface")
	}
	if input.AliasUsed {
		return refuse("invalid_selector", 2, "use the canonical command path rather than a compatibility alias")
	}
	if input.Positionals < 0 || input.Positionals > descriptor.MaximumPositionals {
		return refuse("invalid_selector", 2, "unexpected positional arguments for this canonical operation")
	}
	allowed := make(map[string]bool, len(descriptor.Flags))
	for _, name := range descriptor.Flags {
		allowed[name] = true
	}
	hasAST := false
	for _, explicit := range input.Flags {
		if !allowed[explicit.Name] {
			return refuse("invalid_selector", 2, fmt.Sprintf("--%s is not in this canonical operation's surface", explicit.Name))
		}
		hasAST = hasAST || explicit.Name == "ast"
	}
	if descriptor.RequiresAST && !hasAST {
		return refuse("invalid_selector", 2, "canonical Query requires --ast")
	}
	return canonicalReadLeafAdmission{Canonical: true, Descriptor: descriptor}, nil
}
