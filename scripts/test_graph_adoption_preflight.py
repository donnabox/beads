"""Synthetic mechanics tests; these do not establish representative adoption."""
import contextlib
import copy
import hashlib
import importlib.util
import io
import json
from pathlib import Path
import tempfile
import unittest

SCRIPT = Path(__file__).with_name("graph-adoption-preflight.py")
spec = importlib.util.spec_from_file_location("adoption_preflight", SCRIPT)
preflight = importlib.util.module_from_spec(spec)
spec.loader.exec_module(preflight)


def bundle():
    tables = {
        "issues": [{"id": "old-1", "title": "Existing Issue", "status": "open", "issue_type": "task", "current_revision": 1}],
        "dependencies": [], "config": [], "issue_versions": [], "store_epoch": [],
    }
    return {"schemaVersion": 1, "complete": True,
            "provenance": {"kind": "synthetic", "frozen": True, "head": "observed-head", "workingSetSha256": "a" * 64},
            "tableInventory": list(tables), "tables": tables, "artifacts": {}, "metadata": {}}


def codes(result):
    return {item["code"]: item for item in result["exceptions"]}


class AdoptionPreflightTests(unittest.TestCase):
    def test_external_encodings_and_legacy_pair_identity_excludes_type(self):
        data = bundle()
        data["tables"]["dependencies"] = [
            {"id": "d1", "issue_id": "old-1", "depends_on_external": "external:other:old-2", "type": "blocks"},
            {"id": "d2", "issue_id": "old-1", "depends_on_external": "other-3", "type": "blocks"},
            {"id": "d3", "issue_id": "old-1", "depends_on_id": "https://foreign.invalid/beads/x", "type": "related"},
            {"id": "d4", "issue_id": "old-1", "depends_on_external": "external:other:old-2", "type": "related"},
            {"id": "d5", "issue_id": "old-1", "depends_on_id": "unknown-7", "type": "blocks"},
            {"id": "d6", "issue_id": "old-1", "depends_on_issue_id": "old-1", "depends_on_external": "opaque", "type": "blocks"},
        ]
        result = preflight.report(data, "digest")
        self.assertEqual(result["histograms"]["externalEncodings"], {"external-prefix": 2, "opaque": 1, "url": 1})
        found = codes(result)
        self.assertEqual(found["EXTERNAL_BLOCKER_CONTRACT_REQUIRED"]["count"], 2)
        self.assertEqual(found["LEGACY_PAIR_IDENTITY_COLLISION"]["items"][0]["ids"], ["d1", "d4"])
        self.assertIn("DEPENDENCY_TARGET_UNRESOLVED", found)
        self.assertIn("INVALID_DEPENDENCY_ENDPOINTS", found)
        self.assertEqual(result["counts"]["dependencies"], 6)
        self.assertEqual(result["counts"]["parsedDependencies"], 5)
        self.assertFalse(result["adoptionReady"])

    def test_dependency_properties_require_preservation_without_value_echo(self):
        data = bundle()
        data["tables"]["dependencies"] = [{"id": "d1", "issue_id": "old-1", "depends_on_issue_id": "old-1",
                                            "type": "blocks", "metadata": '{"note":"PRIVATE_METADATA"}',
                                            "thread_id": "PRIVATE_THREAD"}]
        result = preflight.report(data, "digest")
        fields = codes(result)["DEPENDENCY_FIELDS_WITHOUT_ADOPTION_ADAPTER"]["items"][0]["fields"]
        self.assertEqual(fields, ["metadata", "thread_id"])
        self.assertNotIn("PRIVATE_METADATA", json.dumps(result))
        self.assertNotIn("PRIVATE_THREAD", json.dumps(result))

    def test_mixed_endpoint_columns_are_an_explicit_exception(self):
        data = bundle()
        data["tables"]["dependencies"] = [{"issue_id": "old-1", "depends_on_issue_id": "old-1",
                                            "depends_on_id": "external:legacy:2", "type": "blocks"}]
        found = codes(preflight.report(data, "digest"))
        self.assertEqual(found["MIXED_DEPENDENCY_ENDPOINT_LAYOUT"]["items"][0]["legacyTarget"], "external:legacy:2")

    def test_utf8_only_and_number_overflow_refuse(self):
        with self.assertRaises(ValueError):
            preflight.parse_json('{"schemaVersion": 1}'.encode("utf-16"))
        with self.assertRaises(ValueError):
            preflight.parse_json('{"graph_schema_version": 1e400}')

    def test_exact_memory_keys_empty_body_and_candidate_collision(self):
        data = bundle()
        data["tables"]["config"] = [
            {"key": "kv.memory.old-1", "value": ""},
            {"key": "kv.memory. leading/key ", "value": "SECRET_BODY"},
            {"key": "memory.not-a-memory", "value": "not imported"},
            {"key": "kv.memory.schema_version", "value": "1"},
        ]
        result = preflight.report(data, "digest")
        self.assertEqual(result["counts"]["legacyMemories"], 3)
        memory = [row for row in result["identities"] if row["kind"] == "LegacyMemory"]
        self.assertTrue(memory[0]["emptyBody"])
        self.assertEqual(memory[1]["identity"], " leading/key ")
        self.assertIsNone(memory[1]["candidatePath"])
        self.assertIn("CANDIDATE_PATH_COLLISION", codes(result))
        self.assertNotIn("SECRET_BODY", json.dumps(result))

    def test_default_revision_does_not_imply_retained_history(self):
        data = bundle()
        result = preflight.report(data, "digest")
        self.assertIn("CURRENT_RETAINED_STATE_GAP", codes(result))
        data["tables"]["issue_versions"] = [{"issue_id": "old-1", "revision": 1, "durable_state": "broken"}]
        found = codes(preflight.report(data, "digest"))
        self.assertIn("MALFORMED_RETAINED_STATE", found)
        self.assertIn("HISTORY_ATTRIBUTION_GAP", found)
        self.assertIn("HISTORY_EPOCH_GAP", found)
        self.assertIn("HISTORY_CONTINUITY_UNPROVEN", found)
        data["tables"]["issue_versions"][0]["removed_at"] = "2026-09-25"
        self.assertIn("CURRENT_RETAINED_STATE_GAP", codes(preflight.report(data, "digest")))

    def test_partial_unknown_tables_and_unsupported_fields_are_named(self):
        data = bundle()
        data["tables"]["issues"][0].update(status="review", metadata={"secret": "DO_NOT_ECHO"}, assignee="somebody")
        data["tableInventory"] += ["missing_export", "wisps"]
        data["tables"]["wisps"] = [{"id": "ephemeral-1"}]
        result = preflight.report(data, "digest")
        found = codes(result)
        self.assertIn("TABLE_NOT_EXPORTED", found)
        self.assertIn("UNCLASSIFIED_TABLE", found)
        self.assertIn("WISP_ADOPTION_UNSUPPORTED", found)
        self.assertIn("ISSUE_CLASSIFICATION_REVIEW", found)
        self.assertEqual(found["ISSUE_FIELDS_WITHOUT_ADOPTION_ADAPTER"]["items"][0]["fields"], ["assignee", "metadata"])
        self.assertNotIn("DO_NOT_ECHO", json.dumps(result))

    def test_missing_export_is_unknown_not_zero(self):
        data = bundle()
        del data["tables"]["issues"]
        result = preflight.report(data, "digest")
        self.assertIsNone(result["counts"]["issues"])
        self.assertIn("TABLE_NOT_EXPORTED", codes(result))

    def test_backup_artifacts_do_not_claim_restore_or_authority(self):
        data = bundle()
        data["artifacts"] = {name: {"present": True, "sha256": "b" * 64} for name in preflight.REQUIRED_ARTIFACTS}
        data["tableInventory"] += ["dolt_status", "dolt_ignore", "graph_preview_catalog"]
        data["tables"].update(dolt_status=[{"table_name": "issues", "status": "modified"}],
                              dolt_ignore=[{"pattern": "wisps"}], graph_preview_catalog=[])
        found = codes(preflight.report(data, "digest"))
        self.assertNotIn("BACKUP_ARTIFACT_UNVERIFIED", found)
        for code in ["GRAPH_BINDING_NOT_ESTABLISHED", "WORKING_SET_BACKUP_REVIEW", "IGNORED_TABLE_BACKUP_REVIEW",
                     "EXISTING_GRAPH_STATE_REQUIRES_CONTINUITY", "BACKUP_RESTORE_CONTINUITY_UNPROVEN"]:
            self.assertIn(code, found)

    def test_bounded_input_duplicate_members_and_bad_shapes_refuse(self):
        with self.assertRaises(ValueError):
            preflight.parse_json('{"schemaVersion":1,"schemaVersion":1}')
        with self.assertRaises(ValueError):
            preflight.parse_json('{"n":NaN}')
        for mutate in [lambda x: x.update(provenance={}), lambda x: x["tables"].update(issues={}),
                       lambda x: x["tables"].update(unknown=[]), lambda x: x.update(artifacts={"x": "file"}),
                       lambda x: x.update(artifacts={"x": {"sha256": "not-a-digest"}})]:
            data = bundle(); mutate(data)
            with self.assertRaises(ValueError):
                preflight.report(data, "digest")
        data = bundle()
        data["tables"]["issues"] = [{}] * (preflight.MAX_ROWS + 1)
        with self.assertRaises(ValueError):
            preflight.report(data, "digest")

    def test_cli_reads_only_input_and_emits_no_partial_success(self):
        data = bundle()
        frozen = copy.deepcopy(data)
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "source.json"
            raw = json.dumps(data).encode(); path.write_bytes(raw)
            before = sorted(Path(directory).iterdir())
            out, err = io.StringIO(), io.StringIO()
            with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
                self.assertEqual(preflight.main([str(path)]), 0)
            self.assertEqual(path.read_bytes(), raw)
            self.assertEqual(sorted(Path(directory).iterdir()), before)
            self.assertEqual(json.loads(out.getvalue())["inputSha256"], hashlib.sha256(raw).hexdigest())
            self.assertEqual(data, frozen)
            path.write_text('{"schemaVersion":')
            out, err = io.StringIO(), io.StringIO()
            with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
                self.assertEqual(preflight.main([str(path)]), 2)
            self.assertEqual(out.getvalue(), "")
            self.assertEqual(json.loads(err.getvalue())["code"], "invalid_export")


if __name__ == "__main__":
    unittest.main()
