# Metadata-only Link updates in incident Bead Event Sources

## Conclusion

The stored Event and its metadata delta are already valid and implemented. The open issue is the membership of that existing Event in an endpoint Bead's projection. Current normative text does **not unambiguously authorize broadening that projection to metadata-only Link updates**: its specific membership bullet says properties, while the general `updated` definition and wire schemas cover properties or metadata. This is a narrow normative reconciliation, not a missing Event type, missing metadata envelope, or missing journal fact.

Recommend the ruling: **An incident Link's `updated` fact appears in each in-Scope endpoint Bead's Event Source when either its properties or its common metadata changes.** This would permit the smallest repair, preserve the five verbs, and make an endpoint observer see all version-changing mutable content of its incident Links. It requires an explicit ruling/clarification to the specific projection rule before removing the private gap. I have not made that ruling or changed the spec.

## Exact source and inspection limits

Read-only source: `/Users/dbox/repos/wt/vickie/vickie-schema-property-change`, HEAD `61ed530ac6a460598e0e312f5d28ec0e6fa2a33d`, tree `22e448fe0f517c0d6300b732de55494bd70c7938`, matching main `581bad0de6f5eb098b920147f6a7e09809c553ea`. Working tree remains clean. Local source/spec/schema/test inspection only; no tests, builds, servers, network, agents, or source edits. Existing tests below are inspected assertions, not newly executed results.

## Exact failure and what already works

`packages/server/src/transaction-events.ts:257–264` recognizes an `updated` Event whose subject is a Link, gets its endpoints from retained same-group facts, and determines whether the requested Bead is incident. If incident and `data.change` is absent but `data.metadataChange` is present, it throws `TransactionEventProjectionGap("incident-link-metadata")` instead of returning a projection.

The failure occurs **after a successful mutation commit**. `transaction-batch.ts:687–734` suppresses all Event production for an actual semantic no-op; otherwise an update records the previous/fresh revision and every supplied properties/metadata patch. Its changes include the complete Link postimage. `:754–795` also emits the owning source Bead's separate `updated` fact, with `ownedLink` containing the same Link delta, when ownership applies. These committed facts are not absent or malformed merely because the Bead projection refuses them.

The common projector is used by finite reads and SSE packets (`transaction-store.ts:640–656,1503–1588,1590ff`). Consequently:

- Scope and Link Event Sources can serve the metadata-only Link Event.
- A finite incident Bead Event request traversing that group fails; the HTTP handler maps the explicit gap to a private500 (`transaction-http.ts:587–588`).
- A newly opened Bead SSE request that initially encounters the group can refuse before streaming. An already open stream that later reaches it terminates; reconnecting from the preceding Event encounters the same group and fails again.
- An owned source's valid own `updated` Event does not rescue the request: projection throws while processing the earlier incident Link fact.

Existing tests pin precisely these limitations: `transaction-events.test.ts:320–352,533–557` and `transaction-event-sse.test.ts:515–559`. Combined properties+metadata updates already pass the incident branch when `change` is present; a Bead's own metadata update is already included.

## What the law resolves, and what it does not

| Normative source | Settled meaning / remaining limit |
| --- | --- |
| `docs/specs/bdp.md:5368–5373` | Resource Event Sources are deterministic projections of committed Scope Events; IDs derive from checkpoint, ordinal and projection. This prohibits inventing a second mutation/log to fix membership. |
| `:5378–5380,5437–5441,5516–5522,5610–5613` | Metadata change is a Resource transition with one fresh revision and one `updated` fact; its delta can contain `metadataChange` without `change`. This settles Event creation and payload, not every Resource-source membership rule. |
| `:5547–5554` | Link sources include their own lifecycle facts. Bead sources include their own lifecycle facts, incidence facts, and an incident Link's `updated` fact “when that Link's mutable properties change.” This is the specific missing metadata condition. Treating the word properties as also meaning common metadata would contradict their explicit separation elsewhere. |
| `:5437–5466,5477–5489` | An owned mutation also emits an owning source Bead fact; the Link and source facts are separate and keep their original ordinals. The double-report example explicitly says property change, so it does not independently settle metadata-only incident membership. The source's own metadata-bearing owned delta is already required. |
| `:6138–6148,6174–6181,6225–6233` | Resource sources contain the subset required by the abstract model, ordered with stable source-local cursors; `updated` visibility requires the references in both pre/post states. This supplies no extra metadata-membership rule that overrides the specific bullet. |
| `schemas/bdp-v0.schema.json:2878–2920,2985–3041` (`ownedLinkDelta`, `updatedData`) | The canonical closed payload permits `metadataChange` alone, or both document patches; owning Bead form uses `ownedLink` instead. Payload schemas do not decide whether a fact belongs in a particular incident source. No schema extension is needed for inclusion. |

