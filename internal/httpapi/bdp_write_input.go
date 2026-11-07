package httpapi

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// What Write's requests mean for the Issue and Dependency Types: References and Type
// affiliations, the property policy of Issue §3.3, ordered JSON Patch, and Core's
// number admission. Nothing here reads the store; a value that depends on the row
// reaches it through beadEdit.patchFor, inside the write's transaction.

// issueServiceManaged is Issue §3.3's service-managed set: refused on any input, even
// when equal to the stored value (Write, "Properties replacement and protected input").
var issueServiceManaged = []string{
	"created_at", "created_by", "updated_at",
	"started_at", "closed_at", "close_reason",
	"arrival_seq", "closed_by_session", "compaction_level", "compacted_at", "compacted_at_commit", "original_size",
}

// issueWritable is the native write boundary of Issue §3.2: every other member a
// record may show is carried on reads and cannot be changed here.
var issueWritable = []string{
	"title", "description", "design", "acceptance_criteria", "notes", "spec_id",
	"status", "priority", "assignee", "owner", "estimated_minutes", "external_ref",
	"due_at", "defer_until", "labels",
}

// issueNullable are the members on which JSON null is a value, the clear.
var issueNullable = []string{"estimated_minutes", "external_ref", "due_at", "defer_until"}

// issueInstants are the writable members Issue §3.2 types as instants.
var issueInstants = []string{"due_at", "defer_until"}

// issueRequired are the members a complete Issue record has (Issue IT-4).
var issueRequired = []string{"title", "status", "priority"}

const issueClassificationMember = "issue_type"

func writeInvalid(diagnostics ...callableDiagnostic) *callableProblem {
	problem := newCallableProblem(bdpValidationFailed, diagnostics[0].Message)
	problem.Diagnostics = diagnostics
	return problem
}

// writeDiagnostic names a place in the failing properties or metadata document.
func writeDiagnostic(pointer, message string) callableDiagnostic {
	return callableDiagnostic{Message: message, InstanceLocation: pointer}
}

// writeRequestDiagnostic is a cause in a request member. Write says a request-body
// pointer is not a position in a document, so the message names the member instead.
func writeRequestDiagnostic(member, message string) callableDiagnostic {
	return callableDiagnostic{Message: "`" + member + "`: " + message}
}

// writeObjectMember decodes an optional object member, keeping number literals. An
// array or null is not an object: Write has them refused.
func writeObjectMember(members map[string]json.RawMessage, name string) (map[string]any, *callableProblem) {
	raw, sent := members[name]
	if !sent {
		return nil, nil
	}
	object, err := coreDecodeObject(raw)
	if err != nil {
		return nil, newCallableProblem(bdpMalformedRequest, "`"+name+"` is a JSON object")
	}
	return object, nil
}

// coreNumberAdmissible is Core's rule: the exact decimal value must survive conversion
// to the nearest binary64 and back through its shortest round-trip form.
func coreNumberAdmissible(literal string) bool {
	value, err := strconv.ParseFloat(literal, 64)
	if err != nil {
		return false
	}
	exact, ok := new(big.Rat).SetString(literal)
	if !ok {
		return false
	}
	shortest, ok := new(big.Rat).SetString(strconv.FormatFloat(value, 'g', -1, 64))
	return ok && exact.Cmp(shortest) == 0
}

// firstInadmissible finds a number Core refuses, at any depth, and its pointer.
func firstInadmissible(value any, pointer string) (string, bool) {
	switch v := value.(type) {
	case json.Number:
		return pointer, !coreNumberAdmissible(v.String())
	case map[string]any:
		names := make([]string, 0, len(v))
		for name := range v {
			names = append(names, name)
		}
		slices.Sort(names)
		for _, name := range names {
			if found, bad := firstInadmissible(v[name], pointer+"/"+jsonPointerEscape(name)); bad {
				return found, true
			}
		}
	case []any:
		for i, item := range v {
			if found, bad := firstInadmissible(item, pointer+"/"+strconv.Itoa(i)); bad {
				return found, true
			}
		}
	}
	return "", false
}

// checkAdmissibleNumbers holds a properties and a metadata document to Core's model.
func checkAdmissibleNumbers(properties, metadata map[string]any) *callableProblem {
	if pointer, bad := firstInadmissible(properties, ""); bad {
		return writeInvalid(writeDiagnostic(pointer, "the number at `"+pointer+"` in properties is not admissible under Core's number model"))
	}
	if pointer, bad := firstInadmissible(metadata, ""); bad {
		diagnostic := writeDiagnostic(pointer, "the number at `"+pointer+"` in metadata is not admissible under Core's number model")
		diagnostic.Document = "metadata"
		return writeInvalid(diagnostic)
	}
	return nil
}

