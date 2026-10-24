#!/usr/bin/env python3
"""Keep macOS race coverage exhaustive without oversized graphstore or CLI processes.

The existing Linux name-shard helpers establish the exhaustive-discovery pattern.
This lane uses Python so it does not require mapfile from Bash newer than macOS's.
No test environment wrapper is introduced: server availability/skip policy stays
with the existing tests, independently of required two-engine graph qualification.
"""
import argparse
from collections import Counter
from contextlib import contextmanager
import json
import os
from pathlib import Path
import re
import signal
import shutil
import subprocess
import sys
import time

GRAPHSTORE = "github.com/steveyegge/beads/internal/storage/graphstore"
COMMAND = "github.com/steveyegge/beads/cmd/bd"
FLAGS = ["-tags", "gms_pure_go", "-v", "-race", "-short", "-timeout=30m", "-skip", "^TestEmbedded"]
# -json supplies framed verbose output. Explicit -v would override its
# test2json framing and let unterminated application output hide RUN/PASS.
JSON_FLAGS = [flag for flag in FLAGS if flag != "-v"]
GROUPS = 4  # Command groups and the default pure partition helper.
# A native 81-root group exhausted 30m after 79 passing roots.
# Finer graphstore batches retain every alarm and the total scope deadline.
GRAPHSTORE_GROUPS = 8
GRAPHSTORE_SCOPES = ("graphstore-odd", "graphstore-even")
CI_SCOPES = (*GRAPHSTORE_SCOPES, "remaining")
DISCOVERY_SECONDS = 900
GROUP_SECONDS = 2100
OTHER_SECONDS = 3600
TOTAL_SECONDS = 6900  # Finish cleanup before the workflow's 120-minute ceiling.
CLEANUP_SECONDS = 5
_deferred_signals = 0
_pending_signal = None


@contextmanager
def defer_signals():
    # Defer Python handlers, not the OS signal mask inherited by children.
    global _deferred_signals
    _deferred_signals += 1
    try:
        yield
    finally:
        _deferred_signals -= 1


def interrupted(signum, _frame):
    global _pending_signal
    if _deferred_signals:
        _pending_signal = signum
    else:
        raise RuntimeError(f"interrupted by signal {signum}")


def take_interrupt():
    global _pending_signal
    signum, _pending_signal = _pending_signal, None
    return RuntimeError(f"interrupted by signal {signum}") if signum is not None else None


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def write_json(path, value):
    path.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n")


def group_alive(pid):
    try:
        os.killpg(pid, 0)
        return True
    except ProcessLookupError:
        return False
    except PermissionError:
        # Darwin can return EPERM while a just-exited leader is awaiting reap.
        # This is not proof of absence; retain ownership and retry after poll.
        return True


def group_snapshot(pid):
    """Diagnose only the owned group without recording argv or environment."""
    try:
        result = subprocess.run(
            ["ps", "-axo", "pid=,ppid=,pgid=,stat=,comm="],
            capture_output=True, text=True, timeout=2, check=True)
    except (OSError, subprocess.TimeoutExpired, subprocess.CalledProcessError) as error:
        return {"error": type(error).__name__}
    members = []
    for line in result.stdout.splitlines():
        fields = line.split(None, 4)
        if len(fields) != 5:
            continue
        try:
            child, parent, group = map(int, fields[:3])
        except ValueError:
            continue
        if group == pid:
            members.append(dict(pid=child, ppid=parent, pgid=group,
                                state=fields[3], executable=Path(fields[4]).name))
    return {"members": sorted(members, key=lambda item: item["pid"])}


def signal_group(pid, value):
    try:
        os.killpg(pid, value)
    except (ProcessLookupError, PermissionError):
        # retire still requires a reaped owner and proven group absence.
        pass


def retire(process):
    """Reap the direct owner and retire its inherited group, even after owner exit."""
    signal_group(process.pid, signal.SIGTERM)
    deadline = time.monotonic() + CLEANUP_SECONDS
    while time.monotonic() < deadline:
        process.poll()  # Reap the leader before checking group membership.
        if not group_alive(process.pid):
            return
        time.sleep(0.02)
    signal_group(process.pid, signal.SIGKILL)
    process.wait(timeout=CLEANUP_SECONDS)
    deadline = time.monotonic() + CLEANUP_SECONDS
    while group_alive(process.pid) and time.monotonic() < deadline:
        time.sleep(0.02)
    require(not group_alive(process.pid), "owned process group survived cleanup")


