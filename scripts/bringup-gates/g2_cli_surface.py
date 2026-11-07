#!/usr/bin/env python3
"""G2: `bd` stays identical for a v0 client and service, base against head.

Fork-only (.github/workflows/bringup-gates.yml). Builds bd at the merge base of the PR
with bringup-staging and at the PR head, then compares the two three ways:

- the help tree and flag inventory, walked from the built binaries;
- the committed goldens that pin what the binaries do not show (hidden commands and
  flags, the `bd --json` corpus);
- the default wire behaviour, with scripts/bdp-call-parity/wire_proof.py.

A difference fails unless g2_exceptions.json accepts it or a decision record the PR adds
under engdocs/decisions/ names it on a `G2-waives: FINDING-ID` line (a glob is allowed;
this run prints each finding's id). A run that could not compare fails, and no record
waives that.

Exit codes: 0 no unaccepted difference, 1 a difference, 2 could not compare.

    python3 scripts/bringup-gates/g2_cli_surface.py --base-ref bringup-staging --dolt DOLT --out DIR
"""

import argparse
import collections
import concurrent.futures
import difflib
import fnmatch
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import time

HERE = os.path.dirname(os.path.abspath(__file__))
PROOF_DIR = os.path.join(os.path.dirname(HERE), "bdp-call-parity")
sys.path.insert(0, PROOF_DIR)
sys.dont_write_bytecode = True

import native_cases  # noqa: E402
from normalize import diff as json_diff  # noqa: E402

EXCEPTIONS_PATH = "scripts/bringup-gates/g2_exceptions.json"
DECISIONS_DIR = "engdocs/decisions/"
GOLDEN_PATHS = ["cmd/bd/testdata/command_tree/", "cmd/bd/protocol/testdata/corpus/"]
GATE_PATHS = ["scripts/bringup-gates/", "scripts/bdp-call-parity/", ".github/workflows/bringup-gates.yml"]
DOLT_PIN_FILE = "scripts/ci/install-dolt.sh"
WAIVER_LINE = re.compile(r"^G2-waives:\s*(\S.*?)\s*$", re.MULTILINE)
TELEMETRY_ENV = {
    "BD_DISABLE_METRICS": "1",
    "DO_NOT_TRACK": "1",
    "BD_DISABLE_EVENT_FLUSH": "1",
    "DOLT_DISABLE_EVENT_FLUSH": "1",
    "BEADS_DOLT_AUTO_START": "0",
    "BD_NON_INTERACTIVE": "1",
}
DROPPED_ENV = ("BEADS_DIR", "BEADS_DB", "BD_ACTOR", "BEADS_ACTOR", "BEADS_DOLT_SERVER_PORT",
               "BEADS_DOLT_PASSWORD", "BEADS_HTTP_TOKEN", "BEADS_HTTP_TOKEN_COMMAND")
# Below this a walk found too little to mean anything; 307 at 2026-11-06.
MIN_SURFACE_COMMANDS = 100
HEADER = re.compile(r"^([A-Z][A-Za-z &/,-]*):$")
NOT_COMMAND_SECTIONS = {"Usage", "Aliases", "Examples", "Flags", "Global Flags", "Additional help topics"}
COMMAND_LINE = re.compile(r"^  (\S+)(?:\s|$)")
FLAG_LINE = re.compile(r"^\s+(?:-(\S), )?--([\w-]+)(?: (\S+))?(?:\s{2,}(.*))?$")
TREE_HEADING = re.compile(r"^bd(?: [\w-]+)*$", re.MULTILINE)


class CouldNotCompare(Exception):
    pass


def finding(id_, fields, kind=None):
    return {"id": id_, "kind": kind or id_.split(":", 1)[0], "fields": list(fields)}


# --- The help tree and flag inventory, walked from a binary -----------------------------

def parse_subcommands(text):
    """The subcommand names a help text lists, by its sections after the last Usage:."""
    lines = text.splitlines()
    usages = [i for i, line in enumerate(lines) if line == "Usage:"]
    if not usages:
        return []
    start = usages[-1]
    found, section = [], "Usage"
    for line in lines[start:]:
        header = HEADER.match(line)
        if header:
            section = header.group(1)
            continue
        if section in NOT_COMMAND_SECTIONS or not line.strip():
            continue
        match = COMMAND_LINE.match(line)
        if match:
            found.append(match.group(1))
    return found