func jsonPointerEscape(token string) string {
	return strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1")
}

// writeReference is one parsed durable Reference.
type writeReference struct {
	// local is the decoded Scope-relative path, one segment or several; "" for an
	// absolute URI outside this Scope.
	local string
}

// bead names the native id a Bead subject can have: one safe segment within the
// column, and not a Link's. Anything else names no Bead here.
func (ref writeReference) bead() (string, bool) {
	if ref.local == "" || strings.Contains(ref.local, "/") || types.CheckFieldLen("id", ref.local) != nil {
		return "", false
	}
	if _, _, _, isLink := coreParseLinkID(ref.local); isLink {
		return "", false
	}
	return ref.local, true
}

func (ref writeReference) link() (source, target string, generation int64, ok bool) {
	if ref.local == "" || strings.Contains(ref.local, "/") {
		return "", "", 0, false
	}
	return coreParseLinkID(ref.local)
}

// parseReference reads a durable Reference: a Scope-relative path, resolved against
// the Scope and never against the command URL, or an absolute URI. A noncanonical
// spelling is carrier syntax, refused and never normalized into another identity.
func (c *writeCall) parseReference(member, value string) (writeReference, *callableProblem) {
	malformed := func(detail string) (writeReference, *callableProblem) {
		return writeReference{}, newCallableProblem(bdpMalformedRequest, "`"+member+"`: "+detail)
	}
	if value == "" {
		return malformed("a Reference is a nonempty string")
	}
	if value[0] == '@' {
		return malformed("a `@name` binding is not valid in a single operation")
	}
	path := value
	if parsed, err := url.Parse(value); err == nil && parsed.IsAbs() {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return writeReference{}, nil
		}
		if !strings.HasPrefix(value, c.p.scope) {
			return writeReference{}, nil
		}
		path = strings.TrimPrefix(value, c.p.scope)
	}
	if strings.ContainsAny(path, "?#") {
		return malformed("a Resource Reference has no query or fragment")
	}
	lower := strings.ToLower(path)
	if strings.Contains(lower, "%2f") || strings.Contains(lower, "%5c") {
		return malformed("an encoded separator is not a canonical path")
	}
	decoded := make([]string, 0, 1)
	for _, segment := range strings.Split(path, "/") {
		plain, err := url.PathUnescape(segment)
		if err != nil || !coreSafeSegment(plain) {
			return malformed("this is not a canonical Scope-relative path")
		}
		decoded = append(decoded, plain)
	}
	return writeReference{local: strings.Join(decoded, "/")}, nil
}

// subject reads a Resource subject.
func (c *writeCall) subject(member string) (writeReference, *callableProblem) {
	var value string
	if err := json.Unmarshal(c.members[member], &value); err != nil || isJSONNull(c.members[member]) {
		return writeReference{}, newCallableProblem(bdpMalformedRequest, "`"+member+"` is a string naming a Resource")
	}
	return c.parseReference(member, value)
}

