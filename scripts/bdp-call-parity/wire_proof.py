#!/usr/bin/env python3
"""The defaults are the old code: a byte-for-byte comparison with the base commit.

See README.md. The client half runs the native command table through a bd built at the
base commit and through the new bd, each with no --protocol anywhere, and compares every
request. The server half asks a base-built server and the new one the same raw requests.
"""

import argparse
import collections
import json
import os
import re
import subprocess
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.request
import http.server

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import callable_cases as a_cases  # noqa: E402
import native_cases as b_cases  # noqa: E402
import part_a  # noqa: E402
from normalize import MINTED_ID, REVISION, TIMESTAMP, UUID, HEX_HASH  # noqa: E402
from stack import Root, Stack, run_bd  # noqa: E402

REQUEST_ID = re.compile(r'"request_id":"[^"]*"')
CURSOR = re.compile(r'("(?:next_)?cursor":")[^"]*(")|([?&]cursor=)[^&"\s]*')
BUILD_HASH = re.compile(r"\((?:[0-9a-f]{9}: ?)?[0-9a-f]{9,12}\)")
DROPPED_HEADERS = {"date", "content-length", "x-request-id"}
# What a base command says when it will not run over bd connect.
BASE_REFUSED = re.compile(r"not supported by the http backend|not expressible on the v0 wire|not available over bd connect"
                          r"|refuses? (?:over|against)|not served|cannot be served|nothing was closed|http backend"
                          r"|not implemented over|unknown flag|unknown command|the http store|expects? a direct")


class Masker:
    """Everything two independent runs may differ in, applied to raw text."""

    def __init__(self, extra):
        self.extra = extra
        self.minted = {}

    def text(self, value):
        if value is None:
            return None
        for old, new in self.extra:
            value = value.replace(old, new)
        value = TIMESTAMP.sub("<ts>", value)
        value = UUID.sub("<uuid>", value)
        value = REVISION.sub("<rev>", value)
        value = HEX_HASH.sub("<hash>", value)
        value = REQUEST_ID.sub('"request_id":"<request>"', value)
        value = CURSOR.sub(lambda m: (m.group(1) + "<cursor>" + m.group(2)) if m.group(1) else m.group(3) + "<cursor>", value)
        value = BUILD_HASH.sub("(<build>)", value)
        return MINTED_ID.sub(self._mint, value)

    def _mint(self, match):
        token = match.group(0)
        if token not in self.minted:
            self.minted[token] = f"<id{len(self.minted) + 1}>"
        return self.minted[token]

    def headers(self, pairs):
        return [[name, self.text(value)] for name, value in pairs if name.lower() not in DROPPED_HEADERS]


class WireRecorder:
    """Records every request and response whole, as text, and relays it unchanged."""

    def __init__(self, stack):
        self.entries = []
        self.lock = threading.Lock()
        recorder = self

        class Handler(http.server.BaseHTTPRequestHandler):
            protocol_version = "HTTP/1.1"

            def log_message(self, *_):
                pass

            def relay(self):
                length = int(self.headers.get("Content-Length") or 0)
                body = self.rfile.read(length) if length else b""
                forward = {k: v for k, v in self.headers.items()
                           if k.lower() not in ("host", "connection", "content-length", "accept-encoding")}
                request = urllib.request.Request(stack.url + self.path, data=body if length else None,
                                                 method=self.command, headers=forward)
                try:
                    with urllib.request.urlopen(request, timeout=120) as response:
                        status, reply, raw = response.status, list(response.headers.items()), response.read()
                except urllib.error.HTTPError as err:
                    status, reply, raw = err.code, list(err.headers.items()), err.read()
                self.send_response(status)
                for key, value in reply:
                    if key.lower() not in ("connection", "transfer-encoding", "date", "content-length"):
                        self.send_header(key, value)
                self.send_header("Content-Length", str(len(raw)))
                self.end_headers()
                self.wfile.write(raw)
                with recorder.lock:
                    recorder.entries.append({
                        "method": self.command, "target": self.path, "version": self.request_version,
                        "headers": list(self.headers.items()), "body": body.decode("utf-8", "replace"),
                        "status": status, "reply_headers": reply, "reply_body": raw.decode("utf-8", "replace"),
                    })

            do_GET = do_HEAD = do_OPTIONS = do_POST = do_PATCH = do_PUT = do_DELETE = relay

        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.url = f"http://127.0.0.1:{self.server.server_address[1]}"
        threading.Thread(target=self.server.serve_forever, daemon=True).start()

    def take(self):
        with self.lock:
            taken, self.entries = self.entries, []
        return taken

    def stop(self):
        self.server.shutdown()
        self.server.server_close()


