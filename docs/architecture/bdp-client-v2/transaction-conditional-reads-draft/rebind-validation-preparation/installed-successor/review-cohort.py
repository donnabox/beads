"""Data-only exact conditional and cumulative cohort adjudication; no product imports."""
import sys,runpy
from pathlib import Path
name=sys.argv[1]
if name in ('principal-positive','principal-collapse'):
    runpy.run_path(str(Path(__file__).with_name('review-existing.py')))
    raise SystemExit(0)
if name in ('full-positive','cross-batch','cross-sequence'):
    globals().update(runpy.run_path(str(Path(__file__).with_name('review-existing.py'))))
else:
    """Read genuine cohort reports and independently reconcile bindings/cleanup. No imports of product."""
    from pathlib import Path
    import json,os,re,sys
    from gate import P,W,E,load_gate,completed,document,require,sha
    name=sys.argv[1]
    g,_=load_gate(name+"-review");outcome=completed(name)
    cfg=json.loads((P/"commands.json").read_text());spec=cfg["commands"][name]
    results=json.loads((E/(name+".results.json")).read_text())
    executed=[(str(Path(f["name"]).resolve().relative_to(W)),a["fullName"],a["status"]) for f in results["testResults"] for a in f.get("assertionResults",[]) if a["status"] not in ("pending","skipped","todo")]
    require(executed==[("apps/bdptest/src/transactional-matrix.test.ts",spec["selectedFullNames"][0],"passed")],"Wrong actual selected case")
    require(results["success"] and results["numPassedTests"]==1 and results["numFailedTests"]==0 and not results.get("unhandledErrors") and not results.get("errors") and not results.get("numRuntimeErrorTestSuites",0),"Test result failed")
    report=Path(spec["observationPath"]);o=json.loads(report.read_text());r=o["result"];p=json.loads((E/"installed-provenance.json").read_text())
    require(o["sourceCommit"]==g["finalHead"] and Path(o["entry"]).resolve()==Path(p["entryRealpath"]),"Wrong source/entry")
    s=o["baseline"]["served"]
    require(s["sourceCommit"]==g["finalHead"] and s["installedPayloadDigest"]==p["payloadIdentity"],"Wrong payload/source binding")
    require(s["scope"]=="https://reference.invalid/acme/" and s["profile"]=="transactional" and s["history"] is False and o["baseline"]["authorityMode"]=="canonical","Wrong serving configuration")
    for key in ("installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest"):require(re.fullmatch("[0-9a-f]{64}",s[key]) is not None,"Missing serving digest")
    require(s["fixtureDigest"]==sha((W/"packages/conformance/test-support/transactional/fixture-v1.json").read_bytes())==r["artifacts"]["fixtureDigest"],"Fixture differs")
    for ready in o["sessions"]:
        require(ready["payloadIdentity"]==p["payloadIdentity"] and ready["runtimeConfigurationIdentity"]==o["baseline"]["runtimeConfigurationIdentity"],"Session payload/runtime differs")
        require(ready["authorityMode"]=="canonical" and ready["profile"]=="transactional" and ready["history"] is False and ready["claimEligible"] is False,"Session declaration differs")
    require(r["claimEligible"] is False and r["profile"]=="transactional","False qualification")
    require(r["declarations"]["exactConfigurationDigest"]=="5a32a664575b83fe55ea3b73b617fc738aefffd56d83affdf66a61d7fd8100e5","Route configuration differs")

ledger=json.loads((P/'request-ledger.json').read_text());literals=ledger['requests']
control=name=='conditional-control';count=4 if control else 14
if name in ('full-positive','cross-batch','cross-sequence'):
    count=14;control=False
