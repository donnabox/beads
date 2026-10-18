# Ordinary Update command-minimum cumulative plan — proposed, not released

Basis: actual merged main `6f5e607bf338dd166820bfa3f9f480f7cd54c5f1`, tree `a49cb4b2cfc3a7ad9f66d0d790a7e1e4dbad168a`. Original bb9600 plan and actual012d5c0 review remain historical. Authoring/execution remain closed pending root acceptance; recheck actual main/source pins before authoring. No product changes, tests, services or review retry were performed for this correction.

## Purpose and authority

Bind existing ordinary Update behavior to required cumulative obligation `read-update.http.command-minimum-fields`. This is executable evidence integration toward the sole design's cumulative qualification prerequisite, not a new serving fix or public activation. `docs/specs/bdp.md:3581–3588`, the exact catalog requirement, and the operator ruling require mutation POSTs to ignore the three read-minimum fields while preserving ordinary authentication, keys, carrier validation and expectedRevision. Read controls and Transactional rules remain separate. No pending Snapshot, incident-Link, opaque-token or collision decision is inferred.

The existing six-request installed witness at `apps/bdptest/src/installed-reference.test.ts:552–684` already proves singleton/sequence creation, equal retained replay under changed ignored fields, and exact current Resources. It lacks same-row authentication, missing-key, key-conflict, revision-conflict and malformed-carrier discriminators. Reuse its semantics and existing installed composition; do not rewrite Update through TX or add helper APIs.

## Literal row and exact oracles

`scenario.json` freezes 11 raw HTTP actions and 47 assertions; `ledger.json` freezes every actual URL, ordered header line, body byte count/digest and expected status. `coverage-row.json` replaces only this omitted coverage row. Proposed wholeContract credit is finite authored evidence subject to council and actual execution; it grants no public qualification or automatic TX coverage. The unchanged TX partial row must remain partial.

|Order|Action|Expected|
|---|---|---|
|1|singleton-create, full foreign tuple|200 created|
|2|equal singleton replay, empty minimum/duplicate mixed-case epoch/malformed view|200, exact original result|
|3|same key, changed body, partial malformed minimum|409 idempotency-conflict / never|
|4|missing singleton key, position field absent, duplicate epoch/malformed view|400 malformed-request / never|
|5|unknown credential, full tuple|401 unauthenticated / after-state-change|
|6|stale expectedRevision, malformed tuple|409 revision-mismatch / after-state-change|
|7|invalid JSON carrier, partial malformed minimum|400 malformed-request / never|
|8|singleton point GET|200, original complete Resource unchanged|
|9|sequence create, partial malformed minimum, member key only|200 one created result|
|10|equal sequence replay, malformed tuple|200, exact original sequence result|
|11|sequence point GET|200, original complete Resource|

Use existing fixture IDs w1-a/w1-b, isolated per-row installed state, existing operator/unknown credential handles and the same configuration/bootstrap machinery. No dynamic capture or new target binding is needed. GETs omit bodies. Nine POST bodies total 1,112 bytes, maximum190 bytes (derive exact totals from ledger during review). The fields remain raw ordered lines; do not collapse genuine duplicate epoch lines into a header map.

Inside existing `executeExact`, a module-local ledger records genuine input ordinal/method/url/credentialRef before unchanged delegation, then actual returned exactRequest/write/body and response facts. Local indices are zero-based; ledger ordinals are one-based. Retain actual serialized headerLines privately. Compare authored prefix names first, then the exact case-sensitive ordered prefix against the ledger; reject occurrences of any of the three minimum-field names in the generated/credential suffix. Require complete write. The names-only precheck prevents credential values entering a diff if the authored prefix is short. No raw headers enter safe reports; retain full actual response bodies privately. The executor input has no action ID. Join actual returned `ObservedExchange.request.id/method/url` by fixed order afterward; never invent ID or credentials in that type. Assert exactly11 exchanges, no client/setup/native-read traffic, and complete writes matching literal body hashes/lengths. Validate original response schema before storing baselines. A stale-revision dispatch must first establish the literal differs from the actual original revision. Compare complete created/replayed/current Resource and result objects, not a lossy projection; require metadata{}, nonempty revision and exact fixture properties, sequence index0, and absence of TX receipt/context fields. Refusals must leave singleton state unchanged. Source-backed ordinary response retry vocabulary is frozen in the row. Existing transport/header/body limits remain in force; this row does not independently exhaust every transport-boundary case.

Run this oracle in the selected positive and full aggregate. Safe optional report contains counts, request IDs/methods/URLs, write facts, outcomes and equality booleans; no credentials, raw authorization or full bodies. Retain genuine returned runner evidence before cleanup, including expected negative results; propagate oracle/report/cleanup failures. Reuse existing lifecycle and supervisor patterns.

