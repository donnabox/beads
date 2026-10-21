# Per-carrier admission evidence completion

Frozen clean **8e12162e63502660cc562503c9506789919ab604**, direct ordinary successor of **394d3208790a488fd790257da9d48f4b293f129a**. Checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-admission-complete`; branch `codex/vickie-read-update-admission-complete`; origin `https://github.com/gastownhall/bdp.git`. Predecessor frozen unchanged. No outward writes.

## Scope and review dispositions

Exactly four evidence files: `packages/conformance/matrices/read-update-v1.json`, `packages/conformance/test-support/read-update-coverage-v1.json`, `packages/conformance/fixtures/read-update-reference-v1.json`, `apps/bdptest/src/read-update-matrix.test.ts`. No runtime, store, runner, executor, catalog, schema, TX or public-admission changes. Promotion-only cumulative-artifact test untouched. **129 required / 47 authored / 26 whole / 103 uncovered**, both admission flags remain false and `claimEligible:false`. All 45 pre-admission scenario JSON blocks remain byte-identical to 4a; catalog order and exact ordered assertion bindings pass cumulative-artifact tests.

- External F1 / root per-carrier resolution: existing missing-key observations remain on all eight singleton routes. Each of the other seven routes now receives `bad.key` and repeated same-key fields (one case-varied field name), immediate complete unchanged Resource state and alias resolution where applicable, then an identical-body/route successful valid-key correction. Create-Bead retains its eight malformed classes, repeated identical/distinct/case-varied corpus, boundary positives, and their paired successful corrections. The existing eight invalid member-key Sequence classes retain whole-carrier/no-prefix observations. No key namespace, retention or internal drain conclusion follows.
- The media row now observes missing Content-Type, declared `text/plain`, and `text/plain; charset=utf-8; arbitrary=ignored` on **all nine mutation targets**, then actual raw `application/json; charset=iso-8859-1; arbitrary=ignored` success on that same target. The correction reuses the refused key(s), changes the input, and verifies effectful final state. Existing `application/problem+json`, isolated parameter cases and UTF-8 `café雪` remain. Sequence has two independent effects; all refusals leave its prefix absent.
- F2: removed the three unused key alternate-delete seeds/readbacks and fixture bindings. Key has eight setup writes plus eight reads; media retains eleven plus eleven. These are counted manifest actions, not hidden setup. Existing aliases and deletion operands are reused while live; no new allocation seed framework.
- F3: every collection envelope in these rows has the normative `currentBeadCollection` or `currentLinkCollection` schema: **138 key + 82 media** actual collection reads. They additionally assert explicit limit100, exact independently accounted sorted IDs/count, `next:null`, and every complete item equals its latest independently verified earlier Resource capture. Raw success captures still use subsequent client GETs because raw capture supports URLs only. Those GETs pin known identity/type/properties/metadata/endpoints and the correct Bead/Link schema before becoming baselines. No raw-result opaque-revision equality is fabricated.
- F4: **the external 50/200 assertion was rejected after checking actual installed evidence**. Both historical 394d and this exact fresh packed successor return `limits.page={defaultItems:20,maximumItems:100}`. This matches the consumed owner tuple in development-authority.ts. Public Read test fixtures use a separate 50/200 configuration. Parent withdrew its preliminary acceptance of F4. Actual fresh request/response/config is retained in `installed-discovery-limits.json`; explicit limit100 remains valid.
- F5: all budgets unchanged. No runtime route wildcard, additional principal, timeout increase or framework extension. The separate reference-error proposal remains deferred pending root review and measured remaining headroom.

The preceding reviewed plan f23d is preserved as `read-update-required-next-plan.reviewed-f23d.md`. The intermediate plan containing the unverified 50/200 correction is preserved as `read-update-required-next-plan.carrier-delta-unverified-e0ea.md`. Current corrected delta/disposition plan SHA256 **a5bff590fae370611b055130c930d38b1eba0bf4c7bcc67a7e36a3d4579187c1**. Original49b3/acceptedb12f histories remain intact.

## Accounting and discriminators

