# Native code and evidence review — exact local schema serving

## Findings

**No remaining actionable findings.** No Medium-or-higher finding or production defect was identified. All required local implementation and recorded execution gates in this bounded review are supported. The earlier preflight cleanup Low is closed in the exact frozen source. Root owns external council, exact-head CI, ordinary integration and public admission. This review is not a complete Transactional-profile qualification.

**Closed P3 / Low — source provenance wording.** The original `transaction-cold-schema-serving/report.md:62` called declaredSourceCommit “launcher-verified metadata.” Actual `packed.py` checks the package checkout HEAD/clean state, builds/packs it and supplies BDP_TRANSACTIONAL_SOURCE_COMMIT to the candidate harness. `transactional-schema-installed.test.ts:31–39` reads that supplied declaration; `transactional-execution-binding.ts:74–80` inserts it into the binding. `scripts/transaction-installed-observation.mjs` does not receive or check a source commit. The original evidence was soundly bound by the supervising script, actual tar/installed bytes and payload digest; the report assigned the check to the wrong component.

The corrected report now explicitly says “supervisor-verified source attribution bound to a separately built package” and “not child-attested provenance,” identifies packed.py and installedTransactionBinding, and states that the IPC launcher does not verify this caller label. Independently compared the complete original/current report: only this attribution paragraph changed. Original report SHA af5416434406f997193f6a30180d264e12eeaa51a67983dc485e81be003da1be is preserved as `report-before-provenance-wording.md`; corrected report SHA is 5762647d142cb16a1a3e9e88007758015747304adc57e6b957adea2a9486d31c. Read and reconciled `provenance-wording-disposition.json`. Source, summary and cleanup hashes remain unchanged; no rerun or implementation change was needed. This closes the Low.

## Exact basis and scope

Frozen source: c555fe01df5cd20454fb0d663431df816c1270aa, tree062a3ab7e24236417bd915b373a4b44ba9908647, sole ordinary parent4aa2d0be950f157f3f0617278dd37c40f41ec035. The isolated checkout `/Users/dbox/repos/wt/vickie/vickie-local-schema-serving` was clean at entry and final source check. Full base-to-head scope is nine files,1089 additions/5 deletions. All nine frozen bytes and hashes equal the accepted cleanup-corrected preflight; no execution-time source corrections occurred.

Reviewed actual production/test/docs/package diff, accepted plan c464ddba, dispositions72d3a92b, independent closurea356a0cd, root acceptance, packaging amendment, measured preflight, cleanup correction/closure, executed controls, actual source test records, both packed cohorts, final report/source/summary/cleanup and root-packed-audit. The previous static preflight review is supporting history, not a substitute for this code/evidence review.

No tests, builds, installs, services, network, product changes or outward actions were performed by this reviewer. Independent work comprised source reading, hash/byte comparisons, in-memory patch reconstruction, parsing saved results and recomputing the installed payload hashes from files without executing package code.

## Production and compatibility assessment

