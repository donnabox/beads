# Completed-key conflict evidence — frozen implementation

Frozen source **b13055049ed7c1bd09ad2d7af29e0a38a81001ea**, tree **3bda07f860204eba0fcd5b2fe65335dadf99653f**, direct parent **2172092de12752867b43858ebcbdd2c37d4c727f**. Checkout `/Users/dbox/repos/wt/vickie/vickie-completed-key-conflict`; branch `codex/vickie-completed-key-conflict` was pushed through the remote-session guard and `git ls-remote` returned this exact source. No PR, main, launch, bus, runtime, service, schema or public admission change.

The accepted v2 plan is `transaction-after-receipts-next-plan-v2.md` SHA256 `77f6deb26eb19ad12f1642e237dcc6c589eecd65280329559fc757d1ea342dbb`, with root acceptance. Exactly four authorized evidence/test files changed. The sole new catalog row `transactional.idempotency.conflict` is in canonical order and remains **wholeContract:false**. All ten authored rows are partial; the runner still exposes all 250 required plus five optional rows (10 observed pass, 245 unimplemented; required-only unplanned count240). This is diagnostic evidence, not profile qualification.

## What the actual installed row observes

Twelve actions, eleven HTTP exchanges, one same-store restart. One explicit unowned Task is created. The same key is presented five times: three byte-identical original create bodies and two byte-identical changed create bodies targeting `beads/b`. The source guard pins all five exact raw bodies/headers, including the first conflict's `Accept:text/plain`, and detects changing that request back to the original semantics.

Both conflicts return actual409 `idempotency-conflict` with transactional Problem schema, conflict type, retry never and problem media. The text/plain request demonstrates conflict precedence over unacceptable success media in this configured completed-key path. Independent complete one-item current collection tuples and complete one-Event tuples prove no visible extra Bead/Event. Full current record equality ties both collection reads to the actual original receipt record. Whole-body/root comparisons preserve the old completed receipt before and after reopen. The reopened old-root GET uses the existing validated source-receipt authority ledger, checks later generation, same configuration and exact URL, plus explicit context header presence; Event context matches the original committed response with explicit presence on both sides. No refusal capture or receipt authority is minted.

The literal state/Event facts are independently authored; captures compare only opaque runtime values and complete retained postimages. The receipt-reader's existing source-wire validation remains unchanged. `mapping-audit.json` lists actual exchange status, byte counts, assertion outcomes and dynamic source/session provenance from the genuine packed report.

## Bounds and preservation

| Measure | Actual | Accepted effective ceiling |
| --- | ---: | ---: |
| Manifest bytes | 407227 | 410217 |
| Manifest value nodes | 10891 | 11282 |
| Coverage bytes | 60059 | 63067 |
| Total actions | 151 | 151 |
| New row actions/assertions | 12 / 99 | 12 actions; no assertion quota |
| Positive child launches/closed | 18 / 18 | 18 |
| Positive scenario/reopen sessions | 17 | 17 |
| Positive elapsed milliseconds | 3701.057 | 63000 |

All nine previous raw scenario objects and all nine previous raw coverage objects are byte-identical to the assigned main parent. The eight original hash pins remain unchanged; pagination's independently recomputed ninth pins were added, including an executable drift control. `final-source-audit.json` records every hash/length and exact mechanical coverage mapping. Fixture remains716bytes, SHA256 `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`. Exact route/configuration digest remains `634214546089490d235fa404166dc868a0feb5dc2a8af2a566a6ad33664b4b1a`. No route was added. Existing M/Z and receipt request/response bounds remain intact.

New POSTs are173bytes each (<=2048); maximum new actual response is696bytes (<=8192). Aggregate90s, outer120s, positive63s, wire5s and preparation/cleanup10s limits were not changed. The first formatted candidate was measured before execution; final source was remeasured after correcting target-query placement, with2990bytes/391nodes/3008coveragebytes effective headroom.

## Fresh exact-source execution

Clean source was frozen before building and packing. Node24.16.0/pnpm11.20.0 were used throughout. Fresh protocol/server tarballs were installed offline into `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-completed-key-conflict-implementation/install`; actual entry `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-completed-key-conflict-implementation/install/node_modules/@bdp/server/dist/transaction-installed-process.js`. `source.json` binds all four files; `summary.json` records exact commands, exits, tarballs, installed baseline and five controls. The sourceCommit is verified launcher metadata, not independent target attestation.

- PASS: clean pre-build typecheck; emitted build; frozen cumulative source tests **27/27**.
- PASS: fresh packed positive **1/1**, ten authored rows all pass,151actions/18children closed.
- PASS: fresh packed five sensitivity controls **5/5**. Each completed actual wire and failed exactly its intended assertion: `problem-code`, `http-status`, `ordered-state`, `retained-body`, or `ordered-facts`. Each mutated manifest has its own actual report digest; original bundle identity is not substituted for control identity.
- PASS: format, lint (exit0 with existing warnings/info), dependency boundaries, Read evidence verification, git diff check. Full workspace suite and the unrelated remaining43 app tests were not rerun.
- Earlier local positive and five controls also passed, but the fresh frozen package evidence above is authoritative for this source.

Positive installed payload digest `6945e90b33b51b1fc26cb92866399de58d36dee2da4fc09ebe126fe5ccba59ef`; runtime configuration `3124240008a8c851c0c255cdd0980511aed31d92632f933838360231c6f38bbe`; contracts `c67d12264b97e16c76bf68f850eaf394f61896bb72491e06603ef1d6debb335b`. Protocol tarball SHA256 `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`; server tarball SHA256 `502bb785fe3f54503440a7a3c9298c4c607a086470b1f5e5b00f73235ecdb222`. Runtime/package source is unchanged from main217; this does not assert byte identity with an unmeasured historical tarball.

## Retained failed attempts and cleanup

1. Initial source typecheck failed on direct access to a union report field. Fixed by recording the safe full declarations object.
2. Initial cumulative tests failed because two new collection targets placed query syntax in `path`; fixed to the existing separate `target.query` format. No target execution occurred in that failing factory path. Original22failed/5passed log/JSON is retained; corrected27pass log/JSON is separate.
3. All fresh packed tests passed, then the first external collector exited1 because it incorrectly required18children in every negative control. The two controls that fail before reopen correctly have17; others have18. All closed. Original `pack.py` and `collector-first-failure.json` remain; `collect.py` consumes the same genuine completed reports with the corrected17–18 ceiling, without rerunning or relabeling failures as target passes.
4. A later ad-hoc report inspection attempted a nonexistent client action `state` field and exited1; this was a read-only inspection mistake, not a test or target failure. The mapping report uses actual client assertions instead.

No runtime fault or source-scope drift was encountered. No test/build/installed child remains; `remaining-processes.txt` is empty. Checkout remains clean and frozen.

## Limits and follow-on

This row covers one principal, fixed view/epoch and explicit-ID create batch with available completed detail and same-epoch restart. It does not establish every duration while a key is bound, expiry/restore/new epoch, another principal/view, pending/unresolved/concurrent behavior, other operation families or carriers, rate/capacity, internal receipt absence, History/erasure. False flags and these exclusions are in the new coverage reason. Independent code council/root acceptance remain required before integration; no serving or public readiness follows automatically.
