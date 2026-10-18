# Native S1 Snapshot plan review

Reviewed `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-snapshot-plan.md` against frozen G1 `a7430ddf0fc2e56b0923d8c75fcfc752ee89ac88` and its unchanged normative spec/schema, with installed predecessor `e2f1385aa89bc6ae27e18a802a2a651a8f02e87d` as the stated integration baseline. Plan only: no source edits or tests; no new external review consulted.

**Proceed with durable capture groundwork after the two plan refinements below.** The overall boundary is appropriate: one owner, retained finite streams, actual changefeed rendezvous, no Snapshot-driven mutation admission, and no public profile claim. Q1 is a real specification interpretation gap for retained Snapshot reads; Q2 is distinct bookkeeping policy with a normative refusal constraint. Neither should block independent capture/paging work, but neither may be silently counted as complete conformance.

## Required plan refinements

### P2 — distinguish independent completeness evidence from what a generic replica can detect (plan line83)

The proposed negative oracle promises rejection for arbitrary omitted pages/members and mixed valid Snapshot identities. The actual stream wire form has only `items` and `next`; the manifest carries no complete stream count, member digest or signed membership proof (`schemas/bdp-v0.schema.json:4784–4852`, bead/link collection definitions). A generic reader can reject repeated URLs/IDs, owned-record disagreement, a Link missing its in-Scope endpoint, and inconsistent server-supplied continuation association. It cannot detect a valid-looking isolated Bead omitted from an otherwise coherent stream or a premature null next solely from that payload. Likewise an injected coherent alternate page is not necessarily distinguishable by its item bodies alone.

Keep those useful negative controls, but split their authority: server persistence checks verify recorded counts/ranges/identity bindings; fixture-derived expected inventories and an independently captured state oracle detect arbitrary omission/mixing; the replica itself rejects only observable protocol/graph contradictions. Do not add new manifest fields or pretend a test's expected inventory is information a cold generic client receives. The valid rendezvous run must compare the complete staged graph to independent canonical authority reads, as already planned. This is an evidence-scope correction, not a request for a cryptographic extension.

### P2 — do not import receipt principal isolation into Snapshot authorization (plan lines70/86)

“Foreign principal/view” conflates two different cases. Snapshots are bound to Scope epoch and Authorization View; view/history tokens do not identify the principal (`docs/specs/bdp.md:3697–3716`). A caller must independently be authorized; possession of a handle confers no access. Receipt principal isolation is explicit at `:4964`, `:5092–5096` and its404 table at`:4602`, but no equivalent mandatory same-creator-principal Snapshot rule exists. Snapshot law uses unknown or *undisclosable*404 and authorized expired/unavailable410 (`:5899–5916`).

In this fixed all-visible fixture, two authorized principals sharing the same view should not be asserted to differ merely because one did not create the handle. Test missing credentials, genuinely undisclosable/view-incompatible use, and authorized refetch separately. If root deliberately adds a creator-only administrative policy, identify it as that policy and include it in configuration; do not present it as Snapshot law or infer it from receipt code. No new policy is necessary for the first fixed-view implementation.

## Q1: minimum-position controls versus immutable Snapshot anchor

The spec resolves several premises exactly:

- One immutable manifest anchors both streams to one position, and replay is strictly after its checkpoint (`:5742–5755`).
- Manifest refetch returns the same identity, anchor, checkpoint, expiry, ledger and initial pages (`:5899–5907`); stream pages never move to a newer anchor (`:5923–5936`). A supplied minimum cannot justify replacing the Snapshot.
- Minimum reads must return a representation at that position or later, wait/route, or return the listed context/timeout Problem; never report success with an older position (`:5779–5785`).
- Required *current* Scope context remains required on304/412; minimum/epoch/view/expiry checks precede conditionals (`:5801–5814`). An immutable body anchor does not make an old304 context permissible.
- Receipt responses explicitly distinguish current serving observation from retained body history (`:5233–5244`). This clause is expressly receipt/page-specific; it is useful precedent, not a Snapshot ruling.

What remains unstated is whether a retained Snapshot's minimum gates the current serving replica's prefix while preserving its older payload, or gates the captured Snapshot anchor itself. If it gates the latter, waiting cannot advance that handle and the precise refusal for a minimum newer than its anchor but older than current head is not specified. The general read language and Snapshot immutability do not explicitly reconcile this. Do not silently generalize the receipt clause or invent a catch-up timeout for a fully caught-up authority.

