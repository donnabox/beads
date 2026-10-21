# BDP simplicity review catalog

Status: proposed decisions, not protocol changes. Owner: Vickie; product decisions: Donna. This catalog supports the sole [design plan](design.md).

Donna's direction: do not introduce mechanisms that are not needed to make the product work. Existing implementation, specification length, or prior design effort is not evidence of need. We will review one decision at a time, retaining a mechanism when its concrete value earns its implementation and client cost.

Source: [present-contract specification at 5d7694e8](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md), the published PR67 head inspected for this audit. These are current normative requirements or permitted surfaces unless identified otherwise. Recommendations below are proposals. No deletion, profile reshaping, or wire change has been approved by this catalog.

## How we will decide

For each item: name the user action it enables, identify the consumer that needs it, compare the smallest alternative, and state the lost guarantee. A useful distinction deserves a wire representation only if a consumer can act on it. A provider's storage problem need not become a client-visible state machine.

Known requirements from this workstream are a usable cold client, efficient generic Issue operations, Bead and Link CRUD, useful batch composition, domain-neutral Events, and evaluating journal/replica support. Exact revision reads have explicit operator support. Attribution and useful journal metadata also have an identified motivation. More granular History diagnoses, arbitrary history rewrites, and their discovery/navigation machinery have not been tied to an actionable consumer requirement in this review. That is an evidence gap, not proof that no consumer could ever want them.

Do not remove a correctness rule while keeping the feature that depends on it. Retry ownership/fencing prevents duplicate effects; authorization prevents disclosure; a replica needs a consistent bootstrap and a way to detect gaps. The simplification may instead be to narrow the supported feature or move it out of the minimum profile. We will not label an endpoint that silently loses these guarantees conforming.

**Default positions:** simplify or defer speculative surfaces; retain demonstrated functionality; ask for evidence where the tradeoff is real. Every item remains **undecided**. Recommendations are starting positions for discussion, not accepted rulings.

## Review order

Start with H01–H06: what historical reading promises. Then T01–T02: whether atomic mutation and replication must be one product capability, and whether two observation surfaces are justified. Then U01–U05: the combinations responsible for much of the write state machine. Remaining items are smaller cuts or deliberate choices to retain.

| Group | Items | Subject |
| --- | --- | --- |
| History | H01–H10 | Exact reads versus archival explanation and repair machinery |
| Read surface | R01–R07 | Identity syntax, discovery, views, values and metadata |
| Mutation | U01–U08 | Aliases, sequences, retries, patches and admission |
| Transactional | T01–T11 | Atomicity, observation, replication and erasure |

## History

### H01 — One all-or-nothing History capability

**Current:** `historicalResolution: {"version":1}` requires exact revision reads plus the whole applicable History contract, for both Beads and Links. An implementation cannot advertise only exact revision reads. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#historical-resolution)

**Value/cost:** exact revision lookup is explicitly wanted; coupling it to a historical inventory, lineage metadata and diagnostics increases the price of supporting it.

**Recommendation:** make the minimum promise exact, complete revision retrieval when available, including the historical owned state covered by that revision. Never rebind a token or substitute current state. Authorization, applicable erasure and transient service failures still apply. Decide whether to keep a versions listing separately. Do not immediately invent a matrix of new capability flags. **Decision:** what must a provider implement before it can say “I can retrieve an old revision”?

### H02 — Inventories of missing historical pieces

**Current:** `historyMissing` names missing records, property JSON Pointers, or owned-Link sets, with a 64-item bound and a completeness flag; an empty incomplete inventory is valid. It never returns partial Resource content. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#historical-diagnoses-and-missing-state)

**Value/cost:** potentially helps an archival repair tool; requires providers to track and classify internal reconstruction failures. No such repair consumer is established here.

**Recommendation:** remove the standard missing-piece inventory; return a stable revision-unavailable outcome when a complete record cannot be served. Providers can diagnose storage internally. **Lost:** portable component-level repair hints. **Decision:** does any product action need to know *which property or owned set* is absent?

### H03 — Five missing-revision diagnoses

**Current:** `revision-unknown`, `revision-unretained`, `revision-reorganized`, `revision-not-tracked`, and `revision-unrepresentable`, each with evidence and precedence rules. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#historical-diagnoses-and-missing-state)

**Value/cost:** detailed explanations for archival tools, but more durable evidence and more states for every reader. Several outcomes may produce the same user action: the requested revision cannot be read.

