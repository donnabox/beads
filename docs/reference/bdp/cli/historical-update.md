# Historical content in graph Memory updates

In an experimental graph workspace, `bd update ID --from-revision TOKEN` uses an exact retained revision of the same Memory as its editable content base. This initial implementation supports live Memory updates; deleted recovery and native Issue updates are not implemented by this slice.

```sh
bd versions plan --json
bd update plan --from-revision TOKEN --if-revision CURRENT_TOKEN --json
bd update plan --from-revision TOKEN --properties '{"body":"Revised historical content"}' --json
```

`--from_revision` is an alias for the same option. The source token is opaque: it is neither a local history ordinal nor a current revision guard. A source-only request is valid. `--if-revision` retains the ordinary optional guard behavior and checks the actual current revision before acquiring the historical source.

Composition starts with the source's editable properties and common metadata, then applies explicit property and metadata edits. The result keeps the current Type and owned Links. For an Evaluation, the current protected subject group remains bound to the identity, including when the source predates that Evaluation. The ordinary current Type and final graph validation still apply. A historical request cannot change the Type.

An effective write creates one fresh revision. The native preview result's `attribution.historicalUpdate` names `fromRevision` and the actual `previousRevision`; actor and time describe the new operation. These are native preview fields, not a claim that a new public BDP provenance profile or wire encoding has been activated.

If the final content equals the current content, the operation returns `changed: false` and preserves the entire current envelope, including earlier historical context. Selecting a source alone never creates an audit revision.

An unknown, foreign, invalid, or unavailable source refuses without writing. Link historical updates refuse. Ordinary native Issue updates with this flag also refuse in this initial slice; shipping `bd restore` behavior is unchanged.

The approved recovery direction is tracked in [beads #7719](https://github.com/gastownhall/beads/issues/7719). This page documents the implemented live slice and does not claim the recovery feature is complete.
