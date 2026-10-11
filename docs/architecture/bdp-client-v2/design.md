# BDP client platform and protocol design

Status: exploratory design, owned by Vickie with product rulings by Donna. Last updated: 2026-10-11.

This Markdown document on `donnabox/beads:codex/vickie-bdp-client-design` is the single plan of record at Donna's request. It replaces the upstream issue and intentionally overrides the fleet's issue-shaped tracker convention. It is a coordination and design record, not a normative BDP amendment or release commitment. Earlier client/journal proposals remain on this fork branch; the BDP editorial and implementation work authorized below has separate upstream review PRs.

## Current priority: BDP profile structure and implementation

BDP specification and executable profiles take priority over further V2 architecture design. Work is authorized in the BDP repository on normative prose, schemas, examples, validation, reference implementations and conformance. Donna retains product rulings; Vickie owns BDP integration and normative changes. Janet owns Beads CLI implementation and Preview 2 release integration; Trish owns Types for Preview 3. No merge is authorized. PR78 remains the earlier client/journal design discussion; profile implementation belongs in separate reviewable BDP PRs.

Keep one specification file, organized into three cumulative, substantially self-contained parts: **Read → Read+Update → Transactional**. Each part defines the minimum concepts needed for its own semantics and may depend only on earlier parts. Read uses Resource revisions and ordinary read/pagination semantics; it must not depend on transaction identity, Scope epochs, commit positions, Events or replication snapshots. Introduce a concept in a later part only where its semantics require it. Capability-specific material such as History must have an explicit profile location instead of leaking later-profile requirements into basic reads.

| Work | Owner / boundary | First deliverable | Done when |
|---|---|---|---|
| Editorial | Bounded editorial worker; Vickie integrates normative changes | Section allocation, concept-dependency map, old-to-new clause/anchor map and contradiction register | Every existing normative requirement is located or explicitly dispositioned; no normative forward dependency; profile examples and schemas agree; unresolved semantics are visible rather than silently changed |
| Implementation | Bounded implementation worker, isolated files/worktree | Current runtime/conformance inventory, then one complete Read+Update operation through the installed reference endpoint | Request validation, persistence, guards, errors, response and observed resulting state pass positive/negative executable cases; no unqualified profile advertisement |
| CLI crosswalk | Vickie, with independent review as useful; Janet supplies qualified CLI evidence | create/delete/list/show/update matrix | Each row pins CLI and BDP evidence and covers IDs/references, guards, merge/replace, JSON, errors, atomicity and Events; classified equivalent, legitimate shape difference, BDP gap, CLI gap or unresolved ruling; Preview 2 relevance explicit |
| Integration | Vickie | Small separate PRs for editorial movement, semantic rulings and runtime slices | Each PR aligns affected prose/schema/examples/conformance; exact tested head and limits recorded; no CLI parity claim without checking both sides |

Implementation proceeds cumulatively: preserve the Read baseline; make Read+Update executable and qualified; then implement Transactional atomic batch, rollback/concurrency, durable outcomes and ordered Events/changefeed/snapshot recovery. Existing evaluators and fixtures are reusable evidence, not proof that either write profile is served. Apply each profile's required contract without importing a later profile's wire concepts for convenience. The existing BDP implementation stack is the starting point; this does not initiate the later Rust V2 rewrite.

Donna clarified that the spec should state the current contract directly: draft-development chronology, amendment dates, ruling/council identifiers and prior-draft narrative are not reader requirements. Remove that material; retain actual revision and historical-read semantics in a self-contained optional Read section. Current artifact citations must remain valid, but preserving draft evolution for a hypothetical external audience is not a goal. The complexity of incomplete-history/lineage diagnoses remains a separate design question, not an implicit authorization to delete behavior.

Editorial moves and heading/link repairs should be separable from behavior changes. Establish stable clause identities and an anchor migration map early so code/conformance work can proceed while prose moves. Contradictions go into a decision register with evidence and alternatives; only genuine design choices come back to Donna. Proposed Event enrichments and query changes remain proposals until resolved.

Preview 2 is not the delivery target for the profile reorganization or broad write implementation. Record isolated actual compatibility blockers for Janet; release inclusion remains a separate decision. Use current qualified CLI behavior rather than the tip of an unqualified branch. At planning inspection, release tracker #7170 pins CLI staging at `76fde9c9cb0865d430bf12d76fc473490f51128a`; verify this with Janet before parity qualification. BDP main is `182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`; pending PR64 common metadata and PR65 attribution basis must be tracked separately, with exact accepted source chosen before integration. Trish's PR60 is design evidence, not an adopted wire contract.

Use asynchronous bounded workers with disjoint write scopes under this interactive session. Vickie retains design decisions, the integration queue and the agent bus. No separate interactive role is needed to begin the editorial and implementation inventories. Concrete implementation assignments follow those inventories and the agreed profile boundaries.

First-round review surfaces:

