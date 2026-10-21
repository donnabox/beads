# Next delivery: inherited Read conditions on the installed Transactional owner

Planning only, rooted in working candidate `a17af9bc4742be33dfd64071f62ffb2936c82d86`; no main query, product execution, imports, source writes or admission changes. The sole design remains `bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md` (Autonomous profile completion, T2/T3/T4/P and accepted F5/F6/F13). This report supplements that authority. It does not repeat the prior serving audit as an exhaustive proof.

**Recommendation:** implement one bounded cumulative inherited-Read wire scenario over the actual mutable Transactional owner, covering conditional GET/HEAD, ordinary-refusal precedence and complete current-state/validator changes after mutation. This is an executable requirement gap, **not an established new runtime defect**. The component implementation already appears correct. Root explicitly accepted this next planning direction; authoring and execution remain closed until review.

## Actual activation obstruction

1. `packages/server/src/index.ts:288–315` only admits configured `read`, then checks its target's Read evidence. `read-conformance-capability.ts:1–57` contains the existing two-target Read cohort seal, not a Transactional capability. Do not widen that enum/check or reuse its digest for writes.
2. `transaction-http.ts:212,592–600` refuses real Scope/descriptor/directory navigation with the private discovery gap. `transaction-discovery.ts:24–66` already builds actual same-context metadata, but it explicitly advertises `transactional`. Wiring that builder now would advertise the incomplete profile. An out-of-band fixture target is not cold discovery, and a non-attesting header does not make that advertisement harmless.
3. Cumulative TX execution itself is available: `artifact-bundle.ts:126–145,220–225` binds the canonical inherited requirement set; `runner.ts:2792–2820` demands exact TX configuration and private session transport. `runner.ts:255,422` explicitly leaves reports `claimEligible:false`. The current generator/verifier is Read-specific (`scripts/generate-read-cohort.mts:71,291,311–316`; `scripts/read-cohort-evidence.mjs:20,28–29`). A separate reviewed write-profile artifact/verification/admission path is still mandatory under the sole plan, after complete required functionality and same-target observations. A full suite or a collection of private reports cannot stand in for that path.

S1/G2 still independently prevent final activation; this report neither decides them nor creates another question. It also does not propose optional History, Types development or a new provenance mechanism. No unconditional additional runtime defect was established in this focused gate inspection.

## Why this particular wire slice is useful

The current cumulative manifest actually executes mutations and current reads, but it contains no realization for `read.http.conditional-reads` or `read.http.head-conditional-parity`. Its current Resource observations cannot establish that 304/412 are bodyless, that HEAD retains GET decisions, that ordinary 401/404 wins over `If-None-Match: *`, or that a subsequent real mutation changes the observed Resource/ETag while advancing current context. This is a semantic omission, not a diagnostic denominator.

Those laws are settled: spec lines1947–1954 preserve inherited GET/HEAD and native negotiation;1965–1970 bind Resource ETag/cache;1980–2008 require ordinary checks before conditionals, native bodyless304/412, and HEAD parity. The exact inherited catalog rows are `catalog/read-v1.json:811–850`. Production `read-http.ts:10–55` applies conditions only to successful representations and strips HEAD bodies; `transaction-http.ts:664` applies it after routing. Existing tests already establish examples (`transaction-http.test.ts:310,332,715,1675`). Reusing these laws in a same-configuration installed cumulative observation is appropriate; inventing a production fix is not.

## Proposed exact positive schedule

Use the existing Task descriptor/properties schema and operator/anonymous credentials in `apps/bdptest/src/transactional-matrix.test.ts:47–79,147–160`. The Task has no owned-Link declaration. The Scope is `https://reference.invalid/acme/`; all application URLs are explicitly predeclared in exact transport configuration, not discovered. No minimum headers, retained cursors, policy transition, Event feed or Snapshot is added.

POSTs target `operations/batch`, with `Content-Type: application/json`, `Accept: application/json`, and distinct literal Idempotency-Key values `conditional-seed` and `conditional-update`. Their complete bodies are:

```json
{"operations":[{"operation":"createBead","id":"beads/a","type":"https://reference.invalid/acme/types/task","properties":{"title":"before"}}]}
```