def parse_flags(text):
    """name -> {scope, shorthand, type, usage}, from the Flags and Global Flags sections."""
    flags, section = {}, None
    for line in text.splitlines():
        header = HEADER.match(line)
        if header:
            section = header.group(1)
            continue
        if section not in ("Flags", "Global Flags") or not line.strip():
            continue
        match = FLAG_LINE.match(line)
        if match:
            short, name, type_, usage = match.groups()
            flags["--" + name] = {"scope": section, "shorthand": short or "", "type": type_ or "", "usage": usage or ""}
    return flags


def help_of(bd, path, cwd, env):
    args = [bd] + path.split()[1:] + ["--help"]
    done = subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True, timeout=60, stdin=subprocess.DEVNULL)
    return {"exit": done.returncode, "stdout": done.stdout, "stderr": done.stderr}


def walk_surface(bd, golden_paths, cwd, env, workers=8):
    """Every command path the binary lists, plus those its tree golden names: the hidden ones."""
    seen, level = {}, ["bd"] + sorted(golden_paths)
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        while level:
            level = [p for p in dict.fromkeys(level) if p not in seen]
            answers = pool.map(lambda p: help_of(bd, p, cwd, env), level)
            following = []
            for path, answer in zip(level, answers):
                seen[path] = answer
                following += [f"{path} {name}" for name in parse_subcommands(answer["stdout"]) if name != "bd"]
            level = following
    return seen


def diff_lines(base, new):
    return [(line[0], line[1:]) for line in difflib.unified_diff(base.splitlines(), new.splitlines(), lineterm="", n=0)
            if line[:3] not in ("---", "+++") and line[:2] != "@@"]


def compare_surface(base, new, exceptions):
    """Findings between two walks. What exceptions accepts is counted, not reported."""
    findings, accepted = [], collections.Counter()
    help_lines = merged(exceptions, "help_lines")
    flag_names = merged(exceptions, "flags")
    for path in sorted(set(base) | set(new)):
        if path not in new:
            findings.append(finding(f"surface:command:{path}", ["command removed"]))
            continue
        if path not in base:
            findings.append(finding(f"surface:command:{path}", ["command added"]))
            continue
        left, right = base[path], new[path]
        if left["exit"] != right["exit"]:
            findings.append(finding(f"surface:help:{path}", [f"exit: {left['exit']} != {right['exit']}"]))
        if left["stderr"] != right["stderr"]:
            findings.append(finding(f"surface:help:{path}", [f"stderr {s}{t}" for s, t in diff_lines(left["stderr"], right["stderr"])]))
        allowed = set(help_lines.get(path, []))
        fields = []
        for sign, text in diff_lines(left["stdout"], right["stdout"]):
            if text in allowed:
                accepted[f"help line, {path}"] += 1
            else:
                fields.append(f"help {sign}{text}")
        if fields:
            findings.append(finding(f"surface:help:{path}", fields))
        findings += compare_flags(path, parse_flags(left["stdout"]), parse_flags(right["stdout"]), flag_names.get(path, []), accepted)
    return findings, accepted


def compare_flags(path, base, new, allowed, accepted):
    out = []
    for name in sorted(set(base) | set(new)):
        if name not in base:
            if name in allowed:
                accepted[f"flag, {path} {name}"] += 1
            else:
                out.append(finding(f"surface:flag:{path} {name}", [f"flag added: {new[name]}"]))
        elif name not in new:
            out.append(finding(f"surface:flag:{path} {name}", [f"flag removed: {base[name]}"]))
        elif base[name] != new[name]:
            out.append(finding(f"surface:flag:{path} {name}", [f"{key}: {base[name][key]!r} != {new[name][key]!r}"
                                                                 for key in base[name] if base[name][key] != new[name][key]]))
    return out


def merged(exceptions, key):
    """One exception field across every accepted entry: {name: [values]}."""
    out = collections.defaultdict(list)
    for entry in exceptions.get("accepted", []):
        for name, values in entry.get(key, {}).items():
            out[name] += values
    return out


# --- Committed goldens -------------------------------------------------------------------

