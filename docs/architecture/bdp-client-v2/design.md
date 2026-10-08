# BDP client platform and protocol design

Status: exploratory design, owned by Vickie with product rulings by Donna. Last updated: 2026-10-08.

This Markdown document on `donnabox/beads:codex/vickie-bdp-client-design` is the single plan of record at Donna's request. It replaces the upstream issue and intentionally overrides the fleet's issue-shaped tracker convention. It is not a normative BDP amendment, release commitment, or implementation authorization. Work remains on this fork branch until Donna asks otherwise; do not create an upstream issue or PR for it.

## Agreed direction

BDP is the provider interface. A provider writes and owns its HTTP endpoint over its chosen store. Our Rust CLI/client is a surface over BDP; commands need not correspond one-to-one with requests. Any provider meeting the required generic BDP capabilities must run the Issues pack without Issue-specific provider code.

Packs build on Gas City core concepts and carry Bead/Link definitions, skills and extensible CLI commands. Issue commands belong in an Issues pack. A shared provider library is possible, but does not take over the endpoint. Brian's team plans the Postgres store; our team designs the client. Current `bd` must coexist side-by-side. Mode/URL/Scope configuration is a candidate, not settled syntax.

Trish owns Type lifecycle design. Janet owns current Beads/Preview 2 integration, CLI/BDP alignment and qualification. Do not edit their branches, PRs, specs, tests, playgrounds or shared databases. Coordinate dependencies through the existing bus. No V2 implementation or merge is authorized.

## Evidence

The [CLI audit](audit-20261008/README.md) classifies 274 main registry paths and five Preview additions, with source/test contracts and cost analysis. 219 existing BDP tests and 20 additional bounded probes passed. This is parser/schema/fixture/model evidence, not end-to-end Transactional provider qualification. [Verification](audit-20261008/verification.json) pins source commits and checks 287 citation anchors and 160 source files. The [earlier source map](baseline/runtime-boundary-map.md) records 13 runtime interfaces and five paths.

The historical [journal audit](audit-20261008/event-log-audit.md) distinguished Events and change groups. The current review must sharpen that distinction: missing original command intent is not itself an obstacle to forward state replication. Neither before-images nor original selectors are intrinsically necessary to apply complete committed effects. The question is whether today's Event Source contract carries every state/control transition and an adequate atomic replay boundary.

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

The protocol proposals and replica reassessment will be developed below. Decisions remain proposals until Donna rules; their syntax is illustrative and does not change the current schema.
