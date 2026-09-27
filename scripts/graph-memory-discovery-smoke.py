#!/usr/bin/env python3
"""Installed Memory discovery-to-exact-recall proof, authored through normal CLI.

Uses fresh processes, complete summary assertions and caller-owned ordinary Dolt.
Both engines run serially. No SQL, fixtures, mock stores or hidden bootstrap.
"""
import argparse
import copy
import concurrent.futures
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import signal
import time
import threading

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("graph_c0_capture", HERE / "graph-c0-smoke.py")
c0 = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(c0)
SCOPE = "https://example.invalid/disposable-memory-discovery/"
LIMITATIONS = [
    "experimental complete-or-refuse discovery, not CLI pagination or complete Memory R6",
    "records-json is a summary, not full Memory export or connected interchange",
    "title/body literal case-folded search only; no keys, metadata, normalization, ranking or inferred Links",
    "exact saved-version recall is not native/public HTTP History",
    "server concurrent processes do not force engine overlap; no corruption, crash, rollback or acquisition-limit injection",
    "ordinary Dolt 2.1.8 database provisioning remains serialized",
]

def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    expected, searches = {}, []

    def initialization(graph, suffix):
        args = ["init", "--prefix", suffix, "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
        if graph:
            args += ["--graph-mode", "link", "--scope-url", SCOPE]
        if capture.args.server_port:
            args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                     str(capture.args.server_port), "--database", suffix + "_" + capture.root.name.replace("-", "_"),
                     "--server-user", "root"]
        return args

    initialized = c0.envelope(capture.success("normal-graph-init", initialization(True, "discover")))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong graph authority/backend")
    status = c0.envelope(capture.success("discovery-status", ["status", "--graph", "--json"]))
    capabilities = status.get("capabilities", {})
    c0.require(capabilities.get("memoryDiscovery") is True and capabilities.get("memoryDiscoveryPagination") is False and
               capabilities.get("memory") is False and capabilities.get("historyExact") is False,
               "discovery advertised incomplete capability incorrectly")
    for key, value in {"memoryDiscoveryDefaultMatches": 50, "memoryDiscoveryOutputBytes": 1 << 20,
                       "memoryDiscoveryQueryBytes": 4096, "memoryDiscoveryExcerptCodePoints": 160}.items():
        c0.require(status.get("limits", {}).get(key) == value, "wrong disclosed discovery limit: " + key)

    def record(label, value):
        c0.require(value.get("id", "").startswith(SCOPE) and value.get("version") and value.get("attribution"),
                   label + ": incomplete authored record")
        expected[label] = copy.deepcopy(value)
        c0.write_json(capture.output / "expected-records.json", expected)
        return value

    def memory(label, path, title, body):
        return record(label, c0.envelope(capture.success(label, [
            "remember", body, "--id", path, "--title", title, "--actor", "discovery-author", "--json"])))

    def summary(label, query, records, fields=None, details=False, all_rows=False, flags=None):
        args = ["memories"]
        if query is not None:
            args.append(query)
        args += ["--format", "records-json"]
        if details:
            args.append("--details")
        if all_rows:
            args.append("--all")
        args += flags or []
        response = capture.success(label, args)
        output_file = capture.output / capture.records[-1]["artifact"] / "stdout.log"
        c0.require(output_file.stat().st_size <= 1 << 20, label + ": serialized summary exceeds response bound")
        c0.require(set(response) == {"schemaVersion", "preview", "result"} and
                   response.get("schemaVersion") == 1 and response.get("preview") is True, label + ": wrong wrapper")
        result = response["result"]
        c0.require(isinstance(result, dict) and result.get("projection") == "summary" and
                   result.get("scope") == SCOPE and result.get("complete") is True and
                   "next" in result and result["next"] is None and isinstance(result.get("items"), list),
                   label + ": partial/invalid discovery result")
        ordered = sorted(records, key=lambda item: item["id"].encode("utf-16-be"))
        items = result["items"]
        c0.require([item.get("id") for item in items] == [item["id"] for item in ordered],
                   label + ": wrong matches/order or duplicate identity")
        for item, wanted in zip(items, ordered):
            required = {"id", "type", "title", "version", "attribution", "matchedFields"}
            allowed = required | {"excerpt", "details"}
            c0.require(required <= set(item) <= allowed, label + ": body/full-record or unexpected summary fields")
            for key in ["id", "type", "version", "attribution"]:
                c0.require(item[key] == wanted[key], label + ": inconsistent selected " + key)
            c0.require(item["title"] == wanted["properties"]["title"], label + ": title differs")
            matches = [] if fields is None else fields[wanted["id"]]
            c0.require(isinstance(item["matchedFields"], list) and
                       len(item["matchedFields"]) == len(set(item["matchedFields"])) and
                       set(item["matchedFields"]) == set(matches), label + ": wrong field provenance")
            if details:
                c0.require(item.get("details") == {"ownedLinkCount": len(wanted["owned"])},
                           label + ": details leaked payload or lost complete owned count")
            else:
                c0.require("details" not in item, label + ": unrequested details")
            excerpt = item.get("excerpt")
            if "body" in matches:
                c0.require(isinstance(excerpt, dict) and excerpt.get("field") == "body",
                           label + ": body match lacks body excerpt")
            if excerpt is not None:
                c0.require(set(excerpt) == {"field", "text", "truncated"} and excerpt["field"] in matches and
                           isinstance(excerpt["text"], str) and len(excerpt["text"]) <= 160 and
                           isinstance(excerpt["truncated"], bool), label + ": unbounded or untruthful excerpt")
                source = wanted["properties"][excerpt["field"]]
                if len(source) > 160:
                    c0.require(excerpt["truncated"] is True and excerpt["text"] != source, label + ": full long field leaked")
                c0.require("PRIVATE_TAIL_" not in excerpt["text"], label + ": unrelated tail body leaked")
                if query and len(query) < 40:
                    c0.require(query.casefold() in excerpt["text"].casefold(), label + ": excerpt does not show match")
            c0.require("TITLE_PRIVATE_BODY_SENTINEL" not in json.dumps(item), label + ": title-only search leaked body")
        searches.append({"label": label, "query": query, "result": copy.deepcopy(result)})
        c0.write_json(capture.output / "searches.json", searches)
        return items

    def refuse(label, args, code="capability_unavailable", typed=True, structured=None):
        result = capture.run(label, args)
        receipt, out, err = result
        c0.require(receipt["exit_code"] != 0 and out == "", label + ": refusal emitted successful stdout")
        if typed:
            if structured is None:
                structured = "--json" in args or "--format=records-json" in args or (
                    "--format" in args and args[args.index("--format") + 1] == "records-json")
            if structured:
                problem = json.loads(err)
                c0.require(problem.get("code") == code and problem.get("retryable") is False and
                           "result" not in problem, label + ": wrong JSON refusal")
            else:
                # Non-structured format choices retain plain coded stderr;
                # records-json and true --json select JSON failures.
                c0.require(err.startswith(code + ": ") and err.endswith("\n"),
                           label + ": wrong plain coded refusal")
        else:
            c0.require(bool(err), label + ": missing command-line diagnostic")

    def exact(label, selected, original):
        actual = c0.envelope(capture.success(label, ["show", selected["id"], "--version", selected["version"],
                                                  "--readonly", "--json"]))
        c0.require(actual == original, label + ": selected complete saved state differs")
        receipt, _, _ = capture.run(label + "-body", ["recall", selected["id"], "--version", selected["version"], "--readonly"])
        artifact = capture.output / next(item["artifact"] for item in capture.records
                                         if item.get("pid") == receipt.get("pid") and
                                         item.get("started_unix") == receipt.get("started_unix"))
        c0.require(receipt["exit_code"] == 0 and (artifact / "stderr.log").read_bytes() == b"" and
                   (artifact / "stdout.log").read_bytes() == original["properties"]["body"].encode(),
                   label + ": selected body cannot be recalled exactly")

    # First show that empty means a complete result, not a hidden first page.
    summary("empty-corpus", None, [])
    title_only = memory("title-only", "beads/title", "Titleneedle report", "TITLE_PRIVATE_BODY_SENTINEL")
    body_only = memory("body-only", "beads/body", "Body case",
                       "x" * 240 + " bodyneedle 雪😀 " + "y" * 400 + " PRIVATE_TAIL_BODY")
    both = memory("both-fields", "beads/both", "Togetherneedle note",
                  "Togetherneedle original body\r\n  " + "z" * 300 + " PRIVATE_TAIL_BOTH")
    empty = memory("empty-body", "beads/empty", "Empty memo", "")
    unicode = memory("unicode-fold", "beads/unicode", "Straße Σίσυφος",
                     "Straße has expanding fold; Σςσ has sigma variants; cafe\u0301 stays decomposed.")
    issue = record("issue", c0.envelope(capture.success("create-issue",
                   ["create", "Titleneedle excluded Issue", "--id", "beads/work", "--json"])))
    relation = c0.envelope(capture.success("owned-context", ["link", both["id"], issue["id"],
               "--resource-type", SCOPE + "types/preview-related-v2", "--id", "links/context",
               "--properties", '{"note":"original context"}', "--if-source-revision", both["revision"], "--json"]))
    both, link = record("both-linked", relation["source"]), record("context", relation["link"])
    all_memories = [title_only, body_only, both, empty, unicode]
    summary("omitted-query", None, all_memories)
    summary("empty-query", "", all_memories)
    summary("title-match", "TITLEneedle", [title_only], {title_only["id"]: ["title"]})
    summary("body-match", "BODYneedle", [body_only], {body_only["id"]: ["body"]})
    selected = summary("both-match-details", "Togetherneedle", [both],
                       {both["id"]: ["title", "body"]}, details=True)[0]
    summary("unicode-expanding-fold", "STRASSE", [unicode], {unicode["id"]: ["title", "body"]})
    summary("unicode-sigma-fold", "σ", [unicode], {unicode["id"]: ["title", "body"]})
    summary("no-normalization", "café", [])
    summary("literal-query-no-regex", "Together.*", [])
    summary("no-match", "not-present-anywhere", [])
    summary("empty-body-title", "Empty memo", [empty], {empty["id"]: ["title"]})
    capture.passed("complete title/body/both/empty/no-match summaries; Unicode folding; literal matching; bounded excerpts and details")

    for label, args, records in [
        ("human-table-default", ["memories"], all_memories),
        ("human-details", ["memories", "Togetherneedle", "--details", "--format", "table"], [both]),
    ]:
        receipt, out, err = capture.run(label, args)
        c0.require(receipt["exit_code"] == 0 and not err, label + ": table failed")
        for item in records:
            c0.require(item["id"] in out and item["version"] in out, label + ": missing usable id/version")
        c0.require("bd recall" in out and "PRIVATE_TAIL_" not in out and
                   "TITLE_PRIVATE_BODY_SENTINEL" not in out, label + ": body leak or missing follow-up")
    capture.passed("human table/details expose usable recall identity and version without complete bodies")

    selected_before = copy.deepcopy(both)
    updated_properties = {"title": both["properties"]["title"], "body": "Togetherneedle revised body — 雪\r\n"}
    update = c0.envelope(capture.success("edit-after-selection", ["update", both["id"], "--properties",
        json.dumps(updated_properties, ensure_ascii=False), "--if-revision", both["revision"], "--actor", "later-editor", "--json"]))
    both = record("both-body-edited", update["memory"])
    c0.require(both["version"] != selected["version"] and both["owned"] == selected_before["owned"], "edit did not isolate body")
    current_selected = summary("fresh-after-body-edit", "Togetherneedle", [both],
                       {both["id"]: ["title", "body"]}, details=True)[0]
    exact("selected-before-body-edit", selected, selected_before)
    after_body = copy.deepcopy(both)
    changed_link = c0.envelope(capture.success("edit-owned-link-after-selection", ["update", link["id"],
        "--properties", '{"note":"new linked context"}', "--if-revision", link["revision"],
        "--if-source-revision", both["revision"], "--actor", "context-editor", "--json"]))
    both, link = record("both-context-edited", changed_link["source"]), record("context-edited", changed_link["link"])
    c0.require(both["version"] != current_selected["version"] and both["properties"] == after_body["properties"],
               "owned Link edit did not version only source context")
    summary("fresh-after-owned-edit", "Togetherneedle", [both],
            {both["id"]: ["title", "body"]}, details=True)
    exact("selected-before-owned-edit", current_selected, after_body)
    exact("original-selection-after-both-edits", selected, selected_before)
    all_memories = [title_only, body_only, both, empty, unicode]
    c0.require(c0.envelope(capture.success("target-still-unchanged", ["show", issue["id"], "--readonly", "--json"])) == issue,
               "discovery/edit changed Issue target")
    capture.passed("selected id+version recalls exact old body and complete linked state after fresh-process body/Link edits")

    summary("readonly-flag", None, all_memories, flags=["--readonly"])
    prior_readonly = capture.env.get("BD_READONLY")
    try:
        capture.env["BD_READONLY"] = "true"
        summary("readonly-environment", None, all_memories)
    finally:
        if prior_readonly is None:
            capture.env.pop("BD_READONLY", None)
        else:
            capture.env["BD_READONLY"] = prior_readonly
    yaml = capture.work / ".beads" / "config.yaml"
    prior_yaml = yaml.read_bytes() if yaml.exists() else None
    try:
        yaml.write_text("readonly: true\n")
        summary("readonly-configuration", None, all_memories)
    finally:
        if prior_yaml is None:
            yaml.unlink(missing_ok=True)
        else:
            yaml.write_bytes(prior_yaml)
    try:
        yaml.write_text("json: true\n")
        summary("configured-json-explicit-records", None, all_memories)
        receipt, out, err = capture.run("configured-json-explicit-table", ["memories", "--format", "table"])
        c0.require(receipt["exit_code"] == 0 and not err and out.startswith("Memories ("),
                   "configured JSON overrode explicit table")
        for item in all_memories:
            c0.require(item["id"] in out and item["version"] in out,
                       "configured-json table omitted usable id/version")
        c0.require("bd recall" in out and "PRIVATE_TAIL_" not in out and
                   "TITLE_PRIVATE_BODY_SENTINEL" not in out, "configured-json table leaked a body")
        for label, args in [
            ("configured-json-unspecified-format", ["memories"]),
            ("configured-json-explicit-json", ["memories", "--json"]),
            ("configured-json-records-plus-json", ["memories", "--format", "records-json", "--json"]),
            ("configured-json-table-plus-json", ["memories", "--format", "table", "--json"]),
        ]:
            refuse(label, args, structured=True)
    finally:
        if prior_yaml is None:
            yaml.unlink(missing_ok=True)
        else:
            yaml.write_bytes(prior_yaml)
    capture.passed("explicit summary/table formats override configured JSON; unspecified format and explicit JSON still refuse")

    for label, args, code, typed in [
        ("bare-json", ["memories", "--json"], "capability_unavailable", True),
        ("false-json", ["memories", "--json=false"], "capability_unavailable", True),
        ("records-plus-json", ["memories", "--format", "records-json", "--json"], "capability_unavailable", True),
        ("legacy-format", ["memories", "--format", "legacy-json"], "capability_unavailable", True),
        ("invalid-format", ["memories", "--format", "mystery"], "capability_unavailable", True),
        ("empty-format", ["memories", "--format", ""], "capability_unavailable", True),
        ("oversized-query", ["memories", "q" * 4097, "--format", "records-json"], "invalid_selector", True),
        ("unsupported-limit", ["memories", "--limit", "1"], None, False),
        ("unsupported-after", ["memories", "--after", "token"], None, False),
        ("unsupported-full", ["memories", "--full"], None, False),
    ]:
        refuse(label, args, code, typed)
    capture.passed("readonly remains available; disputed JSON formats, oversize query and unsupported paging/body flags refuse")

    bulk = []
    for index in range(50):
        bulk.append(memory("bulk-create-" + str(index), "beads/bulk-" + str(index).zfill(2),
                           "BULKMARK " + str(index).zfill(2), ""))
    bulk_fields = {item["id"]: ["title"] for item in bulk}
    summary("exactly-fifty-default-complete", "BULKMARK", bulk, bulk_fields)
    bulk.append(memory("bulk-create-50", "beads/bulk-50", "BULKMARK 50", ""))
    bulk_fields[bulk[-1]["id"]] = ["title"]
    refuse("fifty-one-default-refuses", ["memories", "BULKMARK", "--format", "records-json"])
    summary("fifty-one-explicit-all-complete", "BULKMARK", bulk, bulk_fields, all_rows=True)
    summary("queryless-explicit-all-complete", None, [*all_memories, *bulk], all_rows=True)
    summary("narrow-query-with-large-corpus", "Empty memo", [empty], {empty["id"]: ["title"]})
    summary("no-match-with-large-corpus", "not-present-anywhere", [])
    c0.require(c0.envelope(capture.success("graph-state-unchanged", ["show", both["id"], "--readonly", "--json"])) == both,
               "search changed current Memory")
    capture.passed("50 matches are complete; 51 refuse without stdout; explicit all and narrow/empty searches remain complete")

    overlap = {"exercised": False, "limitation": "concurrent CLI processes; engine transaction overlap is not forced"}
    if capture.args.server_port:
        concurrent_memory = memory("concurrent-seed", "beads/concurrent", "Concurrentneedle",
                                   "Concurrentneedle initial body.")
        started = threading.Event()

        def concurrent_writer():
            versions = []
            started.set()
            for index in range(4):
                properties = {"title": "Concurrentneedle", "body": "Concurrentneedle writer body " + str(index) + " — 雪"}
                response = c0.envelope(capture.success("concurrent-write-" + str(index),
                    ["update", concurrent_memory["id"], "--properties", json.dumps(properties, ensure_ascii=False),
                     "--unconditional", "--actor", "concurrent-writer-" + str(index), "--json"]))
                c0.require(response.get("changed") is True and response["memory"]["properties"] == properties,
                           "concurrent writer failed")
                versions.append(response["memory"])
            return versions

        observed = []
        with concurrent.futures.ThreadPoolExecutor(max_workers=1) as pool:
            writer = pool.submit(concurrent_writer)
            c0.require(started.wait(timeout=5), "concurrent writer did not start")
            for index in range(4):
                response = capture.success("concurrent-search-" + str(index),
                    ["memories", "Concurrentneedle", "--format", "records-json", "--details", "--readonly"])
                result = c0.envelope(response)
                c0.require(result.get("projection") == "summary" and result.get("scope") == SCOPE and
                           result.get("complete") is True and "next" in result and result["next"] is None and
                           len(result.get("items", [])) == 1, "concurrent search is incomplete")
                item = result["items"][0]
                c0.require(item["id"] == concurrent_memory["id"], "concurrent search changed identity")
                retained = c0.envelope(capture.success("concurrent-selected-state-" + str(index),
                    ["show", item["id"], "--version", item["version"], "--readonly", "--json"]))
                c0.require(all(item[key] == retained[key] for key in ["id", "type", "version", "attribution"]) and
                           item["title"] == retained["properties"]["title"] and
                           item.get("details") == {"ownedLinkCount": len(retained["owned"])} and
                           set(item["matchedFields"]) == {"title", "body"} and
                           item["excerpt"]["field"] == "body" and item["excerpt"]["text"] == retained["properties"]["body"],
                           "concurrent summary combined different versions")
                # All concurrent bodies are below the excerpt bound; comparing
                # their complete short excerpt to this same retained body proves
                # search bytes and the selected version came from one state.
                exact("concurrent-selected-recall-" + str(index), item, retained)
                observed.append({"summary": item, "retained": retained})
            written = writer.result()
        for index, value in enumerate(written):
            record("concurrent-written-" + str(index), value)
        c0.write_json(capture.output / "concurrent-observations.json", observed)
        overlap = {"exercised": True, "writes": len(written), "discoveries": len(observed),
                   "limitation": "concurrent CLI processes; engine transaction overlap is not forced"}
        capture.passed("ordinary-server concurrent writes/discovery select wholly retained state with matching exact body/attribution")

    graph_work = capture.work
    legacy_work = capture.root / "legacy-workspace"
    legacy_work.mkdir(mode=0o700)
    capture.work = legacy_work
    try:
        for label, flags in [("legacy-all-before-init", ["--all"]),
                             ("legacy-details-before-init", ["--details"]),
                             ("legacy-format-before-init", ["--format", "records-json"])]:
            before_tree = c0.tree_digest(capture.work)
            refuse(label, ["memories", *flags])
            c0.require(c0.tree_digest(capture.work) == before_tree, label + ": new option created legacy storage")
        receipt, _, err = capture.run("normal-legacy-init", initialization(False, "dlegacy"))
        c0.require(receipt["exit_code"] == 0, "legacy init failed: " + err)
        legacy_body = "Legacy body remains a key to body map — 雪\n"
        value = capture.success("legacy-remember", ["remember", legacy_body, "--key", "legacy-plan", "--json"])
        c0.require(value.get("key") == "legacy-plan" and value.get("value") == legacy_body, "legacy remember changed")
        legacy_map = capture.success("legacy-json-map", ["memories", "--json"])
        # Baseline outputJSON adds schema_version to object/map output when
        # BD_JSON_ENVELOPE is unset, as in Capture's isolated environment.
        c0.require(legacy_map == {"legacy-plan": legacy_body, "schema_version": 1},
                   "legacy JSON no longer carries the original map and schema marker")
        for spelling in ["json", "JSON"]:
            aliased = capture.success("legacy-format-alias-" + spelling, ["memories", "--format", spelling])
            c0.require(aliased == {"legacy-plan": legacy_body, "schema_version": 1},
                       "case-insensitive legacy format alias changed the exact map/schema marker")
        capture.passed("legacy memories --format json and --format JSON preserve the exact shipped JSON map")
        for label, flags in [("legacy-all", ["--all"]), ("legacy-details", ["--details"]),
                             ("legacy-format-table", ["--format", "table"]),
                             ("legacy-format-records", ["--format", "records-json"])]:
            refuse(label, ["memories", *flags])
        c0.require(capture.success("legacy-map-after-refusals", ["memories", "--json"]) == legacy_map,
                   "new options changed legacy memories")
    finally:
        capture.work = graph_work
    capture.passed("normal legacy workspace preserves memories --json map; new graph discovery flags refuse before legacy store use")
    return {"story_memories": 5, "boundary_memories": 51, "searches": len(searches),
            "expected_records": len(expected), "http_exercised": False, "concurrent_overlap": overlap,
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "searches_sha256": c0.sha256(capture.output / "searches.json")}


