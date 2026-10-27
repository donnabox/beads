# Preview 3 operational companion

These optional exercises accompany the [qualified blog-based runbook](RUNBOOK.md#qualified-source-and-evidence) and the official `graph-preview3` source release. The embedded manual workspace does not start an HTTP service or qualify your local server setup; the retained HTTP exercise uses a separate fresh workspace and separately owned processes. Never reset, reinitialize or open an old workspace with the new binary. Historical transcripts remain source-attributed; obtain actual IDs/revision tokens from your session.

## Appendix A. Read the graph over HTTP

The CLI exercises use embedded Dolt. BDP exposes graph records over HTTP and requires a shared Dolt server, so use a separate fresh workspace for this optional exercise. It does not migrate the workspace used in the [main walkthrough](RUNBOOK.md#using-the-playground). You need the Preview 3 executable, a compatible Dolt installation and `curl`.

Three terminals make the processes visible: one runs Dolt, one runs the read-only BDP server, and one makes requests. These examples use local ports 13395 and 18795; choose unused ports and change every matching argument if necessary.

### Terminal A: start Dolt

From a directory outside an existing workspace, create a new trial directory:

```sh
mkdir beads-preview3-bdp
cd beads-preview3-bdp
mkdir server-data workspace
dolt sql-server -H 127.0.0.1 -P 13395 --data-dir "$PWD/server-data"
```

Leave this terminal running. In Terminal B, navigate to the `workspace` directory you just created.

### Terminal B: initialize and serve

```sh
git init -q
bd init --graph-mode link --scope-url 'http://127.0.0.1:18795/graph/' \
  --server --external --server-host 127.0.0.1 --server-port 13395 --server-user root
bd remember 'Review before publishing.' --id policy-bdp --title 'BDP policy'
bd show policy-bdp --json
bd serve --readonly --addr 127.0.0.1:18795
```

Save the Memory's full canonical ID from the show result. Its Scope URL matches the address at which this server will expose the graph. Leave this terminal running too.

### Terminal C: discover and read

```sh
curl --fail-with-body http://127.0.0.1:18795/graph/bdp.json
```

Discovery advertises the selected Read capabilities and inventory URLs. Copy the advertised Bead inventory URL and request it with `curl --fail-with-body`. Follow its advertised next link when a page has one; do not construct a continuation token yourself. Then request the Memory's exact canonical ID saved in Terminal B:

```sh
curl --fail-with-body 'PASTE_THE_CANONICAL_MEMORY_ID_HERE'
```

Using the returned identity and advertised collection routes keeps the exercise tied to the server's actual projection. The selected Read client accepts literal rootless and segmented IDs. Do not infer Resource kind from a URL segment.

Compare the HTTP record's identity, title, body and revision with the CLI output. This is a selected BDP Read exercise. HTTP writes, a History service and blanket profile conformance are outside its scope. Finish by stopping the BDP server in Terminal B with Ctrl-C, then Dolt in Terminal A with Ctrl-C.

## Appendix B. Revisions, recovery and pins

An ID identifies a continuing Bead; a revision identifies a saved state of it. The convenient CLI form `ID@N` uses that Resource's stable local revision number. `ID@=TOKEN` selects an exact token, including a token consisting only of digits. The selected convenience forms are shell-friendly; quote literal IDs when their own characters require it. An ID containing `@` remains an allocated identity, not permission to reinterpret its suffix. Select it using the literal-selector form supported by the qualified command or its exact returned canonical URI; do not invent a new alias or edit the ID. `N` is local to this store/history authority, so use the full token for a portable citation. The CLI resolves local numbers before storing pins or sending canonical requests; JSON and stored references retain full tokens. Use a revision when you need to cite exactly what a note said, compare an earlier decision, or recover content after an unwanted change. A correction under the same ID differs from the replacement relationship in the [main walkthrough](RUNBOOK.md#retiring-a-memory-using-superseded-by-links): the record continues, and its earlier revisions remain in history.

| Form | Purpose |
| --- | --- |
| `bd versions ID` | Inspect saved revisions and the current live or deleted state. |
| `bd show ID@current` | Select the current live revision explicitly. |
| `bd show ID@N` / `bd show ID@=TOKEN` | Read the specific retained revision identified by the citation. |
| `bd update ID --from-revision N` / `--from-revision =TOKEN` | Use retained editable content as the starting point for a new revision. |
| `--if-revision N` / `--if-revision =TOKEN` | Guard the write against a change to the current state. |

> **Qualification exercise:** The following is the intended release walkthrough. Historical selectors and recovery are included in the landed build; use a fresh workspace and the release executable described in the [runbook setup](RUNBOOK.md#using-the-playground). An unsupported-command refusal is an outstanding acceptance item, not a successful exercise. Replace each uppercase revision placeholder with the exact token from your own output. Keep the leading `=` shown in the commands: `=TOKEN` means an exact token, while a bare positive number means a local history ordinal. The examples quote `=TOKEN` arguments so Zsh does not try to expand them as command names. Bare opaque tokens and `current` are refused on these flags. `--if-source-revision` keeps its existing exact-token syntax.

### Catch up from an earlier revision

Create an isolated Memory for this exercise so that unrelated relationship edits do not change its revision:

```sh
bd create 'CSV recovery exercise' --id revision-demo --type memory \
  --body 'CSV exports must include headings.'
bd show revision-demo --long
bd versions revision-demo
```

Save the full live revision token as `REV_A` for the guard examples below. The human listing also gives this initial state local sequence number 1. You can read it directly with `bd show revision-demo@1`; `bd show revision-demo@=REV_A` selects the full token explicitly. Make a mistaken edit, guarded by that revision, then inspect the resulting record and save its new revision as `REV_B`:

```sh
bd update revision-demo --properties '{"body":"CSV exports omit headings."}' --if-revision '=REV_A'
bd show revision-demo --long
bd show revision-demo@=REV_A
```

The historical read should still say that exports include headings. The same human document format is used for current and historical content; its revision context distinguishes them. Now start from that saved content and apply a deliberate improvement in the same update:

```sh
bd update revision-demo --from-revision '=REV_A' --if-revision '=REV_B' \
  --properties '{"body":"CSV exports include headings, even for an empty selection."}'
bd show revision-demo --long
bd versions revision-demo
```

Save the new live revision as `REV_C`. This is “catch up from here”: restore the earlier editable properties and metadata, then apply the requested edits. The Type, ownership and protected state remain governed by the current record. The operation creates a new revision rather than moving a history pointer backward. `REV_A` selects the source content; `REV_B` guards the current state. They do different jobs.

### Undelete under the same ID

Delete only this disposable exercise record, then inspect its history:

```sh
bd delete revision-demo --force --if-revision '=REV_C'
bd versions revision-demo --long
```

The default timeline omits the full token lines. `--long` includes them, and JSON retains exact tokens. The human history column is **SEQ**, rather than a REV label that suggests the local number is the portable revision token. Numbers remain stable for this Resource even when additional revisions or a deletion/recovery are added; they are not row positions that get renumbered. The history timeline includes the deletion event even after a later recovery. Use `bd versions revision-demo --long` to include the exact token lines; save the terminal deletion token as `REV_D`. The numeric column is a stable local sequence, not a replacement canonical identity. A deletion revision is a terminal state, not a readable live-content revision. Restore from the saved live revision and guard against that deletion state:

```sh
bd update revision-demo --from-revision '=REV_C' --if-revision '=REV_D'
bd show revision-demo --long
bd versions revision-demo
```

The Memory should be live under `revision-demo` with a new revision. Its deletion and earlier revisions remain in history. Ordinary create does not implicitly undelete an ID. Recovery needs an available retained live source for that same subject and the required History/protected baseline; it cannot manufacture content that the store no longer retains.

### Follow and recover a Link under the same ID

Links have their own continuing identities and retained revisions. Use the Link command space for historical composition and recovery: `bd link update`, not an assumed generic `bd update LINK` alias. The landed build includes the Link lifecycle implementation; an older executable's unknown-flag refusal is not a completed recovery exercise.

For this isolated exercise, make an unowned References Link between two fresh memories:

```sh
bd create 'Pin source' --id pin-source --type memory --body 'A source note.'
bd create 'Pin target' --id pin-target --type memory --body 'Original target content.'
bd show pin-target --json
```

Save the target's full token as `TARGET_REV_A`. Use that exact token, not the display number, in the endpoint field:

```sh
bd link create pin-source pin-target --type references --id pin-example \
  --target-revision TARGET_REV_A --properties '{"title":"Original reference"}'
bd link show pin-example --json
bd link versions pin-example
```

Save the new Link's full token as `LINK_REV_A`. Make a change, capture its new token as `LINK_REV_B`, then compose from the earlier Link content:

```sh
bd link update pin-example --properties '{"title":"Revised reference"}' --if-revision '=LINK_REV_A'
bd link show pin-example --json
bd link update pin-example --from-revision '=LINK_REV_A' --if-revision '=LINK_REV_B'
bd link show pin-example --json
bd link versions pin-example
```

Save the restored live Link token as `LINK_REV_C`. The new revision carries the restored editable content; it does not rewind a history pointer or bypass the current Type/ownership/protected rules. Now delete and recover that same Link:

```sh
bd link delete pin-example --force --if-revision '=LINK_REV_C'
bd link versions pin-example --json
```

Save the current terminal deletion token as `LINK_REV_D`, then restore from the retained live source:

```sh
bd link update pin-example --from-revision '=LINK_REV_A' --if-revision '=LINK_REV_D'
bd link show pin-example --json
bd link versions pin-example
```

The Link should be live under the same `pin-example` identity, with a new full revision token. Its deletion event and earlier live revisions remain in the timeline; a deletion event is not a live content source. The unowned References example does not change either endpoint's revision. Owned Links must also satisfy their owning source's current rules and any required `--if-source-revision` guard; do not copy an old source guard from an earlier mutation.

### Change a pin without changing identity

A stored endpoint pin names the revision the Link refers to. `--if-revision` guards the Link write; `--if-source-revision` guards an owning source where required. Neither guard pins an endpoint.

After the preceding recovery, save the current Link token as `LINK_REV_E`. The field-specific values `current` and `unpinned` resolve or clear the pin while preserving the Link's identity:

```sh
bd link update pin-example --target-revision current --if-revision '=LINK_REV_E'
bd link show pin-example --json
```

Save the resulting Link token as `LINK_REV_F` (a true no-op can retain the whole envelope and token), then explicitly clear the target pin:

```sh
bd link update pin-example --target-revision unpinned --if-revision '=LINK_REV_F'
bd link show pin-example --json
```

Omitting a pin field on update preserves its stored selection; it does not request an unpin. On create, an omitted pin field is floating. `unpinned` here is a field value, not a new `ID@unpinned` convenience selector. Owned-source pin restrictions remain enforced. A retained single-record read or a live incidence query does not imply arbitrary historical graph traversal.

### List incident Links and inspect their pins

`bd link list` lists all current, live Links matching the filters. With one Bead ID, it lists incidence; with two, it lists Links from the first to the second. Direction is authored direction, independent of the relationship's English label.

```sh
bd link list
bd link list csv-design
bd link list csv-design csv-encoding
bd link list csv-design csv-encoding --direction both
bd link list --from csv-design --to csv-encoding
bd link list csv-design csv-encoding@current
bd link list csv-design --type references
bd link list csv-design --exact-type references
bd link count csv-design --type references
bd link show encoding-context
```

One-ID incidence defaults to `both`; two-ID selection defaults to `out`. `in` reverses the pair, while `both` includes either direction without duplicating self-Link rows. Other filters are combined with this selection. `--type TYPE` selects the family, including refinements; `--exact-type TYPE` selects only the declared Type. Link count uses this same checked population, not a separate broader scan. Text, JSON and canonical output must agree on the selected identities and count. These CLI filters map onto the selected BDP exact-Type and family semantics; they do not redefine the wire contract or silently change native search, ready, dep or query workflows.

| Selected endpoint | Applicable current live Links |
| --- | --- |
| Bare `ID` | All incident Links, regardless of that endpoint's pin. |
| `ID@current` | Floating endpoints and pins matching the current revision. |
| `ID@N` or `ID@=TOKEN` | Pins matching the selected revision; also floating endpoints if that revision is current. |
| With `--all-revisions` | All current live incident Links regardless of endpoint pin. |

The override does not reconstruct historical Link inventory. Selecting an earlier Bead revision filters the applicability of Links live now. Each endpoint is matched independently. A pin remains visible even when its target has since changed; inspecting the stored Link does not silently float it to today's target.

## Appendix C. Common command reference

### Read the common output

The common Bead and Link rows use the same identity and `[Type]` notation. Local identities omit only the workspace's exact Scope prefix; cross-scope references remain full URLs. A literal segment in an ID is never removed to make the display shorter.

`bd list` and `bd link list` put ID before Type. Creation prints the same row representation for the records it created. `bd show` and `bd link show` use a document view: identity, Type and optional title, selected revision, content and metadata. A Link also shows its exact source and target, including endpoint pins. The normal human revision label is the stable local revision number; JSON and administrative detail retain the full token. If an optional human-label lookup is unavailable or exceeds its budget, output falls back to the exact token without turning a successful read or write into a failure. Resolving an input `@N` still fails closed if its mapping cannot be established. Empty sections are omitted. Use `--long` for administrative detail and explicit JSON output when scripting; human output does not default to JSON. Current and retained reads share this layout, but a retained read must not borrow current comments, owned relationships or other current context and present them as historical content.

Subcommand help focuses on that command's options. Use `bd --help` for global options or `bd help --all` for the complete reference. Type filters use `--type TYPE` for a refinement family and `--exact-type TYPE` for only that declared Type. Graph traversal uses the corresponding `--link-type` / `--exact-link-type`; `--link-type any` includes all Link Types.

The examples mostly provide IDs so that later commands can refer to predictable records. When omitted, generic Bead/Link creation starts with a six-character lowercase URL-safe canonical generated identity. This is the complete allocated name, not a display abbreviation. An explicit `--id-root` places generated IDs under that root; do not combine it with `--id`. Native Issue creation retains its supported allocation conventions and complete-name collision checks. The allocator checks and reserves the complete name atomically across the Scope's lifetime namespace, covering both Resource kinds, prior/deleted identities, aliases and navigation collisions. A collision can cause a retry or bounded lengthening; do not depend on every generated name having exactly six characters. Existing identities are never shortened or rewritten. This local allocation check is not a disconnected-replica merge guarantee: conflicting histories still require checked reconciliation before admission.

### Familiar Issue workflows

Preview 3 also improves familiar Issue workflows. These commands use the same implementation Issue created earlier:

```sh
bd create 'Task: Verify CSV export' --id export-check --type task --force
bd link create export-check export-work --type blocks --id check-waits
bd show export-work --long
bd show export-work export-check --json
bd list --format json
bd ready --type task --json
```

Here `export-check` depends on `export-work`. The informational and Memory relationships do not block readiness: `export-work` should be ready, while `export-check` waits for its blocker. `ready --claim` preserves the same dependency-only rule.

### Optional generated identity exercise

For a generated Memory under an explicit organization root, use:

```sh
bd create 'CSV scratch note' --type memory --body 'An independently named note.' --id-root notes
```

Save the complete returned ID for later commands. Link creation has the same `--id-root` option; generated Links are flat by default, rather than automatically named beneath their source. `--id` plus `--id-root` is an error. A root organizes the name the allocator creates; it does not change Type membership or turn a path into ownership.

### Familiar output and input

Issue `show` supports multiple IDs, `--id`, `--current`, `--refs`, `--children` and interactive `--watch`. Structured output is form-specific in Preview 3:

| Command form | Structured result |
| --- | --- |
| `bd show ISSUE --json` | Familiar native Issue-detail array, with native fields and local Issue IDs. |
| `bd show ID --format graph-json` | Complete canonical Bead record with full ID, actual Type, revision and selected owned Links, in the preview `.result` envelope. |
| `bd show MEMORY --json` / `bd link show LINK --json` | Complete record in the preview `.result` envelope. |
| `bd list --type any --all --format records-json` | Mixed Bead inventory in `.result.items`; canonical IDs and full revision tokens. |

Use the explicit canonical forms when comparing records with BDP Read. The native compatibility array and complete-record envelope are not interchangeable. A successful human presentation change does not change these JSON contracts.

Issue creation and update accept `--body-file FILE`, `--description-file FILE`, `--stdin` and `--description -`. Conflicting sources refuse before consuming stdin. Update still requires its admitted revision guard. List accepts `--format json`, including case variants, and `--tree=false` for flat output.

For an epic hierarchy, `blocked --parent ID` resolves the native parent in the checked read. `show --children` and `show --refs` inspect related Issues. Positional `bd link BLOCKED BLOCKER` retains familiar Dependency feedback and JSON.

### Checked close and deletion

A multi-Issue close containing a missing ID refuses before writing or claiming next work. Other per-Issue policy refusals can still coexist with a successful sibling; this is not a promise that every batch refusal is all-or-nothing.

Issue deletion supports single or named sets, `--from-file FILE` and Dependency-only `--cascade`. Preview the selected set without `--force`, or use `--dry-run`; apply requires `--force` and the admitted guard policy for that form. Test deletion only on disposable data.

Apply removes all incident Links, retains deletion history, and revises surviving sources that own removed Links. Cascade follows Dependencies toward dependent Issues. Informational Links are removed without deleting their other endpoint. Invalid surviving Type or ownership state refuses the whole operation. Missing or non-Issue set members cannot cause earlier members to be deleted first.

The older recorded candidate refused dotted native Issue IDs in set deletion pending the citation-rewrite fix, and its single-Memory deletion refused incident Links. Those are source-specific limitations to recheck on the assembled identity/lifecycle successor, not promises that the correction is installed. Delete preview checks topology and guards but does not simulate final Type validation; apply may still refuse. General per-Type/direction/depth cascade filters are outside this preview.

## Appendix D. Output and traversal limits

| Author a relationship | Read the relationship |
| --- | --- |
| `memory reference` | `memory references`, `memory referenced-by` |
| `memory supersede` | `memory successor`, `memory supersedes` |
| `memory requires` | `memory requirements`, `memory required-by` |
| `memory defines` | `memory definition`, `memory defined-by` |
| `memory subject-to` | `memory constraints`, `memory subjects` |

The five authoring commands express different meanings. The ten reads follow the corresponding relationships in the direction appropriate to the question. Start with `bd memory --help`, then use a subcommand's `--help` for its options. The [main walkthrough](RUNBOOK.md#using-memory-beads-and-links-informative) introduces each pair in context.

The ten Memory reads have deliberate traversal defaults:

| Read commands | Default reach |
| --- | --- |
| `references`, `referenced-by` | One hop. |
| `successor` | Replacement chain, up to 1,000 hops. |
| `supersedes` | One reverse hop. |
| `requirements`, `definition` | Up to 1,000 forward hops. |
| `required-by`, `defined-by` | One reverse hop. |
| `constraints` | Original Subject To declarations, then Requires/Defines closure. |
| `subjects` | Direct incoming Subject To declarations. |

The first eight reads accept `--depth`; `constraints` and `subjects` have fixed rules and refuse it. Reads accept multiple roots, each with its own traversal view. Any summary counts unique identities across those views once; repeated tree occurrences are not extra Resources. Default request limits are 100 unique Beads and 200 unique Links; `--max-nodes` and `--max-links` can raise either to at most 1,000.

Without an explicit Link-Type selection, `bd graph` selects Dependency and its refinements. When the checked reached-incidence evidence contains omitted non-Dependency Links, a note on stderr explains the filter and suggests `--link-type any`. That notice is not evidence about unexplored parts of the Scope. `--exact-link-type TYPE` selects only that declared Type; `--link-type TYPE` includes refinements. Repeat selections to form a union, with one direction and depth applying to that union. The wildcard is a CLI selector, not a new Type identity or a reservation of the name `any`. Existing exact Type URI selection and ordinary configured-name compatibility must be preserved; an actual ambiguous new generic selection must refuse explicitly rather than silently select a different Type. Only selected Link Types determine reachability; eligible Links among reached Beads remain visible. Human output uses `[Type]` for Beads and Links and a plain empty result when nothing is selected. It omits default complete/frontier bookkeeping. When the depth actually stops traversal, a plain boundary notice discloses that limit. Canonical structured traversal context retains the corresponding completeness/frontier evidence. Exceeding resource or output bounds refuses rather than returning an apparently complete graph.

Memory graph JSON uses `schema_version: 2`, `result_kind: "graph"`, `outcome: "success"` and `.data`. Generic `graph`, `ready` and several other routes use `schemaVersion: 1` / `.result`. Their shapes are not interchangeable; inspect the selected command's output before scripting. Full BDP records retain canonical URL identities and references; the common no-prefix identity correction must be applied consistently, not by rewriting an isolated JSON field.

Resource field masks are unavailable. Current-snapshot acquisition has its own budget, so a small selected result does not guarantee that an arbitrarily large store can be captured. Memory relationship reads refuse selected endpoint pins that do not match the captured current revisions. Generic `bd graph` instead displays current Bead records alongside the Links' exact stored pins; it does not reconstruct historical endpoint content or rewrite those pins. Arbitrary historical graph traversal is outside this walkthrough.
