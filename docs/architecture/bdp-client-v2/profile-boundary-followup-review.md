# Independent boundary follow-up and integration review

Reviewed frozen author `9bf885826a1a74c6e387856611de1926e8371b8d`, direct child of
`75d1db9174421c360bf2658aae1e5674e23b326b`. No remaining blocking findings.
This supersedes the three validation blockers recorded in
`profile-boundary-independent-review.md` for the earlier head.

## Exact regressions and alternate routes

Reran the original independent repro inputs against the frozen repair, and again
against the integration candidate. The external driver imports the built private
Markdown module and extracts the exact ownership helper source using Node's type
stripping; it does not alter repository files.

- A heading alias of the shared Normative conformance matrix now inherits both its
  Read rank and its direct excerpt bounds. The nested Transactional-table quote
  is rejected. The allowed direct matrix introduction still succeeds.
- Explicit/generated collisions now fail closed during indexing, before either
  citation text or profile ownership can choose a different occurrence. The
  collision inserted into the actual spec keeps its exact Part inventory and is
  rejected. Explicit/explicit and reversed-order explicit/generated collisions
  also reject. Catalog validation reports the failure at the citation anchor.
- A fenced block between an explicit anchor and a later heading no longer counts
  as empty adjacency. The anchor retains its interior extent, and the later
  heading/subtree is excluded. Matching fences still hide fake heading/anchor
  definitions during index construction.
- Both id and name heading aliases retain the shared restriction; duplicates of
  either spelling reject consistently. The current repository's 144 indexed
  anchors continue to match all 144 IDs in the saved GitHub-rendered Contents HTML.

The changes stay within the current ATX/standalone-anchor utility; no general
Markdown parser or public API was added. The previously reviewed optional History
discovery and no-op storage-cost clarifications remain unchanged and sound. All
capability, successful-key tombstone and failed-disposition rules are preserved.

## Ordinary merge candidate

Frozen clean head: `3a8f9d553a2944bc7d1d7812a284f0160dcbe798`.
Parents:
- Published diagnostic/routing integration `e87d4497ac34f7fcd0f29f9cb81bbd467a6a28ae`.
- Reviewed repair `9bf885826a1a74c6e387856611de1926e8371b8d`.

Checkout: `/Users/dbox/repos/wt/vickie/bdp-profile-boundary-final-review`.
Branch: `codex/vickie-bdp-profile-boundary-integration`.
Origin: `https://github.com/gastownhall/bdp.git`.

The ordinary merge was conflict-free, with no resolution edits. Relative to the
published parent, exactly five reviewed files change: spec prose, catalog citation
validation/tests, the private Markdown utility, and the profile ownership tests.
No schema, server/runtime, diagnostic plan, package export, checked-in evidence or
profile admission change is included. Published routing/alias/pending work and all
Read evidence ancestry are preserved. Author checkout remains untouched.

## Independent checks

Pinned Node 24.16.0 / pnpm 11.20.0, frozen install, build before process tests.
- Exact author: 110 tests across nine artifact/catalog/boundary/inventory files.
- Combined candidate: 295 tests across 14 focused files, including development
  matrix/server/disconnect, sequence executor, SDK transport and all targeted
  catalog checks.
- The development report remains 25 passing plans / 264 actions / 827 passing
  assertions, claimEligible false. Inventory remains 54 bounded / 27 omitted.
- Candidate build, typecheck, dependency boundaries, format, lint, installed-package
  smoke, development smoke and evidence:verify all passed. Lint has the existing
  three warnings and one informational finding only.
- Read evidence verifier confirms seal 6142cf73deae687bc3e9668cc014ea408fa02843,
  run head 4f0ed57906525c36bc55b343b77e0f07edbb7553, evidence commit
  72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5 and 89 target rows.
- git diff --check passed; working tree clean after ordinary merge commit.
  No broad full-suite run, remote CI or outward writes performed.

Logs, exact repro driver/results and reports:
`/Users/dbox/repos/wt/vickie/bdp-integration-logs/profile-boundary-followup/`.
Root owns any subsequent publication/integration decision.

## Published full-suite verification

Root verified published PR83 head `3a8f9d553a2944bc7d1d7812a284f0160dcbe798`
and [Linux CI38111288272](https://github.com/gastownhall/bdp/actions/runs/38111288272).
The run passed 115 files / 7,743 tests with one optional skip, seven preflight
tests, installed-package smoke and the actual `evidence.verified` event.
The event retains seal `6142cf73deae687bc3e9668cc014ea408fa02843`, observed
run head `4f0ed57906525c36bc55b343b77e0f07edbb7553`, evidence commit
`72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5` and 89 applicable target rows.
This is exact-head CI evidence, not full write-profile admission.
