# Comparison-timeout context: frozen implementation and bounded evidence

Frozen source `30035b88380b7cbb3abd5359787db236b70345c7`, tree `cdee4bc7e31511cde094d17c045ea618163b6d7a`, ordinary parent `c555fe01df5cd20454fb0d663431df816c1270aa`. Checkout `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-context`, branch `codex/vickie-comparison-timeout-context`. Working tree is clean. The eight authorized files change by 557 insertions / 35 deletions. Local commit passed the remote session guard and includes `Agent-Signature: Vickie (Codex)`. No push, PR mutation, merge, coordination post, or public activation was performed by this worker.

## Result and boundary

A healthy finite comparison deadline in observed submission now captures the current persisted Scope tuple in one synchronous, metadata-only owner transaction. The expected epoch is checked inside that transaction; stopped/changed-epoch outcomes do not manufacture a current tuple. The HTTP adapter emits the existing 503 Problem and Retry-After with the captured context. Required context missing from a private deadline result fails through the existing sanitized 500 path.

The original epoch-before-deadline ordering, original finite deadline, turn-time dependencies and unobserved result shapes remain intact. The observation callback is excluded from inherited comparison configuration and attached only by observed submission; even an untyped extra callback is overridden. No owner entry or database transaction spans a wait. Store faults reaching observation propagate. No service, runtime, parser, schema, fixture, catalog, public export or profile flag changed.

`source.json` records every final file hash. All production and test files equal the accepted preflight bytes. Only the owned progress document was updated after successful validation to state the actual bounded evidence.

## Authority and retained draft corrections

Execution follows `root-measured-acceptance.md`, accepted plan v2 `941f350fb4634a77000f3bc0108f3ea80c94ee5701aef94bdc456c66a28ad4e6`, dispositions `994e5fa3195a2515db8fcb9b8bfd6f7ad30f8ba5557909230408e29a0a4155db`, native closure `b2a637e8f4d951a6fb0166198822ee8fe55a3a7687ca6690b3c059da7ab17d2c`, and final preflight closure `18dc132e`.

The accepted final preflight hashes remain unchanged: report `79aae00ec518107c72bc904abdd148c842036d3cdfdeafb640fdac859fc80252`, bounds `74200eb1f9d125a8a7856c2c1568ce4553ab198241b46402dac5fa7970e93054`, diff `97b5bd7840fb1101622fbedde0d037d7e40479b98316ddf5f2544967fb607b0e`.

Original draft records are preserved in `initial-preflight/` and `pretyping-preflight/`. Static review corrected a masked control to mutate the submission boundary, and corrected the optional callback annotation / derivative cast for exactOptionalPropertyTypes. These were unexecuted draft corrections, not failed runtime attempts. The originally proposed complete changes-page oracle was replaced, with root approval, by complete Scope Event pages with independently pinned content and `next:null`; this does not prove change-group pagination.

Final-summary correction: the first assembled summary inherited `executed:false` from preflight for controls despite retaining their real execution outcomes. That report-only error is preserved in `summary-before-execution-flag-correction.json`; `summary-execution-flag-correction.json` binds the corrected `executed:true` record. No execution or source changed for that correction.

## Source execution

Pinned Node 24.16.0 and pnpm 11.20.0. Exact argv, cwd, start/end time and exit status are retained in each `*.exit.json`; `source-1.json` through `source-5.json` retain Vitest results. Tests were explicitly selected, not a full-suite run. Between 2026-10-14 17:54:02Z and 17:54:09Z, **18 selected tests passed, 287 were filtered out, zero failed**.

| File | Passed | Filtered |
| --- | ---: | ---: |
| `packages/server/src/transaction-store.test.ts` | 6 | 133 |
| `packages/server/src/transaction-submission.test.ts` | 5 | 31 |
| `packages/server/src/transaction-http.test.ts` | 2 | 68 |
| `packages/server/src/transaction-dependencies.test.ts` | 3 | 22 |
| `packages/server/src/transaction-sequence-alias.test.ts` | 2 | 33 |

The source HTTP cohort used 16 requests / two listeners: the rewritten scenario asserts 12 dispatched and completed exchanges; the preserved basic case has four requests. The new scenario independently commits during comparison wait, observes current pos-1 on timeout, proves no timeout effects through complete graph/Event state and receipt inventory, then releases the actual sequence and checks retained retry / final pos-3 state. Event counts are 0 → 1 → 5, including exact content and complete-page witnesses.

Six new owner cases cover one atomic metadata observation, final epoch mismatch, normal closure, reentrancy, commit uncertainty fencing and metadata failure propagation. New submission cases cover independent progress, defensive epoch stop, and exact legacy shape/caller-callback exclusion.

Static/setup checks:

- `boundaries.exit.json`: exit 0.
- `build.exit.json`: exit 0.
- `format-final.exit.json`: exit 0.
- `format.exit.json`: exit 0.
- `install.exit.json`: exit 0.
- `lint.exit.json`: exit 0.
- `read-seal.exit.json`: exit 0.
- `typecheck.exit.json`: exit 0.