- [BDP PR66](https://github.com/gastownhall/bdp/pull/66), Janet PR64/65 integration at `a9bcf9ed1d869aaba2cb49fe9cca10e3b8d52a94`. Corrected 28 later fixture members to the integrated attribution vocabulary. Build/typecheck/boundaries/format and the full bounded suite pass: 108 files, 7,583 tests, one optional skip. The prior sealed Read cohort remains historical evidence.
- [BDP PR67](https://github.com/gastownhall/bdp/pull/67), cumulative profile structure and present-contract cleanup at `5d7694e8be2f5f4ca72856b288dac0ecd383cd2f`. Removes development chronology and consolidates optional History without losing operative requirements. Maps account for all original blocks and 17 removed ledger entries; 539 citations, 144 anchors and 170 internal links pass audit without lower-profile forward dependencies. Eleven excerpt edits preserve case identities/profiles/counts; nine additive History sibling cases retain original obligations. Author 42 files / 1,636 tests passed before final small wording cleanup, followed by nine affected tests; independent final review passed 58 tests and checked every ledger disposition. Schemas, runtime and sealed evidence are unchanged. Earlier [council dispositions](profile-editorial-council.md) retain the partial external-coverage caveat; final council readiness remains open; fresh combined Read qualification is recorded in PR83 below.
- [BDP PR68](https://github.com/gastownhall/bdp/pull/68), initial durable mutable reference HTTP surface at `ec9df671157a0f4c7a7a484160222254c7777170`. Six Resource singleton mutations use the existing SQLite recovery and shared live reads. Full validation passes: 109 files, 7,588 tests and one optional skip; build, typecheck, boundaries, process restart/replay and installed-package smoke pass. Independent review found no blocker. Development-only, with no full Read+Update or Transactional advertisement.
- [Initial CLI operation crosswalk](preview2-operation-crosswalk.md): five operations with fixed-source evidence, interface translations, conditional guards, metadata/properties semantics, atomicity, JSON and Event boundaries. Qualified CLI pin confirmation and executable cross-repo parity remain outstanding; neither blocks delivery of this source-review artifact.

The first editorial round and initial implementation slice are delivered as drafts, separate from Preview 2. No PR is merged, no release promotion is made, and full write profiles are not qualified. Three-provider council review remains a readiness gate before marking these PRs ready. The continuation below records delivered alias/sequence and atomic mutation work, plus remaining HTTP, recovery and qualification requirements. Transactional commits Resource changes, durable outcome and change group atomically rather than looping over independently committing members.

The crosswalk identifies two further cold-client decisions: portable efficient recent-first ordering and optional owning-source guards. Accepted main also differs from qualified records on metadata and attribution; PR66 integrates the proposed remedy. Janet owns qualified CLI evidence and release impact. Do not treat these future writer/ordering decisions as new Preview 2 Read blockers.


### Autonomous continuation: complete the write profiles

Donna asked for Update and Transactional implementation to continue autonomously after the first-round handoff. Keep bounded workers active and replace completed assignments with the next safe slice; a draft PR alone is not this continuation's stopping condition. Vickie owns integration, review and validation. Continue independently of Preview 2, without merging main or changing Janet/Trish surfaces. Surface genuine product decisions while continuing unaffected work.

| Lane | Initial bounded assignment | Done when |
|---|---|---|
| Update | Complete alias singleton and ordered sequence HTTP paths on the existing durable development receiver; then close remaining HTTP/recovery and conformance gaps | All eight singleton targets and sequences have observed success/refusal/retry/restart behavior; full Read+Update advertisement waits for executable required-contract evidence |
| Transactional | Implement a separate atomic batch/store boundary for Bead+Link local references, one durable outcome and ordered change group | Success commits all effects/outcome/group together; late failure leaves no partial graph/group; retained duplicates and restart recovery satisfy the Transactional contract |
| Review/integration | Review bounded commits, preserve separate PR units and apply council findings before ready | Exact heads validated with appropriate tests, evidence and limitations; no unsupported conformance or parity claims |

Both lanes start from verified runtime `ec9df671157a0f4c7a7a484160222254c7777170` (PR68), incorporating PR66. Update owns development receiver/executable wiring; Transactional initially owns new batch/transaction store modules and tests. Any shared durable-store edits require coordination before changes. Editorial PR67 is the cumulative wording reference; unresolved query/event enrichment proposals and Trish Types are not implementation law.

Continuation progress:

- [PR69](https://github.com/gastownhall/bdp/pull/69), `276d9fde7966e2825360b3ac50f08d29507f27f3`: alias singletons and ordered sequences over real HTTP. 326 focused tests, independent 11-test HTTP rerun, build/typecheck/boundaries and process restart smoke pass. HTTP follow-up is delivered in PR72; full Update qualification remains open.
- [PR70](https://github.com/gastownhall/bdp/pull/70), `64fe7befb24b8d595dac611257cf840588436faa`: private atomic create-Bead/Link batch, pending/terminal receipts, ordered group, late-failure rollback and abrupt-exit recovery. 357 focused tests, independent 13-test owner rerun and installed-package smoke pass. Review repaired Link Event subject/source and private-package inventory defects. No HTTP/public profile claim. Guarded Resource mutations and no-op/final-postimage follow-up are delivered in PR71.
- [Executable qualification checklist](write-profile-next-qualification.md) records 81 Update and 133 Transactional catalog rows, reusable tests and required receiver/recovery observations. These catalogs remain metadata-only; listed cases are not coverage claims.
- Both slices compose cleanly at validation head `28cfbade60b2a7c1bc7a1c3b56b5006f00c878b6`; combined build/typecheck/boundaries, full suite (110 files, 7,607 pass, one optional skip), development restart smoke and installed-package smoke pass. [Verification receipt](write-round2-verification.json). The subsequent editorial council and repairs are recorded in the council report; no readiness result is claimed.
- [PR71](https://github.com/gastownhall/bdp/pull/71), `08aab0ae5b412562726665648b4240f928bd680c`: all six Resource mutations in the private atomic owner, revision guards, ordered intermediate Events/results, normalized final postimages/tombstones, and all-no-op receipt semantics. 364 focused tests and an independent 20-test owner run pass; build/typecheck/boundaries/format/package smoke pass. Receipt retention follow-up is delivered in PR73.
- [PR72](https://github.com/gastownhall/bdp/pull/72), `bd1d0f2f8b534a8760791c86dd0bfb522a22ac99`: development HTTP receiving Problems, shared Read GET/HEAD/ETag/conditional behavior, raw-target checks and mutation Accept comparison before unknown-key admission. Full suite: 109 files, 7,604 pass, one optional skip. Independent 370-test review and build/typecheck/boundaries/format/development/package smokes pass. Deterministic disconnect/recovery follow-up is delivered in PR74.

The next pair composes at `d84ae338b3849ca02912ffa417af39a714861933`: 110 files, 7,624 tests pass, one optional skip, with build/typecheck/boundaries and both development/package smokes passing. [Verification receipt](write-round3-verification.json). Neither profile is advertised; the remaining qualification rows still need executable bindings.

- [PR73](https://github.com/gastownhall/bdp/pull/73), `778216c32c065f129061fb4c453096832f6e20ee`: live receipt expiry/compaction, whole failed-receipt forgetting, terminal-time minimum and atomic maintenance failure fencing. 371 focused tests and independent 27-owner-test review pass, plus build/checks/installed smoke. Next slice adds singleton carriers to the same atomic owner and key namespace.
- [PR74](https://github.com/gastownhall/bdp/pull/74), `84cd9e3bf3afc37ad3c72cb23922cbeb096ca70f`: deterministic real-process disconnect, graceful drain and SIGKILL/restart tests with a private scheduling seam. 131 focused tests plus independent three-process-test review pass, build/checks/both smokes pass. Next slice maps all 81 Update rows to diagnostic executable bindings or explicit omissions, without changing public admission.

These slices compose at `26503a6fc52a01cf7e8181abce4cf39f55fac63c`: 111 files, 7,634 tests pass, one optional skip, with build/typecheck/boundaries and both development/package smokes passing. [Verification receipt](write-round4-verification.json). Receipt workload qualification and full write contracts remain open.

- [PR75](https://github.com/gastownhall/bdp/pull/75), `1d7dd96512db2be991f6ce8bfd75d081099c1904`: six Resource singleton carriers share the Transactional owner, semantic identity, keys and receipts with batches. Independent review exposed and repaired empty-context normalization; a private identity-policy fingerprint refuses earlier stores before recovery/maintenance. Initial 387 focused tests and independent 43-owner-test review; corrected owner suite passes 53 tests, with build/typecheck/boundaries/format checks. Final test-only type annotation passes typecheck and affected-file lint/format. No HTTP, alias or sequence claim.
- [PR76](https://github.com/gastownhall/bdp/pull/76), `d2a32e0c03516130d0ff6bd89486ff48f9330c8e`: development-only diagnostic harness reuses the existing runner and SDK HTTP transport. Nine plans execute 57 actions and 174 assertions; 36 of 81 catalog rows have bounded observations, 45 remain explicit omissions. These are mappings, not passed normative rows. Reports remain `claimEligible: false`, and production artifact admission is unchanged. 270 focused tests plus independent 35-test review and both smokes pass.

These slices compose at `21ba42bf28b483a5ee22c77b6b28b67d464f3045`. Full validation on its immediate predecessor `7fd6936797620d6a4f4c45685ae85cdf50c8bd89` passed 113 files, 7,669 tests and one optional skip, plus build/typecheck/boundaries and both development/package smokes. The only later source change is an explicit test-variable type; final typecheck and affected-file lint/format pass. [Verification receipt](write-round5-verification.json). Editorial PR67 is being integrated with this runtime in a separate candidate; its new diagnostic citations will be audited against the profile-local spec.

[PR77](https://github.com/gastownhall/bdp/pull/77), `d1e62ebc848aff57d46c873013c81efeaf29cd38`, adds alias singleton mutations to the atomic owner. Locator and receipt commit together without Resource revisions, Events, groups or position advance; expired allocations are empty. 464 tests across six files and independent 120-test review pass, including byte-identical refusal of older private formats before pending recovery or expiry maintenance. Build/checks and installed smoke pass. Resource-alias operands remain explicitly refused until durable original-locator/target/miss retry witnesses are implemented. No HTTP or full profile claim.

[PR78](https://github.com/gastownhall/bdp/pull/78), `5a05519d212467e92f390e6c25c9eb0af4d874ec`, extends Update diagnostics to raw credentials/keys, media/Accept, JSON/UTF-8 and body bounds. Fourteen plans execute 105 actions and 398 assertions; 41 of 81 required rows have bounded observations, 40 remain omissions. Independent review passed 136 tests, including unchanged general executor guards; authored verification passed 120 tests, build/checks and both smokes. Fixed test credentials and raw framing stay in the disposable loopback fixture. The next slice binds real process recovery observations.

[PR79](https://github.com/gastownhall/bdp/pull/79), `1fb51f4ef5a6ee607737e72fb50ef0a2ca55ab36`, adds durable per-operation Resource-alias target/miss witnesses. Admission captures witnesses with key/identity/receipt, retries preserve original resolution across slots and carriers, fresh retry locators remain transient, and execution uses captured targets/misses. Completed expiry retains witnesses; failed forgetting and pending retraction permit fresh capture. 479 tests across six files plus independent 201-test review pass; build/checks/installed smoke pass. Earlier private formats are refused unchanged. HTTP, changing authorization and full profile qualification remain outside the slice.

Runtime slices through PR79 compose at `0ccdc9bd8acd808cbddd103119c679a3f2dd42ae`: 113 files, 7,700 tests pass and one optional skip, plus build/typecheck/boundaries and both smokes. [Verification receipt](write-round6-verification.json). That historical runtime-only head excluded editorial PR67; the assembled candidate and fresh Read observations are recorded below.

Latest reviewed delivery:

- [PR80](https://github.com/gastownhall/bdp/pull/80), `5f07a51c0bc0e32652b02ab16cdb2fd4c7f76bc3`: emitted-process disconnect/drain/SIGKILL diagnostics. Seventeen plans, 132 actions and 515 assertions; 44 bounded row mappings and 37 omissions, with `claimEligible: false`. Author 256 tests and independent 68 tests passed. Digests compare canonical parsed JSON; this is not power-loss or mid-commit durability proof.
- [PR81](https://github.com/gastownhall/bdp/pull/81), `4a9137e01817712aedce0b780abb39d02cf1470c`: private Transactional sequence owner with atomic unknown-key admission, ordered separate commits, pending precedence, retry/expiry projection and fenced cleanup. Author 494 tests and independent 127 tests passed. Sequence dependencies and Resource-alias operands remain refused before admission. No HTTP or full-profile claim.
- [Independent review receipts](profile-round7-review.json) pin exact heads and test-report digests for the present-contract cleanup and these two implementation slices. All PRs remain drafts and unmerged.

- [PR82](https://github.com/gastownhall/bdp/pull/82), `0965971b855bd7af0753bc40487a0e0efac1f0eb`: alias allocation/targets, retained targets and misses, fresh-locator comparison and opaque pins. Author and independent review each passed 269 tests; 23 plans / 221 actions / 705 assertions map 48 of 81 rows to bounded observations, with 33 explicit omissions. No runtime/schema/public admission changes. [Independent review](alias-qualification-review.md).
- [PR83](https://github.com/gastownhall/bdp/pull/83), `59915ea916a210f5ce36611eb6d799f7c3124056`: assembled specification/runtime candidate through PR82, including authorized Janet64/65 integration. Fresh Read evidence at `72bdaad` is a two-file commit directly atop observed `4f0ed579`; the current head ordinarily merges diagnostics-only PR82 and repairs ten citations. All 7,726 tests pass, with one optional skip; required real-bd matrix and all build/check/smoke/evidence gates pass. Existing bd1.0.5 baseline, policy and coverage exclusions are unchanged. [Fresh qualification report](read-reseal-review.md). This replaces the three stale-seal failures with new observations; historical evidence is not relabeled. Draft targets main for existing CI, not merge authorization. Council and exact-head CI remain readiness gates; write-profile admission is still disabled. [CI run](https://github.com/gastownhall/bdp/actions/runs/38109034182) is running at exact head 59915ea.

- [PR84](https://github.com/gastownhall/bdp/pull/84), `0fdaa7b50f926aef3e6f9725deea368e63ecf45b`: canonical sequence dependencies with durable creator facts, stable unbound identity and bounded direct comparison. Author 807 regression tests plus final 121 owner tests pass; independent 618 tests, build/typecheck/boundaries/installed smoke pass with no actionable findings. Resource-alias sequence operands remain refused; no HTTP/normative/export/profile changes. This bounded implementation remains subject to the simplicity review and is not yet in PR83.

Current bounded follow-ups:

| Worker | Scope | Acceptance / state |
|---|---|---|
| Transactional implementation | Canonical sequence dependencies | Completed and published as draft PR84 after independent review. Keep isolated from the council’s PR83 head until separately reviewed integration; no further mechanism expansion during the simplicity discussion. |
| Update qualification | Remaining executable observations | Read-only audit of 33 omitted rows and incomplete variants within 48 bounded mappings, against existing runtime. Identify meaningful next tests that need no new mechanism or misleading coverage claim. |
| Integration validation | PR82 on fresh qualified candidate | New isolated integration `59915ea916a210f5ce36611eb6d799f7c3124056` preserves evidence ancestry and merges only reviewed diagnostics. Ten diagnostic citation repairs; 162 focused tests pass. Full 7,726 tests / one skip, all checks and both smokes passed; current evidence gate verifies unchanged run head and evidence commit. [Integration report](alias-qualified-integration-review.md). Published in PR83. |
| Council | Frozen PR83 at 72bdaad | Native and two external reviews use the same three bounded source packets: profile contract, write ownership/recovery, and HTTP/conformance/evidence. Supplied ranges and hashes recorded; no whole-branch coverage claim beyond actual reviewed content. The 72bdaad→59915ea diagnostics-only delta is a separate packet for all reviewers. No readiness change until dispositions are resolved. |

The [earlier candidate diagnosis](profile-write-integration-diagnosis.md) remains historical evidence of the genuine catalog mismatch. The fresh [qualification workflow](read-reseal-feasibility.md) has now been executed for PR83 with unchanged pin and policy. Further changed inputs must satisfy the same verifier, without weakening checks.

The heartbeat continues through successive slices toward tested Update and Transactional support, stopping on completion, an operator pause, supersession or an unavoidable decision. It reports meaningful results and blockers, not routine unchanged polls. First-round completed work above remains the baseline.


## Agreed direction

BDP is the provider interface. A provider writes and owns its HTTP endpoint over its chosen store. Our Rust CLI/client is a surface over BDP; commands need not correspond one-to-one with requests. Any provider meeting the required generic BDP capabilities must run the Issues pack without Issue-specific provider code.

Packs build on Gas City core concepts and carry Bead/Link definitions, skills and extensible CLI commands. Issue commands belong in an Issues pack. A shared provider library is possible, but does not take over the endpoint. Brian's team plans the Postgres store; our team designs the client. Current `bd` must coexist side-by-side. Mode/URL/Scope configuration is a candidate, not settled syntax.

Trish owns Type lifecycle design. Janet owns current Beads/Preview 2 integration, CLI/BDP alignment and qualification. Do not edit their branches, PRs, specs, tests, playgrounds or shared databases. Coordinate dependencies through the existing bus. BDP profile implementation is authorized under the current priority above; V2 client implementation and merges remain out of scope.

## Simplicity review

Donna explicitly directs a YAGNI review: mechanisms need a concrete product purpose, not speculative future utility or existing implementation effort. The [BDP simplicity catalog](yagni-catalog.md) lists 36 current-contract candidates across History, Read, mutation and Transactional/replication, plus four Type-related pointers reserved for coordination with Trish. Each gives the source, benefit/cost, a smaller alternative and the guarantee lost. Review one decision at a time, beginning with what exact historical retrieval requires. No individual removal or profile change is approved yet. Existing bounded implementation may finish, but do not widen its scope while these choices are being reviewed. Catalog recommendations are not normative law and do not justify weakening conformance claims.

## Batch composition direction

Donna agrees that ordinary `bd create`, `bd update` and `bd delete` should work inside `bd batch` when transactionally expressible, with create covering Beads and Links in the new CLI. Treat these as shared semantic command implementations. This is a CLI composition direction, not adoption of a new BDP wire operation or a claim that every current operation is already reducible. Remaining distinctions include alias mutation, bulk target selection and sequence-versus-atomic-batch behavior. Coordinate concrete CLI findings with Janet; do not edit her branch.

## Accepted decisions

### D01 — A usable cold client

Accepted by Donna during the design interview on 2026-10-08: **we want a usable cold client.** Common CLI commands must work efficiently over remote BDP without first downloading, indexing or synchronizing the whole Scope. Local caches and replicas are optional optimizations, not prerequisites for ordinary use.

Evaluate the query and transaction proposals against that requirement. The exact command coverage, supported query shapes, scale envelopes and latency/request/byte budgets remain open; this decision does not promise constant latency for arbitrary queries or accept the proposed wire syntax.

### D02 — Required relationship-aware queries

Accepted by Donna during the design interview on 2026-10-08: every provider claiming Issues-pack compatibility must support generic relationship predicates, filtering and ordering sufficient for queries such as “return ten open tasks with no unfinished blockers.” The Issues pack supplies the domain predicates; providers evaluate the generic operations efficiently. Full-Scope download is not an acceptable prerequisite for these ordinary cold-client commands.

Donna wants to keep the expression language bounded. Treat that as a design goal to validate against the command audit; the operator set, nesting/recursion rules and execution budgets are not yet accepted. A bounded expression shape alone does not bound the amount of data examined.

### D03 — Zero domain-specific event types

Accepted by Donna on 2026-10-09: the generic stream must contain **zero domain-specific event types**. Closing, commenting, assignment and other application actions must be expressed through generic resource/state transitions; no special `close`, `comment` or Issue-only event kind is admitted. This does not settle the exact generic vocabulary or comment data model, and is not authorization to change release code. Compatibility with existing consumers requires an explicit migration decision. The [worked five-verb model](five-verb-event-model.md) maps current actions and shows comment, Link, transaction and recovery behavior; its inline-comment choice remains provisional. A proposed [stream-only identity bridge](five-verb-event-model.md#stream-only-bridge-for-legacy-dependency-identity) gives legacy dependency lifetimes opaque Link references while preserving existing CLI/pair uniqueness; it requires durable mapping and sound baseline/continuity.

Comments need not become Beads to satisfy D03. An inline collection of comment objects could be part of the parent Bead's canonical state and change through a generic update. That choice remains proposed, including owner revisions, efficient append and pagination. Current comments already have string IDs, author, text and creation time; current dependencies have mutable JSON metadata. [Source-grounded clarification](beads-1.3-journal-vs-bdp.md#clarification-zero-domain-events-comments-and-link-updates)

## Evidence

The [CLI audit](audit-20261008/README.md) classifies 274 main registry paths and five Preview additions, with source/test contracts and cost analysis. 219 existing BDP tests and 20 additional bounded probes passed. This is parser/schema/fixture/model evidence, not end-to-end Transactional provider qualification. [Verification](audit-20261008/verification.json) pins source commits and checks 287 citation anchors and 160 source files. The [earlier source map](baseline/runtime-boundary-map.md) records 13 runtime interfaces and five paths.

The historical [journal audit](audit-20261008/event-log-audit.md) distinguished Events and change groups. The current review must sharpen that distinction: missing original command intent is not itself an obstacle to forward state replication. Neither before-images nor original selectors are intrinsically necessary to apply complete committed effects. The question is whether today's Event Source contract carries every state/control transition and an adequate atomic replay boundary.

## Beads 1.3 journal comparison

Information exposed by working 1.3.1 should be considered for generic BDP support when it improves efficiency or consumer convenience. The comparison presents [recommended changes to BDP](normalized-1.3-vs-current-bdp.md#recommended-changes-to-bdp) and [recommended changes to 1.3](normalized-1.3-vs-current-bdp.md#recommended-changes-to-13), including the three working 1.3.1 changes: generic close updates, comments in parent state, and durable Link lifetime identities. Specific payload choices and release scope remain open.

The [detailed side-by-side comparison](normalized-1.3-vs-current-bdp.md) holds the proposed close/comment normalization and Link lifetime identity changes constant, then compares modified 1.3 with both current BDP Events and Scope change groups. It separates the now-aligned vocabulary from remaining payload, transaction, snapshot, epoch, coverage, projection and erasure obligations. These assumptions are a comparison scenario, not implementation or release approval.

Donna requested a comparison of the shipped Beads 1.3 journal with BDP before choosing possible 1.4 changes. The [comparison and recommendation](beads-1.3-journal-vs-bdp.md), backed by an [independent source audit](beads-1.3-journal-source-audit.md), finds substantial overlap with BDP's **changefeed**, especially its postimages. The existing journal's transactional capture is reusable; its flat records lack transaction framing, history epochs and snapshot/checkpoint rendezvous, and some writers/state changes can bypass it.

Recommendation, not an accepted release scope: converge the committed-effects capture, preserve the legacy projection, and add the stronger replication guarantees before claiming BDP compatibility. Preserve comment payloads despite contradictory six-op projection guidance. Coordinate with existing journal PRs 7211/7213/7144. No implementation or release artifacts were changed. D01/D02/D03 are accepted; the bounded-expression-language question remains open while this research is discussed.

## Current questions

1. Propose the smallest coherent generic query and transaction additions that make current Issue workflows efficient, with examples beyond Issues, provider costs, limits, authorization semantics and failure behavior.
2. Identify precisely what an Event-only consumer lacks for an authorized current-state replica. Separate this from historical reconstruction, command replay and full provider backup. Compare enriching an event stream with retaining current change groups; avoid redundant independent sources of truth.

## Decision plan

| Step | Deliverable and done-when predicate | State |
|---|---|---|
| C1 — client/provider contract | Interface/responsibility table names invariant and error ownership, excludes direct store coupling and distinguishes accepted/open choices. | Open; direction agreed |
| C2 — CLI over BDP | Every in-scope behavior has a request plan or explicit gap; representative cases state atomicity, requests, bytes and retry/failure behavior. No proposed behavior is presented as implemented. | Bounded audit delivered; choices and runtime qualification remain open |
| C3 — packs | Inspected Gas City provenance, reuse/change list, dispatch/dependencies/version/trust interface and worked Issue command over generic BDP. | Open; Type lifecycle remains Trish's |
| C4 — Rust/coexistence | Library responsibilities and phased migration cover executable selection, configuration/auth/capabilities, compatibility and recovery, with entry/exit conditions. | Open |
| C5 — review | Exact reviewed revision and findings recorded here; Donna's acceptance and Janet's coexistence acknowledgment recorded with provenance; dependencies have owners. | Open; two human acceptance records remain |

The proposals below are ready for discussion. They remain proposals until Donna rules; their names and syntax are illustrative and do not change the current schema.


## Proposal 1: one query model reused for reads, checks and writes

I recommend one bounded, structured query algebra, with three uses. This keeps the provider interface generic while moving selection work close to the indexes. The Issues pack supplies domain predicates; the provider only evaluates the generic algebra.

| Use | Proposed capability | Why it matters |
|---|---|---|
| Find data | `query` with filtering, ordering, selected fields, counts and anchored relationship predicates | Return ten useful records instead of shipping the candidate corpus to the CLI. |
| Protect a decision | `assertQuery` inside an atomic batch | Check conditions such as absence, exact membership or unchanged ordered results at serialization, including newly inserted matching records. |
| Choose and change | `queryApply` inside an atomic batch | Select a bounded ordered set and apply generic mutation templates to it without a read/network/write race. |

### Read semantics

The initial algebra should include scalar predicates, exact array membership, explicit null/missing rules, deterministic ordering with an ID tie-break, selected property paths, exact existence/count and `existsLink` with direction/type/Link predicates plus a predicate on the opposite endpoint. Negation supplies “no related record satisfies this condition.” The same mechanisms work for appointments, inventory reservations, documents awaiting approval and equipment with expired certificates.

Text search must state its semantics. For current CLI compatibility we need literal substring search, with a protocol-defined Unicode case-folding mode where requested. Database-default collation or a ranked full-text query is not an equivalent substitute. A provider may choose its indexes, but must implement the same meaning. No arbitrary SQL, custom server function or provider-loaded Issue code is needed.

Ordering and projection reduce wire cost; they do not guarantee cheap server execution. Broad relationship predicates and substring queries can still scan. Required query shapes need indexed, measured scale envelopes, with explicit budget failures. Exact counts may be expensive; an estimate must never masquerade as an exact count.

Related queries can share an expiring immutable **read handle** covering both collections at one Scope position and Authorization View. Return it with the first query, rather than requiring a separate round trip. This does not create a client-controlled open write transaction. Providers choose how to retain the view; it need not pin a PostgreSQL connection per client. Full snapshots remain the replication bootstrap; projected query rows are visibly distinct from full canonical records.

Bounded reachability is an additional generic capability for descendants and transitive dependencies. Depth/visited-edge/result limits must be explicit. If a complete negative result cannot be proved within the budget, report incomplete/limit failure instead of returning false. It may be optional for generic BDP implementations, but must be required by any Issues-pack compatibility profile that promises those commands.

### Transaction semantics

`assertQuery` is a read assertion, so a reader does not need mutation permission on every dependency simply to guard a decision. It can assert zero matches, expected cardinality, or a prior observation token representing the ordered result identities and revisions. The provider rechecks or proves equivalence at serialization. A token compresses the request; it does not eliminate validation work. A phantom, changed ordering or changed relevant revision must invalidate the observation.

Assertions see preceding staged effects. They are not blindly checked against final post-state: claiming an unassigned task intentionally changes the state that passed its precondition. The enclosing batch must nevertheless serialize the checks and mutations against concurrent writers.

`queryApply` freezes an ordered selection when reached, then applies a bounded template of existing generic mutations. A take-one selection can update the chosen task and create a lease Link in the same transaction. Bindings initially cover selected identity/revision, named creations and narrowly specified scalar assignments. This is a finite operation plan, not arbitrary server code. The complete mutation remains subject to operation, resource, byte and induced-event limits.

Keep **match cardinality** distinct from **take**: “exactly one exporter exists” differs from “choose one available item among many.” Idempotent replay must return the original selected identities and outcome. Strict queue order cannot silently become `SKIP LOCKED` opportunistic selection; relaxed selection would need a separate explicit contract.

Partial-success multi-close is the one larger control-flow question. A possible optional capability is a flat guarded group: false business precondition skips that group's mutations; true executes them; all applied groups still commit together. Structural/auth/limit/execution errors abort the batch. This is not general exception catching and does not promise parity with every legacy per-item failure until tested. Start with query/assertion/selection, and standardize guarded groups only against the precise compatibility cases. A full Issues profile must require the capability or declare an approved semantic change.

### Concrete CLI journeys

| Journey | Proposed request plan | Target, not a measured result |
|---|---|---|
| `ready --limit 10` | Wake expired dated deferrals, then query scalar + relationship predicates with projection; share one handle for the recent/older hybrid-order buckets. | Usually 2–3 application requests after discovery and O(returned rows) wire payload; bounded graph evaluation stays at provider. Separate handle creation or receipts can add requests. |
| `ready --claim` | One batch performs any required wake and ordered `queryApply(take:1)`, patching holder/status and creating lease state. | One mutation submission for simple ordering; no whole candidate list. Hybrid ordering needs conditional fallback or a guarded probe/retry. Receipt polling and contention retries remain possible. |
| `close A B --claim-next` | Guard each target's policy, apply permitted closes, then choose next work against the staged state. | One atomic unit if guarded groups are selected; otherwise explicitly documented preflight/retry or changed semantics. |
| Search or small graph view | Exact text/relationship query against a consistent handle; project only display data. | No mandatory full-Scope bootstrap for an ordinary cold command. Index/query-budget qualification is still required. |

The [detailed generic proposal](generic-bdp-proposals.md) defines the operator semantics, limits, PostgreSQL implementation options, visibility rules and three worked journeys. It also explains where a local index remains useful, notably fuzzy pairwise comparison and repeated graph analysis.

### Limits of this proposal

All reads and joins operate within the authorized view. A hidden blocker cannot silently become evidence of “no blocker.” A command that promises authoritative global policy needs a principal with the required complete view, or a separately defined conservative completeness contract. Refusal must not itself reveal whether a particular hidden record exists.

These mechanisms enforce a submitted predicate; they do not force every client to submit an Issues-pack policy. Decide separately whether those policies are cooperative-client guarantees or generic authority-enforced constraints. Do not add private provider Issue logic to bridge that distinction.

Clock/lease semantics and no-history operational state remain separate decisions. Query constants can initially use an explicit client cutoff, but strict authoritative lease timing would need a generic, precisely defined transaction-time value rather than casually using server wall-clock functions. No-history heartbeats cannot be claimed solved merely by placing them on a separate versioned Resource. Identity rename, full history restore and native database administration also retain their explicit migration questions.

## Proposal 2: a replica-capable event stream

**The prior answer was too broad. Events can reconstruct a replica.** Missing original commands, selectors, guards and before-images is not the problem. For ordinary forward replication, create inserts a record, update applies a committed delta to the known predecessor, and delete removes it. A deleted record's previous body is needed for undo or retained history, not to remove it from current state.

BDP already provides most of those payloads, including owned-Link source deltas. The earlier fixture probe demonstrates event-folded state matching the change-group postimages for its sample. That is positive evidence, though only for the fixture.

The precise question is whether the **current Event Source delivery contract** carries every transition and enough framing to maintain an atomic, resumable, authorized current-state copy. It presently does not promise all of that.

### The smallest state counterexample

Start with a replica containing Bead A. The current specification permits A to leave a projection while remaining alive in the authority and while the view token stays stable. A change-group tombstone communicates that removal. But there is no actual `deleted` lifecycle fact, and an `updated` fact is exposed only when visible in both pre-state and post-state. An Event-only consumer can therefore retain A incorrectly.

On entry, a previously hidden existing Bead needs a full visible image: it was not just created, and the consumer does not possess a base for an ordinary patch. This needs explicit projection-entry/exit transitions, or a rule requiring resnapshot on every such membership change. The latter is simpler but potentially expensive. Grants/revocations already rotate the view and are a separate reset case. [Current projection and visibility rules](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6037-L6140).

### The remaining replica contract

| Requirement | Why today's Event Source is insufficient or awkward | Minimal event-stream solution |
|---|---|---|
| Snapshot handoff | Snapshot returns a Scope checkpoint; Events accept source-local Event IDs, with no defined rendezvous between them. | Snapshot returns/accepts the exact exclusive stream checkpoint. |
| Atomic transaction delivery | Individual SSE Events share a transaction ID but lack an explicit completion boundary; ordinal gaps can be legitimate. | Deliver a complete transaction frame or explicit completion marker. Persist state and cursor atomically. |
| Canonical initialization | Creation Events omit Type-derived empty owned-Link keys. | Complete creation/entry image, or stable metadata resolution. Full images on every update are unnecessary. |
| Recovery and freshness | Source-local delivery progress is not durable applied state or a Scope-position guarantee. | Epoch/view fences, durable checkpoint, continuity and explicit replay-expiry/reset semantics. |
| Erasure | An archive may retain erased prior content while lifecycle Events show no live-state change. | Erasure records and a bootstrap ledger in the same stream contract. Historical erasure alone is not a counterexample for a copy that retains only unaffected current state. |

Atomic reconstruction from individual Events is not mathematically impossible: additional finite reads reaching the current source end could supply completion information. The framed stream makes that guarantee direct and efficient, especially when the final transaction is followed by silence.

### Recommended shape

Keep one committed transaction envelope with Scope epoch, Authorization View, previous position, position and checkpoint. Make its ordered entries sufficient to reconstruct the projected state:

- Complete creation/projection-entry records.
- Committed property/owned-Link deltas with predecessor and successor revisions.
- Deletion/projection-exit identities, with the meanings kept distinct.
- Erasure and progress/reset control records where needed.

Activity/resource feeds become filtered views of that stream. Final postimages can remain a derived transfer mode for simpler clients and efficient catch-up; they need not be an independently authored truth or mandatory duplication for every small patch.

Current BDP already atomically commits Events and final-state changes together. It does **not** require two independent logs. The proposed change is to make the canonical transition stream reconstruction-complete and make the relationship between its views explicit. Providers remain free to use relational transactions and an outbox/change table, event sourcing, or another compliant implementation. We should require event-fold/read equivalence, not dictate the storage engine's internals.

A command request, a committed transition and a retained historical record remain different concepts. Converting Events back to mutation requests is unnecessary and can duplicate derived effects; directly reducing those facts into replica state is the appropriate operation. Full provider backup adds administrative and historical state beyond this current-state replica contract.

The [source-grounded reassessment](replica-reassessment.md) records the exact counterexamples and qualifications. My recommendation is to evaluate this event-first wire design against the existing delta-plus-postimage group design, preserving snapshot, atomicity, projection and erasure guarantees in both.

## Validation and next decisions

No runtime protocol changes were made and no new runtime conformance was claimed. The original audit tests remain evidence for the baseline only. This revision is source-grounded design review; its [source anchors and migrated files are checked mechanically](design-verification.json).

Before accepting the proposals, require:

1. Differential query and command results on two providers, including a non-Issue domain using the same operators.
2. Phantom/queue races, view changes, strict ordering, expired handles/tokens, limits and ambiguous-retry tests. Publish cold/warm byte/request/latency budgets rather than assuming that expressiveness implies efficiency.
3. Snapshot plus event-fold equivalence to canonical reads through ordinary and owned mutations, projected entry/exit, multiple updates in a transaction, reconnect/crash, quiet-stream completion, erasure and epoch/view reset.
4. Byte/CPU/storage comparison of event deltas versus final postimages on large records, small patches and graph batches. Keep provider internal architecture free.

The first decision I recommend discussing is the shared query algebra and transaction selection/assertion contract. For the log question, first agree that the target is an authorized, atomic current-state replica; that lets us resolve the actual stream gaps without dragging command audit or full backup into the requirement.

## Workspace and migration receipt

This work moved from the former upstream issue into this fork branch after Donna requested a quieter design workspace. The full previous issue and its comments are preserved in [the historical archive](legacy-tracker-archive.md); old links inside evidence are not live authority.

The branch and archive were verified before attempting issue removal. GitHub denied `DeleteIssue` because the current account lacks permission. The former issue is now closed as not planned, its current title/body replaced with a withdrawal, and all four author-written comments deleted. One automatic triage comment remains; actual issue deletion requires an upstream administrator. No replacement issue or PR was created. The fork itself is public.
