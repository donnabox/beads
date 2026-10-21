# Sequence refusal context — independent native code review

**Clean: no actionable production, test, security or documentation findings.** This is a full review of the frozen implementation and its completed source/control/installed artifacts, separate from the earlier static preflight. It does not grant root acceptance, replace the external review/CI gates, or establish whole-profile qualification.

## Exact source and review boundary

Reviewed the complete six-file diff from `d50f554cf6418ba036072ebf6795cb272cd6cbc1` to **`90cc834c3672b02840c68dcd7abb35794f038887`**, tree **`97db367180fb49a20e831fba859e0ac09482966b`**, in clean `/Users/dbox/repos/wt/vickie/vickie-sequence-refusal-context`. The commit has the assigned ordinary parentd50 and829 additions/6 removals. All six current file hashes match `transaction-sequence-refusal-context/source.json`.

Read accepted plan v2, its dispositions/independent closure, root acceptance, corrected preflight and measured execution release. Reviewed the actual owner, service lifetime and HTTP receiver surrounding the diff, the entire new installed test, added service/HTTP tests and progress changes. Compared reported test/control outcomes with structured results and failure stacks, and read the three completed packed observation records. No additional tests, builds, services, installations, controls, network or outward actions were run. Read-only file hashing/arithmetic does not execute project code. Only this review artifact was written.

## Production reasoning

`transaction-service.ts:525–536` adds one private observation. Existing run() checks accepting state and registers the returned promise before invoking the callback. The async wrapper performs owner.withRead synchronously before returning; its callback itself is synchronous and returns a frozen wrapper around the owner's already-frozen three-field context. The owner entry ends before any awaiting caller can deliver the response. No reader capability escapes, no scheduler/wait is added, and the operation remains within service drain/close lifetime.

The actual store withRead at883–958 refuses reentrant/closed/fenced entries, constructs context inside its synchronous entry and invalidates the reader in finally. The new method reads no Resource, key or receipt, invokes no admission or execution function, and returns the same current owner tuple. Entry refusal retains its cause in an internal error; storage errors propagate. Existing HTTP catch handling sanitizes these as empty500 without fault details. Normal runtime closing/not-ready503 still occurs before handler dispatch. This preserves the accepted distinction: requiring a serving observation can expose an exceptional fault; it does not promise every installed fault is prevented by the closing gate.

`transaction-http.ts:316–335` calls the method only for the already authenticated/authorized sequence operation-count branch and supported-carrier unacceptable-response-media branch. Existing carrier-key/syntax/auth order remains, count413 still precedes media negotiation, and unsupported composition remains before406. The count Problem still has code=limit-exceeded and limit=sequence.operations;406 remains empty without Content-Type. Existing common finalization supplies non-attesting/Vary behavior; the context wrapper supplies the current tuple without changing admission or creating a receipt. The rest of direct submit routing and successful sequence execution is untouched. No blanket error decorator, cached fallback, arbitrary comparison/minimum treatment or new semantics was introduced.

The change implements the accepted application of the general spec5768 Scope-response law. Unexpected-fault500 follows4588–4592; it does not import successful-sequence replay recovery as a reason to omit context. The service discovery closure, store/parser/configuration/export/schema/catalog/runtime and public profile gates are unchanged.

## Test and oracle review

The service test uses the real owner while wrapping withRead and admission methods. It verifies exactly one entry before awaiting delivery, identity with the captured context, frozen wrapper/tuple, unchanged inspection, no direct/sequence admission or nextSequence call, later position advancement and immutability of the earlier observation. Same-position discovery comparison is outside the one-entry assertion; it does not conceal an extra observation by the new method. Reentrant/closed-owner cases preserve causes and check HTTP500 sanitization. Excluded carrier-key400, malformed400,401,403,415 and minimum-development500 retain all three absent headers and zero new observations; the development marker is preserved.

The two HTTP parameterizations independently start with413 and406. Both exercise genuine seed positions, canonical count Problem identity, bodyless negotiation, tuple presence/value, cache/Vary and repeated refusals after another commit. Final principal/key/receipt SQLite inventory is read-only, occurs after all live observations and closes the runtime. Exact baseline seed receipts and every rejected key are checked separately from graph/Events. The different-identity key-reuse test uses a separate fixture and correctly expects only the subsequently admitted receipts; it does not contradict the refusal-only absence oracle.

The new installed test grants four exact static routes plus an observed seed receipt root. Its local root guard validates exact canonical href, fixed Scope prefix, one nonempty bounded token segment, no query/fragment/userinfo and8192 length/byte bounds before granting the GET. The source-only guard test has two valid and20 malformed cases, without starting a child. This closes the preflight gap; the generic receipt schema and exact executor alone would not have confined a root path.

