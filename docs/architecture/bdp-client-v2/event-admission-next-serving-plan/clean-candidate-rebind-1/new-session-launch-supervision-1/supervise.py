"""Single-attempt outer supervisor. Local evidence survives failed role checks."""
from pathlib import Path
import datetime
import hashlib
import json
import os
import signal
import subprocess
import sys
import threading
import time

HERE = Path(__file__).resolve().parent
CONTROL = json.loads((HERE / "control.json").read_text())
PACKET = Path(CONTROL["packet"])
PYTHON = "/opt/homebrew/opt/python@3.14/bin/python3.14"
OUTPUTS = ["source-before.json", "source-after.json", "process.json", "stream.jsonl",
           "progress.jsonl", "stderr.txt", "exit.json", "final-review.txt", "terminal-results.json"]


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    require(len(sys.argv) == 1, "No arbitrary command or retry arguments accepted")
    require(CONTROL["rootLaunchReleased"] is True, "CLOSED supplement; root release required")
    os.umask(0o077)
    # Exclusive marker is deliberately retained even when preflight/claim/spawn fails.
    attempt = HERE / "attempt"
    attempt.mkdir(exist_ok=False)
    begin = time.monotonic()
    epoch_begin = time.time()
    started = datetime.datetime.now(datetime.timezone.utc).isoformat()
    events = (attempt / "outer-progress.jsonl").open("x")
    log_lock = threading.Lock()
    state_lock = threading.Lock()
    state = {"reason": None, "termSent": False, "killSent": False, "errors": [],
             "known": {}, "groups": {}, "wrapperPid": None, "forced": False,
             "lastRows": [], "ownershipUnknown": False}
    finished = threading.Event()
    wrapper = None
    checks = []
    guard = None
    final_role_ok = False

    def log(kind, **fields):
        with log_lock:
            events.write(json.dumps({"elapsedSeconds": time.monotonic() - begin,
                                     "kind": kind, **fields}) + "\n")
            events.flush()

    def fail(reason):
        with state_lock:
            if state["reason"] is None:
                state["reason"] = reason

    def inventory():
        # No command arguments or credentials enter samples. Each query is bounded.
        raw = subprocess.check_output(
            ["ps", "-axo", "pid=,ppid=,pgid=,lstart=,comm="], text=True, timeout=0.4)
        rows = {}
        for line in raw.splitlines():
            parts = line.split(None, 8)
            if len(parts) != 9:
                continue
            pid, ppid, pgid = map(int, parts[:3])
            rows[pid] = {"pid": pid, "ppid": ppid, "pgid": pgid,
                         "start": " ".join(parts[3:8]), "comm": parts[8]}
        return rows

    def remember_direct(proc, kind):
        # A Popen child belongs to this supervisor; never infer ownership by name alone.
        rows = inventory()
        row = rows.get(proc.pid)
        if row is not None:
            require(row["ppid"] == os.getpid() and row["pgid"] == proc.pid,
                    "Unexpected direct child ownership")
            with state_lock:
                state["known"][row["pid"]] = row
                state["groups"][row["pgid"]] = kind

    def collect():
        rows = inventory()
        with state_lock:
            owned = {pid for pid, old in state["known"].items()
                     if pid in rows and rows[pid]["start"] == old["start"]}
            # Cover Popen-to-registration windows using verified direct parentage.
            if state["wrapperPid"] in rows:
                row = rows[state["wrapperPid"]]
                if row["ppid"] == os.getpid() and row["pgid"] == row["pid"]:
                    owned.add(row["pid"])
            changed = True
            while changed:
                children = {pid for pid, row in rows.items() if row["ppid"] in owned}
                changed = not children.issubset(owned)
                owned.update(children)
            # The frozen runner starts Claude in a separate session. Bind its PID/PGID
            # to its authenticated process record, start time and expected command.
            process_file = PACKET / "process.json"
            if process_file.exists():
                try:
                    record = json.loads(process_file.read_text())
                    pid = record["pid"]
                    require(type(pid) is int and record["pgid"] == pid, "Bad reviewer identity")
                    require(record["argv"] == CONTROL["reviewerArgv"], "Reviewer command changed")
                    observed = datetime.datetime.fromisoformat(record["started"]).timestamp()
                    require(observed >= epoch_begin - 1, "Stale reviewer process record")
                    if pid in rows:
                        row = rows[pid]
                        process_start = time.mktime(time.strptime(row["start"], "%a %b %d %H:%M:%S %Y"))
                        require(row["pgid"] == pid and abs(process_start - observed) < 2,
                                "Reviewer PID reused or process-group identity changed")
                        require(Path(row["comm"]).name == "claude", "Reviewer executable identity changed")
                        owned.add(pid)
                        state["groups"][pid] = "reviewer"
                except json.JSONDecodeError:
                    # A concurrent initial write may be incomplete; never signal from it.
                    pass
                except Exception as exc:
                    error = "reviewer_identity:" + type(exc).__name__
                    if error not in state["errors"]:
                        state["errors"].append(error)
                    state["ownershipUnknown"] = True
                    if state["reason"] is None:
                        state["reason"] = "reviewer_identity_failure"
            # Repeat after admitting the separately sessioned reviewer.
            changed = True
            while changed:
                children = {pid for pid, row in rows.items() if row["ppid"] in owned}
                changed = not children.issubset(owned)
                owned.update(children)
            for pid in owned:
                row = rows[pid]
                state["known"][pid] = row
                if row["pgid"] != os.getpgrp():
                    state["groups"].setdefault(row["pgid"], "observed-descendant")
            state["lastRows"] = [rows[pid] for pid in sorted(owned)]
            return rows, list(state["lastRows"])

    def send_owned(sig):
        # Refresh identities before every signal. PID/start-time prevents PID-reuse kills.
        rows, owned = collect()
        valid_groups = {row["pgid"] for row in owned if row["pgid"] != os.getpgrp()}
        for pgid in sorted(valid_groups):
            # At least one live member has verified direct/descendant/record ownership.
            try:
                os.killpg(pgid, sig)
            except ProcessLookupError:
                continue
            log("signal", pgid=pgid, signal=int(sig))
        if valid_groups:
            with state_lock:
                state["forced"] = True
        return owned

    def watchdog():
        # Independent of final remote checks, wrapper waits or source hashing.
        while not finished.is_set():
            elapsed = time.monotonic() - begin
            try:
                if elapsed >= 1790:
                    fail("outer_timeout")
                    with state_lock:
                        term = not state["termSent"]
                        state["termSent"] = True
                    if term:
                        send_owned(signal.SIGTERM)
                if elapsed >= 1798:
                    with state_lock:
                        kill = not state["killSent"]
                        state["killSent"] = True
                    if kill:
                        # Freeze launch-capable groups before the final descendant census.
                        send_owned(signal.SIGSTOP)
                        send_owned(signal.SIGKILL)
                if elapsed >= 1800:
                    with state_lock:
                        state["ownershipUnknown"] = bool(state["lastRows"])
                    return
            except BaseException as exc:
                with state_lock:
                    state["errors"].append("watchdog:" + type(exc).__name__)
                    state["ownershipUnknown"] = True
            finished.wait(0.05)

    def check_session():
        nonlocal guard
        require(time.monotonic() - begin < 1790, "No guard launch after outer TERM boundary")
        out = (attempt / ("guard-%03d.stdout" % len(checks))).open("xb")
        err = (attempt / ("guard-%03d.stderr" % len(checks))).open("xb")
        proc = subprocess.Popen([PYTHON, str(guard), "/usr/bin/true"],
                                stdout=out, stderr=err, start_new_session=True)
        out.close()
        err.close()
        remember_direct(proc, "session-check")
        checks.append({"pid": proc.pid, "startedElapsed": time.monotonic() - begin})
        return proc

    def wait_check(proc):
        while proc.poll() is None:
            if time.monotonic() - begin >= 1790 or state["reason"] is not None:
                return False
            collect()
            time.sleep(0.05)
        checks[-1]["exitCode"] = proc.returncode
        return proc.returncode == 0

    def interrupted(_sig, _frame):
        fail("supervisor_interrupted")
        raise KeyboardInterrupt()

    original_signals = {s: signal.signal(s, interrupted) for s in (signal.SIGTERM, signal.SIGINT)}
    timer = threading.Thread(target=watchdog, daemon=True)
    timer.start()
    try:
        require(not any((PACKET / name).exists() for name in OUTPUTS), "Prior packet attempt exists; no retry")
        closed = json.loads((HERE / "closed-binding.json").read_text())
        bound = json.loads((PACKET / "binding.json").read_text())
        expected = dict(closed)
        expected["rootLaunchReleased"] = True
        require(bound == expected, "Packet release must change only rootLaunchReleased false to true")
        frozen = json.loads((HERE / "frozen-packet-index.json").read_text())
        require({str(f.relative_to(PACKET)) for f in PACKET.rglob("*") if f.is_file()} == set(frozen),
                "Frozen packet file set changed")
        for name, digest in frozen.items():
            if name != "binding.json":
                require(sha(PACKET / name) == digest, "Frozen packet bytes changed: " + name)
        require(CONTROL["expectedSession"] == "2026-10-18T02:16:24Z-549f", "Wrong session control")
        guard = PACKET / "guarded-action-session-20261018-021624-549f.py"
        log("preflight_complete", releasedBindingSha256=sha(PACKET / "binding.json"),
            termAfterSeconds=1790, killAfterSeconds=1798, totalSeconds=1800)
        initial = check_session()
        require(wait_check(initial), "Remote role check failed; no reviewer launch")
        require(state["reason"] is None and time.monotonic() - begin < 1788,
                "No wrapper launch near or after outer TERM boundary")
        stdout = (attempt / "wrapper.stdout").open("xb")
        stderr = (attempt / "wrapper.stderr").open("xb")
        env = {key: value for key, value in os.environ.items() if key != "PYTHONOPTIMIZE"}
        wrapper = subprocess.Popen([PYTHON, str(PACKET / "run.py")],
                                   cwd=CONTROL["sourceRoot"], stdout=stdout, stderr=stderr,
                                   start_new_session=True, env=env)
        state["wrapperPid"] = wrapper.pid
        stdout.close()
        stderr.close()
        remember_direct(wrapper, "wrapper")
        log("wrapper_started", pid=wrapper.pid, pgid=wrapper.pid)
        pending_check = None
        next_check = time.monotonic() + 15
        next_sample = 0
        while wrapper.poll() is None:
            now = time.monotonic()
            _, owned = collect()
            if now >= next_sample:
                log("process_sample", owned=owned)
                next_sample = now + 1
            if state["reason"] is not None or now - begin >= 1790:
                break
            if pending_check is not None and pending_check.poll() is not None:
                checks[-1]["exitCode"] = pending_check.returncode
                if pending_check.returncode != 0:
                    fail("role_check_failed_or_superseded")
                    break
                pending_check = None
                next_check = now + 15
            if pending_check is None and now >= next_check:
                pending_check = check_session()
            time.sleep(0.05)
        if pending_check is not None:
            require(wait_check(pending_check), "Pending role check failed")
        if state["reason"] is None:
            # Normal wrapper exit never proves separately sessioned reviewer absence.
            _, owned = collect()
            require(not owned, "Owned process survived normal wrapper exit")
            require(wrapper.returncode == 0, "Wrapper failed; inspect preserved local records")
            final_role_ok = wait_check(check_session())
            require(final_role_ok, "Final role check failed; no completion claim")
            terminal = json.loads((PACKET / "exit.json").read_text())
            require(terminal.get("classification") == "review_complete"
                    and terminal.get("reviewComplete") is True,
                    "Runner did not record a complete review")
    except BaseException as exc:
        fail("wrapper_or_preflight_failure")
        state["errors"].append(type(exc).__name__)
        log("outer_error", errorType=type(exc).__name__)
    finally:
        # A second interrupt must not skip local cleanup evidence.
        for sig in original_signals:
            signal.signal(sig, signal.SIG_IGN)
        remaining = []
        try:
            _, remaining = collect()
            if remaining:
                fail("process_cleanup_required")
                send_owned(signal.SIGTERM)
                stop = min(begin + 1798, time.monotonic() + 1)
                while time.monotonic() < stop:
                    _, remaining = collect()
                    if not remaining:
                        break
                    time.sleep(0.05)
                if remaining:
                    send_owned(signal.SIGSTOP)
                    send_owned(signal.SIGKILL)
                # Reap only within the original absolute total deadline.
                while time.monotonic() - begin < 1800:
                    if wrapper is not None:
                        wrapper.poll()
                    for item in checks:
                        # Individual check Popen objects may already be reaped; waitpid
                        # is nonblocking and cannot consume a second cleanup allowance.
                        try:
                            os.waitpid(item["pid"], os.WNOHANG)
                        except ChildProcessError:
                            pass
                    _, remaining = collect()
                    if not remaining:
                        break
                    time.sleep(0.05)
            if time.monotonic() - begin >= 1790:
                fail("outer_timeout")
            state["ownershipUnknown"] = state["ownershipUnknown"] or bool(remaining)
        except BaseException as exc:
            state["errors"].append("final_cleanup:" + type(exc).__name__)
            state["ownershipUnknown"] = True
        finally:
            finished.set()
            timer.join(timeout=1.5)
        elapsed = time.monotonic() - begin
        if elapsed >= 1800:
            state["reason"] = "outer_timeout"
        classification = state["reason"] or "review_complete_pending_root_adjudication"
        if state["ownershipUnknown"] and classification != "outer_timeout":
            classification = "process_ownership_or_cleanup_failure"
        result = {"classification": classification, "approvalInferred": False,
                  "started": started, "elapsedSeconds": elapsed,
                  "deadlineSeconds": 1800, "termAfterSeconds": 1790, "killAfterSeconds": 1798,
                  "wrapperPid": wrapper.pid if wrapper else None,
                  "wrapperExitCode": wrapper.poll() if wrapper else None,
                  "roleChecks": checks, "finalRoleVerified": final_role_ok,
                  "forcedCleanup": state["forced"], "termSentAtDeadline": state["termSent"],
                  "killSentAtDeadline": state["killSent"], "errors": state["errors"],
                  "remainingOwned": remaining, "ownershipUnknown": state["ownershipUnknown"],
                  "observedOwned": list(state["known"].values()), "observedGroups": state["groups"],
                  "noRetry": True, "localEvidenceAfterSupersessionAllowed": True,
                  "remoteWritesPerformed": False,
                  "limitation": "PID/start-time/ancestry and authenticated runner process record provide owned-group evidence; no system-wide socket census or hard-real-time OS guarantee. Terminal local evidence adds no reviewer execution allowance."}
        with (attempt / "outer-exit.json").open("x") as output:
            output.write(json.dumps(result, indent=2) + "\n")
        log("outer_terminal", classification=classification, remainingOwned=remaining)
        events.close()
        for sig, previous in original_signals.items():
            signal.signal(sig, previous)
    return 0 if classification == "review_complete_pending_root_adjudication" else 124 if classification == "outer_timeout" else 1


if __name__ == "__main__":
    sys.exit(main())
