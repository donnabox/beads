"""A recording proxy between `bd connect --protocol bdp` and `bd serve`.

On BDP, `bd connect` speaks only /bdp/call. This records every request the client really
makes during Part B, and asks of each: is it an invocation at all, does its Callable
declare every argument it sent, and, for a read, does v0 answer the same right now?

A request that is not POST /bdp/call/NAME is what the move exists to end, so it fails
the run (parity.py): see non_call_requests. So does one with `actor` or `author` in its
arguments, which no declaration has, and one that leaves a defaulted parameter to the
service, because the client means v0's behavior by an omission and the default is the
CLI's: see contract_breaches.
"""

import http.server
import json
import threading
import urllib.error
import urllib.parse
import urllib.request

import callable_cases as cases
import part_a


CALL_MOUNT = "/bdp/call/"


class Recorder:
    def __init__(self, stack, declared):
        self.stack = stack
        self.declared = declared
        self.entries = []
        self.lock = threading.Lock()
        recorder = self

        class Handler(http.server.BaseHTTPRequestHandler):
            protocol_version = "HTTP/1.1"

            def log_message(self, *_):
                pass

            def relay(self):
                length = int(self.headers.get("Content-Length") or 0)
                body = self.rfile.read(length) if length else None
                headers = {k: v for k, v in self.headers.items()
                           if k.lower() not in ("host", "connection", "content-length", "accept-encoding")}
                request = urllib.request.Request(stack.url + self.path, data=body, method=self.command, headers=headers)
                try:
                    with urllib.request.urlopen(request, timeout=120) as response:
                        status, reply_headers, raw = response.status, dict(response.headers), response.read()
                except urllib.error.HTTPError as err:
                    status, reply_headers, raw = err.code, dict(err.headers), err.read()
                self.send_response(status)
                for key, value in reply_headers.items():
                    if key.lower() not in ("connection", "transfer-encoding", "date", "content-length"):
                        self.send_header(key, value)
                self.send_header("Content-Length", str(len(raw)))
                self.end_headers()
                self.wfile.write(raw)
                recorder.observe(self.command, self.path, dict(self.headers), body, status, reply_headers, raw)

            do_GET = do_HEAD = do_OPTIONS = do_POST = do_PATCH = do_PUT = do_DELETE = relay

        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.url = f"http://127.0.0.1:{self.server.server_address[1]}"
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()

    def observe(self, method, target, headers, body, status, reply_headers, raw):
        parsed = urllib.parse.urlsplit(target)
        name = parsed.path[len(CALL_MOUNT):] if parsed.path.startswith(CALL_MOUNT) and method == "POST" else None
        entry = {"method": method, "path": parsed.path, "callable": name, "status": status,
                 "query": sorted(urllib.parse.parse_qs(parsed.query, keep_blank_values=True)),
                 "headers": sorted(k.lower() for k in headers)}
        if name:
            try:
                invocation = json.loads(body) if body else {}
            except ValueError:
                invocation = {}
            arguments = invocation.get("arguments", {}) if isinstance(invocation, dict) else {}
            entry["arguments"] = sorted(arguments)
            entry["undeclared"] = sorted(set(arguments) - self.declared.get(name, set()))
            entry["actor_in_arguments"] = sorted({"actor", "author"} & set(arguments))
            entry["left_to_default"] = sorted(set(cases.DEFAULTS.get(name, {})) - set(arguments))
            if isinstance(invocation, dict) and invocation.get("revisionProvenance") is not None:
                entry["provenance"] = invocation["revisionProvenance"]
            if name in cases.ROUTES and (cases.ROUTES[name][0] == "GET" or name == "issue/get-many"):
                # A read changes nothing, so v0 can be asked the same thing at once.
                v0_headers = {k: v for k, v in headers.items()
                              if k.lower() in ("bd-project-id", "bd-wire-revision", "user-agent", "authorization")}
                try:
                    v0_form = part_a.v0_request(name, arguments)
                    if v0_form is None:
                        # The mount refused it for a missing path argument, before v0.
                        entry["replay_problems"] = [] if status != 200 else ["no path argument, yet answered 200"]
                    else:
                        replay = self.stack.http(v0_form[0], v0_form[1], v0_form[2], headers=v0_headers)
                        entry["replay_status"] = replay.status
                        entry["replay_problems"] = part_a.compare(name, replay, _Reply(status, raw, reply_headers), same_server=True)
                except OSError as err:
                    entry["replay_problems"] = [f"replay failed: {err}"]
        with self.lock:
            self.entries.append(entry)

    def stop(self):
        self.server.shutdown()
        self.server.server_close()


