# First cumulative Read+Update executable slice

Base: `1f2d75fbaa67f631e72bd79384c829d7588b7277`. New isolated checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-executable`, branch `codex/vickie-read-update-executable`, origin `https://github.com/gastownhall/bdp.git`. Prior reviewed checkouts stay frozen.

## Concrete scope

1. Add a source-private cumulative Read+Update artifact path built from the canonical catalog selection, exact v2 manifest bytes and exact fixture bytes. Reuse existing bounded parsers, fixture oracle validation, factory branding and runner. Preserve public createConformanceArtifactBundle Read behavior and the existing diagnostic-only development path. The new path is explicitly partial/non-attesting, with required, selected, missing and observed IDs distinct; no claim eligibility or Read sealed evidence change.
2. Author an actual v2 RU manifest and fixture under conformance matrices/fixtures. Start with cold authenticated RU discovery (RU schema, declared mutation routes, preserved default-order semantics), create/read/guarded update/read/delete/read, duplicate-key exact replay and changed-input conflict, stale revision rejection with unchanged afterstate. Map each assertion to actual catalog obligations; select only rows whose entire contract the case observes. Keep uncovered inherited/companion obligations explicit rather than claiming broad rows from one happy path. Rebind inherited discovery prerequisites to the RU replacement; never execute the retired closed Read discovery row. No fabricated lifecycle observations or N/A.
3. Add real exact credentialed HTTP execution against installed P1, using its captured target configuration and process lifecycle. Capture installed payload/contracts/policy/auth identities, Scope/profile, source commit, catalog/manifest/fixture hashes from the actual target/configuration boundary. Reuse the existing raw HTTP executor, transport credential handles, shared assertion evaluator and harness cleanup. No server or evaluator changes.
4. Add negative controls: ordinary Read/development factories still reject inappropriate writer artifacts; omitted cumulative obligation remains visible; mismatched profile/Scope/fixture/provenance fails before execution; modified expected result fails actual HTTP assertions; missing/wrong credential and stale guard preserve afterstate.

## Expected files

- `packages/conformance/src/artifact-bundle.ts`: only a separately named source-private profile factory path plus shared parser factoring as needed, pending root approval of this boundary change.
- `packages/conformance/test-support/`: cumulative canonical artifact assembly/execution binding; reuse `profile-contract-plan.ts` selection rather than another requirement list.
- `packages/conformance/matrices/read-update-v1.json` (manifestVersion 2), `packages/conformance/fixtures/read-update-reference-v1.json`.
- `packages/conformance/src/*test.ts`: artifact/selection/runner negative controls and citation coverage for selected contracts.
- `apps/bdptest/src/read-update-matrix.test.ts` and minimal test support: installed P1 real HTTP execution.
- A small implementation progress document/report with exact selected/uncovered IDs and provenance; no normative edits.

No public package exports or evidence schema/admission changes, no shared server/evaluator implementation edits, no Transactional execution.

## P1 integration dependency

P1 source is still being repaired in `vickie-update-installed-repair`; this branch does not contain its factory. Need root to provide a frozen accepted successor and ordinary-merge authorization (or exact reusable installed worker boundary) before observed integration. Can implement artifact/manifest pieces first after root reviews this plan. Do not bind final evidence to provisional4078 or import arbitrary source from a mutable sibling checkout. Request P1's concrete startup IPC target/config identity shape, credential-handle setup and installed payload digest source; reuse those values rather than parallel identity calculation.

## Done when

- Canonical cumulative selection remains intact; retired discovery cannot execute; missing cases are reported, never passed/N/A.
- Bounded selected cases run via exact HTTP against the same actual installed P1 configuration, validate afterstate and retry/guard outcomes, and include inspectable source/config/artifact identity.
- Counterfactual response/credential/config changes fail meaningful controls.
- Existing Read artifact/runner/evidence behavior and diagnostic RU behavior retain tests.
- Focused tests, complete conformance suite plus relevant bdptest matrix/install tests, build, typecheck, boundaries and format complete successfully; report exact source, observed IDs and remaining obligations.
- Freeze reviewed slice without public admission, evidence claim, push or PR mutation. Subsequent work fills remaining cumulative requirements using this executor instead of expanding the planner.
