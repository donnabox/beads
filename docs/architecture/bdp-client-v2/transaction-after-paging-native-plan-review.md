# Native plan review: atomic sets after receipt paging

Read-only review of `transaction-after-paging-plan.md` against frozen `0ff180a531b3c9ae594e0a104321e80744de38ed` and its normative `docs/specs/bdp.md`/schema. Read the current transaction preparation, normalization, evaluator, owner, and Read Selector implementation. No source changes, tests, external writes or qualification claims.

## Verdict

**Proceed with the independent preparation/identity/staged-execution core.** Both set verbs belong in one bounded slice: they share selection, cardinality, expansion and atomic ownership. The plan correctly excludes sets from sequences, retains the shared Resource evaluator, and requires actual HTTP/paging/restart evidence.

**Do not mark the negative HTTP boundary complete until Selector failure timing is clarified.** The frozen text does not determine one complete mapping for grammar-invalid/unsupported Selectors and `selector.bytes/depth/nodes` limits. This is a narrow wire/admission decision, not a reason to block the settled core or redesign the error model.

## Priority correction: Selector admission must distinguish explicit law from inference

### P2 — Root must supply an existing ruling or approve a narrow clarification before completing negative routes

Plan lines 176–191 correctly identify the issue, but its recommended mapping must remain a proposal:

| Question | Explicit existing law | What is not explicit |
| --- | --- | --- |
| Invalid/out-of-profile grammar | Selection defines the supported grammar (`bdp.md:849–863`). The schema only checks a nonempty string (`selector`). A syntactically invalid carrier never consults key state (`4490–4539`). | Whether invalid embedded Selector grammar is part of carrier syntax with `malformed-request`, or another direct refusal; no set-specific mapping is stated. Pre-key `malformed-request` is coherent, but inferred. |
| Selector complexity | Limits may exist; exceeding one cannot partially mutate (`4275–4278`). Limits describe decoded bytes and parsed structure (`1065–1066`). | The six pre-admission steps do not name Selector complexity. The closed receipt-code prose defines `limit-exceeded` through transaction work limits (`5317–5319`), without `selector.bytes/depth/nodes`. It does not explicitly place those three limits in a phase. |
| After admission | Every nontransient execution failure is retained in a failed receipt; transient retraction is for genuine execution aborts (`4526–4539`). Receipt codes exclude `invalid-parameter`/`malformed-request`. | These rules do not license inserting either code into a receipt, treating permanent grammar failure as transient, or inventing a new code. |
| Equal retry | A syntactically valid equal retained duplicate returns its receipt after negotiation, before fresh execution-capacity controls (`4508–4524`). | A parser that applies operational Selector limits during preparation could accidentally reject such a retry before comparison. |

Read's `readControlProblem` maps grammar to `invalid-parameter` and parser limits to `limit-exceeded`, but it is implementation evidence for a different carrier, not authority to copy the mapping into set receipts.

Recommended clarification, subject to root's ruling: grammar-invalid/out-of-profile Selector text is carrier syntax, rejected before key lookup with direct `malformed-request`; operational Selector limits are distinguished from grammar and given an explicit phase and existing `limit-exceeded` framing. If the chosen phase is operation execution after admission, explicitly include these Selector limits in the receipt prose, with the outer operationIndex and request pointer. If root chooses pre-admission enforcement instead, its position relative to principal/key/negotiation must be stated, preserving the retained-duplicate law. No new error code or wider receipt union is needed in either approach.

The ambiguity cannot be uniquely resolved from the inspected frozen text. It requires a narrow normative decision for complete serving behavior, **not an additional product decision about whether to implement sets**. An existing authoritative ruling would suffice. The core implementation can continue while root resolves it.

### Parsing consequence to make explicit

`selectReadResources` combines source-byte checking, grammar parsing and AST depth/node checks before evaluating candidates. Its parser can hit a structural limit before it reaches a later grammar error. Do not simply invoke it with execution caps during pre-key preparation and call every thrown error malformed. Separate bounded grammar validation from operational-cap evaluation, or retain a prepared parse outcome sufficient to implement the approved precedence. No generic parser framework is required. Any shared helper extraction remains root-coordinated.

Test the eventual decision against an invalid later member, over-complex valid Selector, invalid suffix after a complex prefix, unknown/conflicting/equal/pending keys, and unacceptable Accept; inspect actual receipt/key/effect state. For JSON-body selectors, count UTF-8 bytes of the JSON-decoded string; do not percent-decode its literal text as if it were a GET parameter.

## Concrete implementation guardrails

1. **Preparation and identity:** retain exact Selector strings, lossless numeric values and existing provenance defaults. Numeric equality applies to actual numeric JSON members, not numeric-looking text inside the Selector. Reject `cardinality:{}` (schema `minProperties:1`), negative/fractional bounds and set `name`; do not introduce a safe-integer restriction or a `min <= max` syntax rule. Preserve original outer indices across mixed Resource/set members, including backward labels around intervening sets and wrong-kind/forward labels. No set alias witnesses and no permissive unknown-kind fallback.
2. **Staged lifetime:** enumeration must be a capability guarded by the same active transaction lifetime as `resource`, `putResource` and `deleteResource`. Materialize canonical visible candidates at the operation turn, fix the selected IDs before mutation, then use the shared Resource path. Add a retained-capability-after-evaluation rejection test. Do not let a second owner or public Read snapshot determine matches.
3. **Failures and diagnostics:** use the outer set index on operation-local failures. Synthetic singleton inputs must not leak fabricated `/propertiesChange` or inner operation locations in place of submitted `/change` and outer `/operations/N` pointers. Unapplied patch numeric diagnostics retain the existing optional document-location behavior. Do not confuse a request pointer with a postimage pointer.
4. **Event cap:** enforce one finite transaction-total cap at the common fact-recording path, counting actual emitted facts—including each Link endpoint fact, both roles of a self-Link, and ordinary operations mixed with sets. No-op and matched entries are not Events. A transaction-wide cap failure omits operationIndex, retains the failed receipt and rolls back all staged identities/effects/pages. This cap does not bound an all-noop selected population, scanning cost or retained result bytes; the plan honestly leaves that workload qualification open.
5. **Configuration and compatibility:** use the existing private-format refusal path rather than relying on an old decoder eventually encountering an unfamiliar witness. Old owners otherwise run startup recovery before they need to interpret completed witnesses. Include any configurable expansion/Selector policy in the relevant immutable installation snapshot, and prove changed/old format refusal before maintenance byte-identically. No automatic migration or reseed. The exact private format name is an implementation choice.
6. **Actual HTTP acceptance:** retain the planned real singleton/batch routes, schema-valid matched-plus-expanded receipts, zero-match/all-noop behavior, cardinality retained replay after population changes, complete Link incident failures, at least three pages counting matched, stable Scope across listener restart, and atomic fault controls. These are substantially stronger than testing the pure Selector alone. Keep sets rejected from sequences before member keys or effects.

## Safe independent work now

Preparation for closed shapes and lossless identity; strict new witness variants; mixed-label/index controls; lifetime-guarded staged enumeration; valid bounded Selector evaluation; inclusive cardinality; shared Resource expansion/facts; the finite Event cap; full internal receipts and durable pages; rollback/recovery and compatibility tests. Keep unfinished negative routes explicitly private/unqualified until their narrow admission decision is applied. No need to touch Update's W1 work, Types, schema, public admission, historical evidence, Snapshot/changefeed, or unresolved Read minimum questions.
