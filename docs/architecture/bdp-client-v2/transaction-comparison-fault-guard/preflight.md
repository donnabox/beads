# Maintenance-fault guard: draft preflight

Status: SOURCE DRAFT ONLY. No tests, builds, service starts, pack/install operations, commits, pushes, PRs or coordination writes performed. Static text/arithmetic checks and source formatting only. Root/native review must release execution separately.

Checkout `/Users/dbox/repos/wt/vickie/vickie-comparison-fault-guard`; branch `codex/vickie-comparison-fault-guard`; unmerged parent `427cfef6decbb422b393024e36218affc26007fa`, tree `9075897867d88344a8477e1f4a8e1a00a50f1228`. Parent CI remains blocked. This ordinary dependent draft does not assert PR120 landed. Final integration must preserve ordinary ancestry after review and exact-head CI; no frozen source was rewritten.

Acceptance SHA256 `c365fdbb9982afe73955ee56a6fdaa5b4c7800b471a8069f3b6d9b89e3cf6248`; accepted v2 `72a06f8e370bc6ed5df10d835464cb102c3c8cad09076e0b7b879b365ea8ee85`; independent closure `daa2008ef5adb78f164275025c4542d18c43d2113f2ab6b4ae15809717ee3e1f`.

All five owned parent files match reviewed70ac byte-for-byte. The parent differs from70ac solely in read-update-matrix.test.ts diagnostics (4 insertions/2 deletions). Five owned draft files and read-only caller/schema/test/config hashes are in source-binding.json, with exact owned snapshots under source/.

Draft patch SHA256 `69c3c4fd2e60ff1ec069755efaac37920b55d51402d89f620f4eeb1faf17dd6c`. `git diff --check` passed as a whitespace check, not a test or runtime validation.

## Draft behavior and source checks

- Production changes only comparisonEpoch: ordinary closed owners return undefined; retained ownerFailure or maintenanceFailure enters the existing fault-first guard. No new API or semantics.
- Three existing owner tests retain old deadline assertions and add epoch observations. Four expanded submission cases use plain real owners with no reservations/cancelSequence, spread vi.fn delegates and an explicit preserved admit overload type. Both admission delegates, final observation and wait must remain uncalled.
- One new HTTP case uses existing emitted runtime seams; selective fake setInterval/clearInterval and clock1000. The seed200 round trip named selective interval/fetch compatibility is the first of the two planned requests, not an additional probe. Fake timer count1 then one25ms advance, injected clock-call count1, timer count0, real timers restored before fresh POST. No f.setTime.
- Actual parsed seed200 receipt is schema-checked and compared with independently authored key/status/pos1/outcome/id/type/properties/metadata facts; the readonly inventory must equal that actual receipt. No invented Resource kind field or fabricated receipt identity.
- Internal error proof is the HTTP handler catch at transaction-http.ts614, marker616 and vary617, not runtime fallback: 500, empty bytes, content-length0, private/no-store, non-attesting, exact varyAccept; absent content-type/retry-after/location/development-limit/all three context fields; no sentinel text in actual body or header values.
- HTTP finally restores timers, closes runtime, performs exactly one readonly inventory and logs requests/actualstatus/faultCalls/intervalCancelled/closed/inventoryChecked, then AggregateError retains primary and cleanup errors. Global fixture teardown remains unchanged and idempotently closes before directory removal. Submission fault finally safely closes and restores timers without entering a faulted owner to cancel reservations.

## Exact expanded source selection

1. `transaction-store.test.ts` — returns a stopped comparison observation for a normally closed owner
2. `transaction-store.test.ts` — preserves a fenced owner's comparison observation failure
3. `transaction-store.test.ts` — preserves background maintenance failure in a closed comparison owner
4. `transaction-submission.test.ts` — propagates closed maintenance faults before admission (observed=false)
5. `transaction-submission.test.ts` — propagates closed maintenance faults before admission (observed=true)
6. `transaction-submission.test.ts` — preserves normal closed submission shape (observed=false)
7. `transaction-submission.test.ts` — preserves normal closed submission shape (observed=true)
8. `transaction-submission.test.ts` — keeps full-owner unobserved comparison exact and excludes a caller callback
9. `transaction-submission.test.ts` — distinguishes a stopped final observation at the reached comparison deadline
10. `transaction-service.test.ts` — uses the original finite comparison budget while an earlier sequence dependency is unresolved
11. `transaction-dependencies.test.ts` — proves static mismatch immediately, retains no timeout outcome and keeps the owner
12. `transaction-sequence-alias.test.ts` — uses one bounded comparison deadline and leaves the owner's reservation intact
13. `transaction-http.test.ts` — preserves maintenance faults as bodyless internal errors before direct admission
14. `transaction-http.test.ts` — keeps a live listener from serving through an owner fenced by injected commit uncertainty

