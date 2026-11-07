package httpapi

import "encoding/json"

// callableRow is one Callable: its name, the v0 operation it routes to, and
// where each of its parameters goes. Everything the Callables layer knows about
// an operation is on its row.
//
// path, query and body are the v0 operation's own parameters, under v0's own
// names, so a row carries no semantics of its own.
// TestCallableTableMatchesTheDocument compares each row to openapi.v0.yaml in
// both directions, so a parameter added to v0 fails there until it is added
// here. The table is written by hand rather than generated because that test is
// needed either way and a generator would add a tool and a second artifact;
// its first draft was derived from specs/types/issue/callables.md §§6-7 in the
// bdp repository, which is also where idempotent, atomic, reasons and cli come
// from.
type callableRow struct {
	// name is the Callable's name; renaming a Callable is this one string.
	name string
	// op is the v0 operation, the key into routeTable.
	op string
	// batch marks a CallableBatch; a row without it is a CallableQuery.
	batch bool
	// idempotent and atomic are what the declaration promises (Callables §6).
	idempotent, atomic bool

	// path, query and body are the declared parameters, so none of them is the
	// operation's `actor` or `author`: see actorMember.
	path, query, body []string
	// adapterParams are arguments adapt consumes; v0 never sees them.
	adapterParams []string
	// optional names v0-required parameters adapt can do without, so the
	// declaration does not require them.
	optional []string
	// adapt is the row's product adapter, when it has one.
	adapt func(args map[string]json.RawMessage) *callableProblem
	// actorMember is the v0 body member revisionProvenance.actor.id is sent as,
	// which no declaration has (specs/types/issue/callables.md §2, §3). "" on a
	// CallableBatch means v0 records no actor, and one supplied is refused.
	actorMember string
	// actorOptional says v0 runs the operation with no actor at all.
	actorOptional bool
	// defaults are the service defaults of §4: the value of a parameter the
	// invocation omitted. A supplied value, null included, is never replaced.
	defaults map[string]json.RawMessage

	// reasons are the v0 codes this Callable declares as callable-error.
	reasons []string
	// changed reads `changed` off v0's result for a batch. Nil means every
	// success may have committed something, which v0's result cannot rule out:
	// `false` promises no durable effect, and `true` says there may have been one.
	changed func(result map[string]json.RawMessage) bool
	// cli is the declaration's cliMapping, as JSON. It is client data the
	// server only publishes (Callables §10).
	cli string
}