// createdBeadID reads createBead's optional `id`. The native id policy is one safe
// segment (Issue §3.2), and every other rule of native ids is the role's. A Link's
// spelling is a Resource of another kind, Donna's ruling keeps that identity-taken; a
// route's first segment is navigation, which Core forbids assigning.
func (c *writeCall) createdBeadID() (string, *callableProblem) {
	raw, sent := c.members["id"]
	if !sent {
		return "", nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || isJSONNull(raw) {
		return "", newCallableProblem(bdpMalformedRequest, "`id` is a string")
	}
	ref, problem := c.parseReference("id", value)
	if problem != nil {
		return "", problem
	}
	if ref.local == "" {
		return "", newCallableProblem(bdpMalformedRequest, "`id` names a Resource of this Scope")
	}
	if _, _, _, isLink := ref.link(); isLink {
		return "", newCallableProblem(bdpIdentityTaken, "this identity is a Link's; a Bead cannot take it")
	}
	refuse := func(reason string) (string, *callableProblem) {
		return "", writeInvalid(writeRequestDiagnostic("id", reason))
	}
	switch {
	case strings.Contains(ref.local, "/"):
		return refuse("this provider's Bead identities are one path segment")
	case strings.HasPrefix(ref.local, coreLinkIDPrefix):
		return refuse("`" + coreLinkIDPrefix + "` opens this provider's Link identities")
	case strings.ContainsRune(ref.local, '@'):
		return refuse("`@` is reserved for revision syntax and cannot be in a Bead identity")
	case types.CheckFieldLen("id", ref.local) != nil:
		return refuse("this identity is longer than storage holds")
	case c.s.reservedSegment(ref.local):
		return refuse("this path is protocol navigation and cannot be a Resource identity")
	}
	return ref.local, nil
}

// typeCarrier reads a Type affiliation down to its URI. A pin is a well-formed input
// nothing here can store, so it refuses as a value and never loses its revision.
func typeCarrier(raw json.RawMessage) (string, *callableProblem) {
	var uri string
	if err := json.Unmarshal(raw, &uri); err == nil && !isJSONNull(raw) {
		if parsed, err := url.Parse(uri); err != nil || !parsed.IsAbs() || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return "", newCallableProblem(bdpMalformedRequest, "`type` is an absolute HTTP(S) Type URI")
		}
		return uri, nil
	}
	var pinned map[string]string
	if err := json.Unmarshal(raw, &pinned); err == nil && len(pinned) == 2 && pinned["uri"] != "" && pinned["revision"] != "" {
		return "", writeInvalid(writeRequestDiagnostic("type", "this provider stores no Type definition pin; send the Type URI alone"))
	}
	return "", newCallableProblem(bdpMalformedRequest, "`type` is an absolute HTTP(S) Type URI or exactly `uri` and `revision`")
}

// beadType is Issue §2's table read backwards. Anything outside it refuses: the
// abstract base, a Link Type, and every Type that is not this package's.
func (c *writeCall) beadType(raw json.RawMessage) (string, *callableProblem) {
	uri, problem := typeCarrier(raw)
	if problem != nil {
		return "", problem
	}
	refuse := func(reason string) (string, *callableProblem) {
		return "", writeInvalid(writeRequestDiagnostic("type", reason))
	}
	if name, ok := strings.CutPrefix(uri, coreBeadTypeNamespace); ok {
		switch {
		case uri == coreBeadFamily:
			return refuse("`issue/issue` is abstract and admits no direct instance; send a concrete Type")
		case coreBeadTypeNames[name]:
			return name, nil
		}
		return refuse("the Issue package has no Type `" + name + "`")
	}
	if strings.HasPrefix(uri, coreLinkTypeNamespace) {
		return refuse("this is a Dependency Link Type, and a Bead has an Issue Type")
	}
	if name, ok := strings.CutPrefix(uri, c.p.scope+coreLocalBeadTypes); ok {
		plain, err := url.PathUnescape(name)
		if err != nil {
			return refuse("this is not a Type this Scope configures")
		}
		configured, err := c.p.configuredTypes(c.ctx)
		if err == nil && configured[plain] {
			return plain, nil
		}
		return refuse("this is not a Type this Scope configures")
	}
	return refuse("this Scope writes the Issue and Dependency Types only")
}

// checkProtectedMember refuses input on a member Issue §3.3 does not let a client
// author. It is about the attempt, so a patch that touches one and undoes it refuses.
func checkProtectedMember(name string) *callableDiagnostic {
	switch {
	case slices.Contains(issueServiceManaged, name):
		d := writeDiagnostic("/"+jsonPointerEscape(name), "`"+name+"` is service-managed and cannot be supplied")
		return &d
	case name == issueClassificationMember:
		d := writeDiagnostic("/"+name, "`issue_type` is refused: a record's classification is its `type`")
		return &d
	}
	return nil
}

