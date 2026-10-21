"""D3a exact private callback/report audit; source-only preparation, root gate required."""
import sys
sys.dont_write_bytecode=True
from gate import *
import re,math
from collections import Counter
mode=sys.argv[1]
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build'}.get(mode,'audit-'+mode),child=True)
if mode=='tools':
 require((E/'node-version.log').read_text().strip()=='v24.16.0' and (E/'pnpm-version.log').read_text().strip()=='11.20.0','Actual runtime/package-manager versions differ')
 tools=dict(document(g['baseTools'])['tools'])
 paths=[W/'node_modules'/n for n in ['@biomejs/biome/bin/biome','vitest/vitest.mjs','typescript/bin/tsc','typescript/lib/tsc.js','typescript/lib/getExePath.js','typescript/package.json']]
 paths += [W/'node_modules/.pnpm/@biomejs+cli-darwin-arm64@2.5.7/node_modules/@biomejs/cli-darwin-arm64/biome',W/'node_modules/.pnpm/@typescript+typescript-darwin-arm64@7.0.2/node_modules/@typescript/typescript-darwin-arm64/lib/tsc']
 for p in paths:tools[str(p)]={'realpath':str(p.resolve(strict=True)),'sha256':sha(p.read_bytes())}
 versions={'@biomejs/biome':'2.5.7','vitest':'4.1.10','typescript':'7.0.2'}
 for name,version in versions.items():require(read(W/'node_modules'/name/'package.json')['version']==version,'Installed tool version differs')
 complete('workspace-install',g)
 write('workspace-tools.json',{'tools':tools,'versions':versions,'head':g['finalHead'],'sourceIdentity':g['_sourceIdentity'],'sourceLockSha256':g['sourceLockSha256'],'installExitSha256':sha((E/'workspace-install.exit.json').read_bytes()),'installLogSha256':sha((E/'workspace-install.log').read_bytes())})
