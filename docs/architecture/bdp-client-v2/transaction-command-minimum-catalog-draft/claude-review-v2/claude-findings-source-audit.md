# Claude required-findings source audit

No product correction is required for findings1,2,5,6,7 after inspecting the missing source. They are supplied-context gaps resolved below, not observed failing tests. F2 warrants a fuller evidence inventory rather than asserting that every affected artifact is ungated. Root separately owns F3/F4 and optional wording; this report does not disposition them.

Basis: actual Claude stdout SHA256 `df9e921666cea5524cad204820779a62de2b76ee7cade71fb43c2d60953b0fb7`; frozen draft source binding755192feb51c3fb063a48f1d5189ee8194fac32a0a4ee5e3f49a556844d6c010 and diff6c0e6f47817d0072d2f258ec68c81c5e5e565d4a66169ff567c1e4a4837ab5fc. All source references below are relative to `/Users/dbox/repos/wt/vickie/vickie-command-minimum-catalog`, basebd2178dd6f4b9f7c2d239624d7edaa3d961c3e26 plus the frozen draft. Reviewer is independent of draft author.

## F1 — resolved: dynamic digest expectations, unrelated fixed configuration pin

The missing trailing line is a calculation, not a literal. transactional-cumulative.test.ts:87–89 and cumulative-artifact.test.ts:119–121 both assert:

    expect(actual.digests.catalogDigest).toBe(
      createHash("sha256").update(canonicalJson(selected.catalogDigests)).digest("hex"),
    );

(The latter uses bundle instead of actual.) profile-contract-selection.ts:76–80 reads each current source and assigns `digests[name] = digest(text)`; the full function begins at49. artifact-bundle.ts:241–245 computes cumulative catalogDigest with `sha256(Buffer.from(canonicalJson(cumulative.catalogDigests)))`. Both implementation and independent test expectation naturally move with the changed RU source. No stale literal requires updating.

Coverage/set digests likewise derive from source: test-support/read-update-artifact-set.ts:273–301 builds descriptor.coverageDigest from `digest(coverage.bytes)`, includes selection.catalogDigests and current shards, then hashes canonicalJson(descriptor). read-update-artifact-set.test.ts:335–348 independently reconstructs this descriptor and hashes it. cumulative-artifact.test.ts:119–145 checks current source digests/catalog order and reads current coverage for binding checks; it does not embed a historical coverage SHA.

The fixed3224c3… exactConfigurationDigest at read-update-matrix.test.ts:1263–1265 is a different identity. runner.ts:1580–1598 includes only scope, profile, ordered routes, request header/body limits and optional approvedSyntheticManifestDigest. It contains no catalog/coverage requirement inventory. This metadata-only draft leaves manifest routes and those configuration inputs unchanged. No pin update or new owned path is indicated. It remains correct to distinguish newly constructed catalog/set digests from historical evidence.

## F2 — resolved with expanded source/bound evidence

Search basis: repository-wide references to the two coverage filenames, read-update-v1 catalog path, profile-contract-planning.md and docs/specs/bdp.md; digest/sourceBytes consumers; `BYTES_BASE`, `byteLength`, `MAX_*`, `toBeLessThanOrEqual`, `toHaveLength`, coverage/report/progress length checks in catalog/profile/boundary/artifact and installed matrix files; exact historical bases648157/698614/703415. Searches excluded node_modules/dist so generated duplicates did not become new source authorities. No additional changed-file-specific byte/node authoring fence or report serialization ceiling was found. The relevant four coverage-authoring gates remain read-update-matrix.test.ts:4195–4202,4448/4585–4590,4965–4968,5111. Their measured overages remain243/50 only. TX coverage guards at transactional-cumulative.test.ts:919–920 concern unchanged test-support/transactional/coverage-v1.json, not the changed RU file.

There ARE generic artifact parser limits, so “ungated” would be too broad:

