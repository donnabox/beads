Trish — reconciliation candidate d7197e97 remains unchanged. The full suite ended with 9,374 passed, one failed and one skipped. The required CLI matrix failed during initialization; a controlled probe reproduced ancestor configuration discovery when the temporary workspace lived beneath /Users/dbox, and succeeded with the same executable/settings under /private/tmp. A separate validation harness correction is under native and Claude review. No product source change or main landing; no decision needed.

[Failed-run audit](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/pr150-frozen-main-reconciliation/full-suite-failure-audit.md), [diagnosis](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/pr150-frozen-main-reconciliation/bd-init-diagnosis/root-diagnosis.json). Exact raw failed log is preserved as a gzip/base64 archive with [transport metadata](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/pr150-frozen-main-reconciliation/bd-init-diagnosis/raw-log-transport.json); Git normalized the readable text copy, so its bytes are explicitly distinguished.

```handoff
from: vickie
to: trish
repo: gastownhall/bdp
did: Preserved failed candidate validation and isolated its temporary-directory cause; review underway for unchanged-source rerun.
next: Keep d719 reconciliation pending validation in #59; no change to frozen inputs or requested integration scope.
need: none
```
