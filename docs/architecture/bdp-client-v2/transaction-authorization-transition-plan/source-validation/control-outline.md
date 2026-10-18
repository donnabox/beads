# Closed C1 emitted publication control

This is a preparation outline, not a created mirror or execution release. The positive build and combined source freeze are still required.

T1 imports `../dist/transaction-runtime.js` using a file URL inside its callback (transaction-runtime.test.ts368–370). Runtime imports emitted `./transaction-store.js` and `./transaction-service.js`; service also imports its emitted store. The private store object method at transaction-store.ts1903 is the sole mutation seam. Replace the complete emitted `replaceAuthorizationPolicy()` method with:

```js
replaceAuthorizationPolicy() { return enter(() => authorization); }
```

The derivative consistently retains old metadata, old policy generation/view, and old leases. It does not fabricate a new policy, suppress a test assertion, change service drain or alter source methods. Match exactly one complete method against the actual final build; bind positive/mutant bytes and full patch before root release. Preserve every other emitted file byte.

Copy source tests/support, all package source/dist and manifests, root Vitest configuration, setup files, lock/config/schema/fixture inputs, and a fully resolved local tool/dependency graph. The unchanged Vitest config scans all library package manifests and aliases @bdp libraries to the mirror's source. Its read-conformance setup is a source-only admission mock; it does not intercept the relative emitted store. The dynamic runtime file URL and every relative runtime/service/store edge must resolve inside the mirror. All @bdp package links—including virtual-store fallback links—and copied Vitest entry must resolve inside the copied graph; never allow a workspace fallback or blanket node_modules symlink into the positive workspace. Third-party deduplication is permitted only if inventory proves it cannot resolve a product edge into the positive workspace. Record symlink realpaths and exact source/dist/tool inventories, without importing code.

Run only `narrows current Transactional visibility and fences issued handles after policy publication`, with one Vitest worker, existing30s test bound and120s command bound. Expected seven HTTP dispatches (hard authored maximum17), one owner/directory/listener, zero product children. The seven prerequisites must succeed before the named `authorization-transition:generation-and-lease` assertion sees generation0/all-visible/canMutate=true/unchanged view. Retain the pre-cleanup stream observation showing it remains open and the safe response records. Expected raw exit1 and exactly that named failure; unrelated assertion, transport error, timeout, missing prerequisite, unexpected status or forced cleanup rejects discrimination. Cleanup must still close stream/socket/runtime and remove the known-stopped fixture store; sampled process/temp absence is an additional external check.

No control gate, derivative files, product execution, package preparation, or source changes are included in this source-validation packet.