- `transaction-schema-http.ts:14–36` derives a private lookup from genuine immutable captured contracts. It accepts only exact canonical local schemas/ retrieval identities, excludes query/fragment delimiters and userinfo, copies captured base64 into immutable UTF8 text and retains the captured byte length. It does not fetch schemas, invent canonical/embedded-ID aliases, reject accepted configurations or intercept arbitrary Scope paths. Existing ingestion already validates these captured bytes; no lossy JSON reserialization is introduced.
- Adapter38–50 handles only authenticated GET/HEAD, with none of the minimum fields present through the exact existing hasMinimumControls helper, and an exact manifest hit. Unknown, alternate, unauthenticated and minimum-bearing requests return unhandled without owner observation. Runtime167–171 places this dispatch after existing raw-target/Host/method/body/closing/Origin checks and before ordinary handling; it never replaces a handler404/refusal after the fact. The sole HTTP-module change is the helper's export keyword; no package API exposes it.
- `transaction-service.ts:525–533` adds only schemaObserved under existing run/withRead. The callback returns the actual frozen owner context in one synchronous entry; no key admission, sequence, receipt/Type/discovery read or minimum wait is used. Both the observation wrapper and context are immutable; the reader is released before asynchronous delivery. Reentrancy/closed owner preserves causal failure, service closure prevents new entry and drains active work. Runtime closure drains pending handlers before closing the service, so the adapter's asynchronous fallthrough does not introduce a closed-owner ordering defect.
- Adapter52–64 runs shared HTTP selection once using representation metadata; it needs no additional parsed schema copy. Only selected200 receives raw text and exact byte length. Existing private deliver at runtime87–95 retains the destroyed guard, development marker and defensive HEAD strip. HEAD reports captured GET byte length without transmitting bytes. Selected304/406/412 remain bodyless with selector metadata, including no invented304 length and no fabricated ETag/Last-Modified. Observation faults use the existing sanitized runtime500 path without stale context.
- Source compatibility partitions demonstrate continued startup/offline validation under differing$id, outside-namespace and query-bearing artifacts; outside-namespace capture does not shadow the actual Bead response. The deliberate canonical/reference-base reachability limits remain documented. Local owner-observed context is the accepted conservative serving contract, not a new universal schema classification or minimum-support claim.
- Packaging adds exactly four private output filenames to scripts/smoke-executables.mjs. Package exports remain only the preexisting root/development surfaces; the unchanged generic transaction-*.js denial loop covers this module. No local smoke run is claimed. Public launcher/admission, schema/catalog authority, configuration, store, compiler, shared Read and existing service production are unchanged.

## Source tests, controls and regression limits

Actual adapter result:8 passed,0 failed. Actual service selection:2 passed,63 explicitly filtered. These are10 selected new tests, not65 service passes or a broad regression suite. The deterministic path count is38 source HTTP attempts across seven loopback listeners plus two direct service owners; it is not a retained per-exchange transcript. Tests cover exact root/transitive bytes and input immutability; HEAD/304/406/412 context/media/cache; current position after a real commit; all11 minimum-header rows and auth/method/Origin precedence; compatibility; real disconnect suppression and continued HEAD usability; sanitized observation fault; same-entry context/lifetime/no-admission behavior. The new cleanup hook now attempts all registered cleanup operations and restores mocks even after a failure, aggregating cleanup causes.

Typecheck/build/format/lint/boundaries/historical Read evidence verification all have actual exit0 records. Lint has663 warnings/43 infos; the report correctly makes no warning-free claim. Typecheck passed without the prospective strict-header typing correction mentioned during preflight. No unexpected setup/test failure or replacement execution is hidden. Existing Read evidence verification retains its historical run identity and does not qualify this changed target. Broad workspace tests and local smoke were explicitly not run; exact-head CI remains a separate gate.

Independently verified all five control patches and executed mutant-source hashes against the accepted ledger. Each derivative built, ran exactly its intended selected case and failed at its named predicate; separate adjudication passed:

| Control | Actual failure | Deterministic HTTP path |
|---|---|---:|
|remove-handling|known root status,404 versus200|1|
|reserialize|exact captured schema bytes differ|0|
|omit-transitive|referenced document status undefined|0|
|stale-context|current schema position pos-0 versuspos-1|8|
|minimum-bypass|preserved minimum refusal200 versus500|1|

These failures reach the intended source assertion, not compilation/setup/timeout. Total10 derivative HTTP attempts is within hard32 across three listeners; no installed child is hidden in these controls. Unit sensitivities intentionally add no wire requests. No completed alias/minimum/Snapshot/SSE implementation was reopened.

## Installed evidence and independent binding

Final report/source/summary/cleanup were available and read before this conclusion. Verified identities:

- report.md:5762647d142cb16a1a3e9e88007758015747304adc57e6b957adea2a9486d31c (report-only Low closed as described above).
- source.json:8ebd9984ab2b7a615c1ef2333b1059050af0ae0e4a45e7a4dc01cace1720742f.
- summary.json:aa783dacb0334e1beac7c8d166cf99d9eb4014bdfbf712079076eab3ada15f1f.
- cleanup.json:c7161b4162c68bbf96a50fc05c5371ad8cc6393d2759939ca9603fc1a61059c7.