elif mode=='build':
 write('build-artifacts.json',{'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'artifacts':inventory(),**{key:{'path':str(E/('fresh-build.'+suffix)),'sha256':sha((E/('fresh-build.'+suffix)).read_bytes())} for key,suffix in [('start','started.json'),('exit','exit.json'),('log','log')]}})
else:
 stage=mode;cfg=read(P/'commands.json')['commands'][stage];sel=read(P/'selection.json')[stage]
 result=read(E/(stage+'.results.json'));complete(stage,g)
 def keys(o,required,optional=()):
  require(type(o) is dict and set(required)<=set(o)<=set(required)|set(optional),'Missing/extra report key '+str(required))
 def exact(o,want):
  for k,v in want.items():require(k in o and type(o[k]) is type(v) and o[k]==v,'Typed actual field '+k)
 def integer(v):require(type(v) is int and v>=0,'Actual nonnegative integer')
 def finite(v,limit=None):require(type(v) in [int,float] and math.isfinite(v) and v>=0 and (limit is None or v<limit),'Actual finite bounded value')
 def digest(v):require(type(v) is str and re.fullmatch('[a-f0-9]{64}',v),'Actual SHA256')
 def strings(o):require(type(o) is dict and all(type(k) is str and type(v) is str for k,v in o.items()),'Header map')
 def absent_pid(pid):
  require(type(pid) is int and pid>0,'Actual child PID')
  try:os.kill(pid,0)
  except ProcessLookupError:return
  except OSError:raise RuntimeError('PID absence unknown')
  raise RuntimeError('Child remains')
 def absent_dir(value,prefix):
  tmp=Path(read(E/(stage+'.started.json'))['temporaryDirectory']);p=Path(value)
  require(p.is_absolute() and p.is_relative_to(tmp) and p.name.startswith(prefix) and not p.exists(),'Owned BDP directory absence')
 chosen=[]
 for suite in result['testResults']:
  for case in suite['assertionResults']:
   if case['fullName'] in cfg['selectedFullNames']:
    require(Path(suite['name']).resolve()==(W/cfg['selectedFilesByFullName'][case['fullName']]).resolve(),'Selected callback file')
    chosen.append(case)
   else:require(case['status'] in ['pending','skipped','todo'],'Unselected callback executed')
 require(Counter(x['fullName'] for x in chosen)==Counter(cfg['selectedFullNames']) and all(x['status']=='passed' for x in chosen),'Exact callbacks pass once')
 exact(result,{'success':True,'numFailedTests':0,'numFailedTestSuites':0,'numPassedTests':len(chosen)})
 if 'numRuntimeErrorTestSuites' in result:exact(result,{'numRuntimeErrorTestSuites':0})
 log=(E/(stage+'.log')).read_text();require(not re.search('Unhandled Errors|Unhandled Rejection|Uncaught Exception|Worker exited unexpectedly',log),'Unhandled runtime diagnostic')
 reports=[]
 for line in log.splitlines():
  try:o=json.loads(line.strip(),object_pairs_hook=unique_object,parse_constant=invalid_constant)
  except json.JSONDecodeError:
   require('\"observation\"' not in line,'Malformed observation JSON')
   continue
  if type(o) is dict and 'observation' in o:reports.append(o)
 schemas=read(P/'report-schemas.json');kind=sel['kind']
 def named_fields(fields):
  require(type(fields) is list,'Field occurrence array')
  for item in fields:
   keys(item,['name','value']);require(type(item['name']) is str and type(item['value']) is str,'Actual field values')
 def headers_record(r):
  strings(r['headers']);integer(r['status']);require(100<=r['status']<=599,'HTTP status')
 def same(a,b):
  if type(a) is not type(b):return False
  if type(b) is dict:return set(a)==set(b) and all(same(a[k],v) for k,v in b.items())
  if type(b) is list:return len(a)==len(b) and all(same(x,y) for x,y in zip(a,b))
  return a==b
 def token(v):require(type(v) is str and re.fullmatch('[A-Za-z0-9_-]{1,256}',v),'Actual nonempty context token')
 context_names=['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position']
 def header_context(headers,position,expected=None):
  if position is None:
   require(all(n not in headers for n in context_names),'No invented undisclosed context');return None
  values=[headers.get(n) for n in context_names]
  for v in values:token(v)
  require(values[2]==position,'Actual expected position')
  if expected is not None:require(values[:2]==expected,'Same actual epoch/view')
  return values[:2]
 def finite_url(value):
  require(type(value) is str and 0<len(value.encode())<=8192,'Actual bounded URL string')
 def scalars(o,want):
  require(type(o) is dict and all(k in o and same(o[k],v) for k,v in want.items()),'Typed semantic fields '+str(want))
 def rows(o,key,n):
  r=o.get(key);require(type(r) is list and len(r)==n,'Exact finite array '+key);return r
 def context(c,position):
  keys(c,['scopeEpoch','authorizationView','scopePosition'])
  for v in c.values():token(v)
  require(c['scopePosition']==position,'Actual owner position')
 def directory_absent(value,prefix):absent_dir(value,prefix)
 def cleanup(state,prefix):
  scalars(state,{'closed':True,'stopUnknown':False,'directoryAbsent':True});absent_dir(state['directory'],prefix)
 def source_response(r,want):
  keys(r,['method','url','credential','status','headers'],['body']);headers_record(r)
  for k in ['method','url','credential','status']:require(same(r[k],want[k]),'Exact source returned '+k)
  finite_url(r['url']);header_context(r['headers'],want['position'])
  mode=want['body']
  if mode=='absent':require('body' not in r,'Source response truly has no body')
  else:
   require(type(r.get('body')) is dict,'Actual source JSON object body')
   body=r['body']
   if mode=='problem':
    keys(body,['type','code','retry','status']);require(type(body['status']) is int and body['status']==r['status'] and all(type(body[k]) is str and body[k] for k in ['type','code','retry']),'Typed source Problem')
   elif mode=='descriptor':
    exact(body,{'scope':schemas['canonicalScope'],'profile':'transactional','bdpVersion':'0'})
    require([body.get(k) for k in ['scopeEpoch','authorizationView','headPosition']]==[r['headers'][k] for k in context_names],'Actual descriptor/current response tuple')
   elif mode=='directory':require(same(body,{k:v.rsplit('/',1)[1] for k,v in schemas['operationTargets'].items()}),'Exact canonical operation Directory')
   else:raise RuntimeError('Unknown finite source body grammar')
 def served_binding(binding,proof):
  keys(binding,['served','operationTargets','runtimeConfigurationIdentity','authorityMode']);exact(binding,{'authorityMode':'canonical'});digest(binding['runtimeConfigurationIdentity'])
  require(same(binding['operationTargets'],schemas['operationTargets']),'Exact canonical operation target inventory')
  served=binding['served'];keys(served,['scope','profile','history','sourceCommit','installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest','fixtureDigest','fixtureId','capabilities'])
  exact(served,{'scope':schemas['canonicalScope'],'profile':'transactional','history':False,'sourceCommit':g['finalHead'],'installedPayloadDigest':proof['payloadIdentity'],**schemas['fixtureExpected']})
  for k in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest','fixtureDigest']:digest(served[k])
 def expected_url(actual,want):
  finite_url(actual)
  if type(want) is str:require(actual==want,'Exact finite request URL')
  else:
   keys(want,['receiptUnder']);require(re.fullmatch(re.escape(want['receiptUnder'])+r'[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}',actual),'Actual received receipt identity confined to canonical root')
 def installed_rows(records,wants):
  require(type(records) is list and len(records)==len(wants),'Exact installed retained record count')
  epoch_view=None
  for r,want in zip(records,wants):
   for k in ['id','method','credential','transport']:require(k in r and same(r[k],want[k]),'Exact ordered installed '+k)
   expected_url(r['url'],want['url'])
   if want['transport']=='public-exact':
    require(type(r['status']) is int and r['status']==want['status'],'Exact installed status')
    epoch_view=header_context(r['headers'],want['position'],epoch_view) or epoch_view
    fields=[{k:(epoch_view[0] if v=='$epoch' else epoch_view[1] if v=='$view' else v) for k,v in f.items()} for f in want['fieldOccurrences']]
    require(same(r['fieldOccurrences'],fields),'Exact source-derived occurrence/order/value grammar')
    if r['method']=='HEAD' or r['status'] in [204,304,405,406,412]:require(r['bodyOctets']==0 and r['bodySha256']==sha(b''),'Bodyless installed response')
    if r['id']=='N1':require(r['headers'].get('link')=='<'+schemas['canonicalScope']+'bdp.json>; rel="service-desc"; type="application/json"' and r['headers'].get('cache-control')=='private, no-store','Scope Link/privacy')
   else:
    require(want['id'] in ['wrong-host','pending-minimum'],'Finite raw terminal identity')
    if want['id']=='wrong-host':require(r.get('partialStatus')==400 and type(r.get('partialStatus')) is int,'Actual raw Host refusal status')
  require(len({r['id'] for r in records})==len(records),'No duplicated record IDs')
 def source_report(o):
  keys(o,['observation','callback','elapsedMs','workSettled','cleanupKnown','directories','observations','failureCount']);exact(o,{'observation':'D3-source','workSettled':True,'cleanupKnown':True,'failureCount':0});finite(o['elapsedMs'],4000)
  require(o['callback'] in cfg['selectedFullNames'],'Report callback identity');want=schemas['sourceExpected'][o['callback']]
  require(type(o['directories']) is list and type(o['observations']) is list,'Finite source arrays')
  for d in o['directories']:
   keys(d,['directory','absent']);exact(d,{'absent':True});absent_dir(d['directory'],'bdp-d3-')
  counters=[];raw=[];clones=[];baselines=[];kinds=[]
  for r in o['observations']:
   require(type(r) is dict,'Actual observation object')
   if 'baseline' in r:
    kinds.append('baseline');keys(r,['baseline']);keys(r['baseline'],['calls','entries','assemblies']);require(same(r['baseline'],want['baseline']),'Exact source-derived construction baseline');baselines.append(r)
   elif 'counters' in r:
    kinds.append('counters');keys(r,['counters','records']);keys(r['counters'],['direct','discovery','entries','assemblies','waits','clocks','commits']);require(same(r['counters'],want['counters']),'Typed actual source counters')
    require(type(r['records']) is list and len(r['records'])==want['returnedCount'],'Exact retained returned-response census, separate from started-direct count')
    returned_context=None
    for rr,expected in zip(r['records'],want['returned']):
     source_response(rr,expected);returned_context=header_context(rr['headers'],expected['position'],returned_context) or returned_context
    counters.append(r)
   elif 'counterfactual' in r:
    kinds.append('clone');keys(r,['counterfactual','rejected','wireRequests']);exact(r,{'rejected':True,'wireRequests':0});require(type(r['counterfactual']) is str,'Actual clone label');clones.append(r['counterfactual'])
   else:
    kinds.append('raw');keys(r,['transport','method','target','credential','fieldOccurrences','status','headers','bodyOctets','bodySha256','wireSha256']);exact(r,{'transport':'source-local-raw'});headers_record(r);integer(r['bodyOctets']);digest(r['bodySha256']);digest(r['wireSha256']);require(r['bodyOctets']<=131072,'Actual raw body bound')
    fields=r['fieldOccurrences'];require(type(fields) is list and all(type(f) is list and len(f)==2 and all(type(x) is str for x in f) for f in fields),'Source raw exact tuple occurrence grammar');raw.append(r)
  require(kinds==want['observationKinds'] and len(baselines)==want['baselineCount'],'Exact baseline/returned/raw/clone observation order')
  require(len(raw)==want['wire'] and len(clones)==want['clones'],'Exact raw/clone census')
  for r,expected in zip(raw,want['rawRows']):
   for key in ['method','target','credential','status','fieldOccurrences']:require(same(r[key],expected[key]),'Exact raw source row '+key)
   header_context(r['headers'],None)
  if clones:require(Counter(clones)==Counter(schemas['sourceCounterfactuals']),'Exact clone labels')
  require(len(counters)==want['baselineCount'],'Exact source counter object occurrence')
 # G09 adapter: exact frozen reports, not a report union or conformance claim.
 def g09_json(text):
  require(type(text) is str and len(text.encode('utf-8'))<=131072,'G09 JSON response bound')
  return json.loads(text,object_pairs_hook=unique_object,parse_constant=invalid_constant)
 def g09_exact_schema(h,text,body,head=False):
  require(text==('' if head else body),'G09 exact captured UTF-8 schema')
  exact(h,{'content-length':str(len(body.encode('utf-8'))),'content-type':'application/json','cache-control':'private, no-store','vary':'Accept'})
  require('etag' not in h and 'last-modified' not in h,'G09 no invented conditional validator')
 def g09_problem(h,text,status,code,head=False):
  exact(h,{'content-type':'application/problem+json'})
  require(re.fullmatch('[1-9][0-9]*',h.get('content-length','')),'G09 represented problem length')
  if head:require(text=='','G09 HEAD problem has no transferred body')
  else:
   v=g09_json(text);require(type(v) is dict,'G09 actual problem object')
   family,retry=schemas['G09']['problems'][code]
   exact(v,{'code':code,'status':status,'type':'https://github.com/gastownhall/bdp/problems/'+family,'retry':retry})
   require(h['content-length']==str(len(text.encode('utf-8'))),'G09 problem byte length')
  if status==503:exact(h,{'retry-after':'1'})
 def g09_wire(r):
  import base64
  value=r['wireBase64'];require(type(value) is str and len(value)<=4*((163840+2)//3),'G09 bounded complete wire encoding')
  try:raw=base64.b64decode(value,validate=True)
  except (ValueError,TypeError):raise RuntimeError('G09 malformed base64 wire')
  require(base64.b64encode(raw).decode()==value and 0<len(raw)<=163840 and sha(raw)==r['wireSha256'],'G09 exact full wire bytes/fingerprint')
  offset=0
  while True:
   end=raw.find(b'\r\n\r\n',offset);require(end>=offset and end+4<=16384,'G09 actual executor header budget')
   lines=raw[offset:end].split(b'\r\n');match=re.fullmatch(rb'HTTP/1\.[01] ([1-5][0-9]{2})(?: .*|)',lines[0]);require(match is not None,'G09 actual HTTP status line')
   status=int(match.group(1));headers={}
   for line in lines[1:]:
    require(b':' in line,'G09 actual header separator');name,value=line.split(b':',1)
    name=name.decode('latin1').strip().lower();value=value.decode('latin1').strip()
    require(re.fullmatch("[!#$%&'*+.^_`|~0-9a-z-]+",name) and not any(c in value for c in ['\x00','\r','\n']),'G09 actual header syntax')
    headers[name]=headers[name]+', '+value if name in headers else value
   offset=end+4
   if status>=200:break
   require(status!=101 and 'content-length' not in headers and 'transfer-encoding' not in headers,'G09 bounded informational response')
  require(status==r['status'] and same(headers,r['responseHeaders']),'G09 raw status/header projection equality')
  framed=raw[offset:]
  if r['method']=='HEAD' or status in [204,304]:
   require(framed==b'','G09 truly bodyless raw response');body=framed
  elif 'transfer-encoding' in headers:
   require(headers['transfer-encoding'].lower()=='chunked' and 'content-length' not in headers,'G09 exact framing without ambiguity')
   chunks=[];size=0;cursor=0
   while True:
    end=framed.find(b'\r\n',cursor);require(end>=cursor,'G09 complete chunk header')
    line=framed[cursor:end];part=line.split(b';',1)[0];require(re.fullmatch(rb'[0-9A-Fa-f]+',part),'G09 actual hex chunk size');n=int(part,16);cursor=end+2
    if n==0:
     while True:
      end=framed.find(b'\r\n',cursor);require(end>=cursor,'G09 complete chunk trailer');trailer=framed[cursor:end];cursor=end+2
      if not trailer:break
      require(b':' in trailer and b'\x00' not in trailer,'G09 finite trailer syntax')
     require(cursor==len(framed),'G09 no trailing wire bytes');break
    require(n<=131072-size and cursor+n+2<=len(framed) and framed[cursor+n:cursor+n+2]==b'\r\n','G09 chunk body bound/completeness')
    chunks.append(framed[cursor:cursor+n]);size+=n;cursor+=n+2
   body=b''.join(chunks)
  else:
   body=framed
   if 'content-length' in headers:
    require(re.fullmatch('0|[1-9][0-9]*',headers['content-length']) and int(headers['content-length'])==len(body),'G09 exact content length framing')
  require(len(body)<=131072 and type(r['bodyOctets']) is int and r['bodyOctets']==len(body) and type(r['bodyText']) is str and body.decode('utf-8')==r['bodyText'],'G09 actual raw/body UTF-8/octet equality')
 def g09_source(o):
  rules=schemas['G09'];identity=o['observation'];require(identity in sel['reportIds'],'G09 selected report identity')
  if identity in rules['sourceService']['observations']:
   keys(o,rules['sourceService']['keys']);exact(o,{**rules['sourceService']['positiveCounts'][identity],'closeCalls':1,'closed':True,'removed':True,'failureCount':0})
   positions=rules['servicePositions'][identity];require(type(o['contexts']) is list and len(o['contexts'])==len(positions),'G09 exact actual read-entry census');ev=None
   for c,p in zip(o['contexts'],positions):
    context(c,p);v=[c['scopeEpoch'],c['authorizationView']];require(ev is None or ev==v,'G09 same actual service epoch/view');ev=v
  elif identity=='G09-S04':keys(o,rules['sourceSelector'][identity]['keys']);exact(o,{'calls':8,'urlReads':0,'observations':0})
  elif identity=='G09-S05':
   keys(o,rules['sourceSelector'][identity]['keys']);exact(o,{'calls':8,'abortCalls':1,'abortPropagated':True,'abortSignalForwarded':True})
   require(type(o['rows']) is list and len(o['rows'])==8,'G09 selector row census')
   for r,status,method in zip(o['rows'],rules['sourceSelector'][identity]['statuses'],['GET','HEAD','GET','HEAD','GET','GET','HEAD','GET']):keys(r,['status','method']);exact(r,{'status':status,'method':method})
  else:
   keys(o,rules['sourceWire']['topKeys']);wants=rules['sourceRows'][identity];exact(o,{'attempted':len(wants),'completed':len(wants),'pending':0,'closed':True,'removed':True,'failureCount':0})
   require(type(o['rows']) is list and len(o['rows'])==len(wants),'G09 source wire row census');ev=None
   for ordinal,(r,w) in enumerate(zip(o['rows'],wants),1):
    keys(r,rules['sourceWire']['rowKeys']);exact(r,{'ordinal':ordinal,'method':w['method'],'url':w['url'],'status':w['status'],'settled':True,'received':True,'validated':True,'errorCategory':None});strings(r['headers']);require(type(r['text']) is str,'G09 actual source response text')
    ev=header_context(r['headers'],w['position'],ev) or ev
    if w.get('artifact'):g09_exact_schema(r['headers'],r['text'],rules['sourceArtifacts'][w['artifact']],r['method']=='HEAD')
    elif w.get('problem'):g09_problem(r['headers'],r['text'],w['status'],w['problem'],r['method']=='HEAD')
 def g09_installed(o):
  rules=schemas['G09'];keys(o,rules['installed']['topKeys']);identity=sel['reportIds'][0];wants=rules['installedRows'][identity]
  exact(o,{'observation':identity,'declaredSourceCommit':g['finalHead'],'workspaceDirty':True,'readyValidated':True,'attemptedRequests':len(wants),'completedResponses':len(wants),'pending':0,'directoryAbsent':True,'outcome':'passed','failureCount':0,'claimEligible':False,'authenticationIdentityRetention':'not emitted; actual credential binding is source-attested','evidenceRetention':'Full actual returned wire/body bytes; unavailable partial exchange evidence remains null; child stdout/stderr bytes are not retained'})
  finite(o['elapsedMs'],60000);proof=verify_consumer(g)
  require(type(o['entry']) is str and Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'],'G09 actual external entry bytes')
  ready=o['ready'];keys(ready,rules['installed']['readyKeys']);exact(ready,{'scope':rules['canonicalScope'],'profile':'transactional','history':False,'authorityMode':'canonical','claimEligible':False,'payloadIdentity':proof['payloadIdentity']});digest(ready['runtimeConfigurationIdentity'])
  require(type(ready['listenUrl']) is str and re.fullmatch(r'http://127\.0\.0\.1:([1-9][0-9]{0,4})/',ready['listenUrl']) and 0<int(ready['listenUrl'].split(':')[2][:-1])<=65535,'G09 actual loopback listener')
  require(same(o['artifactBytes'],rules['installedArtifacts']),'G09 exact captured artifact provenance')
  session=o['session'];keys(session,rules['installed']['sessionKeys']);exact(session,{'closePromiseSettled':True,'closeSucceeded':True,'stdoutEnded':True,'stderrEnded':True})
  for key in ['exit','close']:keys(session[key],['code','signal']);exact(session[key],{'code':0,'signal':None})
  absent_pid(session['pid']);require(type(o['rows']) is list and len(o['rows'])==len(wants),'G09 one actual response per scheduled request');ev=None
  for ordinal,(r,w) in enumerate(zip(o['rows'],wants),1):
   keys(r,rules['installed']['rowKeys']);exact(r,{'ordinal':ordinal,'label':w['label'],'method':w['method'],'url':w['url'],'status':w['status'],'settled':True,'received':True,'validated':True,'writeState':'complete','errorCategory':None})
   strings(r['responseHeaders']);digest(r['wireSha256']);integer(r['requestBodyBytes']);digest(r['requestBodySha256']);body=rules['createBody'].encode('utf-8') if w['label']=='create' else b''
   require(r['requestBodyBytes']==len(body) and r['requestBodySha256']==sha(body),'G09 exact source-derived request body bytes')
   g09_wire(r);h=r['responseHeaders'];text=r['bodyText'];exact(h,{'x-bdp-development':'non-attesting'});ev=header_context(h,w['position'],ev) or ev
   if w.get('artifact'):g09_exact_schema(h,text,rules['installedArtifactText'][w['artifact']],r['method']=='HEAD')
   elif w.get('problem'):g09_problem(h,text,w['status'],w['problem'],r['method']=='HEAD')
   elif w.get('gap'):exact(h,{'x-bdp-development-limit':w['gap']});require(text=='','G09 retained private limitation has no body/credit')
   elif w['label']=='create':
    exact(h,{'content-type':'application/json'});receipt=g09_json(text)
    keys(receipt,['id','status','detail','idempotencyKey','scopeEpoch','authorizationView','transaction','requiredPosition','effectPosition','results','next','expiresAt'])
    exact(receipt,{'status':'completed','detail':'available','idempotencyKey':'g09-create','scopeEpoch':ev[0],'authorizationView':ev[1],'requiredPosition':'pos-1','effectPosition':'pos-1','next':None})
    require(type(receipt['id']) is str and re.fullmatch(re.escape(rules['canonicalScope']+'receipts/')+r'[A-Za-z0-9_-]+',receipt['id']),'G09 actual receipt target');token(receipt['transaction'])
    require(type(receipt['expiresAt']) is str and re.fullmatch(r'\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z',receipt['expiresAt']),'G09 actual receipt expiry')
    require(type(receipt['results']) is list and len(receipt['results'])==1,'G09 one create-bead result');item=receipt['results'][0]
    keys(item,['operationIndex','outcome','resource'],['operationName']);exact(item,{'operationIndex':0,'outcome':'created'})
    if 'operationName' in item:require(type(item['operationName']) is str and len(item['operationName'])>0,'G09 actual operation name')
    resource=item['resource'];keys(resource,['id','type','revision','properties']);exact(resource,{'id':rules['canonicalScope']+'beads/g09','type':rules['canonicalScope']+'types/task','properties':{'title':'café'}});require(type(resource['revision']) is str and len(resource['revision'])>0,'G09 actual created revision')
   elif w['status'] in [304,412,406,405]:
    require(text=='' and r['bodyOctets']==0,'G09 conditional/media/method bodyless')
    if w['status']==304:require('content-length' not in h,'G09 actual304 absent length')
    if w['status']==412:exact(h,{'content-length':'0'})
    if w['status']==405:exact(h,{'allow':'GET, HEAD'})
   else:raise RuntimeError('Uncovered G09 finite response oracle')
 if kind=='G09-source':
  require(Counter(o.get('observation') for o in reports)==Counter(sel['reportIds']),'G09 exact selected report census')
  for o in reports:g09_source(o)
 elif kind=='G09-installed':
  require(len(reports)==1,'G09 exact singleton installed observation');g09_installed(reports[0])
 elif kind=='G09-legacy-source':require(reports==[],'Unexpected report from selected legacy source callbacks')
 elif kind=='D3-source':
  require(Counter(o.get('callback') for o in reports)==Counter(cfg['selectedFullNames']) and all(o.get('observation')=='D3-source' for o in reports),'Exactly21 source reports')
  for o in reports:source_report(o)
 elif kind in ['legacy-installed-contracts-suite','schema-zero-wire']:
  require(reports==[],'Unexpected observation from no-report suite')
 elif kind=='legacy-changefeed-suite':
  require(Counter(o['observation'] for o in reports)==Counter(schemas['legacyChangefeedExpectedLabels']),'Exact historical report census')
  for o in reports:
   shape=schemas['legacyChangefeedShapes'][o['observation']];keys(o,shape['required'],shape.get('optional',[]))
   if 'outcome' in o:exact(o,{'outcome':'passed'})
   for field in ['failureCount','hardFailureCount']:
    if field in o:exact(o,{field:0})
   if 'claimEligible' in o:exact(o,{'claimEligible':False})
   if 'directoryAbsent' in o:exact(o,{'directoryAbsent':True})
   for field in ['deadlineExceeded','cleanupDeadlineExceeded']:
    if field in o:exact(o,{field:False})
   if 'cleanupState' in o:
    state=o['cleanupState'];exact(state,{'closed':True,'stopUnknown':False,'directoryAbsent':True});absent_dir(state['directory'],'bdp-current-feed-')
   if 'directory' in o and o['directory'] is not None:absent_dir(o['directory'],'bdp-')
   if 'contextFailures' in o:require(o['contextFailures']==[],'Actual context failures')
  grouped={label:[o for o in reports if o['observation']==label] for label in set(schemas['legacyChangefeedExpectedLabels'])}
  a=grouped['current-changefeed-minimum-owner'][0];scalars(a,{'commits':2,'ownerLifetimes':2,'offlineDeletes':1})
  rr=rows(a,'records',18);current=rr[0]['context'];context(current,'pos-2')
  require([r['role'] for r in rr]==['baseline','baseline']+['table-'+str(i) for i in range(1,17)],'Owner ordered actual calls')
  expected=['page','page','TransactionMinimumUnproved','TransactionMinimumUnproved','page','page',
            'TransactionMinimumFailure','TransactionMinimumFailure','TransactionMinimumFailure',
            'TransactionMinimumFailure','TransactionMinimumFailure','TransactionChangefeedParameterError',
            'TransactionChangefeedParameterError','unsupported','TransactionMinimumUnproved','TypeError','Error','Error']
  require([r['outcome'] for r in rr]==expected and [r['ordinal'] for r in rr]==list(range(1,19)),'Owner outcome/ordinal table')
  for i,r in enumerate(rr):
   require(r['context']==(None if i in [13,15,16,17] else current),'Owner no invented context')
   if i>=2:require(r['admission']==('finite-start-now' if i==14 else 'current-changefeed'),'Owner tag preservation')
  require([rr[i]['reason'] for i in [6,7,8,9,10]]==['cursor-expired','foreign-view','cursor-expired','cursor-expired','foreign-view'],'Cursor/minimum precedence')
  captures=grouped['current-changefeed-minimum-capture']
  require(sorted(o['label'] for o in captures)==sorted(['start-now-unknown','genesis-unknown','retained-unknown','retained-pos-01']),'Capture cases')
  for o in captures:
   scalars(o,{'commits':2,'waits':1,'clocks':4,'closes':1,'captures':1 if o['label'].startswith('start') else 0})
   rr=rows(o,'reads',4);context(rr[0]['context'],'pos-1');context(rr[1]['context'],'pos-1');context(rr[2]['context'],'pos-2');context(rr[3]['context'],'pos-2')
   require(rr[0]['after']=='genesis' and rr[0]['minimum'] is None and rr[0]['admission'] is None,'Baseline traverses actual recording delegate')
   for r in rr[1:3]:
    scalars(r,{'admission':'current-changefeed','tupleFrozen':True,'outcome':'TransactionMinimumUnproved','after':o['originalAfter']})
   require(rr[1]['minimum']==rr[2]['minimum'] and rr[1]['minimum']['minimumPosition']==('pos-01' if o['label'].endswith('pos-01') else 'unknown'),'One immutable tuple')
   require(rr[3]['after']==o['originalAfter'] and rr[3]['minimum']['minimumPosition']=='pos-2' and rr[3]['outcome']=='page','Known positive same anchor')
   require(o['anchors']==([o['originalAfter']] if o['label'].startswith('start') else []),'Once-only start anchor')
   positive=o['positiveGroup'];position=1 if o['originalAfter']=='genesis' else 2
   scalars(positive,{'position':'pos-'+str(position),'previousPosition':'pos-'+str(position-1),'projectionAdvance':False})
   require(rows(positive,'changes',1)[0]['resource']['id']=='https://example.test/feed/beads/'+('a' if position==1 else 'b') and type(positive['eventCount']) is int and positive['eventCount']==len(positive['events']) and positive['eventCount']>0,'Actual whole positive capture group')
   cleanup(o['cleanupState'],'bdp-current-feed-')
  drains=grouped['current-changefeed-minimum-drain'];require(sorted(o['label'] for o in drains)==['retained','start-now'],'Drain cases')
  for o in drains:
   scalars(o,{'readsBeforeClosedProbe':2,'rawClosedProbes':1,'serviceClosedProbes':1,'waits':1,'clocks':2,'closes':1,'commits':1,'closeSettled':True,'captures':1 if o['label']=='start-now' else 0})
   rr=rows(o,'reads',2);require(rr[0]['after']=='genesis' and rr[0]['minimum'] is None and rr[1]['outcome']=='TransactionMinimumUnproved','Actual drain calls')
   require(o['order']==['wait-entered','abort','wait-released','owner-close','close-resolved'],'Drain/owner closure order')
   cleanup(o['cleanupState'],'bdp-current-feed-')
  f=grouped['current-changefeed-minimum-preflight'][0]
  for r,label in zip(rows(f,'rows',4),['already-aborted','invalid-policy','throwing-clock','nonfinite-clock']):
   scalars(r,{'label':label,'initialized':True,'attempted':True,'closed':True,'captures':1,'reads':0,'waits':0,'clocks':0 if label=='invalid-policy' else 1,'closes':1,'commits':1,'outcome':'passed','failureCount':0})
   require(r['errorClass']==('ScopeServerOperationAbortedError' if label=='already-aborted' else 'Error' if label=='throwing-clock' else 'TypeError'),'Preflight actual error class')
   cleanup(r,'bdp-current-feed-')
  legacy=grouped['changefeed-opaque-boundaries'][0];scalars(legacy,{'outcome':'passed','failureCount':0,'captures':2,'reads':4,'waits':0,'clocks':2})
  require(legacy['outcomes']==['default-owner-TypeError','default-service-TypeError','after-TypeError','malformed-TypeError','unsupported-sentinel','foreign-cursor-expired','valid-closed','malformed-before-closed','owner-closed'],'Unchanged old/default strict contract')
  legacy=grouped['changefeed-opaque-http'][0];scalars(legacy,{'outcome':'passed','failureCount':0,'attempted':10,'exchanges':10,'deadlineExceeded':False,'cleanupDeadlineExceeded':False,'closed':True,'directoryAbsent':True})
  require([r['status'] for r in rows(legacy,'records',10)]==[200,200,200,503,503,503,503,503,503,400],'Only accepted legacy status successor')
  directory_absent(legacy['directory'],'bdp-feed-opaque-http-')
  historical=grouped['changefeed-opaque-capture'];require(Counter(o['row'] for o in historical)==Counter(['unknown','pos-01']),'Historical capture row identity')
  for o in historical:
   scalars(o,{'commits':2,'captures':2,'waits':1,'clocks':4,'closes':1});rr=rows(o,'reads',4);anchor=o['positiveAfter'];token(anchor)
   require(o['anchors']==[anchor,anchor],'Actual repeated captured anchor')
   for i,r in enumerate(rr):
    keys(r,['after','minimum','minimumFrozen','admission','outcome','context']);require(r['after']==anchor,'Captured anchor preserved');context(r['context'],'pos-1' if i<2 else 'pos-2')
    keys(r['minimum'],['scopeEpoch','authorizationView','minimumPosition']);require([r['minimum'][k] for k in ['scopeEpoch','authorizationView']]==[rr[0]['context'][k] for k in ['scopeEpoch','authorizationView']],'Historical same tuple identity')
    require(r['minimum']['minimumPosition']==('pos-2' if i==3 else o['row']),'Historical minimum opaque preservation')
    require(type(r['minimumFrozen']) is bool and r['minimumFrozen']==(i!=0),'Actual tuple snapshot freezing')
    require(r['outcome']==('page' if i==3 else 'unproved') and r['admission']==(None if i==3 else 'finite-start-now'),'Historical admission/outcome')
  o=grouped['changefeed-opaque-abort'][0];scalars(o,{'captures':1,'reads':1,'waits':1,'clocks':2,'closes':1,'postCloseProbes':1});require(o['order']==['wait-entered','abort','wait-released','owner-close','close-resolved'],'Historical abort order')
  pre=grouped['changefeed-opaque-preflight'];require(Counter(o['row'] for o in pre)==Counter(['already-aborted','invalid-policy','throwing-clock','nonfinite-clock']),'Historical preflight census')
  for o in pre:
   label=o['row'];scalars(o,{'captures':1,'reads':0,'waits':0,'clocks':0 if label=='invalid-policy' else 1,'errorName':'ScopeServerOperationAbortedError' if label=='already-aborted' else 'Error' if label=='throwing-clock' else 'TypeError'})
  o=grouped['private-source-changefeed-parameter-context'][0];scalars(o,{'attempted':12,'exchanges':12,'hardFailureCount':0,'deadlineExceeded':False,'startupCompleted':True,'closed':True,'directoryAbsent':True,'contextFailures':[]});rr=rows(o,'completedResponses',12);scope='https://example.test/feed/'
  require([r['status'] for r in rr]==[200,200,400,400,400,400,400,400,401,400,200,400],'Historical parameter statuses')
  future=rr[2]['url'];require(re.fullmatch(re.escape(scope+'changes/?after=')+r'cg_[A-Za-z0-9_-]{43}_[A-Za-z0-9_-]{43}_3',future),'Actual future checkpoint URL')
  urls=[scope+'operations/batch',scope+'changes/?after=genesis']+[future]*4+[scope+'changes/?after=genesis']*2+[future,scope+'changes/?after=genesis&start=now',scope+'operations/batch',future]
  methods=['POST','GET','GET','HEAD','GET','HEAD','GET','HEAD','GET','GET','POST','GET']
  accepts=['application/json']*4+['text/event-stream']*4+['application/json']*4
  for i,r in enumerate(rr):
   keys(r,['ordinal','method','url','accept','reconnect','status','bodyOctets','context','contentType','contentLength']);scalars(r,{'ordinal':i+1,'method':methods[i],'url':urls[i],'accept':accepts[i],'reconnect':'malformed' if i in [6,7] else None});integer(r['bodyOctets'])
   require(type(r['context']) is list and len(r['context'])==3,'Historical context vector')
   if i in [8,9]:require(r['context']==[None,None,None],'Historical no disclosed tuple')
   else:
    for v in r['context']:token(v)
    require(r['context'][:2]==rr[0]['context'][:2] and r['context'][2]==('pos-2' if i>=10 else 'pos-1'),'Historical current tuple')
   require(r['contentType'] is None or type(r['contentType']) is str,'Content type nullable string');require(r['contentLength'] is None or type(r['contentLength']) is str and r['contentLength'].isdigit(),'Content length nullable decimal')
   if methods[i]=='HEAD':require(r['bodyOctets']==0,'Historical HEAD bytes')
  o=grouped['changefeed-opaque-http'][0];rr=rows(o,'records',10);scope='https://example.test/feed/';methods=['POST','GET','HEAD','GET','HEAD','GET','HEAD','GET','GET','GET']
  urls=[scope+'operations/batch']+[scope+'changes/?start=now']*6+[scope+'changes/?after=genesis']+[scope+'changes/?start=now']*2
  current=None
  for i,r in enumerate(rr):
   keys(r,['ordinal','method','url','status','bodyOctets','bodySha256','headers']);scalars(r,{'ordinal':i+1,'method':methods[i],'url':urls[i]});integer(r['bodyOctets']);digest(r['bodySha256']);h=r['headers'];require(type(h) is dict and all(type(k) is str and (v is None or type(v) is str) for k,v in h.items()),'Historical nullable header map')
   keys(h,context_names+['content-type','content-length','cache-control','retry-after','etag','x-bdp-development-limit'])
   if i==9:require(all(h.get(k) is None for k in context_names),'Historical malformed no tuple')
   else:
    c=[h.get(k) for k in context_names]
    for v in c:token(v)
    require(c[2]=='pos-1' and (current is None or c==current),'Historical actual tuple');current=c
   if methods[i]=='HEAD':require(r['bodyOctets']==0,'Historical HEAD bytes')
   if 3<=i<=8:require(h.get('content-type')=='application/problem+json' and h.get('cache-control')=='private, no-store' and h.get('retry-after')=='1' and h.get('etag') is None and h.get('x-bdp-development-limit') is None,'Historical timeout metadata')
  for get,head in [(1,2),(3,4),(5,6)]:require(rr[head]['headers'].get('content-length')==str(rr[get]['bodyOctets'])==rr[get]['headers'].get('content-length'),'Historical represented GET/HEAD lengths')
 else:
  require(len(reports)==1,'Exactly one installed observation');o=reports[0]
  require(o['observation']==cfg['expectedObservation'],'Installed observation identity');exact(o,{'outcome':'passed','claimEligible':False,'workspaceDirty':True,'declaredSourceCommit':g['finalHead']})
  proof=verify_consumer(g);require(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'],'Actual external entry bytes')
  if kind=='D3-installed':
   keys(o,schemas['installedD3Required']);exact(o,{'failureCount':0,'workSettled':True,'cleanupKnown':True,'pending':0,'directoryAbsent':True,'entryLiteral':proof['entryRealpath']});finite(o['elapsedMs'],60000);absent_dir(o['directory'],'bdp-d3-installed-')
   keys(o['counts'],['scheduled','completed','aborts','exact','raw']);exact(o['counts'],cfg['expectedCounts'])
   keys(o['child'],['pid','exitSeen','closeSeen','code','signal']);exact(o['child'],{'exitSeen':True,'closeSeen':True,'code':0,'signal':None});absent_pid(o['child']['pid'])
   keys(o['metadata'],schemas['metadataKeys']);exact(o['metadata'],{'scope':schemas['canonicalScope'],'profile':'transactional','history':False,'claimEligible':False,'authorityMode':'canonical'});require(o['metadata']['payloadIdentity']==proof['payloadIdentity'],'Actual child payload identity')
   for k in ['payloadIdentity','policyIdentity','authenticationIdentity','runtimeConfigurationIdentity']:digest(o['metadata'][k])
   require(type(o['metadata']['artifactIdentity']) is str and re.fullmatch('development-schema5-[a-f0-9]{64}',o['metadata']['artifactIdentity']),'Versioned actual artifact identity')
   require(type(o['records']) is list and len(o['records'])==cfg['expectedCounts']['scheduled'],'One actual terminal per scheduled request');exact(o,{'sourceAttribution':'externally bound by exact source/archive/consumer/lock/proof; no child attestation'})
   exact_records=[];raw_records=[]
   for r in o['records']:
    if r.get('transport')=='public-exact':
     keys(r,['id','transport','method','url','credential','fieldOccurrences','status','headers','bodyOctets','bodySha256','wireSha256','requestWriteState']);headers_record(r);exact(r,{'requestWriteState':'complete'});digest(r['bodySha256']);digest(r['wireSha256']);integer(r['bodyOctets']);require(r['bodyOctets']<=131072,'Body cap');named_fields(r['fieldOccurrences']);exact_records.append(r)
    else:
     keys(r,['id','transport','method','url','credential','state','terminal','partialWireAvailable','actualWireOctets'],['partialStatus','actualWireSha256']);exact(r,{'transport':'callback-local-raw','method':'GET'});integer(r['actualWireOctets']);require(r['actualWireOctets']<=163840,'Actual raw wire bound');require(type(r['partialWireAvailable']) is bool and r['partialWireAvailable']==(r['actualWireOctets']>0),'Honest raw availability')
     require(('actualWireSha256' in r)==r['partialWireAvailable'],'No invented wire fingerprint')
     if r['partialWireAvailable']:digest(r['actualWireSha256'])
     keys(r['state'],['settled','responseStarted','written','closeRequested'],['writeAtMs','closeRequestedAtMs','responseAtMs']);exact(r['state'],{'settled':True,'written':True})
     require(all(type(r['state'][k]) is bool for k in ['settled','responseStarted','written','closeRequested']),'Typed actual raw lifecycle')
     for k in ['writeAtMs','closeRequestedAtMs','responseAtMs']:
      if k in r['state']:finite(r['state'][k],60000)
     if cfg['expectedCounts']['aborts']:
      exact(r,{'terminal':'abort','partialWireAvailable':False,'actualWireOctets':0});exact(r['state'],{'closeRequested':True,'responseStarted':False});require('partialStatus' not in r,'No invented close status');finite(r['state']['writeAtMs']);finite(r['state']['closeRequestedAtMs']);require(r['state']['writeAtMs']<=r['state']['closeRequestedAtMs'],'Actual close chronology')
     else:
      exact(r['state'],{'responseStarted':True,'closeRequested':False});require(r['terminal']=='response' and type(r['partialStatus']) is int and r['partialStatus']==400 and r['partialWireAvailable'] is True,'Actual Host refusal response');finite(r['state']['writeAtMs'],60000);finite(r['state']['responseAtMs'],60000);require('closeRequestedAtMs' not in r['state'],'No invented Host close chronology')
     raw_records.append(r)
   require(len(exact_records)==cfg['expectedCounts']['exact'] and len(raw_records)==cfg['expectedCounts']['raw'],'Actual exact/raw census')
   installed_rows(o['records'],schemas['installedExpected'][stage])
  elif kind=='legacy-owner-installed':
   keys(o,schemas['legacyOwnerRequired']);exact(o,{'cleanupKnown':True,'directoryAbsent':True,'exchanges':12});absent_dir(o['directory'],'bdp-discovery-installed-');keys(o['session'],['pid','exit']);exact(o['session'],{'exit':0});absent_pid(o['session']['pid']);keys(o['childTerminal'],['exitSeen','closeSeen','code','signal']);exact(o['childTerminal'],{'exitSeen':True,'closeSeen':True,'code':0,'signal':None});require(len(o['completedResponses'])==12,'Legacy12 actual responses')
   for r in o['completedResponses']:
    keys(r,['method','url','status','bodyOctets','bodySha256','wireSha256']);integer(r['bodyOctets']);digest(r['bodySha256']);digest(r['wireSha256'])
   require([r['status'] for r in o['completedResponses']]==[204,200,200,204,200,200,401,401,401,200,200,200],'Actual legacy12 statuses')
   served_binding(o['binding'],proof)
   for r,want in zip(o['completedResponses'],schemas['legacyOwnerRows']):
    exact(r,want);finite_url(r['url']);integer(r['status'])
    if r['method']=='HEAD' or r['status']==204:require(r['bodyOctets']==0,'Legacy owner bodyless response')
  elif kind=='legacy-schema-installed':
   keys(o,schemas['legacySchemaRequired']);exact(o,{'attemptedRequests':14,'directoryAbsent':True});require(len(o['completedResponses'])==14 and len(o['sessions'])==1,'Schema14 exchanges/one child lifetime')
   require([r['label'] for r in o['completedResponses']]==schemas['schemaLabels'],'Schema request order')
   for session in o['sessions']:
    keys(session,['pid','exit','signal']);exact(session,{'exit':0,'signal':None});absent_pid(session['pid'])
   for r in o['completedResponses']:
    keys(r,['label','method','url','status','requestBodyBytes','requestBodySha256','responseBodyBytes','responseBodySha256','wireSha256','context']);integer(r['requestBodyBytes']);integer(r['responseBodyBytes']);digest(r['requestBodySha256']);digest(r['responseBodySha256']);digest(r['wireSha256'])
   require([r['status'] for r in o['completedResponses']]==[200,200,200,200,200,406,304,412,404,401,405,204,200,200],'Actual schema14 status order')
   served_binding(o['binding'],proof);require(same(o['routes'],schemas['schemaRoutes']) and same(o['observedGrants'],schemas['schemaGrants']) and same(o['artifactBytes'],schemas['schemaArtifacts']),'Actual schema administrative routes/grants/exact artifact bytes')
   exact(o,{'evidenceRetention':'Response fingerprints/context only; full wire/body bytes are not retained'})
   epoch_view=None
   for r,want in zip(o['completedResponses'],schemas['schemaRows']):
    exact(r,{k:want[k] for k in ['label','method','url','status']});finite_url(r['url']);integer(r['status']);c=r['context']
    if want['position'] is None:require(same(c,{}),'No fabricated schema context on authorization/method refusal')
    else:
     keys(c,['epoch','view','position'])
     for v in c.values():token(v)
     require(c['position']==want['position'] and (epoch_view is None or [c['epoch'],c['view']]==epoch_view),'Exact schema current tuple');epoch_view=[c['epoch'],c['view']]
    require(r['requestBodyBytes']<=65536 and r['responseBodyBytes']<=131072,'Schema actual request/response caps')
    if r['label']!='create':require(r['requestBodyBytes']==0 and r['requestBodySha256']==sha(b''),'No invented request body')
    if r['method']=='HEAD' or r['status'] in [204,304,405,406,412]:require(r['responseBodyBytes']==0 and r['responseBodySha256']==sha(b''),'Schema truly empty response')
    if r['label'] in ['root','referenced','root-after']:
     artifact=schemas['schemaArtifacts'][1 if r['label']=='referenced' else 0];require(r['responseBodyBytes']==artifact['bytes'] and r['responseBodySha256']==artifact['sha256'],'Exact local schema response bytes')
  else:raise RuntimeError('Unknown observation grammar')
 write(stage+'.accepted-observation.json',{'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes()),'selectedFullNames':cfg['selectedFullNames'],'reports':reports,'claimEligible':False,'limitations':'Private exact source/build/archive/consumer only. Historical lanes distinct. Failure/partial records remain raw, never reconstructed; no public Write qualification/admission claim.'})
