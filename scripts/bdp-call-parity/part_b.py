"""Part B: does the native command work over `bd connect --protocol bdp`, and match direct mode?

Two workspaces in lockstep: `direct` talks to its Dolt server, `client` is attached
with `bd connect --protocol bdp` to a `bd serve --protocol both` on another Dolt server. Seeds and read-backs run
in direct mode on BOTH sides, on the served workspace for the client side, so neither
can be shaped by the path under test.
"""

import json
import os
import re
import time

import native_cases as cases
from normalize import Masker, diff, unordered
from recorder import Recorder, declared_parameters, summarize
from stack import Stack, run_bd

OP_LINE = re.compile(r"event=request .*? op=(\S+) method=(\S+) path=(\S+) status=(\d+) code=\"?([^\" ]*)\"?")


class Pair:
    def __init__(self, root, name, journal=True):
        self.root = root
        self.direct = Stack(root, name + "-direct", journal=journal)
        self.served = Stack(root, name + "-http", journal=journal, serve=True)
        self.recorder = Recorder(self.served, declared_parameters(self.served))
        self.client = self.served.connect_client(url=self.recorder.url)
        self.log_at = os.path.getsize(self.served.serve_log)

    def seed(self, rows, before, gap=0):
        """Direct mode on both databases. A seed that fails is a harness fault."""
        expanded = []
        for row in rows:
            # bd create refuses --id with --parent, so a seeded child is two steps.
            if row[0] == "create" and "--parent" in row:
                at = row.index("--parent")
                expanded.append(row[:at] + row[at + 2:])
                expanded.append(["update", row[row.index("--id") + 1], "--parent", row[at + 1]])
            else:
                expanded.append(row)
        for row in expanded + before:
            if gap:
                time.sleep(gap)
            for stack in (self.direct, self.served):
                ran = stack.bd(*row)
                if ran.code != 0 and row[0] == "create" and "already exists" in (ran.out + ran.err):
                    continue
                if ran.code != 0:
                    raise RuntimeError(f"seed {row}: {ran.out} {ran.err}")

    def run(self, case):
        for name, body in case["files"].items():
            for path in (self.direct.work, self.client):
                with open(os.path.join(path, name), "w") as handle:
                    handle.write(body)
        suffix = [] if case["plain"] else ["--json"]
        direct = self.direct.bd(*self.argv(case, self.direct), *suffix, actor=case["actor"], stdin=case["stdin"], timeout=60)
        argv = self.argv(case, self.served) + suffix
        self.log_at = os.path.getsize(self.served.serve_log)
        client = run_bd(self.root.bd, argv, self.client, self.served.env(case["actor"]), False, 60, case["stdin"])
        return direct, client, self.new_ops()

    @staticmethod
    def argv(case, stack):
        """$rev:ID is that issue's current revision on this side's database."""
        out = []
        for word in case["argv"]:
            if word.startswith("$rev:"):
                shown = parse(stack.bd("show", word[5:], "--json").out)
                word = shown[0]["revision"]
            out.append(word)
        return out

    def new_ops(self):
        with open(self.served.serve_log, errors="replace") as handle:
            handle.seek(self.log_at)
            text = handle.read()
        self.log_at += len(text.encode())
        return [{"op": m.group(1), "method": m.group(2), "path": m.group(3), "status": int(m.group(4)), "code": m.group(5)}
                for m in OP_LINE.finditer(text)]

    def state(self):
        """Every row, comment, label, edge and memory, read in direct mode on each side."""
        return side_state(self.direct), side_state(self.served)


def side_state(stack):
    ran = stack.bd("export", "--all", "--include-infra", "--include-memories", "--json", timeout=60)
    rows = {}
    for line in ran.out.splitlines():
        try:
            row = json.loads(line)
        except ValueError:
            continue
        rows[row.get("id") or row.get("key") or line] = row
    config = stack.bd("config", "list", "--json", timeout=60)
    rows["(config)"] = parse(config.out)
    return rows


def state_delta(before, after):
    """What this command moved on one side. Prior drift cannot enter it.

    A list in title order, not a map by id, so that minted ids mask the same way on
    both sides whatever order they were minted in.
    """
    moved = []
    for key in set(before) | set(after):
        if before.get(key) != after.get(key):
            row = after.get(key)
            moved.append(row if row is not None else {"removed": key})
    return sorted(moved, key=lambda row: (str(row.get("title", "")) if isinstance(row, dict) else "", json.dumps(row, sort_keys=True)[:40]))


def parse(text):
    text = text.strip()
    if not text:
        return None
    try:
        return json.loads(text)
    except ValueError:
        return None


def clean(pair, ran, extra):
    masker = Masker(extra)
    payload = parse(ran.out)
    if payload is None:
        return {"text": masker.text(ran.out.strip())}
    return {"json": masker.walk(payload)}


def first_line(ran):
    text = (ran.err.strip() or ran.out.strip())
    payload = parse(ran.out)
    if isinstance(payload, dict) and payload.get("error"):
        text = str(payload["error"])
    lines = [line for line in text.splitlines() if line.strip()]
    return lines[0][:300] if lines else ""


