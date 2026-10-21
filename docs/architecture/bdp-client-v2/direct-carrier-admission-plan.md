# Next bounded slice: direct-carrier admission ordering

Authorized bounded implementation from reviewed combined b7eefc9. Root accepted the scope and assigned an isolated feature branch after publication. The fork design document remains plan of record. No new mechanism or normative rewrite is proposed.

## Accepted contract and present gap

Current `docs/specs/bdp.md` lines 4443–4485 prescribe: request byte bounds; carrier syntax; principal; semantic key comparison; unknown-key admission controls; response negotiation. A conflict therefore wins over unacceptable Accept. An equal duplicate negotiates its receipt before capacity controls. A new unacceptable request creates neither receipt nor key binding. Lines 3955–3985 preserve one finite comparison budget while an unresolved dependency becomes comparable or retracts. A pending receipt cannot be disclosed until equality is established. Lines 4517ff distinguish an admitted pending outcome from direct refusals.

The current private `admit()` immediately inserts an unknown key. Calling it and then rejecting Accept would violate no-admission-on-406; putting Accept before the call would lose conflict precedence. `prepareTransaction()` also rejects >32 operations as syntax, while a configured transaction.operations refusal belongs after existing-key comparison. These are receiving/admission composition gaps, not reasons to introduce another key owner.

## Concrete implementation seam

1. In `transaction-batch.ts`, keep lossless decoding, body bounds, closed operation records, label/reference checks and explicit development unsupported-operation restrictions before key access. Separate syntactically valid operation count from the current private maximum-operations admission decision. Keep an independently bounded body/depth parse; this is not permission for unbounded input. Do not interpret unsupported set operations as a protocol-wide syntax ruling.
2. Extend the same atomic owner decision in `transaction-store.ts` with immutable direct-submission controls: acceptable response representation and the private operation-count bound. Preserve default behavior of existing trusted private callers if needed. On an established existing key, compare first and never apply new-request count controls. On an unknown key, check count before negotiation; only an accepted request inserts the pending receipt and owner. Return internal refusal decisions without storing them. Do not run arbitrary external callbacks or HTTP operations inside SQLite. Do not split lookup and insertion into separately authoritative decisions.
3. Thread those same controls through `transaction-comparison.ts`. An unresolved possible match still waits under its original finite deadline; a proven conflict remains immediate. Proven equality can then produce a receipt or negotiation refusal. Retraction permits fresh admission only through the same owner gate with the same deadline. Every attempt reuses the serving request's controls, never inherits another owner's claim.
4. A small private `transaction-submission.ts` maps these receiving/admission decisions for the bounded direct singleton and batch carriers. It accepts a trusted authentication result, not a raw principal asserted by an eventual remote client. The current fixed all-visible policy stays explicit. Preserve malformed-before-key, principal-before-key, conflict-before-negotiation and unknown-limit-before-negotiation ordering. Return transport-independent status/problem/receipt decisions; do not create an HTTP listener or a public API. Distinguish permanent admitted failure receipt from pre-admission refusal and internal/fenced owner faults. No new rate-limiter or admission policy framework is needed for this slice.

Files: transaction-batch.ts, transaction-store.ts, transaction-comparison.ts, new transaction-submission.ts and focused tests; transaction implementation progress note and installed-module smoke allowlist. Shared evaluators/RU modules, schema/spec, exports, catalog admission, Type policies and existing development receiver stay untouched.

## Executable done-when

- Invalid syntax/labels/key never consults key state; rejected principal never consults key state. Body-bound failure precedes parsing.
- Known conflict + unacceptable Accept returns conflict; exact retained/pending/expired duplicate + unacceptable Accept returns bodyless-negotiation decision with unchanged identity/disposition and no detail-renewal.
- Equal duplicate can replay under a now-smaller private operation-count bound. The same oversized request with an unknown key receives direct limit-exceeded/transaction.operations and no receipt. Valid unknown request with unacceptable Accept leaves the key unknown; a later acceptable presentation can acquire it.
- Unknown limit refusal wins over unacceptable Accept. No compensated admit-then-delete is used to emulate rejection.
- Singleton and equivalent one-op batch keep one namespace, numeric/default/context/alias-witness equality and original allocation/receipt identity.
- Unresolved dependency: provable mismatch immediate; possible match no receipt until equality; equality then negotiation; timeout no admission; retraction fresh gate under original deadline; owner change never resets deadline; epoch change stops the attempt. Sequence projections remain nonwaiting and unchanged.
- Concurrent presentations and injected admission failure preserve one atomic owner and no split keys. Existing commit/rollback fencing, original-cause chain, format refusal, receipt retention and dependency tests pass.
- Build/typecheck/boundaries/format/lint, focused tests, installed smoke and unchanged Read evidence pass. New private module remains denied as a package deep import.

## Explicit boundary after this slice

HTTP wire details (raw duplicate headers, target/media checks, socket bounds, actual 200/202 timing, GET/HEAD receipt conditional semantics and current observation headers) still require a receiver. A receiver also requires same-owner authoritative Resource/alias reads and live-claim lifecycle ownership; existing `inspect()` and the parallel RU development owner are not that read authority. This slice cannot qualify those wire rows.

Sequence Resource-alias first lookup remains a separate representation clarification and implementation task; no alias wording is changed here. Fixed local authentication, no policy changes, no erasure and no receipt pagination remain private limits, not a full Transactional profile. The actual diagnostic harness should gain Transactional observations only when a real receiver exists; static HTTP examples and private owner tests are not executable endpoint conformance.
