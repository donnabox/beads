"""Vickie exact-session guard. Does not claim or modify a launch file."""
import base64
import json
import subprocess
import sys

EXPECTED_SESSION = "2026-10-18T02:16:24Z-549f"
if len(sys.argv) < 2:
    raise SystemExit("No guarded command supplied")
response = json.loads(subprocess.check_output(
    ["gh", "api", "repos/donnabox/agent-coordination/contents/launch/vickie.md"],
    timeout=30,
))
if response.get("type") != "file" or response.get("path") != "launch/vickie.md" or response.get("encoding") != "base64":
    raise SystemExit("Unexpected remote launch response; command refused")
launch = base64.b64decode(response["content"], validate=False).decode("utf-8", errors="strict")
sessions = [line[len("session: "):] for line in launch.splitlines() if line.startswith("session: ")]
if sessions != [EXPECTED_SESSION]:
    raise SystemExit("ROLE SESSION SUPERSEDED OR AMBIGUOUS; command refused")
subprocess.run(sys.argv[1:], check=True)