Seed: 141B, SHA256 `d63bac5f6ec230f2b2420b3c0b64ff21dd4fc6d9a8bdf616353d48e35a91d7ae`.

```json
{"operations":[{"operation":"updateBead","bead":"beads/a","propertiesChange":[{"op":"replace","path":"/title","value":"after"}]}]}
```

Update: 130B, SHA256 `c86a262f896fd4b357050b2b952e8d9992aa5b98f4a5d60f1f6f793040770e20`.

All reads target canonical `beads/a` except the two missing-target reads at `beads/missing`; all send `Accept: application/json`. Read rows have no request body. The fixed false entity tag is `"conditional-no-match"`; independently require the actual seed ETag to differ before using it, rather than assuming any arbitrary tag cannot match.

| # | Action | Credential / condition | Expected |
|---|---|---|---|
|1|POST seed|operator, seed key|200 completed effect receipt, position1|
|2|GET A|operator|200 complete original Resource, ETag/revision/context|
|3|HEAD A|operator|200 empty wire body, GET representation metadata|
|4|GET A|operator, If-None-Match:*|304 empty, cache/context retained|
|5|HEAD A|operator, If-None-Match:*|304 empty, matching decision|
|6|GET A|operator, If-Match:"conditional-no-match"|412 empty, no Problem body|
|7|HEAD A|operator, same If-Match|412 empty, matching decision|
|8|GET A|anonymous, If-None-Match:*|401 actual challenge/Problem, not304|
|9|HEAD A|anonymous, same condition|401 empty, required denial metadata|
|10|GET missing|operator, If-None-Match:*|404 actual Problem, not304|
|11|HEAD missing|operator, same condition|404 empty, denial metadata parity|
|12|POST update|operator, update key|200 completed effect receipt, position2|
|13|GET A|operator|200 full updated Resource equal to actual receipt postimage; changed revision/ETag|
|14|HEAD A|operator|200 empty, updated GET metadata and serialized byte length|

Seed and update receipts must first pass actual mutationReceipt schemas and independent full shallow expected receipt/Resource oracles, including id/key/type/title/metadata/revision/result index/effect and required positions. The GETs must pass current Resource schema and equal those real postimages, not only have matching hashes. Epoch/view are captured from served responses, stable through this sequence; current positions1 then2 are actual asserted behavior. An ETag is an HTTP opaque field whose relation to the opaque Resource revision must follow the actual existing emitter; never replace it with a fabricated receipt transaction or position.

For HEAD, compare explicitly required context/cache/ETag/media and actual wire-body absence; compare Content-Length to the corresponding serialized GET body where supplied, preserving permitted HTTP omissions. For304 do not require an invented Content-Length:0. For412 require body absence/no BDP Problem. For ordinary401/404 retain actual body/schema (GET), empty HEAD, and correct current/no-observation context according to existing routing. Avoid deriving anonymous context rules from authenticated responses. Exclude Date from parity; do not weaken required fields. The source author must pin the exact existing response schema IDs/error context for these rows before execution.

## Coherent sensitivity control and ownership

Use the same selected installed scenario in a fresh owner/store, changing only request4's `If-None-Match` from `*` to the verified nonmatching fixed tag. The actual existing Resource then legitimately returns200 with its full original body instead of304. Keep all other requests and expectations intact. Require successful seed/read/HEAD prerequisites, actual complete serialization, and independently validate the control's real200 full Resource. The named conditional-status assertion (and its bodylessness assertion) must fail; schema/setup/transport/deadline/cleanup failure is not a successful control. Freeze the exact expected failure inventory from authored assertions before release—do not guess that an arbitrary generic test failure suffices. This is an input sensitivity control, not a production mutant or proof that every conditional bug is detected.

Proposed owned paths only:
- `packages/conformance/test-support/transactional/manifest-v1.json`: append one inherited `read.http.conditional-reads` scenario.
- `packages/conformance/test-support/transactional/coverage-v1.json`: map the precise partial observation and explicit omissions, with HEAD parity cross-reference only where existing binding rules permit; never mark either entire contract qualified.
- `packages/conformance/src/transactional-cumulative.test.ts`: preserve all parent objects/bytes and historical reconstruction; pin new actions and revised bounded measurements.
- `apps/bdptest/src/transactional-matrix.test.ts`: independent exact route inventory, current-wire full-body/metadata assertions and selected control. Reuse existing raw executor/serving binding/launcher; no shared helper or protocol/parser changes planned.

