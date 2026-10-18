# Next bounded Transactional slice after principal isolation

**Recommend one installed same-key overlap witness across singleton and one-operation batch routes.** This is a missing installed evidence slice for the settled shared-key/exactly-once law, not a newly discovered runtime defect. Use one authenticated principal, one fresh key and an authority-allocated creation so accidental double execution produces two identities/effects. Do not duplicate the current principal-isolation witness, reopen cold activation, or expand the cumulative matrix merely to add a row.

## Basis and gap

Read the sole fork design's active objective and live ledger, plus `transaction-depth-key-order-draft/next-serving-audit.md`. This proposal does not repeat the broad serving audit. Inspected committed actual main `f5511718371a6935614eae6756e796cef71570c2`, tree `3cd1da66ba3a6cbd0bbdaa2442d10002939d826b`, through `/Users/dbox/repos/wt/vickie/vickie-principal-key-isolation`; the in-progress principal test is not counted as merged evidence.

`packages/conformance/catalog/transactional-v1.json:685` requires `transactional.idempotency.concurrent-join`. It is absent from both the current 15-scenario executable manifest and 15-row bounded coverage artifact. This is an executable-evidence gap, not proof that no component test exists. `packages/server/src/transaction-submission.test.ts:299–315` already exercises two overlapping in-process `submitTransaction` calls, equal receipt identity and one inspected group. The installed cross-carrier row exercises sequential presentation and replay; the new sequence member row exercises separate members and retained replay. Neither records two simultaneously outstanding installed HTTP clients sharing one key across routes. The ordinary Update overlap witness uses different keys and expected-revision arbitration, with a different profile/outcome contract; it remains unchanged.

Settled law: `docs/specs/bdp.md:3886–3902` scopes normalized identity to Scope/epoch/principal and requires concurrent equal requests to join one execution. Lines 3910–3940 exclude carrier metadata, equate a one-operation batch and singleton, and permit the duplicate's pending or terminal receipt. Lines 3973–3976 require all mutation routes to consult one authoritative namespace. No Snapshot or incident-Link ruling is needed; no Type, History, retention-expiry or provenance claim is added.

## Existing implementation supports the experiment

- `transaction-http.ts:329–347` maps singleton and batch requests into the same `service.submit`; `176–187` emits a pending submission as 202 with its actual receipt Location, otherwise the terminal representation.
- `transaction-submission.ts:121–134,196–213` uses the same comparison/admission owner for both carriers, returns an existing receipt, and synchronously executes an owned claim. No new hook or scheduler seam is needed.
- `transaction-store.ts:1057–1094` looks up principal/key, canonicalizes against prior identity, and returns existing versus conflict before fresh admission; `1112ff` allocates one pending receipt/transaction and owner. The proposed test should exercise these through installed HTTP, not invoke them or inspect a live SQLite database.
- Reuse the bounded client-side pair machinery as a pattern from `apps/bdptest/src/read-update-guarded-overlap-installed.test.ts:320–370,459–463`: two distinct sockets, attach rejection handling immediately, issue both ends before awaiting either response, retain monotonic lifecycle observations and close all sockets. Copy only the small necessary transport discipline into the TX witness; do not generalize/rewrite the working Update helper.

## Smallest useful installed experiment: six HTTP requests, one child

One fresh external-consumer Transactional service, stable fixture Scope and one operator credential, ordinary empty initial state. No bootstrap HTTP, sleep, server-side pause, fault hook or live database inspection.

1–2. Dispatch concurrently a `create-bead` singleton and the equivalent one-operation `batch`, same literal `Idempotency-Key`, same type and properties, **no supplied Resource id**. The batch adds only its required operation kind/envelope; compare the canonical semantic request, not identical HTTP bytes. A concrete record is the existing fixture's task Type with `{title:"same-key-overlap"}`. Validate the exact admitted carrier shapes during the bounded plan before authoring.

Both responses must be schema-valid receipts for the same key, receipt id, transaction identity, Scope epoch and Authorization View. Accept terminal 200/completed or pending 202 as permitted; do not demand two byte-identical bodies when one is pending. Require at least one actual terminal completed response for this synchronous installed implementation, then retain its entire validated body as the terminal baseline. A different future implementation returning two pending responses needs a separately bounded polling plan; this six-request experiment must not grow an automatic retry loop or call that legal alternative a protocol violation.

