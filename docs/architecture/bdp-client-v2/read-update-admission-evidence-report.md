# Installed Read+Update admission evidence

Frozen clean **394d3208790a488fd790257da9d48f4b293f129a**, ordinary successor of **4a8607ac2131aae4706d00684a2f3f1185cb37e9**. Checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-admission-evidence`, branch `codex/vickie-read-update-admission-evidence`, origin `https://github.com/gastownhall/bdp.git`. Parent/frozen sources untouched. No outward writes.

## Scope and evidence

Four files changed: RU manifest, RU coverage bindings, reference fixture bindings, app matrix tests. No runtime, executor, runner, schema, catalog, TX or public admission changes. The fifth promotion-only cumulative-artifact test remains untouched. **129 required /47 authored /26 whole /103 uncovered**, `claimEligible:false`. Both new flags are false pending independent selected-clause audit. All45 old scenario JSON blocks were independently compared **byte-for-byte unchanged**; exact ordered bindings and catalog order pass the existing cumulative test.

- `read-update.singleton.idempotency-key-required`: missing key through all eight routes; eight malformed spellings (empty,257 characters, quoted, internal space/tab, punctuation, comma-combined, Latin-1); repeated identical/distinct/mixed-case fields;1/256-character and alphabet-positive keys. Every malformed singleton record gets an identical-body/route, only-key-corrected successful submission and full GET postimage. Each invalid-token class also appears after a valid effectful prefix in a sequence; direct rejection leaves both IDs absent, then only the invalid member key changes and both succeed. Duplicate valid keys each get changed-record recovery to prove neither occurrence captured a disposition.
- `read-update.singleton.unsupported-media-type`: missing media through all nine routes, distinct wrong media types, corrected same-key changed-record submissions, and real raw parameter-bearing JSON with UTF-8 `café雪`. The two-member sequence has independent effects; refusal leaves no prefix. `charset=iso-8859-1` and irrelevant parameters leave UTF-8 semantics intact.
- Every refusal is actual raw HTTP with authored header/body/write-state assertions, exact expected Problem plus normative schema. Later writes/GETs use the existing exact client. No synthetic positive response or special owner control.
- Complete Resource collections request explicit limit100, assert independently accounted sorted IDs/count/next:null, and each item deep-equals the latest prior verified Resource capture. Captures independently pin known id/type/properties/metadata/endpoints and absent ownedLinks for this unowned fixture. Updates replace expected bindings, deletions remove items, creations add known items. Resource afterstate is never derived from the negative request. Known aliases use real307/exact Location/empty body or404; there is no alias collection.

## Approved mechanical change and initial authoring failures

Reviewed49b3 and acceptedb12f plan bytes remain in separate historical files. Current mechanical clarification hash **f23d9443563d51cef024263694fa673cc2cbe3076552eb604c3ef041c0395a89**. Parent expressly approved the substitution below; no target limits or scope changed.

The initial raw collection-body capture failed because HTTP `captureValue` accepts only absolute URLs. That failure is retained in `focused-one.log/json`; no passing result was inferred. Arbitrary JSON captures are supported only on programmatic actions. The final collection oracle uses prior successful command/GET Resource captures per item. Raw parameter successes likewise get independently pinned client GET postimages for future collection comparison, not a fabricated opaque raw-result revision equality. No receiver-consumption/drain inference follows from a complete client write.

Other preserved authoring failures: invalid slash-bearing fixture binding names (`initial.log`), alias paths colliding with live Bead local IDs (`initial-two.log`), and TypeScript narrowing/readonly-list errors (`typecheck-one.log`, `typecheck-final.log`). Fixed in the four owned files. They were fixture/test-authoring errors, not demonstrated runtime defects. All later successful commands have separate exit files; no historical exit was retrofabricated.

## Bounded accounting

| Row | Actions | Raw HTTP | Client | Outside-manifest setup | Peak live Beads/Links | Assertions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Key |244|151|93|0|37/4|2036|
| Media |116|73|43|0|11/4|735|

Each row includes11 real setup mutations and11 readbacks among those actions, not hidden fixture writes. Permanent tests recompute action/route and live-identity accounting, verify every malformed singleton/sequence pair changes only its key, and pin count/query/complete-page evidence. Both peaks fit approved64/16 slice ceilings and advertised page maximum100. `final-action-ledger.json` includes per-target counts and maximum authored body bytes. The current manifest is **3,094,308 bytes /82,657 JSON nodes** (`source-ledger.json`), below the unchanged4MiB private cap. No budget widened or case silently removed.

New sensitivity tests: actual duplicate→single and257→256 input corrections cause200 and fail refusal oracles; wrong-media parameter probe causes415 and fails success oracle. Existing post-wire controls replace corrected status/state and must fail named immediate assertions with exactly one application/input alignment. Independent wrong alias Location and stale prior collection-record expectations fail their named oracles. These are real-input/oracle sensitivity controls, not target evidence fabricated from altered responses.

## Verification

Pinned Node24.16.0/pnpm11.20.0, frozen install, build before process tests; external logs/explicit exits in this directory.

- Final local **117 tests pass /2 files**: complete app matrix111 plus cumulative artifacts6. Aggregate47-scenario job15.595s; the two admission scenarios together1.422s. Full local Vitest86.55s. Earlier focused11 tests passed before final accounting/oracle additions; do not sum overlapping runs.
- Fresh packed **146 tests pass /4 files**, exact clean394d320 (`packed/matrix.json`, exit0). Packing/install also exit0. Aggregate47-scenario job14.679s; the two admission rows together1.296s; whole packed Vitest78.66s. Per-row standalone durations were not separately measured. Prior Linux CI>30s remains relevant; no claim these local times predict CI.
- All47 authored sessions pass.917 client observations,44 setup observations and21 inherited native Read observations. Required omitted rows remain visible; the wider catalog result includes87 expected unexecuted/harness-error entries (82 required omissions plus other catalog entries), not additional passed contracts. New-row raw/client counts above come from actual packed row records. `packed-summary.json` retains exact provenance.
- Final typecheck, boundaries, formatting (including explicit large-manifest check), lint and diff-check exit0. Lint75 warnings/two infos; no warning-free claim.
- No full repository suite or extra smoke repeated for this evidence-only slice; parent4a full8451 is historical. Fresh packed matrix/exact-client/development-server/installed-read-oracle suites were rerun here.
- No timeout, forced cleanup or surviving child owned by this run. Source clean/frozen. Existing unrelated process trees were untouched; disposable authorities close through the existing finally paths.

## Exact fresh bindings

- Runtime payload: `819c09e3ade4bbc533c118a956af23e22d44d5b7f87a73aef3d391a98dc11093` (measured, equal to4a).
- Installed contracts: `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`.
- Policy: `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430`.
- Authentication: `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e`.
- Fixture: `ebcf9771f3c45afc8d6a4de097b39ce25b03821d68d4d9015ae798edeb8c6af4`.
- Manifest: `9ec35ef4179afcdfe1ec77f9264ba7c8ef7285d5d23aa5b1b4161b47fd576be5`.
- Exact ordered routes: `fdcbe130767421273c839d15cd8db5fef4b8047e7aabe192a54c1d4110aca6be`; both pins updated. Routes remain exact and bounded; key/media rows occupy canonical indices30/33. No wildcard or credential expansion.

No whole-contract promotion, full-profile qualification, receiver drain, arbitrary principal/Scope, retention, internal admission stage or optional History claim. Root owns independent review, integration and publication.