The plan should separate the header issue from minimum fulfillment: existing304/412 law already requires current context; the unqualified captured-header alternative in Q1 is therefore not a complete solution. Root should choose or clarify a consistent retained200/HEAD/304/412 contract before crediting that HTTP behavior. An explicitly unsupported private retained-minimum path is an honest interim gap. Creation minima are well-defined and can use the existing finite wait loop; retries must capture only after the successful eligible prefix observation. No protocol change or special new error is required to start durable capture.

## Q2: expired issued-handle disposition

The mandatory observable distinction is already clear: authorized expired or no-longer-available handles410; unknown/undisclosable404; inability to serve an unexpired promise is service failure (`:5904–5907`,`:5916`,`:5927–5930`). There is no Snapshot forgetting interval or statement that elapsed time turns an issued handle into an unknown one. Optional `retention.maximumSnapshotLifetime` describes the usable lifetime, not a tombstone-forgetting authorization.

The physical retention method and capacity are implementation choices. However, an arbitrary time-based switch from a still-recognizable issued handle to404 is not positively authorized by current text merely because the server chose to delete its row. Root should approve a concrete policy preserving the required classification, or explicitly keep beyond-horizon behavior unqualified pending clarification. Possible engineering approaches include bounded disposition storage with creation backpressure, or durable authenticated issuance/expiry information that can recognize expired issued selectors without per-handle bodies. These are options, not a new protocol mandate or an instruction to build a token framework. Tokens remain selectors and still require authorization.

This is not Q1's stale-representation question. Snapshot bodies may be released at expiry; the separate issue is how issuance/expiry can still be classified. Do not require immortal body retention, renew expiry, evict an unexpired promise, or invent a minimum forgetting horizon.

## Minimal mandatory contract and implementation seams

| Contract | Exact law | Planned bounded implementation and evidence |
| --- | --- | --- |
| Complete common-prefix state | spec5742–5747,5941–5955 | SQL capture under the existing synchronous owner entry/transaction; Bead and Link streams, owned-record agreement, all in-Scope endpoints. Type/alias records are not extra Snapshot streams. Independent complete fixture oracle. |
| Immutable identity/pages |5899–5903,5923–5936 | Persist exact projected records and fixed page plan; creation/refetch/continuation separated; mutation between every page; restart retains original body/anchor/expiry. No ordinary collection cursor substitution. |
| Snapshot→feed suffix |5750–5755,5932–5936 | Reuse G1 checkpoint encoder and one captured prefix. No pruning presently means every later group is retained. Stage both streams then apply actual finite groups; compare full state and owned closure independently. |
| HTTP creation/refetch |5831–5840,5909–5921 | GET creates; creation HEAD discloses no usable handle and need not materialize one; refetch HEAD never renews. Native auth, method-first405, negotiation, expiry and conditional ordering; distinguish unresolved retained-minimum interpretation. |
| Refusal/lifetime |5904–5907,5916,5927–5930 | Fixed persisted expiry; never-issued versus issued-expired versus corrupt promised data separated. Select concrete disposition policy before promising that distinction. |
| Ledger/authorization |5938–5957,3697–3706 | Actual empty ledger only under unchanged no-erasure fixed-view installation. No erasure or hidden-view credit. Same-view authorized callers are not automatically foreign. |

The proposed4096 rows/32MiB per capture, eight live handles/64MiB aggregate and4MiB typed pages are private engineering limits, not existing normative counts or proof of arbitrary-graph support. M32/E256 bound transactions, not the captured graph. Use actual wire serialization after canonical projection; include next/id/Scope framing, both initial pages, lossless values, and aggregate capture work. A normal legal Resource must fit one page under the actual integrated tuple. Check sizes before loading corrupt persisted bodies and measure representative workloads; do not infer heap safety from encoded bytes.

One owner transaction and durable captured rows provide the appropriate atomicity seam. Keep lossless JSON projection and current Resource shape rather than reinterpreting raw storage as wire identity. Capture failures must roll back retained state, leave mutation keys/receipts/groups/revisions untouched, and preserve existing uncertain-commit fencing. A committed but undisclosed capture may consume capacity until its fixed expiry; tests should account for this instead of assuming every interrupted GET allocated nothing. Hot-journal recovery must still precede configuration refusal; unchanged valid old-file bytes and legitimate recovery are distinct assertions.

The planned source ownership is appropriate. Start from reviewed integrated M32/Event/G1 source, then add capture/store tests; add transport and H1 installed evidence without a second owner. Keep public admission closed and explicit row-level incompleteness. The existing schema/catalog already establish mandatory Snapshot shape and laws; no new normative schema, pagination framework, Snapshot wire fields, Type lifecycle or optional History feature is needed for this slice.
