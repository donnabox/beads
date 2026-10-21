# Production Claude findings: source dispositions

Actual completed read-capable review: exit0 in 509.142s, full stdout5443B SHA aac781c2ee13dec964c4f71c8c020ea5f086ddb03fc9f06613f79918a1ffb09d, empty stderr. This succeeds the separately preserved incomplete170B no-tools attempt; that original remains inconclusive. Review binds frozen dirty candidate1647, patch e5cdabf3acf912d0cef684c2263e5a20b9d5c144d51f13ab2083621683ae186a; no main-integration claim. All source/dependency/input hashes independently rechecked, before/after snapshots equal, recorded cleanup clear and PID/group 60325 plus every sampled PID absent now. No runtime or source change during this audit. I authored this implementation/preparation; this is a source-grounded disposition, not a second independent code approval.

## F1 — optional performance tradeoff; no repair

The fixed spelling indeed never becomes recognized by position() (transaction-consistency.ts26–30); repeated unproved reads can consume the established budget/cadence (139–177). However “provably futile” is too broad: each attempt observes current Resource existence and authorization, and a now-missing/hidden target returns404 before minimum semantics (transaction-read.ts201–208). The selected real owner test publishes an empty policy from the wait callback (transaction-read.test.ts2197–2206) and tests changed visibility. Timeout uses the last actual observation context (consistency174–175; test181–200). Owner corruption/unsupported/cancellation remain terminal, not blindly retried.

Operator ruling explicitly requires existing finite wait then503, with no new cadence policy. A special50ms branch would alter the accepted design and consume no less semantic uncertainty about recognition; do not add it in this slice. Approximately200 observations is an approximate default cadence count, not throughput proof or an unbounded loop. No immediate-return/fabricated503 or stale success substitute.

## F2 — valid redundancy observation; optional cleanup declined

HTTP407 rejects incomplete/non-token tuples. currentReadMinimumHeaders63–72 accepts the remaining lexical tuple, so ordinary immutable Headers leave HTTP580 unreachable. Keeping its conservative refusal does not widen success or break the later excluded-route check586–590. No correctness repair or extra test is required; omit unrelated cleanup churn.

## F3 — implementation limitation, not a missing normative ruling

checkCurrentReadMinimum122–129 deliberately treats unknown position plus foreign epoch/view as private unsupported; HTTP650–651 labels minimum-position. operator-rulings.json explicitly excludes this collision from the two decisions. The new spec paragraph5845–5851 is deliberately current-epoch/view bounded. Do not turn private500 implementation behavior into normative protocol law.

Precision: progress.md720–728 contains historical broad unknown-token gap wording, and the new draft paragraph1997ff describes the current bounded scope/strict-consumer separation. It does not contain an exact newly-added “unknown plus foreign ->500” sentence. The accepted plan and operator ruling carry the explicit exclusion. This does not require editing historical text or expanding the spec.

## F4 — accurate compatibility observation, already the explicit operator decision

Yes, new missing-first applies even when a known minimum is ahead or foreign. This is the selected behavior, not accidental relaxation: operator ruling1 specifies uniform404 before waiting/epoch/view. Source check occurs after actual owner observation and preserves early invalid-parameter/limit refusal. HTTP observed-refusal projection carries actual current context. No minimum-satisfied or stale-representation claim is made for404. Existing source/HTTP assertions were updated under this exact ruling; broader full CI remains a later gate.

## F5 — no reachable HTTP defect; optional fallback declined

The only production minimumFor consumer is service.readAtMinimum (service384). Both valid paths go through waitForTransactionMinimum (397–415), which consumes Unproved and emits catch-up-timeout only at deadline. Invalid delivery-option fallbacks delegate to authority receiving rejection before observation; they cannot produce Unproved or an unconditioned success. Direct source-private minimum facet calls may intentionally observe the private condition; they are not an HTTP route.

Adding a direct Unproved-to503 catch would allow a future incorrectly unwrapped path to bypass the required finite wait. Preserve the current single wait-to-HTTP failure boundary rather than conceal such a wiring defect. No mandatory repair.

## Review context and limits

The review also explicitly checked descriptor equivalence with authority receiving, same immutable request classification/materialization, exception identity through owner entry, continuation nonallocation on thrown conditions and excluded-route labels. These source facts support this bounded patch; they are not installed/runtime qualification. Root must aggregate the remaining source-test/installed scopes, final actual-main integration, installed controls and exact CI. No source amendment recommended from these five findings.
