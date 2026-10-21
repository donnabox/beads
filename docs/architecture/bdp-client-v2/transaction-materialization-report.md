# Measured transaction materialization limits — frozen implementation

Head **076201849f98eeab71bad5c8404fb693a36c6474**
Parent **c0d67388cd849410b38ec7ecb3a328771431dbfa**
Branch `codex/vickie-transaction-materialization`
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-materialization`
Origin is GitHub. Clean frozen working tree; local commit only, no outward writes.

## Result and scope

Final private tuple is **M64/E256**, with32 carrier operations,65536 request/property bytes,262144 complete Resource bytes and65536 encoded Scope Event-source bytes. Root authorized lowering M after the measured M256 case exceeded the pre-agreed memory ceiling. M256 is rejected for this fixture; it was not shipped or advertised. Scope/tuple checks and actual TX evaluator use are enforced, while generic RU optionality stays unchanged.

Six files:

- `packages/server/src/transaction-batch.ts`: one immutable tuple, derived operation cap, TX-only required representation/properties limit subtype, runtime refusal of missing/nonfinite/nonpositive/fractional/raised bounds, actual64/256 count defaults. No new public option/export path.
- `packages/server/src/transaction-store.ts`: encoded Scope check before provisioning and exact tuple passed to evaluator. Existing runtimeLimits configuration binding now captures every bound; compatibility refusal remains after necessary SQLite recovery and before BDP maintenance.
- `packages/server/src/transaction-materialization.test.ts`:18 tests for actual tuple passage/identity, required-limit compiler negatives and runtime integrity refusals with original-key reuse, legacy/changed tuple refusal preserving pending/retained rows, exact Scope boundary/refusal without provisioning, real262144-byte direct/source overflow with unchanged state and retained failure.
- `packages/server/src/transaction-set.test.ts`:64/65 real cumulative no-op match boundary; retained equality/conflict, fresh-key success after shrinking and reopen; raised-bound probes updated honestly.
- `packages/server/src/transaction-owned.test.ts`:256/257 Event boundary with64 selected owned self-Links plus a preceding one-Event update, proving rollback; exact closed owned-update data shape, including combined patches and absence of preimage bodies. Old4096 evidence remains historical, not relabeled.
- `docs/design/transaction-implementation-progress.md`: current tuple/proof/measurements and explicit historical labeling of earlier4096 observations.

No Event serving/accessor/routing, schemas, public admission, RU store/evaluator implementation or unrelated worktree changed. No physical format migration or output-size Problem vocabulary was added. Existing same-key semantic equality and retained dispositions remain unchanged; old installation files require explicit new provisioning and are not rewritten.

## Proof and measured selection

Accepted plan351fa00 is preserved unchanged in `../transaction-materialization-plan.md`; its recommendation was conditional on stress. Current code uses the authorized lower M64 after measurements. Its same patch-aware proof gives:

```
R=262144; Q=S=65536; N=7Q; A=8192; H=1048576; K=M+32
Group <= min(E*(2.5R+N+S)+K*A+H,
             K*(max(9R+4S,5R+2N+2S)+A)+H)
