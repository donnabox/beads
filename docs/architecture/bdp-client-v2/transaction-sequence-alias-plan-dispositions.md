# Transaction sequence alias plan — review dispositions

Original plan preserved byte-for-byte: `transaction-sequence-alias-plan.md`, SHA256 `72b247c74afd51b2a9842901345e9ea4d33a424b4ba5921016076bbad8b1a657`.

Revised plan: `transaction-sequence-alias-plan-v2.md`, SHA256 `575a2f104f3ece1412645099a7b6839ffa278c3a0f877f5d1b528d37314b492d`. Read-only design revision; no implementation, tests, public specification/schema changes, format changes, or outward writes. Exact authoring base remains root-assigned after C1.

Reviewed inputs: `transaction-sequence-alias-claude-plan-review.txt` (normal exit0) and `transaction-sequence-alias-native-plan-review.md`. All seven external findings and native P2 are addressed below.

## External findings

1. **F1 High — shared deferral fact/durable-row agreement: accepted.** §A publishes one immutable prepared alias-operand/creator-dependency fact. §B explicitly covers both store sites and uses the eligible owned durable row's non-null dependency as the requirement to normalize. Preparation and row must agree; malformed/inconsistent template is integrity failure. This durable-row inspection must occur after transient-creator precedence, so it cannot accidentally consult a dependent key before allowed. Alias-only and mixed tests plus the existing execute precondition prove unresolved state cannot reach execution.

2. **F2 High — alias-only templates may return equal: accepted invariant; example corrected.** §C seeds unresolved from original alias operands whenever dependency remains present. A compatible alias template returns pending, never equal; static contradiction returns conflict. Matching spelling is not permission to return a pending receipt. The review's predicted eventual conflict for the *same original locator* is not generally correct: after original turn capture the candidate reuses that witness and can be equal. A distinct fixed canonical target provides the actual eventual-conflict discriminator. Both cases are now required, with direct singleton and one-operation-batch callers.

3. **F3 Medium — forbid live lookup of original template locators: accepted and strengthened by native P2.** The prohibition applies to any candidate occurrence of any original unresolved canonical locator, across all eligible operand slots and equivalent spellings. External wording permitting lookup for the candidate's own slots is narrowed: only genuinely new candidate locators may resolve live. Original locators carry correlated symbols before masking/comparison. A throwing resolver unit witness verifies this rule. Retry-only lookups remain transient and are never owner witnesses.

4. **F4 Medium — additive pending-only compatibility: accepted with executable proof, no bump.** §D fixes format/identityPolicy/terminal codec as unchanged. New tagged state lives in dependency JSON only; identity/witness fields remain empty while unresolved. Startup exclusive ownership/configuration check precedes retraction of all owned pending rows and creator facts; retention/disclosure do not parse pending templates. Required maintain, crash/new-reader, real predecessor old-reader, terminal direct replay and exclusive-owner tests establish the claimed compatibility. Unknown tags fault. A failed proof returns to root; no migration or format bump is delegated. Old-reader test must execute actual predecessor emitted code over a disposable new-written image, not edit the image to fake compatibility.

5. **F5 Medium — operative law/slot anchors: accepted.** §Existing law now cites3062–3080 explicitly;2354 is identified as a forward pointer. §A enumerates updateBead.bead, deleteBead.bead, createLink.source/target including pinned URI forms, with Link-subject and putAlias-target exclusions. No widening of alias support is implied.

6. **F6 Low — asynchronous fallback: accepted and removed.** §B retains synchronous nextSequence from fenced normalization through executeClaim, exclusive SQLite ownership, reentrancy guard and existing fault stage. No await, alternate execution path or modification of the shared direct-carrier execution precondition is proposed. Crash between resolution and execution retracts pending work rather than resuming it.

7. **F7 Low — source ranges: accepted.** Preparation229–293/refusal264–268; complete nextSequence1791–1913; normalizer391–467. Creator-fact writeback for an alias-bearing createLink is explicitly included in the review/test scope. These are references to the inspected frozen runtime basis; implementation must reconcile shifts against the exact C1 descendant.

## Native P2 — symbolic candidate reuse and correlation

Accepted. §C and observation7 include the full required counterexample:

- Original pendingK creates work→work; current work=A.
- Direct candidateK creates work→B.
- Before original turn, work repoints toB.
- Candidate must remain unresolved, then compare equal to original B→B after turn; eager live A/B comparison would falsely conflict.

The swapped endpoint case and a two-original-locator swap ensure symbols follow locator identity, not slots. FixedA/fixedB against original work/work still conflicts immediately. Unification uses fixed canonical/new-locator facts, preserves pins and immutable structure, restricts alias-symbol domain to canonical Bead or its own locator-specific miss, and never clears alias unresolvedness from a satisfiable constraint. No general solver is proposed; the bounded operand set suffices.

## Ownership and validation disposition

C1 HTTP work goes first. Root assigns an exact frozen successor; alias may later implement preparation/dependency/store in an isolated checkout, with HTTP branch/test removal applied after ordinary integration and explicit ownership handoff. Cold discovery/composition remains separate. No source changes or tests were run for this plan revision. Existing bounded HTTP/retry/restart/RU checks remain in the plan, with new discriminatory concurrency and compatibility cases; no conformance whole-row claims follow automatically.
