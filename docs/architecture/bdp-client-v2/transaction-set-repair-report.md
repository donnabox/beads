# Private set-core review repair

Frozen clean local commit **a94ab9885762385cc823aecd7071bda14217d4b2**,
direct child of combined **3e2b8784daefd5893f990785cfe91f1d872a2ffe**.
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-set-repair`, branch
`codex/vickie-transaction-set-repair`; origin `https://github.com/gastownhall/bdp.git`.
No push, PR, bus, launch, merge or public qualification action.

Four changed files only:

- `packages/server/src/transaction-set.ts`
- `packages/server/src/transaction-batch.ts`
- `packages/server/src/transaction-set.test.ts`
- `docs/design/transaction-implementation-progress.md`

The concurrent recovery repair's store/HTTP/recovery-test files are untouched, as are
owner format, compilers, public schema, installed config and profile admission.

## External findings addressed

**F1 — deleteWhere provenance.** Preparation now reuses the existing deleteLink
syntax root for deletion, because that existing root accepts precisely the shared
revisionProvenanceInput that both set collections admit. The synthetic Link ID is
only syntax input; it never reaches semantic identity, storage, evaluation or a returned
diagnostic. Updates select the matching Bead/Link syntax root. There is no second schema
parser or forged prepared brand. The canonical bundle corpus now covers both operations
× both collections × nine cardinality variants × twelve provenance variants (432
combinations, including valid empty/null-field and invalid input shapes). Four actual
owner receipt cases cover update/delete of Beads/Links with valid provenance. Deletion
keeps the final live revision in its deleted identity, mints no subject version and
preserves surviving endpoint state. Bead deletion's supplied provenance is inert.

**F2 — no-op result expansion.** A second fixed private constant bounds matched
Resources at4096 across all set operations in one transaction. Repeated matches and
no-ops count again for each set expansion. The cap is checked before adding that set's
matched/result entries; crossing it rolls back staged effects and returns the retained
transaction-wide413 `limit-exceeded` with `transaction.matchedResources`, no fabricated
operationIndex or partial result prefix. The existing Event cap remains separate.
An evaluator-only option may lower, never raise, the bound; tests reject invalid/raised
values. No owner provisioning or persisted installation identity changes. Existing
retained dispositions remain replayable under unchanged identity rules.

**Limit:** enumeration, canonical-record construction, visibility traversal and Selector
selection still materialize/scan the full candidate collection before this check. This
bounds expanded result entries, not total scan work, memory, or receipt bytes. The note
now states that limitation explicitly. No public numeric maximum is advertised.

**F3 — visible candidate projection.** The shared exported isResourceVisible predicate
replaces bare canRead, so Link visibility includes local endpoints before Selector
matching/cardinality. Two actual owner/evaluator tests narrow the per-turn policy to
hide one endpoint while both Links remain individually readable. The hidden-endpoint
Link contributes neither a match nor mutation nor disclosed result; the visible Link
updates/deletes normally and its match fits cardinality/cap1. The owner installation
itself remains all-visible; these are evaluator probes, not installed variable-policy
qualification. No store/policy mechanism was added.

**F4 — failure type.** Operation-local failure problems require operationIndex. Only
currently implemented whole-transaction limits (inducedEvents/matchedResources) may
omit it, and those branches forbid operationName as well. Compiler-checked type probes
reject unindexed validation failures and arbitrary selector limits while accepting the
implemented whole-transaction match limit. No wider receipt schema change.

**F5 — module layout.** The module comment precedes the import block; declarations now
follow imports, rather than stranding the module comment between declarations/imports.

## Verification

Pinned Node24.16.0 / pnpm11.20.0; frozen install and build before emitted/process tests.
All evidence in this report directory.

- Regression red before runtime repair: **6 failed /24 passed** (`red-corrected.log/.json`):
  delete provenance corpus plus both real deletes, missing match bound, both visibility
  cases. An initial run had8 failures because two newly authored update expectations
  mistakenly expected input-string message rather than the existing `{state,value}`
  output form. That test-oracle mistake was corrected before recording the six intended
  red failures; no runtime change was made for those two extra failures (`red.log`).
- Focused green after fixes:30/30; final suite includes six additional bound/type tests.
- PASS **558 tests /14 files**, no skips, 6.03s (`regressions.log/.json`): every transaction
  test file plus Read Selector and Resource evaluator, including36 set tests and both
  actual hot-journal process-death recovery cases.
- PASS typecheck (including the negative type constraints), build, dependency boundaries,
  changed-file format/lint, diff check.
- PASS installed-package smoke (private export denial and public write-profile startup
  remain fail-closed), development smoke (`claimEligible:false`), unchanged89-row Read
  evidence verification. No sealed artifact regenerated.

### Additional real emitted-owner probes

`emitted-probe.mjs/.log`, disposable SQLite store, no lower cap or evaluator mock:

1.128 Beads ×32 no-op sets =4096 matches completes with4128 result entries, no group
  or effectPosition, unchanged graph. This checks the actual default boundary.
2.Adding a129th matching Bead makes the same32 sets total4128 matches; execution fails
  atomically with the exact match-limit problem, no partial results and unchanged graph.
  Same-key failure replays after durable reopen.
3.The actual format6 emitted opener from frozen
  `0ff180a531b3c9ae594e0a104321e80744de38ed` refuses this format7 database containing both
  terminal and pending set witnesses. Database bytes are identical after refusal. This
  closes the reverse-direction observation missing from the earlier new-opens-old probe.
  No frozen prior source or build output was changed.

No interrupted runs or killed test processes, hook bypasses or test-owned processes
remain. Disposable probe stores were removed. Full workspace tests were not run; the
focused transaction/evaluator regressions and two existing smokes completed.

## Remaining scope

HTTP set preparation still refuses sets pending the separate product decision on
negative Selector admission. This is private valid-set core repair, not public profile
completion. Installed owned/schema/policy integration, owned-source fanout, nonwritable
selected matches, whole-collection work bounds, replication and other Transactional
serving/qualification gaps remain. Root owns independent review/integration/publication.
