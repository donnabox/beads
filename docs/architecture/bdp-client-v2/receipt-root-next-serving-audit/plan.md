# Next bounded recommendation: opaque minima on live alias resolution

Recommendation only. Continue receipt-root review and validation first. This is a source-read diagnosis, not an executed reproduction, delivered fix, full plan, or profile qualification.

## Concrete trigger and current behavior

On a real seeded Transactional owner, create Bead `beads/a`, bind `alias/work` to it, then GET or HEAD that alias with the actual epoch/view and a lexically valid unrecognized minimum such as `unknown` or near-miss `pos-01`. The live locator currently returns private `500` with `x-bdp-development-limit: minimum-position`, rather than entering its existing finite waiter and eventually returning contextual `503 catch-up-timeout` if eligibility remains unproved.

This is not inferred from an empty fixture or a bare owner TypeError. The existing emitted HTTP callback `serves conditioned live aliases and preserves bodyless redirects, refusal context and receiving boundaries` at `packages/server/src/transaction-http.test.ts:3076` seeds two Beads and a real alias, proves normal conditioned redirects, then explicitly asserts this exact opaque-token 500 at lines 3150–3157. No test was executed for this audit.

Source chain, all paths relative to the bound receipt checkout:

- `packages/server/src/transaction-http.ts:213`: alias route captures lexical minima, calls `readAtMinimum(...).resolveAlias`, and constructs bodyless 307 only from a successful target result. There is no strict early parser obstacle here.
- `packages/server/src/transaction-service.ts:383` and `:403`: the tuple is captured; resolveAlias already runs through `waitForTransactionMinimum`, forwards the captured signal, and participates in the service lifetime.
- `packages/server/src/transaction-read.ts:333`: alias resolution creates the private condition with kind `missing`. The same kind is used by initial Bead incident-Link reads at `:100`.
- `packages/server/src/authority-read.ts:769`: live alias mapping, canonical target validation, projected target existence and its observation are computed inside the authority owner entry. Missing/deleted/dangling/hidden results carry that same observation. Receiving/cancellation precede this at `:1062`.
- `packages/server/src/transaction-read.ts:218`: Resource/missing refusals return before minimum comparison, correctly preserving nondisclosure. But successful results at `:231` permit unknown waiting only for `resource` or `initial-collection`; alias success is excluded.
- `packages/server/src/transaction-consistency.ts:115`: unknown + waitable false throws `TransactionMinimumUnsupported`. The waiter at `:139` retries only Unproved or known-behind; HTTP catches Unsupported at `transaction-http.ts:663` as the private development 500.

## Normative basis and limits

Pinned monolithic `docs/specs/bdp.md` SHA256 `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8`:

- Lines 254–269: alias is a repointable locator, not a Resource; alias changes do not revise the target.
- Lines 1424–1441: successful alias GET/HEAD is a temporary bodyless 307 with canonical Location; query/fragment/unknown locator behavior and authorization projection are retained. Do not broaden alias authorization or claim redirect alone authorizes the followed target.
- Lines 5820–5832: lexical receiving is distinct from position support, with malformed-control precedence and HEAD metadata rules.
- Lines 5834–5839 expressly include alias targets in uniform current-context 404 before semantic comparison/waiting.
- Lines 5841–5849 give the remaining-read eligibility/wait/route/problem law. Choosing the existing finite waiter for a live alias's matching opaque minimum is a bounded reference implementation choice under that general law. **The more specific Resource/properties unknown-minimum MUST is not an alias-specific mandate**, because aliases are not Resources.
- Lines 5850–5851 leave retained continuation/Snapshot precedence separate and do not extend unknown waiting to successful incident-Link lists. This recommendation must not change the shared `missing` kind wholesale.

The required-contract matrix supplies inherited alias navigation and TX locator/position obligations (rows Read navigation, TX singleton/locator aliases, and positions/reads/barriers). Its historical implementation descriptions are not treated as current source facts. The sole plan prioritizes actual Transactional serving and branch-only reviewed increments; its public activation predicate remains unmet.

## Smallest implementation boundary for a later reviewed plan

One production file appears sufficient: distinguish alias conditions from successful incident-Link conditions in `transaction-read.ts`, allowing alias success to use the existing same-entry `checkCurrentReadMinimum(..., true)` behavior while preserving its existing 404 bypass. An explicit private alias condition kind is a straightforward option; final spelling belongs to the bounded plan. Keep the exported request classifier's incident-Link result unchanged and avoid generic broadening of `missing`.

No new waiter, owner API, policy, operation, public route, discovery switch or shared HTTP receiver is needed. Existing capture, authority admission, serialization entry, context extraction, cancellation and drain remain in place. Recognized position comparison and foreign recognized epoch/view errors remain unchanged. Opaque foreign epoch/view remains explicitly unsupported; this slice does not invent its policy. Alias-only changes still cannot advance a Resource position or satisfy a known future bound. Retained mutation/sequence alias witnesses and retry identity remain separate from current locator reads.

