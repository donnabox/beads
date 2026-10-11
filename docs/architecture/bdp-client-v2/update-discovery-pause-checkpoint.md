# Operator pause checkpoint — private cumulative RU HTTP composition

Checkout `/Users/dbox/repos/wt/vickie/bdp-update-discovery`; branch `codex/vickie-bdp-update-discovery`; base `cd22ac3e388be3baffaf1d3a4caac7c4bd288502`; GitHub origin configured. No outward writes.

## Completed first commit

`fc2c708` — Extract private Read HTTP route projection without changing admission.

Files: server `index.ts`, new private `read-http-routing.ts`, executable smoke output/deep-import allowlist. The public brand, GET/HEAD listener, alias identity gate, Problem branding, and final conditional/HEAD application remain at the existing boundaries. Route and representation-response helpers are reusable internally.

Verified: frozen install; build; workspace typecheck; boundaries; focused 5 files /125 tests (Read HTTP handler, Node listener, server index, bdptest/bdpbd server composition). Logs `extraction-*.log`. Formatting run on all three changed files. Installed smoke/evidence and wider regression runs were NOT run before pause.

## Explicitly unfinished second commit

`96ae78efc21de2936b910137d74bfd5532deee53` — WIP private canonical RU HTTP composition before qualification.

Files: `packages/server/src/development-authority.ts`, `development-runtime.ts`, new source-only `apps/bdptest/test-support/development-discovery.ts`.

Adds private authority discovery/parsed-read callbacks, dynamic scoped alias classifier, private runtime callback/scopePath seam, reuse of existing receiver's fixed operation table, construction-failure authority cleanup, and source-only `/acme/` canonical discovery/directory composition. Default public facade supplies no callback and has no new options. No store/evaluator/schema/public export/config/evidence edits.

Build and workspace typecheck completed successfully after draft. IMPORTANT: source-only fixture is not imported by a test yet, and the workspace tsconfig directly includes apps/src, not apps/test-support. Thus that typecheck does **not** establish the new fixture typechecks. No new discovery HTTP tests were written or run. No conformance/profile claim. Do not publish this WIP as validated implementation.

Next resume: narrow factory authority argument to needed readonly capabilities (current type gives full authority); validate malformed/missing scopePath and factory return/cleanup; typecheck fixture through actual test import (its response mutation and callback contextual typing merit checking); scrutinize scoped routing/unknown-operation and directory-query behavior; complete cold navigation + actual all-nine-target effects, method/Accept/bounds, owner committed-prefix and pagination snapshot assertions. Add private-entry/package deep-import and unchanged default startup proofs. Then run full planned checks/regressions/evidence/both smokes. Do not advertise 24h retention or untested representation limits. The source fixture currently advertises only page, selector, request-body, sequence and diagnostic bounds; their new HTTP boundary witnesses are still absent.

The plan remains `/Users/dbox/repos/wt/vickie/bdp-integration-logs/update-http-discovery-composition-plan.md`. TX worker separately edits the narrowed authority-read interface and transaction store/read; no overlap here beyond disjoint executable-smoke inventory additions.

No tests were interrupted or killed. Short runs completed. No remaining test process owned by this checkout was observed. Working tree clean after WIP commit; pause honored, no further work initiated.
