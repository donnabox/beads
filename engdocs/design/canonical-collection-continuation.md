# Canonical collection continuation: provider work required

Status: implementation design for Preview 3, October 20, 2026. This does not
change the public CLI or BDP contract. The output contract is
[selection and output](../../docs/reference/bdp/cli/selection-and-output.md).

## Existing boundary

`bd list` and `bd link list` use `graphread.Reader.Select`. It projects one
complete `graphstore.CurrentSnapshot` transaction, bounded to 1,000 live
Resources and a 16 MiB acquisition budget. It has no retained cursor. The
`graphread.Pagination` implementation retains complete snapshots and opaque
cursors only in one process. Every `bd` invocation opens a new process, so
constructing that pager per command would publish a continuation the next
command cannot follow.

`CurrentSnapshot.WriterToken` identifies controlled writer state for equality
checks. It is explicitly neither a Scope epoch nor a retained read handle, and
it does not prove absence of out-of-band SQL changes. Re-reading the live
inventory at an offset or by keyset on `--continue` therefore cannot satisfy
BDP's one logical projected snapshot: concurrent create, delete, retype or
mutation could skip, duplicate or replace a delivered Resource.

## Required provider contract

The authority needs an operation to create a bounded, selected, ordered,
immutable read context and return its first page. Each continuation must
identify the same read context, selection, semantic projection, ordering,
authorization view and Scope epoch. The provider must return a cursor at the
*delivered* boundary, including when a caller's total cap is smaller than a
provider page. It must reject an expired context, changed view, changed epoch,
foreign projection or malformed cursor, without restarting against current
state. Representation masks and output format may change under continuation
because they are applied to the retained complete records.

For this CLI, the read context must survive process exit. A service with a
stable lifetime could own BDP's existing in-memory pager. A directly opened
local workspace needs durable authority-owned snapshot storage instead. The
preview store's existing schema has no such table, and `OpenExisting` performs
no DDL or migration. Adding durable storage requires an explicit Preview
schema-version/upgrade decision, not a CLI-side cache with unspecified
ownership, quota, cleanup or backup behavior.

### Recommended Preview 3 implementation

Add a Preview schema v8 table in the same Dolt database, containing a random
opaque context ID, the full selected and ordered BDP records, binding and
read-context fences, selection identity, creation/expiry instants, and encoded
size. The context payload can be one bounded JSON value; it need not duplicate
per-row mutable Resource tables. A continuation token combines the random
context ID with a validated next-item offset, protected from tampering by an
authority secret or by a stored random per-position token. The provider returns
the BDP `next` URL; the CLI treats its value as opaque.

The first-page path captures current state in one read transaction, selects
from that frozen inventory, and atomically admits the retained payload and
its capacity reservation before emitting output. The read and admission may
be separate transactions because the selected bytes are already captured;
admission cannot re-read current Resource rows. If the admission commit
outcome is unknown, the CLI must refuse to print a successful first page.
A later page reads the retained bytes, never current Resource rows.
Under the current complete-read bound, admission is at most 1,000 items and
16 MiB before an additional retained-state cap; page size has its own bound.
Start with the existing pager's provisional limits (32 contexts, 32 MiB total
retained bytes, 5-minute expiry) only after checking that storing the projected
payload and its envelope fits the cap. Reject at capacity before first-page
output. Remove only expired contexts on the access path; never silently evict
an unexpired one. Expiry uses an authority clock and rejects the cursor even if
the persisted row remains until cleanup.

The existing `graph_preview_scope.authority_id` identifies this workspace,
but is not a restore generation. Add an explicit Scope epoch or equivalent
restore fence. A restore must rotate it and invalidate retained contexts.
If backup/restore includes the cursor table, restoration must not resurrect
continuations from the previous epoch. Local authorization is currently one
view; bind that stable view identity now and verify it on every continuation
so a later authorization feature cannot accidentally broaden an old cursor.
Selection identity includes exact structural predicates and future query IR,
not representation flags. It must be compared at continuation; never decode
the cursor to reconstruct a new query.

`OpenExisting` currently performs no DDL or migration. Choose one explicit
v7-to-v8 path before code lands: (a) a tested opt-in preview upgrade that
creates the table and epoch atomically, or (b) require a newly initialized v8
workspace and document how disposable v7 data is exported and reimported.
Changing `SchemaVersion` alone would strand every existing v7 workspace.
Changing the database schema is justified here only if local cross-process
pagination is in scope; a CLI-owned sidecar would require a separate ownership,
security, quota, cleanup, and restore contract. This table is operational read
state, not a new Resource or History record.

## CLI adapter once the provider exists

`canonical_collection.go` should request a page of at most `--page-size` and
at most the remaining `--limit`, so it never discards fetched items. `--paged`
returns the first page and its native `next`; `--unpaged` follows native
continuations to exhaustion or the total cap. JSONL emits its header, items
and a terminal record with the native continuation; a failure after the header
emits an error terminal. The existing collection renderer and `Paging.PlanPage`
already provide most of this accounting. `--continue` must pass the provider's
opaque value rather than parse or synthesize it.

Tests should cross two separate `bd` processes, mutate between pages, and
prove that exact IDs and revisions appear once in the original order. Other
cases: small total limit, short provider page, changed representation,
changed selection, expiration, foreign workspace/view, restore/epoch change,
capacity refusal before output, and a late JSONL failure. The v7-to-v8 path
needs a real database test, including interrupted upgrade or import as
appropriate. Embedded and server Dolt must agree on capacity admission and
page continuation under concurrent clients. Run the affected graphstore,
graphread and command tests in the appropriate Bazel tiers.

Until this facility exists, the current command's refusal on truncation is
the truthful behavior. Its complete unpaged selection remains useful within
the provider's 1,000 Resource / 16 MiB bounds.
