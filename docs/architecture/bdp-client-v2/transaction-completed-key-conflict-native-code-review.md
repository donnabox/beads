# Completed-key conflict — native code review

## Finding

### P3 — complete the accepted independent response-oracle pins

**Location:** packages/conformance/src/transactional-cumulative.test.ts:1000–1059 (assertCompletedConflictRequests), with the existing context guard at868–905 and new-row declarations in manifest-v1.json.

The new source guard independently pins the twelve-action order, five literal bodies/keys/Accept values, request targets/credentials, root capture and authored request observations. It does not pin the response-side obligations explicitly accepted in plan77f6deb: both409 Problem shape assertions, independent one-record/Event tuple projections and literals, original body/record captures, retained-record equalities, replay/root body equalities, and current-context comparisons/presence. The existing required-context guard filters only the M/Z set rows. Coverage mirror validation derives its expected action/assertion names from the manifest and cannot substitute for these semantic pins.

A concrete source-level weakening is to keep each c-state/c-state-reopened retained-record assertion ID and pointer but replace equalsBinding:c-record with exists:true. The current guard would not detect removal of full revision/provenance equality, the coverage mirror would remain unchanged, and the five installed sensitivity controls target other oracles (the state control mutates ordered-state). This is an inferred regression-protection gap from the complete source, not a mutant test executed by this reviewer. The current manifest itself contains the correct full equality and the current packed evidence is not invalidated by this finding.

**Minimal repair:** extend the owned cumulative test's independent expected assertion/capture objects to the selected response oracles, including both complete state comparisons, three retained body comparisons, exact Event/state tuples and both409 schemas/facts. Include the conflict row in the context-presence/equality guard (and reconcile its presence order with the accepted guard convention). Add a bounded static weakening control, e.g. the retained-record equalsBinding→exists change, which must fail the named guard. No new requests, runtime behavior, coverage credit, sixth installed control or framework is needed.

No runtime defect or false whole-profile claim found in the reviewed delta. Root owns disposition of this Low test-completeness finding and the independent external review.

## Exact source and review scope

- Head b13055049ed7c1bd09ad2d7af29e0a38a81001ea; tree3bda07f860204eba0fcd5b2fe65335dadf99653f; direct parent2172092de12752867b43858ebcbdd2c37d4c727f.
- Checkout `/Users/dbox/repos/wt/vickie/vickie-completed-key-conflict`, clean when inspected.
- Read the full four-file diff (1348 additions /12 removals), the complete new twelve-action/99-assertion manifest row and coverage mirror, accepted plan77f6deb/root acceptance, author report/source/summary/final-source-audit/mapping audit and actual packed positive report.
- Read actual executor semantics for response-metadata equality, JSON-pointer captured equality and ordered array tuples; source-wire receipt validation; retained-source generation/config checks in the app; normative Mutation Transactions and batch admission/negotiation clauses. No tests/builds/network or repository changes were performed. Only this external review report was written.

## Current evidence and clause interpretation

The two selected conflict clauses require different semantic requests under the same bound key to refuse409 before admission. The five literal presentations are three byte-identical create-a bodies and two byte-identical create-b bodies under one key. Changing the explicit identity creates a genuinely different semantic request whose accidental execution would visibly create b. The first conflict's text/plain Accept tests existing comparison-before-response-negotiation behavior; it is not a newly invented ordering rule.

Both current conflict assertions independently require409, transactional Problem schema, conflict type, idempotency-conflict code, status409 and retry never. The source static request guard prevents accidentally replacing changed semantics with an ordinary duplicate. Neither refusal establishes a receipt binding or capture.

The successful original POST's root capture is checked by the unchanged private receipt ledger against actual POST wire. c-original's private read is validated against that original body before c-body/c-record captures are used. The raw HTTP capture is only the URL; full JSON captures use the supported client path. Both later original-body POSTs and the reopened root compare complete retained bodies. Each current collection checks currentBeadCollection schema, no next page, exact ordered one-record literal state and full captured-record equality. This rules out a visible second Bead and preserves opaque revision/provenance fields for this fixture.

Both Event reads check eventPage schema, no next page and exactly one created/a/ordinal0 tuple, together with current epoch/view/position equality to the original committed response. Executor equality can treat absent/absent headers as equal, but this actual manifest also independently requires all three headers on both sides. The presence assertions occur after equality in the two Event action arrays; because all declared assertions execute and must pass, this does not weaken the present successful result. The missing independent source guard/order is covered by the finding above, not presented as an existing false-positive observation.