**Recommendation:** collapse unavailable-history distinctions unless a client action requires one. Preserve separate transient service failure and non-disclosure; decide deliberate erasure separately. **Lost:** precise public diagnosis of tracking, reconstruction and representability. **Decision:** which cases actually change what the client does?

### H04 — Replaced lineages and version navigation graphs

**Current:** rows and `BDP-History-Lineage` distinguish `current` from `replaced`; HTTP links expose direct predecessors/successors, potentially multiple, and latest/version-history relations. Restore branches must not be mistaken for linear succession. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#historical-http-metadata)

**Value/cost:** enables browsing rewritten history; requires retained ancestry and separate authorization for navigational relationships.

**Recommendation:** defer lineage and ancestry navigation. Keep a revision bound to its original value, and serve it if retained; do not require the public protocol to explain the store's history graph. **Lost:** portable branch/restore-history browsing. **Decision:** do we have that product requirement?

### H05 — Rich retained-versions inventory

**Current:** `view=versions` reports `population`, `participation`, lineage, body completeness, and a window with oldest/newest bounds and completeness, over one authorized stable enumeration. Replaced and incomplete retained versions participate. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#retained-versions-pages)

**Value/cost:** a versions list is useful UI functionality; claims about an entire retained population, tracking status and incomplete versions are a much larger promise.

**Recommendation:** if listing is needed, start with revision, available metadata, and continuation; define what is enumerated without population/participation/window assertions. Do not imply missing rows never existed. **Lost:** portable statements about archival completeness. **Decision:** do users need a versions list, and what will they do with completeness metadata?

### H06 — Permanent explanations for removed historical addresses

**Current:** removing a cited version must preserve sufficient evidence for Gone-with-reason, without an expiry exception. This goes beyond preventing Resource ID reuse. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#historical-diagnoses-and-missing-state)

**Value/cost:** durable explanations for broken citations; durable per-version disposition tracking potentially outlives the content forever.

**Recommendation:** allow an unavailable result after ordinary retention ends unless a demonstrated product requires a permanent reason. Retain non-rebinding and separately decide erasure obligations. **Lost:** proof of why every old citation stopped resolving. **Decision:** must a provider remember the reason indefinitely, or only avoid returning the wrong value?

### H07 — Tri-state change-context envelopes

**Current:** `committedAt`, `agent`, and `message` use tagged `present`, `absent` or `undetermined` states, with field-specific rules. History requires context on native versions; legacy absence remains permitted. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#immutable-change-context)

**Value/cost:** timestamp/agent/message are useful journal metadata already discussed. Distinguishing intentionally absent from unknowable imports multiplies shapes throughout reads, rows, Events and receipts.

**Recommendation:** retain useful metadata; consider ordinary optional fields, with omission meaning “not supplied by this response.” Keep three states only if a concrete UI or importer needs that distinction. **Lost:** machine-readable proof of intentional absence. **Decision:** does any consumer act differently on absent versus unknown?

### H08 — Several independent historical disclosure permissions

**Current:** subject-history diagnostics, complete historical records, whole metadata rows, and historical target identity/relationship visibility have distinct gates. These are obligations, not necessarily four API permissions. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#historical-authorization), [exact reads](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#exact-historical-reads)

**Value/cost:** supports historical access despite current deletion or target invisibility; makes the authorization model harder to implement and explain.

**Recommendation:** first simplify the surfaces in H02–H05, then specify a minimum whole-record historical access rule. Do not remove authorization checks while continuing to expose their protected data. **Lost:** some independent metadata/relationship disclosure options. **Decision:** is partial historical discoverability a product requirement?

### H09 — Import/recovery assurance as part of History

**Current:** before publishing imported retained copies, establish origin/version identity and authoritative erasure status; unestablishable status requires rejection or discard. No HTTP import endpoint is defined. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#history-recovery-imports-and-assurance)

**Value/cost:** prevents resurrection of erased content; asks a history reader/provider to implement evidence acquisition for imported archives.

**Recommendation:** separate ordinary historical reads from a portable archival import/erasure-assurance product promise. Keep truthful immutable reads and whatever erasure duty the chosen product actually claims. **Lost:** universal conformance assurance for imported retained copies. **Decision:** must this release support importing arbitrary retained history, or only serving its own?

### H10 — Compatibility projection for versions that predate metadata

**Current:** historical records lacking `metadata` acquire `{}` in their served projection without a new revision; their already committed erasure digests still use the original pre-projection bytes. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#resource-records), [erasure](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#version-erasure)

