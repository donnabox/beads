# Command-minimum cumulative mapping implementation plan

**Plan only.** No product edits or execution are authorized here. Implement after the active Selector author releases the shared spec path; this prerequisite does not displace that serving fix.

## Source and authority

Inspected clean `vickie-command-minimum` at `727e919c2893bd34d5e72ea4659302a999bc4a6f`, tree `734bf9daf40fd17d0fe2bf86b085598589122f2f`, equal to root's verified merged PR123 tree at `757aedc8008386745d59d5005498a9256d6ea76e`. Runtime evidence remains attributed to f2ac; repaired spec/CI evidence belongs to 727e. The older follow-up audit's 5792 authority placement is historical.

Current authority is `docs/specs/bdp.md:3579–3586`, `#mutation-command-preconditions`: mutation POSTs ignore BDP-Minimum-Scope-Position, BDP-Scope-Epoch and BDP-Authorization-View irrespective of presence/completeness/value, without a barrier or command-identity effect; normal syntax/bounds, authorization, idempotency, carrier validation and expectedRevision still apply. Transactional inherits this through `#transactional-http-consistency`, at 5797–5801, explicitly preserving receipt GET/HEAD minimum behavior.

Use one new normative ID, **`read-update.http.command-minimum-fields`**, with requiredProfile read-update, inherited by TX. This is a proposed new ID, not existing coverage. Cite the current shared paragraph exactly; do not point backward to its superseded location, repurpose the read-side `transactional.changefeed.consistency-fields`, duplicate a TX authority row, or introduce a retirement. The rule covers all mutation POSTs, not only requests a client describes as “copied.”

The four optional wording suggestions in `transaction-command-minimum-ci-repair/root-delta-dispositions.md` are not blockers: preserve the operative paragraph, existing explicit TX read carve-out, cumulative applicability and link. Do not add an editorial pass or new interpretation to this mapping. No user ruling is missing.

## First implementation slice: seven paths, metadata and honest omission

| Path | Exact bounded change |
|---|---|
| `packages/conformance/catalog/read-update-v1.json` | Add the one normative row with exact shared-authority citation and a title stating that the three read-consistency fields neither gate mutation POSTs nor affect identity. Keep the preserved admission/read boundary in its requirements. |
| `docs/specs/bdp.md` | Add the matching ID/title in the Read+Update conformance table only; no operative prose or TX table change. This path is actively owned by Selector, so sequence the slice after handoff/rebase. |
| `docs/design/profile-contract-planning.md` | Add the row in catalog order as Inherited, explaining RU singleton/sequence versus TX singleton/batch/sequence observations and preservation of read minima. |
| `packages/conformance/test-support/read-update-coverage-v1.json` | Insert the same row in canonical required order with status omitted, wholeContract false, bindings [], and a reason distinguishing existing private implementation tests from absent cumulative executable evidence. |
| `packages/conformance/src/profile-contract-plan.test.ts` | Add one focused static test, proposed title “inherits the command minimum-field exception without executable credit”: exact ID/clause presence, RU and TX selected, Read excluded, no retirement/duplicate, and omission in current executable artifacts. Preserve the complete inheritance audit. |
| `packages/conformance/src/transactional-cumulative.test.ts` | Update the required count 250→251 at current line80; assert the new ID remains required but absent from the private executable manifest. No transport or limit change. |
| `apps/bdptest/src/transactional-matrix.test.ts` | Update current required250→251 and total report255→256 at lines340–341. Implemented13/pass13 and existing20sessions/21children stay unchanged; the additional report row is unobserved, not a new executed case. |

Why the coverage edit is mandatory: `test-support/read-update-artifact-set.ts:217–244` requires the complete canonical inventory and explicitly requires unplanned rows to be omitted/false/empty. TX private coverage instead mirrors only implemented manifest rows (`transactional-cumulative.test.ts:95–118`), so **do not** invent a TX coverage binding or append an omitted row to its differently shaped inventory. The canonical selector already inherits preserved RU requirements (`profile-contract-selection.ts`); no production selector/runner/factory change is needed.

