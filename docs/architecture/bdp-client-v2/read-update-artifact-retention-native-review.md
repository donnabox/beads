# Native independent review: RU artifact retention

Reviewed frozen `8e3781ca78bc69571f6baa0e27d8d1f593986694` against `cbecfeac9b2279e7424d2499d558e976ff3f9dcf`. Author checkout was not modified. Independent detached review checkout: `/Users/dbox/repos/wt/vickie/vickie-read-update-artifact-native`, clean at the exact source head. Read the revised plan, complete six-file diff, relevant existing startup/compiler/caller code, and finalized author report. Did not read the external code review or external plan review.

## Result

No blocking correctness finding in this bounded change. The implementation matches the approved capture/inventory scope and the disclosed compatibility break. This is not whole-profile qualification or a universal pre-maintenance startup gate.

- `recovery-store.ts:290–326` snapshots the new own data expectations before file provisioning; the inventory is detached and frozen. Explicit uncaptured text is `null`. Missing metadata differs from absence, and default low-level reopening cannot bypass a captured installation.
- The authority serializes the branded immutable compiler capture under the finite aggregate-derived ceiling before opening storage (`development-authority.ts:183–200`). Both configured and built-in paths unconditionally supply their complete Type inventory. The sole non-test store caller is this authority; shared accessor/compiler and TX code are unchanged.
- Under the existing exclusive startup transaction, ordinary identity checks precede exact capture byte length/text and complete Type inventory checks (`recovery-store.ts:630–656`). Capture length uses UTF8 BLOB length before text materialization. Count plus every ID/body comparison detects missing, extra, replaced, or changed Types. Capture JSON is not parsed in storage and no callback enters SQLite.
- These checks precede claimed-key deletion and the startup write/commit (`:674–678`), and therefore precede authority expiry (`development-authority.ts:213`). Writable SQLite journal recovery remains possible first; the actual hot-journal regression verifies this. Resource-Type and retained-runtime compatibility checks remain post-open as documented.
- Capture is retained separately from compact installation identity and remains outside member-writable policy. Exact raw bytes and all compiler-limit witnesses persist for the installation lifetime. Existing compiler identity and hosted-schema paths stay unchanged; offline reconstruction retains validation/replay behavior. Built-in literal descriptors deliberately do not acquire whole exact-artifact conformance credit.
- Format 3→4 refusal affects the nonattesting development facade too; docs disclose the lack of migration/reseed. Physical byte identity is asserted for ordinary refusal, not across necessary SQLite rollback-journal restoration.

## Independent executable evidence

Pinned Node24.16.0 and pnpm11.20.0. Frozen install and build completed successfully before subprocess/emitted tests. **212 tests /5 actual files PASS in6.50seconds**, maxWorkers2: recovery-store, development-retention, sequence-executor, read-sequence-lifecycle, and installed-reference. Includes all new corruption/lifetime cases, real hot-journal mismatch/recovery, existing schema hosting, changed policy/credentials, and relevant ownership/recovery paths. Log `native-tests.log`, structured report `native-tests.json`. The command also supplied a nonexistent development-authority.test.ts filter; Vitest's actual five-file count above is the evidence, not six files.

Two additional external emitted-module probes passed:

1. `native-predecessor.mjs/.log`: actual frozen cbec predecessor writer produces format3; review build produces format4. Both cross-version reopen directions refuse with byte-identical main file and abandoned claim intact. No synthetic version-marker edit.
2. `native-same-length.mjs/.log`: same-byte-length capture substitution and equal-count replaced Type ID both refuse byte-identically before pending cleanup; the correct expectation then recovers exactly one claim. This independently exercises equality after the length/count fast checks.

Typecheck, dependency boundaries (11 packages), unchanged89-row Read evidence verifier, and git diff check passed. Logs are `native-{install,build,typecheck,boundaries,evidence}.log`. No tests were interrupted; all invoked processes completed and disposable probe stores were removed. Source checkout remains clean.

## Limits

No full workspace run, new packed observations, or both-package smoke rerun was performed in this native review; the author reports those separately. Saved capture read is bounded by expected UTF8 length. This does not add a universal cap to low-level caller text or canonical Type-row materialization, nor does it move the explicitly excluded Resource-Type/runtime-limit refusal classes. Administrative configuration is still required to reopen; this is retained evidence/offline reconstruction, not a new auto-loading installer or migration facility. Root owns final composition, broader validation and publication.
