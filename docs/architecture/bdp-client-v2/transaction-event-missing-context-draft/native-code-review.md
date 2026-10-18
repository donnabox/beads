# Native code review — missing Resource Event context

**One Medium validation-bound finding was identified and is now closed by the reviewed one-line successor. No other actionable defect was found in this bounded six-path review.** This conclusion is source review only: compilation, tests, emitted-control behavior, fresh installed observations and cleanup outcomes have not been executed or demonstrated by this review.

## Finding N1 — automatic redirects bypassed the new source dispatch counter — CLOSED

Original location: packages/server/src/transaction-events.test.ts1125, new HTTP callback's fetch call.

Fetch's default redirect following can send further HTTP requests while the test's attempts counter advances only once. That contradicted the new callback's no-redirect/exact11-dispatch contract and could obscure route evidence under a redirect regression.

Root authorized only `redirect: "error"`. I independently verified the final source equals the original snapshot with exactly that one line inserted, and all other five owned files are byte-identical. A redirect now fails the exchange instead of causing a hidden follow-up request. The installed socket executor is unchanged and does not follow redirects. No selection, request count, timeout, owner or child allowance changed.

## Exact source and review binding

Workspace: /Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence; branch codex/vickie-event-missing-context. Parent remains88c9408e908590ab1a4c64c9b9dee1a8b705999c/treeb2d9f2c18502e0778c51123be8b4564af052d803. This is an uncommitted six-path draft, not a clean implementation commit.

Original reviewed packet, preserved:
- report.md: ed3b827e1f6da2e5295e7ec040eebc1116c841a3f75a610fc4f140efd6ad5b8c.
- source-binding.json:83e5b7b63785abbb7986a14447a3ee21cc29d301e26ee85bb6fae9ff595b0f3a.
- complete draft.patch:822603cd33acb3549ad018a2d440ad4e19d5b9e38216cd879a1dc3f38ae90426.
- Original source Event test:27255ef2ee85e20135a1b2078cd1a699bb050ea717d6c532ec068a9f1aecb696.

Final reviewed successor under redirect-repair/:
- source-binding.json:**acf190e742dfb669fe5607085a3ac60c5ecb66f4f86b1f05025424a7873615b0**.
- complete draft.patch:**e0c25b193ef1191d10916728c54eb6a894c5f29992e0da768bd585ef089e98e3**.
- correction.patch:1b82124fd9005dd864a45f4720936c9aa7b8fcb6ef06235f86adcaccefe50927.
- report.md:4bec01b934f0abbbb3adffdb4a9d38a81ee83acc02d2003b79b8125336e99ada.

All final file bytes matched their snapshots and SHA256 values:
| Path | SHA256 |
| --- | --- |
| packages/server/src/transaction-store.ts | b2a444f7997ba4e2b34bd57ed2a8119f0999e6146305b5aae4fc8735ff78bc7c |
| packages/server/src/transaction-http.ts | 9db33008a4c3a66214e2bf523b0b90f79c216c1f3f304c6df0491a0d1904439a |
| packages/server/src/transaction-events.test.ts | 0152561cdf096a48b0b01ec0e0d8bc81208a905bb5bf71dad411694387b58b27 |
| packages/server/src/transaction-event-sse.test.ts | 69f58153b93947f028c37aebce2f57b0a25f725aacb9a1891c7b276b1c9e44c0 |
| apps/bdptest/src/transactional-event-missing-context-installed.test.ts | 828c05f358ae5863c992f93194da15a31df9a200e073e5e3525531ba39fe7d0d |
| docs/design/transaction-implementation-progress.md | c31d464b9d7928639c54acf827844d3ae62691d026ee9557ed6f7b921643b278 |

Read the accepted original plan, six-path addendum684d3c05b331417a71cbd14044e734e878a6ddffabfaeccd97e4a3ad10be1c39, root dispositions, full actual Claude findings/native source dispositions, complete source changes, requests/selection/commands/control packet and referenced source APIs. No applicable AGENTS.md/CLAUDE.md was found within the checkout or its ancestor chain. Unrelated sibling instructions were not applied.

Only read-only source/data inspection and hashes were performed. No tests, typecheck, build, import, runtime, install, packaging, control application, source edit or external review launch. The report write is session-guarded.

## Production behavior and regression boundary

The production diff is confined to the two explicit missing-result unions/branches and two HTTP projector calls. readEvents and beginEventScan both call existing readContext inside the same synchronous enter callback after eventSourceKnown establishes absence. Both result and context are frozen. Successful-path observations and all minimum/cursor/reconnect order remain unchanged.

