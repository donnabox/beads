# Next Transactional serving audit after receipt context

**One further concrete serving defect is established by source: owner-observed changefeed parameter refusals lose their current Scope context.** Recommend a small same-entry context repair before adding another diagnostic evidence row. This is a proposed next slice, not authoring or execution authorization.

## Exact basis and inspection boundary

Read-only source basis is clean candidate `1a37f7209e3ff6f409f43491d670c881259406ba`, tree `a9cca0e1e5c52ac8bbd06ce989a94ca6e4f2b272`, in `/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence`. Receipt successor CI/merge is pending; this is not actual-main attribution. Consulted sole design `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md:7–31,195–197`, prior `transaction-after-event-context-next-audit.md`, historical consistency rulings, current progress, HTTP routing/error projection, service minimum lifecycle, changefeed parser/owner and existing tests. No imports, services, tests, build/install, source edits or review CLI were run. RU preparation remains closed.

## Trigger, actual result and required behavior

An authenticated request supplies a well-shaped private changefeed checkpoint with matching epoch/view tags but a position above the current owner head. URL parsing accepts its grammar. `transaction-store.ts:1713–1716` reads the actual immutable current tuple, checks the token against it, then throws a contextless `TransactionChangefeedParameterError`. `transaction-http.ts:608–615` returns the existing400 invalid-parameter Problem but forwards context only for the Event parameter-error class. Thus all three current Scope response fields are absent despite an actual owner observation.

A second instance of the same loss is malformed selected-SSE Last-Event-ID with valid URL intent: HTTP `:454–466` forwards the selected reconnect value to service; `checkpointPosition` calls `parseCheckpoint` at `transaction-changefeed.ts:79`, after owner readContext. Its syntax error is likewise contextless. Repeated Last-Event-ID is different: HTTP rejects it before owner entry; preserve that contextless branch.

Law: `docs/specs/bdp.md:5798–5804` requires Transactional Scope response fields; `:5830–5835` requires every Scope-bounded read to report the projected-prefix context it observed. `:5845–5849` covers changefeed GET/HEAD and no live HEAD stream. The ordinary invalid-parameter status/retry is defined at `:1898` and already used by this implementation. This proposal changes no error code, unknown-checkpoint classification, authorization rule, minimum ordering or token format. Existing source tests explicitly accept future-token400 in `transaction-changefeed.test.ts:275–295`.

The fabricated future token is a negative-input specimen, not a valid durable client checkpoint or discovery authority. Spec `:6153–6155` requires conforming reconnect clients to use durable applied checkpoints. Do not present a private token edit as portable client navigation. This restriction does not justify dropping an actually observed response context on the server's existing refusal.

## Narrow implementation outline and fanout

1. `packages/server/src/transaction-changefeed.ts`: optional read context on its existing parameter error, following existing `TransactionEventParameterError` at transaction-events.ts:44–51. Thread optional context into private checkpoint syntax validation only when called from `checkpointPosition` with an actual owner snapshot; URL prevalidation still supplies none.
2. `packages/server/src/transaction-store.ts`: pass the already captured `context` on future-head rejection. No second observation, owner opening, fabricated tuple, parser reorder or maintenance behavior change. `readContext` at :590–598 already freezes the three fields.
3. `packages/server/src/transaction-http.ts`: forward either contextual parameter-error class through existing transactionHttpProblem. Keep pre-owner failures contextless. Service return shapes/wait lifecycle need no changes.
4. `packages/server/src/transaction-changefeed.test.ts`: use existing owner/http fixtures and their real current head; independently pin unchanged400 Problem and observed tuple for future checkpoint and malformed selected-SSE reconnect. Cover GET/HEAD and JSON/SSE future-token variants, no live stream on errors, normal metadata/HEAD empty body, malformed URL/repeated reconnect/anonymous401 context absence. Preserve foreign-view/epoch errors, minimum/conditional precedence and corrupt-group500 assertions. Use startup-safe local cleanup and bounded response consumption; do not inherit unbounded/default-redirect fetch behavior in a new witness.
5. One narrowly named new installed test, tentatively `apps/bdptest/src/transactional-changefeed-refusal-context-installed.test.ts`, using the existing genuine packed launcher and private configuration. No new runtime injection/export, discovery activation or public claim.
6. Existing `docs/design/transaction-implementation-progress.md` factual increment only, with durable evidence link after available.

All production callers of this error were searched: private changefeed parser, the single store future-head rejection and HTTP repeated-reconnect check/catch. Existing Event class is a model, not an owned edit. No shared schema/catalog/evaluator/config/runner paths are proposed. Root must arbitrate store/HTTP ownership after current receipt merge and reanchor the actual six paths.

## Bounded validation proposal, not measured execution allowance

Extend one existing CPU owner callback and add one bounded source HTTP callback; exact expanded titles and total existing fixture calls must be measured before release. A useful fresh installed witness is ten requests/one child: (1) real batch seed, (2) finite genesis page obtaining its real checkpoint/context, (3–6) same-tag future checkpoint GET/HEAD × JSON/SSE, (7–8) malformed selected-SSE Last-Event-ID GET/HEAD with valid URL intent, (9) anonymous401, (10) malformed URL400. The latter two remain contextless. Required actual responses 3–8 are400 with the same observed head tuple, not fake successes; the issued seed checkpoint remains unchanged. Hard-bound request target/body/header/response sizes, explicit no-redirect requests, finite work/exchange deadlines and aggregate cleanup before execution. No restart or second installed cohort is inherently needed.

One copied-emitted context-omission derivative can target the new source callback: require correct400/full GET Problem (HEAD metadata/empty body where selected) before a named missing-Scope-field assertion, independently bind copied imports, and retain original failure/process cleanup. Do not accept generic exit1 or package fallback. Fresh exact frozen archive/consumer provenance for the positive is separate from this derivative. Exact source/control request and owner/listener counts remain a measured-preflight gate, not numbers borrowed from prior receipt/Event slices.

## Exclusions and priority

No renewed receipt404/Event404, mutation POST minimum, Selector, alias/current minimum, schema, start=now or comparison-fault repair. Do not infer retained Snapshot/continuation anchor semantics (S1-Q1), metadata-only incident-Link Event membership (G2-Q1), partial/unknown minimum tuple policy or mutable-view behavior. Discovery remains intentionally unavailable until required cumulative serving/evidence gates pass; its refusal alone is not proposed for bypass. No optional History/Types/provenance redesign. Unobserved catalog counts were not used to infer missing functionality.

The narrow loss above is a current source-established response-context defect. It is suitable for independent native/actual Claude plan review; this audit supplies neither an executed reproduction nor a whole-profile qualification claim.

## Source fingerprints
- `packages/server/src/transaction-changefeed.ts`: `b731f660b3900be9914cc94bfe1d436514c9da323d825f9e377b33e3ca99ee19`
- `packages/server/src/transaction-store.ts`: `eec0f4723c8fcfb2245d2e80d77b2de014235cdb208101d523c0ab83fde78eed`
- `packages/server/src/transaction-http.ts`: `0e42f034c3aada6fd0a7696897200b336d2b0c547d69497e46500f6da4c2267a`
- `packages/server/src/transaction-service.ts`: `53f655faf494209996aac2e0f93a4c2c785e471f66aac6af9ea41f92c9cfbe00`
- `packages/server/src/transaction-changefeed.test.ts`: `3ef92bbedf31feef9bdf45a3fb2384e7487b91a5f688235da4abfeae4d94b7bb`
- `docs/specs/bdp.md`: `f036103ffdf5f89fa3a2033570e1e987c4cd203cd80b1c633cb44615123f892e`
