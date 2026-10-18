"""Strict post-run adjudication. No product imports or additional HTTP."""
import json,os,re,sys
from pathlib import Path
from gate import P,E,require,sha,load_gate,completed
name=sys.argv[1] if len(sys.argv)==2 else ''
require(name in ['visibility-handles','active-event'],'Exact cohort required')
g,spec=load_gate('review-'+name)
end=completed(name);start=json.loads((E/(name+'.started.json')).read_text())
commands=json.loads((P/'commands.json').read_text());command=commands['commands'][name]
require(start['argv']==end['argv']==command['argv'] and start['cwd']==end['cwd']==command['cwd'],'Command attribution differs')
require(start['head']==g['finalHead'] and start['workspaceDirty'] is False,'Cohort source attribution differs')
raw=(E/(name+'.json')).read_bytes();require(len(raw)<=1048577,'Report ceiling')
report=json.loads(raw)
def keys(value,names):
 require(isinstance(value,dict) and set(value)==set(names),'Unexpected report fields')
def absent(pid):
 require(type(pid)is int and pid>1,'Missing child PID')
 try:os.kill(pid,0)
 except ProcessLookupError:return True
 except PermissionError:return False
 return False
keys(report,['label','name','attempts','maximumRequests','observations','result','failureCount','failureKinds','timingLimited','child','cleanup','binding','claimEligible'])
n=18 if name=='visibility-handles' else 4
require(report['label']=='private-installed-authorization' and report['name']==name,'Report identity')
require(report['attempts']==report['maximumRequests']==n and report['result']=='passed' and report['failureCount']==0 and report['failureKinds']==[] and report['timingLimited'] is False and report['claimEligible'] is False,'Incomplete/failed witness')
log=(E/(name+'.log')).read_text()
records=[]
for line in log.splitlines():
 try:value=json.loads(line)
 except (ValueError,TypeError):continue
 if isinstance(value,dict) and value.get('label')=='private-installed-authorization':records.append(value)
require(records==[report],'One exact stdout/file final record required')
keys(report['child'],['pid','exitCode','signalCode','pidAbsent'])
child=report['child'];require(child['exitCode']==0 and child['signalCode'] is None and child['pidAbsent'] is True and absent(child['pid']),'Product child did not close cleanly')
keys(report['cleanup'],['closed','directory','directoryRemoved'])
cleanup=report['cleanup'];directory=Path(cleanup['directory']);tmp=E/(name+'.tmp')
require(directory.parent.resolve()==tmp.resolve() and directory.name.startswith('bdp-authorization-installed-'),'Unexpected store ownership')
require(cleanup['closed'] is True and cleanup['directoryRemoved'] is True and not directory.exists(),'Owned store remains')
for pid in end['sampledPids']:require(absent(pid),'Sampled process remains')
proof=json.loads((E/'installed-provenance.json').read_text())
binding=report['binding']
keys(binding,['sourceCommit','sourceTree','installedEntry','installedEntrySha256','wrapper','wrapperSha256','payloadIdentity','artifactIdentity','authorityMode','scope','runtimeConfigurationIdentity','policyIdentity','authenticationIdentity','descriptorSha256','canonicalHost'])
require(binding['sourceCommit']==g['finalHead']==proof['sourceCommit'] and binding['sourceTree']==g['finalTree']==proof['sourceTree'],'External source binding differs')
require(binding['installedEntry']==proof['entryRealpath'] and binding['installedEntrySha256']==proof['entrySha256']==sha(Path(binding['installedEntry']).read_bytes()),'Installed entry differs')
require(binding['wrapper']==str(P/'authorization-bootstrap.mjs') and binding['wrapperSha256']==sha((P/'authorization-bootstrap.mjs').read_bytes()),'External wrapper differs')
require(binding['payloadIdentity']==proof['payloadIdentity'] and re.fullmatch('[0-9a-f]{64}',binding['payloadIdentity']),'Payload identity differs')
require(binding['scope']=='https://event-sse.invalid/acme/' and binding['authorityMode']=='canonical' and binding['canonicalHost']=='event-sse.invalid','Authority differs')
for field in ['runtimeConfigurationIdentity','policyIdentity','authenticationIdentity']:
 require(re.fullmatch('[0-9a-f]{64}',binding[field]) is not None,'Invalid safe configuration identity')
