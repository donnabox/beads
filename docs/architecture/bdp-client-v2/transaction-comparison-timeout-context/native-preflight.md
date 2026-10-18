# Comparison timeout draft: independent native preflight

**Hold execution for one Medium control correction.** No production blocker was identified by this static review. The proposed complete Event-page substitution is sound for this fixture. Once N1 is corrected, preserve this initial preflight and regenerate successor control/bounds/preflight hashes for narrow closure before root releases execution. Nothing below is a test/build pass or code-council signoff.

## N1 — Medium: unobserved-field mutant is hidden by submission projection

Location: `controls/leak-unobserved-fields/source.patch`, `measure.py` control tuple, and bounds control entry; production `transaction-comparison.ts:55` versus `transaction-submission.ts:181–188`; selected test `transaction-submission.test.ts:1110–1138`.

The mutant adds `context: undefined` to comparison's `{kind:"temporarily-unavailable"}` result. Its selected test calls submitTransaction. Submission's switch reconstructs a new `{kind:"refused",status:503,code:"temporarily-unavailable",retry:"after-delay"}` object without copying that field. Consequently the selected `toStrictEqual` assertion remains satisfied and this control cannot produce its promised `legacy comparison result exact` failure. Callback non-invocation assertions also remain satisfied. This is a static dataflow contradiction, not a suspected runtime failure.

Smallest correction: mutate the legacy temporary-unavailable **submission return** by adding the enumerable `context: undefined` there, keeping the same selected strict submission assertion, zero HTTP/children and same18-case source selection. Update the generator, patch path/hash/mutant hash and prose to label this a submission-boundary exact-shape control. Preserve original artifacts. Do not claim that it proves direct comparison's undefined-field sensitivity.

No separate new test/cohort is required for this bounded correction. The selected dependency and alias tests already call comparison with a full owner and assert its legacy result using `toEqual`, and the actual comparison source visibly returns the old bare object when unobserved. Those existing assertions detect defined observed discriminator/context leakage but do not distinguish an extra undefined property; their limitations must not be relabelled as strict direct-comparison coverage. The accepted required sensitivity can be satisfied at the selected strict submission boundary without expanding the footprint or request budget.

## Exact source and artifact verification

Read root-authoring acceptance, accepted plan941f350f, dispositions994e5fa3, closureb2a637e8 and measured preflight. Worktree HEAD is c555fe01df5cd20454fb0d663431df816c1270aa, tree062a3ab7e24236417bd915b373a4b44ba9908647. Exactly eight authorized paths are dirty,550 additions/35 deletions. Independently verified all eight current and base hashes and byte-equality of live git diff with the recorded source.diff.

- preflight.md:2aba0c1f70a31bdd0ccfc801945c337f146c397622923af20d8ba83b3bd73203.
- preflight-bounds.json:e0816f18a489ad29130b63a060c2da1f86be1d1f370eaaa98055e3675349e1cc.
- source.diff:27fc1610da057443639bbd49012e73ec2e9f951f4fdfc20d8801435922809cbe.

This is an uncommitted draft, not a frozen implementation. No source/build/install/test/service/network/package execution was performed. Reviewed pinned Vitest source from the already installed adjacent schema checkout read-only to resolve reporting behavior; no dependency setup was performed here. Only this review artifact was written.

## Production path assessment

Comparison checks epoch before the original deadline; it invokes the observed callback synchronously only after a healthy deadline decision. No owner entry spans wait and no deadline is reset. Store observeComparisonDeadline uses enter plus existing atomic, reads persisted epoch and current position in that boundary, and returns frozen context. Normal closed becomes owner-stopped; reentrancy/fenced/metadata errors that reach the operation propagate. The earlier comparisonEpoch maintenance-fault masking residual remains explicitly documented. No supported live epoch rotation is invented.

Submission alone binds the owner method. Both the inherited options type and actual spread/override exclude a caller observeDeadline field; service/runtime inherit the narrowed submission options without edits. Unobserved full-owner use does not opt in merely because the method exists. Observed deadline and owner-stopped remain distinct through the private result union. HTTP requires deadline context, preserving the existing sanitized500 path if it disappears, and keeps owner-stopped as bare503. Canonical Problem/retry/Retry-After behavior and earlier route exclusions are unchanged. No arbitrary tuple, extra admission, receipt lookup, minimum wait or public configuration is introduced.

