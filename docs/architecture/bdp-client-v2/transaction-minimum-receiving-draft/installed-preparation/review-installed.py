"""Read actual stdout observations; no product imports, dispatch or target mutation."""
import json,re,sys,subprocess,hashlib
from pathlib import Path
from urllib.parse import urlsplit
from gate import P,W,E,load_gate,completed,require,sha,document,verify_build
def exact(value,keys,label):require(isinstance(value,dict) and set(value)==set(keys.split()),label+" fields differ")
def digest(v):return isinstance(v,str) and re.fullmatch("[0-9a-f]{64}",v) is not None
def integer(v):return type(v) is int and v>=0
def absent(pids):
    rows=subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,comm="],text=True,timeout=10).splitlines()
    alive={int(row.split(None,1)[0]) for row in rows if row.strip()}
    require(not (set(pids)&alive),"Observed PID remains present")
    return sorted(pids)
def main():
    require(len(sys.argv)==2,"One cohort required")
    name=sys.argv[1];g,_=load_gate(name+"-review");result=completed(name)
    cfg=json.loads((P/"commands.json").read_text());spec=cfg["commands"][name]
    contracts=json.loads((P/"observation-contract.json").read_text());contract=contracts["cohorts"][name]
    provenance=document(g,"installedProvenance");verify_build(g)
    # Re-read actual Vitest JSON, independently of the supervisor's compact adjudication.
    tests=json.loads((E/(name+".results.json")).read_text())
    cases=[c for f in tests["testResults"] for c in f.get("assertionResults",[]) if c["status"] not in ("pending","skipped","todo")]
    require(len(cases)==1 and cases[0]["fullName"]==spec["selectedFullNames"][0] and cases[0]["status"]=="passed","Exact callback did not pass")
    require(tests.get("success") is True and tests["numFailedTests"]==0 and tests["numPassedTests"]==1,"Test aggregate failed")
    require(not tests.get("errors") and not tests.get("unhandledErrors") and not tests.get("numRuntimeErrorTestSuites",0),"Unhandled test error")
    require(all(not f.get("message") and not f.get("failureMessage") for f in tests["testResults"]),"Suite failure")
    log=E/(name+".log");records=[]
    for line in log.read_text().splitlines():
        stripped=re.sub(r"\x1b\[[0-9;]*m","",line).strip()
        if '"observation"' not in stripped:continue
        require(stripped.startswith("{"),"Malformed observation line")
        row=json.loads(stripped);require(row.get("observation")==contract["label"],"Unexpected observation label");records.append(row)
    require(len(records)==contract["records"],"Observation count differs")
    scope="https://reference.invalid/acme/";pids=[];fixture=(W/"packages/conformance/test-support/transactional/fixture-v1.json").read_bytes()
    fx=json.loads(fixture)
    request_plan=json.loads((P/"requests.json").read_text())
    for ri,o in enumerate(records):
        common="observation outcome declaredSourceCommit workspaceDirty entry entrySha256 binding exchanges completedResponses claimEligible"
        tail=("artifactBytes attempted deadlineExceeded cleanupExceeded launchAttempted startupCompleted closed childPid childExitCode childSignalCode directory directoryAbsent failureCount failureCategories retention startupLimitation" if name=="minimum-positive" else "aborted children")
        exact(o,common+" "+tail,"Observation")
        require(o["outcome"]=="passed" and o["claimEligible"] is False and o["workspaceDirty"] is False,"Failed/dirty/public report")
        require(o["declaredSourceCommit"]==g["finalHead"] and o["entry"]==provenance["entryRealpath"] and o["entrySha256"]==provenance["entrySha256"],"Source/entry mismatch")
        b=o["binding"];exact(b,"served operationTargets runtimeConfigurationIdentity authorityMode","Binding")
        s=b["served"];exact(s,"scope profile history sourceCommit installedPayloadDigest installedContractsDigest policyDigest authenticationDigest fixtureDigest fixtureId capabilities","Served")
        require(s["scope"]==scope and s["profile"]=="transactional" and s["history"] is False and s["sourceCommit"]==g["finalHead"],"Wrong served source/profile")
        require(s["installedPayloadDigest"]==provenance["payloadIdentity"] and s["fixtureDigest"]==sha(fixture) and s["fixtureId"]==fx["id"] and s["capabilities"]==sorted(fx["capabilities"]),"Payload/fixture mismatch")
        require(b["authorityMode"]=="canonical" and digest(b["runtimeConfigurationIdentity"]) and all(digest(s[k]) for k in ("installedPayloadDigest","installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest")),"Invalid identity")
        names={"createBead":"create-bead","updateBead":"update-bead","deleteBead":"delete-bead","createLink":"create-link","updateLink":"update-link","deleteLink":"delete-link","putAlias":"put-alias","deleteAlias":"delete-alias","updateWhere":"update-where","deleteWhere":"delete-where","batch":"batch","sequence":"sequence"}
        require(b["operationTargets"]=={k:scope+"operations/"+v for k,v in names.items()},"Target inventory differs")
        rows=o["completedResponses"];require(o["exchanges"]==len(rows)==contract["exchanges"][ri],"Exchange count differs")
        require([r["status"] for r in rows]==contract["statuses"][ri],"Response status vector differs")
        for r in rows:
            u=urlsplit(r["url"]);require(u.scheme=="https" and u.netloc=="reference.invalid" and u.path.startswith("/acme/") and not u.fragment and not u.username and not u.password,"Response URL outside Scope")
        if name=="minimum-positive":
            require(o["attempted"]==22 and o["failureCount"]==0 and o["failureCategories"]==[],"Failure/attempt count differs")
            require(all(o[k] is True for k in ("launchAttempted","startupCompleted","closed","directoryAbsent")) and all(o[k] is False for k in ("deadlineExceeded","cleanupExceeded")),"Minimum cleanup failed")
            children=[{"pid":o["childPid"],"exitCode":o["childExitCode"],"signalCode":o["childSignalCode"]}]
            directory=Path(o["directory"]);require(directory.is_relative_to(E/(name+".tmp")) and directory.name.startswith("bdp-minimum-receiving-installed-") and not directory.exists(),"Minimum directory survives/escaped")
            require(integer(o["artifactBytes"]) and o["artifactBytes"]>0,"Missing artifact size")
            actual_receipt=rows[9]["url"];actual_page=rows[11]["url"]
            require(re.fullmatch(re.escape(scope)+r"receipts/[A-Za-z0-9_-]+",actual_receipt) is not None,"Invalid observed receipt root")
            require(re.fullmatch(re.escape(actual_receipt)+r"/pages/[A-Za-z0-9_-]+",actual_page) is not None,"Invalid observed page root")
            occurrences=[[1,0,0],[0,0,0],[0,0,0],[1,0,0],[1,0,1],[2,1,1],[1,2,1],[1,1,2],[1,1,1]]+[[1,0,0]]*4+[[1,1,1]]*7+[[1,0,0]]*2
            roles=["writer"]*10+["other","writer","other","other","other"]+["writer"]*5+["anonymous","writer"]
            for r,q in zip(rows,request_plan["installed"]):
                exact(r,"ordinal label method url credential requestBodyBytes requestBodySha256 minimumFieldOccurrences status bodyOctets bodySha256 headers","Minimum response")
                require([r["ordinal"],r["label"],r["method"],r["status"],r["requestBodyBytes"],r["requestBodySha256"]]==[q["ordinal"],q["label"],q["method"],q["expectedStatus"],q["bodyBytes"],q["bodySha256"]],"Request fingerprint/identity differs")
                target=q["target"]
                require(r["url"]==({"receipt":actual_receipt,"page":actual_page}.get(target,scope+target)),"Literal/retained route differs")
                index=q["ordinal"]-1
                require(r["credential"]==roles[index] and r["minimumFieldOccurrences"]==occurrences[index],"Credential-role/header occurrence ledger differs")
                require(integer(r["bodyOctets"]) and r["bodyOctets"]<=32768 and digest(r["bodySha256"]),"Response retention differs")
                h=r["headers"];require(set(h)==set(request_plan["safeRetainedHeaders"]) and all(v is None or isinstance(v,str) for v in h.values()),"Response header allowlist differs")
                require(h["x-bdp-development"]=="non-attesting","Missing private marker")
                if r["bodyOctets"]==0:require(r["bodySha256"]==sha(b""),"Empty body hash differs")
                if r["status"]==400:
                    require(h["content-type"]=="application/problem+json" and h["cache-control"]=="private, no-store","Malformed response representation differs")
                    require(all(h[k] is None for k in ("bdp-scope-epoch","bdp-authorization-view","bdp-scope-position","etag","last-modified","location","link","x-bdp-development-limit")),"Malformed response leaked context")
            require(rows[4]["bodyOctets"]==0 and rows[4]["headers"]["content-length"]==rows[3]["headers"]["content-length"],"HEAD representation differs")
        else:
            require(o["aborted"] is False,"Snapshot deadline aborted")
            children=o["children"]
            methods=(["POST"]+["GET"]*9 if name=="snapshot-capacity" and ri==0 else
                     ["POST"]+["GET"]*7+["HEAD","GET","GET","GET","GET","POST","GET"]+["GET"]*14 if name=="snapshot-capacity" else
                     ["GET","GET","POST","GET","HEAD","GET","GET","GET","GET","HEAD","GET","GET","GET","GET","HEAD","GET","HEAD","GET","GET","GET","GET","GET","GET","HEAD","POST","GET","GET","GET","HEAD","GET","GET"])
            require([r["method"] for r in rows]==methods,"Snapshot method vector differs")
            for index,r in enumerate(rows):
                exact(r,"url status method responseBytes wireSha256","Snapshot response")
                if name=="snapshot-capacity":
                    retained=(ri==1 and (19<=index<=26 or index==28))
                    expected_url=scope+("operations/batch" if index==0 else "snapshot?snapshot=a&snapshot=b" if ri==1 and index==15 else "snapshot")
                else:
                    retained=index in (21,22,23,26,27,28)
                    expected_url=scope+("operations/batch" if index in (2,24) else "snapshot")
                if not retained:require(r["url"]==expected_url,"Snapshot fixed route differs")
                require(r["method"] in ("GET","HEAD","POST") and integer(r["responseBytes"]) and r["responseBytes"]<=1048576 and digest(r["wireSha256"]),"Invalid Snapshot wire record")
        require(len(children)==contract["children"][ri],"Child count differs")
        for c in children:
            exact(c,"pid exitCode signalCode","Child")
            require(type(c["pid"]) is int and c["pid"]>0 and c["exitCode"]==0 and c["signalCode"] is None,"Child failed/unreaped")
            pids.append(c["pid"])
    require(len(set(pids))==len(pids),"Repeated child PID within cohort")
    absent_pids=absent(set(pids)|set(result["sampledPids"]))
    for item in set(result["sampledTempEntries"])|set(result["tempEntriesAfter"]):
        p=Path(item)
        require(not (p.name.startswith(tuple(spec["ownedTempPrefixes"])) and p.exists()),"Runtime directory remains")
    unknown=[p for p in result["unclassifiedTempEntriesStillPresent"] if Path(p).exists()]
    require(all(Path(p).name.startswith("node-compile-cache") for p in unknown),"Unknown retained temp entry requires disposition")
    out={"status":"accepted-private-observation","cohort":name,"source":g["finalHead"],"payload":provenance["payloadIdentity"],"logSha256":sha(log.read_bytes()),"resultsSha256":sha((E/(name+".results.json")).read_bytes()),"exitSha256":sha((E/(name+".exit.json")).read_bytes()),"records":records,"responses":sum(o["exchanges"] for o in records),"children":len(pids),"independentlyAbsentPids":absent_pids,"retainedCaches":unknown,"claimEligible":False,"limitations":"Bodies/schema/equality asserted in source callback; hashes cannot reconstruct bodies. No child Git attestation or public qualification."}
    with (E/(name+".review.json")).open("x") as f:f.write(json.dumps(out,indent=2)+"\n")
    print(json.dumps({"cohort":name,"accepted":True,"responses":out["responses"],"children":len(pids)}))
if __name__=="__main__":main()
