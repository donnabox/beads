# Native ordinary Update command-minimum code review

Reviewed binding881ebb8224d75fd899aae5851dcc6a76b72293ae42595ce56ea795c7141b0952; full six-file patch29cda7d745cd4091de2bcea19b55248e02d336e79b906ae7de62ffc3b33d7870; parent6f5e607bf338dd166820bfa3f9f480f7cd54c5f1. App1bf2661fa7bee90aa43622741d9e71d9770b2982d0a8804fa55ac1afb59e4d5d. All six actual hashes, changed-path set and complete diff verified. Accepted plan a49be8e237ae867f12d43691cf661ad44f5d49f92cce51d0369bb9a2615def18.

One concrete native finding was repaired; no remaining actionable finding in the final bounded code review. This is source/data review, not executed validation.

## N1: stale omission assertions — closed

Initial source updated counts but retained six positive assertions requiring command-minimum in uncoveredIds/omittedRequiredIds, contradicting wholeContract=true. Final app2731/2750/2826/3489/3911/4342 explicitly excludes commandMinimumId; still-omitted singleton.update-bead remains asserted. This closes deterministic failures in aggregate and three existing callbacks without new requests. Root's separate formatting finding is also closed: app diff698additions/25deletions replaces unrelated tab/default-format churn. Original packet/preformat history is preserved.

## Wire, oracle and negative review

App begin records genuine executeExact input ordinal/method/URL/credentialRef; runner.ts1211–1226 supplies no action ID. Actual exactRequest supplies ordered serialized headers and write/body facts. Names-before-values prefix checks protect credential diffs; exact case/order/empty/duplicate authored fields and absent minimum-field suffix are checked. Complete writes, exact body presence/bytes/digest and expected responses are required. Runner exactHeadersMatch1824–1868 separately enforces the complete serializer suffix. Query-free URLs avoid dispatch/report query-redaction ambiguity.

Actual response schema validation precedes baseline storage. Complete original singleton/sequence results, exact Resource metadata/properties/revision and sequence index0 are checked before equal replay and point-read equality. resource-evaluator.ts1125–1169 supports the fixture's strict metadata{} and absence of optional provenance/ownedLinks. Stale literal revision is proved different before dispatch. Row assertions preserve ordinary Problem status/code/retry; actual response checks exclude TX fields and Location.

The control is restricted to this scenario's second local completion, after real input/header fidelity, genuine200 schema and full replay equality. Only assertion-input outcome becomes updated; original private body/completion facts remain unchanged. Finish requires exactly2 exchanges/8 outcomes/7 passed and sole singleton-replay.outcome failure, one application and no cleanupError. Positive and aggregate require11 exchanges/47 passed and zero applications. Vitest callback success requires this structured expected runner failure; a generic process exit1 is not accepted evidence.

The optional bounded writer projects safe fields, rejects tokens/oversize output, creates wx0600 and retains genuine returned evidence before supplemental assertions and after cleanup. Primary/report/cleanup errors still fail. No raw header values/full bodies enter reports; safe files never substitute for successful process/cleanup adjudication. Existing target close/store removal behavior is preserved and still requires runtime observation.

## Preservation and bounds

Independently confirmed exact equality with accepted11-action/47-assertion scenario; all50 old baseline raw spans and129 other coverage spans are byte-identical. Remove only appended object/comma or restore omitted span to reconstruct complete parent manifest/coverage bytes exactly. Five continuation objects/file are untouched. Full corpus test preserves immediate and older cold-parent reconstruction. Profile mapping retains catalog inheritance and unchanged partial TX mapping; claimEligible remains false.

Actual manifest3,689,484 disk/1,912,593 compact bytes/98,833 nodes; coverage720,627 bytes. Old derivative transformation bodies are unchanged in the full app diff. Disjoint appended growth9,413 compact bytes/504 nodes yields existing36-recipe envelope1,913,704 bytes/98,868 nodes under unchanged4MiB/100,000 caps. This is source/arithmetic proof, not executed derivatives. Four coverage allowances72470/22013/17212/7966 preserve historical bases with exact2,320-byte growth. Current manifest pin10a7c6f9 and retained historical80f74625 are used distinctly; route/configuration digest26da5a28 stays pinned, no new route is introduced.

Eight CPU selectors and three runtime invocations match scope. Planned installed totals remain2327expected/2336maximum HTTP,62children including bootstrap (11positive/2negative/2314aggregate). No executed counts claimed. Actual clean package provenance, CPU/installed/control results, cleanup and exact-head CI remain root-owned gates.

No product edit/import/test/build/service was performed by this reviewer. Earlier plan/source-context participation is disclosed; I did not author these six implementation files. One report-write precondition initially counted seven exclusion statements instead of eight (six repairs plus two existing new-row exclusion checks); it failed before writing any artifact, was corrected by direct source inspection, and has no product effect.
