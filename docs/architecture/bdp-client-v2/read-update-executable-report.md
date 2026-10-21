# First cumulative Read+Update executable slice

## Frozen sources

- Checkout: `/Users/dbox/repos/wt/vickie/vickie-read-update-executable`; branch `codex/vickie-read-update-executable`; origin `https://github.com/gastownhall/bdp.git`.
- Authored executable slice: `572d725af9dfae9e56816a5e5019c258805c7229`.
- Separate carrier/readiness correction: `11cbe3cc2cc4b167cd10ebdd08af667e37a03301`.
- Ordinary development integration of P1 `dea6f9e`, `39a7674`, and latest symlink repair `4cf6acd03774fd90e629ba63516ff1365865b00e`; current clean combined source `d3b9524e8e650e6f8a25ea07c0ea87d001e0828e`.
- Final source: `31007d84a91964786cdc1424e4a55d2af83204c6`, clean. Ordinary parent integration `d9b981760184d836e4f1fe43debeb16b2e298fe3` brings the two pre-build type-import fixes and existing reviewed authority-read/pagination/transaction-read/member-diagnostic ancestry (17 changed files relative to this lane). Those inherited semantics are not authored in the executable slice. The final source then clarifies structural readiness in comments/design prose.
- Clean → typecheck → build passed in CI order at the final source. Boundaries, existing evidence verification and git diff check also passed.

## Delivered behavior

The private cumulative factory accepts only the canonical branded Read+Update selection and commits its exact per-file catalog digest map. It preserves all existing public Read and diagnostic development guards, public exports, evidence schema and sealed artifacts. The canonical loader moved once into source-private implementation with test-support reexports; no second requirement set was introduced.

Version-2 manifest and fixture execute seven bounded rows using cold discovery, real credentialed exact HTTP, captured allocated URLs/revisions, and independent GET afterstate. `equalsBinding` is evaluated by shared runner assertions against earlier captures; missing/late/self captures, incompatible values and conflicting expected-value forms fail. The exact client limits captured Bead routes to the same canonical Scope, origin, URL kind, frozen listener and credential handles. Request/response/write observations retain inspected actual transport data and capture provenance; no credentials are copied to reports.

Installed IPC supplies actual Scope/profile/History, payload, contracts, policy and authentication identities. The harness compares each prepared target to a captured baseline before HTTP. The full tagged schema5 contract identity is preserved and equality checked; the existing digest field gets only a strictly parsed component, never an independent recomputation. Source commit/dirty state are launcher Git metadata, explicitly not target attestations.

## Coverage and limits

- 129 normative required IDs inventoried once in canonical order.
- Seven observed rows: discovery, create-bead, update-bead, delete-bead, expected-revision, idempotent-retry, idempotency-conflict.
- Five whole-contract observations; discovery omitted-order fallback and retry in-flight/expired/sequence legs explicitly incomplete.
- 124 required obligations remain uncovered, including all inherited Read and companion rows against this target.
- Unfiltered matrix: 134 applicable rows, seven pass, 127 `harness-error`/`not-implemented` (includes applicable diagnostics). The test passes because it asserts these honest gaps; the matrix does not pass qualification.
- Seven isolated scenario sessions under the same configured server factory; 44 actual client exchanges plus native HTTP observations.
- `claimEligible:false`. No public startup/admission enabled, no Transactional execution, no fresh Read or RU seal, no outward writes.

The machine inventory is `packages/conformance/test-support/read-update-coverage-v1.json`; execution sources are `matrices/read-update-v1.json`, `fixtures/read-update-reference-v1.json`, `apps/bdptest/src/read-update-matrix.test.ts`; design report is `docs/design/read-update-executable-progress.md`. The approved plan and all twelve external review dispositions remain in `read-update-executable-plan.md` beside this report.

## Completed validation

All commands used Node 24.16.0 and pnpm 11.20.0.

- Authored-slice regression: 872 tests / 34 files passed, with structured `read-update-executable-conformance.json` and text `.log`. After the final no-History carrier assertion, focused 60 tests / two files passed (`read-update-executable-carrier-final.log`). Do not add these overlapping counts or describe them as a full workspace suite.
- Build, post-build typecheck, dependency boundaries, focused lint, formatting, git diff check, and existing Read evidence verification passed. Existing sealed 89-row Read artifact untouched.
- Following ordinary 4cf integration: build, post-build typecheck, 29 tests / five files passed (`read-update-executable-symlink-{build,typecheck,tests}.log`).
- Actual fresh tarball installation at clean `d3b9524`: 13 tests / three files passed in 4.14 seconds. Logs and exact observation record: `read-update-installed-symlink/{pack.log,install.log,matrix.log,progress.json}`. The earlier clean 572 run remains preserved in `read-update-installed-final/`.
- Negative controls cover malformed/foreign captured targets, missing revision captures, strict identity tag/version, mismatch of observed policy/configuration before execution, anonymous/unknown credentials, actual stale revision, wrong expected postimage, missing/late/self bindings, public-factory writer rejection, retired/misprofiled rows, and canonical inventory/order/digest binding.

