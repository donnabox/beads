# Decision proposal: set Selector errors and admission

Status: proposed, not a ruling or implementation authority. This supports the sole design.md plan. Frozen evidence: BDP7a768c160eecd6c698c934a47492a5943359c50d; native after-paging review. Product decision owner: Donna. Safe staged-set core work continues independently.

## Decision needed

The spec defines the Selector grammar and operational limits, and orders carrier syntax → principal → idempotency key → fresh capacity → response negotiation → admission. It does not place embedded set-Selector grammar or bytes/depth/nodes failures in that order. Read query errors cannot simply be copied into Transactional receipts: the receipt's closed codes exclude invalid-parameter and malformed-request. The current parser combines grammar and complexity checks, so implementation would otherwise choose wire behavior accidentally.

Evidence: Selection825–889; advertised limits1055–1073; set limits4264–4285; admission4490–4539; receipt codes5308–5325. These are references to the pinned source, not permanent line anchors. Native review identifies this as a real normative gap. Native proposal audit and external plan review both confirm the missing phase rule; no existing ruling was found in the current spec or the fork's design records.

## Recommended contract

1. Invalid or out-of-profile Selector grammar is carrier syntax: direct400 malformed-request before principal/key lookup, with a pointer to the actual selector member. It creates no receipt or key binding.
2. A valid Selector exceeding configured bytes/depth/nodes is a request-complexity refusal for a fresh key: direct413 limit-exceeded in the existing fresh-capacity step, after key comparison and before negotiation/admission. It names selector.bytes/depth/nodes and the actual request pointer. No key or receipt is created.
3. An equal retained/pending retry follows existing receipt/comparison/negotiation rules before those current operational limits. It never reselects or consumes fresh execution capacity. A conflicting key retains its existing precedence.
4. Data-dependent execution limits (examined/matched/mutated Resources, induced Events and duration) retain the existing failed-receipt semantics after admission. No new error codes or receipt-union variants.
5. Selector bytes count the UTF-8 decoded string: percent decoding for a query value, JSON decoding for a body string, never percent-decoding percent characters inside a JSON string. Clarify the existing Read-centric wording rather than silently rewriting set literals.

Why recommend this: syntax is invalid regardless of stored state; static complexity is a capacity decision like transaction.operations and should not consume an idempotency key. Existing outcomes remain retrievable if Selector operational limits change, subject to the existing step1 request-body bounds and current authentication; this does not promise retrieval through a newly smaller transport bound. This requires explicitly separating grammar validation from operational complexity checks; grammar checking remains bounded by the overall request-body admission bound and must be stack-safe. No general expression-language expansion is implied.

## Alternative

Keep the same syntax rule, but enforce Selector complexity during execution after durable admission, storing a failed receipt with limit-exceeded. This requires explicitly adding Selector limits to the receipt-code prose and defining their operationIndex/pointer. It gives clients a retained failure but consumes the key while that record is retained; a new attempt after changing limits needs a new key until the old record is forgotten under the existing retention rules. It is implementable with existing codes, but unlike transaction.operations it treats a static request bound as an execution failure.

## Error precedence and client consequences

The recommended option checks all embedded Selectors before execution. An overcomplex Selector in operation2 therefore returns direct413 even if operation1 would have failed during execution. The failed-receipt alternative instead records operation1’s earlier execution failure without reaching operation2. Both are implementable; this difference must be explicit.

The existing limit-exceeded problem keeps retry:never. An unbound key does not tell the client to retry the identical request immediately; the request or available capacity must change. Equal retained/pending outcomes bypass current Selector caps only, not earlier request framing, body bounds or authentication. The current parser combines grammar with bytes/AST limits and must not be reused unchanged to turn every failure into malformed-request.

## Concrete normative edits after a ruling

- Selection/set input section: classify unsupported/invalid embedded grammar and preserve literal identity.
- Transactional admission step2: add invalid set-Selector grammar with pointer guidance.
- Step5: for the recommended option, add Selector bytes/depth/nodes to fresh-key capacity refusals; state no receipt/key effects and preserve equal-duplicate precedence.
- Advertised limits: distinguish URL query decoding from JSON body decoding for Selector bytes.
- If the alternative is chosen, instead extend existing failed-receipt limit prose to include Selector complexity and its operation-local framing.
- Add actual HTTP negative cases for unknown/conflicting/equal/pending keys, unacceptable Accept, valid overcomplex inputs, invalid suffix after a complex prefix, exact boundary versus boundary+1, and no-key/no-effect versus retained-failure state. Keep Read behavior unchanged.

Neither option is enacted by this document. Public set routes stay gated until the decision and its tests are integrated. The atomic valid-input core, finite Event cap, receipts, rollback/recovery and owned-Link conformance continue.
