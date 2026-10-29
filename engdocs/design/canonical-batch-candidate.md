# Canonical CRUD and atomic batch candidate

The canonical JSONL path composes the same Bead and Link mutations emitted by
`--dry-run-mode batch`. It runs separately from the legacy command-text batch
runner. This is an experimental local implementation on a fresh, explicitly
selected layout15 workspace; it does not advertise the selected public BDP Write
profile or upgrade an existing workspace. Published user docs continue to target
the pinned release.

## Try the separate path

Use a disposable directory and the normally built candidate binary. Initialize
with `bd init --graph-mode link --atomic-batch --scope-url URL`; later relative
references use that configured Scope without repeating it on every command.

```sh
# Emit two creations into one JSONL input.
bd create --type memory --id first --batch-bind a --properties '{"title":"First","body":"Source"}' --dry-run-mode batch > changes.jsonl

# Append a second creation.
bd create --type memory --id second --batch-bind b --properties '{"title":"Second","body":"Target"}' --dry-run-mode batch >> changes.jsonl

# Append a Link using captured identities.
bd link create batch-ref::a batch-ref::b --type references --id citation --dry-run-mode batch >> changes.jsonl

# Apply the complete request as one checked transaction.
bd batch --file changes.jsonl --idempotency-key example --json

```

The JSONL compiler checks every member before execution. `--input-format json`
accepts an already compiled request with its explicit capabilityVersion; it does
not reinterpret the retained version1 wire grammar. Member emission and wrapper
emission do not initialize or open a database. BDP/HTTP request preview refuses
until an actual discovered Write route exists.

The native candidate admits six CRUD operations for generic Memory-family Beads
and informational Links, including prior creation bindings. Local List selections
can capture a fixed set for later update/delete members. A subsequent selection
can inspect the current staged bodies of a prior capture; the original set does
not change when a later edit makes another identity match. Selection and writes
use the same checked transaction. A batch sees staged values, checks literal
guards against the original committed baseline and assigns
public revisions only to final published states. It preserves no-op revisions,
captures final owned Links once, and permanently consumes a create-then-delete
identity without inventing a Resource or revision for it.

Existing durable native Issues also support scalar content and metadata updates.
The supported property fields are `title`, `description`, `design`,
`acceptance_criteria`, `spec_id`, `priority`, `estimated_minutes`, `external_ref`
and `due_at`. Native Dependencies support creation, deletion and metadata updates
for recognized native relationship kinds, including kinds that affect readiness.
Their endpoints must be existing durable local Issues. The concrete Link Type
must agree with its `dep_type` property, for example Type `blocks` with
`{"dep_type":"blocks"}`. These paths keep authoritative native rows and publish
one final native Issue version and owner capture after all final Links. Native
scheduling cycles and hierarchy checks apply to the final graph; readiness settles
once over the complete native population, before final audit and journal emission.
A neighbor whose only change is derived readiness gains no Resource revision.
The native storage API requires a nonempty
actor under the ordinary native write policy, even for a canceled or unchanged
mutation. Fresh CLI invocations use the ordinary graph-write actor resolver
(flag, environment, Git identity and existing fallback), then save that choice;
retry uses the saved actor. Read-only selections and generic records retain their existing policy.

Native optional properties use removal for clearing: explicit null or empty
optional strings are refused because the native DTO would omit that presence.
Dates must have canonical UTC whole-second spelling. This admission happens
before a later selection can observe an unrepresentable staged value. Each
authored scalar also passes the ordinary native field validator, including
values or required-field removals later canceled within one change array. Native
Issue status, assignment, leases, notes, labels, full property replacement,
creation, deletion and reclassification remain outside this slice. Native
Dependency endpoint, classification and property edits also remain unavailable.
Metadata updates preserve the existing native assertion's identity and creation
facts. Deleting an old public Link and creating a distinct new Link on the same
pair is an explicit reassertion with fresh creation facts; this first slice
requires the same native classification. A canceled new Link consumes its public
identity without writing a native row or inventing a Resource revision.

This candidate requires a complete, bounded controlled-writer workspace: every
native Issue and Dependency has a checked canonical mapping, every native Issue
participates in retained history, and native readiness is already consistent.
Wisp/external authority, malformed pairs and nonempty native conversation threads
refuse the decision. This complete-closure gate also applies to native metadata
updates. It does not claim atomicity against arbitrary legacy SQL writers or
cross-replica commits. Exact pins are never discarded to make a native endpoint fit.

The supported List fragment uses the shared Query predicate kernel over a closed
revisionless view: identity, Type, properties and optional metadata. Sources may
scan one local collection or use explicit current References/prior captures;
family tests use the complete checked Type inventory. Missing required identities,
unsupported branches and exhausted limits fail the whole batch. This candidate
does not acquire endpoints or execute Graph/full Query selectors.

## Recovery and result boundaries

Before applying a request, the CLI saves its resolved workspace, Scope, authority,
database, schema, actor, request and key under the workspace's private
`batch-invocations` directory. The transaction stores its successful request
identity and complete result together. `bd batch --retry-file FILE` uses those
frozen values rather than rereading JSONL, arguments or today's configuration.
Equivalent capture renaming changes only returned presentation labels.

