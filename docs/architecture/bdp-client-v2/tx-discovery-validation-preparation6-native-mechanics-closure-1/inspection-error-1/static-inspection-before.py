from pathlib import Path
import json,hashlib,ast,difflib,copy,re
L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs');D=Path('/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2');P=L/'tx-discovery-validation-preparation-6';Q=L/'tx-discovery-validation-preparation-5';M=D/P.name
sha=lambda b:hashlib.sha256(b).hexdigest()
load=lambda p:json.loads(p.read_text())
checks={};errors=[]
def ck(k,v):checks[k]=v; assert v,k
idx=load(P/'preparation-index.json'); ck('preparationIndexExact',sha((P/'preparation-index.json').read_bytes())=='b897b61a370de781f6d83409f71a39f924c7f0c0960b1f46f5cb2a344ad4e6bc')
physical={str(f.relative_to(P)) for f in P.rglob('*') if f.is_file()};ck('1312Physical',len(physical)==1312);ck('indexComplete',physical==set(idx['files'])|{'preparation-index.json'})
for rel,r in idx['files'].items():
 b=(P/rel).read_bytes();ck('index:'+rel,sha(b)==r['sha256'] and len(b)==r['bytes'])
ck('mirrorComplete',{str(f.relative_to(M)) for f in M.rglob('*') if f.is_file()}==physical)
for rel in physical:ck('mirror:'+rel,(P/rel).read_bytes()==(M/rel).read_bytes())
impl=load(P/'implementation-hashes.json');ck('1301Implementation',len(impl['files'])==1301);ck('11Exclusions',len(impl['excludedControlRelativePaths'])==11 and physical-set(impl['files'])==set(impl['excludedControlRelativePaths']))
for rel,h in impl['files'].items():ck('impl:'+rel,type(h)is str and re.fullmatch('[a-f0-9]{64}',h) is not None and sha((P/rel).read_bytes())==h)
orig=load(P/'context-origins.json');ck('1218Origins',len(orig)==1218)
for rel,r in orig.items():
 b=(P/rel).read_bytes();ck('origin:'+rel,len(b)==r['bytes'] and sha(b)==r['sha256'] and Path(r['origin']).read_bytes()==b)
pins=load(P/'reference-pins.json');ck('4105References',len(pins)==4105)
for f,h in pins.items():ck('reference:'+f,sha(Path(f).read_bytes())==h)
helpers=['run.py','gate.py','audit.py','packages.py','guarded-action-session-20261018-021624-549f.py'];H={}
for n in helpers:
 b=(P/n).read_bytes();H[n]=sha(b);ck('helper5:'+n,b==(Q/n).read_bytes());ck('helper2:'+n,b==(L/'tx-discovery-validation-preparation-2'/n).read_bytes());ck('AST:'+n,ast.dump(ast.parse(b))==ast.dump(ast.parse((Q/n).read_bytes())))
for n in ['source-binding.json','selection.json'] :ck('unchanged:'+n,(P/n).read_bytes()==(Q/n).read_bytes())
for n in ['commands.json','source-audit-contract.json','installed-audit-contract.json','package-consumer-contract.json','audit-expectations.json','selection-capacity.json']:
 ck('pathOnly:'+n,(P/n).read_bytes().replace(str(P).encode(),str(Q).encode())==(Q/n).read_bytes())
s=load(P/'report-schemas.json');old=load(Q/'report-schemas.json');prec=load(P/'exact-eleven-precision.json');rev=copy.deepcopy(s)
ck('11Ordinals',[x['ordinal'] for x in prec['deltas']]==[7,10,13,16,19,22,26,29,32,35,38])
for r in prec['deltas']:
 row=rev['installedExpected']['installed-02'][r['ordinal']-1];ck('schemaRow:'+r['id'],row['id']==r['id'] and row['fieldOccurrences']==r['after'] and old['installedExpected']['installed-02'][r['ordinal']-1]['fieldOccurrences']==[]);row['fieldOccurrences']=[]
