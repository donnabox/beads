# Pinned source/spec excerpt inventory

Line numbers refer to exact copied inputs, not newer specification or main.

### docs/specs/bdp.md:5386–5424

5386: ### Events and Event Sources
5387: 
5388: > **Transactional contract.**
5389: >
5390: > The complete Event ordering, transaction framing, and Event Source guarantees
5391: > below are required only by the Transactional profile. Read and Read+Update do
5392: > not publish or expose BDP Events.
5393: 
5394: An **Event** is an immutable authority-generated record of a committed fact.
5395: Each Event has an immutable ID that is unique within exactly one **Event
5396: Source**. An Event is not required to be an independently addressable
5397: Resource, but its Event Source is a Resource.
5398: 
5399: Every Bead and Link is an Event **subject** and has an associated
5400: Resource-scoped Event Source. The subject and the Event Source have distinct
5401: identities, and they may have different lifetimes. An Event Source may remain
5402: observable after its subject is deleted, subject to retention policy.
5403: 
5404: Resource-scoped Event Sources are not independently committed logs. They are
5405: deterministic projections of the semantic Events inside the Scope's committed
5406: change groups. Each Event has a stable ordinal within its change group. Its
5407: source-local opaque ID and cursor are stable functions of the group
5408: checkpoint, the ordinal, and the projection. An implementation may
5409: materialize or index a projection without changing its contents or order.
5410: 
5411: The model defines five domain-independent Event Types:
5412: 
5413: - **created** — a Bead or Link began to exist;
5414: - **updated** — the mutable properties or common metadata of a Bead or Link changed, or an
5415:   owned Link of a Bead changed;
5416: - **deleted** — a Bead or Link ceased to exist;
5417: - **linked** — a Link became incident upon a Bead; and
5418: - **unlinked** — a Link ceased to be incident upon a Bead.
5419: 
5420: Events are the observation-side duals of singleton DML operations. A Resource
5421: read or snapshot bootstrap conveys current state. An Event instead conveys
5422: the committed delta — the change that advances previously observed state.
5423: Every Event identifies its subject by immutable `id` and `type`, and it
5424: carries the transaction in which the fact committed.

### docs/specs/bdp.md:5438–5455