- catalog.ts:85–97 directly JSON.parses normative catalog text and validates shape; it has no source-byte/node/depth ceiling of its own. Scenario IDs retain the128-character shape bound. The new RU file84833→85899 bytes/1303→1313 value nodes is safe by static inspection, without a new per-file authoring allowance.
- When cumulative selections become bundles, artifact-bundle.ts:17–22,159–175,391–413,447–471 applies1,048,576 source bytes,100,000 value nodes,depth128,10,000 entries per container. The combined current core/companion catalog (duplicates merged as the inspected selector does) is184068 compact bytes,3707 nodes,depth5,max container269. Its new row adds955 compact bytes/10 nodes; previous totals183113/3697. This is independent JSON arithmetic, not a product-loader execution claim.
- RU coverage is copied with the ordinary fixture-role source cap (read-update-artifact-set.ts:87–92,151) and parsed by parseBoundedJson at212. Its713193 disk bytes,14632 nodes,depth6,max container2303 remain below every generic bound. The four more restrictive historical authoring ceilings are separate.
- Development coverage is directly JSON.parsed at development-read-update.test.ts:17–19; its sole current executable consumer validates inventory/rows/bindings at67–109, with no category enum or artifact byte/node limit. New49322 disk bytes/1146 nodes/depth6/max container82 are recorded; its full-file growth370 is not subject to another discovered authoring fence.
- Changed docs are consumed as text for citation sections, table/audit ID equality and profile dependency checks. No document byte ceiling was found. The spec change is one213-byte table row, and the planning doc growth is634 bytes. Their operative prose is unchanged.
- Installed RU report/progress checks at1266–1272,1284 onward,1363–1364,2024,2446,2877 are inventory/count or credential-nondisclosure assertions, not whole serialized report byte caps. Wire/body bounds concern actual responses, which this omitted catalog row does not execute. Existing required count changes remain separately covered by root's F4 disposition.

The generic bounds are unchanged and comfortably satisfied by static arithmetic; this does not predict all runtime checks will pass. Retain this expanded basis in the measured disposition instead of broad unqualified “ungated” wording.

## F5 — resolved: category is free text, single-member categories already exist

The only current development coverage consumer is development-read-update.test.ts:17–19,67–109. For omitted entries at78–82 it checks:

    expect(row.bindings).toEqual([]);
    expect(row.reason.length).toBeGreaterThan(0);
    expect(row.category.length).toBeGreaterThan(0);

There is no enum/schema/category dispatch. The previous23 omitted rows use profile-admission5,policy-principals8,owned-contracts3,controlled-scheduling3,retention-recovery-control1,schema-contracts2,disabled-capability1. Concrete single-member precedents are read-update.idempotency.restore/category retention-recovery-control at development/read-update-coverage-v1.json:1210 and read-update.http.cors-idempotency-key/category disabled-capability at1354. The new topic-specific category command-minimum-fields names an otherwise unrepresented omission and its reason clearly states no diagnostic scenario binds the command rule. Grouping multiple rows is not an existing contract; a forced reuse of policy/retention/schema categories would be less accurate. No source correction necessary.

## F6 — resolved: citation binder normalizes both sides

catalog.ts:175 uses:

    normalizeWhitespace(section).includes(normalizeWhitespace(citation.selectedText))

The helper at356–357 is:

    return value.replace(/\s+/g, " ").trim();

The anchor exists at docs/specs/bdp.md:3567 (`### Mutation command preconditions`), enclosing the cited3581–3588 paragraph. The single-line text is the exact whitespace-normalized paragraph. read-update-catalog.test.ts:30–33 invokes this actual validator against the current spec. No citation reshaping or extra precedent is needed; this is an existing explicit binder rule.

## F7 — resolved: exact two-shard authority already enforced

The unchanged test-support/read-update-artifact-set-v1.json contains exactly baseline→matrices/read-update-v1.json and reference-continuation→matrices/read-update-reference-continuation-v1.json. Loader read-update-artifact-set.ts:113–114 requires `input.shards.length !== 2` to throw “exactly two shards”;129–130 additionally requires two distinct IDs and sources. At143–153 supplied manifest source keys are checked against the index. Existing read-update-catalog.test.ts:72 and95–105 pins the same explicit two-name index with toEqual; cumulative-artifact.test.ts:106–116 supplies exactly these two files to the loader in an already selected callback. The new callback's two RU literals plus one TX manifest therefore exhaust the actual current RU shard inventory. Deriving the test list dynamically is unnecessary and would trade away an independent approved-path pin. No source correction necessary.

## Remaining work and limits

The five assigned required findings introduce no additional source paths, metadata semantics or allowance increases. Preserve the planned two measured allowance decisions and defer F3/F4/optional decisions to root. The full public profiles remain incomplete and this omitted row creates no executable credit.

No test/build/service/install/product import, product edit, commit or remote publication occurred. Read-only cat/sed/rg plus Python stdlib JSON/hash/URL-free arithmetic inspected source and measured current metadata; the session guard was used before writing only this report. Negative search results are scoped source inspection, not proof that arbitrary future code cannot add a bound. Exact-source validation and CI remain pending root-controlled work.
