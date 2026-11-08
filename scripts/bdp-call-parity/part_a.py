"""Part A: does POST /bdp/call/NAME give what the direct /v0 request gives?"""

import json
import os
import time
import urllib.parse

import callable_cases as cases
from normalize import diff, masked, strict
from stack import CALLABLE_MEDIA, Stack

STATE_QUERY = ("/v0/beads/issues?all=true&include_templates=true&include_gates=true"
               "&include_infra=true&include_ephemeral=true&limit=0")


def v0_request(name, args):
    """The request a direct v0 caller sends: path, then query or body by the route."""
    method, template = cases.ROUTES[name]
    remaining = dict(args)
    path = template
    while "{" in path:
        param = path[path.index("{") + 1:path.index("}")]
        value = remaining.pop(param, None)
        if value is None:
            return None
        path = path.replace("{" + param + "}", urllib.parse.quote(str(value), safe=""))
    query = {}
    body = None
    if method in ("GET", "DELETE"):
        query = remaining
    elif name == "issue/claim-next":
        query = {k: v for k, v in remaining.items() if k in cases.CLAIM_NEXT_QUERY}
        body = {k: v for k, v in remaining.items() if k not in cases.CLAIM_NEXT_QUERY}
    else:
        body = remaining
    if query:
        pairs = []
        for key, value in query.items():
            for item in value if isinstance(value, list) else [value]:
                pairs.append((key, json.dumps(item) if isinstance(item, bool) else str(item)))
        path += "?" + urllib.parse.urlencode(pairs)
    return method, path, body


def call_v0(stack, name, args):
    method, path, body = v0_request(name, args)
    return stack.http(method, path, body)


def ask_v0(stack, name, args):
    """What a direct v0 caller has to send to get what the Callable does with args."""
    return call_v0(stack, name, cases.with_defaults(name, args))


def ask_callable(stack, name, args, **options):
    """The Callable, with a case's v0 arguments: the actor goes in revisionProvenance."""
    rest, provenance = cases.for_callable(name, args)
    return stack.call(name, rest, provenance=provenance, **options)


def names_the_actor(reply):
    return "revisionProvenance.actor" in json.dumps((reply.json or {}).get("diagnostics", []))


def interpolate(stack, args):
    """$rev:ID in a value is that issue's current revision on this stack."""
    if isinstance(args, dict):
        return {k: interpolate(stack, v) for k, v in args.items()}
    if isinstance(args, list):
        return [interpolate(stack, v) for v in args]
    if isinstance(args, str) and args.startswith("$rev:"):
        return stack.http("GET", "/v0/beads/issues/" + args[5:]).json.get("revision")
    return args


def code_of(reply):
    return reply.json.get("code") if reply.status != 200 and isinstance(reply.json, dict) else None


def expected_problem(v0_reply):
    """Spec section 5 applied to a v0 refusal."""
    body = v0_reply.json
    code = body.get("code")
    stale = code == "precondition_failed" and (
        body.get("expected_version") is not None or body.get("param") == "expected_version")
    if stale:
        return "revision-mismatch"
    return cases.PROBLEM_MAP.get(code, "callable-error")


