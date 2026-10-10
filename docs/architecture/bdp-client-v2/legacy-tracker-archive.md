# Historical issue archive — superseded, do not resume

The operator moved this work to [design.md](design.md) on 2026-10-08. This archive preserves the previous tracker body and comments before removal. It is evidence only, not a live tracker. Original links and wording are historical.

## Former issue body

## Purpose and ownership

Vickie and Donna are designing the **Rust BDP client platform and extensible CLI**. This issue is Vickie's single plan of record. Donna owns product rulings. [Donna's October 8 direction](https://github.com/gastownhall/beads/issues/7403#issuecomment-6069015755) supersedes the initial next-step framing around extracting an in-process Beads runtime.

**BDP is the provider interface.** Store providers expose their own BDP HTTP endpoints. The CLI works through BDP, with efficient support for current CLI functionality rather than a required one-command/one-request mapping. Packages build on the core of Gas City packs and include Bead/Link Type definitions, skills and extensible commands. Issue commands should live in an Issues pack. Donna subsequently ruled that any provider meeting its required generic BDP capability contract MUST support that pack without Issue-specific provider code; protocol/client gaps must be evaluated against that requirement. The new implementation direction is Rust. Brian's team owns the planned Postgres store; our team focuses on the client side.

A shared provider-support library is a possibility, with scope still open; it does not replace provider ownership of the HTTP endpoint. Current `bd` needs side-by-side coexistence. A graph-mode-like selector and target service/Scope URL are a candidate, not a decided executable layout or flag spelling.

Trish retains Type lifecycle design in [BDP #59](https://github.com/gastownhall/bdp/issues/59)/[PR60](https://github.com/gastownhall/bdp/pull/60). **Defer Type details from this exercise and let Trish lead them.** Janet retains Graph Preview 2 integration, current CLI/BDP alignment and qualification under [Beads #7170](https://github.com/gastownhall/beads/issues/7170). Her current Beads integration is a separate stream, not Vickie's target client architecture.

Coordinate dependencies through [donnabox/agent-coordination#1](https://github.com/donnabox/agent-coordination/issues/1). Donna is not a message courier. This is design authority only; no implementation, protocol amendment or merge is authorized here.

## Completed baseline

- S0 onboarding: [receipt](https://github.com/gastownhall/beads/issues/7403#issuecomment-6068385996), registered role and initially unclaimed launch. Donna subsequently issued standalone `resume vickie`; the interactive session is held.
- S1 source map: [receipt](https://github.com/gastownhall/beads/issues/7403#issuecomment-6068513221), [immutable map](https://github.com/donnabox/agent-coordination/blob/809ed7c104255f6cf742063c16ad4ffc8664821b/context/vickie/runtime-boundary-map-20261008.md). Thirteen boundaries and five operation traces; 23 cited runtime files checked against remote blobs. Source analysis only, no runtime qualification.

The map's in-process-versus-service option framing and Type-first next steps are historical after Donna's direction above. Its source observations remain baseline evidence; do not resume that superseded plan.

## Current bounded decision plan

The initial S2–S5 plan is superseded by C1–C5 below. Each design artifact must have an immutable link here, with proposed versus decided behavior identified.

| Step | Deliverable | Completion predicate |
| --- | --- | --- |
| C1 — BDP client/provider contract | Diagram and responsibility table for CLI, command host, client library, BDP, optional provider helper library and provider-owned endpoint/store. | done-when: every module has an owner, interface, invariant/error responsibility and declared dependency; direct store coupling is excluded from client operations; Donna records accepted versus open choices here. |
| C2 — Current CLI over BDP | User-journey/command crosswalk, including Issue behavior, selectors, guards, output/errors and capabilities. | done-when: each in-scope CLI behavior has a BDP operation sequence or explicit gap; representative journeys document atomicity, worst-case round trips/data transfer and failure/retry behavior; no unavailable behavior is described as implemented. |
| C3 — Packs and extensible commands | Source-grounded reuse proposal for Gas City packs, manifest/lifecycle and command execution interface; Issues pack feasibility. Type payloads are treated as a dependency on Trish. | done-when: inspected pack provenance is recorded, reused versus changed pieces are listed, command execution/dispatch/dependencies/versioning/trust boundaries are explicit, and a representative Issue command is worked through BDP without silently relying on a private store interface. |
| C4 — Rust libraries and coexistence | Rust module/library responsibilities, optional provider support, service/Scope targeting and side-by-side current-bd migration. | done-when: alternatives and recommendation cover executable/mode selection, config/auth/scope targeting, capability negotiation, compatibility, phased adoption and recovery; each phase has entry/exit criteria and proposed owner, with spellings and library commitments still labeled open where unruled. |
| C5 — Review and handoff | Consolidated client design and explicit client/Postgres/Type/Preview 2 dependency record. | done-when: the exact reviewed revision and finding dispositions are linked; Donna accepts the design here; Janet acknowledges the Preview 2/coexistence boundary here; downstream unresolved items have named owners and bus handoff links. Implementation requires separate authorization. |

Human-attestation debt: two final acceptance predicates (Donna's design acceptance and Janet's boundary acknowledgment). Type design is delegated to Trish rather than an immediate Vickie completion gate; any later concrete Type interface dependency will be recorded explicitly. Brian's team's storage implementation is outside this plan; no commitment on its behalf is implied.

## Boundaries and next discussion

Do not edit Preview 2 branches, PRs, specs, tests, pinned playgrounds or shared databases. Do not edit Trish's design/PR60. Keep design records separate; current BDP specifications do not change through this issue. Use remote evidence or isolated source snapshots. No V2 implementation or merge.

Donna answered the generic-provider question: **yes, this is a requirement**. Provider-side Issue extensions are not the fallback. Her next requested unit was a thorough current CLI/test-contract scrub for efficient generic BDP implementation, plus a separate Events/log/batch comparison.

C2 audit evidence is now [published](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/README.md), including [279-path main/Preview crosswalk](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/command-crosswalk.csv), four source-cited family reports, request/cost/guard analysis and executable evidence probes. This is a completed bounded audit, not a declaration that every C2 compatibility/efficiency question or runtime parity gate is closed. Remaining decisions include generic query/read consistency, compact transactional assertions, queue selection, partial-close semantics, volatility/identity, and the provider-admin boundary. Detailed Type lifecycle stays deferred.

The independent [journaling report](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/event-log-audit.md) distinguishes Event activity projections, complete replication groups, original command intent and full provider restore. It is a supporting deliverable under this plan, not another plan of record.

Next: use the evidence to settle the cold-client query/consistency contract and bounded mutation guards before choosing the Rust client/pack interfaces. No protocol amendments or implementation authorized. Qualification remains future work against two generic providers with no Issue-specific routes.


## Former comments

### 2026-10-08T20:24:24Z — donnabox

Original comment: https://github.com/gastownhall/beads/issues/7403#issuecomment-6068385996

S0 onboarding receipt — October 8, 2026.

- Remote [roster](https://github.com/donnabox/agent-coordination/blob/86d1dceae5d5ef2477344d2868faa2919ea7c908/roster.json) verified: exactly one Vickie entry, on `Donnas-MacBook-Pro`, backend `codex`; registration used `scripts/coordination-register.sh vickie --backend codex`. Verified roster blob: `7f5658c2e34114cafcf81576e598558b1cca2a08`.
- [Launch file at commit 86d1dceae5d5ef2477344d2868faa2919ea7c908](https://github.com/donnabox/agent-coordination/blob/86d1dceae5d5ef2477344d2868faa2919ea7c908/launch/vickie.md) verified byte-for-byte against prepared content: 6,456 bytes; blob `b9d3a27ac396f6c0a330599c7ee21a279ae48916`. Its single plan of record is `gastownhall/beads#7403` and its session is `unclaimed`.
- [Bus ownership/dependency notice](https://github.com/donnabox/agent-coordination/issues/1#issuecomment-6068380922) routes to Trish and records Janet's preserved release/specification/qualification ownership.

S0 is complete; S1–S5 have not begun. No interactive session claim, V2 implementation or merge, or edits to Preview 2 artifacts, pinned playgrounds or shared databases occurred. Onboarding now stops. Donna will separately issue standalone `resume vickie` before design work starts. The initial bus cursor intentionally precedes the onboarding notice so the first interactive resume reads it incrementally.


### 2026-10-08T20:32:09Z — donnabox

Original comment: https://github.com/gastownhall/beads/issues/7403#issuecomment-6068513221

S1 source boundary map published after Donna's standalone `resume vickie`.

[Runtime boundary map](https://github.com/donnabox/agent-coordination/blob/809ed7c104255f6cf742063c16ad4ffc8664821b/context/vickie/runtime-boundary-map-20261008.md) traces 13 boundaries and five representative paths: Memory create, CLI/HTTP read, native Issue update, owned-Link mutation and Memory deletion/retained reads. It distinguishes source observations, current BDP contracts, Type proposals and unmeasured behavior.

Primary runtime baseline: `donnabox/beads:release/preview2-integration` at `23c54902ca14087c5766eec6be2784ed71d7a216`. BDP main: `182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`. Type PR60: `40061cf111da52cb5a364ba7639b99435175a5f2`. CLI PR102 metadata verified at `94d880d4b7bdb87468281c1a7453b94387aacb2e`, without claiming a complete CLI contract audit.

[Source verification](https://github.com/donnabox/agent-coordination/blob/809ed7c104255f6cf742063c16ad4ffc8664821b/context/vickie/source-verification-20261008.json): 23 cited runtime files match the remote Git blob IDs; 32 runtime citation line bounds and all reference labels checked. Published content read back byte-for-byte. No build, tests, database opens or runtime probes were run; this is not release qualification.

Principal findings: the concrete graph store currently owns transactions and operation-specific validation; native Issue execution already participates within those transactions; HTTP Bead projection's separate Type read depends on the immutable installed catalog; owned-Link writes can version their source. Six explicit runtime questions carry forward to S2/S3, especially the affected-resource boundary for Type adoption.

S1 is complete at the document's stated scope. Next: compare current structure, an in-process graph operation layer and a service-owned runtime against those questions. No option selected, product ruling inferred or Preview 2 artifact changed. S2–S5 and the three acceptance attestations remain open.

_codex-unknown-model-unknown-reasoning on behalf of Donna Box_


### 2026-10-08T20:55:38Z — gascityinc-sarah

Original comment: https://github.com/gastownhall/beads/issues/7403#issuecomment-6068886026

<!-- gc-triage:v1 -->
**Triage:** `kind/chore` · `priority/p3` — Internal design\-stream tracker: its deliverables are design records and it authorizes no implementation, so there is no user\-visible behavior change and nothing user\-impacting to expedite\.
_Auto-triaged by @gascityinc-sarah. Reply or relabel if this is off._


### 2026-10-08T21:03:59Z — donnabox

Original comment: https://github.com/gastownhall/beads/issues/7403#issuecomment-6069015755

Donna's October 8 architecture direction, recorded from the active design conversation:

1. BDP is the provider interface. A new store exposes a BDP HTTP head; the long-term client does not integrate through the current Beads store-provider interface.
2. The CLI is a surface over BDP. Commands need not map one-to-one to protocol operations, but the current CLI functionality must be expressible efficiently through BDP.
3. Packages build on the core of Gas City packs and carry Bead/Link Type definitions, skills and commands for CLI extensibility. Implementing Issue-specific commands as user-defined commands in an Issues pack is the desired direction; its feasibility is to be established, not assumed.
4. A shared library may make provider implementations easier and more consistent. The store provider writes and owns the HTTP endpoint. Library scope and obligations remain to be designed.
5. Current `bd` and the new client need a side-by-side story. A graph-mode-like selector plus a target BDP service/Scope URL is a candidate, not settled flag spelling or executable packaging.
6. The new work targets Rust.
7. Brian's team will build a Postgres store against BDP semantics. Our team focuses on the client side; Vickie does not take over that store implementation.

Type lifecycle design stays with Trish and is deferred from Vickie's immediate exercise. Janet's ongoing integration of generic graph/versioning/Memory/BDP into current Beads remains her Preview 2 stream. The source map is useful baseline evidence, not the target architecture.

This supersedes the earlier next step of choosing between an in-process Beads runtime and a service-owned runtime as the external provider interface, and removes Type dependency resolution as Vickie's immediate gating exercise. Internal library composition remains open within the BDP-first direction. Detailed package execution, compatibility, provider helper-library scope and CLI-to-BDP coverage still need decisions. No implementation or merge is authorized by this design update.

First discussion seam: how much domain behavior an Issues pack can express through general BDP operations, versus whether any provider-side domain extension is allowed. This is an open question, not a claim that current BDP lacks the needed capability. A current-command crosswalk should test semantics, atomicity, round trips and data transferred before proposing protocol changes.

_codex-unknown-model-unknown-reasoning on behalf of Donna Box_


### 2026-10-08T21:32:27Z — donnabox

Original comment: https://github.com/gastownhall/beads/issues/7403#issuecomment-6069460308

Donna's requirement is recorded: **the Issues pack must work against any provider meeting its generic BDP capability requirements, without Issue-specific provider code.** This audit treats missing efficiency or semantics as protocol/client design questions.

Published audit: [overview and findings](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/README.md), [all registered command paths](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/command-crosswalk.csv), [Issues/workflows](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/issue-cli-audit.md), [other CLI families](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/other-cli-audit.md), [supporting commands](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/supporting-cli-audit.md), and the separate [Events/journal/batch report](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/event-log-audit.md).

Measured scope: 274 unique command registry paths on main plus five Preview additions, with static inventories of 3,453/3,621 top-level test declarations and 1,798/2,077 flag-declaration calls. These are inventory counts, not runtime coverage or exhaustive flag combinations. Every registry path has a disposition; high-risk behaviors are tied to existing implementation and tests. Baselines: beads main `5126de8c1a02d01db2481e4fd5e78dd41010e25f`, Preview `6c33ea17af3bffbdea46a5ac04715c6dccd7c9bf`, BDP `182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`.

Findings:

- Bounded generic mutations fit well: create-with-links, formula batches, scalar/metadata updates, comments, known-candidate claims and small state machines.
- Read efficiency is the main pressure: no caller ordering/text functions/joins/traversal/aggregation/projection. The measured synthetic full-scan fallback for one substring match among 50,000 records is 59,300,007 resource-JSON bytes and 500 pages at page size 100; this is not HTTP latency or a universal algorithmic lower bound.
- Graph preflight needs membership/readset protection, not only target revision. Zero-match set cardinality and no-op revision guards can protect bounded cases, but grow with readset, hit limits and require write permission. Single-winner claim is easier than selecting the correct queue front at commit.
- Partial-success multi-close, ready's persisted defer wake, no-history heartbeats, same-ID promotion/rename and native engine administration need explicit compatibility decisions.
- Events serve activity projection; snapshot plus complete change groups serves durable materialized-state journaling. Batch intent is not recoverable from effects: the checked fixture maps three operations to six Events to three final postimages. The report also compares actual bd batch, legacy Issue batchApply and journalops rows.

Validation: **219 existing BDP tests passed** (61 Selector implementation, 158 wire/catalog); **20 additional probes passed** (12 client schema/parser/model, eight fixture/reconstruction). Transactional server execution is not implemented in the checked-in reference server; catalog/fixtures explicitly do not establish runtime conformance. No Go CLI/database suite or end-to-end BDP parity run is claimed. [Verification](https://github.com/donnabox/agent-coordination/blob/dd3d2876c05aede176498e7ed2495f4828c9e44b/context/vickie/bdp-client-audit-20261008/verification.json) checked 287 citation anchors and 160 unique cited source files against remote Git blobs. All 26 published artifacts were verified by Git blob hash at `dd3d2876c05aede176498e7ed2495f4828c9e44b`.

C2's bounded audit is delivered. Full C2 behavioral/efficiency acceptance remains open pending the documented product choices, budgets and future provider qualification. Preview 2, Type artifacts and databases were untouched; no V2 implementation or merge.

_codex-unknown-model-unknown-reasoning on behalf of Donna Box_

