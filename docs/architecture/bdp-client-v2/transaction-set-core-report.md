# Private staged set core checkpoint

Commit `ca603d8ae7afb2362a0c0b8390bff30159ba04cd`, direct child of root integration
`bf462023d9eeea183766ddb01c388d6e9b3c0e96`. Clean local branch
`codex/vickie-transaction-set-core` in
`/Users/dbox/repos/wt/vickie/vickie-transaction-set-core`; GitHub origin configured.
No outward writes. Frozen paging source0ff remains untouched.

**This is a bounded core checkpoint, not an integration-ready or served-set claim.**
Root's separately assigned hot-journal/receipt-ordering repair must be ordinary-
merged and reverified before final integration. Store edits paused immediately on
that handoff: only approved format7 string and staged enumeration were already
present. No independent preflight/service/http/runtime repair was attempted.

## Implemented

- Source-private `prepareTransactionSetCore`: valid bounded set inputs plus ordinary
  Resource members, retaining existing branded RU members and whole Resource label
  preflight. No existing transport calls it. HTTP and normal batch preparation still
  refuse sets; sequence vocabulary stays unchanged.
- Original lossless literals, canonical symbolic creator indices, empty-provenance
  normalization, exact selector strings and singleton/one-op-batch identity survive
  admission, pending replay and restart. Cardinality integer shape checks use exact
  decimal structure, not forced zero or Number.MAX_SAFE_INTEGER. Common patch and
  provenance syntax uses an existing RU syntax-only projection; its placeholder is
  never admitted, resolved or persisted. Negative Selector grammar/complexity stays
  an explicit development gap rather than a wire ruling.
- Lifetime-checked `resources(kind)` facade enumerates the active owner transaction's
  staged rows and freezes them. The existing Selector evaluates canonical projected
  records once at that operation turn. Canonical ID sort fixes expansion order.
- Ordinary and expanded Resource writes share one Resource application/fact path,
  under the existing effects savepoint and terminal receipt commit. Properties-only
  set updates preserve metadata. Matched and per-Resource entries keep outer indices;
  no-op versions/provenance, deletion incidence, Link/self-Link endpoint facts,
  intermediate receipt postimages and final group postimages retain their laws.
- Inclusive cardinality produces the retained failed-receipt conflict. Whole-
  transaction finite Event bound is4096, counting all facts, including ordinary
  members. Evaluator-only lower test bound cannot raise it. Limit failure has the
  existing receipt limit-exceeded framing without an invented operationIndex.
- Approved private format `bdp-private-transaction-set-core-7` protects the expanded
  durable witness vocabulary. Identity codec/policy unchanged. No migration/reseed.
- No shared evaluator, protocol parser/export, schema, normative spec, HTTP route,
  installed configuration, Types, History, profile advertisement or evidence change.

Files: transaction-batch/store/set, new transaction-set.test, private emitted-file
smoke inventory and public-import negatives, transaction implementation progress.

## Verification

Pinned Node24.16.0 / pnpm11.20.0, frozen install; build before tests/process probes.

- PASS final focused regressions: **537 tests /13files**, all transaction test files
  plus Read Selector and Resource evaluator (`regressions-final-core.log`). Includes
  **19new core tests**; no skip. Source tests and existing process tests completed.
- PASS build (`build-final-core.log`), typecheck (`typecheck-final-core.log`),
  boundaries11packages (`check-boundaries-final-core.log`), format
  (`format-check-final-core.log`), lint (`lint-final-core.log`). Lint has only the
  pre-existing3warnings/1info in untouched files; no new warnings.
- PASS installed package smoke (`smoke-core.log`) including new module deep-import
  denial and absence of core helpers from public root; development smoke
  (`smoke-development-core.log`), unchanged89-row Read evidence (`evidence-core.log`).
- PASS actual old-writer probe (`actual-format6-probe.mjs/.log`): imported emitted
  frozen0ff owner creates a real format6 file with pending receipt, closes it, then
  current format7 refuses with database bytes unchanged. This is not an edited
  marker stand-in. Repeat after the root recovery repair because preflight changes.
- No interrupted or terminated runs. No worker test process remains running.
- Full workspace suite was not run by this worker; root owns integrated validation.

The19tests exercise staged selection, stable selected population, canonical result/
Event ordering, set-before-creator label positions and name-insensitive retry,
properties-only metadata preservation, no-op/provenance, cardinality full rollback,
late patch application failure, Link update/deletion/self-Link facts and incidence,
original selection across pending/replay/population change/restart, real numeric
admission/diagnostic paths, durable pages, facade lifetime and frozen rows, total
Event cap, deliberate serving closure, after-effects/after-terminal transient failure,
uncertain COMMIT/sticky fencing/recovery, and mutation reversal/final group state.
Injected failures are not power-loss/disk tests. Current variable authorization is
not claimed by fixed all-visible tests.

Earlier runs found one TypeScript union-narrowing issue and test-only optional-row
annotations, all corrected. First focused run10pass/3fail used invalid actor strings
and an unsupported JSON Patch test operation in fixtures; fixtures were corrected
to actual actor records and a late remove-on-missing-member failure. Runtime syntax
was not relaxed. Final logs above are the verified runs.

## Remaining and next

1. Ordinary-merge root's frozen paging crash/ordering repair, inspect store overlap,
   repeat affected regression/checks/smokes and the actual prior-format probe.
2. Obtain review of Selector grammar/complexity refusal/admission semantics before
   enabling HTTP sets; the valid core is deliberately closed off from served inputs.
3. Review this shared Resource-path refactor and core expansion independently.
4. Later authorized serving slice wires singleton and mixed-batch sets through the
   existing submission pipeline, then closes real set-produced HTTP/page evidence.
   Installed contracts, owned-source fanout, variable-view/erasure projection,
   remaining consistency semantics, replication and truthful discovery stay open.
