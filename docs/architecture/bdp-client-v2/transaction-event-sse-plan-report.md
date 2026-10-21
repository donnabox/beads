# G2b Event Source SSE plan completion

Status: read-only council proposal complete; implementation not started or authorized.

Plan: `transaction-event-sse-plan.md`.
SHA256: `bacc2b820d32c83b7dcdf20d84b25ac38df600d5ae40bb2f6d70398b168a4187`.
Frozen source inspected: `b64d7e46652739c6436259a8c8ed4c5e8049a077`, `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse`, branch `codex/vickie-changefeed-sse`. Source remains clean and unchanged. Root must select the reviewed integrated implementation predecessor after any G2a review repairs.

Remote launch session was checked and remains `2026-10-08T20:26:35Z-81e5`; read-only guard evidence is `transaction-event-sse-resume-guard.md`. No outward writes were performed.

## Concrete proposal

- Owner-issued, WeakMap-authenticated private scan capability carries source/context/lifetime and exact group/ordinal progress. One owner entry inspects at most one group; empty-source idle polling does not reread the historical prefix. No persisted token, synthetic wire cursor, inspect adapter, SQLite iterator or open transaction crosses entries.
- Same-group progress never passes an unreturned Event. The shared pump adopts packet progress only after delivery and yields between empty backlog steps. Existing finite next/null lookahead is preserved through a narrow shared scanner.
- One G2a/G2b write/drain engine, one service stream slot, one bounded packet and pending frame, cancellable backlog/poll/heartbeat paths, existing shutdown and post-header failure behavior.
- Event Sources permit Last-Event-ID without after and replay earliest retained Events when neither is supplied. Genesis remains invalid as an Event ID. Finite JSON continues to ignore Last-Event-ID. Initial identity/minimum/cursor checks preserve source nondisclosure.
- Arithmetic separates 4,194,304-byte packets and 4,194,816-byte framed Events from inherited 169,345,024-byte parsed group bounds. These are not total heap or latency claims.
- Fourteen executable predicates cover owner sparse work, same-group boundaries, actual raw sockets, faults/backpressure/cleanup, exact packed installed Scope/Bead/Link evidence and a 15-second sparse-source idle observation under unchanged timers.

## Deliberate limits and next action

Metadata-only incident-Link projection behavior stays an explicit pending gap. Erasure/revocation/pruning fences and public profile/discovery qualification remain separate work. No Type, selector, schema, catalog, mutation or evaluator changes are proposed.

Root should council-review this plan, especially capability lifetime, finite equivalence, header nondisclosure and packet adoption, then assign an integrated base before implementation. The plan contains exact file ownership and staged done-when predicates.

## Preservation and process state

The original paused plan remains byte-preserved as `transaction-event-sse-plan-paused-deb205.md` (SHA256 `deb205bd0004caee44ab0e0536f7743c8081c3f5f73f828406090eadc315aee8`); `transaction-event-sse-pause.md` remains the historical pause report.

No tests were run for this read-only planning task. No owned background test, review, server or child process was started; none requires termination. Root-owned reviews were not altered. G2b source implementation is still unstarted.
