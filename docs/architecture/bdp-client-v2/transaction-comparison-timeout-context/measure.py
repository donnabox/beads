from pathlib import Path
import subprocess,json,hashlib,re,difflib
r=Path('/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-context');l=Path(__file__).parent
h=lambda b:hashlib.sha256(b).hexdigest()
def git(*a):return subprocess.check_output(['git',*a],cwd=r,text=True).strip()
base='c555fe01df5cd20454fb0d663431df816c1270aa';assert git('rev-parse','HEAD')==base
names=['transaction-comparison.ts','transaction-store.ts','transaction-submission.ts','transaction-http.ts','transaction-store.test.ts','transaction-submission.test.ts','transaction-http.test.ts']
paths=['packages/server/src/'+f for f in names]+['docs/design/transaction-implementation-progress.md']
assert set(git('diff','--name-only').splitlines())==set(paths)
files=[dict(path=p,bytes=(r/p).stat().st_size,sha256=h((r/p).read_bytes()),baseSha256=h(subprocess.check_output(['git','show',base+':'+p],cwd=r))) for p in paths]
selection={
 'transaction-store.test.ts':[
 'captures comparison deadline context in one metadata-only atomic entry','rejects a defensive final comparison epoch mismatch without a tuple','returns a stopped comparison observation for a normally closed owner','rejects reentrant comparison observation without entering an atomic read',"preserves a fenced owner's comparison observation failure",'propagates comparison metadata failure and releases the atomic owner'],
 'transaction-submission.test.ts':[
 'captures current comparison context only at the observed deadline','distinguishes defensive comparison epoch stop before deadline observation','keeps full-owner unobserved comparison exact and excludes a caller callback','returns a provable unresolved conflict immediately and times out possible equality without disclosing a receipt','keeps a single comparison deadline across repeated owner replacement and stops at epoch change'],
 'transaction-http.test.ts':[
 'serves current comparison deadline context without timeout effects or receipt disclosure','maps a finite unresolved comparison to temporary refusal without disclosing or executing it'],
 'transaction-dependencies.test.ts':[
 'waits without a DB lock then joins the canonical direct/batch identity','proves static mismatch immediately, retains no timeout outcome and keeps the owner','competes freshly after retraction with the same comparison deadline'],
 'transaction-sequence-alias.test.ts':[
 "uses one bounded comparison deadline and leaves the owner's reservation intact","waits for a fixed A candidate then conflicts with the owner's actual B witness"]}
for p,titles in selection.items():
 s=(r/'packages/server/src'/p).read_text()
 for title in titles:assert s.count('it('+json.dumps(title)+',')==1 or s.count('it(\n    '+json.dumps(title)+',')==1,(p,title)
commands=[]
for p,titles in selection.items():
 commands.append(dict(file='packages/server/src/'+p,expectedCases=len(titles),parameterExpansion='Each selected literal title belongs to exactly one direct it call; no it.each or generation loop.',titles=titles,args=['exec','vitest','run','packages/server/src/'+p,'-t','|'.join(re.escape(t) for t in titles),'--maxWorkers=2','--reporter=verbose','--reporter=json']))
http=selection['transaction-http.test.ts'][0]
controls=[
 ('remove-deadline-wrapper','transaction-http.ts','''        return context(response, result.context);''','''        return response;''',http,'deadline context presence bdp-scope-epoch',4),
 ('stale-deadline-position','transaction-comparison.ts','''? { kind: "comparison-deadline", context: observed.context }''','''? { kind: "comparison-deadline", context: { ...observed.context, scopePosition: "pos-0" } }''',http,'deadline current position',4),
 ('lose-required-tuple','transaction-comparison.ts','''? { kind: "comparison-deadline", context: observed.context }''','''? { kind: "comparison-deadline", context: undefined as unknown as TransactionReadContext }''',http,'healthy deadline response',4),
 ('leak-unobserved-fields','transaction-submission.ts','''    case "temporarily-unavailable":
      return {
        kind: "refused",
        status: 503,
        code: "temporarily-unavailable",
        retry: "after-delay",
      };''','''    case "temporarily-unavailable":
      return {
        kind: "refused",
        status: 503,
        code: "temporarily-unavailable",
        retry: "after-delay",
        context: undefined as unknown as TransactionReadContext,
      };''',selection['transaction-submission.test.ts'][2],'legacy comparison result exact',0),
 ('omit-final-epoch-check','transaction-store.ts','''if (metadata("epoch") !== expectedEpoch)''','''if (metadata("epoch") !== expectedEpoch && false)''',selection['transaction-store.test.ts'][1],'final deadline epoch check',0),
]
cs=[]
for name,f,a,b,title,oracle,requests in controls:
 p='packages/server/src/'+f;s=(r/p).read_text();assert s.count(a)==1,(name,s.count(a));changed=s.replace(a,b);out=l/'controls'/name;out.mkdir(parents=True,exist_ok=True)
 patch=''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+p,tofile='b/'+p));(out/'source.patch').write_text(patch)
 cs.append(dict(name=name,path=p,patchSha256=h(patch.encode()),mutantSha256=h(changed.encode()),testFile='packages/server/src/'+('transaction-http.test.ts' if requests else ('transaction-submission.test.ts' if 'unobserved' in name else 'transaction-store.test.ts')),title=title,expectedFailure=oracle,plannedHttpUntilFailure=requests,plannedMaxHttp=12 if requests else 0,installedChildren=0,executed=False))