class Side:
    """One workspace behind one bd serve, a recording proxy, and one bd connect."""

    def __init__(self, root, name):
        self.root = root
        self.stack = Stack(root, name, journal=True, serve=True)
        self.recorder = WireRecorder(self.stack)
        self.client = os.path.join(self.stack.dir, "client")
        os.makedirs(self.client)
        self.stack._git_init(self.client)
        self.extra = [(root.path, "<root>"), (self.recorder.url, "<server>"), (self.stack.url, "<server>"),
                      (self.recorder.url.split("//")[1], "<server>"), (self.stack.url.split("//")[1], "<server>")]
        self.connect_record = self.record_connect()

    def record_connect(self):
        ran = run_bd(self.root.bd, ["connect", self.recorder.url], self.client, self.stack.env(), False, 60, None)
        record = self.finish("connect", ["connect", "<server>"], "connect", "text", ran)
        with open(os.path.join(self.client, ".beads", "http_target.json")) as handle:
            record["sidecar"] = Masker(self.extra).text(handle.read())
        return record

    def seed(self, rows, before, gap=0):
        expanded = []
        for row in rows:
            if row[0] == "create" and "--parent" in row:
                at = row.index("--parent")
                expanded.append(row[:at] + row[at + 2:])
                expanded.append(["update", row[row.index("--id") + 1], "--parent", row[at + 1]])
            else:
                expanded.append(row)
        for row in expanded + before:
            if gap:
                time.sleep(gap)
            ran = self.stack.bd(*row)
            if ran.code != 0 and not (row[0] == "create" and "already exists" in ran.out + ran.err):
                raise RuntimeError(f"seed {row}: {ran.out} {ran.err}")

    def argv(self, case):
        out = []
        for word in case["argv"]:
            if word.startswith("$rev:"):
                word = json.loads(self.stack.bd("show", word[5:], "--json").out)[0]["revision"]
            out.append(word)
        return out

    def run_case(self, case):
        key = case["mode"] + " " + case["label"]
        try:
            self.seed(case["seed"], case["before"])
        except RuntimeError as err:
            return {"key": key, "label": case["label"], "mode": case["mode"], "phase": case["phase"],
                    "seed_failed": str(err)[:400], "exit": None, "stdout": "", "stderr": "", "requests": []}
        for name, body in case["files"].items():
            with open(os.path.join(self.client, name), "w") as handle:
                handle.write(body)
        if case.get("wait"):
            time.sleep(case["wait"])
        argv = self.argv(case) + ([] if case["plain"] else ["--json"])
        self.recorder.take()
        ran = run_bd(self.root.bd, argv, self.client, self.stack.env(case["actor"]), False, 60, case["stdin"])
        return self.finish(key, case["argv"], case["phase"], case["mode"], ran, case["label"])

    def finish(self, key, argv, phase, mode, ran, label=None):
        masker = Masker(self.extra)
        requests = []
        for entry in self.recorder.take():
            requests.append({
                "request": {"method": entry["method"], "target": masker.text(entry["target"]), "version": entry["version"],
                            "headers": masker.headers(entry["headers"]), "body": masker.text(entry["body"])},
                "response": {"status": entry["status"], "headers": masker.headers(entry["reply_headers"]),
                             "body": masker.text(entry["reply_body"])},
            })
        return {"key": key, "label": label or " ".join(argv), "argv": argv, "phase": phase, "mode": mode, "exit": ran.code,
                "stdout": masker.text(ran.out), "stderr": masker.text(ran.err), "requests": requests}

    def stop(self):
        self.recorder.stop()
        self.root.stop(self.stack.serve_proc)
        self.root.stop(self.stack.dolt_proc)