3. GET the actual receipt id, after both submissions settle. Require completed available receipt and full equality with the terminal baseline, including allocated Resource, positions and expiry. Pending bodies from either submission must share identity/transaction with this terminal outcome; 202 Location must equal that observed receipt id.

4. GET the ordinary Bead collection with an explicit bounded limit using the existing route/query shape. Require exactly one full Resource, exactly equal to the receipt result, and no continuation. This discriminates accidental double allocation even if a broken implementation later aliases the receipt response.

5. GET the Scope Event page. Require one complete creation Event for that exact Resource/revision and receipt transaction, correct source and ordinal, no continuation, and unchanged serving position. No incident-Link/metadata operation is involved. Actual page schema and full Event facts are asserted in memory; output retains only safe facts/fingerprints.

6. Retry the equal one-operation batch under the same key. Require the entire terminal receipt baseline unchanged and the same serving position. The pair, final state and Event assertions establish one-effect behavior for the observed execution; they do not prove a particular internal locking mechanism.

The two outstanding client requests must be evidenced by actual dispatch/response ordering and distinct sockets. **Client overlap does not prove both requests occupied the server's pending-admission window.** Record whether a 202 actually occurred; if neither did, report terminal replay under overlapping clients, not observed pending-join or wait-bound coverage. That partial boundary is essential and is consistent with the existing source-level pending test. Do not add nondeterministic delays just to force 202.

## Scope and sensitivity

Likely two authored paths: a new `apps/bdptest/src/transactional-same-key-overlap-installed.test.ts` and the existing `docs/design/transaction-implementation-progress.md`. Runtime, catalog, manifest, coverage schema, launcher API, public admission and existing profile evidence remain unchanged. Use current fresh-package/eight-package consumer provenance and exact installed entry binding; no workspace fallback or source attestation invented from parent HEAD. Final source must ordinary-reanchor after principal integration.

Plan the sole counterfactual narrowly before execution: a test-only wire derivative changes just the second request's key to a different valid key, leaving the authority-allocated creation unchanged. Correct service behavior then admits two independent creations; the shared receipt/one-Resource oracle must reject it. This is a different-key stimulus control of witness sensitivity, not a production mutation or proof that every faulty join algorithm is detected. Exact failure IDs, whether it stops at the pair or retains later state, and total control request count must be fixed in the authoring plan rather than guessed here. No additional control is executed or authorized by this report.

Use the existing witness discipline (finite request/body/count/work/cleanup limits, complete response body collection, redirects disabled, finally cleanup and retained failures); six positive HTTP/one child is a proposed workload, not a run result. The transport ledger must distinguish initiated/completed responses, and stop on an unexpected outcome without retries. No new harness framework or full-matrix amplification is justified for this first small installed observation.

No unavoidable product decision blocks this bounded partial slice. If root instead requires proof that the duplicate was pending *inside the server*, this proposal deliberately does not manufacture it: the current supported installed configuration offers no deterministic suspension seam, and an internal pause would change the experiment. Keep that stronger portion of concurrent-join visible as a gap rather than overstating overlap.

## Immutable source pins and inspection limits

SHA256 at f551 main:
- spec: `f036103ffdf5f89fa3a2033570e1e987c4cd203cd80b1c633cb44615123f892e`
- TX catalog: `c9125f914129c71f01cefc5c54b32db80637d37b57e321571c96e81936c31eb4`
- TX manifest: `a11a79f6d0f9ffa5933a680a3e4fed87987789c3deca50a38cda8563132d242e`
- TX coverage: `346e1edaa22af789f89017fdc2fef6bb01b7564ed72a65344e9fe5635887c74f`
- submission: `4336e45d7b4f90eec685b70c7ba059219954bb20271aa2c942ee9e140f0743af`
- store: `fcd6882b29e538db45867c894c06dfbaa6ca588db0c2e403ef1b7abb14f6c706`
- HTTP: `b51d02801e5924190aba77c377e538db0b8619149d6184e19cb56d440f40c2cf`
- submission test: `86342bc18126965f380d0e09218481eceac4a8396afb4543dd7e11d464fd1976`

Read-only source/catalog inspection and stdlib hashes only. No new runtime observation, product edits/imports, test/build/service, council launch, public advertisement or network write. Only this report is written after the remote session guard. Existing principal work and pending rulings remain separate.
