import json,csv,collections
from pathlib import Path
r=Path(__file__).resolve().parent
sets={
'issue':'''assign blocked children close comment comments cook count create create-form defer delete dep duplicate duplicates edit epic find-duplicates formula gate heartbeat label lint list merge-slot mol note orphans priority promote prune purge q query ready reclaim recompute-blocked rename rename-prefix reopen search set-state ship show stale state status statuses supersede swarm tag todo types unclaim undefer update'''.split(),
'other':'''admin ado backup bootstrap branch compact config conflicts context db-proxy-child diff doctor dolt export federation flatten forget gc github gitlab graph history hooks import info init init-safety jira linear link memories metrics migrate migrate-issues migrate-personal notion onboard preflight prime quickstart recall remember repo restore schema send-metrics serve setup sql sync vc version where worktree codex-hook cursor-hook claude-hook compare links unlink versions'''.split(),
'supporting':'''audit human kv mail ping provenance rules upgrade'''.split(),
'event-log':'''batch events'''.split()}
report={'issue':'issue-cli-audit.md','other':'other-cli-audit.md','supporting':'supporting-cli-audit.md','event-log':'event-log-audit.md'}
read_heavy=set('blocked children count dep duplicates epic find-duplicates gate label lint list mol orphans query ready recompute-blocked search stale state status swarm graph memories recall human'.split())
local=set('audit mail rules upgrade version where worktree hooks setup codex-hook cursor-hook claude-hook metrics send-metrics quickstart onboard prime formula cook'.split())
admin=set('admin backup branch conflicts db-proxy-child dolt federation flatten migrate migrate-issues migrate-personal serve sql sync vc'.split())
mixed=set('bootstrap compact config context diff doctor export gc history import info init init-safety preflight repo restore schema'.split())
external=set('ado github gitlab jira linear notion ship'.split())
allrows={x:json.loads((r/f'cli-ast-{x}.json').read_text()) for x in ['main','preview']}
decs={x:json.loads((r/f'command-declarations-{x}.json').read_text()) for x in allrows}
registries={x:{z['name']:z for z in rows if z['kind']=='registry'} for x,rows in allrows.items()}
paths=sorted(set(registries['main'])|set(registries['preview']))
summary={}
for x,rows in allrows.items():
 summary[x]=dict(collections.Counter(z['kind'] for z in rows));summary[x]['unique_registry_paths']=len(registries[x]);summary[x]['unique_test_files']=len({z['file'] for z in rows if z['kind']=='test'})
summary['preview_only']=sorted(set(registries['preview'])-set(registries['main']))
summary['main_only']=sorted(set(registries['main'])-set(registries['preview']))
summary['scope']='Static cmd/bd top-level Go AST declarations plus capability registry, not executed tests, dynamic subtests, nested package tests, alias expansion or exhaustive flag combinations. Family dispositions are architecture analysis, not implementation parity.'
rows=[]
for p in paths:
 root=p.split()[0];families=[k for k,v in sets.items() if root in v];assert len(families)==1,(p,families)
 snap='main' if p in registries['main'] else 'preview'; registry=registries[snap][p]
 declarations=[d for d in decs[snap] if p in d['paths']]
 if root in local:boundary='client-local'
 elif root in admin:boundary='provider-administration-or-explicit-redesign'
 elif root in mixed:boundary='mixed-client-resource-and-provider-semantics'
 elif root in external:boundary='client-external-integration'
 else:boundary='generic-resource-command'
 if root in read_heavy:cost='query/sort/graph scan or warm local replica; see family report'
 elif root in admin:cost='native engine semantics unavailable through portable BDP; retain side-by-side or redesign'
 elif root in local:cost='no BDP data request normally required'
 else:cost='bounded requests/batch where possible; guards, limits and exceptions in family report'
 flags=sorted({f['name'] for d in declarations for f in d.get('flags',[]) if f.get('name')})
 source=';'.join(f"{d['file']}:{d['line']}" for d in declarations) or f"{registry['file']}:{registry['line']} (registry; factory/constant/alias path)"
 rows.append({'command':p,'main':p in registries['main'],'preview':p in registries['preview'],'boundary':boundary,'analysis':report[families[0]],'cost_or_gap':cost,'source':source,'direct_declared_flags':';'.join(flags),'registry_evidence':f"{registry['file']}:{registry['line']}",'evidence_level':'source/test-contract review; no command execution parity'})
with (r/'command-crosswalk.csv').open('w') as f:
 w=csv.DictWriter(f,fieldnames=list(rows[0]),lineterminator='\n');w.writeheader();w.writerows(rows)
(r/'inventory-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