def client_side(bd, dolt, out, only, log):
    wanted = re.compile(only) if only else None

    def selected(group):
        return [c for c in group if not wanted or wanted.search(" ".join(c["argv"]))]

    def text_of(group):
        return [t for t in (b_cases.as_text(c) for c in group) if t]

    root = Root(os.path.abspath(bd), os.path.abspath(dolt), serve_protocol=None, connect_protocol=None)
    records = []
    try:
        side = Side(root, "wire-read")
        records.append(side.connect_record)
        side.seed(b_cases.READ_SEED, b_cases.READ_BEFORE, gap=1.1)
        for case in selected(b_cases.reads()):
            records.append(side.run_case(case))
        for case in text_of(selected(b_cases.reads())):
            records.append(side.run_case(case))
        side.stop()
        log(f"reads done, {len(records)}")
        for name, group in (("wire-write", selected(b_cases.writes())), ("wire-write-text", text_of(selected(b_cases.writes())))):
            side = Side(root, name)
            for case in group:
                records.append(side.run_case(case))
            side.stop()
            log(f"{name} done, {len(records)}")
        both = selected(b_cases.global_cases())
        for index, case in enumerate(both + text_of(both)):
            side = Side(root, f"wire-global-{index}")
            records.append(side.run_case(case))
            side.stop()
        log(f"global done, {len(records)}")
    finally:
        root.reap()
    with open(out, "w") as handle:
        json.dump(records, handle, indent=1)


def first_line(text):
    lines = [line for line in (text or "").splitlines() if line.strip()]
    return lines[0][:300] if lines else ""


def request_lines(record):
    return [f"{r['request']['method']} {r['request']['target']}" + (f"  {r['request']['body'][:300]}" if r["request"]["body"] else "")
            for r in record["requests"]]


def classify(base, new):
    """One command's verdict, and what differed."""
    if base.get("seed_failed") or new.get("seed_failed"):
        return ("same" if base.get("seed_failed") == new.get("seed_failed") else "unexplained"), {"seed": [base.get("seed_failed"), new.get("seed_failed")]}
    wire_base = [r["request"] for r in base["requests"]]
    wire_new = [r["request"] for r in new["requests"]]
    answers_base = [r["response"] for r in base["requests"]]
    answers_new = [r["response"] for r in new["requests"]]
    output_same = base["exit"] == new["exit"] and base["stdout"] == new["stdout"] and base["stderr"] == new["stderr"]
    if wire_base == wire_new and answers_base == answers_new and output_same:
        return "same", {}
    detail = {"base_exit": base["exit"], "new_exit": new["exit"],
              "base_says": first_line(base["stderr"]) or first_line(base["stdout"]),
              "new_says": first_line(new["stderr"]) or first_line(new["stdout"]),
              "base_requests": request_lines(base), "new_requests": request_lines(new)}
    base_refused = base["exit"] != 0 or BASE_REFUSED.search(base["stderr"] + base["stdout"]) is not None
    if wire_base != wire_new:
        detail["first_difference"] = first_difference(wire_base, wire_new)
        if sorted(map(json.dumps, wire_base)) == sorted(map(json.dumps, wire_new)):
            detail["order_only"] = True
        if base_refused and (new["exit"] == 0 or new["exit"] != base["exit"] or BASE_REFUSED.search(base["stderr"] + base["stdout"])):
            return "allowed", detail
        return "unexplained", detail
    if answers_base != answers_new:
        detail["first_answer_difference"] = first_difference(answers_base, answers_new)
        if base_refused and base["exit"] != 0 and new["exit"] == 0:
            return "allowed", detail
        return "state", detail
    detail["base_stdout"], detail["new_stdout"] = base["stdout"][:800], new["stdout"][:800]
    detail["base_stderr"], detail["new_stderr"] = base["stderr"][:800], new["stderr"][:800]
    return "output", detail


