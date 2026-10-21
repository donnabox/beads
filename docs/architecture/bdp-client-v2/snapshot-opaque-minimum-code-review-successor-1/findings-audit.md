# Independent Snapshot CODE findings audit

No integration-blocking source defect identified by this audit. The optional observations should narrow evidence attribution; none requires changing the frozen source before the already prepared bounded validation. Root owns findings disposition and execution release.

## Review integrity

Read the complete actual final review, then inspected the cited production classifier, service capture/lifecycle, common waiter and A/C/D/F test code. Independently parsed all 5,466 raw stream records: exactly one successful terminal result, subtype success and is_error=false, whose text matches final-review.txt. Wrapper exit 0, elapsed 1381.0919885002077 seconds, no timeout, no cleanup requirement and no wrapper error. Before/after source records are identical. Rehashed all 27 preparation entries (using the preserved closed binding for the intentionally released binding entry), 7 explicit source files, 695 context pins and 36 artifact pins. The released binding differs from the closed binding only by rootLaunchReleased=true and matches root-release.json. Actual HEAD and exact seven-file dirty status match; all 679 tracked-source identities are covered by the bound source context. A fresh process census found sampled PIDs 6530 and 98177 absent. Sampled absence does not establish visibility into every possible short-lived process.

- Final review SHA256: f271f1d28e29683948958b543ff9776a0d2d47ce65fdea40a28cff22c285ecba.
- Raw stream SHA256: 0e8db4096665f3879fbdad973278d6200ff67cdbd0b2b9fc0536a34e77ec0990.
- Exit record SHA256: c2b32102e3a8d459932fc4e1f185fc3407293f231a8fb5be81f2129e5a33e67b.
- Terminal-results SHA256: 967f9be8de6e12038a10dc96c8583faa3c0a5430ec1229837140189623775e41.
- Identical before/after SHA256: b4e98597bbaeec59f6465f7f278317c25634c4fdb2a0852421e1e7efd53f4e6c.

## Findings disposition recommendations

**O1 — Accept the evidence limitation; defer optional test expansion.** In transaction-service.test.ts:3347–3349, abort and gate release occur in the same turn. The callback proves that an admitted wait is reached, abort is requested, the deferred wait is then released, and the operation yields ScopeServerOperationAbortedError with the asserted counts. It does not independently distinguish an implementation that settles cancellation before release. The proposed single Promise.resolve() checkpoint is not a robust remedy: rejection/fulfillment propagation through the waiter, service run and observed promise can require multiple promise reactions, leaving a settled flag false at one checkpoint even for premature cancellation. Do not describe this row as proving that abort cannot interrupt an unresolved wait. The separate drain sequence and real owner probes retain their narrower assertions. No timing or workload expansion is necessary for this slice; stronger cancellation discrimination would need a separately reviewed deterministic control.

**O2 — Defer defensive symmetry; preserve the deliberate original-error rule.** Both lexical catches rethrow the original strict TypeError. snapshotTransactionMinimum first invokes the same snapshotCurrentReadMinimum data-only capture, and the strict-error path has no intervening request.kind getter or other admitted user callback before repeating lexical capture. Proxies/accessors are rejected. A non-TypeError from the first capture propagates immediately; the suggested second-catch non-TypeError scenario has no reachable trigger in this admitted path. The catches are intentionally narrow around tuple capture, with request.kind outside. This does not justify refactoring semantics during validation. Unexpected-host mutation scenarios were not exercised or established.

**O3 — Accept aggregate-only attribution; defer per-request instrumentation.** D records eight total waits across its 27 requests. It does not measure a separate wait delta for every row and cannot by itself rule out compensating count changes across requests. Status/order/body/header assertions and the separately selected service branch tests remain useful independent evidence. Report waits=8 as an aggregate, never as observed one-wait-per-503 or zero-wait-per-known-row proof. No additional count or elapsed-time claim is warranted.

**O4 — No correction required.** The owner fixture takes the baseline, asserts concrete inventory values, performs the second inspection, compares the inventories, and independently asserts offlineInspections=2. The equality is not relied upon alone. The comparison precedes the final helper reopen, whose maintenance is expressly excluded from the measured no-expiry comparison.

**O5 — Preserve the explicit residual gap.** Existing emitted disconnect callbacks use strict pos-2. The selected opaque C rows cover service cancellation/drain, not emitted opaque client disconnect. Keep that distinction and do not claim new disconnect coverage or whole-profile qualification.

## Behavior notes and validation boundaries

**N1:** Lexically valid tokens rejected by the strict position parser include opaque words, noncanonical pos-01 and unsafe-integer pos-9007199254740992. For a fresh Snapshot in the actual matching epoch/view they enter the same bounded Unproved path; this is intended classification, not numeric normalization. A directly pins pos-01; unsafe-integer behavior is source-derived here, not a separately selected new test. Malformed lexical tuples, retained requests and foreign contexts retain their separate fences.

**N2:** Unknown minimum plus foreign epoch/view still produces private minimum-position refusal (500), rather than the known-position 410/409 paths. This remains unresolved, excluded policy/conformance territory. Preserving it is not blessing it or claiming normative coverage; this slice only changes eligible fresh matched-context behavior.

**N3:** Snapshot capture still happens synchronously before run/check. Consequently malformed input wins over a closed-service refusal, and malformed-plus-aborted input also retains capture-first ordering. The F rows pin this behavior. Do not transfer the distinct Event first-attempt capture ordering to Snapshot or call this lifecycle-first evidence.

Claude performed no product execution. The completed five-stage static prefix is separate evidence and has its own native audit; it does not turn prospective callback counts into passes. Fresh build and build binding remain required before source HTTP collection (dynamic dist import) and installed execution. The reviewed artifact remains a seven-file dirty successor on 0f215c05678f507dd8a39ee08e288f0901710135, not code contained in that base commit. No release, discovery, admission, S1/G2 or retained-continuation conclusion follows.

No source, test, helper, review result, gate or prior evidence was changed or executed by this audit. Only this report was written through the current-session guard.
