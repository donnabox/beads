# Sequence refusal plan v2 — independent native closure

Disposition: **closed for plan correction review; no blocking or Low findings remain from this bounded review.** All ten external comments and both native comments are corrected or explicitly dispositioned with a source-grounded justification. This is an independent review recommendation, not root acceptance, assignment of a parent, permission to author, or an execution result.

## Reviewed identity and scope

- Revised plan: `transaction-after-direct-refusal-serving-plan-v2.md`, SHA256 `ea8f0a3b9decc2613c45a8966fc49db8f94cdbdbae782357073676583cbc5b36`.
- Revision dispositions: `transaction-sequence-refusal-v2-dispositions.md`, SHA256 `10327a9989db66dbfe9c5491d0d1a14319c91c66c90f99820058b9e252db3896`.
- Read root instructions in `transaction-sequence-refusal-root-plan-dispositions.md`, the completed `transaction-sequence-refusal-claude-plan-review.txt`, and original `transaction-sequence-refusal-native-plan-review.md`.
- Original plan remains preserved at SHA256 `feac0983d3da85dd0b93892b2cf5b1c731f7d67864c49dc416e1f4721ac63306`.
- Actual source inspected read-only: `/Users/dbox/repos/wt/vickie/vickie-direct-refusal-review-repair`, clean HEAD `ed3597da1179d4d333ef89c5f0b55adf230be654`, tree `7544b53aa8aa6d7edae94bd6bf0a4d157ee55985`.

This closure checks the corrections and their consistency with the actual source; it does not reopen the wider profile audit. No tests, builds, services, installation, network actions, or source edits were performed. Only this review artifact was written.

## Comment closure

| Item | Result and evidence in v2 |
| --- | --- |
| External C1 | Closed. Lines 33–42 distinguish the general Scope-response obligation at spec5768 from the negotiation ordering clause and the conditional-response analogy. Applying the general rule to these authorized known-Scope 413/406 responses is explicitly a derivation. Root's rejection of a new-ruling requirement is preserved; no new semantics are invented. |
| External C2 | Closed. Installed requests 10 and 11 repeat both selected refusals with fresh keys at pos-4; requests 12–13 then check complete unchanged graph/Events. Source one-entry and immutable-context checks remain independent, and the plan does not overclaim that finite observations exclude every cache implementation. |
| External C3 | Closed with the explicit root disposition. Line 52 acknowledges that introducing a required current observation can expose an unexpected owner/service fault as sanitized bodyless 500. It cites the existing spec4588–4592 exception, preserves normal closing/not-ready 503, and expressly rejects universal installed-500 unreachability. No runtime edit or new fault policy is proposed. |
| External C4 | Closed. Receipt5233–5244 and minimum-read5783–5786 are supporting, narrower precedents. They do not replace5768 as authority for the selected refusals or import minimum semantics. |
| External C5 | Closed. Line 69 requires all three headers absent and zero new observation calls for carrier-key400, malformed400,401,403,415 and the existing minimum development500, retaining its marker. Line 72 adds a named absent-header over-wrap control. These assertions preserve the selected source scope without declaring a universal normative exemption. |
| External C6 | Closed. Line 70 separates receipt storage from graph inspection, requires exact baseline receipt/key inventory and all refused keys absent, and places the SQLite inspection after every live observation. The fixture's close behavior is acknowledged. Different-identity reuse is a separate case, avoiding a contradictory later absence claim for deliberately admitted keys. |
| External C7 | Explicitly and appropriately deferred. Line 48 preserves the unrelated discovery closure byte-for-byte. Existing owner readContext remains the single context authority; sharing the discovery closure would also involve its minimum/wait/diagnostic behavior. The narrow refusal wrapper and optional same-position comparison meet this slice without that refactor. |
| External C8 | Closed. Lines 69 and 93 require canonical limit-exceeded plus sequence.operations, and distinguish body65536/header16384/response131072 caps from target8192. Measurement and root release precede execution; no cap increase is proposed. |
| External C9 | Closed. Line 23 identifies first direct sequence over-count coverage and its existing schema/parser reachability. It does not mistake transaction.operations coverage for this branch or introduce schema/parser/catalog changes. |
| External C10 | Closed. File six is mandatory: scope the Snapshot context sentence at progress727 and distinguish successful sequence post-runner/replay discussion at615–617 from a nonexecuting refusal observation that can fault. |
| Native N1 | Closed. Line 91 explicitly states the recurring whole-suite positive cost of13 requests/one child and keeps the predecessor negative separately selected. Build precedes emitted consumers, or exact installed-entry/source overrides are supplied. |
| Native N2 | Closed with C6. Receipt/key storage is not inferred from inspect(), unchanged position, or Resource state. Any principal/key projection is confined to existing test fixture code; no production inspection API is added. |

## Source, budget, and scope cross-check

The actual sequence branch in `packages/server/src/transaction-http.ts:303–328` still authenticates/authorizes before the count refusal, returns `limit-exceeded` with `limit: sequence.operations` at the count branch, prepares supported transactions afterward, and returns empty406 before execution. V2 preserves that order. Its narrow service observation uses existing finite-work registration and one synchronous owner read; it neither admits keys nor executes members, and releases the owner entry before transport delivery.

Actual `docs/specs/bdp.md:5768–5774` states the three-field Scope-bounded response rule. Actual4588–4592 includes the shared bodyless500 for an unexpected internal fault anywhere. Actual `transaction-runtime.ts:115–118` checks closing or missing handler and returns temporarily-unavailable before dispatch. That gate supports the normal closing503 statement only; it cannot establish that a live handler never encounters an unexpected owner/read fault. V2 expresses precisely that limit and does not use successful-sequence durable replay as a refusal justification.

The receipt oracle is implementable within the owned HTTP test file: existing `retainedReceipts()` closes the runtime and reads SQLite read-only, while `receiptRows(key)` closes and reopens. The current retained-receipt projection is receipt bodies, so the plan correctly allows a test-local principal/key projection for the required exact inventory. No later live call may follow the final closing oracle. The over-wrap control supplements the two primary source control categories; the omission category must reach413 and406 separately.

The installed ledger is internally exact: two seeds, two initial refusals, one historical receipt read, two finite collection/Event reads, two different-identity key-reuse commits, two repeated refusals, and two final finite reads total **13 requests/one child**. The unchanged direct-refusal regression adds **14/one**. The separately selected predecessor negative reaches the first refusal after two seeds, expected **3/one**. Thus author validation is **30 expected requests/three children**, with **40/three maximum** if the negative unexpectedly reaches all13 requests. The negative must establish the canonical413 code/limit before its named missing-context failure; exit1 alone is insufficient. No retry is silently budgeted. Fixed route authority, validated receipt URL, payload identities, counts, fingerprints versus full-wire distinctions, deadlines and cleanup remain explicit preflight/evidence requirements, not executed proof.

The maximum six owned files are service, service tests, HTTP, HTTP tests, one new installed test, and the mandatory progress document. No discovery extraction, store/evaluator/parser/schema/catalog/export/runtime change, public profile activation, or unrelated work is included. Source exclusions remain descriptive regressions and do not settle unresolved disclosure/minimum precedence.

Root acceptance, reviewed PR117 successor/main parent assignment, shared-file release and measured preflight remain pending. The plan cannot establish merge status, implementation correctness, installed success or whole-profile completion before those later gates and their evidence.
