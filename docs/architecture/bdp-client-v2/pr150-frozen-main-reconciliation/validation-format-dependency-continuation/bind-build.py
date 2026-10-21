import sys
sys.dont_write_bytecode=True
from gate import *
g,_=load_gate('bind-build',child=True)
r=complete('build',g)
write('build-artifacts.json',{'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'artifacts':inventory(),'records':{n:{'path':str(E/('build.'+n)),'sha256':sha((E/('build.'+n)).read_bytes())} for n in ['started.json','exit.json','log']}})
