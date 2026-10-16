# Native receipt projection failure audit

Two mandatory projection defects, with exact existing-test fallout. This audit reviews final-code-v2 (e0bbb4bf) and preserves its actual failing v3 transport result. No product execution or edits were performed by this reviewer.

## P1 — Medium: whole-withheld available receipts retain prohibited expiresAt

transaction-authorization.ts:102–118 changes completed available receipts with nonempty all-withheld results to detail withheld and deletes results/next/allocated, but leaves expiresAt. mutationReceipt oneOf branch4 explicitly forbids expiresAt at schema4089. The preceding branches cannot accept a completed withheld receipt either. The actual source-validation-v3 transport.results.json reports this precise oneOf/4/properties/expiresAt false-schema failure at the new T1 receipt observation. The strict schema assertion correctly discovered a production response-shape defect; do not relax it.

Required correction: whole-withheld root projection also removes expiresAt. Preserve body historical id/status/key/epoch/view/transaction/required/effect positions and current response context. Existing exact expected-object fanout: transaction-runtime.test.ts708–712, transaction-authorization.test.ts319–324 and509–516. All three presently duplicate the same incomplete deletion list and must remove expiresAt too.

## P2 — Medium: nonempty fully withheld compact allocations never collapse

transaction-authorization.ts:96–101 maps every allocated identity to withheld but only the detail-available/results branch can select detail withheld. A real completed expired receipt with one or more allocations consequently remains detail expired with every allocated entry withheld. Spec5240–5250 first requires allocation reauthorization, then requires a completed receipt whose every entry is withheld to carry detail withheld and no results/next/allocated. The expired schema permits per-allocation withheld values in general, but that structural permission does not override this all-withheld semantic rule. Under this fixed empty policy, every nonempty compact allocation is hidden.

Required correction: apply whole-withheld root projection to nonempty all-withheld allocated entries on completed receipts as well, deleting expiresAt/results/next/allocated. Do not collapse an empty allocation list through vacuous every(): a completed update/no-op/alias transaction can legitimately have no compact Resource allocations and remain expired. Do not change failed receipts or durable stored detail; projection is delivery-only.

T9 at authorization.test.ts412–422 currently requires the wrong expired/all-withheld representation. Its actual batch is create a/delete a/create b. Before expiry, the non-record deletion survives, so the mixed root remains available and the issued page remains a page with its exact withheld Resource entry. After real retention compaction (store469–539), only creation allocations0/2 remain; both are withheld, requiring a whole-withheld root. Retain the real expiry/read path and strict historical-field/context equality. Pin actual retained compact allocations if inspecting storage, and assert their absence from delivered root; preserve the existing empty-results nonvacuity guard and cover empty allocated without adding another runtime callback. No request or runtime-cap growth is needed.

## Boundaries checked

Store.observeReceipt calls the shared projector before root-prefix slicing; receipt() also calls it, so repairing that helper covers root delivery without schema/API expansion. Issued pages at store2259–2263 project each entry within their retained boundaries and must remain page-shaped; do not collapse an entire page into a receipt root. Sequence response uses entry projection and has no root expiresAt field. Failed receipt problem/non-record execution facts are retained. The projection must not rewrite stored receipts, receipt-page plans, retention timestamps, creator identities or sequence dependency resolution.

Normative anchors: spec5140–5175 defines available/expired/withheld and expiry compact facts;5231–5254 defines every delivery/current authorization and the all-withheld rule. Complete mutationReceipt schema3896–4092 supplies five exclusive branches. No new product decision, erasure implementation, partial-policy mode or runtime-limit change is proposed.

The original full native review missed P1 and P2; the strict actual T1 schema check exposed P1 and this source audit established P2. Prior reports remain immutable; these are newly established findings, not retroactive passing evidence. Root/author own repair and validation.

Reviewed binding SHA `e0bbb4bf08afbc39b86fc04440ac91c1651f1bbed813d544c1d391de4b707491`; patch `cfac763efe463fe61a9e16780b40c615dddc1ffc4768c4e2d6ebbf356ca96937`. Actual failed transport result SHA `436074562781bf83a7d0501fca65368c1b8b31e6da5101752b00b387da38af27`.