No obvious compilation obstacle was found in the changed types or branches; actual typecheck/build remains required. The selected owner spy asserts one BEGIN IMMEDIATE/COMMIT and exactly two metadata prepares, while independent work proves the captured position advances and the previous frozen tuple stays unchanged.

## Event, graph and receipt proof

The stage0 one-change-page premise was incorrect and is preserved as history. The replacement uses finite `events()` source in transaction-store.ts:1550–1609, whose loop traverses retained groups until actual item/byte bounds; it is distinct from the single-group streaming eventPacket and changes paths. Five small facts fit the configured32-item finite page bound. Actual transaction-batch.ts:738–771 emits Bead created, Link created and source/target linked facts in the asserted order. The independent Bead gives counts0→1→5. No metadata-only Event ruling or change pagination claim is needed.

The test gets a real initial empty Event page and pos0; during exactly one comparison wait it commits independent Bead b, then advances the controlled clock. The observed timeout must be503/currentpos1, unchanged epoch/view, exact canonical Problem and no Location. Before release, complete graph pages and exact one-fact Event page plus missing a establish no candidate effect. Finally releases the original sequence, validates its result, then retries the dependent key. Replay remainspos3 and equals its original result; complete graph and exact five ordered Event facts exclude duplicate mutation. Expected semantic fields are pinned independently; dynamic revision/provenance/Link identity come from previously validated representations, with next:null and unique Event IDs. The receipt inventory separately closes runtime and reads actual durable principal/key/receipt rows, requiring creator/dependent/independent only and full replay equality. Graph state is not used as a receipt-inventory substitute.

## Selection, controls and cleanup

Independently matched each regex against literal direct it declarations: store6, submission5, HTTP2, dependency3, alias2 =18 selected cases. No selected parameter loop or hidden child scenario was found. Nine cases are newly added and one existing HTTP case is replaced. The existing basic HTTP test is four requests; the expanded case is12, so total16/two listeners. Its counters cover dispatch/completion; the independent response completes before the waiting comparison response even though it is dispatched later.

The new HTTP test holds two scheduler deliveries, asserts one comparison wait, releases queued callbacks in finally and awaits original sequence before propagating failures. It retains both primary and drain errors. Pinned Vitest4.1.10 runner chunk-artifact.js:3055–3073 flattens AggregateError.errors into processed test errors, and JSON reporter index.UpGiHP7g.js:3573 emits each error stack/message. Thus the named inner presence/currency/status assertions remain available for adjudication; the wrapper does not mask them. Require actual reports to show only the intended oracle and successful cleanup, not merely nonzero exit or the wrapper title. The preserved basic timeout test still uses its existing callback/drain behavior; no new safety claim is made for failure injection into that older case.

Independently reconstructed all five patches in memory and verified their mutant hashes. Remove-wrapper reaches the required presence assertion after four HTTP dispatches; stale-position reaches currency after four; lose-tuple reaches healthy503-versus-sanitized500 after four; final-epoch-check reaches its zero-HTTP strict mismatch assertion. N1 is the sole unreachability issue. With its narrow correction, expected controls remain12 HTTP/max36 across three listeners, five runs, no installed children. The final mismatch test is explicitly defensive rather than a live authority transition.

## Bodies, routes and execution limits

All12 ledger body byte counts and SHA256 values recomputed correctly: maximum277B, total629B, with116B candidate sent twice. Routes bind the actual fixture Scope: events, sequence/batch, complete bead/link collections with explicitlimit32 and a single bead absence check. No extra receipt fetch, reopen, discovery or remote target is introduced. Source response figures39936B/479232B are clearly prospective fixture estimates, not enforced Fetch caps or measured heap/wire totals. Retain that distinction in final evidence.

The proposed existing installed regressions remain direct14/one child plus sequence13/one, total27/two; source files still enforce those request ceilings. They preserve65536B request,16384B headers,131072B response,8192B target and5s/90s/120s/10s deadlines. These cohorts are regressions only and do not exercise default installed comparison timeout. No new app, controlled package cohort, predecessor timeout pack, retry or broad workspace/matrix run is authorized. Build must precede emitted HTTP consumers; controls need separately built derivatives. Actual counters, exits, selected failures, cleanup and package identities remain pending root release and execution.

Apart from N1, this bounded static review found no further blocker. Source/config/public admission scope is preserved. Root retains the measured-release, frozen code-council and integration decisions.
