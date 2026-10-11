# Transactional sequence dependencies: implementation representation

Implementation plan reviewed by Vickie, 2026-10-11. Steps 1–3 are authorized in an isolated branch; alias admission wording remains a separate clarification. No new wire behavior is adopted.

Source assessment: Evidence is frozen commit
`4a9137e01817712aedce0b780abb39d02cf1470c` in
`/Users/dbox/repos/wt/vickie/bdp-transaction-sequence`. This memo changes no
normative text, runtime source, profile advertisement, or public API. Root owns
review, integration and any normative clarification. The fork design.md remains
the plan of record; this is a bounded implementation handoff.

## Recommendation

Implement canonical `@name` dependencies next, together with unresolved direct
comparison. Keep sequence Resource-alias operands explicitly refused until the
narrow admission-representation clarification below is resolved. Alias singleton
mutations remain supported, including a `putAlias.target` supplied by a successful
Bead creator once dependency support lands. No separate Read+Update key owner is
needed or acceptable.

The central addition is a durable creator-attempt binding fact and a pending
request template whose dependency slots reference that fact. A creator key or
label is insufficient: a failed receipt can expire and its key can be reused
before its original dependent reaches its turn. Final semantic identities still
contain canonical Resource identities or the existing unbound marker, never
attempt identifiers. This is internal storage, not new protocol semantics.

## Normative evidence

All line references below are to `docs/specs/bdp.md` at the frozen commit.

| Clauses | Required behavior |
| --- | --- |
| 1317–1346 | One Scope/epoch/principal key namespace across all carriers. A sequence dependency normalizes to its creator's bound identity; batch allocated references retain their different symbolic-index rule. |
| 1366–1389 | Durable admission records identity or unresolved dependency form, pending receipt and exclusive owner atomically. Commit and retraction are fenced. Recovery retracts pending attempts within a finite bound and never resumes them. |
| 1391–1416, T62 | Reserve unknown dependent keys with all other unknown keys. Pin creator attempt and operation slot. Record creator binding atomically with effects and terminal disposition. Preserve facts across failed-receipt forgetting. Failed creators normalize to the existing unbound marker; terminal dependent identities cannot rebind. |
| 1418–1434, T62a | Direct singleton/batch: immediate provable conflict, otherwise bounded comparison wait outside a DB transaction. No receipt disclosure before equality; unresolved deadline is direct retryable 503, retaining no new outcome. |
| 1436–1452, T62b | A still-present direct request may compete afresh after retraction using the same deadline. It inherits no ownership or identity. Epoch change ends it with retryable 503. |
| 1454–1469 | Completed compact outcomes remain throughout the epoch; failed receipts are forgotten whole after retention. |
| 1472–1544 | Each sequence member commits separately. Foreign pending projects in-progress without waiting. A dependent of pending/transient creator consults no dependent key and releases only its own reservation. Stable dependencies then use normal comparison. |
| 1546–1558, T63 | Hidden public allocation does not remove the internal creator binding. Dependents authorize independently. |
| 4110–4113; conformance row 6509 | An alias resolves when its member is reached; an earlier alias put in the same sequence is observed. |
| 4246–4293 | Exact semantic equality/defaults, bound/unbound/transient distinction, and durable original alias target/miss witnesses; new retry locators are transient comparison input only. |
| 4295–4356 | Claims precede execution atomically across the carrier; existing keys are classified as of each member turn. Retained conflict is not a new retained disposition. |

## Proposed private representation

Use a new private format and identity-policy version. Refuse every older private
format before maintenance/recovery; do not migrate or reseed. A format refusal
must leave the old file byte-identical. This is the established development-store
compatibility policy, not a public migration design.

1. Give each admitted sequence presentation a random internal attempt ID. Preserve
   each member's declaration index and each creator's operation slot. For this
   one-operation-per-member carrier that slot is zero, but represent it explicitly.
2. Add durable creator facts keyed by `(sequence attempt, member index, operation
   slot)`, with `unresolved`, `bound(canonical id, Resource kind)`, `unbound`, or
   `transient`. Record the specific acquired receipt/owner token where the creator
   is newly owned. Facts are immutable once resolved. A foreign receipt alone is
   not a successful presentation: a conflicting creator presentation yields no
   creation binding, even when the preexisting receipt succeeded.