ck('schemaOnly11',rev==old);ck('schemaReverseByteExact',(json.dumps(rev,indent=2,ensure_ascii=False)+'\n').encode()==(Q/'report-schemas.json').read_bytes())
ck('schemaAfter',sha((P/'report-schemas.json').read_bytes())==prec['afterSchemaSha256']);ck('schemaBefore',sha((Q/'report-schemas.json').read_bytes())==prec['beforeSchemaSha256'])
delta=load(P/'successor-delta.json')['files'];ck('28Pairs',len(delta)==28)
for n,r in delta.items():
 a=(P/'predecessor5'/n).read_bytes();b=(P/n).read_bytes();f=(P/(n+'.successor6.patch')).read_bytes();iv=(P/(n+'.successor6.inverse.patch')).read_bytes()
 ck('deltaHashes:'+n,sha(a)==r['beforeSha256'] and sha(b)==r['afterSha256'] and sha(f)==r['forwardSha256'] and sha(iv)==r['inverseSha256'] and (a!=b)==r['changed'])
 ck('forward:'+n,f==''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='predecessor5/'+n,tofile='current6/'+n)).encode())
 ck('inverse:'+n,iv==''.join(difflib.unified_diff(b.decode().splitlines(True),a.decode().splitlines(True),fromfile='current6/'+n,tofile='predecessor5/'+n)).encode())
sb=load(P/'source-binding.json');ck('679Sources',len(sb['trackedSourceFiles'])==679)
for rel,h in sb['trackedSourceFiles'].items():ck('source:'+rel,sha((Path(sb['worktree'])/rel).read_bytes())==h)
for rel,r in sb['ownedFiles'].items():ck('owned:'+rel,sha((Path(sb['worktree'])/rel).read_bytes())==r['afterSha256'])
for n in ['execution-gate.json','execution-gate.closed.json']:
 g=load(P/n);ck('closed:'+n,g['released'] is False and g['approvedStages']==[] and g['bootstrapReuse']==[] and g['staticValidationAuthorized'] is False);ck('implGate:'+n,g['implementationIndexSha256']==sha((P/'implementation-hashes.json').read_bytes()))
ck('gateIdentical',(P/'execution-gate.json').read_bytes()==(P/'execution-gate.closed.json').read_bytes());ck('noActiveRun',not(P/'run').exists())
c=load(P/'commands.json');order=c['finalOrder'];ck('52Stages',len(order)==52 and len(set(order))==52);ck('16ImmediatePairs',sum(x.startswith('audit-') for x in order)==16 and all(i>0 and order[i-1]==x[6:] for i,x in enumerate(order) if x.startswith('audit-')))
hist=load(P/'historical-subset-inventory.json');ck('historicalCounts',[hist[x] for x in ['originalPreparedPaths','actualFlatRunFiles','historicalArchives','consumerMetadataFilesCopied','actualStageExits']]==[957,128,2,3,28]);ck('noReuse',hist['futureOutputReuse'] is False)
for rel,r in hist['files'].items():ck('history:'+rel,(P/rel).read_bytes()==Path(r['origin']).read_bytes() and sha((P/rel).read_bytes())==r['sha256'])
oldidx=load(Q/'preparation-index.json');ck('old957',len(oldidx['files'])+1==957)
for rel,r in oldidx['files'].items():
 expected=hist['preparedActiveGateOnlyException']['observedClosedSha256'] if rel=='execution-gate.json' else r['sha256'];ck('oldFrozen:'+rel,sha((P/'predecessor5'/rel).read_bytes())==expected)
ck('oldIndexExact',(P/'predecessor5/preparation-index.json').read_bytes()==(Q/'preparation-index.json').read_bytes())
ck('noConsumerBody',not(P/'run/consumer').exists() and not(P/'predecessor5/run/consumer').exists())
result={'state':'STATIC VERIFICATION COMPLETE; no helper imported or executed','accepted':True,'checkCount':len(checks),'counts':{'physical':len(physical),'indexed':len(idx['files']),'implementation':len(impl['files']),'origins':len(orig),'referencePins':len(pins),'source':679,'deltaPairs':len(delta),'stages':len(order),'immediateAuditPairs':16},'helpers':H,'schemaReversalByteExact':True,'onlyElevenArraysChanged':True,'currentSourceIdentity':sb['sourceIdentity'],'preparationIndexSha256':sha((P/'preparation-index.json').read_bytes()),'implementationIndexSha256':sha((P/'implementation-hashes.json').read_bytes()),'sourceBindingSha256':sha((P/'source-binding.json').read_bytes()),'reportSchemaSha256':sha((P/'report-schemas.json').read_bytes()),'allChecksPassed':all(checks.values())}
Path('/tmp/data6-native-verification.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
