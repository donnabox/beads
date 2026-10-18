"""CLOSED synthetic-only control launcher. Do not evaluate before other-native/root.
Each externally authorized case has one exclusive attempt and a real independent
120s parent clock (TERM115/KILL119). No original product worker or Node is ever invoked. Real zero-product peers
use the original150/145/149 policy; unit cases use120/115/119.
The actual D4 watchdog uses labeled model stimuli or narrow real peer adapters.
This is not the still-unwritten responsive whole-host wrapper; see design.
"""
import time
ORIGIN=time.monotonic()
import argparse,hashlib,importlib.util,json,math,os,signal,subprocess,sys
from pathlib import Path
sys.dont_write_bytecode=True
HERE=Path(__file__).resolve().parent
SESSION='2026-10-18T02:16:24Z-549f'
PYTHON='/opt/homebrew/opt/python@3.14/bin/python3.14'
def need(ok,why):
 if not ok:raise RuntimeError(why)
def sha(data):return hashlib.sha256(data).hexdigest()
def unique(pairs):
 value={}
 for k,v in pairs:need(k not in value,'duplicate-key');value[k]=v
 return value
def read(path):return json.loads(Path(path).read_bytes(),object_pairs_hook=unique,parse_constant=lambda _:(_ for _ in ()).throw(ValueError('nonfinite-json')))
def write(path,value):
 with Path(path).open('x')as output:json.dump(value,output,indent=2,allow_nan=False);output.write('\n')
def document(pin):
 need(type(pin)is dict and set(pin)=={'path','sha256'},'document-shape');path=Path(pin['path']);need(path.is_absolute()and path.is_file()and not path.is_symlink()and sha(path.read_bytes())==pin['sha256'],'document-pin');return read(path)
def authorized(case,release_path,variant):
 need(__debug__,'assertions-required')
 binding=read(HERE/'binding.json')
 need(binding.get('hostOwnerExecutableImplemented')is True,'CLOSED partial preparation: independent responsive host-owner executable is not implemented; immutable successor required')
 index=read(HERE/'implementation-hashes.json')
 need(type(index)is dict and index and all(type(v)is str and len(v)==64 for v in index.values()),'flat-implementation')
 for name,digest in index.items():
  path=HERE/name;need(not path.is_symlink()and path.is_file()and path.resolve().is_relative_to(HERE)and sha(path.read_bytes())==digest,'harness-input-changed:'+name)
 need(read(HERE/'execution-gate.closed.json')==read(HERE/'execution-gate.json'),'original-closed-gate-changed')
 closed=read(HERE/'execution-gate.json');need(closed=={'session':SESSION,'syntheticReleased':False,'productReleased':False,'approvedCases':[],'noRetry':True},'closed-template-changed')
 release=read(release_path)
 need(set(release)=={'session','scope','syntheticOnly','productReleased','harnessImplementationSha256','subjectImplementationSha256','approvedInvocations','nativeHarnessAcceptance','durableRootRelease','noRetry'},'release-keyset')
 need(release['session']==SESSION and release['scope']=='D4 zero-product synthetic controls only'and release['syntheticOnly']is True and release['productReleased']is False and release['noRetry']is True,'synthetic-only-release')
 need(release['harnessImplementationSha256']==sha((HERE/'implementation-hashes.json').read_bytes())and release['subjectImplementationSha256']==binding['subjectImplementation']['sha256'],'release-binding')
 matrix=read(HERE/'control-matrix.json');allowed={c+'--'+v for c,r in matrix['cases'].items()for v in r['variants']}
 need(type(release['approvedInvocations'])is list and len(set(release['approvedInvocations']))==len(release['approvedInvocations'])and all(v in allowed for v in release['approvedInvocations'])and case+'--'+variant in release['approvedInvocations'],'invocation-not-released')
 native=document(release['nativeHarnessAcceptance']);need(native.get('accepted')is True and native.get('hardLimiterStaticAcceptance')is True and native.get('clockDomainVerified')is True and native.get('ownershipBootstrapAccepted')is True and native.get('harnessImplementationSha256')==release['harnessImplementationSha256']and native.get('subjectImplementationSha256')==release['subjectImplementationSha256'],'other-native-harness-acceptance')
 durable=document(release['durableRootRelease']);need(durable.get('accepted')is True and durable.get('syntheticOnly')is True and durable.get('productReleased')is False and durable.get('harnessImplementationSha256')==release['harnessImplementationSha256'],'durable-root-release')
 P=Path(binding['subjectRoot']);need(P.is_absolute(),'subject-root')
 document(binding['subjectImplementation'])
 subject=read(P/'implementation-hashes.json')
 for name,digest in subject.items():need(type(digest)is str and sha((P/name).read_bytes())==digest,'subject-input-changed:'+name)
 # Pins include original CLOSED gates/indices and all dependencies, so a later
 # product gate release invalidates this preparation until root rebinds it.
 for path,digest in binding['externalPins'].items():need(sha(Path(path).read_bytes())==digest,'external-pin-changed:'+path)
 subprocess.run([PYTHON,'-I','-B',binding['currentSessionGuard'],'/usr/bin/true'],check=True,timeout=min(30,max(.01,115-(time.monotonic()-ORIGIN))))
 return binding,P,release

