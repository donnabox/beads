#!/usr/bin/env python3
"""Read a bounded frozen table-export bundle; emit an advisory report to stdout.

This tool never opens a database, invokes bd, allocates identities, or writes files.
It is not a migration, a backup validator, or proof that an export was consistent.
"""
import argparse
from collections import Counter, defaultdict
import hashlib
import json
import math
from pathlib import Path
import re
import sys

MAX_BYTES = 32 * 1024 * 1024
MAX_ROWS = 100_000
TABLES = {
    "issues", "wisps", "dependencies", "wisp_dependencies", "config", "labels", "comments",
    "issue_versions", "store_epoch", "dolt_ignore", "schema_migrations", "dolt_status",
    "graph_preview_catalog", "graph_preview_scope", "graph_preview_types", "graph_preview_versions",
    "graph_preview_issue_versions", "graph_preview_links", "graph_preview_payloads",
}
ISSUE_FIELDS = {
    "id", "title", "description", "issue_type", "status", "priority", "created_at", "updated_at",
    "closed_at", "created_by", "closed_by", "close_reason", "current_revision", "content_hash",
}
DEPENDENCY_FIELDS = {"id", "issue_id", "depends_on_id", "depends_on_issue_id", "depends_on_wisp_id",
                     "depends_on_external", "type", "created_at", "created_by"}
REQUIRED_ARTIFACTS = ("metadata.json", "config.yaml", "database-backup", "working-set", "table-inventory")


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"duplicate JSON member: {key}")
        result[key] = value
    return result


def finite_float(token):
    value = float(token)
    if not math.isfinite(value):
        raise ValueError("JSON number exceeds finite numeric range")
    return value


def parse_json(raw):
    if isinstance(raw, bytes):
        raw = raw.decode("utf-8")
    return json.loads(raw, object_pairs_hook=unique_object, parse_float=finite_float,
                      parse_constant=lambda value: (_ for _ in ()).throw(ValueError(f"non-JSON number: {value}")))


def nonempty(value):
    return value is not None and value is not False and value != "" and value != 0 and value != [] and value != {}


def candidate_path(identity):
    # Conservative candidate only, not the canonical validator or an assignment.
    # Importers must use graphops.ValidateBeadPath after contract review.
    if isinstance(identity, str) and re.fullmatch(r"[A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*", identity):
        return "beads/" + identity
    return None


def validate(bundle):
    if not isinstance(bundle, dict) or bundle.get("schemaVersion") != 1:
        raise ValueError("expected frozen export bundle schemaVersion 1")
    provenance = bundle.get("provenance")
    if not isinstance(provenance, dict) or provenance.get("kind") not in {
        "synthetic", "release-authored-surrogate", "permitted-workspace"
    }:
        raise ValueError("provenance.kind must explicitly identify synthetic, release-authored-surrogate, or permitted-workspace")
    inventory, tables = bundle.get("tableInventory"), bundle.get("tables")
    if not isinstance(inventory, list) or any(not isinstance(name, str) for name in inventory) or len(set(inventory)) != len(inventory):
        raise ValueError("tableInventory must contain distinct table names")
    if not isinstance(tables, dict) or any(name not in inventory for name in tables):
        raise ValueError("tables must be an object whose names occur in tableInventory")
    count = 0
    for name, rows in tables.items():
        if not isinstance(rows, list) or any(not isinstance(row, dict) for row in rows):
            raise ValueError(f"{name}: export must be an array of row objects")
        count += len(rows)
    if count > MAX_ROWS:
        raise ValueError(f"export exceeds {MAX_ROWS} rows; no partial report is produced")
    if not isinstance(bundle.get("artifacts", {}), dict) or not isinstance(bundle.get("metadata", {}), dict):
        raise ValueError("artifacts and metadata must be objects")
    for name, artifact in bundle.get("artifacts", {}).items():
        if not isinstance(artifact, dict):
            raise ValueError(f"{name}: artifact must be an object")
        digest = artifact.get("sha256")
        if digest is not None and (not isinstance(digest, str) or re.fullmatch(r"[0-9a-f]{64}", digest) is None):
            raise ValueError(f"{name}: sha256 must be a lowercase SHA-256 digest")
    return tables


