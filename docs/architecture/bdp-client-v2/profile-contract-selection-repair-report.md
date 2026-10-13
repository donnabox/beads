# Native configuration-boundary repairs

Frozen successor **9ea408958e08be38cc4e586a4846caf3d6ef6874**, parent **abdd1d4af45cd02ad1238b311462a1b154e44b79**. Same clean isolated checkout `/Users/dbox/repos/wt/vickie/vickie-profile-contract-selection`, branch `codex/vickie-profile-contract-selection`. Prior reviewed commit preserved without rewrite. No outward actions.

Both native P2 findings addressed:

- Reject proxy/revoked-proxy and non-plain configuration records before reflection. Snapshot own descriptors once; require exactly the closed own enumerable data fields. Accessors never run. Subsequent validation and output construction read only the detached captured record, never spread or reread caller input.
- Require primitive strings for commit, digest, Scope and fixture identity inputs before regex, URL parsing or canonical equality. Coercible or boxed objects cannot serialize to indistinguishable empty objects. Profile and History remain strict scalar selections.
- Nested capabilities require a plain dense array with own primitive data elements. Proxies, getters, holes, subclasses, extra keys/custom iteration and object elements reject without invocation. Fresh array storage is sorted/frozen; the caller record/array remains unfrozen and later mutation cannot alter the binding.

Added all three native reproductions as rejection regressions: changing sourceCommit getter, changing ownKeys proxy introducing token, and two unequal coercible policy digest objects. Added nested fixtureId/fixtureDigest/capabilities accessor controls; boxed values for every string field; revoked proxies; capability proxy/index getter/custom iterator/hole/subclass/extra-property controls; symbol/nonenumerable/inherited record members; primitive/capability alias mutation and null-prototype positive case. The suite verifies zero invoked getter/proxy/coercion callbacks where relevant.

Validation with pinned Node24.16.0/pnpm11.20.0:

- PASS typecheck, build, dependency boundaries.
- PASS focused serial suite: **394 tests / 11 files**,4.23s; same original eleven-file regression scope, including43 planner tests.
- PASS Biome checks and git diff --check.
- PASS unchanged sealed Read evidence verification:89 rows, constant674982be2123b9b7aae630bf79c3f774122900cd; no generation/artifact edits.
- Initial typecheck caught TypeScript's mapped-array descriptor length typing; corrected the reflection input typing without changing runtime behavior and reran all gates above.

Logs: `profile-contract-selection-repair-{focused,typecheck,build,boundaries,evidence,format}.log` alongside this report.

Limits unchanged: source-private planning only; binding proves shape/equality, not truth of captured installation observations. No public admission, runtime behavior or evidence format changed. No full workspace suite/served qualification claimed. Awaiting native independent closure and external Claude supplement before integration.