def compare_goldens(raw_diff, exceptions):
    """Findings from `git diff --unified=0`: a changed line no exception names, by file."""
    allowed = merged(exceptions, "golden_lines")
    findings, accepted = [], collections.Counter()
    by_file, name = collections.defaultdict(list), None
    for line in raw_diff.splitlines():
        if line.startswith("diff --git "):
            name = line.split(" b/", 1)[1]
            by_file[name]
        elif line.startswith(("+++", "---", "@@", "index ", "new file", "deleted file", "similarity", "rename ")):
            continue
        elif line[:1] in ("+", "-") and name:
            by_file[name].append(line)
    for name, lines in sorted(by_file.items()):
        kept = []
        for line in lines:
            if line[1:] in allowed.get(name, []):
                accepted[f"golden line, {name}"] += 1
            else:
                kept.append(line)
        if kept or not lines:
            findings.append(finding(f"golden:{name}", kept or ["file changed"]))
    return findings, accepted


# --- The wire proof ----------------------------------------------------------------------

def expected_client_cases():
    def text_of(group):
        return [t for t in (native_cases.as_text(c) for c in group) if t]
    reads, writes, both = native_cases.reads(), native_cases.writes(), native_cases.global_cases()
    return 1 + len(reads) + len(text_of(reads)) + len(writes) + len(text_of(writes)) + len(both) + len(text_of(both))


def wire_view(record):
    """A client record shaped so a difference names its field: lines, and `Name: value` headers."""
    def message(part):
        return {"status": part["status"]} if "status" in part else {
            "method": part["method"], "target": part["target"], "version": part["version"]}

    def side(part):
        return {**message(part), "headers": [f"{k}: {v}" for k, v in part["headers"]], "body": (part["body"] or "").splitlines()}
    return {
        "exit": record.get("exit"),
        "stdout": (record.get("stdout") or "").splitlines(),
        "stderr": (record.get("stderr") or "").splitlines(),
        "sidecar": (record.get("sidecar") or "").splitlines(),
        "requests": [{"request": side(r["request"]), "response": side(r["response"])} for r in record.get("requests", [])],
    }


def query_order_allowed(record, exceptions):
    command = " ".join(record.get("argv") or [])
    for entry in exceptions.get("accepted", []):
        for pattern in entry.get("query_order_commands", []):
            if fnmatch.fnmatchcase(command, pattern):
                return entry["id"]
    return None


def judge_client(base, new, results, exceptions, expected=None):
    """Findings for the client half, and what made a comparison impossible.

    The proof's `query-order` verdict is any command; only the commands the exceptions
    file names get it. A seed that failed, or a command that timed out, compared nothing,
    so neither counts as the same even when both sides did it.
    """
    problems, findings, accepted = [], [], collections.Counter()
    keys_base, keys_new = [r["key"] for r in base], [r["key"] for r in new]
    if sorted(keys_base) != sorted(keys_new):
        only = sorted(set(keys_base) ^ set(keys_new))
        problems.append(f"the two sides ran different commands: {only[:20]}")
    if expected is not None and len(base) != expected:
        problems.append(f"the base side ran {len(base)} commands, the table has {expected}")
    by_key = collections.defaultdict(list)
    for record in new:
        by_key[record["key"]].append(record)
    twins = []
    for record in base:
        twins.append((record, by_key[record["key"]].pop(0) if by_key[record["key"]] else None))
    for record, twin in twins:
        for name, side in (("base", record), ("new", twin)):
            if side and side.get("seed_failed"):
                problems.append(f"{record['key']}: seed failed on {name}: {side['seed_failed'][:200]}")
            if side and side.get("exit") == -1:
                problems.append(f"{record['key']}: timed out on {name}")
    if len(results) < len(base):
        problems.append(f"the proof compared {len(results)} of {len(base)} commands")
    # The proof's results are the base's records in order; any extra at the end has no
    # twin, and the key check above has already said so.
    for result, (record, twin) in zip(results, twins):
        verdict = result["verdict"]
        if verdict == "same":
            continue
        if twin is None:
            continue
        fields = json_diff(wire_view(record), wire_view(twin), limit=20) or [json.dumps(result.get("detail"))[:400]]
        if verdict == "query-order":
            allowed = query_order_allowed(record, exceptions)
            if allowed:
                accepted[f"query order, {record['label']} ({allowed})"] += 1
                continue
        findings.append(finding(f"client:{result['key']}", fields))
    return findings, accepted, problems


