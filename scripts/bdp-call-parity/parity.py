#!/usr/bin/env python3
"""Measure BDP Callables parity. See README.md; run.sh is the way in."""

import argparse
import collections
import json
import os
import signal
import subprocess
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import callable_cases  # noqa: E402
import part_a  # noqa: E402
import part_b  # noqa: E402
import recorder  # noqa: E402
from stack import Root  # noqa: E402


def op_of_callable():
    """Callable -> v0 operation id, read off the served document's own route table."""
    return {
        "issue/context": "getContext", "issue/ready": "listReadyWork", "issue/ready-count": "countReadyWork",
        "issue/stats": "getStats", "issue/blocked": "listBlockedIssues", "issue/dep/cycles": "listDependencyCycles",
        "issue/dep/tree": "getDependencyTree", "issue/list": "listIssues", "issue/create": "createIssue",
        "issue/query": "queryIssues", "issue/count": "countIssues", "issue/show": "getIssue",
        "issue/update": "updateIssue", "issue/dep/list": "listRelatedIssues", "issue/comment": "addComment",
        "issue/claim": "claimIssue", "issue/unclaim": "releaseIssue", "issue/heartbeat": "heartbeatIssue",
        "issue/close": "closeIssue", "issue/advance": "advanceMolecule", "issue/reopen": "reopenIssue",
        "issue/metadata/compare-and-set": "compareAndSetMetadata", "issue/purge": "sweepIssues",
        "issue/prune": "sweepIssues", "issue/delete": "deleteIssues", "issue/get-many": "batchGetIssues",
        "issue/reclaim": "reclaimIssues", "issue/create-many": "batchCreateIssues", "issue/batch": "applyBatch",
        "issue/close-many": "batchCloseIssues", "issue/claim-next": "claimNextIssue",
        "issue/dep/edges": "listDependencies", "issue/dep/count": "countDependencyEdges",
        "issue/blocking-annotations": "listBlockingAnnotations", "issue/dep/add": "addDependencies",
        "issue/dep/remove": "removeDependency", "config/list": "listSettings", "config/get": "getSetting",
        "config/set": "setSetting", "config/unset": "unsetSetting", "memory/list": "listMemories",
        "memory/remember": "rememberMemory", "memory/recall": "getMemory", "memory/forget": "forgetMemory",
        "events/list": "listEvents",
    }


def summarize(names, a_results, b_results):
    ops = op_of_callable()
    rows = []
    for name in names:
        mine = [r for r in a_results if r["callable"] == name]
        failed = [r for r in mine if not r["pass"]]
        refusals = sum(1 for r in mine if r.get("v0") and r["v0"]["status"] != 200)
        successes = sum(1 for r in mine if r.get("v0") and r["v0"]["status"] == 200)
        assigned = [r for r in b_results if name in r["callables"]]
        dialing = [r for r in b_results if any(o["op"] == ops[name] for o in r["ops"])]
        verdicts = collections.Counter(r["verdict"] for r in assigned)
        working = [r for r in dialing + assigned if r["verdict"] == "ok" and any(o["op"] == ops[name] for o in r["ops"])]
        rows.append({
            "callable": name, "v0_operation": ops[name],
            "a": {"cases": len(mine), "successes": successes, "refusals": refusals, "failed": len(failed),
                  "parity": not failed, "failures": [f["label"] for f in failed]},
            "b": {"assigned_verdicts": dict(verdicts),
                  "assigned_commands": sorted({" ".join(r["argv"][:2]) for r in assigned}),
                  "commands_dialing_the_operation": sorted({" ".join(r["argv"]) for r in dialing})[:40],
                  "working_command": bool(working),
                  "all_assigned_ok": bool(assigned) and all(r["verdict"] == "ok" for r in assigned)},
        })
    return rows