Receipt <= K*(2R+Q+A)+32*128+S+H
```

| Tuple candidate | Group byte upper bound | Receipt byte upper bound |
|---|---:|---:|
| M256/E256 |305397760|173346816|
| M128/E256 |304349184|96800768|
| **M64/E256** |**253493248**|**58527744**|

N bounds the combined normalized property/metadata patch, duplicated across two owned Events; it does not give each list N separately. Source and Link provenance charge separate R bodies. At most2 Resource effects/4 Events, no cascade/preimage body, actual R enforcement and new-store/exact-tuple reopen are load-bearing. Both current JSON bounds are below actual Node24.16 V8 MAX_STRING_LENGTH536870888 and SQLite MAX_LENGTH1000000000. This is not a bound on all SQLite row fields or heap use.

Before stress, abort ceilings were fixed at120 seconds and1GiB RSS per child. `stress-supervisor.py` samples the exact child via ps every100ms and always kills/cleans only its own process/store. `stress.mjs` uses actual emitted owner operations, normal <=65536-byte requests, point reads, receipt/pages, exact serialized digest comparison, and durable reopen. It never calls inspect(). Read-only SQL after owner close transfers only length(CAST(body AS BLOB)) for the exact effectPosition; no group body/history is brought into the harness and no DB writes are used. Phase logs separate owner work from that scalar query. Peak measurements include in-process receipt serialization/hash comparisons and page validation, not a claim to isolate runtime-only allocations.

| Workload | Receipt bytes | Group bytes | Child high-water KiB | Outcome |
|---|---:|---:|---:|---|
| M256 candidate:8 exact-R Beads ×32 no-op sets |67124671|0|execution743840; supervisor sampled1154256 on replay|**Aborted**,1.511s; replay inconclusive |
| 128 no-op matches under initial M256 code |33563367|0|802800|Passed,1.104s; comparison workload, not a separately installed M128 qualification |
|128 owned updates/256 Events under initial M256 code;128 exact-R sources, transient numeric arrays|66482|90118209|975216|Passed,7.975s; insufficient comfortable headroom |
|Final M64 tuple:2 exact-R Beads ×32 no-op sets|16782715|0|480000|Passed,0.670s |
|Final M64 tuple:64 owned updates/128 Events with64 exact-R source Beads|33409|45059033|557584|Passed,3.633s |

Final dense case uses a50000-byte raw transient numeric array that expands in each owned patch Event, removed before final properties validation. Largest actual requests were60157/60158 bytes. It measures128 Events; the permanent distinct count test supplies256-Event coverage using64 four-Event deletes. Results report actual counts rather than claiming one workload covers every maximum simultaneously. All stress temp directories were removed. No forced GC or arbitrary huge proof-bound allocation was used.

## Verification

Pinned Node24.16.0/pnpm11.20.0; logs in this directory, with explicit exit files.

- Frozen install, build (before emitted process tests and stress), typecheck, dependency boundaries, format and lint: PASS.
- Initial count red: **4 failed/8 passed/56 intentionally filtered skips** against old counts; both boundary refusals and raised-cap guards discriminated.
- Focused candidate:85 passed, then86 passed after adding actual default-R subject/source overflow test.
- Broader transaction/evaluator regression: **561 passed,16 files,0 failures**,7.43s (`regressions.log/.json`). Includes emitted runtime/page restart, hot-journal recovery, store/receipt/dependency contracts and owned/set behavior.
- Final stronger closed Event data assertions and match retained/fresh witnesses: **68 passed,2 files**,1.19s (`last-assertions.log`). This was the only source-test assertion change after the broad run.
- Packed executable smoke and development smoke: PASS, including unchanged public admission fail-closed behavior and installed nonattesting payload observation. No evidence reseal or conformance claim.
- `git diff --check`: PASS. No test-owned processes remain; all execution sessions completed.

Initial typechecks caught missing test fixture create flags, a derived literal's default-parameter inference, and an optional array element in a new test descriptor; these were fixed and failed logs preserved. No test failure is omitted from the evidence inventory.

## Actual predecessor fixture

`legacy-owner-probe.mjs` opens the frozen clean **01e2d5470941a49240fab7820a6dea1b6deb291e** emitted owner, creates a completed receipt plus a pending key, then closes it. The new owner with a post-expiry clock refuses the old installation; raw DB bytes remain exactly unchanged. The actual old owner reopens, retracts its pending claim and reproduces the exact serialized completed receipt. The probe records old runtimeLimits and hashes of both emitted owner/batch files. It does not rebuild or edit the frozen predecessor. Exit0.

An initial probe compared object prototypes as well as JSON and failed on null-prototype admitted properties versus parsed persisted objects; that log is retained. The corrected witness compares complete JSON serialization, which is the intended receipt equality scope, not object-prototype identity.

## Remaining boundaries

The finite stress runs are not a universal memory, concurrency, HTTP transport or population qualification. Whole candidate collections are still parsed/materialized before guards; C*R memory, visibility traversal, dense-object overhead, accumulated receipt history and multi-group reads remain independent concerns. Raw principal/other SQLite-row fields are not covered by the receipt/group JSON-string proof. No guarantee for corrupt externally edited stores or old-budget files is implied. Full repository suite and Read evidence regeneration were not run in this bounded slice; root owns independent council and integration.
