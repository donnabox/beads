# C3 v2 native plan closure

## Finding and recommendation

1. **Low — correct the candidate key count before implementation.** `transaction-cold-discovery-next-plan-v2.md:76` says “17 required discovery members plus limits.” The actual `schemas/bdp-v0.schema.json:4986–5074` requires **16** members; adding the candidate's optional `limits` produces **17 total**. Correct the prose and derive/pin the candidate exact-key oracle accordingly. Do not add a fictitious eighteenth member or alter the schema. This is the same inventory error root identified, independently confirmed here.

**No remaining material plan blocker found.** Recommend accepting the bounded C3 scope subject to that factual correction and the existing final-base dispatch gate. This is plan closure, not implementation approval, serving readiness, or profile qualification. Root has separately approved the two parser/two type root exports and private service Scope comparison; those are no longer unresolved scope requests.

## Exact reviewed inputs

- Plan: `transaction-cold-discovery-next-plan-v2.md`, SHA256 `bb90ce032603cd3ece6b13728e75d041bc62eb4ad2b51720c1a8253b406f7886`.
- Dispositions: `transaction-cold-discovery-next-plan-dispositions.md`, SHA256 `0fd535fc61d839eb979a46720ad1ead93bdca5a554cef8ebbac8a821b4c7f8f7`.
- Source: `/Users/dbox/repos/wt/vickie/vickie-tx-collection-minimum`, clean HEAD `20d61119b5faf7251a5d285c5c34a1049866ffe5`, tree `3de0c4b5589423d01eb0dcb7df7ad150d184273d`.
- Prior native/external plan findings and root-accepted `transaction-discovery-replay-floor-audit.md` inform closure. Original reviewed plans/reports remain unchanged. Later Snapshot/alias integration is not represented as present in this source; implementation must re-anchor on root's dispatched accepted base.

## Prior-finding closure

| Prior issue | Source-grounded closure |
|---|---|
| F1 floor versus checkpoint | Corrected: owner floor is position `pos-0`; the separately derived exclusive replay checkpoint is `genesis`. Empty head/first previousPosition test the former, actual feed/reopen/expiry tests the latter. No codec, schema or fixture rewrite is proposed. |
| F2 runtime parsing authority | Two production parsers use exact existing schema roots. `read-update-values.ts:128–178` provides the lazy AJV, canonical bundle and guarded snapshot precedent; `read-values.ts:536` exposes the internal snapshot helper without requiring a new generic public validator. Two types/two root exports are a feasible, explicitly approved addition. Server production calls both parsers; optional schema-valid members remain valid generically. |
| F3 nine route roots | All nine are now named, including singular `alias/` and slashless `snapshot`. Prefix routes append a real subject or valid query. The shared operation map is checked against actual targets and an independent literal directory, rather than its own length. No bare-prefix success fiction remains. |
| F4 limits/configuration | Memberwise wire projection excludes cursor TTL. The same captured scalar values feed enforcement and assembly; existing `authority-read.ts:350–400` copies and compares page, Selector and internal TTL values against enforcement. Mutation tests exercise the intended immutable binding. Capturing the configuration before owner entry avoids reentrancy. |
| F5 sequencing | C1's former unmerged dependency is explicitly superseded; later accepted integration remains a named implementation prerequisite. Packed C3 no longer repeats a whole C1 walk. |
| F6 and native schema/owner issue | Structural acceptance and candidate identity/configuration policy are explicitly separate. Well-formed foreign Scope and optional History are not misrepresented as schema failures. The impossible “no lower-profile members” predicate is removed. |
| F7 emitted versus packed seam | The plan uses actual `transaction-installed-contracts.test.ts` for emitted/in-process checks and a separately labelled packed app cohort. Configuration-only tests are not recast as transport evidence. |
| F8 budget | Twelve enumerated requests sum correctly: 3 credentialed GET refusals + 3 HEAD refusals + 3 anonymous refusals + create + independent GET + exact replay. One child/one Bead/no reopen/wait; ceiling16/1, 5s wire, 45s aggregate, 60s outer and explicit10s lifecycle. Direct/emitted and zero-wire smoke checks are separate. The proposed request budget is feasible; timing remains to be measured during implementation. |
| F9 Scope authority | Approved private service.scope from owner.scope plus canonical handler equality is a narrow feasible assertion at `transaction-http.ts:166–190`, before target construction. It closes the double-authority seam without a runtime/config API. |
| F10 listener authority | Port exclusion is confined to the explicit distinct-canonical-authority fixture. Listener-derived mode is correctly allowed to include its port. |
| F11 optional policy | Candidate omission of order/multiplicity/history remains separate from parser law. Correct exact key count as the Low above. |
| F12 ownership/package surface | HTTP tests and smoke inventory are now owned. Existing `smoke-executables.mjs:129` pins emitted files; its transaction-module deep-import denial loop at 472 onward can cover the new private module, with four emitted entries and installed protocol-root parser checks. No export-map expansion is required. |
| F13 owner lifecycle | Configuration is obtained before entry; one synchronous withRead supplies epoch/view/head. The plan preserves entry-refused cause, close/fence checks and prevents reader/context escape. No nested readConfiguration is required. |
| F14 stale anchors | Named symbols and exact inspected base are explicit; final source reconciliation is mandatory before authoring. |

## Boundary and validation observations

The current handler still explicitly refuses Scope, `bdp.json` and `operations/` (`transaction-http.ts:546–553`). Internal descriptor assembly does not remove that gate. Schema parsing exports prove value validity, not target qualification; package smoke imports are zero-wire and cannot stand in for cold discovery. Existing source-derived and installed admission checks remain intact, and any actual Read evidence binding effect must be reported rather than silently resealed.

The source/emitted route suite can exercise receipt refetch, Snapshot branch recognition and classifier agreement independently of the twelve-request packed ledger. It must remain labelled that way. The configuration copy must preserve existing callbacks/TTL enforcement while copying the advertised scalar subtree; the plan does not authorize a broad clone of runtime settings. These are implementation checks already expressed by v2, not new scope.

## Actual review activity and limits

Read-only source, schema, plan, disposition and packaging-script inspection; verified HEAD/tree, clean status and both hashes. No tests, builds, servers, network, subprocess fixtures or source changes were performed. A read command initially used nonexistent `docs/schemas/bdp-v0.schema.json`; the actual canonical `schemas/bdp-v0.schema.json` was then inspected directly. Only this review report was written. No claim is made about future test results or packed C3 behavior.
