# Internal payload symlink refusal

Frozen commit: `4cf6acd03774fd90e629ba63516ff1365865b00e`
Base: `39a76748329fa88eee6cd1ca31b9501c5ef7f395`
Branch: `codex/vickie-update-payload-symlink`
Checkout: `/Users/dbox/repos/wt/vickie/vickie-update-payload-symlink`
Origin: `https://github.com/gastownhall/bdp.git`

Small repair for the native review's executed-but-unbound fast-uri symlink reproduction. Runtime traversal refuses symbolic links encountered within a package payload, including directly read package manifests and payload directory roots. Dependency package roots still resolve through Node normally, supporting pnpm/installed package-root links. Ignored node_modules trees are reached only through declared dependency resolution; no generalized symlink following or attestation framework.

Two files changed: private installed factory and its process tests. Existing isolated copy helper is reused. Three real-process negative cases replace the copied fast-uri utils.js, lib directory, or package.json with an internal symlink to a disposable external target. Each launch refuses without readiness; its store path remains absent and its configured listener port can be rebound. Normal process tests retain the protocol package-root symlink, regular-file byte-change identity discrimination and actual cold-client behavior. No production/shared dependency modified.

The refusal precedes the factory's listener/store startup. It is not a sandbox or a claim to prevent module initialization code from executing before the factory is called. It prevents an unsupported linked payload from returning a misleading qualification identity/readiness.

Validation (Node24.16.0 / pnpm11.20.0, maxWorkers2, external logs here):

- Frozen install, build PASS before process tests.
- Six focused files /54 tests PASS,7.42s (`focused.log`, `focused.json`): installed-reference, startup, import, closure, development-discovery, development-matrix. Includes all51 native review tests plus3 negative link cases.
- Typecheck, boundaries, format, lint PASS (existing3 warnings/1 info).
- Packed smoke and development smoke PASS, including normal package-root links, private inventory/deep-import denial and public Write refusal. Packed payloadIdentity `0c2782d51a18aa31c3c464c866b3d4e8f4a01a1411ab1808861f97b3357aa902`; claimEligible:false, all9 targets and stable Scope.
- Existing Read evidence verification PASS, unchanged89 rows.
- git diff --check PASS; commit complete, clean tree.

No failed/interrupted tests, no full repository rerun, no public admission/evidence changes, no new capability or store/evaluator changes. Predecessors39a/dea remain frozen. No outward writes. Full cumulative eligibility and deployment guarantees remain unclaimed.