def judge_server(results):
    findings, problems = [], []
    if not results:
        problems.append("the server half compared nothing")
    for result in results:
        if result["same"]:
            continue
        fields = json_diff(result.get("base"), result.get("new"), limit=20)
        if result.get("state_same") is False:
            fields.append("state: what the write left behind differs")
        name = " ".join(x for x in (result["label"], result.get("method"), result.get("target")) if x)
        findings.append(finding(f"server:{name}", fields or ["answers differ"]))
    return findings, problems


def wire_floor_problems(base, server):
    """A proof that ran and touched nothing: every command failed, or nothing got an answer."""
    problems = []
    succeeded = sum(1 for r in base if r.get("exit") == 0)
    sent = sum(1 for r in base if r.get("requests"))
    if base and succeeded * 2 < len(base):
        problems.append(f"only {succeeded} of {len(base)} commands exited 0 at the base")
    if base and sent * 2 < len(base):
        problems.append(f"only {sent} of {len(base)} commands sent a request at the base")
    answered = [r for r in server if "status" in r]
    good = sum(1 for r in answered if 200 <= r["status"] < 300)
    if server and not answered:
        problems.append("the server half recorded no status")
    elif answered and good * 10 < len(answered):
        problems.append(f"only {good} of {len(answered)} server answers were 2xx")
    return problems


# --- Decision records and the verdict ----------------------------------------------------

def waivers_from(records):
    """(pattern, record path) for each `G2-waives:` line. An empty record waives nothing."""
    return [(pattern, path) for path, text in records.items() for pattern in WAIVER_LINE.findall(text)]


def apply_waivers(findings, waivers):
    standing, waived = [], []
    for item in findings:
        hit = next((path for pattern, path in waivers if fnmatch.fnmatchcase(item["id"], pattern)), None)
        if hit:
            waived.append({**item, "waived_by": hit})
        else:
            standing.append(item)
    return standing, waived


def exit_code(problems, standing):
    return 2 if problems else (1 if standing else 0)


# --- Running it --------------------------------------------------------------------------

def git(*args, check=True):
    done = subprocess.run(["git"] + list(args), cwd=repo_root(), capture_output=True, text=True)
    if check and done.returncode != 0:
        raise CouldNotCompare(f"git {' '.join(args)}: {done.stderr.strip()[:300]}")
    return done.stdout


def repo_root():
    return os.path.dirname(os.path.dirname(HERE))


def base_env(home):
    env = {k: v for k, v in os.environ.items() if k not in DROPPED_ENV}
    env.update(TELEMETRY_ENV)
    env.update({"HOME": home, "NO_COLOR": "1", "TZ": "UTC", "PYTHONDONTWRITEBYTECODE": "1"})
    return env


def check_dolt(dolt, env):
    if not os.access(dolt, os.X_OK):
        raise CouldNotCompare(f"no dolt at {dolt}")
    subprocess.run([dolt, "config", "--global", "--add", "metrics.disabled", "true"], env=env, check=True, capture_output=True)
    with open(os.path.join(repo_root(), DOLT_PIN_FILE)) as handle:
        pin = re.search(r'^readonly version="([^"]+)"$', handle.read(), re.MULTILINE).group(1)
    first = subprocess.run([dolt, "version"], env=env, capture_output=True, text=True).stdout.splitlines()
    got = first[0].split()[-1] if first else "nothing"
    if got != pin:
        raise CouldNotCompare(f"dolt reports {got}, {DOLT_PIN_FILE} pins {pin}")
    return pin


def build_bd(sha, work, name, log):
    source = os.path.join(work, f"src-{name}")
    output = os.path.join(work, "bin", name, "bd")
    os.makedirs(os.path.dirname(output), exist_ok=True)
    git("worktree", "add", "--detach", source, sha)
    try:
        started = time.time()
        # The real HOME: nothing here runs bd or dolt, and Go's caches live under it.
        with open(os.path.join(work, f"build-{name}.log"), "w") as handle:
            done = subprocess.run(["make", "build", f"BD_BUILD_OUTPUT={output}"], cwd=source, stdout=handle, stderr=subprocess.STDOUT)
        if done.returncode != 0:
            raise CouldNotCompare(f"make build failed at {sha}; see build-{name}.log")
        log(f"built bd at {sha[:9]} in {time.time() - started:.0f}s")
    finally:
        git("worktree", "remove", "--force", source, check=False)
    return output


