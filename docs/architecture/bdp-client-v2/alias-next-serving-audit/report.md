# Next serving recommendation: fresh finite changefeed opaque minima

Recommend the already-deferred **finite JSON `changes/?start=now` matching-context opaque-minimum repair**. This is one consistency/replication serving slice, independent of retained Snapshot S1 and incident-Link G2. It is a read-only source diagnosis on clean alias commit `5ceb213b632e17f28714684efc5e35a62b282bf6`; no reproduction or new test was run. All679 bytes match the reviewed/tested alias successor. Supporting source/context identities are in binding.json.

The current remote launch has the same authorized session, with PR160 published and CI pending at the read, alias source work complete, and the independent local audit passed. The sole design plan still prioritizes actual usable Transactional serving and preserves public activation/S1/G2 gates. This report continues its T3 changefeed/position work; it creates no replacement tracker. The historical required-contract matrix is useful for contract families, not current implementation claims or coverage counts.

## Concrete trigger and current behavior

Seed one real Bead through the actual Transactional owner. Issue finite JSON GET or HEAD to `changes/?start=now` with the returned current scope epoch and authorization view, plus lexically valid `BDP-Minimum-Scope-Position: unknown` (or near-miss `pos-01`). Current source returns private500 with `x-bdp-development-limit: minimum-position`, rather than waiting and returning contextual503 catch-up-timeout when eligibility remains unproved.

The concrete source chain is:

- `transaction-http.ts:468–483`: the changefeed route always uses strict `transactionMinimumHeaders`; opaque tokens become the `unsupported` sentinel. Lexical malformed controls have already been rejected separately at407. The route selects JSON/SSE and reconnect precedence, then calls `service.changes`; `unsupported` becomes the private500.
- `transaction-service.ts:466–490`: `changes` copies the intent and strictly snapshots supplied tuples. A direct well-shaped opaque tuple therefore throws synchronously; the HTTP sentinel instead survives. The registered operation captures `start=now` through `owner.captureChangesStart()` **once**, before invoking the existing waiter. It retries `owner.readChanges` against that same anchor.
- `transaction-store.ts:1825–1888`: `captureChangesStart` captures the actual admission head checkpoint; `readChanges` accepts only an `after` checkpoint and strict minimum, so it has lost the original start-now versus supplied-after provenance. The unsupported sentinel returns before context capture. Otherwise it captures current context, validates checkpoint/view/epoch/future/missing retained row, checks minimum, then reads/projects a complete group and builds the finite page inside one owner entry.
- `transaction-consistency.ts:57–80` distinguishes lexical capture from strict known-position capture; `:136–155` already supports actual matching opaque Unproved through `checkCurrentReadMinimum(..., true)`. Its existing bounded waiter retries Unproved and returns503 with the final actual context; no new wait implementation is needed.

These are source-derived outcomes, not an executed new failing test. Existing changefeed tests demonstrate the relevant implemented behavior, but do not currently test this opaque start-now trigger: `transaction-changefeed.test.ts:543` proves admission-anchor preservation across a known-position wait and drain; `:689` proves known HTTP minimum failure precedes conditional shortcuts; `:750` preserves owner-observed parameter-error context.

## Normative basis and distinct boundary

Pinned monolithic `docs/specs/bdp.md` is unchanged SHA256 `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8`.

-5820–5832 separates lexical receiving from supported position identity and forbids malformed controls becoming conditional/negotiation/SSE success.
-5841–5851 gives the remaining-read eligibility/wait/route/problem law. Using the existing finite wait for this matching opaque changefeed request is a reference implementation choice under that general law; the narrower Resource/properties MUST is not claimed as a changefeed-specific mandate. Retained continuation/Snapshot precedence remains explicitly excluded.
-6036–6053 requires exactly one starting intent and defines `start=now` as future-only;6149–6153 fixes its returned `after` to the checkpoint observed **at admission**. Moving capture after waiting or recapturing on retries would skip intervening commits.
-6185–6189 preserves SSE Last-Event-ID precedence; normative contract identifiers at6994–6995 require changefeed consistency fields and bounded catch-up failure. No new cursor syntax, retention policy or publication claim is proposed.

This is not another finite Event/Snapshot/receipt/alias test. Those completed repairs do not touch the strict changefeed admission/capture/read path. A fresh start-now request has no client-supplied retained anchor, but its one-time captured checkpoint must remain fixed across retries while current observation context advances. That is the distinct semantic risk requiring tests.

## Minimal proposed implementation boundary

Five existing files should suffice: `transaction-http.ts`, `transaction-service.ts`, `transaction-store.ts`, `transaction-changefeed.test.ts`, and `apps/bdptest/src/transactional-changefeed-refusal-context-installed.test.ts`. This is a proposed maximum, not authoring authority. The parser, consistency helper, runtime pump and shared launcher should remain unchanged.

