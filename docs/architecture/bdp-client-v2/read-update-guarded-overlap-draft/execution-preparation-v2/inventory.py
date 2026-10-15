from gate import *
g,_=load_gate("build-inventory");completed("build")
names=git("ls-files","--cached","--others","--exclude-standard").decode().splitlines()
x={"sourceSha256":source_hashes(),"buildExitSha256":sha((E/"build.exit.json").read_bytes()),"inputFiles":{p:sha((W/p).read_bytes()) for p in names},"files":build_outputs(),"attribution":"Actual accepted source build; future clean commit requires identical inputs/outputs."}
with (E/"build-output-inventory.json").open("x") as f:f.write(json.dumps(x,indent=2)+"\n")
