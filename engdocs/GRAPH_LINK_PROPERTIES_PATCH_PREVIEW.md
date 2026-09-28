# Ordered informational Link properties patch preview

This private preview adds `bd update links/PATH --patch` to the existing
informational Link writer. It reuses the ordered property operation evaluator
introduced for Memory: pinned BDP commit
`53bdbd03136875f952af184fce7b3c7af8f74e96`, without advertising HTTP Write,
arbitrary Link Types, full Memory or native History.

For a Memory-owned Link, use both observed guards:

```sh
bd update links/context --patch '[{"op":"add","path":"/note","value":"Related context"}]' \
  --if-revision LINK_VERSION --if-source-revision MEMORY_VERSION
```

These are usage examples, not recorded results. `--patch @changes.json` and
`--patch @-` use the same bounded file/stdin loader as Memory. Input is a nonempty
ordered array of `add`, `replace` and `remove`; unsupported members/operations,
duplicate JSON keys, invalid Unicode and inadmissible numbers are refused.
`--properties`, Issue fields and other writer flags cannot be combined with
`--patch`. Existing full replacement remains available separately.

## Guards, ownership and results

Exactly one Link guard is required: `--if-revision TOKEN` or `--unconditional`.
The source guard is independent: `--if-source-revision TOKEN` or
`--unconditional-source`. Its syntax is checked before input acquisition; its
necessity is determined by the actual source Type inside the write transaction.
A Memory-owned Link requires it. An informational Issue source does not require
it, but a supplied guard is still checked. No source pre-read or refreshed guard
is introduced.

The completed properties must be `{}` or an object containing only a UTF-8
string `note`, as required by the existing installed informational Type. Empty
string and absent note differ. Removing an absent property fails. Temporary
objects, arrays or other members are allowed only when later operations produce
a valid completed representation. Type, identity and endpoints cannot be patched.
A healthy blocking Dependency refuses this informational-only writer.

The existing preview JSON envelope carries `LinkMutationResult`: complete Link,
complete source, `changed`, and optional `replacedSource`. Changed Memory-owned
Links create a new source version containing its full current owned set; an
informational Issue source stays unchanged. A guarded no-op retains both complete
records and attribution. A stale guard is refused even when the requested final
properties equal the current properties.

`replacedSource` describes the actual prior Memory only for a changed operation
with `--unconditional-source`. Link `--unconditional` alone does not authorize
source replacement or produce this disclosure. No-op, refusal, unknown outcome
and Issue-source updates do not manufacture a Memory replacement receipt.
Human output says `Updated` or `Unchanged` with the canonical Link ID and the
existing actual-predecessor summary where applicable. Quiet output retains the
existing policy. There is no automatic retry after an uncertain COMMIT.

## Admission and private limits

Write permission, canonical Scope selection, incompatible flags and guard
syntax are checked before opening input files or reading stdin. The Link Type
and source ownership require the checked storage transaction. Ordinary legacy
mode refuses graph patch intent before input or storage. Source guards on Memory
patches remain unsupported.

`status --graph` exposes `linkPropertiesPatch: true`. The seven `linkPatch*`
limits equal the Memory patch limits: `InputBytes` and `DocumentBytes` are 1 MiB,
`Operations` is 256, `PointerBytes` is 4096, `PointerSegments` and `Depth` are 64,
and `EvaluationBytes` is 16 MiB. Before each operation, the evaluator charges
current canonical document bytes plus supplied value bytes. This cumulative
work proxy is not a heap or elapsed-time bound. A changed result must also fit
the existing 16 MiB whole-workspace current-read acquisition budget, accounting
for owned snapshot copies. No-op patches perform evaluation but no additional
post-write workspace check. This changed-result postcondition is private to the
new patch route; existing replacement policy is unchanged. A shrinking patch
can restore readability only if the complete resulting workspace fits the
current-read budget. Shrinking alone is not a general recovery guarantee.

Input acquisition, syntax and Parse limit failures use `invalid_properties`
(exit 2). Invalid application or completed properties also use that code.
Checked-writer document/depth/evaluation/workspace bounds use
`capability_unavailable` (exit 5). CLI guard syntax is `invalid_selector` (exit 2);
a missing required source guard found by the writer is `invalid_properties`
(exit 2), and stale Link/source guards are `revision_conflict` (exit 4).
A healthy blocking Dependency reaches the installed-Type check and refuses with
`invalid_properties` (exit 2); it does not use the Memory wrong-kind code.
Malformed `links/` selectors on the patch route use `invalid_selector` (exit 2).
The older complete-replacement route retains its existing
`capability_unavailable` (exit 5) classification for that malformed prefix. This
slice does not change the replacement route. These categories describe admission
boundaries, not a universal limit code.

## Evidence scope and compatibility

`scripts/graph-link-properties-patch-smoke.py` authors disposable workspaces
through the installed CLI on embedded and caller-owned ordinary shared-server
Dolt sequentially. It captures command/stdout/stderr/input/binary/source hashes,
complete expected records and fresh-process exact-version reads. It exercises
owned and unowned Links, separate identities with equal endpoints, self-Links,
dual guards, absent/empty note, no-ops, refusals and human/quiet output. A
CLI-authored blocking Dependency is read before and after the wrong-Type patch
refusal; its complete record and Issue source remain unchanged, with both
retained for fresh-process exact-version reads. The cumulative-work refusal
checks its specific evaluation reason, not only the shared limit error code. Forced
transaction overlap, rollback, lost acknowledgments and whole-workspace budget
acceptance belong to the separate real-engine storage tests. The script does
not imply those internal guarantees from a successful CLI recording.

The previous Memory CLI test `link-selector` and installed receipt
`link-before-input` retain their early-refusal slots but now include
`--properties={}` alongside `--patch`. Link patching alone is newly supported;
this explicit incompatible combination remains refused before input. All other
Memory patch expectations remain unchanged. Exact-source verifiers must record
this semantic substitution, not silently drop the old checks.

Full Memory metadata/Inception/derivation, ordered native History, adoption,
custom Type installation, public authenticated/idempotent Write, aliases and
Jim's integration-target review remain separate open delivery work.