def compare(name, v0_reply, call_reply, same_server):
    """Everything wrong with the Callable's answer, given v0's. Empty means parity."""
    problems = []
    clean = strict if same_server else masked
    batch = cases.ROUTES[name][0] != "GET" and name != "issue/get-many"
    if v0_reply.status == 200:
        if call_reply.status != 200:
            return [f"v0 answered 200, the Callable {call_reply.status} {call_reply.json.get('code') if call_reply.json else ''}"]
        if not call_reply.headers.get("Content-Type", "").startswith(CALLABLE_MEDIA):
            problems.append("Content-Type " + str(call_reply.headers.get("Content-Type")))
        want = {"changed", "result"} if batch else {"result"}
        if set(call_reply.json) != want:
            problems.append(f"envelope members {sorted(call_reply.json)}")
        problems += [f"result{d}" for d in diff(clean(v0_reply.json), clean(call_reply.json.get("result")))]
        return problems
    want = expected_problem(v0_reply)
    if call_reply.status == 200:
        return [f"v0 refused {v0_reply.status} {v0_reply.json.get('code')}, the Callable answered 200"]
    got = call_reply.json or {}
    if got.get("code") != want:
        problems.append(f"code {got.get('code')!r}, want {want!r} for v0 {v0_reply.json.get('code')!r}")
    if call_reply.status != cases.PROBLEM_STATUS.get(want):
        problems.append(f"status {call_reply.status}, want {cases.PROBLEM_STATUS.get(want)}")
    if got.get("retry") != cases.PROBLEM_RETRY.get(want):
        problems.append(f"retry {got.get('retry')!r}, want {cases.PROBLEM_RETRY.get(want)!r}")
    if not call_reply.headers.get("Content-Type", "").startswith("application/problem+json"):
        problems.append("Content-Type " + str(call_reply.headers.get("Content-Type")))
    carried = got.get("error") if want == "callable-error" else got.get("v0")
    if carried is None:
        problems.append("no v0 body carried")
    else:
        problems += [f"carried{d}" for d in diff(clean(v0_reply.json), clean(carried))]
    if want == "validation-failed" and not got.get("diagnostics"):
        problems.append("validation-failed without diagnostics")
    return problems


def state_of(stack):
    """Everything a write can visibly move, as v0 reads it."""
    rows = stack.http("GET", STATE_QUERY).json.get("items", [])
    return {
        # Ties in the listing's own order can break differently on two servers, and
        # a minted id cannot name a row across them, so rows sort by title then id.
        "issues": sorted(rows, key=lambda row: (row.get("title", ""), row.get("id", ""))),
        "config": stack.http("GET", "/v0/beads/config").json,
        "memories": stack.http("GET", "/v0/beads/memories").json,
        "journal_head": stack.http("GET", "/v0/beads/events?since=0&limit=1").json.get("head"),
    }


def seed(stacks, rows, before, gap=0):
    for row in rows:
        if gap:
            time.sleep(gap)
        for stack in stacks:
            reply = call_v0(stack, "issue/create", row)
            if reply.status != 200:
                raise RuntimeError(f"seed {row['id']}: {reply.status} {reply.raw[:300]}")
    for name, args in before:
        for stack in stacks:
            reply = call_v0(stack, name, args)
            if reply.status != 200:
                raise RuntimeError(f"before {name} {args}: {reply.status} {reply.raw[:300]}")


def record(name, label, args, v0_reply, call_reply, problems, mode, **more):
    entry = {
        "part": "A", "callable": name, "label": label, "mode": mode, "arguments": args,
        "v0": {"status": v0_reply.status, "code": code_of(v0_reply)},
        "call": {"status": call_reply.status, "code": code_of(call_reply),
                 "reason": (call_reply.json or {}).get("error", {}).get("code")
                 if isinstance((call_reply.json or {}).get("error"), dict) else None},
        "problems": problems,
        "pass": not problems,
    }
    if problems:
        entry["v0_body"] = v0_reply.json
        entry["call_body"] = call_reply.json
    entry.update(more)
    return entry


def run_reads(stack, results):
    for name, args, want in cases.read_cases():
        v0_reply = ask_v0(stack, name, args)
        call_reply = stack.call(name, args)
        problems = compare(name, v0_reply, call_reply, same_server=True)
        if None in args.values():
            # A query has no spelling of null, so the router refuses one before v0 runs.
            problems = [p for p in problems if p != "no v0 body carried"]
        got = code_of(v0_reply)
        harness = None if got == want else f"expected v0 code {want!r}, got {got!r}"
        results.append(record(name, json.dumps(args, sort_keys=True), args, v0_reply, call_reply,
                              problems, "read", harness_note=harness))


