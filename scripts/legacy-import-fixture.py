#!/usr/bin/env python3
"""Generate ordinary bd export fixtures using only a disposable synthetic store."""
import argparse, hashlib, json, os, subprocess, tempfile
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--bd", type=Path, required=True)
parser.add_argument("--output", type=Path, required=True)
parser.add_argument("--source-dir", type=Path, required=True, help="Clean checkout used for the canonical producer build")
a = parser.parse_args()
binary = a.bd.resolve()
source_dir = a.source_dir.resolve()
source = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=source_dir, text=True).strip()
source_tree = subprocess.check_output(["git", "rev-parse", "HEAD^{tree}"], cwd=source_dir, text=True).strip()
if subprocess.check_output(["git", "status", "--porcelain"], cwd=source_dir, text=True).strip():
    parser.error("producer source checkout must be clean")
a.output.mkdir(parents=True, exist_ok=True)
receipts = []
with tempfile.TemporaryDirectory(prefix="iris-legacy-producer-") as temporary:
    root = Path(temporary)
    work, home = root / "workspace", root / "home"
    work.mkdir(); home.mkdir()
    env = {k:v for k,v in os.environ.items() if not k.startswith(("BD_", "BEADS_", "DOLT_", "GIT_"))}
    env.update(HOME=str(home), XDG_CONFIG_HOME=str(home/"config"), GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL=str(home/"gitconfig"),
               BD_DISABLE_METRICS="1", BD_DISABLE_EVENT_FLUSH="1", DOLT_METRICS_DISABLED="1", DOLT_DISABLE_EVENT_FLUSH="1", BEADS_DOLT_AUTO_START="0", GIT_TERMINAL_PROMPT="0")
    (home/"gitconfig").write_text("[user]\n name = Synthetic Fixture\n email = fixture@example.invalid\n")
    subprocess.run(["git","init","-q"], cwd=work, env=env, check=True)
    subprocess.run(["git","config","core.hooksPath","/dev/null"], cwd=work, env=env, check=True)
    def call(*args):
        process = subprocess.run([str(binary),*args], cwd=work, env=env, capture_output=True, text=True, timeout=90)
        receipts.append(dict(args=list(args), code=process.returncode, stdout=process.stdout, stderr=process.stderr))
        if process.returncode: raise RuntimeError(json.dumps(receipts[-1]))
        return process.stdout
    call("init","--prefix","legacy","--skip-hooks","--skip-agents","--non-interactive")
    call("create","Prerequisite","--id","legacy-a","--priority","0","--actor","fixture-author","--json")
    call("create","Dependent work — 記憶","--id","legacy-b","--description","Body\nSecond line","--design","A design","--acceptance","A check","--notes","A note","--labels","one,two","--metadata",'{"custom":{"flag":true},"list":[1,"value"]}',"--actor","fixture-author","--json")
    call("dep","add","legacy-b","legacy-a","--actor","fixture-author","--json")
    call("comments","add","legacy-b","A preserved comment\nSecond line","--author","fixture-commenter","--json")
    call("close","legacy-a","--reason","Synthetic completion","--actor","fixture-author","--json")
    call("remember","Remembered context","--key","fixture-key")
    exported = call("export","--include-memories")
    (a.output/"ordinary.jsonl").write_text(exported)
    version = call("version")
    if source[:9] not in version:
        raise RuntimeError("producer build version does not match clean source HEAD")
    provenance = dict(producerSource=source, producerSourceTree=source_tree, sourceState="clean checkout; canonical make install-force build", binarySHA256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                      producerVersion=version, fixtureSHA256=hashlib.sha256(exported.encode()).hexdigest(),
                      data="exclusively synthetic disposable workspace; no production inputs", receipts=receipts)
    (a.output/"producer.json").write_text(json.dumps(provenance, ensure_ascii=False, indent=2)+"\n")
print(str(a.output/"ordinary.jsonl"))
