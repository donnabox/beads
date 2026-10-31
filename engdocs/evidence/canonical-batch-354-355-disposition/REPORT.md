# PR354/355 — current-candidate source and intent disposition

Both original intents are represented in exact composed candidate 31e68fe91a14eaf7968b51380294be166fe046dc (tree b5fb071aecd419cd642a9f7d2ba52112e873459f). No missing source/test effect from either owned topic was found. This is source incorporation in the candidate, not a main-landing or current combined-behavior PASS. No source, branch/index, PR state, owner post or qualification was changed.

## Exact PR subjects

| PR | Original logical delta | Current read-only PR head/base | State |
| --- | --- | --- | --- |
| [354](https://github.com/donnabox/beads/pull/354) | 18a1374a7ff9e8db3a2d90091f5c362673863101 → e40d9d3e0364503ae305eaf2d27bb8aec2742b56; seven paths, +1298/−13 | a60b3e8af58a54f0cfcdac5d1ecf914b021e8989 / 28c523bbed55ec8bbe3470612f898f4af4ccef2e | OPEN, DRAFT, no merge |
| [355](https://github.com/donnabox/beads/pull/355) | 4cae8f790f60c1b99e76194cba96410398987c03 → f4b2b3b062f08862320f700f499efc6c7bb0e667; six paths, +516/−7 | 128c5d54bc3f7fb72d4bebb271e853f7fbff1ead / a60b3e8af58a54f0cfcdac5d1ecf914b021e8989 | OPEN, DRAFT, no merge |

Each logical delta was enumerated directly with Git diff, not inferred from ancestry or its last commit. Every topic postimage/mode matches its published composed PR head. a60b has ordered parents e40d +28c523; 128c has f4b2 +a60b. In particular,355 must be compared to logical base 4cae, not the later moving354 head: its original f4b2 storage files predate five later354 hardening postimages. The published composition and current candidate retain those later fixes; no older355 storage blob was treated as the final target.

## Full changed-path correspondence

Paths below are relative to the repository. Exact means complete mode/blob equality with the original topic postimage. BUILD differences were read: later additions only, no removal of either topic’s source/test/dependency registration. This checks source/test effects rather than crediting a commit name.

| PR/path | Original effect and current-candidate representation |
| --- | --- |
| 354 internal/storage/graphstore/atomic_batch.go | Exact. Both freshly encoded success before INSERT and recovered success use validateAtomicBatchRetained with Scope/members; changed reader signature retained. Candidate lines173,311,368–391. |
| 354 internal/storage/graphstore/atomic_batch_retained.go | Exact. Bounded canonical/closed successful carriers, operation/identity completeness, full scalar validation, ordinary historical bodies/pins, final owner/capture correlation and producer-order preservation. |
| 354 internal/storage/graphstore/atomic_batch_retained_test.go | Exact. Complete scalar positives and disguised corruptions, body/mapping/variant/transaction failures, owner/pin/capture controls, ordering, bounds and detached returned results. |
| 354 internal/storage/graphstore/atomic_batch_recovery_integration_test.go | Exact. Actual stored corruption rejection and scalar byte-preserving replay, profile isolation, reverse private/public owned capture with current Show/exact ReadVersion, duplicate-membership refusal. |
| 354 internal/storage/graphstore/atomic_batch_concurrency_test.go | Exact. Authored actual two-pool same-key and original-revision competition plus lost-COMMIT-response recovery; actual server precondition/skip remains, so source presence alone is not an executed pass. |
| 354 internal/storage/graphstore/history_exact.go | Preserved with later adaptation: exact seen-ID uniqueness/source/ownership/link validation at candidate360–377 retains recorded capture order. Relative to original 354, only three calls changed: two readVersionBytes calls now use s.readVersionDomainBytes (163,201), and applyIssueRevisionContext now uses s.applyNativeIssueRevisionContext (440). These are actual later composition bytes, not claimed byte identity or newly qualified behavior. |
| 354 internal/storage/graphstore/BUILD.bazel | Original validator source plus all three new test files remain registered; later additions only. |
| 355 cmd/bd/canonical_batch_cli.go | Exact. Strict scalar occurrence checks stay scoped to selected canonical batch, with actual discovered handler check before input consumption; ordinary legacy command-text last-value behavior retained. |
| 355 cmd/bd/canonical_batch_execute.go | Exact. Canonicalize invocation before digest/save and bound canonical bytes; legacy escaped reader has sixfold physical cap followed by canonical logical cap. All prior frozen binding/retry and stdout error messages remain. |
| 355 cmd/bd/canonical_batch_recovery_test.go | Exact. Six roots: journal roundtrip, legacy file compatibility, discovered conflicting file flags, frozen retry inputs/override refusal, offline unreachable source, HTML expansion/logical-bound controls. |
| 355 cmd/bd/canonical_batch_workflow_test.go | Exact. Real registered batch workflow forces stdout failure with a read-only descriptor after confirmed commit; finds saved key, removes input file, replays twice and compares original full result. |
| 355 engdocs/design/canonical-batch-candidate.md | Exact. Documents frozen journals/HTML recovery and explicitly preserves remaining public capability/delivery limits. |
| 355 cmd/bd/BUILD.bazel | Original recovery test/dependency registration retained; later additions only. |

Total: ten complete original postimages exact; two BUILD files retained with additions; one historical-read file retains the original fix with three later bridge substitutions. bindings.json records original/published/current blob IDs and modes for every path.

## Residuals and conditional disposition

There is no additional original 354/355 feature source to author merely because these drafts are still open. Keep their queue state as still needed until the actual integration subject lands. If the accepted candidate (or an exact receiving successor whose changes are accounted for) lands on main with these thirteen effects preserved, disposition both as incorporated by that actual integration, naming the landed commit and receiving evidence. A second independent cherry-pick/merge of already incorporated topic bytes is unnecessary. This report does not close either PR, retarget it, merge it, or select owner state.

Main is only the supplied last-verified pin 6dd6a81f5c733a0b13f92f787e5e16c91affbb1d; no main ref was fetched or moved and no current remote-main claim is made. A bounded thirteen-path comparison at that pin finds ten topic files absent; only the older two BUILD files and pre-354 history_exact.go exist. Therefore this audit does not establish these intents delivered on main. The two read-only gh views above only establish current PR heads/base/state.

Published-body StageA/StageB receipts concern final 128c5d54 and separately pinned earlier subjects: they are real historical scope statements, not fresh behavior qualification for 31e68fe plus later Types/Query/output composition. No tests, build, normal generation, CI observation or runtime occurred in this task. The current three historical bridge substitutions in particular are labeled composition, not silently covered by old reverse-order runtime evidence.

Preserve the declared limits: local native batch Version="2" (internal/canonicalbatch/types.go:14), fresh opt-in layout15 and retained-success-only recovery; no retained failed-key or shared singleton/principal-key implementation promise. Existing canonical-batch-candidate.md:104–118 and137–144 expressly leave those open. No full public BDP Write/provider/transport, restart/power-loss, complete network delivery or blanket CI PASS follows. HTML journal repair permits bounded accepted requests and bounded legacy escaped carriers; it does not fix large-HTML dry-run emission. Confirmed-commit stdout failure followed by immutable saved recovery does not prove arbitrary output/context parity or failure/restart durability. Known335 failure/cleanup history remains literal.

## Bindings and method

Only two PRs were queried read-only, and all source comparisons used immutable local Git objects. No fetch, ref/branch/index mutation, source/W write, test/runner/engine, CI watcher or outward state change. Companion bindings.json contains exact PR metadata, candidate head/tree/parent, logical bases and all thirteen blob/mode comparisons. Original published intent sources:

- /Users/dbox/.local/state/sybil/canonical-crud-batch-20261028/day-hardening/remaining-stack-plan/PR-354-published-body.md — SHA-256 3ff46e3e534e561a470ffdb1945a4d9971efac64612dd62a821800511a55721c
- /Users/dbox/.local/state/sybil/canonical-crud-batch-20261028/day-hardening/remaining-stack-plan/PR-355-published-body.md — SHA-256 8567899a43e9a085d9d55a7ccb6f06657f60e7c6fdbd35dab588738b3776cd0d

bindings.json SHA-256: 3d9a6c676f4ea739710dbdf03276acfb71124e87b2fc8b3b8bfdd5358ab68911.

## Consumer receiving — handoff 6541222459 (2026-10-31)

Sybil receives Janet's independent source comparison into this existing
354/355 owner disposition. **Both topics remain still needed and unlanded;
their source is incorporated in the selected candidate, with the three
354 history bridge adaptations explicitly accounted for.** Sybil retains
354/355 ownership. Janet remains the 371 integration writer; this receipt
transfers no ownership.

This publication preserves the original owner report above as an exact byte
prefix (8,871 bytes, SHA-256
`224dfdd6c06a6913a26b30dded288701c70f3e1ed55d24a771c43215b89e88d9`)
and its unchanged [bindings.json](bindings.json) (12,214 bytes, SHA-256
`3d9a6c676f4ea739710dbdf03276acfb71124e87b2fc8b3b8bfdd5358ab68911`).
Statements above about reads, writes and PR state describe that earlier audit.
The receipt below is a documentation publication, not a new source or runtime
audit. The original local owner artifacts were read without modifying them.

The bounded authority and received evidence are:

- [Exact Janet handoff](https://github.com/donnabox/agent-coordination/issues/1#issuecomment-6541222459).
- [Independent incorporation report](https://github.com/donnabox/agent-coordination/blob/61f48e895f738e03b3dbf4e679bc9d489ee7aafb/context/janet/upstream-and-recovery-20261027/overnight-20261029/BEADS-PR355-INCORPORATION-READONLY-0827.md).
- [Immutable thirteen-path facts](https://github.com/donnabox/agent-coordination/blob/d27fc6092a723313bbf7caa620dec1223e0491ff/context/janet/upstream-and-recovery-20261027/overnight-20261029/BEADS-PR355-INCORPORATION-READONLY-0827.json).
- [Separate ordinary WIP publication source route](https://github.com/donnabox/agent-coordination/blob/3ee58d3bb41ce42897a745826c2bc21fb3b2e202/context/janet/upstream-and-recovery-20261027/overnight-20261029/TRISH-WITNESS-WIP-PUBLICATION-SOURCE-ROUTE-0830.md).

All thirteen original/published/candidate/main mode-and-blob tuples and their
dispositions match the existing owner bindings. The received PR heads, bases,
draft states, selected candidate `31e68fe91a14eaf7968b51380294be166fe046dc`
and tree `b5fb071aecd419cd642a9f7d2ba52112e873459f` also agree. This
comparison checks the two evidence records; it does not rerun Janet's source
inspection or query live PRs. Janet's facts were observed at
`2026-10-31T08:28:32.778638+00:00`; all metadata remains scoped to those snapshots.

Ten complete non-BUILD postimages are exact, including all five non-BUILD
355 paths. Both BUILD files retain their original entries with additions only.
The remaining 354 file preserves its original owned-membership fix with
exactly two calls to `s.readVersionDomainBytes` and one to
`s.applyNativeIssueRevisionContext` substituted by the composition. These
adaptations remain explicit; historical reverse-order evidence does not
qualify their current behavior.

The last verified main is still the evidence pin
`6dd6a81f5c733a0b13f92f787e5e16c91affbb1d`, where ten selected files are
absent. No fresh remote-main assertion is made. Historical `128c5d54` results
stay attached to that subject and do not qualify the current 371 integration;
its recorded native, budget and postguard failures and unrun repaired cases
remain open. Source incorporation supplies no current-runtime green.

Keep the success-only local layout15/encoder2 boundary, actor versus
authenticated-principal distinction, and failed-outcome/shared-key,
public-provider/transport, restart/power-loss, output/context and large-HTML
dry-run limits stated above. Open status alone is not evidence of a missing
355 production repair. The separate WIP source route remains context: current
generation, affected checks, hooks/configuration, source/remote facts and any
matching finite publication grant belong to its owners.

The next disposition action waits for an actual main landing that retains
all thirteen effects. At that point, refresh the individual heads, bases,
dependents, discussions and operative policy before a deliberate incorporated
retirement tied to the landed commit. Existing draft-closure and substantive
human/different-merger gates remain; proposed policy 372 is not adopted.
This bounded receipt changes no PR state, retargets no stack, creates no
feature or duplicate campaign, starts no test replay or watcher, and requests
no routine acknowledgment. No launch file is read or changed.

Validation for this receipt: original report/bindings hashes, JSON parsing,
all thirteen evidence correspondences, relative Markdown links and diff
whitespace. No build, test, CI or engine execution was performed.
