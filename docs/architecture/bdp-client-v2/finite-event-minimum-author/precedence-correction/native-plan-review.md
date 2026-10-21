# Native review of first-attempt capture correction plan

Approved as a bounded correction plan. No actionable finding or additional product decision. Root still owns authoring/execution release.

The proposed call-local captured flag and minimum value move selected lexical capture into the first unchanged-waiter attempt, inside run. This restores the actual baseline ordering: synchronous run.check; waiter policy validation; initial clock/deadline validation; abort check; selected capture; owner access. No duplicated waiter checks or error policy is introduced. First capture executes synchronously before the first await and subsequent attempts use the frozen tuple, preserving B's caller-mutation discrimination and the original deadline. Capture TypeError is not retryable, so leaving captured false on capture failure is harmless. Undefined/sentinel/excluded paths remain unchanged.

The five rows discriminate the discovered precedence defect without new wire traffic: aborted yields ScopeServerOperationAbortedError after one clock read; invalid budget yields the exact policy TypeError before any clock read; throwing clock preserves sentinel identity; NaN clock yields the exact clock TypeError; valid preflight then malformed tuple yields the exact capture TypeError. The malformed epoch is structurally typed but lexically invalid under the real snapshot helper. Each row observes one returned promise immediately and expects zero owner reads/waits. Four total initial clock calls is correct. Isolated fixture/service ownership avoids sharing an owner among multiple closing services, and the existing reverse cleanup order closes service then owner/store.

Five added direct callbacks/stores and five events calls are explicitly budgeted. A–E remain byte-preserved, with original32+3 HTTP and deadlines; overall six-file scope remains six, correction touches only service and its test. Existing regression traffic remains separate. No need to alter production owner/HTTP/waiter/parser/installed code or add a public seam.

Verified all679 current source hashes against this correction binding, unchanged HEAD b76c3b7df77ea9b33611bc9027db651833e20abb, and predecessor binding/patch/native-finding hashes. Original draft remains frozen.

Pins:

- plan.md: c2767bd47aa3740c52eafd5d277316d26d56b94af295b6761ff126a9ee2e0d5f
- binding.json: 5b3b3f5bf3b654a57ea8db758aaf809d75f11300c253db1a360dc169a2e8320d
- original source-binding.json: e439fb06b06dabf70445688412f1a30388298bc0880cbad59a4ccd547ed4c764
- original draft.patch: ee6d34ccc8d253f845e1e7d07c469220883698382f788f69566291b6623fbf2d
- original native-code-review.md: 89df19d3ed51f330de465e7c964e4fcec65a8994fd5b4836b2b9d5c7d72310ef

No tests, product imports, builds, installs, services or source changes. Only this guarded review artifact was written.
