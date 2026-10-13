# Native review: durable receipt paging plan

Reviewed `transaction-after-minimum-plan.md` against frozen `d3252087a31529cf97b5bfcfe2c402f02c694b91` in `vickie-transaction-point-minimum`. The source checkout remains clean. This is a source/spec plan review, not execution evidence for paging; no tests were run and no repository files were changed.

## Verdict

Proceed with the approved representation after making the two configuration requirements below explicit. Neither needs a product ruling or new protocol semantics. Full immutable facts plus atomically recorded page boundaries is a suitable bounded extension of the existing owner. There is no architectural blocker to the proposed slice.

## Required plan clarifications

### P2 — Bind the receipt partition bound to the advertised maximum, not merely a compatible smaller value

Plan lines 99–103 say to check the authority-chosen bound against `page.maximumItems`, without defining that relation. In this configured fixture, use the same positive integer for the persisted receipt bound and advertised `page.maximumItems`. A mere `receiptBound <= maximumItems` test is insufficient: receipt bound 2 with advertised maximum 32 would paginate five results even though all fit the advertised bound. The existing spec requires all such results inline with `next: null` (`bdp.md:5067–5072`, `5167–5175`). `page.defaultItems` is not the receipt partition authority.

Use one validated snapshot for store construction, advertised maximum and the Read pagination maximum, preserving a valid default at or below it. Add explicit controls that five results with maximum 5 remain inline, while maximum 2 produces 2/2/1; reject divergent private configurations. There is no need to add a public option or alter the protocol.

### P2 — Validate shared configuration before opening the durable owner

The plan requires incompatible-bound refusal before mutation (acceptance 11). Current `transaction-runtime.ts:223–231` opens the store before constructing the service; service validation presently occurs in `transaction-service.ts:62–67`. Opening an existing store runs pending retraction, creator-fact cleanup and expiry maintenance after its installation check (`transaction-store.ts:390–420`). Merely replacing the service's current `>=32` check would therefore discover a store/service bound mismatch after maintenance.

Validate the shared bound and any advertised/pagination consistency before `openTransactionStore`, then pass the same detached values to both components. Persisting that bound in the installation fingerprint handles changed-bound reopen refusal before recovery. Keep a defensive service/owner compatibility assertion, but do not rely on that later assertion to prove byte-identical startup refusal.

Prove both prior-format and mismatched-bound refusal on a disposable existing file containing pending and expiry-eligible records. Compare the complete file bytes before and after refusal, and verify that compatible reopen still performs normal recovery. Do not create the new page table, rewrite metadata, retract pending work or compact receipts before the old-format/configuration check. No migration, reseed or automatic retry with a new store.

## Additional implementation acceptance detail

**Fail closed on an inconsistent recorded page plan.** The proposed “no page records means full array” fallback (plan line 97) is correct only for an actually inline receipt. A retained array larger than the persisted bound must have its recorded chain. Validate contiguous, nonoverlapping in-range slices, correct parent/principal binding and final coverage before exposing a prefix/page. Missing or inconsistent available detail must produce a service/integrity failure, not an oversized inline response, regenerated page identity, silent truncation, or an expiry response. One disposable corruption regression can distinguish this from the already-planned rollback/no-orphan cases. This is defensive integrity validation, not a new normative storage requirement or a request for a generic cursor framework.

## Reviewed invariants to preserve

- Allocate page identities in the same terminal SQLite transaction as effects, group and full receipt. Exercise late failure and uncertain-commit fencing; no successful pages for failed/pending/retracted receipts.
- Keep the full internal result array separate from wire projection. Existing `expireRow`, `creatorFact` and sequence projection depend on it. Tail-page creations must survive as allocation tombstones; singleton/alias receipts must remain inline and preserve cross-carrier identity.
- Retain issued page identity after completed-detail expiry so authorized known pages return 410, while unknown, foreign-principal and prior-epoch pages remain uniform 404. Failed receipt forgetting remains distinct. Tokens confer no authority.
- Existing receipt-page law explicitly orders authentication, principal/epoch nondisclosure, minimum comparison/wait, then page expiry and representation (`bdp.md:4590–4624`). Recheck lookup and context after each wait. Do not hold a page body across a wait or add a second clock-based expiry lookup to initial committed delivery.
- Stored boundaries and canonical URLs survive restart and a changed listener address. Current graph state must not reconstruct historical result postimages. Current projection and current serving context remain distinct from immutable execution facts.
- HEAD/conditional behavior, private/no-store, validator omission and bounded service shutdown apply to pages. Supported facet AbortSignal cancellation does not imply immediate HTTP socket-disconnect cancellation.

## Scope and normative unknowns

No unresolved normative choice blocks this paging slice. Fixed all-visible, schema-free/unowned tests cannot qualify variable-view authorization or erasure. Explicit batches cannot satisfy the catalog's large-set expansion witness. Unsupported minimum tuple forms, absent/hidden Resource ordering and retained query continuation remain existing private gaps; the explicit receipt-page ordering rule does not settle those separate questions. Public discovery/admission, catalogs and sealed Read evidence must not be relabeled by this implementation.