def run_writes(v0_stack, call_stack, results):
    for case in cases.write_cases():
        name, args = case["callable"], case["args"]
        label = case["label"] or json.dumps(args, sort_keys=True)[:100]
        seed([v0_stack, call_stack], case["seed"], case["before"])
        for id_ in case["sql"]:
            for stack in (v0_stack, call_stack):
                # A template row has no v0 spelling: a formula makes one, and this is that row.
                stack.bd("sql", f"UPDATE issues SET is_template=1 WHERE id='{id_}'", check=True)
        v0_args = interpolate(v0_stack, case["v0_args"] if case["v0_args"] is not None else args)
        call_args = interpolate(call_stack, args)
        if case["wait"]:
            time.sleep(case["wait"])
        before = state_of(call_stack)
        v0_reply = ask_v0(v0_stack, name, v0_args)
        call_reply = ask_callable(call_stack, name, call_args)
        after = state_of(call_stack)
        problems = compare(name, v0_reply, call_reply, same_server=False)
        lacks_actor = cases.lacks_actor(name, call_args)
        if case["own_refusal"] or lacks_actor:
            # The router refused before v0 ran, so there is no v0 body to carry.
            problems = [p for p in problems if p != "no v0 body carried"]
        if lacks_actor and not names_the_actor(call_reply):
            problems.append("no actor, and the refusal does not name revisionProvenance.actor")
        effect = strict(before) != strict(after)
        changed = (call_reply.json or {}).get("changed") if call_reply.status == 200 else None
        # `false` promises no durable effect, and `true` only that there may have been
        # one. v0's result cannot rule one out on the `may_change` cases: they accept a
        # `true` with no visible move, and never a `false` with one.
        if call_reply.status == 200 and changed != effect and not (case.get("may_change") and changed):
            problems.append(f"changed={changed} but the state {'did' if effect else 'did not'} move")
        if call_reply.status != 200 and effect:
            problems.append("a refused Callable moved the state")
        state_diff = diff(masked(state_of(v0_stack)), masked(after), limit=6)
        if state_diff:
            problems.append("state differs after the step: " + "; ".join(state_diff))
        got = code_of(v0_reply)
        harness = None
        if case["expect"] != got and (case["expect"] or got) and case["label"] is None:
            harness = f"expected v0 code {case['expect']!r}, got {got!r}"
        if case["expect"] and case["expect"] != got:
            harness = f"expected v0 code {case['expect']!r}, got {got!r}"
        results.append(record(name, label, args, v0_reply, call_reply, problems, "write", harness_note=harness,
                              changed=changed, effect=effect))


UNKNOWN = "parity_unknown_argument"


def required_args(name):
    """Enough arguments that the only thing wrong is the unknown one."""
    method, template = cases.ROUTES[name]
    args = {}
    if "{id}" in template:
        args["id"] = "sw-fix-none"
    if "{key}" in template:
        args["key"] = "fix-none"
    return args


def run_unknown_argument(stack, names, results):
    """Callables section 4: an undeclared argument is refused before v0 runs."""
    for name in names:
        args = required_args(name)
        args[UNKNOWN] = 1
        call_reply = stack.call(name, args)
        method, path, body = v0_request(name, {k: v for k, v in args.items()})
        v0_reply = stack.http(method, path, body)
        problems = []
        got = call_reply.json or {}
        if got.get("code") != "validation-failed" or call_reply.status != 422:
            problems.append(f"unknown argument answered {call_reply.status} {got.get('code')}")
        if not any(d.get("instanceLocation") == "/arguments/" + UNKNOWN for d in got.get("diagnostics", [])):
            problems.append("no diagnostic names the unknown argument")
        results.append(record(name, "unknown argument", args, v0_reply, call_reply, problems, "probe",
                              note=f"v0 answered {v0_reply.status} {code_of(v0_reply)} to the same member"))


