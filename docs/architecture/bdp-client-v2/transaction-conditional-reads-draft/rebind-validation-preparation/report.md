# Corrected conditional rebind — CLOSED validation preparation

No tests, build, product imports, package creation, consumer install, child launch, or validation command ran during this preparation. Only source/report reads, JSON parsing, AST parsing, hashing, and guarded external packet writes were performed. Both execution gates remain absent; their example documents have `released: false`, empty approved prefixes, and closed phases. Root alone may create release documents and actual gates.

## Source and scope

The corrected reviewed draft is based on `1ba65f2f192cc84e44eb09ef89b9693ec6cdaf05`, tree `773614c5ebc6c97d90c4c629c7c7a02ed51bc2ef`. Its four exact source hashes are in `prepared-source.json`, sourced from `rebind-correction/source-binding.json` SHA256 `1095d0093a729c9138c22c6419da528fefbece42be5707dd16db1238e91d61ee`. The corrected full patch SHA256 is `bc7e620ea97a37933901ee4bf2e2d818685178bd974a262e88b025e513e2a60a`. Those source hashes were rechecked during preparation; this does not attest future integrated bytes.

The conditional row is `transactional.http.conditional-reads`, appended after the sixteen unchanged parent rows. Its fourteen actions contain exactly 194 portable assertions. The positive must actually complete all fourteen; the nonmatching If-None-Match control must stop after exactly four with precisely the two expected failures, genuine 200/full current body, and initial validator equality. The app retains implementation-specific 412 metadata checks; this packet does not restore those as portable spec requirements. Matching 304 validators must equal the earlier GET validator.

The new adjudicator joins the supplemental session to actual successful launch records by scenario/reopen/sessionOrdinal, checks every actual session ordinal, preserves all sixteen parent first presentations in full/cross cohorts, and requires exactly one final non-reopen conditional session. Selected conditional runs use ordinal zero. Existing principal/member/cross evidence continues through the original reviewer with the new configuration digest. Credential handles and redacted body fingerprints are corroboration, not independent wire credential attestation or reconstruction of full JSON bodies. Full JSON/schema/postimage equality still executes in the frozen app callback before redaction.

## Initial source-only gate

`source-preflight/` is an independent closed dirty-source packet. After root accepts the corrected code/council and explicitly releases it, its exact ordered commands are typecheck, format-check, lint, boundaries, cumulative-full, binding-source, receipt-first-page, app-static. The gate binds actual dirty status, HEAD/tree, exact four-file patch, every tracked input, source lock, tools, and the command digest before and after execution. It contains no build or installed command.

There are 56 selected source cases: whole cumulative file 42, physical-binding regressions 3, receipt first-page fixtures 8, and app static checks 3. App static callbacks do not launch the constructed dist path, so they can run before build. `read-seal` imports conformance/server dist and is reserved for the final packet after fresh-build. These source-only results cannot become final integrated or installed evidence.

The eight unchanged first-page fixtures were individually inventoried from `packages/conformance/src/runner.test.ts` (source hash in `source-workload.json`):

- id-first: 5 HTTP attempts, one listener, one logical reopen.
- page-first: 5 attempts, one listener, one logical reopen.
- wrong-parent: 4 attempts, one listener, rejects wrong parent.
- empty-page: 4 attempts, one listener, rejects empty page.
- changed-page: 5 attempts, one listener, one logical reopen, rejects changed page.
- cancel-page: 4 attempts, one listener, last request intentionally aborted and drained.
- null-next: 1 attempt, one listener, stops on null next link.
- absent-next: 1 attempt, one listener, stops on absent next link.

These are 29 expected attempts, with a diagnostic ceiling of 40; eight two-byte seed POSTs contain 16 total body bytes. They create no product child or owner store. Their finally blocks close all connections and await listener close. They exercise the restored principal index-zero fixture shape without adding new source tests.

