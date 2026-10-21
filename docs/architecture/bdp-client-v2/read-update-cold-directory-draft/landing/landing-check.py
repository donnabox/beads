"""Closed root-invoked read-only PR138 landing inspection. Never ready/merge/push.
One shot, retained raw evidence; no poll loop, test retry, Git fetch or product run.
"""
from pathlib import Path
import subprocess,json,hashlib,re,datetime
P=Path(__file__).resolve().parent;D=P.parent;L=D.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-update-cold-directory")
INPUTS_SHA='85909bfee0e22fed03eab6a4e665ea640a7200ab091f9fffc888aaaea43a6f1e'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def require(ok,message):
    if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(["git",*args],cwd=W,timeout=15)
def read(p):return json.loads(p.read_text())
def main():
    require(sha(P/"inspection-inputs.json")==INPUTS_SHA,"Inspection pins changed")
    cfg=read(P/"inspection-inputs.json");release=read(P/"root-inspection-release.json")
    require(release.get("released") is True and release["scriptSha256"]==sha(Path(__file__)) and release["inputsSha256"]==INPUTS_SHA,"Root inspection release absent/changed")
    council_pin=release["rootCouncilAcceptance"];council_path=Path(council_pin["path"])
    require(council_path.is_relative_to(D) and sha(council_path)==council_pin["sha256"],"Final root council document missing/changed")
    council=read(council_path)
    require(council.get("accepted") is True and council["head"]==cfg["head"] and council["tree"]==cfg["tree"] and council["sourceBindingSha256"]==cfg["sourceBinding"]["sha256"],"Exact-candidate final council not accepted")
    require(sha(Path(cfg["sourceBinding"]["path"]))==cfg["sourceBinding"]["sha256"],"Final source document changed")
    subprocess.run(["python3",str(L/"guarded-action.py"),"/usr/bin/true"],check=True,timeout=40)
    R=P/"inspection";R.mkdir(exist_ok=False)
    def write(name,v):
        with (R/name).open("x") as f:json.dump(v,f,indent=2);f.write("\n")
    def capture(name,endpoint,raw=False):
        argv=["gh","api","--method","GET","repos/gastownhall/bdp/"+endpoint]
        with (R/name).open("xb") as out,(R/(name+".stderr")).open("xb") as err:
            result=subprocess.run(argv,stdout=out,stderr=err,timeout=90 if raw else 40)
        write(name+".capture.json",{"argv":argv,"exit":result.returncode,"sha256":sha(R/name),"bytes":(R/name).stat().st_size,"at":datetime.datetime.now(datetime.timezone.utc).isoformat()})
        require(result.returncode==0,"Read-only capture failed: "+name)
        return None if raw else read(R/name)
    def local():
        require(git("rev-parse","HEAD").decode().strip()==cfg["head"] and git("rev-parse","HEAD^{tree}").decode().strip()==cfg["tree"],"Local candidate drift")
        require(not git("status","--porcelain").strip(),"Candidate must remain clean")
        require(git("show","-s","--format=%P","HEAD").decode().strip().split()==cfg["parents"],"Ordinary integration parents changed")
        require(cfg["parents"]==[cfg["base"]] and cfg["authorCandidate"]==cfg["head"] and git("show","-s","--format=%P",cfg["authorCandidate"]).decode().strip().split()==[cfg["authoringParent"]],"Author/integration ancestry differs")
        git("merge-base","--is-ancestor",cfg["base"],cfg["head"])
        require(set(git("diff","--name-only",cfg["base"],cfg["head"]).decode().splitlines())==set(cfg["files"]) and len(cfg["files"])==6,"Feature scope differs")
        require(hashlib.sha256(git("diff","--binary",cfg["base"],cfg["head"])).hexdigest()==cfg["featureDiffSha256"],"Feature diff changed")
        for n,digest in cfg["files"].items():require(sha(W/n)==digest,"Source changed: "+n)
        require(sha(W/".github/workflows/ci.yml")==cfg["workflowSha256"],"Workflow differs")
    def current(pr,main):
        require(pr["head"]["sha"]==cfg["head"] and pr["base"]["ref"]=="main" and pr["base"]["sha"]==main["object"]["sha"]==cfg["base"] and pr["state"]=="open","Current PR head/base/remote main changed")
    try:
        local()
        pr=capture("pr.json",f"pulls/{cfg['pr']}");remote=capture("main.json","git/ref/heads/main");current(pr,remote)
        run=capture("run.json",f"actions/runs/{cfg['run']}");job=capture("job.json",f"actions/jobs/{cfg['job']}")
        require(run["id"]==cfg["run"] and job["id"]==cfg["job"] and job["run_id"]==run["id"] and run["head_sha"]==job["head_sha"]==cfg["head"] and run["event"]=="pull_request" and run["run_attempt"]==cfg["runAttempt"],"Wrong CI identity/attempt")
        require(run["status"]==job["status"]=="completed","CI incomplete; no partial logs downloaded")
        capture("job.log",f"actions/jobs/{cfg['job']}/logs",True)
        lines=(R/"job.log").read_text().splitlines();checkouts=[]
        for i,line in enumerate(lines[:-1]):
            if "[command]/usr/bin/git log -1 --format=%H" in line:
                m=re.search(r"(?:^|\s)([0-9a-f]{40})\s*$",lines[i+1]);require(m is not None,"Missing actual checkout SHA");checkouts.append(m.group(1))
        require(len(checkouts)==1,"Ambiguous actual checkout")
        synthetic=capture("synthetic.json","git/commits/"+checkouts[0])
        require(synthetic["sha"]==checkouts[0] and synthetic["tree"]["sha"]==cfg["tree"] and [p["sha"] for p in synthetic["parents"]]==[cfg["base"],cfg["head"]],"Actual tested synthetic tree/parents differ")
        tree=capture("workflow.json",f"contents/.github/workflows/ci.yml?ref={checkouts[0]}")
        import base64
        require(hashlib.sha256(base64.b64decode(tree["content"])).hexdigest()==cfg["workflowSha256"],"Synthetic workflow differs")
        require(run["conclusion"]==job["conclusion"]=="success" and job["steps"] and all(x["status"]=="completed" and x["conclusion"]=="success" for x in job["steps"]),"CI/job/steps not all successful")
        require(set(cfg["requiredSteps"]).issubset({x["name"] for x in job["steps"]}),"Required workflow steps absent")
        checks=capture("checks.json",f"commits/{cfg['head']}/check-runs?per_page=100")
        require(checks["total_count"]==len(checks["check_runs"]) and checks["check_runs"],"Missing/paginated checks")
        require(any(x["id"]==cfg["job"] for x in checks["check_runs"]) and all(x["status"]=="completed" and x["conclusion"]=="success" for x in checks["check_runs"]),"Current-head checks not all green")
        pr2=capture("pr-final.json",f"pulls/{cfg['pr']}");main2=capture("main-final.json","git/ref/heads/main");current(pr2,main2);local()
        require(sha(council_path)==council_pin["sha256"],"Council changed during inspection")
        summary=[line for line in lines if any(n in line for n in ["Test Files","Tests  ","Duration  ",cfg["expectedNewTestTitle"],"e2e:ready:preflight","Read-cohort evidence gate"])]
        write("outcome.json",{"inspection":"EXACT_CANDIDATE_CI_AND_COUNCIL_CHECKS_PASSED","head":cfg["head"],"tree":cfg["tree"],"base":cfg["base"],"candidateParents":cfg["parents"],"actualSynthetic":checkouts[0],"syntheticParents":[cfg["base"],cfg["head"]],"actualSteps":len(job["steps"]),"rawLogSha256":sha(R/"job.log"),"rootCouncilAcceptance":council_pin,"summaryLinesForHumanReview":summary,"suiteTotals":None,"humanFullLogReviewRequired":True,"ready":False,"mergeAuthorized":False,"limitation":"Read-only point-in-time inspection. Root must freshly recheck head/base/checks immediately before any merge and independently verify actual merge parents/tree. Remote child PID absence cannot be inferred."})
    except BaseException as error:
        write("failure.json",{"exceptionClass":type(error).__name__,"reason":str(error),"ready":False,"retryAuthorized":False});raise
if __name__=="__main__":main()