**Value/cost:** compatibility for old records/digests; two canonical representations to reason about. No externally deployed BDP population has been established, though local data may still need migration.

**Recommendation:** investigate actual stored data before keeping a permanent protocol compatibility exception. Prefer one current record contract and an explicit local migration/import boundary if feasible. **Lost:** transparent reads of unchanged legacy artifacts. **Decision:** what real data requires this exception?

## Read surface

### R01 — Hierarchical opaque Resource IDs

**Current:** IDs allow arbitrary safe multi-segment paths; normalization, encoded separator rejection and path/control collisions are specified. Views use queries partly to avoid those collisions. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#scope-discovery-and-human-documentation)

**Recommendation:** ask whether one opaque ID segment covers real stores; simplify grammar if it does. **Value/cost:** human path organization versus parsing and routing complexity. **Lost:** existing multi-segment identity spellings; migration must be explicit. **Decision:** is hierarchical identity needed, rather than labels/properties?

### R02 — Discovery through a Link header and another document

**Current:** GET Scope yields a `service-desc` Link; the client follows it and never treats the Scope body as discovery. Most collection URLs are nevertheless fixed derivations. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#scope-discovery-and-human-documentation)

**Recommendation:** compare this with returning JSON discovery directly at the Scope URL. **Value/cost:** human landing-page freedom and Web discovery conventions versus another cold-client request and parser path. **Lost:** independent landing-page representation without negotiation. **Decision:** does that flexibility earn an extra discovery step?

### R03 — Multiple ways to retrieve the same data

**Current:** full record, `view=properties`, `view=links`, filtered Link collection, and optional `include=links` aggregate have separate query/validator rules. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#resource-views), [aggregate](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#optional-incident-link-aggregate)

**Recommendation:** require one basic form per operation; retain extra forms only for measured request/byte savings. `include=links` may help a cold show command; do not discard that benefit without comparison. **Lost:** convenience or one-request retrieval. **Decision:** which forms materially improve our CLI?

### R04 — Snapshot-stable ordinary pagination

**Current:** every page preserves selected membership, Resource revisions and authorization projection, even in Read. An expired cursor fails explicitly. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#collection-retrieval-and-selection)

**Recommendation:** review the cost, but provisionally retain consistent pagination; it prevents skipped/duplicated or mixed-version results. A weaker keyset continuation is simpler only if callers accept moving data. This does not require exposing Transactional epochs in Read. **Lost if weakened:** one coherent enumeration. **Decision:** is that guarantee needed for ordinary list, or only explicit exports/snapshots?

### R05 — Exact-decimal equality plus binary64 admission

**Current:** numbers must survive exact-decimal → binary64 → shortest decimal without value change; the model governs no-ops, identity and digests. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#revisions)

**Recommendation:** keep one interoperable numeric rule, then test whether its present formulation imposes unnecessary token-level machinery. Compare explicitly specified binary64 semantics with rejection of lossy input; do not silently round. **Lost if simplified carelessly:** consistent retry equality and values across providers. **Decision:** which numeric values and precision does the product promise?

### R06 — External endpoints and opaque revision pins

**Current:** either endpoint may be external when the other is local; references can be URI strings or `{uri,revision}`. Pins are echoed rather than validated, and URI-only comparison drives graph incidence. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#references)

**Recommendation:** provisionally retain citation pins and external targets because Memory has an identified use; ask separately whether external *sources* are needed. Do not add validation/resolution machinery merely because a pin exists. **Lost if removed:** version-specific or external citations. **Decision:** which reference directions need generic support now?

### R07 — Several metadata planes

**Current:** Type-validated `properties`, mutable generic `metadata`, immutable per-version `attribution`, and `changeContext` are distinct. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#resource-records), [attribution](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#carried-attribution)

**Recommendation:** retain distinctions with different validation/mutability/meaning; simplify duplicate representations rather than conflating audit metadata with editable content. Metadata and attribution have concrete alignment motivation in Janet's integrated changes. **Cost:** more fields and rules in every representation. **Decision:** can the immutable journal metadata be one small envelope without losing meaning?

## Mutation

### U01 — Aliases mandatory for every writer

**Current:** aliases are optional in Read, required in both write profiles; put/delete alias add a separate locator namespace with no Resource revision. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#discovery-profile-membership)