require(re.fullmatch('development-schema5-[0-9a-f]{64}',binding['artifactIdentity']) is not None,'Artifact identity shape')
ledger=json.loads((P/'ledger.json').read_text());require(binding['descriptorSha256']==ledger['artifactSha256'],'Descriptor differs')
observations=report['observations'];require(isinstance(observations,list),'Observations absent')
wire=[x for x in observations if isinstance(x,dict) and 'index'in x]
other=[x for x in observations if x not in wire]
require(len(wire)==n and [x['index'] for x in wire]==list(range(1,n+1)),'Actual response ordering/count differs')
expected=ledger['cohorts'][name]['rows']
uuid=r'[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}'
old=wire[0]['context'];require(len(old)==3 and all(isinstance(x,str)for x in old),'Initial context')
require(re.fullmatch(uuid,old[0]) and old[1]=='development-all-visible' and old[2]=='pos-1','Initial context values')
def checktarget(actual,expected):
 require(isinstance(actual,str) and actual.startswith('/acme/') and len(actual)<=4096 and not any(ord(x)<33 for x in actual),'Unsafe observed target')
 if expected.startswith('<'):return
 if expected=='events/?after=<actual final Event id>':
  require(actual.startswith('/acme/events/?after=') and len(actual)>len('/acme/events/?after='),'Event continuation absent')
 else:require(actual=='/acme/'+expected,'Literal target differs')
for observed,row in zip(wire,expected,strict=True):
 index,method,target,status,_=row
 require(observed['method']==method and observed['status']==status and observed['canonicalHost']=='event-sse.invalid','Wire method/status/Host differs')
 checktarget(observed['target'],target)
 context=observed['context'];require(isinstance(context,list)and len(context)==3 and context[0]==old[0] and (context[1]=='development-all-visible' or re.fullmatch(uuid,context[1])),'Context scope/view invalid')
 require(context[2]==('pos-1' if name=='active-event' or index==1 else 'pos-2'),'Position differs')
 if observed.get('stream'):
  keys(observed,['index','method','target','stream','canonicalHost','status','context','closed','clientAbortedBeforeClose','wireBytes','completeFrames','idFrames'])
  require(observed['stream']is True and observed['closed']is True and observed['clientAbortedBeforeClose']is False,'Stream not closed by server')
  require(type(observed['wireBytes'])is int and 0<observed['wireBytes']<=2097152 and type(observed['completeFrames'])is int and observed['completeFrames']>=0 and type(observed['idFrames'])is int and 0<=observed['idFrames']<=observed['completeFrames'],'Stream bounds')
 else:
  fields=['index','method','target','status','bodyBytes','writeComplete','canonicalHost','bodySha256','requestBodyBytes','context','contentType','contentLength']
  if method=='POST':fields.append('requestBodySha256')
  keys(observed,fields)
  require(observed['writeComplete']is True and type(observed['bodyBytes'])is int and 0<=observed['bodyBytes']<=32768 and re.fullmatch('[0-9a-f]{64}',observed['bodySha256']),'Finite wire completeness')
  require(observed['contentType'] in ['application/json','application/problem+json'],'Content type')
  require(isinstance(observed['contentLength'],str)and re.fullmatch('[0-9]+',observed['contentLength']),'Representation length absent')
  if method=='HEAD':
   require(observed['bodyBytes']==0 and observed['bodySha256']==sha(b'') and int(observed['contentLength'])==wire[index-2]['bodyBytes'] and observed['contentType']==wire[index-2]['contentType'],'HEAD representation differs')
  else:require(observed['bodyBytes']==int(observed['contentLength'])>0,'Body length differs')
  if method!='POST':require(observed['requestBodyBytes']==0,'Unexpected GET/HEAD body')