class _Reply:
    def __init__(self, status, raw, headers):
        self.status = status
        self.raw = raw
        self.headers = headers
        try:
            self.json = json.loads(raw) if raw else None
        except ValueError:
            self.json = {"unparsed": raw[:200].decode(errors="replace")}


def declared_parameters(stack):
    """Callable -> every argument its declaration publishes, from GET /bdp/call."""
    document = stack.http("GET", "/bdp/call").json
    declared = {}
    for path, operations in document.get("paths", {}).items():
        if "/bdp/call/" not in path or "post" not in operations:
            continue
        name = path.split("/bdp/call/")[1]
        schema = operations["post"]["requestBody"]["content"][next(iter(operations["post"]["requestBody"]["content"]))]["schema"]
        declared[name] = set(_arguments_of(schema, document))
    return declared


def _arguments_of(schema, document):
    while "$ref" in schema:
        at = document
        for part in schema["$ref"].split("/")[1:]:
            at = at[part]
        schema = at
    arguments = schema.get("properties", {}).get("arguments")
    if arguments is None:
        for branch in schema.get("allOf", []) + schema.get("oneOf", []):
            names = _arguments_of(branch, document)
            if names:
                return names
        return []
    while "$ref" in arguments:
        at = document
        for part in arguments["$ref"].split("/")[1:]:
            at = at[part]
        arguments = at
    return arguments.get("properties", {}).keys()


def non_call_requests(wire):
    """The keys of summarize()'s rows that are not invocations: the run's failure."""
    return [key for key in wire if key.startswith(NON_CALL)]


def contract_breaches(wire):
    """What the client must never send: the run's failure, one line each."""
    out = []
    for key, row in wire.items():
        if row.get("actor_in_arguments"):
            out.append(f"{key}: {' '.join(row['actor_in_arguments'])} in arguments")
        if row.get("left_to_default"):
            out.append(f"{key}: left {' '.join(row['left_to_default'])} to the service's default")
    return out


NON_CALL = "(not an invocation) "


def summarize(entries):
    """One row per Callable the client invoked, and one per anything else it sent."""
    rows = {}
    for entry in entries:
        key = entry["callable"] or f"{NON_CALL}{entry['method']} {entry['path']}"
        row = rows.setdefault(key, {"requests": 0, "query": set(), "arguments": set(), "headers": set(), "undeclared": set(),
                                    "actor_in_arguments": set(), "left_to_default": set(), "with_provenance": 0,
                                    "replayed": 0, "replay_failures": []})
        row["requests"] += 1
        row["query"] |= set(entry["query"])
        row["arguments"] |= set(entry.get("arguments", []))
        row["headers"] |= set(entry["headers"])
        row["undeclared"] |= set(entry.get("undeclared", []))
        row["actor_in_arguments"] |= set(entry.get("actor_in_arguments", []))
        row["left_to_default"] |= set(entry.get("left_to_default", []))
        row["with_provenance"] += 1 if "provenance" in entry else 0
        if "replay_problems" in entry:
            row["replayed"] += 1
            if entry["replay_problems"]:
                row["replay_failures"].append({"path": entry["path"], "problems": entry["replay_problems"][:3]})
    return {key: {k: sorted(v) if isinstance(v, set) else (v[:5] if isinstance(v, list) else v) for k, v in row.items()}
            for key, row in sorted(rows.items())}
