import json,os,pathlib,subprocess,tempfile,time,shutil,sys,hashlib
root=pathlib.Path(__file__).parent
variant=sys.argv[1]
assert variant in ('dense-m32','mixed-m32','noop-m32')
script=root/(variant+'.mjs')
directory=tempfile.mkdtemp(prefix='bdp-materialization-closure-')
env={**os.environ,'BDP_STRESS_DIRECTORY':directory}
started=time.monotonic();peak=0;outcome='completed'
with (root/(variant+'.log')).open('x') as log:
 p=subprocess.Popen(['/Users/dbox/.nvm/versions/node/v24.16.0/bin/node',str(script)],env=env,stdout=log,stderr=subprocess.STDOUT)
 try:
  while p.poll() is None:
   sample=subprocess.run(['ps','-o','rss=','-p',str(p.pid)],capture_output=True,text=True).stdout.strip()
   if sample:peak=max(peak,int(sample))
   if peak>1024*1024:outcome='RSS ceiling exceeded';p.kill();break
   if time.monotonic()-started>120:outcome='120s timeout';p.kill();break
   time.sleep(.1)
  p.wait(timeout=10)
 finally:
  if p.poll() is None:p.kill();p.wait(timeout=10)
  shutil.rmtree(directory)
 result={'variant':variant,'script':str(script),'scriptDigest':hashlib.sha256(script.read_bytes()).hexdigest(),'supervisorDigest':hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest(),'exit':p.returncode,'outcome':outcome,'elapsedSeconds':round(time.monotonic()-started,3),'sampledPeakRSSKiB':peak,'limitRSSKiB':1024*1024,'timeoutSeconds':120,'directoryRemoved':not os.path.exists(directory)}
 (root/(variant+'.summary.json')).write_text(json.dumps(result,indent=2)+'\n');(root/(variant+'.child.exit')).write_text(str(p.returncode)+'\n')
 code=0 if p.returncode==0 and outcome=='completed' else 1
 (root/(variant+'.supervisor.exit')).write_text(str(code)+'\n');print(json.dumps(result));sys.exit(code)
