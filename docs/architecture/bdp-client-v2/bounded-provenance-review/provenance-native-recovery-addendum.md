# Native Codex addendum — incompatible recovery-store refusal

Reviewed exact commit `90228b35015f124a242345e10c4e8557289314fa` against `6b64c1cede8e434ae6e09781b372a8c899f7bc91`. The delta contains only `packages/server/src/recovery-store.ts` and its test file. No source edits were made; focused tests ran in a private external `git archive` snapshot under Node 24.16.0.

**No actionable finding. The bounded repair is sound.**

The store-format discriminator advances from 2 to 3 because retained Resource/outcome and request-identity representations changed. On reopen, the existing format check runs before allocation integrity checks, claim sweeping, startup metadata write or commit. A mismatch throws `store-mismatch`, rolls back and closes. New explicit provisioning still uses exclusive file creation, so passing `create` cannot overwrite or silently reseed an existing store. The existing JSON codec/function name remains unchanged appropriately: this is a representation-compatibility boundary, not a new JSON syntax interpretation.

The added test exercises old-format retained, expired and pending key states together, retained legacy Resource bytes, whole SQLite-file byte identity after rejection, and refusal of explicit reseeding. Existing format-1 rejection and ordinary current-format recovery tests remain in the focused suite. This prevents an old store from being partly recovered under incompatible representation semantics; it deliberately refuses migration rather than claiming to convert historical outcomes or identities.

The focused `recovery-store.test.ts` suite passed independently: **80 tests / 1 file**, with output in `native-recovery-tests.log`. Prior provenance source coverage through `6b64c1c` remains unchanged. No full-suite repeat, fresh cohort execution, write-profile qualification or public readiness claim is implied by this addendum. Final evidence qualification at `90228b3` remains root-owned.
