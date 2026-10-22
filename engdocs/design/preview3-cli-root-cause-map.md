# Preview 3 CLI parity: root-cause map (discussion draft)

This map groups the exceptions in [the 43-command parity inventory](preview3-cli-parity.md) by work that could remove several exceptions at once. It is a design hypothesis, not a new release gate or a claim that open PRs are merged. The as-built comparison remains shipping v1.3.1 versus Preview 3 integration at `b6d0493ba`. Donna has selected basic Bead/Link CRUD plus the ready invariant as the first focus; she has not declared other entries release-blocking.

**Order of work:** first prove the underlying operation can produce correct,
atomic native and canonical graph state on embedded and shared-server Dolt;
then make shipping command forms, flags, defaults and output match where the
model admits them. The legacy CLI adapter is lowest priority unless it reduces
implementation cost. This order does not defer a selector that could mutate
the wrong object, an ignored flag, or a misleading success: those are
correctness requirements, not presentation polish. Capability proof may use
exact canonical paths and typed results before every shipping alias is wired.

## The atomicity distinction

The transaction primitive exists. Ordinary `bd batch` executes its narrow script in one transaction and rolls back on a command error ([`cmd/bd/batch.go`](../../cmd/bd/batch.go)). The native `BatchApplier` is also all-or-nothing for its four supported item kinds ([`issueops/batchapplier.go`](../../issueops/batchapplier.go)); graphstore's `withTx` wraps graph writes, retained bytes and coordination in one transaction ([`internal/storage/graphstore/backend.go`](../../internal/storage/graphstore/backend.go)). Graph-mode link currently does not admit the `bd batch` command through its route gate ([`cmd/bd/graph_preview.go`](../../cmd/bd/graph_preview.go)). The native batch role does not itself publish graph catalog identities, canonical Links or their retained History.

Atomic *commit* is distinct from all-or-nothing *item policy*. `BatchCloser` intentionally skips per-item refusals and commits successful siblings together ([`issueops/batchcloser.go`](../../issueops/batchcloser.go)). Graph `close --claim-next` uses one graph transaction with that policy ([`internal/storage/graphstore/issue_close_batch.go`](../../internal/storage/graphstore/issue_close_batch.go)); ordinary multi-ID graph `close` loops over separate single-Issue transactions ([`cmd/bd/graph_preview_workflow.go`](../../cmd/bd/graph_preview_workflow.go)). Shipping v1.3.1 direct `bd close A MISSING` pre-resolves both arguments and changes nothing, while its ordinary proxied path has per-item outcomes. Matching the tested direct behavior for a missing selector is narrower than changing all close refusals to rollback.

