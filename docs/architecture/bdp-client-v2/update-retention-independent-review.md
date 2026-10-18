# Independent retention fixture review

Frozen head: `e211324e462752a654670139efcee1b07d42afdd`, base `8c0e043f6bfc5dae8b66044721a0aee170d66c97`.
Isolated clean detached checkout: `/Users/dbox/repos/wt/vickie/bdp-update-retention-review`.
No blocking findings. No edits, commits, integration or outward writes.

## Runtime seam and lifecycle

The private authority defaults its optional clock to Date.now and uses it for actual startup expiry and the existing member/lifecycle clocks. member-executor still reads terminal(options.clock) when constructing the retained disposition and adds the unchanged 24-hour retention interval; sequence-executor uses that same member clock for expiry. Read pagination retains Date.now. This is a clock input seam, not a manufactured timestamp/outcome or direct store modification. The fixed fixture clock changes only on reopen, so this plan does not distinguish startup time from a later terminal time within a continuously running process; correct terminal clock threading was verified in source, not claimed as a separate timing experiment.

No public DevelopmentServerOptions/configuration/CLI flags/package exports changed. The public development facade still calls the private runtime without scheduler/clock overrides. No shared store, evaluator, schema, minimum retention floor or public profile admission change. The generated source-only child passes its epoch to the private runtime; no actual database, retained outcome, Resource or tombstone fields are seeded/backdated.

The action owner binds controls to its active disposable Scope and retention mode. It refuses malformed inputs, extras, unsafe/noninteger/negative time, backwards time and elapsed values beyond seven days. Ordinary process controls are refused in retention mode; retention controls reject absence of that mode. The generated harness additionally requires a controlled child and safe nonnegative epoch. Reopen waits for normal close and zero exit before opening the same owned store/Scope/port; cancellation checks cover after close and after startup. Failure cleanup kills/awaits all owned children before deleting support/store directories. The existing IPC/HTTP/exit deadlines remain bounded.

## Observations and claim limits

The plan creates all retained outcomes through real HTTP. It observes a retained original creation postimage before expiry while live state differs, and a retained failed deletion while its subject is now live. At the fixed boundary, creation returns expired without its original outcome, changed semantics still conflict, and the captured current revision still guards a same-value update. The failed key executes a first successful deletion against the live subject; read-back proves deletion and the new retained deleted outcome survives an additional reopen. Before-boundary retry does not renew the tested boundary.

At-boundary behavior matches the existing recovery store's retain_until<=now expiry. Failure forgetting is permitted behavior of this implementation, not a requirement that all implementations forget at exactly 24 hours. No advertised retention.idempotency promise is made: retention-minimum remains omitted. Live periodic maintenance, other successful mutation categories, alias witnesses, allocated expired-creator bindings, internal tombstone fields and restore are still explicitly unqualified. Selected response fields, state and guards are tested, not whole response equality or unobservable internal lookups.

The prior 27 plans are unchanged. Exactly two rows gain bounded observations; three still-omitted rows only receive more accurate omission reasons. Inventory 57 bounded / 24 omitted remains 81 rows. No complete-row/profile qualification claim.

## Independent validation

Pinned Node 24.16.0 / pnpm 11.20.0 frozen install and build completed before process tests.
- 242 tests passed across eight files in 8.30 seconds: matrix (including 12 new malformed/control/failure/abort negatives), original HTTP/disconnect tests, diagnostic inventory, Read+Update catalog, sequence executor, Read/sequence lifecycle and recovery store.
- Saved actual report: 28 plans / 302 actions / 974 passing assertions; claimEligible:false.
- Typecheck, boundaries, formatting, lint, evidence verification, installed package smoke and development smoke all passed. Lint retains baseline 3 warnings / 1 informational finding.
- Existing Read seal verifies unchanged: 6142cf73deae687bc3e9668cc014ea408fa02843, run 4f0ed579, evidence 72bdaad, 89 rows. No new Read observation or reseal.
- Clean checkout and no surviving owned test subprocess. No full repository suite run; future integration/new-head CI remains required.

External logs and structured outputs are adjacent with update-retention-review-* prefix. Author report remains separate and unchanged. No fixes requested by this bounded review.
