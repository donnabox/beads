# SSE plan council dispositions

Target plan7a7f4c638216dfba2bf151dce950969f87d07ad682a16dbdba2e7fa1f563b723 against dfcf. Native root review and Claude90497 are complete. Source authoring remains gated behind S1b and H1b integration; this document assigns bounded plan corrections, not protocol rulings.

## Accepted corrections

Native1: choose effective cursor from the representation actually selected; mixed Accept with JSON winning must keep ignoring Last-Event-ID. Pure selection may precede owner validation, but rejection/error disclosure order remains intact.
Native2: heartbeat comments share the same backpressure pump; no accumulating periodic writes while waiting for drain. Native3: post-header errors destroy/close rather than reenter finite500 handling, and pre-abort/admission-registration/disconnect paths cannot leak a live slot.

Claude F1: correct numeric wording. stringifyJsonValue uses JSON.stringify primitive number spelling; admitted numeric value fidelity and original lexical spelling are different. Reuse finite serialization for exact canonical group bytes and safe immutable JSON handling. Test numeric values and admitted representation equality, never survival of original1.0e2 spelling.
F2: name the exact changefeed SSE406 test that becomes supported, and preserve Event-source SSE406 until G2b. Selection belongs only to the served changefeed branch; generic JSON/204 behavior remains unchanged.
F4: keep the existing explicit query-intent requirement in G2a, including header-without-query400. The frozen finite test alone cannot prove SSE semantics, but the shared normative requirement that omitting both is an error supports preserving it. Only start=now-plus-header remains a deferred interpretation. This narrows the native review's earlier uncertainty without relaxing the parser or adding a claim.
F5: inventory and reuse the existing Node disconnect-signal helper pattern in index.ts. Coordinate any narrow private extraction without changing public exports; do not duplicate it unknowingly. Abort is normal cancellation, not a fabricated owner500. Pin unaffected finite minimum reads and real disconnect behavior.
F6: choose raw header/query syntax before observation as appropriate, but validate effective checkpoint epoch/view inside the existing first service.changes owner entry. Do not obtain a separate context or export a new codec merely to satisfy inaccurate pre-admission wording.
F7: cite existing narrated SSE GET/HEAD rows and run transactional-http-contracts.test.ts. Preserve exact no-store,no-transform and required context/no validators. Vary:Accept is legitimate representation metadata; absence from an illustrative fixture is not itself a ban or a reason to remove it. Add an actual-wire assertion; change normative fixtures only if a concrete schema conflict is demonstrated.
F8: separate cancellable stream registrations from finite active promises; service close first rejects new registration/cancels streams, then drains tracked work. Distinguish ordinary shutdown from owner fault in diagnostics. Preserve mutation drain.
F9: enumerate installed request, header, socket-idle and keepalive deadlines. Establish actual pinned-Node lifecycle behavior by bounded observation rather than guessing. New raw stream recorder has a disclosed stream observation budget longer than the idle witness; do not widen existing finite5s exchange deadlines.
F10: preserve x-bdp-development:non-attesting on actual stream headers/delivery. No native-browser EventSource observation or CORS support is claimed.
F11: preserve signatures and behavior of current read-http exports/consumers. Any private additive helper gets inherited Read/RU/Transaction regression tests.

## Reviewer claims not accepted as new protocol requirements

F3 proposes rejecting Last-Event-ID:genesis because it is never emitted on a later group. The existing spec explicitly reserves genesis as the distinguished initial checkpoint; never assigning it to a later position is not a ban on client submission as a checkpoint. An initial/genesis intent's lack of a later-position epoch tag is deliberate, not independently an erasure bypass. Do not invent invalid-parameter solely from this review. Add explicit genesis-header coverage/disposition under a supported query-intent combination, preserving the current checkpoint codec; leave the already-deferred start=now-plus-header combination deferred. Future erasure/replay-floor checks must apply to the effective position including0. No erasure capability is claimed now.

F12's lack of an SSE observedHeadPosition field is not a demonstrated ambiguity: caught-up eligibility is an authority-side fence, and the owner already supplies current head per poll. Preserve that internal context for future fence work; do not add a wire field/event or new product question solely to expose it. Client finite-read catch-up and authority stream eligibility are distinct purposes. Existing required erasure fence remains unimplemented/disclosed.

## Next action

After H1b repair freezes, its author can revise the original plan with these native/external dispositions, preserving reviewed7a7 bytes. Actual G2a authoring waits for a root-assigned integrated Snapshot/H1b base and exclusive service/HTTP/runtime ownership. No need to repeat the entire council for mechanical clarifications; genuinely widened seams or new semantics require focused review.
