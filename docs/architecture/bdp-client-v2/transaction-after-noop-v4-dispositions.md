# Deletion-safety plan v4 — native-v3 inventory correction

Status: plan-only successor, awaiting independent native closure and root acceptance. No implementation authorization.

Native-v3 closure identified one omitted mechanical pin: `packages/conformance/src/transactional-cumulative.test.ts:930` expects receipt-pagination at scenario index8. Inserting deletion-safety at index2 moves that existing row to index9. V4 adds an explicit8→9 inventory entry at lines930–931, preserving the receipt row’s18 actions, raw manifest/coverage bytes, semantic assertions and existing preservation hashes. No change to request/action/cohort limits, scope, semantics or evidence claims.

V4 is otherwise the v3 plan, apart from version/status wording. The full native P3/external F1–F9 dispositions in `transaction-after-noop-v3-dispositions.md` remain applicable. V3 `transaction-after-noop-next-plan-v3.md` SHA256 `f8bb465200455bdb1d57d07a614a0fc079c149dab48be2cd746acd4d55246ed7` is unchanged, as are original and v2. No source edits, tests, builds or network operations were performed.