## Next bounded work

Complete real inherited Read and companion bindings against this same installed Read+Update configuration, adapting discovery prerequisites to the RU replacement without importing the closed Read shape. Then fill remaining RU document/Link/ownership/alias/sequence/authorization/retention/error contracts. Use the existing machine inventory and executor rather than expanding planning abstractions. Final profile qualification requires the complete actual cumulative run and reviewed installed target, followed by the separate approved admission/evidence work; this slice supplies neither that gate nor its claim.

## Final carrier-review disposition

The external carrier review closed N1/N2. Its two new low findings are addressed by explicitly documenting that ready-unexecuted certifies only structural readiness, conditional on the execution gate enforcing matching exact configuration; the actual installed harness does so before requests. The prose now describes the new private RU path and restores the version-1 non-Read rejection explanation. No additional mirror test or planning framework was introduced. Canonical merged order remains pinned by the existing new artifact test.

## Final frozen-source validation

At clean `31007d84a91964786cdc1424e4a55d2af83204c6`, clean → typecheck → build completed in CI order, followed by 1,474 passing tests / 41 files in 39.92 seconds. This suite includes all conformance tests and the relevant installed, Read, development, authority/pagination/transaction-read/member regressions introduced through the authorized ordinary ancestry. It is a targeted regression suite, not the full workspace suite. Structured output: `read-update-executable-final-tests.json`; human-readable output: `read-update-executable-final-tests.log`. Final boundaries, existing Read evidence verification and diff check passed.

Fresh final-source server/protocol tarballs were installed into `read-update-installed-31007d8/install` with install scripts disabled. The actual installed factory then passed 13 tests / three files in 4.55 seconds. `read-update-installed-31007d8/progress.json` verifies the exact source above with `sourceDirty:false`, seven scenario sessions, 44 actual client exchanges, seven planned IDs, 129 required IDs and 124 uncovered IDs. `claimEligible:false` remains unchanged. Pack/install/matrix logs are preserved beside this progress record. This is the authoritative latest observation; the earlier source-pinned observations are retained as history.

No source edits remain pending. Root may begin independent review of the frozen source while this report and logs remain external to the checkout. No unresolved product decision was introduced by this bounded slice. Final profile qualification and completion of all cumulative obligations remain future work.

Exact final observed identities:

```json
{
  "baselineReceipt": {
    "event": "qualification.ready",
    "scope": "https://reference.invalid/acme/",
    "claimEligible": false,
    "configurationIdentity": "0d292d7e7fd502ffd141986d47793c194d861de72ca5379e4eca630c3b33a198",
    "artifactIdentity": "development-schema5-e1dff401b5caafb04a4d935775d67eeefeae7bcf306d5d100b979a64eef148e7",
    "payloadIdentity": "6503d708974149ece46a9e3ef5d676cdd13ef74393b2bff1f417937392e5889c",
    "policyIdentity": "dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430",
    "authenticationIdentity": "66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e",
    "profile": "read-update",
    "history": false
  },
  "reportBinding": {
    "artifacts": {
      "catalogDigest": "03ca39f24a988a16497d80a52c1fffad7e2c5dbd27a2a521679cc0183672b184",
      "manifestDigest": "f07bafeffc3e1799b9b9a17cd112b3fd35b3cd85ebfe92a23a3d72d1a52f1ec8",
      "fixtureDigest": "529c89518946a9a6f17159f17a44026018d27744a753f51caeeabbd599f6974f"
    },
    "exactConfigurationDigest": "67049d455adac264014f46fd1423527ebe6fcd08cb987808893dceb120d83748"
  }
}
```

## Native review correction — pending successor repair

The independent native review found that the update-bead no-op response itself lacks a comparison to the prior postimage/revision; the later GET alone does not detect a response-only wrong revision. It also found that this row's metadata-patch object/effective-Type schema clauses are not exercised by the properties-only case. Therefore the frozen source's five-whole-contract/124-uncovered classification is overstated and must not be used as a current completeness claim. A successor will add direct response equality plus a response-only counterfactual and mark the still-unobserved update contract incomplete (or supply actual additional coverage). The frozen source and historical progress JSON remain unchanged for independent review. Test-run counts and observed HTTP exchanges above remain factual; they do not establish the missing contracts.