The three binding cases make 1, 1, and 2 attempts respectively across 1, 1, and 2 listeners: close-abort and cohort-abort must drain their registered socket promises before mock child close; rebind must read bodies from the first and second physical listeners while base mode stays fixed. There are four total attempts and two deliberate aborts, no product children/stores. Cumulative source cases add fourteen synthetic attempts across twelve listeners. App static cases add none. Total source inventory is 47 expected attempts, 58 diagnostic maximum, 24 listeners, and three deliberate aborts. Attempt counts must not be represented as completed HTTP evidence.

## Final integrated packet, still closed

The outer packet requires ordinary integration of the reviewed PR147 successor, its exact-head green CI and council disposition, and a clean final candidate. Actual final base/head/tree/ordered parents, patch, tracked inputs, successor commit, and root acceptance remain explicitly unknown/null until root binds the real result. No future hashes have been invented. The final candidate must preserve these reviewed four-file bytes, have only those four changes relative to its accepted final base, and have both the reviewed base and serving successor as ancestors as required by the gate. Any rebase/conflict/content change requires new review and pins.

The full command order is in `commands.json`: source checks; fresh-build; read-seal and selected source cases; build inventory; Node/package metadata; protocol/server tar creation and source-to-tar verification; separate offline external consumer preparation/install; root acceptance of its actual generated lock; installed payload verification; root acceptance of installed provenance and full consumer inventory; then each of seven installed cohorts immediately followed by its data-only review. There is no workspace entry fallback.

Selected installed workload:

- Conditional positive: 14 completed requests, 2 children, 271 request-body bytes.
- Conditional nonmatching-header control: exactly 4 completed requests, diagnostic maximum 14, 2 children, 141 actual/271 maximum request-body bytes.
- Principal positive: 7 requests, 2 children, 750 request-body bytes.
- Principal collapsed namespace control: exactly 2 requests, diagnostic maximum 7, 2 children, 300 actual/750 maximum request-body bytes.
- Full cumulative: 234 requests, 25 children, 18,877 request-body bytes.
- Cross-batch semantic-conflict control: 225 requests, 25 children, 17,821 request-body bytes.
- Cross-sequence fresh-key control: 226 requests, 25 children, 18,025 request-body bytes.

Across the seven distinct installed cohorts: 712 expected requests, 727 diagnostic maximum, 83 children, seven temporary owner-store directories, and 56,185 expected/56,765 maximum request-body bytes. Source and installed combined prospective attempts are 759 expected/785 diagnostic maximum. These are finite prospective workload descriptions, not completion or qualification percentages. Old Read conditional catalog evidence remains unplanned; all reports must retain `claimEligible: false`.

## Supervision and provenance

Each command uses the inherited 120-second wrapper, TERM at 115 and KILL at 120, no automatic retry, exclusive start/result markers, exact selected-case verification, source before/after checks, process-group samples, PID/temporary-directory cleanup checks, and root review of results. Failed or timed-out stages halt the ordered prefix; a new attempt requires a distinct reviewed packet. Unexpected files are preserved instead of erased.

`supervise.py`, `inventory.py`, `packages.py`, and all three principal contract/ledger/control files are byte-identical to the previous reviewed preparation. Packaging compares source archive members and installed external payload, accounts only for intentional workspace dependency rewriting, pins eight production dependencies, checks symlink containment, and records the actual installed entry and consumer inventory. A fresh build and fresh external consumer are required; prior successful build/install evidence cannot be relabeled.

`adaptation.patch` shows changes to inherited gate and reviewers. `reference-pins.json` binds the original packet and reviewed plan/author/correction/measurement artifacts. `preparation-hashes.json` binds this complete closed packet including the source-preflight subpacket. Source hashes, derivative specimen/compact hashes, 194-assertion ledger, JSON syntax, and Python AST syntax were checked without loading product modules or executing validation.

Release readiness: ready for root inspection of this preparation only. Initial source preflight needs explicit root release. Final integrated phases additionally need actual ordinary successor integration, exact source/ancestry/CI/council bindings, fresh build, actual generated consumer lock, and actual installed provenance acceptance. This is not a claim that validation passed or that the write profile is complete.
