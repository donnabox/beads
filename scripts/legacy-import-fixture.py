#!/usr/bin/env python3
"""Generate ordinary bd export fixtures using only a disposable synthetic store.

The output must be a fresh directory owned by a single publisher. A sibling
temporary directory is renamed only after both fixture and receipt are complete.
Existing output leaves (including symlinks) are refused, never updated in place.
The prechecks assume no other process creates the destination concurrently;
portable directory rename does not provide an atomic no-replace race guarantee.
"""
import argparse, hashlib, json, os, re, subprocess, tempfile
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--bd", type=Path, required=True)
parser.add_argument("--output", type=Path, required=True)
parser.add_argument("--source-dir", type=Path, required=True, help="Clean checkout used for the canonical producer build")
parser.add_argument("--prefix", default="legacy", help="Ordinary source Issue prefix (completeness qualification uses web-app)")
parser.add_argument("--profile", choices=("basic", "completeness"), default="basic", help="Keep the original fixture or exercise prefix, field and memory-key compatibility")
a = parser.parse_args()
output = a.output.absolute()
if os.path.lexists(output):
    parser.error("output must be a fresh directory; existing output is never replaced")
binary = a.bd.resolve()
source_dir = a.source_dir.resolve()
source_env = {k:v for k,v in os.environ.items() if not k.startswith(("BD_", "BEADS_", "DOLT_", "GIT_"))}
def source_state():
    def git(*args):
        return subprocess.check_output(["git", *args], cwd=source_dir, env=source_env, text=True, timeout=30).strip()
    head, tree = git("rev-parse", "HEAD"), git("rev-parse", "HEAD^{tree}")
    if git("status", "--porcelain"):
        parser.error("producer source checkout must be clean")
    return head, tree
source, source_tree = source_state()
binary_hash = hashlib.sha256(binary.read_bytes()).hexdigest()
receipts = []
with tempfile.TemporaryDirectory(prefix="iris-legacy-producer-") as temporary:
    root = Path(temporary)
    work, home = root / "workspace", root / "home"
    work.mkdir(); home.mkdir()
    env = {k:v for k,v in os.environ.items() if not k.startswith(("BD_", "BEADS_", "DOLT_", "GIT_"))}
    env.update(HOME=str(home), XDG_CONFIG_HOME=str(home/"config"), GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL=str(home/"gitconfig"),
               BD_DISABLE_METRICS="1", BD_DISABLE_EVENT_FLUSH="1", DOLT_METRICS_DISABLED="1", DOLT_DISABLE_EVENT_FLUSH="1", BEADS_DOLT_AUTO_START="0", GIT_TERMINAL_PROMPT="0")
    (home/"gitconfig").write_text("[user]\n name = Synthetic Fixture\n email = fixture@example.invalid\n")
    def call(*args):
        process = subprocess.run([str(binary),*args], cwd=work, env=env, capture_output=True, text=True, timeout=90)
        receipts.append(dict(args=list(args), code=process.returncode, stdout=process.stdout, stderr=process.stderr))
        if process.returncode: raise RuntimeError(json.dumps(receipts[-1]))
        return process.stdout
    # Check structured build identity rather than finding HEAD text in a branch
    # name or other human output. Older canonical builds identify HEAD in build;
    # newer producers also expose commit. Both must agree when present.
    version_info = json.loads(call("version", "--json"))
    identities = [version_info.get("build")] if isinstance(version_info, dict) else [None]
    if isinstance(version_info, dict) and "commit" in version_info:
        identities.append(version_info["commit"])
    if any(not isinstance(reported, str) or not re.fullmatch(r"[0-9a-f]{9,40}", reported) or not source.startswith(reported) for reported in identities):
        raise RuntimeError("producer build version does not match clean source HEAD")
    version = call("version")
    subprocess.run(["git","init","-q"], cwd=work, env=env, check=True)
    subprocess.run(["git","config","core.hooksPath","/dev/null"], cwd=work, env=env, check=True)
    complete = a.profile == "completeness"
    prerequisite = a.prefix + ("-a1b" if complete else "-a")
    dependent = a.prefix + ("-b2c" if complete else "-b")
    call("init","--prefix",a.prefix,"--skip-hooks","--skip-agents","--non-interactive")
    call("create","Prerequisite","--id",prerequisite,"--priority","0","--actor","fixture-author","--json")
    fields = []
    if complete:
        fields = ["--type","bug","--assignee","fixture-worker","--estimate","20","--external-ref","gh-fixture-123","--spec-id","spec-fixture","--due","2099-01-02","--defer","2099-01-01","--mol-type","swarm"]
    call("create","Dependent work — 記憶","--id",dependent,"--description","Body\nSecond line","--design","A design","--acceptance","A check","--notes","A note","--labels","one,two","--metadata",'{"custom":{"flag":true},"list":[1,"value"]}',"--actor","fixture-author","--json",*fields)
    call("dep","add",dependent,prerequisite,"--actor","fixture-author","--json")
    call("comments","add",dependent,"A preserved comment\nSecond line","--author","fixture-commenter","--json")
    call("close",prerequisite,"--reason","Synthetic completion","--actor","fixture-author","--json")
    call("remember","Remembered context","--key","fixture-key")
    if complete:
        call("create","Persistent context","--id",a.prefix+"-abcdef","--type","decision","--status","pinned","--labels","context,記憶","--metadata",'{"unicode":"記憶","empty":{},"number":0}',"--actor","fixture-author","--json")
        call("create","Operational record","--id",a.prefix+"-d4e","--type","event","--event-category","fixture.started","--event-actor","fixture-actor","--event-target",dependent,"--event-payload",'{"synthetic":true}',"--actor","fixture-author","--json")
        # Reserve direct canonical spellings as well as their decoded keys so
        # importer allocation must be independent of exported record order.
        for key in ["has space","has%20space"," leading","trailing ","記憶","%E8%A8%98%E6%86%B6","100%",".","..","path/part","double//separator","back\\slash","tab\tline\nbreak",prerequisite]:
            call("remember","Value for "+key,"--key",key)
        # Empty values are valid config-backed export records even though the
        # remember front door correctly refuses empty memory content.
        call("config","set","kv.memory.empty-value","")
    exported = call("export","--include-memories")
    if source_state() != (source, source_tree):
        raise RuntimeError("producer source changed during fixture generation")
    if hashlib.sha256(binary.read_bytes()).hexdigest() != binary_hash:
        raise RuntimeError("producer binary changed during fixture generation")
    fixture_bytes = exported.encode("utf-8")
    provenance = dict(producerSource=source, producerSourceTree=source_tree, sourceState="clean checkout; canonical make install-force build", binarySHA256=binary_hash,
                      producerVersion=version, fixtureSHA256=hashlib.sha256(fixture_bytes).hexdigest(),
                      data="exclusively synthetic disposable workspace; no production inputs", profile=a.profile, issuePrefix=a.prefix, receipts=receipts)
    receipt_bytes = (json.dumps(provenance, ensure_ascii=False, indent=2)+"\n").encode("utf-8")
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="."+output.name+"-staging-", dir=output.parent) as pending:
        staging = Path(pending)
        (staging/"ordinary.jsonl").write_bytes(fixture_bytes)
        (staging/"producer.json").write_bytes(receipt_bytes)
        if os.path.lexists(output):
            parser.error("output appeared during generation; existing output is never replaced")
        staging.rename(output)
print(str(output/"ordinary.jsonl"))
