# P0 wire adoption — 2026-09-14

The current wire input is BDP `53bdbd03136875f952af184fce7b3c7af8f74e96`,
selected on Beads base `8af1e139770e3adbd81013558ce989a869ab99af`.
This is a Read DTO adoption. The complete upstream bundle contains 153
definitions; the upstream named Read projection selects 42 and accounts for
111 excluded definitions. Fifteen History Read-side definitions extend the old
27; exactly five old definitions change and 22 remain identical. Schema
support supplies no History, Read+Update, Transactional, serving or admission
capability. Provider mapping, installed Type contracts, graph ownership and
observed public HTTP conformance remain separate work.

## Current input identities

| Input | SHA-256 | Upstream Git blob |
| --- | --- | --- |
| `schemas/bdp-v0.schema.json` | `e20cfd088e836155f1a0ff00d764f8ad2e0a45e8429227bd1a1e7dcecf279ef2` | `a12889462395327ee48e1d2f262778b7edb6d802` |
| `fixtures/reference-domain/reference-domain.json` | `389dd6f72cd4ba2a5addee70060c23bc5eae8492b8fd04a79d04fbc9ddf30cb7` | `b84318b4fac80d88897b0a96136d2605ef243631` |
| `packages/conformance/fixtures/read-reference-v1.json` | `d02c97a1bdcf02cf366d894cc08c8c170806b3716521dfacd9a6ff35337614d4` | `9048a17b3044d613e0fce53e572c9d02c2ff38f2` |
| `packages/conformance/fixtures/read-bdpbd-v1.json` | `d97177f776d01b85d24b7f136879ceadc25ce0cadc751461d7eefd166adf78cc` | `c663ba72393dc575bf526495846f88aa291a2940` |
| `packages/conformance/catalog/read-v1.json` | `c1d2ef592fe0fe006f3368874a70d6113c67a4772bf79f4fcc7cb8c2018d91c2` | `b2a3f61556db0bb2c1b44604dd904ecd0c44eae5` |
| `packages/conformance/matrices/read-v1.json` | `ca175737db61eba6e013f429248088ed05ccae31e26f1eb4e30539bd2dfd62a1` | `9b8009941f092a27ee9e028e58e87b241d65a1f4` |
| `packages/conformance/src/schema-read-projection.ts` | `1b90533fc3e3ece5e3c627b06899918234efee4b441920a4517051aa5bf4e884` | `4aec522d8bad8f0f2429c7e5e1a68e8846d72e3e` |
| `packages/protocol/src/read-values.ts` | `ca91848bb179a8ea7e1dc05a777e12941ba94507700ed74e37fac4d77f413436` | `010f93a272ad2d3c138f88e86776dd25cb1d4c61` |
| `packages/protocol/src/history-values.ts` | `05b43bfd0bdfa9301b06688980b8833014b2509dc4f7d896fecf2cd3f391a2c9` | `64c0804913eba161cee2b4a7bfa9ac588d06159f` |
| `fixtures/history/wire.json` | `1336b221a50338b4ba1aa4b6d90ee440aaea1de5212541137556d531425df903` | `d4b81e4efd3843bc528d04f72b6d6c456d1282b5` |
| `docs/specs/bdp.md` | `f07d61947fc11cf00c5f062cbd64fdb251fa00651bc553479768db9372999d4b` | `2532f6f7a1761ba3954894bcdb5070675f4c8d36` |

The named projection digest is
`0feaa86a2ba5180d6396e1b52b0b2ee339b0a79a0650ecc0c0e6045b17d053e7`.
Sixteen derived roots reach 41 selected definitions; `protocolProfile` remains
sealed despite being unreachable. Both the ordered projection witness and its
support manifest reproduce through a joint gate over the complete bundle,
pinned upstream source declarations and current Read matrix. The projection
pair bytes depend only on the ordered seal and selected definitions; the matrix
contributes roots to the manifest and closure checks. The manifest never selects its own authority. No executable
TypeScript or network fetch is introduced into the Go package.

The current Read catalog and executable manifest contain 49 rows, with no old
ID removed. New rows cover Accept negotiation, conditional reads and HEAD
conditional parity. Existing internal-fault, malformed-response and positional
18-code problem-table plans changed. These are contract inputs, not 49 observed
Beads HTTP passes. The internal-fault plan now also requires public-http.

