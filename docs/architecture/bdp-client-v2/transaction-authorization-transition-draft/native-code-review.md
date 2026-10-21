# Native authorization code review — original frozen draft

Two mandatory findings; no other concrete blocker established in the fifteen-file review. Root has accepted both for a narrow successor. This report reviews the original frozen bytes, not that future repair.

## N1 — Medium: reject non-string durable authorization views

`packages/server/src/transaction-authorization.ts:49–67`, especially 58–61, casts parsed JSON to the typed snapshot and passes `authorizationView` directly to `RegExp.test`. A generation-1 record with `authorizationView: ["00000000-0000-4000-8000-000000000000"]`, `readPolicy: "empty"`, and `canMutate: false` satisfies the regex through JavaScript array-to-string coercion. The outer field set is correct, so it is accepted. `transaction-store.ts:611–618` installs that result during reopen.

This violates the accepted strict, closed durable-policy contract: context/view fields are no longer strings and the returned shallow-frozen object contains a mutable array. View equality, cursor identity and response serialization can then consume malformed policy rather than refusing startup. This is a storage-corruption admission defect, not evidence of a remotely writable policy endpoint.

Required correction: explicitly require a string before the UUID expression. Add a rejected-open durable-copy specimen containing the singleton-array UUID to the existing migration/reopen callback. It should fail before any usable owner is returned and preserve existing copy cleanup. No parser permissiveness or policy semantics should change. The existing wrong-shape/missing-policy specimens do not cover this coercion.

## N2 — Medium: make the issued receipt-page withholding oracle sensitive

`packages/server/src/transaction-authorization.test.ts:364–405` creates `[create a, create b, delete b]` with a two-item receipt prefix. The previously issued continuation therefore contains only the non-record deletion result. Its post-publication assertion at 392–393 correctly expects that deletion to remain unchanged, but never exercises a Resource-bearing page entry. Removing the new page-entry projection at `transaction-store.ts:2259–2263` would leave this test passing. Root withholding and inline transport receipts do not replace this distinct, required issued-page disclosure check.

Required correction within the same callback and request/owner budget: reorder the real transaction to `[create a, delete a, create b]`. Assert the actual issued page contains the full third Resource before publication and the exact same page boundary/identity with its entry reduced to `{operationIndex: 2, outcome: "withheld"}` after publication. Keep root mixed-record/non-record equality, current context, failed-receipt equality and expired allocation assertions (indices 0 and 2) intact. Do not fabricate a page or weaken the strict full-object oracles.

## Scope and non-findings

Read the complete fifteen-file change, both new authorization files, the new runtime, service and queued-SSE callbacks, and the relevant unchanged owner/read/stream/recorder boundaries. Reviewed against the accepted Scope-wide successor plan `8f74cee8992d8d5b7c0c46cd640fe30971acb7022ebf19611dece0565cd2a521`, administrative lifetime correction `0b1e7bf427faaaad163bdda1af3bc0cda8e2b839e294e34c2c1751295d4be111`, transport preparation/report and frozen draft report/measurements.

The source trace supports exact format-10 migration with format-11 policy storage; unchanged retained graph/receipt/group/Snapshot facts; immutable per-entry policy capture; mutation pause/drain with a finite administrative timeout and cancellation that preserves admitted writes; direct pending-claim/sequence refusal; rollback versus uncertain/post-commit fencing; current root/page/sequence receipt projection; empty historical Changes advances and Event projection; Snapshot storage integrity separated from request-view fencing; and serialized expected-context lease acquisition before headers. Runtime passes the selected context into service stream acquisition, and publication aborts registered leases before leaving the owner entry. The queued-frame witness explicitly observes only the already accepted first frame and rejects further writes. No new public administration route or profile claim is introduced.

Normative cross-checks used `docs/specs/bdp.md` at the frozen base: 805–829 (authorization projection), 3709–3718 (opaque changing view), 5230–5275 (current receipt delivery, withheld entries and historical execution facts), and 5751–5758 (invisible Change projection). No unresolved product decision is inferred or resolved by this review.