class Runner:
    def __init__(self, root, output, json_output=None):
        self.root, self.output = root, output
        self.json_output = json_output
        self.deadline = time.monotonic() + TOTAL_SECONDS
        self.receipts = []

    def run(self, label, argv, timeout):
        timeout = min(timeout, self.deadline - time.monotonic())
        require(timeout > 0, "macOS test dispatch total deadline exhausted")
        started = time.monotonic()
        receipt = dict(label=label, argv=argv, timeoutSeconds=timeout, exitCode=None,
                       failure=None, groupGone=False)
        self.receipts.append(receipt)
        write_json(self.output / "processes.json", self.receipts)
        print(f"{label}: starting (bound {timeout:.0f}s)", flush=True)
        process = None
        failure = None
        with (self.output / (label + ".stdout")).open("wb") as out, (self.output / (label + ".stderr")).open("wb") as err:
            try:
                with defer_signals():
                    process = subprocess.Popen(argv, cwd=self.root, stdout=out, stderr=err, start_new_session=True)
                    receipt["pid"] = process.pid
                pending = take_interrupt()
                if pending is not None:
                    raise pending
                process.wait(timeout=timeout)
                require(process.returncode == 0, f"{label} exited {process.returncode}")
                left_descendants = group_alive(process.pid)
                if left_descendants:
                    receipt["remainingProcesses"] = group_snapshot(process.pid)
                # Preserve the original verdict even if diagnostic sampling
                # happens to observe a process that exits meanwhile.
                require(not left_descendants, f"{label} left descendants after exit")
            except BaseException as error:
                failure = error
                receipt["failure"] = str(error) or type(error).__name__
            finally:
                with defer_signals():
                    if process is not None:
                        try:
                            if group_alive(process.pid):
                                retire(process)
                        except BaseException as error:
                            receipt["cleanupError"] = str(error)
                            if failure is None:
                                failure = error
                        receipt["exitCode"] = process.poll()
                        receipt["groupGone"] = not group_alive(process.pid)
                    receipt["elapsedSeconds"] = time.monotonic() - started
                    write_json(self.output / "processes.json", self.receipts)
                pending = take_interrupt()
                if pending is not None and failure is None:
                    failure = pending
                    receipt["failure"] = str(pending)
                    write_json(self.output / "processes.json", self.receipts)
        if self.json_output is not None and "-json" in argv and argv[1] == "test":
            # Preserve actual Go JSON, including failed/truncated runs. Discovery
            # and dispatcher diagnostics never enter the equivalence input.
            with (self.output / (label + ".stdout")).open("rb") as source, self.json_output.open("ab") as target:
                shutil.copyfileobj(source, target)
        print(f"{label}: exit={receipt['exitCode']} elapsed={receipt['elapsedSeconds']:.2f}s", flush=True)
        if failure is not None:
            for suffix in ("stdout", "stderr"):
                with (self.output / f"{label}.{suffix}").open("rb") as saved:
                    saved.seek(0, os.SEEK_END)
                    saved.seek(max(0, saved.tell() - 16000))
                    print(saved.read().decode(errors="replace"), file=sys.stderr)
            raise failure
        return self.output / (label + ".stdout")


def partition(names, group_count=GROUPS):
    require(type(group_count) is int and group_count > 0, "invalid partition count")
    require(names and len(names) == len(set(names)), "empty or duplicate discovered test set")
    eligible = sorted(name for name in names if not name.startswith("TestEmbedded"))
    require(eligible, "no eligible package tests")
    groups = [eligible[index::group_count] for index in range(group_count)]
    require(Counter(name for group in groups for name in group) == Counter(eligible), "incomplete group partition")
    return groups


def selected_group_numbers(scope, count):
    numbers = list(range(1, count + 1))
    if scope in GRAPHSTORE_SCOPES:
        return [number for number in numbers if number % 2 == (1 if scope == "graphstore-odd" else 0)]
    return numbers


