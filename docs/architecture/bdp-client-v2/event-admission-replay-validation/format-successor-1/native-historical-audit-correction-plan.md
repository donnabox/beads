# Historical installed audit failure: diagnosis and bounded helper plan

**Confirmed auditor schema bug; no product defect established.** At audit.py134 the common installed predicate reads `o['entryLiteral']`, but the unchanged historical restart witness emits only `entry` and `entrySha256` (installed source970–981). The new replay witness alone adds entryLiteral (2515–2518). Requiring its newer diagnostic field from historical evidence causes KeyError before the remaining audit. My earlier preparation review missed this shared-field assumption; its clean readiness verdict did not establish this historical schema compatibility. Preserve that prior review and this correction rather than rewriting history.

## Actual evidence

Historical callback raw0/accepted in5.367056s; immediate auditor raw1/notaccepted in9.291281s. Callback exit3866d42688926fb105001418b9ea469841a86a669ae525111037d20efcc99386; raw logacaaa901831f86b6d5f60742e2a1efabaf58e98750f8a78c1d70b8e28e34821e; results e729d29ce8e7026e8874c9dc9e2f9af270465fb5c456588d53bfcaa14313d0ba. Auditor exitbf93062294c7bb612f511447728de5da599aa0548efa65d26db2481b35f423b8; traceback log2372646868c5fe0ca437c11c64a15b5845e822d23995cafbf959d7f15b2d36e3. No accepted historical-audit result exists. New replay10 callback/audit and required4 have not run.

Actual historical report has five requests, four finite responses200/200/200/400 and separate successful SSE; two children8668/8671 closed exit0, no signal/stopUnknown, drained transport, pending0 and removed store. Current directory absence and all48 sampled PIDs from these two stages independently checked. These facts do not turn the failed auditor into a pass or add a closeEvent field to historical evidence. Source679 still matches07c407cf and the gate is now releasedfalse/approvedStagesempty.

Actual historical entry resolves to proof entry, entrySha256 equalsbcf15317, and child payload equals verifiedda80f6d7. The missing literal field is not missing installed provenance: common resolved path, file digest, child payload, exact environment injection and consumer proof already supply that binding. Historical does not claim the newer exact literal observation.

## Exact correction proposal — helper only, not released by this report

In a distinct immutable helper successor, replace only the existing audit.py134 combined line with:

```python
proof=verify_consumer(g)
require(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'External entry/payload')
if not history:require(o['entryLiteral']==proof['entryRealpath'],'External entry literal')
```

Use the existing branch indentation. This keeps common provenance strict and the new callback literal requirement mandatory. Do not add `.get(..., proof...)` defaults, fabricate historical fields, remove entry/hash/payload verification, or change source/tests/report contracts. Keep exact source679, selection, workload, callbacks, deadlines and raw0 acceptance. Preserve original helper/indices/run records in place; create complete before/after correction patch with actual audit.py before-image, inverse reconstruction, dependent hashes and CLOSED root gates. Native exact-delta review before execution. No repeat actual CODE council needed for this report-consumer-only correction.

## Remaining predicate scan

Read all predicates after the failure through accepted report write against actual historical JSON and both frozen report constructors. No second schema mismatch found:
- Seed history244bytes/two creates versus new three operations already branches; history events2 versus new baselineEvents8 already branches.
- History sessions2/create true,false/exit0/closed versus new sessions1 plus mandatory closeEvent already branches; no historical closeEvent or new pre-close live object is invented.
- Requests5 versus10, finite responses4 versus9, historical ordinals1/2/3/5 and new1–9 are distinct. All common headers/digests/context fields are present. New opaque terminal body length/Retry-After/frames/drained checks are appropriately nonhistory-only.
- Historical SSE request4, session2, lastEventId, bounded beforeIntentionalCloseMs and two sensitivity controls match actual JSON. New SSE request10/liveBeforeClose/framesBeforeClose/marker and one control are correctly separate.
- Both streamAtReport shapes contain required status/headers/frames/drained/destroyed/wire fields; only new branch enforces equality with its richer recorded stream identity. Common stream_headers deliberately checks the reported subset, not unavailable whole-header telemetry.
- Store prefix and absence branch correctly. Actual frame data equals the second historical Event, and its report says bounded observed frames, not indefinite absence.

This scan uses data reads/comparisons and frozen source; it did not invoke/import the auditor or produce a replacement accepted result. The new callback has no actual result yet, so its compatibility remains static pending fresh execution.

## Revalidation boundary

The implementation source and dependency inputs are unchanged. Semantically this defect does not invalidate prior static/build/source/package/consumer passes; the direct affected requirement is a fresh historical callback plus corrected immediate audit, followed by the still-unrun new callback/audit and required4. Such a minimal continuation would require explicit pinned reuse of the15 prior stages and build/consumer/source origins, never relabeling them as fresh or treating the failed audit as accepted.

For the existing proven successor architecture, recommend the simpler established execution contract: original bootstrap2 only reused via unchanged17 dependency/11 tool/8 record mapping; all19 later stages and4 required checks fresh in the separate helper successor. This reruns unchanged preparation deliberately to avoid introducing a new15-stage continuation mechanism while repairing this bounded helper. Root may choose an explicitly reviewed continuation instead; this report authorizes neither. No automatic retry, timing expansion or product/source correction. Root must approve the concrete helper change and frozen execution scope before any run.
