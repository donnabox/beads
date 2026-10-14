"""Draft source-only gate. No command is authorized by this file."""
from pathlib import Path
import hashlib
import json
import subprocess

P = Path(__file__).resolve().parent
W = Path("/Users/dbox/repos/wt/vickie/vickie-command-minimum")
E = P / "run"
PARENT = "122800511ee6306e4105e1f73bf3658f35c90a14"
TREE = "088290751951c353797b984f690b7d9fd96deb80"
OWNED = {
    "docs/specs/bdp.md",
    "packages/server/src/transaction-http.ts",
    "packages/server/src/transaction-http.test.ts",
    "packages/server/src/transaction-service.test.ts",
    "apps/bdptest/src/transactional-collection-minimum-installed.test.ts",
    "apps/bdptest/src/installed-reference.test.ts",
    "docs/design/transaction-implementation-progress.md",
}
sha = lambda data: hashlib.sha256(data).hexdigest()

def require(condition, message):
    if not condition:
        raise RuntimeError(message)

def git(*args):
    return subprocess.check_output(["git", *args], cwd=W)

def source_hashes():
    return {name: sha((W / name).read_bytes()) for name in sorted(OWNED)}

def load_gate(name):
    # No execution-gate.json is supplied by this preparation.
    gate = json.loads((P / "execution-gate.json").read_text())
    require(gate.get("released") is True, "Execution gate is closed")
    require(gate.get("authorizationKind") == "measured-source-execution", "Measured source release required")
    acceptance = Path(gate["rootAcceptancePath"])
    require(acceptance.is_absolute(), "Absolute root acceptance path required")
    require(sha(acceptance.read_bytes()) == gate["rootAcceptanceSha256"], "Root acceptance changed")
    require(sha((P.parent / "transaction-command-minimum-implementation-plan-v2.md").read_bytes()) ==
            "f8824197e765e8228679044ab32fd7a54ebd26ec3a892614fdd9b5a15cf07c82", "Plan changed")
    required = {"commands.json", "gate.py", "supervise.py"}
    require(set(gate["preparationSha256"]) == required, "Preparation bindings incomplete")
    for rel, digest in gate["preparationSha256"].items():
        require(sha((P / rel).read_bytes()) == digest, "Preparation changed: " + rel)
    ledger = json.loads((P / "commands.json").read_text())
    spec = ledger["commands"][name]
    require(spec["phase"] == "source" and spec["runnable"] is True, "Later-phase binding not prepared")
    require(name in gate["approvedCommandIds"], "Command not released: " + name)
    require(spec["cwd"] == str(W), "Wrong checkout")
    require(spec["supervisorSeconds"] == 120, "Unexpected supervisor budget")
    require(git("rev-parse", "HEAD").decode().strip() == PARENT, "Draft parent changed")
    require(git("rev-parse", "HEAD^{tree}").decode().strip() == TREE, "Parent tree changed")
    require(set(gate["sourceSha256"]) == OWNED, "Seven source hashes required")
    require(source_hashes() == gate["sourceSha256"], "Draft source changed")
    changes = set(git("diff", "--name-only", "HEAD").decode().splitlines())
    require(changes <= OWNED, "Unowned tracked change")
    require(not git("ls-files", "--others", "--exclude-standard").strip(), "Untracked source present")
    require(sha(git("diff", "--binary", "HEAD")) == gate["sourceDiffSha256"], "Draft diff changed")
    # A partial or failed earlier attempt blocks later commands, even if the
    # caller accidentally leaves them approved. No hidden retries/waivers.
    for started in E.glob("*.started.json"):
        prior = json.loads(started.read_text())
        result_path = E / (started.name.removesuffix(".started.json") + ".exit.json")
        require(result_path.exists(), "Earlier attempt has no final result")
        result = json.loads(result_path.read_text())
        require(prior["sourceSha256"] == gate["sourceSha256"], "Earlier attempt used different source")
        require(result.get("supervisorExit") == 0, "Earlier attempt needs root disposition")
    for dependency in spec.get("requires", []):
        result = json.loads((E / (dependency + ".exit.json")).read_text())
        require(result.get("supervisorExit") == 0, "Prerequisite did not finish cleanly: " + dependency)
    return gate, spec