3. Add an explicit receipt admission state: `resolved(identity, witnesses)` or
   `unresolved(template, dependency edges)`. Keep owner, transaction identity and
   receipt unchanged when resolving the template. Do not overload identity text
   with a pseudo-canonical URL or user-forgeable JSON sentinel.
4. The template preserves lossless operation data, normalized defaults and every
   nondependency semantic field. Each recognized reference slot records its
   creator fact key and reference kind; a pinned reference preserves its pin
   byte-exactly while substituting only the URI. JSON strings in properties remain
   ordinary strings. Use the existing tagged semantic codec and private unbound
   sentinel for final identity; never round numeric literals through native Number.
5. Durable edges keep facts alive while any owned unresolved dependent needs them.
   Facts can be garbage-collected when no such edge remains. A terminal dependent
   retains its ordinary final identity/witnesses, so it does not need its creator
   receipt or fact thereafter. Recovery retracts pending receipts and removes the
   corresponding edges/facts atomically; it does not reconstruct a runnable tail.

Facts attached to a presentation slot handle both newly executed and replayed
creators. For a newly owned creation, write its bound/unbound fact in the same
SQLite commit as Resource effects, group and terminal receipt. For a retained or
expired creator, establish the presentation's binding from the authoritative
retained allocation during its member turn; never infer it from response JSON.
For a permanent refusal of the creator presentation (including a key conflict),
record unbound. For a foreign pending/transient presentation, record transient.
The existing RU `MemberCreatorBinding` and `turn` helper demonstrate this
separation, but its key state machine and identity bytes must not be copied.

Do not optimize a supplied-ID creator to success before it executes. The supplied
URL is known, but that creator can permanently fail: its dependent then has the
unbound identity, not the supplied URL. Conservatively retain a dependency hole
until that presentation's binding is stable. This follows sequence binding rules;
it does not import the different batch supplied-ID normalization rule.

## Admission, execution, replay and cleanup

Whole-carrier syntax checks still happen before any key effects: duplicate keys,
unknown/forward/wrong-kind labels, malformed references and unsupported operations
reject everything. In this slice Resource-alias operands remain a preflight
refusal. Support the six Resource kinds plus the two existing alias mutations;
only their protocol-defined reference slots may accept `@name`.

Admission creates every unknown receipt/owner and every unresolved template/fact
edge in one SQLite transaction. Existing key states are not replaced. Do not
classify an existing dependent by semantic identity before its creator's turn.
An all-unknown carrier must be observed wholly admitted or wholly absent by a
competing presentation, including an injected storage failure midway through it.

At each ordered member turn:

1. Read the creator presentation facts first. Any transient fact produces the
   transient sequence in-progress projection before dependent-key comparison or
   alias resolution. Retract only the dependent reservation recorded as owned by
   this exact attempt; do not inspect or modify a foreign dependent receipt to
   construct this result. Independent later members still proceed.
2. Once all dependencies are stable, bind canonical IDs or the private unbound
   marker and construct the ordinary one-operation Transactional semantic identity.
   For an owned unresolved receipt, verify exact owner and referenced facts, then
   freeze that final identity. For an existing key, apply sequence precedence:
   foreign pending remains in-progress even if a static mismatch is visible;
   otherwise compare the normalized presentation at this turn.
3. A permanently failed creator causes the dependent's own retained
   `binding-unavailable` failed receipt, with no graph/group effects. Its identity
   includes the unbound marker. A retry through a new successful creator conflicts;
   another failed creator with matching remaining fields compares equal.
4. A bound dependency feeds the existing evaluator using the captured canonical
   reference. The Resource can have been modified/deleted since the creator
   committed; ordinary guards/not-found rules apply at the dependent's turn.
5. Commit effects, group, receipt and any creation binding fact through the existing
   `executeClaim` atomic boundary. A creation depending on another creation may
   itself bind a later member, so this must work transitively, not just at depth one.

Disconnect does not cancel an admitted carrier. Internal abort/retraction releases
only exact owned pending tails and their dependency edges; a committed prefix stays
committed. A new attempt with reused keys receives fresh owner/transaction IDs for
unknown members. Old claims and fact writes cannot mutate that replacement.

