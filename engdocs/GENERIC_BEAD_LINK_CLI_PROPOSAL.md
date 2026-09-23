# Generic Beads and Links: CLI proposal

**Status: proposed CLI contract for team review, 2026-09-23.** These commands describe target behavior, not shipped functionality. Please review the command shapes, defaults, compatibility changes and unresolved model choices on their own merits. Implementation sequencing and effort are intentionally outside this review.

The proposal extends familiar `bd` commands over one authoritative graph. Select graph storage once with `bd init --graph-mode link`; later commands use that workspace automatically. Create a Link with `link`, inspect it with `show`, change its properties with `update links/<id>`, and remove it with `unlink`. No shell alias is required.

**Suggested reading:** §3 command selection, §4 grammar and compatibility deltas, §6 Link lifecycle, §10 Memory commands, then §17 open decisions. Each **Proposed** choice is open for feedback. “Required” describes the proposed contract; it does not imply approval or implementation.

## 1. Starting points

This starts from [Chris’s Memory proposal #5877](https://github.com/gastownhall/beads/issues/5877), its later [export clarification](https://github.com/gastownhall/beads/issues/5877#issuecomment-5555018914) and [delivery-budget clarification](https://github.com/gastownhall/beads/issues/5877#issuecomment-5555009574), plus [Donna’s feedback](https://github.com/gastownhall/beads/issues/5877#issuecomment-5779625219) and [Scope-integrity discussion](https://github.com/gastownhall/beads/issues/5877#issuecomment-5779725641).

History behavior draws on [#5898](https://github.com/gastownhall/beads/issues/5898) and [Jim’s `versions` work](https://github.com/gastownhall/beads/pull/6661). The protocol reference is [public BDP at `1fe8cf32`](https://github.com/gastownhall/bdp/blob/1fe8cf32afabd02ca62d86548409f82dd756e357/docs/specs/bdp.md): [owned Links](https://github.com/gastownhall/bdp/pull/17), wildcard ownership, pins, and [History wire definitions](https://github.com/gastownhall/bdp/pull/30) already exist. Protocol definitions do not establish runtime support. Common top-level application metadata remains a proposal, and Task/Bug nominal typing remains an explicit open question.

## 2. Domain contract

| Term | Meaning |
|---|---|
| Scope | The authority and identity namespace within which local integrity is enforced |
| Bead | Resource with permanent identity, immutable declared Type, mutable properties and common metadata |
| Link | Independently identified directed Resource with immutable Type and endpoint references, mutable properties and common metadata |
| Reference | An endpoint or reference-valued field; not another name for a Link |
| Issue | Issue-typed Bead; Chris’s “Task” means this workflow kind, not just the `task` Issue subtype |
| Dependency | Issue-domain Link between Issue-typed Beads, with the existing dependency-kind semantics |
| Memory | Ordinary descriptor-backed Bead with Markdown content and mandatory retained History; no Issue workflow state |
| Revision | Opaque current Resource concurrency token, compared only for equality |
| Version | Immutable retained historical state with a portable, nonreusable address |
| Local revision | Jim’s store-local history ordinal, useful for display/order; not a portable version address |
| Alias | Mutable, reusable locator under `alias/`; never canonical identity or stored endpoint data |

One fact has one authoritative representation. Generic, Issue, Memory, CLI, and supported BDP writes share validation and transactional effects. Specialized physical tables are permissible; duplicate authoritative Issue/Dependency stores are not.

**Major open design question:** how should mutable Issue classifications relate to nominal graph Types? The previous draft treated `issue_type` as a settled solution; Donna’s review reopens that assumption. The rest of this draft uses the single-Issue-Type/property approach provisionally so compatibility examples remain concrete. It does **not** resolve the nominal typing gap.

Existing Issue event history, Dolt history, retained Bead versions, and Scope identity recovery are distinct facilities.

### 2.1 Issue subtypes versus immutable declared Type — decision required

Today `bd create --type task` and `bd update ID --type bug` use a mutable Issue classification. It would be useful for Task, Bug and other classifications to be nominal Types conforming to Issue, so descriptors and endpoint constraints can express their meaning. Under current BDP, however, a Resource’s declared Type is immutable. `conformsTo` does not make a Resource retypeable, and swapping in new schema content under an existing Type ID is not an escape hatch.

| Choice | Benefit | Cost / unresolved work |
|---|---|---|
| One immutable Issue Type; mutable `issue_type` property | Preserves current classification changes and identity; smallest integration | Task/Bug are not nominal Types. Their constraints remain domain rules/property predicates. This is a real modeling limitation, not completed nominal typing |
| Immutable nominal Task/Bug Types | First-class subtype schemas and endpoint conformance | Existing task→bug edits must refuse or create another identity. Delete/create is not a compatible replacement for preserving the same Issue, references and History |
| Explicit atomic reclassification with stable identity | Nominal Types plus familiar classification changes | Requires a reviewed BDP model change: Type-dependent endpoint validation, incoming/outgoing constraints and ownership, historical Type identity, caches, migration, concurrent edits and client compatibility all need defined behavior |

**Recommendation pending ruling:** keep the property model only as an explicitly limited early slice; decide the intended full model before freezing public Issue descriptors or claiming complete compatibility. Do not silently choose immutable nominal Types and break `update --type`, or claim the property model provides nominal Task/Bug typing. Type mutation for Beads is currently prohibited; allowing it would be an explicit design change. Link Type/endpoint immutability remains the current contract too.

### 2.2 Where common metadata belongs — recommendation, not a ruling

The requirement is an open application metadata object on **both** Beads and Links. Its values are durable state: changes affect revision/History and owned-source state. Placement is not yet settled.

| Placement | Schema / Type implications | Tradeoff |
|---|---|---|
| Top-level `metadata`, beside `properties` | One shared Resource-member schema: JSON object with open members, plus ordinary JSON/admission/size limits. No universal nominal Bead/Link Type required | Uniform on every Type, separate from domain properties; needs a BDP member/update/History amendment and client support |
| `properties.metadata` | Ordinary property patching works. To guarantee availability on every Type, every applicable schema must admit it; an optional reusable schema alone cannot guarantee that | Fewer wire members, but reserves a domain property name and can conflict with closed third-party schemas or different domain meanings |
| `properties.metadata` supplied by universal root Types | Would introduce universal Bead and Link Type IDs plus required conformance, or equivalent mandatory descriptor rules | A larger Type-system change merely to share a record. Existing BDP explicitly publishes no universal root Type IDs; conformance intersects schemas and cannot override a child schema that rejects `metadata` |

**Recommended draft default:** keep top-level `metadata`; define its open-object schema once in the common Resource contract. The `--metadata` CLI option remains useful either way, but the resulting JSON shape and patch mapping must be decided before publication. This document’s examples use the recommendation, not an already-approved BDP field. If nested metadata is selected instead, revise the entire JSON/patch/descriptor contract consistently; do not publish both independent authoritative records.


An informational Link does not affect readiness merely because it touches an Issue. Workflow Dependency Types require Issue endpoints. Memory never becomes ready, blocked, assigned, claimed, closed, or ephemeral. Applications interpret content and metadata; Beads does not execute Memory text or infer Links from Markdown URLs.

## 3. Invocation, compatibility, and selection

### 3.1 Initialize the storage model once; keep command meanings explicit

**Reconciliation:** the older repository spec made `--graph-mode` select between separate Issue/Dependency and Bead/Link graphs, and allowed a config default. This proposal instead uses fresh workspace opt-in and one authoritative graph for Issues, Dependencies and other Resources. Storage selection should not also change the meaning of an established CLI flag.

**Proposed correction for review:** `graph-mode` selects the workspace’s storage model at initialization/migration. It is not an ordinary per-command choice of storage or of what an existing flag means.

```sh
bd init --graph-mode link --scope-url https://example.invalid/demo
bd remember "The plan" --id beads/plan
bd show beads/plan
```

Init persists and validates a durable workspace-format marker. Every subsequent command reads it automatically. In graph mode, old Issue commands and new generic commands operate on the **same** authoritative records. Unsupported paths refuse before effects; no fallback to an old Issue table. Existing workspaces remain legacy until explicit migration. During opt-in rollout, plain `bd init` retains its existing legacy default; selecting graph mode once opts the new workspace in.

There is **no storage override** on an ordinary command. Retain the drafted root `--graph-mode` spelling and `BD_GRAPH_MODE` only as assertions of the existing workspace format: matching succeeds; mismatch reports `workspace_mode_mismatch` without opening another store or migrating. If both are supplied and disagree, refuse rather than ignore one. Thus the originally requested `bd --graph-mode link show beads/plan` still works, but `bd show beads/plan` is sufficient. `config set graph-mode …` cannot convert a workspace; migration owns that transition. This changes the older unimplemented flag design and must be recorded as an amendment, not attributed to a prior settled ruling.

To avoid making old flags mean something different after init, use **additive generic syntax**:

| Existing form retained | Explicit generic form proposed |
|---|---|
| `bd create TITLE --type task` (Issue classification) | `bd create --resource-type memory --properties JSON` (declared Resource Type) |
| `bd link A B --type blocks` (Dependency) | `bd link SOURCE TARGET --resource-type related` (registered Link Type) |
| `bd update ID --title TEXT`, `--type bug` | `bd update beads/PATH --patch JSON`; `bd update links/PATH --patch JSON` |
| `bd list --type task` (Issue filter) | `bd list --kind bead --resource-type TYPE`; `bd list --kind link` |
| `bd show ISSUE-ID` (Issue projection) | `bd show beads/PATH` or `links/PATH` (Resource projection) |
| `bd types`, `bd graph`, `bd status` | `bd types --kind bead`, `bd graph --view generic BEAD`, `bd status --graph` |

`--resource-type` is proposed new spelling, not a shipped flag. It replaces the previous draft’s mode-dependent generic `--type`; the benefit is preserving established `--type` meanings without carrying a mode flag everywhere. New commands such as `links`, `alias`, `compare` and `changes` already identify their function. Generic mutation flags (`--patch`, `--properties`) and explicit canonical Resource selectors identify the generic contract; mixing incompatible legacy and generic flags is an error rather than precedence magic. `--metadata` alone on a familiar Issue ID keeps its old Issue contract; on an explicit canonical Resource selector it uses the new guarded Resource contract. Whole-resource replacement is always explicit.

Memory-specific commands work over canonical Memory; Issue-specific commands (`ready`, `close`, `dep`, etc.) retain Issue meanings in the initialized graph. The `bd link … --type related` mixed-Memory bridge from Chris remains an explicit compatibility proposal (§14), not an inference that every Type string is generic.

### 3.2 Selectors and identity

Generic commands accept:

- Exact local canonical paths: `beads/plan`, `beads/team/release-policy`, `links/policy-citation`.
- Absolute canonical Resource URLs in the selected Scope.
- Explicit aliases: `alias/releases/latest`, resolved once to a canonical Resource.
- Familiar Issue IDs where the command expects an Issue. Mapping to graph identity is recorded, not guessed across Scopes.

Generic selectors do not use fuzzy prefix matching. Existing Issue commands retain their documented partial-ID rules. Ambiguous or wrong-kind selectors fail before mutation. Explicit resource roots and `--` disambiguate unusual shell arguments.

`--id` on generic creation accepts the full local canonical path of the correct kind. Omitting it requests a generated identity. An allocated identity cannot be reused after deletion or erasure. Reopening or restoring uses the same identity. A canonical path is not renameable.

Aliases are resolved before submitting a mutation; guards bind to the resulting canonical identity, so concurrent alias repointing cannot redirect it. Persist canonical references, never aliases. Cross-Scope aliases require explicit successful resolution before use as canonical references; an opaque external URI is not falsely described as a resolved alias.

### 3.3 Type names

`--resource-type` accepts an installed absolute Type ID or an unambiguous installed CLI name. The authority validates against the pinned installed descriptor closure, not a live network fetch. Unknown/uninstalled Types fail; no typo creates a new Type.

Proposed built-in CLI names: `issue`, `memory`, `related`, and qualified Dependency names such as `dependency/blocks`, `dependency/related`, `dependency/parent-child`. Actual immutable descriptor URLs must be published by `types get`; examples do not invent their production URLs. `related` is a baseline informational Type, not a declaration that every informational Type inherits from it.

Readers preserve unknown-but-installed Types and their properties. Lack of a specialized renderer does not make them unreadable. Custom Type installation is in the complete proposed CLI; built-in Types suffice for a basic mixed-Bead workflow.

## 4. Command grammar

The following is the proposed generic grammar. Brackets denote optional arguments, not literal shell syntax. `JSON` means a JSON literal, `@file`, or `@-` for stdin. An invocation may consume stdin only once.

```text
bd create --resource-type TYPE [--id beads/PATH] [--properties JSON] [--metadata JSON]
bd show RESOURCE [--version VERSION | --at INSTANT] [--json]
bd list --kind bead|link [--resource-type TYPE] [--conforms-to TYPE]
       [--source BEAD] [--target REFERENCE] [--limit N] [--after CURSOR | --all]
       [--full] [--json]
bd update RESOURCE (--patch JSON | --properties JSON)
         [--metadata JSON | --metadata-patch JSON] GUARD [--json]
bd update RESOURCE (--metadata JSON | --metadata-patch JSON) GUARD [--json]
bd edit RESOURCE [--field properties|metadata] [--json]
bd delete RESOURCE [GUARD] [--force] [--erase | --unlink-incident] [--json]
bd link SOURCE TARGET --resource-type TYPE [--id links/PATH]
       [--target-version VERSION] [--properties JSON] [--metadata JSON]
       [SOURCE-GUARD] [--request-id TOKEN] [--json]
bd unlink LINK GUARD [SOURCE-GUARD] [--json]
bd unlink SOURCE TARGET --resource-type TYPE [--target-version VERSION]
         GUARD [SOURCE-GUARD] [--json]
bd links BEAD [--direction in|out|both] [--resource-type TYPE]
        [--version VERSION] [--resolve-targets]
        [--limit N] [--after CURSOR | --all] [--json]
bd graph --view generic BEAD [--direction in|out|both] [--resource-type TYPE]
        [--depth N] [--max-nodes N] [--max-links N] [--json]
bd types --kind bead|link [--json]
bd types get TYPE [--json]
bd types install FILE [--dry-run] [--json]
bd alias list [--target RESOURCE] [--json]
bd alias set alias/PATH RESOURCE (--if-absent | --if-target RESOURCE) [--json]
bd alias remove alias/PATH --if-target RESOURCE [--json]
bd versions RESOURCE [--limit N] [--after CURSOR | --all] [--json]
bd compare RESOURCE --from VERSION --to VERSION [--json]
bd restore RESOURCE --version VERSION [GUARD] [--apply] [--json]
bd changes (--after CHECKPOINT | --since INSTANT | --from-start)
          [--resource-type TYPE] [--limit N] [--json]
bd request show TOKEN [--json]
bd status --graph [--json]
```

`GUARD` is `--if-revision REVISION` or explicit `--unconditional`. Restoring a deleted Resource instead uses `--if-deleted TOKEN`, returned by its authorized deletion-state read; see §11. `SOURCE-GUARD` is `--if-source-revision REVISION` or explicit `--unconditional-source`, required for mutations of source-owned Links. Generic Link update/delete also accepts `SOURCE-GUARD`, even though the condensed shared grammar above omits it. Read-only delete/restore previews need no guard; apply requires one and revalidates it. Preview returns the observed revision or deletion token for the subsequent explicit apply.

All generic durable mutations accept optional `--request-id TOKEN` and optional `--message TEXT` for change context. Actor selection follows existing `--actor` rules; attribution is an assertion, not proof of authenticated identity. `--json` and diagnostic rules are in §13. `--dry-run` on creation/update/link validates without reservation, allocation, or history; its success is not a promise a later write will pass.

This grammar deliberately does not overload existing `create --file` (Markdown Issue batches), `create --graph` (Issue graph plans), `show --as-of` (Dolt commit/branch), `show --refs` (legacy Issue references), or `diff` (Dolt comparisons).

### 4.1 Exact compatibility deltas at a glance

“Unchanged” below means the command’s public meaning/output is retained; its backing storage changes after opt-in. “Changed” is a deliberate behavior change, not hidden under “generic support.” §14 covers the remaining command families.

| Surface | What exists at the inspected Beads baseline | Proposed result | Compatibility classification |
|---|---|---|---|
| `init --graph-mode link` | No production root graph-mode flag/bootstrap | Provision and persist graph workspace once | **New option and storage format**; existing workspaces require migration |
| `create TITLE --type task`, `update ID --type bug` | Create/change Issue classification | Same meaning, subject to the unresolved nominal-Type decision | **Unchanged provisional contract; major open risk** |
| `create --resource-type TYPE --properties JSON` | Absent | Create any installed Bead Type | **New additive form**; not a changed `--type` |
| `link A B [--type blocks]` | Dependency, B blocks A | Same semantics/output against canonical Links | **Unchanged** |
| `link S T --resource-type TYPE` | Absent | Generic Link creation with Link identity | **New additive form** |
| `link S T --type related` with Memory endpoint | Memory endpoint unsupported | Bridge to registered generic informational Link | **Changed accepted target kinds**, explicitly proposed; both-Issue legacy behavior preserved |
| `show ISSUE-ID`, bare `list`, `types`, `graph`, `status` | Issue projections | Same default projections | **Unchanged**; no mode-dependent flag reinterpretation |
| `show beads/PATH`, `show links/PATH`; `list --kind …`; `types --kind …`; `graph --view generic`; `status --graph` | Generic forms absent | Resource read/discovery/inspection | **New explicit selectors/options and new JSON contract** |
| `update links/PATH --patch …`; Resource `edit`, guarded delete | Absent for generic Resources | Shared property/metadata mutation contract | **New forms**; old Issue guards/exit codes retained |
| `unlink`, `links`, `alias`, `compare`, `changes`, `request show`, `scope …` | No equivalent generic command family | Defined in this specification | **New commands**, not renamed existing ones |
| `versions ISSUE-ID` (Jim’s separate stack) | Local recorded Issue history, not integrated at graph baseline | Retain existing projection; canonical Resource selector adds generic History | **Existing candidate retained + additive selector contract** |
| `restore ISSUE-ID [--apply]` | Pre-compaction Issue-content recovery | Same meaning; explicit `--version` selects retained-version restoration | **New option**, no silent reassignment of bare `restore` |
| `remember BODY`, including one token | Store-dependent recall/write and invented key | Always writes; generated canonical ID, no invented key | **Changed behavior** per Chris R31 |
| `remember --key K` on existing Memory | Upsert without the new generic guard contract | Canonical adapter with guard or explicit unconditional update | **Proposed breaking write-precondition change**, still a ruling to make |
| `recall`, `forget` | Keyed; empty body confused with absence; forget hard-deletes | Canonical/alias selectors; empty-body correctness; forget preserves History | **Changed selectors, lifecycle, and proposed guards** |
| `memories` text | Key/value listing/search | Bounded compact summaries | **Changed default output/pagination** |
| `memories --json` | Complete key→body map with `schema_version` collision defect | Keep body map and fix collision; `--format records-json` is separate | **Preserved schema with explicit bug fix; new opt-in schema** |
| `prime` and generated context | May inject Memory bodies | Selective retrieval guidance without bodies | **Changed behavior** per Chris R31 |
| `export`, automatic export | Existing Issue-oriented behavior | No newly introduced Memory records by default; explicit full Memory export; whole-store backup includes it | **Exposure contract made explicit**; `--format graph-json` is a new opt-in format |
| `delete BEAD --unlink-incident` | No generic atomic cascade form | Optional atomic incident cleanup plus deletion | **New proposed destructive option**, default remains restrictive |

## 5. Resource creation, reading, and updates

### 5.1 Create

`bd create --resource-type …` creates one generic Bead. `bd link … --resource-type …` creates one generic Link. Both return canonical identity, immutable Type, initial revision, and committed state; a Memory also returns its stable version address. The legacy forms retain their documented response projections. Creation is atomic with required domain initialization and History. A failed request leaves no partial Resource. An explicit ID collision is a conflict, not update/upsert.

`properties` and `metadata` must be JSON objects. Missing means `{}` unless the descriptor or domain requires values. Reject duplicate JSON member names rather than silently selecting one value. Validate the complete resulting Resource, installed Type constraints, domain policy, and limits before commitment. Do not coerce strings into numbers/booleans, discard unknown metadata, or silently round unsupported numbers. Report a failing JSON Pointer where disclosure is permitted.

An Issue created generically receives the same defaults, labels, domain validation, and derived effects as ordinary `bd create`. An unsupported Issue property must fail rather than disappear. A Memory created generically follows the same title/body and History rules as `remember`.

### 5.2 Show and discovery

`bd show` reads one explicitly selected Resource. It may return the complete selected Memory body. It never recursively reads endpoint bodies. Links expose their own identity, properties, metadata, and stored endpoints. Current reads return a revision; versioned reads return the selected version address and its historical state.

`bd list` retains its Issue default; explicit `--kind bead` selects generic Beads and `--kind link` selects Links. `--source` and `--target` are Link filters and fail for a Bead collection. `--resource-type` matches declared Type; `--conforms-to` explicitly requests conformance matching. Filters intersect. Resource identity determines ascending deterministic order; provider collation must implement the published canonical-ID ordering consistently.

**Proposed defaults:** one page of 50, maximum requested page size 500. `--all` consumes pages to completion; it is incompatible with `--after`. `--full` explicitly requests complete records, including Memory bodies and metadata. Default Memory summaries contain identity, Type, title, selected version address, and match information; never complete bodies or metadata values. Search excerpts may be bounded body fragments as §10 specifies, not an accidental full-record projection.

**Generic operations are sufficient.** Use generic reads, collections, Link operations and installed descriptors for ordinary CRUD and traversal. A CLI can fetch an authorized complete record and render a summary without emitting its body to the caller. Compact display does not by itself demand a separate Memory service or a new BDP summary operation. The earlier draft’s prohibition on downloading full records was stronger than this requirement and is withdrawn.

`remember`/`recall`/`forget`/`memories` add Chris’s authoring, body-output, search and compatibility policies over that same graph. Efficient server-side search/projection may later be useful, including to avoid downloading a whole corpus, but it is an explicit performance/capability decision—not a prerequisite invented for generic operations. Generic schema/ownership/History laws still have to be enforced by the authority; client-only rendering is not permission to bypass those laws. Measure bounded search cost before claiming economical remote Memory discovery.

Every paged result carries `next` and `complete`. Continuations bind to query, Scope, authorization view, and a stable read state. Changed/expired continuation state produces a typed refusal, never silent gaps, duplication, or restart at page one. An implementation may invalidate on change instead of retaining a snapshot. `--all` either completes that same traversal or fails visibly; it does not claim a consistent result by stitching unrelated pages.

### 5.3 Update properties and metadata

`bd update beads/…` and `bd update links/…` use the same update contract:

- `--properties JSON` explicitly replaces the whole properties object.
- `--patch JSON` is an ordered RFC 6902 patch relative to the properties object, limited to `add`, `replace`, and `remove`. `replace`/`remove` require an existing location; `null` is a value, not deletion. Array operations retain RFC semantics.
- `--metadata JSON` replaces the whole common metadata record; `--metadata-patch JSON` uses the same patch subset relative to metadata.
- Either or both records may change in one transaction. Omitted records and members outside a patch are preserved. Whole replacement and patch of the same record cannot be combined.
- An immutable member supplied through a Resource-shaped payload is rejected. A properties member coincidentally named `source` remains ordinary Type-defined data; it cannot change the endpoint.
- Equality is semantic JSON equality under the declared numeric admission contract, not input formatting. A no-op retains the Resource revision and creates no retained version or change event. A new attribution message alone is not a change.

Memory metadata is durable versioned state. Link metadata changes advance the Link revision, and the owning source’s version where ownership applies. An open metadata schema does not waive size, JSON validity, or supported numeric limits.

`bd edit` reads the selected mutable record and its guard before launching `$VISUAL`/`$EDITOR`; save submits that guard. For an owned Link it also captures the source guard. A conflict leaves the edited file available and reports its path; it does not overwrite the newer record or silently rebase the edit. Immutable identity/Type/endpoints are shown as context, not editable fields.

## 6. Link lifecycle, immutable endpoints, and ownership

### 6.1 Creation and direction

```sh
bd link beads/task beads/policy --resource-type related --id links/policy-citation
bd show links/policy-citation --json
bd update links/policy-citation \
  --metadata-patch '[{"op":"add","path":"/reason","value":"release policy"}]' \
  --if-revision '<revision returned by show>'
```

For a descriptor that admits a `weight` property, its update is:

```sh
bd update links/weighted-relation \
  --patch '[{"op":"replace","path":"/weight","value":0.8}]' \
  --if-revision '<observed Link revision>'
```

Examples use common metadata on `related` rather than pretending its descriptor accepts arbitrary properties. Property changes are always validated against the actual installed Link Type.

The source and target references are immutable **including the pin already defined by the public Reference contract**. Generic Link update accepts no `--source`, `--target`, `--target-version`, `--resource-type`, or legacy `--type` option. Issue classification updates remain the separate provisional contract in §2.1. To retarget or repin, create a distinct Link and explicitly remove the old one. Until an atomic replacement capability exists, these are two visible operations; the CLI must not claim atomic replacement. Its ordering and any multiplicity conflict must be reported, not hidden behind a convenience command.

Generic direction is always source → target. For a blocking Dependency this preserves existing `bd link A B` semantics: **A depends on B; B blocks A**. Help must say this explicitly.

### 6.2 Identity, duplicates, retries, and unlink

Generic Links with equal Type and endpoints may coexist. A new request with a new/generated Link ID creates another Link unless a declared policy prohibits it. Do not deduplicate by endpoint tuple. Reusing a request ID for the same operation is retry recovery, not another Link; see §8.

This sharpens Chris’s repeat-reference no-op rule: repeating the same identified Link/request is a no-op or replay, while deliberately creating a second Link is a distinct change. An endpoint tuple alone cannot distinguish those intentions. The proposal text needs that amendment.

`bd unlink links/ID` is canonical removal and acts immediately, like `dep remove`. It requires its guard, including a source guard when owned. `bd delete links/ID` is the same deletion with the ordinary preview/`--force` wrapper.

Pair form `bd unlink SOURCE TARGET --resource-type TYPE` is convenience only: canonicalize endpoints, include target version selection, find exactly one matching live Link, and guard that Link. Zero matches is `not_found`; multiple matches is `ambiguous_link` with visible candidate IDs and no deletion. There is no implicit “remove all.” Users select Link IDs individually; a later transactional facility may remove an explicitly enumerated set.

Old `dep add/remove/relate/unrelate` retain their domain-specific duplicate and pair semantics. They cannot silently consume unrelated generic multiedges. Migration must establish the Dependency projection and its identity mapping; §14 records the unresolved compatibility work.

### 6.3 Owned Links

For Memory, **every outgoing Link belongs to its durable historical state**, including custom informational Types. The public descriptor already expresses this with `ownsOutgoing: {"*": {"max": N}}`, where the declared maximum bounds the complete owned set (including explicitly owned Types). Use that existing wildcard contract; select and expose the Memory limit rather than inventing new ownership machinery. It is not limited to built-in `related`.

Creating, changing properties/metadata on, or deleting an owned Link atomically changes the Link and accepts a new source Memory version containing the complete outgoing Link set. The target is unchanged. A failed source or Link guard rolls back the whole operation. A no-op Link update changes neither.

Results expose the affected source identity, new source revision and version address as well as the Link result. Ownership cannot depend on which CLI verb was used. Historical source reads inline the owned Link state for that version; they must not join in today’s Link properties.

**Open:** Issue ownership. Jim currently snapshots every outgoing Dependency, while #5898 discusses narrower ownership. Proposed starting point is to preserve those recorded semantics and explicitly decide ownership of other outgoing Issue Links before freezing the Issue descriptor. Do not silently expand the History snapshot, or stop recording an existing Dependency mutation, because the generic representation changed.

## 7. Scope integrity, pins, and external references

Within the current Scope, an endpoint must name an existing live Bead satisfying the Link Type’s endpoint constraints. Link creation validates both local endpoints and aggregate/domain constraints within the same transaction as the write. Local aliases resolve to exact identities. A Link Resource cannot be substituted for a Bead endpoint.

Outside the Scope, an endpoint may be a fully qualified opaque URI. Acceptance does not depend on that service being available, and does not certify its existence, Type, authorization, or version. Represent it using the public Reference union: a URI or `{uri, revision}`. Local/external status follows the URI; there is no endpoint-Type sentinel in this public contract. Any application expectation about a remote Type is separately labeled unverified. Local endpoint Type validation reads the actual Bead; references do not carry a claimed declared Type.

**Proposed initial restriction:** new generic Links must have a local source. Outgoing cross-Scope targets are allowed; externally sourced Links are deferred by this CLI proposal even though public BDP can admit a Link with a local target. Both-external Links are already prohibited by the checked public BDP contract, not an unresolved BDP question.

`--target-version` pins the target selection, defaulting to latest when omitted. Creating a pin to a local retained version requires its successful resolution under the Memory/History validation layer. Invalid selector, never existed, gone with reason, permission failure, and unsupported History are distinct failures. No fallback to latest is allowed. A new Link to a deleted local Bead is refused under the current live-endpoint rule, even if a historical version survives; changing that rule requires a reviewed amendment.

An external pin may remain unverified; inspection labels that fact and exact recall can refuse. This intentionally amends #5877/#5898’s earlier unconditional unknown-target refusal at the Scope boundary. It does not weaken local validation.

Stored endpoints never change because a target changes or disappears. `links --resolve-targets` is explicit, bounded observation: return each stored reference alongside observed status/time or unresolved reason; do not mutate Links or source versions. Default inspection makes no network fetch. Authentication goes only to explicitly configured origins, not arbitrary Link targets or redirects.

Unresolved external blocking Dependencies remain blocking/unknown, never “satisfied.” **Open compatibility question:** define their representation without falsely claiming the external endpoint is a verified Issue. Existing external Dependencies must survive migration or cause a specific preflight refusal. Generic cross-Scope support alone does not prove this compatibility case.

## 8. Concurrency, atomicity, and uncertain outcomes

**Proposed new generic/Memory default:** updates and destructive lifecycle writes require an observed revision, or explicit `--unconditional`. Capturing a revision immediately before an otherwise unconditional update does not protect a user who edited stale content. `edit` carries the token from the original read. Legacy Issue mutations retain existing precondition semantics, including `update --if-status`, `--if-assignee`, and exit 13 on those failures.

Create enforces never-allocated identity. For owned Link creation, the source guard protects the observed outgoing set; there is no nonexistent-Link revision guard. Link update/unlink guard both the Link and owning source. `--unconditional` never bypasses Type, integrity, identity, ownership, or authority checks. `--force` is destructive-operation confirmation, not a concurrency bypass.

A logical Memory update accepts one version containing its combined field and owned-Link changes. Existing multi-step domain commands must define their logical acceptance boundary; Jim’s different direct/UOW version counts cannot become accidental CLI semantics. Explicit ordered operations in a batch may legitimately produce separate transitions, even if a later operation reverses an earlier one.

Two CLI processes from the authority workspace must get valid serial outcomes or typed conflicts. Process-local locking is insufficient. A cancelled client cannot release authority while an unaccounted database operation continues. These are acceptance requirements, not a mandate for a particular locking scheme.

`--request-id` is proposed durable retry identity, scoped to principal and authority. Same ID and same normalized operation returns its prior disposition without another mutation; changed content is `request_id_conflict`. A CLI-generated ID is exposed before submission in recovery diagnostics and included in the result. `bd request show TOKEN [--json]` retrieves disposition where that capability is supported. Do not equate caller-chosen Resource ID with transaction idempotency.

If transport fails after possible commit, return `outcome_unknown`, the request ID, and the exact status/retry command. Do not say “not applied,” and do not issue a new mutation automatically. A provider without durable request disposition advertises that limitation and refuses caller-required retry semantics; it does not simulate them with a volatile CLI cache. A partial implementation may expose a narrower capability, but complete advertised remote writes need an honest recovery contract.

## 9. Deletion, restoration, erasure: unresolved integrity policy

**This section is a review decision, not a settled ruling.** Donna leans toward B (local references prevent deletion); Chris’s current Memory scenarios preserve references to deleted targets; base BDP blocks *all* live incident Links. These are materially different contracts.

| Question | Proposed B-based draft behavior | Alternative / tradeoff |
|---|---|---|
| Local live incoming informational Links | Refuse deletion; list authorized blockers without exposing hidden relationships | Allow dangling-to-deleted identity, preserving Chris’s R20 experience |
| Local live outgoing Links | Base-compatible default refuses; explicit `--unlink-incident` may atomically clean up (§9.1) | Individual unlinking is possible, but never described as a race-free deletion algorithm |
| Incoming live Link pinned to retained history | Still blocks under the simple live-incident rule | Exempting pins improves retirement but changes base BDP integrity semantics |
| Links appearing only in immutable historical snapshots | Do not block current deletion; never rewrite snapshots | Blocking on all history can make deletion permanently impossible |
| Cross-Scope references pointing here | Cannot be enumerated or enforced as a local integrity guarantee | Their readers receive an observed deleted/gone/unresolved outcome |
| Issue’s existing delete-and-cleanup behavior | Preserve explicit Issue-domain cleanup atomically, subject to the new informational-Link policy | Optional generic cascade does not imply changing defaults or silently deleting unrelated informational Links |
| Erasure with live local Links | No implicit integrity bypass; report blockers even with `--force` | A separately reviewed explicit atomic erase-and-unlink workflow could remove them |

Until the policy is ratified, preview must report `deletion_policy_unresolved` for affected cases. An explicitly partial release may expose unreferenced Memory deletion while refusing the unresolved incident-Link cases; it must not claim complete #5877 support. Erasure must disclose that other Beads’ historical snapshots may contain references to the erased identity; erasing a target’s retained bodies does not silently rewrite another Bead’s history.

For the cases admitted by the chosen policy:

- `forget` immediately removes a current Memory, preserving retained versions; empty body is still a present Memory.
- `delete` previews by default; `--force` applies the same disappearance. Preview is read-only and must revalidate on apply. No body dump is needed to describe the deletion.
- Aliases bound to the deleted Resource are released. Canonical identity remains permanently reserved. Current reads say deleted; historical reads still resolve retained versions.
- `delete --erase --force` removes retained content for that identity and publishes the required version-removal facts. Authorized historical resolution returns gone with reason `erasure`; never another version. Failure cannot leave an advertised successful half-erasure.
- Ordinary deletion is not erasure, archival, or a clock-driven state. Retention removal is distinct from deliberate erasure. A citation grants no retention hold by itself.
- Restoring from retained state revives the same identity with a new revision/version, subject to current validation and endpoint integrity. Aliases are separate nonversioned state; restore does not steal reused aliases or fabricate old alias state from a content version.

### 9.1 Optional atomic deletion of incident Links — proposed decision

Donna’s race concern is correct: “unlink what I can see, then delete the Bead” is not atomic. Restrictive deletion can safely fail if another Link arrives, but an unbounded CLI retry loop is not an acceptable implementation of “delete this Bead and its incident Links.”

**Proposed syntax:** `bd delete beads/PATH --unlink-incident` previews; add `--force` and the usual Bead guard to apply. This flag removes all current in-Scope incident Links, incoming and outgoing, then the selected Bead in **one authority transaction**. It never recursively deletes neighboring Beads, removes unknown external references, erases retained versions, or rewrites historical snapshots. Default deletion still refuses incident Links. `--erase` is not implicitly included; combining the two needs a separately reviewed erasure contract and otherwise refuses.

Required atomic semantics if approved:

1. Select and validate the entire incident set inside the transaction. All selected Links and the Bead must be authorized and removable; otherwise nothing changes. Hidden blockers may be disclosed only as a non-disclosing refusal.
2. Preserve every affected surviving source’s ownership/revision/History effects. Incoming owned-Link removals can version **other** Beads; preview and result disclose these effects where authorized. Deleting a source does not manufacture a new live version for the deleted Resource.
3. Serialize against concurrent Link creation, update and removal. If a new Link is admitted first, it is included in the committed incident set; if deletion wins, creation must fail because its local endpoint no longer exists. Any backend conflict aborts the whole operation. There must be no committed intermediate state with only some Links removed.
4. The flag expressly authorizes **all incident Links at commit time**, not a fixed preview list. Preview says that the set can change. A user who needs exact-set approval requires a separate reviewed set-precondition contract; the ordinary Bead revision alone cannot fence changes to unowned/incoming Links. Do not pretend a preview provides that stronger guard.
5. Return the deleted Bead identity, all authorized deleted-Link identities/counts and surviving source revisions/versions. A timeout after possible commit follows the normal uncertain-outcome contract; no client-side compensating loop.

The public Transactional profile can express incident `deleteWhere` followed by `deleteBead` within one transaction. That is different from two Read+Update requests. Local CLI realization may use its existing authority transaction; a remote Read+Update-only route must refuse unless it advertises an atomic equivalent. This convenience does not require unrelated commands to advertise full BDP Transactional support. **Open:** whether to add this convenience to the public singleton delete operation, support it only where a transactional route exists, and when to schedule it. The flag is a proposal, not an assertion that base `DeleteBead` already cascades.

## 10. Memory CLI: Chris’s API on the generic model

```text
bd remember [BODY] [--body-file FILE | --stdin] [--title TITLE]
            [--id beads/PATH | --update MEMORY | --key LEGACY-KEY]
            [--metadata JSON] [--inception JSON]
            [--derived-from JSON] [--derivation-completeness complete|incomplete|unknown]
            [GUARD] [--request-id TOKEN] [--json]
bd recall (MEMORY | --key LEGACY-KEY) [--version VERSION | --at INSTANT] [--json]
bd memories [SEARCH] [--limit N] [--after CURSOR | --all]
            [--details] [--format table|records-json|legacy-json]
bd forget (MEMORY | --key LEGACY-KEY) GUARD [--json]
```

### 10.1 Authoring

`remember` always writes; one positional token is body text, never an implicit recall. Exactly one body source is permitted. Stdin/file input is explicit. New Memory needs a non-whitespace title or body. A nonempty body with no title gets a deterministic short title; it does not get an invented key. A title with empty body is valid. Body bytes are preserved as UTF-8 text; frontmatter is not parsed as Bead properties.

**Proposed title derivation:** first nonblank body line, Markdown heading marker stripped, Unicode whitespace collapsed, at most 80 Unicode code points with a visible ellipsis if shortened. The returned title is authoritative; clients must not depend on independently deriving it.

`--id` names a new immutable identity. `--update` updates an existing Memory under the guard; omitted fields remain unchanged. An explicitly empty body clears the body. There is no positional ambiguity between ID and content. All Memory authoring flags are conveniences over the installed Memory descriptor and common metadata, not another storage model.

Memory properties include title, body, Inception, `derived-from`, and `derivation-completeness`. The CLI flags map to the descriptor’s published field names; the JSON names in this draft retain Chris’s spellings. Inception is editable origin/transfer data. Change attribution belongs to the accepted version and is not editable Memory content.

`derived-from` is a structured list of references, not a distinguished Link Type. Absent means unknown, `[]` explicitly asserts direct authorship, and a populated list names derivation inputs. Completeness is tri-state and never defaults to `complete`. If stronger reference validation is required for this field, the Memory adapter must supply it explicitly: public BDP treats reference-shaped values inside `properties` as authored data, not graph References. Whether to require that stronger validation is open; it is not automatically supplied by generic BDP. Body URLs remain uninterpreted.

Memory always records History. Reject `--storage-class unversioned`, `--no-history`, and ephemeral Memory before mutation. Disabling optional Issue history cannot disable Memory history. No archive/unarchive command or scheduled archival mechanism is introduced.

### 10.2 Recall and discovery

`recall` emits the complete body on stdout and diagnostics on stderr. Empty body is success with zero content bytes; not-found is failure. `--json` returns identity, Type, selected version, attribution, Inception, derivation, metadata, owned references, and full body together. Historical recall cannot substitute current content.

`memories` searches current Memories with deterministic canonical-ID ordering, default limit 50, maximum page 500. **Proposed initial search:** literal Unicode case-folded substring matching across title and body, with no stemming, embedding ranking, or application-specific taxonomy. Matching field and excerpt provenance are explicit. Metadata search is a later opt-in capability, not implicit body exposure.

Default rows contain identity, title, selected version, and at most a 160-code-point match excerpt. `--details` adds attribution/Inception summary and reference counts without full bodies or metadata values. Exact selected version addresses accompany excerpts so a later read can select the state that matched. Continuation and `--all` obey §5.2; an empty successful result means a completed query with no matches.

`--format records-json` is the canonical summary envelope, including generated/unaliased identities. **Keep existing `bd memories --json` as the complete legacy name→body map** (`--format legacy-json` equivalent). It is explicitly body-carrying and mutually exclusive with records-json, paging, and `--details`; `--all` is redundant and may be accepted. A real memory named `schema_version` remains data, not a place to put an envelope version. Old consumers do not receive the new summary schema accidentally.

### 10.3 Legacy keys without a third identity system

Canonical identity plus aliases replaces the proposal’s separate mutable key field. `--key` remains a compatibility selector/upsert spelling, not a property on a canonical Memory.

**Proposed mapping:** conversion assigns an available valid memorable canonical path derived reversibly from the legacy key, otherwise a generated path with the original key bound as an alias. Preserve exact legacy-key lookup through its canonical path or an alias binding. The public alias/canonical-segment uniqueness rule must be respected; do not create an alias that collides with its own chosen canonical segment. Publish the encoding in the migration report; never normalize two distinct keys into one.

Key lookup resolves the exact migration-recorded canonical-path or alias mapping to the canonical Memory; it never fuzzy-matches a title. `remember --key K` updates an existing binding with its required guard; for an absent key it creates and binds a new Memory atomically. A key reused after deletion gets a new canonical identity if its former path is permanently reserved. An explicit creation request cannot reclaim the old identity. `recall`/`forget` prefer an explicit canonical selector; `--key K` is also accepted to force legacy-key lookup where a spelling could be confused with a path.

The legacy JSON map includes these exact compatibility key mappings, excludes Memories without such a mapping, and preserves empty bodies. Repointing an alias used by a mapping changes the map without versioning Memory content. The canonical records projection includes all Memories. Keyed Go/HTTP adapters use the same mapping and empty-versus-absent distinction. No dual authoritative key/value store remains after cutover.

`prime`, onboarding, initialization, and task loading teach literal `memories`, `recall`, and `remember` commands; they do not inject bodies. Guidance reports its budget assumptions and separately whether it was delivered completely, truncated, or delivered whole over budget. Memory data does not become startup instructions.

## 11. History surface and Jim’s versioning work

### 11.1 Commands and selectors

`bd versions RESOURCE` lists retained versions in acceptance order, newest first, with opaque version address, attribution, accepted time, and available local ordinal. It reports recording/participation and coverage, including when recording began; “no recorded history,” “history disabled,” “backend unsupported,” “not found,” and “history removed” are not interchangeable empty arrays.

`bd show RESOURCE --version V` and `bd recall MEMORY --version V` resolve exactly V. **Proposed selection format:** an explicit `--version` argument accepts the opaque version token returned for that Resource or its complete advertised version address. Do not invent `@v3`, concatenate a local ordinal into a URI, or parse provider token internals. Cross-resource version mismatch fails.

`--at INSTANT` selects state as of an RFC 3339 instant under the provider’s disclosed History semantics, excluding later acceptance. If the provider cannot distinguish states at its timestamp precision or cannot establish the requested boundary, it refuses `as_of_unavailable`; it does not silently pick a plausible version. Exact version selection is the reproducible alternative. As-of History is not proof of what an agent actually saw, especially when history arrived later by import/pull; consumers keep their own exact read receipts.

`bd compare RESOURCE --from V1 --to V2` returns changes in durable properties, metadata, and owned Links, plus selected identities/attribution. Same version is a successful empty difference. Either unavailable version fails the comparison as a whole. It does not compare arbitrary Dolt commits; existing `bd diff` remains that facility.

`bd restore RESOURCE --version V` previews restoring a retained state. `--apply` accepts a new attributed version under the current revision guard, or `--if-deleted TOKEN` for a deleted Resource. The token fences intervening restoration/deletion and is not reusable after a lifecycle transition. Restore is not rollback of the database, resurrection of erased versions, or reallocation of identity.

An authorized `bd show RESOURCE --json` of a deleted identity returns the typed `deleted` failure with a `deletionToken` in its error details; it does not pretend a live record was found. Restore preview can also return this token. An unauthorized caller is not entitled to lifecycle disclosure. A restore result explicitly reports that aliases were not restored and identifies any requested compatibility-name conflicts; content restoration does not fail merely because a former alias now belongs elsewhere.

Restoring a historical owned-Link set is atomic and subject to present integrity. **Open mechanism:** deleted Link IDs cannot be reallocated, and endpoint changes cannot be updates. The History profile must define whether exact Link identities can be restored as the same Resources or materialized through new Link identities with explicit mapping. Until settled, restore of an affected owned set refuses before any mutation; restoring only the body cannot be passed off as full restoration.

`bd changes` starts from an opaque checkpoint, an initial instant, or explicit beginning. Results include accepted versions, disappearances, restorations, and retained-version removals with reasons, plus the next checkpoint and completeness. Filtering Memory changes cannot skip checkpoint advancement across other event kinds. Checkpoints are Scope/view-bound, do not derive from wall time or local version ordinals, and fail explicitly when expired or truncated. A late-arriving version is observable as arrival, not silently missed because its original timestamp predates the reader’s checkpoint.

### 11.2 Compatibility with existing versioning

Preserve Jim’s `bd versions ISSUE-ID [--json]` projection and its `local_revision` field. Canonical Resource selectors use the generic envelope; a store-local ordinal is not a portable version address or a concurrency token.

`versioned-history.enabled` and `BD_VERSIONED_HISTORY_ENABLED` control optional Issue recording. Recorded versions remain readable when recording is off. Graph current revisions and mandatory Memory History are independent of that switch. Report when recording began and any missing coverage; do not fabricate earlier versions.

No command may claim complete Memory History while exact addressing, comparison, restore or change/removal observations are unavailable. Content hashes, reader witnesses, globally historical Scope snapshots, branches and complete BDP Transactional/Replication are not prerequisites for this narrower contract.

## 12. Traversal, Types, aliases, and data movement

### 12.1 Links and graph views

`bd links BEAD` returns incident Link records, default direction both; `in` and `out` are relative to the selected Bead. It does not overload legacy `show --refs`. Pagination uses the same stable-state contract as collections.

With `--version V`, outgoing owned Links are those in V. A request for historical incoming completeness requires an actual historical incident capability; otherwise it refuses. It must not relabel today’s incoming Links as historical. Default output contains endpoint references, not endpoint bodies or fetched target state.

`bd graph --view generic` renders a bounded generic traversal. Proposed defaults: direction both, depth 1, maximum 100 nodes/200 Links. Cycles terminate by identity; parallel Links retain distinct identities. Results expose frontier and completeness when bounded. Memory nodes are summaries; external nodes are opaque reference stubs. Ordinary `bd graph` remains the Issue Dependency view. There is no implicit remote traversal.

### 12.2 Types and aliases

`bd types get` returns exact descriptor identity, kind, validation properties, endpoint restrictions, conformance, declared ownership and applicable History policy. `bd types install FILE` validates and installs the complete pinned closure atomically; an incomplete closure fails without partial installation. Reinstalling identical content is a no-op. Different contract bytes under an already installed Type ID conflict; evolving a Type requires another Type ID. No declared Resource-Type mutation or destructive Type uninstall is introduced here; the Issue classification question remains open in §2.1.

`alias set --if-absent` creates a binding; `--if-target OLD` repoints it conditionally. Target must resolve to a live in-Scope canonical Bead under the checked public alias contract; Link aliases would be a separate extension and are not introduced here. Aliases cannot chain. `alias remove` guards the old target. Aliases do not advance content revisions or History versions, though authority/audit state may record their changes. Canonical data and pins must remain unchanged when aliases are reused.

### 12.3 Export, import, and backup

Proposed explicit generic controls:

```text
bd export --format graph-json [--include-memory] [--full] [--output FILE]
bd import --format graph-json --file FILE [--dry-run]
```

`bd export` (legacy JSONL or explicit generic format) excludes Memory by default. `--include-memory` permits Memory summaries; `--include-memory --full` permits complete Memory records. Full records are required for a restorable interchange unit; summaries are discovery artifacts and import must refuse them as content. Do not emit endpoint Links to excluded local content while claiming a complete connected export: expose omissions or reject a requested connected export.

Automatic export creates or refreshes no Memory records, including summaries. Preserve manually exported Memory entries already in a file without treating them as permission to refresh them. Stored configuration cannot silently enable body-carrying export. Whole-store backup includes all Memory content and retained History because otherwise it is not a whole-store backup.

Import of a selected connected unit preserves identities and reference meaning under its declared import mode, or rejects it atomically. Never merge by similar title, key, alias, or matching path in another Scope. Identity-preserving restore/replication and copying content into a new Scope are different operations; cross-Scope copy needs an explicit mapping, not an implicit `import` guess. Full connected interchange remains capability-gated until its shared contract is implemented; do not claim it from JSONL parsing alone.

## 13. Output, errors, limits, and capability reporting

### 13.1 Output profiles

Legacy Issue commands keep their documented JSON/text contracts. New generic commands use a versioned CLI envelope, distinct from raw BDP wire DTOs. Proposed JSON shape:

```json
{
  "schemaVersion": 1,
  "scope": "https://example.invalid/demo",
  "result": {
    "kind": "link",
    "id": "https://example.invalid/demo/links/policy-citation",
    "type": "<installed related Type URL>",
    "revision": "<opaque revision>",
    "source": "https://example.invalid/demo/beads/task",
    "target": "https://example.invalid/demo/beads/policy",
    "properties": {},
    "metadata": {"reason": "release policy"}
  },
  "changed": true,
  "requestId": "<request ID>"
}
```

This illustrates the proposed CLI envelope and common-metadata amendment, **not an already-valid BDP response**. Collections use `items`, `next`, and `complete`. Summaries identify `projection: "summary"`; full records identify `projection: "full"`. Do not manufacture `revision` from the version ordinal. Historical records add `version` and `attribution`; owned-Link mutation results add `sourceChange: {id, revision, version}`. A deletion result identifies the Resource and its final live revision, with `deleted: true`; it does not invent a new live Resource revision.

Human mutation output includes the canonical path and current revision; History-aware writes also print the version address. `--quiet` suppresses prose but never errors; `--silent` on creation prints only the canonical ID, following existing capture ergonomics. JSON stdout contains exactly one document and no progress text. Recall’s plain stdout is body-only. Diagnostics go to stderr. New structured errors are one stderr JSON document with `code`, `message`, `details`, and `retryable`; stdout is empty on failure. Existing legacy error shapes are preserved.

### 13.2 Error contract

Proposed new generic exits: 0 success (including a clearly labeled preview/no-op), 2 usage/invalid selector or payload, 3 not found or gone, 4 conflict/precondition/constraint refusal, 5 unavailable/unsupported/authority refusal, 6 uncertain mutation outcome. Existing Issue exits, especially conditional-update exit 13, are unchanged. A successful bounded page is not an error; its `complete: false` is explicit. A failed all-results request is nonzero and cannot mark its partial results complete.

Stable codes include:

| Family | Examples / required distinction |
|---|---|
| Input/type | `invalid_selector`, `ambiguous_selector`, `wrong_kind`, `invalid_properties`, `unknown_type`, `immutable_member` |
| Identity/lifecycle | `not_found`, `deleted`, `gone` with disclosed reason, `identity_reserved` |
| Concurrency | `revision_conflict`, `source_revision_conflict`, `request_id_conflict` |
| Integrity | `incident_links_exist`, `invalid_endpoint`, `constraint_violation`, `ambiguous_link` |
| History | `history_disabled`, `history_unsupported`, `history_unretained`, `as_of_unavailable` |
| Availability | `not_authority`, `graph_not_initialized`, `capability_unavailable`, `route_unavailable`, `permission_denied` |
| Traversal | `continuation_invalid`, `continuation_expired`, `checkpoint_truncated`, `limit_exceeded` |
| Uncertainty | `outcome_unknown`; never represented as confirmed rollback |

Non-disclosing authorization failures may withhold existence or blocking IDs; do not leak hidden graph state to make diagnostics more detailed. Errors identify a supported next action, not an opaque internal stack trace.

### 13.3 Limits and capability discovery

`bd status --graph --json` reports selected backend/route, Scope identity, authority state, schema/protocol generation, installed built-ins, and operation support. Proposed capability names cover `genericRead`, `genericWrite`, `memory`, `historyExact`, `historyAsOf`, `historyCompare`, `historyRestore`, `historyChanges`, `ownedLinks`, `commonMetadata`, `requestStatus`, `customTypes`, `connectedInterchange`, and `backupContinuity`. A missing capability is not an empty result.

Status also reports advertised limits and their units: property/metadata/body bytes, patch count/depth, page size, traversal bounds, cursor lifetime, and response limits. Exact storage-size defaults need measured backend admission work; do not choose an arbitrary low cap that breaks existing Issue payloads. An oversized mutation fails before effects. Complete recall is complete or a typed size refusal, never a silently shortened body. If a surface intentionally emits an excerpt or bounded traversal, it labels the bound and completeness.

## 14. Changes to the existing CLI surface

This matrix is the compatibility contract by command family. Detailed flags remain as documented in the existing [CLI reference](https://github.com/gastownhall/beads/blob/16871ee15/docs/CLI_REFERENCE.md), except the explicit changes below. Every actual leaf command and direct writer needs an implementation inventory before complete A; this table does not falsely claim that census has already passed.

| Existing surface | Required target behavior / change |
|---|---|
| `create`, `q`, `create-form`, `create --file`, `create --graph` | Preserve Issue defaults, batch meanings, familiar IDs and outputs; store canonical Issues. Generic `create` is selected with `--resource-type`. No hijacking `--file` for generic JSON |
| `update`, `edit`, `note`, `priority`, `assign` | Preserve Issue fields and conveniences; shared transaction/domain validation. Existing `--type` provisionally remains mutable Issue subtype, pending §2.1; existing metadata replacement and set/unset flags keep their semantics |
| `close`, `reopen`, `ready`, `blocked`, `defer`, `claim` and equivalent flags | Issue-only semantics; wrong-kind Memory rejection. Generic status/dependency changes produce the same readiness effects. No hidden maintenance writes in generic reads |
| `show`, `list`, `search`, `query`, `count`, `status`, `statuses`, `types`, `stale`, `lint` | Legacy Issue projections remain default; explicit generic options/selectors add declared Types/Resources. Explicit selected Memory can be shown. Issue workflow filters never select Memory |
| `link` | Default Issue form remains Dependency shorthand, default `blocks`, direction unchanged. The new `--resource-type` form requires an installed Type and returns Link identity |
| Chris’s `bd link … --type related` with Memory endpoint | Add a compatibility bridge to registered generic `related`; do not reinterpret default blocking as informational. Both-Issue `--type related` retains the legacy Dependency meaning. `--resource-type related` always selects the generic installed Type; help displays the resolved Type |
| `unlink` (new) | Generic identity-first deletion with guarded pair convenience; no accidental multi-Link deletion |
| `dep add/remove/list/tree/cycles/relate/unrelate`, `children`, `epic`, `duplicate`, `duplicates`, `supersede`, `graph` | Preserve Issue relationship policy, direction, duplicate/reciprocal behavior and output; use canonical Links. Generic multiedges outside the Dependency projection must not be mistaken for legacy Dependency rows |
| `comments`, `comment`, `label`, `tag`, `state`, `set-state`, `todo`, gates, merge slots, swarms, molecules/formulas and work generators | Preserve documented Issue workflows and logical transactions. Do not generalize them to arbitrary Types without a separate domain contract |
| `remember`, `recall`, `memories`, `forget`, `prime`, onboarding/setup guidance | Apply §10; canonical Memory storage, explicit capture/read, no body injection, legacy JSON exception, empty-body correctness |
| `history`, `show --as-of`, `diff`, `restore`, `provenance` | Preserve existing meanings: Issue/Dolt events, Dolt state, compaction recovery, external-artifact provenance. New generic History is §11; Inception is not `provenance` |
| Jim’s `versions` and history config | Preserve Issue surface/local ordinal; add generic projection; optional Issue recording never gates graph revisions or Memory History |
| `batch`, bulk Issue mutation, graph-plan creation, import | Preserve advertised transaction boundary; same policy as single writes. No blanket partial success. Generic Transactional API is not required merely because Issue batch already exists |
| `--ephemeral`, wisps, `--no-history`, `--storage-class`, promotion/demotion | Preserve Issue persistence semantics, labels and dependencies. Identity cannot be reused on disappearance. Reject invalid Memory combinations; do not substitute Memory for a wisp |
| `export`, auto-export/hooks, `import`, `backup`, bootstrap recovery | Apply §12 and §15; preserve complete graph/History/identity obligations or refuse before effects |
| Dolt `pull/push`, `branch`, `vc`, federation, multi-repo sync/worktrees | Respect one authority; replicas submit writes through supported routing. Independent offline graph minting/merge is not silently supported. Unsupported routes refuse with an adoption preflight entry |
| `rename-prefix` and repository moves | Cannot rename canonical graph identities. Proposed compatibility strategy changes Issue display/locator mappings with explicit disclosure; if exact old semantics require identity changes, resolve this before claiming the complete proposed contract |
| `compact`, `flatten`, `gc`, `prune`, `purge`, admin cleanup/reset | Preserve promised versions, nonreuse and removal observations, or refuse with a named reason. Replay of current Issue tables is not sufficient to reconstruct retained History/Memory |
| `sql` | SQL reads remain available under documented backend access. Arbitrary direct SQL writes are outside graph compatibility; do not advertise them as a supported mutation path |
| Integrations (Jira/Linear/etc.), hooks, background/proxied writers | Same mutation fence and authority as CLI. Old writers refuse after canonical cutover, even via a schema-skew override |
| Help, completions, config, doctor, context/info/where, upgrade, utilities | Update help/completion and diagnosis for selected workspace format/Scope; otherwise retain existing role. Every utility capable of changing data must be classified as a writer, not exempted by its name |

**Open compatibility decisions:** the explicit mixed-endpoint `link --type related` bridge; unique legacy Dependency projections versus generic multiedges; Issue ID display/rename mapping; and which old multi-clone workflows can submit to the authority. These need concrete migration examples and leaf-command evidence. Unsupported previews are permitted; unresolved documented Issue workflows prevent full compatibility.

## 15. Initialization, deployment, routing, and recovery

### 15.1 Real initialization

**Proposed syntax:** `bd init --graph-mode link` opts a fresh workspace into graph storage and persists that selection; subsequent commands need no repeated mode flag. Add a canonical Scope URL input. Do not reuse existing Issue-plan `create --graph` as a workspace switch.

```sh
bd init --graph-mode link --scope-url https://example.invalid/demo
bd create --resource-type memory --id beads/plan \
  --properties '{"title":"Plan","body":"First durable graph record."}'
bd show beads/plan
```

The example URL is a placeholder for an operator-owned canonical Scope URL; local CLI use does not require starting HTTP, but the URL is permanent identity and must be chosen with future serving in mind. Do not silently mint a disposable localhost identity that later has to move. A sensible default Scope-identity scheme is an open deployability decision; explicit `--scope-url` makes the initial contract executable without inventing one.

Successful init provisions the schema, authority/identity state, complete built-in descriptor catalog and selected backend settings atomically or as an explicitly recoverable setup operation. It must not report success while `show` still depends on a separate hidden mint/bootstrap step. Init does not create `beads/plan`; the documented create command does.

For an ordinary released shared Dolt server, retain actual existing flags:

```sh
bd init --graph-mode link --server --external --server-host 127.0.0.1 --server-port 3307 \
  --scope-url https://example.invalid/shared-demo
```

Existing engine selection and credential rules remain; `--server` here means Dolt SQL storage, not a BDP URL. macOS and Linux qualify separately, both embedded and shared server. Missing existing-workspace metadata or a mismatched database identity must fail before migration or fabrication of a new empty database. Fresh initialization deliberately creating its requested database is a different case.

### 15.2 Remote routing and administration

Retain the drafted route surface, while avoiding the `init --server` boolean collision:

```text
bd init --graph-mode link --bdp-server URL
bd client local
bd client server --server URL [--insecure-http]
bd serve --addr ADDRESS [--scope-url URL] [--auth-token-file FILE]
bd scope status
bd scope recover [--ledger FILE] [--rotate-url URL]
bd scope promote [--rotate-url URL] [--steal]
bd scope ledger snapshot --output FILE
bd scope ledger apply FILE
```

`scope …` is a **proposed move** of drafted graph administration away from root `restore`/`promote`, which already have Issue meanings and now also need History restore. These are operator commands, not first-demo prerequisites. Recovery/lease mechanisms remain implementation choices constrained by public identity guarantees.

Shared-server authority can serve BDP; embedded graph support means local commands and permitted client use, not an embedded HTTP server. A remote client reports unsupported capabilities truthfully. Route failure never falls back to mutating a local stale replica. Credentials come from the configured secure channel/token file, not printed command results. Existing non-loopback binding/authentication safeguards in the serving design still apply.

### 15.3 Migration and backup

Proposed `bd migrate graph --dry-run` reports IDs, Type mapping, unsupported Dependency kinds, external references, payload admission, legacy Memory conversion, old-writer fences and recovery requirements. `--apply` performs a verified cutover; retain the old database for rollback before any graph writes. After new writes, describe forward recovery rather than promising a lossless downgrade.

A restorable backup bundle includes graph content, Memory History, pinned descriptors, identity allocation/nonreuse state and whatever external continuity evidence is required. Recovery on a fresh machine must either prove continuity of the same Scope or explicitly require a new Scope/refuse. An old/incomplete backup cannot silently reopen allocation with reusable names. Recovery has no permission to remint a version address for different bytes.

## 16. User scenarios and observable acceptance criteria

### 16.1 Four-Bead CLI demonstration

The following shell sketch names four Beads and two informational Links. The two Issue commands retain familiar domain semantics. Revision placeholders mean values actually returned by preceding reads; the final executable demo must capture them, not hardcode invented tokens.

```sh
# Fresh opted-in workspace, using either supported backend (§15).
bd init --graph-mode link --prefix demo --scope-url https://example.invalid/demo

bd create "Release package" --id demo-release --type task
bd create "Complete prerequisite" --id demo-prereq --type task

bd remember "Release only after the prerequisite is complete." \
  --id beads/plan --title "Release policy"
bd remember "This protects compatibility for existing users." \
  --id beads/rationale --title "Policy rationale"

bd link beads/demo-release beads/plan \
  --resource-type related --id links/uses-policy
bd show beads/plan --json
bd link beads/plan beads/rationale \
  --resource-type related --id links/explains-policy \
  --if-source-revision '<plan revision>'

bd show links/explains-policy --json
bd show beads/plan --json
bd update links/explains-policy \
  --metadata-patch '[{"op":"add","path":"/confidence","value":"reviewed"}]' \
  --if-revision '<Link revision>' --if-source-revision '<new plan revision>'

# Existing issue direction: prerequisite blocks release.
bd link demo-release demo-prereq
bd ready
bd close demo-prereq
bd ready

bd show beads/plan
bd links beads/plan
bd memories --format records-json
bd recall beads/plan
```

Also demonstrate the reverse adapter direction: create a blocking Dependency through `bd link … --resource-type dependency/blocks`, observe it with `bd dep list` and `ready`, change an Issue through `bd update`, and observe its domain state through `bd show`. The sketch initializes the Issue prefix `demo`, matching its explicit Issue IDs.

Exit the process, reopen the workspace, and run `bd show beads/plan` again. The same identities, data and revisions must remain visible after reopening.

### 16.2 Acceptance matrix

| Area | Observable acceptance |
|---|---|
| Init and compatibility | Persisted graph format survives reopen; ordinary `bd show beads/plan` works without a flag; mode mismatch refuses; existing Issue `--type` meanings do not silently change |
| Nominal Issue Types | Ratified §2.1 decision explains task→bug conversion, identity, endpoint/ownership and History effects; no property workaround advertised as nominal typing |
| Mixed graph | Two Issues and two Memories; Issue→Memory and Memory→Memory Links inspectable by identity; no second authoritative representation |
| Link properties | Update a descriptor-admitted property and common metadata; same Link ID/Type/endpoints; new revision; immutable endpoint/Type attempts rejected |
| Owned state | Link create/update/unlink changes source Memory once per logical operation, never target; historical source retains exact previous outgoing set |
| Dependencies | Generic and legacy creation/update produce matching Issue/Dependency behavior and readiness; Memory endpoints rejected for workflow dependencies |
| Multi-links | Two intentional equal-endpoint Links coexist; pair unlink refuses ambiguity; ID unlink removes only one; retry ID does not create a duplicate |
| Concurrency | Two independent CLI processes produce serial success or conflict; stale editor and stale owned-source guard cannot lose another write |
| Atomic failures | Invalid properties/endpoint/guard and cancellation leave no partial content, History or dependency effects; uncertain outcome is distinguished from rollback |
| Scope boundary | Invalid local target rejected atomically; unreachable external reference admitted as unverified; unresolved external blocker is not ready |
| Memory retrieval | Positional `remember` always writes; empty body remains readable/deletable; summaries and linked task views do not inject full bodies; explicit recall is complete |
| Naming | Alias rebind does not change versions or existing stored references; deleted canonical name cannot be reused; legacy keys round-trip without collisions |
| History | Stable exact addresses, no-op behavior, comparison, guarded restore, change/removal checkpoints and typed gone; Jim’s local ordinal never masquerades as a portable ID |
| Atomic incident cleanup | Concurrent link creation yields serialized include-or-refuse behavior; no partial unlink/delete or lost owned-source effects; Read+Update-only remote route cannot emulate atomicity with a loop |
| Deletion | Chosen §9 policy tested for live/historical, latest/pinned, local/external, owned outgoing and Issue cleanup; open policy is not called a pass |
| Continuations | Stable complete traversal or explicit invalidation under concurrent changes; no silent omitted/duplicate records; bounded traversal labels its frontier |
| Exposure | Default/manual/automatic export postures, explicit full records, legacy memories JSON exception and whole-store backup verified independently |
| Deployability | Existing workspace preflight; safe conversion/cutover and old-client refusal; backup recovery on a fresh machine; independent uncoached reproduction before claiming deployable adoption |
| Complete contract | Every documented Issue/Dependency workflow accounted for; Memory-required History and custom Types complete; missing capability is visible and prevents a full completion claim |


## 17. Decisions to review in this draft

These are still decisions, not approvals inferred from the comments:

1. **Nominal Issue Types and reclassification (§2.1):** property-only compatibility, immutable nominal Task/Bug Types with a compatibility break, or a deliberate stable-identity Type-change operation. This is an overarching model gap, not an implementation detail.
2. **Common metadata placement (§2.2):** recommended top-level Resource member versus reserved `properties.metadata`; whether any universal root Type is warranted. Recommendation: no root Types solely for metadata.
3. **Init and grammar (§3):** persist graph storage once; no per-command storage switch; preserve existing `--type` and use new `--resource-type` for generic Types. This replaces the older mode-dependent CLI proposal and needs explicit review.
4. **Deletion (§9):** restrictive default, local latest/pinned behavior, historical-reference boundary, and whether to adopt atomic `--unlink-incident`. Decide remote capability/transaction mapping and effects on other owning sources.
5. **Guard ergonomics:** required guards/explicit unconditional writes for new generic and Memory mutation, including owned-source guards and existing-key `remember`; these change some Memory command behavior and are labeled in §4.1.
6. **Integration, not missing BDP design:** implement the existing public pin/ownership/History contracts, choose Memory ownership limits, and map metadata updates and additional CLI guards. Resolve owned-set restoration and Issue ownership against Jim’s actual writer. Do not re-open wildcard ownership or invent a mandatory remote-summary protocol.
7. **Remaining compatibility/defaults:** mixed-endpoint `related` bridge, exact legacy-key mapping under alias namespace constraints, Dependency multiedges, immutable identity versus prefix rename/moves, output/error formats and bounded search defaults.

