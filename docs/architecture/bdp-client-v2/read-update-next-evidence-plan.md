# Next installed Read+Update evidence slice: retained retries across carriers

Plan only, 2026-10-13. The sole plan of record remains the fork design document; this is its bounded implementation proposal. No source edits, tests, profile promotion or outward actions performed.

## Basis and recommendation

Inspected clean frozen `31aab4db0f61e3d8b1139a7b78ff1ff10f890836` at `/Users/dbox/repos/wt/vickie/vickie-profile-review-closure`. The actual cumulative coverage contains **129 required IDs, 43 authored scenarios, 26 whole-contract flags and 103 uncovered IDs**: 17 authored partial rows and 86 omitted rows. The matrix is 2,038,702 bytes, below its existing 4 MiB bound. The installed run remains `claimEligible:false`.

Author **two new row-local scenarios**, not a runtime change:

1. `read-update.idempotency.cross-carrier` — retained success/failure and conflict comparison across singleton and sequence, in both directions.
2. `read-update.sequence.idempotent-retry` — retained results reprojected at the retry's current indices/names, while independent GETs prove current state was neither overwritten nor freshly returned as the old result.

These are useful client guarantees: a client can switch between a singleton and a sequence while retrying, and can reconstruct an interrupted workflow without reapplying completed writes. Existing installed evidence only exercises same-carrier singleton replay and some retained failed local bindings. The two omitted rows are not filled by those neighboring assertions. This proposal addresses that evidence gap using the already implemented durable lifecycle; it does not create retry semantics.

Keep **both new wholeContract flags false** for this bounded slice. If implemented as proposed, the useful change is 45 authored scenarios, 19 partial plus 26 whole, still 103 uncovered. Those are prospective inventory arithmetic, not a delivery result or a target to optimize.

## Source and harness feasibility

- Both authority methods call the same `lifecycle.submit(carrier, principal, …)` (`packages/server/src/development-authority.ts:307–328`). They do not own separate stores or key maps.
- The installed HTTP receiver dispatches singleton and sequence into those methods (`development-runtime.ts:451–467`). A singleton has the key header; a sequence has member keys and no carrier key.
- `sequence-executor.ts:176–189` takes retained disposition and applies the **present** operationIndex/name, parsing Problems with their status. Existing source test `sequence-executor.test.ts:601` exercises cross-carrier renamed positions. This is supporting implementation evidence, not installed HTTP qualification.
- Existing `installed.exact-request` in `apps/bdptest/test-support/read-update-exact-client.ts` already supports the needed POST bodies, header keys, captured Bead GETs and static discovered mutation targets. Nested sequence fields can be authored literals at the existing fixed Scope; this plan does **not** require recursive body-binding interpolation.
- Runtime configuration, installed artifact capture, credentials, exact-route guard, runner and executor remain unchanged. The matrix's Scope → service-desc → Operation Directory prefix supplies mutation target bindings. Current reference task Type accepts the simple authored properties already used throughout W2; no new Type/schema fixture is required.

## Exact citation scope and limits

### `read-update.idempotency.cross-carrier`

Catalog anchors: `#idempotency-keys` and `#operation-directory-and-singleton-targets`. The former at `docs/specs/bdp.md:3198–3200` defines the shared namespace; semantic normalization follows at `:3202–3229`. The latter begins at `:2064` and defines singleton operation records and carrier equivalence. The new observation covers one authenticated principal and one Scope, identical operation records between carriers, retained Resource successes and a retained resource-not-found failure, changed-record conflicts and absence of reexecution.

Partial: no elapsed retention/expiry/tombstone witness, in-flight presentation, changed authorization, other principal/Scope, alias/pinned/local-binding comparison, or exhaustive eight-operation corpus. Do not reinterpret “one retention rule” as qualified merely because an immediate retry works. Exact equality of unchanged records is sufficient here; broader semantic-identity variants remain on their own omitted ID.

### `read-update.sequence.idempotent-retry`

Catalog anchors: `#duplicate-keys-and-retained-dispositions` (`:3245`; retained equality/current positions `:3285–3289`, retained postimage `:3307–3310`) and `#sequence-request-envelope`. Observe success and failure repositioning, present versus absent operationName, retained original Resource images, a fresh independent inserted member, and refusal to reexecute the old failure after its missing subject becomes live.

