# Discovery replay-floor token audit

## Conclusion

Correct C3's proposed `minimumReplayPosition: "genesis"` to this owner's **genesis position `"pos-0"`**, and separately verify replay with the **genesis checkpoint `"genesis"`**. No feed-codec change, fixture rewrite, new retention semantics, or operator product ruling is needed for the fixed no-pruning owner.

The external F1 identifies a real position/checkpoint conflation in the plan. Its inference that the client should feed `minimumReplayPosition` directly to `changes/?after=` is not established by the full contract. That proposed test would itself conflate the two kinds. Opaque spelling alone is not decisive: `pos-N` is not a required universal prefix, and the schema's shared wireToken cannot distinguish semantic roles. The decisive evidence is the corpus's explicit position/checkpoint split and equality relationships, applied to this owner's actual encodings.

This is a bounded plan correction, not evidence that currently served Transactional discovery is wrong: that serving path remains disabled. No C3 acceptance or public discovery activation follows from this audit.

## Exact inspected basis

Frozen `/Users/dbox/repos/wt/vickie/vickie-tx-collection-minimum` at `20d61119b5faf7251a5d285c5c34a1049866ffe5`; relevant normative text and store/changefeed codecs are unchanged from main `581bad0de6f5eb098b920147f6a7e09809c553ea`. Read `transaction-cold-discovery-next-plan.md`, external F1 in `transaction-cold-discovery-next-claude-review.txt`, exact normative clauses, fixture examples and protocol tests. No source/spec/test edits, builds, tests, network actions or processes were launched beyond read-only file inspection. A pure-code probe was unnecessary: the relevant codec branches are explicit.

## Distinct symbols and contracts

| Symbol/field | Meaning and primary evidence | Current owner's concrete value |
| --- | --- | --- |
| Scope position P | One authority-defined ordered position; clients compare only equality, never arithmetic or lexical order. Each epoch has a distinguished genesis position (`docs/specs/bdp.md:3680–3693`). | `pos-N`; initial P0=`pos-0` |
| `headPosition`, current `BDP-Scope-Position`, `observedHeadPosition` | Current projected prefix's position (`:3743–3747`, `:5791–5798`, `:6105–6108`). | `readContext().scopePosition`, constructed `pos-${metadata("position")}` at `transaction-store.ts:582–590` |
| `minimumReplayPosition` | Oldest position whose exclusive Scope-changefeed cursor is still legal (`:3743–3747`); after erasure at P it advances to at least P (`:6464–6478`). | P0=`pos-0` for the present no-pruning/no-erasure owner |
| `checkpoint` | Epoch/view-qualified replay handle naming the exclusive boundary at a position; Snapshot has both `scopePosition` and `checkpoint` (`:5854–5858`, `:5932–5935`). | Initial K0=`genesis`; positive Kn=`cg_<epoch-tag>_<view-tag>_N` |
| `after`, SSE change-group id / Last-Event-ID | Supplied replay checkpoint, not a request to parse position spelling (`:5966–5972`, `:6076–6078`, `:6105–6125`). | `genesis` or a valid `cg_…` value |

The phrase “oldest position still legal as an exclusive … cursor” describes the boundary's semantics; it does not erase the explicit checkpoint field or define a conversion from every position token to a replay request. A client that possesses a Snapshot/group checkpoint uses that checkpoint. A client requesting complete retained history can use reserved `genesis`. Clients cannot infer ordering by parsing `pos-N` or use discovery to silently advance an expired cursor; `:6123–6125` expressly prohibits the latter.

## Why examples are more than a prefix convention

- Normative discovery example `docs/specs/bdp.md:3729–3730` uses `opaque-position-*` for both head and floor, while the Snapshot/changefeed examples deliberately use separate `opaque-checkpoint-*` values. Neither prefix is prescribed syntax.
- `fixtures/transactional/discovery.json:33` and `fixtures/history/wire.json:83` use `pos-17` as floor. Those literals alone would not prohibit an implementation choosing another opaque position alphabet.
- Stronger evidence is `fixtures/transactional/changefeed.json:739,777`: the erasure floor is `pos-44`, while the same group uses checkpoint `ckpt-44`. `packages/protocol/src/transactional-wire.test.ts:628,658` explicitly requires `minimumReplayPosition === group.position`, and the same test separately requires the SSE id and durable replay token to equal `group.checkpoint`. This independently distinguishes the field's role, not just its spelling.
- `schemas/bdp-v0.schema.json:5023–5028` gives head and floor the common wireToken syntax. The shared alphabet is not a semantic union or a guarantee that arbitrary tokens are interchangeable. The wire profile at `docs/specs/bdp.md:6563–6584` expressly permits different native encodings and reserves `genesis` as the initial checkpoint; it never requires every initial position field to use that spelling.