NOT_SERVED = re.compile(r"not supported by the http backend|not expressible on the v0 wire|not available over bd connect")


def classify(direct, client, output_diff, state_diff, order_only=False):
    if client.code == -1:
        return "hung"
    if direct.code == 0 and client.code == 0 and (NOT_SERVED.search(client.err) or NOT_SERVED.search(client.out)):
        # The command printed a refusal, wrote nothing, and exited 0.
        return "refused-exit-zero"
    if direct.code == 0 and client.code != 0:
        return "refused"
    if direct.code != 0 and client.code == 0:
        return "accepted-only-over-connect"
    if direct.code != 0 and client.code != 0:
        return "both-fail" if not output_diff and direct.code == client.code else "both-fail-differently"
    if state_diff:
        return "state-differs"
    if output_diff:
        return "order-differs" if order_only else "output-differs"
    return "ok"


def run_case(pair, case):
    extra = [(pair.direct.work, "<ws>"), (pair.client, "<ws>"), (pair.served.work, "<ws>"),
             (pair.recorder.url, "<server>"), (pair.served.url, "<server>")]
    wait = case.get("wait", 0)
    try:
        pair.seed(case["seed"], case["before"])
    except RuntimeError as err:
        return {"part": "B", "callables": case["callables"], "label": case["label"], "argv": case["argv"], "phase": case["phase"],
                "verdict": "seed-failed", "direct_exit": None, "client_exit": None, "direct_says": str(err)[:400], "client_says": "",
                "output_diff": [], "state_diff": [], "ops": [], "direct_out": "", "client_out": "", "client_err": ""}
    write = case["phase"] != "read"
    if wait:
        time.sleep(wait)
    before = pair.state() if write else None
    direct, client, ops = pair.run(case)
    state_diff = []
    if write:
        after = pair.state()
        moved_direct = state_delta(before[0], after[0])
        moved_client = state_delta(before[1], after[1])
        state_diff = diff(Masker(extra).walk(moved_direct), Masker(extra).walk(moved_client), limit=8)
    left, right = clean(pair, direct, extra), clean(pair, client, extra)
    if direct.code != 0 and client.code != 0:
        # Two refusals are compared by what they say, on either stream.
        left["stderr"] = Masker(extra).text(direct.err.strip())
        right["stderr"] = Masker(extra).text(client.err.strip())
    output_diff = diff(left, right, limit=10)
    order_only = bool(output_diff) and unordered(left) == unordered(right)
    verdict = classify(direct, client, output_diff, state_diff, order_only)
    return {
        "part": "B", "callables": case["callables"], "label": case["label"], "argv": case["argv"], "phase": case["phase"],
        "mode": case["mode"], "verdict": verdict, "direct_exit": direct.code, "client_exit": client.code,
        "direct_says": first_line(direct) if direct.code else "", "client_says": first_line(client) if client.code else "",
        "output_diff": output_diff, "state_diff": state_diff,
        "ops": ops, "direct_out": direct.out[:1500], "client_out": client.out[:1500], "client_err": client.err[:800],
    }


def run(root, log, only=None):
    """only: a regex on the command line; the rest of the cases do not run."""
    wanted = re.compile(only) if only else None

    def selected(group):
        return [c for c in group if not wanted or wanted.search(" ".join(c["argv"]))]

    def text_of(group):
        return [t for t in (cases.as_text(c) for c in group) if t]

    results = []
    pair = Pair(root, "b-read")
    # created_at has one-second precision on the way out, and listings break ties on it:
    # a gap keeps the two sides from straddling a second differently.
    pair.seed(cases.READ_SEED, cases.READ_BEFORE, gap=1.1)
    for case in selected(cases.reads()):
        results.append(run_case(pair, case))
    for case in text_of(selected(cases.reads())):
        results.append(run_case(pair, case))
    log(f"part B: reads done, {len(results)} results")
    entries = list(pair.recorder.entries)
    pair.recorder.stop()
    root.stop(pair.served.serve_proc)
    # Text runs repeat the fixture ids, so they get a database of their own.
    for name, group in (("b-write", selected(cases.writes())), ("b-write-text", text_of(selected(cases.writes())))):
        pair = Pair(root, name)
        for case in group:
            results.append(run_case(pair, case))
        log(f"part B: {name} done, {len(results)} results")
        entries += pair.recorder.entries
        pair.recorder.stop()
        root.stop(pair.served.serve_proc)
    both = selected(cases.global_cases())
    for index, case in enumerate(both + text_of(both)):
        pair = Pair(root, f"b-global-{index}")
        results.append(run_case(pair, case))
        entries += pair.recorder.entries
        pair.recorder.stop()
        root.stop(pair.served.serve_proc)
        root.stop(pair.served.dolt_proc)
        root.stop(pair.direct.dolt_proc)
    log(f"part B: global cases done, {len(results)} results")
    return results, summarize(entries)