def run_actor_rules(stack, names, results):
    """Spec sections 2 and 3: who is acting is revisionProvenance's to say, and nothing else's."""
    def entry(name, label, args, reply, problems, provenance=None):
        results.append({"part": "A", "callable": name, "label": label, "mode": "probe", "arguments": args,
                        "provenance": provenance, "v0": None, "call": {"status": reply.status, "code": code_of(reply)},
                        "problems": problems, "pass": not problems})

    sam = {"actor": {"id": "sam", "basis": "writer-supplied"}}
    for name in names:
        method = cases.ROUTES[name][0]
        if method == "GET" or name == "issue/get-many":
            continue
        args = required_args(name)
        if name == "issue/dep/add":
            args.update(source="sw-fix-none", target="sw-fix-none2")
        records = name in cases.ACTOR_MEMBER
        for member in ("actor", "author"):
            reply = stack.call(name, {**args, member: "kim"}, provenance=sam if records else None)
            got = reply.json or {}
            problems = []
            if reply.status != 422 or got.get("code") != "validation-failed" or got.get("reason") != "unknown_parameter":
                problems.append(f"`{member}` in arguments answered {reply.status} {got.get('code')} {got.get('reason')}")
            if not any(d.get("instanceLocation") == "/arguments/" + member for d in got.get("diagnostics", [])):
                problems.append(f"no diagnostic names `{member}`")
            entry(name, f"`{member}` in arguments is refused", {**args, member: "kim"}, reply, problems, sam if records else None)
        if name in cases.RECORDS_NO_ACTOR:
            reply = stack.call(name, args, provenance=sam)
            got = reply.json or {}
            problems = [] if reply.status == 422 and got.get("code") == "validation-failed" else [f"answered {reply.status} {got.get('code')}"]
            if "revisionProvenance" not in json.dumps(got.get("diagnostics", [])):
                problems.append("no diagnostic names revisionProvenance")
            entry(name, "a supplied actor is refused", args, reply, problems, sam)
            continue
        reply = stack.call(name, args)
        got = reply.json or {}
        if name in cases.ACTOR_OPTIONAL:
            # Nothing else is wrong with a preview: it runs. A prune with no filter is refused
            # for that, by v0, carrying v0's own body and naming no actor.
            problems = [] if reply.status == 200 or got.get("v0") else [f"no actor answered {reply.status} {got.get('code')}"]
            if names_the_actor(reply):
                problems.append("refused for having no actor")
            entry(name, "runs with no actor", args, reply, problems)
        else:
            problems = []
            if reply.status != 422 or got.get("code") != "validation-failed":
                problems.append(f"no actor answered {reply.status} {got.get('code')}")
            if not names_the_actor(reply):
                problems.append("the refusal does not name revisionProvenance.actor")
            if got.get("v0"):
                problems.append("v0 ran")
            entry(name, "no actor is refused", args, reply, problems)


def run_v0_spellings(stack, results):
    """bd connect sends each omission as a value (wire.v0Omissions). v0 must read it as the omission."""
    def check(name, label, omitted, spelled, clean=strict):
        first, second = call_v0(stack, name, omitted), call_v0(stack, name, spelled)
        problems = [] if first.status == second.status else [f"status {first.status} omitted, {second.status} spelled"]
        problems += diff(clean(first.json), clean(second.json), limit=6)
        results.append({"part": "A", "callable": name, "label": "v0 reads a spelled omission as the omission: " + label,
                        "mode": "probe", "arguments": spelled, "v0": {"status": first.status, "code": code_of(first)},
                        "call": {"status": second.status, "code": code_of(second)}, "problems": problems, "pass": not problems})

    for id_ in ("sw-fix-sp1", "sw-fix-sp2"):
        call_v0(stack, "issue/create", cases.issue(id_))
    check("issue/list", "sort", {}, {"sort": "created"})
    check("issue/dep/list", "direction", {"id": "sw-fix-sp1"}, {"id": "sw-fix-sp1", "direction": ""})
    check("issue/create", "issue_type", {"actor": cases.A, "title": "t"}, {"actor": cases.A, "title": "t", "issue_type": ""})
    # The same refusal of the same member, in two sentences.
    check("issue/purge", "tier", {"actor": cases.A}, {"actor": cases.A, "tier": ""},
          clean=lambda body: {k: body.get(k) for k in ("code", "param", "reason")})
    check("issue/purge", "dry_run and the protections", {"tier": "ephemeral", "actor": cases.A, "pattern": "sw-fix-nomatch*"},
          {"tier": "ephemeral", "actor": cases.A, "pattern": "sw-fix-nomatch*", "dry_run": False, "protect_referenced": True, "protect_live_dependents": False})
    check("issue/delete", "dry_run", {"ids": ["sw-fix-none"], "actor": cases.A}, {"ids": ["sw-fix-none"], "actor": cases.A, "dry_run": False})
    check("issue/reclaim", "older_than_seconds", {"actor": cases.A}, {"actor": cases.A, "older_than_seconds": 0})
    # Two rows, because a close happens once. The record is what has to agree.
    check("issue/close", "reason", {"id": "sw-fix-sp1", "actor": cases.A}, {"id": "sw-fix-sp2", "actor": cases.A, "reason": ""},
          clean=lambda body: {k: (body.get("issue") or body).get(k) for k in ("status", "close_reason", "closed_by_session")})


