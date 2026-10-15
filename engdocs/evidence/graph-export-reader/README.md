# Internal graph export reader evidence

This independently implements the storage acquisition slice of the
[export milestone](../../design/graph-export-milestone.md), supporting
[donnabox/agent-coordination#10](https://github.com/donnabox/agent-coordination/issues/10).
It starts from qualified integration76fde9c9 and has no PR86 ancestry. It does
not add a public format, serializer, CLI, BDP profile or migration operation.

The [frozen manifest](manifest.json) identifies reader/test/BUILD content given
to the same [council prompt](prompt.txt). The [native review](native.txt) found no
confirmed defects. [Actual Claude implementation review](implementation-claude.txt) completed
with no confirmed reader correctness defect and three Low findings. All three
are addressed: first check the persisted binding, remove an impossible by-value
mutation assertion, and reread detachment on the same Store. The bounded
[repair manifest](repair-manifest.json), [prompt](repair-prompt.txt) and
[native repair review](repair-native.txt) cover those repairs and new cleanup
tests. Actual Claude repair review completed as recorded below. Runtime remains
unqualified.

[Pure evidence](pure.log) records four roots and17subtests:21PASS,0FAIL,0SKIP.
The exact-package runner uses canonical build flags and constructs no backend.
It also compiles the authored cgo-tagged engine tests, without executing them.
The first child run passed in0.877s; parent captured raw evidence in0.848s.
An initial recursive-package command was stopped during unnecessary CLI
prebuild, exit143, before any test or DB workload. An unused import exposed
by compilation was fixed before these passing runs.

Command:

```sh
./scripts/test.sh -v -count=1 -run '^(TestExportCurrentSnapshotContextBeforeOpen|TestExportCommentOriginalUTF8|TestExportSnapshotClosure|TestExportReadProjectionIncludesUnjoinedComments)$' ./internal/storage/graphstore
```

[Lint](lint.log) passed formatting and native/windows/darwin profiles with zero
findings. [Documentation link check](docs.log) passed. Package BUILD source
manifest check and diff whitespace checks passed; local Bazel was not run.

Pending engine proof: lifecycle/reopen/empty/mixed state, original comment
identity/content/timestamp, raw prefix, deleted allocation disclosure, invalid
ownership/alternative-plane refusal, aggregate acquisition limits, unchanged
logical source rows, and forced writer interleaving in embedded and ordinary
SQL server modes. These authored cases are **unrun** until Janet explicitly
hands back the engine slot. No Iris DB/server work was started.

The [repair pure run](repair-pure.log) passed nine roots plus19subtests
(28PASS/0FAIL/0SKIP) in0.913s. It includes actual driver rollback, comment
EOF-close plus rollback, SQL scan/close/rollback, ErrBadConn logical cleanup,
and separate physical Store.Close failure through a test-owned SQL connector.
The [child run](cleanup-child.log) also passed its five roots plus two subtests.
These tests prove database/sql cleanup propagation, not engine acknowledgments.
Late non-EOF iteration error injection remains a nonblocking coverage gap.
The initial [parallel lint collision](repair-lint.log) is preserved; the
[sequential retry](repair-lint-retry.log) passed allprofiles with0findings.
Actual backend cleanup qualification remains required before readiness. Installed binary,
encoded-output budget, permissions and cleanup-before-publication proof belong
to the later independently reviewed exporter/CLI milestone. Current limits
measure persisted acquisition bytes, not exact heap or JSON output size.


## Completed cleanup audit and final repair

The [actual bounded Claude review](repair-claude.txt) completed with exit0 and
found no reader correctness defect. Its Low coverage findings are repaired:
nested Issue/owned-Link caller mutation uses an immutable canonical oracle;
fresh and reopened reads check unchanged source rows; direct comment feeds
exercise valid nonempty rows, UTF-8/owner/overflow/count refusals and mid-feed
Next failure. Physical-close accounting and receiver names are corrected.
The early binding check is deliberate authority-before-content admission;
shared current-read checks retain their other callers' behavior.

One reviewer claim was rejected against the real Go database/sql boundary:
non-EOF Next errors auto-close rows, and simultaneous driver Close errors are
suppressed. The tests preserve exposed Next and rollback causes, nil partial
maps and exactly one driver Close; they do not fabricate an unavailable cause.
The [final pure run](final-pure.log) passed11 roots/40 total cases with0failures
and0skips in1.009s. [Lint](final-lint.log) passed all profiles. The
[manifest](final-manifest.txt), [prompt](final-prompt.txt) and
[native confirmation](final-native.txt) identify the repaired-source successor
council; actual Claude was stopped at Donna's overnight pause (exit143, no
findings). [Attempt record](final-review-interrupted.json) preserves the boundary;
rerun that frozen audit on resume. No completed successor council is claimed. Real mixed snapshot/query/backend cleanup
proof remains unrun until explicit Janet engine handback.
