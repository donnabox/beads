# C1 v2 native plan closure

## Result

**PASS — no remaining material plan defect found.** All seven external findings and both native findings are resolved at the planning boundary. Recommend authorizing the bounded C1 implementation as written. This does not verify implementation, executed tests, public admission, or whole-contract coverage.

Reviewed plan SHA256 `4252bf3e2f951c10ad6bfa351bf7d5212349ea18dc506b770728d949dc7b2dfa`; dispositions `transaction-cold-client-plan-dispositions.md` SHA256 `bfc74810721632383bd8a532d09f1f4d30e6600856b4558a314ff6ec3c77acf1`. Both hashes independently checked. Original plan/reviews preserved. Source remains `/Users/dbox/repos/wt/vickie/vickie-schema-property-change`, frozen HEAD `61ed530ac6a460598e0e312f5d28ec0e6fa2a33d`, tree `22e448fe0f517c0d6300b732de55494bd70c7938` matching main581.

## Closure dispositions

| Prior finding | Closure verified |
| --- | --- |
| External1 / Native1: installed deterministic wait feasibility | Direct owner/runtime proof and genuine installed observations are explicitly separated. No installed seam or metadata substitution. Fixed-head installed timeout is feasible with the real1000ms budget; four cases disclose the nominal4s floor. Concurrent installed success is not claimed as an observed behind attempt. |
| External2: materialized control refusal | Named existing `invalid-parameter`400 and `limit-exceeded`413 are preserved by returning from sourceCondition, not throwing a plain Problem or broadly catching errors. Concrete behind-minimum controls and no-wait/no-retention assertions are required. The external review's400 limit-exceeded claim is correctly rejected. See narrow reachability check below. |
| External3: exact route subset | One private shared predicate accepts only initial beads/links collections with no continuation. Types, incident-Link views, cursors and discovery remain excluded; existing point subset stays separate. Both call sites must use the helper. |
| External4: problem framing and discriminator | The defect is accurately described as explicit minimum-control refusal, not stale ordinary reads. Counterfactuals target timeout/incompatible minima instead of ordinary post-commit freshness. |
| External5 / Native2: abort forwarding | The already-owned HTTP path will forward request.signal to the existing facet/service/wait. Point and collection pre-abort and actual disconnect controls are required; synchronous work and the current bounded sleep are expressly not preempted. No new timing SLA, installed seam or lifetime framework. |
| External6: pinned negative guard | Exact existing negative list is deliberately split; only initial collection entries move positive. Cursor, alias, Scope/descriptor/directory cases stay, with Types and both existing/absent incident anchors added. No forced uniform response for unrelated invalid paths. |
| External7: collection revision/validator | Body record revisions are named explicitly; no collection ETag added. Wildcard304 and nonmatching If-Match412 retain existing semantics and required context after minimum checks. |

## Materialized refusal scope checked against source

`authority-read.ts:752–760` wraps only pagination-limit validation and Selector validation in `control`. `control` at519–526 catches only `ReadPaginationError`/`ReadSelectorError`; provider/store exceptions are not converted. `read-request.ts:954–968` maps Selector grammar to invalid-parameter, complexity limits and pagination invalid-limit to limit-exceeded. `read-pagination.ts:581–592` enforces the configured range.

Within the accepted initial Bead/Link collection subset, these are the named materialized control refusals that can reach the existing source-condition callback. Authorization capture refusals do not invoke that callback (`authority-read.ts:982–996`). Point absence remains resource-not-found and its existing explicit visibility gap. Continuation, Type and incident-anchor materialization remain unreachable through the new minimum predicate. Errors from later finish/pagination allocation happen after sourceCondition and are not broadened by returning from this callback.

Therefore the proposed callback early return can preserve the already-materialized control Problem without a shared AuthorityReadPlane edit or duplicate parsing. The plan appropriately requires a stop if implementation discovers wider reachability; do not implement this as a catch-all status class or arbitrary-error bypass. Minimum checks still run on every valid representation before finish/conditional projection.

## Abort-test feasibility checked against source

`transaction-http.ts:222–227` has the single missing forwarding site. The service already passes receiving.signal into the wait (`transaction-service.ts:236–245`), the wait checks abort before each attempt, and AuthorityRead checks an already-aborted signal before observe (`authority-read.ts:1050–1064`). A pre-aborted direct Request/handler test can therefore prove no owner attempt; using only an already-aborted client fetch would not exercise that server path and must not be substituted.

For actual Node disconnect, the direct runtime's existing `minimumReads.wait` seam can hold progress after an observed behind attempt while the client socket is closed. `node-http-disconnect.ts:14–24` exposes the existing server-side abort behavior. A test must synchronize on server-side abort observation before releasing its gate and then assert no subsequent attempt; client-side close or an elapsed short sleep alone is insufficient. That is feasible with test-local observation of existing signal/lifecycle behavior, without new installed IPC or product API. Existing HTTP tests already contain actual client-disconnect fixtures. The revised plan explicitly demands a forwarding-sensitive test and bounded cleanup, so this is implementation guidance within its acceptance criteria, not an outstanding plan blocker.

## Limits

Local read-only source inspection and hash checks only; no tests, builds, servers, network, agents, source changes or outbound actions. Only this report was written. One initial dispositions filename guess was absent; the actual file was found and its requested hash verified. All stated execution ceilings remain proposed bounded implementation gates, not measured results. No renewed feature audit or additional optional obligations were added.
