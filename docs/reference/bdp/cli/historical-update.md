# Historical content in graph updates

In an experimental graph workspace, `bd update ID --from-revision TOKEN` uses an exact retained revision of the same Bead as its editable content base. This implementation supports live Memory and admitted native Issue updates, plus same-ID Memory and admitted native Issue recovery.

```sh
bd versions plan --json
bd update plan --from-revision TOKEN --if-revision CURRENT_TOKEN --json
bd update plan --from-revision TOKEN --properties '{"body":"Revised historical content"}' --json
```

`--from_revision` is an alias for the same option. The source token is opaque: it is neither a local history ordinal nor a current revision guard. A source-only request is valid. `--if-revision` retains the ordinary optional guard behavior and checks the actual current revision before acquiring the historical source.

Composition starts with the source's editable properties and common metadata, then applies explicit property and metadata edits. The result keeps the current Type and owned Links. For an Evaluation, the current protected subject group remains bound to the identity, including when the source predates that Evaluation. The ordinary current Type and final graph validation still apply. A historical request cannot change the Type.

An effective write creates one fresh revision. The native preview result's `attribution.historicalUpdate` names `fromRevision` and the actual `previousRevision`; actor and time describe the new operation. These are native preview fields, not a claim that a new public BDP provenance profile or wire encoding has been activated.

For a live Bead, if the final content equals the current content, the operation returns `changed: false` and preserves the entire current envelope, including earlier historical context. Selecting a source alone never creates an audit revision.

An unknown, foreign, invalid, or unavailable source refuses without writing. Informational Links support the checked historical flow below. Native Dependency historical updates require their separate native writer and currently refuse. Shipping `bd restore` behavior is unchanged.

## Recovering a deleted Bead

Deletion creates a fresh terminal revision D and keeps the final live snapshot separate. `bd versions ID --json` returns D in `result.currentState.revision`; `result.terminal_versions` lists retained terminal rows separately from citable live snapshots. Human output also shows the current deletion revision. Ordinary `bd show ID` remains a `gone` refusal. A terminal token is never a live historical content source, including after recovery.

```sh
bd delete plan --force --json
bd versions plan --json
bd update plan --from-revision LIVE_SOURCE_TOKEN --if-revision DELETION_TOKEN --json
```

The current guard, when supplied, checks D. The last live revision and the historical source cannot substitute for D. Recovery preserves the allocated ID and creates one fresh live revision even when the selected content equals the final live content. The native result records `attribution.inception: "revert"`, with the historical source and D recorded separately. New ordinary writes preserve the current Inception and clear any previous operation's historical context; no-ops preserve the complete envelope.

The service separately acquires the retained last-current Type and protected initialization facts. Missing or corrupt baseline/source state refuses atomically. A legacy preview deletion that never retained a distinct D remains readable through its existing history, but cannot be recovered by pretending its last live token is D. Recovery does not recreate removed Links or historical owned membership. The first Evaluation subject group remains bound to the identity through deletion, recovery and later Type changes.

A lost commit response remains an unknown outcome. The service does not automatically replay it; a supplied stale D guard refuses if a recovery already committed. Existing guard omission stays valid under ordinary unconditional Update behavior.

The approved recovery direction is tracked in [beads #7719](https://github.com/gastownhall/beads/issues/7719). Executable BDP schemas/SDK and full release qualification remain separate work.


## Live native Issue updates

A native Issue supports source-only historical updates, `--properties` edits, ordered `--patch`/`--replace-properties` edits and common-metadata edits. Admitted editable fields are title, description, design, acceptance criteria, priority, nullable estimate, assignee, notes, external reference, spec ID and due date. Existing notes overwrite and active-assignment protections apply; use `--force` only when deliberately replacing notes. Current labels, parent and owned structure stay current, including when the selected snapshot contains older Links.

Changed native members outside that setter list, such as status or issue type, refuse with a native-domain-mapping diagnostic. Ordered operations apply to the selected historical editable properties inside the transaction. A replacement must retain required title and priority; omitted optional fields clear through their ordinary native setters. A removed retained native source refuses using its recorded availability information; the service never substitutes another revision. The implementation crosswalk is recorded in `engdocs/GRAPH_NATIVE_HISTORICAL_UPDATE.md`.


## Native Issue recovery boundaries

Native graph deletion also returns a distinct terminal D. The final live Issue remains a separate historical snapshot, and ordinary deleted `show` still refuses as gone. Recovery uses the same `update --from-revision` command and optional actual-current guard shown above. It preserves the graph allocation and native Issue ID, current Type and retained creation facts, then creates a fresh live revision with Inception `revert`. Equal content is still an effective deleted-to-live transition.

The service separately requires the authorized last-current native baseline and its participation evidence. Missing, removed, malformed or unavailable source/baseline state refuses atomically. A legacy deletion without D is never upgraded by treating its final live token as a deletion revision. A native row without retained participation admission cannot be silently enrolled through recovery.

Native notes overwrite and active-assignment protections continue to apply against the last-current baseline. Use `--force` deliberately when historical notes replace current notes, including when `--properties` explicitly edits a different field. Current native workflow classification/status remain under their ordinary domain authority: a source that requires an unsupported status/classification rollback refuses.

Deletion removed native labels and node-local leases; recovery does not replay them from the source or last-current payload. Removed parent/Dependency/owned-Link structure also stays removed. Native workflow status and assignee alone do not regrant an old lease. The native journal records materializing the missing row; graph Inception separately identifies the same-allocation recovery. Shipping native `bd restore` remains unchanged.

Graph Issue history can now contain both native snapshots and graph terminal revisions. Its local ordering includes D; underlying native Issue history keeps its own ordinals. Terminal tokens remain unavailable as editable content sources before and after recovery.

## Informational Link updates and recovery

Use a retained Link revision to restore its editable properties and metadata, or to recover a deleted informational Link with its original ID:

```sh
bd link versions LINK --json
bd link update LINK --from-revision LIVE_SOURCE_TOKEN --if-revision CURRENT_TOKEN --json
bd link update LINK --from-revision LIVE_SOURCE_TOKEN --patch '[{"op":"replace","path":"/note","value":"Revised historical note"}]' --json
# After deletion, obtain D from currentState.revision in link versions:
bd link update LINK --from-revision LIVE_SOURCE_TOKEN --if-revision DELETION_TOKEN --json
```

The source supplies properties and common metadata before explicit edits. The current Link Type, source/target identities, stored endpoint pins and ownership remain current; recovery uses the separately validated final live baseline for those facts. Selecting an old Link revision does not restore its old pins. Existing explicit `--source-revision` and `--target-revision` flags remain independent pin edits; owned-source pins still refuse.

For an owned Link, `--if-source-revision TOKEN` independently checks the current owning Bead. An effectful Link change creates its fresh Link revision and one new owning Bead revision in the same transaction. Source-only live equality preserves both complete envelopes. Recovery always appends a fresh live Link revision, even for identical content.

`bd link versions LINK --json` returns the current deletion descriptor and lists every retained D separately in `terminal_versions`, including old D rows after recovery. A D token is never citable live content. Missing/deleted endpoints, stale guards, unknown or deleted source revisions, unavailable pins and invalid current Type constraints refuse atomically. Native Dependency Links currently refuse historical updates rather than bypassing their Issue-domain writer.

The informational Link implementation crosswalk is recorded in `engdocs/GRAPH_LINK_HISTORICAL_UPDATE.md`.
