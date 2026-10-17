# Native code review of actual staged reconciliation

Verdict: no actionable merge-resolution blocker found. The staged result preserves both parent intents within the accepted scope. This is code/tree review only; formatting, typechecking, runtime, installed validation and exact-head CI have not been run by this reviewer and are not claimed here.

## Exact reviewed state

- Worktree: /Users/dbox/repos/wt/vickie/vickie-pr150-main-reconciliation.
- HEAD/first parent: d4f3f4be88c7837b71d6646518ca3589d464ba3d. MERGE_HEAD/second parent: 8d6671ed19b0c0406428ccc2d2161768d01e8b9b. Common base: 9df574a9ac980272fa31be9f443a771233b69c6a.
- No unmerged index entries and no unstaged changes. This remains a staged merge, not an already-created merge commit.
- Complete701-path staged path→blob map SHA256 (sorted JSON): f716d2df944c970d989e919eb408e1cfdda6b64fc1061de053db4365c7b7c185.
- Binary staged diff versus HEAD SHA256 fdfbe85a69804c5319b0228ae2bfb2309abddaf8453983de3890e7b6f3049b89; versus MERGE_HEAD 06727ea5e0ec9670bd7f955236e85c390fa5eec542deb302d19c5d30be2cba13.

## Preservation and scope

Independently compared every staged blob with both parents and their common base. There are9 overlapping changed paths,91 incoming-only paths and26 d4-only paths using physical delete/add accounting. Taking incoming for incoming-changed paths and d4 otherwise yields exactly13 deliberate deviations: the eight overlapping test files retaining d4 behavior, the four candidate-only schema readers, and Write's normative transplant. No undeclared path differs from that union. Old monolith navigation is exact incoming bytes. All main-only relocation/helper/layout/draft files are preserved, with Write as the explicit planned exception.

Read the complete staged differences for all12 affected tests. Independently reconstructed each from d4 using only the old→Core schema path substitution and, for six JSON.parse/readFileSync expressions, the exact displayed line wrapping/trailing comma. All12 reconstructed texts equal staged bytes. Thus callback bodies, oracles, selectors, deadlines, requests, fixtures, lifecycle controls and prior conditional/Event evidence behavior remain d4's. This covers collection-minimum, depth-key-order, matrix, cumulative, HTTP, sequence, set and Snapshot overlaps plus Event SSE, fresh-depth, minimum-current-resource and sequence-depth hidden readers.

All118 production TypeScript files under apps/packages src remain byte-identical to d4. Every package manifest, sealed conformance catalog and docs/design/evidence file also equals d4. Production exports/discovery claims are unchanged. The incoming citation helper and evidence-generation/verification path relocation remain incoming, not newly invented runtime behavior. The workflow keeps d4's codex/vickie-* CI route. No newer-main or downstream serving-slice content is introduced.

## Normative transplant

Derived the actual common-base→d4 monolith changes independently with a line diff: exactly two replacement hunks (80→75 bytes and279→1310 bytes). Each old hunk occurs once in incoming Write. Replacing those two spans produces the staged Write file byte-for-byte. There is no extra normative editing, duplicate block or heading/map change. Staged Write SHA256: 6df5a0f07f9cee293685a0aee0ee0300b0f575d69b668f035da2044706c77d8c.

The rule retains lexical validation precedence, ordinary missing/deleted/hidden Resource/properties/alias/initial view=links-anchor404, actual current refusal context, and bounded unknown-position wait→503 for existing authorized current Resource/properties reads. It explicitly leaves retained continuations/Snapshots and successful incident-Link lists outside that extension. The exact whitespace-normalized sealed sentence “It never reports success with an older position.” remains inside the Transactional HTTP consistency section. The historical catalog bytes/source identity are unchanged.

The new Core schema equals d4's old root schema and the packaged protocol schema byte-for-byte, SHA2560baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227. The old physical root schema is absent. Public schema $id and package distribution path remain intact. Navigation SHA256d38501c7cd5ad220b747b403e5acd152b70b3ddd7d30aeb60f2a09b49a480985 equals incoming; citation helper SHA256ebaad4ee8e797375ad3186ab54feb80d0bef518707825c4a5ee6834a1363b64d also equals incoming.

## Live old-path sweep

Read-source sweep across apps/packages/scripts finds no remaining executable repository-root schema read. Remaining schema strings classify as public URLs/$id; intentional protocol-package-local ../schemas reads; installed schema-policy references to packages/protocol/schemas; historical seal comments; and the layout assertion that the old root file is absent. None is a missed relocation.

Remaining docs/specs/bdp.md references classify as sealed catalog/report/citation data, test-local synthetic metadata, incoming readSpecificationSource compatibility reads, and layout/navigation assertions. The helper assembles Core+Write+conformance in memory for historical citation validation, with repository-boundary checks; it does not recreate a competing normative monolith. Incoming Core/Types/Issue/Memory status language is preserved: proposed Types/Issue remain WIP, Memory remains a separate unadopted PR. No claims of adopting those designs are introduced by this merge.

## Diagnostic and limits

Read-only git diff --cached --check returns2 solely for specs/profiles/write/write.md:4665, a blank line at EOF. This blank line is inherited from incoming Write: exact two-hunk reconstruction proves the transplant did not introduce it. Report the diagnostic honestly; it is not a semantic merge blocker and does not justify widening the resolution.

This review verifies composition, explicit adaptations and preservation, not a fresh semantic audit of every inherited parent line. No product imports, tests, builds, formatter, installations or services were executed, and IW was not edited. Actual merged-tree validation and council disposition remain root-owned. Historical passing parent evidence is not relabeled as a pass for this staged tree.
