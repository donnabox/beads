# Native plan review — first W2 installed sequence observations

Read-only against combined `3e2b8784daefd5893f990785cfe91f1d872a2ffe` in
`/Users/dbox/repos/wt/vickie/vickie-installed-profiles-integration`.
Reviewed `read-update-w2-sequence-plan.md`, current required catalog, normative spec,
canonical schemas, actual sequence preparation/lifecycle/receiving code, and installed
exact client/raw-action seams. No source edits, tests, generated qualification or
outward actions.

**Proceed with bounded observations after the scope corrections below.** No new
runtime/protocol/product decision is necessary. The plan should not currently promise
three whole candidates: only member-problem-shape has a complete proposed observable
oracle for its cited clauses; local-bindings and carrier-rejection need the distinctions
below. Useful partial observations should proceed without adding scheduler/fault seams.

## Required corrections

### P2 — local-bindings cannot infer commit-before-binding from final results

Plan lines35–43 designate this a whole candidate. Its first catalog clause is literally
“The binding becomes available only after the create commits” (spec2745). A dependent
update and a final GET can pass when an implementation resolves an allocated identity
from staged state before committing the creator and only commits later. They prove
identity resolution and final durable effect, not the timing premise. The plan correctly
recognizes the same issue for order-and-partial-commit at line61; apply that discipline
here too.

Keep `local-bindings.wholeContract:false` with named timing limitation unless a real
commit-before-dependent discriminator is independently available. Do not add an
installed scheduler hook or owner calls merely to force this first slice complete.
The successful/failure/retained afterstate cases remain valuable bounded observations.

### P2 — dependent failure retention needs an independent key discriminator

Plan line39's live-alias setup is a good improvement: clear the collision, replay the
original keys, observe no new Bead, then prove a fresh key can use that path. It detects
re-execution of the *creator*. But a broken implementation that retains the creator
failure and recomputes `binding-unavailable` afresh without retaining the dependent
key would produce the same response and state on every such replay.

The local-bindings second selected clause explicitly calls `binding-unavailable` “a
retained failure” (spec2857–2858). Add this bounded existing-HTTP discriminator after
the retained replay: submit a valid changed-semantic mutation of a separately known
live Bead using the original dependent key. Require `idempotency-conflict` and an
unchanged independent GET. Run that same mutation under a fresh key and require its
known effect/GET. The namespace is shared across singleton and sequence carriers
(spec3198–3200), so this can use the existing singleton lane without new nested
bindings. Changed semantics should be obvious (different concrete subject and patch),
not a disputed equivalent-reference case.

This proves dependent key retention, but does not close the separately deferred
creator-rekey/dependent identity clause. Preserve that limitation.

### P2 — carrier-rejection's final cited clause has uncovered refusal classes

Plan lines45–55 calls its list “the complete refusal set” and predicts only400/413.
The catalog's final citation “are direct problem responses and execute nothing” refers
to the whole list at spec3157–3162, including unauthenticated principal, unsupported
body media, oversized JSON and invalid patch JSON Pointer, plus rate-limit/unavailable
authority. The proposed cases omit those classes.

Options within the bounded work:

- Keep this row partial and name unobserved receiving/lifecycle branches; proceed
  with the comprehensive syntax and 32/33 tests as planned.
- Add currently reachable401,415,request-too-large413 and malformed-patch-pointer400
  cases, each with no-effect and corrected-key witnesses, then assess actual remaining
  applicability. A body-size probe needs a narrow exact-harness body budget allowing
  the service cap+1; the current equal65536 client/service cap would reject locally.
  Do not change the actual advertised/runtime cap or claim a client rejection as HTTP.

A rate limiter must not be invented if this installed configuration has none; record
that applicability fact rather than demanding an artificial429. Similarly do not add
shutdown/fault hooks merely for W2. Absent an observation or justified applicability
assessment for the remaining clause branches, preserve `wholeContract:false`.
When adding receiving cases, use each real Problem status/code, not line55's400-or413
shortcut. Negotiation is also distinct;406 is bodyless, not a Problem to force through
that assertion, and is not needed just because this particular catalog excerpt exists.

### P3 — describe Problem schemas as extensible, not closed

Plan line41 says “closed binding-unavailable shape.” The canonical
`sequenceMemberProblem` composes `readUpdateProblem`, requires status/index and forbids
`outcome`, but permits Problem extension members (spec3143–3147). The sequence envelope
and success results are closed; Problem objects are not. Reuse the untouched canonical
schema and assert required type/code/status/retry/index/name plus absence of outcome.
Do not reject otherwise valid Problem extensions with an exact-object oracle. Retained
response equality can still compare what this actual service originally returned.

## Accepted details / implementation bounds

- Real sequence URL discovery and the immutable advertised32/body65536 values have
  existing installed support. No new route/runtime functionality is needed.
- The33-member test is well designed: raw body below byte cap, all33 independent absence
  reads, corrected32 plus final singleton under original keys with changed markers.
  Exactly32 success is necessary. Duplicate/invalid-name/key variants should preserve
  unique valid sentinels; corrected semantic changes reveal accidentally captured keys.
- Forward/unknown/self/wrong-kind references are static whole-carrier failures, unlike
  a durable missing/wrong-kind Resource at a member turn. Keep that distinction; do not
  reinterpret a missing durable subject as malformed carrier syntax.
- Member-problem-shape has a plausible whole candidate: named/unnamed success/failure,
  exact result count/indices/names, independent would-be status/type/code/retry values,
  canonical schema and post-observation sensitivity controls. An identity-taken failure
  uses retry `never`, whereas alias-path-taken uses `after-state-change`; static oracles
  must preserve that difference. Shared field-shape claims do not require manufacturing
  every unrelated optional failure condition.
- The allocated creator postimage must not be equated to current GET after its dependent
  changes it. Capture identity from creation, assert initial properties there, then
  independently assert the dependent's static final properties and current identity.
- `order-and-partial-commit` remains partial; eventual response order is not evidence
  of never-parallel execution or immediate between-member visibility.
- Existing literal raw bodies suffice for decoded duplicate member names and lone
  surrogate escapes; normal JSON serialization would erase the intended malformed
  spelling. No new raw-body language is needed.
- If the client lane is chosen, optional carrier header omission plus exact sequence
  route and an unchanged normative sequenceResponse wrapper are bounded. Preserve
  omitted-vs-invalid key distinction, singleton tests, canonical target/credential
  guards and current capture provenance. Source-local Type literals must first agree
  with actual discovered inventory. No direct store observations enter conformance.
- No schema/configuration/runtime changes, installed scheduling/clock seams or new
  product rules are required for the corrected first slice. Actual mismatches should
  be reported before modifying behavior. The base has27whole/102uncovered, not the
  historical28/101; recompute exact reviewed IDs after implementation.

The three P2 scope corrections were relayed to root promptly for the worker. This is
plan review only; it provides no executable evidence or new qualification credit.