def discovered_names(path):
    # -list . also lists benchmarks; those do not run in normal `go test`.
    # Examples and fuzz seed tests DO run normally and must remain in the union.
    names = [line for line in path.read_text().splitlines()
             if re.fullmatch(r"(?:Test|Example|Fuzz)\w*", line)]
    require(names and len(names) == len(set(names)), "empty or duplicate compiled test discovery")
    return names


def verify_events(path, packages, expected_roots=None):
    packages = set(packages)
    started, terminal, package_ends = Counter(), {}, Counter()
    roots = set()
    with path.open() as stream:
        for line in stream:
            event = json.loads(line)
            require(isinstance(event, dict), "Go JSON event is not an object")
            action, package, name = event.get("Action"), event.get("Package"), event.get("Test")
            require(isinstance(action, str), "Go JSON event is missing its action")
            require(action not in ("fail", "build-fail"), "Go JSON contains failure")
            if action.startswith("build-"):
                continue
            require(package in packages, f"unexpected package in receipt: {package}")
            if name:
                key = (package, name)
                if action == "run":
                    started[key] += 1
                    require(started[key] == 1, f"duplicate test execution: {name}")
                    if "/" not in name:
                        roots.add(name)
                elif action in ("pass", "skip"):
                    require(started[key] == 1 and key not in terminal, f"test terminal without unique run: {name}")
                    terminal[key] = action
            elif action in ("pass", "skip"):
                package_ends[package] += 1
    require(set(started) == set(terminal), "missing test terminal receipt")
    require(package_ends == Counter({package: 1 for package in packages}), "missing or duplicate package terminal receipt")
    if expected_roots is not None:
        require(roots == set(expected_roots), "executed package roots differ from discovered selection")
    return dict(roots=sorted(roots), tests=len(terminal),
                skipped=sorted(f"{package}:{name}" for (package, name), action in terminal.items() if action == "skip"),
                packages=sorted(packages))


def policy_flags(package_parallel=None, test_parallel=None):
    flags = list(FLAGS)
    for flag, value in (("-p", package_parallel), ("-parallel", test_parallel)):
        if value is not None:
            require(type(value) is int and value > 0, f"{flag} must be a positive integer")
            flags.extend([flag, str(value)])
    return flags