Retention is orthogonal to fact lifetime. A successful creator's expired compact
receipt still supplies its allocation. Forgetting a failed creator must not delete
facts needed by outstanding dependents. After a dependent fails, its own failure
retention starts at its terminal time, not its creator's. Completed dependent
identity survives detail expiry unchanged. Maintenance and fault handling must
preserve these distinctions.

## Direct comparisons must land before dependencies are enabled

Add a private asynchronous comparison/admission seam over the same store owner;
keep the low-level synchronous SQLite steps short. It returns an internal
`comparison pending` state, not the owner's receipt, when equality is not yet
provable. It never introduces another key namespace or lets HTTP policy enter the
store. The current fixed all-visible authorization boundary stays explicit.

A structural comparator checks operation count/kind and normalized known fields.
Compare exact values/defaults/object equivalence; treat only actual dependency
slots as holes. Preserve pins and surrounding reference structure. A demonstrable
mismatch returns conflict immediately. A possible match waits on owner-state change
or a bounded timer outside SQLite and outside the store's synchronous busy guard.
No execution or pending-receipt disclosure occurs until ordinary equality is known.

Use one monotonic, finite, authority-selected comparison deadline per explicit
request, distinct from transaction duration. It is a private runtime constant/seam,
not a new discovery field. Recheck deadline, store state and epoch on every wake;
notifications are hints, not authority. Timeout is a transport-independent
`temporarily-unavailable` result for later direct 503 mapping; do not retain it.
An unresolved comparison cannot return the sequence-only in-progress problem.

Recommended T62b choice: allow fresh competition after retraction within the same
deadline. Reuse the ordinary fresh-admission function and acquire a new token; do
not transfer a previous template or receipt. If a competitor wins, compare again
without extending the deadline. Once equality has been proven and the caller
joined, stop using this pre-comparison retry path. Sequence calls never wait or
use the direct retry-after-retraction permission.

## Feasible code seams

- `transaction-batch.ts`: retain whole-carrier protocol preparation, but represent
  sequence dependency slots separately from `PreparedTransaction`'s batch-local
  labels. Add pure known-field/template normalization and final binding helpers.
  Do not feed a sequence `@name` to batch normalization: it would encode a creator
  index rather than the committed identity. Reuse reference-slot rules, lossless
  values and semantic codec; keep TX defaults/context folding intact.
- `transaction-store.ts`: own schema, reservation/fact/edge lifecycle, fenced
  finalization, retention/recovery and one-step comparison results. Extend
  `admitSequence`, `nextSequence`, `executeClaim` and cleanup around one owner.
  Compare/execute against captured normalized state, not a later re-resolution.
- A new private dependency/identity helper is preferable to growing an unrelated
  policy branch tree in `transaction-batch.ts`. A private asynchronous comparison
  helper can keep timers/wakeup management out of SQLite mutations. Any emitted
  private modules need installed-package allowlist and denied-deep-import tests.
- `transaction-sequence.ts`: project outcomes after the store establishes internal
  creator facts. Do not use projected results/problems as the binding registry.
  Its stepping seam remains private; no HTTP, discovery, auth-policy expansion or
  exports are part of this slice.
- Shared `member-identity.ts`, `semantic-identity.ts`, alias/Resource evaluators and
  `sequence-executor.ts` are evidence/reusable helpers only. Changes to shared
  modules require root coordination. Do not import the RU admission engine.

## Smallest safe implementation sequence and done-when

1. **Pure templates plus durable facts:** add the private representation and exact
   normalization/partial-comparison tests; add schema refusal and fenced fact/edge
   primitives. Keep public preparation's dependency refusal until later. Done when
   facts survive creator failed-receipt expiry and old-owner writes cannot bind a
   replacement, with no existing direct/carrier behavior regression.
2. **Direct unresolved comparison:** implement bounded async comparison over that
   representation with deterministic clock/wakeup tests. Done when provable
   mismatch is immediate, possible equality never leaks a pending receipt, timeout
   leaves the reservation intact, and retraction competition keeps one deadline
   and fresh ownership. No open SQLite transaction during a wait.