All 13 prior spec-example purposes remain, with current line ranges recorded
in [PROVENANCE](../internal/httpapi/bdpwire/schema/PROVENANCE). Eleven JSON bodies
remain identical; the two higher-profile discovery examples changed and are
still refusal controls. The 42-case History wire fixture is vendored verbatim:
22 selected Read examples round-trip, and 20 excluded examples remain explicitly
classified. Illustrations carry no History execution or capability evidence.

## Boundaries and checks

Hand-written DTOs remain the approved generator fallback. Current parity checks
bind all 42 selected definitions in both directions, preserve old members and
fixture floors, and add exact reference/restriction resolution, union branches,
booleans, integer constants, nullable members and conditional problem guards.
The historical Link Go type has the exact ordinary Link fields and tags, held
to the pinned schema alias mechanically, and its own strict decode/marshal
methods. Ordinary LinkRecord's encoding/json behavior remains unchanged.
Strict new codecs preserve three-state context and present empty messages,
missing-state cardinality/uniqueness, paired bounds and page structure. Multiple
current lineage rows remain valid. Date-time and URL/Type/provider admission
semantics remain outside transport shape checks.

The carrier guard rejects malformed UTF-8/unpaired escaped surrogates before
JSON decoding can repair values or member names, including raw properties and
extensions. Explicit null clears reused nullable destinations. Existing nested
error paths and the documented lenient ordinary-json legacy envelopes remain.
New strict Marshal methods reject malformed Go string bytes before encoding.

The source validation record is appended to
[the P0 verification document](BDP_P0_VERIFICATION_ROWS.md). Broader source council,
final repository gates and publication are independent gates; this adoption
record grants none of them.

## Preserved previous provenance

The 2026-09-10 adoption remains BDP
`19923f5bb6cc3f4ee4c508e36df3bd4c5c52344b`, with named projection
`b4c13b1d8e78bd556ace7db9c65729f86ea43428c069168bc3aba84bbe073d1a`.
Its complete source/fixture bytes are retained by Beads ancestor
`8af1e139770e3adbd81013558ce989a869ab99af`. The original P0 pin/results and the
2026-09-10/12 verification records retain their own source identities; none is
relabelled as evidence for this successor.

The exact previous PROVENANCE file follows (SHA-256
`0105ad62b0d3da45691fc7436c308ccd98137efd7035571f5cbaf3e09c593f45`):

