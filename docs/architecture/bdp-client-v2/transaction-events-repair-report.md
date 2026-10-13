# Finite Event council repairs

Frozen clean source: **5909666cb86094b467e51cb98cbaa8e584c6a472**.
Direct parent: **e2f1385aa89bc6ae27e18a802a2a651a8f02e87d**.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-events-repair`.
Branch: `codex/vickie-transaction-events-repair`, GitHub origin. No outward writes.

Exactly three changed files:

- `packages/server/src/transaction-store.ts`: readEvents minimum gate and page accounting only.
- `packages/server/src/transaction-http.ts`: alias/receipt route classification before Event parsing.
- `packages/server/src/transaction-events.test.ts`: discriminating regression and missing controls.

The frozen original, G1 extraction, runtime/service/Event module, tuple and group-bound literal were not edited. The M32 follow-up remains separately authored and must be integrated ordinarily later.

## Council dispositions

**Medium 1 closed — gate before retained scan.** A deterministic red test observed two group-query preparations for a behind minimum. readEvents now checks the supplied minimum immediately after obtaining its same-entry context, before cursor membership, query preparation, group copy/parse/projection or page serialization. Unknown source and unsupported-control handling remain in their existing earlier positions. The counterfactual now observes zero group reads. The synchronous entry cannot advance its own head while scanning, so no later recheck is needed. HTTP grammar still rejects malformed cursor syntax in its established parser; no new error code or product precedence was invented.

**Medium 2 closed — retain alias and receipt route ownership.** Actual emitted HTTP reproduced 400 invalid-parameter for query-bearing alias, known receipt and unknown receipt URLs. Classified malformed/query aliases now retain their established resource-not-found result; receipt roots are dispatched to their existing receipt identity checks rather than Event parsing. GET and HEAD controls include a future minimum and remain uniform 404, without disclosing a receipt or alias. Query-free alias/minimum behavior and ordinary Resource Event handling are unchanged.

**Low 3 closed — linear exact wire accounting.** The loop no longer constructs and serializes a growing candidate page for each Event. It serializes the empty envelope once, subtracts the four literal bytes of null, accumulates each serialized candidate Event plus comma bytes once, and adds the exact serialized candidate continuation URL. This is algebraically the same bytes as the prior `{source,events,next}` serialization. A selected final page with null is smaller; an overflow continuation is the one reserved when the last accepted Event was added. The item ceiling and 4 MiB byte bound are unchanged. A serializer observation proves one empty-envelope serialization plus 20 individual Event serializations for 20 accepted Events, using UTF8, newline, quote and backslash data. The existing actual-byte long-source splitting test continues to pass. This removes the within-page quadratic work; it does not bound total retained-log scan time or parsed group heap.

## Added observations

- Actual HTTP malformed and cross-source/never-member cursor results are 400 invalid-parameter; HEAD has the same refusal and no body.
- Successful HEAD content-length equals the corresponding GET's actual UTF8 byte count, with matching content type and no HEAD body.
- The unresolved metadata-only incident branch is observed over actual HTTP: bodyless 500 with `x-bdp-development-limit: incident-link-metadata` and the nonattesting marker, for GET and HEAD. Scope and Link sources remain available. This tests an explicit private gap, not an accepted metadata membership rule.
- Twelve unrelated groups contribute zero Events to a resumed Bead source; it returns a true empty end with current context. A later contributing mutation becomes available through the same old cursor.
- Injected retained-read corruption controls cover missing group, wrong epoch and wrong ordinal, asserting the exact integrity error; restoring the read seam yields the original valid pages and owner state. These are bounded read-seam injections, not disk-fault qualification.
- Review example correction: actual `beads/%61?view=events` returns 400 invalid-parameter through the existing canonical Resource-path check, not a distinct-identity 404. The new test preserves the observed current result; no canonicalization semantics changed.

## Validation

Pinned Node **24.16.0**, pnpm **11.20.0**. Frozen dependency install completed. Builds completed before emitted HTTP tests; no clean/build overlapped test execution.

- `red.log`: **2 failed / 12 passed** (14 tests), confirming both Medium defects before production edits. The HTTP test retained all three route witnesses through soft assertions; minimum test observed 2 versus expected 0 group queries.
- `green.log`: 14/14 after fixes.
- `extended2.log`: 19/19 after added controls (one intervening run corrected the review's inaccurate encoded-path expectation).
- `final-tests.log` and `final-tests.json`: **238 tests / 5 files passed**, 4.21 s. Files: transaction-events, transaction-service, transaction-http, transaction-store, transaction-materialization.
- `final-typecheck.log`, `final-build.log`, `boundaries.log`, `format.log`, `lint.log`, `evidence.log`, `smoke-development.log`, `smoke.log`: all completed successfully. Existing workspace lint/large cumulative manifest notices remain; no new diagnostics in the three touched files.
- Read seal artifact and constant are unchanged; historical verification passes. Neither these tests nor the installed smoke qualify a public Transactional profile. No broad workspace suite or new Read cohort was run.

## Limits / later integration

G1's frozen extraction is not included here. Its ordinary merge must preserve the early minimum gate and incremental byte accounting if readEvents is moved onto the shared group accessor. M32's separately owned bound-literal update was deliberately left untouched.

Not every retained-integrity branch has a dedicated fault test: inconsistent endpoint witnesses, missing incident endpoint witnesses and the mathematically unreachable first-valid-Event overflow remain supported by source/proof and existing semantic tests, not newly claimed injected coverage. The unknown/never-member cursor 400 convention is unchanged; an untrusted token alone cannot prove that a missing row once existed. No new corruption status distinction was introduced.

Metadata-only incident semantics await the operator; SSE, mutable policy, pruning/erasure, public admission and product-level semantics remain outside this repair. Root owns independent review, integration and publication. Source is frozen with clean git status.