Partial: the citation's admitted-work/lost-response recovery obligation is **not** established by deliberately retrying a completed response. No disconnect, concurrency barrier, expiry, crash or restart is introduced here. Names used on these rows create no `@name` dependencies, so reorderings are valid. Never use final result order as proof of serial execution or intermediate commit visibility.

No coverage credit is assigned to `sequence.retained-failure`, `idempotency.semantic-identity`, `principal-isolation` or other neighboring rows simply because a subcase resembles them. In particular, the retained-failure row also cites re-keyed creators/dependents, outside this proposal.

## Proposed row A: shared keys with observable state divergence

Use distinct case keys and Resource IDs under the configured Scope. Follow discovered targets. All mutations use the existing operator credential and valid small JSON bodies.

1. Singleton create Bead A under key A with properties `{n:1}`. Assert 200, created outcome, exact authored properties/metadata, canonical identity and complete normative result schema. Capture its original Resource, revision and disposition.
2. Fresh-key update A to `{n:2}`; independent GET asserts exact authored state and captures the complete current record. Present the original create/key A as a named sequence member. Assert retained created outcome, original Resource/revision, current sequence position/name, and no duplicate result. GET A must still equal the newer record. This distinguishes replay from reexecution and from refreshing the retained image.
3. Present key A with a changed creation record through sequence: assert member idempotency-conflict (outer HTTP200), full Problem tuple/schema, and unchanged A. Present the original singleton again: the original retained disposition must remain intact.
4. Reverse direction with Bead B: initial sequence create/key B; fresh-key update to a different state; singleton retry receives original created disposition, **without operationIndex or operationName**, while GET retains the newer state. Changed singleton input returns HTTP409 conflict; a final original-input retry proves no overwrite of the retained record.
5. Singleton update of an actually absent Bead F under key F returns HTTP404 resource-not-found. Prove absence first. Create F with a fresh key and known state. Retry the failed update as a sequence member: outer200/member404 with the retained Problem, then GET proves no application. Fresh-key otherwise identical update succeeds and changes F; retrying F still returns the original failure.
6. Repeat the failure direction once in reverse: initial sequence missing-subject update, then create the subject, singleton retry still404, fresh-key positive control succeeds. This proves a shared failure namespace rather than only successful-create sharing.

Resource GETs can use fixed `family:http` actions with `target:{binding:"beads",path:"..."}` for authored IDs and the existing captured-ID client GET route for allocated IDs. Prefer explicit IDs for this row to avoid unnecessary binding machinery; do not add fixture bindings solely for these paths. Full original-disposition comparison is possible within the same carrier; cross-carrier comparisons explicitly compare the disposition fields and independently assert current index/name rather than comparing envelopes that lawfully differ.

## Proposed row B: replay in present sequence positions

1. Submit three independent members: allocated create A (named), update a missing fixed Bead F (unnamed, retained404), and explicit create B (named). Assert full response schema, exact count/index/name pattern, authored success bodies and full failure tuple. Capture both created Resource records, including allocated A identity.
2. Fresh singleton operations advance A and B to distinct known properties and create F in a known state. GET all three independently and capture current records. The actions that establish this divergence must themselves be asserted, not treated as unchecked setup.
3. Retry the three original keys/records in a new order, rename one member, remove a previous name, add a name to the former unnamed member, and insert one independent fresh create. Assert **the entire ordered index/name/outcome-or-code array**, all expected absence-of-name cases, exact length, original retained Resource records and the retained failure. The fresh member must commit and be readable.
4. GET A/B/F and the fresh member. A/B remain at the newer state; F remains at its pre-retry state. Thus old results cannot be reconstructed from current GETs, and old failure cannot be reexecuted successfully. Optionally repeat the same reordered request once to check its freshly inserted member is now retained too; no process seam is needed.

For the failed disposition, compare every actual non-positional field using literals or existing captures; do not silently narrow equality to `code` alone. Use the existing normative closed response/Problem wrappers. A historical `detail`/diagnostic field, if present, must be accounted for rather than discarded by a test helper.

## Independent oracles and negative controls

Static authored properties/metadata/outcomes and direct GETs are independent of captured replay values. Captured opaque revisions/identities are used only for equality where the protocol intentionally leaves their content opaque. Do not derive expected poststate from the very response being asserted.

In the existing app test's response-corruption layer, add a compact row-local control set, each applied once **after actual HTTP** and required to fail a named assertion:

- substitute the newer live record into one retained success;
- change one retained revision while preserving shape;
- return a success for the old retained failure after F exists;
- reintroduce an old index or name, including an unwanted name on the currently unnamed member;
- return a singleton result carrying a sequence envelope field;
- corrupt one independent final GET to the replay's old postimage;
- turn the changed-input conflict into success or alter the final original-key replay.

Schema validity alone is insufficient. Controls must run the shared assertions, report exactly one application and fail for the intended semantic assertion (schema failure is appropriate for the forbidden singleton envelope). Keep original exact wire records unchanged; a corrupted assertion input is explicitly a negative control, never the target observation. Do not edit the raw executor or synthesize target responses.

## Ownership, dependencies and done-when

Expected edited files only:

- `packages/conformance/matrices/read-update-v1.json` — two actual HTTP/client scenarios in canonical catalog order.
- `packages/conformance/test-support/read-update-coverage-v1.json` — two omitted→bounded entries with row-local assertion IDs and explicit partial limitations; flags remain false.
- `apps/bdptest/src/read-update-matrix.test.ts` — targeted actual-installed tests, row-local negative controls and exact route-digest pin updated from actual added fixed GET routes if necessary.
- A narrow existing progress document only if root requests one. No normative catalog/spec/schema, fixture, Type/provenance, runner/executor, H1/M1, runtime/store, CLI/staging or shared configuration edits.

`packages/conformance/src/cumulative-artifact.test.ts` already verifies inventory/order/bindings dynamically and pins the unchanged whole-contract list. It should need no edit. The app's103 uncovered assertion should stay unchanged. Added static GET routes can legitimately change the existing exactConfigurationDigest pin; record the exact route delta, recompute from the actual mode, and do not relabel old evidence. Configuration/contracts/payload/source binding mechanisms remain intact.

No product ruling or new runtime mechanism is needed. Selector complexity, metadata-only incident Link behavior, Types/provenance and optional History are excluded. If the implementation unexpectedly requires a new binding operation, route wildcard, test clock, policy replacement or runtime seam, stop that expansion and report it instead of widening this slice.

Done-when after separate authorization: build with pinned Node24.16/pnpm11.20; focused cumulative-artifact and these actual installed rows pass; each negative control discriminates; full current installed RU matrix passes within existing cooperative90s/outer120s budgets; a fresh packed target run records actual source/payload/configuration with both new rows, all129 IDs preserved, `claimEligible:false`,26whole/103uncovered unchanged; applicable type/boundary/format checks pass. Preserve the current five-second individual exact HTTP timeout, fresh per-row disposable stores and existing finally cleanup. No timing sleeps, concurrency claim, long retention wait or new budget mechanism. Freeze clean code and external exact run evidence for independent review. No implementation has been started by this plan.

## Accepted council dispositions before implementation

F1: Add fourth owned file, fixtures/read-update-reference-v1.json, for named fixed-path client GETs. Use client GETs so the existing post-wire control layer can discriminate afterstate; raw HTTP remains available for absence but is not needed here.
F2: Ordered tuples project operationIndex only. Per-member assertions check name presence/absence, outcome, resource equality and every normative Problem field, with optional detail explicitly absent. No mixed optional-field tuple and no runner change.
F3: Existing sequence.local-bindings/dependent-key-conflict already observes sequence-to-singleton changed-input conflict. Novel evidence is retained-success/failure replay across carriers and current-position projection; conflicts remain row-local controls, not a novelty claim.
F4: Correct exact source citations: carrier equivalence2109–2111, retained equality/current positions3286–3288, retained postimage3304–3305. Reauthorization3307–3311 is excluded.
F5: Recompute both exactConfigurationDigest pins from actual expanded fixed routes; captured-ID GETs add no fixed route.
F6: A/B labels are presentation labels only. Canonical insertion follows required catalog order: sequence.idempotent-retry precedes idempotency.cross-carrier.
F7: Measure unchanged43-row baseline before authoring. Preserve cooperative90s/outer120s; if approached, report rather than thin observations or widen budgets.

Implementation authorized by root from8e45ef4d2c9674296f217ac0cf75bc706f49ff09; both rows remain partial. No runtime/executor/H1/M1 changes.

Implementation correction: names apply to creating operations, not updateBead. Row B adds a name to an initially unnamed successful create and removes the other creator name; the retained missing-subject update remains unnamed. Root accepted this grammar-preserving correction. Baseline43 installed rows passed11.84s test/12.32s total with current budgets.
