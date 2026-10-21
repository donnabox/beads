# Independent native review — frozen installed Event restart witness

Verdict: **two blocking witness defects; do not release execution as written.** No tests, imports, builds, services, package creation, or source edits were performed. This is source-based code review, not an execution verdict. The separately reported baseline smoke failure and conditional Claude timeout are not evidence about this unexecuted callback.

## N1 — default full-suite execution fails before the witness starts

`apps/bdptest/src/transactional-event-sse-installed.test.ts:322–332` requires both installed environment overrides, then rejects an entry inside this worktree. The default package script is `vitest run --reporter=verbose`; `vitest.config.ts` includes `apps/**/*.test.ts`. `.github/workflows/ci.yml:64–69` builds then invokes `pnpm test`, with only the bd matrix executable/require variables. It supplies neither BDP_TRANSACTIONAL_INSTALLED_ENTRY nor BDP_TRANSACTIONAL_SOURCE_COMMIT. The historical module-level fallback still constructs the local dist entry; the new callback will then unconditionally throw its explicit-override error. An already-green baseline full-suite run cannot cover this dirty addition.

Bounded correction: keep the same real five-request callback executable in ordinary development mode, using the established local-dist fallback and clear development/non-installed attribution. For installed mode, require both explicit entry/source values, an external entry, and later external artifact/provenance acceptance. A partial override must fail. Do not silently skip the new callback or describe local-dist execution as installed evidence. Root should reconcile this distinction in the successor plan/report and freeze the revised callback before review/execution. No shared helper or production change is needed.

## N2 — unattributed creation is required to return provenance that this runtime deliberately omits

The seed at test lines 354–366 supplies only operation/id/type/properties. The independent Resource expectation at lines 732–742 nevertheless requires revisionProvenance with committedAt present and agent/message undetermined; Event expectations at lines 764–769 carry that expectation forward.

The actual authority path is explicit: `transaction-store.ts:1330–1331` passes observeCommitTime but `recordChangeContext: false` into evaluateTransaction; `transaction-batch.ts:708–729` forwards those options to evaluateResourceMutation. `resource-evaluator.ts:840–856` sets needsContext only for the native-context flag or an explicit agent/message property. For this seed both are absent. `resource-evaluator.ts:1107–1126` therefore omits revisionProvenance from the Resource, and `transaction-batch.ts:771–777` includes it in created Event data only when the Resource has it. Existing evaluator tests distinguish omitted attribution from the explicit `recordChangeContext: true` native-default case. Thus the correct current runtime fails this callback during request 1; the restart/SSE observations are never reached.

Bounded correction: independently require the complete unattributed Resource shape for the unchanged 244-byte seed, explicitly verify absence of revisionProvenance, and require the matching complete created Event data without provenance. Preserve receipt revisions, time/cursor comparisons, schema checks, effectPosition presence, and all other independent expectations. Do not turn on native provenance, add seed provenance, broaden production behavior, or weaken the whole-Resource comparator merely to make the test pass. Recorded provenance remains a separate supported condition; this fixture does not request it.

## Remaining reviewed behavior

The intended semantic composition is otherwise sound against plan-v3 and the pinned spec: two explicit creates in one batch, a two-Event initial page with independent ordinals and receipt transaction/revision links, actual first-child normal exit, same-directory create:false reopen, exact original E1 suffix after actual opaque E0, Last-Event-ID E0 overriding query E1, and contextual foreign-source rejection against independently established Bead a. Context-bearing responses require pos-1 and stable epoch/all-visible view. No extra mutation, lookup, History, partial ACL, Snapshot-minimum or incident-Link ruling is introduced.

The finite helper pins five-second transport timeout, exact canonical Host and actual listener dial, complete write state, media/schema/context, and safe response fingerprints. beginRequest limits the callback to five attempts; success requires four finite responses plus the SSE observation and the exact session/method ordinal inventory. Launch logic permits exactly two child attempts and success checks two distinct actual PIDs. Launchers use IPC and introduce no hidden HTTP requests. The seed is 244 bytes, SHA256 `2ffd2dc1c98cf8ec53338127cfad4edba630ac386b22f6c693b3931626559f03`. Counts remain prospective until released and observed.

Both sensitivity controls use the same positive comparators with actual saved data: a cloned restored E1 with wrong ordinal, and the actual E1 frame compared against saved E0. They add zero network requests and no children. They are local comparator discrimination, not server rejection evidence.

## Lifetime, cleanup, and observation limits

I traced the actual raw SSE recorder and installed launcher. The code checks ready/status/media/context and the first complete frame, then synchronously checks nonaborted signal, live socket and elapsed time before intentional close with no intervening await. This is necessary because until may return on a buffered frame despite stored failure, and close itself does not expose stored failure. After drain it checks status again and requires exactly one observed id-bearing frame, fully equal to E1. This establishes the bounded observed-frame claim only, as the report states.

