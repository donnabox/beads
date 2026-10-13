# M1 external review closure

Frozen clean local commit: `80d9ef0941a6cb7d46c4b6ad18cab76a0fcfcd4f`, direct child of published integrated `315b04ec40f9ba78e3db9ee0b5e10e1c441c7fe6`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-runtime-closure`.
Branch: `codex/vickie-transaction-runtime-closure`; origin `https://github.com/gastownhall/bdp.git`.
Only transaction-runtime-configuration.ts and its test changed. No Snapshot, H1b, HTTP normalization, store format, public admission or qualification changes.

## Findings disposition

**F1 — fixed.** Each credential digest now hashes `{revision: "transaction-credential-1", token}`. Actual emitted runtime authentication metadata is tested against the versioned association preimage. Domain separation does NOT prevent guessing low-entropy tokens; metadata digests are not secrecy guarantees. Existing predictable private fixture tokens demonstrate there is no claimed entropy provenance validation. H1b must generate actual installed-run credentials using a CSPRNG. This alters credential/authentication/policy/runtime digests by design; stored configuration identity/format is unchanged.

**F2 — receiving allegation false; scheduler fixed.** Both exact predecessor 31aab4d and current transaction-submission.ts:62 declare the comparison parameter `Omit<NonNullable<Parameters<typeof compareAndAdmitTransaction>[4]>, "receiving">`. TransactionServiceOptions derives that exact parameter, and TransactionRuntimeTestSeams derives the service option. There was no type-accepted receiving seam at either source, and none is added. Scheduler is typed turn-only and service consumes only turn. Capture now rejects extra delay (including own undefined) before provisioning; a positive turn-only capture remains accepted. Existing scheduler caller regressions pass.

**F3 — covered.** Root reported the exact integrated 315b04 full suite passed 8,427 tests, including all callers. This closure additionally ran transaction-owned, transaction-events and transaction-changefeed tests alongside configuration/runtime/HTTP/service: 150 tests across seven files, all passing. Root's full run is predecessor evidence, not relabeled as this source's full run.

**F4 — disclosure corrected.** The metadata descriptor-closure check is a defensive invariant over the capture's own derivation, not an independent installation proof. The store's persisted-configuration byte comparison is the actual reopen binding. The policy identity includes the fixed all-visible policy description and live view; it depends on credential associations and is not independent policy proof. Forged-input tests exercise defensive checks only.

**F5 — fixed.** Frozen metadata returns the captured authorityMode. Real default compiled canonical and listener runtimes prove different runtimeConfigurationIdentity values with equal artifact, policy and authentication identities. Actual Host behavior is also exercised for both modes.

**F6 — explicit private restriction.** Actual socket tests prove mixed-case Canonical.Test and trailing-dot canonical.test. are refused with 400, alongside foreign/listener and explicit default-port hosts. Comment documents exact-byte private Host matching; no normalization behavior is added, and complete HTTP hostname-equivalence qualification remains open. H1b trusted dial mapping must preserve the canonical Host bytes.

**F7 — no new machinery.** Recursive freeze receives only bounded detached descriptors, compiler output or internally constructed records in current code. The review's cyclic input path is unreachable. No speculative cycle framework added.

**F8 — scope and evidence clarified.** Existing leaf tests hash modified defaults directly; their contribution to runtime identity follows the implementation embedding settings wholesale. They are not real served overrides or direct per-leaf runtime differential observations. The authority-mode differential added here is actual served metadata. Repeated numeric callback-range bounds are not a correctness defect; broader deduplication is deferred.

## Verification

Pinned Node 24.16.0 / pnpm 11.20.0. Frozen install, clean build, typecheck, repository format check, dependency boundaries, changed-file lint and git diff check exit 0.
Focused tests: transaction-runtime-configuration, transaction-runtime, transaction-http, transaction-service, transaction-owned, transaction-changefeed and transaction-events. 150 pass, seven files, zero failures/skips, maxWorkers=2; 3.99 seconds Vitest duration. Actual JSON/verbose logs retained alongside this report. Build preceded emitted-runtime and child-process tests.
No full suite or packed smoke rerun for this two-file closure. No interrupted/timed-out commands. Vitest cleanup inspection retained in cleanup.log; all owned commands exited. No push/PR/bus/launch writes. Parent owns review, ordinary integration and any publication. Frozen M1 and 315b04 checkouts unchanged.
