# Native review: Snapshot core efficiency repair

**No new actionable core finding. The bounded repair is accepted; integration dependencies remain separately tracked.**

Reviewed exact `390321d788b6d469d3e23b931be7f83ef32aa1b9` against `a0a068c689f5588701c2b7a350bb677dc4f114ae` in `/Users/dbox/repos/wt/vickie/vickie-snapshot-core-efficiency-repair`. Frozen author source stayed clean and untouched. Read the author report and five-file diff plus surrounding owner/record-codec paths. No new external review was read. Independent tests used `/Users/dbox/repos/wt/vickie/vickie-snapshot-efficiency-native-review`.

## Reviewed guarantees and limits

- **Bounded ledger materialization:** `transaction-snapshot.ts:235–295` checks count/compact sizes in SQL, checks expiry classification and forbids expired entries with retained detail, then retrieves at most9 live IDs and rejects overflow before decoding dispositions. Normal capture decodes at most8 existing live dispositions plus its new manifest. Permanent-cap refusal precedes the live decoding path. This does not bound all SQL scans to constant work; expiry may still process a bounded set of due entries. The author correctly distinguishes the first expiry work from a subsequent exhausted-cap refusal, and makes no O(1) or population-performance claim.
- **Cold expired integrity timing:** detail-free expired commitments are no longer eagerly decoded/hashed on unrelated capture. Direct manifest/page lookup still validates the disposition commitment, context, lifetime and counts before token recognition or expired diagnosis. This is explicit in source and closure documentation and pinned by the new cold-corruption test. Malformed/nonboolean expiry flags and an expired flag hiding retained records/pages fail startup before committed pending recovery. The change is not represented as adversarial tamper resistance or continuous integrity scanning.
- **Graph map/index integrity:** all Resource counts/encoded sizes are checked before loading; each record still goes through the admitted codec, canonical kind/identity and endpoint-index checks. The complete validated map supplies local endpoint existence, outgoing owned inventory, exact sorted IDs, and full inline/first-class Link equality. Unowned Beads skip outgoing traversal and still reject unexpected owned planes. Indexing occurs after record validation; no evaluator semantics or mutation visibility was changed. Existing4096/32MiB encoded bounds remain, while decoded objects, inline data and response construction require additional heap as disclosed.
- **Lazy SQL:** the cache accepts only this module's finite literal statements, prepares on first use after provisioning tables exist, and stays inside the owner lifetime. Retained pages still receive independent readback/commitment checks. The parse/prepare regression proves one authority decode plus one first-page decode per Resource in its owned topology and eliminates the former per-Bead Link scan.
- **Context and compatibility:** `transaction-store.ts:247` persists the actual fixed authorization view in format10; exact installation comparison occurs before BDP pending-claim recovery/maintenance. Format9-shaped/missing-view and changed-view fixtures preserve ordinary file bytes and pending claims on refusal. They are format-shaped fixtures, not a fresh execution of an old binary. The shared `readContext()` now validates canonical position for startup sweep as well as reads and reads epoch directly, avoiding pre-initialization use during provisioning. Necessary SQLite hot-journal recovery remains outside the no-BDP-maintenance promise.
- The successor-token framing repair and both9→10/99→100 boundary regressions are unchanged and pass. Fixed permanent65536 issuance capacity, eight live handles, no migration, no HTTP Snapshot route, and no Q1/public profile qualification remain explicit.

## Independent verification

Pinned Node24.16.0/pnpm11.20.0, frozen install in a clean isolated clone:

- PASS pre-build typecheck.
- PASS clean build.
- PASS172 tests /2 files (`transaction-snapshot.test.ts`40 and `transaction-store.test.ts`132),18.72s. Includes changed startup/corruption/parse/cache tests and inherited page, rollback, expiry, reopen and finite-bound cases.
- PASS four additional native probes,0.560s (40 unrelated cases filtered): deleting a local target, changing an inline owned Link while preserving the first-class Link, and changing a Link endpoint index each cause the intended integrity refusal and leave zero Snapshot/disposition/page rows; a forged ninth future live handle is rejected before any disposition JSON.parse call. These are disposable review tests, not production source changes.
- Frozen author tree remained clean; review clone retains only the four probe tests. No owned test child remains. No source repair, commit, push or outward action.

Logs/probe patch: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/snapshot-core-efficiency-native/` (`install.log`, `typecheck.log`, `build.log`, `focus.log/json`, `probes.log/json`, `probes.patch`).

## Integration accounting

Did not duplicate root's full suite or fresh packed run. Root reported the combined source's sole inherited `transaction-read.test.ts:587` exact-count mismatch: two point checks now observe both position and epoch, so four reads replace two. This matches the inspected shared-context change; root owns that narrow integration-test correction. It was not independently executed here and is not a new core behavior finding.

The author's installed smoke failure is an acknowledged packaging allowlist dependency on separately reviewed5208, incorporated by root in63f2cad. This review neither reran that known failing standalone smoke nor treats it as passing. No full-suite, universal heap/performance, standalone packaging, or profile qualification claim is made for390321d.
