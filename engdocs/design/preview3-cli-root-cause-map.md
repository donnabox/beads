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

**Canary interpretation:** ordinary `bd batch` is a control for native
transaction plumbing and rollback, not a graph-mode release canary. Its narrow
grammar admits create, update, close, dependency add and dependency remove;
it does not admit delete or the full top-level command flag sets. A stronger
graph canary must combine an Issue and a canonical Link, inject failure after
the native edge but before retained Link/History publication, and show that
both embedded and shared-server Dolt leave no partial state. A linked delete
canary should likewise prove either complete intended cleanup or a refusal
that changes nothing; the cleanup policy still needs a command contract.

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
embedded Dolt; the server subtests were skipped, so the assembled combination
also passed the broader embedded `TestGraphPreview` run at local-only
`5b0552c61`: 273 top-level passes, one server-only top-level skip, zero
failures. It still needs real shared-server qualification. #163 separately owns the alias
and must be reconciled without dropping its tests. A separate CLI `ListLinks` call followed by
deletion would make the decision stale and would not close this class bug. The native role checks
dependents from both dependency planes without filtering by kind
([`internal/storage/issueops/delete_role.go`](../../internal/storage/issueops/delete_role.go));
the graph plan must preserve that rule for every admitted native kind.

## Candidate shared work

| Root cause / seam | Exception cluster | Existing substrate | Focused work that could remove the class |
| --- | --- | --- | --- |
| **Graph transaction composition**: native Issue writes and canonical graph projection must land together. | `create --deps`, parent/dependency batches, linked `delete`, `update --parent`, `duplicate`, `supersede`, `set-state`; potentially multi-Issue update/delete. | `graphstore.withTx`; #197–#202 prove local create/dependency and Markdown batches can be atomic. Draft #209 proves an opt-in, single-Issue linked-delete composition without changing the default refusal. Draft #222–#223 prove a checked parent-only move across native edges, old/new canonical Links, child History and installed CLI on both engines; combined edits and detach remain gaps. | Continue typed, transaction-scoped Issue and Link composition. Keep validation, native edge changes, graph catalog, retained versions and source-owned History in one transaction. Prove rollback after each stage on embedded and shared-server Dolt. Avoid exporting a raw transaction to CLI callers or assuming native `BatchApplier` alone maintains graph invariants. |
| **Batch outcome policy and command routing**, separate from the transaction engine. | Multi-ID `close` mismatch; multi-ID `reopen`, `update`, `delete`; `--claim-next` coupling and JSON result reporting. | Native `BatchCloser` commits survivors as one act; graph `CloseIssues` uses it; the ordinary graph `close` path still loops over `CloseIssueWithOptions`. `bd batch` is not a graph-mode route. | Make a per-command contract table: preflight failures, per-item policy refusals, commit grouping, history and output. First reconcile `close A MISSING` without assuming every policy refusal rolls back. Then add typed multi-resource graph operations only where the selected command contract calls for them. |
| **One Issue selector model**: native ID, abbreviated ID and canonical graph path are different identities. | Explicit-ID children, positional `link`, `dep` operands, `create --deps` targets, parent filters/updates, list/show/update/delete selectors. | `IssueSelector` resolves exact bare native IDs and exact graph paths inside a transaction; #180 aligns bare flat create identity, #193–#194 add positional abbreviated IDs in the write transaction. | Extend a shared transaction-scoped resolver with shipping exact/unique-prefix precedence, explicit-path exactness, ambiguity and deleted/non-Issue rules. Have every write consume its result within the same transaction; no CLI read-then-write lookup. Reuse it across commands instead of reimplementing partial-ID rules. |
| **Relationship semantics and checked read projection**: native dependencies, canonical Links and readiness have different views. | Default `dep list` omits nonblocking kinds; `ready` must ignore informational Memory Links; parent/ready/list/graph/blocked views; Link kind filters. | #165, #174 and #175 expose kind and blocker rules; #169–#172 and #189–#190/#203 add checked parent/ready reads. Graph reads already use read transactions, but several collections refuse after bounded full scans. | Centralize relationship-kind classification and reuse the same checked projection for `dep`, `ready`, `list`, `children`, `blocked` and graph views. Test informational vs blocking edges, native and graph Links, hierarchy and row caps across both engines. Pagination is a separate deferred contract; keep honest refusal above a cap. |
| **Later: legacy CLI compatibility adapter**: the graph store returns canonical records while shipping commands promise Issue-shaped feedback. | Human `show`, JSON `create`/`close`/`delete`/positional `link`, update result shapes, error categories and help. | #151/#152/#156 and other candidate PRs restore individual forms; #216 is a narrow `tag` facade over #196's checked label writer. Graph flag admission is currently command-specific. Draft [#214](https://github.com/donnabox/beads/pull/214) adds a separate canonical mutation envelope with `success`/`failure`/`unknown` and preserves typed operation results; no command consumes it yet, so it does not change current CLI behavior or qualify BDP pin adoption. | After the core operations are proven, define command-specific shipping projections from checked typed results, with a distinct explicit canonical records mode. Use golden differential workflows for human, JSON, stderr and exit codes. Build the adapter earlier only where it materially reduces the cost of the capability work or prevents a false success. |

Graph transaction composition, relationship reads and batch outcome policy are
the first capability probes. Selector work joins them where identity affects
correctness; convenience abbreviations and output rendering follow after the
operations are sound. The final row is the deliberate later compatibility
pass. Full Memory, gate, diff, generic query and external Link endpoints are
product-scope questions, not evidence that the transaction engine is missing.
`external:` dependency targets are an agreed explicit graph-mode refusal;
local targets are the selected implementation path.

## Proposed proof before widening a module

1. Take two or more failing operations from a row and express the same storage invariant through one private seam; demonstrate that a shared fix removes both failures. Start with exact, unambiguous selectors and typed results where possible.
2. Test success, stale/ambiguous selector, later-stage injected failure, no-op and concurrent writer outcomes. Check native rows, canonical catalog, retained versions and History after success and refusal.
3. Run the same installed workflows on embedded and real shared-server Dolt. Record skipped tiers exactly. Then add differential CLI forms, defaults, flags and output checks against shipping. Keep each open PR a candidate until it is selected, reviewed and merged.

This draft does not decide whether all specialized commands must ship in Preview 3, nor does it change the command-by-command limitations in the main matrix.
