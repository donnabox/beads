# Valid but unrecognized minimum-position identity

**The existing contract constrains this case but does not uniquely select its semantic response.** The latest operator ruling explicitly separates alphabet-valid unknown identities. No affirmative policy for them was found in the targeted rulings below. The deliberate private 500 is not a normative completed implementation; neither lexical 400 nor expiry can be inferred from failure to interpret a position.

Basis: actual merged 67a2842e3f1918e70978f745587cf11eaf0df5a8/tree596abe870220fb61ce606f8382a9da52d554d063. Read committed bytes and verified the five inspected working files equal those objects. No product edits, imports, tests, services, additional review launches or user questions.

## What is already decided

- Spec5819–5832: missing/repeated/lexically invalid minimum fields receive400 malformed-request after ordinary authentication/receiving, before owner lookup, without projected context. All three absent requests no minimum. It expressly says lexical validity does not establish known/supported identity.
- Spec3690–3705 and6625–6642: positions belong to an authority-defined total order and are opaque/equality-only to clients; the ASCII alphabet is not a universal counter grammar. An unknown string is not evidence of an expired position. Do not reinterpret it as genesis/current head or order it lexically.
- Recognized tuples: consistency.ts80–89 checks epoch, then view, then its private position ordering. Existing epoch incompatibility maps to410 cursor-expired; view mismatch to409 foreign-view. An ahead recognized position waits under the original finite deadline (96–143); expiry of that wait maps to503 catch-up-timeout/after-delay with the actual observed context (HTTP600–614).
- Spec5330–5333 defines catch-up-timeout as inability to serve at/after the requested position within the authority's bound. Spec5834–5838 permits representation/wait-or-route/established failure, never stale success. This strongly supports bounded503 as a conservative unknown-identity fallback, but does not expressly say whether a token that identifies no interpretable position is an invalid semantic parameter instead. The closed Problem table contains400 invalid-parameter and503 catch-up-timeout, but names alone do not decide applicability.
- Spec1889–1893/1958–1962 reserves bodyless500 for unexpected faults. The known private capability refusal cannot be advertised as a compliant unknown-token policy.

## Current behavior and precise gap

consistency.ts26–31 understands only genesis and safe canonical pos-N. snapshotTransactionMinimum rejects otherwise alphabet-valid values at45–54; transactionMinimumHeaders catches that into undefined at57–67. Thus complete current epoch/view plus unknown, pos-01, or an overflowing counter passes lexical receiving but loses semantic interpretation. It is distinct from a recognized future pos-N: the latter can be compared and may catch up.

HTTP394–395 performs lexical receiving first. Fresh Snapshot416–417 and ordinary current-read564–567 deliberately return the minimum-position development500 for unsupported identity. Changes/Event paths pass an explicit unsupported sentinel (460–521); receipt paths preserve their existing service nondisclosure flow (528–561). Service494–511 can retain another-principal/unknown receipt404 before returning an own disclosed unsupported result; it does not fabricate a wait or turn unknown into position zero. Supported ordinary read/service checks are at read.ts139–166 and service.ts316–346. The current parser can reject unknown position before even reaching epoch/view comparison; no new collision precedence is inferred here.

## Smallest decision, if root elects to resolve this next

Limit the question to an authenticated, disclosed, current read with the **actual current epoch and view**, valid field syntax, and an identity whose ordering cannot be established. Recommend: **treat the minimum as unsatisfied, retain the existing finite read budget, and return503 catch-up-timeout/after-delay if eligibility remains unprovable**. This uses the existing failure vocabulary and protects the minimum without claiming expiry. It is a recommended clarification, not an already-recorded answer or implementation release.

Concrete example: current actual context is E/V/pos-2; client sends E/V/opaque_unknown. The request must not return the pos-2 representation as satisfying that minimum, must not return lexical malformed-request, and must not assert410 merely because the token is unknown. Under the recommendation it ends503 after the bounded attempt; any response context must come from a real owner observation, never from the requested token. An immediate semantic400 invalid-parameter is the narrower alternative if the operator wants permanently uninterpretable identities distinguished; that classification needs explicit wording too. No new status/code, indefinite waiting, fabricated expiry, fabricated position, or retention mutation is proposed.

Do not bundle foreign epoch/view plus unknown-token precedence, missing/hidden Resource ordering, retained Snapshot/continuation S1, or incident-Link G2 into this question. Their established checks/holds remain unchanged. No catalog or broad implementation plan is needed to make this one decision.

## Search and attribution

Read current spec's position/character profile, receiving, consistency and Problem definitions; consistency/read/service/HTTP paths above; the receiving operator ruling, prior minimum-rule audit, command-field ruling and compatible-missing root dispositions. The receiving ruling explicitly lists unknown identities as separate unresolved; the command ruling applies only to mutation POST ignore. The prior audit and next-serving-slice label unknown policy unresolved but are analysis, not operator authority. This bounded search does not claim to recover unsaved conversation or every historical artifact.

### Exact source pins
- docs/specs/bdp.md: 63e93be562334c84c1e967da2fce212e8ef515aa3c9448924c5f9a272ea7b80b
- packages/server/src/transaction-consistency.ts: 00a420c3e9f5dcb1b8814a6476d5ce52edfb7adbbd1c64ab082833f17322a336
- packages/server/src/transaction-read.ts: c3b486baf6b9e48d4a716d36b4162c2ba1f0c7672a6f48988192e2591eda7910
- packages/server/src/transaction-service.ts: 53f655faf494209996aac2e0f93a4c2c785e471f66aac6af9ea41f92c9cfbe00
- packages/server/src/transaction-http.ts: ebb3550110bd6fda638572966e5faa05c15f062f37a10947f6e442fc7d4570f4

### Read record pins
- transaction-minimum-read-receiving-plan/operator-ruling.json: c09f81e8a825efee610eb9b8a9b3a42c5b8a83728ded9d92e180c3cdb6197608
- transaction-minimum-read-rule-audit.md: 35fa1a378662c78a02b03f8bf847b87cedbc25aff6b6a406a94b414beb624fd9
- transaction-selector-command-operator-rulings.md: 5f7c78447789da4df73c616c174a23832d4b72b4c00238a3c7c26240993d2611
- transaction-compatible-missing-minimum-root-code-dispositions.md: 5ebe28756982b46baaaa848fd91bcaa80800e2e57461d74a23fb8bc47d065401
- transaction-minimum-receiving-draft/next-serving-slice.md: 969f5c9caf0b942e14722586d7d311a64862eae43a2a065d36d1d1b3b8af5020
