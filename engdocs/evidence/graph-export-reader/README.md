# Internal graph export reader evidence

This independently implements the storage acquisition slice of the
[export milestone](../../design/graph-export-milestone.md), supporting
[donnabox/agent-coordination#10](https://github.com/donnabox/agent-coordination/issues/10).
It starts from qualified integration76fde9c9 and has no PR86 ancestry. It does
not add a public format, serializer, CLI, BDP profile or migration operation.

The [frozen manifest](manifest.json) identifies reader/test/BUILD content given
to the same [council prompt](prompt.txt). The [native review](native.txt) found no
confirmed defects. The external implementation review is still running; no
completed council or runtime readiness is claimed from this checkpoint.

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

Callback faults do not cover real driver rollback/connection/rows-close faults.
Those cleanup tests remain required before export readiness. Installed binary,
encoded-output budget, permissions and cleanup-before-publication proof belong
to the later independently reviewed exporter/CLI milestone. Current limits
measure persisted acquisition bytes, not exact heap or JSON output size.