3. **Enable canonical dependencies:** remove only the `@name` refusal after the
   prior seams are executable. Implement creator facts, unbound retained failure,
   transient precedence and ordered execution for all supported reference slots.
   Done when the cases below pass through the actual owner/stepper, not only helper
   unit tests. Keep Resource-alias preflight refusal and fixed schema-free/unowned
   Types restriction; no full Transactional claim.

The three steps may be reviewed as local commits in one isolated clone. Do not
publish a partially enabled dependency path between steps. Each is useful work
while the alias clarification is pending.

Required executable acceptance matrix:

- Bead creation → Link endpoints; Link creation → update/deleteLink; Bead creation
  → update/deleteBead and putAlias target; pinned endpoint; multi-hop creator chain.
  Each commits independently and concurrent unrelated writes can interleave.
- Renamed labels, exact-number equivalence, omitted defaults/empty context, and
  equivalent canonical singleton/one-op batch retrieve the same dependent receipt.
  Different pins, guards or other known fields conflict, including while unresolved.
- Allocated and supplied-ID creator failure both give retained binding-unavailable;
  later success under reused creator key conflicts with retained unbound dependent;
  another permanent creator failure compares equal. Exercise multiple dependencies
  and repeated reference slots, not just one source slot.
- Creator success then detail expiry still binds; creator failure then forgetting
  and key reuse cannot change the original unresolved dependent. Dependent terminal
  identity remains stable through its own expiry/retention lifecycle.
- Foreign pending creator yields transient dependent with no dependent-key lookup,
  even if the dependent key already conflicts. It retracts only its own unknown-key
  reservation. Foreign pending dependent wins over static mismatch when creator is
  stable. Overlapping presentations never split an all-unknown key set or double-run.
- A dependent reads current Resource/guard state after creator commit, including
  deletion and update by another request. Late permanent failure retains only its
  own failed outcome; earlier independent commits remain.
- Inject faults before/after admission, at creator effects/fact/receipt commit, and
  at dependent finalization. No graph/group/fact partial commit. Crash/reopen
  retracts pending tails without auto-resume; graceful close alone is not the crash
  test. Retry replays prefix and admits the tail anew; stale tokens stay fenced.
- Direct possible-match wait resolves to original receipt/conflict; timeout retains
  no new receipt and does not retract owner; no DB lock held; retraction allows only
  fresh same-deadline competition; repeated owner churn does not reset budget;
  epoch change stops waiting. Sequence overlap remains nonwaiting.
- Existing alias witness, retention, Resource/batch/singleton tests remain green.
  Frozen-format refusal is byte-identical and same-format restart works. Pinned
  build/typecheck/boundaries/format/lint and installed-package smoke remain required.

## Alias member-turn timing: narrow clarification for the next slice

The observable behavior is already clear: lines 4110–4113 require first alias
resolution at the member turn, including observation of an earlier `putAlias`.
The current Transactional admission sentence (1366–1368) permits normalized
identity or “the durable unresolved dependency form below”; that form (1391–1399)
is specifically described for not-yet-committed creators. It does not explicitly
say how an alias-only dependent reserves its unknown key before that first lookup.
Capturing its target/miss at admission would violate the explicit member-turn rule.

Recommended narrow clarification, for root to evaluate/materialize: a pending
sequence reservation may also retain unresolved alias-reference slots until that
member's first turn, preserving all other normalized fields. At that turn capture
its original locator→target/miss witnesses atomically with identity finalization
under the exact owner; all existing original-witness replay laws remain unchanged.
Direct comparisons against this unresolved form follow the existing finite
comparison rules. No new public receipt representation, new equality component,
new wait budget, or change to alias lookup timing is proposed.

This is a specification coverage question about permitted admission state, not a
request to choose new observable alias behavior. It need not interrupt canonical
creator-dependency implementation. Do not silently enable alias slots using this
interpretation until root has reviewed it.

After clarification, tests must include `putAlias x→A; updateBead alias/x`, alias
repoint/deletion between admission and member turn, a miss later bound before the
turn, and direct competition while the alias template is unresolved. Subsequent
retries still use the original witnesses across reference slots; fresh retry
locators resolve only for transient comparison. A sequence combining `@name` and
alias slots must resolve both at its eligible member turn without rewriting a
prior terminal identity. Alias put target continues rejecting alias spellings.
