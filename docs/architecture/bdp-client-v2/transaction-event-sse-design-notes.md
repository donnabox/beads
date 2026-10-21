# G2b preparation: Event Source SSE

Root research notes, 2026-10-13. Supporting material for the sole fork design.md; not an implementation assignment or settled design. G2a changefeed SSE is still being authored. These observations are from frozen0944fbaa19ab2b3399a7547eb6acf74ed9f5c5ff. Reconcile against its reviewed G2a successor before proposing exact file edits.

## Required behavior already in the spec

The Event replay and live observation section requires both the discovered Scope events/ Resource and Bead/Link ?view=events Sources to support SSE. Event frames repeat the complete JSON Event id/type in their id/event fields. Public after is exclusive and source-local; absence begins with earliest retained Events. Last-Event-ID overrides the original after only for live representation. JSON must preserve its existing behavior. No new log, cursor wire format, subscription API, or browser authentication promise is needed.

Use G2a's reviewed media selection, private conditional handling, raw duplicate-header handling, bounded frame delivery, separate stream cancellation registry and actual runtime settings. Do not copy the pump or invent a second lifecycle. Preserve unknown-source disclosure ordering, retained deleted-Resource Sources, exact initial context, HEAD/conditionals, no-store/no-transform and Vary. Native EventSource transport progress is distinct from replication checkpoint durability; Event Sources still cannot supply erasure reconciliation alone.

## Existing owner seam and performance risk

transaction-store.ts readEvents validates retained source identity, minimum and context, then parses/verifies the public Event cursor. It traverses committed groups one at a time and projects with transaction-events.ts. It returns a finite page plus same-entry context, but no private scan continuation. The initial position is last public Event's group or1. A sparse or empty Resource Source would rescan unrelated history on every polling tick if G2b merely repeated service.events(source,lastEvent).

The implementation plan must explicitly handle this. An internal owner-issued scan continuation may record how far this particular source/context has been examined, without becoming a public Event ID or acknowledging an Event not sent. Validate a supplied public cursor before entering that path. Bind progress to source, epoch/view and owner lifetime, preserve replay/corruption checks, and invalidate it on context change. Do not turn a synthetic head checkpoint into an Event cursor or emit an invented head Event. Decide whether to reuse bounded finite pages or yield one Event per owner observation; document the resulting memory/read-ahead bound. The same-group ordinal must survive interruption so a cursor between two source Events never skips the second.

No SQLite iterator/lock may escape an owner entry or survive a socket drain. No scan progress may pass an Event omitted merely because a page/frame bound was reached. An empty completed scan may advance internal position only using the owner observation that proves it empty. New commits between scan, admission and next poll must remain observable. An independent probe should count inspected historical groups across repeated idle polls and show that sparse-source cost does not grow by rescanning the entire prefix each tick.

## Evidence to require in the eventual bounded plan

- Scope, Bead and Link replay and live frames match separately observed finite Event records, including stable source-local IDs and same-group ordinals. A self-Link yields both endpoint facts in its Bead source. Deleted subjects retain their Source while retained identity permits it.
- Empty/sparse Sources remain live; unrelated mutations do not fabricate Events or change a remembered public Event ID. A later relevant mutation arrives once. Multiple relevant Events in one group are not skipped across page limits or reconnect.
- Header override, source-mismatched/stale/foreign/malformed cursor handling, JSON-vs-SSE negotiation and current-context conditional ordering have discriminating actual-wire cases. Do not reuse a changefeed checkpoint as an Event ID.
- Backpressure, disconnect, pre-abort, service shutdown and post-header faults reuse the one bounded pump and prove slot/owner cleanup. Installed tarball tests use the dedicated bounded stream recorder, not the finite executor.
- Read/RU, finite Events/changefeed, Snapshot and G2a behavior remain intact. No public admission or whole-row claim until independently audited exact-source evidence.

Metadata-only incident-Link membership remains the existing explicit product gap; do not resolve it through transport work. Required erasure withholding/publication fences and mutable-view/replay retention remain later work, not claims supplied by all-visible no-pruning stream tests. The plan must say which of these limitations remain observable refusals and which triggered behavior is unavailable in the configured owner.

Next: after G2a freezes, assign a bounded implementation plan with concrete private owner/service/pump file ownership, council review and exact installed done-when predicates. Do not begin store edits concurrently with the G2a writer.
