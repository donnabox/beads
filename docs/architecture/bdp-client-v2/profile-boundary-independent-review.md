# Independent boundary/clarity repair review

Frozen reviewed head: `75d1db9174421c360bf2658aae1e5674e23b326b`.
Base: `e538e23b6a4a7371b14945f59009323c302e60a3`.
Isolated checkout: `/Users/dbox/repos/wt/vickie/bdp-profile-boundary-review`.
Branch: `codex/vickie-bdp-profile-boundary-review`.
No source edits, new commit, merge or outward writes. Author checkout untouched.

## Blocking findings in validation commit 6c1cf50

1. **Shared-parent restriction is lost through an explicit alias.**
   `packages/conformance/src/profile-catalog-boundaries.test.ts:53` applies the
   direct-text restriction only when the literal anchor belongs to sharedAnchors,
   while lines 40–41 propagate the parent rank to any explicit alias. Insert
   `<a id="matrix-alias"></a>` immediately before `### Normative conformance matrix`
   in the actual spec. The alias's section contains “These rows name the
   Transactional obligations.” and its ownership is Read. assertProfileCitation
   accepts that higher-profile quote as Read because the alias skips the
   direct-text restriction. Preserve the effective target's shared restrictions
   as well as its rank, and add this regression.

2. **Duplicate explicit/generated names select different section and owner.**
   `packages/conformance/src/markdown.ts:62` uses the first matching anchor, but
   the ownership map at profile-catalog-boundaries.test.ts:38/41 uses later
   assignments. Insert `<a id="transactional-discovery"></a>` immediately before
   the shared Normative conformance matrix in the actual spec; its exact Part
   inventory is unchanged. Citation text lookup selects the original Transactional
   discovery section while ownership resolves to the later shared Read alias.
   The exact quote “The Transactional-profile discovery representation adds the
   surface for” is accepted as Read. Reject ambiguous explicit/generated anchor
   collisions consistently rather than selecting different occurrences.

3. **Fenced content is treated as empty adjacency.**
   `packages/conformance/src/markdown.ts:49` tests the masked string between an
   explicit anchor and the next heading. A real intervening fenced block becomes
   whitespace, causing an interior anchor to alias the following heading and
   acquire its entire subtree. For example, a Read anchor followed by a text fence
   and then `## Part III — Transactional` incorrectly includes the next Part's
   rule. Use original intervening content for heading-alias adjacency while still
   using the masked index to ignore fake headings/anchors inside fences.

These are bounded fixes to the existing advertised anchor/ownership behavior, not
requests for a general Markdown parser or new API. The exact external repro uses
built markdown.js plus the unmodified ownership functions extracted from the test
source with Node's TypeScript stripping. Script and output are in
`profile-boundary-independent/reproduce.mjs` and `reproduce.log`.

## Sound portions / prose review

The current corpus anchor inventory independently matches the saved GitHub
Contents HTML: all 144 rendered IDs and all 144 indexed IDs match, with zero
missing/extra anchors. The Part slug keeps the two spaces surrounding the removed
em dash as two hyphens. Generated duplicate headings and literal numbered slug
collisions work in the covered cases. Matching code fences hide synthetic headings
and anchors, ordinary subtree boundaries stop at equal/higher headings, and exact
Part names/top-level inventory are enforced for the current corpus. No actual
catalog citation in the current spec was found crossing the profile boundary;
the blocking findings concern the guard's intended future regression protection.

The separate prose commit `75d1db9` is sound. historicalResolution was already an
optional complete Read capability inherited without becoming required; all three
unchanged closed discovery schemas admit historyCapability and omit it from their
required arrays. The optional/optional/optional table row and distinction from
Transactional history fences reflect that existing rule. The no-op storage-cost
sentence now acknowledges that distinct successful keys can grow retained state
without new revisions. Successful no-op tombstones, immutable replay/conflict
identity, failure forgetting and logical-Scope replacement rules are unchanged.
No schema, capability or retention behavior is changed by this prose.

## Independent validation

Pinned Node 24.16.0 / pnpm 11.20.0; frozen install and build completed.
- 106 focused tests passed across nine artifact/catalog/profile/inventory files.
- Typecheck, boundaries, formatting, lint and evidence:verify passed. Lint retains
  the unchanged baseline three warnings and one informational finding.
- Read evidence seal remains 6142cf73deae687bc3e9668cc014ea408fa02843, run head
  4f0ed57906525c36bc55b343b77e0f07edbb7553, evidence commit
  72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5; 89 target rows verified.
- Three independent adversarial repros expose the gaps despite the passing suite.
  No full repository suite or remote CI was run for this review.

Logs: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/profile-boundary-independent/`.
Checkout remains clean at the frozen reviewed head. Root has assigned all three
repairs to the author; follow-up review is pending a new frozen head.