These seven paths do not touch any active Selector server/test/installed path. The only ownership intersection is docs/specs/bdp.md, deliberately sequenced rather than concurrently authored. Before drafting, reanchor to the actual post-Selector parent and recheck counts/citations; stop if unrelated catalog changes invalidate these exact deltas.

## Bounded verification for the mapping slice

Measure and freeze the seven-file diff first. Run one selection of seven CPU-only cases: the new test; RU catalog's “strictly parses and binds every citation to the current specification text”, “scopes every row to the Read+Update profile with the profile's id prefix”, and “mirrors the specification's Read+Update conformance rows exactly and in order”; profile-contract plan's “locks every inheritance-audit row and replacement to the catalogs”; TX cumulative's “binds the canonical cumulative TX selection and exact source digests without qualification”; cumulative artifact's “pins cumulative source digest map, catalog order and every coverage binding”. Use exact describe-qualified selectors where applicable. Static inspection must confirm zero HTTP/listeners/stores/children for these selected cases. Do not run the whole cumulative test files, which contain loopback tests, or the installed matrix merely to check two count edits.

Appropriate format/type/lint and unchanged Read evidence verification remain required; source syntax/citation checks are not installed proof. Before execution, retain exact argv, selection, source/diff hashes and a 120-second per-command bound, with no automatic retry. Actual exact-head CI remains a separate integration gate. No new fixture, package, archive or public admission is needed for this first slice. Historical artifacts/digests stay immutable; newly constructed selections naturally have a new catalog digest.

## Subsequent executable proof: separately measured and released

Mapping closes the missing-ID defect only. Before giving this obligation any whole-contract or public-profile credit, author a narrowly measured executable row under that **same ID** in an approved RU shard and the private TX manifest, with their respective coverage bindings. Likely paths are `matrices/read-update-v1.json`, RU coverage, `test-support/transactional/manifest-v1.json`, TX coverage, and the existing `read-update-matrix.test.ts`/`transactional-matrix.test.ts` observation/control guards. Do not add a new shard, fixture capability or execution seam simply to avoid current bounds.

Use the existing RU6 and TX18 cases as concrete oracle sources, not fabricated runner reports. Observe valid absent/partial/empty/malformed/foreign/future field combinations over applicable singleton/sequence and TX batch carriers; same-key same-body replay; changed-body conflict; ordinary authorization/revision/carrier admission; full unchanged/poststate and exact result/receipt/sequence shapes. TX must additionally retain subsequent receipt GET/HEAD minimum refusals and compatible read. Empty fields need actual raw serializer request-header observations: fetch accepting an empty input value alone is insufficient. Do not interpret invalid ordinary HTTP syntax as permitted by the ignore rule.

Bind all observations to actual launched payload/source/contract/fixture/configuration identities, exact granted routes and genuine request/response evidence. Retain the already demonstrated named old-command500 sensitivity and receipt-bypass200-vs503 sensitivity, with source/derivative identities and failure-specific adjudication. Do not silently replay old controls or assume their historical execution covers new DSL bindings. Measure request bytes/captures/canonical nodes, controls, listeners/children/temp ownership and runtime ceilings from the drafted rows before any release; this plan does not invent a final request count or authorize an installed run.

Existing private f2ac evidence proves its authored cases only. A new catalog digest or manifest is not retroactively bound by that run. Fresh exact-source archive/consumer/runtime dependency and installed-session binding is needed for new executable claims. Declare unobserved field/operation partitions and keep wholeContract:false until a reviewed clause-level partition is actually covered; claimEligible remains false. Public qualification still requires every applicable cumulative Read/RU/TX obligation and the ordinary admission path, independently of this row.

**Blockers:** no semantic blocker. The active spec-file ownership is the only immediate scheduling constraint. Executable proof remains a later measured prerequisite, not a reason to expand this metadata slice or interrupt Selector work. Root owns prioritization, release, CI and merge.