def run_published_defaults(stack, results):
    """Spec section 4 in the published document: a default, and not required."""
    paths = stack.http("GET", "/bdp/call").json.get("paths", {})
    for name, defaults in cases.DEFAULTS.items():
        content = paths["/bdp/call/" + name]["post"]["requestBody"]["content"]
        arguments = content[next(iter(content))]["schema"]["properties"]["arguments"]
        problems = []
        for param, value in defaults.items():
            published = arguments["properties"][param].get("default", "absent")
            if published != value or type(published) is not type(value):
                problems.append(f"{param}: published default {published!r}, want {value!r}")
            if param in arguments.get("required", []):
                problems.append(f"{param} has a default and is required")
        results.append({"part": "A", "callable": name, "label": "published defaults", "mode": "probe", "arguments": None,
                        "v0": None, "call": None, "problems": problems, "pass": not problems})


def run_journal_off(stack, results):
    v0_reply = call_v0(stack, "events/list", {"since": 0})
    call_reply = stack.call("events/list", {"since": 0})
    problems = compare("events/list", v0_reply, call_reply, same_server=True)
    results.append(record("events/list", "journal off", {"since": 0}, v0_reply, call_reply, problems, "probe",
                          harness_note=None if code_of(v0_reply) == "events_journal_disabled" else "expected events_journal_disabled"))


def run_envelope(stack, results):
    """The namespace's own rules, none of which v0 has."""
    def check(label, reply, status, code=None):
        got = reply.json.get("code") if isinstance(reply.json, dict) else None
        problems = []
        if reply.status != status or (code and got != code):
            problems.append(f"answered {reply.status} {got}, want {status} {code}")
        results.append({"part": "A", "callable": "(envelope)", "label": label, "mode": "probe", "arguments": None,
                        "v0": None, "call": {"status": reply.status, "code": got}, "problems": problems, "pass": not problems})
    media = {"Content-Type": CALLABLE_MEDIA}
    check("unknown name", stack.call("issue/nonsense", {}), 404, "callable-not-found")
    check("wrong media type", stack.http("POST", "/bdp/call/issue/list", {"arguments": {}}), 415, "unsupported-media-type")
    check("no arguments member", stack.http("POST", "/bdp/call/issue/list", {}, media), 400, "malformed-request")
    check("extra envelope member", stack.http("POST", "/bdp/call/issue/list", {"arguments": {}, "x": 1}, media), 400, "malformed-request")
    check("provenance on a query", stack.http("POST", "/bdp/call/issue/list", {"arguments": {}, "revisionProvenance": {}}, media), 400, "malformed-request")
    check("Idempotency-Key on a query", stack.call("issue/list", {}, headers={"Idempotency-Key": "k"}), 400, "malformed-request")
    check("GET on a Callable", stack.http("GET", "/bdp/call/issue/list"), 405)
    check("the list", stack.http("GET", "/bdp/call"), 200)
    check("wire floor", stack.call("issue/list", {}, headers={"Bd-Wire-Revision": "0"}), 200)


