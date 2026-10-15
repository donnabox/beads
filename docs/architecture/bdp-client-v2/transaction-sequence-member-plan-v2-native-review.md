# Native review — sequence-member plan v2

Two Medium corrections are required before authoring. No production defect is established by this plan; its value is bounded regression evidence for separately retained sequence-member transactions, failure continuation and replay. The ten-action narrowing is technically viable after the corrections below and root acceptance of its partial proof boundary.

Reviewed plan SHA256: 972e4827075598e6223648229199ff3ed91e3ec317acc1961f4a7f6cababc936.
Companion Markdown SHA256: 70ebff17cedab807f029a7d564cf3608eba21a5d4e69271613b49e6b0d340e52.
Companion JSON SHA256: 803a7980b6626b39bf0235198f2bb8f10fd6b7196e4738df6862b5bc581d4865.

Inspected main/base 453b6dab386746bb03aac2c0a293572b8a71c969, tree40782d97c9764bc19ae78c46d1d39475e3db317a, and current frozen command-minimum evidence source in vickie-command-minimum-evidence, full diff ea04e6ba34c9e1d9b724348f32407c6fd6bd78b795cad0481e7f4093b0f21762. The plan honestly binds an older preserved pending snapshot; the required accepted merged-parent reanchor remains mandatory. This report does not authorize drafting or execution.

## Required corrections

### N1 — Medium: selected negative expects fields absent from deletion projection

Location: plan line52, “independent actual-success assertions pos3/deletedkind/id/type/properties/metadata”.

Actual source: packages/server/src/resource-evaluator.ts:1073–1084 constructs outcome deleted with deleted.resourceKind and deleted.resource containing exactly id, type and revision. It does not include properties or metadata. transaction-batch.ts:712–719 incorporates that outcome; transaction-sequence.ts:57–76 preserves it while rewriting operationIndex for the current envelope.

Impact: if properties/metadata are implemented as required-present literal checks on the deleted snapshot, the intended real-wire negative will fail its independent positive oracle even though serving behavior is correct. That masks the exact seven baseline failures and makes the planned control impossible.

Fix: pin the complete deleted identity shape {resourceKind:"bead",resource:{id,type,revision}}; assert absence of properties/metadata there if desired. Compare deleted.resource.revision to the independently validated first creation revision in the app. Keep full properties/metadata checks on the first and last successful Resource records. Preserve middle-no-resource and the seven planned failing baseline IDs; no request/count/source-production change is needed.

### N2 — Medium: exact static selection omits the preservation/fence callback that must change

Location: plan lines58,76–80 and companion staticSelections.

Actual source: packages/conformance/src/transactional-cumulative.test.ts:836, callback “preserves every original TX scenario and coverage object while adding one bounded partial cross-carrier row”. Its lines855–866 reconstruct a historical parent after stripping only minimum and cross-carrier rows; lines875ff pin the exact fourteen-row order; lines916ff pin total210 actions and canonical disk/normal/node/coverage fences. The new sequence row necessarily affects these checks. The selected minimum callback at2724 also currently reconstructs its historical parent after removing only the minimum row.

Impact: the six selected callbacks do not exercise the canonical size fences, exact row order/action count or historical cross-carrier reconstruction. Leaving the reconstruction unchanged fails later CI; overwriting its historical digest would erase preservation evidence. Recounting artifacts externally is useful but does not execute these amended test guards.

Fix: add the existing :836 callback to the exact static selection, yielding seven CPU-only callbacks and no new registration, HTTP, listener or child work. Explicitly remove the new sequence row before applying the existing historical minimum/cross reconstruction, keeping all historical parent digests. Separately bind the actual new immediate parent and its 28 raw objects. Include the corresponding staged fence checks after root’s measured allowance decision. The four-file ownership scope can remain unchanged. If keeping six callbacks is important, root can explicitly replace an unchanged generic callback; do not silently omit the affected preservation/fence test.

## Source-confirmed viability and boundaries

