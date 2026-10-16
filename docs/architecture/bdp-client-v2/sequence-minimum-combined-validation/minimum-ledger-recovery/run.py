"""One fixed same-evidence adjudication; no product/test invocation."""
from pathlib import Path
import hashlib,json,sys
R=Path(__file__).resolve().parent;P=R.parent
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def main():
 i=json.loads((R/'inputs.json').read_text());a=json.loads((R/'root-recovery-release.json').read_text())
 require(a['accepted'] is True and a['head']==i['head'] and a['stage']==i['stage'],'Recovery not released')
 require(a['inputsSha256']==sha(R/'inputs.json') and a['runnerSha256']==sha(Path(__file__)),'Recovery pins changed')
 require(str(Path(sys.executable).resolve())==i['pythonRealpath'] and sha(Path(i['python']))==i['pythonSha256'],'Wrong pinned Python')
 require(sys.version_info[:2]==(3,14),'Python 3.14 required')
 def unchanged():return all(sha(Path(p))==h for p,h in i['files'].items())
 require(unchanged(),'Original inputs changed')
 require(not (P/'run/minimum-installed.adjudication.json').exists(),'Adjudication already exists')
 require(not (R/'run').exists(),'Recovery attempt already exists')
 sys.path.insert(0,str(P));import supervise
 original=supervise.load_gate
 def recovery_load(name):
  require(name==i['stage'],'Only fixed adjudication permitted')
  g,s=original(name)
  require(g['finalHead']==i['head'] and s['argv']==['/usr/bin/python3',str(P/'review-minimum.py')],'Original stage differs')
  s=dict(s);s['argv']=[i['python'],str(P/'review-minimum.py')]
  return g,s
 supervise.load_gate=recovery_load;supervise.E=R/'run';sys.argv=[str(__file__),i['stage']]
 code=supervise.main()
 retained=unchanged();ad=P/'run/minimum-installed.adjudication.json'
 valid=code==0 and retained and ad.exists() and json.loads(ad.read_text()).get('accepted') is True
 result={'accepted':valid,'head':i['head'],'stage':i['stage'],'originalInputsUnchanged':retained,'inputsSha256':sha(R/'inputs.json'),'runnerSha256':sha(Path(__file__)),'rootReleaseSha256':sha(R/'root-recovery-release.json'),'exitSha256':sha(R/'run'/ (i['stage']+'.exit.json')),'adjudicationSha256':sha(ad) if ad.exists() else None,'newHTTP':0,'newProductChildren':0}
 with (R/'verification.json').open('x') as f:f.write(json.dumps(result,indent=2)+'\n')
 require(valid,'Recovery failed; preserve all outputs and stop')
 return 0
if __name__=='__main__':sys.exit(main())
