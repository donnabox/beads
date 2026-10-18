# Update guarded-request overlap: bounded feasibility

**Feasible as a narrow private installed observation, with an explicit limit: existing helpers prove overlapping client invocations, not simultaneous server admission or two requests physically on the wire before either completes.** No new concurrent conformance DSL or production scheduler is needed for the former. If root requires the stronger transport-overlap claim, a small test-local HTTP observer exposing socket/finish milestones is necessary; current helper results cannot establish it retrospectively. This note proposes no implementation authorization.

Basis: clean candidate `3df82ac480f50c8f486b4e456336366e7acbdf87` in `/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence`; accepted Update audit `62f11d551e70563b95343a0f109dd576bbf59fa78687beef97d47ecc44f6387e`. Receipt plan remains untouched. No tests, imports, services, builds, installs or source edits were performed.

## Existing seams and exact overlap limitation

`scripts/installed-reference-observation.mjs:104–143` exports `requestInstalled`: each call synchronously constructs a separate Node HTTP request and calls `incoming.end(body)`, returning a Promise settled by response end/error. It contains no shared await/queue/lock. Therefore a caller can invoke A, immediately invoke B in the same JavaScript turn, record both invocation IDs before any response callback can execute, then await `Promise.allSettled([A,B])`. Attach both fulfillment and rejection observers immediately and record ordered invocation/settlement events. Both promises are client-outstanding before waiting; avoid claiming observable Promise internal state. `allSettled` is preferable to early `Promise.all` rejection so cleanup drains both requests.

The helper does not expose ClientRequest, socket connection or finish events, and uses Node's default Agent. Starting two promises does not prove two connected sockets or two fully transmitted requests at a given instant; Node/OS/server scheduling may serialize them. No test can require simultaneous synchronous owner evaluation: this owner deliberately serializes work. A future test-local observer could use two distinct connections, establish both before sending bodies, and record finish milestones; even that would prove transport overlap, not overlapping durable admissions. Do not conflate those claims or add a server-side barrier merely to guarantee them.

The current complete observer at `:145` awaits requests sequentially and performs a much broader navigation/reopen scenario; do not run or extend all of it just for this witness. Reuse its helpers in a separately selected test. The cumulative executor at `packages/conformance/src/runner.ts:637` also advances its scenario actions sequentially; two manifest actions are not concurrency. Keep this first witness outside the manifest unless a separately reviewed existing observation seam can carry it honestly.

`launchInstalledReference` at `:17–85` accepts the actual child entry, checks ready/configuration/artifact/payload/policy/authentication identities, and offers bounded close/exit. The existing test uses emitted `packages/server/dist/installed-reference-process.js` (`apps/bdptest/src/installed-reference.test.ts:80–84`). A fresh-pack witness must instead pass the external installed consumer's actual entry; merely retaining the word installed in a test title is insufficient. `read-update-matrix.test.ts:69–70` demonstrates an existing entry override, but this note does not propose running that matrix or borrowing its sequential action mechanism.

## Minimal meaningful positive: ten HTTP requests, one child/owner/store

1–3: GET canonical Scope probe (204), its actual service-description Link (200), then the discovered operations directory (200). Parse discovery/directory, use returned createBead/updateBead target references. No Type inventory traversal is claimed; Type is the existing explicitly provisioned local task fixture.

4: POST createBead, unique seed key, properties `{n:0}` (200). Validate full result/Resource and retain its actual revision R.

5–6: Two immediately initiated updateBead POSTs, distinct keys A/B, the same subject and `expectedRevision:R`, replacing `/n` with respectively1 and2. Preserve request-to-result association rather than assuming which wins. Require exact status multiset `[200,409]` at a named assertion. Success must be a complete `outcome:updated` postimage with the chosen n, unchanged identity/type/metadata and revision different from R. Failure must be the ordinary direct Problem: conflict type, status409, code`revision-mismatch`, retry`after-state-change`, no fabricated TX receipt/context or sequence operationIndex. Validate against the actual RU schema and assert required concrete fields; inspect any optional diagnostic fields before deciding exact whole-object shape in a formal plan.

7: GET subject (200), full equality to the winner.

8–9: Replay A and B separately with their original bodies/keys. Require original respective200/409 and full parsed body equality to each retained original disposition (including opaque revisions), not regenerated expected values.

10: GET subject (200), still full winner equality. Mutation responses require private/no-store and no ETag/Location, and must not acquire TX tuple/receipt fields. Source law: `docs/specs/bdp.md:2098–2114,3498–3499,6775`.

This distinguishes execution, conflict, durable key-specific replay and unchanged state. It does not prove arbitrary schedules, crash durability, Link/owned-Link races or whole-profile qualification. The evaluator checks the revision at `resource-evaluator.ts:1023–1027`; RU execution reaches it via `member-executor.ts:593` inside `recovery-store.ts:799–801` and its synchronous transaction wrapper at`:381–390`.

## Sensitive emitted negative, no production hook

Copy the complete built child dependency graph into an isolated derivative. Remove only the emitted equivalent of the evaluator's expectedRevision mismatch guard. Leave source, positive build, requests, keys, timeouts and evaluator behavior otherwise identical. Bind actual child entry/module resolution to the derivative (including workspace package links); patch-count/before-after hashes and an import inventory must show the child loads the mutated module. Do not rebuild mutated source or point a positive child at an unrelated file.

Run the same first six requests once. Both distinct valid updates should now complete200; retain both safe statuses/outcomes and fail the named `[200,409]` assertion. Stop before the remaining four positive reads/replays. This is a meaningful guard sensitivity control even if the server serializes requests; it proves the oracle rejects omission of revision enforcement, not that every race implementation bug is detected. Unexpected transport/schema/fault failure is not an accepted control result.

Proposed bound: positive10 + derivative6 =16 HTTP, two sequential children/owners/store directories, no restart/retry, no extra source HTTP fixture. A tiny local event-log/order assertion adds no wire traffic. This is prospective only; exact fixtures, body bytes, named selectors and fresh package identities require a formal measured plan.

## Cleanup and readiness constraints

Retain existing helper request5s and child-ready/close10s bounds; proposed per-case30s and external120s command ceiling, with no automatic retry. Both outstanding requests must settle or be explicitly destroyed by the test-local transport path before child close. Preserve primary and close/removal failures independently, emit safe outcome/PID/response-count records even on failure, and independently check child/temp absence. Existing `requestInstalled` buffers response text without a ceiling and exposes no abort handle; do not claim a32768-byte enforced cap or guaranteed local cancellation from it. A formal plan must either accept its existing bounded-fixture limitation explicitly or add only a test-local bounded response reader/cancellation path, with no production API change. This also favors keeping the witness separate from the large existing observer.

The experiment is worth a formal plan if root wants direct Update concurrency confidence after receipt repair. It should remain accurately labeled overlapping client calls unless transport milestones are added. There is no demonstrated Update runtime defect here and no need to delay the proven receipt-context serving repair for this proposal.
