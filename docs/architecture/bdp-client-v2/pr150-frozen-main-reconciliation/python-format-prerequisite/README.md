# CLOSED isolated Python format prerequisite

Preparation only. No wheels downloaded, environment created, tool imports executed, or validation stages run. The only network reads in preparation were official PyPI JSON metadata for pip26.1.1, rfc3339-validator0.1.4, and six1.17.0. Saved metadata pins those pure wheels. The other five wheel URLs/SHA256/version identities are copied exactly from the preserved original pip install report.

Root controls `root-release.json`; both attempts are initially closed. Its preparation-index hash must match this packet. Release `provision` alone first. After inspecting accepted exit, installed exact-set result, verified wheel hashes, raw logs, source/old-tool invariance, and environment file/symlink inventory, root may add `probe` and bind the actual provision exit and environment JSON hashes. No fabricated future hashes.

Each command is one exclusive attempt, bounded120s with TERM at115 and owned-group TERM/KILL cleanup. Records are exclusive; any existing attempt directory or existing target environment stops provisioning. All partial wheels, environments, logs and exit records are preserved after failure. No automatic retry, widening, or partial-env reuse. The runner records stdout/stderr, command/start/process samples/exit, source/tool before and after and the full resulting environment file/symlink map. No input files or prior evidence are modified. Tool prerequisites are outside the logical25-stage validation chain and never represent product or example-script passes.

Provisioning downloads all eight exact wheels inside its single bound, verifies SHA256 before creating the environment, creates a new venv from the pinned base Python, records the ensurepip bootstrap pip version, and explicitly installs final pip26.1.1 with the other seven pinned distributions using offline `--no-index --find-links --require-hashes --only-binary=:all:`. No resolver index or optional extras. The exact installed set must equal the eight pins with no duplicates or additions. The old environment is independently inventoried and must remain unchanged. Archive SHA256 values are authoritative pins now; downloaded archive bytes will be checked only during the released attempt.

The separate probe uses the literal new interpreter with `-I -B`; no product imports. It constructs bare `FormatChecker()`, requires date-time registration, checks a valid RFC3339 timestamp and `bad-date` through both direct conforms and Draft202012Validator, and records actual provider path/version and environment identity. It requires a root-accepted provision exit plus exact full environment file/symlink inventory and must leave that inventory unchanged.

Concrete commands (closed until root release):

```
/opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/pr150-frozen-main-reconciliation/python-format-prerequisite/run.py provision
/opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/pr150-frozen-main-reconciliation/python-format-prerequisite/run.py probe
```

Preparation verification was data-only: AST parsing of the three scripts, exact requirement/pin agreement, original report version/URL/hash preservation, clean candidate identity and all701 source bytes. No helper or product execution occurred.