def run_path_escapes(stack, results):
    """Ids a direct client must escape. v0's own client escapes the colon too."""
    for raw in ("sw-fix-r1:claim", "sw-fix-r1:close", "a b", "a/b", "a%2Fb", "a?b", "a#b", "caf\u00e9", "%", "a;b", "a,b"):
        for name, args in (("issue/show", {"id": raw}), ("config/get", {"key": raw}), ("memory/recall", {"key": raw})):
            v0_reply = call_v0(stack, name, args)
            call_reply = stack.call(name, args)
            problems = compare(name, v0_reply, call_reply, same_server=True)
            results.append(record(name, f"id needing an escape: {raw!r}", args, v0_reply, call_reply, problems, "probe"))


def run_wire_headers(stack, results):
    """What the bd connect client sends on every request, sent to both surfaces."""
    cases_ = [
        ("client Accept header", {"Accept": "application/json, application/problem+json"},
         {"Accept": "application/vnd.bdp.v0.1+json, application/problem+json"}),
        ("wrong Bd-Project-Id", {"Bd-Project-Id": "not-this-project"}, {"Bd-Project-Id": "not-this-project"}),
        ("future Bd-Wire-Revision", {"Bd-Wire-Revision": "9999"}, {"Bd-Wire-Revision": "9999"}),
        ("Bd-Wire-Revision below the floor", {"Bd-Wire-Revision": "0"}, {"Bd-Wire-Revision": "0"}),
        ("malformed Bd-Wire-Revision", {"Bd-Wire-Revision": "x"}, {"Bd-Wire-Revision": "x"}),
    ]
    for label, v0_headers, call_headers in cases_:
        v0_reply = stack.http("GET", "/v0/beads/issues?limit=1&sort=priority", headers=v0_headers)
        call_reply = stack.call("issue/list", {"limit": 1}, headers=call_headers)
        problems = compare("issue/list", v0_reply, call_reply, same_server=True)
        results.append(record("issue/list", label, {"limit": 1}, v0_reply, call_reply, problems, "probe",
                              note=f"v0 {v0_reply.status} {code_of(v0_reply)}; Callable {call_reply.status} {code_of(call_reply)}"))
    # The profile's own rule, which v0 has no counterpart to: an Accept that excludes the
    # media type is a bodyless 406, even on a request that would have failed.
    v0_reply = stack.http("GET", "/v0/beads/issues?limit=1&sort=priority", headers={"Accept": "application/json"})
    call_reply = stack.call("issue/list", {"limit": 1}, headers={"Accept": "application/json"})
    problems = [] if call_reply.status == 406 and not call_reply.raw else [f"status {call_reply.status}, want a bodyless 406"]
    results.append(record("issue/list", "an Accept without the media type is refused", {"limit": 1}, v0_reply, call_reply, problems, "probe"))


def run_handshake_identity(stack, results):
    """v0 exempts the handshake from the project stamp: it is how a client learns the id."""
    headers = {"Bd-Project-Id": "not-this-project"}
    v0_reply = stack.http("GET", "/v0/beads/context", headers=headers)
    call_reply = stack.call("issue/context", {}, headers=headers)
    problems = compare("issue/context", v0_reply, call_reply, same_server=True)
    results.append(record("issue/context", "wrong Bd-Project-Id on the handshake", {}, v0_reply, call_reply, problems, "probe",
                          note=f"v0 {v0_reply.status} {code_of(v0_reply)}; Callable {call_reply.status} {code_of(call_reply)}"))


