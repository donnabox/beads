# Independent audit of actual Claude Event-plan findings

Verdict: accept the bounded installed Event restart witness, subject to the concrete wording below and root disposition. No additional production defect or need to edit shared helpers was established. Findings 1, 7, 8 and 9 offer useful explicit constraints; 3 and 6 need accurate observation limits; 2 and 4 contain material recommendations that must not be adopted; 5 mixes a real attribution concern with stale workspace assumptions. No authoring or execution is released by this report.

I read actual plan-v2, actual Claude stdout, the launcher, SSE recorder, installed bootstrap, installed Event/alias callbacks, Event cursor/store code, installed binding and served-configuration factory, and the existing outer validation supervisor. The thirteen Event review source pins still match. Current fresh-depth checkout is clean at ddfb5f43a058fe0f16ff9e04e528b4e7bfc03d87; b0213805 was chronology, not a current execution-source pin. I have not edited either checkout or any plan/review packet and have not run tests, imports, builds or servers. Only this guarded report is written.

## 1. Sensitivity controls — accept as bounded strengthening, not a newly diagnosed comparator defect

The existing alias callback uses the same comparator for an actual positive and a deliberately corrupted clone (lines 339–342 and 357–360). Add the two request-free controls after collecting real evidence: mutate a clone of the observed/restored E1 record and require the exact positive full-Event comparator to reject it; give the same positive SSE comparator the saved E0 as the expected Event while retaining the observed E1 frame and require rejection. Avoid an independent trivial comparison that does not exercise the positive oracle. Retain immutable originals. Record both under localSensitivityControls, separate from wire operations and serving evidence. Counts remain five HTTP requests and two child lifetimes. No claim that these local controls test the server's rejection behavior.

## 2. Cleanup envelope — accept clarification; reject late-exit success

Launcher close is memoized and can await exit for 10 seconds, then kill/drain for another 10 seconds (transaction-installed-observation.mjs:22–60). Plan-v2 already explicitly distinguishes its absolute 10-second callback cleanup/reporting bound from launcher termination, tracks unsettled promises, fails unknown-stop/drain timeout, preserves the store, and says supervisor intervention is never success. Thus Claude's proposed tolerance of late child exit on a reported pass directly contradicts the existing plan and must be rejected. Raising the callback timeout is not required merely to make an unsuccessful child teardown wait complete inside the callback; a bounded failure can leave cleanup to the owner supervisor. Do not expand timeouts as a substitute for proper gates.

Corrected language: “The 45-second work budget and shared 10-second final-cleanup budget bound callback work and reporting; they do not assert that the launcher's failure path terminates within 10 seconds. A callback pass requires every launched child to have a positively observed normal exit (code 0, no signal), every stream drained, no launch/stop uncertainty, and no cleanup timeout. The outer accepted result additionally requires no remaining owned processes and no forced cleanup. Any drain timeout, unknown startup stop, late exit after the callback has failed, or supervisor intervention remains failed evidence even if the process eventually exits. Preserve the directory and aggregate errors.” Keep the 60-second callback/120-second supervisor envelope only with a bounded failure/report path and root-reviewed supervisor. Underlying close promises must have rejection observers before races; no bare 20-second await after budget expiry.

## 3. Diagnostic trace — accept limitation; prefer truthful preserved evidence within current scope

The launcher counts/discards stdout/stderr and the installed bootstrap deliberately suppresses exception detail behind a fixed safe message (launcher:13–18; bootstrap:71–77). A callback cannot attach to the child's stderr before a failed launch returns no handle. Attaching after a successful launch cannot recover those lost startup bytes. Merely collecting the fixed stderr message does not recover the cause and must not be advertised as doing so. Do not edit the shared launcher/bootstrap to solve this in the one-file witness.