row=next(x for x in r['scenarios'] if x['id']=='transactional.http.conditional-reads')
require(row['state']==('fail' if control else 'pass') and not row.get('prerequisiteFailure') and not row.get('cleanupError') and not row.get('actions'),'Conditional row state/lifecycle differs')
require(len(row['exchanges'])==count,'Conditional exact prefix differs')
supp=o['conditionalSupplemental']
require(set(supp)=={'mode','session','dispatches','responses','expectedExchanges','provenance'},'Conditional projection shape differs')
require(supp['mode']==('nonmatching-if-none-match' if control else 'positive') and supp['expectedExchanges']==count,'Conditional mode/count differs')
require(supp['provenance']=='authored credential handles and actual completed raw request fingerprints; safe response metadata and body digests; full report joined by action order; no wire credential attestation','Projection attribution differs')
require(len(supp['dispatches'])==len(supp['responses'])==count,'Missing actual dispatch/response')
contract=json.loads((P/'conditional-report-contract.json').read_text())
session=supp['session'];require(set(session)=={'scenario','reopen','sessionOrdinal'},'Conditional session fields differ')
ordinal=session['sessionOrdinal'];require(type(ordinal) is int and ordinal>=0 and session['scenario']=='transactional.http.conditional-reads' and session['reopen'] is False,'Conditional session identity differs')
require(ordinal<len(o['sessions']) and all(o['sessions'][ordinal].get(k)==v for k,v in session.items()),'Conditional actual session join differs')
require(sum(x.get('scenario')=='transactional.http.conditional-reads' for x in o['sessions'])==1,'Conditional session duplicated')
require(all(x.get('sessionOrdinal')==i for i,x in enumerate(o['sessions'])),'Actual session ordinal inventory differs')
if name in ('full-positive','cross-batch','cross-sequence'):
    parent=contract['sessionChecks']['full']['parentFirstPresentations']
    require([x['scenario'] for x in o['sessions'] if x.get('reopen') is False]==parent+['transactional.http.conditional-reads'],'Parent first presentations/tail order differ')
    require(ordinal==len(o['sessions'])-1 and all(x['scenario'] in parent for x in o['sessions'][:ordinal]),'Conditional not actual final session')
else:require(ordinal==0 and len(o['sessions'])==1,'Selected conditional session differs')
allowed=set(json.loads((P/'conditional-report-contract.json').read_text())['responseHeaderAllowlist'])
failures=[];epoch=view=None
for i,(literal,dispatch,response,ex) in enumerate(zip(literals[:count],supp['dispatches'],supp['responses'],row['exchanges'],strict=True)):
    require(dispatch=={k:literal[k] for k in ('method','url','credentialRef')},'Actual conditional dispatch differs')
    require(ex['request']['id']==literal['id'] and ex['request']['method']==literal['method'] and ex['request']['url']==literal['url'],'Report/order join differs')
    require(not ex.get('transportError') and not ex.get('harnessError'),'Conditional transport/harness failure')
    exact=ex['request']['exact'];status=200 if control and i==3 else literal['status']
    require(exact['writeState']==response['writeState']=='complete' and exact['bodyPresent']==(literal['bodyUtf8'] is not None) and exact['bodyOctets']==literal['bodyBytes'],'Completed request framing differs')
    require(response['action']==literal['id'] and response['status']==ex['response']['status']==status,'Actual response status differs')
    require(response['requestBodyBytes']==literal['bodyBytes'] and response.get('requestBodyDigest')==literal['bodySha256'],'Actual request fingerprint differs')
    expected_fields={'action','status','bodyBytes','bodyDigest','headers','requestBodyBytes','writeState'}|({'requestBodyDigest'} if literal['bodyUtf8'] is not None else set())
    require(set(response)==expected_fields,'Unsafe/absent response field')
    require(isinstance(response['bodyBytes'],int) and 0<=response['bodyBytes']<=8192 and re.fullmatch('[0-9a-f]{64}',response['bodyDigest']),'Response fingerprint/bound differs')
    headers=response['headers'];require(set(headers)<=allowed and headers.get('cache-control')=='private, no-store','Response headers differ')
    context=['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position']
    # Reference-only authentication context absence; not portable assertion credit.
    if i in (7,8):
        require(all(k not in headers for k in context) and headers.get('www-authenticate')=='Bearer realm="bdp-transaction-development"','Anonymous context/challenge differs')
    else:
        require(all(isinstance(headers.get(k),str) and headers[k] for k in context),'Missing authorized context')
        if i==0:epoch=headers[context[0]];view=headers[context[1]]
        require(headers[context[0]]==epoch and headers[context[1]]==view and headers[context[2]]==('pos-2' if i>=11 else 'pos-1'),'Context position/identity differs')
    if literal['method']=='HEAD' or status in (304,412):
        require(response['bodyBytes']==0 and response['bodyDigest']==sha(b''),'Body-forbidden response not empty')
    require([a['id'] for a in ex['assertions']]==literal['assertionIds'],'Complete conditional assertion inventory differs')
    failures.extend(literal['id']+':'+a['id'] for a in ex['assertions'] if not a['passed'])
