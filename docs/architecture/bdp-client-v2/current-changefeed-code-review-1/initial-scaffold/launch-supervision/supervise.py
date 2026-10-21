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
             "lastRows": [], "ownershipUnknown": False, "direct": {},
             "termAttemptCount": 0, "killAttemptCount": 0, "authenticatedReviewers": {},
             "unanchoredGroupsCurrent": {}, "unanchoredGroupsObserved": {}}
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
        remaining = begin + 1800 - time.monotonic()
        if remaining <= 0:
            raise subprocess.TimeoutExpired("process inventory after absolute boundary", 0)
        raw = subprocess.check_output(
            ["ps", "-axo", "pid=,ppid=,pgid=,state=,lstart=,comm="], text=True,
            timeout=min(0.4, remaining))
        rows = {}
        for line in raw.splitlines():
            parts = line.split(None, 9)
            if len(parts) != 10:
                continue
            pid, ppid, pgid = map(int, parts[:3])
            rows[pid] = {"pid": pid, "ppid": ppid, "pgid": pgid,
                         "state": parts[3], "start": " ".join(parts[4:9]), "comm": parts[9]}
        return rows

    def register_direct(proc, kind):
        # Register the unreaped Popen handle before any inventory, close or file write.
        # All waits/polls use the same lock, so fallback cannot race PID reuse.
        with state_lock:
            state["direct"][proc.pid] = {"proc": proc, "kind": kind}
            state["groups"][proc.pid] = kind

    def poll_direct(proc):
        with state_lock:
            return proc.poll()

    def collect():
        rows = inventory()
        with state_lock:
            owned = {pid for pid, old in state["known"].items()
                     if pid in rows and rows[pid]["start"] == old["start"]}
            # Include every registered direct child, including session checks.
            for pid, direct in state["direct"].items():
                if direct["proc"].returncode is None and pid in rows:
                    row = rows[pid]
                    require(row["ppid"] == os.getpid() and row["pgid"] == pid,
                            "Registered direct-child identity changed")
                    owned.add(pid)
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
                        authenticated = {"pid": pid, "pgid": row["pgid"], "start": row["start"],
                                         "recordStarted": record["started"], "argv": record["argv"]}
                        if row["state"].startswith("Z"):
                            # Zombie comm may be <defunct>. Only a reviewer previously
                            # authenticated while live can retain ownership in that state.
                            require(state["authenticatedReviewers"].get(pid) == authenticated,
                                    "Unbound or changed zombie reviewer identity")
                        else:
                            require(Path(row["comm"]).name == "claude", "Reviewer executable identity changed")
                            state["authenticatedReviewers"][pid] = authenticated
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
            # A zombie leader may leave live children reparented in its same
            # authenticated group. Include the complete group before deciding
            # whether there is anything signalable; zombie presence is not absence.
            authenticated_groups = {rows[pid]["pgid"] for pid in owned
                                    if rows[pid]["pgid"] != os.getpgrp()}
            owned.update(pid for pid, row in rows.items() if row["pgid"] in authenticated_groups)
            # A reaped leader and never-sampled reparented child may leave no
            # surviving PID/start anchor. Census EVERY previously authenticated
            # PGID anyway. Numeric PGID history alone never authorizes a signal:
            # surviving unanchored membership is unknown/failure, not absence.
            unanchored = {}
            for pgid in set(state["groups"]) - authenticated_groups:
                members = [row for row in rows.values() if row["pgid"] == pgid]
                if members:
                    unanchored[pgid] = members
                    state["unanchoredGroupsObserved"].setdefault(pgid, members)
            state["unanchoredGroupsCurrent"] = unanchored
            if unanchored:
                state["ownershipUnknown"] = True
                if state["reason"] is None:
                    state["reason"] = "known_group_membership_unanchored"
            for pid in owned:
                row = rows[pid]
                state["known"][pid] = row
                if row["pgid"] != os.getpgrp():
                    state["groups"].setdefault(row["pgid"], "observed-descendant")
            state["lastRows"] = [rows[pid] for pid in sorted(owned)]
            return rows, list(state["lastRows"])

    def send_owned(sig):
        # Fresh PID/start-time/ancestry verification for non-child groups. A failed
        # inventory never prevents safe signaling of registered unreaped direct children.
        complete = True
        try:
            rows, owned = collect()
            with state_lock:
                if state["unanchoredGroupsCurrent"]:
                    complete = False
        except Exception as exc:
            rows = None
            owned = []
            complete = False
            with state_lock:
                error = "signal_inventory:" + type(exc).__name__
                if error not in state["errors"]:
                    state["errors"].append(error)
                state["ownershipUnknown"] = True
        # A fresh all-zombie group needs reaping/absence, not TERM/STOP/KILL.
        valid_groups = {row["pgid"] for row in owned
                        if row["pgid"] != os.getpgrp() and not row["state"].startswith("Z")}
        signaled = []
        with state_lock:
            for pid, direct in state["direct"].items():
                # returncode None is safe here: this parent retains the Popen handle,
                # no other waitpid occurs, and poll/reap is excluded by this lock.
                if direct["proc"].returncode is None:
                    if rows is not None:
                        members = [row for row in rows.values() if row["pgid"] == pid]
                        if members and all(row["state"].startswith("Z") for row in members):
                            continue
                    try:
                        if os.getpgid(pid) == pid:
                            valid_groups.add(pid)
                    except ProcessLookupError:
                        pass
            for pgid in sorted(valid_groups):
                if time.monotonic() >= begin + 1800:
                    complete = False
                    break
                try:
                    os.killpg(pgid, sig)
                except ProcessLookupError:
                    continue
                except OSError as exc:
                    complete = False
                    state["errors"].append("signal:" + type(exc).__name__)
                    continue
                signaled.append(pgid)
                state["forced"] = True
        for pgid in signaled:
            log("signal", pgid=pgid, signal=int(sig))
        log("signal_pass", signal=int(sig), complete=complete, signaledGroups=signaled)
        return complete

    def watchdog():
        # Independent of final remote checks, wrapper waits or source hashing.
        # Failed inventory/signaling leaves success false and is retried inside1800.
        while not finished.is_set():
            elapsed = time.monotonic() - begin
            if elapsed >= 1800:
                fail("outer_timeout")
                with state_lock:
                    state["ownershipUnknown"] = state["ownershipUnknown"] or bool(state["lastRows"])
                return
            try:
                if elapsed >= 1790:
                    fail("outer_timeout")
                    with state_lock:
                        term = not state["termSent"]
                        if term:
                            state["termAttemptCount"] += 1
                    if term:
                        success = send_owned(signal.SIGTERM)
                        with state_lock:
                            state["termSent"] = success
                if elapsed >= 1798:
                    with state_lock:
                        kill = not state["killSent"]
                        if kill:
                            state["killAttemptCount"] += 1
                    if kill:
                        # A failed STOP pass must not skip the KILL pass or its retry.
                        send_owned(signal.SIGSTOP)
                        success = send_owned(signal.SIGKILL)
                        with state_lock:
                            state["killSent"] = success
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
        register_direct(proc, "session-check")
        out.close()
        err.close()
        checks.append({"pid": proc.pid, "startedElapsed": time.monotonic() - begin})
        return proc

    def wait_check(proc):
        while poll_direct(proc) is None:
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
        register_direct(wrapper, "wrapper")
        state["wrapperPid"] = wrapper.pid
        stdout.close()
        stderr.close()
        log("wrapper_started", pid=wrapper.pid, pgid=wrapper.pid)
        pending_check = None
        next_check = time.monotonic() + 15
        next_sample = 0
        while poll_direct(wrapper) is None:
            now = time.monotonic()
            _, owned = collect()
            if now >= next_sample:
                log("process_sample", owned=owned)
                next_sample = now + 1
            if state["reason"] is not None or now - begin >= 1790:
                break
            if pending_check is not None and poll_direct(pending_check) is not None:
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
            require(not owned and not state["unanchoredGroupsCurrent"],
                    "Owned or unanchored known-group process survived normal wrapper exit")
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
                        poll_direct(wrapper)
                    for direct in list(state["direct"].values()):
                        poll_direct(direct["proc"])
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
            # Inventory failure must not cancel cleanup with direct children alive.
            # Keep retrying safe escalation within the existing absolute boundary.
            while time.monotonic() < begin + 1800:
                send_owned(signal.SIGSTOP)
                send_owned(signal.SIGKILL)
                for direct in list(state["direct"].values()):
                    poll_direct(direct["proc"])
                try:
                    _, remaining = collect()
                    if not remaining:
                        break
                except Exception:
                    pass
                time.sleep(0.05)
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
                  "wrapperExitCode": poll_direct(wrapper) if wrapper else None,
                  "roleChecks": checks, "finalRoleVerified": final_role_ok,
                  "forcedCleanup": state["forced"], "termSentAtDeadline": state["termSent"],
                  "killSentAtDeadline": state["killSent"],
                  "termSignalPassAttempts": state["termAttemptCount"],
                  "killSignalPassAttempts": state["killAttemptCount"],
                  "signalStateMeaning": "Sent flags mean a complete successful inventory/signal pass, not a scheduled attempt; logs preserve per-group signals and failed passes.",
                  "errors": state["errors"],
                  "remainingOwned": remaining, "ownershipUnknown": state["ownershipUnknown"],
                  "observedOwned": list(state["known"].values()), "observedGroups": state["groups"],
                  "authenticatedReviewers": list(state["authenticatedReviewers"].values()),
                  "unanchoredKnownGroupsAtLastInventory": state["unanchoredGroupsCurrent"],
                  "unanchoredKnownGroupsObserved": state["unanchoredGroupsObserved"],
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
