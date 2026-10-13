# Owned-Link evidence follow-up

Frozen clean head **4d13f4104764d2a4757909c7ee48bdba97789374**, parent **c0d67388cd849410b38ec7ecb3a328771431dbfa**.
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-owned-evidence`, branch `codex/vickie-transaction-owned-evidence`, origin `https://github.com/gastownhall/bdp.git`. Local commit only; no outward writes. Frozen owned01e2 and all prior review sources remain unchanged.

Three files, tests and documentation only:

- `transaction-owned.test.ts`: two new actual-owner tests. Native `revisionProvenance: {agent,message}` on creation/update/singleton deletion/deleteWhere pins exact committedAt/agent/message output. Creation and update assert Link, source, Link Event and nested/source Event attribution equality. Delete and deleteWhere assert fresh source attribution, identity-only nested deletion, per-member source revision chain and empty final owned set. A mixed first updateWhere preserves the first sorted member as a no-op and changes the second; the next set's cardinality requires both staged matching members; a final explicit update follows. Assertions pin seven receipt entries, five per-member source revisions (including unchanged no-op), four source transitions/eight facts, intermediate member properties and final normalized source+Link postimages.
- `transaction-service.ts`: only stale introductory schema-free/unowned comment corrected. Event author owns active service changes; retain that author's header correction if ordinary merge overlaps.
- `transaction-implementation-progress.md`: historical schema-free/unowned statements explicitly dated by slice, owned deletion attribution corrected, new evidence and deferred parsing optimization described.

Normative inputs checked against MintInput/RevisionProvenanceInput and bdp.md owned Event/provenance/no-op law5437–5509. No runtime/evaluator/sourceRecipe/format/policy/identity/limit/public surface change. The4096 boundary source test and historical evidence remain unchanged; it ran again at this c0 configuration but is not relabeled for a later materialization tuple. External owned review F3 repeated parse optimization is intentionally deferred during separate materialization work.

Validation with Node24.16.0/pnpm11.20.0:

- Frozen-lockfile install; build before emitted/process cases, final build exit0.
- Focused owned suite **17/17 passed**, one file,3.09s,maxWorkers2; final `tests-final.json/.log`. Includes both new tests, actual process-kill recovery, compiled HTTP and existing real4096 boundary (2.08s).
- Typecheck, dependency boundaries, format check, lint and historical89-row Read evidence verifier pass. Final lint has34 inherited warnings/one info, no errors. No new evidence generated or resealed.
- Initial lint identified unsafe optional chaining in a new assertion; replaced with explicit missing-Event guard, reran focused suite and final checks. Initial passing17test run and superseded lint log retained.
- Broad all-repo tests and package smokes not repeated: production behavior/export/build inventory did not change. No stopped/hung runs or active test-owned processes.

Remaining: separate Event/materialization/G1 lanes; installed profile qualification and broad protocol gaps remain as recorded in the burn-down. This adds evidence, not a Transactional profile claim.
