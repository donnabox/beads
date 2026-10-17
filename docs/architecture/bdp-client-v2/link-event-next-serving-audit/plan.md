# Next serving recommendation after the frozen Link draft

Supporting read-only audit under the sole design.md; **not a new implementation authority or source release**. Root is validating the unpublished Link candidate. Recommend one next bounded slice: **fresh Snapshot creation/preflight GET/HEAD with a lexically valid unrecognized minimum position and the actual current epoch/view**.

## Concrete gap and useful outcome

The existing packed-runtime test already demonstrates the gap: `apps/bdptest/src/transactional-snapshot-installed.test.ts:796` (“observes fresh Snapshot minima with exact current context and retained gaps over packed bytes”) sends a fresh no-query GET using actual context plus `minimumPosition: "unknown"` in exchange 19 (`:894–908`) and expects private 500 `minimum-position`. The callback has 31 real exchanges (`:953`), including seed writes; these are historical/current oracles, not evidence for a repaired implementation.

This is not missing coverage alone. Three production seams deliberately reject this input:

- `packages/server/src/transaction-http.ts:421–427`: fresh creation uses strict `transactionMinimumHeaders`; a complete lexical but unrecognized position yields the private refusal. Retained manifest/page requests with controls are independently refused here.
- `packages/server/src/transaction-service.ts:422–446`: `snapshot()` freezes through strict `snapshotTransactionMinimum`, then uses the existing finite minimum waiter for create/preflight only.
- `packages/server/src/transaction-store.ts:1558–1571`: `observeSnapshot()` strictly snapshots the tuple and calls `checkTransactionMinimum` before expiry maintenance, capacity/integrity preflight, capture, record copy or handle/page allocation.

The bounded outcome is contextual 503 catch-up-timeout after the existing finite wait when eligibility remains unproved, with HEAD parity and no stale Snapshot success, fabricated expiry, or usable handle. This removes an actual replica-bootstrap serving refusal. Installed proof must still use the existing private configured target convention: it neither invents discovery nor qualifies a cold public Transactional client.

## Basis, and why this does not decide S1/G2

Pinned spec `docs/specs/bdp.md:5841–5843` permits the remaining-read authority to wait/route or return the normative consistency failure rather than report an older success. `:5901–5911` defines **fresh** safe-GET Snapshot creation and HEAD metadata/decision without a usable handle; subsequent GET is independent. `:5876–5878` keeps minimum enforcement ahead of conditional shortcuts.

The Resource/properties-specific unknown-token MUST at `:5844–5849` must **not** be described as applying to Snapshots. Its explicit caveat at `:5850–5851` does not settle Snapshot/retained-continuation precedence. This proposal is the narrow reference implementation's bounded-wait choice under the general remaining-read law, for a fresh current observation and matched context only. The already-reviewed fresh Snapshot plan (`transaction-snapshot-creation-minimum-plan-reviewed-v1.md`, Basis/Smallest implementation) explicitly separates fresh current capture from retained-anchor ambiguity, and current production already checks known minima before fresh work. Do not claim a new universal precedence rule for full capacity, integrity failure, or foreign/unknown combinations.

S1 remains retained anchor versus serving-observation interpretation. No retained manifest/page minimum behavior changes. G2 concerns incident-Link metadata Event membership and is unrelated. Unknown opaque plus foreign epoch/view remains unsupported; known-position epoch/view handling remains unchanged. No operator question is needed for this scoped implementation choice, subject to native/Claude plan review and root acceptance.

## Smallest proposed scope and required discrimination

Production can remain the same three Snapshot entry files. Reuse the existing lexical capture/current-check helpers and finite waiter; no token codec, storage format, public export, schema, spec, or `transaction-consistency.ts` redesign is proposed. `checkCurrentReadMinimum` (`transaction-consistency.ts:115–132`) already distinguishes unproved same-context eligibility from unsupported unknown-plus-foreign, and the waiter already catches unproved eligibility without changing its original deadline.

A final implementation plan should pin these details before authoring:

1. Select only the existing canonical no-query creation route and owner/service create/preflight operations. Forward one immutable lexical tuple through both direct-capture and terminal-preflight HTTP branches. Preserve the pure HTTP candidate and **one service call**; no second capture or renewed wait after HEAD/304/412/406 preflight.
2. Check the actual current owner context before `snapshots.expire`, preflight, capture/copy/allocation, on every retry. Direct owner calls reclassify their supplied tuple rather than trusting service provenance. Keep no-minimum, retained, known-position and foreign-context behavior intact.
3. Preserve Snapshot's **existing** entry/error order. Unlike Event's first-attempt capture, current `service.snapshot()` synchronously captures controls and checks kind before `run()`/waiter preflight. Do not transplant the Event capture arrangement and silently reorder malformed tuple, invalid kind, service lifecycle, clock/policy or abort errors. Freeze request kind and tuple once; aborted admitted waits still prevent later owner entries, and deferred waits must be released before awaiting rejection/close.
4. Use a real-owner call and offline Snapshot table observations to prove unknown-minimum attempts neither allocate nor expire retained state. Do not use a history-scanning helper as supposed proof of no scan. Keep capacity/integrity checks on eligible preflight genuine.

