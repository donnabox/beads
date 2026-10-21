# Native receipt missing-context plan review

**One blocking plan defect (N1, Medium): installed request 2 uses the wrong closed-schema field.** No additional blocker found in the proposed three-production-path repair, caller inventory, or bounded source/control selection. This is plan review only, not implementation or execution approval.

Reviewed frozen plan SHA-256 `618e69a22bf6f0d5acc01dbbee6001dbd63b17d16fee4529c18e000bfcf47d18` and ledger SHA-256 `133d93f01d6fb85bf111b016a49a0467470ce09694b8dff7dd1b40b591f5ce9d`. Actual source: clean `3df82ac480f50c8f486b4e456336366e7acbdf87`, tree `d5f02c936cf080666a8e73bb5f253e52086705f4`, ordinary parent88c9408e. All seven existing planned-path hashes match the ledger; the eighth path is correctly prospective. Actual merged Event-parent reanchor remains required.

## N1 — invalid installed update-where body

Location: ledger `installed.requests[1].body`; plan “Fresh installed witness” body-byte statement. Request2 currently sends `propertiesChange` to `operations/update-where`.

Source: transaction-batch.ts187–210 passes the singleton's actual fields into the mixed set parser; transaction-set.ts81–90 permits/requires `change`, rejects extra fields, and only later translates `input.change` into a Resource singleton's `propertiesChange` at121. The normative closed set schema is docs/specs/bdp.md4867–4880. Existing actual-wire test transaction-http.test.ts2290–2305 uses `change`.

Impact: request2 is rejected before a completed receipt/page exists. Its expected200/pos2, page grant, remaining receipt checks and14-response witness cannot succeed. This is not a paging-capacity or production-serving defect.

Smallest correction: replace only request2's field name `propertiesChange` with `change`; retain Selector, cardinality32, patch, routes, keys, order and request count. Independently recomputed exact resulting body length179B (formerly189), SHA-256 `ba388a9fdbed4eae2e5221a425aa476720d2307d22553870faccf5403648cb2d`; aggregate body bytes3999 (formerly4009); maximum remains3728. Update the plan/ledger binding in a preserved successor after the ongoing review. No source or current-plan edit was made here.

## Verified repair and privacy boundaries

- docs/specs/bdp.md4632–4633 and4650–4656 explicitly cover receipt/page nondisclosure and authentication → nondisclosure → minimum → expiry ordering. Lines5798–5804/5830–5835 supply the current observed Scope-context obligation. The plan uses that existing law and the current all-visible owner policy without inventing receipt visibility or resolving pending Snapshot/incident-Link decisions.
- Root owner readReceipt at transaction-store.ts854–884 already returns a frozen absence/context observation in one entry/transaction. The proposed service helper change retains that tuple rather than rereading the owner. The supported minimum path and unsupported nondisclosure path are both accounted for; visible unsupported behavior remains unchanged.
- Page readReceiptPage at2008–2089 has two distinct early namespace defenses (marker and parent-prefix), followed by actual saved-parent/epoch/token observations. Keeping those private defense results missing-only while adding readContext to the three observed-absence branches preserves the existing lookup and integrity boundaries. readContext failure must remain a fault, as the plan explicitly states. Genuine missing results return before minimum waits or retention disposal.
- Service receiptPage429–468 and root receipt/receiptAtMinimum470–497/570–583 are the necessary typed projection seams; HTTP532/544 are the two missing final projections.401 remains separate/contextless. No new owner read, public API, enumeration, schema or runtime setting is needed.
- Repo-wide `readReceiptPage`, `receiptPage`, `receiptAtMinimum` and result-type searches found the documented production caller chain. Existing successful page consumers in recovery/owned/set tests do not require shape changes. The one strict root hidden-result object in service.test.ts237 and four stale HTTP header loops at328–359,2028–2088,2150–2181,2184–2214 are explicitly owned. Other inspected root404 assertions are partial. Namespace-defense exact-shape/no-query regression280–290 remains valid.

## Witness, selection and count checks

The paging strategy is otherwise source-grounded: store default maximum32 at224/281, page-plan slicing723–739, and transaction-batch.ts890–906 append one matched summary followed by one result for each selected Resource. A real32-match update therefore has33 results,32 inline and one tail. Existing installed set witness205–252 independently demonstrates this page construction. The fixture starts a new store; the32 unique createBead records use valid batch discriminator/type/properties shapes, and the later create-bead singleton correctly omits its discriminator. Both singleton bodies remain keyless; all three distinct keys are HTTP headers. Literal descriptor/schema IDs and the required numeric n property align. The remaining request shapes and bounded canonical id/next authority have no identified issue.

Verified the ledger's existing callback titles against their source registrations and enclosing describe names, including the false/true service expansion; new titles remain prospective literals. All three anchored selectors match the intended ledger names (3 store,5 service,6 HTTP). Source HTTP counts independently expand to7+18+10+17+3 existing plus13 new =68. Owner/directory totals account for the two reopen lifetimes. The14 source cases, one13-request control, one14-request installed case reconcile to16 executions/95HTTP/18 owner opens/16directories/8listeners/one product child/two readonly inventory handles. These are prospective source-derived bounds, not actual execution measurements.

The new store14 page calls and service13 calls are internally consistent with their described operations. The control's eight named failures correspond to six first-head404 GET/HEAD observations plus two later-head404 GET observations;401 and all prerequisite status/body/schema checks must remain successful. Dynamic emitted runtime import exists at transaction-http.test.ts14–16, so the required fresh build and derivative relative-module/copy binding are material. No extra installed mutant is needed.

The14 installed route/status entries and existing body hashes were independently recounted; descriptor/schema content totals328B. Current literal body total4009 is arithmetically correct for the invalid body, and becomes3999 under N1. Response cap32768, request caps,45s cohort/60s test/5s exchanges and120s external command bound are unchanged prospective limits. Actual schema/body sizing, discovered test names, cleanup and stage outcomes remain measured-draft/execution gates; this review does not infer them from static arithmetic.

The planned full Problem assertions before soft tuple assertions, bounded response consumption, redirect prohibition, primary-plus-cleanup error preservation, no automatic retry, independent final PID/store checks and genuine source/archive/payload binding are appropriate. Safe reports do not recover full bodies from fingerprints, and an externally supplied source label alone is not attestation.

## Disposition

Correct N1 in the plan/ledger successor; retain all other proposed scope and counts. After council acceptance, mechanically reanchor to the actual merged Event parent before authoring the eight paths. No source edits, tests, imports, builds, installs, services, new reviews, or outward actions occurred in this audit. Public qualification and authoring release remain outside this conclusion.
