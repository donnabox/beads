# Native review — sequenceObserved deep freeze repair

**No actionable finding in this bounded two-file correction.** Reviewed the complete correction and surrounding service/projection/fixture behavior against HEAD `1ba65f2f192cc84e44eb09ef89b9693ec6cdaf05`. No product commands, tests or source edits were performed by this reviewer.

Binding SHA256: `ba25cca4d9eecd1d9e8d3e1343fb021f71b44287af513fe4b72207381f649ed1`.
Correction SHA256: `581be6bb5f53d6e132a31f46c8aed5305a3a4a8c49984e9984887a132abeea90`.
The actual two-path Git diff matches that patch byte for byte. Both predecessor snapshots match HEAD. All eighteen previously frozen paths retain their recorded hashes.

Final source hashes:
- transaction-service.ts: `4faf6b2187cacd2c50fe8437819e60b4c216f3e8485fff3251f5ed9058f7645d`
- transaction-service.test.ts: `32010b6fb1107076534c544f2bbb93e62cc39c89cd0fbc3c482f0736cc28cfe2`

## Production behavior

The change at service lines643–655 replaces only the local recursive freeze. The explicit stack visits every nested JSON container, including descendants beneath an already-frozen parent, and freezes each object without rebuilding values or skipping shared references. Freezing parent before children instead of after children is equivalent for these internally projected, acyclic JSON values: there are no caller getters/proxies or new mutations between visits. Object.values is local and nonrecursive. No broad spread of the pending graph or depth-dependent matcher is introduced.

The projection still occurs in the same owner.withRead callback after sequence execution, captures the same current authorization and context, and returns the same frozen outer wrapper. Member order, retained-key handling, committed effects, error mapping, lifetime ownership and limits are unchanged. This fixes the actual recursively freezing service presentation path, rather than weakening a depth limit or changing what sequence execution commits.

## Regression and typing bridge

The new callback at lines693–771 exercises the previously missed sequenceObserved API with one real owner/service and a literal12000-level array, then the same principal/carrier/key again. Independent iterative comparison checks every key/container/primitive leaf and complete expected created result/Resource; only the emitted opaque revision is taken from the real response. Shallow sorted key lists passed to matchers do not contain the deep graph.

The freeze check deliberately traverses identical objects and verifies every reachable observed container, including the outer response, body, entries, Resource and deeply nested arrays. Full replay equality, same context/position, identical inspect inventory/resources/groups/identities and unchanged Event page establish the replay's lack of new effects. The Event reads are actual source API calls, not HTTP. This is one additional callback/owner/directory and zero HTTP; no reopen or installed-payload claim.

The final typing correction is narrow and sound: after the one-result length assertion, bind the first projected entry as Record<string,unknown> and then read resource. The old optional property access was ill-typed because projection inferred an empty object shape. The assertion establishes the index for this observed plain array; the subsequent independent full JSON comparison checks its actual shape. This does not turn an invalid output into a pass or erase a runtime check.

The existing fixture registers directory cleanup before open, owner cleanup after open and service cleanup after service creation; reverse awaited cleanup closes service/owner before removing its store. It is not a generic guarantee that cleanup continues after a thrown close, and the supplied run records—not that assumption—establish successful cleanup here.

## Actual evidence and limits

Read the actual result/exit records: red has the unique new callback failing with RangeError at the old source freeze646; green has1 pass/67 pending; complete service suite has68 passes/0 failures. The original typecheck TS2339 failure is retained; successor typecheck and targeted formatting pass. The full runtime pass belongs to pre-typing-bridge test bytes. The final bridge changes only the local access/type expression; no claim of a post-bridge runtime rerun is made.

All seven recorded stages report stable source, no forced cleanup, empty remaining groups and no runtime directories remaining. Independently checked all16 sampled PIDs absent. These are source-alias runs; they do not replace the fresh archive/consumer thirteen-request installed regression that exposed the service fault.

No further correction is recommended before the separately required actual code-review disposition and fresh packaged validation. Prior failed installed request7/500 and diagnostic/setup failures must remain distinct historical evidence. This review does not approve publication or claim full profile qualification.