New decisions return one `results` entry per member with an `ids` array, including
empty selections and empty set writes. Resource outcomes describe final states,
not intermediate postimages. Read-only/empty decisions retain complete recovery
results without a public transaction token. Recovery of the earlier local scalar
candidate preserves its retained `id` spelling; it does not reinterpret that
stored outcome as a new array result.

Unknown commit outcomes remain unknown and print the recovery filename. A
rendering failure after confirmed commit reports that the decision committed and
retains recovery information. Successful decisions and consumed identities must
not be pruned in a way that permits reexecution or identity reuse. Failed-outcome
retention and a shared singleton/batch principal-key namespace are still open.

The local result uses existing checked native records inside an experimental
mutation envelope. It is not a newly adopted BDP wire DTO. Public receipt URLs,
Scope snapshots and intermediate Resource revisions are not introduced.

## Remaining coverage

Saved invocation files retain the original workspace, Scope, actor, request and
key. Retry does not re-read the source file or substitute today's defaults.
The journal uses canonical JSON so HTML-rich content does not expand beyond its
recovery budget; older escaped files remain readable within bounded physical
and canonical limits. If output fails after a confirmed commit, retry returns
the retained result, including its original revisions, rather than applying the
request again. Retained outcomes are checked for structural and identity
consistency without validating historical properties against today's Types.
Owned capture arrays retain the producer's order.

Native execution refuses native mutations outside the finite content and
Dependency slice above, exact Type pins and supplied revision provenance before
effects. The supported staged List fragment
does not imply full Query/Graph evaluation. Ordinary CRUD content copy through
`--from-revision` uses the existing checked writers and composes with the new
canonical object edits. It is a separate historical-content adapter; ordinary
canonical batch update members have no generic copy field. Types
materialization and current-read Query have separate owners.

The complete target still needs singleton durable delivery, failed-outcome
retention, Events correlation, the remaining native domain publishers, complete output/transport
parity and required normal CI before full capability or release claims. Existing
layout11 writers retain their admission boundaries. A layout15 store is never an
automatic fallback for an old workspace.

## Code and validation

- [Compiler](../../internal/canonicalbatch/README.md): strict JSONL/native admission, References, arguments and stable request identity.
- [Staging](../../internal/canonicalbatchstate/state.go): detached ordered working graph; no I/O or invented public revisions.
- [Publisher](../../internal/storage/graphstore/atomic_batch.go): one checked native transaction, final publication, allocations and retained success.
- [CLI](../../cmd/bd/canonical_batch_cli.go): offline member/wrapper emission and explicit native dispatch.
- [Delivery](../../cmd/bd/canonical_batch_execute.go): frozen invocation journal and result recovery.

Native regression cases cover final graph publication and one version per
Resource, close/reopen recovery and capture renaming, no-op revisions, guarded
whole-prefix rollback, allocation-only reservations and deletion of an owner
with its owned Link. Separate CLI workflows exercise actual fresh embedded
initialization, emitted JSONL composition, saved-invocation replay and ordinary
CRUD object edits/guards/deletion. These are cgo tests: run their named cases in
the embedded-capable lane described in [Testing](../TESTING.md), and verify actual
RUN/PASS/zero skips rather than treating a compile or absent test as qualification.
The default and other required Bazel gates still apply.

Additional native cases cover staged List/set publication, fixed membership,
fresh reads of captured bodies, missing-identity rollback, authoritative family
matching and close/reopen of empty retained decisions. The actual CLI workflows
also cover selection/set/empty replay and historical copy combined with keyed
properties and strict metadata patches. Pure tests check shared predicate parity
and strict scalar/array retained-carrier admission. Runtime proof of those paths
does not establish a passing full repository gate or a public Write profile.

Native domain cases additionally cover multiple staged Issue edits plus native
Dependency metadata, one final version per owner, exact final owned-Link capture,
ordinary read equivalence, retained replay, canceled native edits and failure after
final publication facts. Actorless unchanged/canceled native requests must refuse
without retaining any decision or effects. The actual CLI workflow composes native
batch content updates and the separate ordinary historical-copy adapter with the
same property and metadata flags. Native Show explicitly requests `--format
graph-json`; the legacy native detail output remains a distinct presentation.

Dependency CRUD regressions additionally cover final scheduling/readiness and
ordinary Link readback, a derived-only neighbor without a Resource version,
distinct-ID same-pair reassertion and detached deleted backings, transient identity
consumption, final cycle refusal, a valid final graph that reverses an old edge,
legacy history-participation refusal, and rollback after final facts/decision.
A waits-for metadata change uses the existing readiness kernel without emitting
a new-assertion human audit. The actual CLI workflow emits native Link creation,
metadata update and deletion into JSONL, applies them through the separate batch
route and compares creation with ordinary Show. These targeted passes do not
replace the required whole-repository CI or establish a public Write profile.