require(failures==(ledger['control']['expectedFailureIds'] if control else []),'Wrong genuine conditional failure inventory')
initial=supp['responses'][1]['headers'].get('etag');require(isinstance(initial,str) and initial and initial!='"conditional-no-match"','Initial validator/control precondition differs')
for index in ([3] if control else [3,4]):require(supp['responses'][index]['headers'].get('etag')==initial,'Conditional earlier-validator equality differs')
if control:require(supp['responses'][3]['bodyBytes']>0 and supp['responses'][3]['bodyDigest']==supp['responses'][1]['bodyDigest'],'Control genuine full current body differs')
if name.startswith('conditional-'):
    require(r['selectedScenarioIds']==['transactional.http.conditional-reads'] and len(r['scenarios'])==1 and len(o['sessions'])==1,'Wrong selected conditional execution')
    require(sum(x['requestBodyBytes'] for x in supp['responses'])==spec['workload']['requestBodyBytes'],'Conditional body workload differs')
    manifest=json.loads((W/'packages/conformance/test-support/transactional/manifest-v1.json').read_text())
    if control:
        row_source=next(x for x in manifest['scenarios'] if x['id']=='transactional.http.conditional-reads')
        control_action=next(a for a in row_source['actions'] if a['id']=='conditional-get')
        header=next(h for h in control_action['raw']['headerLines'] if h['name']=='if-none-match')
        require(header=={'name':'if-none-match','value':'*'},'Control source seam differs')
        header['value']='"conditional-no-match"'
        require(r['artifacts']['manifestDigest']==sha(json.dumps(manifest,separators=(',',':'),ensure_ascii=False).encode()),'Control derivative differs')
    else:require(r['artifacts']['manifestDigest']==sha((W/'packages/conformance/test-support/transactional/manifest-v1.json').read_bytes()),'Canonical manifest differs')
    logs=[]
    for line in (E/(name+'.log')).read_text().splitlines():
        line=re.sub(r'\x1b\[[0-9;]*m','',line).strip()
        if not line.startswith('{'):continue
        try:v=json.loads(line)
        except json.JSONDecodeError:continue
        if v.get('kind')=='installed-tx-conditional':logs.append(v)
    require(len(logs)==1,'Missing/duplicate conditional log')
    log=logs[0];require(log['sourceCommit']==g['finalHead'] and log['control']==control and log['exchanges']==count and log['childLaunches']==log['closedChildren']==2 and log['claimEligible'] is False and log['conditionalSupplemental']==supp,'Conditional log/report/closure differs')
for pid in outcome['sampledPids']:
    try:os.kill(pid,0)
    except ProcessLookupError:continue
    raise RuntimeError('Sampled process remains '+str(pid))
tmp=E/(name+'.tmp');require(not any(x.name.startswith('bdp-tx-installed-') for x in tmp.iterdir()) if tmp.exists() else True,'Owned store remains')
result={'command':name,'sourceCommit':g['finalHead'],'sourceTree':g['finalTree'],'conditionalHTTP':count,'failedIds':failures,'reportSha256':sha(report.read_bytes()),'rootReviewRequired':True,'claimEligible':False,'limits':'Full JSON schema/postimage equality executes in frozen app before redaction; safe fingerprints do not independently reconstruct bodies. Actual dispatch handles are not wire credential attestation.'}
with (E/(name+'.conditional-review.json')).open('x') as f:f.write(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
