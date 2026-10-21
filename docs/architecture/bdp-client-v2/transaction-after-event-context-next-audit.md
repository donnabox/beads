# Next bounded work after observed Event404 context

**Recommend another concrete serving repair before resuming the deferred sequence witness: retain the same-entry current context on authenticated, nondisclosable Mutation Receipt root/page404 responses.** This is a source-established gap in the existing contract. It requires no new minimum precedence or Event-membership ruling. This audit does not authorize implementation or claim an executed reproduction.

## Source and search boundary

Actual main basis is88c9408e908590ab1a4c64c9b9dee1a8b705999c/treeb2d9f2c18502e0778c51123be8b4564af052d803 at /Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence, plus the frozen Event repair bindingacf190e742dfb669fe5607085a3ac60c5ecb66f4f86b1f05025424a7873615b0/fullpatche0c25b193ef1191d10916728c54eb6a894c5f29992e0da768bd585ef089e98e3. This is not yet a claim that the Event successor is merged.

Read the sole design at bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md:7–31 and current live ledger:69–94, the prior transaction-after-sequence-serving-next-audit.md, current progress/discovery/HTTP/Event/receipt store and service paths, relevant receipt tests, normative receipt/consistency clauses and current cumulative catalog/coverage. The search was bounded to outstanding serving seams and their immediate evidence; no repeat audit of completed alias, Selector, command-minimum, Snapshot-creation, schema, SSE, comparison or cross-carrier fixes. No tests, services, build/setup/install, product imports, new Claude review or outward mutation occurred. Only this artifact is written through the session guard.

## Concrete failed scenario and established law

A trusted authenticated client requests a canonical receipt URL belonging to another principal, or a canonical unknown receipt URL, using GET or HEAD. With no minimum fields, the current implementation returns the required404/resource-not-found but omits all three current Scope response fields. A well-shaped unknown page under a canonical receipt root has the analogous missing-context path.

- docs/specs/bdp.md:4632–4633 explicitly lists unknown/foreign/retracted/forgotten/prior-epoch receipt404 and the same page nondisclosure rule. Lines4650–4656 settle authentication → principal/epoch nondisclosure → minimum → page expiry → representation ordering. That receipt-specific precedence is already decided.
- docs/specs/bdp.md:5116–5133 defines principal-bound receipt URLs, uniform404 and the namespace-root/nonlisting distinction.
- docs/specs/bdp.md:5798–5804 requires the Transactional Scope response fields;5830–5835 requires every Scope-bounded read to report its observed projected-prefix context. The successful-receipt sentence does not exempt observed404 reads from that general law.
- Current private configuration is fixed/all-visible; the same permitted Scope tuple on unknown and foreign-principal receipt404 says nothing about which receipt exists. Do not return receipt-body fields, another principal's receipt facts or a receipt's old required/effect position as the current observation.

This does not ask whether missing Resource reads beat foreign/behind minima generally. Receipt nondisclosure ordering has its own express rule, and the primary demonstration needs no minimum at all.

## Exact loss points

| Path | Actual source | Smallest repair seam |
|---|---|---|
| Receipt root owner | transaction-store.ts:854–884 defines ReceiptRead with context and already returns frozen {receipt:undefined,context} from the same entered atomic observation, including absence. | Preserve this actual observation; no extra HTTP-side or unrelated owner read is needed. |
| Receipt root service | transaction-service.ts:50–53 models404 without context; helper:202–210 discards the entire missing observation by returning undefined. Ordinary receipt:575–581 and canonical/unsupported minimum paths:476–497 reconstruct contextless404. | Carry the existing owner observation through only authenticated missing-result variants. Keep401 distinct/contextless; preserve supported/unsupported minimum ordering and existing receipt checks. |
| Receipt page owner/service | transaction-store.ts:2019–2057 returns missing-only for absent/foreign/unknown-page cases before its later readContext. transaction-service.ts:441–466 maps missing to contextless refused404. | For actual canonical page absence, capture immutable current context inside the existing owner entry and forward it. Inventory namespace/syntactic defense branches rather than blindly broadening all malformed/pre-owner refusals. Preserve nondisclosure-before-minimum/expiry and fail through existing integrity handling if context observation faults. |
| HTTP | transaction-http.ts:532 and:544 call transactionHttpProblem(code) on page/root refusals, dropping any possibility of supplied context. | Pass the observed404 context to the existing projector; preserve body/status/retry/cache and HEAD semantics. No routing, discovery or admission change. |