## One small negative

New local option `replay-outcome` applies only when currentScenario equals this row, localIndex === 1 (one-based ordinal2), and the actual authored-header prefix equals singleton-replay. URL/body/status alone must never select it. First validate genuine wire200/schema/full replay equality, then change only assertion-input `outcome` from created to updated and recompute bodyText/bodyOctets. Preserve original wire facts/body and mark the projection as deliberate. This schema-valid counterfactual must yield exactly one failed assertion: scenario `read-update.http.command-minimum-fields`, exchange `singleton-replay`, assertion `outcome` (display join `singleton-replay.outcome`). Four first-action assertions and three remaining second-action assertions pass: 7 passed / 1 failed; runner stops after2 HTTP. Adjudicate structured results: exactly2 exchanges/8 outcomes,7 passed and only singleton-replay/outcome failed. Thrown text is not an identity oracle. Require application count1 for negative/0 for positive+aggregate and increment the shared counterfactual count. Unexpected wire failures must not be relabeled as that expected failure. This is assertion sensitivity, not a production mutation or proof of every possible defect.

## Six owned paths and preservation

1. `packages/conformance/matrices/read-update-v1.json`: append the new row at its actual catalog position after all existing executable baseline IDs.
2. `packages/conformance/test-support/read-update-coverage-v1.json`: replace only the existing omitted row.
3. `apps/bdptest/src/read-update-matrix.test.ts`: local ledger/oracle/control and safe per-invocation writer, one CPU pin callback, two selected runtime callbacks, affected global counts, current manifest/coverage digest/byte/node pins and four measured source-growth allowances. Specifically app4674 must pin the measured successor manifest;80f74625 remains immediate-parent reconstruction only. No unrelated digest changes.
4. `packages/conformance/src/read-update-artifact-set.test.ts`: corpus split/count/preservation expectations.
5. `packages/conformance/src/cumulative-artifact.test.ts`: exact reviewed whole-ID whitelist in catalog order and current data counts.
6. `packages/conformance/src/profile-contract-plan.test.ts`: replace the obsolete RU-omission expectation with the new RU evidence mapping, retaining exact catalog inheritance/Read exclusion and unchanged partial TX mapping.

Preserve all50 baseline and5 continuation scenario raw objects, all129 other coverage raw objects, fixture/catalog/index/production/API bytes. Removing only the appended row/comma reconstructs the complete parent manifest; restoring the single original omitted span reconstructs complete parent coverage. Existing cold-directory row and all old controls retain their semantics. Baseline51+continuation5; authored56, whole35, omitted74, uncovered95, required130; status buckets bounded-observation55→56 and omitted75→74 are separate from whole34→35. Never convert unimplemented or unsupported rows to success. The historical cold-parent reconstruction must explicitly remove both successor additions when it refers to its earlier ancestor; additionally pin reconstruction of immediate6f5 parent (its manifest bytes equal bb9600).

## Measurements and immutable budgets

`measurements.json` is data-only arithmetic on committed bytes and literal row. Proposed new-span indent2 formatting: manifest3,689,484 disk /1,912,593 compact bytes /98,833 nodes; coverage720,627 bytes. Final formatting must remeasure without reformatting old spans. All36 old derivative recipes act on unchanged rows; exact additive delta is9,413 compact bytes/504 nodes. Their combined envelope becomes1,913,704 bytes/98,868 nodes, leaving1,132 nodes beneath the unchanged100,000 parser cap. New response-only control adds no manifest derivative. Preserve the prior exhaustive inventory and reproduce every changed recipe if implementation alters more than this append.

Proposed exact historical data allowances are72,470 /22,013 /17,212 /7,966, retaining historic bases and explaining only2,320 coverage-byte growth. These are review proposals, not execution/runtime cap changes. No4MiB/1MiB/parser/HTTP/time increase. Ordered route set/order stays284 and the existing configuration digest stays26da5a286d9887bebbb0fcf73a92c5d09aa7b7fcfdcc9ed91f767f4db1578c69; independently derive with the existing pure configuration path during authoring. stdlib route projection equality is recorded separately, not presented as an actual installed identity.

|Cohort|Expected HTTP|Maximum HTTP|Children|
|---|---:|---:|---:|
|New selected positive|11|11|2|
|New selected negative|2|11|2|
|Existing full aggregate|2314|2314|58|
|Total|2327|2336|62|

