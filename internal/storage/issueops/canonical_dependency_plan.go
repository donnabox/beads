package issueops

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/depid"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// CanonicalDependencyPlanLimit bounds each complete native Issue population and
// the total final pair transitions in this private candidate. It is an execution
// ceiling, not paging, a schema limit, or an advertised BDP capability.
const CanonicalDependencyPlanLimit = 1000

// CanonicalDependencyPair names actual native Issue IDs in one checked local
// population. Neither field is a public URL, alias, or revision selector.
type CanonicalDependencyPair struct {
	SourceID string
	TargetID string
}

// CanonicalExistingDependency asserts a native baseline identity and
// classification. SQL application must verify the actual row in its transaction;
// this pure planner cannot prove existence, ownership, or creation facts.
type CanonicalExistingDependency struct {
	Pair     CanonicalDependencyPair
	NativeID string
	Type     types.DependencyType
}

// CanonicalDependencyMetadataFinal supplies an exact final metadata object for
// an existing assertion. Its identity, endpoints, Type and thread stay unchanged.
type CanonicalDependencyMetadataFinal struct {
	Existing CanonicalExistingDependency
	Metadata json.RawMessage
}

// CanonicalDependencyInsertFinal describes a new assertion on an absent native
// pair. Missing metadata uses the ordinary new-assertion empty-object default.
// The native writer, not this request, supplies identity and creation facts.
type CanonicalDependencyInsertFinal struct {
	Pair     CanonicalDependencyPair
	Type     types.DependencyType
	Metadata json.RawMessage
}

// CanonicalDependencyReassertFinal replaces an old public assertion with a
// distinct new public assertion on the same native pair and classification.
// Graphstore must prove the two public identities are distinct. Creation facts
// and initial thread belong to the fresh native assertion, not the old row.
type CanonicalDependencyReassertFinal struct {
	Existing CanonicalExistingDependency
	Metadata json.RawMessage
}

// CanonicalDependencyFinalRequest describes final pair transitions, not batch
// members to replay. A pair appears exactly once across all four slices.
// Population is the complete admitted native Issue inventory; its actual
// completeness and durability remain transaction-local SQL checks.
type CanonicalDependencyFinalRequest struct {
	Actor      string
	Population []string
	Metadata   []CanonicalDependencyMetadataFinal
	Insert     []CanonicalDependencyInsertFinal
	Delete     []CanonicalExistingDependency
	Reassert   []CanonicalDependencyReassertFinal
}

// CanonicalDependencyTransition is the closed vocabulary produced by the pure
// normalizer. No request field can choose a transition outside this vocabulary.
type CanonicalDependencyTransition string

const (
	CanonicalDependencyMetadataTransition CanonicalDependencyTransition = "metadata"
	CanonicalDependencyInsertTransition   CanonicalDependencyTransition = "insert"
	CanonicalDependencyDeleteTransition   CanonicalDependencyTransition = "delete"
	CanonicalDependencyReassertTransition CanonicalDependencyTransition = "reassert"
)

// CanonicalDependencyIntent is a final pair transition for the domain SQL
// composite. Existing is nil only for Insert; Metadata is nil only for Delete.
// Type always agrees with Existing.Type when an existing assertion is present.
// It never supplies timestamps, thread edits, lifecycle, public allocation,
// endpoint correction, or reclassification of an existing assertion.
type CanonicalDependencyIntent struct {
	Transition CanonicalDependencyTransition
	Pair       CanonicalDependencyPair
	Existing   *CanonicalExistingDependency
	Type       types.DependencyType
	Metadata   json.RawMessage
}

// CanonicalDependencyIntentPlan is fully detached from its request. Population
// and Changes use UTF-16 code-unit order (source, then target for pairs).
// It proves finite input validity, not native final-graph validity or SQL
// admission. The SQL composite must normalize its own incoming request, rather
// than trust an externally constructed or modified plan.
type CanonicalDependencyIntentPlan struct {
	Actor      string
	Population []string
	Changes    []CanonicalDependencyIntent
}

