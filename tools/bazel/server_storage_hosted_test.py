"""Exercise capture and the fail-closed cross-invocation gate on synthetic evidence."""

import argparse
import json
from pathlib import Path
import tempfile
import unittest

import server_storage_hosted as proof


class HostedProofTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name)
        self.download = self.base / "download"
        self.download.mkdir()
        self.args = argparse.Namespace(directory=str(self.download), source="qualified-head", run="123", attempt="1")
        for part, (label, count) in proof.PARTS.items():
            original = self.base / ("original-" + part)
            output = original / "output"
            logs = output / "bazel-out/k8-fastbuild/testlogs"
            xmls = proof.testlog_xmls(logs, label, count)
            for k, name in enumerate(xmls, 1):
                xml = Path(name)
                xml.parent.mkdir(parents=True, exist_ok=True)
                test = "TestConformance" if part == "conformance" else "TestCase" + str(int(part) * 4 + k)
                xml.write_text(f'<testsuite><testcase name="{test}"/></testsuite>')
                xml.with_name("test.log").write_text("original log\n")
            events = [
                {"id": {"started": {}}, "started": {"command": "test", "uuid": "invocation-" + part}},
                {"id": {"workspace": {}}, "workspaceInfo": {"localExecRoot": str(output / "execroot/_main")}},
                {"id": {"convenienceSymlinksIdentified": {}}, "convenienceSymlinksIdentified": {
                    "convenienceSymlinks": [{"path": "bazel-testlogs", "target": "bazel-out/k8-fastbuild/testlogs"}]}},
                {"id": {"targetConfigured": {"label": label}}, "configured": {"targetKind": "sh_test rule"}},
            ]
            events += [{"id": {"testResult": {"label": label, "shard": k, "run": 1, "attempt": 1}},
                        "testResult": {"status": "PASSED"}} for k in range(1, count + 1)]
            events.append({"id": {"finished": {}}, "finished": {"overallSuccess": True, "exitCode": {"name": "SUCCESS"}}})
            bep = original / "bep.json"
            bep.write_text("".join(json.dumps(e) + "\n" for e in events))
            for n in ("bazel-exec.log.zst", "bazel-profile.json", "bazel-test.log"):
                (original / n).write_text("original diagnostic\n")
            captured = self.base / ("captured-" + part)
            capture_args = argparse.Namespace(**vars(self.args))
            capture_args.directory, capture_args.part, capture_args.bep = str(captured), part, str(bep)
            proof.capture(capture_args)
            artifact = self.download / f"server-storage-hosted-{part}-1"
            artifact.mkdir()
            (captured / part).rename(artifact / part)

    def folder(self, part="0"):
        return self.download / f"server-storage-hosted-{part}-1" / part

    def rehash(self, part="0"):
        folder = self.folder(part)
        path = folder / "metadata.json"
        data = json.loads(path.read_text())
        data["files"] = {str(p.relative_to(folder)): proof.digest(p) for p in folder.rglob("*")
                         if p.is_file() and p != path}
        path.write_text(json.dumps(data))

    def verify(self):
        logs = self.base / "derived"
        logs.mkdir(exist_ok=True)
        proof.verify_parts(self.args, logs, lister=lambda script, shard, total: {f"TestCase{shard}"})

    def test_partition_identity_population(self):
        self.assertEqual(proof.PARTS, {
            "0": ("//internal/storage/dolt:dolt_server_hosted_part0_test", 4),
            "1": ("//internal/storage/dolt:dolt_server_hosted_part1_test", 4),
            "2": ("//internal/storage/dolt:dolt_server_hosted_part2_test", 4),
            "3": ("//internal/storage/dolt:dolt_server_hosted_part3_test", 4),
            "conformance": ("//internal/storage/dolt:dolt_server_conformance_test", 1),
        })

    def test_complete_exact_union(self):
        self.verify()
        # Analysis is repeatable; the original invocations are never rerun.
        self.verify()

    def test_missing_or_unexpected_artifact(self):
        for bad in ("missing", "unexpected"):
            with self.subTest(bad=bad):
                path = self.download / "server-storage-hosted-0-1"
                moved = (self.base / "saved") if bad == "missing" else (self.download / "extra")
                path.rename(moved)
                with self.assertRaises(ValueError):
                    self.verify()
                moved.rename(path)

    def test_wrong_source_run_attempt(self):
        path = self.folder() / "metadata.json"
        original = path.read_text()
        for field in ("source", "run", "attempt", "part"):
            with self.subTest(field=field):
                data = json.loads(original)
                data[field] = "wrong"
                path.write_text(json.dumps(data))
                with self.assertRaises(ValueError):
                    self.verify()
                path.write_text(original)

    def test_failed_cached_retried_or_duplicate_result(self):
        path = self.folder() / "bep.json"
        original = path.read_text()
        for mode in ("failed", "cached", "retry", "duplicate", "old-target", "missing", "zero-shard"):
            with self.subTest(mode=mode):
                events = [json.loads(line) for line in original.splitlines()]
                result = next(e for e in events if "testResult" in e)
                if mode == "failed":
                    result["testResult"]["status"] = "FAILED"
                elif mode == "cached":
                    result["testResult"]["cachedLocally"] = True
                elif mode == "retry":
                    result["id"]["testResult"]["attempt"] = 2
                elif mode == "duplicate":
                    events.append(result)
                elif mode == "old-target":
                    result["id"]["testResult"]["label"] = proof.PREFIX + "dolt_server_full_test"
                elif mode == "zero-shard":
                    result["id"]["testResult"]["shard"] = 0
                else:
                    events.remove(result)
                path.write_text("".join(json.dumps(e) + "\n" for e in events))
                self.rehash()
                with self.assertRaises(ValueError):
                    self.verify()
                path.write_text(original)
                self.rehash()

    def test_failed_or_missing_finish(self):
        path = self.folder() / "bep.json"
        original = path.read_text()
        for mode in ("failed", "missing", "nonzero"):
            with self.subTest(mode=mode):
                events = [json.loads(line) for line in original.splitlines()]
                end = next(e for e in events if "finished" in e)
                if mode == "missing":
                    events.remove(end)
                elif mode == "nonzero":
                    end["finished"]["exitCode"]["code"] = 1
                else:
                    end["finished"]["overallSuccess"] = False
                path.write_text("".join(json.dumps(e) + "\n" for e in events))
                self.rehash()
                with self.assertRaises(ValueError):
                    self.verify()
                path.write_text(original)
                self.rehash()

    def test_duplicate_invocation(self):
        path = self.folder("1") / "bep.json"
        events = [json.loads(line) for line in path.read_text().splitlines()]
        next(e for e in events if "started" in e)["started"]["uuid"] = "invocation-0"
        path.write_text("".join(json.dumps(e) + "\n" for e in events))
        metadata = self.folder("1") / "metadata.json"
        data = json.loads(metadata.read_text())
        data["invocation"] = "invocation-0"
        metadata.write_text(json.dumps(data))
        self.rehash("1")
        with self.assertRaises(ValueError):
            self.verify()

    def test_bad_xml_or_shard_assignment(self):
        xml = next((self.folder() / "testlogs").rglob("test.xml"))
        original = xml.read_text()
        for body in ("", '<testcase name="TestCase1"><skipped/></testcase>',
                     '<testcase name="TestCase1"><failure/></testcase>',
                     '<testcase name="TestWrongShard"/>',
                     '<testcase name="TestCase1"/><testcase name="TestCase1"/>'):
            with self.subTest(body=body):
                xml.write_text("<testsuite>" + body + "</testsuite>")
                self.rehash()
                with self.assertRaises(ValueError):
                    self.verify()
                xml.write_text(original)
                self.rehash()

    def test_conformance_root_is_not_substitutable(self):
        xml = next((self.folder("conformance") / "testlogs").rglob("test.xml"))
        xml.write_text('<testsuite><testcase name="TestWrongRoot"/></testsuite>')
        self.rehash("conformance")
        with self.assertRaises(ValueError):
            self.verify()

    def test_hash_drift_and_unbound_file(self):
        log = self.folder() / "bazel-test.log"
        log.write_text("changed")
        with self.assertRaises(ValueError):
            self.verify()
        self.rehash()
        (self.folder() / "unbound").write_text("extra")
        with self.assertRaises(ValueError):
            self.verify()


if __name__ == "__main__":
    unittest.main()
