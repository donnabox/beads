# Historical content in graph updates

In an experimental graph workspace, `bd update ID --from-revision TOKEN` uses an exact retained revision of the same Bead as its editable content base. This implementation supports live Memory and admitted native Issue updates, plus same-ID Memory recovery. Native Issue recovery is not implemented by this slice.

```sh
bd versions plan --json
bd update plan --from-revision TOKEN --if-revision CURRENT_TOKEN --json
bd update plan --from-revision TOKEN --properties '{"body":"Revised historical content"}' --json
```

`--from_revision` is an alias for the same option. The source token is opaque: it is neither a local history ordinal nor a current revision guard. A source-only request is valid. `--if-revision` retains the ordinary optional guard behavior and checks the actual current revision before acquiring the historical source.

Composition starts with the source's editable properties and common metadata, then applies explicit property and metadata edits. The result keeps the current Type and owned Links. For an Evaluation, the current protected subject group remains bound to the identity, including when the source predates that Evaluation. The ordinary current Type and final graph validation still apply. A historical request cannot change the Type.

An effective write creates one fresh revision. The native preview result's `attribution.historicalUpdate` names `fromRevision` and the actual `previousRevision`; actor and time describe the new operation. These are native preview fields, not a claim that a new public BDP provenance profile or wire encoding has been activated.

For a live Bead, if the final content equals the current content, the operation returns `changed: false` and preserves the entire current envelope, including earlier historical context. Selecting a source alone never creates an audit revision.

An unknown, foreign, invalid, or unavailable source refuses without writing. Link historical updates refuse. Shipping `bd restore` behavior is unchanged.

## Recovering a deleted Memory

Deletion creates a fresh terminal revision D and keeps the final live snapshot separate. `bd versions ID --json` returns D in `result.currentState.revision`; `result.terminal_versions` lists retained terminal rows separately from citable live snapshots. Human output also shows the current deletion revision. Ordinary `bd show ID` remains a `gone` refusal. A terminal token is never a live historical content source, including after recovery.

```sh
bd delete plan --force --json
bd versions plan --json
bd update plan --from-revision LIVE_SOURCE_TOKEN --if-revision DELETION_TOKEN --json
```

The current guard, when supplied, checks D. The last live revision and the historical source cannot substitute for D. Recovery preserves the allocated ID and creates one fresh live revision even when the selected content equals the final live content. The native result records `attribution.inception: "revert"`, with the historical source and D recorded separately. New ordinary writes preserve the current Inception and clear any previous operation's historical context; no-ops preserve the complete envelope.

The service separately acquires the retained last-current Type and protected initialization facts. Missing or corrupt baseline/source state refuses atomically. A legacy preview deletion that never retained a distinct D remains readable through its existing history, but cannot be recovered by pretending its last live token is D. Recovery does not recreate removed Links or historical owned membership. The first Evaluation subject group remains bound to the identity through deletion, recovery and later Type changes.

A lost commit response remains an unknown outcome. The service does not automatically replay it; a supplied stale D guard refuses if a recovery already committed. Existing guard omission stays valid under ordinary unconditional Update behavior.

The approved recovery direction is tracked in [beads #7719](https://github.com/gastownhall/beads/issues/7719). Native Issue recovery, executable BDP schemas/SDK, and full release qualification remain separate work.


## Live native Issue updates

A live native Issue supports source-only historical updates, `--properties` edits and common-metadata edits. Admitted editable fields are title, description, design, acceptance criteria, priority, nullable estimate, assignee, notes, external reference, spec ID and due date. Existing notes overwrite and active-assignment protections apply; use `--force` only when deliberately replacing notes. Current labels, parent and owned structure stay current, including when the selected snapshot contains older Links.

Changed native members outside that setter list, such as status or issue type, refuse with a native-domain-mapping diagnostic. Historical native `--patch` and `--replace-properties` also refuse in this slice. A removed retained native source refuses using its recorded availability information; the service never substitutes another revision. The implementation crosswalk is recorded in `engdocs/GRAPH_NATIVE_HISTORICAL_UPDATE.md`.