def first_difference(left, right):
    for index, (a, b) in enumerate(zip(left, right)):
        if a != b:
            return {"index": index, "base": a, "new": b}
    return {"index": min(len(left), len(right)), "base_count": len(left), "new_count": len(right)}


def compare_clients(base_path, new_path, out):
    with open(base_path) as handle:
        base = json.load(handle)
    with open(new_path) as handle:
        new = json.load(handle)
    results = []
    by_key = collections.defaultdict(list)
    for record in new:
        by_key[record["key"]].append(record)
    for record in base:
        twin = by_key[record["key"]].pop(0) if by_key[record["key"]] else None
        if twin is None:
            results.append({"key": record["key"], "verdict": "unexplained", "detail": {"missing": "new"}})
            continue
        verdict, detail = classify(record, twin)
        if record.get("sidecar") != twin.get("sidecar"):
            verdict, detail = "unexplained", {"sidecar": [record.get("sidecar"), twin.get("sidecar")]}
        results.append({"key": record["key"], "verdict": verdict, "requests": len(record.get("requests", [])), "detail": detail})
    counts = collections.Counter(r["verdict"] for r in results)
    with open(os.path.join(out, "client-comparison.json"), "w") as handle:
        json.dump({"counts": counts, "requests_compared": sum(len(r.get("requests", [])) for r in base), "results": results}, handle, indent=1)
    return results, counts


STATIC_PROBES = [
    ("GET", "/"), ("HEAD", "/"), ("OPTIONS", "/"), ("POST", "/"), ("DELETE", "/"),
    ("GET", "/bdp"), ("HEAD", "/bdp"), ("OPTIONS", "/bdp"), ("POST", "/bdp"), ("PUT", "/bdp"),
    ("GET", "/bdp/"), ("GET", "/bdp/beads"), ("POST", "/bdp/beads"), ("GET", "/bdp/links"), ("OPTIONS", "/bdp/links"),
    ("GET", "/bdp/nonsense"), ("GET", "/bdp/call"), ("POST", "/bdp/call"), ("HEAD", "/bdp/call"),
    ("GET", "/bdp/call/issue/show"), ("POST", "/bdp/call/issue/show"), ("POST", "/bdp/call/nosuch"),
    ("GET", "/sw-fix-r1"), ("HEAD", "/sw-fix-r1"), ("POST", "/sw-fix-r1"), ("OPTIONS", "/sw-fix-r1"), ("GET", "/nope"), ("GET", "/no/such/route"),
    ("GET", "/v0"), ("GET", "/v0/"), ("GET", "/v0/nonsense"), ("POST", "/v0/nonsense"), ("GET", "/v0/beads"), ("GET", "/healthz"),
    ("POST", "/healthz"), ("GET", "/v0/beads/context"), ("HEAD", "/v0/beads/context"), ("POST", "/v0/beads/context"),
    ("GET", "/v0/beads/issues/sw-fix-none"), ("GET", "/v0/beads/issues/sw-fix-r1:nosuchverb"), ("POST", "/v0/beads/issues/sw-fix-r1:nosuchverb"),
    ("GET", "/v0/beads/events:watch?since=999999"),
]


def raw_request(stack, method, target, headers=None, body=None):
    request = urllib.request.Request(stack.url + target, data=body, method=method, headers=headers or {})
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            return response.status, list(response.headers.items()), response.read()
    except urllib.error.HTTPError as err:
        return err.code, list(err.headers.items()), err.read()


def masked_answer(stack, reply):
    status, headers, raw = reply
    masker = Masker([(stack.root.path, "<root>"), (stack.url, "<server>"), (stack.url.split("//")[1], "<server>")])
    return {"status": status, "headers": masker.headers(headers), "body": masker.text(raw.decode("utf-8", "replace"))}


def lockstep_probe(base, new, label, method, target, body=None, headers=None, results=None):
    answers = [masked_answer(stack, raw_request(stack, method, target, headers, body)) for stack in (base, new)]
    results.append({"label": label, "method": method, "target": target, "same": answers[0] == answers[1],
                    **({} if answers[0] == answers[1] else {"base": answers[0], "new": answers[1]})})


