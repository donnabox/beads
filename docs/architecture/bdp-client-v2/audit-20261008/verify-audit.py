import json,re,hashlib,subprocess
from pathlib import Path
r=Path(__file__).resolve().parent
baselines=json.loads((r/'baselines.json').read_text())
maps={}
for b in baselines:
 p=r/(b['dir']+'-remote-tree.json')
 if not p.exists():
  p.write_bytes(subprocess.check_output(['gh','api',f"repos/{b['repo']}/git/trees/{b['sha']}?recursive=1"]))
 t=json.loads(p.read_text());assert not t.get('truncated'),b
 maps[b['sha']]={x['path']:x['sha'] for x in t['tree'] if x['type']=='blob'}
checked={};citations=0
for p in r.glob('*.md'):
 for match in re.finditer(r'https://github.com/([^/]+/[^/]+)/blob/([0-9a-f]{40})/([^\s)#]+)(?:#L(\d+)(?:-L(\d+))?)?',p.read_text()):
  repo,sha,path,start,end=match.groups();bs=[b for b in baselines if b['sha']==sha and b['repo']==repo]
  if not bs:continue
  b=bs[0];f=r/b['dir']/path;assert f.is_file(),(p.name,path)
  data=f.read_bytes();digest=hashlib.sha1(f'blob {len(data)}\0'.encode()+data).hexdigest();assert maps[sha][path]==digest,(p.name,path,'remote mismatch')
  lines=len(data.splitlines())
  if start: assert 1<=int(start)<=int(end or start)<=lines,(p.name,path,start,end,lines)
  citations+=1;checked[(repo,sha,path)]=dict(repo=repo,commit=sha,path=path,git_blob_sha=digest,lines=lines)
result={'method':'GitHub immutable recursive Git trees compared to git-blob SHA1 of every cited local source file; all Markdown blob citation line bounds checked. Content provenance, not runtime equivalence.','citations_checked':citations,'unique_cited_source_files':len(checked),'sources':list(checked.values()),'checks':{'existing_tests':219,'existing_test_files':3,'extra_client_probes':12,'extra_journal_probes':8,'transactional_runtime_conformance':False,'go_cli_database_tests_run':False}}
(r/'verification.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='sources'},indent=2))