post_hashes={'visibility-handles':[(3728,'a778feb6f8559743228db36139ccc653370c43a30cdcf3d2017db2e85349b054'),(132,'8f22a1dee8d5c16750f476c2c0f4883ca91c257014e07b9efef9d2e678a98220'),(3728,'a778feb6f8559743228db36139ccc653370c43a30cdcf3d2017db2e85349b054')],'active-event':[(244,'77b717d875eb96e4f16756f40001700f9cf4848be09a7acfb2c7b44ccc36f072')]}
require([(x['requestBodyBytes'],x['requestBodySha256'])for x in wire if x['method']=='POST']==post_hashes[name],'Actual request fingerprint differs')
if name=='visibility-handles':
 require(len(other)==3,'Publication/timing observation count')
 publication,dispatch,response=other
 keys(publication,['publication','oldStreamClosedAtPublication'])
 policy=publication['publication'];keys(policy,['revision','bindingKind','generation','authorizationView','readPolicy','canMutate'])
 require({k:v for k,v in policy.items()if k!='authorizationView'}=={'revision':1,'bindingKind':'scope','generation':1,'readPolicy':'empty','canMutate':False},'Policy publication differs')
 require(type(publication['oldStreamClosedAtPublication'])is bool,'Publication closure marker type')
 for value,field in [(dispatch,'cursorDispatchAgeMs'),(response,'cursorResponseAgeMs')]:
  keys(value,[field,'requiredBelowMs']);require(value['requiredBelowMs']==500 and type(value[field])in [int,float] and 0<=value[field]<500,'Cursor timing prerequisite not met')
 require(response['cursorResponseAgeMs']>=dispatch['cursorDispatchAgeMs'],'Timing chronology')
 new=policy['authorizationView'];require(new!=old[1]and re.fullmatch(uuid,new),'No fresh view')
 require(all(x['context'][1]==old[1]for x in wire[:7]) and all(x['context'][1]==new for x in wire[7:]),'Before/after view mismatch')
 require([x['index']for x in wire if x.get('stream')]==[7,14],'Stream count differs')
 require(wire[6]['idFrames']==0 and wire[13]['idFrames']==0,'Revoked Changes emitted frames')
 require(wire[15]['target']==wire[16]['target'],'HEAD is not actual prior receipt')
else:
 require(len(other)==1,'Event publication observation count');value=other[0];keys(value,['deliveredFrames','publishedView','oldStreamClosed'])
 require(value['deliveredFrames']==2 and value['oldStreamClosed']is True and value['publishedView']!=old[1]and re.fullmatch(uuid,value['publishedView']),'Event publication/closure differs')
 require([x['index']for x in wire if x.get('stream')]==[3] and wire[2]['idFrames']==2,'Exact Event stream count')
 require(all(x['context'][1]==old[1]for x in wire[:3])and wire[3]['context'][1]==value['publishedView'],'Event view transition differs')
with (E/(name+'.review.json')).open('x')as f:
 json.dump({'accepted':True,'cohort':name,'actualResponses':n,'productChildren':1,'reportSha256':sha(raw),'logSha256':sha((E/(name+'.log')).read_bytes()),'rawExit':end['exit'],'supervisorExit':end['supervisorExit'],'pid':child['pid'],'pidAbsent':True,'storeAbsent':True,'sourceCommit':g['finalHead'],'sourceTree':g['finalTree'],'payloadIdentity':binding['payloadIdentity'],'installedEntrySha256':binding['installedEntrySha256'],'wrapperSha256':binding['wrapperSha256'],'claimEligible':False,'limitations':'Full bodies/schema and dynamic-handle relationships are asserted by the separately hash-bound observer; safe hashes alone do not reproduce those semantic assertions. Source is externally attributed, not child-attested. No migration/reopen or general ACL qualification.'},f,indent=2);f.write('\n')
print(json.dumps({'accepted':True,'cohort':name,'actualResponses':n,'productChildren':1}))