// callableTable is every declared Callable. `health` and `watchEvents` stay
// plain routes.
var callableTable = []callableRow{
	{
		name: "issue/context", op: OpGetContext,
	},
	{
		name: "issue/ready", op: OpListReadyWork,
		query: []string{
			"assignee", "unassigned", "type", "exclude_type", "label", "label_any", "exclude_label",
			"label_pattern", "label_regex", "priority", "parent", "metadata_field",
			"has_metadata_key", "include_ephemeral", "include_deferred", "sort", "limit", "brief",
		},
		cli: `{"flags": [{"names": ["-n"], "parameter": "limit"}, {"names": ["-p"], "parameter": "priority"}, {"names": ["-a"], "parameter": "assignee"}, {"names": ["-u"], "parameter": "unassigned"}, {"names": ["-s"], "parameter": "sort"}, {"names": ["-l"], "parameter": "label"}, {"names": ["-t"], "parameter": "type"}]}`,
	},
	{
		name: "issue/ready-count", op: OpCountReadyWork,
		query: []string{
			"assignee", "unassigned", "type", "exclude_type", "label", "label_any", "exclude_label",
			"label_pattern", "label_regex", "priority", "parent", "metadata_field",
			"has_metadata_key", "include_ephemeral", "include_deferred",
		},
	},
	{
		name: "issue/stats", op: OpGetStats,
		query: []string{"assignee", "skip_blocked"},
	},
	{
		name: "issue/blocked", op: OpListBlockedIssues,
		query: []string{"parent", "label", "label_any", "exclude_label"},
	},
	{
		name: "issue/dep/cycles", op: OpListDependencyCycles,
	},
	{
		name: "issue/dep/tree", op: OpGetDependencyTree,
		query: []string{"root_id", "direction", "max_depth", "status"},
		cli:   `{"flags": [{"names": ["-d"], "parameter": "max_depth"}], "positionals": [{"position": 0, "parameter": "root_id"}]}`,
	},
	{
		name: "issue/list", op: OpListIssues,
		defaults: jsonMembers(`{"sort": "priority"}`),
		query: []string{
			"status", "type", "assignee", "label", "label_any", "exclude_label", "parent", "all",
			"include_templates", "include_gates", "include_infra", "include_ephemeral",
			"created_before", "created_after", "metadata_field", "has_metadata_key", "q",
			"no_assignee", "title", "id", "title_contains", "desc_contains", "notes_contains",
			"external_contains", "external_ref", "spec", "label_pattern", "label_regex",
			"updated_after", "updated_before", "closed_after", "closed_before", "defer_after",
			"defer_before", "due_after", "due_before", "priority", "priority_min", "priority_max",
			"empty_description", "no_labels", "no_parent", "exclude_type", "exclude_status",
			"pinned", "no_pinned", "deferred", "overdue", "mol_type", "wisp_type", "sort", "cursor",
			"limit", "brief",
		},
		cli: `{"flags": [{"names": ["-s"], "parameter": "status"}, {"names": ["-a"], "parameter": "assignee"}, {"names": ["-t"], "parameter": "type"}, {"names": ["-l"], "parameter": "label"}, {"names": ["-n"], "parameter": "limit"}]}`,
	},
	{
		name: "issue/create", op: OpCreateIssue, batch: true,
		defaults:    jsonMembers(`{"issue_type": "task"}`),
		actorMember: "actor",
		body: []string{
			"id", "title", "description", "design", "acceptance_criteria", "notes",
			"issue_type", "status", "priority", "assignee", "owner", "labels", "estimated_minutes",
			"external_ref", "due_at", "defer_until", "sender", "metadata", "ephemeral",
			"no_history", "parent_id", "inherit_labels_from_parent", "dependencies", "await_type",
			"await_id", "timeout", "waiters", "waits_for", "force_id_prefix",
		},
		reasons: []string{"dependency_cycle"},
		cli:     `{"flags": [{"names": ["-t", "--type"], "parameter": "issue_type"}, {"names": ["-p"], "parameter": "priority"}, {"names": ["-s"], "parameter": "status"}, {"names": ["-a"], "parameter": "assignee"}, {"names": ["-d"], "parameter": "description"}, {"names": ["-l"], "parameter": "labels"}, {"names": ["--parent"], "parameter": "parent_id"}, {"names": ["--acceptance"], "parameter": "acceptance_criteria"}], "fromFile": [{"names": ["--body-file"], "parameter": "description"}, {"names": ["--design-file"], "parameter": "design"}], "memberFlags": [{"member": "spawner_id", "parameter": "waits_for"}, {"member": "gate", "names": ["--waits-for-gate"], "parameter": "waits_for"}], "positionals": [{"position": 0, "parameter": "title"}]}`,
	},
	{
		name: "issue/query", op: OpQueryIssues,
		query: []string{"q", "all", "sort", "reverse", "limit"},
		cli:   `{"flags": [{"names": ["-n"], "parameter": "limit"}, {"names": ["-a"], "parameter": "all"}, {"names": ["-r"], "parameter": "reverse"}], "positionals": [{"position": 0, "parameter": "q"}]}`,
	},
	{
		name: "issue/count", op: OpCountIssues,
		query: []string{
			"status", "type", "assignee", "priority", "priority_min", "priority_max", "label",
			"label_any", "title", "id", "title_contains", "desc_contains", "notes_contains",
			"created_after", "created_before", "updated_after", "updated_before", "closed_after",
			"closed_before", "empty_description", "no_assignee", "no_labels", "metadata_field",
			"include_ephemeral", "has_metadata_key", "include_infra", "parent", "no_parent",
			"exclude_type", "exclude_status", "group_by",
		},
		cli: `{"flags": [{"names": ["-s"], "parameter": "status"}, {"names": ["-p"], "parameter": "priority"}, {"names": ["-a"], "parameter": "assignee"}, {"names": ["-t"], "parameter": "type"}, {"names": ["-l"], "parameter": "label"}]}`,
	},
	{
		name: "issue/show", op: OpGetIssue,
		path:  []string{"id"},
		query: []string{"include_comments", "include_dependents", "brief_deps"},
		cli:   `{"positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/update", op: OpUpdateIssue, batch: true,
		actorMember: "actor",
		path:        []string{"id"},
		body: []string{
			"patch", "claim", "expected_version", "expected_status", "expected_assignee",
			"force_close_policy", "force_assignee_transfer", "force_notes_overwrite",
			"allow_template",
		},
		reasons: []string{
			"already_claimed", "dependency_cycle", "dependency_exists", "not_claimable",
			"not_closable", "notes_overwrite_refused", "precondition_failed", "template_read_only",
		},
		changed: changedWhen("changed"),
		cli:     `{"memberFlags": [{"member": "title", "parameter": "patch"}, {"member": "description", "names": ["-d"], "parameter": "patch"}, {"member": "design", "parameter": "patch"}, {"member": "acceptance_criteria", "names": ["--acceptance"], "parameter": "patch"}, {"member": "notes", "parameter": "patch"}, {"member": "append_notes", "parameter": "patch"}, {"member": "priority", "names": ["-p"], "parameter": "patch"}, {"member": "issue_type", "names": ["-t", "--type"], "parameter": "patch"}, {"member": "status", "names": ["-s"], "parameter": "patch"}, {"member": "assignee", "names": ["-a"], "parameter": "patch"}, {"member": "parent_id", "names": ["--parent"], "parameter": "patch"}, {"member": "labels", "names": ["--set-labels"], "parameter": "patch"}, {"member": "add_labels", "names": ["-l", "--add-label"], "parameter": "patch"}, {"member": "remove_labels", "names": ["--remove-label"], "parameter": "patch"}, {"member": "estimated_minutes", "names": ["-e", "--estimate"], "parameter": "patch"}, {"member": "external_ref", "parameter": "patch"}, {"member": "due_at", "names": ["--due"], "parameter": "patch"}, {"member": "defer_until", "names": ["--defer"], "parameter": "patch"}, {"member": "await_id", "parameter": "patch"}], "positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/dep/list", op: OpListRelatedIssues,
		defaults: jsonMembers(`{"direction": "out"}`),
		path:     []string{"id"},
		query:    []string{"direction", "type"},
		cli:      `{"flags": [{"names": ["-t"], "parameter": "type"}], "positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/comment", op: OpAddComment, batch: true,
		actorMember: "author",
		atomic:      true,
		path:        []string{"id"},
		body:        []string{"text"},
		cli:         `{"fromFile": [{"names": ["-f", "--file"], "parameter": "text"}], "positionals": [{"position": 0, "parameter": "id"}, {"position": 1, "parameter": "text"}]}`,
	},
	{
		name: "issue/claim", op: OpClaimIssue, batch: true,
		actorMember: "actor",
		idempotent:  true,
		path:        []string{"id"},
		reasons:     []string{"already_claimed", "not_claimable"},
		changed:     changedUnless("already_claimed"),
		cli:         `{"positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/unclaim", op: OpReleaseIssue, batch: true,
		actorMember: "actor",
		path:        []string{"id"},
		body:        []string{"expected_assignee", "force"},
		reasons:     []string{"already_claimed", "not_releasable", "precondition_failed"},
		changed:     changedWhen("changed"),
		cli:         `{"flags": [{"names": ["--if-assignee"], "parameter": "expected_assignee"}], "positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/heartbeat", op: OpHeartbeatIssue, batch: true,
		actorMember: "actor",
		path:        []string{"id"},
		reasons:     []string{"already_claimed", "lease_not_held"},
		cli:         `{"positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/close", op: OpCloseIssue, batch: true,
		defaults:    jsonMembers(`{"reason": "Closed"}`),
		actorMember: "actor",
		idempotent:  true,
		path:        []string{"id"},
		body:        []string{"reason", "session", "force", "expected_version", "auto_close_molecule"},
		reasons:     []string{"issue_pinned", "not_assignee", "not_closable", "template_read_only"},
		changed:     changedByClose,
		cli:         `{"flags": [{"names": ["-r"], "parameter": "reason"}, {"names": ["-f"], "parameter": "force"}, {"names": ["--if-revision"], "parameter": "expected_version"}], "fromFile": [{"names": ["--reason-file"], "parameter": "reason"}], "positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/advance", op: OpAdvanceMolecule, batch: true,
		actorMember: "actor",
		path:        []string{"id"},
		body:        []string{"auto_claim"},
		changed:     changedWhen("claimed"),
		cli:         `{"positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/reopen", op: OpReopenIssue, batch: true,
		actorMember: "actor",
		idempotent:  true,
		path:        []string{"id"},
		body:        []string{"reason", "expected_version"},
		changed:     changedUnless("already_open"),
		cli:         `{"flags": [{"names": ["-r"], "parameter": "reason"}, {"names": ["--if-revision"], "parameter": "expected_version"}], "positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/metadata/compare-and-set", op: OpCompareAndSetMetadata, batch: true,
		actorMember: "actor",
		path:        []string{"id"},
		body:        []string{"key", "expected", "value"},
		changed:     changedWhen("swapped"),
		cli:         `{"positionals": [{"position": 0, "parameter": "id"}]}`,
	},
	{
		name: "issue/purge", op: OpSweepIssues, batch: true,
		defaults:    jsonMembers(`{"dry_run": true, "tier": "ephemeral", "protect_referenced": false, "protect_live_dependents": true}`),
		actorMember: "actor", actorOptional: true,
		body: []string{
			"tier", "closed_before", "pattern", "protect_referenced", "dry_run",
			"protect_live_dependents", "limit", "retain_ids",
		},
		changed: changedUnlessDryRun("swept"),
	},
	{
		name: "issue/prune", op: OpSweepIssues, batch: true,
		defaults:    jsonMembers(`{"dry_run": true, "tier": "durable", "protect_referenced": true, "protect_live_dependents": false}`),
		actorMember: "actor", actorOptional: true,
		body: []string{
			"tier", "closed_before", "pattern", "protect_referenced", "dry_run",
			"protect_live_dependents", "limit", "retain_ids",
		},
		changed: changedUnlessDryRun("swept"),
	},
	{
		name: "issue/delete", op: OpDeleteIssues, batch: true,
		defaults:    jsonMembers(`{"dry_run": true}`),
		actorMember: "actor", actorOptional: true,
		atomic:  true,
		body:    []string{"ids", "cascade", "force", "dry_run", "expected_version"},
		changed: changedUnlessDryRun("deleted"),
		cli:     `{"flags": [{"names": ["--if-revision"], "parameter": "expected_version"}]}`,
	},
	{
		name: "issue/get-many", op: OpBatchGetIssues,
		body: []string{"ids"},
	},
	{
		name: "issue/reclaim", op: OpReclaimIssues, batch: true,
		defaults:    jsonMembers(`{"older_than_seconds": 600}`),
		actorMember: "actor",
		body: []string{
			"older_than_seconds", "ids", "assignees", "labels", "labels_any",
			"exclude_labels", "any_replica",
		},
		changed: changedWhen("reclaimed"),
	},
	{
		name: "issue/create-many", op: OpBatchCreateIssues, batch: true,
		actorMember: "actor",
		atomic:      true,
		body:        []string{"items"},
	},
	{
		name: "issue/batch", op: OpApplyBatch, batch: true,
		actorMember: "actor",
		atomic:      true,
		body:        []string{"items", "provenance", "force_id_prefix", "skip_per_edge_cycle_check"},
		reasons: []string{
			"already_claimed", "dependency_cycle", "dependency_exists", "issue_pinned",
			"not_assignee", "not_closable", "notes_overwrite_refused", "precondition_failed",
			"template_read_only",
		},
	},
	{
		name: "issue/close-many", op: OpBatchCloseIssues, batch: true,
		actorMember: "actor",
		body:        []string{"items", "session", "force", "auto_close_molecule"},
		changed:     changedByBatchClose,
	},
	{
		name: "issue/claim-next", op: OpClaimNextIssue, batch: true,
		actorMember: "actor",
		atomic:      true,
		query: []string{
			"assignee", "unassigned", "type", "exclude_type", "label", "label_any", "exclude_label",
			"label_pattern", "label_regex", "priority", "parent", "metadata_field",
			"has_metadata_key", "include_ephemeral", "include_deferred", "sort",
		},
		changed: changedWhen("claimed"),
		cli:     `{"flags": [{"names": ["-p"], "parameter": "priority"}, {"names": ["-a"], "parameter": "assignee"}, {"names": ["-u"], "parameter": "unassigned"}, {"names": ["-s"], "parameter": "sort"}, {"names": ["-l"], "parameter": "label"}, {"names": ["-t"], "parameter": "type"}]}`,
	},
	{
		name: "issue/dep/edges", op: OpListDependencies,
		query: []string{"issue_id", "type"},
	},
	{
		name: "issue/dep/count", op: OpCountDependencyEdges,
		query: []string{"issue_id", "direction", "type", "status"},
	},
	{
		name: "issue/blocking-annotations", op: OpListBlockingAnnotations,
		query: []string{"issue_id", "is_blocked"},
	},
	{
		name: "issue/dep/add", op: OpAddDependencies, batch: true,
		actorMember: "actor",
		idempotent:  true, atomic: true,
		body:          []string{"edges"},
		adapterParams: []string{"source", "target", "type"}, adapt: adaptFlatDependency,
		optional: []string{"edges"},
		reasons:  []string{"dependency_cycle", "dependency_exists"},
		// No rule: v0's `added` echoes the request whether or not an edge was new.
		cli: `{"flags": [{"names": ["-t"], "parameter": "type"}, {"names": ["--blocked-by", "--depends-on"], "parameter": "target"}], "positionals": [{"position": 0, "parameter": "source"}, {"position": 1, "parameter": "target"}]}`,
	},
	{
		name: "issue/dep/remove", op: OpRemoveDependency, batch: true,
		actorMember: "actor",
		idempotent:  true,
		body:        []string{"issue_id", "depends_on_id"},
		changed:     changedWhen("removed"),
		cli:         `{"positionals": [{"position": 0, "parameter": "issue_id"}, {"position": 1, "parameter": "depends_on_id"}]}`,
	},
	{
		name: "config/list", op: OpListSettings,
	},
	{
		name: "config/get", op: OpGetSetting,
		path: []string{"key"},
		cli:  `{"positionals": [{"position": 0, "parameter": "key"}]}`,
	},
	{
		name: "config/set", op: OpSetSetting, batch: true,
		path: []string{"key"},
		body: []string{"value"},
		cli:  `{"positionals": [{"position": 0, "parameter": "key"}, {"position": 1, "parameter": "value"}]}`,
	},
	{
		name: "config/unset", op: OpUnsetSetting, batch: true,
		path: []string{"key"},
		cli:  `{"positionals": [{"position": 0, "parameter": "key"}]}`,
	},
	{
		name: "memory/list", op: OpListMemories,
		query: []string{"search"},
		cli:   `{"positionals": [{"position": 0, "parameter": "search"}]}`,
	},
	{
		name: "memory/remember", op: OpRememberMemory, batch: true,
		body: []string{"content", "key"},
		cli:  `{"positionals": [{"position": 0, "parameter": "content"}]}`,
	},
	{
		name: "memory/recall", op: OpGetMemory,
		path: []string{"key"},
		cli:  `{"positionals": [{"position": 0, "parameter": "key"}]}`,
	},
	{
		name: "memory/forget", op: OpForgetMemory, batch: true,
		path: []string{"key"},
		cli:  `{"positionals": [{"position": 0, "parameter": "key"}]}`,
	},
	{
		name: "events/list", op: OpListEvents,
		query:   []string{"since", "limit"},
		reasons: []string{"events_journal_disabled", "events_journal_truncated"},
	},
}

// jsonMembers is a table literal's JSON object, by member. A malformed literal
// panics at init, which is every test.
func jsonMembers(document string) map[string]json.RawMessage {
	var members map[string]json.RawMessage
	if err := json.Unmarshal([]byte(document), &members); err != nil {
		panic("callable table: " + document + ": " + err.Error())
	}
	return members
}

// callableIndex finds a row by name.
var callableIndex = func() map[string]*callableRow {
	index := make(map[string]*callableRow, len(callableTable))
	for i := range callableTable {
		index[callableTable[i].name] = &callableTable[i]
	}
	return index
}()

// CallableRoute is where one Callable lands on v0, for the client's tests to hold
// their own tables to (internal/httpclient/wire): they may not read this package's
// table, and the client may not import this package at all.
type CallableRoute struct {
	Op, Method, Path                    string
	PathParams, QueryParams, BodyParams []string
	// ActorMember is the v0 member revisionProvenance.actor.id is sent as.
	ActorMember string
	// ActorOptional says the operation runs with no actor at all.
	ActorOptional bool
	// Defaults are the service defaults the mount applies on omission, by parameter.
	Defaults map[string]json.RawMessage
}

// CallableRoutes is the table, by Callable name.
func CallableRoutes() map[string]CallableRoute {
	routes := make(map[string]CallableRoute, len(callableTable))
	for i := range callableTable {
		row := &callableTable[i]
		rt, ok := routeForOp(row.op)
		if !ok {
			continue
		}
		routes[row.name] = CallableRoute{
			Op: row.op, Method: rt.method, Path: rt.specPathOf(),
			PathParams: row.path, QueryParams: row.query, BodyParams: row.body,
			ActorMember: row.actorMember, ActorOptional: row.actorOptional, Defaults: row.defaults,
		}
	}
	return routes
}

// memberTrue reports whether a member of a v0 result says something happened:
// present and neither null, false, zero nor empty. v0's bodies are compact, so
// these spellings are the only ones it emits.
func memberTrue(result map[string]json.RawMessage, member string) bool {
	switch string(result[member]) {
	case "", "null", "false", "0", "[]", "{}", `""`:
		return false
	}
	return true
}

func changedWhen(member string) func(map[string]json.RawMessage) bool {
	return func(result map[string]json.RawMessage) bool { return memberTrue(result, member) }
}

func changedUnless(member string) func(map[string]json.RawMessage) bool {
	return func(result map[string]json.RawMessage) bool { return !memberTrue(result, member) }
}

// changedUnlessDryRun is the two erasing operations: a preview commits nothing,
// even though its count says what a real run would remove.
func changedUnlessDryRun(count string) func(map[string]json.RawMessage) bool {
	return func(result map[string]json.RawMessage) bool {
		return !memberTrue(result, "dry_run") && memberTrue(result, count)
	}
}

// changedByClose: a re-close writes nothing of its own, but it can still have
// closed the molecule above the step.
func changedByClose(result map[string]json.RawMessage) bool {
	return !memberTrue(result, "already_closed") || memberTrue(result, "auto_closed_molecule")
}

// changedByBatchClose: `already_closed` is present only on a successful item,
// and an item that is present and false is one this batch closed.
func changedByBatchClose(result map[string]json.RawMessage) bool {
	var outcomes []map[string]json.RawMessage
	if err := json.Unmarshal(result["outcomes"], &outcomes); err != nil {
		return true
	}
	for _, outcome := range outcomes {
		if raw, ok := outcome["already_closed"]; ok && string(raw) == "false" {
			return true
		}
	}
	return false
}