Consequently, claiming `genesis` is forbidden solely because it lacks a pos- prefix would be wrong. But this implementation has already chosen `pos-0` for its initial position, and its group's previousPosition/current read context consistently use that choice. C3 should not introduce a second equality-unequal spelling for that same position field merely because the corresponding checkpoint is `genesis`.

## Current implementation is internally coherent without C3

`transaction-changefeed.ts:53–85` deliberately separates codecs:

- `transactionCheckpoint(scope, context, 0)` returns literal `genesis`; positive values include epoch/view hashes and a positive integer suffix.
- `checkpointPosition(scope, context, "genesis")` returns0; positive checkpoints must match the current Scope/epoch/view. `pos-0` is not a checkpoint and is correctly rejected by this function and the request classifier.
- `transaction-store.ts:1645–1650` converts the current position ordinal to a checkpoint for start=now. `:1667–1690` decodes after, verifies retained groups and returns group checkpoints in next URLs. `projectTransactionChangeGroup` returns both position and checkpoint separately.
- The no-pruning invariant at `transaction-store.ts:1672–1675` treats missing positive groups through head as corruption rather than an expired replay floor. Receipt maintenance does not delete groups. Hence the *position boundary* remains0 while replay from its paired checkpoint remains legal.

The reserved zero checkpoint intentionally has no encoded epoch/view hashes. Changing that established behavior is not necessary for this discovery assembly; explicit context controls and later positive checkpoints retain their existing contracts. This audit makes no new claim about prior-epoch use of a bare genesis request.

## Smallest corrected plan and evidence

Replace the plan's line31 instruction “use the existing checkpoint function at zero, not pos-0” with two named facts:

1. The current fixed no-pruning owner's advertised replay **position** floor is its initial position `pos-0`, in the same position namespace as readContext, group.position and group.previousPosition. Bind this fixed applicability explicitly in the private builder; do not derive it from the current head, latest receipt, Snapshot anchor or latest checkpoint.
2. Independently derive the replay **checkpoint** at that floor through `transactionCheckpoint(scope, capturedContext, 0)` and assert it is `genesis`. Use that checkpoint in actual readChanges/after requests; do not call checkpointPosition on the discovery floor itself.

Bounded direct tests can establish: empty-owner head equals the floor; the first committed group's previousPosition equals that same floor; subsequent descriptions advance head while leaving floor unchanged; `after=genesis` returns the first complete group after commits, ordinary reopen and receipt expiry; group positions/checkpoints maintain their distinct expected relations. These observations need no full-history scan in production or additional store accessor. Production may use the fixed private position value under the disclosed invariant; independent tests must tie it to actual owner observations rather than validate a constant against itself.

Do not mutate the existing illustrative fixtures to use this runtime's codec. Their distinct opaque examples and existing equality tests are coherent with the correction. Do not extend the feed parser to accept all `pos-N`, which would introduce a new checkpoint authority/epoch/view path and exceed C3. No schema, catalog, protocol parser, public API, store format or terminal identity change is needed for this floor correction.

If a future pruning/erasure implementation is introduced, its real authority must expose the applicable oldest retained position and corresponding replay policy; a hardcoded initial floor would then cease to be truthful. That future owner seam is outside this no-pruning C3 slice. If root instead wants discovery to expose a directly replayable positive checkpoint field, that is a separate protocol design change—not necessary to repair this plan.

## Residual ambiguity, narrowly scoped

There is some prose shorthand in “position … as … cursor,” but the full corpus already distinguishes the two serialized roles sufficiently for the existing implementation. There is no unresolved choice of retention policy or need for an operator ruling on initial-floor spelling here. This audit does not settle unrelated malformed/minimum request rules, retained Snapshot anchors, erasure implementation, or C3's other external findings.