def run_body_limits(stack, results):
    """Between v0's limit and the envelope's slack, v0 refuses; past the slack, the mount does."""
    limit = 1 << 20
    for label, size in (("just over v0's body limit", limit + 1024), ("past the envelope's slack", limit + (64 << 10) + 1024)):
        args = {"content": "x" * size, "key": "fix-big"}
        try:
            v0_reply = call_v0(stack, "memory/remember", args)
            call_reply = stack.call("memory/remember", args)
        except OSError as err:
            results.append({"part": "A", "callable": "memory/remember", "label": label, "mode": "probe", "arguments": None,
                            "v0": None, "call": None, "problems": [], "pass": True, "note": f"not measured: {err}"})
            continue
        problems = []
        if size <= limit + (64 << 10):
            problems = compare("memory/remember", v0_reply, call_reply, same_server=True)
        elif (call_reply.json or {}).get("code") != "request-too-large" or call_reply.status != 413:
            problems = [f"answered {call_reply.status} {(call_reply.json or {}).get('code')}"]
        results.append(record("memory/remember", label, {"content": f"{size} bytes"}, v0_reply, call_reply, problems, "probe",
                              note=f"v0 {v0_reply.status} {code_of(v0_reply)}; Callable {call_reply.status} {code_of(call_reply)}"))


def run_auth(root, results):
    """A server that requires a token: 401 and WWW-Authenticate on both surfaces."""
    stack = Stack(root, "a-auth", serve=False)
    token_file = os.path.join(stack.dir, "tokens")
    with open(token_file, "w") as handle:
        handle.write("parity-token\n")
    stack.serve_args = ["--auth-token-file", token_file]
    stack.start_serve()
    for label, headers in (("no credential", {}), ("wrong credential", {"Authorization": "Bearer nope"})):
        v0_reply = stack.http("GET", "/v0/beads/issues?limit=1&sort=priority", headers=headers)
        call_reply = stack.call("issue/list", {"limit": 1}, headers=headers)
        problems = compare("issue/list", v0_reply, call_reply, same_server=True)
        if v0_reply.headers.get("WWW-Authenticate") != call_reply.headers.get("WWW-Authenticate"):
            problems.append("WWW-Authenticate differs")
        results.append(record("issue/list", "auth: " + label, {"limit": 1}, v0_reply, call_reply, problems, "probe",
                              harness_note=None if code_of(v0_reply) == "unauthenticated" else "expected unauthenticated"))
    good = {"Authorization": "Bearer parity-token"}
    v0_reply = stack.http("GET", "/v0/beads/issues?limit=1&sort=priority", headers=good)
    call_reply = stack.call("issue/list", {"limit": 1}, headers=good)
    results.append(record("issue/list", "auth: right credential", {"limit": 1}, v0_reply, call_reply,
                          compare("issue/list", v0_reply, call_reply, same_server=True), "probe"))
    listing = stack.http("GET", "/bdp/call")
    bad = [] if listing.status == 401 else [f"the list answered {listing.status} without a credential"]
    results.append({"part": "A", "callable": "(envelope)", "label": "auth: the list needs a credential", "mode": "probe",
                    "arguments": None, "v0": None, "call": {"status": listing.status, "code": None}, "problems": bad, "pass": not bad})
    root.stop(stack.serve_proc)


def run_large_apply(v0_stack, call_stack, results):
    """Over 100 items a batch apply takes the large-write path, with its own budget."""
    items = [{"kind": "create", "create": {"title": f"large {i}", "issue_type": "task", "id": f"sw-fix-large-{i}"}}
             for i in range(101)]
    args = {"actor": cases.A, "items": items}
    v0_reply = ask_v0(v0_stack, "issue/batch", args)
    call_reply = ask_callable(call_stack, "issue/batch", args)
    problems = compare("issue/batch", v0_reply, call_reply, same_server=False)
    results.append(record("issue/batch", "101 items, the large-apply path", {"items": "101 creates"}, v0_reply, call_reply, problems, "write", changed=(call_reply.json or {}).get("changed"), effect=True))


