# Native review: bounded Transactional set conformance

## Findings

No actionable findings in the reviewed seven-file change. This is a clean review of the bounded private evidence slice, not Transactional qualification, promotion of either row, or approval of omitted Selector/policy/concurrency behavior.

## Exact target and scope

- Candidate: `9eb9339a0b25a5ce9c16e90c18010874bab8d571`; tree `922177e400b929caf44fc52708df39bdf2c35b83`.
- Direct parent: `2019649745e7d11be70de4089361619e7b8ecd7f`.
- Immutable author checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-set-conformance`.
- Independent detached checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-set-conformance-native`.
- Read the entire seven-file diff, both new manifest scenarios and assertion/capture chains, additive fixture bindings, coverage rows, accepted revised plan `57aeeded7f755dbf0487e2a1b2f45541d0a8f13b66efef5c6ee1cfee6e103822`, root acceptance, and the shared review prompt. Checked relevant existing receipt-observation implementation and selected normative clauses at spec lines 4240–4275, 4579–4582, 5143–5178 and 5295–5323.
- No author source edits, public actions, credential retrieval, shared databases, or other owners' assets were used. No runtime/schema/spec/public-export changes occur in this delta.

## Review conclusions

1. `transactional-observation.ts:146–162` adds only the two exact canonical POST suffixes. It inherits actual bounded wire/header/body checks, normative receipt schema validation, canonical receipt identity, original credential/session linkage, and inline completed/failed receipt constraints. Sequence, ordinary Resource captures, foreign paths, query/trailing-slash variants and wrong methods do not gain receipt authority. Ordinary recapture still revokes it. Dynamic GET compares the complete retained body with original wire facts; reopen preserves logical identity while requiring a later physical generation.
2. `manifest-v1.json:1088–2087` checks min-cardinality failure, failed-receipt identity/context through actual GET/HEAD and reopen, unchanged independently established full preimages, staged batch cardinality failure at operation index 1, tentative creation absence, and original Resources unchanged. The fresh-key allowed-cardinality comparator is concrete and effectful. Its receipt assertions are a bounded comparator; this report does not elevate them to exhaustive set state coverage.
3. `manifest-v1.json:2090–3495` seeds Beads in reverse ID order and an unowned Link, independently asserts literal IDs/types/endpoints/properties/metadata, then checks matched count, exact sorted result identities, both operation indexes, updated outcomes and literal postimages. Point reads equal the validated receipt postimages and also assert literal properties/metadata. Link deletion checks exact deleted identity, point absence, unaffected Beads and durable graph/receipt facts after reopen. This avoids accepting mutually consistent but arbitrary target-generated state as the only oracle.
4. The added app controls at `transactional-matrix.test.ts:820–936` actually discriminate the named wrong count/order/index/poststate assertions. Valid cross-wiring reaches the authorized alternate receipt and fails the expected-delete-count oracle; it does not invent an authority refusal. Omitting either declared route refuses at exact plan preflight after only the already-closed bootstrap, with no exchanges. Denied receipt credentials produce actual linked 401 exchanges. Loopback source self-disagreement and target invalidity remain separately classified; they are not described as installed target qualification.
5. Original four scenario objects and coverage objects remain byte-pinned; the negative old-assertion control rejects a changed source object. Independently compared all four original parsed objects against the direct parent and verified exact action/assertion binding order for all six rows. All six wholeContract flags and overall claimEligible remain false. Required denominator stays 250 plus five optional rows; six pass and 249 remain unplanned/harness-error in the unfiltered actual report. These missing-plan outcomes are visible, not counted as passes.

## Independent execution

Pinned Node `24.16.0`, pnpm `11.20.0`. Commands executed in the detached native checkout:

```sh
node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs install --frozen-lockfile
node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs build
node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs exec vitest run packages/conformance/src/transactional-observation.test.ts packages/conformance/src/transactional-cumulative.test.ts apps/bdptest/src/transactional-matrix.test.ts --maxWorkers=2 --reporter=verbose
```