selection.json provides exact argv and suffix-anchored escaped name union. Every literal title or %s template occurs exactly once in its selected file; false/true expansion is statically14. This is not runner discovery. On approved execution the report must contain exactly these14 passes, with all other cases skipped. Unexpected selection or any failure stops; no automatic rerun.

Proposed guarded/supervised positive command after dependency setup and successful build, NOT RUN:

```sh
node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs exec vitest run packages/server/src/transaction-store.test.ts packages/server/src/transaction-submission.test.ts packages/server/src/transaction-service.test.ts packages/server/src/transaction-dependencies.test.ts packages/server/src/transaction-sequence-alias.test.ts packages/server/src/transaction-http.test.ts -t '(?:returns a stopped comparison observation for a normally closed owner|preserves a fenced owner'"'"'s comparison observation failure|preserves background maintenance failure in a closed comparison owner|propagates closed maintenance faults before admission \(observed=false\)|propagates closed maintenance faults before admission \(observed=true\)|preserves normal closed submission shape \(observed=false\)|preserves normal closed submission shape \(observed=true\)|keeps full-owner unobserved comparison exact and excludes a caller callback|distinguishes a stopped final observation at the reached comparison deadline|uses the original finite comparison budget while an earlier sequence dependency is unresolved|proves static mismatch immediately, retains no timeout outcome and keeps the owner|uses one bounded comparison deadline and leaves the owner'"'"'s reservation intact|preserves maintenance faults as bodyless internal errors before direct admission|keeps a live listener from serving through an owner fenced by injected commit uncertainty)$' --maxWorkers=1 --reporter=verbose
```

Use pinned Node24.16.0/pnpm11.20.0 with PATH explicitly bound. No dependencies were installed or linked for drafting. Formatting used the parent checkout existing absolute node_modules/.bin/biome binary; final pass formatted4TS files, fixed1 (17ms). No global format or source-execution command ran. Progress markdown was edited as plain source.

## Static resource ledger

| Selection | Cases | Directories | Owner opens | Readonly inventory handles | HTTP | Listener lifetimes | Installed children |
|---|---:|---:|---:|---:|---:|---:|---:|
| Owner3 + submission6 + service1 + dependency1 + alias1 |12|12|12|0|0|0|0|
| New HTTP positive |1|1|1|1|2|1|0|
| Existing HTTP fencing with one reopen |1|1|2|0|3|2|0|
| Source positive subtotal |14|14|15|1|5|3|0|
| Isolated old-guard HTTP derivative |1|1|1|1|2|1|0|
| Source + derivative |15|15|16|2|7|4|0|

Rechecked fixture implementations and selected bodies: each of the12 non-HTTP cases opens once in one directory; no selected reopen. Existing fencing HTTP reopens once in its original directory. Each new HTTP positive/control performs one additional read-only DatabaseSync inventory after close; these are not owner reopens. Total18 transaction-file SQLite connection lifetimes, not18 stores. Build/derivative/package directories are separate from these store-directory counts. All counts are planned maximum successful-path workload, not measured execution outcomes; premature failures may use fewer requests. No new invocation beyond the14 positives and one control is authorized by this file.

Source requests are enumerated verbatim with UTF-8 lengths/SHA256 in requests.json. All POSTs target actual runtime.targets.batch, use Bearer plus32a token, application/json and their authored idempotency-key. Existing fencing GET targets actual scope+beads/. Actual host/ports and fetch-added headers are not fabricated. Source POST byte counts: new-positive#1=150, new-positive#2=164, old-fencing#1=120, old-fencing#3=120, old-guard-control#1=150, old-guard-control#2=164.

## Counterfactual and execution gate