**Recommendation:** require a concrete use case before making every store implement mutable names; consider optional naming support. **Lost:** guaranteed alias portability across all writers. **Decision:** does basic writing require aliases, or just CRUD on canonical identities?

### U02 — Aliases accepted inside mutation operands

**Current:** original alias target/miss is retained for replay, while a new retry locator is resolved transiently for identity comparison; a sequence must observe alias state at its member's turn. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#duplicate-keys-and-retained-dispositions), [sequence target](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#readupdate-sequence-target)

**Recommendation:** consider resolving a human-friendly alias to canonical identity before submitting a mutation. **Cost:** durable witnesses and turn-time resolution become simpler. **Lost:** one-request alias writes and same-sequence alias rebinding semantics; adds a client read and creates an explicit resolution moment. **Decision:** do those server-side semantics earn their cost?

### U03 — Independently committing sequence carrier

**Current:** `/sequence` commits members separately, continues after member failures, survives response loss, and supports creation labels/dependencies. Atomic batch is a separate later-profile carrier. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#readupdate-sequence-target), [durability](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#durability-and-recovery)

**Recommendation:** treat this as a distinct product requirement from atomic bd batch. Consider singleton calls for independent work and atomic batch for all-or-nothing work. **Lost:** one-round-trip partial-success execution and its efficiency; client looping is not equivalent. **Decision:** which customer workflow needs partial commit plus server-side dependency chaining?

### U04 — Semantic retry equivalence across carriers

**Current:** singleton, sequence member and one-operation atomic batch can share a key namespace and normalize to equal operations; spelling, defaults, numbers, labels and canonical references affect comparison rules. Unresolved dependencies can require a bounded direct comparison wait. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#mutation-transactions)

**Recommendation:** compare a simpler contract: stable replay of the same normalized request in one carrier, with changed carrier/request conflicting. Retain cross-carrier equivalence only if a client uses it. **Lost:** flexible retries expressed through a different API form. **Decision:** must a retry be semantically interchangeable, or just safely repeatable?

### U05 — Different lifetime rules for success and failure receipts

**Current:** Transactional completed keys retain compact outcomes for the epoch; failed receipts can be forgotten after retention; details can expire separately while allocations survive. Read+Update retains committed-key tombstones for the logical Scope’s lifetime; Transactional retains them for the epoch. Both can expire detail and eventually forget failed dispositions. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#transactional-limits), [receipts](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#mutation-receipt-responses)

