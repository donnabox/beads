"""Run the frozen existing supervisor against this packet's local closed gate."""
from pathlib import Path
import hashlib,runpy,sys
base=Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-command-minimum-later-preparation-v4/supervise.py")
if hashlib.sha256(base.read_bytes()).hexdigest()!="7f0d08b5a1c819131c2efd7f5c258b565414cbe70c65cd8b8b0ffe9ec3ba0065":
    raise RuntimeError("Reviewed supervisor changed")
sys.path.insert(0,str(Path(__file__).resolve().parent))
runpy.run_path(str(base),run_name="__main__")