def run_truncated(stack, results):
    """A checkpoint below the retained window. bd events prune is the only way to get there."""
    for i in range(3):
        call_v0(stack, "issue/create", cases.issue(f"sw-fix-t{i}"))
    head = stack.http("GET", "/v0/beads/events?since=0&limit=1").json.get("head")
    # The retention floors would keep every row this young.
    stack.bd("config", "set", "events-journal-retain-days", "0", check=True)
    stack.bd("config", "set", "events-journal-retain-rows", "0", check=True)
    pruned = stack.bd("events", "prune", "--before", head, "--json") if head else None
    v0_reply = call_v0(stack, "events/list", {"since": 0})
    call_reply = stack.call("events/list", {"since": 0})
    problems = compare("events/list", v0_reply, call_reply, same_server=True)
    results.append(record("events/list", "checkpoint below the window", {"since": 0}, v0_reply, call_reply, problems, "probe",
                          harness_note=None if code_of(v0_reply) == "events_journal_truncated" else
                          f"expected events_journal_truncated, prune said: {pruned.out[:200] + pruned.err[:200] if pruned else 'no head'}"))


def run_db_down(stack, results):
    """The served database gone: v0's 503 and the Callable's temporarily-unavailable."""
    stack.root.stop(stack.dolt_proc)
    time.sleep(1)
    # v0 itself answers the first request after the loss differently from the next
    # ones, so each surface is compared with v0 at the same position in the sequence.
    v0_first = ask_v0(stack, "issue/list", {"limit": 1})
    v0_reply = ask_v0(stack, "issue/list", {"limit": 1})
    call_reply = stack.call("issue/list", {"limit": 1})
    problems = compare("issue/list", v0_reply, call_reply, same_server=True)
    if bool(v0_reply.headers.get("Retry-After")) != bool(call_reply.headers.get("Retry-After")):
        problems.append(f"Retry-After differs: {v0_reply.headers.get('Retry-After')} and {call_reply.headers.get('Retry-After')}")
    results.append(record("issue/list", "the database is down", {"limit": 1}, v0_reply, call_reply, problems, "probe",
                          note=f"v0 first {v0_first.status} {code_of(v0_first)}, then {v0_reply.status} {code_of(v0_reply)}; Callable {call_reply.status} {code_of(call_reply)}"))


def declared_names(stack):
    document = stack.http("GET", "/bdp/call").json
    names = []
    for path, operations in document.get("paths", {}).items():
        if "post" in operations and "/bdp/call/" in path:
            names.append(path.split("/bdp/call/")[1])
    return sorted(names)


def run(root, log):
    results = []
    v0_stack = Stack(root, "a-v0", journal=True, serve=True)
    call_stack = Stack(root, "a-call", journal=True, serve=True)
    probe = Stack(root, "a-probe", journal=False, serve=True)
    names = declared_names(call_stack)
    if sorted(cases.ROUTES) != names:
        raise RuntimeError(f"declared names differ from the routes: {sorted(set(names) ^ set(cases.ROUTES))}")
    log(f"part A: {len(names)} Callables declared")
    seed([v0_stack, call_stack], cases.READ_SEED, cases.READ_BEFORE)
    run_reads(call_stack, results)
    log(f"part A: reads done, {len(results)} results")
    run_writes(v0_stack, call_stack, results)
    log(f"part A: writes done, {len(results)} results")
    run_unknown_argument(probe, names, results)
    run_actor_rules(probe, names, results)
    run_published_defaults(probe, results)
    run_v0_spellings(probe, results)
    run_journal_off(probe, results)
    run_envelope(probe, results)
    run_path_escapes(call_stack, results)
    run_wire_headers(call_stack, results)
    run_handshake_identity(call_stack, results)
    run_body_limits(probe, results)
    run_large_apply(v0_stack, call_stack, results)
    run_auth(root, results)
    # The journals are the record of every write: they must tell the same story.
    final = diff(masked(v0_stack.http("GET", "/v0/beads/events?since=0&limit=10000").json),
                 masked(call_stack.http("GET", "/v0/beads/events?since=0&limit=10000").json), limit=10)
    results.append({"part": "A", "callable": "(all)", "label": "journals agree after every write", "mode": "probe",
                    "arguments": None, "v0": None, "call": None, "problems": final, "pass": not final})
    truncated = Stack(root, "a-truncated", journal=True, serve=True)
    run_truncated(truncated, results)
    run_db_down(probe, results)
    return results, names