Reopened private-root evidence has actual completed wire, original sourceReceiptObservation, later served generation, unchanged logical configuration/runtime identity and exact URL. The unchanged app loop verifies each of these and requires the successful client context-presence assertions. The five controls run real target exchanges and change only declared expectations. Each observed exchange has complete actual request wire, expected actual status and exactly its named failed assertion; the scenario fails as intended. These are oracle-sensitivity controls, not injected server-corruption or internal no-admission proofs.

Coverage remains wholeContract:false for all ten rows. The new reason expressly limits this to one principal/fixed view/epoch, available completed detail, explicit-ID create batch and same-epoch restart. It excludes all-duration/expiry, failed-retention transitions, new epochs, other principals/views, pending/concurrent/cross-carrier cases, other operations, capacity, undisclosed internal receipt creation, History and erasure. Partial credit is appropriate; no additional selected law is implicitly claimed.

## Independent preservation and bounds audit

Read-only raw JSON span extraction compared parent to candidate, separately for both artifacts: all nine previous manifest objects and all nine previous coverage objects are byte-identical. The pagination spans independently match the added ninth pins: manifest138750B/c531c260349347cdc3e54ba68282f1a266a5fa207d3775660615f836a4768782; coverage13057B/0735edbca2fbf384f9f8147140a353dc084d9ad85facaddf2f5073f56a7256ee. The original eight pins remain unchanged.

The new row has12 actions and99 unique ordered bindings; its coverage bindings exactly equal the actual row-local action:assertion mirror. All ten whole flags are false. Four changed-file SHA-256 values match source.json/summary.json. No runtime, fixture, schema, route configuration, old installed-source identities, historical evidence or public admission file changed.

Measured source remains within the explicit accepted authoring ceilings: manifest407227B/10891 nodes, coverage60059B, total151 actions; effective headroom2990B/391 nodes/3008B. Fixture716B/SHA3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85 and configuration digest634214546089490d235fa404166dc868a0feb5dc2a8af2a566a6ad33664b4b1a remain pinned. New POSTs173B and response maximum696B fit the unchanged2KiB/8KiB row bounds. Existing M/Z and pagination bounds and63/90/120/5/10 timing controls remain intact.

## Inspected executions, not independently rerun

Author frozen cumulative:27 passed. Fresh packed positive:1 selected passed,48 other app tests skipped; ten authored rows pass,151 actions,17 scenario/reopen sessions and18 children, all closed. Actual full result contains245 harness-error rows with category not-implemented / no executable plan; they are retained visible gaps, not target failures suppressed by this review. Required-only unplanned count240 and five optional gaps remain distinct. claimEligible is false.

Fresh packed controls:5 selected passed,44 other app tests skipped. Actual failed assertion/status pairs are problem-code/409, http-status/409, ordered-state/200, retained-body/200 and ordered-facts/200. Five distinct mutated manifest digests are retained. Two early controls launch17 children because their row stops before reopen; the other three launch18. All launched children close. The author preserves the first collector's incorrect universal18 expectation separately and recollects the same completed genuine reports with the correct17–18 accounting. No test failure was silently reclassified as target success.

The report also retains the initial TypeScript union-field failure and initial malformed target-query factory failures, followed by corrected source and final successful checks. Author type/build/static/boundaries/Read-seal exits are recorded. remaining-processes.txt contains only a newline (no process entries). This reviewer launched no servers/test children and needs no cleanup.

Inspected report hashes:

- report.md:5eb4fca95a38458d7ac02ccb4d594ee433ea82f299c6edffe333f06f94ae1487
- source.json:b9e352806d20f5b1fc78d667c860e4f4abf478a1e2f88bde16c863b2bc0e075f
- summary.json:4e7b89e216a9eaa394bdfe816341c264d48f2319382a07c4c98cfd4fb0e65a05
- mapping-audit.json:ad8c14c24e8994b8df5e1d011a6a19bfecd29d83e8b364c2d27df0b68244158c
- final-source-audit.json:5932b9613aeae4e8eab31e4c1ea1ba169959714efbe7d506d6bf16cded43a5b1

All reviewer commands were read-only source/evidence inspections and completed successfully. No broad runtime re-audit, full suite, independent installed execution, whole-row promotion or main/PR action is claimed.