Lint exited 0 with 666 warnings and 43 informational diagnostics; this is not a warning-free claim. Historical Read evidence verification passed without relabeling or resealing it.

## Counterfactual controls

Five isolated derivatives were built and executed once each, between 17:54:25Z and 17:54:39Z. Each build exited 0, each selected test exited 1 at the intended named assertion, and each adjudication passed. The original source was unchanged. Patches, exact executed mutant hashes, raw results and named failure stacks remain under `controls/`.

| Control | Intended inner assertion | Build / test exit | HTTP requests to failure |
| --- | --- | --- | ---: |
| `remove-deadline-wrapper` | `deadline context presence bdp-scope-epoch` | 0 / 1 | 4 |
| `stale-deadline-position` | `deadline current position` | 0 / 1 | 4 |
| `lose-required-tuple` | `healthy deadline response` | 0 / 1 | 4 |
| `leak-unobserved-fields` | `legacy comparison result exact` | 0 / 1 | 0 |
| `omit-final-epoch-check` | `final deadline epoch check` | 0 / 1 | 0 |

The three HTTP derivatives each reached four exchanges, then drained the held sequence and listener through the existing finally path. Total control HTTP usage was 12, within the accepted maximum 36, with three listeners and no installed children. Counts follow the exact selected path and named failure location, not a separate raw wire recorder. Vitest records retain the intended inner AssertionError rather than relying on a generic AggregateError label. The legacy-shape control proves submission-boundary sensitivity to an enumerable undefined context; it is not a direct-comparison structural-equality claim.

There were no unexpected source, build, static or installed failures and no automatic retries. The five nonzero test exits are expected counterfactual evidence, not green tests.

## Fresh installed regression evidence

After the clean commit, `packed.py` rebuilt, packed and installed the frozen source, then ran only the two approved existing regressions. All build/pack/install commands exited 0. Direct refusal ran 17:56:08–17:56:09Z (14 requests / one child); sequence refusal ran 17:56:09–17:56:10Z (13 requests / one child). Both tests and their independent postprocessors passed. Total **27 requests / two children**, no retries or predecessor cohort. Both actual observations retain `claimEligible:false`.

- Protocol tar SHA256: `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`.
- Server tar SHA256: `b9ebfae7b6fd2f5602688fadf0110da3749d5af012fdf4599867c6af2f84bf35`.
- Installed entry SHA256: `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`.
- direct: payload `2c62c66724c2c48fcd1de19360abbb90a3d5bb7c37257a095af752ba9e0b322d`, configuration `6fc8e98500180438f54bd0197b095ed5bb4fc09475ad770c658428bf0893b8b4`, contracts `1f369030c8696ff445c235b24637cc1ac41f0ab724a1b2402511347f74da344f`, fixture `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`.
- sequence: payload `2c62c66724c2c48fcd1de19360abbb90a3d5bb7c37257a095af752ba9e0b322d`, configuration `d1718dcaff57ee0b51474b894f13e52ebd868e99416439f00f75553eac92bc2d`, contracts `1f369030c8696ff445c235b24637cc1ac41f0ab724a1b2402511347f74da344f`, fixture `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`.

`packed/source.json`, `packed/summary.json`, `packed/exits.json`, test JSON/logs and `*-postprocessing.json` retain exact source/package identities, canonical route observations, body/wire fingerprints, current contexts and actual outcomes. `declaredSourceCommit` is **supervisor-verified source attribution bound to a separately built package, not child-attested provenance**. The launcher does not independently verify this caller label.

The installed tests preserve their existing 5s request, 90s scenario, 120s outer and 10s close limits, plus existing body/header/response/target caps. These are ordinary installed regressions; they do not observe the default installed comparison timeout. Evidence retains fingerprints and contexts, not complete wire/body transcripts.

## Cleanup and remaining limits

Both installed children exited normally (PIDs 82412 and 82445, exit 0, signal null). Their absence was independently checked after execution and again during the final audit; packed supervisors are also absent. Each test reported actual temporary directory absence. The external postprocessor validates that report but does not independently inspect a separately retained directory path. No owned test/build/service process remains. Disposable derivative and package evidence directories remain as review artifacts, with no active processes.

The source response-size estimate (39,936 bytes largest / 479,232 bytes aggregate) remains prospective, not measured heap or an enforced source transport cap. Live epoch rotation is unsupported; mismatch tests are defensive. The earlier comparisonEpoch closed shortcut after background maintenance failure remains outside this repair. No public activation, cumulative qualification, optional History, Snapshot or product-ruling change is claimed.

Final record hashes:

- `source.json`: `562896d01e09f070ee741c40b781973971310443a09fc35f9509507fdca172e5`.
- `summary.json`: `368397ba571a32f249b3b7cc0346a310febc993c8d78722d33a1ee90ccb0d4cc`.
- `cleanup.json`: `62618dee6b325d3b404b4e0c12b6b1d1f18f39725e6c22bd2e3abd05d636d128`.

Root owns code council, publication and integration. Parent PR119 later merged with the same c555 tree; this candidate remains the ordinary c555 descendant above, with no execution attribution rewritten to the main merge commit.
