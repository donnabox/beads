# Static draft arithmetic only: reads text/Git; never imports or executes product code.
from pathlib import Path
import json,hashlib,subprocess,re,difflib,collections
W=Path('/Users/dbox/repos/wt/vickie/vickie-command-minimum');E=Path(__file__).parent
J=lambda x:json.dumps(x,separators=(',',':'),ensure_ascii=False)
H=lambda x:hashlib.sha256(x if isinstance(x,bytes) else x.encode()).hexdigest()
bead='https://example.test/bead';link='https://example.test/link'
malformed={'bdp-minimum-scope-position':'not a position!','bdp-scope-epoch':'','bdp-authorization-view':'view-one, view-two'}
foreign={'bdp-minimum-scope-position':'pos-999','bdp-scope-epoch':'foreign-epoch','bdp-authorization-view':'foreign-view'}
minimum=lambda pos='@actual-position':{'bdp-minimum-scope-position':pos,'bdp-scope-epoch':'@observed-epoch','bdp-authorization-view':'@observed-view'}
create=lambda n='a':{'operation':'createBead','id':f'beads/{n}','type':bead,'properties':{'title':n}}
single=lambda n='a':{k:v for k,v in create(n).items() if k!='operation'}
batch=lambda rows=None:{'operations':rows if rows is not None else [create()]}
targets={'createBead':'create-bead','updateBead':'update-bead','deleteBead':'delete-bead','createLink':'create-link','updateLink':'update-link','deleteLink':'delete-link','putAlias':'put-alias','deleteAlias':'delete-alias','updateWhere':'update-where','deleteWhere':'delete-where','batch':'batch','sequence':'sequence'}
rows=[]
def add(group,method,route,body=None,key=None,headers=None,status=200,note=None):
 text=None if body is None else body if isinstance(body,str) else J(body)
 rows.append({'group':group,'step':1+sum(r['group']==group for r in rows),'method':method,'scopeRelativeRoute':route,'bodyText':text,'bodyBytes':len(text.encode()) if text is not None else 0,'bodySha256':H(text) if text is not None else None,'key':key,'copiedOrSpecialHeaders':headers or {},'expectedStatus':status,'note':note})
def post(g,target,body,key=None,headers=None,status=200,note=None):add(g,'POST','operations/'+targets.get(target,target),body,key,headers,status,note)
for i,field in enumerate(malformed):
 for carrier,value in [('batch',''),('createBead','not a token!'),('sequence','token-one, token-two')]:
  key=f'{carrier}_{i}';body=single(key) if carrier=='createBead' else batch([{**create(key),**({'idempotencyKey':key} if carrier=='sequence' else {})}]);k=None if carrier=='sequence' else key
  post('A',carrier,body,k,{field:value});add('A','GET',f'beads/{key}');post('A',carrier,body,k,foreign)
