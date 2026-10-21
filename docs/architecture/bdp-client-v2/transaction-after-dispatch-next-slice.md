# Next bounded slice after same-key dispatch

Recommend a **private installed Transactional guarded-batch contest**, after the current same-key dispatch slice. No new serving defect was established by this bounded selection. This extends an existing settled behavior into actual HTTP evidence; it does not rewrite ordinary Update or activate discovery.

Basis: committed main `98c600767e8e91a79e109cb741beb91e23b9bebe`, tree `fde859381752d15c8caabcbc946d3c2e384b9cf6`, read with git show through the same-key checkout. No mutable author draft was used. Read the two requested prior audits first. Current transaction-http/store/resource-evaluator bytes match their f551 predecessors, so this selection does not repeat their broad serving audit or resurrect closed defects.

## Three related requirement IDs; one scenario

| Candidate | Existing evidence/source definition | Exact useful addition |
| --- | --- | --- |
| **transactional.batch.revision-guard-race** — primary | Required catalog row at transactional-v1.json264ff. No corresponding executable row in the current Transactional manifest or coverage file. Store test539–569 already admits two guarded requests, executes them serially, checks one transition and a retained failure/reopen. Ordinary RU's installed overlap is separate-profile evidence. | Real installed clients dispatch two distinct-key batches using the same actual observed revision, without awaiting the first response. Establish exactly one completed receipt and one failed revision-mismatch receipt, then the complete resulting graph/Event state. Record client dispatch/lifetimes, not physical admission overlap or an internal scheduling proof. |
| **transactional.batch.rollback** — companion | Existing manifest row has nine actions: seed, seed-facts, Events-before, a sequential three-operation failing batch, unchanged/absent/Events-after, fresh create and point read. Coverage is explicitly wholeContract:false and excludes concurrency. Store test671–694 checks rollback of earlier update/Link deletion after a guarded failure. | Put a distinct supplied-ID marker creation before each competing guarded update. Only the winner's marker may survive; the loser's earlier staged creation and induced Event must be absent from complete graph/Event observations. This combines the atomicity oracle with the actual guard contest instead of adding another independent rollback scenario. |
| **transactional.idempotency.failed-retained** — narrow companion | Current eleven-action row covers missing-target failure, later creation, equal retry retaining the failure, fresh-key success and changed-identity conflict. Its coverage remains wholeContract:false; retention expiry/forgetting is omitted. | Replay the actual losing batch with the same key and exact body, and GET its issued receipt. Require full retained failed-receipt equality and no new observable effect. This demonstrates the retained outcome of this guard contest; it does not prove the retention interval, expiry or forgetting branch. |

These are three related case IDs, not a recommendation for three registrations or whole-contract credit. Existing rows and historical evidence remain unchanged. Do not use their presence/count as qualification.

## Small coherent scenario sketch

A prospective **seven-request positive** on one fresh ordinary installed Transactional child/store is sufficient for the above bounded facts:

1. Seed one supplied-ID Bead and independently validate its complete Resource and actual revision from the terminal receipt.
2–3. Dispatch two batch POSTs without awaiting either response. Each has a distinct idempotency key and two operations: create its own distinct marker Bead; then update the shared seed with the same captured expectedRevision but a different literal title. No alias, owned Link, provenance, fault seam or scheduler manipulation is needed.
4. Replay the exact losing batch/key.
5. GET the actual validated losing receipt URL.
6. GET a complete current Bead collection with next:null.
7. GET a complete finite Scope Event page with next:null.

No initial probe, discovery bootstrap, polling, separate singleton rewrite or extra Resource GET is required for that sketch. Exact bodies, title, negative control, deadlines, output fields and byte budgets still require a real implementation plan and measured gate; this report releases none. Seven is the enumerated positive request count, not an executed observation or an approved complete validation workload.

The accepted same-key transport discipline can be reused locally after that slice freezes: canonical Host while dialing bound loopback, exact URL grants/query preservation, two bounded outstanding requests, immediate rejection handlers, bounded response streaming and full cleanup. Do not invent concurrent server admission merely because both client promises are outstanding. A healthy source-grounded fixture should produce terminal results; unexpected pending/service/transport outcomes must be retained honestly and must not earn this proof.