Service return inference carries the added context without a service-source change. HTTP passes the supplied tuple to transactionHttpProblem; it does not reconstruct context in a second owner entry. Authentication, malformed route/input handling, unsupported minimum and existing fault mapping remain in their previous positions. New metadata-integrity failures are deliberately propagated through the existing internal-fault path, as accepted by the addendum; no synthetic tuple or contextless fallback was added.

The SSE regression changes only the canonical unknown404 row's observed flag to true. Its malformed query400 row remains false; all other table rows, anonymous context exclusion, hook lifetime and timeout are unchanged. The three exact owner missing shapes now include independently expected configured epoch/view/current position. No other formerly reviewed exact missing shape is left stale by the diff.

The progress document accurately labels the implementation as an unexecuted draft, preserves historical attribution and excludes profile admission, retained Snapshot minima and metadata-only incident-Link membership.

## Test oracles and body bounds

The four new owner cases independently exercise finite/SSE × Bead/Link atpos-0 and after a genuine unrelated commit atpos-1. They assert exact missing shape, frozen result/context and unchanged old context, then preserve unknown-before-unsupported/future-minimum/invalid-cursor behavior. No new scan capability or cached live object escapes. These checks support the concrete owner change; they are not overstated as general concurrency verification.

The source HTTP case uses the explicit emitted-runtime import. All status/Problem/media/framing prerequisites precede the tuple oracle. GET validates the full independently authored Problem; HEAD validates empty body and corresponding representation length. The eight tuple checks are soft solely so the one control reaches both branches. Its report explicitly defers verdict to final Vitest output and does not turn a zero caught-exception count into success.

The source response ceiling is enforced during reading: each chunk is checked before retention, excess aborts the request, read failure awaits cancellation, and finally releases the reader lock. Bounded concat/fatal UTF-8 decoding occurs afterward. The report correctly distinguishes retained32768-byte chunks/concat/body strings from native transport allocations and parsed-object overhead. It makes no whole-process memory guarantee.

The installed witness uses the existing strict socket executor and real serializer observations. Authored canonical URLs, methods and Accept order are explicitly asserted; batch target is checked against the actual ready binding. Its credential handles contain only Authorization, with an empty anonymous handle. Key/type/Accept/body metadata checks use actual available ExactRequestObservation fields, not an invented method field. Safe reports retain whitelisted response headers, fingerprints and binding metadata, never the secret-bearing full request header array.

Receipt and Event schema checks, exact known source/one created Event/next:null, independentpos-1, complete404 Problem, all eight context tuples, HEAD framing and anonymous401/challenge/no-context are concrete oracles. No receipt/Type/discovery fetch or successful live SSE GET is added.

## Measured selection and lifecycle

The static title expansion is consistent with the source: eight callbacks in transaction-events.test.ts and the single existing SSE regression callback. The four owner rows use named object fields, so the declared titles correspond to the table expansion. Control selects only the new source HTTP callback; installed selects one exact title.

The source/ledger model is nine positive callbacks39HTTP, one emitted-control callback11HTTP and one installed callback11HTTP/one child: **11 case executions,61 expected/max authored HTTP,11 owners/directories,4 listener lifetimes,1 installed child**. Existing SSE callback decomposition remains28, with its source untouched except the expectation flag. These are source-derived counts, not observed execution totals. Six full-CI callbacks are added; the existing SSE callback is not a new registration.

Independently recomputed source seed131 bytes, installed seed152 bytes, both digests and descriptor170/schema166 bytes with336 aggregate. The raw executor's32768 body/65536 response-header/163840 wire bounds are actual API settings/default arithmetic. Request header/body bounds remain16384/65536.

Both new callbacks retain primary failures, abort work, attempt runtime/child close, independently attempt directory removal, preserve deadline expiry through cleanup, and emit safe observations before rethrowing distinct failures. The installed launcher already bounds pre-ready failure reaping; an unset child variable is correctly reported as startup-unobserved rather than proof no child existed. External PID/temp verification remains required. Existing owner/SSE cleanup mechanisms are reused within their accepted scope.

## Remaining execution gates

The emitted control is specified, not yet prepared or executed. It must remove exactly two context arguments from a copied positive emitted graph and bind the copied source test's relative import to that graph. Bare dependency and module resolution must also be verified. Expected outcome is one failed case with exactly eight named missing tuples after all11 completed prerequisite responses, no unexpected error/timeout and actual cleanup. No generic exit1 credit or extra installed mutant.

The existing commands/selection packet still carries the original source-binding hash. Reanchor the execution preparation to the reviewed successor before release, as the successor report explicitly requires; do not silently relabel the original packet. This is pending preparation, not a source defect or execution authorization.

Fresh packaging/installed attribution must follow a clean ordinary final commit, source/build/archive/member/dependency verification and separately released installed11/1. Actual results, final reporters, process absence, exact-head CI and root integration remain outstanding. No runtime success or public qualification is inferred from this code review.
