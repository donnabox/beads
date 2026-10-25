# Extending Beads: bringing work and knowledge into Preview 3

> **DONNA: WRITE THE OPENING HERE.**

<!-- Editorial draft. Examples target Janet's October 24 playground source 0fcca128. Donna owns release wording and publication. -->

## Start with work you already have

A new graph workspace does not have to start with invented tasks. You can bring in the current Issues, local dependencies, comments and key/value memories from an existing ordinary Beads workspace, then explore that information through the new model. That gives you work and knowledge you recognize before you start adding richer relationships. Preview 3 supports a bounded conversion into a fresh destination; it does not transfer your database's complete history or replace your backup.

### Export from the existing workspace

Use the ordinary `bd` installation that already owns your source workspace. Export its current records to a file you can read from the new workspace, and keep the original workspace available for comparison. Memories are excluded from ordinary export by default, so include them explicitly if you want to bring them forward. The export command's normal filtering still applies.

```sh
# Check the ordinary Beads installation in the existing workspace.
bd version

# Export current Issues, dependencies, comments and key/value memories.
bd export --include-memories --output /tmp/preview3-export.jsonl

```

An ordinary export carries current data. Earlier revisions, Dolt branches and the source database's history do not come with it. Keep this file unchanged while you test the conversion so that you have a clear source for comparison.

### Prepare a fresh destination and try a dry-run

Switch to the selected Preview 3 executable before preparing the destination. Choose a new directory outside any existing Beads workspace, and initialize Git there so workspace discovery has a local boundary. The Scope URL supplies the workspace's identity; initializing it does not contact or start a web service. Import must happen before you create sample records, because this preview requires a genuinely fresh destination.

```sh
# Confirm that this terminal now uses the selected Preview 3 build.
bd version

# Create a new practice directory outside the existing workspace.
mkdir beads-preview3-trial

# Enter the new directory.
cd beads-preview3-trial

# Give workspace discovery a local Git boundary.
git init -q

# Initialize a fresh embedded graph workspace with its own Scope identity.
bd init --graph-mode link --scope-url 'https://example.invalid/preview3-trial/'

# Save the complete mixed Bead inventory before import.
bd list --type any --all --format records-json > before-import.json

# Validate the export through the import transaction, then roll it back.
bd import /tmp/preview3-export.jsonl --dry-run --json > import-preview.json

# Inspect the proposed counts and history boundary.
jq '.result | {dry_run, issues, memories, dependencies, comments, history}' import-preview.json

# Read the inventory again after the dry-run.
bd list --type any --all --format records-json > after-preview.json

# Check that the dry-run left the visible inventory unchanged.
cmp before-import.json after-preview.json

```

A successful dry-run reports `dry_run: true`; `cmp` finishes without output. The result also includes identity maps for the records the importer accepts. Dry-run reserves nothing, so generated Link IDs are provisional and may differ when you apply the import.

### Apply once, then inspect familiar content

After a successful dry-run, apply the same file and options to the same still-fresh destination. Save the result maps so that you can inspect recognizable Issues and Memories. Use the mixed inventory to check both kinds of Bead: `--all` alone changes Issue status coverage, while `--type any` selects the broader population. Compare content and relationships with the export before relying on the converted copy.

```sh
# Apply the same export once to the fresh destination.
bd import /tmp/preview3-export.jsonl --json > import-result.json

# Inspect imported record counts and their destination identities.
jq '.result | {issues, memories, dependencies, comments, ids, memory_ids, dependency_ids}' import-result.json

# List Issues and Memories together, including closed Issues.
bd list --type any --all

# Browse the imported Memories.
bd memories

```

An admitted Issue keeps its public ID, and its local dependencies become Links. Existing comments remain in the Issue's comment feed; they do not become Memory Beads. For a classic key/value Memory, the key becomes its public ID and the exact value becomes its body, while its title is derived independently from that body.

The importer refuses unsupported or conflicting data as a whole. This includes missing or cross-Scope dependency endpoints, incompatible IDs, unsupported custom Issue classifications or statuses, and source history payloads. Repeating an import into an occupied workspace refuses; deleting records does not make a used destination fresh again. This is a way to seed a new graph workspace, not a lossless in-place migration or portable graph round trip.