victim={'operation':'createLink','id':'links/victim','type':link,'source':'beads/a','target':'beads/b'}
post('B','batch',batch([create('a'),create('b'),victim]),'all-seed',malformed)
bops=[('createBead',single('c')),('updateBead',{'bead':'beads/c','propertiesChange':[{'op':'replace','path':'/title','value':'c2'}]}),('createLink',{'id':'links/l','type':link,'source':'beads/a','target':'beads/b','properties':{'title':'l'}}),('updateLink',{'link':'links/l','propertiesChange':[{'op':'replace','path':'/title','value':'l2'}]}),('putAlias',{'alias':'alias/work','target':'beads/c'}),('updateWhere',{'collection':'beads','selector':'$[?@.id]','change':[{'op':'add','path':'/marked','value':True}]}),('deleteAlias',{'alias':'alias/work'}),('deleteLink',{'link':'links/l'}),('batch',batch([create('d')])),('sequence',batch([{**create('e'),'idempotencyKey':'all-sequence'}])),('deleteBead',{'bead':'beads/d'}),('deleteWhere',{'collection':'links','selector':'$[?@.id]'})]
for i,(t,b) in enumerate(bops):post('B',t,b,None if t=='sequence' else f'all-{t}',minimum('pos-999') if i%2==0 else foreign)
for r in ['beads/?limit=32','links/?limit=32','alias/work']:add('B','GET',r,status=404 if r=='alias/work' else 200)
post('C','batch',batch(),'guard-seed',malformed)
for key,status in [('bad key',400),('good-key',200)]:post('C','createBead',single('key-corrected'),key,malformed,status)
stale={'bead':'beads/a','expectedRevision':'wrong','propertiesChange':[{'op':'replace','path':'/title','value':'changed'}]}
post('C','updateBead',stale,'guard-failed',malformed);add('C','GET','beads/a');post('C','updateBead',stale,'guard-failed',malformed)
post('C','updateBead',{**stale,'expectedRevision':'00000000-0000-4000-8000-000000000000'},'guard-valid',malformed,note='Body hash is a template, not observed bytes. Actual revision is randomUUID() from first Resource; exactly36 ASCII bytes, so byte length is exact.')
add('C','GET','beads/a');post('C','createBead',single('negotiated'),'negotiate',{**malformed,'accept':'text/plain'},406);post('C','createBead',single('negotiated'),'negotiate',malformed);add('C','GET','beads/negotiated')
seq=batch([{**create('sequence-corrected'),'idempotencyKey':'sequence-member'}]);post('C','sequence',seq,'forbidden-header',malformed,400);post('C','sequence',seq,None,malformed)
for carrier in ['batch','createBead','sequence']:
 for status in [401,403]:
  key=f'{carrier}_{status}';body=single(key) if carrier=='createBead' else batch([{**create(key),**({'idempotencyKey':key} if carrier=='sequence' else {})}])
  post('D',carrier,body,None if carrier=='sequence' else key,{'bdp-minimum-scope-position':'pos-999','authorization':'@absent' if status==401 else '@reader-credential'},status);add('D','GET',f'beads/{key}',status=404);post('D','batch',batch([create(key+'_replacement')]),key)
post('E','batch',batch(),'receipt-minimum',malformed)
for h,status in [(minimum('pos-99'),503),({**minimum(),'bdp-authorization-view':'foreign'},409),({**minimum(),'bdp-scope-epoch':'foreign'},410),(minimum(),200)]:add('E','GET','@observed-exact-receipt-root',headers=h,status=status)
add('E','HEAD','@observed-exact-receipt-root',headers=minimum('pos-99'),status=503)
for carrier,name,copied in [('create-bead','copied-singleton',foreign),('sequence','copied-sequence',{'bdp-minimum-scope-position':'not a token!'})]:
 inp={'id':'beads/'+name,'type':'https://reference.invalid/acme/types/task','properties':{'title':name}};inp=inp if carrier!='sequence' else {'operations':[{'operation':'createBead',**inp,'idempotencyKey':name}]}
 note='Exact canonical Scope from referenceConfiguration default https://reference.invalid/acme/; listener port only selects transport endpoint. Literal bytes/hash, not executed observation.'
 post('RU',carrier,inp,None if carrier=='sequence' else name,copied,note=note);add('RU','GET','beads/'+name);post('RU',carrier,inp,None if carrier=='sequence' else name,{'bdp-minimum-scope-position':'','bdp-scope-epoch':'epoch-one, epoch-two','bdp-authorization-view':'not a view!'},note=note)
