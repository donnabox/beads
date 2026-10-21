# Next serving audit after sequence-member planning

**Recommend a narrow serving repair before another evidence-only increment: preserve observed Scope context on unknown Resource Event-view404 responses.** This is a source-established mismatch, not an executed reproduction or implementation release. The ten-action sequence evidence plan remains independently reviewed work; this recommendation does not alter it.

## Basis and authority

Read-only clean actual merged main `88c9408e908590ab1a4c64c9b9dee1a8b705999c`, tree `b2d9f2c18502e0778c51123be8b4564af052d803`, at `/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence`. Read sole design Autonomous profile completion (design.md:7–31): prioritize actual cold-client/discovery, consistency and Event/view behavior; do not advertise an incomplete profile or invent semantics. Inspected relevant current store/service/HTTP/Event/consistency/read paths, existing Event tests, prior after-cross-carrier audit, discovery-serving v2, and current normative clauses. No tests/build/setup/install/service, product import, Claude invocation, source edit or remote publication was performed. Only this report is written through the session guard.

The earlier PR102–126 work is present in this merged source: command minimum fields are ignored on mutation POSTs; Selector preparation/admission ordering is fixed; sequence alias turn resolution, current Resource/Type/incident/alias minima, fresh Snapshot minima, schema delivery, saved Last-Event-ID replay, refusal/deadline context and maintenance fault propagation are not proposed again. PR126's classifier/credential-helper changes address conformance input admission, not the Event404 projection below.

## Concrete trigger and mismatch

On an authorized fixed-view Transactional service, use a canonical unknown Bead or Link identity, e.g. `GET /acme/beads/never?view=events`, with no minimum fields. The identity has never been committed; no erasure/history interpretation is needed. Use finite JSON or selected SSE; HEAD follows the same initial routing.

| Boundary | Actual code behavior | Existing required behavior |
|---|---|---|
| Owner absence observation | `transaction-store.ts:621–627` queries the current identity table. `readEvents:1563–1567` runs this inside `enter`, then returns frozen `{kind:"missing"}` before copying context. `beginEventScan:1647–1651` does the same. | Return the same absence, with the Scope epoch/view/position observed in that same owner entry. No second observation or receipt-derived context. |
| HTTP finite/SSE projection | `transaction-http.ts:497` and `:513` convert missing to `transactionHttpProblem("resource-not-found")` without the optional context argument. Service `events:390–402` and `beginEvents:404–423` forward the owner result. | Preserve404 resource-not-found while attaching the three actual context fields. HEAD remains bodyless; SSE negotiation must not start a stream on404. |
| Normative authority | `docs/specs/bdp.md:5661–5668` requires unknown/out-of-projection Event identities to return the same404. `:5800–5804` specifies the three fields on Transactional Scope-bounded responses; `:5830–5835` says every Scope-bounded read reports the projected-prefix context it observed. | Nondisclosure requires the same404, not loss of already-observed Scope context. The context contains Scope/view/head facts, not whether an undisclosed Resource exists. |

This has an actual owner observation; it is distinct from a rejected route, unauthenticated request or malformed pre-owner input. Existing general Read projection already carries context when a problem includes an observation (`transaction-http.ts:214–238`), providing the analogous established boundary. Existing tests explicitly accept missing-only owner shapes at `transaction-events.test.ts:221`, `:283–285`, `:893–895`; the HTTP unknown case at `:459–463` checks404 but not these headers. That explains the current test gap without treating prior green tests as proof of the required response.

No new response body, status/retry policy, public discovery/profile claim or administration API is necessary. The primary proof uses no minimum fields. Existing unknown-source-before-minimum/reconnect precedence stays intact; adding observed Scope context does not authorize changing that precedence or resolving behind/foreign visibility questions.

## Smallest bounded repair candidate for council

Likely five paths, with root exclusively arbitrating shared ownership:

