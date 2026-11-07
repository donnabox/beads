// Package typeref is the one place bd serve and its client know how a Type is
// spelled on the BDP surfaces: the shared identities of Issue and Dependency Types
// §2, and the Types rule of specs/types/issue/callables.md §2, which says where an
// Issue Callable carries a Bead's Type or a dependency's kind.
//
// Core's projection (internal/httpapi) reads its fixed tables from here, and so do
// the Callables mount and `bd connect --protocol bdp`, so the three cannot drift.
// What a Scope-local name means is the Scope's: this package spells the shared
// namespaces and the prefix, and leaves the configured vocabulary to its caller.
package typeref

import "strings"

// The shared identities. A native name is a suffix of its namespace, and the
// tables below are what stops a suffix being derived from anything else.
const (
	BeadNamespace = "https://github.com/gastownhall/bdp/types/issue/"
	LinkNamespace = "https://github.com/gastownhall/bdp/types/dependency/"
	BeadFamily    = BeadNamespace + "issue"
	LinkFamily    = LinkNamespace + "dependency"

	// LocalBeadTypes and LocalLinkTypes follow a Scope URL (which ends in `/`) to
	// spell a configured name's Scope-local identity.
	LocalBeadTypes = "types/issue/"
	LocalLinkTypes = "types/dependency/"
)

// BeadTypeNames is Issue §2's Bead table: the native classifications with a shared
// identity.
var BeadTypeNames = map[string]bool{
	"task": true, "bug": true, "feature": true, "epic": true, "chore": true, "decision": true,
	"message": true, "molecule": true, "gate": true, "spike": true, "story": true,
	"milestone": true, "event": true,
}

// LinkTypeNames is Issue §2's Link table.
var LinkTypeNames = map[string]bool{
	"blocks": true, "parent-child": true, "conditional-blocks": true, "waits-for": true,
	"related": true, "discovered-from": true, "replies-to": true, "relates-to": true,
	"duplicates": true, "supersedes": true, "authored-by": true, "assigned-to": true,
	"approved-by": true, "attests": true, "tracks": true, "until": true, "caused-by": true,
	"validates": true, "delegated-from": true,
}

// Kind says which table a member's value is read against.
type Kind uint8

const (
	// None is a member the Types rule does not reach.
	None Kind = iota
	// Bead is a Bead's Type, a classification.
	Bead
	// Link is a dependency's kind.
	Link
)

func (k Kind) String() string {
	switch k {
	case Bead:
		return "Bead"
	case Link:
		return "Link"
	}
	return "none"
}

// The Types rule, as callables.md §2's table states it. Operation and component
// names are openapi.v0.yaml's.
var (
	// parameters are v0 query parameters, by operation. Their names stay.
	parameters = map[string]map[string]Kind{
		"listReadyWork":        {"type": Bead, "exclude_type": Bead},
		"countReadyWork":       {"type": Bead, "exclude_type": Bead},
		"listIssues":           {"type": Bead, "exclude_type": Bead},
		"countIssues":          {"type": Bead, "exclude_type": Bead},
		"claimNextIssue":       {"type": Bead, "exclude_type": Bead},
		"listRelatedIssues":    {"type": Link},
		"listDependencies":     {"type": Link},
		"countDependencyEdges": {"type": Link},
	}
	// adapterParameters are arguments no v0 operation declares (callables.md §3).
	adapterParameters = map[string]map[string]Kind{
		"addDependencies": {"type": Link},
	}
	// members are v0 component members, by component.
	members = map[string]map[string]Kind{
		"Issue":                       {"issue_type": Bead},
		"BatchGetIssue":               {"issue_type": Bead},
		"IssueWithCounts":             {"issue_type": Bead},
		"IssueDetails":                {"issue_type": Bead},
		"IssueWithDependencyMetadata": {"issue_type": Bead},
		"TreeNode":                    {"issue_type": Bead},
		"BlockedIssue":                {"issue_type": Bead},
		"CreateIssueRequest":          {"issue_type": Bead},
		"IssuePatchBody":              {"issue_type": Bead},
		"BatchCreateItem":             {"issue_type": Bead},
		"ApplyCreateItem":             {"issue_type": Bead},
		"ApplyPatchBody":              {"issue_type": Bead},
		"Dependency":                  {"type": Link},
		"DependencyEdge":              {"type": Link},
		"CreateIssueDependency":       {"type": Link},
		"BatchCreateDependency":       {"type": Link},
		"ApplyDepAddItem":             {"type": Link},
		"Blocker":                     {"type": Link},
	}
	// renamed is a v0 member the rule spells otherwise.
	renamed = map[string]string{"issue_type": "type"}
)

// Members v0 declares only as `object`, which the rule reaches all the same:
// EventRecord.issue is an Issue, and EventRecord.dep's `kind` is a Link Type
// spelled `type`.
const (
	eventRecord        = "EventRecord"
	eventIssue         = "issue"
	eventDep           = "dep"
	eventDepKind       = "kind"
	issueComponent     = "Issue"
	problemComponent   = "Problem"
	callableTypeMember = "type"
)

// IsReference says a value is spelled as a Type reference: an absolute HTTP(S) URI,
// Core's typeReference without the pin, which nothing here stores.
func IsReference(value string) bool {
	return strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://")
}

// CallableName is the name a v0 member or parameter has on a Callable.
func CallableName(v0 string) string {
	if name, ok := renamed[v0]; ok {
		return name
	}
	return v0
}

// ParameterKind is what the rule says of one of an operation's v0 query parameters
// or adapter arguments.
func ParameterKind(op, name string) Kind {
	if kind := parameters[op][name]; kind != None {
		return kind
	}
	return adapterParameters[op][name]
}

// MemberKind is what the rule says of one v0 component member.
func MemberKind(component, member string) Kind {
	return members[component][member]
}

// ReferenceSchema is the schema of a member the rule reaches: a Type reference, or
// an array of them where v0 has an array. note is v0's description.
func ReferenceSchema(kind Kind, v0 map[string]any) map[string]any {
	description := kind.String() + " Type reference (callables.md §2)."
	if text, ok := v0["description"].(string); ok && text != "" {
		description += " " + text
	}
	reference := map[string]any{"type": "string", "format": "uri", "pattern": "^https?://.+"}
	if v0["type"] == "array" {
		return map[string]any{"type": "array", "items": reference, "description": description}
	}
	reference["description"] = description
	return reference
}

// ProjectComponent is a v0 component schema under the rule: each member it
// reaches renamed and spelled as ReferenceSchema. A component it does not reach
// comes back as it is. The schema is not modified.
func ProjectComponent(component string, schema map[string]any) map[string]any {
	reach := members[component]
	if len(reach) == 0 {
		return schema
	}
	properties, _ := schema["properties"].(map[string]any)
	out := make(map[string]any, len(schema))
	for key, value := range schema {
		out[key] = value
	}
	projected := make(map[string]any, len(properties))
	for name, value := range properties {
		if kind := reach[name]; kind != None {
			v0, _ := value.(map[string]any)
			projected[CallableName(name)] = ReferenceSchema(kind, v0)
			continue
		}
		projected[name] = value
	}
	out["properties"] = projected
	if required, ok := schema["required"].([]any); ok {
		names := make([]any, len(required))
		for i, raw := range required {
			if name, ok := raw.(string); ok && reach[name] != None {
				names[i] = CallableName(name)
			} else {
				names[i] = raw
			}
		}
		out["required"] = names
	}
	return out
}