The existing `observe` helper performs one bootstrap launch plus one selected scenario launch; budget the selected positive14HTTP and control14HTTP as28HTTP/four children, sequential, without reopen or extra hidden HTTP. Use existing90s cohort/finite request and120s outer bounds; no limit increases. Keep output outside owned runtime stores, retain safe complete response metadata/fingerprints before cleanup, and prove returned/sampled process and owned-store closure. Fresh package/eight-dependency external-consumer provenance and explicit installed entry remain mandatory for installed attribution. No workspace fallback earns qualification. These are prospective bounds; code-derived lifecycle and exact assertion inventories must be frozen before execution. Current whole-matrix/cohort totals and derivative growth must be measured after authoring; do not silently append traffic to historical whole-run limits or relax numeric fences.

Retain old manifest objects and prior runs exactly. Selected positive/control output remains claimEligible:false and explicitly omits cold discovery, all other inherited-read surfaces, full conditional precedence/date/validator combinations, S1/G2 and complete profile evidence. A later write-profile seal must consume complete, freshly bound executable observations and actual discovery readiness; this plan does not solve that final admission step.

## Source pins

Sole-plan local content SHA256 `80f72d23c855880b16f0a80a54d97120ae9495951e5e0c05f0e1bd3ba80d5162`. Working source below was inspected without product execution. The source snapshot is the actual branch candidate, not a main readiness claim.

- `docs/specs/bdp.md`: `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8`
- `packages/server/src/index.ts`: `34f9d9de8b48852c64919e69ce93f149e9cdae5142c15c75e4c35f08f15efc44`
- `packages/server/src/read-conformance-capability.ts`: `58d4529dc40bbf646f02bda4c7fa3fad9b83d21f3080152cc04867dbe3fa6f60`
- `packages/server/src/transaction-http.ts`: `81c130dd18f604ab1d13c03a2f52a98e11a6988f45d62af98b4b8c7499d14529`
- `packages/server/src/read-http.ts`: `31934105eff6fff8ded8ce2ca9e3c19028991ca7c3d5f51b848004177e8c9268`
- `packages/server/src/transaction-http.test.ts`: `9eb452b8e9d80b53f28cd61b40385886c3a61a992aadb1c9c201d9a188081cce`
- `packages/server/src/transaction-discovery.ts`: `bb8055e2089f640a5d2071e389ac2011d08c038427265b50ab3d03c02047cfef`
- `packages/conformance/src/runner.ts`: `ac20c7fbfe15a5e42a6559e89cd3467562efe9bc74d462478414414ed3efede1`
- `packages/conformance/src/artifact-bundle.ts`: `bbf54915eb792433ae492fd73a5334717e74e0942c485a72daa5105928546dd6`
- `packages/conformance/src/profile-contract-selection.ts`: `5b1caf61688bfa6a5465b5f8dc817b84f02be76bfc5bdd886a19e749b6441543`
- `packages/conformance/catalog/read-v1.json`: `769ff9f24cfdfe74a47670240b04149b520af23b2c71a431883fd15e988adeb0`
- `packages/conformance/test-support/transactional/manifest-v1.json`: `cd90592e76cace65b115dbfedf5f604c9c4011ed2e1a0fad954c1edb12aca6cd`
- `packages/conformance/test-support/transactional/coverage-v1.json`: `0c991feb128a50625397e337464badcb17d22a0c31e1d57a6b66fa1032a92a51`
- `apps/bdptest/src/transactional-matrix.test.ts`: `762f84132bdfae18a730e07ac3a759da2a55289644a7b97888f314a55dffae4e`
- `scripts/generate-read-cohort.mts`: `42111310e1e76df5dee938648bc84144c30fe6a9f15c3b4d30caf670584b6847`
- `scripts/read-cohort-evidence.mjs`: `451d77cf880329238f044e3f0e92d681d6493c509fcb3b050f6e292ce3f2a21e`
