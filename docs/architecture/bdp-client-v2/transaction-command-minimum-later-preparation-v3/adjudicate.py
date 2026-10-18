"""Read retained outputs only. Candidate findings never replace root acceptance."""
import json,re,sys,os
from pathlib import Path
from gate import E,load_gate,completed,require,sha
name=sys.argv[1]
g,_=load_gate(name+"-review")
expected=0 if name=="fresh-tx" else 1
exit_record=completed(name+"-test" if name!="fresh-tx" else name,expected)
testid=name+"-test" if name!="fresh-tx" else name
results=json.loads((E/(testid+".results.json")).read_text())
rows=[a for suite in results["testResults"] for a in suite["assertionResults"] if a["status"] not in ("pending","skipped","todo")]
require(len(rows)==1,"Not exactly one selected case")
row=rows[0]
require(not results.get("unhandledErrors"),"Unhandled errors present")
if name!="fresh-tx":
    require(results["numFailedTests"]==1 and results["numPassedTests"]==0 and row["status"]=="failed","Control did not fail exactly one case")
    title=("ignores partial and malformed copied read controls on each mutation carrier" if name=="restore-command"
           else "keeps receipt minimum enforcement after commands ignore copied controls")
    require(row["fullName"]==title,"Wrong failing test")
    messages=row["failureMessages"];require(len(messages)==1,"Multiple failures require root disposition")
    message=re.sub(r"\x1b\[[0-9;]*m","",messages[0])
    marker="copied command accepted before effects: " if name=="restore-command" else "receipt minimum remains enforced: "
    require(marker in message,"Named inner oracle absent")
    observed,_=json.JSONDecoder().raw_decode(message.split(marker,1)[1])
    if name=="restore-command":
        require(observed=={"status":500,"developmentLimit":"minimum-position","contentLength":"0"},"Wrong500 diagnostic")
    else:
        require(observed["status"]==200 and observed["body"]["status"]=="completed" and observed["body"]["idempotencyKey"]=="receipt-minimum","Wrong receipt bypass")
        require(observed["context"]["bdp-scope-position"]=="pos-1","Wrong bypass context")
        require(observed["body"]["requiredPosition"]=="pos-1","Wrong bypass receipt position")
    value={"candidateMatches":True,"control":name,"selectedTitle":title,"diagnostic":observed,
           "expectedPrefixHttp":1 if name=="restore-command" else 2,
           "httpCountBasis":"Reached named oracle in exact reviewed source prefix; no independent request-counter instrumentation.",
           "rootReviewRequired":True,"limitation":"Inspect retained full log/failure messages for unrelated or cleanup failures; this is not acceptance."}
else:
    require(results["success"] and results["numPassedTests"]==1 and results["numFailedTests"]==0,"Installed test did not pass once")
    require(row["fullName"]=="ignores command read controls while enforcing a later installed receipt minimum","Wrong installed test")
    reports=[]
    for line in (E/"fresh-tx.log").read_text().splitlines():
        if line.startswith("{"):
            item=json.loads(line)
            if item.get("observation")=="private-installed-command-minimum":reports.append(item)
    require(len(reports)==1,"Missing/duplicate actual observation report")
    report=reports[0];provenance=json.loads(Path(g["installedProvenance"]["path"]).read_text())
    require(report["outcome"]=="passed" and report["claimEligible"] is False and report["closed"] is True and report["childExitCode"]==0 and not report["aborted"],"Bad installed lifecycle")
    require(report["declaredSourceCommit"]==g["finalHead"] and report["workspaceDirty"] is False,"Wrong source attribution")
    require(report["entry"]==provenance["entryRealpath"] and report["entrySha256"]==provenance["entrySha256"],"Wrong entry")
    served=report["binding"]["served"]
    require(served["sourceCommit"]==g["finalHead"] and served["installedPayloadDigest"]==provenance["payloadIdentity"],"Wrong payload")
    require(served["scope"]=="https://reference.invalid/acme/" and served["profile"]=="transactional" and served["history"] is False,"Wrong private configuration")
    for key in ("installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest"):
        require(re.fullmatch("[0-9a-f]{64}",served[key]) is not None,"Missing binding "+key)
    observations=report["completedResponses"];require(report["exchanges"]==len(observations)==18,"Not18 responses")
    scope=served["scope"];receipt=observations[13]["url"]
    require(re.fullmatch(re.escape(scope)+r"receipts/[A-Za-z0-9_-]+",receipt) is not None and len(receipt.encode())<=8192,"Invalid observed receipt route")
    expected_routes=[
      ("GET","beads/?limit=32",200),("POST","operations/create-bead",200),("GET","beads/a",200),
      ("POST","operations/create-bead",200),("POST","operations/batch",200),("GET","beads/b",200),
      ("POST","operations/batch",200),("POST","operations/sequence",200),("GET","beads/c",200),
      ("POST","operations/sequence",200),("POST","operations/batch",409),("POST","operations/update-bead",200),
      ("GET","beads/a",200),("GET",receipt,503),("GET",receipt,409),("GET",receipt,200),
      ("HEAD",receipt,503),("GET","beads/?limit=32",200)]
    for item,(method,url,status) in zip(observations,expected_routes):
        require((item["method"],item["url"],item["status"])==(method,url if url.startswith("https:") else scope+url,status),"Wire route/status order mismatch")
        require(0<=item["responseBytes"]<=32768,"Response bound")
        for key in ("wireSha256","bodySha256"):require(re.fullmatch("[0-9a-f]{64}",item[key]) is not None,"Missing wire fingerprint")
    require(len({x["url"] for x in observations})==9,"Not nine URLs")
    try:os.kill(report["childPid"],0)
    except ProcessLookupError:pass
    else:raise RuntimeError("Reported child PID still exists; independently inspect before acceptance")
    value={"candidateMatches":True,"observation":report,"rootReviewRequired":True,
           "actualCompletedResponses":18,"declaredLaunches":1,
           "limitation":"Full response assertions execute in the reviewed test; retained report has fingerprints, not full bodies or header values. Root independently checks provenance/configuration/cleanup; no public admission."}
value["resultsSha256"]=sha((E/(testid+".results.json")).read_bytes())
value["logSha256"]=sha((E/(testid+".log")).read_bytes())
with (E/(name+".review-candidate.json")).open("x") as f:f.write(json.dumps(value,indent=2)+"\n")
print(json.dumps({"candidate":name,"matches":True,"rootReviewRequired":True}))