This is a selected policy as well as a transaction-composition gap. Draft
[#64](https://github.com/donnabox/beads/pull/64) deliberately introduced
Preview 2 multi-Issue close with partial success: `bd close A MISSING` leaves
`A` closed, reports the missing ID, and exits nonzero. Its tests and design
must be preserved or explicitly superseded with attribution. Donna selected
the direct shipping behavior for Preview 3: if any requested Issue ID is
missing, `bd close A MISSING` changes no Issue. Resolve all named selectors
inside one write transaction before closing any item. This decision covers
missing IDs, not every per-item policy refusal; `BatchCloser` may still report
other outcomes individually. It does not require a new transaction engine.
Draft [#226](https://github.com/donnabox/beads/pull/226)–[#227](https://github.com/donnabox/beads/pull/227)
prove and route this narrower rule on embedded and real shared-server Dolt;
they remain unmerged candidates and do not change the pinned integration.

**Canary interpretation:** ordinary `bd batch` is a control for native
transaction plumbing and rollback, not a graph-mode release canary. Its narrow
grammar admits create, update, close, dependency add and dependency remove;
it does not admit delete or the full top-level command flag sets. A stronger
graph canary must combine an Issue and a canonical Link, inject failure after
the native edge but before retained Link/History publication, and show that
both embedded and shared-server Dolt leave no partial state. A linked delete
canary should likewise prove either complete intended cleanup or a refusal
that changes nothing; the cleanup policy still needs a command contract.

Shipping `bd create --graph` is a separate batch contract from Markdown
`bd create --file`. Its JSON plan gives every node a symbolic key; inline
dependencies resolve that key before treating it as an existing Issue ID,
and plan edges can refer to keys on either end
([`cmd/bd/graph_apply.go`](../../cmd/bd/graph_apply.go)). The shipping plan
schema has no arbitrary update or close items. It does have an
`assign_after_create` node option that updates the newly created Issue's
assignee later in the same transaction. That bounded form is absent from the
pinned integration and addressed by unmerged draft #240; generic update/close
policy belongs to the separate `bd batch` contract.
The graph-mode Markdown batch in #201–#202 creates Issues and local Links
atomically, but its
headings have no caller-assigned key and its Dependency targets must already
exist ([`internal/storage/graphstore/issue_create_batch.go`](../../internal/storage/graphstore/issue_create_batch.go)). Thus that batch proof does not close the
same-plan graph-JSON gap. The next storage proof should allocate and validate
all node identities and edges in one checked transaction, then publish native
Issues, canonical Links, retained versions and History together; it must
refuse a plan it cannot represent without committing a subset. Harry Miller's
open [upstream #7521](https://github.com/gastownhall/beads/pull/7521)
supplies a status/creation timestamp invariant for shipping graph plans and
should be preserved when that path is implemented.

The common create writer is close to the needed primitive, but its public
request cannot yet express the plan. `ExecuteCreateBatch` prepares every
item and assigns all native IDs before its single `CreateIssuesInTxWithResult`
call ([`internal/storage/issueops/create_batch.go`](../../internal/storage/issueops/create_batch.go));
the public `BatchCreateItem` contract permits a dependency only on an
earlier item with a caller-supplied ID
([`issueops/batchcreator.go`](../../issueops/batchcreator.go)). Graph-mode
Markdown creation resolves every dependency against an already stored Issue
*before* invoking that writer. The reusable seam is therefore a checked
plan-to-native-ID mapping inside the same write transaction, followed by
validation and publication of the complete edge set. A loop of single-Issue
creates or a CLI lookup after the first write would lose all-or-nothing
behavior. This is a storage capability proof first; the shipping JSON field
surface and output envelope can follow after the plan and rollback invariants
pass on both engines.

The native heterogeneous batch already exposes a reusable in-transaction
body: `PlanApplyBatch` checks symbolic keys and request shape before opening a
transaction, and `ApplyBatchInTx` binds the keys, creates the Issues and adds
their edges under the caller's `*sql.Tx`
([`internal/storage/batch_apply.go`](../../internal/storage/batch_apply.go),
[`internal/storage/issueops/batch_apply.go`](../../internal/storage/issueops/batch_apply.go)).
The graph adapter can lower the shipping JSON plan to that checked request,
run the body inside `graphstore.withTx`, then publish Bead identities, every
native edge's canonical Link, retained versions and source-owned History
before commit. It must validate that every plan field and endpoint is
representable by the installed graph Types; an existing native edge must not
gain a second canonical Link. A useful first test is two generated-ID nodes
with a forward symbolic edge, followed by an injected failure after the native
edge and before Link retention. Embedded and shared-server Dolt must both
show no nodes, edge, catalogue entries or versions after that failure.

Draft [#231](https://github.com/donnabox/beads/pull/231) at `b90818e85`
implements that first storage slice for new durable Issues and metadata-free
key-to-key edges. The injected post-native-edge failure and successful
canonical projection passed on embedded and real shared-server Dolt. It
explicitly refuses existing-Issue endpoints, metadata references, edge
metadata, updates and closes, and by itself does not wire the CLI. The broader
embedded graphstore sweep passed 177 test/subtest results with 21 server-only
skips. This is a checked capability proof, not full graph-JSON parity.

Stacked draft [#232](https://github.com/donnabox/beads/pull/232) at
`767e13b7c` wires that exact bounded storage case to `bd create --graph`
for new keyed Issues and local key-to-key dependencies. Its installed workflow
passed embedded and real shared-server Dolt. Unsupported plan fields and
existing-ID endpoints refuse before writing. These drafts are not integrated or
selected for release.

Stacked draft [#233](https://github.com/donnabox/beads/pull/233) at
`bf61fe93a` extends the same transaction to a new keyed source and an exact
existing local Issue target. The target resolves before native writes; alias
duplicates and missing targets refuse without partial state. Its injected
post-native-edge rollback, caller-request immutability check and installed CLI
workflows passed embedded and real shared-server Dolt. Existing sources,
partial or external targets, metadata and deferred assignment remain
unavailable.

Stacked draft [#234](https://github.com/donnabox/beads/pull/234) at
`6f68b2d8a` admits initial Issue text, references, assignment, owner, estimate,
due date and labels through the checked native writer. Its installed
create/show/Link workflow passed embedded and real shared-server Dolt without
skips. Other graph-plan forms remain explicit refusals; its four current-head
hosted checks passed, with review pending. Stacked draft
[#235](https://github.com/donnabox/beads/pull/235) at `02a480db4` then admits
one exact existing local Issue source linked to one new keyed target. Source
resolution, native batch, canonical Link, old/new source versions and History
share one transaction. Focused storage rollback/retention and installed CLI
workflows passed embedded and real shared-server Dolt without skips. Multiple
Links from one existing source refuse until intermediate-version retention is
proven. #235 has four passing current-head hosted checks; review is pending.

The next same-source batch proof has a specific retention seam. Native
`ApplyBatchInTx` applies dependency items in request order, and each genuinely
inserted edge mints an `issue_versions` row for its source. Its `ItemResult`
reports whether the edge changed but carries no post-edge Issue revision or
owned-Link snapshot. Graph `recordIssueMappingInTx` reads only the final native
revision and final canonical owned set. Merely removing #235's one-Link guard
would therefore leave intermediate native revisions without matching graph
Issue History. A safe extension pairs each changed edge with its native
revision and the owned-Link set at that point inside the same transaction.
Stacked draft [#237](https://github.com/donnabox/beads/pull/237) at
`5173b3945` does this for the admitted existing-source-to-new-target form:
it checks the native revision sequence, maps each canonical snapshot, checks
the aggregate Link cap, and rolls back if retaining the second version fails.
Focused historical-content, rollback and installed CLI create/Link/versions
tests passed embedded and real shared-server Dolt without skips on the earlier
head; the current head adds a two-distinct-existing-sources rollback and
History test, which passed both engines without skips. Related embedded
graphstore tests and its four current-head hosted checks passed. Stacked
draft [#238](https://github.com/donnabox/beads/pull/238)
at `0f4f12cec` admits new exact existing-to-existing local Issue edges in a
graph plan that also creates a node. Same-transaction endpoint selection,
native edge insertion, canonical Link mapping, source History and fault
rollback passed embedded and real shared-server Dolt, including the installed
CLI path. Its four current-head hosted checks passed; review is pending.
Stacked draft [#239](https://github.com/donnabox/beads/pull/239) at
`bdeed74fb` admits the proven same-type, metadata-empty, thread-empty
change-free re-add when the native pair already has a matching canonical
Link. It verifies the native item and source revision did not change; a mixed
plan retains only the genuinely inserted edge revision. Focused storage and
installed CLI workflows passed embedded and real shared-server Dolt without
skips. Native metadata-changing re-adds can mint a version despite reporting
no new edge, so they and unmapped native edges remain explicit refusals.
#239's four current-head hosted checks passed; review is pending. Stacked
draft [#240](https://github.com/donnabox/beads/pull/240) at `2c4b7b093`
admits only shipping's deferred `assign_after_create` update on a newly
created keyed Issue. The native create and update retain one final canonical
inception snapshot in the same graph transaction; an injected post-batch
fault rolls back both. Focused storage and installed CLI workflows passed
embedded and real shared-server Dolt without skips. Hosted CI and review are
pending. Stacked draft [#241](https://github.com/donnabox/beads/pull/241)
at `382c37318` admits initial node metadata and symbolic `metadata_refs`
to new keyed Issues. The native post-create metadata splice, deferred
assignment, canonical Link and one final Issue inception snapshot pass
focused embedded and real shared-server Dolt workflows with fault rollback.
Its broader embedded graph-batch selection passed with seven server-only
skips; hosted CI and review are pending. None of #237–#241 is merged;
arbitrary update and close items are not part of shipping `create --graph`'s
schema.

Test-only draft [#236](https://github.com/donnabox/beads/pull/236) at
`5fbf1cb91` pins the different *new-source* inception rule: native create plus
one edge produces two native `issue_versions` rows, while the graph catalog is
allocated only after the whole plan lands and retains one canonical creation
snapshot. This passed embedded and real shared-server Dolt without skips.
That intentional compression does not justify dropping intermediate graph
History for an existing, already allocated source. #236 hosted CI and review
are pending. It is the base of #237, which preserves this rule while mapping
each later native revision of an already allocated source.

A first commit in #231 at `cb7da57c5` extends the graph batch adapter to keep an
edge from a later item to an earlier item with an explicit native ID inside
the existing transaction. A staged Link-retention failure rolls back the
native Issues, catalog, retained state and coordination. Focused storage
tests pass on embedded and disposable real shared-server Dolt. The Markdown
parser does not expose explicit IDs, reverse same-batch edges still refuse,
and this does not wire a shipping CLI form. It is a substrate check, not
shipping graph-JSON plan parity. Contributor
[upstream #5648](https://github.com/gastownhall/beads/pull/5648)
addresses dependency-type alias normalization in ordinary `bd link`, `bd batch`
and create forms; its tests and attribution should be preserved when a CLI
plan route is added.

The native `Deleter` already supplies a bounded, transactional deletion plan
([`issueops/deleter.go`](../../issueops/deleter.go)): without `--force` or
`--cascade`, dependents outside the named set cause a whole-request refusal;
`--force` leaves surviving dependents orphaned; `--cascade` deletes their
transitive closure; `--dry-run` evaluates the same snapshot without writing.
The graph Issue deleter currently refuses *every* incident canonical Link
before calling that native writer
([`internal/storage/graphstore/issue_delete.go`](../../internal/storage/graphstore/issue_delete.go)).
That refusal protects graph integrity but does not prove the cleanup operation
is impossible. The missing proof is a single checked graph transaction that
plans the native deletion and the exact canonical Link tombstones together,
including retained versions and owner History, then rolls back on a staged
failure. Start with one named Issue and one dependency Link; widen to
informational Links, a named multi-Issue set, `--force`, and bounded
`--cascade` only after their outcome contracts are pinned. `bd batch` cannot
exercise this seam because its script grammar has no delete operation.
Calling the public `Unlink` before `DeleteIssue` is not the proof: each opens
its own transaction, and the dependency unlink helper updates the source
Issue's version and mapping. The deletion plan has to account for that source
version and the final-live predecessor of the deletion terminal in one
transaction, including when the source itself is being deleted.
Draft [#209](https://github.com/donnabox/beads/pull/209) is an unmerged
storage proof of this narrower composition: an explicit method tombstones
bounded incident Links and deletes one Issue in one transaction, while the
default refusal remains. Outgoing and incoming Dependencies, an informational
Link, stale guards and staged rollbacks passed on embedded and real
shared-server Dolt without skips. It does not select the CLI force/cascade
policy or multi-ID outcome contract.
Stacked draft [#228](https://github.com/donnabox/beads/pull/228)–[#229](https://github.com/donnabox/beads/pull/229)
extend that storage proof to a checked named set and cascade closure. Focused
tests prove missing-selector refusal and staged rollback on both engines.
Stacked draft [#230](https://github.com/donnabox/beads/pull/230) admits
multi-ID, cascade, and file-selected Issue sets through that composer;
its installed deletion workflows passed both engines without skips, and its
broader embedded graph CLI sweep passed 256 top-level tests with one server-only
skip. These PRs remain unmerged candidates. Dotted native IDs explicitly refuse
until [upstream #7377](https://github.com/gastownhall/beads/pull/7377)
fixes citation rewriting, preserving that contributor's work.

The next CLI slice must preserve the distinction between the delete plan and
confirmation. Shipping's single-ID `bd delete ID` previews a deletion and
refuses when a surviving Issue depends on `ID`; `bd delete ID --force` applies
it and orphans that dependent. In graph mode, incoming native Dependency Links
are the same dependent relationship, while informational Links are merely
incident cleanup. Thus the safe command mapping is:

| Form | Checked outcome needed before CLI wiring |
| --- | --- |
| `bd delete ID` | One read transaction plans Link cleanup, but refuses any incoming Dependency from a surviving Issue. No write. |
| `bd delete ID --force` | One write transaction tombstones all incident Links, removes the Issue, and leaves incoming dependents live and unlinked. |
| `bd delete ID --dry-run` | Same dependent policy and snapshot as the unforced preview, with no write. |
| `bd delete ID --cascade` | Requires a bounded transitive Issue closure and all of its canonical Link tombstones in one write transaction; #209 does not establish this. |

#209's opt-in preview is a **force-mode plan**: it lists incident Links even
when an incoming Dependency exists. Stacked draft
[#210](https://github.com/donnabox/beads/pull/210) supplies a separate
unforced preview that checks dependents inside the same read transaction;
focused embedded and real shared-server tests passed. Stacked draft
[#212](https://github.com/donnabox/beads/pull/212) wires the single-Issue CLI
preview and `--force` routes. Its installed workflow passed both engines and
its six current-head hosted checks passed, with review pending. A disposable
combined checkout included
#163's `--dry-run` alias and passed existing Issue/Memory delete workflows on
embedded Dolt; the server subtests were skipped, so that earlier assembled combination
also passed the broader embedded `TestGraphPreview` run at local-only
`5b0552c61`: 273 top-level passes, one server-only top-level skip, zero
failures. #230 subsequently reconciled #163's alias with the set route and
passed the installed delete-set, single-Issue, and Memory workflows on real
shared-server Dolt as well as embedded Dolt. A separate CLI `ListLinks` call followed by
deletion would make the decision stale and would not close this class bug. The native role checks
dependents from both dependency planes without filtering by kind
([`internal/storage/issueops/delete_role.go`](../../internal/storage/issueops/delete_role.go));
the graph plan must preserve that rule for every admitted native kind.

## Candidate shared work

| Root cause / seam | Exception cluster | Existing substrate | Focused work that could remove the class |
| --- | --- | --- | --- |
| **Graph transaction composition**: native Issue writes and canonical graph projection must land together. | `create --deps`, parent/dependency batches, graph-JSON symbolic batches, linked `delete`, `update --parent`, `duplicate`, `supersede`, `set-state`; potentially multi-Issue update/delete. | `graphstore.withTx`; #197–#202 prove local create/dependency and Markdown batches can be atomic. Draft #231 proves new-Issue symbolic edges and canonical Link retention on both engines; #232 wires that bounded `bd create --graph` form; #233 adds an exact existing local target and #234–#235 admit initial fields and one existing-source edge. #236 pins new-source inception History; #237 maps multiple existing-source edge revisions with their intermediate canonical owned-Link sets; #238 admits new existing-to-existing edges, #239 admits a checked change-free re-add, #240 admits deferred assignment, and #241 admits symbolic metadata references, still inside one transaction. Draft #209 proves single-Issue linked deletion; #228–#229 prove bounded named-set and cascade storage composition, and #230 wires those set forms in the CLI. Draft #222–#223 prove a checked parent-only move; stacked #224–#225 prove parent-plus-scalar composition through one native Issue writer, old/new canonical Links and installed CLI on both engines. Empty detach remains a policy gap. | Continue typed, transaction-scoped Issue and Link composition. Keep validation, native edge changes, graph catalog, retained versions and source-owned History in one transaction. Prove rollback after each stage on embedded and shared-server Dolt. Do not infer no-write from native `ItemResult.Changed=false`: a metadata refresh can still mint a source version. Avoid exporting a raw transaction to CLI callers or assuming native `BatchApplier` alone maintains graph invariants. |
| **Batch outcome policy and command routing**, separate from the transaction engine. | Multi-ID `close` mismatch; multi-ID `reopen`, `update`, `delete`; `--claim-next` coupling and JSON result reporting. | Native `BatchCloser` commits survivors as one act; graph `CloseIssues` uses it. Draft #226–#227 preflight all close selectors in that transaction and route ordinary multi-ID close through it; `bd batch` remains outside the graph-mode route. | Keep a per-command contract table: preflight failures, per-item policy refusals, commit grouping, History and output. Qualify #226–#227 at current head, then add typed multi-resource operations only where a selected command contract calls for them. |
| **One Issue selector model**: native ID, abbreviated ID and canonical graph path are different identities. | Explicit-ID children, positional `link`, `dep` operands, `create --deps` targets, parent filters/updates, list/show/update/delete selectors. | `IssueSelector` resolves exact bare native IDs and exact graph paths inside a transaction; #180 aligns bare flat create identity, #193–#194 add positional abbreviated IDs in the write transaction. A local-only assembly at `9f92ab4aa` reused that resolver for `create --deps` and passed focused embedded and real-server workflows; its candidate bases have not converged in an open PR. | Extend the same transaction-scoped resolver with shipping exact/unique-prefix precedence, explicit-path exactness, ambiguity and deleted/non-Issue rules. Have every write consume its result within the same transaction; no CLI read-then-write lookup. Reuse it across commands instead of reimplementing partial-ID rules. |
| **Relationship semantics and checked read projection**: native dependencies, canonical Links and readiness have different views. | Default `dep list` omits nonblocking kinds; `ready` must ignore informational Memory Links; parent/ready/list/graph/blocked views; Link kind filters. | #165, #174 and #175 expose kind and blocker rules; #169–#172 and #189–#190/#203 add checked parent/ready reads. Graph reads already use read transactions, but several collections refuse after bounded full scans. | Centralize relationship-kind classification and reuse the same checked projection for `dep`, `ready`, `list`, `children`, `blocked` and graph views. Test informational vs blocking edges, native and graph Links, hierarchy and row caps across both engines. Pagination is a separate deferred contract; keep honest refusal above a cap. |
| **Installed Type closure and shared checked selection**: list and count currently choose different populations. | Default `list` can include every Bead Type while `count` is Issue-only; nominal `--type` is exact rather than including refinements. Legacy Issue-classification flags also need precedence over declared Type names. | Installed Type descriptors expose refinement relationships; Melinda's #217 supplies a bounded Memory-family collection reader, but neither establishes the agreed list/count CLI contract in the pinned integration. | Reuse one checked Type-resolution and candidate-selection seam for both commands. Donna selected omitted Type = Issue family, `--type X` = X plus transitive refinements, `--exact-type X` = declared X only, and `--type any` = all Bead Types. Preserve Issue visibility/status/pinned defaults and legacy `bug`/`task` classification filters. Count the complete selection before display limits; refuse incomplete acquisition. Resolve `any` URI reservation and mixed-result JSON admission with Type/output owners before wiring. |
| **Later: legacy CLI compatibility adapter**: the graph store returns canonical records while shipping commands promise Issue-shaped feedback. | Human `show`, JSON `create`/`close`/`delete`/positional `link`, update result shapes, error categories and help. | #151/#152/#156 and other candidate PRs restore individual forms; #216 is a narrow `tag` facade over #196's checked label writer. Graph flag admission is currently command-specific. Draft [#214](https://github.com/donnabox/beads/pull/214) adds a separate canonical mutation envelope with `success`/`failure`/`unknown` and preserves typed operation results; no command consumes it yet, so it does not change current CLI behavior or qualify BDP pin adoption. | After the core operations are proven, define command-specific shipping projections from checked typed results, with a distinct explicit canonical records mode. Use golden differential workflows for human, JSON, stderr and exit codes. Build the adapter earlier only where it materially reduces the cost of the capability work or prevents a false success. |

Graph transaction composition, relationship reads and batch outcome policy are
the first capability probes. Selector work joins them where identity affects
correctness; convenience abbreviations and output rendering follow after the
operations are sound. The final row is the deliberate later compatibility
pass. Full Memory, gate, diff and external Link endpoints are product-scope
questions, not evidence that the transaction engine is missing. Donna selected
Issue-only `bd query` compatibility for Preview 3 without requiring a generic
Memory evaluator; this is an intentional scope boundary.
`external:` dependency targets are an agreed explicit graph-mode refusal;
local targets are the selected implementation path.

`duplicate` and `supersede` share one later composition seam: shipping
resolves two Issues, adds a typed native dependency, then closes the source
through a separate call ([`cmd/bd/duplicate.go`](../../cmd/bd/duplicate.go)).
The graph route currently refuses both. A shared checked graph operation
should resolve both selectors and commit the native edge, canonical Link,
source closure, retained versions and History together, with a staged-failure
rollback proof. Calling public `AddDependency` then `CloseIssue` would repeat
shipping's partial-write exposure and cannot qualify this class. This is
separate from the first selected CRUD/ready slice; neither command is marked
release-blocking by this map.

For multi-Issue and cascade delete, the native `DeleteInTx` body already
preflights missing IDs, checks outside dependents, bounds the transitive
closure, and rewrites surviving neighbors in one transaction
([`internal/storage/issueops/delete_role.go`](../../internal/storage/issueops/delete_role.go)).
That is a useful planner, but calling it alone would leave graph catalog,
canonical Link tombstones, retained terminal versions and History out of sync.
The next proof must resolve the *complete* deletion set to live graph Issue
paths before writing, account for non-graph and wisp members explicitly, and
delete the native rows and every canonical incident Link in the same graph
transaction. Dotted child IDs also expose a citation-rewrite boundary bug;
upstream contributor [gastownhall/beads#7377](https://github.com/gastownhall/beads/pull/7377)
has the relevant fix and conformance tests. It remains open with conflicts, so
any graph cascade work should preserve that prior work and attribution.

## Proposed proof before widening a module

1. Take two or more failing operations from a row and express the same storage invariant through one private seam; demonstrate that a shared fix removes both failures. Start with exact, unambiguous selectors and typed results where possible.
2. Test success, stale/ambiguous selector, later-stage injected failure, no-op and concurrent writer outcomes. Check native rows, canonical catalog, retained versions and History after success and refusal.
3. Run the same installed workflows on embedded and real shared-server Dolt. Record skipped tiers exactly. Then add differential CLI forms, defaults, flags and output checks against shipping. Keep each open PR a candidate until it is selected, reviewed and merged.

This draft does not decide whether all specialized commands must ship in Preview 3, nor does it change the command-by-command limitations in the main matrix.
