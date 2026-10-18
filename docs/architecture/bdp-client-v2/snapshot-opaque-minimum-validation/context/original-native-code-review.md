# Independent native Snapshot opaque minimum code review

**Not ready for execution: one concrete type blocker and one narrow planned-oracle omission.** Production behavior and bounded test architecture otherwise match accepted plan-v3 on inspection. No compiler, formatter, lint, tests, product imports, builds or saved helpers were run; the findings below are source-based, not invented diagnostic output.

## Required corrections

**N1 — helper forwards explicit undefined to an exact optional option.** `packages/server/src/transaction-service.test.ts:3207–3212` declares `minimumReads: TransactionServiceOptions["minimumReads"]`, which includes undefined, and explicitly constructs `{ ownerLifetime: "service", read: f.read, minimumReads }`. The receiving optional property is `minimumReads?: TransactionMinimumWaitOptions` (`transaction-service.ts:71`) and tsconfig.base.json enables exactOptionalPropertyTypes. This violates that option type even though every current caller supplies an object. Narrow fix: use `NonNullable<TransactionServiceOptions["minimumReads"]>` for the helper parameter, as the existing helper at transaction-service.test.ts:1470 already does. No runtime, callback, assertion, count or deadline change is needed. Root independently confirmed this type reasoning; no actual typecheck has been performed.

**N2 — E does not yet assert the accepted exact Retry-After value.** New installed callback at `transactional-snapshot-installed.test.ts:1133–1138` invokes unchanged snapshotTimeout, which at :780 only checks `Number(retry-after) > 0`, then checks cache/limit marker. Accepted plan-v3 E requires actual Retry-After1 on both timeout responses, matching the fixed reference timeout mapping. Add a callback-local exact `response.headers["retry-after"] === "1"` assertion for GET and HEAD. Keep the historical shared helper byte-identical. This is a missing bounded evidence assertion, not a production protocol defect; no extra requests or fields need be added.

Visible extra blank lines/layout may require routine formatting in a successor. No formatter was run and no diagnostic result is claimed. Keep any such edits confined to the owned delta and preserve the original packet.

## Production and preservation assessment

Read the entire 962-line seven-file patch and relevant consistency, Snapshot storage, service lifecycle, HTTP executor, receipt and installed launcher helpers. HTTP changes exactly the receiving capture function and retains the undefined-minimum/retained fence and terminal one-service-call branch. Service strict capture catches only its helper, gates TypeError, separately lexically recaptures and rethrows original on failure; kind/lifecycle/waiter/owner failures stay outside those catches. Known/no-minimum paths retain their order. Owner revalidates independently, captures only newly admitted opaque kind, keeps original request reference for known/no-minimum paths and uses current eligibility checking only for unrecognized minima. Actual context then instant then minimum remains before maintenance/preflight/capture. Unknown foreign stays unsupported. No S1/G2, format/schema/public-admission or token-codec extension is introduced.

All 679 frozen source hashes matched; the 672 outside the seven owned files remain identical to baseline. Exact current Git patch matches the frozen artifact. Inspection confirms only the declared historical owner unknown assertions, handler minimum-loop extension and installed exchange19 oracle changed in existing callbacks; shared installed lifecycle helpers remain untouched. Other Event/parser and production prefix/suffix regions are unchanged. The new helper is test-local, not a shared runtime seam.

## Callback and lifecycle assessment

A is synchronous. It measures real baseline1/1/2 at original time, advances time, performs exactly fourteen Snapshot calls including pos-01 and four retained pre-entry refusals, then compares counts and disposition before ordinary final reopen. Two offline inspections/three owner opens use one store; report captures immutable inventory values and makes no post-reopen expiry claim.

B freezes tuple/kind across a genuine independent commit, with setup plus three attempts, two deferred waits/probes and final position1 timeout. C cancels after a real first wait, then separately drains three attempts while closing; all gates are released before awaiting task/close cleanup, and its closed probe bypasses the counting delegate. F's seven rows preserve synchronous malformed/retained/lifecycle refusals versus asynchronous pre-abort with one clock call. D performs27 real network requests, actual waits8, exact5/8/8/4/2 distribution and only observable HEAD/retained-body assertions. Existing source18-network and amended15-handler remain separate. The selection of22 callbacks/stores and prospective45 network/15 handler calls is feasible; it is not observed execution evidence.

E constructs one exact executor, guards before dispatch and asserts final3, checks actual serialized body presence/octet/digest, canonical Host/minimum fields, nonempty wire, real seed receipt schema and independently specified resource result, actual context and timeout body/cache/marker semantics. The launcher is genuine; no default-mode override is forced. Callback-local cleanup awaits child.close, verifies exit0/no signal and unexpired deadline before removal. Unknown startup ownership and late/failed close preserve directory evidence; generic unchanged helpers aggregate primary/cleanup failures. Report uses actual PID/exit/directory state and does not equate a directory with a surviving process. The fixed5/45/60-second bounds remain; outer validation still owns process census and installed provenance.

## Validation packet reconciliation (not a source blocker)

Actual new E report has `ok` boolean, not historical `outcome`; responses contain only method/url/status/responseBytes/wireSha256/elapsedSeconds, **no headers**. Thus pending auditor's unconditional outcome/header expectations must be replaced during its already-required schema finalization. Do not invent retained headers or claim independent audit of values absent from the report: they remain pinned passing callback assertions. Actual source reports match author observation-schemas.json: A1/B2/C2/D1/F7 =13 records; E1 and historical1 are separate. The validation author has been informed, and its gates remain closed pending actual successor binding.

## Frozen review identity

Base HEAD `0f215c05678f507dd8a39ee08e288f0901710135`, tree `03371e5c2e1ea756e1501e0e26cccb7aa2f5d04e`, branch codex/vickie-snapshot-opaque-minimum with exactly seven dirty paths.

- `source-binding.json`: `c2dada60d8fb1b6243628f3f09d30b7b012c4214852b6c5d0ba76e00f88865fb`.
- `full.patch`: `7f6ed2cd9b5203928cc91ccf5368c333bad5868f916552665b0b15a59cac1da1`.
- `author-index.json`: `a3d68cb4373ab870fc31f66cf6c62446e3eb2c6c4b4872d51ea5dbc90cd1aa42`.
- `selection.json`: `e332adb50a83dab333490391a86d5badb1b45a715c2c0505de54c77d63e2df83`.
- `workload.json`: `f021016d5acc43d9d80b55f5a3e2bb569d89f43737f8cf85795edc0cdd17bc96`.
- `observation-schemas.json`: `6e95f1093724bd2d843e21d20612405873c75f8e09a49ebcd8a53bd4bbc4921b`.
- `preservation.json`: `b122a56dc24b791c233f5951c7189f88fed69be93e2ee51c10b378fd9182e051`.

Plan-v3 SHA256 `e2df7de56582fb376eb36c4a385df0d3923e2b3dddeee15223edf698c0302811`; pinned spec `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8`. All indexed author artifacts were independently rehashed. Guard verified remote role session `2026-10-08T20:26:35Z-81e5` before this sole report write. No source/input/review packet mutation or execution release.