def run_wire_proof(base_bd, new_bd, dolt, out, only, env, log):
    command = [sys.executable, os.path.join(PROOF_DIR, "wire_proof.py"), "--base", base_bd, "--new", new_bd, "--dolt", dolt, "--out", out]
    if only:
        command += ["--only", only]
    with open(os.path.join(out, "wire-proof.log"), "w") as handle:
        proc = subprocess.Popen(command, env=env, stdout=handle, stderr=subprocess.STDOUT)
        try:
            code = proc.wait()
        finally:
            if proc.poll() is None:
                proc.terminate()
                proc.wait()
    log(f"wire proof exited {code}")
    return code


def load_json(path):
    try:
        with open(path) as handle:
            return json.load(handle)
    except (OSError, ValueError) as err:
        raise CouldNotCompare(f"cannot read {path}: {err}") from err


def tail(path, count=15):
    try:
        with open(path, errors="replace") as handle:
            return "".join(handle.readlines()[-count:])
    except OSError:
        return ""


def print_findings(title, items, write):
    if not items:
        return
    write(f"\n{title} ({len(items)}):")
    for item in items:
        by = f"  [waived by {item['waived_by']}]" if item.get("waived_by") else ""
        write(f"  {item['id']}{by}")
        for field in item["fields"][:12]:
            write(f"      {str(field)[:400]}")
        if len(item["fields"]) > 12:
            write(f"      ... and {len(item['fields']) - 12} more; see g2-report.json")


def write_step_summary(path, report):
    state = {0: "PASS", 1: "FAIL: differences", 2: "FAIL: could not compare"}[report["exit"]]
    lines = [f"## G2 CLI surface: {state}", "", f"Base `{report['base'][:12]}`, head `{report['head'][:12]}`.", ""]
    for title, items in (("Could not compare", report["could_not_compare"]),):
        if items:
            lines += [f"### {title}", ""] + [f"- {item}" for item in items] + [""]
    for title, items in (("Differences", report["differences"]), ("Waived by a decision record", report["waived"])):
        if items:
            lines += [f"### {title} ({len(items)})", ""]
            for item in items[:200]:
                lines.append(f"- `{item['id']}`: `{str(item['fields'][:3])[:300]}`")
            lines += [""] if len(items) <= 200 else [f"- ...{len(items) - 200} more in the artifact", ""]
    if report["accepted"]:
        lines += ["### Accepted by g2_exceptions.json", ""] + [f"- {k}: {v}" for k, v in sorted(report["accepted"].items())] + [""]
    lines += [f"Took {report['seconds']:.0f}s."]
    with open(path, "a") as handle:
        handle.write("\n".join(lines) + "\n")


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--base-ref", required=True, help="the branch or commit the PR goes into; the base is the merge base with --head")
    parser.add_argument("--head", default="HEAD", help="the commit under test")
    parser.add_argument("--dolt", required=True, help="the pinned dolt binary")
    parser.add_argument("--out", required=True, help="an empty or new directory for the report and logs")
    parser.add_argument("--bd-base", help="use this bd for the base instead of building it")
    parser.add_argument("--bd-new", help="use this bd for the head instead of building it")
    parser.add_argument("--only", help="partial: the wire proof's client half runs only matching commands")
    parser.add_argument("--skip-wire", action="store_true", help="partial: no wire proof")
    parser.add_argument("--exceptions", help="read the exceptions from this file rather than from --head (tests)")
    parser.add_argument("--summary", help="append a Markdown summary here (GITHUB_STEP_SUMMARY)")
    args = parser.parse_args(argv)
    os.environ.update(TELEMETRY_ENV)
    for name in DROPPED_ENV:
        os.environ.pop(name, None)
    partial = bool(args.only or args.skip_wire)
    if partial and os.environ.get("GITHUB_ACTIONS"):
        parser.error("--only and --skip-wire are for local work; a gate run compares everything")

    def log(message):
        print(message, flush=True)

    out = os.path.abspath(args.out)
    os.makedirs(out, exist_ok=True)
    if os.listdir(out):
        parser.error(f"{out} is not empty: an old result must not pass for this run's")
    started = time.time()
    work = tempfile.mkdtemp(prefix="g2-gate.")
    report = {"base": "", "head": "", "partial": partial, "could_not_compare": [], "differences": [], "waived": [],
              "accepted": {}, "records": [], "notes": [], "seconds": 0.0, "exit": 2}
    try:
        run_gate(args, out, work, report, log)
    except CouldNotCompare as err:
        report["could_not_compare"].append(str(err))
    except Exception as err:  # A broken gate is a failed gate, whatever broke it.
        report["could_not_compare"].append(f"{type(err).__name__}: {err}")
    finally:
        shutil.rmtree(work, ignore_errors=True)
        git("worktree", "prune", check=False)
    code = exit_code(report["could_not_compare"], report["differences"])
    report["exit"] = code
    report["seconds"] = time.time() - started
    with open(os.path.join(out, "g2-report.json"), "w") as handle:
        json.dump(report, handle, indent=1)
    print_findings("Differences", report["differences"], log)
    print_findings("Waived by a decision record", report["waived"], log)
    for key, count in sorted(report["accepted"].items()):
        log(f"accepted: {key}: {count}")
    for note in report["notes"]:
        log(f"note: {note}")
    for problem in report["could_not_compare"]:
        log(f"COULD NOT COMPARE: {problem}")
    if args.summary:
        write_step_summary(args.summary, report)
    verdict = {0: "PASS", 1: f"FAIL: {len(report['differences'])} differences", 2: "FAIL: could not compare"}[code]
    if partial:
        verdict = f"PARTIAL RUN, NOT A GATE RESULT ({verdict})"
    log(f"\nG2 {verdict}. base {report['base'][:12]} head {report['head'][:12]}, {report['seconds']:.0f}s. Report: {out}/g2-report.json")
    return code