def write_summary(path, meta, rows, a_results, b_results, wire):
    lines = ["# BDP Callables parity", "",
             f"Commit {meta['commit']}. {meta['seconds']:.0f}s. bd {meta['bd']}.", "",
             f"A: {sum(r['a']['parity'] for r in rows)} of {len(rows)} Callables at full parity "
             f"({sum(r['pass'] for r in a_results)} of {len(a_results)} cases pass).",
             f"B: {sum(r['b']['working_command'] for r in rows)} of {len(rows)} have a command that dials the "
             f"operation and matches direct mode; {sum(r['b']['all_assigned_ok'] for r in rows)} have every variant ok.",
             f"B cases: {dict(collections.Counter(r['verdict'] for r in b_results))}.", "",
             "| Callable | A cases (ok path / refusals / failed) | B variants by verdict |", "| --- | --- | --- |"]
    for r in rows:
        verdicts = ", ".join(f"{k} {v}" for k, v in sorted(r["b"]["assigned_verdicts"].items())) or "no command dials it"
        lines.append(f"| `{r['callable']}` | {r['a']['successes']} / {r['a']['refusals']} / {r['a']['failed']} | {verdicts} |")
    lines += ["", "## Part A failures", ""]
    for r in a_results:
        if not r["pass"]:
            lines.append(f"- `{r['callable']}` {r['label']}: " + "; ".join(p[:240] for p in r["problems"][:3]))
    lines += ["", "## Part B, everything not ok", ""]
    for r in b_results:
        if r["verdict"] != "ok":
            says = r["client_says"] or r["direct_says"] or (r["output_diff"] or r["state_diff"] or [""])[0]
            lines.append(f"- {r['verdict']} ({r['mode']}): `bd {' '.join(r['argv'])}` ({', '.join(r['callables'])}): {says[:240]}")
    lines += ["", "## What the client sent during Part B", "",
              "Recorded by a proxy between `bd connect` and `bd serve`. Each read was asked of v0 at once and compared.", "",
              f"{sum(row['requests'] for row in wire.values())} requests; "
              f"{sum(row['requests'] for key, row in wire.items() if key in recorder.non_call_requests(wire))} were not `POST /bdp/call/NAME`.", "",
              "| Callable | requests | arguments | undeclared | replayed / failed |", "| --- | ---: | --- | --- | --- |"]
    for name, row in wire.items():
        lines.append(f"| `{name}` | {row['requests']} | {' '.join(row['arguments'])} | "
                     f"{' '.join(row['undeclared'])} | {row['replayed']} / {len(row['replay_failures'])} |")
    with open(path, "w") as handle:
        handle.write("\n".join(lines) + "\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--bd", required=True)
    parser.add_argument("--dolt", required=True)
    parser.add_argument("--out", required=True)
    parser.add_argument("--part", choices=("a", "b", "both"), default="both")
    parser.add_argument("--only", help="Part B: run only commands matching this regex")
    parser.add_argument("--keep", action="store_true", help="keep the temporary root, for a postmortem")
    args = parser.parse_args()
    os.makedirs(args.out, exist_ok=True)

    def log(message):
        print(message, flush=True)

    # SIGTERM takes the finally below, so a killed run still reaps its daemons.
    signal.signal(signal.SIGTERM, lambda *_: sys.exit(143))
    started = time.time()
    root = Root(os.path.abspath(args.bd), os.path.abspath(args.dolt))
    a_results, b_results, wire, names = [], [], {}, sorted(callable_cases.ROUTES)
    try:
        if args.part in ("a", "both"):
            a_results, names = part_a.run(root, log)
        if args.part in ("b", "both"):
            b_results, wire = part_b.run(root, log, args.only)
    finally:
        root.reap(keep=args.keep)
        left = root.own_daemons()
        log(f"daemons left under {root.path}: {left or 'none'}")
    here = os.path.dirname(os.path.abspath(__file__))
    meta = {
        "commit": subprocess.run(["git", "rev-parse", "HEAD"], cwd=here, capture_output=True, text=True).stdout.strip(),
        "bd": subprocess.run([args.bd, "version"], capture_output=True, text=True, env=root.env).stdout.strip()[:120],
        "seconds": time.time() - started,
    }
    rows = summarize(names, a_results, b_results)
    with open(os.path.join(args.out, "results.json"), "w") as handle:
        json.dump({"meta": meta, "callables": rows, "client_wire": wire, "part_a": a_results, "part_b": b_results}, handle, indent=1)
    write_summary(os.path.join(args.out, "summary.md"), meta, rows, a_results, b_results, wire)
    log(f"wrote {args.out}/results.json and {args.out}/summary.md")
    # Exit 0: a run that measured anomalies still measured. A harness fault raises, and
    # so does the one thing that is no anomaly: a client request that is not an invocation.
    strays = recorder.non_call_requests(wire)
    if strays:
        sys.exit("bd connect sent requests other than POST /bdp/call/NAME:\n  " + "\n  ".join(strays))


if __name__ == "__main__":
    main()