5438: UpdatedData { previousRevision: Revision revision: Revision change?:
5439: PropertyChange // changes properties metadataChange?: PropertyChange //
5440: changes common metadata ownedLink?: OwnedLinkChange revisionProvenance?:
5441: RevisionProvenance // the new version's carried attribution }
5442: 
5443: OwnedLinkChange {
5444:   operation: created | updated | deleted
5445:   link: LinkState                // created: the owned Link's complete record
5446:       | OwnedLinkDelta           // updated: the owned Link's own delta
5447:       | ResourceIdentity         // deleted: id, type, and final live revision
5448: }
5449: 
5450: OwnedLinkDelta { id: URI // the owned Link's canonical URL type: TypeId
5451: previousRevision: Revision // the Link's revisions, not the source's
5452: revision: Revision change?: PropertyChange // changes the Link's properties
5453: metadataChange?: PropertyChange // changes the Link's common metadata
5454: revisionProvenance?: RevisionProvenance // the Link's new version's carried
5455: attribution }

### docs/specs/bdp.md:5503–5557

5503: 
5504: `CreatedData` and `DeletedData` carry no owned-Link data. A Bead is created
5505: with an empty owned set for every Link Type its Type owns, and the record's
5506: empty `ownedLinks` entries follow from the Type Descriptor rather than from
5507: the Event: a consumer that reconstructs a record from Events alone cannot
5508: know which empty entries the record carries without the Type Descriptor,
5509: and the canonical record read or the snapshot, not the Event stream, is
5510: where that key set is authoritative. A Bead with a live owned Link cannot
5511: be deleted, so a `deleted` Bead Event never has owned Links to report.
5512: 
5513: A source Bead's owned-Link `updated` Event is in addition to, not instead
5514: of, the facts the Link mutation already induces: the Link's own `created`,
5515: `updated`, or `deleted` fact, and the `linked` or `unlinked` fact at each
5516: in-Scope endpoint, including the source itself. Within a change group, the
5517: facts induced by one owned-Link operation are ordered: the Link's lifecycle
5518: fact first, then the graph facts at its in-Scope endpoints, source before
5519: target — a self-Link's one endpoint Bead receiving its `source` fact before
5520: its `target` fact — then the source's `updated` fact last. Ordinals are
5521: assigned in that order and never renumbered by projection. A Bead-scoped
5522: Event Source for an owning source therefore reports an owned Link's
5523: property change twice, under two subjects: once as the incident Link's
5524: `updated` fact and once as the source's own `updated` fact carrying the
5525: same delta.
5526: 
5527: A no-op owned-Link update — one whose resulting `properties` and `metadata`
5528: both equal, under the RFC 6902 Section 4.6 comparison, their respective values
5529: immediately before it — retains the Link's revision and emits no Event, and
5530: it does not version the source: there is no transition for the source's
5531: version to cover. An omitted change leaves its corresponding value unchanged.
5532: 
5533: A consumer that holds the source's record at `previousRevision` advances it
5534: to `revision` by applying `ownedLink` to the entry keyed by `link.type` in
5535: the record's `ownedLinks` member: for `created`, inserting `link` in
5536: ascending code-unit order of `id`; for `updated`, locating the entry whose
5537: `id` equals `link.id` and whose `revision` equals `link.previousRevision`,
5538: applying `link.change` to its `properties` when present and
5539: `link.metadataChange` to its `metadata` when present, leaving each omitted
5540: value unchanged, and setting its `revision` to `link.revision` and its
5541: `revisionProvenance` to `link.revisionProvenance`, removing that member when
5542: the delta carries none; for `deleted`, removing the entry whose `id` equals
5543: `link.id`; then setting the record's `revision` to the Event's `revision` and
5544: its `revisionProvenance` to the Event's `revisionProvenance`, removing the
5545: member when the Event carries none. A consumer whose held revision is not
5546: `previousRevision`, or whose held entry is not at `link.previousRevision`, is
5547: not positioned to apply the delta; it re-reads the record or resumes from a
5548: snapshot. Replicas do not need the delta at all: the containing change
5549: group's `changes` member carries the source Bead's complete postimage,
5550: `ownedLinks` inline, beside the Link's own postimage or tombstone.
5551: 
5552: `CreatedData` contains the complete initial properties and metadata, because
5553: creation is the delta from absence to the initial state; previously retained
5554: Events may omit metadata as described above. For a Link, it also contains the
5555: Link's source and target endpoint references. `UpdatedData` contains the
5556: committed Property Change for properties and/or metadata — or the owned-Link
5557: change — rather than a resulting state snapshot.

### docs/specs/bdp.md:5583–5606

5583: A Link-scoped Event Source reports `created`, `updated`, and `deleted` facts
5584: about that Link. A Bead-scoped Event Source reports:
5585: 
5586: - `created`, `updated`, and `deleted` facts about the Bead;
5587: - `linked` and `unlinked` facts when a Link becomes or ceases to be incident
5588:   upon the Bead; and
5589: - `updated` facts whose subject is an incident Link when that Link's mutable
5590:   properties change.
5591: 
5592: Link creation produces a `created` fact about the Link and a `linked` fact at
5593: each in-Scope endpoint Bead. Link deletion produces a `deleted` fact about
5594: the Link and an `unlinked` fact at each in-Scope endpoint Bead. No
5595: Bead-scoped fact or Event Source exists for an opaque out-of-Scope endpoint.
5596: A wider Event Source may cover a collection, a graph Scope, or a complete
5597: service.
5598: 
5599: For a self-Link, whose source and target are the same Bead, that one endpoint
5600: Bead receives two graph facts in the same group: one whose `endpoint` is
5601: `source` and one whose `endpoint` is `target`. Both count against the
5602: transaction's Event-expansion limit. These derived facts and the incident
5603: Link view do not mutate the Bead or advance its Resource revision; a
5604: source Bead whose Type owns the Link's type is versioned by the owned
5605: change itself, under [Owned Links](#owned-links), not by these
5606: derived facts.

### docs/specs/bdp.md:5644–5657

5644: ### Transactional revision and History extensions
5645: 
5646: A semantic no-op emits no `updated` Event. A change to properties or common
5647: metadata is one Resource transition with one new revision and one `updated`
5648: Event. A deleted identity's revision is the same final live revision reported
5649: by the Scope changefeed tombstone and `DeletedData.revision`.
5650: 
5651: Revision allocation failures after Transactional admission follow the complete
5652: transaction rollback and retained-failure rules. Under advertised History,
5653: all versions in one atomic transaction share the authority-observed commit
5654: instant; independently committed sequence members retain their own instants.
5655: Context does not alter Event-time meaning. The operation-local context inputs
5656: also apply to set operations and batch members, with fan-out to every version
5657: actually minted. No carrier-wide override is introduced.

### docs/specs/bdp.md:5770–5851

5770: ### Snapshots and strict reads
5771: 
5772: > **Transactional/Replication only — Transactional profile.**
5773: >
5774: > Implementations of the Read and Read+Update profiles may skip snapshot bootstrap,
5775: > snapshot/changefeed rendezvous, minimum-position reads, and the strict
5776: > replica-freshness contract in this section.
5777: 
5778: A first-class Scope **Snapshot** contains the complete live Bead and Link
5779: state visible in one Authorization View at one transaction-consistent Scope
5780: epoch and position. One immutable snapshot manifest anchors separate typed
5781: Bead and Link page streams to the same handle, view, position, and expiry. A
5782: small Scope may inline both complete streams. A replica stages all pages and
5783: publishes the replacement atomically only after both streams finish.
5784: Ordinary collection queries are not a replication bootstrap.
5785: 
5786: The snapshot checkpoint is the precise exclusive position from which Scope
5787: changefeed replay begins. Until the snapshot's advertised expiry, the
5788: authority retains every later projected group required to continue from that
5789: checkpoint. A cursor presented too late, from another Scope epoch, or from
5790: another Authorization View fails explicitly and never silently skips history.
5791: BDP v0 assumes a global retention window rather than per-client retention
5792: pins.
5793: 
5794: Ordinary reads are strict by default. Each read observes one
5795: transaction-consistent prefix of its Authorization View that can be
5796: linearized during the request, and it reports its Scope epoch, view token,
5797: and visible position. A client may require a minimum checkpoint bound to that
5798: same view. A replica that is behind must route, wait, or fail explicitly; it
5799: must not return older state as if current. Weaker consistency modes, if
5800: added, require explicit client selection.
5801: 
5802: ### Transactional HTTP consistency
5803: 
5804: Transactional Scope-bounded responses carry:
5805: 
5806: ```http
5807: BDP-Scope-Epoch: opaque-scope-epoch
5808: BDP-Authorization-View: opaque-authorization-view
5809: BDP-Scope-Position: opaque-position-42
5810: ```
5811: 
5812: On reads, a Transactional client requests a minimum visible position by sending those
5813: same epoch and view fields plus:
5814: 
5815: ```http
5816: BDP-Minimum-Scope-Position: opaque-position-42
5817: ```
5818: 
5819: If any of these three request fields is present on a read, the request MUST
5820: contain exactly one nonempty value for each field, using the
5821: [Event-ID and checkpoint character profile](#event-id-and-checkpoint-character-profile). Missing,
5822: repeated, or lexically invalid values MUST return `400 malformed-request`
5823: after ordinary transport, target/method, and authentication/authorization checks,
5824: but before any owner read or target-existence lookup. All three absent means
5825: no minimum is requested. Lexical validity does not establish that a position
5826: identity is known or supported; semantic handling follows the rules below.
5827: 
5828: This receiving refusal observes no projected prefix and MUST NOT report
5829: `BDP-Scope-Epoch`, `BDP-Authorization-View`, or `BDP-Scope-Position`. It MUST NOT
5830: become a conditional `304` or `412`, a negotiation `406`, or an SSE stream.
5831: Earlier ordinary target/method and Origin-policy refusals remain unchanged.
5832: HEAD retains the corresponding GET response metadata and sends no body.
5833: 
5834: For an ordinary unknown, deleted, or authorization-hidden Bead or Link
5835: Resource (including its properties view), alias target, or initial Bead
5836: view=links anchor, the authority MUST return the uniform 404 resource-not-found
5837: before semantic minimum comparison or waiting. This refusal reports the actual
5838: current observation context and does not disclose the hidden target. Ordinary
5839: request validation and the lexical receiving rule above still apply.
5840: 
5841: For the remaining reads, the authority returns a representation at that position
5842: or later, waits or routes to an eligible replica, or returns the normative
5843: foreign-view, cursor-expired, or catch-up-timeout problem. It never reports
5844: success with an older position. For an existing authorized current Bead or Link
5845: Resource or properties view, a lexically valid but unrecognized minimum position
5846: with the actual current epoch and view does not prove eligibility. The authority
5847: MUST use its bounded minimum-read wait and, if eligibility remains unproved,
5848: return 503 catch-up-timeout with retry: "after-delay"; it MUST NOT substitute
5849: a stale success, fabricated expiry, or lexical malformed-request refusal.
5850: This rule does not establish precedence for retained continuations or Snapshots,
5851: nor extend unrecognized-position waiting to successful incident-Link lists.

### docs/specs/bdp.md:5860–5890

5860: Intermediaries must not cache or transform the stream.
5861: 
5862: ### Transactional Resource read context
5863: 
5864: Every Transactional Scope-bounded read reports the Scope epoch, Authorization
5865: View token, and position of the transaction-consistent projected prefix it
5866: observed. It reports them through the fields defined under
5867: [Transactional HTTP consistency](#transactional-http-consistency). Read and Read+Update use Resource
5868: revisions, entity tags, and snapshot-preserving pagination without exposing
5869: Scope-history tokens.
5870: 
5871: ### Transactional conditional reads
5872: 
5873: A `304` is not delivery or application of a change group or erasure record
5874: and MUST NOT advance a client's durable checkpoint.
5875: 
5876: Scope epoch/view, minimum-position and erasure/expiry checks MUST precede any
5877: conditional shortcut. Required current Scope context remains required on
5878: 304 and 412 responses.
5879: 
5880: The inherited HEAD rules also cover `changes/`, Scope `events/`, Resource
5881: `view=events`, snapshot handles/streams and receipt/page URLs. For acceptable
5882: SSE, HEAD returns the response metadata and ends without frames; it MUST NOT
5883: remain open merely to stream events. HEAD acknowledges neither delivery nor
5884: application and creates no mutation transaction.
5885: 
5886: ### Transactional HTTP field support
5887: 
5888: The inherited [Mutation command preconditions](#mutation-command-preconditions)
5889: also apply here: HTTP conditional fields on operation command targets do not
5890: guard newly produced receipts.

### docs/specs/bdp.md:5970–6006

5970: The manifest's `id` MUST support GET and HEAD while the snapshot handle
5971: remains valid. GET retrieves the same manifest, including its identity,
5972: anchor, checkpoint, expiry, erasure ledger, and initial stream pages; it MUST
5973: NOT create a replacement snapshot or extend `expiresAt`. The authority retains
5974: or reconstructs that manifest. Existing epoch/view validity and erasure fences
5975: apply. An authorized expired or no-longer-available snapshot handle returns
5976: `410 cursor-expired`; an unknown or undisclosable target retains uniform
5977: `404 resource-not-found`. Actual service inability before the promised expiry
5978: is a service failure, not an expiration diagnosis.
5979: 
5980: The following snapshot table summarizes the existing scoped response laws
5981: and shared HTTP failures. It does not exclude ordinary applicable failures,
5982: including authentication and rate limiting.
5983: 
5984: | Snapshot target | Method | Response |
5985: | --- | --- | --- |
5986: | discovered creation target | GET, HEAD | `200` snapshot manifest metadata/body as appropriate; ordinary authentication, authorization, request, consistency, and service failures |
5987: | manifest `id` or supplied stream continuation | GET, HEAD | `200` the retained manifest or stream page; `404 resource-not-found` for an unknown/undisclosable target; `410 cursor-expired` for an authorized expired/unavailable handle, including timed or erasure expiry; ordinary epoch/view, minimum-position, and service failures |
5988: | either target | other methods | `405`, `Allow: GET, HEAD`, with CORS `OPTIONS` handled under the shared rules |
5989: 
5990: The shared native `406`, conditional `304`/`412`, and bodyless `500` rules
5991: apply. No conditional shortcut bypasses a snapshot fence. HEAD has the
5992: lifecycle and body rules under [Conditional reads and HEAD](#conditional-reads-and-head).
5993: 
5994: Each non-null `next` is an absolute URL that the server generates. That URL is
5995: bound to the snapshot identity, the stream kind, the Scope epoch, the
5996: Authorization View, the position, and the expiry. Bead and Link streams may be
5997: paged independently, but they never move to a newer anchor. The authority
5998: keeps the snapshot continuable through `expiresAt`. If the authority cannot
5999: serve a page before then, that is a service failure, not an expired-cursor
6000: result. A client stages all pages, then exposes the replacement generation
6001: atomically only after both streams end.
6002: 
6003: `checkpoint` is the opaque, epoch-qualified value that is supplied as the
6004: exclusive `after` cursor to the Scope changefeed. The authority retains every
6005: group after `scopePosition` through `expiresAt`. Collection cursors remain
6006: query snapshots. They do not substitute for this complete projected Scope

### schemas/bdp-v0.schema.json:2877–2922

2877:     "ownedLinkDelta": {
2878:       "type": "object",
2879:       "required": ["id", "type", "previousRevision", "revision"],
2880:       "properties": {
2881:         "id": {
2882:           "$ref": "#/$defs/absoluteHttpUrl"
2883:         },
2884:         "type": {
2885:           "$ref": "#/$defs/absoluteHttpUrl"
2886:         },
2887:         "previousRevision": {
2888:           "$ref": "#/$defs/revision"
2889:         },
2890:         "revision": {
2891:           "$ref": "#/$defs/revision"
2892:         },
2893:         "change": {
2894:           "$ref": "#/$defs/propertyChange"
2895:         },
2896:         "metadataChange": {
2897:           "$ref": "#/$defs/propertyChange"
2898:         },
2899:         "revisionProvenance": {
2900:           "$ref": "#/$defs/revisionProvenance"
2901:         }
2902:       },
2903:       "anyOf": [
2904:         {
2905:           "required": ["change"],
2906:           "properties": {
2907:             "change": true
2908:           }
2909:         },
2910:         {
2911:           "required": ["metadataChange"],
2912:           "properties": {
2913:             "metadataChange": true
2914:           }
2915:         }
2916:       ],
2917:       "additionalProperties": false
2918:     },
2919:     "ownedLinkChange": {
2920:       "type": "object",
2921:       "required": ["operation", "link"],
2922:       "properties": {

### schemas/bdp-v0.schema.json:2985–3043

2985:     "updatedData": {
2986:       "type": "object",
2987:       "required": ["previousRevision", "revision"],
2988:       "properties": {
2989:         "previousRevision": {
2990:           "$ref": "#/$defs/revision"
2991:         },
2992:         "revision": {
2993:           "$ref": "#/$defs/revision"
2994:         },
2995:         "change": {
2996:           "$ref": "#/$defs/propertyChange"
2997:         },
2998:         "metadataChange": {
2999:           "$ref": "#/$defs/propertyChange"
3000:         },
3001:         "ownedLink": {
3002:           "$ref": "#/$defs/ownedLinkChange"
3003:         },
3004:         "revisionProvenance": {
3005:           "$ref": "#/$defs/revisionProvenance"
3006:         }
3007:       },
3008:       "oneOf": [
3009:         {
3010:           "type": "object",
3011:           "properties": {
3012:             "change": true,
3013:             "metadataChange": true,
3014:             "ownedLink": false
3015:           },
3016:           "anyOf": [
3017:             {
3018:               "required": ["change"],
3019:               "properties": {
3020:                 "change": true
3021:               }
3022:             },
3023:             {
3024:               "required": ["metadataChange"],
3025:               "properties": {
3026:                 "metadataChange": true
3027:               }
3028:             }
3029:           ]
3030:         },
3031:         {
3032:           "type": "object",
3033:           "required": ["ownedLink"],
3034:           "properties": {
3035:             "ownedLink": true,
3036:             "change": false,
3037:             "metadataChange": false
3038:           }
3039:         }
3040:       ],
3041:       "additionalProperties": false
3042:     },
3043:     "deletedData": {

### schemas/bdp-v0.schema.json:3191–3225

3191:     "eventPage": {
3192:       "title": "BDP Event Source page",
3193:       "type": "object",
3194:       "required": ["source", "events", "next"],
3195:       "properties": {
3196:         "source": {
3197:           "$ref": "#/$defs/absoluteHttpUrl"
3198:         },
3199:         "events": {
3200:           "type": "array",
3201:           "items": {
3202:             "$ref": "#/$defs/event"
3203:           }
3204:         },
3205:         "next": {
3206:           "oneOf": [
3207:             {
3208:               "$ref": "#/$defs/absoluteHttpUrl"
3209:             },
3210:             {
3211:               "type": "null"
3212:             }
3213:           ]
3214:         }
3215:       },
3216:       "additionalProperties": false
3217:     },
3218:     "selector": {
3219:       "type": "string",
3220:       "minLength": 1
3221:     },
3222:     "cardinality": {
3223:       "type": "object",
3224:       "properties": {
3225:         "min": {

### packages/server/src/transaction-http.ts:149–156

149:   response.headers.set("x-bdp-development-limit", feature);
150:   return response;
151: }
152: export function hasMinimumControls(request: Request): boolean {
153:   return ["bdp-minimum-scope-position", "bdp-scope-epoch", "bdp-authorization-view"].some((field) =>
154:     request.headers.has(field),
155:   );
156: }

### packages/server/src/transaction-http.ts:403–467

403:     if (authorization.kind === "unauthenticated") return transactionHttpProblem("unauthenticated");
404:     if (malformedTransactionMinimumHeaders(request.headers))
405:       return transactionHttpProblem("malformed-request");
406:     if (url.pathname === new URL("snapshot", scope).pathname) {
407:       const fields = [...url.searchParams.keys()];
408:       const id = url.searchParams.get("snapshot"),
409:         stream = url.searchParams.get("stream"),
410:         token = url.searchParams.get("page");
411:       let target: TransactionSnapshotRequest;
412:       // These closed query shapes are local route conventions. Unknown well-
413:       // shaped identities go to the owner; they are never guessed as expired.
414:       if (new Set(fields).size !== fields.length)
415:         return transactionHttpProblem("invalid-parameter");
416:       if (fields.length === 0) target = { kind: "create" };
417:       else if (fields.length === 1 && id !== null && id !== "") target = { kind: "manifest", id };
418:       else if (fields.length === 3 && id && token && (stream === "beads" || stream === "links"))
419:         target = { kind: "page", id, token, stream: stream === "beads" ? "bead" : "link" };
420:       else return transactionHttpProblem("invalid-parameter");
421:       // Only fresh creation has an unambiguous current observation. Retained
422:       // anchors and unsupported tuple shapes keep their explicit private gap.
423:       const minimum = hasMinimumControls(request)
424:         ? currentReadMinimumHeaders(request.headers)
425:         : undefined;
426:       if (hasMinimumControls(request) && (target.kind !== "create" || minimum === undefined))
427:         return developmentGap("minimum-position");
428:       function snapshotResponse(result: Awaited<ReturnType<Service["snapshot"]>>): HttpResponse {
429:         if (result.kind === "problem") {
430:           const response = context(transactionHttpProblem(result.code), result.context);
431:           if (result.retryAfter !== undefined)
432:             response.headers.set("retry-after", String(result.retryAfter));
433:           return response;
434:         }
435:         if (result.kind === "eligible")
436:           return context(
437:             {
438:               status: 200,
439:               headers: new Headers({
440:                 "content-type": "application/json",
441:                 "cache-control": "private, no-store",
442:               }),
443:             },
444:             result.context,
445:           );
446:         return context(jsonResponse(result.body), result.context);
447:       }
448:       if (target.kind === "create") {
449:         // A pure successful-JSON candidate may select direct capture, but never
450:         // authorizes an owner-free response. Nonallocating decisions still need
451:         // authoritative capacity/integrity checks and same-entry context.
452:         const candidate = applyReadHttpSemantics(request, {
453:           status: 200,
454:           headers: new Headers({ "content-type": "application/json" }),
455:         });
456:         if (request.method === "HEAD" || candidate.status !== 200) {
457:           const eligible = snapshotResponse(
458:             await service.snapshot({ kind: "preflight" }, minimum, request.signal),
459:           );
460:           // Candidate and eligible share successful JSON with no validator;
461:           // their media/conditional decisions agree. This branch is terminal:
462:           // never start a second capture or renew the minimum wait deadline.
463:           return applyReadHttpSemantics(request, eligible);
464:         }
465:       }
466:       return snapshotResponse(await service.snapshot(target, minimum, request.signal));
467:     }

### packages/server/src/transaction-http.ts:657–679

657:         else if (error instanceof TransactionEventProjectionGap)
658:           response = developmentGap(error.message);
659:         else if (error instanceof TransactionMinimumUnsupported)
660:           response = developmentGap("minimum-position");
661:         else if (error instanceof TransactionMinimumVisibilityGap)
662:           response = developmentGap("minimum-position-visibility");
663:         else response = empty(500);
664:       } finally {
665:         readPrincipals.delete(request);
666:       }
667:       response.headers.set("x-bdp-development", "non-attesting");
668:       if (request.method === "POST") response.headers.set("vary", "Accept");
669:       if (
670:         "kind" in response &&
671:         (response.kind === "changefeed-stream" || response.kind === "event-stream")
672:       )
673:         return response;
674:       return {
675:         ...(request.method === "GET" || request.method === "HEAD"
676:           ? applyReadHttpSemantics(request, response)
677:           : response),
678:         kind: "finite",
679:       };

### packages/server/src/transaction-consistency.ts:1–31

1: /** Source-private minimum receiving and canonical semantic subset.
2:  * Alphabet-valid unknown identities and continuation comparisons remain separate. */
3: import { setTimeout as delay } from "node:timers/promises";
4: import { isProxy } from "node:util/types";
5: import { ScopeServerOperationAbortedError } from "./read-request.js";
6: import type { TransactionReadContext } from "./transaction-store.js";
7: 
8: export interface TransactionMinimum {
9:   readonly scopeEpoch: string;
10:   readonly authorizationView: string;
11:   readonly minimumPosition: string;
12: }
13: const token = /^[A-Za-z0-9_-]{1,256}$/;
14: /** Receiving syntax only: unknown token identities are not malformed fields. */
15: export function malformedTransactionMinimumHeaders(headers: Headers): boolean {
16:   const values = [
17:     headers.get("bdp-minimum-scope-position"),
18:     headers.get("bdp-scope-epoch"),
19:     headers.get("bdp-authorization-view"),
20:   ];
21:   return (
22:     values.some((value) => value !== null) &&
23:     values.some((value) => value === null || !token.test(value))
24:   );
25: }
26: const position = (value: unknown): number | undefined => {
27:   if (value === "genesis") return 0;
28:   if (typeof value !== "string" || !/^pos-(0|[1-9][0-9]*)$/.test(value)) return undefined;
29:   const number = Number(value.slice(4));
30:   return Number.isSafeInteger(number) ? number : undefined;
31: };

### packages/server/src/transaction-consistency.ts:33–131

33: export function snapshotCurrentReadMinimum(value: unknown): TransactionMinimum {
34:   if (!value || typeof value !== "object" || isProxy(value))
35:     throw new TypeError("unsupported minimum tuple");
36:   const fields = Object.getOwnPropertyDescriptors(value);
37:   if (
38:     Object.keys(fields).sort().join(",") !== "authorizationView,minimumPosition,scopeEpoch" ||
39:     Reflect.ownKeys(value).length !== 3 ||
40:     Object.values(fields).some((field) => !("value" in field))
41:   )
42:     throw new TypeError("unsupported minimum tuple");
43:   const scopeEpoch: unknown = fields.scopeEpoch?.value;
44:   const authorizationView: unknown = fields.authorizationView?.value;
45:   const minimumPosition: unknown = fields.minimumPosition?.value;
46:   if (
47:     typeof scopeEpoch !== "string" ||
48:     !token.test(scopeEpoch) ||
49:     typeof authorizationView !== "string" ||
50:     !token.test(authorizationView) ||
51:     typeof minimumPosition !== "string" ||
52:     !token.test(minimumPosition)
53:   )
54:     throw new TypeError("unsupported minimum tuple");
55:   return Object.freeze({ scopeEpoch, authorizationView, minimumPosition });
56: }
57: export function snapshotTransactionMinimum(value: unknown): TransactionMinimum {
58:   const minimum = snapshotCurrentReadMinimum(value);
59:   if (position(minimum.minimumPosition) === undefined)
60:     throw new TypeError("unsupported minimum tuple");
61:   return minimum;
62: }
63: export function currentReadMinimumHeaders(headers: Headers): TransactionMinimum | undefined {
64:   try {
65:     return snapshotCurrentReadMinimum({
66:       scopeEpoch: headers.get("bdp-scope-epoch"),
67:       authorizationView: headers.get("bdp-authorization-view"),
68:       minimumPosition: headers.get("bdp-minimum-scope-position"),
69:     });
70:   } catch {
71:     return undefined;
72:   }
73: }
74: export function transactionMinimumHeaders(headers: Headers): TransactionMinimum | undefined {
75:   try {
76:     return snapshotTransactionMinimum({
77:       scopeEpoch: headers.get("bdp-scope-epoch"),
78:       authorizationView: headers.get("bdp-authorization-view"),
79:       minimumPosition: headers.get("bdp-minimum-scope-position"),
80:     });
81:   } catch {
82:     return undefined;
83:   }
84: }
85: export class TransactionMinimumFailure extends Error {
86:   constructor(
87:     readonly reason: "behind" | "foreign-view" | "cursor-expired" | "catch-up-timeout",
88:     readonly context: TransactionReadContext,
89:   ) {
90:     super(`transaction minimum read: ${reason}`);
91:   }
92: }
93: /** Runs within the representation's owner entry, after authorization. */
94: export function checkTransactionMinimum(
95:   minimum: TransactionMinimum,
96:   context: TransactionReadContext,
97: ): void {
98:   if (minimum.scopeEpoch !== context.scopeEpoch)
99:     throw new TransactionMinimumFailure("cursor-expired", context);
100:   if (minimum.authorizationView !== context.authorizationView)
101:     throw new TransactionMinimumFailure("foreign-view", context);
102:   const required = position(minimum.minimumPosition),
103:     current = position(context.scopePosition);
104:   if (required === undefined || current === undefined)
105:     throw new Error("invalid owner minimum position");
106:   if (current < required) throw new TransactionMinimumFailure("behind", context);
107: }
108: /** An unknown identity is not behind or malformed, and carries no ordering claim. */
109: export class TransactionMinimumUnproved extends Error {
110:   constructor(readonly context: TransactionReadContext) {
111:     super("current Resource minimum eligibility unproved");
112:   }
113: }
114: export class TransactionMinimumUnsupported extends Error {}
115: export function checkCurrentReadMinimum(
116:   minimum: TransactionMinimum,
117:   context: TransactionReadContext,
118:   waitable: boolean,
119: ): void {
120:   if (position(context.scopePosition) === undefined)
121:     throw new Error("invalid owner minimum position");
122:   if (position(minimum.minimumPosition) === undefined) {
123:     if (!waitable) throw new TransactionMinimumUnsupported("minimum-position");
124:     if (minimum.scopeEpoch !== context.scopeEpoch)
125:       throw new TransactionMinimumFailure("cursor-expired", context);
126:     if (minimum.authorizationView !== context.authorizationView)
127:       throw new TransactionMinimumFailure("foreign-view", context);
128:     throw new TransactionMinimumUnproved(context);
129:   }
130:   checkTransactionMinimum(minimum, context);
131: }

### packages/server/src/transaction-consistency.ts:137–178

137: /** No SQLite entry survives an await. The original finite deadline never renews. */
138: export async function waitForTransactionMinimum<T>(
139:   attempt: () => T | Promise<T>,
140:   options: TransactionMinimumWaitOptions = {},
141:   signal?: AbortSignal,
142: ): Promise<T> {
143:   const budget = options.budgetMs ?? 1000,
144:     now = options.now ?? (() => performance.now()),
145:     wait = options.wait ?? ((milliseconds) => delay(milliseconds));
146:   if (
147:     !Number.isSafeInteger(budget) ||
148:     budget <= 0 ||
149:     budget > 2147483647 ||
150:     typeof now !== "function" ||
151:     typeof wait !== "function"
152:   )
153:     throw new TypeError("finite minimum-read wait policy required");
154:   const started = now();
155:   if (!Number.isFinite(started)) throw new TypeError("invalid minimum-read clock");
156:   const deadline = started + budget;
157:   let previous = started;
158:   if (!Number.isFinite(deadline)) throw new TypeError("invalid minimum-read deadline");
159:   for (;;) {
160:     if (signal?.aborted) throw new ScopeServerOperationAbortedError();
161:     try {
162:       return await attempt();
163:     } catch (error) {
164:       if (
165:         !(error instanceof TransactionMinimumUnproved) &&
166:         (!(error instanceof TransactionMinimumFailure) || error.reason !== "behind")
167:       )
168:         throw error;
169:       const current = now();
170:       if (!Number.isFinite(current) || current < previous)
171:         throw new TypeError("invalid minimum-read clock");
172:       previous = current;
173:       if (current >= deadline)
174:         throw new TransactionMinimumFailure("catch-up-timeout", error.context);
175:       await wait(Math.min(5, deadline - current));
176:     }
177:   }
178: }

### packages/server/src/transaction-service.ts:423–465

423:       request: TransactionSnapshotRequest,
424:       supplied?: TransactionMinimum,
425:       signal?: AbortSignal,
426:     ) {
427:       // Without a minimum, preserve the synchronous owner path: it does not
428:       // observe signal or claim interruption of retained reads or copying.
429:       if (supplied === undefined) return run(async () => owner.observeSnapshot(request));
430:       let captured: { minimum: TransactionMinimum } | { error: TypeError };
431:       try {
432:         captured = { minimum: snapshotTransactionMinimum(supplied) };
433:       } catch (error) {
434:         if (!(error instanceof TypeError)) throw error;
435:         captured = { error };
436:       }
437:       let minimum: TransactionMinimum;
438:       let kind: TransactionSnapshotRequest["kind"];
439:       if ("error" in captured) {
440:         // Revalidate only data fields synchronously; kind/lifecycle errors are
441:         // outside both catches. Excluded requests retain the original refusal.
442:         try {
443:           minimum = snapshotCurrentReadMinimum(supplied);
444:         } catch {
445:           throw captured.error;
446:         }
447:         kind = request.kind;
448:         if (kind !== "create" && kind !== "preflight") throw captured.error;
449:       } else {
450:         minimum = captured.minimum;
451:         kind = request.kind;
452:       }
453:       if (kind !== "create" && kind !== "preflight")
454:         throw new TypeError("minimum Snapshot observation requires fresh creation or preflight");
455:       // Freeze both controls before the first asynchronous attempt. Each owner
456:       // entry revalidates the tuple; no SQLite entry survives a finite wait.
457:       const copied = Object.freeze({ kind });
458:       return run(() =>
459:         waitForTransactionMinimum(
460:           () => owner.observeSnapshot(copied, minimum),
461:           minimumReads,
462:           signal,
463:         ),
464:       );
465:     },

### packages/server/src/transaction-store.ts:1558–1635

1558:     observeSnapshot(request: TransactionSnapshotRequest, supplied?: TransactionMinimum) {
1559:       let minimum: TransactionMinimum | undefined;
1560:       let observedRequest = request;
1561:       let unrecognizedMinimum = false;
1562:       if (supplied !== undefined) {
1563:         let captured: { minimum: TransactionMinimum } | { error: TypeError };
1564:         try {
1565:           captured = { minimum: snapshotTransactionMinimum(supplied) };
1566:         } catch (error) {
1567:           if (!(error instanceof TypeError)) throw error;
1568:           captured = { error };
1569:         }
1570:         if ("error" in captured) {
1571:           try {
1572:             minimum = snapshotCurrentReadMinimum(supplied);
1573:           } catch {
1574:             throw captured.error;
1575:           }
1576:           const kind = request.kind;
1577:           if (kind !== "create" && kind !== "preflight") throw captured.error;
1578:           observedRequest = Object.freeze({ kind });
1579:           unrecognizedMinimum = true;
1580:         } else minimum = captured.minimum;
1581:       }
1582:       if (minimum && observedRequest.kind !== "create" && observedRequest.kind !== "preflight")
1583:         throw new TypeError("minimum Snapshot observation requires fresh creation or preflight");
1584:       return enter(() => {
1585:         const context = readContext(),
1586:           now = instant();
1587:         if (minimum) {
1588:           if (unrecognizedMinimum) checkCurrentReadMinimum(minimum, context, true);
1589:           else checkTransactionMinimum(minimum, context);
1590:         }
1591:         try {
1592:           atomic(() => snapshots.expire(now, context));
1593:           if (observedRequest.kind === "preflight") {
1594:             atomic(() => snapshots.preflight(now, context));
1595:             return Object.freeze({ kind: "eligible" as const, context });
1596:           }
1597:           if (observedRequest.kind === "create") {
1598:             const manifest = atomic(() => {
1599:               const value = snapshots.capture(now, context);
1600:               options.fault?.("after-snapshot-copy");
1601:               return value;
1602:             });
1603:             options.fault?.("after-snapshot-commit");
1604:             return Object.freeze({ kind: "manifest" as const, body: manifest, context });
1605:           }
1606:           if (observedRequest.kind === "manifest")
1607:             return Object.freeze({
1608:               kind: "manifest" as const,
1609:               body: snapshots.manifest(observedRequest.id, now, context),
1610:               context,
1611:             });
1612:           return Object.freeze({
1613:             kind: "page" as const,
1614:             body: snapshots.page(
1615:               observedRequest.id,
1616:               observedRequest.stream,
1617:               observedRequest.token,
1618:               now,
1619:               context,
1620:             ),
1621:             context,
1622:           });
1623:         } catch (error) {
1624:           if (!(error instanceof TransactionSnapshotFailure)) throw error;
1625:           return Object.freeze({
1626:             kind: "problem" as const,
1627:             code: error.code,
1628:             ...(error.retryAfter === undefined ? {} : { retryAfter: error.retryAfter }),
1629:             context,
1630:           });
1631:         }
1632:       });
1633:     },
1634:     /** Owner-only compatibility entry. HTTP uses the paired observation above;
1635:      * supplied retained-minimum controls remain explicitly unsupported. */

### packages/server/src/transaction-store.ts:1685–1720

1685:       const source = eventSource(sourceUrl);
1686:       const initialMinimumSource =
1687:         (after === undefined || admission === "current-event") &&
1688:         (source.kind === "scope" || source.kind === "bead" || source.kind === "link");
1689:       const minimum =
1690:         supplied === undefined || supplied === "unsupported"
1691:           ? supplied
1692:           : initialMinimumSource
1693:             ? snapshotCurrentReadMinimum(supplied)
1694:             : snapshotTransactionMinimum(supplied);
1695:       return enter(() => {
1696:         // Unknown/hidden sources are nondisclosing before initial Event minimum
1697:         // checks. A retained identity remains a source after Resource deletion.
1698:         if (!eventSourceKnown(source))
1699:           return Object.freeze({ kind: "missing" as const, context: readContext() });
1700:         if (minimum === "unsupported") return Object.freeze({ kind: "unsupported" as const });
1701:         const context = readContext(),
1702:           head = Number(context.scopePosition.slice(4));
1703:         // Eligible current Event minima precede retained group reads. Direct after
1704:         // calls without current-event intent retain their strict minimum check.
1705:         if (minimum !== undefined) {
1706:           if (initialMinimumSource) checkCurrentReadMinimum(minimum, context, true);
1707:           else checkTransactionMinimum(minimum, context);
1708:         }
1709:         const start = after === undefined ? undefined : parseTransactionEventCursor(after, context);
1710:         if (start) checkTransactionEventCursor(start, context, source.url);
1711:         let cursorMatched = start === undefined;
1712:         const events: import("./transaction-batch.js").TransactionEvent[] = [];
1713:         // Keep brackets/envelope once and replace the literal null with each
1714:         // candidate continuation string. Each Event is serialized exactly once.
1715:         const envelopeBytes =
1716:           Buffer.byteLength(json({ source: source.url, events: [], next: null })) - 4;
1717:         let eventBytes = 0;
1718:         let next: string | null = null;
1719:         const retainedGroupAtOrAfter = retainedGroupReader();
1720:         let at = start?.position ?? 1;

### packages/server/src/transaction-store.ts:1766–1836

1766:     /** Initial source/minimum/cursor observation, never a wire or durable cursor. */
1767:     beginEventScan(
1768:       sourceUrl: string,
1769:       after?: string,
1770:       supplied?: TransactionMinimum | "unsupported",
1771:       invalidReconnect = false,
1772:       admission?: "initial-scope-sse" | "current-event",
1773:     ):
1774:       | TransactionEventPacket
1775:       | { readonly kind: "missing"; readonly context: TransactionReadContext }
1776:       | { readonly kind: "unsupported" } {
1777:       if (
1778:         typeof sourceUrl !== "string" ||
1779:         (after !== undefined && typeof after !== "string") ||
1780:         typeof invalidReconnect !== "boolean"
1781:       )
1782:         throw new TypeError("invalid Event scan controls");
1783:       const source = eventSource(sourceUrl);
1784:       const initialScopeSseMinimum =
1785:         admission === "initial-scope-sse" &&
1786:         source.kind === "scope" &&
1787:         source.url === new URL("events/", scope).href &&
1788:         after === undefined &&
1789:         !invalidReconnect;
1790:       const currentEventMinimum =
1791:         admission === "current-event" &&
1792:         (source.kind === "scope" || source.kind === "bead" || source.kind === "link");
1793:       const minimum =
1794:         supplied === undefined || supplied === "unsupported"
1795:           ? supplied
1796:           : initialScopeSseMinimum || currentEventMinimum
1797:             ? snapshotCurrentReadMinimum(supplied)
1798:             : snapshotTransactionMinimum(supplied);
1799:       return enter(() => {
1800:         if (!eventSourceKnown(source))
1801:           return Object.freeze({ kind: "missing" as const, context: readContext() });
1802:         if (minimum === "unsupported") return Object.freeze({ kind: "unsupported" as const });
1803:         const context = readContext();
1804:         if (minimum !== undefined) {
1805:           if (initialScopeSseMinimum || currentEventMinimum)
1806:             checkCurrentReadMinimum(minimum, context, true);
1807:           else checkTransactionMinimum(minimum, context);
1808:         }
1809:         if (invalidReconnect) throw new TransactionEventParameterError(undefined, context);
1810:         const cursor =
1811:           after === undefined ? undefined : parseTransactionEventCursor(after, context);
1812:         if (cursor) checkTransactionEventCursor(cursor, context, source.url);
1813:         return eventPacket(
1814:           {
1815:             source,
1816:             context,
1817:             head: Number(context.scopePosition.slice(4)),
1818:             at: cursor?.position ?? 1,
1819:             ordinal: cursor?.ordinal ?? -1,
1820:           },
1821:           context,
1822:           cursor && after !== undefined ? { ...cursor, id: after } : undefined,
1823:         );
1824:       });
1825:     },
1826:     advanceEventScan(scan: TransactionEventScan): TransactionEventPacket {
1827:       return enter(() => {
1828:         const state = eventScans.get(scan);
1829:         if (!state) throw new TypeError("foreign or fabricated Event scan");
1830:         if (!eventSourceKnown(state.source)) throw new Error("retained Event source identity lost");
1831:         const context = readContext();
1832:         if (
1833:           context.scopeEpoch !== state.context.scopeEpoch ||
1834:           context.authorizationView !== state.context.authorizationView ||
1835:           Number(context.scopePosition.slice(4)) < state.head
1836:         )

### packages/server/src/transaction-snapshot.ts:20–40

20: import { TRANSACTION_RUNTIME_LIMITS } from "./transaction-batch.js";
21: import type { TransactionReadContext } from "./transaction-store.js";
22: 
23: export const TRANSACTION_SNAPSHOT_LIMITS = Object.freeze({
24:   lifetimeMs: 5 * 60 * 1000,
25:   resources: 4096,
26:   recordBytes: 32 * 1024 * 1024,
27:   liveHandles: 8,
28:   liveRecordBytes: 64 * 1024 * 1024,
29:   issuedHandles: 65536,
30:   dispositionBytes: 1024,
31:   snapshotPageItems: 32,
32:   pageBytes: 4 * 1024 * 1024,
33:   manifestBytes: 10 * 1024 * 1024,
34: });
35: export const TRANSACTION_SNAPSHOT_TOKEN_POLICY = "hmac-sha256-ordinal-1";
36: export const transactionSnapshotSchema = `
37: CREATE TABLE snapshots(id TEXT PRIMARY KEY, disposition TEXT NOT NULL) STRICT;
38: CREATE TABLE snapshot_records(snapshot TEXT NOT NULL, kind TEXT NOT NULL, ordinal INTEGER NOT NULL, id TEXT NOT NULL, body TEXT NOT NULL, PRIMARY KEY(snapshot,kind,ordinal), UNIQUE(snapshot,id)) STRICT;
39: CREATE TABLE snapshot_pages(snapshot TEXT NOT NULL, kind TEXT NOT NULL, ordinal INTEGER NOT NULL, token TEXT NOT NULL UNIQUE, start INTEGER NOT NULL, end INTEGER NOT NULL, bytes INTEGER NOT NULL, digest TEXT NOT NULL, PRIMARY KEY(snapshot,kind,ordinal)) STRICT;
40: `;

### packages/server/src/transaction-snapshot.ts:550–615

550:     return freeze(result);
551:   }
552:   function resolve(
553:     id: string,
554:     now: number,
555:     context: TransactionReadContext,
556:     kind?: Kind,
557:     token?: string,
558:   ) {
559:     if (typeof id !== "string" || !idPattern.test(id))
560:       throw new TransactionSnapshotFailure("resource-not-found");
561:     const d = disposition(id, context);
562:     if (!d) throw new TransactionSnapshotFailure("resource-not-found");
563:     // Recognize an actually issued continuation before expiry diagnosis. Retain
564:     // only bounded counts and a secret seed, not every expired page/range. The
565:     // authenticated ordinal is an opaque token component, not a client cursor.
566:     if (kind !== undefined) {
567:       const match =
568:         typeof token === "string" && token.length <= 52
569:           ? /^sp_(0|[1-9][0-9]{0,3})_([A-Za-z0-9_-]{43})$/.exec(token)
570:           : null;
571:       const ordinal = Number(match?.[1]);
572:       if (
573:         !match ||
574:         ordinal >= d.pageCounts[kind === "bead" ? 0 : 1] ||
575:         !timingSafeEqual(Buffer.from(token!), Buffer.from(pageToken(d.seed, id, kind, ordinal)))
576:       )
577:         throw new TransactionSnapshotFailure("resource-not-found");
578:     }
579:     // Qualified witnesses separate unexpired foreign-view and same-view expiry.
580:     // The simultaneous foreign+expired collision is not an admission claim.
581:     if (d.view !== context.authorizationView) throw new TransactionSnapshotFailure("foreign-view");
582:     if (d.expired || now > d.expires) throw new TransactionSnapshotFailure("cursor-expired");
583:     return { d, ranges: plan(id, d) };
584:   }
585: 
586:   function manifest(
587:     id: string,
588:     now: number,
589:     context: TransactionReadContext,
590:   ): TransactionSnapshotManifest {
591:     const { d, ranges } = resolve(id, now, context);
592:     return manifestValue(id, d, context, page(id, ranges, "bead"), page(id, ranges, "link"));
593:   }
594:   function manifestValue(
595:     id: string,
596:     d: Disposition,
597:     context: TransactionReadContext,
598:     beads: TransactionSnapshotPage,
599:     links: TransactionSnapshotPage,
600:   ): TransactionSnapshotManifest {
601:     const result: TransactionSnapshotManifest = {
602:       id: transactionSnapshotUrl(scope, id),
603:       scope,
604:       scopeEpoch: d.epoch,
605:       authorizationView: d.view,
606:       scopePosition: `pos-${d.position}`,
607:       checkpoint: transactionCheckpoint(scope, context, d.position),
608:       expiresAt: new Date(d.expires).toISOString(),
609:       erasures: [],
610:       beads,
611:       links,
612:     };
613:     integrity(bytes(text(result)) <= limits.manifestBytes, "serialized manifest bound");
614:     return freeze(result);
615:   }

### packages/server/src/transaction-snapshot.ts:795–813

795:     return manifestValue(id, d, context, firstPage("bead"), firstPage("link"));
796:   }
797:   return Object.freeze({
798:     expire,
799:     manifest,
800:     page(id: string, kind: Kind, token: string, now: number, context: TransactionReadContext) {
801:       if ((kind !== "bead" && kind !== "link") || typeof token !== "string")
802:         throw new TransactionSnapshotFailure("resource-not-found");
803:       const { ranges } = resolve(id, now, context, kind, token);
804:       return page(id, ranges, kind, token);
805:     },
806:     preflight(now: number, context: TransactionReadContext): void {
807:       capture(now, context, false);
808:     },
809:     capture(now: number, context: TransactionReadContext): TransactionSnapshotManifest {
810:       return capture(now, context, true);
811:     },
812:   });
813: }

### packages/server/src/transaction-events.ts:40–46

40: // extra bytes for suffixes, cursor and envelope. Scope source is smaller.
41: export const TRANSACTION_EVENT_FIRST_PAGE_BYTES = 5 * R + N + A + 1024;
42: if (TRANSACTION_EVENT_FIRST_PAGE_BYTES > TRANSACTION_EVENT_PAGE_BYTES)
43:   throw new Error("transaction tuple cannot support the finite Event page budget");
44: export class TransactionEventProjectionGap extends Error {}
45: export class TransactionEventParameterError extends Error {
46:   constructor(

### packages/server/src/transaction-events.ts:188–273

188: }
189: 
190: /** Uses only this group's final upserts/creation/incidence facts, never live Links. */
191: export function projectTransactionEventGroup(
192:   value: unknown,
193:   context: TransactionReadContext,
194:   source: TransactionEventSource,
195:   scope: string,
196:   expectedPosition: number,
197:   visible = true,
198: ): readonly TransactionEvent[] {
199:   if (!value || typeof value !== "object") throw new Error("invalid retained Event group");
200:   const group = value as TransactionEventGroup;
201:   if (
202:     group.scopeEpoch !== context.scopeEpoch ||
203:     group.authorizationView !== (visible ? context.authorizationView : INITIAL_TRANSACTION_VIEW) ||
204:     group.position !== `pos-${expectedPosition}` ||
205:     group.checkpoint !== group.position ||
206:     !Array.isArray(group.events) ||
207:     group.eventCount !== group.events.length ||
208:     !Array.isArray(group.changes) ||
209:     typeof group.transaction !== "string"
210:   )
211:     throw new Error("retained Event group context mismatch");
212:   const endpoints = new Map<string, readonly [unknown, unknown]>();
213:   const linkIds = new Set<string>();
214:   const remember = (id: string, from: unknown, to: unknown) => {
215:     reference(from);
216:     reference(to);
217:     const old = endpoints.get(id),
218:       next = [from, to] as const;
219:     if (old && stringifyJsonValue(old) !== stringifyJsonValue(next))
220:       throw new Error("inconsistent retained Link endpoints");
221:     endpoints.set(id, next);
222:   };
223:   for (const change of group.changes) {
224:     if (change.resourceKind === "link") {
225:       linkIds.add(change.resource.id);
226:       if (change.operation === "upsert" && "source" in change.resource)
227:         remember(change.resource.id, change.resource.source, change.resource.target);
228:     }
229:   }
230:   for (const event of group.events) {
231:     const data = event.data as Record<string, unknown>;
232:     if (!data || typeof data !== "object") throw new Error("invalid retained Event data");
233:     if (
234:       event.type === "linked" ||
235:       event.type === "unlinked" ||
236:       (event.type === "created" && linkIds.has(event.subject))
237:     )
238:       remember(event.subject, data.source, data.target);
239:   }
240:   const result: TransactionEvent[] = [];
241:   for (const [index, event] of group.events.entries()) {
242:     if (
243:       event.ordinal !== index ||
244:       event.source !== `${scope}events/` ||
245:       event.transaction !== group.transaction ||
246:       !["created", "updated", "deleted", "linked", "unlinked"].includes(event.type)
247:     )
248:       throw new Error("invalid retained Event framing");
249:     const data = event.data as Record<string, unknown>;
250:     let include = source.kind === "scope";
251:     if (visible && source.kind !== "scope") {
252:       include =
253:         event.subject === source.subject && ["created", "updated", "deleted"].includes(event.type);
254:       if (source.kind === "bead" && (event.type === "linked" || event.type === "unlinked")) {
255:         if (data.endpoint !== "source" && data.endpoint !== "target")
256:           throw new Error("invalid endpoint fact");
257:         include = reference(data[data.endpoint]) === source.subject;
258:       }
259:       if (source.kind === "bead" && event.type === "updated" && linkIds.has(event.subject)) {
260:         const pair = endpoints.get(event.subject);
261:         if (!pair) throw new Error("incident Link update lacks same-group endpoints");
262:         include = pair.some((endpoint) => reference(endpoint) === source.subject);
263:         // Literal incident rule currently mentions properties only. Do not silently
264:         // decide metadata-only membership while the spec disposition is pending.
265:         if (include && data.change === undefined && data.metadataChange !== undefined)
266:           throw new TransactionEventProjectionGap("incident-link-metadata");
267:       }
268:     }
269:     if (visible && include)
270:       result.push(renderTransactionEvent(context, source.url, group.checkpoint, event));
271:   }
272:   return result;
273: }

### packages/server/src/transaction-batch.ts:739–787

739:     const outcome = result.outcome;
740:     results.push({ operationIndex, ...(op.name ? { operationName: op.name } : {}), ...outcome });
741:     const resource = outcome.resource;
742:     if (op.name && outcome.outcome === "created" && resource) bindings.set(op.name, resource.id);
743:     // A retained revision is an observed no-op, not a committed transition.
744:     if (result.changed.length === 0 && result.deleted.length === 0) return;
745:     const subject = resource ?? before;
746:     if (!subject) throw new Error("successful transition lacks Resource state");
747:     // Evaluator IDs are local lookup paths; group identity is always canonical.
748:     // One owned operation changes at most the Link and its source Bead.
749:     for (const id of result.changed) {
750:       const row = tx.resource(id);
751:       if (!row) throw new Error("changed Resource lacks staged postimage");
752:       const record = readStoredResource(row, options.scope);
753:       changes.set(record.id, {
754:         operation: "upsert",
755:         resourceKind: "source" in record ? "link" : "bead",
756:         resource: record,
757:       });
758:     }
759:     for (const id of result.deleted) {
760:       if (
761:         !before ||
762:         !outcome.deleted ||
763:         before.id !== `${options.scope}${id}` ||
764:         outcome.deleted.resource.id !== before.id ||
765:         outcome.deleted.resource.type !== before.type ||
766:         outcome.deleted.resource.revision !== before.revision
767:       )
768:         throw new Error("deleted Resource lacks matching previous identity");
769:       changes.set(before.id, { operation: "tombstone", ...outcome.deleted });
770:     }
771:     if (outcome.outcome === "created") {
772:       scopeEvent(subject, "created", {
773:         revision: subject.revision,
774:         properties: subject.properties,
775:         metadata: subject.metadata ?? {},
776:         ...(subject.revisionProvenance ? { revisionProvenance: subject.revisionProvenance } : {}),
777:         ...("source" in subject ? { source: subject.source, target: subject.target } : {}),
778:       });
779:     } else if (outcome.outcome === "updated") {
780:       if (!before) throw new Error("update lacks previous state");
781:       scopeEvent(subject, "updated", {
782:         previousRevision: before.revision,
783:         revision: subject.revision,
784:         ...(input.propertiesChange !== undefined ? { change: input.propertiesChange } : {}),
785:         ...(input.metadataChange !== undefined ? { metadataChange: input.metadataChange } : {}),
786:         ...(subject.revisionProvenance ? { revisionProvenance: subject.revisionProvenance } : {}),
787:       });

### packages/server/src/transaction-batch.ts:805–849

805:     if (outcome.sourceRevision !== undefined) {
806:       if (!("source" in subject) || !outcome.source?.startsWith(options.scope))
807:         throw new Error("owned transition lacks canonical Link/source identity");
808:       const localSource = outcome.source.slice(options.scope.length);
809:       const oldSource = previous.get(localSource),
810:         row = tx.resource(localSource);
811:       const source = row ? readStoredResource(row, options.scope) : undefined;
812:       if (
813:         !oldSource ||
814:         !source ||
815:         "source" in source ||
816:         "source" in oldSource ||
817:         source.id !== outcome.source ||
818:         source.revision !== outcome.sourceRevision
819:       )
820:         throw new Error("owned transition lacks matching source versions");
821:       let link: unknown;
822:       if (outcome.outcome === "created") link = subject;
823:       else if (outcome.outcome === "deleted")
824:         link = { id: subject.id, type: subject.type, revision: subject.revision };
825:       else {
826:         if (!before || (input.propertiesChange === undefined && input.metadataChange === undefined))
827:           throw new Error("owned update lacks committed patch");
828:         link = {
829:           id: subject.id,
830:           type: subject.type,
831:           previousRevision: before.revision,
832:           revision: subject.revision,
833:           ...(input.propertiesChange !== undefined ? { change: input.propertiesChange } : {}),
834:           ...(input.metadataChange !== undefined ? { metadataChange: input.metadataChange } : {}),
835:           ...(subject.revisionProvenance ? { revisionProvenance: subject.revisionProvenance } : {}),
836:         };
837:       }
838:       // Up to four facts: Link lifecycle, source/target incidence, owning Bead.
839:       // The source Event is last and never carries a direct Resource patch.
840:       scopeEvent(source, "updated", {
841:         previousRevision: oldSource.revision,
842:         revision: source.revision,
843:         ownedLink: { operation: outcome.outcome, link },
844:         ...(source.revisionProvenance ? { revisionProvenance: source.revisionProvenance } : {}),
845:       });
846:     }
847:     return undefined;
848:   }
849:   try {

### packages/server/src/transaction-runtime.ts:130–159

130:         !raw.startsWith("/") ||
131:         raw.startsWith("//") ||
132:         request.headers.host !==
133:           new URL(captured.authorityMode === "canonical" ? scope : listenUrl).host
134:       ) {
135:         deliver(response, transactionHttpProblem("malformed-request"));
136:         return;
137:       }
138:       const url = new URL(raw, scope);
139:       const methodFailure = handler.methodResponse(url, request.method ?? "GET");
140:       if (methodFailure) {
141:         // Refused bodies are not consumed and cannot keep this connection alive.
142:         methodFailure.headers.set("connection", "close");
143:         deliver(response, methodFailure);
144:         return;
145:       }
146:       const headers = new Headers();
147:       for (let i = 0; i < request.rawHeaders.length; i += 2)
148:         headers.append(request.rawHeaders[i] ?? "", request.rawHeaders[i + 1] ?? "");
149:       const supplied = Buffer.from(headers.get("authorization") ?? "");
150:       const authorization = captured.authenticate(supplied);
151:       signal = createNodeHttpDisconnectSignal(request, response);
152:       const input = new Request(url, { method: request.method ?? "GET", headers, signal });
153:       const lastEventIds: string[] = [];
154:       for (let i = 0; i < request.rawHeaders.length; i += 2)
155:         if (request.rawHeaders[i]?.toLowerCase() === "last-event-id")
156:           lastEventIds.push(request.rawHeaders[i + 1] ?? "");
157:       const body = request.method === "POST" ? await receive(request) : { text: "" };
158:       if (closing || response.destroyed) {
159:         deliver(response, transactionHttpProblem("temporarily-unavailable"));

### packages/server/src/transaction-runtime.ts:306–346

306:   server.on("timeout", (socket) => socket.destroy());
307:   const closeListener = () =>
308:     new Promise<void>((resolve, reject) => {
309:       server.close((error) => (error ? reject(error) : resolve()));
310:       server.closeAllConnections();
311:     });
312:   let owner: ReturnType<typeof openTransactionStore> | undefined;
313:   try {
314:     await new Promise<void>((resolve, reject) => {
315:       server.once("error", reject);
316:       server.listen(captured.port, settings.transport.host, () => {
317:         server.removeListener("error", reject);
318:         resolve();
319:       });
320:     });
321:     const address = server.address();
322:     if (!address || typeof address === "string")
323:       throw new Error("missing private transaction address");
324:     listenUrl = `http://127.0.0.1:${address.port}/`;
325:     scope = configuredScope ?? new URL("acme/", listenUrl).href;
326:     owner = openTransactionStore({
327:       file: path.join(captured.directory, "transaction.sqlite"),
328:       scope,
329:       create: captured.create,
330:       types,
331:       ...(contracts ? { contracts } : {}),
332:       receiptMaximumItems: settings.receiptMaximumItems,
333:       maintenanceIntervalMs: settings.maintenanceIntervalMs,
334:       ...(capturedSeams.clock ? { clock: capturedSeams.clock } : {}),
335:     });
336:     service = createTransactionService(owner, {
337:       ownerLifetime: "service",
338:       maximumStreams: settings.changefeedStream.maximumStreams,
339:       ...(capturedSeams.scheduler ? { scheduler: capturedSeams.scheduler } : {}),
340:       comparison: { budgetMs: settings.comparisonBudgetMs, ...capturedSeams.comparison },
341:       minimumReads: { budgetMs: settings.minimumReadBudgetMs, ...capturedSeams.minimumReads },
342:       read: captured.readOptions(scope),
343:     });
344:     configurationMetadata = captured.metadata(owner.readConfiguration(), scope);
345:     schemas = createTransactionSchemaHttp(scope, contracts, service.schemaObserved);
346:     handler = createTransactionHttpHandler(service, scope, capturedSeams.maximumOperations);

### packages/server/src/transaction-snapshot.test.ts:1219–1288

1219: it("checks each fresh Snapshot minimum before any handle, record, page or expiry work", () => {
1220:   const f = fixture();
1221:   f.run([create("a")]);
1222:   const context = f.owner.observeSnapshot({ kind: "preflight" }).context;
1223:   const minimum = { ...context, minimumPosition: "pos-2" };
1224:   const tuple = {
1225:     scopeEpoch: minimum.scopeEpoch,
1226:     authorizationView: minimum.authorizationView,
1227:     minimumPosition: minimum.minimumPosition,
1228:   };
1229:   for (const kind of ["create", "preflight"] as const) {
1230:     for (let attempt = 0; attempt < 12; attempt++)
1231:       expect(() => f.owner.observeSnapshot({ kind }, tuple)).toThrow("behind");
1232:     expect(() => f.owner.observeSnapshot({ kind }, { ...tuple, scopeEpoch: "foreign" })).toThrow(
1233:       "cursor-expired",
1234:     );
1235:     expect(() =>
1236:       f.owner.observeSnapshot({ kind }, { ...tuple, authorizationView: "foreign" }),
1237:     ).toThrow("foreign-view");
1238:     let unknown: unknown;
1239:     try {
1240:       f.owner.observeSnapshot({ kind }, { ...tuple, minimumPosition: "unknown" });
1241:     } catch (error) {
1242:       unknown = error;
1243:     }
1244:     expect(unknown).toBeInstanceOf(TransactionMinimumUnproved);
1245:     expect(unknown).toMatchObject({ context });
1246:   }
1247:   f.offline((db) => {
1248:     for (const table of ["snapshots", "snapshot_records", "snapshot_pages"])
1249:       expect(db.prepare(`SELECT count(*) AS n FROM ${table}`).get()?.n).toBe(0);
1250:   });
1251:   const captured = f.owner.observeSnapshot(
1252:     { kind: "create" },
1253:     { ...tuple, minimumPosition: "pos-1" },
1254:   );
1255:   expect(captured).toMatchObject({ kind: "manifest", context });
1256:   if (captured.kind !== "manifest") throw new Error("expected manifest");
1257:   expect(captured.body.beads.items.map((row) => row.id)).toEqual([`${scope}beads/a`]);
1258:   for (const request of [
1259:     { kind: "manifest", id: handle(captured.body) },
1260:     { kind: "page", id: handle(captured.body), stream: "bead", token: "unknown" },
1261:   ] as const)
1262:     expect(() => f.owner.observeSnapshot(request, tuple)).toThrow("requires fresh creation");
1263:   expect(f.owner.observeSnapshot({ kind: "manifest", id: handle(captured.body) })).toMatchObject({
1264:     kind: "manifest",
1265:     body: captured.body,
1266:     context,
1267:   });
1268: });
1269: 
1270: it("does not expire retained Snapshot details on an ineligible fresh attempt", () => {
1271:   const f = fixture();
1272:   f.run([create("a")]);
1273:   const first = f.owner.observeSnapshot({ kind: "create" });
1274:   if (first.kind !== "manifest") throw new Error("expected manifest");
1275:   f.setTime(1000000 + limits.lifetimeMs + 1);
1276:   const minimum = {
1277:     scopeEpoch: first.context.scopeEpoch,
1278:     authorizationView: first.context.authorizationView,
1279:     minimumPosition: "pos-2",
1280:   };
1281:   for (const kind of ["create", "preflight"] as const)
1282:     expect(() => f.owner.observeSnapshot({ kind }, minimum)).toThrow("behind");
1283:   f.offline((db) => {
1284:     expect(db.prepare("SELECT count(*) AS n FROM snapshot_records").get()?.n).toBe(1);
1285:     expect(
1286:       JSON.parse(String(db.prepare("SELECT disposition FROM snapshots").get()?.disposition))
1287:         .expired,
1288:     ).toBe(false);

### packages/server/src/transaction-snapshot.test.ts:1446–1502

1446:     pages: Number(db.prepare("SELECT count(*) AS n FROM snapshot_pages").get()?.n),
1447:     disposition: String(db.prepare("SELECT disposition FROM snapshots").get()?.disposition),
1448:   });
1449:   let baseline: ReturnType<typeof inventory> | undefined;
1450:   let after: ReturnType<typeof inventory> | undefined;
1451:   let offlineInspections = 0;
1452:   f.offline((db) => {
1453:     offlineInspections++;
1454:     baseline = inventory(db);
1455:     expect(baseline).toMatchObject({ snapshots: 1, records: 1, pages: 2 });
1456:     expect(JSON.parse(baseline.disposition).expired).toBe(false);
1457:   });
1458:   f.setTime(1000000 + limits.lifetimeMs + 1);
1459:   const minimum = {
1460:     scopeEpoch: first.context.scopeEpoch,
1461:     authorizationView: first.context.authorizationView,
1462:     minimumPosition: "opaque",
1463:   };
1464:   const unproved = (kind: "create" | "preflight", minimumPosition = "opaque") => {
1465:     let failure: unknown;
1466:     try {
1467:       observe({ kind }, { ...minimum, minimumPosition });
1468:     } catch (error) {
1469:       failure = error;
1470:     }
1471:     expect(failure).toBeInstanceOf(TransactionMinimumUnproved);
1472:     expect(failure).toMatchObject({ context: first.context });
1473:   };
1474:   for (const kind of ["create", "preflight"] as const) {
1475:     unproved(kind);
1476:     expect(() => observe({ kind }, { ...minimum, scopeEpoch: "foreign" })).toThrow(
1477:       expect.objectContaining({ reason: "cursor-expired", context: first.context }),
1478:     );
1479:     expect(() => observe({ kind }, { ...minimum, authorizationView: "foreign" })).toThrow(
1480:       expect.objectContaining({ reason: "foreign-view", context: first.context }),
1481:     );
1482:     expect(() => observe({ kind }, { ...minimum, minimumPosition: "pos-2" })).toThrow("behind");
1483:   }
1484:   unproved("create", "pos-01");
1485:   let preEntryExcludedCalls = 0;
1486:   for (const request of [
1487:     { kind: "manifest", id: handle(first.body) },
1488:     { kind: "page", id: handle(first.body), stream: "bead", token: "unused" },
1489:   ] as const) {
1490:     preEntryExcludedCalls++;
1491:     expect(() => observe(request, minimum)).toThrow("unsupported minimum tuple");
1492:     preEntryExcludedCalls++;
1493:     expect(() => observe(request, { ...minimum, minimumPosition: "pos-1" })).toThrow(
1494:       "requires fresh creation",
1495:     );
1496:   }
1497:   // No await: normal background maintenance must not race the advanced clock.
1498:   f.offline((db) => {
1499:     offlineInspections++;
1500:     after = inventory(db);
1501:     expect(after).toEqual(baseline);
1502:     expect(JSON.parse(after.disposition).expired).toBe(false);

### packages/server/src/transaction-snapshot-http.test.ts:73–169

73: async function fixture(seamless = false) {
74:   const directory = mkdtempSync(path.join(tmpdir(), "bdp-snapshot-http-"));
75:   cleanups.push(() => rmSync(directory, { recursive: true, force: true }));
76:   let now = 1000000,
77:     key = 0;
78:   const options: TransactionRuntimeOptions = {
79:     directory,
80:     create: true,
81:     port: 0,
82:     scope,
83:     authorityMode: "listener",
84:     credentials: [
85:       { token, principal: "writer", canMutate: true },
86:       { token: reader, principal: "different", canMutate: false },
87:     ],
88:     ...(seamless
89:       ? {
90:           artifacts: {
91:             descriptors: types.map((type) => ({
92:               retrievalUri: type.id,
93:               utf8: JSON.stringify(type),
94:             })),
95:             schemas: [],
96:             limits: { ...CONTRACT_ARTIFACT_CEILINGS },
97:           },
98:         }
99:       : { types }),
100:   };
101:   const seams = seamless ? {} : { clock: () => now, maintenanceIntervalMs: 2147483647 };
102:   let runtime = await startTransactionRuntime(options, seams);
103:   cleanups.push(() => runtime.close());
104:   const file = path.join(directory, "transaction.sqlite");
105:   async function get(url = `${scope}snapshot`, init: RequestInit = {}) {
106:     const canonical = new URL(url),
107:       dial = new URL(canonical.pathname + canonical.search, runtime.listenUrl);
108:     return fetch(dial, {
109:       ...init,
110:       headers: {
111:         authorization: `Bearer ${token}`,
112:         ...Object.fromEntries(new Headers(init.headers)),
113:       },
114:       redirect: "manual",
115:     });
116:   }
117:   async function offline<T>(run: (db: DatabaseSync) => T): Promise<T> {
118:     await runtime.close();
119:     const db = new DatabaseSync(file);
120:     try {
121:       return run(db);
122:     } finally {
123:       db.close();
124:       runtime = await startTransactionRuntime({ ...options, create: false }, seams);
125:     }
126:   }
127:   return {
128:     get runtime() {
129:       return runtime;
130:     },
131:     options,
132:     get,
133:     offline,
134:     setTime(time: number) {
135:       now = time;
136:     },
137:     async reopen() {
138:       await runtime.close();
139:       runtime = await startTransactionRuntime({ ...options, create: false }, seams);
140:     },
141:     async state() {
142:       return offline((db) => ({
143:         snapshots: db.prepare("SELECT * FROM snapshots ORDER BY id").all(),
144:         records: db.prepare("SELECT * FROM snapshot_records ORDER BY snapshot,kind,ordinal").all(),
145:         pages: db.prepare("SELECT * FROM snapshot_pages ORDER BY snapshot,kind,ordinal").all(),
146:         position: db.prepare("SELECT value FROM metadata WHERE name='position'").get(),
147:         receipts: db.prepare("SELECT * FROM receipts ORDER BY key").all(),
148:         groups: db.prepare("SELECT * FROM groups ORDER BY position").all(),
149:       }));
150:     },
151:     async write(operations: unknown[]) {
152:       const response = await get(runtime.targets.batch, {
153:         method: "POST",
154:         headers: { "content-type": "application/json", "idempotency-key": `write_${key++}` },
155:         body: JSON.stringify({ operations }),
156:       });
157:       expect(response.status).toBe(200);
158:       const result = (await response.json()) as { status: string };
159:       expect(result.status).toBe("completed");
160:       return result;
161:     },
162:     async capture() {
163:       const response = await get();
164:       expect(response.status).toBe(200);
165:       const value = (await response.json()) as TransactionSnapshotManifest;
166:       expect(manifestSchema(value), JSON.stringify(manifestSchema.errors)).toBe(true);
167:       return value;
168:     },
169:   };

### packages/server/src/transaction-snapshot-http.test.ts:326–378

326: it("retained HEAD uses UTF-8 bytes, conditional decisions carry current context, and expiry precedes shortcuts", async () => {
327:   const f = await fixture();
328:   await f.write(Array.from({ length: 32 }, (_, n) => create(`b${n}`)));
329:   await f.write([create("tail")]);
330:   const snapshot = await f.capture();
331:   expect(snapshot.beads.next).not.toBeNull();
332:   const other = await f.capture();
333:   const wrongKind = new URL(present(snapshot.beads.next));
334:   wrongKind.searchParams.set("stream", "links");
335:   await checkedProblem(await f.get(wrongKind.href), 404, "resource-not-found");
336:   const borrowed = new URL(present(snapshot.beads.next));
337:   borrowed.searchParams.set("snapshot", present(new URL(other.id).searchParams.get("snapshot")));
338:   await checkedProblem(await f.get(borrowed.href), 404, "resource-not-found");
339:   for (const url of [snapshot.id, present(snapshot.beads.next)]) {
340:     const get = await f.get(url),
341:       text = await get.text();
342:     expect(Buffer.byteLength(text)).toBeGreaterThan(text.length);
343:     const head = await f.get(url, { method: "HEAD" });
344:     expect(head.status).toBe(200);
345:     expect(await head.text()).toBe("");
346:     expect(head.headers.get("content-length")).toBe(String(Buffer.byteLength(text)));
347:     expect(context(head)).toEqual(context(get));
348:     for (const [header, value, status] of [
349:       ["if-none-match", "*", 304],
350:       ["if-match", '"other"', 412],
351:     ] as const) {
352:       const conditional = await f.get(url, { headers: { [header]: value } });
353:       expect(conditional.status).toBe(status);
354:       expect(context(conditional)).toEqual(context(get));
355:       expect(await conditional.text()).toBe("");
356:       if (status === 304) expect(conditional.headers.get("content-length")).toBeNull();
357:     }
358:   }
359:   f.setTime(1300000);
360:   expect((await f.get(snapshot.id)).status).toBe(200);
361:   f.setTime(1300001);
362:   for (const url of [snapshot.id, present(snapshot.beads.next)]) {
363:     await checkedProblem(
364:       await f.get(url, { headers: { "if-none-match": "*" } }),
365:       410,
366:       "cursor-expired",
367:     );
368:     const head = await f.get(url, { method: "HEAD" });
369:     expect(head.status).toBe(410);
370:     expect(await head.text()).toBe("");
371:   }
372:   const wrong = new URL(present(snapshot.beads.next));
373:   wrong.searchParams.set("page", `sp_1_${"a".repeat(43)}`);
374:   await checkedProblem(await f.get(wrong.href), 404, "resource-not-found");
375:   wrong.searchParams.set("stream", "links");
376:   await checkedProblem(await f.get(wrong.href), 404, "resource-not-found");
377:   await f.reopen();
378:   await checkedProblem(await f.get(present(snapshot.beads.next)), 410, "cursor-expired");

### packages/server/src/transaction-snapshot-http.test.ts:380–438

380:   expect(state.snapshots).toHaveLength(2);
381:   expect(state.records).toHaveLength(0);
382:   expect(state.pages).toHaveLength(0);
383: });
384: 
385: it("closed query syntax, unsupported controls and foreign creation context fail before capture or conditional success", async () => {
386:   const f = await fixture(),
387:     before = await f.state();
388:   for (const query of [
389:     "?view=events",
390:     "?snapshot=x&snapshot=y",
391:     "?snapshot=x&page=x",
392:     "?snapshot=x&stream=wrong&page=x",
393:     "?unknown=x",
394:     "?snapshot=",
395:   ])
396:     await checkedProblem(await f.get(`${scope}snapshot${query}`), 400, "invalid-parameter");
397:   for (const headers of [
398:     { "bdp-scope-epoch": "epoch" },
399:     { "bdp-authorization-view": "view" },
400:     { "bdp-minimum-scope-position": "pos-0" },
401:     {
402:       "bdp-scope-epoch": "epoch",
403:       "bdp-authorization-view": "view",
404:       "bdp-minimum-scope-position": "pos-0",
405:     },
406:   ])
407:     for (const url of [
408:       `${scope}snapshot`,
409:       `${scope}snapshot?snapshot=unknown`,
410:       `${scope}snapshot?snapshot=unknown&stream=beads&page=unknown`,
411:     ])
412:       for (const method of ["GET", "HEAD"]) {
413:         const response = await f.get(url, {
414:           method,
415:           headers: { ...headers, "if-none-match": "*" },
416:         });
417:         if (url === `${scope}snapshot` && Object.keys(headers).length === 3) {
418:           expect(response.status).toBe(410);
419:           expect(response.headers.get("bdp-scope-position")).toBe("pos-0");
420:           expect(response.headers.get("x-bdp-development-limit")).toBeNull();
421:           if (method === "HEAD") expect(await response.text()).toBe("");
422:           else await checkedProblem(response, 410, "cursor-expired");
423:         } else if (Object.keys(headers).length < 3) {
424:           expect(response.status).toBe(400);
425:           expect(response.headers.get("x-bdp-development-limit")).toBeNull();
426:           expect(response.headers.get("bdp-scope-position")).toBeNull();
427:           if (method === "HEAD") expect(await response.text()).toBe("");
428:           else await checkedProblem(response, 400, "malformed-request");
429:         } else {
430:           expect(response.status).toBe(500);
431:           expect(await response.text()).toBe("");
432:           expect(response.headers.get("x-bdp-development-limit")).toBe("minimum-position");
433:           expect(response.headers.get("bdp-scope-position")).toBeNull();
434:         }
435:       }
436:   expect(await f.state()).toEqual(before);
437:   await checkedProblem(await f.get(`${scope}snapshot?snapshot=unknown`), 404, "resource-not-found");
438: });

### packages/server/src/transaction-service.test.ts:1102–1243

1102: describe("fresh Snapshot minimum service", () => {
1103:   it.each(["create", "preflight"] as const)(
1104:     "freezes %s controls and retries after a real independent commit",
1105:     async (kind) => {
1106:       const f = fixture(),
1107:         entered = Promise.withResolvers<void>(),
1108:         release = Promise.withResolvers<void>();
1109:       const service = f.open({
1110:         minimumReads: {
1111:           wait: async () => {
1112:             entered.resolve();
1113:             await release.promise;
1114:           },
1115:         },
1116:       });
1117:       const initial = await service.snapshot({ kind: "preflight" });
1118:       const tuple = {
1119:         scopeEpoch: initial.context.scopeEpoch,
1120:         authorizationView: initial.context.authorizationView,
1121:         minimumPosition: "pos-1",
1122:       };
1123:       const request: TransactionSnapshotRequest = { kind };
1124:       const pending = service.snapshot(request, tuple);
1125:       try {
1126:         await entered.promise;
1127:         Object.assign(request, { kind: "manifest", id: "not-a-handle" });
1128:         tuple.minimumPosition = "pos-999";
1129:         requireReceipt(await service.submit(batch([create("after-wait")])));
1130:         let closed = false;
1131:         const closing = service.close().then(() => {
1132:           closed = true;
1133:         });
1134:         await Promise.resolve();
1135:         expect(closed).toBe(false);
1136:         release.resolve();
1137:         const result = await pending;
1138:         expect(result.context).toEqual({ ...initial.context, scopePosition: "pos-1" });
1139:         expect(result.kind).toBe(kind === "create" ? "manifest" : "eligible");
1140:         if (result.kind === "manifest") {
1141:           expect(result.body.scopePosition).toBe("pos-1");
1142:           expect(result.body.checkpoint).not.toBe("genesis");
1143:           expect(result.body.beads.items).toHaveLength(1);
1144:           expect(result.body.beads.items[0]).toMatchObject({
1145:             id: `${scope}beads/after-wait`,
1146:             properties: { title: "after-wait" },
1147:           });
1148:         }
1149:         await closing;
1150:       } finally {
1151:         release.resolve();
1152:         await Promise.allSettled([pending]);
1153:       }
1154:     },
1155:   );
1156:   it.each(["create", "preflight"] as const)(
1157:     "keeps %s timeout and pre-abort distinct from owner work",
1158:     async (kind) => {
1159:       const f = fixture();
1160:       let tick = 0;
1161:       const wait = vi.fn(async () => {});
1162:       const observeSnapshot = vi.fn(f.owner.observeSnapshot);
1163:       const service = createTransactionService(
1164:         { ...f.owner, observeSnapshot },
1165:         {
1166:           ownerLifetime: "service",
1167:           read: f.read,
1168:           minimumReads: { budgetMs: 2, now: () => tick++, wait },
1169:         },
1170:       );
1171:       cleanup.push(() => service.close());
1172:       const initial = await service.snapshot({ kind: "preflight" });
1173:       observeSnapshot.mockClear();
1174:       const tuple = {
1175:         scopeEpoch: initial.context.scopeEpoch,
1176:         authorizationView: initial.context.authorizationView,
1177:         minimumPosition: "pos-1",
1178:       };
1179:       await expect(service.snapshot({ kind }, tuple, AbortSignal.abort())).rejects.toMatchObject({
1180:         name: "ScopeServerOperationAbortedError",
1181:       });
1182:       expect(wait).not.toHaveBeenCalled();
1183:       expect(observeSnapshot).not.toHaveBeenCalled();
1184:       await expect(service.snapshot({ kind }, tuple)).rejects.toMatchObject({
1185:         reason: "catch-up-timeout",
1186:         context: initial.context,
1187:       });
1188:       expect(wait).toHaveBeenCalled();
1189:       expect(() => service.snapshot({ kind: "manifest", id: "unknown" }, tuple)).toThrow(
1190:         "requires fresh creation",
1191:       );
1192:       expect(
1193:         await service.snapshot({ kind: "create" }, { ...tuple, minimumPosition: "genesis" }),
1194:       ).toMatchObject({
1195:         kind: "manifest",
1196:         context: initial.context,
1197:         body: { scopePosition: "pos-0", checkpoint: "genesis" },
1198:       });
1199:     },
1200:   );
1201:   it("uses one service call for each fresh HTTP branch, with terminal preflight", async () => {
1202:     const f = fixture();
1203:     let tick = 0;
1204:     const service = f.open({
1205:       minimumReads: { budgetMs: 2, now: () => tick++, wait: async () => {} },
1206:     });
1207:     const initial = await service.snapshot({ kind: "preflight" });
1208:     const snapshot = vi.fn(service.snapshot);
1209:     const http = createTransactionHttpHandler({ ...service, snapshot }, scope);
1210:     for (const [method, headers, status, kind] of [
1211:       ["GET", {}, 200, "create"],
1212:       ["HEAD", {}, 200, "preflight"],
1213:       ["GET", { "if-none-match": "*" }, 304, "preflight"],
1214:       ["GET", { "if-match": '"absent"' }, 412, "preflight"],
1215:       ["GET", { accept: "text/plain" }, 406, "preflight"],
1216:     ] as const) {
1217:       for (const minimumPosition of ["genesis", "pos-1", "unknown"]) {
1218:         snapshot.mockClear();
1219:         const response = await http.handle(
1220:           new Request(`${scope}snapshot`, {
1221:             method,
1222:             headers: {
1223:               ...headers,
1224:               "bdp-scope-epoch": initial.context.scopeEpoch,
1225:               "bdp-authorization-view": initial.context.authorizationView,
1226:               "bdp-minimum-scope-position": minimumPosition,
1227:             },
1228:           }),
1229:           { kind: "authenticated", principal: "p", canMutate: true },
1230:         );
1231:         expect(response.status).toBe(minimumPosition === "genesis" ? status : 503);
1232:         expect(response.headers.get("bdp-scope-position")).toBe("pos-0");
1233:         expect(snapshot).toHaveBeenCalledTimes(1);
1234:         expect(snapshot.mock.calls[0]?.[0]).toEqual({ kind });
1235:       }
1236:     }
1237:   });
1238: });
1239: 
1240: describe("internal discovery owner observation", () => {
1241:   it("copies exactly one owner entry and parses only after releasing it", async () => {
1242:     const f = fixture();
1243:     let entries = 0;

### packages/server/src/transaction-events.test.ts:145–215

145: it.each([false, true])(
146:   "projects Scope, Link and Bead facts through deletion/restart (owned=%s)",
147:   (owned) => {
148:     const f = fixture(scope, owned);
149:     f.run([bead("a"), bead("b"), link()]);
150:     f.run([patch, remove]);
151:     const global = collect(f.owner, `${scope}events/`),
152:       a = collect(f.owner, `${scope}beads/a?view=events`),
153:       l = collect(f.owner, `${scope}links/l?view=events`);
154:     expect(l.events.map((e) => e.type)).toEqual(["created", "updated", "deleted"]);
155:     expect(a.events.map((e) => [e.type, e.subject])).toEqual(
156:       owned
157:         ? [
158:             ["created", `${scope}beads/a`],
159:             ["linked", `${scope}links/l`],
160:             ["updated", `${scope}beads/a`],
161:             ["updated", `${scope}links/l`],
162:             ["updated", `${scope}beads/a`],
163:             ["unlinked", `${scope}links/l`],
164:             ["updated", `${scope}beads/a`],
165:           ]
166:         : [
167:             ["created", `${scope}beads/a`],
168:             ["linked", `${scope}links/l`],
169:             ["updated", `${scope}links/l`],
170:             ["unlinked", `${scope}links/l`],
171:           ],
172:     );
173:     expect(l.events[1]?.data).toMatchObject({ change: patch.propertiesChange });
174:     for (const event of a.events) {
175:       const original = global.events.find(
176:         (e) => e.transaction === event.transaction && e.ordinal === event.ordinal,
177:       );
178:       expect(original).toBeDefined();
179:       expect(event.data).toEqual(original?.data);
180:       expect(event.id).not.toBe(original?.id);
181:       expect(event.id).toMatch(/^[A-Za-z0-9_-]{1,256}$/);
182:     }
183:     f.reopen();
184:     expect(collect(f.owner, `${scope}beads/a?view=events`)).toEqual(a);
185:     expect(collect(f.owner, `${scope}links/l?view=events`)).toEqual(l);
186:   },
187: );
188: it.each([false, true])(
189:   "uses same-group endpoints for create/update/delete and self-Link incidences (owned=%s)",
190:   (owned) => {
191:     const f = fixture(scope, owned);
192:     f.run([bead("a")]);
193:     f.run([link("l", "beads/a", "beads/a"), patch, remove]);
194:     const events = collect(f.owner, `${scope}beads/a?view=events`).events;
195:     const incidence = events.filter((e) => e.type === "linked" || e.type === "unlinked");
196:     expect(incidence.map((e) => (e.data as { endpoint: string }).endpoint)).toEqual([
197:       "source",
198:       "target",
199:       "source",
200:       "target",
201:     ]);
202:     expect(
203:       events.filter((e) => e.type === "updated" && e.subject === `${scope}links/l`),
204:     ).toHaveLength(1);
205:     expect(f.owner.inspect().groups.at(-1)).toMatchObject({
206:       changes: expect.arrayContaining([
207:         {
208:           operation: "tombstone",
209:           resourceKind: "link",
210:           resource: expect.objectContaining({ id: `${scope}links/l` }),
211:         },
212:       ]),
213:     });
214:   },
215: );

### packages/server/src/transaction-events.test.ts:316–372

316: it("does not emit failed/no-op transactions and does not expire Events with receipt detail", () => {
317:   const f = fixture();
318:   const committed = f.run([bead("a"), bead("b"), link()]);
319:   expect(committed.status).toBe("completed");
320:   const before = collect(f.owner, `${scope}events/`);
321:   const noop = f.run([{ ...patch, propertiesChange: [{ op: "replace", path: "/n", value: 0 }] }]);
322:   expect(noop.status).toBe("completed");
323:   expect(noop.effectPosition).toBeUndefined();
324:   const failed = f.run([
325:     patch,
326:     {
327:       operation: "updateBead",
328:       bead: "beads/missing",
329:       propertiesChange: [{ op: "replace", path: "/n", value: 1 }],
330:     },
331:   ]);
332:   expect(collect(f.owner, `${scope}events/`)).toEqual(before);
333:   expect(failed.status).toBe("failed");
334:   f.advance(86400001);
335:   expect(f.owner.maintain()).toEqual({ completedExpired: 2, failedForgotten: 1 });
336:   f.reopen();
337:   expect(f.owner.receipt("p", committed.id)?.detail).toBe("expired");
338:   expect(collect(f.owner, `${scope}events/`)).toEqual(before);
339: });
340: it("isolates unresolved metadata-only incident membership while Scope/Link sources remain available", () => {
341:   const f = fixture();
342:   f.run([bead("a"), bead("b"), link()]);
343:   const metadataChange = [{ op: "add", path: "/note", value: "own" }];
344:   expect(f.run([{ operation: "updateBead", bead: "beads/a", metadataChange }]).status).toBe(
345:     "completed",
346:   );
347:   expect(collect(f.owner, `${scope}beads/a?view=events`).events.at(-1)).toMatchObject({
348:     subject: `${scope}beads/a`,
349:     type: "updated",
350:     data: { metadataChange },
351:   });
352:   expect(f.run([{ ...patch, metadataChange }]).status).toBe("completed");
353:   expect(collect(f.owner, `${scope}beads/a?view=events`).events.at(-1)).toMatchObject({
354:     subject: `${scope}links/l`,
355:     type: "updated",
356:     data: { change: patch.propertiesChange, metadataChange },
357:   });
358:   f.run([
359:     {
360:       operation: "updateLink",
361:       link: "links/l",
362:       metadataChange: [{ op: "add", path: "/n", value: 1 }],
363:     },
364:   ]);
365:   expect(collect(f.owner, `${scope}links/l?view=events`).events.at(-1)?.data).toHaveProperty(
366:     "metadataChange",
367:   );
368:   expect(collect(f.owner, `${scope}events/`).events.at(-1)?.data).toHaveProperty("metadataChange");
369:   expect(() => collect(f.owner, `${scope}beads/a?view=events`)).toThrow(
370:     TransactionEventProjectionGap,
371:   );
372: });

### packages/server/src/transaction-events.test.ts:549–578

549:     await post([remove, { operation: "deleteBead", bead: "beads/a" }], "delete");
550:     expect((await get("beads/a")).status).toBe(404);
551:     expect((await get("beads/a?view=events")).status).toBe(200);
552:     expect(
553:       (await post([bead("c"), bead("d"), link("metadata", "beads/c", "beads/d")], "metadata-seed"))
554:         .status,
555:     ).toBe(200);
556:     expect(
557:       (
558:         await post(
559:           [
560:             {
561:               operation: "updateLink",
562:               link: "links/metadata",
563:               metadataChange: [{ op: "add", path: "/note", value: "pending-ruling" }],
564:             },
565:           ],
566:           "metadata-update",
567:         )
568:       ).status,
569:     ).toBe(200);
570:     for (const method of ["GET", "HEAD"]) {
571:       const gap = await get("beads/c?view=events", {}, method);
572:       expect(gap.status).toBe(500);
573:       expect(gap.headers.get("x-bdp-development-limit")).toBe("incident-link-metadata");
574:       expect(gap.headers.get("x-bdp-development")).toBe("non-attesting");
575:       expect(await gap.text()).toBe("");
576:     }
577:     expect((await get("links/metadata?view=events")).status).toBe(200);
578:     expect((await get("events/")).status).toBe(200);

### packages/server/src/transaction-event-sse.test.ts:515–557

515: it("an initial metadata-only incident Link gap refuses Bead SSE while Scope and Link remain available", async () => {
516:   const f = await runtimeFixture();
517:   await f.run([create("a"), create("b"), link(), metadataUpdate()]);
518:   const gap = f.stream("beads/a?view=events");
519:   await gap.closed;
520:   expect(gap.status).toBe(500);
521:   expect(gap.headers["x-bdp-development-limit"]).toBe("incident-link-metadata");
522:   expect(gap.body).toBe("");
523:   for (const target of ["events/", "links/l?view=events"]) {
524:     const finite = await f.finite(target),
525:       live = f.stream(target);
526:     await live.until(() => eventFrames(live).length === finite.events.length);
527:     expect(eventFrames(live).map(decoded)).toEqual(finite.events);
528:     await live.close();
529:   }
530: });
531: it("a later projection gap closes without a second response and repeat reconnect reaches the same gap", async () => {
532:   const f = await runtimeFixture((error) => {
533:     expect(
534:       causes(error).some((e) => e instanceof Error && e.message === "incident-link-metadata"),
535:     ).toBe(true);
536:   });
537:   await f.run([create("a"), create("b"), link()]);
538:   const finite = await f.finite("beads/a?view=events"),
539:     last = finite.events.at(-1)!.id;
540:   const live = f.stream("beads/a?view=events", [["Last-Event-ID", last]]);
541:   await live.ready;
542:   expect(live.status).toBe(200);
543:   await f.run([metadataUpdate()]);
544:   await live.closed;
545:   expect(eventFrames(live)).toEqual([]);
546:   expect(live.wireBytes.toString().match(/HTTP\/1\.1 /g)).toHaveLength(1);
547:   for (let n = 0; n < 2; n++) {
548:     const reconnect = f.stream("beads/a?view=events", [["Last-Event-ID", last]]);
549:     await reconnect.ready;
550:     expect(reconnect.status).toBe(200);
551:     await reconnect.closed;
552:     expect(eventFrames(reconnect)).toEqual([]);
553:   }
554:   expect((await f.finite("links/l?view=events")).events.at(-1)!.data).toHaveProperty(
555:     "metadataChange",
556:   );
557: });

### packages/server/src/transaction-http.test.ts:4594–4735

4594: it("rejects malformed minimum tuples across read routes without entering the actual owner", async () => {
4595:   const directory = mkdtempSync(path.join(tmpdir(), "bdp-tx-http-"));
4596:   const failures: unknown[] = [];
4597:   let owner: ReturnType<typeof openTransactionStore> | undefined;
4598:   let service: ReturnType<typeof createTransactionService> | undefined;
4599:   const entries: string[] = [];
4600:   try {
4601:     const scope = "https://example.test/minimum/";
4602:     owner = openTransactionStore({
4603:       file: path.join(directory, "owner.sqlite"),
4604:       scope,
4605:       create: true,
4606:       types,
4607:     });
4608:     // Copy the frozen owner's method descriptors into a mutable delegating facade.
4609:     // Every call still reaches the actual owner; setup observations are counted separately.
4610:     const facade = { ...owner };
4611:     for (const [name, value] of Object.entries(facade)) {
4612:       if (typeof value === "function")
4613:         Object.defineProperty(facade, name, {
4614:           value: (...args: unknown[]) => {
4615:             entries.push(name);
4616:             return Reflect.apply(value, owner, args);
4617:           },
4618:         });
4619:     }
4620:     let opaque = 0;
4621:     service = createTransactionService(facade, {
4622:       ownerLifetime: "service",
4623:       read: {
4624:         selectorLimits: { bytes: 4096, depth: 32, nodes: 256 },
4625:         advertisedLimits: {
4626:           page: { defaultItems: 2, maximumItems: 32 },
4627:           selector: { bytes: 4096, depth: 32, nodes: 256 },
4628:           cursorTtlMilliseconds: 1000,
4629:         },
4630:         pagination: {
4631:           scope,
4632:           defaultPageItems: 2,
4633:           maxPageItems: 32,
4634:           cursorTtlMs: 1000,
4635:           retainedStateCapacity: 100,
4636:           maxRetainedCursorPositionsPerSnapshot: 20,
4637:           retainedSnapshotByteCapacity: 100000,
4638:           retainedSnapshotNodeCapacity: 10000,
4639:           maxOpaqueTokenLength: 64,
4640:           tokenGenerationAttempts: 3,
4641:           idleCleanup: "on-demand",
4642:           clock: Date.now,
4643:           generateOpaqueToken: () => `minimum_${++opaque}`,
4644:         },
4645:       },
4646:     });
4647:     const handler = createTransactionHttpHandler(service, scope);
4648:     const setupEntries = [...entries];
4649:     entries.length = 0;
4650:     let calls = 0,
4651:       directGetBytes = 0;
4652:     const authorization = { kind: "authenticated", principal: "p", canMutate: true } as const;
4653:     for (const suffix of [
4654:       "",
4655:       "bdp.json",
4656:       "operations/",
4657:       "beads/",
4658:       "beads/missing",
4659:       "links/missing",
4660:       "alias/missing",
4661:       "snapshot",
4662:       "changes/",
4663:       "beads/missing?view=events",
4664:       "links/missing?view=events",
4665:       "events/",
4666:       "receipts/",
4667:       "receipts/unknown",
4668:       "receipts/unknown/pages/unknown",
4669:     ])
4670:       for (const method of ["GET", "HEAD"]) {
4671:         const response = await handler.handle(
4672:           new Request(new URL(suffix, scope), {
4673:             method,
4674:             headers: {
4675:               "bdp-minimum-scope-position": "genesis",
4676:               accept: "text/event-stream",
4677:               "if-none-match": "*",
4678:               "if-match": '"unmatched"',
4679:             },
4680:           }),
4681:           authorization,
4682:         );
4683:         calls++;
4684:         expect(response.kind, suffix).toBe("finite");
4685:         expect(response.status, `${method} ${suffix}`).toBe(400);
4686:         if (method === "GET") {
4687:           expect(checkedProblem(record(response.body))).toMatchObject({
4688:             code: "malformed-request",
4689:             status: 400,
4690:             retry: "never",
4691:           });
4692:           directGetBytes = Buffer.byteLength(JSON.stringify(response.body));
4693:           expect(directGetBytes).toBeGreaterThan(0);
4694:           // GET length is emitted later by the Node delivery boundary.
4695:           expect(response.headers.get("content-length")).toBeNull();
4696:         } else {
4697:           expect(response.body).toBeUndefined();
4698:           expect(Number(response.headers.get("content-length"))).toBe(directGetBytes);
4699:         }
4700:         for (const name of [
4701:           "bdp-scope-epoch",
4702:           "bdp-authorization-view",
4703:           "bdp-scope-position",
4704:           "etag",
4705:           "location",
4706:           "link",
4707:           "x-bdp-development-limit",
4708:         ])
4709:           expect(response.headers.has(name), name).toBe(false);
4710:       }
4711:     for (const [method, authenticated, minimum, status] of [
4712:       ["GET", false, true, 401],
4713:       ["POST", true, true, 405],
4714:     ] as const) {
4715:       const response = await handler.handle(
4716:         new Request(`${scope}receipts/`, {
4717:           method,
4718:           headers: minimum ? { "bdp-minimum-scope-position": "genesis" } : {},
4719:         }),
4720:         authenticated ? authorization : { kind: "unauthenticated" },
4721:       );
4722:       calls++;
4723:       expect(response.status).toBe(status);
4724:       if (status === 405) expect(response.headers.get("allow")).toBe("GET, HEAD");
4725:     }
4726:     expect(entries, "malformed receiving must precede every owner read").toStrictEqual([]);
4727:     const malformedEntries = [...entries];
4728:     const absent = await handler.handle(new Request(`${scope}receipts/`), authorization);
4729:     calls++;
4730:     expect(absent.status).toBe(404);
4731:     expect(entries, "ordinary absent-minimum receipt root still observes the owner").toStrictEqual([
4732:       "readReceipt",
4733:     ]);
4734:     expect(calls).toBe(33);
4735:     console.log(

### packages/server/src/transaction-http.test.ts:4857–4925

4857: it("rejects raw repeated minimum fields and keeps receiving HEAD and SSE refusals finite", async () => {
4858:   await minimumReceivingFixture(async (scope, signal) => {
4859:     const tuple = [
4860:       "BDP-Minimum-Scope-Position",
4861:       "genesis",
4862:       "BDP-Scope-Epoch",
4863:       "epoch",
4864:       "BDP-Authorization-View",
4865:       "view",
4866:     ];
4867:     const observations: unknown[] = [];
4868:     let getLength: string | null = null;
4869:     for (let index = 0; index < 6; index++) {
4870:       const method = index === 4 ? "HEAD" : "GET";
4871:       const fields =
4872:         index < 3
4873:           ? [...tuple, tuple[index * 2]!.toLowerCase(), tuple[index * 2 + 1]!]
4874:           : index === 3
4875:             ? [...tuple.slice(0, 1), "genesis,genesis", ...tuple.slice(2)]
4876:             : [
4877:                 "BDP-Minimum-Scope-Position",
4878:                 "genesis",
4879:                 ...(index === 5 ? ["Accept", "text/event-stream"] : []),
4880:               ];
4881:       const response = await minimumReceivingWire(scope, method, "beads/", fields, signal);
4882:       observations.push({
4883:         method,
4884:         status: response.status,
4885:         bytes: response.bytes,
4886:         contentType: response.headers.get("content-type"),
4887:         marker: response.headers.get("x-bdp-development-limit"),
4888:       });
4889:       expect(response.status).toBe(400);
4890:       expect(response.headers.get("content-type")).toBe("application/problem+json");
4891:       if (method === "GET") {
4892:         expect(checkedProblem(record(JSON.parse(response.text)))).toMatchObject({
4893:           code: "malformed-request",
4894:           status: 400,
4895:           retry: "never",
4896:         });
4897:         expect(Number(response.headers.get("content-length"))).toBe(response.bytes);
4898:         getLength = response.headers.get("content-length");
4899:       } else {
4900:         expect(response.text).toBe("");
4901:         expect(response.bytes).toBe(0);
4902:         expect(response.headers.get("content-length")).toBe(getLength);
4903:       }
4904:       for (const name of [
4905:         "bdp-scope-epoch",
4906:         "bdp-authorization-view",
4907:         "bdp-scope-position",
4908:         "etag",
4909:         "location",
4910:         "link",
4911:         "x-bdp-development-limit",
4912:       ])
4913:         expect(response.headers.has(name)).toBe(false);
4914:     }
4915:     expect(observations).toHaveLength(6);
4916:     console.log(
4917:       JSON.stringify({
4918:         observation: "minimum-receiving-raw",
4919:         requests: 6,
4920:         responses: observations,
4921:       }),
4922:     );
4923:   });
4924: }, 20000);
4925: 

### apps/bdptest/src/transactional-snapshot-installed.test.ts:175–205

175:   run: (signal: AbortSignal) => Promise<T>,
176:   milliseconds = 45000,
177: ): Promise<T> {
178:   const controller = new AbortController();
179:   const timer = setTimeout(
180:     () => controller.abort(new Error("installed Snapshot test deadline")),
181:     milliseconds,
182:   );
183:   return finishObservation(
184:     () => run(controller.signal),
185:     async () => {
186:       clearTimeout(timer);
187:     },
188:     controller.signal,
189:   );
190: }
191: async function withInstalled(
192:   run: (f: Awaited<ReturnType<typeof start>>) => Promise<void>,
193:   signal?: AbortSignal,
194: ): Promise<void> {
195:   if (!signal) return withDeadline((shared) => withInstalled(run, shared));
196:   signal.throwIfAborted();
197:   const f = await start(signal);
198:   await finishObservation(
199:     () => run(f),
200:     () => f.close(),
201:     signal,
202:     (ok) => f.report(ok),
203:   );
204: }
205: async function start(signal: AbortSignal) {

### apps/bdptest/src/transactional-snapshot-installed.test.ts:205–388

205: async function start(signal: AbortSignal) {
206:   signal.throwIfAborted();
207:   const directory = mkdtempSync(path.join(tmpdir(), "bdp-snapshot-installed-"));
208:   const operator = randomBytes(32).toString("base64url"),
209:     reader = randomBytes(32).toString("base64url");
210:   const configuration = {
211:     scope,
212:     directory,
213:     create: true,
214:     artifacts: artifacts(),
215:     credentials: [
216:       { token: operator, principal: "writer", canMutate: true },
217:       { token: reader, principal: "reader", canMutate: false },
218:     ],
219:   };
220:   let child: Awaited<ReturnType<typeof launchInstalledTransaction>>;
221:   try {
222:     child = await launchInstalledTransaction(entry, configuration, signal);
223:   } catch (error) {
224:     return failureAfterCleanup(error, async () => {
225:       rmSync(directory, { recursive: true, force: true });
226:     });
227:   }
228:   let baseline: ReturnType<typeof installedTransactionBinding>;
229:   try {
230:     baseline = installedTransactionBinding(child.ready, sourceCommit, fixture);
231:     expect(child.ready.claimEligible).toBe(false);
232:     expect(child.ready.history).toBe(false);
233:   } catch (error) {
234:     return failureAfterCleanup(error, () =>
235:       finishObservation(
236:         () => child.close(),
237:         async () => {
238:           rmSync(directory, { recursive: true, force: true });
239:         },
240:       ),
241:     );
242:   }
243:   const children = [child.child];
244:   let key = 0;
245:   const observations: {
246:     url: string;
247:     status: number;
248:     method: string;
249:     responseBytes: number;
250:     wireSha256: string;
251:   }[] = [];
252:   async function exchange(
253:     url = `${scope}snapshot`,
254:     method: HttpMethod = "GET",
255:     headers: Record<string, string> = {},
256:     body?: unknown,
257:     credentialRef = "operator",
258:   ) {
259:     const canonical = new URL(url);
260:     if (
261:       canonical.username !== "" ||
262:       canonical.password !== "" ||
263:       canonical.hash !== "" ||
264:       canonical.origin !== new URL(scope).origin ||
265:       !canonical.pathname.startsWith(new URL(scope).pathname)
266:     )
267:       throw new Error("Snapshot observation escaped the configured Scope");
268:     const address = new URL(child.ready.listenUrl as string);
269:     // Authorization-view is authored protocol context, not credential decoration.
270:     const execute = createRawHttpExchangeExecutor(
271:       { transport: "plain", host: address.hostname, port: Number(address.port) },
272:       {
273:         requestTimeoutMs: 5000,
274:         maximumBodyBytes: 1024 * 1024,
275:         exactMode: {
276:           scope,
277:           profile: "transactional",
278:           routes: [{ method, url }],
279:           maximumRequestHeaderBytes: 16384,
280:           maximumRequestBodyBytes: 65536,
281:           defaultCredentialRef: "operator",
282:           credentialHandles: [
283:             { id: "operator", headerNames: ["authorization"] },
284:             { id: "reader", headerNames: ["authorization"] },
285:             { id: "anonymous", headerNames: [] },
286:           ],
287:           resolveCredentials: ({ credentialRef }) =>
288:             credentialRef === "anonymous"
289:               ? {}
290:               : {
291:                   authorization: `Bearer ${credentialRef === "reader" ? reader : operator}`,
292:                 },
293:         },
294:       },
295:     );
296:     const response = await execute({
297:       method,
298:       url,
299:       signal,
300:       credentialRef,
301:       raw: {
302:         headerLines: Object.entries(headers).map(([name, value]) => ({ name, value })),
303:         ...(body === undefined ? {} : { bodyBytes: Buffer.from(JSON.stringify(body)) }),
304:       },
305:     });
306:     expect(response.exactRequest?.writeState).toBe("complete");
307:     expect(
308:       response.exactRequest?.headerLines.find((h) => h.name.toLowerCase() === "host")?.value,
309:     ).toBe("reference.invalid");
310:     expect(response.wireResponseBytes?.length).toBeGreaterThan(0);
311:     if (headers["bdp-authorization-view"] !== undefined)
312:       expect(
313:         response.exactRequest?.headerLines.find(
314:           (h) => h.name.toLowerCase() === "bdp-authorization-view",
315:         )?.value,
316:       ).toBe(headers["bdp-authorization-view"]);
317:     if (
318:       response.bodyOctets === undefined ||
319:       !Number.isSafeInteger(response.bodyOctets) ||
320:       response.bodyOctets < 0
321:     )
322:       throw new Error("completed response body octets unavailable");
323:     observations.push({
324:       url,
325:       method,
326:       status: response.status,
327:       responseBytes: response.bodyOctets,
328:       wireSha256: createHash("sha256").update(response.wireResponseBytes!).digest("hex"),
329:     });
330:     expect(response.headers["x-bdp-development"]).toBe("non-attesting");
331:     return response;
332:   }
333:   return {
334:     exchange,
335:     get exchangeCount() {
336:       return observations.length;
337:     },
338:     baseline,
339:     signal,
340:     async write(operations: unknown[], credential = "operator") {
341:       const response = await exchange(
342:         `${scope}operations/batch`,
343:         "POST",
344:         {
345:           "content-type": "application/json",
346:           "idempotency-key": `snapshot_${key++}`,
347:         },
348:         { operations },
349:         credential,
350:       );
351:       const receipt = json<{ status: string }>(response);
352:       expect(receipt.status).toBe("completed");
353:     },
354:     async capture() {
355:       return json<Manifest>(await exchange());
356:     },
357:     async reopen() {
358:       await child.close();
359:       child = await launchInstalledTransaction(entry, { ...configuration, create: false }, signal);
360:       children.push(child.child);
361:       expect(installedTransactionBinding(child.ready, sourceCommit, fixture)).toEqual(baseline);
362:     },
363:     async close() {
364:       await finishObservation(
365:         () => child.close(),
366:         async () => {
367:           rmSync(directory, { recursive: true, force: true });
368:         },
369:       );
370:     },
371:     report(ok: boolean) {
372:       // Never log raw credentials, exactRequest headers or authored request bodies.
373:       console.log(
374:         JSON.stringify({
375:           observation: "private-installed-snapshot",
376:           outcome: ok ? "passed" : "failed",
377:           aborted: signal.aborted,
378:           declaredSourceCommit: sourceCommit,
379:           workspaceDirty,
380:           entry: resolvedEntry,
381:           entrySha256,
382:           binding: baseline,
383:           children: children.map((process) => ({
384:             pid: process.pid,
385:             exitCode: process.exitCode,
386:             signalCode: process.signalCode,
387:           })),
388:           exchanges: observations.length,

### apps/bdptest/src/transactional-snapshot-installed.test.ts:422–481

422: it("observes packed-capable canonical Snapshot streams, retained bytes and changefeed-to-point convergence", async () => {
423:   await withInstalled(async (f) => {
424:     // Hosting is observed over the same installed listener, not reconstructed from metadata.
425:     for (const artifact of artifacts().descriptors) {
426:       const hosted = await f.exchange(artifact.retrievalUri, "GET", {}, undefined, "reader");
427:       expect(hosted.status).toBe(200);
428:       expect(json<unknown>(hosted)).toEqual(JSON.parse(artifact.utf8));
429:     }
430:     const denied = await f.exchange(
431:       `${scope}operations/batch`,
432:       "POST",
433:       {
434:         "content-type": "application/json",
435:         "idempotency-key": "reader_denied",
436:       },
437:       { operations: [createBead("reader-must-not-create")] },
438:       "reader",
439:     );
440:     expect(json<{ code: string }>(denied, 403).code).toBe("forbidden");
441:     expect(
442:       json<{ code: string }>(await f.exchange(`${scope}beads/reader-must-not-create`), 404).code,
443:     ).toBe("resource-not-found");
444:     const empty = json<Manifest>(await f.exchange(undefined, "GET", {}, undefined, "reader"));
445:     expect(empty.checkpoint).toBe("genesis");
446:     expect(empty.beads).toEqual({ items: [], next: null });
447:     expect(empty.links).toEqual({ items: [], next: null });
448:     const names = Array.from({ length: 34 }, (_, i) => `b${String(i).padStart(2, "0")}`);
449:     await f.write(names.slice(0, 32).map(createBead));
450:     await f.write(names.slice(32).map(createBead));
451:     const edges = names.map((name, i) => ({
452:       operation: "createLink",
453:       id: `links/e${String(i).padStart(2, "0")}`,
454:       type: linkType,
455:       source: `beads/${name}`,
456:       target: "beads/b00",
457:       properties: { title: `edge ${i}` },
458:       metadata: { seeded: `e${String(i).padStart(2, "0")}` },
459:     }));
460:     await f.write(edges.slice(0, 32));
461:     await f.write(edges.slice(32));
462:     const snapshot = await f.capture();
463:     expect(snapshot.scope).toBe(scope);
464:     expect(snapshot.scopePosition).toBe("pos-4");
465:     expect(snapshot.checkpoint).not.toBe("genesis");
466:     expect(Date.parse(snapshot.expiresAt)).toBeGreaterThan(Date.now());
467:     expect(snapshot.beads.items).toHaveLength(32);
468:     expect(snapshot.links.items).toHaveLength(32);
469:     expect(snapshot.beads.next).not.toBeNull();
470:     expect(snapshot.links.next).not.toBeNull();
471:     const retainedMinimum = await f.exchange(snapshot.beads.next as string, "GET", {
472:       "bdp-scope-epoch": snapshot.scopeEpoch,
473:       "bdp-authorization-view": snapshot.authorizationView,
474:       "bdp-minimum-scope-position": "pos-4",
475:     });
476:     expect(retainedMinimum.status).toBe(500);
477:     expect(retainedMinimum.headers["x-bdp-development-limit"]).toBe("minimum-position");
478:     const beads = await collect(f, snapshot.beads),
479:       links = await collect(f, snapshot.links);
480:     expect(beads.rows.map((r) => r.id)).toEqual(names.map((n) => `${scope}beads/${n}`));
481:     expect(links.rows.map((r) => r.id)).toEqual(edges.map((e) => `${scope}${e.id}`));

### apps/bdptest/src/transactional-snapshot-installed.test.ts:796–956

796: it("observes fresh Snapshot minima with exact current context and retained gaps over packed bytes", async () => {
797:   await withInstalled(async (f) => {
798:     const empty = await f.exchange(); // 1: actual initial context and empty handle
799:     expect(checkedManifest(empty)).toMatchObject({
800:       scopePosition: "pos-0",
801:       checkpoint: "genesis",
802:       beads: { items: [], next: null },
803:       links: { items: [], next: null },
804:     });
805:     const unauthorized = await f.exchange(
806:       undefined,
807:       "GET",
808:       snapshotMinimum(empty, "genesis"),
809:       undefined,
810:       "anonymous",
811:     ); // 2
812:     expect(unauthorized.status).toBe(401);
813:     for (const name of ["bdp-scope-epoch", "bdp-authorization-view", "bdp-scope-position"])
814:       expect(unauthorized.headers[name]).toBeUndefined();
815:     await f.write([
816:       createBead("a"),
817:       createBead("b"),
818:       {
819:         operation: "createLink",
820:         id: "links/ab",
821:         type: linkType,
822:         source: "beads/a",
823:         target: "beads/b",
824:         properties: { title: "edge" },
825:       },
826:     ]); // 3
827:     const current = snapshotMinimum(empty, "pos-1"),
828:       future = snapshotMinimum(empty, "pos-2");
829:     function state(manifest: Manifest, title = "café 🦉 a", position = "pos-1") {
830:       expect(manifest.scopePosition).toBe(position);
831:       expect(manifest.checkpoint).not.toBe("genesis");
832:       expect(manifest.beads.items.map((row) => [row.id, row.properties])).toEqual([
833:         [`${scope}beads/a`, { title }],
834:         [`${scope}beads/b`, { title: "café 🦉 b" }],
835:       ]);
836:       expect(manifest.links.items.map((row) => [row.id, row.properties])).toEqual([
837:         [`${scope}links/ab`, { title: "edge" }],
838:       ]);
839:       expect(manifest.beads.next).toBeNull();
840:       expect(manifest.links.next).toBeNull();
841:     }
842:     const firstWire = await f.exchange(undefined, "GET", current); // 4
843:     const first = checkedManifest(firstWire);
844:     state(first);
845:     exactSnapshotContext(firstWire, empty, "pos-1");
846:     for (const [method, conditions, status] of [
847:       ["HEAD", {}, 200],
848:       ["GET", { "if-none-match": "*" }, 304],
849:       ["GET", { "if-match": '"absent"' }, 412],
850:       ["GET", { accept: "text/plain" }, 406],
851:     ] as [HttpMethod, Record<string, string>, number][]) {
852:       // 5–8
853:       const wire = await f.exchange(undefined, method, { ...current, ...conditions });
854:       expect(wire.status).toBe(status);
855:       expect(wire.bodyText).toBe("");
856:       exactSnapshotContext(wire, empty, "pos-1");
857:     }
858:     for (const [method, conditions] of [
859:       ["GET", {}],
860:       ["HEAD", {}],
861:       ["GET", { "if-none-match": "*" }],
862:       ["GET", { "if-match": '"absent"' }],
863:       ["GET", { accept: "text/plain" }],
864:     ] as [HttpMethod, Record<string, string>][]) // 9–13
865:       snapshotTimeout(
866:         await f.exchange(undefined, method, { ...future, ...conditions }),
867:         empty,
868:         "pos-1",
869:         method === "HEAD",
870:       );
871:     for (const [header, status, code, family] of [
872:       ["bdp-scope-epoch", 410, "cursor-expired", "gone"],
873:       ["bdp-authorization-view", 409, "foreign-view", "conflict"],
874:     ] as const)
875:       for (const method of ["GET", "HEAD"] as const) {
876:         // 14–17
877:         const wire = await f.exchange(undefined, method, { ...current, [header]: "foreign" });
878:         expect(wire.status).toBe(status);
879:         exactSnapshotContext(wire, empty, "pos-1");
880:         if (method === "HEAD") expect(wire.bodyText).toBe("");
881:         else
882:           expect(json<Record<string, unknown>>(wire, status)).toMatchObject({
883:             code,
884:             status,
885:             type: `https://github.com/gastownhall/bdp/problems/${family}`,
886:           });
887:       }
888:     const laterWire = await f.exchange(undefined, "GET", current); // 18
889:     const later = checkedManifest(laterWire);
890:     state(later);
891:     expect(later.id).not.toBe(first.id);
892:     exactSnapshotContext(laterWire, empty, "pos-1");
893:     for (const headers of [
894:       // 19–21
895:       { ...current, "bdp-minimum-scope-position": "unknown" },
896:       { "bdp-minimum-scope-position": "pos-1" },
897:       {
898:         "bdp-scope-epoch": current["bdp-scope-epoch"],
899:         "bdp-authorization-view": current["bdp-authorization-view"],
900:         "bdp-minimum-scope-positions": "pos-1",
901:       },
902:     ]) {
903:       const wire = await f.exchange(undefined, "GET", headers);
904:       if (headers["bdp-minimum-scope-position"] === "unknown") {
905:         snapshotTimeout(wire, empty, "pos-1");
906:         expect(wire.headers["x-bdp-development-limit"]).toBeUndefined();
907:       } else {
908:         expect(json<{ code: string }>(wire, 400).code).toBe("malformed-request");
909:         expect(wire.headers["x-bdp-development-limit"]).toBeUndefined();
910:         for (const name of ["bdp-scope-epoch", "bdp-authorization-view", "bdp-scope-position"])
911:           expect(wire.headers[name]).toBeUndefined();
912:       }
913:     }
914:     const retained = await f.exchange(first.id, "GET", current); // 22
915:     expect(retained.status).toBe(500);
916:     expect(retained.headers["x-bdp-development-limit"]).toBe("minimum-position");
917:     for (const method of ["GET", "HEAD"] as const) {
918:       // 23–24
919:       const wire = await f.exchange(first.id, method);
920:       expect(wire.status).toBe(200);
921:       exactSnapshotContext(wire, empty, "pos-1");
922:       expect(wire.bodyText).toBe(method === "HEAD" ? "" : firstWire.bodyText);
923:     }
924:     await f.write([
925:       {
926:         operation: "updateBead",
927:         bead: "beads/a",
928:         propertiesChange: [{ op: "replace", path: "/title", value: "after" }],
929:       },
930:     ]); // 25
931:     const advanced = await f.exchange(undefined, "GET", future); // 26
932:     state(checkedManifest(advanced), "after", "pos-2");
933:     exactSnapshotContext(advanced, empty, "pos-2");
934:     const old = await f.exchange(first.id); // 27
935:     expect(old.bodyText).toBe(firstWire.bodyText);
936:     exactSnapshotContext(old, empty, "pos-2");
937:     await f.reopen();
938:     for (const method of ["GET", "HEAD"] as const) {
939:       // 28–29
940:       const wire = await f.exchange(first.id, method);
941:       expect(wire.status).toBe(200);
942:       exactSnapshotContext(wire, empty, "pos-2");
943:       expect(wire.bodyText).toBe(method === "HEAD" ? "" : firstWire.bodyText);
944:     }
945:     snapshotTimeout(
946:       await f.exchange(undefined, "GET", snapshotMinimum(empty, "pos-3")),
947:       empty,
948:       "pos-2",
949:     ); // 30
950:     const final = await f.exchange(undefined, "GET", future); // 31
951:     state(checkedManifest(final), "after", "pos-2");
952:     exactSnapshotContext(final, empty, "pos-2");
953:     expect(f.exchangeCount).toBe(31);
954:   });
955: }, 60000);
956: 

### apps/bdptest/src/transactional-snapshot-installed.test.ts:972–1192

972: it("observes installed fresh Snapshot opaque minimum timeout", async () => {
973:   const started = performance.now();
974:   let directory: string | undefined;
975:   let child: Awaited<ReturnType<typeof launchInstalledTransaction>> | undefined;
976:   let binding: ReturnType<typeof installedTransactionBinding> | undefined;
977:   let launchStarted = false,
978:     closeAttempted = false,
979:     closed = false,
980:     attempted = 0;
981:   const responses: {
982:     method: string;
983:     url: string;
984:     status: number;
985:     responseBytes: number;
986:     wireSha256: string;
987:     elapsedSeconds: number;
988:   }[] = [];
989:   await withDeadline((signal) =>
990:     finishObservation(
991:       async () => {
992:         signal.throwIfAborted();
993:         directory = mkdtempSync(path.join(tmpdir(), "bdp-snapshot-opaque-installed-"));
994:         const operator = randomBytes(32).toString("base64url");
995:         const configuration = {
996:           scope,
997:           directory,
998:           create: true,
999:           artifacts: artifacts(),
1000:           credentials: [{ token: operator, principal: "writer", canMutate: true }],
1001:         };
1002:         const validReceipt = snapshotAjv.compile({
1003:           $ref: `${snapshotSchemaBundle.$id}#/$defs/mutationReceipt`,
1004:         });
1005:         launchStarted = true;
1006:         child = await launchInstalledTransaction(entry, configuration, signal);
1007:         binding = installedTransactionBinding(child.ready, sourceCommit, fixture);
1008:         expect(child.ready.claimEligible).toBe(false);
1009:         expect(child.ready.history).toBe(false);
1010:         const address = new URL(child.ready.listenUrl as string);
1011:         const batchUrl = `${scope}operations/batch`,
1012:           snapshotUrl = `${scope}snapshot`;
1013:         const execute = createRawHttpExchangeExecutor(
1014:           { transport: "plain", host: address.hostname, port: Number(address.port) },
1015:           {
1016:             requestTimeoutMs: 5000,
1017:             maximumBodyBytes: 1024 * 1024,
1018:             exactMode: {
1019:               scope,
1020:               profile: "transactional",
1021:               routes: [
1022:                 { method: "POST", url: batchUrl },
1023:                 { method: "GET", url: snapshotUrl },
1024:                 { method: "HEAD", url: snapshotUrl },
1025:               ],
1026:               maximumRequestHeaderBytes: 16384,
1027:               maximumRequestBodyBytes: 65536,
1028:               defaultCredentialRef: "operator",
1029:               credentialHandles: [{ id: "operator", headerNames: ["authorization"] }],
1030:               resolveCredentials: () => ({ authorization: `Bearer ${operator}` }),
1031:             },
1032:           },
1033:         );
1034:         async function exchange(
1035:           method: HttpMethod,
1036:           minimum: Record<string, string> = {},
1037:           body?: string,
1038:         ) {
1039:           signal.throwIfAborted();
1040:           expect(++attempted, "installed Snapshot dispatch bound").toBeLessThanOrEqual(3);
1041:           const url = method === "POST" ? batchUrl : snapshotUrl;
1042:           const requestStarted = performance.now();
1043:           const response = await execute({
1044:             method,
1045:             url,
1046:             signal,
1047:             credentialRef: "operator",
1048:             raw: {
1049:               headerLines: [
1050:                 { name: "accept", value: "application/json" },
1051:                 ...Object.entries(minimum).map(([name, value]) => ({ name, value })),
1052:                 ...(body === undefined
1053:                   ? []
1054:                   : [
1055:                       { name: "content-type", value: "application/json" },
1056:                       { name: "idempotency-key", value: "snapshot-opaque-seed" },
1057:                     ]),
1058:               ],
1059:               ...(body === undefined ? {} : { bodyBytes: Buffer.from(body) }),
1060:             },
1061:           });
1062:           expect(response.url).toBe(url);
1063:           expect(response.exactRequest?.source).toBe("raw-http1-serializer");
1064:           expect(response.exactRequest?.writeState).toBe("complete");
1065:           expect(response.exactRequest?.bodyPresent).toBe(body !== undefined);
1066:           expect(response.exactRequest?.bodyOctets).toBe(
1067:             body === undefined ? 0 : Buffer.byteLength(body),
1068:           );
1069:           expect(response.exactRequest?.bodyDigest).toBe(
1070:             body === undefined ? undefined : createHash("sha256").update(body).digest("hex"),
1071:           );
1072:           const headers = response.exactRequest?.headerLines ?? [];
1073:           expect(headers.find((header) => header.name.toLowerCase() === "host")?.value).toBe(
1074:             "reference.invalid",
1075:           );
1076:           expect(headers.find((header) => header.name.toLowerCase() === "accept")?.value).toBe(
1077:             "application/json",
1078:           );
1079:           if (body !== undefined) {
1080:             expect(
1081:               headers.find((header) => header.name.toLowerCase() === "content-type")?.value,
1082:             ).toBe("application/json");
1083:             expect(
1084:               headers.find((header) => header.name.toLowerCase() === "idempotency-key")?.value,
1085:             ).toBe("snapshot-opaque-seed");
1086:           }
1087:           for (const [name, value] of Object.entries(minimum))
1088:             expect(headers.find((header) => header.name.toLowerCase() === name)?.value).toBe(value);
1089:           expect(response.wireResponseBytes?.length).toBeGreaterThan(0);
1090:           if (
1091:             response.bodyOctets === undefined ||
1092:             !Number.isSafeInteger(response.bodyOctets) ||
1093:             response.bodyOctets < 0
1094:           )
1095:             throw new Error("completed response body octets unavailable");
1096:           responses.push({
1097:             method,
1098:             url,
1099:             status: response.status,
1100:             responseBytes: response.bodyOctets,
1101:             wireSha256: createHash("sha256").update(response.wireResponseBytes!).digest("hex"),
1102:             elapsedSeconds: (performance.now() - requestStarted) / 1000,
1103:           });
1104:           expect(response.headers["x-bdp-development"]).toBe("non-attesting");
1105:           return response;
1106:         }
1107:         const seed = await exchange(
1108:           "POST",
1109:           {},
1110:           JSON.stringify({ operations: [createBead("opaque-snapshot")] }),
1111:         );
1112:         const receipt = json<unknown>(seed);
1113:         expect(validReceipt(receipt), JSON.stringify(validReceipt.errors)).toBe(true);
1114:         expect(receipt).toMatchObject({
1115:           status: "completed",
1116:           idempotencyKey: "snapshot-opaque-seed",
1117:           requiredPosition: "pos-1",
1118:           effectPosition: "pos-1",
1119:           results: [
1120:             {
1121:               outcome: "created",
1122:               resource: {
1123:                 id: `${scope}beads/opaque-snapshot`,
1124:                 type: beadType,
1125:                 properties: { title: "café 🦉 opaque-snapshot" },
1126:               },
1127:             },
1128:           ],
1129:         });
1130:         expect(seed.headers["bdp-scope-position"]).toBe("pos-1");
1131:         const minimum = snapshotMinimum(seed, "opaque");
1132:         for (const method of ["GET", "HEAD"] as const) {
1133:           const response = await exchange(method, minimum);
1134:           snapshotTimeout(response, seed, "pos-1", method === "HEAD");
1135:           expect(response.headers["retry-after"]).toBe("1");
1136:           expect(response.headers["cache-control"]).toBe("private, no-store");
1137:           expect(response.headers["x-bdp-development-limit"]).toBeUndefined();
1138:         }
1139:         expect(attempted).toBe(3);
1140:         expect(responses.map(({ method, status }) => [method, status])).toEqual([
1141:           ["POST", 200],
1142:           ["GET", 503],
1143:           ["HEAD", 503],
1144:         ]);
1145:       },
1146:       async () => {
1147:         if (!child) {
1148:           if (launchStarted) throw new Error("Snapshot child ownership was not returned");
1149:           if (directory) rmSync(directory, { recursive: true, force: true });
1150:           return;
1151:         }
1152:         closeAttempted = true;
1153:         await child.close();
1154:         if (child.child.exitCode !== 0 || child.child.signalCode !== null)
1155:           throw new Error("Snapshot child did not close cleanly");
1156:         signal.throwIfAborted();
1157:         closed = true;
1158:         if (directory) rmSync(directory, { recursive: true, force: true });
1159:       },
1160:       signal,
1161:       (ok) => {
1162:         console.log(
1163:           JSON.stringify({
1164:             observation: "private-installed-snapshot-opaque-minimum",
1165:             ok,
1166:             declaredSourceCommit: sourceCommit,
1167:             workspaceDirty,
1168:             entry: resolvedEntry,
1169:             entrySha256,
1170:             binding,
1171:             attempted,
1172:             exchanges: responses.length,
1173:             responses,
1174:             launchStarted,
1175:             startupCompleted: child !== undefined,
1176:             closeAttempted,
1177:             closed,
1178:             childPid: child?.child.pid ?? null,
1179:             childExitCode: child?.child.exitCode ?? null,
1180:             childSignalCode: child?.child.signalCode ?? null,
1181:             ownership: child ? "returned" : launchStarted ? "unknown" : "not-started",
1182:             directory: directory ?? null,
1183:             directoryAbsent: directory === undefined || !existsSync(directory),
1184:             aborted: signal.aborted,
1185:             elapsedSeconds: (performance.now() - started) / 1000,
1186:             claimEligible: false,
1187:           }),
1188:         );
1189:       },
1190:     ),
1191:   );
1192: }, 60000);

### apps/bdptest/src/transactional-event-sse-installed.test.ts:80–97

80: it("installed Scope/Bead/Link Event SSE matches finite bytes and sparse Bead survives fifteen seconds", async () => {
81:   const controller = new AbortController();
82:   const deadline = setTimeout(
83:     () => controller.abort(new Error("installed SSE cohort deadline")),
84:     45000,
85:   );
86:   const directory = mkdtempSync(path.join(tmpdir(), "bdp-sse-installed-")),
87:     token = randomBytes(32).toString("base64url");
88:   let child: Awaited<ReturnType<typeof launchInstalledTransaction>> | undefined;
89:   let stream: ReturnType<typeof recordRawSse> | undefined;
90:   const failures: unknown[] = [];
91:   let report: unknown;
92:   try {
93:     child = await launchInstalledTransaction(
94:       entry,
95:       {
96:         scope,
97:         directory,

### apps/bdptest/src/transactional-event-sse-installed.test.ts:320–409

320: it("installed Event cursors survive restart with Last-Event-ID precedence and source binding", async () => {
321:   // The full suite executes this same witness locally; only externally bound artifacts
322:   // can supply installed attribution. A declared source string alone proves no provenance.
323:   const entryOverride = process.env.BDP_TRANSACTIONAL_INSTALLED_ENTRY;
324:   const sourceOverride = process.env.BDP_TRANSACTIONAL_SOURCE_COMMIT;
325:   const hasEntryOverride = entryOverride !== undefined;
326:   const hasSourceOverride = sourceOverride !== undefined;
327:   if (hasEntryOverride !== hasSourceOverride)
328:     throw new Error("Event restart overrides must include both installed entry and source");
329:   expect(source).toMatch(/^[0-9a-f]{40}$/);
330:   const attribution = hasEntryOverride ? "external-installed-candidate" : "development-local";
331:   if (hasEntryOverride) {
332:     if (!entryOverride || !sourceOverride || !/^[0-9a-f]{40}$/.test(sourceOverride))
333:       throw new Error(
334:         "Event restart installed overrides must be nonempty and source must be a commit",
335:       );
336:     const installedRelative = path.relative(realpathSync(root), entry);
337:     if (!installedRelative.startsWith(`..${path.sep}`) && !path.isAbsolute(installedRelative))
338:       throw new Error("Event restart installed attribution requires an external entry");
339:   }
340:   const hash = (value: string | Uint8Array) => createHash("sha256").update(value).digest("hex");
341:   const witnessPath = fileURLToPath(import.meta.url);
342:   const witnessSource = {
343:     fingerprintScope:
344:       "witness file bytes and tracked working-tree diff only; excludes untracked inputs and installed payload provenance",
345:     head: execFileSync("git", ["rev-parse", "HEAD"], { cwd: root, encoding: "utf8" }).trim(),
346:     fileSha256: hash(readFileSync(witnessPath)),
347:     diffSha256: hash(execFileSync("git", ["diff", "HEAD"], { cwd: root })),
348:     dirtyPaths: execFileSync("git", ["diff", "--name-only", "HEAD"], {
349:       cwd: root,
350:       encoding: "utf8",
351:     })
352:       .trim()
353:       .split("\n")
354:       .filter(Boolean),
355:   };
356:   const controller = new AbortController();
357:   const started = performance.now(),
358:     workEnd = started + 45000;
359:   const directory = mkdtempSync(path.join(tmpdir(), "bdp-event-restart-installed-"));
360:   const token = randomBytes(32).toString("base64url");
361:   function safeFailureMessage(error: unknown): string {
362:     return (error instanceof Error ? error.message : "non-Error failure").replaceAll(
363:       token,
364:       "[credential redacted]",
365:     );
366:   }
367:   const configuration = {
368:     scope,
369:     directory,
370:     create: true,
371:     artifacts,
372:     credentials: [{ token, principal: "writer", canMutate: true }],
373:   };
374:   const seedKey = "event-restart-two-created-v1";
375:   const seedBytes = Buffer.from(
376:     JSON.stringify({
377:       operations: ["a", "b"].map((id) => ({
378:         operation: "createBead",
379:         id: `beads/${id}`,
380:         type,
381:         properties: { n: 0 },
382:       })),
383:     }),
384:   );
385:   type Handle = Awaited<ReturnType<typeof launchInstalledTransaction>>;
386:   type Session = {
387:     ordinal: number;
388:     create: boolean;
389:     state: "launching" | "owned" | "unknown-no-handle";
390:     handle?: Handle;
391:     pid?: number | undefined;
392:     listenUrl?: string | undefined;
393:     binding?: ReturnType<typeof installedTransactionBinding>;
394:     exit?: { code: number | null; signal: NodeJS.Signals | null };
395:     closing?: Promise<void>;
396:     closed: boolean;
397:     stopUnknown: boolean;
398:   };
399:   const sessions: Session[] = [];
400:   const failures: { phase: string; error: unknown }[] = [];
401:   const pending = new Set<Promise<unknown>>();
402:   const localSensitivityControls: string[] = [];
403:   const requests: { ordinal: number; session: number; method: string; target: string }[] = [];
404:   const responses: unknown[] = [];
405:   let phase = "setup",
406:     cleanupStarted = false,
407:     abortedForCleanup = false,
408:     removed = false;
409:   let binding: ReturnType<typeof installedTransactionBinding> | undefined;

### apps/bdptest/src/transactional-event-sse-installed.test.ts:1794–1810

1794: it("observes installed current Event replay minimum admission across sources", async () => {
1795:   // The full suite executes this same witness locally; only externally bound artifacts
1796:   // can supply installed attribution. A declared source string alone proves no provenance.
1797:   const entryOverride = process.env.BDP_TRANSACTIONAL_INSTALLED_ENTRY;
1798:   const sourceOverride = process.env.BDP_TRANSACTIONAL_SOURCE_COMMIT;
1799:   const hasEntryOverride = entryOverride !== undefined;
1800:   const hasSourceOverride = sourceOverride !== undefined;
1801:   if (hasEntryOverride !== hasSourceOverride)
1802:     throw new Error("Event replay minimum overrides must include both installed entry and source");
1803:   expect(source).toMatch(/^[0-9a-f]{40}$/);
1804:   const attribution = hasEntryOverride ? "external-installed-candidate" : "development-local";
1805:   if (hasEntryOverride) {
1806:     if (!entryOverride || !sourceOverride || !/^[0-9a-f]{40}$/.test(sourceOverride))
1807:       throw new Error(
1808:         "Event replay minimum installed overrides must be nonempty and source must be a commit",
1809:       );
1810:     const installedRelative = path.relative(realpathSync(root), entry);

### apps/bdptest/src/transactional-minimum-receiving-installed.test.ts:126–141

126: it("rejects malformed read minimum tuples before observation while preserving receipt nondisclosure", async () => {
127:   const controller = new AbortController();
128:   const deadlineReason = new Error("installed minimum receiving deadline");
129:   let deadlineExceeded = false,
130:     cleanupExceeded = false,
131:     launchAttempted = false,
132:     closed = false;
133:   const deadline = setTimeout(() => {
134:     deadlineExceeded = true;
135:     controller.abort(deadlineReason);
136:   }, 45000);
137:   const failures: unknown[] = [];
138:   const categories: string[] = [];
139:   const fail = (error: unknown, category: string) => {
140:     if (!failures.includes(error)) {
141:       failures.push(error);

### apps/bdptest/src/transactional-minimum-receiving-installed.test.ts:350–470

350:       "seed-post-ignores-malformed-minimum",
351:       "POST",
352:       `${scope}operations/batch`,
353:       ["BDP-Minimum-Scope-Position", "invalid,invalid"],
354:       "writer",
355:       seedBody,
356:       "minimum-receiving-seed",
357:     );
358:     const seeded = receipt(seed, "minimum-receiving-seed", "pos-1");
359:     expect(receiptResults(seeded)).toHaveLength(32);
360:     expect(seeded.next).toBeNull();
361:     const updated = await exchange(
362:       "issue-real-page",
363:       "POST",
364:       `${scope}operations/update-where`,
365:       [],
366:       "writer",
367:       setBody,
368:       "minimum-receiving-paged",
369:     );
370:     const paged = receipt(updated, "minimum-receiving-paged", "pos-2");
371:     expect(receiptResults(paged)).toHaveLength(32);
372:     const receiptRoot = paged.id,
373:       pageUrl = paged.next;
374:     if (typeof receiptRoot !== "string" || typeof pageUrl !== "string")
375:       throw new Error("actual issued receipt/page required");
376:     for (const url of [receiptRoot, pageUrl]) {
377:       const parsed = new URL(url);
378:       expect(parsed.href).toBe(url);
379:       expect(parsed.origin).toBe(new URL(scope).origin);
380:       expect([parsed.search, parsed.hash, parsed.username, parsed.password]).toStrictEqual([
381:         "",
382:         "",
383:         "",
384:         "",
385:       ]);
386:       granted.add(url);
387:     }
388:     expect(receiptRoot.startsWith(`${scope}receipts/`)).toBe(true);
389:     expect(receiptRoot.slice(`${scope}receipts/`.length)).toMatch(/^[A-Za-z0-9_-]+$/);
390:     expect(pageUrl.startsWith(`${receiptRoot}/pages/`)).toBe(true);
391:     expect(pageUrl.slice(`${receiptRoot}/pages/`.length)).toMatch(/^[A-Za-z0-9_-]+$/);
392:     const expectedContext = context(updated);
393:     expect(expectedContext.slice(0, 2)).toStrictEqual(context(seed).slice(0, 2));
394:     const tuple = [
395:       "BDP-Minimum-Scope-Position",
396:       "pos-2",
397:       "BDP-Scope-Epoch",
398:       expectedContext[0]!,
399:       "BDP-Authorization-View",
400:       expectedContext[1]!,
401:     ];
402:     const point = `${scope}beads/r00`;
403:     const first = await exchange("all-absent", "GET", point);
404:     expect(first.status).toBe(200);
405:     const resource = record(JSON.parse(first.text));
406:     expect(validResource(resource), JSON.stringify(validResource.errors)).toBe(true);
407:     expect(resource).toStrictEqual({
408:       id: point,
409:       type: beadType,
410:       properties: { n: 1 },
411:       metadata: {},
412:       revision: resource.revision,
413:     });
414:     expect(typeof resource.revision).toBe("string");
415:     expect(
416:       receiptResults(paged).find((result) => record(result.resource ?? {}).id === point),
417:     ).toStrictEqual({ operationIndex: 0, outcome: "updated", resource });
418:     expect(context(first)).toStrictEqual(expectedContext);
419:     let malformedLength: string | null = null;
420:     const malformed = (response: Response, head = false) => {
421:       expect(response.status).toBe(400);
422:       expect(response.headers.get("content-type")).toBe("application/problem+json");
423:       expect(response.headers.get("cache-control")).toBe("private, no-store");
424:       for (const name of [
425:         ...contextNames,
426:         "etag",
427:         "last-modified",
428:         "location",
429:         "link",
430:         "x-bdp-development-limit",
431:       ])
432:         expect(response.headers.has(name), name).toBe(false);
433:       if (head) {
434:         expect(response.text).toBe("");
435:         expect(response.bytes).toBe(0);
436:         expect(response.headers.get("content-length")).toBe(malformedLength);
437:       } else {
438:         expect(problem(response, 400, "malformed-request")).toMatchObject({ retry: "never" });
439:         expect(Number(response.headers.get("content-length"))).toBe(response.bytes);
440:         malformedLength = response.headers.get("content-length");
441:       }
442:     };
443:     malformed(await exchange("missing-two", "GET", point, ["BDP-Minimum-Scope-Position", "pos-2"]));
444:     malformed(
445:       await exchange("head-missing-epoch", "HEAD", point, [
446:         "BDP-Minimum-Scope-Position",
447:         "pos-2",
448:         "BDP-Authorization-View",
449:         expectedContext[1]!,
450:         "If-None-Match",
451:         "*",
452:       ]),
453:       true,
454:     );
455:     for (let index = 0; index < 3; index++)
456:       malformed(
457:         await exchange(`raw-duplicate-${index}`, "GET", point, [
458:           ...tuple,
459:           tuple[index * 2]!.toLowerCase(),
460:           tuple[index * 2 + 1]!,
461:         ]),
462:       );
463:     malformed(
464:       await exchange("single-comma-value", "GET", point, [
465:         tuple[0]!,
466:         "pos-2,pos-2",
467:         ...tuple.slice(2),
468:       ]),
469:     );
470:     for (const url of [receiptRoot, pageUrl])

### scripts/transaction-installed-observation.mjs:1–96

1: /** Bounded inherited-IPC launcher. No request or observation emulation. */
2: import { spawn } from "node:child_process";
3: import { mkdir } from "node:fs/promises";
4: export async function launchInstalledTransaction(entry, configuration, signal) {
5:   signal?.throwIfAborted();
6:   // Explicit launcher-owned provisioning; reopen keeps the directory/create:false.
7:   if (configuration.create) await mkdir(configuration.directory, { recursive: true, mode: 0o700 });
8:   signal?.throwIfAborted();
9:   const child = spawn(process.execPath, [entry], {
10:     cwd: "/tmp",
11:     stdio: ["ignore", "pipe", "pipe", "ipc"],
12:   });
13:   let bytes = 0;
14:   for (const stream of [child.stdout, child.stderr])
15:     stream.on("data", (chunk) => {
16:       bytes += chunk.length;
17:       if (bytes > 65536) child.kill("SIGKILL");
18:     });
19:   const exited = new Promise((resolve) =>
20:     child.once("exit", (code, signal) => resolve({ code, signal })),
21:   );
22:   async function bounded(promise) {
23:     let timer;
24:     try {
25:       return await Promise.race([
26:         promise,
27:         new Promise((_, reject) => {
28:           timer = setTimeout(() => {
29:             child.kill("SIGKILL");
30:             reject(new Error("installed transaction process deadline"));
31:           }, 10000);
32:         }),
33:       ]);
34:     } finally {
35:       clearTimeout(timer);
36:     }
37:   }
38:   async function drainFailure(error) {
39:     child.kill("SIGKILL");
40:     try {
41:       await bounded(exited);
42:     } catch (cleanup) {
43:       throw new AggregateError([error, cleanup], "installed transaction cleanup failed");
44:     }
45:     throw error;
46:   }
47:   let closing;
48:   const close = () =>
49:     (closing ??= (async () => {
50:       try {
51:         if (child.connected) child.send({ command: "close" });
52:         const result = await bounded(exited);
53:         if (result.code !== 0 || result.signal !== null)
54:           throw new Error("installed transaction child did not close cleanly");
55:       } catch (error) {
56:         await drainFailure(error);
57:       } finally {
58:         signal?.removeEventListener("abort", onAbort);
59:       }
60:     })());
61:   const onAbort = () => {
62:     void close().catch(() => {});
63:   };
64:   signal?.addEventListener("abort", onAbort, { once: true });
65:   if (signal?.aborted) onAbort();
66:   let detachReadyAbort = () => {};
67:   try {
68:     const ready = await bounded(
69:       new Promise((resolve, reject) => {
70:         const abort = () => reject(signal.reason);
71:         signal?.addEventListener("abort", abort, { once: true });
72:         detachReadyAbort = () => signal?.removeEventListener("abort", abort);
73:         if (signal?.aborted) {
74:           abort();
75:           return;
76:         }
77:         child.once("error", reject);
78:         child.once("message", resolve);
79:         void exited.then(() => reject(new Error("installed transaction exited before ready")));
80:         child.send({ configuration });
81:       }),
82:     );
83:     signal?.throwIfAborted();
84:     if (!ready || ready.event !== "qualification.ready" || JSON.stringify(ready).length > 65536)
85:       throw new Error("invalid installed ready message");
86:     return { child, ready, close };
87:   } catch (error) {
88:     try {
89:       await close();
90:     } catch (cleanup) {
91:       throw new AggregateError([error, cleanup], "installed transaction startup cleanup failed");
92:     }
93:     throw error;
94:   } finally {
95:     detachReadyAbort();
96:   }

### apps/bdptest/test-support/transactional-execution-binding.ts:1–110

1: import { isDeepStrictEqual } from "node:util";
2: import { installedContractsDigest } from "./installed-contract-identity.js";
3: /** Source-only installed binding. Runner retains sole receipt observation authority. */
4: import { createHash } from "node:crypto";
5: import {
6:   createRawHttpScenarioTarget,
7:   type ExactHttpModeOptions,
8:   type ConformanceFixture,
9:   type ExecutableScenario,
10: } from "@bdp/conformance";
11: import {
12:   bindServedProfileConfiguration,
13:   assertSameServedProfileConfiguration,
14:   type TransactionalRunContext,
15:   type TransactionalSession,
16: } from "@bdp/conformance/testing";
17: export interface InstalledTransactionChild {
18:   readonly ready: Record<string, unknown>;
19:   close(): Promise<unknown>;
20: }
21: export function installedTransactionBinding(
22:   ready: Record<string, unknown>,
23:   sourceCommit: string,
24:   fixture: { bytes: Uint8Array },
25: ) {
26:   if (
27:     ready.profile !== "transactional" ||
28:     ready.history !== false ||
29:     ready.claimEligible !== false ||
30:     ready.authorityMode !== "canonical" ||
31:     typeof ready.scope !== "string" ||
32:     typeof ready.runtimeConfigurationIdentity !== "string" ||
33:     !/^[a-f0-9]{64}$/.test(ready.runtimeConfigurationIdentity)
34:   )
35:     throw new Error("installed Transactional metadata unavailable");
36:   if (typeof ready.listenUrl !== "string")
37:     throw new Error("installed physical session unavailable");
38:   const url = new URL(ready.listenUrl);
39:   if (
40:     url.protocol !== "http:" ||
41:     url.hostname !== "127.0.0.1" ||
42:     !url.port ||
43:     url.username ||
44:     url.password ||
45:     url.pathname !== "/" ||
46:     url.search ||
47:     url.hash
48:   )
49:     throw new Error("invalid installed dial binding");
50:   const targetNames = {
51:     createBead: "create-bead",
52:     updateBead: "update-bead",
53:     deleteBead: "delete-bead",
54:     createLink: "create-link",
55:     updateLink: "update-link",
56:     deleteLink: "delete-link",
57:     putAlias: "put-alias",
58:     deleteAlias: "delete-alias",
59:     updateWhere: "update-where",
60:     deleteWhere: "delete-where",
61:     batch: "batch",
62:     sequence: "sequence",
63:   };
64:   const operationTargets = Object.freeze(
65:     Object.fromEntries(
66:       Object.entries(targetNames).map(([name, suffix]) => [
67:         name,
68:         `${ready.scope}operations/${suffix}`,
69:       ]),
70:     ),
71:   );
72:   if (!isDeepStrictEqual(ready.targets, operationTargets))
73:     throw new Error("installed operation target inventory mismatch");
74:   const served = bindServedProfileConfiguration({
75:     scope: ready.scope,
76:     profile: "transactional",
77:     history: false,
78:     sourceCommit,
79:     installedPayloadDigest: ready.payloadIdentity as string,
80:     installedContractsDigest: installedContractsDigest(ready.artifactIdentity),
81:     policyDigest: ready.policyIdentity as string,
82:     authenticationDigest: ready.authenticationIdentity as string,
83:     fixture,
84:   });
85:   return Object.freeze({
86:     served,
87:     operationTargets,
88:     runtimeConfigurationIdentity: ready.runtimeConfigurationIdentity,
89:     authorityMode: "canonical" as const,
90:   });
91: }
92: export function createTransactionalExecutionBinding(options: {
93:   readonly baseline: ReturnType<typeof installedTransactionBinding>;
94:   readonly sourceCommit: string;
95:   readonly fixtureSource: { bytes: Uint8Array };
96:   readonly exactMode: ExactHttpModeOptions;
97:   readonly launch: (
98:     scenario: string,
99:     reopen: boolean,
100:     signal: AbortSignal,
101:   ) => Promise<InstalledTransactionChild>;
102: }) {
103:   let active:
104:     | (TransactionalSession & { controller: AbortController; pending: Set<Promise<unknown>> })
105:     | undefined;
106:   let generation = 0;
107:   let presentation:
108:     | { scenario: ExecutableScenario; scope: string; seed: number; fixture: ConformanceFixture }
109:     | undefined;
110:   let reopening = false;

### packages/server/src/transaction-http.ts:586–602

586:     const minimum = hasMinimumControls(request)
587:       ? currentReadMinimumHeaders(request.headers)
588:       : undefined;
589:     if (hasMinimumControls(request) && !minimum) return developmentGap("minimum-position");
590:     if (minimum && alias === undefined) {
591:       // Aliases have their own conditioned authority path, not a canonical
592:       // ReadRequest or a member of the current Resource/collection subset.
593:       const parsed = requestFromUrl(url, scope);
594:       if (isReadProblem(parsed) && parsed.status === 400) return problem(parsed);
595:       // Preserve the early unknown-token label on excluded valid routes and parser 404s.
596:       if (
597:         !transactionMinimumHeaders(request.headers) &&
598:         (isReadProblem(parsed) || currentReadMinimumKind(parsed) === "strict")
599:       )
600:         return developmentGap("minimum-position");
601:       if (isReadProblem(parsed) || !isCurrentTransactionMinimumRead(parsed, scope))
602:         return developmentGap("minimum-position-route");

### packages/server/src/transaction-read.ts:29–59

29: /** Current private minimum subset; full request validation remains with Read. */
30: export function isCurrentTransactionMinimumRead(request: ReadRequest, scope: string): boolean {
31:   if (isProxy(request)) return false;
32:   switch (request.kind) {
33:     case "resource": {
34:       if (request.resource !== "type")
35:         return request.resource === "bead" || request.resource === "link";
36:       // Locality is checked before Read captures the request. Do not execute
37:       // an ID getter or method here; canonical syntax remains Read's decision.
38:       const id = Object.getOwnPropertyDescriptor(request, "id");
39:       return (
40:         id !== undefined &&
41:         "value" in id &&
42:         typeof id.value === "string" &&
43:         id.value.startsWith(scope)
44:       );
45:     }
46:     case "properties":
47:       return request.resource === "bead" || request.resource === "link";
48:     case "collection":
49:       return (
50:         (request.collection === "beads" ||
51:           request.collection === "links" ||
52:           request.collection === "types") &&
53:         request.continuation === undefined
54:       );
55:     case "bead-links":
56:       return request.continuation === undefined;
57:     default:
58:       return false;
59:   }

### packages/server/src/transaction-read.ts:360–363

360: /** Validated TX interpretation of the generic private Read observation. A
361:  * continuation's snapshot is its original selected data; current is the new
362:  * entry that checked whether that retained page can still be served. Neither
363:  * is inferred from a mutation receipt, and this does not implement HTTP barriers. */

### apps/bdptest/src/transactional-matrix.test.ts:1597–1605

1597:   const prefix = `${direction}-survives-prefix`;
1598:   const links = direction === "incoming" ? ["l", "r"] : ["r", "l"];
1599:   const operations = [
1600:     {
1601:       operation: "updateBead",
1602:       bead: "beads/b",
1603:       propertiesChange: [{ op: "replace", path: "/title", value: prefix }],
1604:       metadataChange: [{ op: "replace", path: "", value: { phase: prefix } }],
1605:     },

### apps/bdptest/src/transactional-matrix.test.ts:1723–1729

1723:     expect(body.operations[0]).toEqual({
1724:       operation: "updateBead",
1725:       bead: "beads/b",
1726:       propertiesChange: [{ op: "replace", path: "/title", value: `${direction}-survives-prefix` }],
1727:       metadataChange: [
1728:         { op: "replace", path: "", value: { phase: `${direction}-survives-prefix` } },
1729:       ],