// setIssueMember validates one authored member and sets it on the patch: value is its
// JSON value, and absent says the member is being removed.
func setIssueMember(patch *issueops.IssuePatch, name string, value any, absent bool) *callableDiagnostic {
	refuse := func(message string) *callableDiagnostic {
		d := writeDiagnostic("/"+name, "`"+name+"` "+message)
		return &d
	}
	if !absent && value == nil && !slices.Contains(issueNullable, name) {
		return refuse("is not nullable")
	}
	text := func() (string, *callableDiagnostic) {
		if absent {
			return "", nil
		}
		s, ok := value.(string)
		if !ok {
			return "", refuse("is a string")
		}
		return s, nil
	}
	switch name {
	case "title":
		if absent {
			return refuse("is required")
		}
		s, bad := text()
		if bad != nil {
			return bad
		}
		if strings.TrimSpace(s) == "" {
			return refuse("must not be blank")
		}
		if err := types.ValidateIssueTitle(s); err != nil {
			return refuse("is refused: " + err.Error())
		}
		patch.Title = issueops.Field[string]{Set: true, Value: s}
	case "description", "design", "acceptance_criteria", "notes", "spec_id":
		s, bad := text()
		if bad != nil {
			return bad
		}
		field := issueops.Field[string]{Set: true, Value: s}
		switch name {
		case "description":
			patch.Description = field
		case "design":
			patch.Design = field
		case "acceptance_criteria":
			patch.AcceptanceCriteria = field
		case "notes":
			patch.Notes = field
		case "spec_id":
			patch.SpecID = field
		}
	case "status":
		if absent {
			return refuse("is required")
		}
		s, bad := text()
		if bad != nil {
			return bad
		}
		if s == "" || types.CheckFieldLen(name, s) != nil {
			return refuse("is a nonempty status within storage's bound")
		}
		patch.Status = issueops.Field[issueops.Status]{Set: true, Value: issueops.Status(s)}
	case "priority":
		if absent {
			return refuse("is required")
		}
		number, ok := integerOf(value)
		if !ok {
			return refuse("is an integer")
		}
		if err := types.ValidateIssuePriority(number); err != nil {
			return refuse("is 0 to 4")
		}
		patch.Priority = issueops.Field[int]{Set: true, Value: number}
	case "assignee", "owner":
		s, bad := text()
		if bad != nil {
			return bad
		}
		if types.CheckFieldLen(name, s) != nil {
			return refuse("is longer than storage holds")
		}
		field := issueops.Field[string]{Set: true, Value: s}
		if name == "assignee" {
			patch.Assignee = field
		} else {
			patch.Owner = field
		}
	case "estimated_minutes":
		field := issueops.Field[*int]{Set: true}
		if !absent && value != nil {
			number, ok := integerOf(value)
			if !ok {
				return refuse("is an integer or null")
			}
			field.Value = &number
		}
		if err := types.ValidateIssueEstimatedMinutes(field.Value); err != nil {
			return refuse("is a nonnegative estimate")
		}
		patch.EstimatedMinutes = field
	case "external_ref":
		field := issueops.Field[*string]{Set: true}
		if !absent && value != nil {
			s, ok := value.(string)
			if !ok {
				return refuse("is a string or null")
			}
			if types.CheckFieldLen(name, s) != nil {
				return refuse("is longer than storage holds")
			}
			field.Value = &s
		}
		patch.ExternalRef = field
	case "due_at", "defer_until":
		field := issueops.Field[*time.Time]{Set: true}
		if !absent && value != nil {
			at, bad := instantOf(value)
			if bad != "" {
				return refuse(bad)
			}
			field.Value = &at
		}
		if name == "due_at" {
			patch.DueAt = field
		} else {
			patch.DeferUntil = field
		}
	case "labels":
		labels := []string{}
		if !absent {
			items, ok := value.([]any)
			if !ok {
				return refuse("is an array of strings")
			}
			seen := map[string]bool{}
			for _, item := range items {
				s, ok := item.(string)
				if !ok || s == "" || seen[s] || types.CheckFieldLen("label", s) != nil {
					return refuse("is an array of distinct nonempty strings within storage's bound")
				}
				seen[s] = true
				labels = append(labels, s)
			}
		}
		patch.Labels.Replace = issueops.Field[[]string]{Set: true, Value: labels}
	default:
		d := writeDiagnostic("/"+jsonPointerEscape(name), "`"+name+"` is outside this store's write boundary; it cannot be changed here")
		return &d
	}
	return nil
}

// integerOf reads an integer-valued JSON number that fits an int.
func integerOf(value any) (int, bool) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	exact, ok := new(big.Rat).SetString(number.String())
	if !ok || !exact.IsInt() || !exact.Num().IsInt64() {
		return 0, false
	}
	asInt := exact.Num().Int64()
	if int64(int(asInt)) != asInt {
		return 0, false
	}
	return int(asInt), true
}

// instantOf reads an RFC 3339 instant. These columns hold whole seconds, and a value
// that would round is refused: Write forbids rounding to obtain admission, and a record
// that came back with another time would be the silent kind of wrong.
func instantOf(value any) (time.Time, string) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, "is an RFC 3339 instant or null"
	}
	at, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}, "is an RFC 3339 instant or null"
	}
	if at.Nanosecond() != 0 {
		return time.Time{}, "is held to whole seconds here; send it without a fraction"
	}
	return at.UTC(), ""
}

