# Scope start=now reconnect — native plan review

## Findings and recommendation

**No actionable plan finding.** The proposed bounded HTTP change implements the explicitly supplied operator ruling through an existing exclusive-checkpoint path. The source/installed tests and ignored-header counterfactual can distinguish the relevant behavior without new runtime machinery or product semantics. Recommend root acceptance subject to the parallel council and normal exact-source authoring gates. This review is neither implementation authorization nor executed proof.

## Exact basis

Reviewed `transaction-start-now-resume-plan.md`, independently verified SHA256 `02cdb76e8f0f7282e8fa8887ffc161533b5672139cee282c9d374a17d02b2e11`. Read source in `/Users/dbox/repos/wt/vickie/vickie-noop-conformance`, inspected8d66855ec4e26d3af0b3d7d22d621e1f37005294/treea1ae75b822e1d7fadd739d818d09a25fb788e380, the tree root identifies with merged main5995f6b08ebd045e4343059c2da1cffd38b09b80. No remote lookup was performed. The earlier metadata/incident and other minimum rulings are not inferred or resolved here.

## Production mechanism and boundaries

`transaction-http.ts:428–445` first parses the starting URL and selects media, then gathers raw Last-Event-ID values. The existing special start/header refusal precedes the repeated-header check and effective selection. Removing only that refusal admits the combination to the same established path: selected SSE with exactly one field becomes `{after: field}`, repeated fields reject, and other media ignores the header. Headerless start remains unchanged. Missing/invalid/both/repeated/unknown query intents still fail in `transactionChangefeedRequest` before effective-header selection; the plan correctly does not let a header repair an otherwise invalid URL.

`transaction-service.ts:366–390` snapshots the effective intent before asynchronous work. Explicit after is reused across minimum retries; only a remaining start intent captures the owner head. `checkpointPosition` distinguishes genesis, malformed checkpoints, foreign epoch and view; the owner reads the group strictly after the resulting position. This makes the change sufficient without modifying service, store, cursor codec, stream pump or lifetime registry. Finite JSON start-now retains its captured anchor and ignores malformed/repeated Last-Event-ID.

The source regression plan preserves raw header multiplicity and independently checks empty/malformed/repeated fields and a genuine Event ID from the wrong token domain. It reuses actual checkpoint segments for authority failures and checks those before conditional shortcuts. Valid HEAD/304/412 requests retain observed-head context but return before a stream response is allocated (`streamResponse` and the response-kind branch at HTTP120–138,446–455). The new controlled-minimum case checks the saved checkpoint survives retry; existing headerless anchoring/cancellation/lifecycle tests remain gates. Event Source paths are untouched.

## Installed proof and counterfactual

The eight-request sequence is executable with existing fixtures: an empty Scope accepts start-now atpos0; G1 is committed and received; finite genesis read supplies an independently validated parity comparison; stream1 is closed before offline G2; reconnect uses exactly the same start-now URL plus saved applied G1. The offline G2 frame is required **before** writing G3. Thus ignoring the header cannot be hidden by receiving only later live G3. Final ordered[G2,G3], checkpoint IDs and finite parity reject inclusive/repeated/skipped replay in this fixed scenario.

Existing exact finite executors are constructed per request from a canonical same-Scope URL and literal method route, so checkpoint-based finite URLs need no shared allowlist or conformance-capture extension. The raw recorder preserves response bytes/headers/full frames and has explicit close/closed handling. Groups must first be checked against independent expected creates, not merely SSE-versus-finite equality. The plan states that prerequisite and distinguishes the small test-local application/checkpoint save from general durable-replica proof.

The ignored-header derivative is stronger than restoring the old500 guard: it must establish reconnect200 and then fail the named missing-offline-G2 observation. `recordRawSse.until()` rejects when its bounded socket closes without the predicate, allowing an explicitly labelled failure without waiting indefinitely or sending G3. A generic setup/schema/auth failure does not satisfy that discriminator. Derivative sources/packages/reports remain separately identified; no clean-source relabelling.

Accounting is consistent: new positive3POST+3finiteGET+2SSE=8requests/1child; existing idle scenario2POST+1finiteGET+1SSE=4requests/1child; canonical total12/2. The separate derivative adds at most8/1, combined ceiling20/3, with expected early stop atrequest5. SSE requests are deliberately nonterminal observations, not completed finite exchanges. Existing5-second finite,25-second installed recorder,45-second cohort,60-second outer and10-second cleanup bounds remain; the two clean positives run serially. The existing source fixture uses its own5-second recorder bound and controlled minimum wait, distinct from the installed25-second recorder. No deadline increase is needed by the described sequence. Actual final elapsed time, wire counts and cleanup still require measurement.

## Spec/catalog/evidence alignment

The exact proposed normative replacement is scoped to Scope changefeed automatic reconnect and adds exclusive replay after a saved checkpoint for start-now. Existing durable-application obligations and invalid/expired/foreign refusal remain. The initial intent wording is clarified rather than removing the required starting-query grammar. Event replay's separate after-only rule is not broadened.

Repository search confirms the maintained selectedText occurs in `packages/conformance/catalog/transactional-v1.json:1219` and `docs/design/w1-transactional-packet.md:6325`; those are both in the eight-file inventory. The existing mirror guard can add the reconnect ID without changing catalog identity/count/order. Other reconnect requirements remain intact. No manifest/coverage/claim flags are changed, and deletion-evidence author ownership does not overlap. Catalog digest changes are intentional current-source alignment, not permission to reseal or relabel historical Read evidence. The plan correctly requires stopping for any unforeseen binding failure or production scope expansion.

Progress prose replaces the live start/header gap while preserving finite-header ignorance and initial anchor behavior; historical assessments remain historical. Public discovery, whole-row qualification and general replication claims stay closed.

## Review limits

Only local plan/spec/catalog/runtime/test/recorder reads, source searches and SHA calculation. No tests, builds, installs, services, network, source edits, agents or outward actions. No owned processes; only this report was written. No installed outcome, parser qualification or performance guarantee is inferred from static feasibility. Full code council and exact-source fresh positive/derivative observations remain required.
