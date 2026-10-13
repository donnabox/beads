# G2b Event Source SSE — paused WIP, not implementation authority

Paused on the operator's request, 2026-10-13. Read-only inventory and tentative representation notes only. This document is incomplete, has not had council review, and does not authorize implementation. Sole fork design remains authority.

Inspected frozen source: `b64d7e46652739c6436259a8c8ed4c5e8049a077`, `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse`, branch `codex/vickie-changefeed-sse`. G2a remains unchanged. Root supporting notes: `transaction-event-sse-design-notes.md` (original0944 references rechecked at b64). No G2b source edits or tests were made/run.

## Verified current contracts and seams

- `docs/specs/bdp.md:5350–5373`: Resource Event Sources are deterministic projections of committed groups; stable source-local IDs derive from checkpoint/ordinal/projection; materialization/indexing may preserve that content and order.
- `:5547–5565`: Link lifecycle versus Bead lifecycle/incidence/incident properties membership; self-Link yields both endpoint facts.
- `:5629–5640`: retained deleted-subject Source may remain readable; unknown/out-of-view identities404; history-expired410 requires retained-history authority/disclosure policy.
- `:6130–6290`: Scope `events/` and Bead/Link `?view=events`, exclusive source-local Event IDs, omitted after starts earliest retained, SSE id/type repeat full JSON Event, Last-Event-ID overrides initial query cursor. Event delivery is application observation, not atomic replica or erasure reconciliation.
- `:6536–6559`: no ETag/Last-Modified for finite Event pages; GET/HEAD/conditions/current context and SSE cache policy.
- `:6563–6576`: ASCII token profile; genesis is a Scope checkpoint and is never an Event ID. Unlike changefeed's mandatory URL intent, an Event request can omit after; Last-Event-ID without after therefore has an existing initial-cursor interpretation.
- `transaction-store.ts:1370–1470` readEvents: source identity check before minimum/cursor checks, then context/minimum, cursor shape+epoch/view/source+actual membership validation. Reads groups one at a time, projects via `projectTransactionEventGroup`, applies32-item/4MiB page bounds, retains same-group ordinal continuation.
- `transaction-store.ts:560` retainedGroupReader: prepares two statements once per owner observation, bounded length before body fetch/parse, no escaping SQLite iterator.
- Current polling via service.events(last emitted Event) would repeatedly re-scan unrelated historical groups for sparse/empty sources. Public cursor validation cannot simply be skipped to fix that.
- `transaction-sse.ts`, service.stream registry, runtime raw header/disconnect bridge and media/conditional helpers from G2a are reusable. Do not duplicate pump/lifecycle, modify selectors, or change finite Event outputs.

## Tentative bounded representation (requires plan completion/review)

Prefer owner-issued **private scan progress**, not a synthetic Event ID or new persisted log. An immutable owner-lifetime capability can carry canonical source, epoch/view, next group position, within-group last inspected/included ordinal, and last observed head. An owner-local WeakMap can authenticate the capability without an unbounded strong registry. A reopened/different owner rejects it. Fixed all-visible view remains the only configured policy; do not invent principal-specific Event namespaces.

Initial entry validates source existence, minimum and the effective public Event cursor (including actual source membership), exactly as finite reads do. Subsequent steps validate owner capability/source/context and monotonic head inside the same owner entry. No SQLite reader, transaction or callback escapes across socket drain.

A candidate minimal step scans at most one committed group per owner entry and returns at most the existing finite page item/byte bound. Empty steps may advance private progress only through a prefix proved empty in that observation. A return flag distinguishes more historical groups from current end, so the pump yields between bounded backlog steps but polls only at current end. This avoids rescanning the full sparse prefix each tick and avoids monopolizing one synchronous owner entry over all history. Need finish exact helper factoring and benchmark assertions before authoring.

Within-group rule: if a relevant Event cannot fit the remaining packet, retain progress in that group after the last included ordinal; never jump to the next group. Only after all relevant Events in the group have been processed may progress advance to group+1. Stable ordinals are Scope-group ordinals and may be sparse in a Resource source. Self-Link endpoint ordinals must both survive page boundaries.

The pump must keep the packet's prospective progress private until all its frames have been handled. A blocked frame prevents another pull; abort discards the private packet/capability. Reconnect starts from a real public Event ID and revalidates membership. Empty scan progress never emits an ID/checkpoint/comment ID or acknowledges an Event not delivered.

Finite readEvents should reuse any extracted per-group scanning/projection/byte-accounting helper but preserve its exact existing page/next semantics. This is a narrow internal helper seam, not a second Event projection implementation. No storage format change or new table is currently indicated.

## Known remaining design details to finish

1. Exact owner API/capability shape and how to share finite byte accounting without changing finite next/null lookahead behavior.
2. Exact packet/pump adapter factoring: reuse one write/drain/abort/heartbeat engine for changefeed groups and Event records. Preserve G2a output and lifecycle tests.
3. Source nondisclosure for malformed/repeated reconnect headers: effective header validation must not introduce an earlier unknown-source disclosure decision. Determine a minimal invalid-input path through the owner's existing source/minimum/cursor order rather than inventing a public token. Query syntax validation retains current behavior.
4. Initial SSE response/HEAD/304/412 capability disposal, post-header owner/projection failure close behavior, raw header character/duplicate checks, and fixed-context invalidation tests.
5. Exact Event frame bound from current4MiB page/first-Event proof plus framing (not the changefeed186MiB figure), packet memory accounting, no strong-token leak, and no read-ahead while blocked.
6. Required installed witness file scope, bounded real idle/replay/reconnect evidence, and deterministic sparse-scan instrumentation (group-body reads counted per observation; no source-only synthetic installed results).
7. Finish executable done-when matrix and proposed ordered commits; no implementation before root/council approval.

## Explicit limits, not new policy

The metadata-only incident-Link update membership question remains unresolved in the existing Bead projection. It currently raises `TransactionEventProjectionGap("incident-link-metadata")`; Scope/Link sources can still project the fact. G2b must preserve that explicit pre-header private gap/post-header close, not silently skip the Event or decide membership.

Mutable authorization views, erasure withholding/publication fences, pruning/expired-source disclosure are unavailable transitions in this fixed all-visible/no-pruning owner. Do not claim those obligations qualified by streams. Unknown sources404 and stale/foreign public cursor checks remain executable. Previously scanned immutable history is not continuously rehashed; every newly inspected group retains existing integrity checks. No general guarantee against arbitrary out-of-owner historical tampering is proposed.

No browser EventSource/CORS claim, selector behavior change, public profile admission, catalog credit, Type lifecycle or optional History work. G2a review fixes, if any, take precedence and must be merged as root directs before G2b authoring.
