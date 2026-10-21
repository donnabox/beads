# Native CODE review — frozen current-observation draft

## Findings

**N1 — P2: stale Link Event HTTP aggregate guarantees a selected callback failure.** `packages/server/src/transaction-events.test.ts:2233` still requires twelve status-500 responses. The foreign-epoch and foreign-view GET/HEAD pairs at lines 2165–2166 now correctly require 410 and 409. The only remaining 500 pairs are SSE, continuation (`after`), reconnect (`last-event-id`), and unsupported media at lines 2168–2171: four pairs, eight responses. The new per-response assertions and the old aggregate cannot both pass. Change only this aggregate from 12 to 8; retain the 35-request workload, timeout count, route exclusions and individual response assertions. This was independently confirmed from the frozen callback after the author's preparation audit raised it. No execution was needed or performed to establish the contradiction.

No additional actionable production, lifecycle or semantic blocker identified in the reviewed diff. This is not approval to execute or publish the unchanged draft: N1 requires disposition/correction.

## Verified scope and behavior

Reviewed the full sixteen-file diff and relevant existing helpers against the accepted successor-2 plan and its eleven dispositions. The five production changes preserve strict consumers while extending current-observation eligibility: unsupported lexical positions on an admitted waitable path check actual epoch/view before reporting unproved eligibility; Type/current Types and existing initial collections use that current path; existing Resource/incident absence handling remains separate; discovery and receipt pages capture caller tuples once and re-observe current owner context on retries. Receipt page parent/principal/token nondisclosure remains ahead of minimum eligibility, and eligibility remains ahead of read-triggered expiry. No public profile/discovery admission, production timeout, schema/export, retained Snapshot S1, or incident projection G2 expansion is introduced.

Inspected the nine new source cases: A's captured tuple across one real intervening commit and observed context; B's missing/hidden Type and receipt page nondisclosure ordering; five C drain cases with immediate settlement observers, bounded entered-versus-settled races, owner probe, clock advance/release in finally and a returned post-close entry refusal; D's synchronous clock advance/read/owner-close/offline comparison; and E's real locally issued receipt page, scope-local Types, 37 HTTP exchanges and finite runtime cleanup. C's single microtask close check is a bounded schedule observation, not proof of every possible premature settlement. D deliberately performs no asynchronous yield between its retention-sensitive actions.

Inspected the new installed callback's 32-Bead mutation, 32-match update, actual issued next page, Link mutation, positive schema/content assertions, 503/410/409 and HEAD metadata parity, principal/token refusal, raw request budgets, request/operation drainage, child exit plus close event, shared cleanup deadline and conditional directory removal. Its 45/5/10/60 second bounds are unchanged. Startup cleanup before a child is returned remains the launcher's inherited contract. Installed assertion coverage and future external-consumer provenance must remain distinct from workspace schema/runner attribution.

The static selection contains 97 unique source cases: 88 legacy and nine new, across nine test files. Every declared literal title/template exists in the actual source, and the anchored per-file patterns match all listed expanded full names. This is static selection verification, not Vitest collection or proof that no runtime selection surprises exist. New source reports total nine (A, B, five C, D, E); the separate installed callback adds one. Reviewed report schema fields against their actual report construction and assertion/counter paths. The advertised new source quantities are source-derived: 47 service operations, eight direct owner page calls, 65 target attempts, 17 waits and 74 minimum-clock calls, excluding E's uninstrumented owner activity. E and F each author 37 HTTP requests. Selected installed cohorts remain exactly 22 and 37, with immediate separate audit intent; the unchanged five-request sibling belongs to later full-file/CI coverage, not focused 59.

Reviewed legacy fanout across Type/incident, alias, receipt, Snapshot, finite Event, Scope Event SSE and finite changefeed fixtures. Strict/reconnect/retained exclusions and explicit unsupported parser sentinels remain present. Both late unknown-current HTTP wait checkpoints are ten while the earlier known-position checkpoint stays eight. N1 above is the outstanding legacy aggregate mismatch.

## Identity and preservation

At review end HEAD remains `d9c3eb9e61b1982b9bd56aa77e4cdc114384e886`. All 679 actual tracked-file SHA-256 values equal the frozen binding; exactly the sixteen authorized files are dirty. All 663 other files match both the binding and HEAD Git blobs. Independently checked all sixteen before/after hashes against HEAD/current bytes, installed module prefix, exact unchanged five-request sibling slice, and full historical progress-document prefix. Current `git diff --binary` equals the frozen patch hash. Original artifacts were not changed.

- Author index: `1d2dcf86e5d2e6547cacf21653ec2874c0416095295e2ab8ba302f64e4dd2692` (43 entries; indexed file hashes verified).
- Source binding: `d81b271b04b78af8da81b1f85bc84c9d2b2ff1856979d1a92c8d79215165843a`.
- Full patch: `194419e1b5210531b781fd822dd8e0de1be4d2c18c70e4d9a77a503a0a78bf7c`.
- Selection: `6db62011c00a9bede399241e088e4881746bc8a464e0deb1dc91d00d0c2803a0`.
- Report schemas: `c08334d504d8fd0bef23866bea32d7af1832475676b0a46952115ddcc7e83c12`.
- Workload: `18750e1973de9a07a6d557d48bcef2956f29bc38e7fc4d5b4aa46f2a246db1db`.

## Limits

Read-only source/data review only. No product or helper imports, compiler, formatter, lint, test collection, tests, build, package installation, runtime, external review, commit or push was executed. Formatting/static validity and runtime assertions remain unverified; visible layout is not an observed formatter result. No wire census, profile completion, public admission or remote process-cleanup claim follows from these source-derived counts. Only this separate report was written through the session guard.