def execute(root, output, go="go", package_parallel=None, test_parallel=None, json_output=None, scope="all", source_sha=None):
    require(scope in ("all", "graphstore", *CI_SCOPES), "unknown execution scope")
    graph_scope = scope == "graphstore" or scope in GRAPHSTORE_SCOPES
    require(source_sha is None or re.fullmatch(r"[0-9a-f]{40}", source_sha), "invalid source revision")
    flags = policy_flags(package_parallel, test_parallel)
    json_flags = [flag for flag in flags if flag != "-v"]
    output.mkdir(parents=True, exist_ok=False)
    if json_output is not None:
        json_output = Path(json_output).resolve()
        require(not json_output.is_relative_to(output.resolve()), "combined Go JSON must be outside individual evidence")
        json_output.write_bytes(b"")
    runner = Runner(root, output, json_output)
    summary = dict(passed=False, failure=None, groups=[], commandGroups=[], scope=scope, source=source_sha)
    try:
        packages_file = runner.run("package-discovery", [go, "list", "-race", "-tags", "gms_pure_go", "./..."], DISCOVERY_SECONDS)
        packages = packages_file.read_text().splitlines()
        require(packages and len(packages) == len(set(packages)) and {GRAPHSTORE, COMMAND} <= set(packages), "package discovery missing partitioned packages or contains duplicates")
        require(all(re.fullmatch(r"[A-Za-z0-9_./-]+", package) for package in packages), "malformed package discovery")
        other = sorted(set(packages) - {GRAPHSTORE, COMMAND})
        require(other, "missing remaining package coverage")
        partitions = []
        for label, package, summary_key in (("graphstore", GRAPHSTORE, "groups"), ("cmd-bd", COMMAND, "commandGroups")):
            if (graph_scope and package != GRAPHSTORE) or (scope == "remaining" and package == GRAPHSTORE):
                continue
            names_file = runner.run(label + "-discovery", [go, "test", *flags, "-list", ".", package], DISCOVERY_SECONDS)
            names = discovered_names(names_file)
            groups = partition(names, GRAPHSTORE_GROUPS if package == GRAPHSTORE else GROUPS)
            plan = dict(package=package, discovered=sorted(names),
                        excluded=sorted(set(names) - set(sum(groups, []))), groups=groups,
                        selectedGroups=selected_group_numbers(scope, len(groups)))
            partitions.append((label, package, summary_key, plan))
        graph_plan = next((plan for _, package, _, plan in partitions if package == GRAPHSTORE), None)
        command_plan = next((plan for _, package, _, plan in partitions if package == COMMAND), None)
        scoped_packages = [GRAPHSTORE] if graph_scope else sorted(set(packages) - {GRAPHSTORE}) if scope == "remaining" else sorted(packages)
        # Preserve the default full-dispatch evidence shape. Explicit CI scopes
        # advertise their exact package ownership and source for the union gate.
        write_json(output / "plan.json", dict(packages=sorted(packages), otherPackages=other,
                   **{key: graph_plan[key] if graph_plan else [] for key in ("discovered", "excluded", "groups", "selectedGroups")},
                   command=command_plan, flags=flags, jsonFlags=json_flags,
                   scope=scope, source=source_sha, scopedPackages=scoped_packages))
        # Separate Go processes reset only each package alarm, not any test body.
        # The same total deadline and strict owned-process cleanup still apply.
        for prefix, package, summary_key, plan in partitions:
            for index, group in enumerate(plan["groups"], 1):
                if index not in plan["selectedGroups"]:
                    continue
                label = f"{prefix}-{index}"
                write_json(output / (label + "-manifest.json"), group)
                if not group:
                    summary[summary_key].append(dict(label=label, roots=[], empty=True))
                    continue
                selected = "^(" + "|".join(group) + ")$"
                events = runner.run(label, [go, "test", *json_flags, "-json", "-count=1", "-run", selected, package], GROUP_SECONDS)
                census = verify_events(events, [package], group)
                write_json(output / (label + "-census.json"), census)
                summary[summary_key].append(dict(label=label, **census))
        if not graph_scope:
            events = runner.run("other-packages", [go, "test", *json_flags, "-json", *other], OTHER_SECONDS)
            summary["other"] = verify_events(events, other)
        summary["passed"] = True
    except BaseException as error:
        summary["failure"] = str(error) or type(error).__name__
        raise
    finally:
        write_json(output / "summary.json", summary)
        print(json.dumps(summary, sort_keys=True), flush=True)


