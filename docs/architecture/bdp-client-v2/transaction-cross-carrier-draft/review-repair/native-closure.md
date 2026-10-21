# Native closure — cross-carrier review successor v2

**No remaining blocker.** The final successor implements only the accepted guard/report amendment and root dispositions4/6/8/9 in the two owned TypeScript tests. One Medium issue found in the initial successor's new digest check is now closed by the exact narrow correction below. Manifest/coverage and all modeled request/control data are unchanged. This is static closure, not executed type/schema/test/build proof or execution authorization.

## Digest correction found and closed

Initial repair app a9db5b61b6a8ca0a999d9d2e294ec4fb11dc7cd385188ab6f15a35089f5caa24 compared a derivative compact-byte manifestDigest with the original pretty-disk bundle manifestDigest. `artifact-bundle.ts:246` hashes the supplied raw bytes. Thus even an unchanged compact clone would satisfy that “changed” assertion. The derivative-to-derivative distinctness check was already sound.

Root authorized the smallest correction. Patch `canonical-digest-correction.patch`, SHA256 `1fa6dfe262f304d499947cc850a47d8d7ba1e19e0bbc99e2c68eb0af3df8a6e8`, adds SHA256(JSON.stringify(canonical)) and substitutes only that comparator. Both derivatives now use the same compact serialization form as canonical. Independently calculated compact digests are canonical `344c0c425f56075198e69e0ad44fcad004ee7cf25c8ee8966ef37add568ea87c`, batch `deba6872e8280be8823cf3c758757c397ee2238685de9ad970d2ef85c7fc7c34`, sequence `896d9c734284bc1c59dce02a68d4138c38ac71cefadee467c51d402fa3d2c569`. The unchanged-clone case would now compare equal and fail the intended assertion; all three actual data digests differ. These were standalone JSON/hash calculations, not product helper execution. Initial repair is preserved under v1/.

## Exact final boundary

W=`/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft`; evidence=`/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-cross-carrier-draft/review-repair`.

Parent remains eff171839fe4f37638284a4c3346428979d26b5d, tree5c14bcc20ec7a72118bc07d2ed5896703211a660. This is a frozen uncommitted draft, not a clean newly committed candidate. Exactly the original four paths are dirty; no extra path appears.

| Final file | Bytes | SHA256 |
| --- | ---: | --- |
| packages/conformance/test-support/transactional/manifest-v1.json |684206|726fc623bbb115a0f39e0bb3aaa586ea37254522b254e55e555ea95587f14cb9|
| packages/conformance/test-support/transactional/coverage-v1.json |99245|a3a15b5143c118b24965698ec10289821664c90e935820b51c8cf76c05bdb1df|
| packages/conformance/src/transactional-cumulative.test.ts |89418|1c361acc6edc460f1a224331d1010da9555a3700066b1ace41f75842aadf7bc8|
| apps/bdptest/src/transactional-matrix.test.ts |84322|5420e0bad74cc3906c72af02027b2de58c698e24c209e2fd9f67d013c5ca5ed2|

Final source-binding.json SHA256 `ae34ce221b13e6be85871e9cc4030d7f75f1c7dc396c6e82e52b07feaa972ddb`; complete final.patch `68cb0cbac92b770887a411cdd876dd9da84455bdc56dda67f4c61d2fc088d6a7`; amendment-against-original.patch `8727bf12f2cddb79029925eb36cd3e79d6a8b4d19d335e584d7d853ebe2a3111`; author report.md `ae56891a1b3cfb5e96c0610a69412d4e94533533d69d524ab84bf045b969ab4d`.

Recomputed all final source hashes/sizes, matched actual git diff bytes to final.patch, and reconstructed the whole amendment from preserved original before files to actual final files. Both JSON artifacts match the original reviewed bytes, so the38-derivative measurements,24 parent raw pins/reconstruction, row hash,230 new assertions,199 actions/2102 assertions and route preimage retain their original measurement basis. The reviewed source guards have not changed those data.

## Accepted changes and type feasibility

The already accepted staged guard/report amendment remains exact: manifest disk688302 and normal699694; coverage disk/normal100269; canonical18171nodes; derivative705812/18327; matching historical-base deltas200022/29607/5152. The genuine optional full-control writer remains `{sourceCommit, entry, ...observed}` with the accepted prefix and no new observation seam. Historical comments are retained.

**Claude4:** assertCrossCarrierControls constructs each actual changed bundle from compact JSON bytes and separately resolves every json-schema assertion in its loaded v2 scenario actions. Bundle loading is not represented as schema resolution. Existing factory signature accepts the manifest/fixture byte sources. Loaded actions/assertions have the exported discriminated union; after `kind === "json-schema"`, schema is a string accepted by `SchemaValidator.resolve`. Existing schema/createHash imports are reused. Each derivative differs from the corrected compact canonical digest and the two derivative digests are distinct. All these checks execute in the existing registered deletion/cross-carrier static case, without a child, request or new output variable. Runtime schema success remains for the authorized source test to establish.

**Claude6:** only the existing presence/equality test title and row filter expand. The new row's two client presence groups and eight response-metadata comparisons satisfy the current filter's authored data requirements. There is no new registration or assertion-data change. The full cumulative file remains the selected target, so its changed title does not invalidate the command selector.

**Claude8:** the existing positive cross-carrier exchange loop retains its all-assertions-pass check and adds2048 POST bodyOctets/8192 decoded response bytes. This reuses the existing exact-request/response observation types and neighboring conflict-row pattern. It adds no traffic and does not alter runtime request/response limits. The8192 bound is a fixture expectation awaiting observation, not a measured result.

**Claude9:** when the hook reaches its expected ordinary-wire count it now rejects all three exact new operation URLs, in addition to the unchanged carrier-key check. The later legitimate authored row uses none of those routes. It adds no hook invocation, capture, broad URL policy or extra response retention.

No obvious static type-signature or narrowing contradiction was found; this is not a typecheck claim. No source outside these two test files changed relative to the reviewed original four-file draft. No accepted control mutation, expected failure/positive set, raw request body, schema/catalog, fixture, production behavior or runtime deadline changed. Deferred findings3/5/7 remain explicitly dispositioned rather than silently implemented.

## Selection, workload and remaining gate

The two exact app-static titles in execution-preparation/commands.json still occur unchanged. The schema/digest helper remains folded into the existing deletion preflight. Static registration count is still39 cumulative cases; app remains59 expanded cases,52 observe invocations and the same nine installed selections. Source workload remains expected14/max24 loopback HTTP with12 listeners and zero installed children; installed cohorts remain expected1641/max1728 HTTP and189 children. No added weakening or service case appears.

Execution preparation must bind these final four source hashes, rather than the original draft or first repair hashes, before any command. Root's release, source checks, later clean source freeze, final package provenance and actual installed/CI observations remain separate. No tests, schema validator, typecheck, lint, build, install, service, pack, network action or repository write was performed by this review. Read-only text/hash/diff/count inspection and this closure artifact were the only reviewer actions.