Aggregate includes2270 authored +44 setup/client requests; no hidden directory or bootstrap wire is excluded. Retained replay is explicitly two row actions; automatic transport retries0. Default helper still starts a bootstrap child plus the selected row's child. The aggregate adds one child and11 requests over actual6f5. Full-CI incremental expected load is24 HTTP/5 children including the two new selected callbacks; unchanged old explicitly selected controls and smallExecutionSet cohorts exclude this row: app4039–4064 is a closed include-list retaining singleton.create-bead plus one continuationId. No exclusion-list edit is needed; recheck this mechanism at freeze. Verify source invocation inventory at freeze. Selected30s, aggregate90s cooperative, request5000ms and external120s limits remain unchanged. Previous timings do not guarantee headroom: timeout stops work and requires diagnosis, never automatic limit growth.

## Finite validation and release path

Actual original Claude review completed; all22 findings are source-dispositioned. Root acceptance and native closure of this narrow successor precede authoring; no automatic full review retry is proposed. After ordinary actual-main reanchor, author only these six paths and freeze final formatted raw preservation/measurements/all affected derivative inventory. Reuse pinned Node24.16/pnpm11.20 and existing setup/build/type/format/lint/boundaries/Read-seal checks. No installed execution until concrete source and commands are reviewed/released.

CPU selection: the current cold-directory seven callbacks, with the obsolete profile-contract title changed to `maps ordinary Update command-minimum evidence while TX mapping remains partial`, plus new `pins command minimum literals, ordinary admission and parent preservation` (8 total). Existing seven are precisely those in cold-directory-draft/commands.json: corpus split; cumulative digest/catalog/bindings; profile mapping; cold directory pins; numeric growth; metadata closure; bounded R closure. Final commands must match literal actual registration names, not guessed regex expansion.

New selected titles: `observes ignored command minimum fields with ordinary Update admission` and `rejects a changed retained outcome in command minimum assertion input`. Third selected runtime callback is unchanged `executes the partial cumulative RU matrix against actual installed factory HTTP and afterstate`. Fresh clean-commit build/archive/source/installed payload verification and external eight-package consumer remain required; use existing installed workflow, no workspace fallback. Verify actual response counts, exact negative7pass/1fail, full aggregate required/uncovered states and genuine child/store cleanup, then complete exact-head fullCI before landing. No new framework or public advertisement is authorized by this plan.

## Precise claim, report and artifact boundaries

wholeContract applies only to this mutation POST command-ignore obligation. Spec3581–3588 ends with scope preservation, not a new ordinary Update read-minimum receiving rule. Ordinary Read dispatch development-runtime.ts376–435 and read-http.ts9–54 remain untouched. Existing read.http.conditional-reads and read.http.head-conditional-parity remain partial with identical coverage objects; they are preservation context, not new credit. No invented RU minimum rows or borrowed TX proof. Full Resource/replay/header equality is a harness-local installed supplement to47 portable assertions; detached manifest replay does not reproduce that private oracle.

The safe writer belongs only to the app test file, with a test-only absolute per-invocation BDP_COMMAND_MINIMUM_REPORT_PATH. No new registry/export/CI setting is needed. Initial exclusive wx/0600 creation followed only by same-invocation rewrites,262144-byte ceiling and token/wrongToken checks supplement a strict allowlist. Preserve genuine wire facts; explicitly label replay-outcome as assertion-input-only, separate from projected outcomes. Omit absent-row reports or mark not-run; no invented action IDs, credential fields or observations. Report/cleanup errors fail. Caller stage environment lives in external execution preparation, not a seventh product file.

Five existing first occurrences prove route-order preservation (zero-based): GETbeads/w1-a index2 and GETbeads/w1-b index3 in document-patch-merge-replace; POSTcreate-bead258/update-bead259/sequence266 in idempotency-key-required. New row therefore does not relocate routes. Its RU catalog index78 follows every existing executable RU row (maximum75); inherited Read comes first. Full positional route comparison remains284/SHA664c7550, independently derived from actual committed parent plus new literal.

The fixture seeds only its fixed Read scenario list, which excludes this row. App923–954 creates a fresh per-scenario directory/child and conditionally seeds only that list. New row has no client/native-read/setup requests. App788–797 declares operator and unknown handles, with unknown resolving wrongToken rather than a harness denial. resource-evaluator.ts139–160 maps revision-mismatch to409/after-state-change. Default recordChangeContext:false plus no explicit actor/context produces metadata{} without provenance; keep strict full Resource equality. These are source proofs, not new runtime observations.

This packet and historical draft commands are external bdp-integration-logs evidence, not committed product files. Nine POST bodies sum1112 bytes (max190); GET bodies absent. Frozen ledger expectedStatus is an expectation, never actual response evidence. Final freeze must recheck all dependent count/hash/data assertions including current baseline digest, not only eight selected CPU callbacks. Original report-clarifications remains historical; this successor incorporates its safety rules and supersedes only old size/base arithmetic. The current sole-design hash is refreshed for root PR138/council chronology; historical external pins remain recorded.