# S invokes handle directly; no wire requests or listener. Six bodies preserved independently.
valid=batch([{**create(),'idempotencyKey':'unused'}]);nested='leaf'
for _ in range(40):nested={'nested':nested}
unsupported=batch([{**create(),'properties':{'nested':nested},'idempotencyKey':'unsupported'}])
for i,(b,status) in enumerate([(valid,400),('{',400),(valid,401),(valid,403),(valid,415),(unsupported,500)]):add('S-direct','POST','operations/sequence',b,status=status,note='Direct handler Request, not TCP HTTP; six existing exclusions; header/auth mapping in source.')
# Fresh TX fixed canonical Scope; all eight POST bodies are literal, exact bytes/hash.
scope='https://reference.invalid/acme/';bt=scope+'types/task'
def inp(n):return {'id':'beads/'+n,'type':bt,'properties':{'title':n}}
a=inp('a');b={'operations':[{'operation':'createBead',**inp('b')}]};c={'operations':[{'operation':'createBead',**inp('c'),'idempotencyKey':'minimum-c'}]}
add('TX18','GET','beads/?limit=32');post('TX18','createBead',a,'minimum-a',malformed);add('TX18','GET','beads/a');post('TX18','createBead',a,'minimum-a',{**foreign,'bdp-minimum-scope-position':'pos-99'});post('TX18','batch',b,'minimum-b',{'bdp-scope-epoch':'partial'});add('TX18','GET','beads/b');post('TX18','batch',b,'minimum-b',{'bdp-scope-epoch':''});post('TX18','sequence',c,None,minimum('pos-99'));add('TX18','GET','beads/c');post('TX18','sequence',c,None,malformed);post('TX18','batch',{'operations':[{'operation':'createBead',**inp('conflict-extra')}]},'minimum-b',{**foreign,'bdp-minimum-scope-position':'pos-99'},409);post('TX18','updateBead',{'bead':'beads/a','expectedRevision':'wrong','propertiesChange':[{'op':'replace','path':'/title','value':'not-admitted'}]},'minimum-stale',{**foreign,'bdp-minimum-scope-position':'pos-99'});add('TX18','GET','beads/a')
for h,status in [(minimum('pos-99'),503),({**minimum('pos-3'),'bdp-authorization-view':'foreign'},409),(minimum('pos-3'),200)]:add('TX18','GET','@observed-exact-receipt-root',headers=h,status=status)
add('TX18','HEAD','@observed-exact-receipt-root',headers=minimum('pos-99'),status=503);add('TX18','GET','beads/?limit=32')
counts=collections.Counter(r['group'] for r in rows);assert dict(counts)=={'A':27,'B':16,'C':13,'D':18,'E':6,'RU':6,'S-direct':6,'TX18':18}
# All RU body literals use the fixed canonical Scope, independent of listener port.
for r in rows:
 r['bodyShaKind']='revision-template-only' if r.get('note','') and 'randomUUID' in r['note'] else 'literal' if r['bodyText'] else 'absent'
stats={g:{'requests':counts[g],'methods':dict(collections.Counter(r['method'] for r in rows if r['group']==g)),'bodyBytesTotal':sum(r['bodyBytes'] for r in rows if r['group']==g),'largestBodyBytes':max(r.get('bodyBytesMax',r['bodyBytes']) for r in rows if r['group']==g),'routes':sorted(set(r['scopeRelativeRoute'] for r in rows if r['group']==g))} for g in counts}
for g in stats:stats[g]['routeTemplateSha256']=H(J(stats[g]['routes']))
assert len(stats['TX18']['routes'])==9
out={'status':'prospective static arithmetic; no test, build, service or package execution','serialization':'UTF-8 compact JSON; insertion order mirrors source. Variable receipt/context values are explicitly symbolic. No credential bytes.','sourceBindingSha256':H((E/'source-binding.json').read_bytes()),'requests':rows,'groups':stats,'controls':[{'id':'restore-command-refusal','selected':'A','expectedHttp':1,'maximumHttp':27,'observedRequired':'500 with developmentLimit minimum-position/contentLength0 at copied command accepted before effects','bodySource':'A step1'},{'id':'bypass-receipt-root-minimum','selected':'E','expectedHttp':2,'maximumHttp':6,'observedRequired':'first future GET incorrectly200 with original full receipt/current context at receipt minimum remains enforced','bodySource':'E steps1–2'}]}
(E/'requests.json').write_text(json.dumps(out,indent=2)+'\n')
# Derivative patch proposals only; never apply or run.
p='packages/server/src/transaction-http.ts';old=subprocess.check_output(['git','show','HEAD:'+p],cwd=W,text=True);new=(W/p).read_text();delta=old.replace(new,'') if new in old else None
begin=old.index('      // Unsupported controls cannot be silently ignored');end=old.index('      if (operation === "sequence")',begin);block=old[begin:end];assert len(block.splitlines())==9
mutant=new.replace('      if (operation === "sequence")',block+'      if (operation === "sequence")',1)
assert mutant==old
receipt='      const result = hasMinimumControls(request)\n        ? await service.receiptAtMinimum(access, url.href, minimum ?? "unsupported")\n        : service.receipt(access, url.href);'
assert new.count(receipt)==1
for name,text in [('restore-command-refusal',mutant),('bypass-receipt-root-minimum',new.replace(receipt,'      const result = service.receipt(access, url.href);'))]:
 patch=''.join(difflib.unified_diff(new.splitlines(True),text.splitlines(True),fromfile='a/'+p,tofile='b/'+p));(E/(name+'.patch')).write_text(patch)
print(json.dumps(stats,indent=2))