Create exactly one separate derivative only after release, preserving the positive source. Restore only comparisonEpoch to `() => (closed ? undefined : enter(() => metadata("epoch")))`; no other mutation. Build the derivative before the emitted HTTP test. Select only the new HTTP title above, maxWorkers1. Expected discriminator: actual second response503, nested named assertion maintenance failure is internal500. Require requests2, faultCalls1, intervalCancelled=true, closed=true, inventoryChecked=true and sole retained seed receipt. AggregateError must be inspected through its nested errors (use the existing root reporter that exposes causes or equivalent report inspection); a generic AggregateError/setup failure/timeout is not sensitivity evidence. Preserve derivative patch/hash, built-source identity, raw logs/JSON and all failed attempts. No control has been created or executed.

The selective interval/fetch interaction is explicitly unproved. If the seed round trip fails or exceeds the existing test deadline, save that attempt and stop for the exact compatibility obstacle; no alternate fetch, extra probe, sleep or deadline increase. All source test/hook deadlines remain inherited unchanged (no override in selected cases/config); maxWorkers1 proposed, never more than2. Each released command needs root observable supervisor with120s bound, progress logs, exit/elapsed and descendant cleanup capture. Build first; static/typecheck/format/lint/boundary/Read gates need their own root-approved bounded commands. Do not use the source request ledger as a ceiling for unreviewed broad suites.

## Optional existing healthy installed14/1 (separate later gate)

Root acceptance permits proposing the unchanged apps/bdptest/src/transactional-direct-refusal-installed.test.ts single case, observes current context on installed direct refusals within14requests. It is NOT run here and is not a default-installed maintenance-fault proof. No new installed fault seam or public flag.

After green source/control and root acceptance, freeze a clean exact commit, verify git status is clean externally, build that exact source and pack @bdp/protocol + @bdp/server into a fresh external artifact directory. Install both actual tarballs with exact production dependencies in a separate fresh consumer (not the repo node_modules). Record pack/install commands, resolved lock/dependency graph, tar hashes and installed server/protocol payload hashes; the normal package scripts have no new build or public configuration changes. Require the actual realpath installed entry under that consumer at node_modules/@bdp/server/dist/transaction-installed-process.js and declare the frozen source SHA via BDP_TRANSACTIONAL_SOURCE_COMMIT paired with BDP_TRANSACTIONAL_INSTALLED_ENTRY. Do not use workspace dist fallback or a historical payload with a new source label. Package/consumer paths and actual hashes remain pending because packaging is not authorized at this gate.

The existing test launches exactly one child via process.execPath with cwd/tmp and inherited IPC. Bind ready metadata to canonical authority, profiletransactional, historyfalse, claimEligiblefalse, actual loopback listenUrl, exact12 operation targets, runtimeConfigurationIdentity and served configuration (source, payload, contracts, policy, auth, fixture). The helper records the declared source; it does not independently prove tar provenance. Its workspaceDirty flag is only reported, so the external clean-source and payload checks are mandatory.

requests.json includes all14 static request payloads/routes/keys, including33-operation capacity request. Largest installed body is 4616 UTF-8 bytes. Exactly14 HTTP, one installed child, one new store directory/owner/listener, no restart or inventory inspection in this cohort. Routes are POST operations/create-bead, POST operations/batch, GET beads/?limit=32, GET events/, plus only the actually captured seed receipt URL. Schema/fixture/config identities remain the existing ones; random credential is not printed. No synthetic receipt path.

Unchanged installed bounds: per-request5s; cohort90s AbortController; case120s; request header16384/body65536 bytes; response body131072 bytes; launcher stdout+stderr65536 bytes; startup/each close wait10s with explicit kill-and-drain failure handling. Launch failure can consume a second10s drain bound, so do not describe10s as a universal total cleanup maximum. Keep outer command supervision120s per accepted plan, record whichever test/command deadline wins and stop; no timeout extension. maxWorkers1 and exact single-file/title selector. Capture PID, exitcode/signal, no-active-child assertion, directoryAbsent and full existing report plus logs/hashes. Existing report retains fingerprints/context only, explicitly not complete wire/body bytes; do not claim otherwise.

If optional installed cohort is later approved, combined source+control+healthy-installed workload is21HTTP,1installedchild,16store directories,17owner opens,2readonly inventory handles,5listener lifetimes. It remains a distinct report and identity, not cumulative profile qualification.

## Handoff

Only the accepted five source paths changed. Historical reports remain frozen. No profile claim, release, Preview2, Type/History or protocol change. Root/native draft gate is next; parent remains unmerged and CI-blocked. No execution is implied by these commands or static counts.
