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
The ordinary human view separates the concept, whole local ID, revision and
metatype. Show presents the descriptor fields with readable JSON and explicitly
summarizes its retained schema environment. Use `--format detail` or `--json` for
the complete Resource, including all schema resources, metadata and owned Link
captures. List shows identities and affiliations instead of repeating the full
descriptor payload. These views do not change stored records or machine output.
The initial Type-family Lists contain four Types, three BeadTypes and one
LinkType. Creating `foo` adds one real local BeadType definition. A fresh
publication2 workspace leaves `propertiesSchema` absent when no local constraints
are supplied. Absence adds no constraints; all inherited constraints still apply.
An existing publication1 workspace retains its original explicit open-object JSON
Schema 2020-12 descriptor. The new binary does not rewrite or upgrade it. A
duplicate ID refuses rather than replacing the definition. Case and whole
segmented IDs remain significant.

The last command is an **unsaved prototype**. Its JSON result says
`prototype: true` and `persisted: false`, carries the exact selected Foo Type
pin, and may carry a proposed ID. It allocates no instance ID or revision and
does not save a Bead. Plain output begins `UNSAVED PREVIEW` and explains that the
proposed instance will not appear in `bd list`. The saved `foo` Type definition
does appear there. This temporary boundary is deliberate; it is not persisted
custom-instance support. The preview accepts only the checked generated
definition for its exact retained publication; it does not validate arbitrary
authored Type constraints or treat `null` as an omitted schema.

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