// issueFromProperties is createBead's properties: Issue §3.2 less what §3.3 protects.
// Omitted required members are all named at once (IT-4). A status of `closed` cannot be
// created: closed_at is required with it and service-managed.
func issueFromProperties(properties map[string]any) (*types.Issue, *callableProblem) {
	var diagnostics []callableDiagnostic
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	slices.Sort(names)
	patch := issueops.IssuePatch{}
	for _, name := range names {
		if bad := checkProtectedMember(name); bad != nil {
			diagnostics = append(diagnostics, *bad)
			continue
		}
		if bad := setIssueMember(&patch, name, properties[name], false); bad != nil {
			diagnostics = append(diagnostics, *bad)
		}
	}
	for _, name := range issueRequired {
		if _, sent := properties[name]; !sent {
			diagnostics = append(diagnostics, writeDiagnostic("", "`"+name+"` is required: this package publishes no creation defaults"))
		}
	}
	if len(diagnostics) > 0 {
		return nil, writeInvalid(diagnostics...)
	}
	if patch.Status.Value == types.StatusClosed {
		return nil, writeInvalid(writeDiagnostic("/status", "`status` of `closed` needs `closed_at`, which is service-managed: create the Issue, then use the `issue/close` Callable"))
	}
	issue := &types.Issue{
		Title:              patch.Title.Value,
		Description:        patch.Description.Value,
		Design:             patch.Design.Value,
		AcceptanceCriteria: patch.AcceptanceCriteria.Value,
		Notes:              patch.Notes.Value,
		SpecID:             patch.SpecID.Value,
		Status:             patch.Status.Value,
		Priority:           patch.Priority.Value,
		Assignee:           patch.Assignee.Value,
		Owner:              patch.Owner.Value,
		EstimatedMinutes:   patch.EstimatedMinutes.Value,
		ExternalRef:        patch.ExternalRef.Value,
		DueAt:              patch.DueAt.Value,
		DeferUntil:         patch.DeferUntil.Value,
		Labels:             patch.Labels.Replace.Value,
	}
	return issue, nil
}

// beadEdit is updateBead's edit members, checked as far as they can be without the row.
type beadEdit struct {
	nativeType       string
	hasType          bool
	properties       map[string]any
	hasProperties    bool
	propertiesChange []writePatchItem
	hasPropChange    bool
	metadataRaw      json.RawMessage
	metadata         map[string]any
	hasMetadata      bool
	metadataChange   []writePatchItem
	hasMetaChange    bool
}

// beadEdit reads the edit members. Within a plane replacement and patch exclude each
// other, even when empty, and guard or provenance alone is not an update.
func (c *writeCall) beadEdit() (*beadEdit, *callableProblem) {
	edit := &beadEdit{}
	_, edit.hasProperties = c.members["properties"]
	_, edit.hasPropChange = c.members["propertiesChange"]
	_, edit.hasMetadata = c.members["metadata"]
	_, edit.hasMetaChange = c.members["metadataChange"]
	_, edit.hasType = c.members["type"]
	if !edit.hasType && !edit.hasProperties && !edit.hasPropChange && !edit.hasMetadata && !edit.hasMetaChange {
		return nil, newCallableProblem(bdpMalformedRequest, "an update names at least one edit member; a guard or provenance alone is not an update")
	}
	if edit.hasProperties && edit.hasPropChange {
		return nil, newCallableProblem(bdpMalformedRequest, "`properties` and `propertiesChange` are one plane; send one")
	}
	if edit.hasMetadata && edit.hasMetaChange {
		return nil, newCallableProblem(bdpMalformedRequest, "`metadata` and `metadataChange` are one plane; send one")
	}
	var problem *callableProblem
	if edit.hasType {
		if edit.nativeType, problem = c.beadType(c.members["type"]); problem != nil {
			return nil, problem
		}
	}
	if edit.properties, problem = writeObjectMember(c.members, "properties"); problem != nil {
		return nil, problem
	}
	if edit.metadata, problem = writeObjectMember(c.members, "metadata"); problem != nil {
		return nil, problem
	}
	edit.metadataRaw = c.members["metadata"]
	if edit.hasPropChange {
		if edit.propertiesChange, problem = parseWritePatch(c.members["propertiesChange"], "propertiesChange"); problem != nil {
			return nil, problem
		}
	}
	if edit.hasMetaChange {
		if edit.metadataChange, problem = parseWritePatch(c.members["metadataChange"], "metadataChange"); problem != nil {
			return nil, problem
		}
	}
	if problem := checkAdmissibleNumbers(edit.properties, edit.metadata); problem != nil {
		return nil, problem
	}
	for _, change := range []struct {
		items    []writePatchItem
		document string
	}{{edit.propertiesChange, ""}, {edit.metadataChange, "metadata"}} {
		for _, item := range change.items {
			if pointer, bad := firstInadmissible(item.value, ""); bad {
				diagnostic := writeDiagnostic(item.path, "a patch value has an inadmissible number at `"+pointer+"`")
				diagnostic.Document = change.document
				return nil, writeInvalid(diagnostic)
			}
		}
	}
	// What the policy forbids is the attempt, so it is decided before any of it
	// applies, and a protected edit undone later in the same patch still refuses.
	var diagnostics []callableDiagnostic
	for name := range edit.properties {
		if bad := checkProtectedMember(name); bad != nil {
			diagnostics = append(diagnostics, *bad)
		}
	}
	for _, item := range edit.propertiesChange {
		if len(item.tokens) == 0 {
			diagnostics = append(diagnostics, writeDiagnostic("", "a patch of the properties root would touch service-managed members; name the members"))
			continue
		}
		if bad := checkProtectedMember(item.tokens[0]); bad != nil {
			diagnostics = append(diagnostics, *bad)
		}
	}
	if len(diagnostics) > 0 {
		slices.SortFunc(diagnostics, func(a, b callableDiagnostic) int { return strings.Compare(a.InstanceLocation, b.InstanceLocation) })
		return nil, writeInvalid(diagnostics...)
	}
	return edit, nil
}