Keep selected finite start-now intent explicit through service→owner, so only that path uses lexical capture and the existing matching-opaque comparison. Do not globally loosen `readChanges(after, minimum)` or infer fresh intent from a checkpoint string. A small private default-strict intent discriminator/entry seam is appropriate for the later plan; retain existing direct-after semantics by default. Capture the caller tuple once and keep the original admission anchor once. Each retry checks the actual current context inside the existing owner entry before reading/projecting groups. Preserve checkpoint validation and corruption order, lifecycle registration, signal propagation, and existing pre-wait anchor-capture order; do not quietly move abort/policy/clock preflight ahead of it.

At HTTP select the extension only for finite JSON start-now. **Finite JSON already ignores Last-Event-ID completely**: do not copy the finite Event path's header-absent eligibility fence. A saved/malformed/repeated reconnect field must remain ignored on JSON under the existing rules. SSE retains strict minimum parsing and its real Last-Event-ID-over-start-now behavior. Explicit after/genesis, retained cursors, foreign opaque epoch/view, parser/lexical receiving, successful incident-Link minima and S1/G2 stay unchanged. Known-position paths retain their behavior. Accept/conditional failures must not become an owner-free shortcut past a selected minimum.

## Focused prospective evidence

Reuse the existing known anchor-through-wait/drain, checkpoint/future/missing-row, actual finite GET/HEAD/restart, known minimum ordering and parameter-context callbacks. Do not replace or re-instrument those historical callbacks merely to count them. Detailed rendered selection and aggregate workload belong to a later bounded plan, with no blanket suite.

Add only the discriminating new source scenarios: unknown/pos-01 direct-owner Unproved versus service contextual timeout; immutable captured tuple and exactly one start capture while a real intervening commit advances context; no stale success or skipped admission anchor; abort/release/drain and closed-owner behavior; excluded after/SSE paths. Reuse the current known-position positive to prove the same captured anchor includes a later committed group. Count actual delegated owner calls/context/anchors, not HTTP-inferred SQL or wait attempts.

A proposed new emitted HTTP cohort can be bounded at10 exchanges/one runtime: seedPOST200; known eligible finite GET/HEAD200; matching unknown GET/HEAD503; matching pos-01 GET/HEAD503; unchanged explicit-after opaque500; unchanged SSE opaque500 (no stream admitted); malformed tuple400. Use the existing deterministic short clock/wait seam rather than widening callback limits. Include conditional/media headers on opaque cases, actual context and Retry-After/noETag, positive GET body length plus empty HEAD matching metadata. JSON Last-Event-ID ignore behavior can be exercised by headers on existing selected requests, without extra dispatch. Exact source counters/deadlines need final plan review, not invented observations here.

Append a separate installed **four-exchange/one-child/one-store** witness to the existing refusal-context file: real seedPOST200, known eligible finite start-now GET200, matching opaque GET503, paired HEAD503. Derive epoch/view from actual receipt and verify the eligible page's captured checkpoint; no fabricated discovery or replica data. Preserve its historical10-request callback byte-for-byte and select it separately as relevant regression. New callback should use the proven request5s/work45s/one-total-cleanup10s/callback60s pattern, actual raw response fingerprints, child exit+close/drain and owned-directory evidence. Actual default wait1000ms is configuration, not a latency assertion. Honest fresh package/consumer/ready bindings remain required; counts above are prospective only.

Fresh emitted build must precede module collection; this test suite imports emitted runtime. Native+actual Claude plan/code review, exact bounded execution, process/store cleanup and exact-head CI remain root gates. This recommendation launches none of them.

## Why this next, and what remains held

The preceding alias recommendation already explicitly deferred this start-now gap until its capture/checkpoint seam could be reviewed; reuse that diagnosis and the original `transaction-changefeed-plan.md` admission-anchor law rather than start a competing plan. The later parameter-context repair already settled contextual parser refusal and finite-header/SSE separation; preserve it.

Cold public discovery is still a real milestone blocker: HTTP intentionally refuses Scope/bdp.json/operations at `transaction-http.ts:610–618`. Enabling it would advertise an incomplete profile; no partial descriptor or admission bypass is a substitute for finishing its required serving predicates. This changefeed repair advances one of those predicates without claiming cold navigation is now complete. Other retained minima, S1, G2, foreign opaque policy, Types, optional History, completed-key-conflict diagnostics and YAGNI remain separate. No new human product ruling is identified for this fresh matching-context finite path; the private capture seam needs normal bounded review, not policy invention.

Recommendation complete. No source, tests, services, dependencies, branches, PRs, main, shared database or coordination bus were changed. Only this report and minimal binding.json were written.
