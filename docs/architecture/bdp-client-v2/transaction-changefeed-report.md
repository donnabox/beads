# G1 finite changefeed checkpoint

Frozen clean head **a7430ddf0fc2e56b0923d8c75fcfc752ee89ac88**, direct parent **8569b85d785f96325ca948580106f454c48ed034** (finite Event predecessor containing materialization076201849f98eeab71bad5c8404fb693a36c6474).
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-changefeed`, branch `codex/vickie-transaction-changefeed`, origin GitHub. Local commit only; no outward writes. Parent/frozen review sources unchanged. Event and materialization reviews may still produce ordinary-integration follow-ups; this is a candidate for independent review, not a declaration those reviews are finished.

## Implemented boundary

Seven files: new transaction-changefeed.ts/test.ts; owner/service/HTTP seams; private smoke output/denied-import inventory; implementation progress doc. No runtime listener/configuration, mutation evaluator/admission, schema, formats, public exports, profile admission, installed evidence or normative edits.

- Private canonical changes/ finite GET/HEAD serves one complete committed group, absolute next or null, explicit after/genesis/start=now, current epoch/view/head context. No all-history serving array or diagnostic inspect path.
- Shared owner-internal retainedGroupAtOrAfter extracts the existing finite Event one-row query/byte check; both paths check BLOB length before copying/parsing. The extraction preserves Event framing and cursor behavior. G1 additionally checks predecessor adjacency, projectionAdvance=false, nonempty visible facts, empty actual erasures and duplicate StateChange identities. Scope projection validates zero-based contiguous all-visible ordinals and reuses exactly the Event Source renderer.
- Outward group checkpoints use deterministic bounded Scope+epoch/view tags and internal position; raw group records remain unchanged. Outward position fields remain pos-N, genesis position pos-0, reserved checkpoint genesis. Current-context future/unrecognized tokens are local400 parameter refusals; actual missing committed rows within the no-pruning prefix fail as owner corruption, never silently skip or claim client error. Epoch/view mismatches use existing410/409 semantics.
- start=now captures admission head once before the service minimum loop. Subsequent attempts observe a fresh current prefix while retaining that original anchor, so waiting cannot skip intervening commits. Body/head/current fields are paired within one owner entry, and tracked finite work drains on close.
- Finite JSON ignores Last-Event-ID; it does not supply or override explicit intent. Shared HTTP logic provides ordinary conditional/HEAD behavior and406 for SSE-only Accept. Current canonical minimum controls wait/refuse before conditional shortcuts. Unsupported tuple forms remain explicit bodyless private gaps after authentication. No new cancellation subsystem or CORS policy.

## Bound and evidence

Source tuple remains M64/E256/O32/R262144/Q65536/S65536. G1 imports the Event module's source-derived stored-group bound253,493,248. Conservative public source/ID addition16,851,456 plus page/header138,264 gives **270,482,968 encoded bytes**, below actual pinned Node24.16 MAX_STRING_LENGTH536,870,888 by266,387,920. Module guard derives the bound from actual constants. External `transaction-changefeed-bound.mjs/.log` pins clean8569 source and performs arithmetic only; no giant proof allocation. This is not a bound on live heap, concurrency or accumulated history. No transaction is rejected retrospectively because of a newly invented response budget.

Actual-owner tests include256 small-body Events in one whole group despite receipt page maximum2; a200KB+ postimage; a40KB canonical Scope with bounded Event/checkpoint IDs and absolute next; staged mutation rollback/no-op/alias absence from history; owned/self-Link/reversal/create-delete facts versus final postimages; exact Scope Event identity parity; restart; corruption; reentrant write refusal during group read; uncertain terminal-COMMIT sticky fence and recovery. New fault injection is a controlled DatabaseSync error after COMMIT, not power-loss testing.

A test-only replica stages changes then atomically publishes graph+checkpoint. It compares exact authority Resource states, checks member-for-member owned inline/first-class agreement, refuses conflicting upserts/tombstones, context/gaps/reorder/conflicting duplicates, and survives a publication failure by replaying from the unchanged durable cursor. The oracle never parses/orders position spelling. It is not a production replication client or Snapshot implementation.

Actual loopback HTTP tests exercise pages/absolute next, current headers, GET/HEAD/content-length,304/412, epoch/view/timeout direct bodies against normative transactionalProblem, authentication/method routing, ignored Last-Event-ID, required intent, SSE-only406 and ordinary reopen. Stable configured Scope differs from loopback listener as in the existing fixture; no public discovery is fabricated.

## Verification

Pinned Node24.16.0/pnpm11.20.0, frozen install, build before emitted/process tests.

- Final Transactional regression command: vitest run packages/server/src/transaction-*.test.ts --maxWorkers=2, verbose plus JSON.
- **428/428 tests pass across17 files,7.35s**; new feed file20 tests. `regressions-final.log/.json` preserve exact results. This includes Event, receipt, sequence, owned, set, installed contract, minimum, fault and real hot-journal recovery regressions.
- Final build, typecheck, dependency boundaries, format, lint, both installed package smokes and historical89-row Read evidence verification pass. Lint has34 inherited warnings/one info and no errors. The new private module's emitted inventory is expected and both bare/dist package deep imports are denied. Public profile admission remains fail-closed.
- Initial focused failures were test-fixture mistakes (wrong reader body field, using alias in the set/batch preparation seam, extra fields in minimum tuple, singleton wire carrier shape), corrected before final runs. Initial type errors/explicit-any warnings were also test-only and resolved; logs retained. No stopped/hung processes or inconclusive runs.
- Broad whole-repository suite not run; root integration may do that. No evidence artifact regenerated/resealed.

## Remaining contracts and review dependencies

SSE/Event follow, Snapshot creation/streams/rendezvous, discovery changes/headPosition/minimumReplayPosition, pruning/late cursor expiry, hidden projections/view changes, erasure, broader minimum controls, set HTTP admission and installed cumulative TX qualification remain open. Empty ledger and all-visible policy do not qualify hidden/erasure branches. Actual Event metadata-only incident ambiguity does not affect Scope Event projection reused here and remains its own explicit gap.

Any later Event review fix or materialization evidence/tuple follow-up must merge normally into a new successor or integration candidate and rerun affected tests/bound calculation. Current proof is pinned to the actual tuple, not a promise that future counts stay unchanged. No harness metadata expansion in this slice.