// patchFor is the native patch that takes current to this edit's complete result, or
// the refusal of that result. Equal members contribute nothing, so an equal result is
// an empty patch, which the role answers as the no-op it is.
func (edit *beadEdit) patchFor(current *types.Issue) (issueops.IssuePatch, *callableProblem) {
	patch := issueops.IssuePatch{}
	if edit.hasType && edit.nativeType != nativeTypeOf(current) {
		patch.IssueType = issueops.Field[issueops.IssueType]{Set: true, Value: issueops.IssueType(edit.nativeType)}
	}
	if edit.hasProperties || edit.hasPropChange {
		stored, err := coreBeadProperties(current)
		if err != nil {
			return patch, newCallableProblem(bdpUnavailable, "the record could not be read")
		}
		desired := map[string]any{}
		if edit.hasProperties {
			for name, value := range edit.properties {
				desired[name] = value
			}
			// The caller-writable set replaces what is writable. The service keeps
			// its own members and completes the result (Write, "Properties replacement").
			for _, name := range issueServiceManaged {
				if value, ok := stored[name]; ok {
					desired[name] = value
				}
			}
		} else {
			applied, problem := applyWritePatch(stored, edit.propertiesChange, "")
			if problem != nil {
				return patch, problem
			}
			desired = applied
		}
		if problem := diffIssueProperties(&patch, stored, desired, current); problem != nil {
			return patch, problem
		}
	}
	if edit.hasMetadata || edit.hasMetaChange {
		stored := coreMetadataObject(current.Metadata)
		desired := edit.metadata
		raw := edit.metadataRaw
		if edit.hasMetaChange {
			applied, problem := applyWritePatch(stored, edit.metadataChange, "metadata")
			if problem != nil {
				return patch, problem
			}
			encoded, err := marshalCoreJSON(applied)
			if err != nil {
				return patch, newCallableProblem(bdpUnavailable, "the metadata could not be encoded")
			}
			desired, raw = applied, encoded
		}
		if !coreJSONEqual(stored, desired) {
			patch.Metadata.Replace = issueops.Field[json.RawMessage]{Set: true, Value: raw}
		}
	}
	return patch, nil
}

// nativeTypeOf is the classification a row reads as: an empty column is the default.
func nativeTypeOf(issue *types.Issue) string {
	if issue.IssueType == "" {
		return string(types.TypeTask)
	}
	return string(issue.IssueType)
}