// NormalizeCanonicalDependencyFinalRequest validates without side effects and
// returns one detached, deterministic final intent plan. Any error returns the
// zero plan. Actual native rows, expected row identity/classification, complete
// authority, template policy, cycles, hierarchy and readiness remain SQL-domain
// checks. Existing deletion uses ordinary removal admission, so it does not
// acquire the add-time dotted-child prohibition.
func NormalizeCanonicalDependencyFinalRequest(request CanonicalDependencyFinalRequest) (CanonicalDependencyIntentPlan, error) {
	if request.Actor == "" || !utf8.ValidString(request.Actor) {
		return CanonicalDependencyIntentPlan{}, fmt.Errorf("%w: canonical Dependency plan requires a nonempty UTF-8 actor", storage.ErrValidation)
	}
	if len(request.Population) > CanonicalDependencyPlanLimit {
		return CanonicalDependencyIntentPlan{}, fmt.Errorf("%w: canonical Dependency population exceeds %d Issues", storage.ErrValidation, CanonicalDependencyPlanLimit)
	}
	count := 0
	for _, size := range []int{len(request.Metadata), len(request.Insert), len(request.Delete), len(request.Reassert)} {
		if size > CanonicalDependencyPlanLimit-count {
			return CanonicalDependencyIntentPlan{}, fmt.Errorf("%w: canonical Dependency plan exceeds %d pair transitions", storage.ErrValidation, CanonicalDependencyPlanLimit)
		}
		count += size
	}
	population := make(map[string]bool, len(request.Population))
	plan := CanonicalDependencyIntentPlan{Actor: request.Actor, Population: append([]string{}, request.Population...), Changes: make([]CanonicalDependencyIntent, 0, count)}
	for _, id := range request.Population {
		if err := validateCanonicalNativeDependencyEndpoint(id); err != nil {
			return CanonicalDependencyIntentPlan{}, err
		}
		if population[id] {
			return CanonicalDependencyIntentPlan{}, fmt.Errorf("%w: duplicate canonical Dependency population identity %q", storage.ErrValidation, id)
		}
		population[id] = true
	}
	seen := make(map[CanonicalDependencyPair]bool, count)
	add := func(transition CanonicalDependencyTransition, pair CanonicalDependencyPair, existing *CanonicalExistingDependency, kind types.DependencyType, metadata json.RawMessage) error {
		if err := validateCanonicalDependencyIntentPair(pair, population); err != nil {
			return err
		}
		if seen[pair] {
			return fmt.Errorf("%w: duplicate or cross-transition canonical Dependency pair %q -> %q", storage.ErrValidation, pair.SourceID, pair.TargetID)
		}
		seen[pair] = true
		if !utf8.ValidString(string(kind)) || !kind.IsWellKnown() {
			return fmt.Errorf("%w: canonical Dependency plan requires a recognized native classification, got %q", storage.ErrValidation, kind)
		}
		if existing != nil && existing.NativeID != depid.New(pair.SourceID, pair.TargetID) {
			return fmt.Errorf("%w: canonical Dependency expected native ID does not match its pair", storage.ErrValidation)
		}
		if transition == CanonicalDependencyDeleteTransition {
			if err := ValidateRemoveDependencyRequest(publicops.RemoveDependencyRequest{Actor: request.Actor, IssueID: pair.SourceID, DependsOnID: pair.TargetID}); err != nil {
				return err
			}
		} else {
			if err := ValidateAddDependenciesRequest(publicops.AddDependenciesRequest{Actor: request.Actor, Edges: []publicops.DependencyEdge{{IssueID: pair.SourceID, DependsOnID: pair.TargetID, Type: kind}}}); err != nil {
				return err
			}
			if metadata == nil && transition != CanonicalDependencyMetadataTransition {
				metadata = json.RawMessage(`{}`)
			}
			if err := validateCanonicalMetadata(metadata); err != nil {
				return err
			}
			canonical, err := graph.CanonicalizeJSON(metadata)
			if err != nil {
				return fmt.Errorf("%w: canonical Dependency metadata: %v", storage.ErrValidation, err)
			}
			metadata = append(json.RawMessage(nil), canonical...)
		}
		if existing != nil {
			copyExisting := *existing
			existing = &copyExisting
		}
		plan.Changes = append(plan.Changes, CanonicalDependencyIntent{Transition: transition, Pair: pair, Existing: existing, Type: kind, Metadata: metadata})
		return nil
	}
	for _, item := range request.Metadata {
		if err := add(CanonicalDependencyMetadataTransition, item.Existing.Pair, &item.Existing, item.Existing.Type, item.Metadata); err != nil {
			return CanonicalDependencyIntentPlan{}, err
		}
	}
	for _, item := range request.Insert {
		if err := add(CanonicalDependencyInsertTransition, item.Pair, nil, item.Type, item.Metadata); err != nil {
			return CanonicalDependencyIntentPlan{}, err
		}
	}
	for _, item := range request.Delete {
		if err := add(CanonicalDependencyDeleteTransition, item.Pair, &item, item.Type, nil); err != nil {
			return CanonicalDependencyIntentPlan{}, err
		}
	}
	for _, item := range request.Reassert {
		if err := add(CanonicalDependencyReassertTransition, item.Existing.Pair, &item.Existing, item.Existing.Type, item.Metadata); err != nil {
			return CanonicalDependencyIntentPlan{}, err
		}
	}
	slices.SortFunc(plan.Population, graph.CompareCodeUnits)
	slices.SortFunc(plan.Changes, func(a, b CanonicalDependencyIntent) int {
		if source := graph.CompareCodeUnits(a.Pair.SourceID, b.Pair.SourceID); source != 0 {
			return source
		}
		return graph.CompareCodeUnits(a.Pair.TargetID, b.Pair.TargetID)
	})
	return plan, nil
}

func validateCanonicalNativeDependencyEndpoint(id string) error {
	// depid.New documents this separator as impossible inside either natural
	// key component; preserve that invariant before deriving an expected ID.
	if id == "" || !utf8.ValidString(id) || strings.ContainsRune(id, '\x1f') {
		return fmt.Errorf("%w: canonical Dependency population requires nonempty UTF-8 native IDs without the native key separator", storage.ErrValidation)
	}
	return nil
}

func validateCanonicalDependencyIntentPair(pair CanonicalDependencyPair, population map[string]bool) error {
	if err := validateCanonicalNativeDependencyEndpoint(pair.SourceID); err != nil {
		return err
	}
	if err := validateCanonicalNativeDependencyEndpoint(pair.TargetID); err != nil {
		return err
	}
	if pair.SourceID == pair.TargetID {
		// Reuse the native ordinary sentinel and message rather than inventing
		// another self-edge diagnosis for this bounded candidate.
		return fmt.Errorf("%w: %s cannot depend on itself", publicops.ErrSelfDependency, pair.SourceID)
	}
	if !population[pair.SourceID] || !population[pair.TargetID] {
		return fmt.Errorf("%w: canonical Dependency pair is outside the admitted native population", storage.ErrValidation)
	}
	return nil
}
