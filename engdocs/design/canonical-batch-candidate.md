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
and informational Links, including prior creation bindings. A batch sees staged
values, checks literal guards against the original committed baseline and assigns
public revisions only to final published states. It preserves no-op revisions,
captures final owned Links once, and permanently consumes a create-then-delete
identity without inventing a Resource or revision for it.

## Recovery and result boundaries

Before applying a request, the CLI saves its resolved workspace, Scope, authority,
database, schema, actor, request and key under the workspace's private
`batch-invocations` directory. The transaction stores its successful request
identity and complete result together. `bd batch --retry-file FILE` uses those
frozen values rather than rereading JSONL, arguments or today's configuration.
Equivalent capture renaming changes only returned presentation labels.

Unknown commit outcomes remain unknown and print the recovery filename. A
rendering failure after confirmed commit reports that the decision committed and
retains recovery information. Successful decisions and consumed identities must
not be pruned in a way that permits reexecution or identity reuse. Failed-outcome
retention and a shared singleton/batch principal-key namespace are still open.

The local result uses existing checked native records inside an experimental
mutation envelope. It is not a newly adopted BDP wire DTO. Public receipt URLs,
Scope snapshots and intermediate Resource revisions are not introduced.

## Remaining coverage

Native execution refuses selection/set members, native Issue/Dependency mutation,
exact Type pins and supplied revision provenance before effects. Pure staged List
and set expansion are implemented separately, but need the checked revisionless
selector integration before the native gate can open. Full Query/Graph evaluation
is not inferred from that subset. Types materialization and current-read Query
have separate owners.

The complete target still needs singleton durable delivery, failed-outcome
retention, Events correlation, native domain publishers, complete output/transport
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
