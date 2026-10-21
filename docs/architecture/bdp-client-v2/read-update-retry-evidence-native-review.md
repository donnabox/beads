# Native review: installed Read+Update retry evidence

Reviewed exact frozen `29e715162691e6d0a4bcc1ce699f33e243ebecdd` against `8e45ef4d2c9674296f217ac0cf75bc706f49ff09` in `vickie-read-update-retry-evidence`. The frozen author checkout remains clean and untouched. Independent build, tests and review-only sensitivity probes ran in `/Users/dbox/repos/wt/vickie/vickie-retry-native-review`. This review used the author report and approved plan/dispositions, not an external implementation review.

## Finding

**P2 — Compare complete retained failure objects, including unexpected fields.** `packages/conformance/matrices/read-update-v1.json:67907` (cross-carrier `f-replay`, with the same pattern at `g-replay` around line68700 and sequence `initial` around64562) checks the four original Problem fields plus selected absent fields, but does not reject an added `title`, `instance` or arbitrary extension. The normative Problem schema intentionally permits these fields/extensions; schema validation is not complete disposition equality. No original failure is captured as a whole object. The later repeated sequence result-array equality only relates the two reordered presentations, not their initial failure.

Four independent post-wire controls reproduce the gap, each applied exactly once and all leaving the selected row **pass**:

1. Add `title: "not the retained failure"` to cross-carrier `f-replay` member0.
2. Add `instance: "urn:review:changed-instance"` to singleton `g-replay`.
3. Add `reviewExtension: "not retained"` to cross-carrier `f-replay` member0.
4. Add a title to the sequence's initial failed member1; all later retries omit it, yet the sequence row passes.

The real target's original failures have exactly `type`, `code`, `status`, and `retry`, plus current `operationIndex` in a sequence. Thus these controls demonstrate a missing equality oracle, not an observed runtime failure. Spec lines3286–3288 require returning the retained disposition with only present positional projection, and the approved plan requires all non-positional failure data to be compared. Use existing full-object literal equality for these independently known fixture Problems (including exact current index where relevant), or an equally complete comparison. Add added-field controls for both carriers/sequence projection. No runner, normative schema or runtime change is needed; Problem extensions remain valid generally.

## Confirmed evidence and limits

- The four-file delta is evidence only: two manifest rows, coverage entries, seven fixed fixture Bead GET bindings, and test/control additions. No runtime/runner/executor/schema/catalog change.
- Independently compared parsed JSON for all43 prior manifest scenarios: all are unchanged. Existing26 whole flags remain; all129 required IDs are retained.45 authored rows =26 whole +19 partial;103 uncovered =19 partial +84 omitted. Both newly authored rows remain explicitly partial and `claimEligible` remains false.
- Successful retry oracles compare complete retained Resource records against earlier real captures, while independent GETs assert deliberately newer known properties/metadata/IDs/Types and captured opaque revision relations. Fresh-key updates positively distinguish retained missing-subject failures from an operation that now could succeed. Changed-semantics conflicts are followed by unchanged afterstate and original-key replay.
- Singleton/sequence key reuse and reordered member positions are real installed HTTP. Names are applied only to creator operations; the failed update remains unnamed. Ordered index tuples pin cardinality and position; separate presence/absence/name assertions stay within the assertion grammar.
- All13 authored controls execute once after actual wire capture and fail their specific named assertion. They test assertion sensitivity, not a defective target or octet-transparent proxy. The review probes use that same explicitly post-wire seam; recorded observations are not rewritten.
- The seven new route bindings are exactly the documented fixed Bead GETs. The closed route digest is repinned; existing captured allocated-Bead GET validation remains same-origin, canonical Scope/kind constrained. No new wildcard, credential behavior or transport route mechanism was introduced.
- Author packed artifacts were inspected as author evidence: clean exact29e715 source,45 passing sessions,781 client observations,44 setup observations,21 native Read observations; no claim that this reviewer independently reran pack/install. Source/payload provenance remains launcher/reviewer-bound, not independent remote attestation.
- No credit for lost-response recovery, crash/restart, expiry/in-flight behavior, changed authorization, other principals/Scopes, all operation kinds or complete semantic-identity variants. No public profile admission or whole-row promotion follows from this slice.

## Independent verification

Pinned Node24.16.0/pnpm11.20.0; isolated clone installed with frozen lockfile.

- PASS clean pre-build `pnpm typecheck`.
- PASS `pnpm build` (includes clean).
- PASS97 tests /2 files (`apps/bdptest/src/read-update-matrix.test.ts`, `packages/conformance/src/cumulative-artifact.test.ts`),61.60s. Includes full45 installed cohort, old regressions and all13 new controls.
- REPRODUCED four oracle-insensitivity probes,4 pass /91 filtered,2.72s. These passing probe tests deliberately require the corrupted assertion inputs to leave the row pass; they are findings, not conformance passes.
- Frozen author `git status --short` empty before/after. No review-owned installed child remains after tests. No source repair, commit, push or public action was performed.

Logs: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/read-update-retry-native/`: `install.log`, `typecheck.log`, `build.log`, `baseline.log`, `baseline.json`, `probe-final.log`, `probe-final.json`, and `probe.patch` (complete four-probe addition against frozen source). The earlier three-probe run is retained as `probe.log`/`probe.json`; final four-probe run supersedes its count. Review clone retains only this review-only test modification.

Full repository regression/static suite and independent packed installation were not rerun. No other actionable finding from this bounded review. Repair the failure equality oracle before treating the stated retained-failure observation as complete; the conservative26/103 inventory needs no count change.