## Keep useful knowledge beside the work

A team learns as it works, and Memory Beads keep that knowledge available for later use. A Memory has a title and a Markdown body; its optional `status` is an author-chosen label with no automatic approval or workflow effect. References Links let readers follow useful context without turning every connection into a blocking Dependency. The [Memory Type overview](https://github.com/gastownhall/bdp/blob/21fa91d3e189bff3832bd977c3d6be49f2b5d93c/specs/types/memory/memory.md#overview-informative) and [Memory CLI introduction](https://github.com/gastownhall/bdp/blob/1979925130a5744add8e45910c8a64931570bc13/docs/cli-design/memory-cli.md) describe that model; the commands below use the selected Preview 3 build's admitted spellings.

### Record notes and the work they explain

Consider a team adding CSV export to a product. Start with a design note, an encoding note and an interoperability note, each with its own identity. `remember` takes the body as its argument; `--key` selects the same public identity as `--id`, and `--title` supplies readable display text. An identity keeps selecting the same record after its wording changes.

```sh
# Record the CSV design under an explicit identity and title.
bd remember 'Export the selected rows as CSV.' \
  --key csv-design --title 'CSV export design' --create-only

# Record the encoding note that the design will reference.
bd remember 'Use UTF-8.' \
  --key csv-encoding --title 'CSV encoding' --create-only

# Record the spreadsheet checks that tie design decisions to implementation.
bd remember 'Test the exported UTF-8 file in spreadsheet tools; keep the results with the implementation work.' \
  --key csv-interop --title 'CSV interoperability checks' --create-only

```

The general Bead command can create the same kind of note by selecting its installed Type. In this build, `types/memory` is an explicit Type selector; use `bd types` to inspect the vocabulary installed in your workspace. The following command is an alternative way to create the first note, so choose it instead of the first `remember` call above.

```sh
# Alternatively, create the design note through the general Bead command.
bd create --id csv-design --type types/memory \
  --properties '{"title":"CSV export design","body":"Export the selected rows as CSV."}'

```

The notes hold knowledge; Issues record the work of implementing and reviewing the feature. The fixed practice IDs below use `--force` to allow a spelling outside the workspace's configured native Issue prefix. We can then inspect a Memory as a Bead, recall just its body, or browse summaries to choose what to read.

```sh
# Create the implementation Issue.
bd create 'Add CSV export' --id export-work --type task --force \
  --description 'Implement CSV export for the selected rows.'

# Create the review Issue.
bd create 'Review CSV export' --id review-work --type task --force \
  --description 'Review the completed CSV export feature.'

# Read the design note's body.
bd recall csv-design

# Browse Memories whose content mentions CSV.
bd memories CSV

# Read the same design note as a complete Bead.
bd show csv-design

```

### Connect context, including cycles

References explain useful connections between knowledge and work. The design uses the encoding note, the encoding note points to interoperability checks, and those checks refer back to the design. The checks also point to the implementation Issue, which refers to the design it implements. This gives us one cycle entirely among Memories and another that passes through an Issue; neither is a blocking Dependency.

```sh
# Explain which encoding note the design uses.
bd memory reference csv-design csv-encoding --id encoding-context \
  --title 'Uses this encoding' --body 'Text encoding for exported files'

# Connect the encoding choice to its spreadsheet checks.
bd memory reference csv-encoding csv-interop --id encoding-checks \
  --title 'Interoperability checks for this encoding'

# Close the cycle among the three Memories.
bd memory reference csv-interop csv-design --id checks-design \
  --title 'Design these checks validate'

# Connect the interoperability checks to the work recording their results.
bd memory reference csv-interop export-work --id checks-work \
  --title 'Implementation and verification work'

# Close the cycle that passes through the implementation Issue.
bd memory reference export-work csv-design --id design-context \
  --title 'Implementation context'

# Make review wait for the implementation Issue to finish.
bd link create review-work export-work --type blocks --id review-after-export

```

References cycles are permitted. A traversal stops expanding a Bead it has already visited and keeps the Links that show the relationship, including a cycle-closing Link. The `blocks` Dependency retains its separate workflow meaning: review waits for implementation.

```mermaid
flowchart LR
    V["Issue: review-work"] -->|blocks| I["Issue: export-work"]
    I -->|references| D["Memory: csv-design"]
    D -->|references| E["Memory: csv-encoding"]
    E -->|references| N["Memory: csv-interop"]
    N -->|references| D
    N -->|references| I
```

A plain References Link has its own revision; adding or editing it does not revise either endpoint. Use a forward read to find context, or a reverse read to find records referring to a note. The graph viewer can follow the current References family across the connected example.

```sh
# Find the design note's direct References.
bd memory references csv-design

# Find Beads that refer to the design note.
bd memory referenced-by csv-design

# Explore the References graph far enough to encounter both cycles.
bd graph csv-design --link-type types/references --direction out --depth 4

```

### Revise a note or replace it

An edit maintains the same piece of knowledge under the same identity. Each accepted change saves a new revision, while an unchanged update leaves the revision alone. A separately maintained replacement is a different Memory, connected to the old one by a Superseded By Link. These operations answer different questions: what changed in this note, and which note should I follow next?

```sh
# Clarify the encoding note while keeping its identity.
bd update csv-encoding --properties '{"body":"Use UTF-8 without a byte-order mark."}'

# Rename the note and label it reviewed in one edit.
bd update csv-encoding --properties '{"title":"CSV text encoding","status":"reviewed"}'

# List the saved revisions of this identity.
bd versions csv-encoding

# Read the second saved revision created by this example.
bd show csv-encoding@2

# Recall just the body from that saved revision.
bd recall csv-encoding@2

```

Use the actual revision numbers and opaque tokens reported by your store. `@2` selects a store-local revision number; `@=TOKEN` selects the exact token for a portable citation. A retained Resource revision does not reconstruct an entire past graph, and this preview's focused Memory relationship reads can refuse endpoint pins that no longer match the captured current state.

```sh
# Record a separately maintained guide that replaces the short note.
bd remember 'Encoding, headings and spreadsheet examples.' \
  --key csv-guide --title 'CSV interoperability guide' --create-only

# Connect the retiring note to its replacement.
bd memory supersede csv-encoding csv-guide --id encoding-replacement

# Follow the replacement path from the old note.
bd memory successor csv-encoding

# Find the direct predecessors of the replacement guide.
bd memory supersedes csv-guide

```

The Link points from old to new, revises the retiring Memory and leaves the replacement unchanged. The old identity and earlier revisions remain readable, and existing References keep their targets until their authors change them. Each retiring Memory has at most one successor, and replacement chains cannot cycle.

## Give constraints a structure

Some knowledge states what work must do, avoid or follow. A Constraint Bead records such an obligation in its body, and its Type tells a reader what kind of rule it is. Policy, Decision, Procedure, Requirement, Scenario and Specification refine Constraint, which itself refines Memory. Use Constraint directly when none of those more specific Types fits; a metadata label does not create Type membership.

| Type | What it records |
| --- | --- |
| Policy | Governing rules for a system, project or team. |
| Decision | Choices that constrain later work. |
| Procedure | How work is carried out, with the flow in its body. |
| Requirement | An outcome or condition that must be met. |
| Scenario | Required behavior in a particular situation. |
| Specification | Intended results at any useful level of detail. |

### Define the parts and require other rules

For CSV export, a Specification can define a selected-row Requirement and an empty-selection Scenario. A customer-data Policy can remain independently maintained and govern many features. Its detailed Policies belong to its own definition, even when another Requirement also relies on one of them. Each record has its own identity and can be revised or referenced independently.

```sh
# Describe the feature's intended result.
bd create 'CSV export format' --id csv-format --type types/specification \
  --body 'Export selected rows and follow the rules linked below.'

# State which rows must appear in the output.
bd create 'Export every selected row' --id selected-rows --type types/requirement \
  --body 'Every selected row appears exactly once.'

# State the expected outcome for an empty selection.
bd create 'Export an empty selection' --id empty-export --type types/scenario \
  --body 'Produce headings and no data rows.'

# Record the independently maintained customer-data Policy.
bd create 'Protect customer data' --id customer-data --type types/policy \
  --body 'Keep personal information out of logs and shared sample exports.'

# Record the Policy's log-handling detail.
bd create 'Redact application logs' --id redact-logs --type types/policy \
  --body 'Remove personal information before writing log entries.'

# Record the Policy's sample-data detail.
bd create 'Use anonymized sample exports' --id sample-exports --type types/policy \
  --body 'Use anonymized records in shared sample exports.'

```

Defines makes a child part of its parent's definition. Requires names another Constraint that must also be met, including one maintained in a separate defining tree. Defines refines Requires, so a Requires-family read includes both kinds of Link without changing their labels.

```sh
# Make the selected-row Requirement a defining part of the Specification.
bd memory defines csv-format selected-rows --id defines-rows

# Make the empty-selection Scenario another defining part.
bd memory defines csv-format empty-export --id defines-empty

# Give the customer-data Policy its log-handling detail.
bd memory defines customer-data redact-logs --id defines-logs

# Give the customer-data Policy its sample-data detail.
bd memory defines customer-data sample-exports --id defines-samples

# Require the independently maintained Policy from the Specification.
bd memory requires csv-format customer-data --id requires-policy

# Require the same log-handling Policy from the selected-row Requirement.
bd memory requires selected-rows redact-logs --id requires-log-policy

```

The two defining trees remain distinct. The selected-row Requirement can require the log-handling Policy without becoming its defining parent. A Defines child has at most one defining parent, and defining trees cannot cycle; Requires can cross between trees and may cycle, with repeated expansion stopped during a read.

```mermaid
flowchart TD
    S["Specification: csv-format"] -->|defines| R["Requirement: selected-rows"]
    S -->|defines| C["Scenario: empty-export"]
    S -->|requires| P["Policy: customer-data"]
    P -->|defines| L["Policy: redact-logs"]
    P -->|defines| E["Policy: sample-exports"]
    R -->|requires| L
```

Both kinds of Link are owned by their source Constraint. A Link change revises that source without revising the target. The following reads distinguish the Specification's own defining parts from the broader set of constraints it requires.

```sh
# Read the Specification's defining Requirement and Scenario.
bd memory definition csv-format

# Discover its defining parts and the independently maintained Policies it requires.
bd memory requirements csv-format

# Find the selected-row Requirement's direct defining parent.
bd memory defined-by selected-rows

# Find Constraints that directly require the log-handling Policy.
bd memory required-by redact-logs

```

### Declare which work is governed

A Subject To Link explicitly declares that a Constraint governs a Bead. The implementation Issue and a Memory describing a preview build can both be subject to the CSV Specification. Starting with either subject, a focused read follows the declared Constraints and their Requires family, including Defines. The paths tell a reader how each rule was reached.

```sh
# Record the preview build as a second kind of subject.
bd remember 'The CSV preview build is ready for review.' \
  --key csv-preview --title 'CSV preview build' --create-only

# Declare the Specification governing the implementation Issue.
bd memory subject-to export-work csv-format --id work-obligation

# Declare that the same Specification governs the build Memory.
bd memory subject-to csv-preview csv-format --id build-obligation

# Discover the rules explicitly governing both subjects.
bd memory constraints export-work csv-preview

# Find the subjects that directly declare the CSV Specification.
bd memory subjects csv-format

```

Subject To is owned by its source, so a Link change revises the subject while leaving the Constraint unchanged. It does not spread through another Subject To declaration: a document-review Policy governing the Specification does not automatically govern its implementation Issue. These declarations help people and agents discover governing rules; they do not execute prose, prove compliance or change Issue readiness.

## Read a proposal as structured knowledge

A substantial proposal contains several kinds of knowledge: an explanation of its purpose, requirements that constrain the design, and scenarios that describe its expected use. [Chris's Memory proposal](https://github.com/gastownhall/beads/issues/5877) gives us a concrete example. We can represent it as one Specification, with 32 Requirement Beads and 16 Scenario Beads as its defining children. Each child retains its original text and complete title, including its `Rn` or `Sn` prefix; R27a accounts for the extra Requirement beyond R1–R31.

### Start with the Specification

For this example, use a separate, focused Scope containing the proposal and its extracted Beads. The [example data](preview3-blog-assets/memory-proposal-data.json) retains the complete source titles and bodies; it is editorial example data, not a format accepted by `bd import`. The Specification's introductory body is newly authored prose, while the children carry the actual proposal text. The compact displays below illustrate the arrangement and abbreviate long titles for the page; they are not an exact terminal transcript or a claim that the entire proposal has been implemented.

```sh
# Read the Specification's title and introductory body.
bd show memory-proposal

```

```text
ID:       memory-proposal
Type:     specification
Title:    Memory Beads
Revision: [current revision]

Keep durable project knowledge in Beads so people and agents can
find it, read it selectively, and understand its relationship to
their work.

Give each Memory a persistent identity and its own content.
Preserve its evolution through recorded revisions, and connect it
to tasks and other knowledge through explicit Links.

The defining Requirements and Scenarios describe the proposed
behavior. They can be read individually or explored together
through this Specification.
```

The introductory body gives readers a starting point without making them read every detail. The Requirements and Scenarios remain independently addressable, so a discussion or implementation task can refer to a particular part of the proposal.

```sh
# List Requirement Beads in this focused Scope.
bd list --type requirement

```

```text
ID           TYPE          TITLE
memory-r1    requirement   R1. Memory is a…
memory-r2    requirement   R2. Canonical identity does…
…
memory-r30   requirement   R30. Core preserves evidence…
memory-r31   requirement   R31. Every target addition…
```

These excerpts show the first two and last two entries in the proposal's order. R27a and the other intervening Requirements are present in the stored data. This presentation order does not promise automatic numeric sorting by the CLI.

```sh
# List Scenario Beads in this focused Scope.
bd list --type scenario

```

```text
ID           TYPE       TITLE
memory-s1    scenario   S1. Record durable knowledge…
memory-s2    scenario   S2. Search before writing…
…
memory-s15   scenario   S15. Temporary knowledge carries…
memory-s16   scenario   S16. An application…
```

The ellipses shorten this presentation; the stored titles and bodies remain complete. In a shared Scope, these commands would also list other matching Beads: the Type filter does not select membership in this particular Specification.

### Follow the definition and its requirements

The lists show what exists, but they do not explain how those Beads form a document. Defines Links make each extracted Requirement and Scenario part of the Specification's definition. Both kinds of child are directly defined by the root Specification, as siblings. A Scenario may exercise several Requirements, so its place in the document does not force it to belong to just one of them.

```sh
# Explore the Specification's defining Requirements and Scenarios.
bd memory definition memory-proposal

```

```text
memory-proposal [specification]
├─ defines → memory-r1  [requirement]
├─ defines → memory-r2  [requirement]
│  …
├─ defines → memory-r30 [requirement]
├─ defines → memory-r31 [requirement]
├─ defines → memory-s1  [scenario]
├─ defines → memory-s2  [scenario]
│  …
├─ defines → memory-s15 [scenario]
└─ defines → memory-s16 [scenario]
```

Defines is a refinement of Requires: a defining child is something that must be met for its parent to be met. The broader requirements read therefore includes the same defining children, along with any additional Requires Links. It preserves `defines` labels rather than relabeling those edges as `requires`.

```sh
# Discover the Specification's defining parts and any other required Constraints.
bd memory requirements memory-proposal

```

With only the Defines Links shown above, this read reaches the same children. If the Specification later requires an independently maintained Policy, that Policy appears in this broader view without becoming part of the Specification's defining tree. The proposal's S-number citations do not automatically create dependency Links among the extracted children.

```sh
# Read the first Requirement's title and complete content.
bd show memory-r1

# Read the first Scenario's Markdown body.
bd recall memory-s1

```

Each extracted Bead can be discussed, referenced and revised independently. The Specification preserves its relationship to the proposal, while the body carries the actual meaning. Discovering a Requirement establishes what must be met, not that anyone has met it.

## Where are we?

> **DONNA: WRITE THE CLOSING AND CALL TO ACTION HERE.**