```text
# bdpwire vendoring provenance (spec B8: the PROVENANCE file). pin_test.go checks every line below against the
# bytes on disk and the constants in schema.go; edit it only when re-pinning (GENERATOR.md, "Re-pinning").
#
# Header lines name the upstream repository, the pinned commit, the bundle's $id, and the identity of the one
# upstream file that is NOT vendored but derived from: docs/specs/bdp.md at the pinned commit, by its git blob
# sha1 (`git hash-object docs/specs/bdp.md` at that commit), so the derived entries below are anchored to a
# file whose identity is stated here rather than to themselves.
#
# File lines are four columns:
#   sha256  local-path  upstream-path[#Lopener-Lcloser]  git-blob-sha1 (or "-" for a derived file)
upstream: https://github.com/gastownhall/bdp
commit: 19923f5bb6cc3f4ee4c508e36df3bd4c5c52344b
schema-id: https://github.com/gastownhall/bdp/schemas/bdp-v0.schema.json
spec: docs/specs/bdp.md
spec-blob: 79049a703ef957e3eed7cbe56c093356c47b0658

# Vendored verbatim from the commit above (byte-identical; the blob sha1 is what
# `git hash-object` and the GitHub trees API report for the upstream file).
e4c4b7bebd75fe06cd4f7a39c5731c774bf23437dd7d01114819b436b624e4d3  bdp-v0.schema.json  schemas/bdp-v0.schema.json  494e3fe67c2929a91bf5e4d44df0ea119ef41d44
389dd6f72cd4ba2a5addee70060c23bc5eae8492b8fd04a79d04fbc9ddf30cb7  fixtures/reference-domain.json  fixtures/reference-domain/reference-domain.json  b84318b4fac80d88897b0a96136d2605ef243631
950de807247595e6d638cf9269ad5b4cd8f221d26d276b7b7262ea9492589b67  fixtures/read-reference-v1.json  packages/conformance/fixtures/read-reference-v1.json  6e932fc6facc92cb1330831f8605bedbf5f59332
339d6d10abe6cd6e8ca19cbf9ee05c4f5d2f145cc30577107e971f0e244782e7  fixtures/read-bdpbd-v1.json  packages/conformance/fixtures/read-bdpbd-v1.json  1221dfdab31887b3a956f1fa87131a848baff473
8867d1eed53f9f37047e11835c0a5a2021d9ff3140f64245cd38735f754d9924  conformance/read-v1.catalog.json  packages/conformance/catalog/read-v1.json  c78e13e14d32f985a1581b7c30b03e41f80cb6af
4e087ff545b514e9cc0608ffebc8ccd7deeb0a9baa2cdd67c5450a9ea9e0900a  conformance/read-v1.matrix.json  packages/conformance/matrices/read-v1.json  cc975aaa1d6fc87a57435356cb65da8da7012f79

# Derived: the JSON example fences of docs/specs/bdp.md (the file named by `spec:` above, identified by
# `spec-blob:`). Each file is the verbatim content of one ```json fence — the lines strictly between the
# opener and the closer — plus one trailing newline. The upstream column gives the fence's line range in that
# file, opener through closer inclusive; the file name's leading number is the opener line. pin_test.go
# reproduces every entry from a local copy of the spec when BDP_SPEC_AT_PIN names one (offline; never fetched).
d3350e0b9a9360d5315cb64733429a35c376be117c57ad8436c0383e85ec67ad  spec-examples/0583-scope-aggregate-constraints-1.json  docs/specs/bdp.md#L583-L593  -
38ea86fb9fa6cf3056d77d5ff2f11a4bb168bd50af5036e827c0a0046779ddcf  spec-examples/1662-scope-discovery-and-human-documentation-1.json  docs/specs/bdp.md#L1662-L1671  -
45ecab6de10fe3150f7e836c7a5feeda95fd5cb63d3913e2aff846f8d92fd1f9  spec-examples/1676-scope-discovery-and-human-documentation-2.json  docs/specs/bdp.md#L1676-L1686  -
88bcdc17076504dbfe3ccf6aa42d305a402e8bdd1bc301dfe6aea791783efdd4  spec-examples/1691-scope-discovery-and-human-documentation-3.json  docs/specs/bdp.md#L1691-L1709  -
b0c0447081a88d1345f4786edbf751c4f413ee50b125d76618a83b8088d07e52  spec-examples/1817-advertised-limits-1.json  docs/specs/bdp.md#L1817-L1830  -
77309a73a26160ed54e727b298885219e9de33197ab6bee2e0b58bdeb10c1379  spec-examples/1994-resource-records-1.json  docs/specs/bdp.md#L1994-L2005  -
8476638732420e3af34af8ef236ea717bcd00b6e00b0484a1eff521e9252c708  spec-examples/2009-resource-records-2.json  docs/specs/bdp.md#L2009-L2020  -
32b40661b3f30e6410026904ec350d941b070952cd1b9b6f63414367c56057ae  spec-examples/2044-resource-records-3.json  docs/specs/bdp.md#L2044-L2049  -
d72e834701a9dd9c43a2c41dab01ab02585eb9539a1b02153bd3af21faa695dd  spec-examples/2053-resource-records-4.json  docs/specs/bdp.md#L2053-L2058  -
e6dfc5f5d1cffac8d9dbbe872c74c3aa08e7d559c751f06934edab0c7421d5a9  spec-examples/2112-resource-views-1.json  docs/specs/bdp.md#L2112-L2137  -
7200e7698da791d5e63d7dfd83dabc75db362cf432c813ed6059edc1a0b333bb  spec-examples/2227-types-and-type-descriptors-1.json  docs/specs/bdp.md#L2227-L2243  -
4e1b54eebc5dd19a39ea758e2438dc8036eb296000754e3b0c458b660bf1303a  spec-examples/2269-types-and-type-descriptors-2.json  docs/specs/bdp.md#L2269-L2280  -
fb2dafdabe75945ad885a3a99bca714b5e0f63e0c65fd1027f228267d688493d  spec-examples/2286-types-and-type-descriptors-3.json  docs/specs/bdp.md#L2286-L2305  -
```
