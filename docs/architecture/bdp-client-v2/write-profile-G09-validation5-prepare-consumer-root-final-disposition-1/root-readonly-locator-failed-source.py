from pathlib import Path
P=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs/write-profile-G09-validation-preparation-5')
for n in ['run/consumer/package.json','run/consumer/pnpm-workspace.yaml']:
 print(n,(P/n).read_text())
t=(P/'packages.py').read_text();a=t.index('def consumer');print(t[a:a+4200])
