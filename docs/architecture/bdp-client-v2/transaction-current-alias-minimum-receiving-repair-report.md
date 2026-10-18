# Alias minimum receiving repair

Frozen **3959a25e70bc6dc9d143abfead187ccaf84ca11e**, tree **7d6212ea04046d7c45a337b293a13a482a397caf**, direct **961125320766b3f02c3ff73ad6029f11117fedae**. Isolated checkout `/Users/dbox/repos/wt/vickie/vickie-alias-minimum-receiving-repair`, branch `codex/vickie-alias-minimum-receiving-repair`. Original author/reviewer checkout remains frozen. Root owns council and integration; no profile qualification follows.

## Native P2 disposition

Three files changed: transaction-read.ts, transaction-service.ts, transaction-read.test.ts. The new source-private alias options helper reuses the existing descriptor-safe ordinary options capture: proxy rejection precedes prototype/key reflection, accessors are rejected without execution, all own data descriptors (including non-enumerable signals) are retained, and only signal/sourceCondition field names are allowed. It additionally validates the signal's native shape using the authority's same predicate before service wait code can inspect it. Captured option data is detached and frozen. No public root export or runtime package entry was added.

Direct alias minimum resolution injects its immutable configured minimum only after successful capture. When capture rejects, it delegates the original invalid input to authority delivery, preserving its observed Promise and closed/reentrant precedence. This is not a permissive ordinary-read fallback: capture can reject only the same option shapes or signal shape that authority rejects before any owner entry. The tests exercise every such branch with a live, minimum-satisfied alias and assert zero entries. The service captures inside existing lifecycle admission, before selecting a wait signal; failure returns the same facet rejection rather than entering the wait loop with unchecked data.

An own data sourceCondition value is deliberately replaced by the facet's configured minimum, as before. It is opaque caller data; even a revoked proxy stored as the *value* is not inspected. A sourceCondition accessor is invalid and rejected without invocation. Tests also show a caller-supplied pos-0 condition cannot weaken the facet's pos-2 minimum. No new condition semantics are inferred.

The inherited minimum `perform` analogue is explicitly out of scope and unchanged. Ordinary facets, HTTP routing, fixtures, installed scenario, public APIs, schema, configuration, admission and all bounds remain unchanged. Entire installed test equals the parent byte-for-byte, SHA256 **69fc8bd2efdbf7a30d9dbdb88a7772ced92ae8bd724a4ba159a88980e2c2df47** (44,601 bytes); preservation and source hashes are recorded separately.

## Actual validation

Pinned Node24.16.0 / pnpm11.20.0. `exits.json` contains full commands, durations and exits, with corresponding logs.

- Clean → typecheck → build before tests: pass. Final frozen clean → typecheck → build and fresh pack/install: pass.
- Affected Read/service/HTTP/authority/consistency and actual app disconnect regressions: **334 tests / 6 files pass**. No full workspace run.
- Format, lint, boundaries, existing Read seal and diff check: exit0.
- New direct and service adversarial cases cover null/scalars/arrays/nonplain prototypes, ordinary/revoked proxies, symbol/unknown and non-enumerable unknown fields, signal/sourceCondition accessors, invalid signals including getter-bearing plain objects. All return Promise rejection while open, execute no getter/proxy trap and enter no owner read. Non-enumerable own pre-aborted signals reject before entry; valid null-prototype data with non-enumerable signal succeeds. Reentrant invalid delivery returns authority reentrant rejection. Closed direct facet returns a rejected Promise with existing `server-closed`; closed service preserves synchronous admission error. Existing ordinary HTTP and actual Request cancellation tests remain green.
- Disposable exact archive countercheck: replacing only the two production files with parent9611253 causes both new cases to fail on invalid null options resolving successfully. Restoring frozen3959 source passes both. Expected red exit1/green exit0 retained; authoritative checkout never mutated for the probe.
- Fresh offline installed protocol/server proof: four separate minimum scenarios **38/1, 10/1, 20/1, 18/1** pass, retaining new alias **18 requests/1 child/7 URLs**. Four independent configurations/stores, not one combined cohort. Actual source declaration, payload/config/fixture/policy/auth identities and completed-response hashes retained in summary. All four children exit0 and PIDs are absent.
- Same fresh server payload: two targeted installed RU positives pass (retained retries and reference syntax), 200 tests intentionally filtered.

Installed payload digest **91e78a58669ad359cb82174a3ff3075e22a5bbfffe3790ad2ffc78ba536e887c**. Protocol tarball **4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227**; server tarball **6e2007d4ff92503cc4573ce1e87bdedc8eec9e995ca2b85beede53e71c616a2a**. SourceCommit is checked launcher metadata, not independent target attestation.

## Retained failure and limits

The first new test run had one failure: its direct closed-error expectation incorrectly used `reason:closed`; actual existing authority contract is `ScopeServerClosedError` with `code:server-closed`. The test was corrected to that contract without changing production. Original receiving.log/json retained. Subsequent334 affected tests and exact frozen proof pass. No runtime failure was suppressed.

Public claims remain false. Existing all-visible authorization limitations, incompatible missing gap, and cooperative cancellation limits are unchanged. No inherited minimum-perform repair, new lifecycle mechanism or hidden-target qualification is claimed. All owned runs completed, disposable source restored, installed children absent and author checkout clean. No PR/main action is performed by this lane.