The installed flow is exactly13 requests. Both413 and406 occur atpos-2 and repeat with fresh keys atpos-4. Refused406 key reuse changes properties; refused413 key reuse changes identity. It independently expects IDs, authored properties, metadata, finite collection/Event counts and next:null, while relating opaque revisions/transaction IDs to actual receipts. Historical receipt body remains unchanged while its serving headers advance. These checks are meaningful within their small fixture and do not claim every possible cache implementation or all semantic carriers are covered.

## Completed evidence inspected

Source records preserve two initial typecheck failures caused by Vitest's overloaded admit spy typing. The final test-only cast explicitly goes through unknown with an explanatory comment and still delegates to the real owner; no production behavior was changed to make the typecheck pass. The third typecheck passes. Build, formatting, lint, boundaries and unchanged Read evidence verification records pass. `focused-1.json` contains **373 passed tests in seven actual test files**, zero failed/pending; its27 nested suite count must not be called27 files.

Five deliberate control failures are present in structured results and their stacks:

- Omitted context,413-first and406-first: the respective named epoch-presence assertions fail.
- Stale context,413-first and406-first: the respective named position assertions reportpos-1 versuspos-2.
- Over-wrapped exclusion: the named carrier-key malformed-request epoch-absence assertion fails before the no-observation spy.

These are three derivative categories/five selected runs. They fail the intended predicates, not setup, compile or timeout checks. They demonstrate specific sensitivity, not universal proof against every fault. The new source tests and the historical direct-refusal tests remain separate.

Completed packed observations show:

| Cohort | Exact source | Requests/children | Result |
| --- | --- | --- | --- |
| New sequence positive |90cc834|13/1|Pass; two installed-file tests pass, including the static root guard.413/406 tuples observed atpos-2 andpos-4.|
| Unchanged direct regression |90cc834|14/1|Pass; prior conflict, direct capacity and fresh/retained negotiation behavior retained.|
| Predecessor negative |d50f554|3/1|Expected failure after genuine413; seedpos-1/2 succeed, refusal context is empty. Named epoch-presence assertion and separate postprocessor verdict both identify the intended missing-context failure. Static root-guard test passes.|

Total **30 requests/three children**, within the prospectively accepted40/three ceiling. No replacement cohort is present. The negative checks canonical status/code/limit before calling the named context predicate. Its real failure stack reaches that predicate at installed test289/385, so the evidence is stronger than exit1 alone.

Reviewed summary/session records show child30256,30965 and31219 exited0, temporary data directories absent and author post-run PID-absence checks recorded. This reviewer did not launch or poll those processes. All three postprocessors pass; only the negative marks its required label checked. Response assertions operated on real in-memory bytes; retained observations honestly contain fingerprints/context, not complete wire transcripts.

Independently recomputed package archive and installed entry hashes from the completed records. Positive/direct use90cc, negative usesd50; source/tree and six source-file hashes match their records. The unchanged entry hash is the same across payloads, but actual installed service/HTTP module hashes differ, as expected, and the payload digests differ. An equal entry module is not treated as proof of equal server packages. Bodies remain within the measured5728 maximum/11970 positive total; runtime/executor limits and deadlines were not raised.

## Documentation and remaining gates

The progress document now scopes the prior HEAD/304/412/406 statement to fresh Snapshot creation and labels the successful sequence post-runner observation/replay discussion accurately. The new paragraph describes only the two selected nonexecuting refusals, discloses exceptional500 and preserves the no-qualification boundary. Since accepted preflight, only test spy typing and this evidence-boundary prose changed; production and installed scenario hashes remain as reviewed at preflight.

The author subsequently confirmed final report/source/aggregate records complete and stable. Read the complete report (SHA256 `06ee8ce82888ac9c04d5170eeb51a72ac4ff219d3d34b683b5e7489bc07df6d2`) and reconciled source (`1f324b21ea0f0337a66b7f2c67a8f034e6759267629f1cfab035f5a2958d8459`) and summary (`db3c42169a1b734072767401147d462d91f74cab5ce3b555e903dcbef9d7458d`) with the already inspected cohort records and underlying artifact hashes. The report also preserves an optional failed Prettier availability check separately from passing repository Biome formatting, and does not claim lint was warning-free. These preparation failures are not hidden or counted as installed retries. No discrepancy affecting this review's conclusion was found.

Root still owns final report reconciliation, external council disposition, publication, exact-head CI and integration. This clean code review provides no public discovery/profile activation, complete response-context coverage or cumulative qualification claim.