1. `packages/server/src/transaction-store.ts`: extend only the two missing Event result variants with immutable `TransactionReadContext`; obtain it inside the already-entered absence branch. Preserve `eventSourceKnown` ordering, existing minimum checks, cursor errors and integrity propagation. Do not read context in HTTP after the owner call.
2. `packages/server/src/transaction-http.ts`: pass the returned context to the existing problem projector in the two missing Event branches. No common error catch rewrite or routing change.
3. `packages/server/src/transaction-events.test.ts`: reconcile the three exact missing-only expectations and extend focused owner/HTTP oracles. The existing file already contains real-owner and emitted-runtime Event fixtures. Preserve identity-before-minimum/reconnect assertions rather than deleting them to accept the new shape.
4. One narrowly named installed Event-context test under `apps/bdptest/src/`, using existing packed entry/configuration/credential/transport helpers. Do not expand the broad transactional matrix or change its current evidence rows/controls.
5. `docs/design/transaction-implementation-progress.md`: distinguish the completed observed404 path from pending Snapshot minima and metadata Event membership; preserve historical evidence citations.

Service result inference should carry the new context without a service source edit; this must be confirmed in the measured draft. If another typed caller needs a change, report the actual dependency before widening ownership. No schema/catalog/fixture/config/export or store-format modification is expected. The active sequence evidence lane owns its four artifact/app paths; this proposed store/HTTP/Event test lane is disjoint except root-owned progress documentation and final ancestry integration.

## Minimal prospective validation

Before execution, author exact callback expansion, resource lifetimes and command selectors in a measured draft; this report does not grant a test budget. Prefer extending the existing owner missing cases and a compact HTTP table rather than running the full Event file. Required source oracles: no-minimum finite and initial SSE absence from both Bead/Link sources carries the exact same-entry immutable context; current head advances after an independent commit;404/body/retry remains unchanged; HEAD metadata matches/body empty; no stream begins; anonymous pre-owner401 has no context. Existing known Event response, identity/minimum/reconnect precedence, close/fenced owner and fault propagation remain targeted regressions. Build before emitted-runtime tests.

A small fresh installed witness can be **11 requests/one child**: one genuine create seed atpos1, one known Scope Event read as context/normal-path comparator, eight unknown Bead/Link reads spanning JSON/SSE × GET/HEAD, and one anonymous unknown-view request. No minimum tuple, polling, redirects, discovery activation, hidden Type fetch or additional transaction is needed. Derive epoch/view from actual seed/known response and pinpos1 only as this fixture's expected position; all eight unknown responses remain404 and context-complete. One live child/one store, ordinary owned cleanup, retain actual response fingerprints/source/package/config bindings and independent PID/temp absence. Exact unchanged deadlines and payload bytes must be measured and reviewed before release.

Meaningful discriminator: in a disposable emitted/source derivative restore the two old HTTP missing projections while retaining the owner context. Select the smallest JSON/SSE header cases; require actual404 and otherwise-valid body, but the named context assertion must fail. This demonstrates context propagation, not genericexit1. A separate owner assertion on the new missing result prevents an HTTP-only reconstruction from satisfying the design. No installed fault seam or extra installed mutant is needed unless council identifies a real evidence gap. Preserve original failures and no automatic retries.

## Other inspected blockers and exclusions

- Cold discovery remains an actual product-delivery blocker: canonical Scope/descriptor/Operation Directory are deliberately refused at HTTP:195/228/560–567 despite internal assembly and observed-minimum support. Reuse the existing discovery-serving plan when the root's truthful-advertisement gate is satisfied. Wiring it now would claim an incomplete cumulative profile; it is not a safe isolated shortcut.
- Metadata-only incident-Link Event membership is still explicitly unresolved (`transaction-events.ts:257–264`). No membership rule is inferred.
- Retained Snapshot/continuation minimum comparison remains unresolved (`transaction-http.ts:406–411`). Fresh Snapshot support is already implemented; do not repeat it or extend its comparison semantics here.
- Private unsupported tuple forms and source-only bare-signal receiving limitations were inspected but no additional concrete cold-wire defect was established in this bounded pass. HTTP uses captured tuple objects and actual Request signals; do not turn speculative API hardening into the next serving claim.
- Dynamic authorization-view/erasure/restore work needs a configuration-specific trigger and established law; no new policy-edit API or optional History/Types expansion is proposed.

This audit is deliberately bounded. It establishes one concrete missing-context path from source and existing normative requirements, not a universal serving audit or reproduced runtime result. Root/native/actual Claude plan review and measured authoring/execution release are still required.