The transport cleanup uses bounded waits and retains directories when runtime closure is unknown; recorder `closed` is a resolving close notification, so the attached observation continuation does not create an unhandled rejection path. The required administrative timeout is the accepted correction, not an invented scheduler delivery deadline. No additional production defect was established in these paths.

## Independence and evidence limits

I did not author this authorization implementation. I had prior read-only plan/source audit involvement and broader workstream plan authorship; root remains the independent integrator and disposition authority. This was a source review only: shell/Git reads, hashes, JSON inspection and report writing. I ran no product import, compiler, test, service, external reviewer or package command. Author/root static results are separate evidence; this report does not claim runtime success or council/merge approval.

## Exact attribution

Original binding SHA-256: `2928d83d55f2496bb3687ac0c3e87b1a536bea45fa511cd0a7ff4f2aceb5bcac`.

Original full patch SHA-256: `1ec29342e379efaa70740d96a3d34bbbef65e19e30e65ec962ccf258af9c9ac0`.

Base/HEAD: `67a2842e3f1918e70978f745587cf11eaf0df5a8`. All fifteen original source hashes were checked before inspection. All fifteen preserved frozen source snapshots were checked again at report creation. Live-source equality at that final check: **false**. A separately authorized successor must receive its own delta closure; it is not covered by this original report.

| Source path | Original SHA-256 |
|---|---|
| `packages/server/src/transaction-authorization.test.ts` | `2e689cac25b5174ce0e63a47d3060e544c3a188f62ad849b15332558219e86bd` |
| `packages/server/src/transaction-authorization.ts` | `fcaa399cde76b9579799a378c7a2272a749688945c8542fea24ed0a824e4c169` |
| `packages/server/src/transaction-changefeed.test.ts` | `ae4b6a0d4ae6e4e74e032074c850e5ced2de14d565078c2de88e2e62fa5e9c81` |
| `packages/server/src/transaction-changefeed.ts` | `1e678ecb1b2c3e59023dbb6a2e6516763907babd79f7926ca572a4764e40fc1e` |
| `packages/server/src/transaction-event-sse.test.ts` | `de14cbbbe935205c5ec43ef6cf6a055d3f1ca2b9632ee684c4c13fbfe9eed042` |
| `packages/server/src/transaction-events.ts` | `6c5ab4482f0b00110a71a96b16d44ca5fe713cb5f397ff78778cc6646f632a3a` |
| `packages/server/src/transaction-http.ts` | `c6bac496033781306d1c63e570e1c65b5786d41ebb8dc565c5a8be44e91ccfa9` |
| `packages/server/src/transaction-read.ts` | `66e71ff74a8139b12f0c7581c30a6798cc8724bb488cac8517119b9678ace212` |
| `packages/server/src/transaction-runtime-configuration.ts` | `57cd5d38c7199011d05589bc2b8ac1a0e2a4a9dc7585d019f9de5801a0a4d808` |
| `packages/server/src/transaction-runtime.test.ts` | `5c0dafced7ebfba91a9529fd3fb9b958701f0fa155c4901313667a23a042de47` |
| `packages/server/src/transaction-runtime.ts` | `48cb5411407c456d7b3f301a4b67970061be9f51c94201de832ba5deff6ea2d4` |
| `packages/server/src/transaction-service.test.ts` | `1cd7e1ccddbec75ca1c0d94d036a358fca98405e749000d28a0384cd9bfd4922` |
| `packages/server/src/transaction-service.ts` | `8e2a00fba538fc38fd81241301d38b232190f7241fa695e1dcdaf10641e97dc8` |
| `packages/server/src/transaction-snapshot.ts` | `d74aad937b0a07ea2878ef2c275cec6dbdb033cf112006e2569d64537520b15c` |
| `packages/server/src/transaction-store.ts` | `bc10df9d02316e395b8c4ea7db6134ab2646c8e8abd83ea785b0ea09f2f0c82c` |
