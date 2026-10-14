# Native plan review: Transactional sequence Resource aliases

## Finding

**P2 — Specify symbolic treatment of original unresolved locators on the candidate side before applying correlation conflicts.** Location: `transaction-sequence-alias-plan.md:47–51`, and discriminator 7 at line 89. The plan extends the existing partial comparer, whose first step currently normalizes all candidate aliases by live lookup (`packages/server/src/transaction-dependencies.ts:49`), then checks correlations and masks only the original dependency slots (lines 53–79). That is insufficient for deferred original aliases: an occurrence of one of the original unresolved locators in the candidate must inherit that future witness, including when it occurs in a different slot. The plan's statement about two different resolved canonical targets is valid only when both targets are actually fixed comparison facts, not when one was prematurely obtained from an original unresolved locator.

Smallest counterexample:

1. `alias/work` currently points to A. Reserve a sequence member under key K: `createLink(source: alias/work, target: alias/work)`; its turn has not started.
2. A direct singleton or one-operation batch presents K with otherwise identical fields and `source: alias/work, target: beads/B`.
3. Before the original member's eligible turn, repoint work to B. The original executes B→B. Ordinary retained-witness comparison makes the direct request B→B too, so it is equal.

At step 2, normalizing the candidate through the current live lookup yields A/B. Applying the proposed same-original-alias correlation check to those values would incorrectly return conflict. Correct partial comparison is unresolved (`comparison-pending`), not equality and not conflict. The mismatch is not yet provable. This is a design gap in the planned extension, not a claim of a current runtime regression: current sequence preparation refuses these aliases.

**Minimal correction:** state that candidate references to any original unresolved canonical locator carry that locator's correlated unknown, in every eligible reference slot, before static comparison. Preserve fixed canonical references and genuinely new retry locators as comparison facts; do not save their lookups as owner witnesses. Propagate the existing creator facts/pins and the alias correlations sufficiently to reject only proven incompatibilities. Add the bounded scheduler case above, plus a cross-slot occurrence, to discriminator 7. The canonical A/B case with no unresolved original locator must still conflict immediately, and matching alias spellings must still never obtain a pending receipt before equality is established. This requires no new public semantics or comparison loop.

Authority: `docs/specs/bdp.md:3221–3241` preserves original locator targets/misses across equivalent spellings and reference slots, resolving only previously unrecorded locators afresh; `3994–4014` requires an already provable mismatch for immediate conflict and otherwise bounded comparison without handing out an unproven-equal pending receipt. The final normalizer already implements cross-slot witness reuse (`transaction-batch.ts:426–445`).

## Remaining review disposition

The proposed all-key reservation and eligible-turn normalization agree with the inspected contract. `admitSequence` performs one atomic reservation pass; existing keys are classified in their eligible turn, and `nextSequence` preserves transient creator precedence and foreign-pending projection (`transaction-store.ts:1720–1875`). The proposed overlay can reuse `normalizeTransaction`'s captured maps and execution consumption, avoiding a second alias lookup. No reason was found to introduce a new wire rule or change separate sequence commitment.

The plan correctly distinguishes sequence's nonwaiting projection from direct carrier waiting and preserves the existing single finite deadline, epoch handling and no-open-transaction wait (`transaction-comparison.ts:31–55`). The planned success/miss/repoint/retry/reopen cases cover the original-versus-new witness distinction; the finding above adds the missing pre-turn candidate distinction.

Compatibility remains an explicit implementation gate. An additive pending-only form appears feasible from inspected startup order: exclusive locking and configuration checks precede unconditional deletion of owner-bearing pending receipts and creator-fact cleanup (`transaction-store.ts:514–557`), before later comparison parsing. That is source evidence supporting the proposed option, not a completed old-reader compatibility proof. The implementation must retain the plan's old-reader/recovery test or select the private format bump with refusal-before-maintenance proof. Final terminal witness encoding should stay unchanged.

The four production files and four test files are a plausible bounded scope. The planned process crash, exact-owner cleanup, mixed creator/alias, all-key atomicity, malformed carrier, singleton/one-operation-batch duplicate and actual HTTP tests are meaningful. No conformance artifact, RU runtime, public discovery or profile claim change is needed for this slice. No additional speculative whole-profile tests are requested.

## Exact review scope and limits

Reviewed plan SHA256 `72b247c74afd51b2a9842901345e9ea4d33a424b4ba5921016076bbad8b1a657`. Inspected source checkout `/Users/dbox/repos/wt/vickie/vickie-schema-property-change`, HEAD `61ed530ac6a460598e0e312f5d28ec0e6fa2a33d`, tree `22e448fe0f517c0d6300b732de55494bd70c7938`, the supplied tree-equivalent context for main `581bad0de6f5eb098b920147f6a7e09809c553ea`; working tree clean. Read actual preparation, normalizer, dependency comparer, owner admission/turn/recovery, direct comparison loop, evaluator overlay consumption, existing sequence/dependency tests, and cited normative clauses. Commands were read-only git/rg/sed/nl and Python hashing. No tests, builds, network, services, credentials, agents or source writes. This is plan review only; implementation and qualification are not approved by this report.
