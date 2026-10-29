# Type-definition development preview

This unreleased development profile makes the four intrinsic Type definitions
ordinary, locally owned Beads. Use it to inspect definitions and create a local
BeadType definition before packs and aliases are designed. Published user docs
continue to describe the release in `docs/cli-docs.pin`.

Use a new disposable directory with a build of this candidate:

```sh
bd init --graph-mode link --graph-profile types \
  --scope-url https://example.invalid/my-types/
bd show type-for-bead-type
bd show type-for-type
bd show type-for-link-type
bd show type-for-refines
bd list --type type
bd list --type bead-type
bd list --type link-type
bd type create --id foo
bd show foo --json
bd create --id example --type foo --properties '{"title":"Example"}' --json
```

The four intrinsic definition IDs belong to the selected local Scope. Their
intrinsic protocol identities remain common across workspaces. Show returns the
complete persisted record, including exact Type affiliations and owned Refines
Links; closing and reopening the CLI preserves that record and its revision.
The initial Type-family Lists contain four Types, three BeadTypes and one
LinkType. Creating `foo` adds one real local BeadType definition with the fixed
open-object JSON Schema 2020-12 properties schema. A duplicate ID refuses rather
than replacing the definition. Case and whole segmented IDs remain significant.

The last command is an **unsaved prototype**. Its JSON result says
`prototype: true` and `persisted: false`, carries the exact selected Foo Type
pin, and may carry a proposed ID. It allocates no instance ID or revision and
does not save a Bead. Plain output states that no Bead was saved. This temporary
boundary is deliberate; it is not persisted custom-instance support.

The profile is embedded-only and definition-only. Existing dependency and mixed
graph initialization retain their defaults. Existing workspaces are never
migrated, replaced or repaired by selecting this profile. A failed bootstrap
may leave an explicitly incomplete workspace; do not reuse it as a ready store.
Server initialization, ordinary Issue/Memory workflows, general Type update or
delete, refinement-chain authoring, packs and aliases are outside this profile.
Its managed agent guidance advertises the supported commands and installs no
Claude Stop hook.

For contributors, the native initializer and definition reader establish the
checked interpretation context. The immutable `resourceget.DefinitionDocument`
carrier validates representation shape; it does not grant authority or infer
definition membership from ID spelling. Full native and CLI qualification,
including default-layout regression checks, follows [TESTING.md](TESTING.md).