The specific bullet has a natural properties-only reading. The general definition does not compel metadata inclusion in every source, since sources are expressly subsets. Conversely, changing a private500 into silent omission would select that narrower behavior without resolving the apparent incomplete propagation of metadata through the spec. Root/product should make the one membership choice explicitly, not infer a consensus from schemas accepting the payload.

## Minimal decision and alternatives

**Decision:** Does the incident-Link `updated` membership rule cover a Link transition caused solely by common metadata?

1. **Include (recommended):** each in-Scope endpoint Bead includes the existing Link-subject `updated` fact for a change to properties or metadata. Owned source additionally includes its distinct source-subject `updated`/`ownedLink` fact. Clarify the projection bullet and its double-report example to name both documents. No new Event, revision, ordinal, schema member, or transaction is introduced.
2. **Exclude:** endpoint Beads omit a Link-subject `updated` fact unless properties actually changed; Scope and Link sources still include it, and an owning source still includes its own required `ownedLink` fact. Clarify that common-metadata-only transitions are excluded. This is a larger implementation question than testing `data.change !== undefined`: `transaction-batch.ts:729–732` preserves a supplied properties patch even when that patch leaves properties equal and only metadata changes. If exclusion means actual properties changed, define/derive that condition from committed transition facts; a supplied-patch-presence shortcut is not justified. Do not broaden this assessment into a store-format repair before choosing this alternative.

Neither alternative should keep an otherwise valid retained group unservable, synthesize `change: []` to evade the guard, convert the metadata fact to a properties fact, or mint new incident Events. Unknown administrative/erasure semantics are outside this decision.

## Smallest implementation boundary after an inclusion ruling

- Production: remove the metadata-only refusal in the existing shared `projectTransactionEventGroup` in `transaction-events.ts`, keeping the same-group endpoint lookup and all validation. Remove its now-unused error class/import/catch in `transaction-http.ts` if no remaining throw uses it (current search finds only this one). Do not swallow all projection errors; malformed retained groups must still fail closed.
- Tests: replace the explicit-gap expectations in `transaction-events.test.ts` and `transaction-event-sse.test.ts` with concrete projection, streaming and reconnect assertions. Extend the existing source-private installed Event test only for a bounded metadata scenario under ordinary installed configuration and exact payload/source identity.
- Normative alignment: separately authorized small spec clarification for incident membership and the owned double-report wording; no Event payload schema or evaluator change. Audit only catalog citations directly affected by that wording; do not manufacture whole-row credit.
- No changes to committed group bytes, revision allocation, changefeed contents, receipt retention, Event ID derivation, cursor grammar, page/packet limits, store format, or public admission. Existing page builders bound actual serialized bytes and the projection remains a subset of Events already bounded at commit. Do not raise budgets to implement the membership choice.

## Concrete acceptance discriminators after the ruling

Use a few existing typed Beads/Links and independent expected metadata/property values; retain current finite/SSE budgets.

1. **Finite membership:** update only Link metadata and check the same Link-subject fact appears once in each in-Scope endpoint's source, once in the Link source and once in Scope. Compare data, ordinal, transaction, time and revisions across sources; source-local ID/source differ by projection. A nonincident Bead must not receive it. Read actual Link state and verify the committed metadata, unchanged properties, and correct fresh revision.
2. **Owned and self-Link:** owning source gets Link-subject metadata update followed by its distinct source-subject owned delta with exact Link delta/provenance correspondence; target gets only the incident Link update. An incident self-Link update appears once, not twice (the two-fact rule concerns linked/unlinked). Unowned endpoint revisions remain unchanged; owned source revision changes for ownership, not because projection was read.
3. **Semantic boundary:** metadata no-op emits no new Event; effectful metadata plus an equal properties patch behaves the same membership-wise as metadata-only; a true mixed update retains both supplied deltas. Failed/rolled-back update yields no fact. These distinguish semantic inclusion from a `change`-presence workaround.
4. **Live SSE:** open from a known existing Bead-source cursor, commit metadata-only update, then a later ordinary update. Observe both actual frames in order and compare decoded Events with independent finite results. The stream must remain usable rather than close at the former gap.
5. **Reconnect and restart:** reconnect using the preceding cursor and observe metadata update exactly once; reconnect using its newly returned source-local ID and observe only later facts. Include same-group later ordinal (e.g. owning source delta), not only another transaction. Reopen store and assert stable projected IDs/data; no live-Link lookup may change old membership after subsequent Link deletion. Foreign-source cursors remain rejected by the existing checks.
6. **Sensitivity controls:** omission of the incident metadata fact must fail exact membership; duplication fails cardinality; missing metadata delta or changed revision/ordinal fails independent payload relations; terminating at the old gap fails observed later-frame/reconnect assertions. Scope/Link-only success must not satisfy Bead assertions.

These are a proposed bounded repair and its discriminators, conditional on the explicit membership ruling. They neither qualify the whole Event contract nor the Transactional profile.