Corrected language: “Preserve the actual available observation trace: request ordinal, allowlisted response facts, safe Event data, caught launcher/cleanup errors and ownership state. On any second-launch failure, unconditionally preserve the store and immutable source/artifact/config identities. Startup failures may expose only a generic launcher error; no private in-child cause/stack trace is available through this helper. A later source diagnosis may inspect preserved evidence under a separately scoped action.” Do not log credentials/configuration bodies or exact request headerLines. This is diagnostic limitation, not permission to suppress a failure or rewrite expected behavior.

## 4. Pre-handle ownership and process groups — accept missing handle; reject blanket group ban

Startup rejection never returns {child,ready,close}; the callback cannot truthfully report a PID it did not observe. Plan-v2's conservative launch-in-flight/unknown-stop flag and mandatory store preservation are correct. Change its evidence wording to distinguish callback-known PIDs from supervisor-observed group members; use stop state “unknown-no-handle” where appropriate. The helper does not remove the directory; remove the ambiguous suggestion that it might do so.

Claude's group conclusion is wrong for the intended owner architecture. The inspected existing supervisor (pr147-ci-failure/validation-final/run.py:57) deliberately starts the runner with start_new_session=True. Launcher children inherit that separately owned group. Killing the whole owned group on failure may intentionally kill Vitest too; it must never target the invoking agent's group or an unproven unrelated group. A pre-handle child may be missed by callback PID capture and even by sampling if short-lived, so restricting forced cleanup to callback-known PIDs weakens cleanup. Do not invent unseen PID observations or claim complete birth accounting from samples.

Corrected language: “The outer supervisor creates and owns an isolated session/process group for this one Vitest invocation and its inherited children, records actual group samples, and inspects the owned group after completion. On failure/deadline/survivors it may terminate that owned group, including Vitest. Any such action fails the observation. Startup rejection before handle return is unknown-no-handle in callback evidence; preserve the store regardless of whether the helper attempted child cleanup. Report only actually captured PIDs. No successful report tolerates a survivor.” The future Event supervisor must itself be inspected against this contract; the existing supervisor is concrete precedent, not evidence that an Event supervisor already exists.

## 5. Installed attribution — accept fresh binding requirement; reject stale literal/dirty allowances

The module's fallback source is live HEAD; source-only env override is allowed by existing code; bindServedProfileConfiguration validates 40-hex syntax and digests, not that sourceCommit corresponds to the package bytes. This genuine attribution obligation already appears in plan-v2's explicit external installed entry/source requirement. Neither a literal b0213805 pin nor a permissive two-CI-file dirty whitelist closes it.

Current checkout is clean at ddfb5f43, with the CI repair committed. Future Event authoring belongs on an isolated branch, followed by a fresh exact source binding, reviewed one-file diff, fresh protocol/server package hashes and a verified external consumer. Require both installed entry and source override for the installed-evidence run; bind the source commit to the actual package provenance and record the witness source/diff separately. Reject source-only override/local dist being labeled installed evidence. Verify entry/payload identities and environment against the new execution binding; never copy a stale baseline literal or assume dirty paths. Ordinary local fallback remains development behavior. Preserve the historical callback unchanged; enforce new witness/run admission without widening its scope. Root will choose the exact branch and source/package binding before execution.

## 6. Bounded SSE window — accept wording and live state checks, avoid invented timing proof

The recorder exposes only complete frames that reached its data handler. Intentional socket.destroy can discard unobserved bytes; frames drops incomplete trailing segments. Therefore “all complete id-bearing frames observed in this bounded window are [E1]” is accurate; global/exhaustive absence is not. E1 as the first complete id-bearing frame excludes observed replay from E0 and proves precedence over query E1, without a sleep/absence claim.

Keep the adjacent synchronous pre-close checks signal.aborted===false and recorder.destroyed===false with no intervening await before close(). These are the decisive available health checks; until and close do not export failure. Re-read status200 and frame/content after drain. Capture elapsedMs immediately before intentional close and require it below the pinned 5000ms deadline if adding a literal check. Do not invent an arbitrary “well under” timing margin or use elapsedMs after later work as transport proof: it is a live wall-time getter, not recorded close time. Timing alone cannot replace destroyed/signal gates. A transport timeout or early socket death fails even when a frame was already buffered.