Expected author scope is four existing files maximum: `transaction-read.ts`, `transaction-read.test.ts`, `transaction-http.test.ts`, and `apps/bdptest/src/transactional-alias-installed.test.ts`. This is a proposed scope, not authority to edit them. Recheck this diagnosis against the eventual clean, reviewed green receipt successor before authoring; the current binding is intentionally the frozen dirty receipt draft.

## Focused evidence proposal, without redundant whole suites

1. Extend selected current-alias source coverage to distinguish direct Unproved from service timeout for a real live alias at matching opaque and near-miss minima. Record actual owner observations and finite waiter activity, not inferred counts. Mutate caller tuple after capture and repoint during a controlled wait: a different live target must not become a stale 307 or satisfy an opaque bound. A transition to missing may return the already-required current 404, so keep that as a distinct nondisclosure oracle rather than insisting all waits time out.
2. Reuse the existing known-position `conditions alias mapping and context in one entry without revising either target` and `rechecks a repointed alias after waiting for a real Resource effect` callbacks. Retain the receiving/cancellation/lifecycle and provider-failure callbacks; add only unknown-path discrimination where their existing known-path checks cannot demonstrate it. Missing/deleted/dangling/hidden and malformed controls must still avoid waiting. Successful incident-Link unknown behavior remains excluded, with an explicit preservation check.
3. Amend the existing seeded HTTP opaque-500 row to the new contextual 503 and exercise GET/HEAD pairing, positive GET length plus empty HEAD, retry metadata, no Location/ETag or conditional/negotiation shortcut. The existing callback already uses budget 2 and deterministic clock/wait seams, so additional waits do not require widening its default timeout. Retain known 307, 409/410, receiving 400, query/missing404, authentication and method boundaries. Exact row selection and counts belong to the next plan, not this audit.
4. Append a separate installed callback rather than rewrite the historical sequence/retry/restart cohort. A compact proposed witness is **six raw exchanges, one child and one store**: real batch POST200 to seed one Bead; put-alias POST200; eligible GET307; matched opaque GET503; paired HEAD503; absent-alias opaque GET404. Validate both receipts and their actual context, including alias requiredPosition without Resource-position advance. Do not auto-follow the 307 or infer discovery from a test-authored closed target set. Use callback-local finite executor/count/allowlist and strict lifecycle ownership; honest external packed-payload/source bindings; existing historical evidence remains distinct. Proposed new callback limits are request5s/work45s/cleanup10s/callback60s, not a change to the historical cohort's bounds.

Fresh exact emitted build must precede collecting `transaction-http.test.ts` (module imports `../dist/transaction-runtime.js`) or installed callbacks. After native and actual Claude plan/code disposition, later validation should run selected source semantics and the new plus relevant historical installed cohorts with observed stage/process/store cleanup. This audit authorizes none of that execution.

## Comparison and deferred alternatives

Scope/Bead/Link finite Event minima and fresh Snapshot minima are already separate completed increments; receipt-root matching opaque minima are the current frozen candidate under CODE review. Alias resolution is not covered by those families: it has no Resource representation/validator, uses a repointable locator, and shares a private classification with the still-excluded incident-Link path. Its missing-before-minimum and same-entry mapping rules are the distinct evidence needed here. Do not redo the completed Event/Snapshot mechanics or widen receipt validation.

Public cold discovery remains a real delivery blocker. `transaction-discovery.ts` builds internal truthful data and the service has an internal observed discovery seam, but `transaction-http.ts:212`/`:615` intentionally refuses public Transactional discovery. Enabling it now would advertise a still-incomplete profile; this recommendation does not fabricate navigation or use an admission bypass. It removes a concrete serving blocker before that separate activation milestone.

Fresh `start=now` changefeed opaque minima are another plausible later gap, but current `changes`/`readChanges` share strict minimum handling with retained after-checkpoints. Their distinction needs a separately reviewed capture/checkpoint seam, making them less bounded than live aliases. Retained Snapshot/page minima (S1), incident-Link metadata/Event membership (G2), foreign opaque context policy, optional History, Types/editorial/provenance work and Janet's surfaces remain deferred. None blocks this live, matching-context, current-alias slice. No new human product ruling is identified; normal root/native/Claude review and a clean source rebind are required before authoring.

## Audit status

All 679 current tracked bytes match receipt author binding `c1c6bd6b…`, on HEAD `08812c6c752c06cd0edf1af2878b2367692e5b32` with the exact six dirty receipt files. No source mutation, test, product import, build, formatter, install, service, fetch of main, publication or coordination-bus action was performed. This report and its source/context binding are the only new artifacts. Recommendations and source-derived prospective traffic are not measured observations, completion percentages or qualification evidence.
