# Native review — callback-only evaluator timeout proposal

No blocking issue found in the proposed explicit10,000ms timeout for this one existing callback. The bounded change is supported by the callback's unchanged workload and the observed prior near-threshold CI duration. It is not a diagnosis of the original delay, proof of a flaky test, or authorization to merge/retry without the remaining gates.

Reviewed proposal SHA-256 `305dab085dea4fd6a9dea92c210c95cb16a10d86233cdf3fbb37e6aee24be271` against clean candidate3df82ac480f50c8f486b4e456336366e7acbdf87. Independently checked all11 listed test/helper/evaluator/runtime/store source hashes and equality to parent88c. The callback file also equals the prior PR126 successor01a file byte-for-byte (SHA722728cac9fb3d85282c0c03aebdd81afc8a7ef0dae7f414ebc67635c7d90992). The existing six Event files remain outside this proposed edit.

## Source-grounded assessment

At development-contracts.test.ts628–666 the4000-item uniqueItems fixture first requires a direct instance limit-work refusal, then submits identical properties over real HTTP and checks bodyless500, Resource absence, correction under the same previously unclaimed key,200 and exact final Resource equality. installed-schema-evaluator.ts597–605 performs pairwise uniqueness evaluation; installed-schema-value.ts4–18/60–75 retains the16,777,216 logical-work ceiling and throws on exhaustion. development-contracts.ts97–112 compiles the actual HTTP validator with the same ceilings. The test thus exercises the full-cap refusal directly and through the serving path; lowering the budget/input or removing either phase would alter the proof.

Adding the third Vitest argument10000 changes only that callback's wall-clock failure threshold. It does not add requests/evaluations, weaken assertions, retry a command, raise production evaluator/transport ceilings, or change hook/global/job timeouts. It can delay detection of a stuck callback by up to an additional5s; a larger test timeout is not an abort or cleanup mechanism. The proposed normal formatting around the call should be verified to preserve every callback-body token.

Optional wording precision: change the suggested comment's “allow CI contention headroom” to “allow CI wall-clock headroom.” Contention is plausible, but the retained records do not isolate its cause. The proposal prose already makes that uncertainty explicit; this is not a blocker.

## Workload and lifecycle

Independently counted seven successful-path HTTP exchanges: fixture149–155 performs Scope probe, descriptor and operations-directory GETs; callback655–665 performs refused POST, absent GET, corrected POST and final GET. This is one in-process listener, one recovery-store authority and one bdp-contracts- directory, zero product subprocesses/restarts. Direct compile/evaluation is CPU work, not hidden HTTP. These counts are source-derived; targeted output does not contain a seven-response transport trace.

Fixture110–122 registers the directory and successfully started server. afterEach29–32 closes registered servers before removing directories. Runtime482–560 starts one listener, opens the authority and handles startup cleanup; close565–592 closes connections, drains admitted tasks and closes the authority. Its existing10s request/header/socket settings remain unchanged. The existing sequential afterEach can skip directory removal if close fails; the timeout patch neither fixes nor worsens that limitation. Do not claim stronger cleanup solely from exit0.

The isolated record reports only Node compile cache remaining. A fresh read-only check found PID97783 absent and targeted.tmp containing only node-compile-cache; no owned bdp-contracts- directory remains there. There is no retained descendant process inventory from that run, so this review does not manufacture complete process attestation.

## Actual evidence, kept distinct

- Original exact-head PR127 CI: callback5029ms, declared5000ms timeout, sole failure; suite755.84s; whole job960s. Its failure and skipped downstream gates remain unchanged.
- Prior PR126 successor raw log5508: the identical callback passed4430ms, leaving570ms under the default. This is a real observation, not a worst-case bound.
- Root's isolated unchanged-source run: exactly1passed/14skipped; callback813.2792089999999ms; supervisor2.2838637500535697s, exit0/no timeout. It does not reproduce or explain the full-suite delay and does not turn the original CI green.

Read and hash-checked targeted-start.json7265702858907cc20e4da065cd120bde1edaf9db3e0552ed9dfd10d7d311b3a3; targeted-exit.json8488976803ddee6a529fa2d8a629f33c5f87659ffc83110bc65e0b67f676fcfc; targeted.log9116dd97984a60286950dec9b6754942ba006fa7690092e06796d3175560e458; targeted.json4b850cb45597e9e2ce32c33f8b178a5b7df35fa6d2fa9ea8e0c51615975bca18.

## Narrow next gate

The proposed one-file timeout/comment successor is reasonable with exact diff/body preservation review, one selected callback validation plus formatting as root chooses, actual Claude supplemental review, and fresh exact-head full CI. Retain the local Event installed witness under its original3df attribution; a test-only timeout does not manufacture new installed observations. A full-suite result is still needed to establish the successor's gates, including the currently skipped preflight/smoke/Read seal. No reviewer tests, imports, builds, source edits, CI polls or reruns were performed.
