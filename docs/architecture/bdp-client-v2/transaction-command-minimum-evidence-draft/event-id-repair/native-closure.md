# Native closure: Event-ID and header-diagnostic successor

No blocking finding in this narrow successor. The previously incorrect stored-ID expectation is replaced with an exact private served-ID oracle, while the manifest keeps the remaining ordered Event facts. This closes the source defect identified in event-id-source-audit.md; it does not constitute test execution or public qualification.

Reviewed source: ordinary base 453b6dab386746bb03aac2c0a293572b8a71c969; final five-file diff SHA256 ea04e6ba34c9e1d9b724348f32407c6fd6bd78b795cad0481e7f4093b0f21762. Compared the preserved event-id-repair/before snapshots, complete correction.patch 623a7545ec1b160efa53cecb4cae8a9473e419d53071075eb24f8cc9f6d1d0dd, current files, current renderer and frozen measurement specimens.

## Source closure

- App lines 2510–2523 independently construct the private served ID from validated current epoch/view, fixed fixture Event source, literal checkpoints pos-1/pos-2/pos-3 and ordinal zero. The SHA256 tagged JSON/base64url construction matches transaction-events.ts:73–100. It neither derives expected IDs from received Event IDs nor imports the renderer. This is a source-bound private oracle, not a new normative wire spelling.
- Only the incorrect /id column and three stored-ID literals leave the new manifest ordered-facts assertion. Source, type, subject, subjectType and ordinal remain ordered and exact; schema, Event data/revisions and full Resource comparisons remain. A/B Event-to-receipt transaction correlation remains; C has no receipt baseline and no three-transaction distinctness claim.
- Coverage explicitly attributes exact private projected IDs to the app supplement. The manifest retains 11 actions and 192 unique assertion IDs with the exact coverage mirror. Its row compact digest is 4344f3701bc99031485bc4f7e327f738bc138084ff10e43c1623cdfa57d068b6.
- App lines 2364–2367 check exact-case names before comparing authored name/value arrays. A future shifted credential header therefore fails a names-only comparison before values can enter this diagnostic. The authored-prefix equality, Boolean prohibition on BDP fields in generated suffixes, complete runner request assertions and explicit whole-json mode guard remain.
- The report-before-cleanup callback repair, literal assertions, route grants, capture boundary, selectors, HTTP/child budgets, time limits and all approved size fences are unchanged. No runner/runtime API or production file changes occur in this correction.

The final cumulative hash differs from the author's 6ca35fc3b9d7d35b9f0fc7e513cf1eeb3b95a81a7e772755432dc435de346f9f only by the independently inspected arrow-expression layout at lines 2699–2701. The mapping and other four final hashes are unchanged. No data specimen changed with that formatting.

## Independent static recount

Parsed all 40 retained derivative JSON specimens without importing product code or running their tests. Maximum two-space JSON plus final newline is 756548 bytes; maximum node count is 19726. Canonical disk, normal form and nodes are 738624, 750430 and 19570. Coverage disk/normal bytes are 108139. These agree with retained measurements.

| Fence | Actual | Unchanged limit | Headroom |
|---|---:|---:|---:|
| Canonical disk | 738624 | 742879 | 4255 |
| Canonical normal form | 750430 | 754685 | 4255 |
| Canonical nodes | 19570 | 19639 | 69 |
| Coverage bytes | 108139 | 109115 | 976 |
| Derivative normal form maximum | 756548 | 760803 | 4255 |
| Derivative nodes maximum | 19726 | 19795 | 69 |

Reconstruction by removing only the new first raw object and its separator reproduces both complete parent files byte-for-byte, preserving all 26 historical raw objects:
- Manifest parent/reconstruction SHA256 726fc623bbb115a0f39e0bb3aaa586ea37254522b254e55e555ea95587f14cb9.
- Coverage parent/reconstruction SHA256 a3a15b5143c118b24965698ec10289821664c90e935820b51c8cf76c05bdb1df.

Recounted controls.json against actual derivative specimens and action assertions. The ordered failed/passed partitions are complete, disjoint and unchanged in meaning:
- Batch body conflict: action m-replay-b, 21 assertions, expected 4 responses/2 children; compact manifest digest 2b271f5df6c633ae6eadcef002df19aabd73f3c6db49e7bef680aa070d1fc9cb; normal bytes 750437, compact bytes 366123, nodes 19570.
- Read-only credential: action m-create-a, 20 assertions, expected 1 response/2 children; compact manifest digest 8e5e40ac289daa764e5e38c7f7f18d936fc7a8cfeabf4543786c72c6f8c75e71; normal bytes 750428, compact bytes 366114, nodes 19570.

Updated control metadata SHA256 is c13e9fd4c286a900f9ba46d2087957e98a9adfd419130f646fdc048b9203a093. These are static expected oracles, not observed failures. No separate external adjudicator update is claimed.

## Final source binding

- packages/conformance/test-support/transactional/manifest-v1.json: 29729151cf6da641545910ec3880f9609c13b319af7bbe99d3d2ad47833dab96
- packages/conformance/test-support/transactional/coverage-v1.json: 4619813e9446e0e714a64c11eac184531950f57720e32c7ef414c7e3342121af
- packages/conformance/src/transactional-cumulative.test.ts: 0d1f87199222e2426b048eae1a5aadd11c20c04298d31e3198ca46897e5492d3
- apps/bdptest/src/transactional-matrix.test.ts: e319109ab0e9f7258e6918ae789ed7a5672eb7aabd0ec8725ae38e3634164e68
- packages/conformance/src/profile-contract-plan.test.ts: e5be4864e456431db74bd325ec64d785bf3c5301a6be5b230bfcc7351ade4fc4

## Limits

This is the independent read-only closure of this correction and its static measurements. It preserves the earlier whole-contract false/RU omission and private-diagnostic limits. No tests, builds, imports, services, installed execution or Claude launch were performed for this closure. Dynamic raw-to-client action-9 epoch/view equality is still outside the demonstrated boundary; original-wire receipt equality uses the existing private helper. Root retains execution-release authority and the separate actual Claude successor review.
