"""Retain fresh build outputs only after exact thirty selected source results; no product import."""
from gate import load_gate,E,W,build_outputs,require
import json
from pathlib import Path
g,s=load_gate("build-inventory")
c=json.loads((Path(__file__).parent/"commands.json").read_text())["commands"]["static-cases"]
r=json.loads((E/"static-cases.results.json").read_text())
actual=[(str(Path(suite["name"]).resolve().relative_to(W)),a["fullName"],a["status"]) for suite in r["testResults"] for a in suite["assertionResults"] if a["status"] not in ("pending","skipped","todo")]
expected=[(a["path"],a["fullName"],"passed") for a in c["selectedCases"]]
require(sorted(actual)==sorted(expected) and len(actual)==30,"Unexpected selected static callbacks")
require(r["success"] and r["numPassedTests"]==30 and r["numFailedTests"]==0 and not r.get("unhandledErrors"),"Static callbacks failed")
with (E/"build-output-inventory.json").open("x") as f:f.write(json.dumps({"sourceSha256":g["sourceSha256"],"files":build_outputs(),"staticSelected":actual},indent=2)+"\n")