| Row | Actions | Raw HTTP | Client | Outside-manifest setup | Peak live Beads/Links | Assertions | Max authored body |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Key |284|196|88|0|36/3|2303|661 bytes|
| Media |174|140|34|0|11/4|1247|424 bytes|

Actual packed reports independently yield the same 196/88 and 140/34 exchange counts. `final-action-ledger.json` includes each raw carrier, expected status/media and exact per-route counts. New refusals reuse the immediately preceding exact state as their baseline and always have a full immediate after-snapshot; known aliases also have direct resolution witnesses. Peak bounds remain below the approved64/16 fixture ceiling and complete page capacity100.

Final formatted manifest **3,302,903 bytes /88,362 JSON nodes**, **891,401 bytes** below4MiB;1,681 total actions/9,983 assertions. Relative to394d: +98 actions/+779 assertions/+208,595 bytes. Growth is measured, not a promise for the deferred slice. Static tests recompute action/family/body/identity/schema accounting and same-body key / same-key changed-media corrections. All new raw bodies remain far below65536 bytes.

Four additional actual-input controls exercise independent new coverage: remove the repeated delete-Link key and observe200 instead of expected400; replace malformed update-Bead key with a valid key and observe200; change parameterized non-JSON Sequence media to JSON and observe200 instead of415; change accepted parameterized create-Link media to non-JSON and observe415 instead of200. Each fails the named status oracle. Existing malformed boundary, full afterstate, alias Location, stale collection record and post-wire corruption controls remain. These are test sensitivity controls, not fabricated target evidence.

## Verification

Pinned Node24.16.0/pnpm11.20.0. Frozen install and build precede emitted process tests. Explicit exit files are retained for every new verification command.

- Focused initial:16 pass/100 intentionally skipped, one file,12.31s. Includes two actual new rows and the four new sensitivity controls; later full runs include the skipped tests.
- Complete local app matrix + cumulative artifacts: **122 pass /2 files**,83.16s; new rows together1.454s. Includes exact ordering/bindings and unchanged cumulative production guards.
- Fresh packed install at exact clean8e12162: **151 pass /4 files**,86.14s Vitest (86.534s supervised command); pack protocol, pack server and isolated npm install all exit0. Actual47-row aggregate13.951s; both admission rows together1.667s. Per-row standalone duration not measured. Local times do not predict Linux CI; unchanged per-exchange5s, diagnostic cancellation90s and outer aggregate120s remain.
- All47 authored sessions pass,903 client observations/44 setup observations/21 inherited native Read observations. The wider catalog retains87 expected unexecuted/harness-error entries (82 required omissions plus other catalog entries); these are not passed contracts. `packed-summary.json` retains exact observations/provenance and row counts.
- Typecheck, dependency boundaries, explicit4MiB formatting and explicit4MiB changed-file lint, repository lint, diff check all pass. Repository lint reports75 existing warnings/two infos and its default large-file skip; the separate explicit-size lint covers the changed manifest.
- No full repository suite or extra standalone smoke repeated for this evidence-only successor. Packed matrix/exact-client/development-server/installed-Read-oracles are exercised. No timeouts, forced cleanup or surviving test-owned installed child; source remains clean. No hook bypass.

## Exact fresh bindings

- source: `8e12162e63502660cc562503c9506789919ab604`, `sourceDirty:false`.
- actual installed payload: `819c09e3ade4bbc533c118a956af23e22d44d5b7f87a73aef3d391a98dc11093` (measured, unchanged from394d).
- contracts: `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`.
- policy: `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430`.
- authentication: `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e`.
- fixture: `64814924334103e9375c043de472f8bf15c8b71a267a176e06cd0443b8249751`.
- manifest: `ea794cc31863536b920a8abaa685e372c11b6a82abe1c67af26c83714398c5d4`.
- exact ordered routes: `a8d5687519da4040d2f23989f86f9bc804d30ac6c91a85074397eaf7c92de712`; both pins updated after deriving actual insertion-order binding, no set-order substitution.

No whole-contract promotion, public profile qualification, receiver drain, arbitrary policy/Scope, concurrency or retention claim. Parent owns independent revised-source review and promotion decisions.