def server_proof(base_bd, new_bd, dolt, out, log):
    base_root = Root(os.path.abspath(base_bd), os.path.abspath(dolt), serve_protocol=None, connect_protocol=None)
    new_root = Root(os.path.abspath(new_bd), os.path.abspath(dolt), serve_protocol=None, connect_protocol=None)
    results = []
    try:
        base = Stack(base_root, "server-base", journal=True, serve=True)
        new = Stack(new_root, "server-new", journal=True, serve=True)
        part_a.seed([base, new], a_cases.READ_SEED, a_cases.READ_BEFORE)
        invocation = b'{"arguments":{"id":"sw-fix-r1"}}'
        for method, target in STATIC_PROBES:
            lockstep_probe(base, new, "probe", method, target, results=results)
        for label, target in (("probe: an invocation", "/bdp/call/issue/show"), ("probe: an invocation", "/bdp/call/issue/context")):
            lockstep_probe(base, new, label, "POST", target, body=invocation,
                           headers={"Content-Type": a_cases.__dict__.get("CALLABLE_MEDIA", "application/vnd.bdp.v0.1+json"),
                                    "Accept": "application/vnd.bdp.v0.1+json"}, results=results)
        lockstep_probe(base, new, "probe: conditional", "GET", "/", headers={"If-None-Match": "*"}, results=results)
        lockstep_probe(base, new, "probe: wrong host", "GET", "/", headers={"Host": "rebind.example"}, results=results)
        log(f"server: probes done, {len(results)}")
        for name, args, _ in a_cases.read_cases():
            form = part_a.v0_request(name, a_cases.with_defaults(name, args))
            if form is None:
                continue
            method, path, body = form
            lockstep_probe(base, new, f"read {name}", method, path, results=results,
                           body=None if body is None else json.dumps(body).encode(),
                           headers=None if body is None else {"Content-Type": "application/json"})
        log(f"server: reads done, {len(results)}")
        for case in a_cases.write_cases():
            name = case["callable"]
            args = case["v0_args"] if case["v0_args"] is not None else case["args"]
            label = case["label"] or json.dumps(args, sort_keys=True)[:100]
            part_a.seed([base, new], case["seed"], case["before"])
            for id_ in case["sql"]:
                for stack in (base, new):
                    stack.bd("sql", f"UPDATE issues SET is_template=1 WHERE id='{id_}'", check=True)
            if case["wait"]:
                time.sleep(case["wait"])
            answers = []
            for stack in (base, new):
                form = part_a.v0_request(name, a_cases.with_defaults(name, part_a.interpolate(stack, args)))
                if form is None:
                    answers.append(None)
                    continue
                method, path, body = form
                answers.append(masked_answer(stack, raw_request(
                    stack, method, path, None if body is None else {"Content-Type": "application/json"},
                    None if body is None else json.dumps(body).encode())))
            states = [Masker([]).text(json.dumps(part_a.state_of(stack), sort_keys=True)) for stack in (base, new)]
            same = answers[0] == answers[1] and states[0] == states[1]
            entry = {"label": f"write {name}: {label}", "same": same}
            if not same:
                entry.update({"base": answers[0], "new": answers[1], "state_same": states[0] == states[1]})
            results.append(entry)
        log(f"server: writes done, {len(results)}")
        # The credential posture: what asks for one must be what asked for one.
        for stack in (base, new):
            path = os.path.join(stack.dir, "tokens")
            with open(path, "w") as handle:
                handle.write("a-token\n")
            os.chmod(path, 0o600)
            stack.serve_args = ["--auth-token-file", path]
            stack.restart_serve()
        for method, target in STATIC_PROBES:
            lockstep_probe(base, new, "probe, auth on, no token", method, target, results=results)
            lockstep_probe(base, new, "probe, auth on, token", method, target, headers={"Authorization": "Bearer a-token"}, results=results)
        log(f"server: auth probes done, {len(results)}")
    finally:
        base_root.reap()
        new_root.reap()
    with open(os.path.join(out, "server-comparison.json"), "w") as handle:
        json.dump({"compared": len(results), "different": sum(not r["same"] for r in results), "results": results}, handle, indent=1)
    return results


