# RU admission bare-media comparators — frozen author handoff

Clean frozen **6f3e3bd02a09844422c5a00651cb14c90290d55b**, direct successor of **8e12162e63502660cc562503c9506789919ab604**. Checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-admission-comparators`, branch `codex/vickie-read-update-admission-comparators`, origin gastownhall/bdp. Remote role-session guard passed before work and again before commit. No push, promotion, public claim or source changes outside the assigned successor.

Three files changed: `packages/conformance/matrices/read-update-v1.json`, `packages/conformance/test-support/read-update-coverage-v1.json`, and `apps/bdptest/src/read-update-matrix.test.ts`. The allowed fixture required no edit. No runtime/schema/catalog/runner changes or deadline/cap changes. Whole26, required129, authored47, uncovered103; both new flags and claimEligible remainfalse pending review.

## Dispositions

- **F1 implemented.** Nine new raw `media-route-<target>-corrected-bare` actions immediately follow the parameterized successes. Each clones exactly the target, credential reference, raw body, Idempotency-Key (or both sequence member keys), empty captures, and complete result fact/schema assertions, changing only its action ID and Content-Type to bare `application/json`. No setup, route, Resource, capture or extra afterstate is added. Deletion and alias cases repeat their actual retained requests and require the same successful outcome. The existing afterstate checks follow the comparator. This compares media acceptance and asserted result facts without inventing raw JSON-body captures or an unobservable opaque-revision equality.
- **F2 partition documented, no Cartesian expansion.** Coverage reasons explicitly name createBead-only grammar/boundary/distinct-repetition and structured-suffix/non-ASCII decoder classes, representative categories on every other carrier, and pending independent audit. Per root’s ruling, these are intentional equivalence partitions, not unobserved requirements demanding every class×route. Same-body successful key correction, immediate no-effect state and subsequent intended afterstate distinguish retained400 from200; no hypothetical invalid200 retained disposition is introduced. Media reason records the added bare comparator as implemented and awaiting review, not yet promoted.
- **F3 strengthened.** The static completion test compares each missing/new malformed/repeated or wrong-media refusal’s whole afterstate action (except its ID) to the exact carrier baseline, including both Resource collections and alias witnesses for alias carriers. Header matching is case-insensitive by name. Both declared-wrong media variants must retain the correction’s exact key(s), including sequence keys. Bare comparators must deep-equal their parameterized originals after only the allowed differences, and be immediately adjacent.
- **F4 strengthened and actual runner behavior clarified.** Every raw input control now pins the perturbed exchange’s entire ordered failed-assertion set, actual response status, and that it is the last observed exchange. The accepted-update/refused-create perturbations can change later state, but the runner stops on the first failed exchange; downstream assertions are not executed in these control runs. Thus we do not require an impossible whole-row single-failure count or call hypothetical downstream effects observed failures. A new control changes the bare create-Link comparator to text/plain: actual415 fails its exact status/schema/result fact oracles despite a retained successful key.

## Preservation and budgets

`audit.py` verifies original45 scenarios byte-for-byte unchanged; all46 other8e scenario blocks are also byte-identical. Only the media scenario gains nine actions. Exact ordered assertion bindings pass; both route digest pins remain `a8d5687519da4040d2f23989f86f9bc804d30ac6c91a85074397eaf7c92de712`.

Final formatted manifest: **3,328,407 bytes**, +25,504; **89,024 nodes**, +662. Headroom:865,897 bytes below4MiB and10,976 nodes below100,000. The node delta fits the reserved1,500; no budget widened. `source-ledger.json` is exact; the pre-format `edit-ledger.json` records larger temporary whitespace only.

| Row | Actions | Raw HTTP | Client | Assertions | Peak Beads/Links |
| --- | ---: | ---: | ---: | ---: | ---: |
| Key |284|196|88|2303|36/3|
| Media |183|149|34|1341|11/4|

All added actions are immediate replays, so live counts and setup remain unchanged. Total new-row actions467; no hidden setup or extra state population.

## Executed validation

Pinned Node24.16.0/pnpm11.20.0, frozen install, build before process tests, maxWorkers2. Logs/scripts and explicit exits are local to this directory.

- Focused admission tests: **18 passed**,99 intentionally unselected; `focused.log/json`, exit0.
- Complete app matrix + cumulative binding suite: **123 passed /2 files**,90.96s; `local.log/json`, exit0.
- Typecheck, boundaries, format, explicit large-manifest formatting, lint and diff-check: exit0 (`checks.json`). Lint retains75 warnings/two infos; no warning-free claim.
- Fresh clean-source protocol/server tarballs and install, then packed matrix/exact-client/development-server/installed-read-oracles: **152 passed /4 files**,82.56s; `packed/matrix.log/json`, exit0. All47 authored sessions pass. Aggregate47-row job13.784s; both admission rows1.493s. These local timings are not CI headroom proof.
- New bare comparator sensitivity passes; all existing raw controls now pass exact local failure-set checks. No previous evidence was relabeled. No full8k suite or redundant smokes were run for this bounded evidence delta.

Fresh source/config/contracts/fixture/routes and observation counts are in `packed-summary.json`; full observations are `packed/progress.json` and `packed/admission.json`. Manifest binding is `0c6e1c9e9853e82e714e919ba23fa33dbf267c8a36ae36aa9457c1a9d3efc64a`. Runtime payload is measured unchanged from8e (`819c09e3ade4bbc533c118a956af23e22d44d5b7f87a73aef3d391a98dc11093`); this is not the separate a88b integrated runtime payload.

Exact source stayed clean after packing. `cleanup.json` and process snapshot show no surviving run-owned installed child and no timeout/forced termination. Other agents’ source/processes were untouched. Root owns independent fix review, any precise promotion edits, integration and publication; this author handoff is not that approval.