def report(bundle, input_digest):
    tables = validate(bundle)
    inventory = set(bundle["tableInventory"])
    findings = defaultdict(list)

    def add(code, identity, reason, **details):
        findings[code].append({"identity": str(identity), "reason": reason, **details})

    provenance = bundle["provenance"]
    if provenance["kind"] != "permitted-workspace":
        add("REPRESENTATIVE_SOURCE_UNAVAILABLE", "source", "This input is a surrogate; representative permitted workspace adoption remains unproven.")
    if provenance.get("frozen") is not True or not provenance.get("head") or not provenance.get("workingSetSha256"):
        add("SNAPSHOT_PROVENANCE_INCOMPLETE", "source", "Require frozen source, commit identity and separate working-set fingerprint; declarations are not independently verified.")
    if bundle.get("complete") is not True:
        add("EXPORT_COMPLETENESS_UNATTESTED", "source", "Exporter did not attest complete row exports; no absence conclusion is safe.")
    for name in sorted(inventory - set(tables)):
        add("TABLE_NOT_EXPORTED", name, "Present table was not exported; contents remain unknown.")
    for name in sorted(inventory - TABLES):
        add("UNCLASSIFIED_TABLE", name, "Table requires an adoption/backup preservation decision.")
    for name in ["issues", "dependencies", "config"]:
        if name not in inventory:
            add("EXPECTED_TABLE_ABSENT", name, "Missing legacy table; empty data is not inferred.")

    identities, candidates = [], defaultdict(list)
    issue_ids = set()
    histograms = {"issueTypes": Counter(), "issueStatuses": Counter(), "dependencyTypes": Counter(), "externalEncodings": Counter()}
    for table in ["issues", "wisps"]:
        seen = set()
        for index, row in enumerate(tables.get(table, [])):
            identity = row.get("id")
            label = identity if isinstance(identity, str) and identity else f"{table}[{index}]"
            if not isinstance(identity, str) or not identity:
                add("INVALID_IDENTITY", label, "Expected a nonempty string ID; no mapping proposed.")
                continue
            if identity in seen:
                add("DUPLICATE_IDENTITY", identity, f"Duplicate ID in {table} export.")
            seen.add(identity)
            path = candidate_path(identity)
            identities.append({"kind": "Issue" if table == "issues" else "Wisp", "identity": identity, "candidatePath": path})
            if path:
                candidates[path].append(table + ":" + identity)
            else:
                add("IDENTITY_MAPPING_REVIEW", identity, "Identity needs canonical-path/alias mapping; preserve original spelling.")
            if table == "wisps":
                add("WISP_ADOPTION_UNSUPPORTED", identity, "Ephemeral/unversioned plane has no graph adoption adapter.")
                continue
            issue_ids.add(identity)
            for field, bucket, allowed in [("issue_type", "issueTypes", {"task", "bug", "feature", "epic", "chore"}),
                                           ("status", "issueStatuses", {"open", "in_progress", "blocked", "deferred", "closed"})]:
                value = row.get(field)
                if not isinstance(value, str):
                    add("INVALID_ISSUE_FIELD", identity, f"{field} must be a string.")
                    value = "<invalid>"
                histograms[bucket][value] += 1
                if value not in allowed:
                    add("ISSUE_CLASSIFICATION_REVIEW", identity, f"{field} needs custom/infra classification review.", field=field, value=value)
            unsupported = sorted(key for key, value in row.items() if key not in ISSUE_FIELDS and nonempty(value))
            if unsupported:
                add("ISSUE_FIELDS_WITHOUT_ADOPTION_ADAPTER", identity, "Current graph create slice cannot import these populated fields losslessly.", fields=unsupported)
    memory_count = 0
    for row in tables.get("config", []):
        key = row.get("key")
        if not isinstance(key, str):
            add("INVALID_CONFIG_KEY", "config", "Non-string key cannot be classified.")
            continue
        if key.startswith("kv.memory."):
            exact = key[len("kv.memory."):]
            memory_count += 1
            path = candidate_path(exact)
            identities.append({"kind": "LegacyMemory", "identity": exact, "candidatePath": path,
                               "emptyBody": row.get("value") == ""})
            if path:
                candidates[path].append("memory:" + exact)
            add("LEGACY_MEMORY_CONVERSION_REQUIRED", exact, "Preserve exact key and empty body; do not invent actor/history. Canonical path and alias decision remains required.")
            if not isinstance(row.get("value"), str):
                add("INVALID_MEMORY_BODY", exact, "Memory body is not a string.")
        elif key in {"types.custom", "types.infra", "status.custom"} and nonempty(row.get("value")):
            add("CONFIGURATION_REVIEW", key, "Configured custom/infra vocabulary requires explicit preservation and type admission.")
    for path, owners in sorted(candidates.items()):
        if len(owners) > 1:
            add("CANDIDATE_PATH_COLLISION", path, "Provisional mappings collide; no allocation or automatic renaming is permitted.", owners=owners)

    dependencies, pairs = [], defaultdict(list)
    for table in ["dependencies", "wisp_dependencies"]:
        for index, row in enumerate(tables.get(table, [])):
            identity = str(row.get("id") or f"{table}[{index}]")
            source = row.get("issue_id")
            typed = [(key, row[key]) for key in ("depends_on_issue_id", "depends_on_wisp_id", "depends_on_external")
                     if row.get(key) is not None]
            if not typed and row.get("depends_on_id") is not None:
                typed = [("depends_on_id", row["depends_on_id"])]
            if len(typed) != 1 or not isinstance(source, str) or not isinstance(typed[0][1] if typed else None, str):
                add("INVALID_DEPENDENCY_ENDPOINTS", identity, "Expected one source and exactly one target; preserve malformed row for review.")
                continue
            column, target = typed[0]
            if column != "depends_on_id" and row.get("depends_on_id") is not None:
                add("MIXED_DEPENDENCY_ENDPOINT_LAYOUT", identity, "Both typed and legacy target fields are populated; explicit schema interpretation is required.",
                    legacyTarget=row["depends_on_id"] if isinstance(row["depends_on_id"], str) else "<non-string>")
            unsupported = sorted(key for key, value in row.items() if key not in DEPENDENCY_FIELDS and nonempty(value))
            if unsupported:
                add("DEPENDENCY_FIELDS_WITHOUT_ADOPTION_ADAPTER", identity, "Preserve populated Dependency properties; no adoption projection is qualified.", fields=unsupported)
            typ = row.get("type", "<missing>")
            if not isinstance(typ, str):
                typ = "<invalid>"
            histograms["dependencyTypes"][typ] += 1
            dependencies.append({"identity": identity, "source": source, "target": target, "targetColumn": column, "type": typ})
            pairs[(source, target)].append(identity)
            if table == "wisp_dependencies" or column == "depends_on_wisp_id":
                add("WISP_DEPENDENCY_UNSUPPORTED", identity, "Wisp endpoint requires explicit persistence/lifetime handling.")
            if source not in issue_ids:
                add("DEPENDENCY_SOURCE_UNRESOLVED", identity, "Source is not in exported durable Issues.", source=source)
            external = column == "depends_on_external" or target.startswith(("external:", "https://", "http://"))
            if external:
                encoding = "external-prefix" if target.startswith("external:") else "url" if target.startswith(("https://", "http://")) else "opaque"
                histograms["externalEncodings"][encoding] += 1
                add("EXTERNAL_BLOCKER_CONTRACT_REQUIRED" if typ == "blocks" else "EXTERNAL_REFERENCE_REVIEW", identity,
                    "Preserve opaque reference; never fabricate a local Issue or infer remote satisfaction. Current blocking descriptor admits local Issues only.", target=target, encoding=encoding)
            elif target not in issue_ids and column != "depends_on_wisp_id":
                add("DEPENDENCY_TARGET_UNRESOLVED", identity, "Missing local or differently prefixed reference; locality cannot be guessed.", target=target)
            if typ != "blocks":
                add("DEPENDENCY_TYPE_WITHOUT_ADAPTER", identity, "Only existing blocks relationships have a qualified Issue-domain graph adapter.", type=typ)
    for pair, ids in sorted(pairs.items()):
        if len(ids) > 1:
            add("LEGACY_PAIR_IDENTITY_COLLISION", " → ".join(pair), "Legacy Dependency identity excludes Type; do not silently turn duplicates into generic multiedges.", ids=ids)

    versions = defaultdict(list)
    for index, row in enumerate(tables.get("issue_versions", [])):
        identity = row.get("issue_id")
        revision = row.get("revision")
        if not isinstance(identity, str) or not isinstance(revision, (int, str)):
            add("INVALID_RETAINED_IDENTITY", f"issue_versions[{index}]", "Missing Issue identity or revision.")
            continue
        versions[(identity, str(revision))].append(row)
        snapshot = row.get("durable_state")
        try:
            if isinstance(snapshot, str):
                snapshot = parse_json(snapshot)
            if not isinstance(snapshot, dict) and not row.get("removed_at"):
                raise ValueError("non-object durable state")
        except (ValueError, RecursionError):
            add("MALFORMED_RETAINED_STATE", identity, "Retained state is not a valid JSON object.", revision=revision)
        if row.get("attribution_status") not in {"claimed", "unknown"}:
            add("HISTORY_ATTRIBUTION_GAP", identity, "Attribution status missing/unrecognized; do not invent an actor.", revision=revision)
        if not row.get("epoch"):
            add("HISTORY_EPOCH_GAP", identity, "Retained version lacks epoch.", revision=revision)
        if row.get("removed_at"):
            add("REMOVED_HISTORY_REQUIRES_PRESERVATION", identity, "Removed state must remain reserved and discoverable under the reviewed History contract.", revision=revision)
    if not tables.get("store_epoch"):
        add("STORE_EPOCH_UNAVAILABLE", "store_epoch", "No exported epoch; continuity is not established.")
    for row in tables.get("issues", []):
        identity, revision = row.get("id"), row.get("current_revision")
        if not isinstance(identity, str):
            continue
        current = versions.get((identity, str(revision)), [])
        if revision is None or len(current) != 1 or current[0].get("removed_at"):
            add("CURRENT_RETAINED_STATE_GAP", identity, "Current Issue does not have exactly one live retained row. Default revision 1 is not a backfill.")
    add("HISTORY_CONTINUITY_UNPROVEN", "issue_versions", "This inventory checks presence/shape, not full content equality, historical reachability, writer completeness or restoration.")
    for table in ["labels", "comments"]:
        if tables.get(table):
            add("AGGREGATE_IMPORT_REQUIRED", table, "Preserve Issue aggregate rows through a qualified adoption transaction; current creation is not an importer.", count=len(tables[table]))

    artifacts = bundle.get("artifacts", {})
    for name in REQUIRED_ARTIFACTS:
        entry = artifacts.get(name, {})
        if entry.get("present") is not True or not entry.get("sha256"):
            add("BACKUP_ARTIFACT_UNVERIFIED", name, "Require separately preserved artifact and fingerprint; SQL table export alone cannot establish recoverability.")
    metadata = bundle.get("metadata", {})
    binding = {key: metadata.get(key) for key in ("graph_scope_url", "graph_authority_id", "graph_workspace", "graph_schema_version")}
    if any(value is None or value == "" for value in binding.values()):
        add("GRAPH_BINDING_NOT_ESTABLISHED", "metadata.json", "Scope, authority and canonical workspace binding require explicit first-adoption or continuity decision; do not infer from _project_id.")
    if any(name.startswith("graph_preview_") for name in inventory):
        add("EXISTING_GRAPH_STATE_REQUIRES_CONTINUITY", "graph_preview_*", "Preserve allocation/tombstones, retained versions, Type descriptors, scope row and matching local metadata as one backup unit.")
    if "dolt_ignore" not in tables:
        add("IGNORED_TABLE_INVENTORY_UNAVAILABLE", "dolt_ignore", "No ignored-table export; clone-local backup coverage cannot be classified.")
    if tables.get("dolt_ignore"):
        add("IGNORED_TABLE_BACKUP_REVIEW", "dolt_ignore", "Ignored and clone-local tables need separate backup classification; commit identity alone is insufficient.")
    if tables.get("dolt_status"):
        add("WORKING_SET_BACKUP_REVIEW", "dolt_status", "Working-set changes exist. Commit-only backup watermarks cannot establish inclusion of graph writes.")
    add("BACKUP_RESTORE_CONTINUITY_UNPROVEN", "backup", "Artifact inventory is not a fresh-machine restore test. Commit-based backup change detection requires qualification against graph working-set writes.")
    exceptions = [{"code": code, "count": len(items), "items": items} for code, items in sorted(findings.items())]
    return {
        "schemaVersion": 1, "readOnly": True, "adoptionReady": False,
        "inputSha256": input_digest, "sourceKind": provenance["kind"],
        "backendObserved": metadata.get("backend"),
        "counts": {"tablesPresent": len(inventory), "tablesExported": len(tables),
                   "issues": len(tables["issues"]) if "issues" in tables else None,
                   "wisps": len(tables["wisps"]) if "wisps" in tables else None,
                   "dependencies": len(tables["dependencies"]) if "dependencies" in tables else None,
                   "wispDependencies": len(tables["wisp_dependencies"]) if "wisp_dependencies" in tables else None,
                   "parsedDependencies": len(dependencies),
                   "legacyMemories": memory_count if "config" in tables else None,
                   "retainedRows": len(tables["issue_versions"]) if "issue_versions" in tables else None},
        "histograms": {key: dict(sorted(value.items())) for key, value in histograms.items()},
        "identities": identities, "dependencies": dependencies, "bindingObserved": binding,
        "exceptions": exceptions,
        "limits": ["No database was opened; source snapshot consistency and provenance are exporter declarations.",
                   "No identity, schema, alias, Type, actor or history was allocated, changed or invented.",
                   "Candidate paths are conservative review aids, not approved mappings; use graphops/depid in any eventual importer.",
                   "Bodies, config values and Issue metadata values are omitted from the report; identities and external references remain visible.",
                   "No migration, backup restore, multi-clone writer exclusion or production adoption is qualified."],
    }


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("bundle", type=Path, help="frozen JSON export bundle; never a database directory")
    args = parser.parse_args(argv)
    try:
        with args.bundle.open("rb") as source:
            raw = source.read(MAX_BYTES + 1)
        if len(raw) > MAX_BYTES:
            raise ValueError(f"bundle exceeds {MAX_BYTES} bytes; no partial report is produced")
        result = report(parse_json(raw), hashlib.sha256(raw).hexdigest())
        print(json.dumps(result, ensure_ascii=False, indent=2, allow_nan=False))
        return 0
    except (OSError, ValueError, TypeError, RecursionError) as exc:
        print(json.dumps({"code": "invalid_export", "message": str(exc)}), file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