## Transactional distinctions that cannot be copied from RU

- Both terminal mutation responses are HTTP200. The loser is a schema-valid `status:failed, detail:available` Mutation Receipt with its own id/key/transaction, `problem.code:revision-mismatch`, embedded status409 and retry after-state-change. It is **not** the RU direct409 response. For this two-operation batch, operationIndex is1. Do not invent a pointer merely from the case name; pin the actual failure-extension contract during the detailed plan.
- Both pair requests carry distinct keys. Equality/replay applies only to the losing request and its own retained receipt. “after-state-change” does not authorize execution again under that retained key; spec5353–5356 explicitly means a new request under a new key.
- Seed commits pos-1; the winner commits one group at pos-2. The loser contributes no effectPosition/group. Its failed receipt requiredPosition is the preceding committed position, here pos-2. Actual current headers must be checked independently from receipt execution facts.
- The winning receipt has two ordered entries. The complete current collection contains exactly seed-at-winner-revision plus the winning marker, and no losing marker. Full Resource equality ties current state to independently checked winning postimages.
- The complete Event page contains the seed creation plus the winner's marker creation and seed update, with actual transaction correlation, revisions, data and within-group ordinals. No loser Event may appear. Use actual source-specific projected IDs if exact IDs are asserted, not guessed pos-only spellings.
- Full losing receipt equality across original response, replay and GET plus unchanged current graph/Event state supplies the retained/no-extra-effect check. This does not prove crash durability or that the losing marker's ID reservation was released; that latter fact would need another specifically justified operation and is not claimed here.

Settled law: spec2156–2160 makes a stale revision a no-state-change refusal; 4173–4183 makes an atomic batch failure roll back the entire transaction; 3420 supplies the revision-mismatch Problem row; 5353–5356 fixes failed-receipt retry meaning. Source supports the proposed oracle: resource-evaluator.ts1022–1027 checks the revision before final writes; transaction-batch.ts749–759 attaches operationIndex to the actual failure; transaction-store.ts1232/1266–1275 rolls back the effects savepoint and retains the failure, while1285–1306 publishes only the successful group's Events/position; transaction-http.ts176–189 emits terminal receipts as200. Thus the current source supports implementation rather than establishing a new product defect.

## Why this next, and what stays excluded

This adds a real profile distinction missing from the installed guard-contest evidence while sharing the semantic seed/guard/winner/loser concept with accepted RU evidence. It preserves Transactional response, retained-retry and whole-batch atomicity rules instead of duplicating RU's response expectations. A single bounded scenario can support narrowly attributed facts for the three IDs; future cumulative mapping must stay explicit and partial, and standalone Vitest evidence does not itself add catalog credit.

No new settled serving repair is justified from these inspected paths. Cold discovery remains under the existing cumulative admission gate, and the known pending product decisions remain outside this proposal. No History/Types/provenance work, public-admission bypass, diagnostic percentage or broad catalog redesign is proposed. Root owns priority, planning and later authorization.

## Exact source pins

- catalog/transactional-v1.json: `c9125f914129c71f01cefc5c54b32db80637d37b57e321571c96e81936c31eb4`
- transactional manifest-v1.json: `a11a79f6d0f9ffa5933a680a3e4fed87987789c3deca50a38cda8563132d242e`
- transactional coverage-v1.json: `346e1edaa22af789f89017fdc2fef6bb01b7564ed72a65344e9fe5635887c74f`
- transaction-store.test.ts: `7cd155ddd959ec31a207adbe6f8eb0a26f0af968a144abb58915720436bf0774`
- transaction-store.ts: `fcd6882b29e538db45867c894c06dfbaa6ca588db0c2e403ef1b7abb14f6c706`
- resource-evaluator.ts: `89606787e455f4ed555b807d91c5331a376431807a9a581de1f94e538fcab4ff`
- transaction-batch.ts: `70d941d416979edd54e250d51318cfb7e5d4ddb315ea6131de7ca5f56e98a014`
- docs/specs/bdp.md: `f036103ffdf5f89fa3a2033570e1e987c4cd203cd80b1c633cb44615123f892e`

 No product edits/imports/tests/builds/services/reviewer launches occurred; only this guarded external report was written.