- Spec docs/specs/bdp.md:2720–2765 and4057–4080 support ordered separate member outcomes, continuation and the receipt specialization. The catalog member-transactions clause at transactional-v1.json:866–876 includes pending admission; this terminal-only witness does not prove simultaneous initial pending admission or ownership/interleaving. Keeping wholeContract:false and claimEligible:false is correct.
- transaction-store.ts:1105–1112 allocates each new receipt root and transaction; :1258–1295 retains a failed receipt at the previous position and commits effective members separately. For create-a / missing-delete / create-b, required/effect positions pos1, failed requiredpos1, and finalpos2 follow the actual fixture.
- transaction-batch.ts:700–714 assigns the failed singleton receipt problem operationIndex0. transaction-sequence.ts:66–76 replaces it with current envelope index1. Recovery via equivalent singleton batch bodies and the same three member keys is supported by the existing normalization/retained-key path; replay carries no outer idempotency key.
- The three batch recovery POSTs are needed to obtain genuine receipt roots absent from the inline sequence envelope. The three real receipt GETs then exercise helper original-wire JSON equality. runner.ts:2468–2487 accepts only scoped URL raw captures; transactional-observation.ts:206–244 authorizes and retains the actual response and :303–307 checks GET equality before projection. No full report-body inference or new capture API is necessary.
- transaction-store.ts:1543–1635 finite Event reading iterates retained groups under item/byte limits. Two small Events can be checked with next:null. The one-group Changes read is different; dropping s-groups avoids an impossible complete-page assertion. It removes Changes-group enumeration credit, not the independent terminal receipt/position/Event proof. Root must accept that narrower scope explicitly.
- The Event-ID construction should follow the current repaired private source-bound oracle at transactional-matrix.test.ts:2510–2535 and transaction-events.ts:73–100, using actual validated context plus fixed fixture source/checkpoints. It must remain a private reference-implementation spelling assertion, not a protocol ID-format claim.
- Full original Resources, failure Problem/absences, whole sequence replay and Event data/revisions remain justified. Pairwise receipt/transaction distinction is independently assertable because all three actual receipts expose validated fields; this differs from the predecessor’s C sequence with no receipt baseline.
- crossCarrierWire at app:1897–1924 rejects sequence routes after its 2/3/10 ordinary-response prefix. The proposed exact new-body/no-outer-key handoff after cross.finish(), plus exact later replay position, can preserve this guard. Forward all nonexempt traffic to the old sentinel; particularly do not exempt all sequence URLs or all traffic after activation. The new seven-ordinary-response ledger must finish, and its three receipt GETs remain checked through their separate client/helper path. commandMinimumWire’s completed guard at2344–2353 already discriminates its own sequence body digest rather than banning all future sequence URLs.
- Preserve report retention before cleanup/postchecks through the existing local onObserved callback. A callback must not invent a report when execution did not return one. Separate genuine report failure IDs from hook/harness/transport/cleanup errors.

## Counts and finite workload

Independently recomputed every companion body byte count and digest. Positive POST bodies are414+147+66+146+414=1187 bytes, largest414; negative408 bytes. The ten actions consist of seven ordinary wire actions and three internal receipt GETs.

The four-cohort arithmetic agrees with current source’s203-response/22-child positive and194/195-response/22-child cross controls: after adding the ten-action row, positive213/23, old controls204/23 and205/23, plus selected negative1/2. Total623 expected HTTP,632 allowed maximum,71 children. These remain proposed source arithmetic, not observed execution. The one selected negative must stop after its first sequence exchange because its retained baseline assertions fail.

The recurring increment decomposition is arithmetically consistent with the retained source inventory:40 full traversals add400 HTTP/40 children; five ready failures add five children and zero wires; global receipt denial reaches action3 and adds3 HTTP/1 child; selected new control adds1 HTTP/2 children. Thus404 expected HTTP/48 children; conservative420 HTTP permits the denied/control rows to progress to their ten-action bounds.54→55 expanded observe invocations is consistent. Exact expanded registration/title and per-specimen inventory must be rebound in the measured draft; no execution count is certified here. Seven rather than six selected static callbacks is the only required local workload amendment from this review.

Existing90-second cohort,120-second test/supervisor,63-second positive elapsed,5-second exchange and10-second preparation/cleanup limits should stay unchanged. The cited earlier CI headroom cannot predict the merged predecessor’s new cost. The plan correctly requires actual predecessor durations before release. No suggestion here increases runtime limits.

Current canonical bytes/nodes after the Event repair are738624 disk/750430 normal/19570 nodes; coverage108139, and the40 predecessor derivative maxima are756548/19726. These supersede the plan’s explicitly historical anticipated values only when it is reanchored. All new row and41 changed derivative measurements must be measured again; no allowance follows from this review.

## Optional simplification and conclusion

Keep the ten actions. Removing any recovery POST/GET would lose one member’s actual retained receipt witness; dropping replay or complete state/Event checks would lose the proposed no-effect or continuation proof. The already proposed removal of the impossible Changes action is the useful simplification.

Implement composition as one narrow local wrapper around the existing cross-carrier and new member observers, with exact handoff/replay ledger checks. Avoid a general observer dispatcher or production seam. Extend existing callbacks and preserve their negative weakening checks; do not add broad diagnostic cohorts.

After N1/N2 and actual merged-parent reanchor, this is a bounded evidence prerequisite consistent with Transactional-first priority. No new settled serving repair is established by this review, and no pending minimum/Event semantics, public discovery or profile admission is decided.

Read-only source/text/data inspection only. No product changes, tests, builds, services, imports, actual Claude launch or runtime execution performed. Report written under the required live-session guard.
