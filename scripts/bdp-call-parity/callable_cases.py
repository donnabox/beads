"""Part A's cases: every Callable, called the way v0 is called.

ROUTES is written from specs/types/issue/callables.md section 1, independently of
internal/httpapi/callables_table.go, so a routing mistake in the table shows as a
difference rather than being copied into the harness.

Reads run both ways against ONE server, back to back. Writes run in lockstep on
two servers holding the same state: v0 on one, the Callable on the other.
"""

ROUTES = {
    "issue/context": ("GET", "/v0/beads/context"),
    "issue/ready": ("GET", "/v0/beads/ready"),
    "issue/ready-count": ("GET", "/v0/beads/ready:count"),
    "issue/stats": ("GET", "/v0/beads/stats"),
    "issue/blocked": ("GET", "/v0/beads/blocked"),
    "issue/dep/cycles": ("GET", "/v0/beads/dependencies/cycles"),
    "issue/dep/tree": ("GET", "/v0/beads/dependencies/tree"),
    "issue/list": ("GET", "/v0/beads/issues"),
    "issue/create": ("POST", "/v0/beads/issues"),
    "issue/query": ("GET", "/v0/beads/issues:query"),
    "issue/count": ("GET", "/v0/beads/issues:count"),
    "issue/show": ("GET", "/v0/beads/issues/{id}"),
    "issue/update": ("PATCH", "/v0/beads/issues/{id}"),
    "issue/dep/list": ("GET", "/v0/beads/issues/{id}/related"),
    "issue/comment": ("POST", "/v0/beads/issues/{id}/comments"),
    "issue/claim": ("POST", "/v0/beads/issues/{id}:claim"),
    "issue/unclaim": ("POST", "/v0/beads/issues/{id}:release"),
    "issue/heartbeat": ("POST", "/v0/beads/issues/{id}:heartbeat"),
    "issue/close": ("POST", "/v0/beads/issues/{id}:close"),
    "issue/advance": ("POST", "/v0/beads/issues/{id}:advanceMolecule"),
    "issue/reopen": ("POST", "/v0/beads/issues/{id}:reopen"),
    "issue/metadata/compare-and-set": ("POST", "/v0/beads/issues/{id}:casMetadata"),
    "issue/purge": ("POST", "/v0/beads/issues:sweep"),
    "issue/prune": ("POST", "/v0/beads/issues:sweep"),
    "issue/delete": ("POST", "/v0/beads/issues:delete"),
    "issue/get-many": ("POST", "/v0/beads/issues:batchGet"),
    "issue/reclaim": ("POST", "/v0/beads/issues:reclaim"),
    "issue/create-many": ("POST", "/v0/beads/issues:batchCreate"),
    "issue/batch": ("POST", "/v0/beads/issues:batchApply"),
    "issue/close-many": ("POST", "/v0/beads/issues:batchClose"),
    "issue/claim-next": ("POST", "/v0/beads/issues:claimNext"),
    "issue/dep/edges": ("GET", "/v0/beads/dependencies"),
    "issue/dep/count": ("GET", "/v0/beads/dependencies:count"),
    "issue/blocking-annotations": ("GET", "/v0/beads/dependencies/blocking"),
    "issue/dep/add": ("POST", "/v0/beads/dependencies:add"),
    "issue/dep/remove": ("POST", "/v0/beads/dependencies:remove"),
    "config/list": ("GET", "/v0/beads/config"),
    "config/get": ("GET", "/v0/beads/config/{key}"),
    "config/set": ("PUT", "/v0/beads/config/{key}"),
    "config/unset": ("DELETE", "/v0/beads/config/{key}"),
    "memory/list": ("GET", "/v0/beads/memories"),
    "memory/remember": ("POST", "/v0/beads/memories"),
    "memory/recall": ("GET", "/v0/beads/memories/{key}"),
    "memory/forget": ("DELETE", "/v0/beads/memories/{key}"),
    "events/list": ("GET", "/v0/beads/events"),
}

# The one POST that also has a query: its filter is listReadyWork's vocabulary.
CLAIM_NEXT_QUERY = {
    "assignee", "unassigned", "type", "exclude_type", "label", "label_any", "exclude_label",
    "label_pattern", "label_regex", "priority", "parent", "metadata_field", "has_metadata_key",
    "include_ephemeral", "include_deferred", "sort",
}

# Spec section 5. Everything else is callable-error.
PROBLEM_MAP = {
    "already_exists": "identity-taken",
    "idempotency_expired": "identity-taken",
    "invalid_cursor": "invalid-parameter",
    "not_found": "resource-not-found",
    "internal": "temporarily-unavailable",
    "busy": "temporarily-unavailable",
    "db_unavailable": "temporarily-unavailable",
    "unauthenticated": "unauthenticated",
    "invalid_argument": "validation-failed",
}
PROBLEM_STATUS = {
    "identity-taken": 409, "invalid-parameter": 400, "resource-not-found": 404,
    "temporarily-unavailable": 503, "unauthenticated": 401, "validation-failed": 422,
    "revision-mismatch": 409, "callable-error": 409,
}
PROBLEM_RETRY = {
    "identity-taken": "after-state-change", "invalid-parameter": "never", "resource-not-found": "after-state-change",
    "temporarily-unavailable": "after-delay", "unauthenticated": "after-state-change",
    "validation-failed": "never", "revision-mismatch": "after-state-change",
    "callable-error": "after-state-change",
}

# Spec sections 2 and 3: no declaration has these. The v0 member an invocation's
# revisionProvenance.actor.id is sent as.
ACTOR_MEMBER = {
    "issue/comment": "author",
    **{name: "actor" for name in (
        "issue/create", "issue/update", "issue/claim", "issue/unclaim", "issue/heartbeat", "issue/close",
        "issue/advance", "issue/reopen", "issue/metadata/compare-and-set", "issue/purge", "issue/prune",
        "issue/delete", "issue/reclaim", "issue/create-many", "issue/batch", "issue/close-many",
        "issue/claim-next", "issue/dep/add", "issue/dep/remove")},
}
# Spec section 3: these run with no actor.
ACTOR_OPTIONAL = {"issue/purge", "issue/prune", "issue/delete"}
# Spec section 3: v0 records no actor on these, so one supplied is refused.
RECORDS_NO_ACTOR = ("config/set", "config/unset", "memory/remember", "memory/forget")

# Spec section 4: what the router supplies for an omitted parameter. v0 has none of them.
# Spelled as v0 spells them; types_rule.py gives the Callable's spelling.
DEFAULTS = {
    "issue/list": {"sort": "priority"},
    "issue/create": {"issue_type": "task"},
    "issue/dep/list": {"direction": "out"},
    "issue/close": {"reason": "Closed"},
    "issue/purge": {"dry_run": True, "tier": "ephemeral", "protect_referenced": False, "protect_live_dependents": True},
    "issue/prune": {"dry_run": True, "tier": "durable", "protect_referenced": True, "protect_live_dependents": False},
    "issue/delete": {"dry_run": True},
    "issue/reclaim": {"older_than_seconds": 600},
}


def with_defaults(name, args):
    """The arguments a direct v0 caller has to send to get what the Callable does on omission."""
    return {**DEFAULTS.get(name, {}), **args}


def for_callable(name, args):
    """A case's arguments are v0's. The Callable takes the actor in revisionProvenance."""
    member = ACTOR_MEMBER.get(name)
    if member is None or member not in args:
        return args, None
    rest = {k: v for k, v in args.items() if k != member}
    return rest, {"actor": {"id": args[member], "basis": "writer-supplied"}}


def lacks_actor(name, args):
    """True when the router itself refuses these arguments for naming nobody."""
    member = ACTOR_MEMBER.get(name)
    return member is not None and name not in ACTOR_OPTIONAL and member not in args


A = "parity"
FUTURE = "2099-01-01T00:00:00Z"
PAST = "2001-01-01T00:00:00Z"


def issue(id_, title=None, **fields):
    body = {"actor": A, "id": id_, "title": title or id_, "issue_type": "task", "priority": 2}
    body.update(fields)
    return body