def load(name,path):
 need(name not in sys.modules,'module-already-loaded:'+name)
 spec=importlib.util.spec_from_file_location(name,path);module=importlib.util.module_from_spec(spec);sys.modules[name]=module;spec.loader.exec_module(module);return module

def child(case,release_path,attempt,variant):
 binding,P,_=authorized(case,release_path,variant)
 need(os.environ.get('D4_SYNTHETIC_ATTEMPT')==str(attempt.resolve())and os.environ.get('D4_OUTER_FD')and os.environ.get('D4_NATIVE_LIMITER'),'outer-owner-attempt-required')
 need(attempt.resolve().parent==HERE/'attempts'and read(attempt/'started.json')['case']==case and read(attempt/'started.json')['variant']==variant,'attempt-root')
 # Import only pinned pure Python subjects after all authorization. packages.py
 # has top-level invocation and is NOT imported. schema-audit.mjs is NOT run.
 os.environ['D4_VALIDATION_PHASE']='final'
 sys.path.insert(0,str(HERE));sys.path.insert(0,str(P))
 scanner=load('scanner',P/'scanner.py');gate=load('gate',P/'gate.py');audit=load('audit',P/'audit.py');watchdog=load('d4_immutable_watchdog',P/'run.py')
 need(audit.need is scanner.need,'actual-shared-scanner-import')
 controls=load('scanner_controls',HERE/'scanner_controls.py')
 if variant.startswith('host-'):
  host=load('host_controls',HERE/'host_controls.py');receipt=host.run(case,watchdog,P,attempt,variant.removeprefix('host-'))
 elif case[0]in ['M','J']:receipt=controls.run(case,scanner,attempt)
 elif case[0]=='D':
  reports=load('report_controls',HERE/'report_controls.py');receipt=reports.run(case,audit,gate,P,attempt)
 else:
  model=load('watchdog_controls',HERE/'watchdog_controls.py');receipt=model.run(case,watchdog,audit,gate,P,attempt)
 if variant=='unit' and case in ['M01','P11','D06']:
  gates=load('gate_controls',HERE/'gate_controls.py')
  receipt['originalGateUnit']=gates.consumer_bindings(gate,attempt)if case=='D06'else gates.stage_receipts(gate,P,attempt)
 unresolved=receipt.get('unresolved',[])
 # Extra limitations are explicit blockers, never silently omitted or skipped.
 declared=read(HERE/'control-matrix.json')['cases'][case]
 need(variant in declared['variants'],'variant-not-declared')
 unresolved=list(dict.fromkeys([*unresolved,*declared['remainingUnsupported']]))
 value={'case':case,'variant':variant,'status':'UNRESOLVED'if unresolved else'ACTUAL_HOST_CONTROL_OBSERVED'if variant.startswith('host-')else'UNIT_CONTROL_PASSED','executed':True,'productCalls':0,'scope':declared['scope'],'receipt':receipt,'unresolved':unresolved,'subjectImplementationSha256':binding['subjectImplementation']['sha256'],'hostContainmentCredit':variant.startswith('host-') and not unresolved,'profileCredit':False,'syntheticAcceptance':False}
 write(attempt/'control-result.json',value);return 2 if unresolved else 0

def parent(case,release_path,variant):
 raise RuntimeError('external-native-limiter-and-outer-host-required')

def main():
 parser=argparse.ArgumentParser();parser.add_argument('--case',required=True);parser.add_argument('--variant',default='unit');parser.add_argument('--release',required=True,type=Path);parser.add_argument('--internal-attempt',type=Path);args=parser.parse_args()
 if args.internal_attempt is not None:return child(args.case,args.release,args.internal_attempt,args.variant)
 return parent(args.case,args.release,args.variant)
if __name__=='__main__':
 try:code=main()
 except BaseException as exc:
  # Preserve partial control error in own attempt directory; never repair or
  # invent absent subject outputs. External root supervisor remains required.
  target=os.environ.get('D4_SYNTHETIC_ATTEMPT')
  if target:
   path=Path(target)
   if path.parent==HERE/'attempts'and path.is_dir()and not(path/'control-error.json').exists():
    write(path/'control-error.json',{'type':type(exc).__name__,'reason':str(exc)[:256],'syntheticAcceptance':False,'productCalls':0})
  raise
 raise SystemExit(code)
