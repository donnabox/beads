# Finite Transactional Event serving — frozen combined implementation

Frozen clean source: **`8569b85d785f96325ca948580106f454c48ed034`**.

Ancestry: finite checkpoint `97ae736b3849abb78e30fbf5fefca20e2f1e1cd6` and materialization `076201849f98eeab71bad5c8404fb693a36c6474` were ordinarily merged as `c425ea8`; final bound/gate follow-up is the frozen head. Both parent slices descend from `c0d67388cd849410b38ec7ecb3a328771431dbfa`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-events-finite`.
Branch: `codex/vickie-transaction-events-finite`. No outward writes.

The materialization tuple is now integrated, and its exact group bound gates body materialization. Source is immutable for review. **Metadata-only incident Link projection remains a named private gap** pending the operator's disposition; no specification rule was invented. Materialization's independent native/external review remains a separate dependency; this report does not substitute for that review.

## Final combined validation

Pinned Node 24.16.0 / pnpm 11.20.0. `combined-typecheck.log` and `combined-build.log` pass, with build completed before emitted tests. `combined-focused.log` / `.json`: **300 tests across 7 files passed**, 4.32 s: Event, service, HTTP, owned, materialization, set and store. This includes 13 Event tests and the added Event service lifecycle test. No full workspace suite was rerun.

`combined-{boundaries,format,lint,evidence,smoke-development,smoke}.log` all pass. Lint reports 34 existing warnings and one info; the cumulative matrix's configured size warning remains. Both development process and installed tarball/public-admission smoke checks pass. Historical Read seal verification remains unchanged at 89 rows; no new profile or Event conformance qualification is claimed.

The only merge conflict was the progress document's appended sections; both full sections were retained. The store merged automatically: inspected result preserves the actual owner tuple, early Scope bound, exact configuration compatibility gate and Resource evaluator options. No schema or evidence files were edited. Final git status is clean.

## Bound/gate evidence

The imported frozen tuple is M64/E256/R262144/Q65536/S65536. The finite module derives the same closed-shape group ceiling **253493248 encoded bytes**, using K=M+32, N=7Q, A=8192 and H=1048576. The owner first selects position and SQLite `length(CAST(body AS BLOB))`, checks the bound, and only then selects/parses the body. The targeted counterfactual overrides the reported length and traps any body retrieval, proving refusal occurs before body copy; it is an injected guard-order test, not a claim of an actual disk fault or maximum-size workload.

The first-Event/page proof explicitly includes Resource-source expansion and source repetition in page/continuation: **5R + 7Q + 8192 + 1024 = 1778688 < 4194304**. Module admission rejects an incompatible future tuple rather than inventing a response 413. Actual-byte tests use a long stable Scope, verify every schema-valid page is below 4 MiB and demonstrate pagination below the item ceiling. A corrupt missing group is refused; no pruning is implemented.


## Implemented

- Source-private `transaction-events.ts`: canonical source classification, opaque fixed-width epoch/view/source tags plus position/ordinal, membership validation, same-group endpoint reconstruction, unchanged fact payloads/ordinals and source-specific rendering. Scope rendering is reusable for later changefeed serving.
- Same-owner `readEvents` obtains source existence/current context and iterates one committed group at a time. No all-log copy or SQLite iterator escapes. Finite pages split groups, include complete Events only, and reserve actual encoded continuation bytes inside a 4 MiB page budget.
- Source identity survives deletion; receipt expiry does not prune Events. Link views contain lifecycle facts. Bead views contain own lifecycle facts, endpoint incidences and incident property updates. Own-Bead metadata and combined Link property/metadata updates are observed without inventing the unresolved metadata-only incident rule.
- Private service/HTTP integration preserves authorization/raw target/Origin/Host gates, applies minimum waits before conditional success, returns current context, and shares JSON/HEAD/conditional handling. Local reference conventions are documented: malformed/never-member cursor `invalid-parameter`, unknown source 404 before minimum. Tagged context mismatches preserve foreign-view/epoch refusal distinctions.
- No public discovery/profile/export, schema, sealed evidence or Type lifecycle change. SSE, mutable authorization views, erasure, Snapshot and changefeed serving remain out of scope.

## Observed validation before tuple integration

Pinned Node 24.16.0 / pnpm 11.20.0; build completed before emitted HTTP tests. All logs are adjacent to this report.

- `focused.log` / `focused.json`: **96 tests in 4 actual files passed** (Event, service, HTTP, owned), 3.88 s. A supplied nonexistent `transaction-minimum.test.ts` filter did not add a test file; no fifth-file coverage is claimed.
- `finite-final-before-tuple.log`: 12 Event tests passed after additional direct metadata/expiry assertions; 1.22 s. Earlier red runs exposed fixture errors (invalid descriptor enum and malformed empty patch), subsequently corrected. No production defects were concealed by those corrections.
- The expiry oracle advances strictly beyond expiresAt, asserts two successful receipts compact and one failed receipt is forgotten, verifies the committed receipt is expired, reopens and compares unchanged Event pages.
- Service test proves facet AbortSignal cancellation, bounded drain and closed-owner refusal. Actual HTTP disconnect does not claim immediate cancellation; it can drain the finite wait budget.
- Build, typecheck, boundaries, format, lint, historical Read evidence verification and both executable smokes passed. Existing lint warnings and the cumulative 1.7 MiB matrix format-size notice remain. New files' non-null assertion warnings were removed.
- Historical Read seal remains `674982be2123b9b7aae630bf79c3f774122900cd`, observed run `e5f7d0c43c151426350ceededdc77fe7e372db9b`, evidence commit `dbaf63a24219e1ea3f80394746853598b0d4464a`, 89 rows. This does not qualify the new Event surface.

## Limits and required next review

The unresolved metadata-only incident case is deliberately a private development gap. SSE, changing views, erasure, Snapshot and changefeed serving are not implemented by this slice; no public Transactional capability is enabled. Canonical Scope/Bead/Link Event routes only are covered here. Local unknown-source/cursor conventions do not establish universal error precedence.

The finite implementation scans retained groups synchronously under one owner entry; it bounds page and per-group encoded materialization, not total retained-log length, parsed heap overhead or synchronous CPU time. Cancellation is observed between minimum-wait attempts, not in the middle of a synchronous group scan. No pruning or policy mutation is implemented. Source-context tags are not credentials or an authenticated proof of issuance; matching-context requests still undergo retained source-membership checks.

Root owns independent council review, materialization-review disposition, combined full workspace validation and outward publication. No further edits are planned to this frozen checkout.
