# Canonical batch compiler

This pure module compiles the reviewed canonical JSONL authoring surface into
an explicitly selected **experimental version 2** native request. It does not
change or execute the legacy command-text `bd batch` runner, interpret version 1,
advertise a protocol capability, acquire Resources, reserve keys or mutate a
workspace. Version 2 remains an implementation candidate until paired protocol
adoption and actual provider qualification.

`Compile` freezes and validates all input before returning a Batch. JSONL has
one object per nonblank physical line; native JSON requires its own exact
`capabilityVersion`. Native input accepts neither arguments nor a version flag.
The compiler rejects duplicate JSON keys, malformed Unicode, lossy/nonfinite
protocol numbers, unknown member fields, conflicting edits, invalid References,
forward/wrong-kind bindings, and missing/unused/duplicate arguments. It uses the
existing `graphops` JSON admission, `graphpatch` patch grammar and `clirevision`
selector parser. Arguments resolve only designated Resource operands and
explicit selector sources; properties and AST literal values remain literal.
`DecisionJSON` supplies comparison material with capture labels normalized to
defining operation slots. It never rewrites ordinary data or changes `NativeJSON`;
retained outcomes can project equivalent renamed labels without reexecuting.
Caller-assigned `id` stays a complete canonical local ID or absolute owning URI;
ordinary References are absolute in the compiled request.

`CompileMember` validates an independently emitted CRUD or selection member while
retaining unresolved structural binding names. Whole-document `Compile` then
checks their definitions and ordering. `ParseReference` and `ParseOperand` share
the existing CLI URL boundary: plain absolute URLs are literal, local exact pins
use `ID@=TOKEN`, URL pins use `[URL]@=TOKEN`, and batch operands recognize
`batch-ref::NAME`. Offline ordinal/current lookup refuses. Accessors return
copies of every mutable value.

The selection envelope uses the shared review AST and infers collection kinds;
it rejects historical acquisition and public-revision tests of staged content.
**AST shape validation is not semantic/execution admission.** The execution
provider must admit every supported selector, lexical binding, Type policy,
complete set bound, single-Scope read and authorization before publication.
It must refuse unsupported selection rather than run a local preselection or
commit a prefix. This module makes no claim of transaction serializability,
final revision publication, durable keyed recovery or an installed product path.

The unchanged embedded `query.schema.json` and four `testdata/*.json` fixtures
come from `gastownhall/bdp` source `ba8d18c6355cbd6dd3499befe03641874bb01c73`:

- `docs/cli-design/query-ir/query.schema.json` (SHA-256
  `bf5fbf0b89f74273122c28b944907dd3e0fdb057d9266fdee4de9c0dfcbaeaa8`)
- `docs/cli-design/overview/examples/csv-basics.batch.json`
- `docs/cli-design/overview/examples/csv-bindings.batch.json`
- `docs/cli-design/overview/examples/csv-review.batch.json`
- `docs/cli-design/overview/examples/csv-cleanup.batch.json`

These are review artifacts, not a new public schema identity. The offline shape
compiler has no schema-network loader. The package's detached member API is
shared by dry-run adapters and the separate transaction planner.