## 7. Restart identity source basis — accept explicit references with policy qualification

Add the concrete bases: transaction-store.ts583–630 provisions/persists epoch and loads policy on reopen; 592–611 verifies installation/capture bytes; readContext at 663–668 reads persisted position and current authorization.authorizationView; transaction-events.ts95–102 hashes epoch/view/source/checkpoint/ordinal. Initial all-visible authorization and no policy transitions in this witness keep the view stable. Avoid generalizing “authorizationView is constant” to all supported policy transitions: actual readContext uses current authorization state. These references scope future failures to persistence/configuration/identity changes without inventing a new protocol rule.

## 8. Literal position and required receipt fields — accept within the fresh-store witness

The completed receipt schema permits effectPosition without requiring it. Because this witness creates a fresh store and performs exactly one effective two-create batch, explicitly require receipt.effectPosition === 'pos-1', nonempty/valid transaction identity, both independent Resource postimages, and current position 'pos-1' on every post-commit context-bearing response (including SSE and contextual400). Keep independent expected ordinals0/1 and transaction links; reuse observed opaque IDs without synthesizing cursor text. These are concrete fixture invariants, not a claim that opaque position syntax is universally portable. No second mutation is allowed to silently invalidate the literal.

## 9. Foreign-source rejection preconditions — accept without another HTTP request

transaction-store.ts691–701 and1659–1670 check source identity/all-visible policy before cursor source validation. Pin A's exact identity as the successfully created beads/a postimage and configure/verify the existing all-visible policy identity, with no policy transition or deletion. Then E0 is a real Scope cursor applied to a distinct known Bead source and the expected contextual400 is discriminating. State those fixture preconditions; do not add another lookup request or claim a 404 under a different policy is a protocol defect.

## Remaining author/code-review gates

Freeze exact body/key/selector bytes and request/session counts before execution. Inspect the local finite helper for canonical Host, explicit expected status, complete response schema/context and no credential leakage. Verify installed entry/source/payload binding against fresh external artifacts. Review all close races, startup ownership flags, directory retention/removal and actual supervisor isolation/cleanup. A timing-dependent stream-disconnect failure must remain a preserved failed observation. These are future implementation/validation obligations, not verified passing behavior.

## Input hashes

- plan-v2.md: `0f16ffaaf73e108e122fecf5a24b534811446a51333139a730f78329ae278367`
- actual Claude stdout: `df5839ec4eca7bc32f359698023c645bc103de06c7c4fb49ada40c854485e9d5`
- actual Claude binding: `0627669baab40642979aa6294e3261d62d82faa253e90beb19c6b1b74b1502c1`
- existing supervisor (precedent only): `c7a1b63e16484d1cf484148d1ab4355c25a376daa94ab5d5eeff1867af42feb6`
- scripts/transaction-installed-observation.mjs: `6f772f3776d3c3effa788676d853cc2932a6ee59ad63ac3338ad1b487408914b`
- packages/server/test-support/raw-sse-recorder.ts: `09ff8343ac1a5dbad77eaf24fc2745cc28ea870ada6b05dc4ea1644c0ad5ef81`
- packages/server/src/transaction-installed-process.ts: `6bd16f940f1f751189fc76419e3a18ca26cd5f54da3cfa3946b30c3956a2f8b7`
- packages/conformance/test-support/profile-contract-plan.ts: `18c953443e6695cfa48f89677231fecffc349df02bd080f107a5c739db10b728`
- packages/server/src/transaction-store.ts: `f7fe71da918c135eb43cb4659f186df6cd442aed9c42c2835e95cde05c782a9e`
- apps/bdptest/test-support/transactional-execution-binding.ts: `cde506c72a8710eafe58522900f6db70ed3701cf333560252dc8e7c5e706cf89`

All thirteen Event-related source pins in the Claude binding match current files. Conditional-review pinned artifacts were not changed.
