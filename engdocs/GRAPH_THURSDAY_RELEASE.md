# Thursday blog release: Memory, mixed graph and script access

Donna's September 28 team-call decision sets the release target for Thursday, October 1, 2026: the blog points readers to `versioned-beads/beads:integration`, where the following workflow must actually run. This is a release checkpoint within the [full graph delivery plan](BEAD_GRAPH_DELIVERY_PLAN.md), not cancellation of its remaining destination. Source transfers and landed hashes remain in the [landing ledger](GRAPH_INTEGRATION_LANDING.md).

## Acceptance on the actual integration commit

| Requirement | Observable acceptance | Current gap |
|---|---|---|
| Memory create, update and delete via CLI | Fresh normal initialization, create a Memory, read from another process, update it, delete it, and verify current reads/enumeration exclude it. Failure, stale guard and Link interactions have explicit tests. | Create/read transfer compiles but is not qualified on Jim's branch. Update is reusable from the preserved candidate. Memory deletion is new work; Link unlink is not Memory deletion. |
| Memory–Memory and Memory–Issue Links via CLI | Create both relationships using an installed Type, enumerate/read endpoints and Links, change/unlink through admitted commands, verify persistence across process restarts. | Reuse qualified mixed-Link work and reconcile the target's current Issue writers. Specify/test deletion with incident Links before a public contract is adopted. |
| Useful Issue CLI compatibility | Publish an exact tested command/flag matrix. Minimum intended workflow: Issue create/show/list, selected update, ready/blocked, blocking Dependency add/remove, close/reopen. Ordinary Issue workspaces keep their existing behavior. Unsupported graph commands must refuse before legacy writes. | Four old wrappers would double-record versions against current target writers. Reconcile exactly-once/no-op behavior and test the advertised subset. Breadth beyond the documented subset is not required for Thursday. |
| Read/enumerate from Python or shell | Check in runnable Python and shell examples that call the real BDP HTTP endpoint to enumerate Memory and Issue records and read a record by identity. Exercise pagination, response/error handling and actual protocol values against the installed release server; do not shell out to the CLI for record access. | BDP Read serving and independent public-client/script round trips are mandatory for Thursday. Reuse qualified collection/selector/pagination work on the new combined source. CLI JSON is not an acceptable substitute; full HTTP Write and full History remain separate. |
| CI covers every added/changed surface | Required destination CI executes new graph tests on embedded and ordinary shared-server Dolt, installed fresh-process workflow, CLI compatibility, script examples and relevant failure/concurrency controls. Missing tools/server, zero discovered tests, or an unexecuted new test must fail the relevant gate. Keep source/head and artifacts identifiable. | Normal upstream CI alone is not proof the optional graph tests ran. Wire dedicated gates and audit discovery/runfiles/env/skip policy for each transferred package. |
| Agent usage through Steph's skills/hooks | Run Steph's actual delivered artifacts in an isolated end-to-end scenario: an agent records/links Memory and another session retrieves/updates it through the supported interface. Verify resulting records and ordinary Issue behavior. | Steph owns skills/hooks implementation and is integrating today or tomorrow. Per Donna, do not work on or block today's CLI/CI work on this dependency. Pick up her integration when available; do not substitute a hand-written skill or mock and call it passed. |

The exact command spelling and incident-Link deletion policy must follow reviewed proposal decisions. Reversible implementation may proceed; unresolved durable semantics return to Donna before adoption. Historical retained-state handling must be stated precisely; deletion must not accidentally promise irreversible erasure or identifier reuse.

## Execution order

1. Land the small graphops/wire prerequisites and the proven recorder-state correction when their combined-source CI is green. Mirror real merge commits and preserve original PR provenance.
2. Qualify and land installed Memory initialization/create/read, then its existing update path and mixed Issue/Memory/Link workflows. Add the missing Memory deletion behavior with coherent guards, Link handling and retained-state proof.
3. Integrate BDP Read serving and deliver Python/shell examples that read and enumerate through BDP, including a multi-page collection round trip. Publish the exact supported Issue CLI matrix separately.
4. Make the whole Thursday journey a required CI integration scenario on both storage modes. Retain focused tests for the independent failure boundaries; do not replace them with a happy-path recording.
5. When Steph's artifacts arrive, exercise them in the same release checkout. Publish the final tested integration commit, setup commands, examples, supported subset and explicit limits before the blog points at it.

Wednesday is the intended point to stop adding release scope and run the complete installation/user journey. Thursday publication depends on those results; dates are commitments to pursue, not a claim of completion. Report missing requirements rather than calling component tests a release.

## Boundaries

Full chronological/native History, the unresolved atomic commit-stamp primitive, complete authenticated HTTP Write/durable outcomes and production adoption/recovery remain in the full plan. They are not additional Thursday requirements unless needed by the six behaviors above. Preserve held PR44/45 decisions and existing blogs. No engine forks, fabricated timestamps, application locks or new open-ended infrastructure work.

The original September 23 attempt clock and missed checkpoint remain recorded. The Thursday target does not silently restart that clock. Keep implementation planning/review in Donna's fork; CLI shape feedback remains upstream Proposal6703.
