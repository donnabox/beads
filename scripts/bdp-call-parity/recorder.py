"""A recording proxy between `bd connect` and `bd serve`.

The operator plans to move the client from /v0 to /bdp/call. This records every
request the client really makes during Part B, and asks of each: which Callable
would carry it, does that Callable declare every parameter, and, for a read, does
the Callable answer it the same right now?
"""

import http.server
import json
import threading
import urllib.error
import urllib.parse
import urllib.request

import callable_cases as cases
import part_a


def match_route(method, path):
    """The (Callable, path arguments) a v0 request corresponds to, or (None, {})."""
    segments = path.split("/")
    for name, (route_method, template) in cases.ROUTES.items():
        if route_method != method or name == "issue/prune":
            continue
        pattern = template.split("/")
        if len(pattern) != len(segments):
            continue
        args = {}
        for want, got in zip(pattern, segments):
            if "{" in want:
                param = want[want.index("{") + 1:want.index("}")]
                suffix = want[want.index("}") + 1:]
                if not got.endswith(suffix):
                    break
                args[param] = urllib.parse.unquote(got[:len(got) - len(suffix)] if suffix else got)
            elif want != got:
                break
        else:
            return name, args
    return None, {}


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
                recorder.observe(self.command, self.path, dict(self.headers), body, status, raw)

            do_GET = do_POST = do_PATCH = do_PUT = do_DELETE = relay

        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.url = f"http://127.0.0.1:{self.server.server_address[1]}"
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()

    def observe(self, method, target, headers, body, status, raw):
        parsed = urllib.parse.urlsplit(target)
        query = urllib.parse.parse_qs(parsed.query, keep_blank_values=True)
        try:
            members = json.loads(body) if body else {}
        except ValueError:
            members = {}
        name, args = match_route(method, parsed.path)
        if name in ("issue/purge",) and members.get("tier") == "durable":
            name = "issue/prune"
        entry = {"method": method, "path": parsed.path, "callable": name, "status": status,
                 "query": sorted(query), "body": sorted(members) if isinstance(members, dict) else ["(not an object)"],
                 "headers": sorted(k.lower() for k in headers)}
        if name:
            undeclared = sorted((set(query) | set(members)) - self.declared.get(name, set()))
            entry["undeclared"] = undeclared
            arguments = dict(args)
            arguments.update({k: (v[0] if len(v) == 1 else v) for k, v in query.items()})
            if isinstance(members, dict):
                arguments.update(members)
            if method == "GET" or name == "issue/get-many":
                # A read changes nothing, so it can be asked again through the mount.
                call_headers = {k: v for k, v in headers.items()
                                if k.lower() in ("bd-project-id", "bd-wire-revision", "user-agent", "authorization")}
                try:
                    replay = self.stack.call(name, arguments, headers=call_headers)
                    v0_reply = _Reply(status, raw)
                    entry["replay_status"] = replay.status
                    entry["replay_problems"] = part_a.compare(name, v0_reply, replay, same_server=True)
                except OSError as err:
                    entry["replay_problems"] = [f"replay failed: {err}"]
        with self.lock:
            self.entries.append(entry)

    def stop(self):
        self.server.shutdown()
        self.server.server_close()


class _Reply:
    def __init__(self, status, raw):
        self.status = status
        self.raw = raw
        self.headers = {}
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


def summarize(entries):
    """One row per v0 route the client used."""
    rows = {}
    for entry in entries:
        key = entry["callable"] or f"(none) {entry['method']} {entry['path']}"
        row = rows.setdefault(key, {"requests": 0, "query": set(), "body": set(), "headers": set(), "undeclared": set(),
                                    "replayed": 0, "replay_failures": []})
        row["requests"] += 1
        row["query"] |= set(entry["query"])
        row["body"] |= set(entry["body"])
        row["headers"] |= set(entry["headers"])
        row["undeclared"] |= set(entry.get("undeclared", []))
        if "replay_problems" in entry:
            row["replayed"] += 1
            if entry["replay_problems"]:
                row["replay_failures"].append({"path": entry["path"], "problems": entry["replay_problems"][:3]})
    return {key: {k: sorted(v) if isinstance(v, set) else (v[:5] if isinstance(v, list) else v) for k, v in row.items()}
            for key, row in sorted(rows.items())}