All raced promises receive fulfillment/rejection observers before the race. Launch-in-flight records exist before calls; a late returned child is retained and close is initiated, while stopUnknown stays true. Pre-handle rejection remains unknown ownership and preserves the store. First-child close is bounded by remaining work time and actual code0/signal-null is required before reopen. The 45-second signal path initiates stream close before the launcher abort observer starts child close.

Final cleanup computes one absolute ten-second deadline for stream and all child drains. closeSession is invoked even when bounded subsequently finds no remaining time, so stop initiation is not skipped. Close promises are memoized/tracked; timeouts and unknown ownership cannot become a pass through a later exit. Primary/cleanup failures aggregate, timers are cleared, and removal occurs only after no failures, two known normal exits, drained stream and no pending operations. The callback correctly retains the diagnostic store for semantic failure, second-launch failure and unknown/failed cleanup.

These checks do not prove helper termination inside the cleanup bound: launcher failure cleanup may itself wait longer, and pre-handle failures expose no callback PID/private child trace. The separate future supervisor must own the process group, fail on forced intervention/survivors, and bind actual fresh external artifacts. No such supervisor was launched or accepted by this review. Dynamic installed configuration digest import and schema validation have not been executed/typechecked here.

## Preservation and bindings

The original historical callback slice, including its trailing newline, is byte-identical to HEAD; SHA256 `2f95f8917bd75ae75b023f6899c8e24fa2548f34bc18ec5dd59a4fc8c1640ba2`. The diff contains two imports and one appended callback, with only that one tracked file dirty and no untracked files. Source HEAD is `ddfb5f43a058fe0f16ff9e04e528b4e7bfc03d87`; source file SHA256 is `73cbc5a2cf14a9d154d8f2d82d40ee392400e9901e9485a00cf2ac5ece3aef07`; full binary diff SHA256 is `37d024c3337005443fc72ff001d675faddfd3388f1fbc303b4d02cad9a1a9a2d`. Both source and diff pins matched at initial read and immediately before report mutation.

Reviewed plan-v3 SHA256: `8cc581a267ad5e88de93548ba7d6d73fc38da5df7d53a7e719dd54647e4e95da`. Author report/source-binding read as frozen unexecuted claims. Pinned spec sections on deterministic/source-bound IDs, exclusive after, and SSE Last-Event-ID precedence agree with the intended witness. This review does not qualify any profile or authorize main/release/Preview2 changes.

Source/reference hashes:

- `apps/bdptest/src/transactional-event-sse-installed.test.ts`: `73cbc5a2cf14a9d154d8f2d82d40ee392400e9901e9485a00cf2ac5ece3aef07`
- `packages/server/test-support/raw-sse-recorder.ts`: `09ff8343ac1a5dbad77eaf24fc2745cc28ea870ada6b05dc4ea1644c0ad5ef81`
- `scripts/transaction-installed-observation.mjs`: `6f772f3776d3c3effa788676d853cc2932a6ee59ad63ac3338ad1b487408914b`
- `apps/bdptest/test-support/transactional-execution-binding.ts`: `cde506c72a8710eafe58522900f6db70ed3701cf333560252dc8e7c5e706cf89`
- `packages/server/src/transaction-events.ts`: `6c5ab4482f0b00110a71a96b16d44ca5fe713cb5f397ff78778cc6646f632a3a`
- `packages/server/src/transaction-http.ts`: `81c130dd18f604ab1d13c03a2f52a98e11a6988f45d62af98b4b8c7499d14529`
- `packages/server/src/transaction-runtime-configuration.ts`: `57cd5d38c7199011d05589bc2b8ac1a0e2a4a9dc7585d019f9de5801a0a4d808`
- `packages/server/src/resource-evaluator.ts`: `6784629841a92baaeed648862cfa84d071356ef82e6d29250b254a9eddf793a2`
- `packages/server/src/transaction-batch.ts`: `830ad57a680db7148fc75cd9cc10cadfc6a812448fd1704c1b0f47b13611808d`
- `packages/server/src/transaction-store.ts`: `f7fe71da918c135eb43cb4659f186df6cd442aed9c42c2835e95cde05c782a9e`
- `packages/server/src/transaction-events.test.ts`: `0152561cdf096a48b0b01ec0e0d8bc81208a905bb5bf71dad411694387b58b27`
- `packages/server/src/resource-evaluator.test.ts`: `2a35e3add8eabe18d27a69c173b709e1d958c16191c049a4877d92a7b27fa7f1`
- `packages/server/src/transaction-store.test.ts`: `7cd155ddd959ec31a207adbe6f8eb0a26f0af968a144abb58915720436bf0774`
- `schemas/bdp-v0.schema.json`: `0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227`
- `docs/specs/bdp.md`: `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8`
- `package.json`: `f081baa3202b4815663d360d3ad1608206efe1b023086a4c8feff4d7fb97d8d0`
- `vitest.config.ts`: `a84575ea9ef0b6e75b6484024a627fa89422aea2fdbeb26127e09714dba0641f`
- `.github/workflows/ci.yml`: `268c9563e8d2f4fa4b7314492cb6c4994f7c76da9b26cda7a74c7e5b3c6f664a`
