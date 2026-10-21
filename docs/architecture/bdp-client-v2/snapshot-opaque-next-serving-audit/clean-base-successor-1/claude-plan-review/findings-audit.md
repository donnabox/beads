# Independent receipt PLAN findings audit

Recommendation: accept findings1–11 as bounded plan precision, preserving the qualifications below; retain finding12 as an actual review limitation. No design, workload, source-scope or deadline expansion is needed. Findings1–4 clarify implementation choices already required by the behavioral plan, not a new product ruling. This audit does not relabel the completed Claude review or infer code approval.

## Actual review evidence

Claude completed exit0 in1123.968341s, timeout false, one successful nonempty terminal result, no parse errors or cleanup required. Independently parsed all4645 JSONL records; terminal text exactly equals final-review.txt SHA256 `2780d62729197c2cb06b2573b3905c9d2484aadd48a5dc62ba64d94f08e01637`. Verified stream SHA256 `6ebcddc4216eefe71b8feb6baba33f8e0d3881fde6dfbeda64dd1a71a01295bd`, empty stderr, terminal-results equality, source-before/source-after equality, all679 current source pins and all bound context/artifact hashes. Sampled PIDs64568/64725/64905/65224 are absent at this audit. No saved helper/parser was executed or imported; data-only parsing/hash inspection and read-only source inspection were used.

## Finding dispositions

1. **Accept exact owner calls, correct rationale.** Replace root capture with `snapshotCurrentReadMinimum(minimum)` and semantic call with `checkCurrentReadMinimum(condition, context, true)` at transaction-store.ts959–971. The helper's true selects Unproved for actual matching epoch/view; mismatched context remains unsupported and known positions keep the existing check. Do not repeat the reviewer assertion that receiptAtMinimum is the only minimum-bearing caller as an absolute API fact: owner.readReceipt is directly callable. The owner raises a synchronous Unproved observation; receiptAtMinimum's finite waiter supplies bounded retry/deadline. No owner loop is added.

2. **Accept explicit branch-local parsing.** HTTP551–575 currently shares one strict minimum. Split computation so page alone retains strict transactionMinimumHeaders and its undefined/unsupported distinction, while root alone uses currentReadMinimumHeaders with the has-controls fence. Merely replacing the shared parser would bypass the page sentinel and alter its private-limit response. Existing page code/semantics remain unchanged.

3. **Accept exact label-first tuple spelling.** Keep `%s`, using first-column strings false/true/opaque false/opaque true with corresponding flags/positions, preserving historical final names. The old plan did not require an object table rendered with `%s`, so the review overstates inevitability of a title change; explicit tuples remove ambiguity without altering the intended selection or workload.

4. **Accept principal versus epoch/view distinction.** Principal-hidden/missing roots return contextual404 before comparison. A disclosed own root with unknown position plus foreign epoch/view preserves private unsupported500. Do not turn either reference fence into a universal normative foreign-token policy.

5. **Accept D receiving/auth labels.** Rows7/10 are shared pre-route400/401 fences, not receipt-owner observations. Keep ten wire requests and avoid inventing ten receipt-branch entries.

6. **Accept paired HEAD length oracle without more traffic.** read-http.ts42–45 supplies serialized body length and strips HEAD; transaction-runtime.ts93–95 supplies GET wire length. D/E each already pair identical503 Problem GET/HEAD outputs, so assert GET serialized length/actual bytes, HEAD matching positive length and zero body bytes. Different conditional inputs do not alter the503 Problem. This is a wire oracle, not a requirement that every direct handler GET object already contains a length header.

7. **Accept deliberate no-signal scope, correct uniqueness.** receiptAtMinimum indeed supplies no signal to its waiter and HTTP supplies none to it. However receiptPage also calls waitForTransactionMinimum without a signal at service585–606. Do not repeat 'the only minimum-bearing path'. State that the receipt-root preexisting no-signal contract is deliberately retained; pages remain untouched and B proves finite close/drain, not disconnect cancellation.

8. **Accept excluded private-gap label for D5/6.** Existing unknown-position classification precedes foreign epoch/view semantics and keeps private500. Those rows preserve an excluded implementation gap, not conformance to normative409/410 or a resolution of foreign unknown policy.

9. **Accept explicit raw owner probe.** Use f.owner.readReceipt directly after close, expect closed error; it is separately counted and bypasses the live-attempt delegate. This makes the already intended B2+one-closed-probe accounting unambiguous.

10. **Accept external-supervisor attribution.**120s is the proposed validation-stage supervisor limit, not an in-repository test/CI timeout guarantee. Name future packet/runner enforcement and distinguish5s request,45s work,10s cleanup,60s installed callback. The validation runner remains to be prepared/released; do not claim it already executed or CI enforces120s.

11. **Accept precise5850–5851 citation.** The caveat limits extension of the narrower Resource MUST; the affirmative receipt choice still rests on general remaining-read law and explicit subsequent receipt GET/HEAD controls. Silence in the caveat alone is not a new receipt-specific MUST.

12. **Preserve actual context-access limitation.** Prompt described context/ as beside the prompt without giving its absolute artifact directory; Claude's stdin review cannot infer that file location from cwd. Stream contains Glob searches for context/**, /Users/dbox/repos context/*, /Users/dbox/.codex **/context/plan-v2.md and /private/tmp */context/*.json. It did not access the immutable context snapshots. It reviewed full plan/selection reproduced in stdin and actual repository/spec source. Therefore do not attribute independent reading of design/rebind/CI artifacts or679-byte verification to Claude. Those facts were separately verified by root/native/runner. Absolute context directory and file paths belong in future prompts; preserve all completed prompt/output/stream bytes and this limitation, with no blind rerun. Read-only tools could read an explicitly located JSON artifact, but no such reading occurred here.

## Closure

No substantive disagreement with the bounded receipt behavior or six-file decomposition. The concrete corrections clarify existing intended behavior; selection remains12 source callbacks/36 HTTP and installed22+5, with unchanged clocks/stores/deadlines. Original native plan closure remains historical and did not catch these text ambiguities/context usability defect. Future plan precision successor should preserve that chronology and bind this audit; root separately accepts and releases any later source work.
