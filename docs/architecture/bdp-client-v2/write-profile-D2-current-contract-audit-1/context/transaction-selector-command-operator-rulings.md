# Operator rulings: Selector admission and command minimum fields

Recorded 2026-10-14T21:00:43.587274+00:00. Authority: direct operator answers in the active Vickie chat, session2026-10-08T20:26:35Z-81e5. These supersede only the two pending questions below; metadata-only Link Event membership and retained Snapshot/continuation minimum S1-Q1 remain unresolved.

## Selector admission

Question: “One decision blocks proper errors for set operations (`updateWhere`/`deleteWhere`). I recommend rejecting invalid selector syntax with 400 before key lookup, without saving a receipt. For valid selectors that exceed size/depth limits, should a new request get 413 without consuming its idempotency key, or a saved failure that consumes the key? I recommend the first: existing equal retries still return their saved outcome, conflicting keys still get 409, and limits are checked before any operation executes. The spec does not currently choose this behavior; PR integration continues meanwhile.”

Operator answer: **“Reject with 413; leave key unused (recommended)”**.

Accepted implementation direction: reject invalid Selector syntax with400 before key lookup and without saving a receipt; valid over-limit Selector on fresh request gets413 before execution with key unused. Equal retained retries preserve their saved outcome; conflicting keys remain409 ahead of fresh capacity checks. Preserve ordinary framing/body bounds, authentication, complete-carrier validation, fixed limits and existing operation semantics. The current plan must reconcile exact parser error categories/bytes/depth/nodes and precedence with this chosen recommendation; do not silently expand unrelated private depth or execution-limit semantics. Use existing malformed-request/limit-exceeded Problem shapes under reviewed normative amendment.

## Command minimum fields

Question: “A separate protocol choice blocks mutation requests that carry copied read headers: should the three minimum-position request fields be ignored on mutation POSTs? I recommend yes: they remain read-only controls, so normal authorization, idempotency and expected-revision rules govern the command, and a later receipt GET still enforces its minimum. The spec currently leaves this applicability undefined.”

Operator answer: **“Ignore on mutation POSTs (recommended)”**.

Accepted implementation direction: three read-minimum fields are ignored on mutation POSTs. Ordinary authorization, idempotency and expected-revision rules govern commands. Receipt GET still enforces its minimum. Preserve global transport/header/body bounds and existing Read behavior; no inferred command-side barrier or new field.

## Priority and review

These concrete serving fixes take priority over planned sequence member-transaction evidence. PR122 cross-carrier evidence remains in CI at c02ccc3706f2e258d380805162a8c47dbb73f4e1. Prepare bounded implementation plans and native+Claude review before authoring, coordinate shared transaction HTTP/spec edits, keep ordinary ancestry and exact evidence. No public profile advertisement, release or Preview2 promotion follows from these rulings.