def run_gate(args, out, work, report, log):
    head = git("rev-parse", "--verify", args.head + "^{commit}").strip()
    target = git("rev-parse", "--verify", args.base_ref + "^{commit}").strip()
    base = git("merge-base", target, head).strip()
    report["base"], report["head"] = base, head
    log(f"G2: base {base[:12]} (merge base with {args.base_ref}), head {head[:12]}")

    home = os.path.join(work, "home")
    empty = os.path.join(work, "empty")
    os.makedirs(home)
    os.makedirs(empty)
    env = base_env(home)
    for tool in ("ps", "lsof", "git"):
        if not shutil.which(tool):
            raise CouldNotCompare(f"{tool} is not on PATH; the proof's cleanup needs it")
    dolt = os.path.abspath(args.dolt)
    pin = check_dolt(dolt, env)
    log(f"dolt {pin}")

    exceptions, gate_findings = read_exceptions(args, base, head, report)
    records = read_decision_records(base, head)
    report["records"] = sorted(records)
    changed_gate = git("diff", "--name-only", base, head, "--", *GATE_PATHS).split()
    if changed_gate:
        report["notes"].append(f"this PR changes the gate itself, so read it: {', '.join(changed_gate)}")

    base_bd = os.path.abspath(args.bd_base) if args.bd_base else build_bd(base, work, "base", log)
    new_bd = os.path.abspath(args.bd_new) if args.bd_new else build_bd(head, work, "new", log)

    findings = list(gate_findings)
    accepted = collections.Counter()

    started = time.time()
    surface_env = {**env, "PATH": os.pathsep.join(["/usr/bin", "/bin"])}
    surface = {}
    for name, bd, sha in (("base", base_bd, base), ("new", new_bd, head)):
        surface[name] = walk_surface(bd, tree_paths_at(sha), empty, surface_env)
        if len(surface[name]) < MIN_SURFACE_COMMANDS or surface[name]["bd"]["exit"] != 0:
            raise CouldNotCompare(f"the {name} walk found {len(surface[name])} commands, root exit {surface[name]['bd']['exit']}")
        failed = [p for p, a in surface[name].items() if a["exit"] != 0]
        if failed:
            report["notes"].append(f"{name}: --help exited non-zero on {failed[:10]}")
        with open(os.path.join(out, f"surface-{name}.json"), "w") as handle:
            json.dump(surface[name], handle, indent=1)
    found, ok = compare_surface(surface["base"], surface["new"], exceptions)
    findings += found
    accepted += ok
    log(f"surface: {len(surface['base'])} commands at base, {len(surface['new'])} at head, {len(found)} findings, {time.time() - started:.0f}s")

    found, ok = compare_goldens(git("diff", "--unified=0", base, head, "--", *GOLDEN_PATHS), exceptions)
    findings += found
    accepted += ok
    log(f"committed goldens: {len(found)} findings")

    problems = []
    if not args.skip_wire:
        started = time.time()
        wire_out = os.path.join(out, "wire")
        os.makedirs(wire_out)
        proof_code = run_wire_proof(base_bd, new_bd, dolt, wire_out, args.only, env, log)
        log(f"wire proof took {time.time() - started:.0f}s")
        client_results = os.path.join(wire_out, "client-comparison.json")
        server_results = os.path.join(wire_out, "server-comparison.json")
        if not (os.path.exists(client_results) and os.path.exists(server_results)):
            raise CouldNotCompare(f"the wire proof exited {proof_code} without a comparison:\n{tail(os.path.join(wire_out, 'wire-proof.log'))}")
        base_records = load_json(os.path.join(wire_out, "client-base.json"))
        new_records = load_json(os.path.join(wire_out, "client-new.json"))
        comparison = load_json(client_results)
        server = load_json(server_results)
        report["compared"] = {"commands": len(base_records), "requests": comparison.get("requests_compared"), "server_answers": len(server["results"])}
        found, ok, client_problems = judge_client(base_records, new_records, comparison["results"], exceptions,
                                                  None if args.only else expected_client_cases())
        findings += found
        accepted += ok
        server_found, server_problems = judge_server(server["results"])
        findings += server_found
        problems += client_problems + server_problems
        if not args.only:
            problems += wire_floor_problems(base_records, server["results"])
        log(f"wire: {report['compared']}, {len(found)} client findings, {len(server_found)} server findings")
        if proof_code != 0 and not (found or server_found or ok):
            # The proof fails on any command it calls neither same nor query-order, and
            # any server answer that differs. With none of those, something else failed.
            problems.append(f"the wire proof exited {proof_code} and reports no difference:\n{tail(os.path.join(wire_out, 'wire-proof.log'))}")

    standing, waived = apply_waivers(findings, waivers_from(records))
    report["differences"], report["waived"] = standing, waived
    report["accepted"] = dict(accepted)
    report["could_not_compare"] += problems