def run_backend(args, backend, deadline):
    selected = argparse.Namespace(**vars(args))
    selected.output_dir = args.output_dir / backend
    selected.server_port = args.server_port if backend == "server" else None
    selected.server_root = args.server_root if backend == "server" else None
    selected.total_timeout = max(0.01, deadline - time.monotonic())
    capture = c0.Capture(selected)
    began = time.monotonic()
    failure, interrupted, result = None, False, {}
    previous_signals = {}

    def interrupted_run(signum, _frame):
        capture.stop()
        raise KeyboardInterrupt(signum)

    try:
        for signum in [signal.SIGINT, signal.SIGTERM]:
            previous_signals[signum] = signal.signal(signum, interrupted_run)
        result = exercise(capture)
        c0.require(c0.sha256(args.bd) == capture.binary_hash, "installed binary changed during capture")
    except BaseException as exc:
        interrupted = isinstance(exc, KeyboardInterrupt)
        failure = f"{type(exc).__name__}: {exc}"
    finally:
        capture.stop()
        for signum, handler in previous_signals.items():
            signal.signal(signum, handler)
        if capture.active:
            failure = failure or "owned child processes remain after capture"
        summary = {"passed": failure is None, "backend": backend, "failure": failure,
                   "interrupted": interrupted, "elapsed_seconds": time.monotonic() - began,
                   "checks": capture.passes, "commands": capture.records,
                   "active_child_count": len(capture.active), "private_root": str(capture.root),
                   "installed_binary_sha256": capture.binary_hash,
                   "harness_sha256": c0.sha256(Path(__file__)),
                   "capture_helper_sha256": c0.sha256(HERE / "graph-c0-smoke.py"),
                   "limitations": LIMITATIONS, **result}
        c0.write_json(capture.output / "summary.json", summary)
    print(f"{'PASS' if summary['passed'] else 'FAIL'} {backend}: {capture.output}", flush=True)
    return summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", type=Path, required=True, help="absolute installed bd executable")
    parser.add_argument("--output-dir", type=Path, required=True, help="new capture directory")
    parser.add_argument("--backend", choices=["embedded", "server", "both"], default="both")
    parser.add_argument("--server-port", type=int, help="caller-owned ordinary loopback Dolt server")
    parser.add_argument("--server-root", type=Path, help="disposable caller-owned server root, provenance only")
    parser.add_argument("--command-timeout", type=float, default=120)
    parser.add_argument("--total-timeout", type=float, default=900, help="whole sequential capture budget")
    args = parser.parse_args()
    c0.require(args.bd.is_absolute() and args.bd.is_file() and os.access(args.bd, os.X_OK),
               "--bd must name an absolute executable installed binary")
    c0.require(args.output_dir.is_absolute(), "--output-dir must be absolute")
    c0.require(args.command_timeout > 0 and args.total_timeout > 0, "timeouts must be positive")
    if args.backend in {"server", "both"}:
        c0.require(args.server_port is not None and 0 < args.server_port < 65536,
                   "server/both require --server-port for an ordinary caller-owned server")
    if args.server_root is not None:
        c0.require(args.server_root.is_absolute() and args.server_root.is_dir(),
                   "--server-root must name an existing absolute disposable server directory")
    args.output_dir.mkdir(mode=0o700, parents=True, exist_ok=False)
    deadline = time.monotonic() + args.total_timeout
    modes = ["embedded", "server"] if args.backend == "both" else [args.backend]
    results = []
    failure = None
    try:
        for backend in modes:
            c0.require(time.monotonic() < deadline, "total capture deadline expired before next backend")
            result = run_backend(args, backend, deadline)
            results.append(result)
            if result["interrupted"]:
                break
    except BaseException as exc:
        failure = f"{type(exc).__name__}: {exc}"
    passed = failure is None and len(results) == len(modes) and all(item["passed"] for item in results)
    c0.write_json(args.output_dir / "summary.json", {
        "passed": passed, "failure": failure, "requested_backends": modes,
        "backends": [{key: value for key, value in result.items() if key != "commands"} for result in results],
        "harness_sha256": c0.sha256(Path(__file__)), "installed_binary_sha256": c0.sha256(args.bd),
        "active_child_count": sum(result["active_child_count"] for result in results),
        "command_count": sum(len(result["commands"]) for result in results),
        "check_count": sum(len(result["checks"]) for result in results),
        "limitations": LIMITATIONS,
    })
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