Natural test ownership is at most **seven files**: the three production files plus existing `transaction-snapshot.test.ts`, `transaction-service.test.ts`, `transaction-http.test.ts`, and `apps/bdptest/src/transactional-snapshot-installed.test.ts`. Reuse their fixtures; do not force this into the preceding Event slice's six-file count or duplicate its helpers.

Existing tests to preserve/amend narrowly:

- Snapshot owner `:1218`, “checks each fresh Snapshot minimum before any handle, record, page or expiry work”: two obsolete unknown-token TypeError assertions (create/preflight, `:1238–1239`) need explicit successor disposition. Its known/foreign/retained controls remain. “does not expire retained Snapshot details on an ineligible fresh attempt” is the natural maintenance-order oracle.
- Service `:1093`, fresh Snapshot minimum service: create/preflight mutation-after-call capture, real independent commit, original deadline, pre-abort and close/drain controls. Add deterministic unproved retries and cancellation/drain discrimination without replacing known-position tests or claiming unknown becomes recognized after a write.
- HTTP `:2805`, “serves complete fresh Snapshot minima with exact timeout, foreign and conditional context”: retain existing known behavior; add opaque matched-context GET/HEAD and shortcut-discriminating conditional/media requests. All use actual observed context; lexical receiving failures and retained refusals stay distinct.
- Installed 31-exchange callback above: explicitly supersede only the old fresh unknown-token 500 oracle in new evidence, preserving prior evidence attribution. A separate small installed GET/HEAD witness should use real seed/context acquisition and actual packed runtime; prove contextual catch-up-timeout/retry/cache/body semantics and retained/private controls without public-admission bypass. Enumerate all setup requests, allocations, children and stores in the later plan rather than guessing an inherited Event workload.

No existing timeout is widened. Build emitted runtime **before** collecting source HTTP suites that import it. Static/source checks and immediate audits precede fresh packing/external-consumer provenance and installed observations; commit only after all accepted source/installed audits. Native + actual Claude review and exact-head green CI remain integration requirements. This audit authorizes none of that execution.

## Alternatives and dependencies

The frozen Link draft closes the identified initial finite Scope/Bead/Link Event unknown-minimum family; do not duplicate it. Scope/descriptor/operation discovery is still explicitly gated by the sole plan's complete same-target functionality and evidence predicates, not fixed by advertising partial capability. Unknown-plus-foreign context and retained Snapshot/continuation minima remain policy-sensitive; incident-Link metadata remains G2. Ordinary Update has no demonstrated regression requiring a rewrite, and completed-key-conflict bookkeeping is lower priority than this witnessed serving refusal.

Root should first finish the current Link validation/publication decision, then reanchor this recommendation to its actual accepted bytes/ordinary descendant. No dependency on newer main, no main merge, and no changes to the frozen PR150 checkpoint, Janet, CLI/staging/playgrounds/shared databases, Trish Types, optional History or provenance proposals. A precise bounded plan/council review is the next step, not immediate authoring.

## Read checkpoint and preservation

Remote role session was checked by the guarded report write: `2026-10-08T20:26:35Z-81e5`. Sole design/live worker ledger were read. The ledger still described the earlier Link code-review/static checkpoint; root's latest assignment reports native + actual Claude CODE without blockers and 33 source callbacks passed, with installed validation in progress. This audit did not independently run validation or inspect/poll processes and makes no publication/completion claim.

- Actual worktree: `/Users/dbox/repos/wt/vickie/vickie-link-event-minimum`; branch `codex/vickie-link-event-minimum`; HEAD `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385` with the frozen six-file Link draft, **not a clean published Link commit**.
- Author binding SHA256 `14f6ceef6491d4b07b60555e1bd91886bf2f9f10bfe679f57abd23e0a93ab8a1`; full patch `7c8d17d90eb8832bfb191ada8deb82f548ce5487cf6c07fcd6c66c3ff3b5aa3b`. All 679 current tracked input hashes match the frozen after-map and exact dirty six-file scope.
- Pinned monolithic spec SHA256 `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8`.
- Sole plan read-checkpoint SHA256 `cfc3edfd471ac400c5a4ea1ec61621bff1e58ec363693d120c69d5695c47280e` (mutable execution ledger; not a future review binding).
- Historical fresh Snapshot reviewed-plan SHA256 `1f24e59f8fd1e9df2d2ec275353f1e730db27e852199a886d168c244da30ec7b`; historical consistency-audit SHA256 `aa25a080d1cba17a9e757a5476ea4283d36573c0c6305251bbfba99bc497fc42`. Historical line references were reconciled to the current monolithic spec above.
- Only this supporting report was written. No source, old evidence, review packet, live plan, coordination, build, test, dependency, service or process state was changed.
