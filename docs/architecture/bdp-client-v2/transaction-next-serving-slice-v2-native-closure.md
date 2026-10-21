# Next serving slice v2 — native closure

No actionable finding remains in the native P3 clarification. Exact plan-v2 SHA256 **a55f40fb886452a43587468935c00d6ac557f206adb640ab89dd6543d7ef4439** differs from frozen originald245c319 only in the service-capture test bullet.

The clarified positive comparator leaves the original signal active, replaces the caller object's signal with a pre-aborted signal while the first wait is held, commits a real Resource effect satisfying the original minimum, then releases the wait. It requires success at the new eligible context with the expected second owner entry. This discriminates captured options from rereading the mutated caller signal on retry. The separate original-abort/no-next-entry case remains.

The clarification changes no production scope, API, wire semantics, request/child/deadline bound or planned source counterfactual. The prior native review otherwise remains applicable, including existing direct-vs-service lifecycle differences. No tests/builds/network/source changes were performed; only this review report was written. Root owns plan acceptance and later exact-base authoring authorization.