def write_reports(out, results, counts, server):
    lines = ["# Default settings against the base commit", "",
             f"Client half: {dict(counts)}.",
             f"Server half: {len(server)} asked in lockstep, {sum(not r['same'] for r in server)} different.", "",
             "## Commands the base refused or got less far with (`allowed`)", ""]
    for r in results:
        if r["verdict"] == "allowed":
            d = r["detail"]
            lines += [f"- `{r['key']}`: base exit {d['base_exit']}, new exit {d['new_exit']}. Base said: {d['base_says'][:200]}",
                      f"  new sends: {'; '.join(x[:120] for x in d['new_requests'][:6])}"]
    for verdict, title in (("state", "Same requests, an answer moved"), ("output", "Same wire, different words"), ("unexplained", "UNEXPLAINED")):
        lines += ["", f"## {title} (`{verdict}`)", ""]
        for r in results:
            if r["verdict"] == verdict:
                d = r["detail"]
                what = d.get("first_difference") or d.get("first_answer_difference") or {
                    "base": d.get("base_says"), "new": d.get("new_says")}
                lines.append(f"- `{r['key']}`: exit {d.get('base_exit')} / {d.get('new_exit')}: {json.dumps(what)[:500]}")
    lines += ["", "## Server half, every difference", ""]
    for r in server:
        if not r["same"]:
            lines.append(f"- {r['label']} {r.get('method', '')} {r.get('target', '')}: {json.dumps(r.get('base'))[:300]} / {json.dumps(r.get('new'))[:300]}")
    with open(os.path.join(out, "summary.md"), "w") as handle:
        handle.write("\n".join(lines) + "\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--base", help="a bd built at the base commit")
    parser.add_argument("--new", help="the bd under test")
    parser.add_argument("--dolt", required=True)
    parser.add_argument("--out", required=True)
    parser.add_argument("--only", help="the client half runs only commands matching this regex")
    parser.add_argument("--skip-client", action="store_true", help="compare the client records already in --out")
    parser.add_argument("--skip-server", action="store_true", help="keep the server comparison already in --out")
    parser.add_argument("--side", help=argparse.SUPPRESS)
    args = parser.parse_args()
    os.makedirs(args.out, exist_ok=True)

    def log(message):
        print(message, flush=True)

    if args.side:
        client_side(args.new, args.dolt, args.side, args.only, log)
        return

    base_records = os.path.join(args.out, "client-base.json")
    new_records = os.path.join(args.out, "client-new.json")
    if not args.skip_client:
        children = []
        for bd, path, name in ((args.base, base_records, "base"), (args.new, new_records, "new")):
            command = [sys.executable, os.path.abspath(__file__), "--side", path, "--new", bd, "--dolt", args.dolt, "--out", args.out]
            if args.only:
                command += ["--only", args.only]
            logfile = open(os.path.join(args.out, f"client-{name}.log"), "w")
            children.append((name, subprocess.Popen(command, stdout=logfile, stderr=subprocess.STDOUT)))
        for name, child in children:
            if child.wait() != 0:
                sys.exit(f"the {name} side failed; see client-{name}.log")
    results, counts = compare_clients(base_records, new_records, args.out)
    log(f"client: {dict(counts)}")
    server_path = os.path.join(args.out, "server-comparison.json")
    if args.skip_server:
        with open(server_path) as handle:
            server = json.load(handle)["results"]
    else:
        server = server_proof(args.base, args.new, args.dolt, args.out, log)
    log(f"server: {sum(not r['same'] for r in server)} of {len(server)} different")
    write_reports(args.out, results, counts, server)
    if counts.get("unexplained"):
        sys.exit(f"{counts['unexplained']} unexplained; see {args.out}/summary.md")


if __name__ == "__main__":
    main()