The frozen Event repair addresses its own store/HTTP branches (:1567/:1653 and HTTP:499/:514); it does not alter these receipt paths. The gap is independent, not an excuse to reopen that accepted repair.

Existing regression evidence also explains the miss: transaction-http.test.ts:328–359 explicitly checks context presence only for410 in a table that includes authenticated foreign receipt404. That stale expectation must become status401-only contextless, while the same404/body/nondisclosure precedence remains. transaction-receipt-pages.test.ts:280–290 pins missing-only shape for namespace defense cases; treat that as an explicit shape/fanout review input, not permission to delete the guard. Source shape consumers need a bounded inventory before assigning paths or test counts.

## Bounded next slice

Have the author prepare one root/page-context plan on the actual merged Event successor, with native and actual Claude review. Production seams are store (page missing observation only), service (preserve root/page absence context), and HTTP (projection); exact tests/docs ownership follows an actual caller inventory.

Required oracles should use the existing receipt fixtures: a real authenticated missing root and foreign-principal receipt share404 and the same current tuple; advance an independent committed head and show that current context changes while the404 body stays uniform; root/page GET and HEAD retain the same representation metadata with no HEAD body; anonymous401 remains pre-owner/contextless. A genuine paged receipt supplies its actual page URL for the foreign-principal case—do not invent client receipt-follow authority or silently add a page-size seam. Preserve the explicit incompatible/unsupported minimum and nonexistent-page-before-retention tests. Do not expand into authority restore/erasure merely to generate every nondisclosure category.

Use an emitted projection-omission negative only after exact source consumer/module and callback counts are measured; require the specific context failure with correct404 prerequisites. Reuse the useful lessons of the Event control, including preserving real failed attempts and distinguishing package shim normalization from source drift. A fresh packed installed witness remains separately bound and budgeted; this audit supplies no guessed request/child allowance or execution release.

## Deferred work and eligibility limits

- Snapshot/collection-continuation S1-Q1 remains the pending anchor-versus-current-serving minimum decision (HTTP:406–411 and saved decision audit). Fresh Snapshot minima are completed. Do not infer the pending answer from receipts.
- Metadata-only incident-Link Event membership G2-Q1 remains explicit at transaction-events.ts:257–264; this receipt slice does not encounter or resolve it.
- The sequence-member witness remains deferred, not rejected or ready to author. Existing requirement transactional.sequence.member-transactions (catalog/transactional-v1.json:866) still lacks a cumulative executable row. After serving repair, the precise existing evidence follow-up is the reviewed ten-action success/failure/success + receipt recovery + replay + complete Scope Event witness, mechanically reanchored and corrected by accepted A dispositions. Recovery keys belong in HTTP idempotency-key headers; batch body keys are rejected at transaction-batch.ts:119, while HTTP:295 forbids a sequence outer key. Outstanding B review/workload concerns remain; do not silently relabel the deferred plan as accepted implementation. Preserve full equality, seven exact negative failures, parent reconstruction, old sentinels and measured recurring workload limits.
- Current cumulative coverage declares claimEligible:false and every row wholeContract:false (coverage-v1.json; transactional-cumulative.test.ts:107–113,194). A new witness would provide real installed/configuration-bound observations for a missing required slice, not certify atomic all-member admission, pending ownership, expiration/reauthorization or the entire cumulative profile.
- Cold discovery is still deliberately refused at transaction-http.ts:195/:228/:562–569 despite the existing internal builder/minimum seam. Activating it now would advertise an incomplete profile. Root's cumulative Read+Update+Transactional evidence and truthful-advertisement gate remains intact.

The receipt-context repair is the recommended next concrete serving increment. The sequence evidence lane is the subsequent existing-contract path toward discovery eligibility once its review/reanchor is complete; neither substitutes for pending decisions or whole-profile proof.