bead='https://example.test/bead';link='https://example.test/link'
create=lambda name:dict(operation='createBead',id='beads/'+name,type=bead,properties=dict(title=name))
dump=lambda v:json.dumps(v,separators=(',',':'),ensure_ascii=False)
sequence=dump(dict(operations=[dict(**create('a'),name='a',idempotencyKey='creator'),dict(operation='createLink',type=link,source='@a',target='@a',idempotencyKey='dependent')]))
candidate=dump(dict(operations=[dict(operation='createLink',type=link,source='beads/a',target='beads/a')]))
independent=dump(dict(operations=[create('b')]))
req=[('GET','events/',''),('POST','operations/sequence',sequence),('POST','operations/batch',candidate),('POST','operations/batch',independent),('GET','beads/?limit=32',''),('GET','links/?limit=32',''),('GET','events/',''),('GET','beads/a',''),('POST','operations/batch',candidate),('GET','beads/?limit=32',''),('GET','links/?limit=32',''),('GET','events/','')]
ledger=[dict(dispatchIndex=i+1,method=m,scopeRelativeTarget=t,body=b,bodyBytes=len(b.encode()),bodySha256=h(b.encode())) for i,(m,t,b) in enumerate(req)]
# Finite conservative fixture response estimate, deliberately not a wire transcript.
# <=5 facts: each has <=10 scalar fields of <=512 escaped bytes (includes fixed
# epoch/UUID/view/source/cursor strings) + <=2048 bytes exact small fact content.
estimate=dict(largestPlannedBodyEstimate=5*(10*512+2048)+4096,totalNewScenarioResponseEstimate=12*(5*(10*512+2048)+4096),basis='Conservative small-fixture envelope estimate: at most5 facts, fixed UUID/loopback/cursor scalars <512B, each tested empty/single-title data <2048B,4096B envelope. Not an enforced response cap or executed measurement; source fetch fixture retains existing behavior.',fixedSourceMaximumEventItems=5,configuredPageMaximumItems=32)
diff=subprocess.check_output(['git','diff','--no-ext-diff'],cwd=r);(l/'source.diff').write_bytes(diff)
bounds=dict(stage='draft-unexecuted',base=base,tree=git('rev-parse','HEAD^{tree}'),branch=git('branch','--show-current'),files=files,diffSha256=h(diff),diffStat=git('diff','--stat'),sourceSelection=commands,selectedCases=sum(len(t) for t in selection.values()),newCases=9,replacedCases=1,sourceHttp=dict(newScenario=12,preservedBasicTimeout=4,total=16,listeners=2,installedChildren=0,newScenarioWaits=1,newScenarioSchedulerDeliveries=2,preservedBasicSchedulerDeliveries=2),requestLedger=ledger,maxRequestBody=max(x['bodyBytes'] for x in ledger),totalRequestBody=sum(x['bodyBytes'] for x in ledger),responseEstimate=estimate,controls=cs,controlTotals=dict(plannedHttp=12,maxHttp=36,listeners=3,installedChildren=0,selectedRuns=5),installedProposal=dict(directRequests=14,sequenceRequests=13,totalRequests=27,children=2,newTimeoutScenario=False,requestBytes=65536,headerBytes=16384,responseBytes=131072,targetBytes=8192,exchangeMs=5000,scenarioMs=90000,outerMs=120000,closeMs=10000),executionOccurred=False)
(l/'preflight-bounds.json').write_text(json.dumps(bounds,indent=2)+'\n')
print('bounds',h((l/'preflight-bounds.json').read_bytes()),'diff',h(diff),'selected',bounds['selectedCases'],'requestBytes',bounds['maxRequestBody'],bounds['totalRequestBody'])