def tree_paths_at(sha):
    """Every command path the tree golden at this commit names, hidden ones included."""
    text = git("show", f"{sha}:cmd/bd/testdata/command_tree/tree.golden", check=False)
    return [m for m in TREE_HEADING.findall(text)]


def read_exceptions(args, base, head, report):
    findings = []
    if args.exceptions:
        with open(args.exceptions) as handle:
            text = handle.read()
    else:
        text = git("show", f"{head}:{EXCEPTIONS_PATH}", check=False)
        if not text:
            raise CouldNotCompare(f"{EXCEPTIONS_PATH} is not at {head[:12]}")
    try:
        exceptions = json.loads(text)
    except ValueError as err:
        raise CouldNotCompare(f"{EXCEPTIONS_PATH}: {err}") from err
    for entry in exceptions.get("accepted", []):
        if not entry.get("id") or not entry.get("reason"):
            raise CouldNotCompare(f"{EXCEPTIONS_PATH}: every entry needs an id and a reason: {str(entry)[:200]}")
    at_base = git("show", f"{base}:{EXCEPTIONS_PATH}", check=False)
    if at_base and at_base != git("show", f"{head}:{EXCEPTIONS_PATH}", check=False):
        findings.append(finding("gate:exceptions-file-changed", [f"{EXCEPTIONS_PATH} differs from the base: widening what G2 accepts takes a decision record"]))
    elif not at_base:
        report["notes"].append(f"{EXCEPTIONS_PATH} is new in this PR: nothing at the base to hold it to")
    return exceptions, findings


def read_decision_records(base, head):
    """The .md files this PR adds under engdocs/decisions/, by path."""
    added = git("diff", "--diff-filter=A", "--name-only", base, head, "--", DECISIONS_DIR).split("\n")
    return {path: git("show", f"{head}:{path}") for path in added if path.endswith(".md")}


if __name__ == "__main__":
    sys.exit(main())