The final summary's copied command/cohort records exactly match their underlying exit/result files. Candidate positive runs the frozen c555 harness against independently packed c555 server/protocol; predecessor negative uses the same harness against separately built/packed genuine4aa. No source/run is relabeled.

**Positive14 requests/one child passes.** Independently checked all14 method/URL/status/body-length/body-hash entries against the preflight and the ordered nine-route digest3a12834387c186457a8a857c7dc693bee45a7fdf44295ed9a65f3720531535e6. Type inventory and descriptor are validated before exact observed grants; actual root$id and actual$ref resolve to the independently configured captured reference. Root174B and referenced118B response hashes equal captured bytes, including noncanonical whitespace/non-ASCII/numeric spelling. One98B mutation body validates against the real closure and produces the independently checked Resource/receipt atpos-1. The final root remains byte-identical and moves frompos-0 topos-1. HEAD is bodyless with174-byte length; selected406/304/412 preserve current context and expected metadata. Unknown/auth/method/discovery exclusions remain their explicit outcomes. No hidden receipt, pagination, external fetch or reopen occurs.

**Predecessor3 requests/one child fails for the intended reason.** The first two Type reads pass their real schema/content assertions. The third recorded response is404, and installed.json/log points to exact():239 with `known installed schema status: expected 404 to be 200`. The separate postprocessor validates one failed test,200/200/404, actual parent attribution and cleanup, and passes. This is an expected negative, not a passing installed scenario or generic any-failure acceptance. Aggregate17 requests/two children remains below hard28/two; one attempt per target, no replacement.

Independently recomputed all four tar hashes, compared every archive member to its corresponding installed file, checked the installed entry hashes, and recomputed the entire runtime dependency-closure payload from disk:

| Target | Payload digest | Server tar SHA256 |
|---|---|---|
|c555|0431b66bb506d46c3414587cf34f6d8007472a3aeaf2ec1488b3e59ae56b1200|f940337b0b94eaf5930b6174e11c056947307245f262ec9ba60c13f941d06370|
|4aa|f1bc055701323b1a8adfb474b8b651a9522d06e91deaba53efeb5afab5520859|97a227add3a3ef9588daf6a4037d692646947b8ba7f175e4e294a00100675bc5|

Protocol tar is identical4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227. Entry hash equality is expected because the bootstrap did not change; it is not payload equality. Actual runtime configuration/contracts/policy/authentication/fixture bindings are retained per cohort. Different fresh authority epochs/authentication configuration digests are not rewritten as equal.

Both child sessions record exit0/no signal and store-directory removal. Author postprocessors and final cleanup record independently checked PIDs60913/61977 absent; root-packed-audit independently confirms both absences and pins the exact reviewed packed-summary hashes. This closes the preflight requirement that PID absence be checked outside the test's exit fields. Recorded supervising processes are also absent. No process or cleanup failure is disguised as success.

## Limits and decision

The accepted14-request test remains a recurring14/one-child cost. Source/control and installed counts are separate; outer command supervision does not widen the inner5s exchange/90s cohort/120s test/10s close contract. Memory figures are the static retained-copy inventory (456 capture bytes,612 base64 characters,292 UTF8-equivalent/582 UTF16 content bytes plus174B decode maximum in this fixture), not measured RSS or a universal transport/heap bound. Full body/wire transcripts are not retained; source/preflight expected bytes plus actual fingerprints/context are retained.

This code and its bounded evidence support the accepted exact local schema retrieval slice. They do not close canonical/embedded-ID/outside-namespace/query traversal, minimum semantics, cold discovery, queued comparison-timeout context, pending metadata-only Event membership, mutable-view/erasure applicability or cumulative public admission. No new semantics or exemption is inferred. Root can combine this clean native review with external council/CI; no additional execution is requested by this review. Parent reports PR118 merged and PR119 published at c555 against a base tree identical to 4aa, with PR119 CI and external review pending. This is integration context supplied by root, not a network or CI check by this reviewer.
