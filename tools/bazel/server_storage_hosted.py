#!/usr/bin/env python3
"""Capture separate hosted invocations and check their exact storage-tier union."""

import argparse
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
import xml.etree.ElementTree as ET

import check_shard_coverage as coverage
from equivalence import read_bep, testlog_xmls

SCRIPT = ".github/scripts/server-storage-test-shard.sh"
PREFIX = "//internal/storage/dolt:"
PARTS = {str(i): (PREFIX + f"dolt_server_hosted_part{i}_test", 4) for i in range(4)}
PARTS["conformance"] = (PREFIX + "dolt_server_conformance_test", 1)


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def invocation(path, part):
    label, count = PARTS[part]
    events = [json.loads(line) for line in Path(path).read_text().splitlines() if line.strip()]
    starts = [e["started"] for e in events if "started" in e]
    ends = [e["finished"] for e in events if "finished" in e]
    require(len(starts) == len(ends) == 1, "missing or ambiguous invocation")
    require(starts[0].get("command") == "test" and starts[0].get("uuid"), "not a test invocation")
    require(ends[0].get("overallSuccess") and ends[0].get("exitCode", {}).get("name") == "SUCCESS" and
            ends[0].get("exitCode", {}).get("code", 0) == 0,
            "invocation failed")
    results = [e for e in events if "testResult" in e.get("id", {})]
    require(len(results) == count, "missing, duplicate or unexpected test results")
    shards = set()
    for event in results:
        identity, result = event["id"]["testResult"], event["testResult"]
        require(identity.get("label") == label, "unexpected target; canonical full target must not run here")
        require(int(identity.get("run", 1)) == int(identity.get("attempt", 1)) == 1, "retried/repeated test")
        index = int(identity.get("shard", 1))
        require(index not in shards and 1 <= index <= count, "duplicate or invalid shard")
        shards.add(index)
        require(result.get("status") == "PASSED", "test result not passed")
        require(not result.get("cachedLocally") and not result.get("cachedRemotely") and
                not result.get("executionInfo", {}).get("cachedRemotely"),
                "cached result cannot prove fresh hosted execution")
    require(shards == set(range(1, count + 1)), "incomplete shard population")
    tested, _, logs = read_bep(path)
    require(tested == {label: count}, "configured/result label population differs")
    return starts[0]["uuid"], logs


def capture(args):
    folder = Path(args.directory) / args.part
    folder.mkdir(parents=True, exist_ok=False)
    # Retain the original even when admission fails; never manufacture a union BEP.
    shutil.copyfile(args.bep, folder / "bep.json")
    uuid, logs = invocation(folder / "bep.json", args.part)
    require(logs is not None, "missing original testlogs location")
    label, count = PARTS[args.part]
    files = {"bep.json": digest(folder / "bep.json")}
    for xml in testlog_xmls(logs, label, count):
        for source in (Path(xml), Path(xml).with_name("test.log")):
            require(source.is_file(), "missing original test XML/log: " + str(source))
            destination = folder / "testlogs" / source.relative_to(logs)
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, destination)
            require(digest(source) == digest(destination), "artifact copy drift")
            files[str(destination.relative_to(folder))] = digest(destination)
    for name in ("bazel-exec.log.zst", "bazel-profile.json", "bazel-test.log"):
        source = Path(args.bep).parent / name
        require(source.is_file(), "missing original diagnostic: " + name)
        shutil.copyfile(source, folder / name)
        files[name] = digest(folder / name)
    metadata = dict(part=args.part, source=args.source, run=args.run, attempt=args.attempt,
                    invocation=uuid, files=files)
    (folder / "metadata.json").write_text(json.dumps(metadata, indent=2) + "\n")


def verify_parts(args, logs, lister=coverage.listed_tests):
    root = Path(args.directory)
    expected_artifacts = {f"server-storage-hosted-{part}-{args.attempt}" for part in PARTS}
    require({p.name for p in root.iterdir()} == expected_artifacts, "missing, duplicate or unexpected artifact identity")
    tested, invocations = {}, set()
    for part, (label, count) in PARTS.items():
        artifact = root / f"server-storage-hosted-{part}-{args.attempt}"
        require({p.name for p in artifact.iterdir()} == {part}, "artifact contains an unexpected part")
        folder = artifact / part
        metadata = json.loads((folder / "metadata.json").read_text())
        require((metadata["part"], metadata["source"], metadata["run"], metadata["attempt"]) ==
                (part, args.source, args.run, args.attempt), "wrong part/source/run/attempt")
        require(not any(p.is_symlink() for p in folder.rglob("*")), "symlink artifact member")
        actual_files = {str(p.relative_to(folder)) for p in folder.rglob("*") if p.is_file() and p.name != "metadata.json"}
        require(set(metadata["files"]) == actual_files, "unbound or missing artifact member")
        require({"bep.json", "bazel-exec.log.zst", "bazel-profile.json", "bazel-test.log"} <= actual_files, "missing original diagnostics")
        for name, expected in metadata["files"].items():
            relative = Path(name)
            require(not relative.is_absolute() and ".." not in relative.parts, "unsafe artifact path")
            require(digest(folder / relative) == expected, "artifact hash drift")
        uuid, _ = invocation(folder / "bep.json", part)
        require(uuid == metadata["invocation"] and uuid not in invocations, "duplicate invocation")
        invocations.add(uuid)
        for xml in testlog_xmls(folder / "testlogs", label, count):
            require(Path(xml).is_file(), "missing XML")
            tree = ET.parse(xml)
            require(not tree.findall(".//failure") and not tree.findall(".//error"), "failed/error XML")
            require(tree.findall(".//testcase"), "empty XML")
            names = [c.get("name", "") for c in tree.findall(".//testcase") if c.get("name", "").startswith("Test") and "/" not in c.get("name", "")]
            require(len(names) == len(set(names)), "duplicate top-level test case")
            if part == "conformance":
                require(names == ["TestConformance"], "conformance root missing or substituted")
            destination = logs / Path(xml).relative_to(folder / "testlogs")
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(xml, destination)
        tested[label] = count
    suites = [(PARTS[str(i)][0], SCRIPT, 4, i * 4, 16) for i in range(4)]
    lines, problems = coverage.check(tested, str(logs), suites, lister=lister, whole=[PARTS["conformance"][0]])
    require(not problems, "\n".join(problems))
    print("\n".join(lines))
    print("Five independent current invocations: all16 original shards plus conformance verified")


def verify(args):
    # A new derived view permits offline reanalysis without a test rerun.
    with tempfile.TemporaryDirectory(prefix="server-storage-verified-") as directory:
        verify_parts(args, Path(directory))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("operation", choices=("capture", "verify"))
    parser.add_argument("--directory", required=True)
    parser.add_argument("--source", required=True)
    parser.add_argument("--run", required=True)
    parser.add_argument("--attempt", required=True)
    parser.add_argument("--part", choices=tuple(PARTS))
    parser.add_argument("--bep")
    args = parser.parse_args()
    if args.operation == "capture":
        if not args.part or not args.bep:
            parser.error("capture requires --part and --bep")
        capture(args)
    else:
        verify(args)


if __name__ == "__main__":
    main()