All exited 0. Focused result: **93 tests / 3 files passed**, 41.82 seconds. Logs: `transaction-set-conformance-native/{install,build,focused}.log`. This includes public/private brand boundaries, caps, old-object byte preservation, actual loopback negative controls, the actual local installed HTTP cohort and all new app countercontrols. `git diff --check 2019649 9eb9339` also exited 0.

The independent positive cohort reported 2,267.347 ms, 74 actions, nine scenario/reopen sessions, ten child launches and ten closed children. All five exact reopened receipt IDs were checked with source linkage, increasing generation and stable logical configuration. Its source is exactly 9eb9339; it uses the local built process, not the author's packed package. Local payload/policy/authentication hashes accordingly are not relabeled as the author's packed identities.

The initial shared clone checkout hit Git's alternate-store nesting limit. Flattening the new clone's own alternate path list to the existing seven read-only object directories allowed the exact detached checkout. This changed only new clone Git metadata; neither source checkout nor source files changed. Both source and review worktrees were clean at completion, and process inspection found no review-owned installed/test child remaining.

## Artifact audit and bounds

Independent audit is saved in `transaction-set-conformance-native/audit.json`.

| Artifact | Bytes | JSON value nodes | SHA256 |
| --- | ---: | ---: | --- |
| manifest | 130887 | 3647 | `65bf4e20a4cdf9753ea5e3cdfbe9808ca85394fcea82c408ecb3142bcb02b63d` |
| coverage | 23604 | 541 | `6a2af134253b25a19ae2917ff5842c58723bae437439b86966d9d79d0f27bb81` |
| fixture | 721 | 15 | `ac4d5834911126fcf4b3b436ddfbf2d0c1f1644c0b80bff587e5a875676a2724` |

New rows are 16 actions / 113 assertions and 22 actions / 157 assertions: 38 new actions, 270 new assertions, 74 total actions and 487 total assertions. They fit the accepted <=40 action increment. Manifest planning headroom is only **185 bytes below 128 KiB**, and 1,353 value nodes below 5,000. This is a future edit constraint, not a present failure or a changed parser cap. Runtime/parser/deadline limits were not widened. The singleton row has three live Resources; the cardinality row has two plus one tentative batch creation. Both remain below the four-Resource plan ceiling.

## Author packed evidence, separately inspected

The author run completed during this review. Read actual `transaction-set-conformance/packed-9eb9339a/{exits.json,tx-safe-summary.json,tx-positive.json,set.log,tx-regressions.log,ru-matrix.log}`. All recorded stages exit 0: build, protocol/server pack, install, installed set test, TX regressions and RU matrix. Test totals: installed set 1; TX regressions 44 across five files; RU checks 209 across four files. These are author executions, not additional native reruns.

`tx-positive.json` is 412,327 bytes, SHA256 `ee766518a167ff7c8381ade3e3cfc9f7019c6268466886db67e08effc9aa4b13`. It identifies exact source 9eb9339 and the newly installed node_modules server process entry. It contains six passing rows, 249 explicit missing-plan outcomes and claimEligible false. Both new rows' dynamic receipt exchanges have the expected original set source, status 200 and passing assertions. The packed summary reports 2,570.915 ms, nine sessions, ten closed children and five exact reopened IDs with increasing generations.

Packed identities inspected: payload `1b12629fe54c39c8f7864dbb13bdba16bcbe4298b0a5bd420fa19daaa4fb7ec2`; contracts `c67d12264b97e16c76bf68f850eaf394f61896bb72491e06603ef1d6debb335b`; exact route configuration `d5a731626ec5d3f6d2fd84ede17d8c2c74239cba1231813faa5218b7bbbc3d51`. Manifest/fixture identities agree with the frozen source audit. The summary also retains separate policy, authentication and logical configuration hashes; no raw credentials are reproduced here.

## Limits

No whole workspace suite, new packed build, typecheck/lint/boundary/Read-seal rerun, exhaustive numeric/policy/concurrency matrix, owned fanout, pagination, Event inventory, cold discovery or public admission review was performed in this bounded native pass. Existing omitted areas and both partial rows remain explicitly unqualified. This review supplies evidence for root's integration decision; it does not make that decision or promote claims.
