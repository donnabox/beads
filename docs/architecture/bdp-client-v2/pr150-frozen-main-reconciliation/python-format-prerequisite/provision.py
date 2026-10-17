from pathlib import Path
import hashlib,json,os,subprocess,sys,urllib.request
P=Path(__file__).parent; I=P.parent; A=P/'provision'; target=I/'tools/python-format-venv'
assert sys.argv[1:]==['--released-child'] and A.is_dir()
assert not target.exists() and not target.is_symlink(), 'Preserve pre-existing or partial environment; no retry'
pins=json.loads((P/'wheel-pins.json').read_text()); wheels=A/'wheels';wheels.mkdir()
for pin in pins:
    assert pin['url'].startswith('https://files.pythonhosted.org/')
    assert Path(pin['filename']).name==pin['filename'] and pin['filename'].endswith('.whl')
    path=wheels/pin['filename']
    with urllib.request.urlopen(pin['url'],timeout=20) as response,path.open('xb') as output:
        while chunk:=response.read(1024*1024): output.write(chunk)
    actual=hashlib.sha256(path.read_bytes()).hexdigest()
    assert actual==pin['sha256'], f"Wheel hash mismatch: {pin['name']}"
    print(json.dumps({'wheel':pin['filename'],'sha256':actual,'verified':True}),flush=True)
subprocess.run([sys.executable,'-I','-B','-m','venv',str(target)],check=True)
python=target/'bin/python'
bootstrap=subprocess.check_output([str(python),'-I','-B','-c',"import importlib.metadata,json;print(json.dumps({'bootstrapPip':importlib.metadata.version('pip')}))"],text=True)
(A/'bootstrap.json').write_text(bootstrap);print(bootstrap.strip(),flush=True)
command=[str(python),'-I','-B','-m','pip','--isolated','install','--no-index','--find-links',str(wheels),'--require-hashes','--only-binary=:all:','--no-compile','--disable-pip-version-check','--report',str(A/'pip-install-report.json'),'-r',str(P/'requirements.txt')]
(A/'install-command.json').write_text(json.dumps(command,indent=2)+'\n')
subprocess.run(command,check=True)
code="import importlib.metadata as m,json,re;print(json.dumps(sorted((re.sub(r'[-_.]+','-',d.metadata['Name']).lower(),d.version) for d in m.distributions())))"
actual=json.loads(subprocess.check_output([str(python),'-I','-B','-c',code],text=True))
expected=sorted([[x['name'].lower().replace('_','-'),x['version']] for x in pins])
assert actual==expected, {'actual':actual,'expected':expected}
report={'literalInterpreter':str(python),'realInterpreter':str(python.resolve()),'installed':actual,'expected':expected,'exactSet':True,'bootstrap':json.loads(bootstrap),'finalPip':'26.1.1','wheelSha256Verified':{x['filename']:x['sha256'] for x in pins}}
(A/'installed.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report),flush=True)
