from pathlib import Path
L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs');p=L/'live-ledger.md';s=p.read_text();anchor='### Current verified main and active work\n';assert s.count(anchor)==1
new='''### Current verified main and active work

**Current main is PR131 merge18104a03b850553641f24fe83e7318493fa102ca**, tree10810a8016a451e241ea974487681dcddd63c2e4, exact candidate/CI/ordinary ancestry/actual tree verified. Its depth/key ordering repair is delivered; earlier PR128–130 pending wording below is historical. No release/public profile qualification.

**Sequence member evidence is authored and statically validated.** Four-path [source binding](sequence-member-draft/source-binding.json), SHAa6eb425d, and fullpatch53c66bc9 append a ten-action partial witness: success/failure/success, three retained receipt GETs, complete replay, current records and two linked Events. [Author report](sequence-member-draft/source-author/report.md) preserves the initial target correction. [Independent native review](sequence-member-draft/native-code-review.md), SHA6a36c3cb, has no actionable findings and closes the mechanical formatting successor. [Measurement audit](sequence-member-draft/source-author/root-measurement-audit.md) verifies all41 changed specimens and all28 old raw objects; final disk790009/normal802361/nodes20956/coverage117702, derivative maximum808479/21112. Measured source fences retain original margins and runtime limits.

[Root source acceptance](sequence-member-draft/root-source-acceptance.json), SHA15e91b6a, binds all eight sequential stages and eight exact static callbacks passing, unchanged Read89 evidence and38 sampled PIDs independently absent. Seven Node compile-cache directories are classified and retained; no product stores remain. Initial format-check failure is preserved separately; only new-row wrapping and disk-fence/comment changed. No installed HTTP evidence is credited yet. Actual Claude code-review packet is being assembled; launch and installed execution remain separately gated. The closed fresh-package scaffold preserves four cohorts623 expected/632 maximumHTTP and71 children. These are workload bounds, not completion metrics.

[Bounded ordinary Update audit](update-after-pr131-audit.md), SHA235e59a0, found no new serving defect in copied Read headers, retained identity/disclosure ordering or revision guards; existing ordinary Update evidence remains distinct. Snapshot/continuation minimum and incident-Link metadata Event questions remain pending; independent work continues.

Workers: update_implementation finished source and closed installed scaffold; cross_carrier_plan_v2_review finished native code review; cross_carrier_native_review is completing measured formatting closure and text-only Claude assembly. Root owns actual council launch, installed validation, publication and next slice. All work stays in the owned sequence worktree; no Janet/Trish/shared-state changes.

### Preserved earlier evidence and chronology
'''
s=s.replace(anchor,new);p.write_text(s)
p=L/'checkpoint-ledger.py';s=p.read_text();anchor="s=(D/'design.md').read_text();a=s.index('### Live worker ledger\\n');";assert anchor in s
extra="""for name in ['update-after-pr131-audit.md']:
 shutil.copyfile(L/name,D/name)
for stem in ['sequence-member-draft','sequence-member-draft/source-author','sequence-member-draft/format-repair','sequence-member-draft/execution-preparation','sequence-member-draft/execution-preparation/run','sequence-member-draft/execution-preparation-v2','sequence-member-draft/execution-preparation-v2/run','sequence-member-draft/claude-preparation','sequence-member-draft/installed-preparation']:
 for src in (L/stem).glob('*'):
  if src.is_file() and src.suffix in ['.md','.json','.txt','.log','.py','.patch']:
   if src.name in ['manifest-row.json','coverage-row.json']:continue
   if stem.endswith('claude-preparation') and src.name in ['stdout.txt','stderr.txt','exit.json','cleanup.json','source-after.json','head-after.json'] and not (src.parent/'exit.json').exists():continue
   dst=D/stem/src.name;dst.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(src,dst)
"""
s=s.replace(anchor,extra+anchor).replace('docs: checkpoint verified PR131 merge and sequence continuation','docs: checkpoint sequence source validation and council preparation');p.write_text(s)
(L/'bus-current.txt').write_text('''Vickie for Janet: PR131 merged18104a03 and tree verified. Sequence-member ten-action evidence now passes eight static checks; native review finds no remaining issue. Actual Claude code review and fresh installed positive/negative cohorts are next. Bounded ordinary Update audit found no additional serving defect in reviewed areas. No public profile qualification, release or CLI/staging/shared DB changes.

```handoff
from: vickie
to: none
repo: gastownhall/bdp
did: Sequence-member draft statically validated; native review and measured preservation closed.
next: Vickie completes actual Claude review and fresh installed four-cohort validation.
need: none
```
''')