def verify_scopes(root, source_sha, package_parallel=None, test_parallel=None):
    """Recheck all source-matched scope outputs and their exact test/package union."""
    require(re.fullmatch(r"[0-9a-f]{40}", source_sha or ""), "missing aggregate source revision")
    summaries = sorted(root.rglob("summary.json"))
    require(len(summaries) == len(CI_SCOPES), "all and only the three scope artifacts are required")
    flags = policy_flags(package_parallel, test_parallel)
    json_flags = [flag for flag in flags if flag != "-v"]
    scopes, coverage, packages = set(), Counter(), None
    graph_roots, graph_discovered = Counter(), None
    for path in summaries:
        directory = path.parent
        summary = json.loads(path.read_text())
        plan = json.loads((directory / "plan.json").read_text())
        scope = plan.get("scope")
        require(scope in CI_SCOPES and scope not in scopes, "duplicate or unexpected scope")
        scopes.add(scope)
        require(summary.get("passed") is True and summary.get("failure") is None and summary.get("scope") == scope,
                "scope did not complete successfully")
        require(plan.get("source") == summary.get("source") == source_sha, "scope source revision differs")
        require(plan.get("flags") == flags and plan.get("jsonFlags") == json_flags, "scope changed test policy")
        discovered = plan.get("packages")
        require(isinstance(discovered, list) and discovered == sorted(set(discovered)) and
                {GRAPHSTORE, COMMAND} <= set(discovered), "invalid complete package discovery")
        require(sorted((directory / "package-discovery.stdout").read_text().splitlines()) == discovered,
                "package plan differs from compiled package discovery")
        if packages is None:
            packages = discovered
        require(discovered == packages, "scopes disagree on complete package discovery")
        graph_scope = scope in GRAPHSTORE_SCOPES
        expected = [GRAPHSTORE] if graph_scope else sorted(set(packages) - {GRAPHSTORE})
        require(plan.get("scopedPackages") == expected, "scope package selection differs")
        coverage.update(expected)
        prefix, package, key = ("graphstore", GRAPHSTORE, "groups") if graph_scope else ("cmd-bd", COMMAND, "commandGroups")
        selected = plan if graph_scope else plan.get("command")
        require(isinstance(selected, dict), "missing partition discovery")
        require(sorted(discovered_names(directory / (prefix + "-discovery.stdout"))) == selected.get("discovered"),
                "partition plan differs from compiled test discovery")
        groups = partition(selected["discovered"], GRAPHSTORE_GROUPS if package == GRAPHSTORE else GROUPS)
        numbers = selected_group_numbers(scope, len(groups))
        require(selected.get("groups") == groups and selected.get("selectedGroups") == numbers
                and len(summary.get(key, [])) == len(numbers), "scope did not retain its exact complete partition")
        require(selected.get("excluded") == sorted(set(selected["discovered"]) - set(sum(groups, []))),
                "scope changed intentional test exclusions")
        if graph_scope:
            if graph_discovered is None:
                graph_discovered = selected["discovered"]
            require(selected["discovered"] == graph_discovered, "graphstore scopes disagree on complete test discovery")
        expected_labels = ["package-discovery", prefix + "-discovery"]
        for position, index in enumerate(numbers):
            group = groups[index - 1]
            label = f"{prefix}-{index}"
            if graph_scope:
                graph_roots.update(group)
            if not group:
                require(summary[key][position] == dict(label=label, roots=[], empty=True), "invalid empty group")
                continue
            expected_labels.append(label)
            census = verify_events(directory / (label + ".stdout"), [package], group)
            require(json.loads((directory / (label + "-census.json")).read_text()) == census and
                    summary[key][position] == dict(label=label, **census), "group proof differs from actual output")
        if scope == "remaining":
            expected_labels.append("other-packages")
            other = sorted(set(packages) - {GRAPHSTORE, COMMAND})
            require(plan.get("otherPackages") == other and other, "remaining package set differs")
            require(summary.get("other") == verify_events(directory / "other-packages.stdout", other),
                    "remaining package proof differs from actual output")
        receipts = json.loads((directory / "processes.json").read_text())
        require([receipt.get("label") for receipt in receipts] == expected_labels, "missing or unexpected process receipt")
        require(receipts and all(receipt.get("exitCode") == 0 and receipt.get("failure") is None and
                                receipt.get("groupGone") is True and not receipt.get("cleanupError")
                                for receipt in receipts), "scope process/cleanup proof failed")
    require(scopes == set(CI_SCOPES) and coverage == Counter({package: 2 if package == GRAPHSTORE else 1 for package in packages}),
            "scope union omits or duplicates package coverage")
    require(graph_roots == Counter(sum(partition(graph_discovered, GRAPHSTORE_GROUPS), [])),
            "graphstore scope union omits or duplicates test roots")
    return dict(passed=True, source=source_sha, scopes=sorted(scopes), packages=packages)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=Path("artifacts/macos-go-test"))
    parser.add_argument("--package-parallel", type=int)
    parser.add_argument("--test-parallel", type=int)
    parser.add_argument("--json-output", type=Path, help="concatenate raw execution JSON for PR-core/Bazel equivalence")
    parser.add_argument("--scope", choices=("all", "graphstore", *CI_SCOPES), default="all")
    parser.add_argument("--source-sha", help="actual checkout revision for CI scope composition")
    parser.add_argument("--verify-scopes", type=Path, help="verify the three downloaded CI scope artifacts without rerunning tests")
    args = parser.parse_args()
    signal.signal(signal.SIGTERM, interrupted)
    signal.signal(signal.SIGINT, interrupted)
    try:
        if args.verify_scopes is not None:
            print(json.dumps(verify_scopes(args.verify_scopes.resolve(), args.source_sha, args.package_parallel, args.test_parallel), sort_keys=True))
        else:
            execute(Path(__file__).resolve().parents[2], args.output.resolve(),
                    package_parallel=args.package_parallel, test_parallel=args.test_parallel, json_output=args.json_output,
                    scope=args.scope, source_sha=args.source_sha)
    except Exception as error:
        print(f"macOS test dispatch failed: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