# The state every read sees. Writes never touch it: they make their own issues.
READ_SEED = [
    issue("sw-fix-r1", "alpha one", priority=1, assignee="alice", labels=["red"],
          metadata={"k": "v"}, description="needle in description", notes="needle in notes",
          external_ref="ext-1", due_at="2030-01-01T00:00:00Z"),
    issue("sw-fix-r2", "alpha two", labels=["blue"], dependencies=[{"target_id": "sw-fix-r1", "type": "blocks"}]),
    issue("sw-fix-r3", "beta closed", priority=3),
    issue("sw-fix-r4", "the epic", issue_type="epic", labels=["red", "blue"]),
    issue("sw-fix-r5", "child of epic", parent_id="sw-fix-r4"),
    issue("sw-fix-r6", "claimed by bob", issue_type="bug", priority=0),
    issue("sw-fix-r7", "commented"),
    issue("sw-fix-r8", "deferred", defer_until="2099-01-01T00:00:00Z"),
    issue("sw-fix-r9", "ephemeral", ephemeral=True),
    issue("sw-fix-ra", "related out", dependencies=[{"target_id": "sw-fix-r7", "type": "related"}]),
]
READ_BEFORE = [
    ("issue/close", {"id": "sw-fix-r3", "actor": A, "reason": "done"}),
    ("issue/claim", {"id": "sw-fix-r6", "actor": "bob"}),
    ("issue/comment", {"id": "sw-fix-r7", "author": "alice", "text": "first"}),
    ("config/set", {"key": "fix.read", "value": "v1"}),
    ("memory/remember", {"content": "remembered thing", "key": "fix-read"}),
]

READY_FILTERS = [
    {"assignee": "alice"}, {"unassigned": True}, {"type": "bug"}, {"exclude_type": ["epic"]},
    # Spec section 2: v0 matches nothing for a name it does not know; that name has no
    # Type reference, so the Callable refuses it.
    {"type": "parity-unknown"},
    {"label": ["red"]}, {"label_any": ["red", "blue"]}, {"exclude_label": ["red"]},
    {"label_pattern": "re*"}, {"label_regex": "^r"}, {"priority": 1}, {"parent": "sw-fix-r4"},
    {"metadata_field": ["k=v"]}, {"has_metadata_key": "k"}, {"include_ephemeral": True},
    {"include_deferred": True},
]
LIST_FILTERS = [
    {"status": ["open"]}, {"status": ["open", "closed"]}, {"type": "bug"}, {"assignee": "alice"},
    {"label": ["red"]}, {"label_any": ["red", "blue"]}, {"exclude_label": ["red"]},
    {"parent": "sw-fix-r4"}, {"all": True}, {"include_templates": True}, {"include_gates": True},
    {"include_infra": True}, {"include_ephemeral": True}, {"created_before": FUTURE},
    {"created_after": PAST}, {"metadata_field": ["k=v"]}, {"has_metadata_key": "k"},
    {"q": "alpha"}, {"no_assignee": True}, {"title": "alpha one"}, {"id": ["sw-fix-r1", "sw-fix-r2"]},
    {"title_contains": "alpha"}, {"desc_contains": "needle"}, {"notes_contains": "needle"},
    {"external_contains": "ext"}, {"external_ref": "ext-1"}, {"spec": "x"}, {"label_pattern": "re*"},
    {"label_regex": "^r"}, {"updated_after": PAST}, {"updated_before": FUTURE}, {"closed_after": PAST},
    {"closed_before": FUTURE}, {"defer_after": PAST}, {"defer_before": FUTURE}, {"due_after": PAST},
    {"due_before": FUTURE}, {"priority": 1}, {"priority_min": 0}, {"priority_max": 1},
    {"empty_description": True}, {"no_labels": True}, {"no_parent": True}, {"exclude_type": ["epic"]},
    {"exclude_status": ["closed"]}, {"pinned": True}, {"no_pinned": True}, {"deferred": True},
    {"overdue": True}, {"mol_type": "swarm"}, {"wisp_type": "patrol"}, {"sort": "priority"},
    {"sort": "created"}, {"limit": 2}, {"brief": True},
]
COUNT_FILTERS = [
    {"status": "open"}, {"type": "bug"}, {"assignee": "alice"}, {"priority": 1}, {"priority_min": 0},
    {"priority_max": 1}, {"label": ["red"]}, {"label_any": ["red", "blue"]}, {"title": "alpha one"},
    {"id": "sw-fix-r1"}, {"title_contains": "alpha"}, {"desc_contains": "needle"},
    {"notes_contains": "needle"}, {"created_after": PAST}, {"created_before": FUTURE},
    {"updated_after": PAST}, {"updated_before": FUTURE}, {"closed_after": PAST}, {"closed_before": FUTURE},
    {"empty_description": True}, {"no_assignee": True}, {"no_labels": True}, {"metadata_field": ["k=v"]},
    {"include_ephemeral": True}, {"has_metadata_key": "k"}, {"include_infra": True},
    {"parent": "sw-fix-r4"}, {"no_parent": True}, {"exclude_type": ["epic"]},
    {"exclude_status": ["closed"]}, {"group_by": "status"}, {"group_by": "priority"},
    {"group_by": "type"}, {"group_by": "assignee"}, {"group_by": "label"},
]


def read_cases():
    """(callable, arguments, the v0 code the refusal should carry, or None for success)."""
    cases = [("issue/context", {}, None)]
    cases += [("issue/ready", {}, None), ("issue/ready", {"limit": 2}, None), ("issue/ready", {"brief": True}, None),
              ("issue/ready", {"sort": "priority"}, None), ("issue/ready", {"sort": "oldest"}, None),
              ("issue/ready", {"sort": "hybrid"}, None)]
    cases += [("issue/ready", f, None) for f in READY_FILTERS]
    cases += [("issue/ready", {"sort": "nonsense"}, "invalid_argument"),
              ("issue/ready", {"priority": 99}, None),
              ("issue/ready", {"limit": -1}, "invalid_argument"),
              ("issue/ready", {"limit": "many"}, "invalid_argument")]
    cases += [("issue/ready-count", {}, None)] + [("issue/ready-count", f, None) for f in READY_FILTERS]
    cases += [("issue/ready-count", {"priority": 99}, None)]
    cases += [("issue/stats", {}, None), ("issue/stats", {"assignee": "alice"}, None),
              ("issue/stats", {"skip_blocked": True}, None), ("issue/stats", {"skip_blocked": "maybe"}, "invalid_argument")]
    cases += [("issue/blocked", {}, None), ("issue/blocked", {"parent": "sw-fix-r4"}, None),
              ("issue/blocked", {"label": ["blue"]}, None), ("issue/blocked", {"label_any": ["blue", "red"]}, None),
              ("issue/blocked", {"exclude_label": ["blue"]}, None)]
    cases += [("issue/dep/cycles", {}, None)]
    cases += [("issue/dep/tree", {"root_id": "sw-fix-r2"}, None),
              ("issue/dep/tree", {"root_id": "sw-fix-r1", "direction": "up"}, None),
              ("issue/dep/tree", {"root_id": "sw-fix-r2", "direction": "down"}, None),
              ("issue/dep/tree", {"root_id": "sw-fix-r2", "direction": "both"}, None),
              ("issue/dep/tree", {"root_id": "sw-fix-r2", "max_depth": 1}, None),
              ("issue/dep/tree", {"root_id": "sw-fix-r2", "status": "open"}, None),
              ("issue/dep/tree", {"root_id": "sw-fix-none"}, "not_found"),
              ("issue/dep/tree", {}, "invalid_argument"),
              ("issue/dep/tree", {"root_id": "sw-fix-r2", "direction": "sideways"}, "invalid_argument")]
    cases += [("issue/list", {}, None)] + [("issue/list", f, None) for f in LIST_FILTERS]
    cases += [("issue/list", {"limit": 0}, None), ("issue/list", {"cursor": "not-a-cursor"}, "invalid_cursor"),
              ("issue/list", {"sort": None}, "invalid_argument"),
              ("issue/list", {"sort": "nonsense"}, "invalid_argument"),
              ("issue/list", {"limit": "many"}, "invalid_argument")]
    cases += [("issue/query", {"q": "status=open"}, None), ("issue/query", {"q": "priority<2", "all": True}, None),
              ("issue/query", {"q": "status=open", "sort": "priority"}, None),
              ("issue/query", {"q": "status=open", "sort": "priority", "reverse": True}, None),
              ("issue/query", {"q": "status=open", "limit": 1}, None),
              ("issue/query", {}, "invalid_argument"), ("issue/query", {"q": "this is = not ( valid"}, "invalid_argument")]
    cases += [("issue/count", {}, None)] + [("issue/count", f, None) for f in COUNT_FILTERS]
    cases += [("issue/count", {"group_by": "nonsense"}, "invalid_argument")]
    cases += [("issue/show", {"id": "sw-fix-r1"}, None), ("issue/show", {"id": "sw-fix-r7", "include_comments": True}, None),
              ("issue/show", {"id": "sw-fix-r1", "include_dependents": True}, None),
              ("issue/show", {"id": "sw-fix-r2", "brief_deps": True}, None),
              ("issue/show", {"id": "sw-fix-none"}, "not_found")]
    cases += [("issue/dep/list", {"id": "sw-fix-r2", "direction": "out"}, None),
              ("issue/dep/list", {"id": "sw-fix-r1", "direction": "in"}, None),
              ("issue/dep/list", {"id": "sw-fix-ra", "direction": "out", "type": ["related"]}, None),
              ("issue/dep/list", {"id": "sw-fix-r2"}, None),
              ("issue/dep/list", {"id": "sw-fix-r2", "direction": None}, "invalid_argument"),
              ("issue/dep/list", {"id": "sw-fix-r2", "direction": ""}, "invalid_argument"),
              ("issue/dep/list", {"id": "sw-fix-none", "direction": "out"}, "not_found")]
    cases += [("issue/get-many", {"ids": ["sw-fix-r1", "sw-fix-r2"]}, None),
              ("issue/get-many", {"ids": ["sw-fix-r1", "sw-fix-none"]}, None),
              ("issue/get-many", {}, "invalid_argument"), ("issue/get-many", {"ids": "sw-fix-r1"}, "invalid_argument")]
    cases += [("issue/dep/edges", {"issue_id": ["sw-fix-r2"]}, None),
              ("issue/dep/edges", {"issue_id": ["sw-fix-r2", "sw-fix-ra"], "type": ["blocks"]}, None),
              ("issue/dep/edges", {}, "invalid_argument")]
    cases += [("issue/dep/count", {"issue_id": ["sw-fix-r2"], "direction": "out"}, None),
              ("issue/dep/count", {"issue_id": ["sw-fix-r1"], "direction": "in", "type": ["blocks"]}, None),
              ("issue/dep/count", {"issue_id": ["sw-fix-r1"], "direction": "in", "status": "open"}, None),
              ("issue/dep/count", {"issue_id": ["sw-fix-r2"]}, "invalid_argument"),
              ("issue/dep/count", {"direction": "out"}, "invalid_argument")]
    cases += [("issue/blocking-annotations", {"issue_id": ["sw-fix-r2"]}, None),
              ("issue/blocking-annotations", {"issue_id": ["sw-fix-r2"], "is_blocked": True}, None),
              ("issue/blocking-annotations", {}, "invalid_argument")]
    cases += [("config/list", {}, None), ("config/get", {"key": "fix.read"}, None),
              ("config/get", {"key": "fix.absent"}, None), ("config/get", {"key": "bad key!"}, None)]
    cases += [("memory/list", {}, None), ("memory/list", {"search": "remembered"}, None),
              ("memory/recall", {"key": "fix-read"}, None), ("memory/recall", {"key": "fix-absent"}, "not_found")]
    cases += [("events/list", {"since": 0}, None), ("events/list", {"since": 0, "limit": 2}, None),
              ("events/list", {}, "invalid_argument"), ("events/list", {"since": -1}, "invalid_argument")]
    return cases