**Storage cost:** successful semantic no-ops retain key tombstones too. Many distinct successful keys can therefore grow storage without creating more Resource revisions. Forgetting those keys would allow a formerly no-op request to execute after the state changes, so this cost cannot be removed while preserving the same late-retry promise. [Retention rule](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#outcome-retention)

**Recommendation:** compare a clearly bounded replay window with the present epoch-long no-repeat promise. Do not quietly expire successful deduplication while claiming its guarantee. **Lost if narrowed:** safe very-late retry under the old key. **Decision:** how long must a client be able to retry, and what happens after that time?

### U06 — JSON Patch subset and array semantics

**Current:** `add`, `replace`, `remove` over properties and metadata, including root replacement and array insertion/append; `move`, `copy`, `test` excluded. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#property-change-values)

**Recommendation:** provisionally retain precise removal, null assignment and append where comments/CLI require them. Compare a smaller set/unset/append model before changing syntax; a shallow merge alone loses functionality. **Cost:** paths, ordering, errors and array indexes. **Decision:** which mutation forms do real commands need?

### U07 — Many separately advertised limit dimensions

**Current:** byte/count/depth limits span requests, selectors, patches, sequences, transactions, diagnostics and retention, with joint-feasibility requirements. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#advertised-limits), [mutation limits](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#mutation-limits)

**Recommendation:** keep bounds needed for safety and chunking; advertise only dimensions clients can use to construct a viable request. Provider-internal budgets may need only a stable refusal. **Lost:** detailed preflight predictability. **Decision:** which advertised values alter client behavior rather than mirror server configuration?

### U08 — Allocated identities in dependent non-atomic sequences

**Current:** authority-allocated creator IDs combine with `@name` consumers and key reservation, requiring durable creator-attempt/slot facts that outlive some receipts. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#sequence-local-resource-references), [transactions](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#mutation-transactions)

**Recommendation:** retain convenient allocated IDs for create and local references inside atomic batch; separately question dependent partial-commit sequences. Preallocated client IDs simplify some workflows, but are not automatically equivalent to “run only if that creator succeeded.” **Lost if narrowed:** convenient one-call dependent partial success. **Decision:** do we need this combination, rather than each ingredient separately?

## Transactional and replication

### T01 — Atomic writes require the full replication profile

**Current:** Transactional includes atomic batch, receipts, epochs/positions, Events, changefeed and snapshots. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#part-iii--transactional)

**Recommendation:** examine whether atomic mutation can be complete without mandatory replica-serving machinery. Preserve the requested three-section organization while deciding the capability boundary; do not introduce a fourth spec by assumption. **Lost if separated:** one profile label guarantees the entire replica contract. **Decision:** should a transactional store have to serve replicas?

### T02 — Both Event Sources and a state changefeed

**Current:** Event Sources describe observable facts; change groups carry committed state and replication/control information. Their cursors, representations and consumers differ. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#events-and-event-sources), [groups](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#change-groups-and-replication)

**Recommendation:** prefer one authoritative committed-change model, with optional projections if there is demonstrated value. An existing Event stream does not become a replica log merely by renaming it. **Lost if one surface disappears:** specialized subscription/query behavior unless recreated as a projection. **Decision:** which consumers need two distinct protocol surfaces?

### T03 — Induced endpoint and owning-source facts

**Current:** a Link mutation can generate its own fact, `linked`/`unlinked` endpoint facts, and an owning-source update. **Value:** subscribers can follow one subject without reconstructing the graph. **Cost:** event fanout, ordering, projection and erasure rules. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#events-and-event-sources)

**Recommendation:** keep actual source-state changes; ask whether mandatory endpoint notifications are needed. The current spec already permits Events to be derived deterministic projections; it does not mandate physically storing each fact. The potential cut is an observable protocol obligation, not a storage optimization. **Lost:** cheap direct per-Bead observation for some consumers. **Decision:** is provider-side fanout needed for efficiency?

### T04 — Receipts as independently paginated, reauthorized retained results

**Current:** receipts expose pending/terminal states, retained detail, expired allocations, pagination, and reauthorized withheld/erased entries. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#mutation-receipt-responses)

**Recommendation:** retain retry outcome discovery; compare a compact receipt plus bounded immediate result with the full retained-result document service. Result pagination may be needed for set operations, so decide T10 with it. **Lost:** retrieval of large original result sets after disconnect. **Decision:** what must the client recover beyond success/failure, allocated IDs and commit identity?

### T05 — Strict replica reads and minimum-position barriers

**Current:** Transactional reads are strict by default; a replica must route, wait or fail if behind. Clients may demand a minimum checkpoint. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#snapshots-and-strict-reads)

**Recommendation:** preserve explicit freshness semantics but determine which deployment requires every read to be linearizable. An authority-only first implementation avoids replica-serving complexity without lying about freshness. **Lost if weakened:** default read-after-write/current-read guarantees. **Decision:** are replicas serving ordinary CLI reads in this product stage?

### T06 — Multi-stream snapshot manifests and lifecycle

**Current:** one snapshot handle anchors separate typed Bead/Link page streams, expiry and replay rendezvous; installation publishes only when complete. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#snapshots-and-strict-reads), [snapshot format](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#scope-snapshots)

**Recommendation:** if replicas remain required, retain one coherent snapshot/checkpoint; compare one typed record stream with two coordinated streams and optional inline variants. **Lost:** independent stream fetching, possibly parallelism. Ordinary paged lists are not a safe replacement bootstrap. **Decision:** what shape is needed for our scale and consumers?

### T07 — Permanent erasure ledger and per-version digest proof

**Current:** version erasure propagates through changefeed and snapshots; a permanent projected ledger survives restore/epoch/view changes, carrying `sha-256-jcs` digests and identity markers. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#version-erasure)

**Recommendation:** separate two questions: required removal everywhere under provider control, and proof of which erased bytes once existed. The earlier interest in proof is relevant; it still needs an identified consumer and cost. Do not claim erasure while retaining content in receipts/caches. **Lost if proof deferred:** verifiable content identity across holes. **Decision:** does the product require replication-wide historical erasure and digest proof now?

### T08 — Erasure plus live correction and owned-version cascades

**Current:** erasing a live version requires a same-group successor or tombstone; erasing an inline owned Link version erases containing source versions, with content-free replacement deltas and special Event withholding. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#version-erasure)

**Recommendation:** review whether corrective erasure is in scope before implementing every combination. If supported, preserve atomic valid live state and cleanup of actual embedded content; ordinary deletion is not equivalent. **Lost if deferred:** portable historical redaction/correction. **Decision:** which concrete workflow needs a corrected successor rather than ordinary update/retention?

### T09 — Special live-stream exception across erasure fences

**Current:** a fully caught-up admitted SSE stream may cross an erasure fence; lagging streams close, old checkpoints expire, and publication/backpressure races have special ordering rules. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#version-erasure)

**Recommendation:** start with the same explicit close/resnapshot rule for every affected stream, if erasure frequency and snapshot cost permit. **Lost:** uninterrupted streaming across erasure for caught-up clients; may increase transfer cost. **Decision:** is avoiding this resnapshot important enough to justify the exception?

### T10 — Set mutation beside explicit CRUD batches

**Current:** `updateWhere`/`deleteWhere` select a staged set, apply cardinality constraints and commit atomically, with bounded evaluation and result reporting. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#set-mutation)

**Recommendation:** provisionally retain where it enables real cold-client bulk commands. Read-then-explicit-CRUD adds requests and does not protect against membership changes. Evaluate individual guard/count options against actual operations rather than discarding the generic mechanism. **Lost if removed:** efficient atomic selection and mutation. **Decision:** which bulk workflows and membership guarantees must be supported?

### T11 — Mandatory live SSE alongside finite polling

**Current:** Event Sources and the Scope changefeed have finite replay and live SSE forms, with reconnect, framing and backpressure rules. [Sources](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#scope-changefeed), [Event observation](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#event-replay-and-live-observation)

**Recommendation:** consider finite polling first unless a latency or connection-load requirement justifies SSE. This is independent of the erasure-specific exception in T09. **Lost:** low-latency push and reduced idle request overhead. Retain complete-group application and reliable replay in either transport. **Decision:** what live-update latency and scale must the initial product support?

## Type-related candidates to review with Trish

These are catalog pointers, not a parallel Type lifecycle design or authorization to change Trish's work:

- **P01 — multiple inheritance and schema intersection:** is every composition rule needed for installed packs, including open-root restrictions and contradictory intersections? [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#effective-type-contracts)
- **P02 — explicit owned-Link entries plus wildcard and overlapping maxima:** Memory provides a stated ownership motivation; ask whether all configuration combinations are needed. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#owned-links)
- **P03 — external endpoint `bead` shape policy:** it checks spelling without dereferencing, so determine what useful guarantee lies between opaque external URI and forbidden external endpoint. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#link-endpoint-constraints)
- **P04 — Scope maximum multiplicity as a built-in policy:** identify actual invariants; do not remove enforcement and replace it with race-prone client checks. [Source](https://github.com/gastownhall/bdp/blob/5d7694e8be2f5f4ca72856b288dac0ecd383cd2f/docs/specs/bdp.md#scope-aggregate-constraints)

Already-deferred features such as endpoint unions, minimum multiplicity, acyclicity and tuple uniqueness are not counted as existing implementation obligations. Proposed V2 query additions are also outside this current-contract catalog.

## Dependency cuts worth discussing first

- H01 controls whether H02–H09 are prerequisites for exact revision lookup. Removing the bundle can simplify the base product without deleting useful optional future work.
- U01/U02/U03/U04/U08 interact: mutable aliases, partial commit, server-created IDs and flexible retry equivalence cause much of the durable reservation/witness/wait machinery. Decide the combination before optimizing its implementation.
- T01 determines whether every atomic writer needs T02 and T05–T09. T07/T08 determine whether T09 is needed at all.
- T10 affects T04: large set results are one reason receipts need pagination. Removing pagination while allowing unbounded result obligations does not simplify the contract coherently.
- Types/ownership influence historical completeness, authorization closure, Event fanout and erasure cascades. Keep those dependencies visible and involve Trish before Type changes.

## Decision record

No catalog item is accepted yet. For each decision record: chosen behavior, concrete use case, retained guarantee, removed/deferred obligation, affected schema/spec/runtime/conformance surfaces, and any coordination dependency. Done when the chosen behavior has a short coherent contract and aligned executable evidence. Implemented draft features can still be removed; tests prove behavior, not product necessity.

First proposed decision is H01: make exact revision reading independent of richer archival explanation. Its minimum still requires an immutable address, the complete requested historical state, authorization and applicable erasure checks, and distinct temporary service failures. Then review H02: does any client need an inventory of the missing reconstruction pieces? Recommendation: defer that inventory until a repair consumer needs it.
