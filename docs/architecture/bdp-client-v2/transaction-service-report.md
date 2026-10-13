# T2a: private same-owner service and atomic receipt observation

Frozen head: bddc4486fe39676bb84d63e05dda22f1bfa77a1a
Parent: 1949f25cbd0e1c639a26e482f4a82763acf8aa20
Checkout: /Users/dbox/repos/wt/vickie/vickie-transaction-service
Branch: codex/vickie-transaction-service
Origin: https://github.com/gastownhall/bdp.git

Clean local commit. No push/PR/bus/launch writes. No active test processes or interrupted runs. This branch has not yet incorporated root's separate clean-checkout typecheck repair to read-pagination-public.test.ts; no exact repair commit was supplied during this slice. Root can ordinary-merge that disjoint test change during integration.

## Implementation and ownership

Six files: new transaction-service.ts/test.ts; narrow transaction-store.ts receipt read/helper plus tests; progress note; private emitted smoke inventory/root negative check. No evaluator, HTTP/Read routing, U2 contracts, schema, storage format, protocol or public export edits.

The service takes one existing owner under explicit ownerLifetime:'service'. Successful construction transfers shutdown; failed construction leaves the caller's owner usable. One Read plane, existing submitTransaction/finite comparison policy, existing sequence stepper and existing Node scheduler share this owner. There is no second connection, key table or write policy. Accepted work is tracked before execution. Close refuses new calls, drains accepted work even if the response was dropped, closes the Read plane/cursors, then closes the owned database. Cached Read facets also reject new work after closing.

Sequence turns yield through the existing lifecycle scheduler contract. Scheduler failure retracts only this attempt's unexecuted tail, preserves committed prefix/foreign pending reservations, and exposes the original cause. Invalid inline/repeated callbacks cannot execute a member twice. If reservation cleanup fails, service entry is fenced and the original plus cleanup causes remain observable through acyclic aggregate/cause traversal. Ordinary successful cleanup leaves the owner reusable; reopening retracts abandoned pending state.

The owner.readReceipt entry performs principal-bound lookup, expiry cleanup, current fixed all-visible projection, detached/deep-frozen receipt bytes and typed current epoch/view/head capture in one guarded SQLite transaction. It shares existing receipt lookup/retention logic and the existing Read context calculation. The existing receipt() behavior is preserved. Tests inspect the real DatabaseSync transaction/reentry guard for both receipt selection and metadata context selection; no mock receipt or synthetic position is used.

Direct submit delivery rereads that current receipt observation atomically; it does not combine the execute/admit result object with a subsequent diagnostic position. Later service receipt reads use the same entry. A retained postimage stays historical after Resource deletion; requiredPosition/effectPosition and original execution view remain unchanged as current context advances. Fixed-view authorization is actual current owner configuration, not a actor/provenance-derived principal. Unauthenticated reads refuse before lookup; principal mismatch, unknown/root URLs, forgotten failed and retracted receipts are absent without observation disclosure.

## Normative boundaries

Current docs/specs/bdp.md sections Mutation Receipt responses (around 5090–5240), mutation surface/non-disclosure ordering (around 4600–4624), and admitted sequence disconnect rules guide this slice. The native required-contract matrix remains authoritative about unfinished serving obligations.

`page.maximumItems` explicitly bounds inline receipt results in the current Mutation Receipt responses text (around 5168). Each supported private batch has at most32 single-Resource operations and at most32 results; singleton aliases produce one. This service therefore requires configured advertised page.maximumItems>=32 until durable receipt paging exists. This is an interim supported-configuration restriction, not a protocol minimum. The maximum32 receipt is tested with next:null. No maximum representation-byte limit is advertised by this slice. No page URL is invented or issued.

This is source-private service composition, not complete receipt HTTP serving. It implements only existing fixed all-visible/schema-free/unowned capabilities. Variable grants, erasure projection, durable receipt pages, set expansion, installed contract evaluation and sequence Resource-alias member-turn behavior remain gaps; no full profile/directory/evidence claim is made. Sequence() accepts a branded prepared carrier/trusted principal; it is not a raw HTTP sequence receiver. Receipts root/method/HEAD/conditional/minimum-position semantics still need HTTP composition. Required Snapshot/changefeed paths remain later work. No pending execution is cancelled because a caller drops delivery.

## Verification

Pinned Node24.16.0 / pnpm11.20.0; frozen install. Build preceded emitted child-process tests and smokes.

PASS 518 tests /11 files /2 workers (4.82s), regressions.log:
transaction-service, transaction-store, transaction-read, transaction-sequence, transaction-dependencies, transaction-submission, authority-read, read-pagination, read-pagination-public, read-sequence-lifecycle, sequence-executor.

New acceptance includes actual same-owner canonical afterstate and one-group cross-carrier replay; current observation versus retained execution facts; nested immutability; trusted identity versus attribution; principal nondisclosure; completed retention/failed forgetting/reopen; pending duplicate and retraction; finite unresolved dependency comparison; detached direct/sequence drain; cached Read facet closure; scheduler/tail ownership/failure cause tests; largest inline receipt; atomic receipt/metadata entry and all existing fencing cases extended to readReceipt.

Three built-process service cases terminate without graceful close: after admission, during tentative effects and after completed delivery. Reopen retracts uncommitted work or preserves exact original completed receipt; retry produces exactly one group. These are actual abrupt process exits and injected SQL faults, not disk/power-loss qualification.

PASS build, typecheck, check:boundaries (11 packages), format:check, lint (existing3warnings/1info), git diff --check.
PASS installed-package smoke (new private module inventory and negative public-root/deep imports), actual non-attesting development smoke (existing RU runtime). Existing development smoke is not evidence that TX HTTP exists.
PASS unchanged evidence:verify: constant674982be2123b9b7aae630bf79c3f774122900cd; runHead e5f7d0c43c151426350ceededdc77fe7e372db9b; evidenceCommit dbaf63a24219e1ea3f80394746853598b0d4464a; schemaReadProjection bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc;89rows. No reseal/relabel.

An initial expanded test run had four fixture failures (missing JSON-Patch request field and omitted child-process pagination callbacks after JSON serialization). They were corrected; final build and all518 regressions passed. Full repository suite was not rerun in this bounded worker slice. All logs are adjacent to this report; no stopped/inconclusive run is represented as a pass.

## Recommended next partition (root review before edits)

T2b should ordinary-integrate reviewed U1 routing and root's test-only typecheck fix, then add a private TX HTTP receiver/runtime over this same service. Keep receiving normalization separate from authoritative owner admission; reuse direct submission precedence, add truthful sequence syntax/auth/negotiation admission, and derive directory entries solely from actual wired operations. Receipt GET/HEAD/root/method handling must preserve nondisclosure and use one readReceipt observation; every mutation receipt delivery uses current observation. Wire trusted transport credentials separately from revision attribution and test cold discovery/navigation/restart in installed processes. No second store or diagnostic inspect adapter.

Receipt minimum-position/conditional handling and sequence envelope serving context need explicit current-spec tests before their HTTP claims; do not merely decorate an old response with later head. Root controls set operations and U2 installed-contract expansion sequencing. Durable pages, variable policy/erasure and required replication Snapshot/changefeeds must remain explicit later qualification slices, not waived by the currently supported inline/all-visible configuration. T2a alone does not enable public Transactional admission.