def W(name, args, expect=None, seed=(), before=(), changed=None, v0_args=None, label=None, sql=(), wait=0, own_refusal=False, may_change=False):
    return {"callable": name, "args": args, "expect": expect, "seed": list(seed), "before": list(before), "sql": list(sql), "wait": wait, "own_refusal": own_refusal,
            "changed": changed, "v0_args": v0_args, "label": label, "may_change": may_change}


def write_cases():
    c = []
    # issue/create
    c += [W("issue/create", {"actor": A, "title": "minted id", "issue_type": "task", "priority": 2}, changed=True),
          W("issue/create", {"actor": A, "title": "no type or priority"}, changed=True, label="issue_type omitted is task"),
          W("issue/create", {"actor": A, "title": "typed", "issue_type": "bug"}, changed=True, label="issue_type supplied"),
          W("issue/create", {"actor": A, "title": "null type", "issue_type": None}, expect="invalid_argument", changed=False, label="issue_type null is refused"),
          W("issue/create", {"actor": A, "title": "empty type", "issue_type": ""}, expect="invalid_argument", changed=False, label="an empty issue_type is supplied, and v0 refuses it"),
          W("issue/create", {"actor": A, "title": "no priority", "issue_type": "task"}, changed=True),
          W("issue/create", {"actor": A, "id": "sw-fix-c01", "title": "every field", "description": "d", "design": "ds",
                             "acceptance_criteria": "ac", "notes": "n", "issue_type": "bug", "status": "open", "priority": 1,
                             "assignee": "alice", "owner": "o@example.test", "labels": ["l1", "l2"], "estimated_minutes": 5,
                             "external_ref": "ext-c01", "due_at": "2030-01-01T00:00:00Z", "defer_until": "2031-01-01T00:00:00Z",
                             "sender": "s", "metadata": {"a": 1}}, changed=True),
          W("issue/create", issue("sw-fix-c02", ephemeral=True), changed=True),
          W("issue/create", issue("sw-fix-c03", ephemeral=True, no_history=True), expect="invalid_argument", label="ephemeral with no_history"),
          W("issue/create", issue("sw-fix-c04", no_history=True), changed=True),
          W("issue/create", issue("sw-fix-c05", parent_id="sw-fix-c06", inherit_labels_from_parent=True),
            seed=[issue("sw-fix-c06", labels=["inh"])], changed=True),
          W("issue/create", issue("sw-fix-c07", dependencies=[{"target_id": "sw-fix-c06", "type": "blocks"}]), changed=True),
          W("issue/create", issue("sw-fix-c08", issue_type="gate", await_type="timer", timeout=3600000000000, await_id="", waiters=["w"]), changed=True),
          W("issue/create", issue("sw-fix-c09", waits_for={"spawner_id": "sw-fix-c06", "gate": "all-children"}), changed=True, label="waits_for"),
          W("issue/create", {"actor": A, "title": "wrong prefix", "id": "zz-c10", "issue_type": "task"}, changed=False, expect="invalid_argument"),
          W("issue/create", {"actor": A, "title": "forced prefix", "id": "zz-c11", "issue_type": "task", "force_id_prefix": True}, changed=True),
          W("issue/create", {"actor": A}, expect="invalid_argument", changed=False),
          W("issue/create", {"title": "no actor"}, expect="invalid_argument", changed=False),
          W("issue/create", issue("sw-fix-c12", priority=99), expect="invalid_argument", changed=False),
          W("issue/create", issue("sw-fix-c13", issue_type="nonsense"), expect="invalid_argument", changed=False),
          W("issue/create", issue("sw-fix-c06"), expect="already_exists", changed=False),
          W("issue/create", issue("sw-fix-c14", dependencies=[{"target_id": "sw-fix-none", "type": "blocks"}]), expect="invalid_argument", changed=False, label="missing dependency target"),
          W("issue/create", issue("sw-fix-c15"), seed=[issue("sw-fix-c15")],
            before=[("issue/delete", {"ids": ["sw-fix-c15"], "actor": A, "force": True})], label="recreate a deleted id", changed=None)]
    # issue/update
    c += [W("issue/update", {"id": "sw-fix-u01", "actor": A, "patch": {"title": "new"}}, seed=[issue("sw-fix-u01")], changed=True),
          W("issue/update", {"id": "sw-fix-u02", "actor": A, "patch": {"title": "t", "description": "d", "design": "ds",
             "acceptance_criteria": "ac", "notes": "n", "priority": 0, "issue_type": "bug", "status": "in_progress", "assignee": "z",
             "estimated_minutes": 3, "external_ref": "e-u02", "due_at": "2030-02-02T00:00:00Z", "defer_until": "2031-02-02T00:00:00Z"}},
            seed=[issue("sw-fix-u02")], changed=True),
          W("issue/update", {"id": "sw-fix-u03", "actor": A, "patch": {"add_labels": ["x"], "remove_labels": ["y"]}}, seed=[issue("sw-fix-u03", labels=["y"])], changed=True),
          W("issue/update", {"id": "sw-fix-u04", "actor": A, "patch": {"labels": ["only"]}}, seed=[issue("sw-fix-u04", labels=["y"])], changed=True),
          W("issue/update", {"id": "sw-fix-u05", "actor": A, "patch": {"append_notes": "more"}}, seed=[issue("sw-fix-u05", notes="n")], changed=True),
          W("issue/update", {"id": "sw-fix-u06", "actor": A, "patch": {"metadata": {"set": {"a": 1}, "unset": ["b"]}}}, seed=[issue("sw-fix-u06", metadata={"b": 2})], changed=True),
          W("issue/update", {"id": "sw-fix-u07", "actor": A, "patch": {"parent_id": "sw-fix-u08"}}, seed=[issue("sw-fix-u07"), issue("sw-fix-u08", issue_type="epic")], changed=True),
          W("issue/update", {"id": "sw-fix-u09", "actor": A, "patch": {"due_at": None, "defer_until": None}}, seed=[issue("sw-fix-u09", due_at="2030-01-01T00:00:00Z", defer_until="2030-01-01T00:00:00Z")], changed=True, label="explicit nulls clear"),
          W("issue/update", {"id": "sw-fix-u10", "actor": A, "patch": {}, "claim": True}, seed=[issue("sw-fix-u10")], changed=True),
          W("issue/update", {"id": "sw-fix-u11", "actor": A, "patch": {"title": "sw-fix-u11"}}, seed=[issue("sw-fix-u11")], changed=False, label="no-op patch"),
          W("issue/update", {"id": "sw-fix-u12", "actor": A, "patch": {}}, expect="invalid_argument", seed=[issue("sw-fix-u12")], changed=False, label="empty patch"),
          W("issue/update", {"id": "sw-fix-u13", "actor": A, "patch": {"title": "n"}, "expected_version": "$rev:sw-fix-u13"}, seed=[issue("sw-fix-u13")], changed=True),
          W("issue/update", {"id": "sw-fix-u14", "actor": A, "patch": {"title": "n"}, "expected_version": "1"}, expect="precondition_failed", seed=[issue("sw-fix-u14")], changed=False),
          W("issue/update", {"id": "sw-fix-u15", "actor": A, "patch": {"title": "n"}, "expected_status": "closed"}, expect="precondition_failed", seed=[issue("sw-fix-u15")], changed=False),
          W("issue/update", {"id": "sw-fix-u16", "actor": A, "patch": {"title": "n"}, "expected_assignee": "nobody"}, expect="precondition_failed", seed=[issue("sw-fix-u16")], changed=False),
          W("issue/update", {"id": "sw-fix-u17", "actor": "carol", "patch": {}, "claim": True}, expect="already_claimed", seed=[issue("sw-fix-u17")], before=[("issue/claim", {"id": "sw-fix-u17", "actor": "bob"})], changed=False),
          W("issue/update", {"id": "sw-fix-u18", "actor": A, "patch": {}, "claim": True}, expect="not_claimable", seed=[issue("sw-fix-u18")], before=[("issue/close", {"id": "sw-fix-u18", "actor": A})], changed=False),
          W("issue/update", {"id": "sw-fix-u19", "actor": A, "patch": {"status": "closed"}}, expect="not_closable", seed=[issue("sw-fix-u20"), issue("sw-fix-u19", dependencies=[{"target_id": "sw-fix-u20", "type": "blocks"}])], changed=False),
          W("issue/update", {"id": "sw-fix-u21", "actor": A, "patch": {"status": "closed"}, "force_close_policy": True}, seed=[issue("sw-fix-u22"), issue("sw-fix-u21", dependencies=[{"target_id": "sw-fix-u22", "type": "blocks"}])], changed=True),
          W("issue/update", {"id": "sw-fix-u23", "actor": A, "patch": {"notes": "over"}}, seed=[issue("sw-fix-u23", notes="existing")], label="notes overwrite without force"),
          W("issue/update", {"id": "sw-fix-u24", "actor": A, "patch": {"notes": "over"}, "force_notes_overwrite": True}, seed=[issue("sw-fix-u24", notes="existing")], changed=True),
          W("issue/update", {"id": "sw-fix-u25", "actor": "carol", "patch": {"assignee": "dave"}}, seed=[issue("sw-fix-u25")], before=[("issue/claim", {"id": "sw-fix-u25", "actor": "bob"})], label="assignee transfer without force"),
          W("issue/update", {"id": "sw-fix-u26", "actor": "carol", "patch": {"assignee": "dave"}, "force_assignee_transfer": True}, seed=[issue("sw-fix-u26")], before=[("issue/claim", {"id": "sw-fix-u26", "actor": "bob"})], changed=True),
          W("issue/update", {"id": "sw-fix-u27", "actor": A, "patch": {"title": "t"}}, expect="template_read_only", seed=[issue("sw-fix-u27")], sql=["sw-fix-u27"], changed=False, label="template row, no allow_template"),
          W("issue/update", {"id": "sw-fix-u28", "actor": A, "patch": {"title": "t"}, "allow_template": True}, seed=[issue("sw-fix-u28")], sql=["sw-fix-u28"], changed=True),
          W("issue/update", {"id": "sw-fix-none", "actor": A, "patch": {"title": "n"}}, expect="not_found", changed=False),
          W("issue/update", {"id": "sw-fix-u29", "patch": {"title": "n"}}, expect="invalid_argument", seed=[issue("sw-fix-u29")], changed=False),
          W("issue/update", {"id": "sw-fix-u30", "actor": A}, expect="invalid_argument", seed=[issue("sw-fix-u30")], changed=False),
          W("issue/update", {"id": "sw-fix-u31", "actor": A, "patch": {"nonsense": 1}}, expect="invalid_argument", seed=[issue("sw-fix-u31")], changed=False),
          W("issue/update", {"id": "sw-fix-u32", "actor": A, "patch": {"priority": 99}}, expect="invalid_argument", seed=[issue("sw-fix-u32")], changed=False)]
    # The declared reasons a plain case above does not reach.
    c += [W("issue/update", {"id": "sw-fix-y01", "actor": A, "patch": {"parent_id": "sw-fix-y02"}}, expect="dependency_cycle", seed=[issue("sw-fix-y01"), issue("sw-fix-y02", parent_id="sw-fix-y01")], changed=False),
          W("issue/update", {"id": "sw-fix-y03", "actor": A, "patch": {"parent_id": "sw-fix-y04"}}, expect="dependency_exists", seed=[issue("sw-fix-y04"), issue("sw-fix-y03", dependencies=[{"target_id": "sw-fix-y04", "type": "blocks"}])], changed=False),
          W("issue/batch", {"actor": "carol", "items": [{"kind": "update", "update": {"target": {"id": "sw-fix-y05"}, "patch": {"assignee": "dave"}}}]}, expect="already_claimed", seed=[issue("sw-fix-y05")], before=[("issue/claim", {"id": "sw-fix-y05", "actor": "bob"})], changed=False),
          W("issue/batch", {"actor": A, "items": [{"kind": "dep_add", "dep_add": {"source": {"id": "sw-fix-y06"}, "target": {"id": "sw-fix-y07"}, "type": "related"}}]}, expect="dependency_exists", seed=[issue("sw-fix-y07"), issue("sw-fix-y06", dependencies=[{"target_id": "sw-fix-y07", "type": "blocks"}])], changed=False),
          W("issue/batch", {"actor": A, "items": [{"kind": "close", "close": {"target": {"id": "sw-fix-y08"}}}]}, expect="issue_pinned", seed=[issue("sw-fix-y08", status="pinned")], changed=False),
          W("issue/batch", {"actor": "carol", "items": [{"kind": "close", "close": {"target": {"id": "sw-fix-y09"}}}]}, expect="not_assignee", seed=[issue("sw-fix-y09")], before=[("issue/claim", {"id": "sw-fix-y09", "actor": "bob"})], changed=False),
          W("issue/batch", {"actor": A, "items": [{"kind": "close", "close": {"target": {"id": "sw-fix-y10"}}}]}, expect="not_closable", seed=[issue("sw-fix-y11"), issue("sw-fix-y10", dependencies=[{"target_id": "sw-fix-y11", "type": "blocks"}])], changed=False),
          W("issue/batch", {"actor": A, "items": [{"kind": "update", "update": {"target": {"id": "sw-fix-y12"}, "patch": {"notes": "o"}}}]}, expect="notes_overwrite_refused", seed=[issue("sw-fix-y12", notes="n")], changed=False),
          W("issue/batch", {"actor": A, "items": [{"kind": "close", "close": {"target": {"id": "sw-fix-y13"}}}]}, expect="template_read_only", seed=[issue("sw-fix-y13")], sql=["sw-fix-y13"], changed=False),
          W("issue/purge", {"tier": "ephemeral", "actor": A, "dry_run": False, "pattern": "sw-fix-y14", "retain_ids": True}, seed=[issue("sw-fix-y14", ephemeral=True)], before=[("issue/close", {"id": "sw-fix-y14", "actor": A})], changed=True, label="retain_ids keeps a tombstone"),
          W("issue/create", issue("sw-fix-y14"), expect="idempotency_expired", changed=False, label="a tombstoned id"),
          W("issue/batch", {"actor": A, "items": [{"kind": "create", "create": {"id": "sw-fix-y14", "title": "t", "issue_type": "task"}}]}, expect="idempotency_expired", changed=False, label="a tombstoned id")]
    # issue/comment
    c += [W("issue/comment", {"id": "sw-fix-m01", "author": "alice", "text": "hi"}, seed=[issue("sw-fix-m01")], changed=True),
          W("issue/comment", {"id": "sw-fix-m02", "author": "   ", "text": "blank"}, expect="invalid_argument", seed=[issue("sw-fix-m02")], changed=False, label="a blank name is v0's refusal"),
          W("issue/comment", {"id": "sw-fix-m03", "text": "no author"}, expect="invalid_argument", seed=[issue("sw-fix-m03")], changed=False),
          W("issue/comment", {"id": "sw-fix-m04", "author": "alice"}, expect="invalid_argument", seed=[issue("sw-fix-m04")], changed=False),
          W("issue/comment", {"id": "sw-fix-none", "author": "alice", "text": "x"}, expect="not_found", changed=False)]
    # issue/claim
    c += [W("issue/claim", {"id": "sw-fix-l01", "actor": A}, seed=[issue("sw-fix-l01")], changed=True),
          W("issue/claim", {"id": "sw-fix-l02", "actor": A}, seed=[issue("sw-fix-l02")], before=[("issue/claim", {"id": "sw-fix-l02", "actor": A})], label="reclaim by holder"),
          W("issue/claim", {"id": "sw-fix-l03", "actor": "carol"}, expect="already_claimed", seed=[issue("sw-fix-l03")], before=[("issue/claim", {"id": "sw-fix-l03", "actor": "bob"})], changed=False),
          W("issue/claim", {"id": "sw-fix-l04", "actor": A}, expect="not_claimable", seed=[issue("sw-fix-l04")], before=[("issue/close", {"id": "sw-fix-l04", "actor": A})], changed=False),
          W("issue/claim", {"id": "sw-fix-none", "actor": A}, expect="not_found", changed=False),
          W("issue/claim", {"id": "sw-fix-l05"}, expect="invalid_argument", seed=[issue("sw-fix-l05")], changed=False),
          W("issue/claim", {"id": "sw-fix-l05", "actor": "   "}, expect="invalid_argument", changed=False, label="a blank actor is v0's refusal"),
          W("issue/claim", {"id": "sw-fix-l05", "actor": "a\u0007b"}, expect="invalid_argument", changed=False, label="a control rune is v0's refusal"),
          W("issue/claim", {"id": "sw-fix-l05", "actor": "x" * 300}, expect="invalid_argument", changed=False, label="an over-long actor is v0's refusal"),
          W("issue/claim", {"id": "sw-fix-l05", "actor": "  padded  "}, changed=True, label="v0 trims the actor")]
    # issue/unclaim
    c += [W("issue/unclaim", {"id": "sw-fix-n01", "actor": "bob"}, seed=[issue("sw-fix-n01")], before=[("issue/claim", {"id": "sw-fix-n01", "actor": "bob"})], changed=True),
          W("issue/unclaim", {"id": "sw-fix-n02", "actor": "carol"}, expect="already_claimed", seed=[issue("sw-fix-n02")], before=[("issue/claim", {"id": "sw-fix-n02", "actor": "bob"})], changed=False),
          W("issue/unclaim", {"id": "sw-fix-n03", "actor": "carol", "force": True}, seed=[issue("sw-fix-n03")], before=[("issue/claim", {"id": "sw-fix-n03", "actor": "bob"})], changed=True),
          W("issue/unclaim", {"id": "sw-fix-n04", "actor": "carol", "expected_assignee": "alice"}, expect="precondition_failed", seed=[issue("sw-fix-n04")], before=[("issue/claim", {"id": "sw-fix-n04", "actor": "bob"})], changed=False),
          W("issue/unclaim", {"id": "sw-fix-n05", "actor": "carol", "expected_assignee": "bob"}, seed=[issue("sw-fix-n05")], before=[("issue/claim", {"id": "sw-fix-n05", "actor": "bob"})], changed=True),
          W("issue/unclaim", {"id": "sw-fix-n06", "actor": A}, expect="not_releasable", seed=[issue("sw-fix-n06")], changed=False),
          W("issue/unclaim", {"id": "sw-fix-n07", "actor": A}, expect="not_releasable", seed=[issue("sw-fix-n07")], before=[("issue/close", {"id": "sw-fix-n07", "actor": A})], changed=False),
          W("issue/unclaim", {"id": "sw-fix-none", "actor": A}, expect="not_found", changed=False),
          W("issue/unclaim", {"id": "sw-fix-n08"}, expect="invalid_argument", seed=[issue("sw-fix-n08")], changed=False)]
    # issue/heartbeat
    c += [W("issue/heartbeat", {"id": "sw-fix-h01", "actor": "bob"}, seed=[issue("sw-fix-h01")], before=[("issue/claim", {"id": "sw-fix-h01", "actor": "bob"})], wait=1.2, changed=True, label="holder renews"),
          W("issue/heartbeat", {"id": "sw-fix-h02", "actor": "carol"}, expect="already_claimed", seed=[issue("sw-fix-h02")], before=[("issue/claim", {"id": "sw-fix-h02", "actor": "bob"})], changed=False),
          W("issue/heartbeat", {"id": "sw-fix-h03", "actor": A}, expect="lease_not_held", seed=[issue("sw-fix-h03")], changed=False),
          W("issue/heartbeat", {"id": "sw-fix-none", "actor": A}, expect="not_found", changed=False),
          W("issue/heartbeat", {"id": "sw-fix-h04"}, expect="invalid_argument", seed=[issue("sw-fix-h04")], changed=False)]
    # issue/close
    c += [W("issue/close", {"id": "sw-fix-x01", "actor": A, "reason": "done"}, seed=[issue("sw-fix-x01")], changed=True),
          W("issue/close", {"id": "sw-fix-x02", "actor": A}, seed=[issue("sw-fix-x02")], changed=True, label="reason omitted is Closed"),
          W("issue/close", {"id": "sw-fix-x16", "actor": A, "reason": ""}, seed=[issue("sw-fix-x16")], changed=True, label="an empty reason is supplied, not omitted"),
          W("issue/close", {"id": "sw-fix-x17", "actor": A, "reason": None}, expect="invalid_argument", seed=[issue("sw-fix-x17")], changed=False, label="reason null is refused"),
          W("issue/close", {"id": "sw-fix-x03", "actor": A, "reason": "again"}, seed=[issue("sw-fix-x03")], before=[("issue/close", {"id": "sw-fix-x03", "actor": A})], changed=False, label="re-close"),
          W("issue/close", {"id": "sw-fix-x04", "actor": A, "session": "sess-1"}, seed=[issue("sw-fix-x04")], changed=True),
          W("issue/close", {"id": "sw-fix-x05", "actor": A, "expected_version": "$rev:sw-fix-x05"}, seed=[issue("sw-fix-x05")], changed=True),
          W("issue/close", {"id": "sw-fix-x06", "actor": A, "expected_version": "1"}, expect="precondition_failed", seed=[issue("sw-fix-x06")], changed=False),
          W("issue/close", {"id": "sw-fix-x07", "actor": A}, expect="not_closable", seed=[issue("sw-fix-x08"), issue("sw-fix-x07", dependencies=[{"target_id": "sw-fix-x08", "type": "blocks"}])], changed=False),
          W("issue/close", {"id": "sw-fix-x09", "actor": A, "force": True}, seed=[issue("sw-fix-xa9"), issue("sw-fix-x09", dependencies=[{"target_id": "sw-fix-xa9", "type": "blocks"}])], changed=True),
          W("issue/close", {"id": "sw-fix-x10", "actor": "carol"}, seed=[issue("sw-fix-x10")], before=[("issue/claim", {"id": "sw-fix-x10", "actor": "bob"})], label="close another's claim"),
          W("issue/close", {"id": "sw-fix-x11", "actor": A}, expect="issue_pinned", seed=[issue("sw-fix-x11", status="pinned")], changed=False, label="pinned"),
          W("issue/close", {"id": "sw-fix-x12", "actor": A}, expect="template_read_only", seed=[issue("sw-fix-x12")], sql=["sw-fix-x12"], changed=False),
          W("issue/close", {"id": "sw-fix-x13", "actor": A, "auto_close_molecule": True}, seed=[issue("sw-fix-x14", issue_type="molecule"), issue("sw-fix-x13", parent_id="sw-fix-x14")], changed=True),
          W("issue/close", {"id": "sw-fix-none", "actor": A}, expect="not_found", changed=False),
          W("issue/close", {"id": "sw-fix-x15"}, expect="invalid_argument", seed=[issue("sw-fix-x15")], changed=False)]
    # issue/advance
    mol = [issue("sw-fix-ma0", issue_type="molecule"), issue("sw-fix-ma1", parent_id="sw-fix-ma0"), issue("sw-fix-ma2", parent_id="sw-fix-ma0")]
    mol2 = [issue("sw-fix-mb0", issue_type="molecule"), issue("sw-fix-mb1", parent_id="sw-fix-mb0"), issue("sw-fix-mb2", parent_id="sw-fix-mb0")]
    c += [W("issue/advance", {"id": "sw-fix-ma1", "actor": A}, seed=mol, label="advance a step"),
          W("issue/advance", {"id": "sw-fix-mb1", "actor": A, "auto_claim": True}, seed=mol2, label="advance and claim next"),
          W("issue/advance", {"id": "sw-fix-mc1", "actor": A}, seed=[issue("sw-fix-mc1")], label="no molecule"),
          W("issue/advance", {"id": "sw-fix-none", "actor": A}, expect="not_found", changed=False),
          W("issue/advance", {"id": "sw-fix-mc2"}, expect="invalid_argument", seed=[issue("sw-fix-mc2")], changed=False)]
    # issue/reopen
    c += [W("issue/reopen", {"id": "sw-fix-o01", "actor": A, "reason": "again"}, seed=[issue("sw-fix-o01")], before=[("issue/close", {"id": "sw-fix-o01", "actor": A})], changed=True),
          W("issue/reopen", {"id": "sw-fix-o02", "actor": A}, seed=[issue("sw-fix-o02")], changed=False, label="already open"),
          W("issue/reopen", {"id": "sw-fix-o03", "actor": A, "expected_version": "1"}, expect="precondition_failed", seed=[issue("sw-fix-o03")], before=[("issue/close", {"id": "sw-fix-o03", "actor": A})], changed=False),
          W("issue/reopen", {"id": "sw-fix-none", "actor": A}, expect="not_found", changed=False),
          W("issue/reopen", {"id": "sw-fix-o04"}, expect="invalid_argument", seed=[issue("sw-fix-o04")], changed=False)]
    # issue/metadata/compare-and-set
    name = "issue/metadata/compare-and-set"
    c += [W(name, {"id": "sw-fix-s01", "actor": A, "key": "k", "value": "v"}, seed=[issue("sw-fix-s01")], changed=True, label="set absent key"),
          W(name, {"id": "sw-fix-s02", "actor": A, "key": "k", "expected": "v", "value": "w"}, seed=[issue("sw-fix-s02", metadata={"k": "v"})], changed=True, label="swap"),
          W(name, {"id": "sw-fix-s03", "actor": A, "key": "k", "expected": "other", "value": "w"}, seed=[issue("sw-fix-s03", metadata={"k": "v"})], changed=False, label="lost race"),
          W(name, {"id": "sw-fix-s04", "actor": A, "key": "k", "expected": "v"}, seed=[issue("sw-fix-s04", metadata={"k": "v"})], changed=True, label="delete by omitting value"),
          W(name, {"id": "sw-fix-none", "actor": A, "key": "k", "value": "v"}, expect="not_found", changed=False),
          W(name, {"id": "sw-fix-s05", "actor": A}, expect="invalid_argument", seed=[issue("sw-fix-s05")], changed=False)]
    # issue/purge, issue/prune
    closed_e = [issue("sw-fix-p01", ephemeral=True), issue("sw-fix-p11", ephemeral=True)]
    close_e = [("issue/close", {"id": "sw-fix-p01", "actor": A}), ("issue/close", {"id": "sw-fix-p11", "actor": A})]
    closed_d = [issue("sw-fix-q01"), issue("sw-fix-q11")]
    close_d = [("issue/close", {"id": "sw-fix-q01", "actor": A}), ("issue/close", {"id": "sw-fix-q11", "actor": A})]
    c += [W("issue/purge", {"tier": "ephemeral", "actor": A, "dry_run": True, "pattern": "sw-fix-p*"}, seed=closed_e, before=close_e, changed=False, label="dry run"),
          W("issue/purge", {"tier": "ephemeral", "actor": A, "dry_run": False, "pattern": "sw-fix-p01"}, changed=True, label="real run with pattern"),
          W("issue/purge", {"tier": "ephemeral", "actor": A, "dry_run": False, "pattern": "sw-fix-nomatch*"}, changed=False, label="real run, nothing matches"),
          W("issue/purge", {"tier": "ephemeral", "actor": A, "dry_run": False, "closed_before": PAST}, changed=False, label="cutoff before everything"),
          W("issue/purge", {"tier": "ephemeral", "actor": A, "dry_run": True, "limit": 1, "protect_referenced": False, "protect_live_dependents": True, "retain_ids": True}, changed=False, label="dry run, every knob"),
          W("issue/purge", {"actor": A}, changed=False, label="everything omitted previews the ephemeral tier"),
          W("issue/purge", {}, changed=False, label="no actor either"),
          W("issue/purge", {"tier": "nonsense", "actor": A}, expect="invalid_argument", changed=False),
          W("issue/prune", {"tier": "durable", "actor": A, "dry_run": True, "pattern": "sw-fix-q*"}, seed=closed_d, before=close_d, changed=False, label="dry run"),
          W("issue/prune", {"tier": "durable", "actor": A, "dry_run": False, "pattern": "sw-fix-q01", "protect_referenced": True}, changed=True, label="real run with pattern"),
          W("issue/prune", {"tier": "durable", "actor": A, "dry_run": False, "closed_before": PAST}, changed=False, label="cutoff before everything"),
          W("issue/prune", {"actor": A}, expect="invalid_argument", changed=False, label="tier omitted is durable, which wants a filter"),
          W("issue/prune", {"actor": A, "pattern": "sw-fix-q*"}, changed=False, label="everything else omitted previews"),
          W("issue/prune", {"pattern": "sw-fix-q*"}, changed=False, label="no actor either"),
          W("issue/prune", {"tier": "durable", "actor": A, "dry_run": False, "closed_before": "yesterday"}, expect="invalid_argument", changed=False)]
    held_live = [issue("sw-fix-pa1", ephemeral=True), issue("sw-fix-pa2", ephemeral=True, dependencies=[{"target_id": "sw-fix-pa1", "type": "blocks"}]),
                 issue("sw-fix-pb1", ephemeral=True), issue("sw-fix-pb2", ephemeral=True, dependencies=[{"target_id": "sw-fix-pb1", "type": "blocks"}])]
    close_live = [("issue/close", {"id": "sw-fix-pa1", "actor": A}), ("issue/close", {"id": "sw-fix-pb1", "actor": A})]
    cited = [issue("sw-fix-qa1"), issue("sw-fix-qa2", description="see sw-fix-qa1"), issue("sw-fix-qb1"), issue("sw-fix-qb2", description="see sw-fix-qb1")]
    close_cited = [("issue/close", {"id": "sw-fix-qa1", "actor": A}), ("issue/close", {"id": "sw-fix-qb1", "actor": A})]
    c += [W("issue/purge", {"actor": A, "dry_run": False, "pattern": "sw-fix-pa1"}, seed=held_live, before=close_live, may_change=True, label="protect_live_dependents omitted holds it back"),
          W("issue/purge", {"actor": A, "dry_run": False, "pattern": "sw-fix-pb1", "protect_live_dependents": False}, may_change=True, label="protect_live_dependents supplied"),
          W("issue/purge", {"tier": "durable", "actor": A, "dry_run": True, "pattern": "sw-fix-qa*"}, seed=cited, before=close_cited, changed=False, label="tier supplied to purge"),
          W("issue/purge", {"actor": A, "dry_run": True, "pattern": "sw-fix-p*", "protect_referenced": True}, changed=False, label="protect_referenced supplied"),
          W("issue/prune", {"actor": A, "dry_run": False, "pattern": "sw-fix-qa1"}, may_change=True, label="protect_referenced omitted holds it back"),
          W("issue/prune", {"actor": A, "dry_run": False, "pattern": "sw-fix-qb1", "protect_referenced": False}, may_change=True, label="protect_referenced supplied"),
          W("issue/prune", {"tier": "ephemeral", "actor": A, "dry_run": True, "pattern": "*"}, changed=False, label="tier supplied to prune"),
          W("issue/prune", {"actor": A, "dry_run": True, "pattern": "sw-fix-q*", "protect_live_dependents": True}, changed=False, label="protect_live_dependents supplied"),
          W("issue/purge", {"actor": A, "dry_run": None}, expect="invalid_argument", changed=False, label="null dry_run is refused"),
          W("issue/purge", {"actor": A, "tier": None}, expect="invalid_argument", changed=False, label="null tier is refused"),
          W("issue/purge", {"actor": A, "tier": ""}, expect="invalid_argument", changed=False, label="an empty tier is supplied, and v0 refuses it")]
    # issue/delete
    c += [W("issue/delete", {"ids": ["sw-fix-d01"], "dry_run": True}, seed=[issue("sw-fix-d01")], changed=False, label="dry run"),
          W("issue/delete", {"ids": ["sw-fix-d01"], "actor": A, "dry_run": False}, changed=True, label="real delete"),
          W("issue/delete", {"ids": ["sw-fix-d02"], "actor": A, "dry_run": False}, expect="invalid_argument", seed=[issue("sw-fix-d02"), issue("sw-fix-d03", dependencies=[{"target_id": "sw-fix-d02", "type": "blocks"}])], label="has dependents"),
          W("issue/delete", {"ids": ["sw-fix-d04"], "actor": A, "dry_run": False, "force": True}, seed=[issue("sw-fix-d04"), issue("sw-fix-d05", dependencies=[{"target_id": "sw-fix-d04", "type": "blocks"}])], changed=True, label="force orphans"),
          W("issue/delete", {"ids": ["sw-fix-d06"], "actor": A, "dry_run": False, "cascade": True}, seed=[issue("sw-fix-d06"), issue("sw-fix-d07", dependencies=[{"target_id": "sw-fix-d06", "type": "blocks"}])], changed=True, label="cascade"),
          W("issue/delete", {"ids": ["sw-fix-d08"], "actor": A, "dry_run": False, "expected_version": "$rev:sw-fix-d08"}, seed=[issue("sw-fix-d08")], changed=True),
          W("issue/delete", {"ids": ["sw-fix-d09"], "actor": A, "dry_run": False, "expected_version": "1"}, expect="precondition_failed", seed=[issue("sw-fix-d09")], changed=False),
          W("issue/delete", {"ids": ["sw-fix-none"], "actor": A, "dry_run": False}, expect="not_found", changed=False),
          W("issue/delete", {"actor": A}, expect="invalid_argument", changed=False),
          W("issue/delete", {"ids": [], "actor": A}, expect="invalid_argument", changed=False),
          W("issue/delete", {"ids": ["sw-fix-da1"], "actor": A}, seed=[issue("sw-fix-da1")], changed=False, label="dry_run omitted previews"),
          W("issue/delete", {"ids": ["sw-fix-da1"]}, changed=False, label="no actor either"),
          W("issue/delete", {"ids": ["sw-fix-da1"], "dry_run": False}, changed=True, label="no actor, a real run"),
          W("issue/delete", {"ids": ["sw-fix-da2"], "dry_run": None}, expect="invalid_argument", seed=[issue("sw-fix-da2")], changed=False, label="null dry_run is refused")]
    # issue/reclaim
    c += [W("issue/reclaim", {"actor": A, "older_than_seconds": 86400}, changed=False, label="nothing stale"),
          W("issue/reclaim", {"actor": A, "older_than_seconds": 0, "ids": ["sw-fix-e01"]}, seed=[issue("sw-fix-e01")], before=[("issue/claim", {"id": "sw-fix-e01", "actor": "bob"})], label="zero age, by id"),
          W("issue/reclaim", {"actor": A, "older_than_seconds": 0.001, "ids": ["sw-fix-e02"], "assignees": ["bob"], "labels": ["el"], "labels_any": ["el"], "exclude_labels": ["no"], "any_replica": True}, seed=[issue("sw-fix-e02", labels=["el"])], before=[("issue/claim", {"id": "sw-fix-e02", "actor": "bob"})], label="every filter"),
          W("issue/reclaim", {"older_than_seconds": 60}, expect="invalid_argument", changed=False),
          W("issue/reclaim", {"actor": A}, changed=False, label="older_than_seconds omitted is 600"),
          W("issue/reclaim", {"actor": A, "older_than_seconds": None}, expect="invalid_argument", changed=False, label="null older_than_seconds is refused"),
          W("issue/reclaim", {"actor": A, "older_than_seconds": -1}, expect="invalid_argument", changed=False)]
    # issue/create-many
    c += [W("issue/create-many", {"actor": A, "items": [{"title": "many one", "issue_type": "task", "priority": 2, "labels": ["m"]}, {"title": "many two", "issue_type": "task", "dependencies": [{"target_id": "sw-fix-r1", "type": "related"}]}]}, changed=True),
          W("issue/create-many", {"actor": A, "items": []}, expect="invalid_argument", changed=False, label="empty items"),
          W("issue/create-many", {"actor": A, "items": [{"description": "no title"}]}, expect="invalid_argument", changed=False),
          W("issue/create-many", {"items": [{"title": "t"}]}, expect="invalid_argument", changed=False),
          W("issue/create-many", {"actor": A, "items": [{"title": "t", "issue_type": "task", "dependencies": [{"target_id": "sw-fix-none", "type": "blocks"}]}]}, label="missing dependency target")]
    # issue/batch
    c += [W("issue/batch", {"actor": A, "items": [
              {"kind": "create", "create": {"key": "k1", "title": "b1", "issue_type": "task", "priority": 2}},
              {"kind": "update", "update": {"target": {"id": "sw-fix-b01"}, "patch": {"title": "b-new"}}},
              {"kind": "close", "close": {"target": {"id": "sw-fix-b02"}, "reason": "done"}},
              {"kind": "dep_add", "dep_add": {"source": {"id": "sw-fix-b03"}, "target": {"id": "sw-fix-b01"}, "type": "blocks"}}]},
            seed=[issue("sw-fix-b01"), issue("sw-fix-b02"), issue("sw-fix-b03")], changed=True, label="every kind"),
          W("issue/batch", {"actor": A, "items": [{"kind": "create", "create": {"title": "b2", "issue_type": "task"}}, {"kind": "close", "close": {"target": {"id": "sw-fix-none"}}}]}, expect="not_found", changed=False, label="rolls back on a miss"),
          W("issue/batch", {"actor": A, "items": [{"kind": "nonsense"}]}, expect="invalid_argument", changed=False),
          W("issue/batch", {"actor": A, "items": []}, expect="invalid_argument", changed=False, label="empty items"),
          W("issue/batch", {"items": [{"kind": "create", "create": {"title": "b3"}}]}, expect="invalid_argument", changed=False),
          W("issue/batch", {"actor": A, "items": [{"kind": "update", "update": {"target": {"id": "sw-fix-b04"}, "patch": {"title": "t"}, "expected_version": "1"}}]}, expect="precondition_failed", seed=[issue("sw-fix-b04")], changed=False),
          W("issue/batch", {"actor": A, "items": [{"kind": "dep_add", "dep_add": {"source": {"id": "sw-fix-b05"}, "target": {"id": "sw-fix-b06"}, "type": "blocks"}}]}, expect="dependency_cycle", seed=[issue("sw-fix-b05"), issue("sw-fix-b06", dependencies=[{"target_id": "sw-fix-b05", "type": "blocks"}])], changed=False),
          W("issue/batch", {"actor": A, "provenance": "parity test", "items": [{"kind": "create", "create": {"title": "b4", "issue_type": "task", "id": "sw-fix-b07"}}]}, changed=True, label="provenance"),
          W("issue/batch", {"actor": A, "force_id_prefix": True, "skip_per_edge_cycle_check": True, "items": [{"kind": "create", "create": {"title": "b5", "issue_type": "task", "id": "zy-b08"}}]}, changed=True, label="force prefix")]
    # issue/close-many
    c += [W("issue/close-many", {"actor": A, "items": [{"id": "sw-fix-z01", "reason": "r"}, {"id": "sw-fix-z02"}]}, seed=[issue("sw-fix-z01"), issue("sw-fix-z02")], changed=True),
          W("issue/close-many", {"actor": A, "items": [{"id": "sw-fix-none"}]}, changed=False, label="per-item miss answers 200"),
          W("issue/close-many", {"actor": A, "items": [{"id": "sw-fix-z03"}, {"id": "sw-fix-none2"}]}, seed=[issue("sw-fix-z03")], changed=True, label="one hit, one miss"),
          W("issue/close-many", {"actor": A, "items": [{"id": "sw-fix-z04"}]}, seed=[issue("sw-fix-z04")], before=[("issue/close", {"id": "sw-fix-z04", "actor": A})], changed=False, label="already closed"),
          W("issue/close-many", {"actor": A, "force": True, "session": "s", "auto_close_molecule": True, "items": [{"id": "sw-fix-z05"}]}, seed=[issue("sw-fix-z05")], changed=True, label="every knob"),
          W("issue/close-many", {"actor": A, "items": []}, expect="invalid_argument", changed=False, label="empty items"),
          W("issue/close-many", {"items": [{"id": "sw-fix-z06"}]}, expect="invalid_argument", seed=[issue("sw-fix-z06")], changed=False)]
    # issue/claim-next: narrow by label so the rest of the state never decides the row.
    c += [W("issue/claim-next", {"actor": A, "label": ["next1"]}, seed=[issue("sw-fix-g01", labels=["next1"]), issue("sw-fix-g02", labels=["next1"], priority=0)], changed=True, label="by label"),
          W("issue/claim-next", {"actor": A, "label": ["next1"], "priority": 2, "type": "task", "sort": "priority", "unassigned": True}, changed=True, label="more filters"),
          W("issue/claim-next", {"actor": A, "label": ["nothing-has-this"]}, changed=False, label="nothing ready"),
          W("issue/claim-next", {"label": ["next1"]}, expect="invalid_argument", changed=False),
          W("issue/claim-next", {"actor": A, "sort": "nonsense"}, expect="invalid_argument", changed=False)]
    # issue/dep/add
    c += [W("issue/dep/add", {"actor": A, "edges": [{"issue_id": "sw-fix-f01", "depends_on_id": "sw-fix-f02", "type": "blocks"}]}, seed=[issue("sw-fix-f01"), issue("sw-fix-f02")], changed=True, label="edges form"),
          W("issue/dep/add", {"actor": A, "edges": [{"issue_id": "sw-fix-f01", "depends_on_id": "sw-fix-f02", "type": "blocks"}]}, label="same edge again (declared idempotent)", may_change=True),
          W("issue/dep/add", {"actor": A, "source": "sw-fix-f03", "target": "sw-fix-f02"},
            v0_args={"actor": A, "edges": [{"issue_id": "sw-fix-f03", "depends_on_id": "sw-fix-f02", "type": "blocks"}]}, seed=[issue("sw-fix-f03")], changed=True, label="flat, type omitted"),
          W("issue/dep/add", {"actor": A, "source": "sw-fix-f04", "target": "sw-fix-f02", "type": "related"},
            v0_args={"actor": A, "edges": [{"issue_id": "sw-fix-f04", "depends_on_id": "sw-fix-f02", "type": "related"}]}, seed=[issue("sw-fix-f04")], changed=True, label="flat, typed"),
          # Spec section 2: a kind outside Issue section 2's table is the Scope's own Link Type.
          W("issue/dep/add", {"actor": A, "source": "sw-fix-f05", "target": "sw-fix-f02", "type": "parity-kind"},
            v0_args={"actor": A, "edges": [{"issue_id": "sw-fix-f05", "depends_on_id": "sw-fix-f02", "type": "parity-kind"}]}, seed=[issue("sw-fix-f05")], changed=True, label="flat, a kind of the Scope's own"),
          W("issue/dep/add", {"actor": A, "edges": [{"issue_id": "sw-fix-f02", "depends_on_id": "sw-fix-f01", "type": "blocks"}]}, expect="dependency_cycle", changed=False),
          W("issue/dep/add", {"actor": A, "edges": [{"issue_id": "sw-fix-f01", "depends_on_id": "sw-fix-f02", "type": "related"}]}, expect="dependency_exists", changed=False, label="other type on a held edge"),
          W("issue/dep/add", {"actor": A, "edges": [{"issue_id": "sw-fix-f01", "depends_on_id": "sw-fix-none", "type": "blocks"}]}, expect="invalid_argument", changed=False, label="missing target"),
          W("issue/dep/add", {"actor": A, "edges": []}, expect="invalid_argument", changed=False),
          W("issue/dep/add", {"edges": [{"issue_id": "sw-fix-f01", "depends_on_id": "sw-fix-f02", "type": "blocks"}]}, expect="invalid_argument", changed=False),
          W("issue/dep/add", {"actor": A, "source": "sw-fix-f01"}, expect="invalid_argument",
            v0_args={"actor": A}, changed=False, own_refusal=True, label="flat without target; v0 has no edges")]
    # issue/dep/remove
    c += [W("issue/dep/remove", {"actor": A, "issue_id": "sw-fix-v01", "depends_on_id": "sw-fix-v02"}, seed=[issue("sw-fix-v02"), issue("sw-fix-v01", dependencies=[{"target_id": "sw-fix-v02", "type": "blocks"}])], changed=True),
          W("issue/dep/remove", {"actor": A, "issue_id": "sw-fix-v01", "depends_on_id": "sw-fix-v02"}, changed=False, label="edge already gone"),
          W("issue/dep/remove", {"actor": A, "issue_id": "sw-fix-none", "depends_on_id": "sw-fix-v02"}, changed=False, label="missing issue"),
          W("issue/dep/remove", {"actor": A, "issue_id": "sw-fix-v01"}, expect="invalid_argument", changed=False),
          W("issue/dep/remove", {"issue_id": "sw-fix-v01", "depends_on_id": "sw-fix-v02"}, expect="invalid_argument", changed=False)]
    # config
    c += [W("config/set", {"key": "fix.w1", "value": "v"}, changed=True),
          W("config/set", {"key": "fix.w1", "value": "v"}, label="same value again", may_change=True),
          W("config/set", {"key": "fix.w2"}, expect="invalid_argument", changed=False),
          W("config/set", {"key": "fix key!", "value": "v"}, changed=True, label="key needing an escape"),
          W("config/set", {"key": "issue_prefix", "value": "zz"}, expect="invalid_argument", changed=False, label="a protected key"),
          W("config/unset", {"key": "fix.w1"}, changed=True),
          W("config/unset", {"key": "fix.w1"}, label="already gone", may_change=True),
          W("config/unset", {"key": "fix key!"}, changed=True, label="key needing an escape")]
    # memory
    c += [W("memory/remember", {"content": "keyed", "key": "fix-w1"}, changed=True),
          W("memory/remember", {"content": "keyed", "key": "fix-w1"}, label="same content again", may_change=True),
          W("memory/remember", {"content": "replaced", "key": "fix-w1"}, changed=True),
          W("memory/remember", {"content": "the key is derived from this"}, changed=True),
          W("memory/remember", {"key": "fix-w2"}, expect="invalid_argument", changed=False),
          W("memory/remember", {"content": ""}, expect="invalid_argument", changed=False),
          W("memory/forget", {"key": "fix-w1"}, changed=True),
          W("memory/forget", {"key": "fix-w1"}, expect="not_found", changed=False)]
    return c
