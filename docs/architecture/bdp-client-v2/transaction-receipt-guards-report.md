# Receipt guard and witness follow-up

Frozen clean **27bad2badc765fc02aeec9456e1fc5f9dd49ea5c**, parent
**064db8a99629122f130a91ee8b5189432ffe97e9**, branch
`codex/vickie-transaction-receipt-guards`, checkout
`/Users/dbox/repos/wt/vickie/vickie-transaction-receipt-guards`.
GitHub origin verified; no outward writes. Prior064 and other review trees unchanged.

Exactly three files: transaction-batch.ts (private receipt helper/type only),
transaction-set.test.ts, docs/design/transaction-implementation-progress.md. No store,
evaluator, protocol, schema, HTTP, installation identity, cap or public exports changed.
The owned-transition author was notified that this section is frozen and disjoint
from applyResource; parent owns ordinary integration.

## Review dispositions

1. Private required Problem shape now restricts code to the13 existing receipt codes
   and status to the six existing receipt statuses (400,403,404,409,413,422). The
   evaluator conversion checks membership before terminal retention, without inventing
   defaults. Existing required type/retry checks remain. Valid values and insertion
   order remain unchanged. This is a required-field/subset guard, not a replacement
   for existing evaluator/parser validation of code/status/retry combinations, URI
   grammar and optional diagnostics. No new wire semantics or protocol-wide type edit.
2. Ten malformed evaluator probes remove type/status/retry/code or provide foreign
   code/illegal status. Each now restores the genuine evaluator and immediately
   executes the SAME key on the still-open owner. This proves live reservation
   cleanup instead of relying on startup sweep. Reopen happens only afterwards,
   and separately proves durable replay of the valid404 receipt. Prior report's
   attribution of cleanup to reopen was inaccurate; that report remains historical.
3. Runtime red against exact064: five of ten malformed cases reached the actual
   terminal receipt write and were rejected by the fixture's canonical mutationReceipt
   AJV oracle instead of the expected integrity guard. They were missing code,
   rate-limited, unknown code, status500 and status200. The other five already refused.
   This is direct malformed-retention red evidence for the changed guard; no need to
   mutate the older frozen8d source or claim a new red for previously fixed fields.
4. Added five invalid induced-Event test bounds (0,-1,1.5,4097,Infinity). Actual4096
   Event boundary remains explicitly deferred to owned/materialization work; no
   costly1366-Link test or new cap changes in this slice.
5. Compiler negatives now exercise missing type/status/retry/code on whole-limit
   Problems as well as operation-local ones. Read-only rate-limited and nonreceipt
   statuses are rejected statically. Prior064 typecheck produced three intended
   failures for the newly restricted subset probes; final typecheck passes.
6. Repository prose now names the first observed platform ceiling: V8's536870888
   code units before SQLite's1000000000 bytes. For the measured ASCII payload family,
   2048 maximum-sized postimages alone exceed that string ceiling. This remains an
   arithmetic projection, not an attempted allocation failure. No materialization
   fix or new limit claimed.
7. Rewrapped the touched deletion-provenance paragraph. No broader editorial cleanup.

## Verification

Pinned Node24.16.0 / pnpm11.20.0; external command logs and explicit exits here.

- Frozen install: PASS.
- `red-parent`: intended **5 failed /5 passed /43 filtered skips**. Actual schema
  failures show the predecessor retained malformed receipt Problems.
- `type-red`: intended3 compiler errors for invalid code/status subset membership.
- Focused green: **53 passed**,881ms.
- Final regressions: **566 passed /14 files**, no skips, **6.15s**; all transaction
  test files plus Resource/alias evaluators, including emitted HTTP restart/pages
  and actual hot-journal process-crash recovery. JSON + exact command recorded.
- Typecheck, clean build, dependency boundaries, touched TS format/lint: PASS.
- Packed executable smoke and development smoke: PASS.
- Diff check and clean commit: PASS; no hooks bypassed.

No full repository suite, public profile qualification or evidence relabeling.
No interrupted runs or test-owned processes remain; disposable test stores cleaned.
Materialization plan is separately frozen in transaction-materialization-plan.md
(SHA25697f07f2bc25191dcf2c6d9ff5640161d5e4d2a53c698f29423a6aa5e57919ac8),
not implemented by this commit.
