# Command POST minimum fields — bounded implementation proposal

Status: plan only, for native and Claude council followed by root authorization. No source authoring or execution is authorized by this document. This report was prepared by reading source, not by running its tests or service.

## Authority and parent

The direct operator ruling in `transaction-selector-command-operator-rulings.md`, recorded 2026-10-14T21:00:43Z, is “Ignore on mutation POSTs (recommended).” The three fields are `BDP-Minimum-Scope-Position`, `BDP-Scope-Epoch`, and `BDP-Authorization-View`. Normal authorization, idempotency and expected revision still govern commands; subsequent receipt reads still enforce their minimum. Global transport, header and body bounds remain unchanged.

This is a supporting slice of the sole plan at `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, [fork design](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/design.md). It is not a competing roadmap or profile activation decision. Metadata-only Link Event membership and retained Snapshot/continuation minimum S1-Q1 remain unresolved and outside this slice.

Inspected source: `/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft`, clean `c02ccc3706f2e258d380805162a8c47dbb73f4e1`, tree `088290751951c353797b984f690b7d9fd96deb80`, parent `eff171839fe4f37638284a4c3346428979d26b5d`. PR122 is pending; its four-file evidence delta does not change production from eff1718. Root must assign the actual ordinary authoring parent and reconcile merged ancestry before publication. No source from another lane is substituted into evidence.

## Exact contract amendment

Add an explicit paragraph to `docs/specs/bdp.md`, “Mutation command preconditions” (current lines 3565–3578), applicable to Read+Update and inherited by Transactional:

> On mutation POSTs to operation command targets, authorities MUST ignore the BDP-Minimum-Scope-Position, BDP-Scope-Epoch and BDP-Authorization-View request fields. Their presence, absence, completeness and field values do not impose a minimum-position barrier or affect command identity. Ordinary HTTP field syntax and transport/header/body bounds still apply, as do authorization, idempotency, carrier validation and expectedRevision. This rule does not change minimum controls on reads, including subsequent Mutation Receipt GET or HEAD requests.

In “Transactional HTTP consistency” (current 5766–5788), scope the minimum-visible-position paragraph to reads and cross-reference the command exception. Preserve response-context requirements, receipt current-context rules (5233–5244), conditional-read ordering (5800 onward), and lower-profile distinctions (3557–3563). No schema or catalog changes are needed: there is no new wire field, Problem, receipt shape or capacity vocabulary.

“Ignore” includes a lone field, partial tuples, empty field values, read-token-invalid strings such as `not a position!`, accepted comma-joined strings, unknown opaque positions, and complete foreign or far-future tuples. These are values of otherwise receivable HTTP fields. It does not authorize CR/LF injection, illegal HTTP framing, unlimited fields, oversized headers/body/target, an unsupported Content-Type, or a method on an unrelated route. Do not parse a copied tuple merely to decide whether to ignore it. The command law does not resolve malformed/partial read tuples or retained-page semantics.

## Source audit and smallest production change

| Boundary | Current source and behavior | Proposed action |
|---|---|---|
| All Transactional operation POSTs | `packages/server/src/transaction-http.ts:256–303`: canonical operation lookup, target/method/receiving limits, then `hasMinimumControls` forces development500 after its own auth checks | Delete only this command-specific refusal block. Retain normal shared receiving checks and subsequent carrier-specific admission. |
| Singleton/batch | Same file 336–398 forwards carrier/body/key/auth/Accept to `service.submit`; it does not forward or parse these fields | No service/submission/comparison/store change. |
| Sequence | Same file 304–335 checks prohibited header key, parses full sequence, auth, operation count, composition and negotiation; `sequenceObserved` receives principal/prepared carrier only | No sequence parser/evaluator change. Removing the special early block restores the same order as an identical request without copied fields, including syntax/auth combinations. Do not preserve the special block's accidental precedence. |
| Complete operation inventory | `transaction-discovery.ts:8–21` lists 12 mutation targets: create/update/delete Bead, create/update/delete Link, put/delete Alias, updateWhere, deleteWhere, batch, sequence | Assert this literal inventory in the new source observation; the 17 discovery members are not 17 commands. Use valid under-limit Selectors only, avoiding the separate Selector admission ruling. |
| Incoming transport | `transaction-runtime.ts:110–147` validates raw target/Host/method, builds Headers by appending raw fields, authenticates and receives bounded content; it does not filter these three fields | No transport or global bound change; ordinary duplicate-field joining still occurs. A comma-joined value test does not claim a raw duplicate-header wire observation. |
| Receipt reads | `transaction-http.ts:523–547` calls `transactionMinimumHeaders`, then `receiptAtMinimum`/`receiptPage`; consistency parser and finite wait stay intact | Preserve `hasMinimumControls` and its source-private export. Do not strip fields globally, change the helper to return false, or bypass receipt routing. |
| Other reads | Point/properties, Type, incident/alias/collection, Snapshot, Changes/Events, schema adapter and discovery paths retain their own minimum decisions | No changes or inference about currently unsupported read forms. |
| Read+Update | `development-runtime.ts:440–477` checks media/key/body, then calls mutate/sequence with body/key/Accept; none of the three fields is read. Server maxHeaderSize and timeouts remain at 483 onward | Already compliant. Add a small emitted installed-factory regression; no RU runtime or result-shape change. RU has eight singleton targets plus sequence, no batch, durable Mutation Receipts or Scope-history fields. |

The only production edit is deletion of the Transactional command block. In particular, normal key validation, expectedRevision, credentials, policy, parsing, retry comparison, response context and receipt retention are not rewritten.

## Proposed ownership: six files

1. `docs/specs/bdp.md` — the two tightly scoped normative paragraphs above.
2. `packages/server/src/transaction-http.ts` — remove the early command-only minimum refusal.
3. `packages/server/src/transaction-http.test.ts` — replace obsolete command refusal test, rename its auth companion, and add bounded actual HTTP comparisons below.
4. `apps/bdptest/src/transactional-collection-minimum-installed.test.ts` — append one scenario using existing `withInstalled`, exact raw executor, minimum/context helpers and current reporting. Do not run or refactor its other scenarios.
5. `apps/bdptest/src/installed-reference.test.ts` — append one six-exchange RU emitted-factory regression using existing launcher/request helper; no shared helper or launcher change.
6. `docs/design/transaction-implementation-progress.md` — state the new command rule, explicitly supersede command-minimum development500 descriptions (including the retained-refusal inventory around 1727), and keep read gaps qualified as reads. Preserve historical reports/failed expectations and avoid unresolved local-log links in published prose.

No runtime configuration, store/evaluator, public exports, evidence admission, schema/catalog/DSL, fixture artifact, startup switch, CORS policy, parser or source snapshot change. The Selector worker also needs transaction-http/spec; root serializes this slice first, then gives that worker the reviewed successor for reanchoring. A conflict or broader needed seam returns to root.

## Source test design and exact selection

Proposed literal titles below are the execution selectors; the draft must include an inventory that proves exactly these cases expand, and bind complete bodies, assertions and helper calls. These are planned counts, not executed results. No unfiltered HTTP file or matrix invocation.

| ID/title | Actual operations and independent oracle | HTTP | Fixtures / owner opens |
|---|---|---:|---:|
| A: `ignores partial and malformed copied read controls on each mutation carrier` | Replace old line1091 refusal case. Three fields × singleton createBead, batch, sequence = nine unique targets/keys. For each: successful command, full Resource GET, equal retry with only copied controls changed/removed. Independently pin ID/type/properties/metadata absence and result/receipt or sequence envelope; carry only opaque revision/receipt IDs from validated first result. Equal replay must not create another Resource or change revision/context. Distribute empty, invalid-token and comma-joined values across the three carrier columns. | 27 | 1 / 1 |
| B: `ignores complete copied tuples on every Transactional mutation target` | One seed batch creates a,b and an unowned victim Link. Execute each of 12 targets once: create c; update c; create/update l; putAlias(c); updateWhere over all then-live Beads; deleteAlias; delete l; batch create d; sequence create e; delete d; deleteWhere over remaining Links. Three final reads: complete Beads (a,b,c,e, exact expected fields and next:null), Links (empty,next:null), known alias404. Commands alternate complete same-epoch/view future position and complete foreign tuple. Validate every ordinary receipt/sequence result and independent effect expectation, including selected-set counts. Capture actual context from seed; assert no minimum wait callback is entered by commands. Do not infer a position from number of alias commands. | 16 | 1 / 1 |
| C: `copied read controls preserve key revision and negotiation admission` | Seed a (1); invalid direct key400 then only-key-corrected create (2–3); stale expectedRevision update produces failed receipt, unchanged GET, equal retained failed retry, fresh-key valid-revision update, changed GET (4–8); bad Accept406 then same-key acceptable create and GET (9–11); prohibited sequence header key400 then same member/body without header key succeeds (12–13). Copied controls are present throughout. Validate complete Problem/receipt fields and effect/non-effect state; the two revision mutations use different expected revisions as required. | 13 | 1 / 1 |
| D: `preserves authentication on commands carrying copied read controls without reserving keys` | Rename and retain existing auth test at1320: three carriers × unauthenticated/reader. 401/403, Resource404, actual receipt-row inventory empty, corrected authorized batch using same key succeeds. Keep ordinary auth/key scope and all assertions. Existing helper closes/read-only-queries/reopens for each of six keys. | 18 | 1 / 7, six read-only SQLite handles |
| E: `keeps receipt minimum enforcement after commands ignore copied controls` | Seed command with ignored malformed controls (1); same receipt GET with actual epoch/view + future pos99 →503 catch-up-timeout/Retry-After/current context (2); foreign view→409 (3); foreign epoch→410 (4); compatible minimum→200 and full original receipt equality (5); future HEAD→503/bodyless with same required fields (6). Use existing deterministic finite-budget minimumReads seam, no real sleep enlargement. | 6 | 1 / 1 |

A–E total: five tests, 80 HTTP exchanges, five store directories, eleven owner/listener instances, six read-only inventory handles; no installed subprocesses. Failed assertions must still close every owner/listener and restore timers/mocks. No read-side fallback is accepted as proof of command behavior.

Two isolated source controls, each built from a disposable copy and run once:
- Restore exactly the old operation-only minimum refusal block; select A. Named first successful-command assertion must fail with actual500/development marker (expected1, max27 HTTP; one fixture/listener). Preserve actual response and nested assertion, not merely generic test exit1.
- Bypass only receipt-root minimum dispatch, leaving commands/read parser/helpers intact; select E. The first receipt future GET must incorrectly return200 and fail the named503 oracle (expected2, max6 HTTP; one fixture/listener). Do not mutate `hasMinimumControls` globally.

Total TX source/control execution: seven case executions, expected83/max113 HTTP, seven directories, thirteen owner/listener instances, six read-only inventory handles. Any control unexpectedly passes or fails earlier is not accepted sensitivity and stops for disposition.

RU regression title: `ignores copied read minimum fields on Update singleton and sequence commands`. One emitted installed child/store, six requests: create via singleton, GET, equal retry with changed copied controls; create via sequence, GET, equal retry with changed copied controls. Use existing reference configuration/types, exact known Resource postimages and opaque revision equality. Assert applicable RU disposition and absence of Transactional epoch/view/position response fields and receipt envelope. The first pair uses a complete foreign/future tuple and the sequence pair a malformed partial value. No invented RU batch or durable receipt. Local emitted evidence is labelled separately from the fresh consumer TX proof. One execution only; 30s existing test bound and 120s external supervisor. Its child/HTTP inventory is additional to TX source totals: aggregate planned source/control89 HTTP, hard119, one child, fourteen listener/owner instances, eight directories.

## Fresh installed TX wire scenario

Append title `ignores command read controls while enforcing a later installed receipt minimum`. Reuse the existing captured immutable contracts, two trusted credentials and same-owner factory in the collection-minimum installed file. One fresh child/store; exact18 requests, nine distinct canonical URLs: four command targets (create-bead, update-bead, batch, sequence), three Resource points, the explicit-limit collection URL, and one validated observed receipt root. Do not cold-fetch discovery or Types as hidden setup.

1. GET Beads with explicit limit100: empty complete collection, actual genesis context; reuse this exact collection URL at18.
2. POST createBead a with a malformed complete copied tuple; completed validated receipt, actual current context.
3. GET a: full independently expected fields, captured revision.
4. Exact same-key/body singleton replay with complete foreign/future tuple; full receipt equality.
5. POST batch create b with one copied field; completed receipt.
6. GET b: full expected Resource.
7. Same batch key/body replay with empty/different copied controls; full receipt equality.
8. POST sequence containing one create c/member key with actual epoch/view and future pos99; validated sequence result, actual context.
9. GET c: full expected Resource.
10. Same sequence replay with malformed copied tuple; full result equality.
11. Batch original key with changed body and copied tuple: direct409 idempotency-conflict, normative Problem/context, no new resource.
12. POST singleton update a with wrong expectedRevision and copied tuple: failed receipt/revision-mismatch, no effect.
13. GET a: deep equality to3, including revision.
14. GET receipt from2 with actual epoch/view/future pos99:503 catch-up-timeout, required context and Retry-After.
15. Same receipt GET with foreign view:409 foreign-view.
16. Same receipt GET with compatible tuple:200 and complete receipt equality to2.
17. Same receipt HEAD future minimum:503, empty body and required headers.
18. GET Beads with explicit limit100: exactly sorted a,b,c, every full Resource equal its independently validated earlier record, next:null. This also rejects the conflicting body's proposed extra identity.

Use the existing raw executor's declared resolver channel for `bdp-authorization-view`; do not weaken its credential-like-header firewall. Validate the captured receipt URL as an exact same-Scope `/receipts/<single token>` root (no query/fragment/userinfo, existing URI/target bounds) before adding the read route. No sequence member receipt fetch, pagination fetch, restart or extra probe is hidden in the18.

Keep the existing 90s scenario AbortController/120s test and supervisor,5s exchange timeout,16KiB request headers,64KiB request body,32KiB response-body limit. Add a scenario-local exact18 counter/route-set assertion; do not widen the shared helper's existing80 maximum or treat it as permission for extra calls. Expected and maximum are18/one child; failed startup consumes that child allowance. No installed predecessor/negative cohort is proposed: the two source controls are the discriminators. Unexpected execution failure stops, preserves evidence and does not authorize a replacement child.

Use one frozen fresh package/install from the final source. Reuse the reviewed supervisor/package-verification approach with rebound paths and exact source/entry overrides. Match source→tar→consumer members and all actual runtime dependency bytes, recompute payload identity, verify selected Node24.16.0, configuration/contract/policy/authentication bindings, and check every observed wire status/body/projection. Source metadata is supervisor-verified attribution bound to separately built package, not child-attested provenance. The existing helper retains safe fingerprints/statuses; disclose that it does not retain complete wire bodies in its report. Never persist raw credentials. Exact request expectations and full response assertions execute before fingerprint reporting. Record actual child exit and independently verify PID and directory absence after cleanup, separately from the helper's `closed` flag. Preserve original plus cleanup errors and report evidence even if directory removal fails; do not silently refactor other scenarios' lifecycle behavior.

## Draft, measured gate and execution commands

First author isolated source/test/doc draft only after plan council/root release. Before any execution submit exact six-file diff/hashes, source test title expansion, each request/body/route and byte count, largest body/response estimate, all fixture/reopen/read-only-handle counts, control patches and intended failure assertions. The proposed counts above are not retrospective evidence. Any count, scope, semantic or bound mismatch requires root disposition before execution; do not weaken assertions to fit. Confirm all normal body/target/header limits unchanged in the diff. Include alias/sets ordinary successful fixture shape checks, so a Selector grammar mistake cannot masquerade as ignored-minimum coverage.

After root measured release, pinned commands use Node `/Users/dbox/.nvm/versions/node/v24.16.0/bin/node` and pnpm `/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs`, explicit isolated cwd and external120s per-command supervisor:

- offline frozen-lock dependency setup if needed; `pnpm typecheck`, `pnpm build` before emitted consumers;
- `pnpm exec vitest run packages/server/src/transaction-http.test.ts --testNamePattern <anchored escaped union of A–E literal titles> --maxWorkers=1 --reporter=verbose --reporter=json --outputFile=<outside-checkout-json>`;
- each disposable source control rebuilds then selects exactly A or E, with named expected-failure and cleanup postprocessing;
- `pnpm exec vitest run apps/bdptest/src/installed-reference.test.ts --testNamePattern '^ignores copied read minimum fields on Update singleton and sequence commands$' --maxWorkers=1 ...`;
- format/lint, dependency boundaries and unchanged Read evidence verification; no catalog/artifact regeneration, broad workspace suite, smoke cohort or unfiltered matrix;
- freeze clean signed ordinary source, then one protocol/server pack, offline consumer install and exact byte/dependency audit; select only the new18-request TX case in the existing installed file with explicit final source/installed entry env bindings.

All loops must publish progress and bounded actual counts; preserve stdout/stderr, JSON, exit status, named assertion failures, temporary paths, observed PID lifecycle and final absence checks. Stop after unexpected failure and retain the original attempt; root decides whether a repaired source/run is authorized. Static/type failures are recorded, never relabelled as runtime results.

## Done when and exclusions

- Six-file source inventory and production deletion are exact; all12 command targets share the ignored-field boundary, and read handling remains unchanged.
- Normative text states the operator's applicability rule for both write profiles, without admitting invalid HTTP or weakening expectedRevision/auth/idempotency.
- A–E and RU regression pass; both source controls fail at the intended wire oracle and clean up; ordinary failures/replays and every named target's actual effects are checked.
- Frozen fresh installed18/one-child passes with actual provenance/configuration/wire and cleanup evidence; response assertions are complete even where retained reports are fingerprints only.
- Native and Claude review, root acceptance, exact CI and ordinary ancestry gates remain required. No full-profile qualification, public advertisement, release, Preview2, new barrier, optional History/Types work, or resolution of retained read-minimum semantics follows.

No implementation or test was executed while preparing this plan. Only this supporting report is written; source checkout remains clean and frozen.