// diffIssueProperties turns the difference between stored and desired into native
// fields. A member the native write path has no field for is admitted only unchanged.
func diffIssueProperties(patch *issueops.IssuePatch, stored, desired map[string]any, current *types.Issue) *callableProblem {
	names := make([]string, 0, len(stored)+len(desired))
	for name := range stored {
		names = append(names, name)
	}
	for name := range desired {
		if _, also := stored[name]; !also {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	var diagnostics []callableDiagnostic
	for _, name := range names {
		have, hasStored := stored[name]
		want, hasDesired := desired[name]
		if issueMemberEqual(name, have, hasStored, want, hasDesired) {
			continue
		}
		if bad := setIssueMember(patch, name, want, !hasDesired); bad != nil {
			diagnostics = append(diagnostics, *bad)
		}
	}
	if len(diagnostics) > 0 {
		return writeInvalid(diagnostics...)
	}
	if patch.Status.Set && patch.Status.Value == types.StatusClosed && current.Status != types.StatusClosed {
		return writeInvalid(writeDiagnostic("/status", "a direct edit to `closed` needs `closed_at`, which is service-managed: use the `issue/close` Callable"))
	}
	return nil
}

// issueMemberEqual is Write's no-op comparison for one member, under the native
// spellings of "nothing": storage omits an empty text, an empty label set and a null,
// so each of them equals an absent member, while a priority or estimate of 0 is a value.
func issueMemberEqual(name string, have any, hasStored bool, want any, hasDesired bool) bool {
	empty := func(value any, present bool) bool {
		if !present || value == nil {
			return true
		}
		switch v := value.(type) {
		case string:
			return v == ""
		case []any:
			return len(v) == 0
		}
		return false
	}
	if empty(have, hasStored) && empty(want, hasDesired) {
		return true
	}
	if !hasStored || !hasDesired {
		return false
	}
	if slices.Contains(issueInstants, name) {
		a, errA := instantOf(have)
		b, errB := instantOf(want)
		if errA == "" && errB == "" {
			return a.Equal(b)
		}
	}
	if name == "labels" {
		a, okA := have.([]any)
		b, okB := want.([]any)
		if okA && okB {
			return coreJSONEqual(sortedTexts(a), sortedTexts(b))
		}
	}
	return coreJSONEqual(have, want)
}

func sortedTexts(items []any) []any {
	sorted := slices.Clone(items)
	slices.SortFunc(sorted, func(a, b any) int { return strings.Compare(fmt.Sprint(a), fmt.Sprint(b)) })
	return sorted
}

// coreJSONEqual is RFC 6902 §4.6's equality: object members in any order, and numbers
// by value.
func coreJSONEqual(a, b any) bool {
	switch left := a.(type) {
	case json.Number:
		right, ok := b.(json.Number)
		if !ok {
			return false
		}
		x, okX := new(big.Rat).SetString(left.String())
		y, okY := new(big.Rat).SetString(right.String())
		if !okX || !okY {
			return left == right
		}
		return x.Cmp(y) == 0
	case map[string]any:
		right, ok := b.(map[string]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for name, value := range left {
			other, present := right[name]
			if !present || !coreJSONEqual(value, other) {
				return false
			}
		}
		return true
	case []any:
		right, ok := b.([]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for i := range left {
			if !coreJSONEqual(left[i], right[i]) {
				return false
			}
		}
		return true
	}
	return a == b
}

// writePatchItem is one checked RFC 6902 item.
type writePatchItem struct {
	op     string
	path   string
	tokens []string
	value  any
}

// parseWritePatch reads an ordered patch. Everything wrong with its shape or a pointer's
// syntax is malformed-request, and a move, copy or test is not admitted.
func parseWritePatch(raw json.RawMessage, member string) ([]writePatchItem, *callableProblem) {
	malformed := func(detail string) ([]writePatchItem, *callableProblem) {
		return nil, newCallableProblem(bdpMalformedRequest, "`"+member+"`: "+detail)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || items == nil {
		return malformed("is an array of patch items")
	}
	out := make([]writePatchItem, 0, len(items))
	for i, itemRaw := range items {
		at := "item " + strconv.Itoa(i) + " "
		var members map[string]json.RawMessage
		if err := json.Unmarshal(itemRaw, &members); err != nil || members == nil {
			return malformed(at + "is an object")
		}
		var item writePatchItem
		if err := json.Unmarshal(members["op"], &item.op); err != nil || isJSONNull(members["op"]) {
			return malformed(at + "requires `op`")
		}
		if err := json.Unmarshal(members["path"], &item.path); err != nil || isJSONNull(members["path"]) {
			return malformed(at + "requires `path`")
		}
		valueRaw, hasValue := members["value"]
		switch item.op {
		case "add", "replace":
			if !hasValue {
				return malformed(at + "requires `value`")
			}
			if len(members) != 3 {
				return malformed(at + "has a member Write does not admit")
			}
			value, err := decodeJSONValue(valueRaw)
			if err != nil {
				return malformed(at + "has an unreadable `value`")
			}
			item.value = value
		case "remove":
			if hasValue {
				return malformed(at + "forbids `value`")
			}
			if len(members) != 2 {
				return malformed(at + "has a member Write does not admit")
			}
		default:
			return malformed(at + "has `op` " + strconv.Quote(item.op) + "; Write admits add, replace and remove")
		}
		tokens, ok := parseJSONPointer(item.path)
		if !ok {
			return malformed(at + "has an invalid JSON Pointer")
		}
		item.tokens = tokens
		out = append(out, item)
	}
	return out, nil
}

// decodeJSONValue decodes any JSON value, keeping number literals.
func decodeJSONValue(raw json.RawMessage) (any, error) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

// parseJSONPointer is RFC 6901: empty is the root, anything else opens with `/`, and `~`
// only introduces ~0 or ~1.
func parseJSONPointer(pointer string) ([]string, bool) {
	if pointer == "" {
		return nil, true
	}
	if pointer[0] != '/' {
		return nil, false
	}
	tokens := strings.Split(pointer[1:], "/")
	for i, token := range tokens {
		for j := 0; j < len(token); j++ {
			if token[j] == '~' && (j+1 == len(token) || (token[j+1] != '0' && token[j+1] != '1')) {
				return nil, false
			}
		}
		tokens[i] = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
	}
	return tokens, true
}

// applyWritePatch runs an ordered patch on a copy of document. A missing replace or
// remove target and a nonobject final result are validation-failed.
func applyWritePatch(document map[string]any, items []writePatchItem, plane string) (map[string]any, *callableProblem) {
	var root any = cloneJSONValue(document)
	for _, item := range items {
		next, reason := applyPatchItem(root, item.tokens, item)
		if reason != "" {
			diagnostic := writeDiagnostic(item.path, "`"+item.op+"` at `"+item.path+"`: "+reason)
			diagnostic.Document = plane
			return nil, writeInvalid(diagnostic)
		}
		root = next
	}
	final, ok := root.(map[string]any)
	if !ok {
		diagnostic := writeDiagnostic("", "the result of this patch is not a JSON object")
		diagnostic.Document = plane
		return nil, writeInvalid(diagnostic)
	}
	return final, nil
}

// applyPatchItem applies one item at or below node and answers what stands in its place.
func applyPatchItem(node any, tokens []string, item writePatchItem) (any, string) {
	if len(tokens) == 0 {
		if item.op == "remove" {
			return nil, ""
		}
		return cloneJSONValue(item.value), ""
	}
	token, rest := tokens[0], tokens[1:]
	switch container := node.(type) {
	case map[string]any:
		child, exists := container[token]
		if len(rest) > 0 {
			if !exists {
				return nil, "the target does not exist"
			}
			updated, reason := applyPatchItem(child, rest, item)
			if reason != "" {
				return nil, reason
			}
			container[token] = updated
			return container, ""
		}
		switch item.op {
		case "add":
			container[token] = cloneJSONValue(item.value)
		case "replace":
			if !exists {
				return nil, "the target does not exist"
			}
			container[token] = cloneJSONValue(item.value)
		case "remove":
			if !exists {
				return nil, "the target does not exist"
			}
			delete(container, token)
		}
		return container, ""
	case []any:
		index, append_, ok := patchArrayIndex(token, len(container), item.op == "add" && len(rest) == 0)
		if !ok {
			return nil, "the target does not exist"
		}
		if len(rest) > 0 {
			updated, reason := applyPatchItem(container[index], rest, item)
			if reason != "" {
				return nil, reason
			}
			container[index] = updated
			return container, ""
		}
		switch item.op {
		case "add":
			if append_ {
				return append(container, cloneJSONValue(item.value)), ""
			}
			return slices.Insert(container, index, cloneJSONValue(item.value)), ""
		case "replace":
			container[index] = cloneJSONValue(item.value)
			return container, ""
		default:
			return slices.Delete(container, index, index+1), ""
		}
	}
	return nil, "the target does not exist"
}

// patchArrayIndex reads an array token: the RFC 6901 index grammar, plus `-` and the
// one-past-the-end index for an insert.
func patchArrayIndex(token string, length int, inserting bool) (index int, appending, ok bool) {
	if token == "-" {
		return length, true, inserting
	}
	if token == "" || (len(token) > 1 && token[0] == '0') {
		return 0, false, false
	}
	for _, digit := range token {
		if digit < '0' || digit > '9' {
			return 0, false, false
		}
	}
	index, err := strconv.Atoi(token)
	if err != nil {
		return 0, false, false
	}
	if index < length {
		return index, false, true
	}
	return index, inserting && index == length, inserting && index == length
}

func cloneJSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for name, item := range v {
			out[name] = cloneJSONValue(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneJSONValue(item)
		}
		return out
	}
	return value
}
