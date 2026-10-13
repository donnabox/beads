# Schema provenance clarity: native review

Reviewed the stable, uncommitted four-file diff in `/Users/dbox/repos/wt/vickie/vickie-schema-provenance-clarity` against `d907a1e1576c5795c17edb5f15528f47edfd7d6b`. Reviewed diff SHA-256: `10d5f26aeecc02d7a2d53f9731bce46fe239655179c2d77708f91b84813d4972`. No source edits or tests in the root checkout. Verdict: **A–F resolved; no additional actionable defect found in this bounded diff.** Root owns execution validation and eventual commit identity.

- **A:** The progress paragraph now accurately separates exact-configuration route pins from artifact byte binding. `runner.ts` hashes Scope/profile/routes/request ceilings (plus a synthetic-manifest digest when present), not descriptor/schema content. Installed contract/configuration identities and capture assertions cover the latter separately.
- **B:** Packed evidence is explicitly pinned to original `2c10d19` and combined `d907a1e`; no successor is relabeled. The stated d907 fresh packed 112-test run is supported by `replication-integration-validation/report.md` and its cited packed artifacts. The text explicitly requires later changes to receive their own validation.
- **C:** The two affected row-local reasons now state that properties document omission is an installed choice, not a mandatory universal spelling. A parsed JSON comparison confirms every other inventory value is unchanged. No status, binding, whole flag or gate changes.
- **D:** The added `contracts.types[id]` equality compares actual canonical descriptor strings between the explicit cohort and independent empty-artifact default capture. `development-contracts.ts` builds this map using `stringifyJsonValue(descriptor)`, so this assertion is stronger than object equality and directly supports byte-identical remapped shipped descriptor wording.
- **E:** The text identifies the eleven probes across three rows and names the diagnostics row explicitly as the subject of the bounds observations. It retains the partial/truncation limitations.
- **F:** The new schema-document branch now checks that its diagnostic exists before assigning metadata. It removes that newly added non-null assertion without changing neighboring controls or allowing a missing diagnostic to pass silently.

Read-only diff check passed. This review does not re-attest historical unlogged commands discussed in external finding G; current root validation is separate. No runtime/schema/public admission or evidence artifact changes occur in the four-file diff.

## Event successor integration boundary

Separately inspected frozen `dd3c2c081208ef0da262abbc9d0158ce04c46587`, direct child of d907. Its exact files are:

- `packages/server/src/transaction-events.test.ts`
- `packages/server/src/transaction-http.ts`
- `docs/design/transaction-event-review-closure.md` (new)

These do not overlap the four schema-clarity files. The HTTP delta reuses an existing alias classification; the rest is focused evidence and its disclosure. Shared `docs/design/transaction-implementation-progress.md` is untouched. Event source remains clean and immutable; no integration or outward write was performed by this review.
