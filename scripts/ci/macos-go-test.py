#!/usr/bin/env python3
"""Keep macOS race coverage exhaustive without one oversized graphstore process.

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
import subprocess
import sys
import time

GRAPHSTORE = "github.com/steveyegge/beads/internal/storage/graphstore"
FLAGS = ["-tags", "gms_pure_go", "-v", "-race", "-short", "-timeout=30m", "-skip", "^TestEmbedded"]
GROUPS = 4
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
    def __init__(self, root, output):
        self.root, self.output = root, output
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
                require(not group_alive(process.pid), f"{label} left descendants after exit")
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
        print(f"{label}: exit={receipt['exitCode']} elapsed={receipt['elapsedSeconds']:.2f}s", flush=True)
        if failure is not None:
            for suffix in ("stdout", "stderr"):
                with (self.output / f"{label}.{suffix}").open("rb") as saved:
                    saved.seek(0, os.SEEK_END)
                    saved.seek(max(0, saved.tell() - 16000))
                    print(saved.read().decode(errors="replace"), file=sys.stderr)
            raise failure
        return self.output / (label + ".stdout")


def partition(names):
    require(names and len(names) == len(set(names)), "empty or duplicate discovered test set")
    eligible = sorted(name for name in names if not name.startswith("TestEmbedded"))
    require(eligible, "no eligible graphstore tests")
    groups = [eligible[index::GROUPS] for index in range(GROUPS)]
    require(Counter(name for group in groups for name in group) == Counter(eligible), "incomplete group partition")
    return groups


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
        require(roots == set(expected_roots), "executed graphstore roots differ from discovered selection")
    return dict(roots=sorted(roots), tests=len(terminal),
                skipped=sorted(f"{package}:{name}" for (package, name), action in terminal.items() if action == "skip"),
                packages=sorted(packages))


def execute(root, output, go="go"):
    output.mkdir(parents=True, exist_ok=False)
    runner = Runner(root, output)
    summary = dict(passed=False, failure=None, groups=[])
    try:
        packages_file = runner.run("package-discovery", [go, "list", "-race", "-tags", "gms_pure_go", "./..."], DISCOVERY_SECONDS)
        packages = packages_file.read_text().splitlines()
        require(packages and len(packages) == len(set(packages)) and GRAPHSTORE in packages, "package discovery missing graphstore or contains duplicates")
        require(all(re.fullmatch(r"[A-Za-z0-9_./-]+", package) for package in packages), "malformed package discovery")
        other = sorted(set(packages) - {GRAPHSTORE})
        require(other, "missing non-graphstore package coverage")
        names_file = runner.run("graphstore-discovery", [go, "test", *FLAGS, "-list", ".", GRAPHSTORE], DISCOVERY_SECONDS)
        names = discovered_names(names_file)
        groups = partition(names)
        write_json(output / "plan.json", dict(packages=sorted(packages), otherPackages=other, discovered=sorted(names),
                   excluded=sorted(set(names) - set(sum(groups, []))), groups=groups, flags=FLAGS))
        # Separate Go processes reset only the package alarm, not any test body.
        for index, group in enumerate(groups, 1):
            label = f"graphstore-{index}"
            write_json(output / (label + "-manifest.json"), group)
            if not group:
                summary["groups"].append(dict(label=label, roots=[], empty=True))
                continue
            selected = "^(" + "|".join(group) + ")$"
            events = runner.run(label, [go, "test", *FLAGS, "-json", "-count=1", "-run", selected, GRAPHSTORE], GROUP_SECONDS)
            census = verify_events(events, [GRAPHSTORE], group)
            write_json(output / (label + "-census.json"), census)
            summary["groups"].append(dict(label=label, **census))
        events = runner.run("other-packages", [go, "test", *FLAGS, "-json", *other], OTHER_SECONDS)
        summary["other"] = verify_events(events, other)
        summary["passed"] = True
    except BaseException as error:
        summary["failure"] = str(error) or type(error).__name__
        raise
    finally:
        write_json(output / "summary.json", summary)
        print(json.dumps(summary, sort_keys=True), flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=Path("artifacts/macos-go-test"))
    args = parser.parse_args()
    signal.signal(signal.SIGTERM, interrupted)
    signal.signal(signal.SIGINT, interrupted)
    try:
        execute(Path(__file__).resolve().parents[2], args.output.resolve())
    except Exception as error:
        print(f"macOS test dispatch failed: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
