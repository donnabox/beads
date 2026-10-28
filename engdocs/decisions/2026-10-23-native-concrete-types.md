# Native concrete Types in Preview 3

> Historical identity mapping: the generated path/hex spellings and generation
> below are superseded by [meaningful Type identities](2026-10-25-meaningful-type-identities.md).
> This record preserves the earlier decision and its evidence.

Status: draft implementation preparation. Release integration and playground qualification compose this candidate with the importer and CLI owners' candidates. The existing product ledger is gastownhall/beads#7656; the BDP contract authority remains gastownhall/bdp#59 and #282.

## Declared affiliation

Issue and Dependency are abstract family Types. A native Issue declares its concrete classification Type; a native Dependency declares its concrete kind Type. The native classification column controls the declaration through a checked adapter.

| Family | Abstract identity within Scope | Concrete identity within Scope | Native column |
| --- | --- | --- | --- |
| Issue | `types/issue` | `types/native-issue/` followed by lowercase hex of the UTF-8 classification | `issues.issue_type` |
| Dependency | `types/dependency` | `types/native-dependency/` followed by lowercase hex of the exact native kind | `dependencies.type` |

For example, native `task` is `types/native-issue/7461736b` and native `blocks` is `types/native-dependency/626c6f636b73`. Prefix each relative identity with the actual Scope URL. These are Scope-local identities. Distinct classifications, Unicode byte sequences and families remain distinct. Native aliases normalize at the existing domain boundary. Memory `types/decision` and native Issue `decision` have different identities.

The native Issue limit remains 32 Unicode scalars; Dependency kinds retain their native 32-byte limit. Canonical identities keep the complete encoded value. The opaque private catalog row key does not become a public identity.

## Admission and lifetime

Native writers authorize classifications and enforce their existing lifecycle and scheduling rules. A descriptor's presence does not authorize a new custom Issue after its configuration is removed. Ordinary CLI Dependency creation keeps its existing known-kind admission; stored/imported native-valid unfamiliar kinds can retain their exact declared meaning.

Fresh initialization installs the fixed Memory definitions, the abstract native bases and concrete native builtins. Configured unused custom Issue definitions can be derived for discovery. On first authorized native use, a custom Issue or unfamiliar Dependency declaration acquires an immutable binding in the existing checked Type catalog, in the same transaction as the native write, graph mapping and retained state. Resource identity collisions, stale guards and write faults refuse atomically.

Bindings survive reclassification, configuration removal and deletion. Current and exact retained reads validate their namespace, decoded value, kind, canonical bytes and fingerprint. Reads do not register Types. Discovery may therefore show a retained custom Type that current configuration no longer authorizes for new use. The preview's bounded acquisition limits still apply.

Native reclassification uses the native update writer. The `--type` setter and an explicit `issue_type` property must agree. Omitted affiliation preserves the current value. Replacing ordinary properties without a classification setter also preserves affiliation; ordered explicit classification edits retain their order and final-state validation. Native and graph state, ownership captures and history commit together.

## Product selectors and portable identity

```sh
bd create "Investigate failure" --type bug
bd update BEAD --type feature
bd list --type issue
bd list --exact-type task
bd list --type any
```

In the Beads product, `--type X` on a collection selects X and its refinements; `--exact-type X` selects the exact declared Type. Bead listing defaults to the Issue family. Bare native classifications such as `task` resolve through this adapter. Use `types/decision` for Memory Decision and `decision` for the native Issue shorthand. `--type` on create/update remains a setter.

Generic BDP relative operands retain their literal Scope resolution. Relative `task` in that interface resolves against Scope without a native namespace rule. Portable structured output and canonical equivalence use the actual full concrete URI, including any supported definition reference. The wire `type` predicate remains nominal, `conformsTo` remains family selection, and raw BDP collection requests retain their domain-neutral default. The existing `--conforms-to` collection alias retains its canonical absolute URI spelling.

Native search, ready-work and dependency workflow filters retain their domain meanings. Generic Bead counting shares the selected Type predicates; native Issue grouping and filters retain the native count route. Legacy/native JSON can keep classification fields, while canonical records expose and validate the actual declared Type.

Bare `any` is the collection wildcard. Full nominal URI handling does not inspect a URI basename to manufacture that wildcard. The uniform policy for pre-existing configured `any` classifications and other reserved shorthand collisions remains an explicit integration review item; no data is renamed or removed here.

## History, import and rollout

This correction requires Type generation 11 and adds no Type-specific DDL. The common public-identity integration independently sets the workspace schema; a composed candidate must match both its schema and Type generation. Earlier Type generations are refused unchanged. It performs no destructive reinitialization or in-place historical rewrite. Existing workspaces and their compatible binaries remain available; qualified integration can use a separate fresh sibling workspace.

Retained native Issues derive affiliation from their retained native classification, preserving earlier Type identities after reclassification. Retained Dependency snapshots keep their concrete kinds. Current reads, ownership groups, selection and export use the same checked mapping.

Iris owns importer assembly. Supported current-state export/import into a fresh workspace must qualify against the composed candidate, preserving native values and the target Scope's authoritative mapping. Current-state interchange does not transfer complete native history or preserve historical revision tokens across Scopes. Preserve the old workspace for those exact historical references. Rollback uses the old binary with its unchanged old workspace; a new generation is not silently downgraded.

Full public HTTP Type-description conformance remains deferred under the existing Preview 3 scope ruling. The compatible closed descriptor shape records ancestry and describes abstract bases in text; it adds no unsupported abstract field or invented finite ownership maximum. This candidate does not activate general Type registration, Type CRUD or a new protocol profile.
