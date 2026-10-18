# Native council review: durable receipt pages

## Verdict and basis

**Bounded pass; no blocking findings.** Reviewed frozen `0ff180a531b3c9ae594e0a104321e80744de38ed` against its direct parent `d3252087a31529cf97b5bfcfe2c402f02c694b91`. Author checkout `vickie-transaction-receipt-pages` remained untouched and clean. Independent build/tests used `bdp-transaction-receipt-pages-native-review`, at the same commit. No AGENTS.md exists in this checkout or the searched ancestor chain; repository CONTEXT.md was read.

Reviewed the complete source diff, author report, paging plan and native/external plan reviews, plus the d325 external minimum-read findings. This verdict applies to explicit-mutation paging in the existing fixed all-visible private receiver, not complete Transactional qualification.

## Independent evidence

Pinned Node 24.16.0 / pnpm 11.20.0; frozen install and build passed before tests. **81 tests across five files passed, zero skips, 1.83 seconds**, maxWorkers 2: receipt-pages, transaction-http, transaction-runtime, transaction-service, transaction-consistency. This includes actual HTTP schema/conditional/HEAD checks, changed-listener process restart, injected commit failure and corruption controls. Logs: `transaction-receipt-pages-native-{install,build,tests}.{log,exit}`; structured results `transaction-receipt-pages-native-tests.json`.

Two independent emitted-code counterexamples passed. Reproduction source: `transaction-receipt-pages-native-probes.mjs`; output/exit: matching `.log` and `.exit`.

1. **Actual prior-format refusal:** the d325 emitted owner created a real prior-schema file with one completed receipt and one pending admission. Opening it with the candidate and an expiry-advanced clock refused installation mismatch. Complete file bytes were identical; the pending admission and available receipt detail remained; no page table was added. This supplements the author's format-marker fixture with an actual predecessor-produced file.
2. **Mid-chain failure:** injected a throw on the second `receipt_page_ranges` insert, after the first range was inserted. All resources, groups, receipts, parent page plans and ranges rolled back. Reopen plus the original key admitted and completed once, yielding 2/2/1. Prototype restoration and disposable-directory cleanup ran in finally blocks.

No broad workspace suite, additional smoke, or evidence regeneration was run for this independent bounded review. Author full-matrix/smoke/evidence results are reported by the author, not claimed as reproduced here. Injected SQL/process failures are not physical disk-fault qualification.

## Paging plan F1–F7 dispositions

| Finding | Disposition and actual implementation |
| --- | --- |
| F1: bound ambiguity | **Closed.** `transaction-runtime.ts:81–84` captures one validated maximum before opening the owner; that value configures persisted partitioning, advertised maximum and Read pagination maximum. `transaction-service.ts:57–64` defensively requires equality. Default 32; HTTP cases 1/2/5 prove five results inline at 5 and paged at 2. The external suggestion to derive the partition from defaultItems is not adopted; the approved maximum-bound reading matches the explicit inline rule. |
| F2: unsupported minimum versus page expiry | **Closed.** `transaction-store.ts:1429–1458` establishes principal/epoch and issued-page identity, then handles unsupported or supported minimum controls before `expireRow`; HTTP passes the unsupported sentinel. The actual expired-page tests prove foreign 404, authorized private gap, timeout 503, foreign-view 409, then satisfied-minimum expiry 410. |
| F3: old next:null service invariant | **Closed.** The root check permits a bounded nonempty prefix with same-parent next URL and a full prefix when next exists. The page counterpart checks size/parent/next; the owner additionally validates the complete persisted chain. POST/replay/root/page HTTP paths pass. |
| F4: page key and redundant successor | **Closed.** Token primary key, unique `(receipt,start)`, parent plan carrying principal/key, contiguous recorded ranges, no stored next pointer. Next URLs derive from saved tokens/boundaries. The bounded reader validates the full chain rather than claiming large-set lookup performance. |
| F5: default and raw/wire separation | **Closed.** Default remains 32; raw full receipts and singleton sequence/creator/expiry paths retain complete facts. Only observed delivery projects a prefix. Full allocation extraction includes tail-page creations; bound-one alias/singleton/no-op/failed/sequence controls pass. |
| F6: old-format and bound refusal | **Closed with justified disposition.** Separate fixtures exercise format and bound mismatch, both byte-identically, despite sharing the same installation-mismatch diagnostic. Distinct error strings are unnecessary to distinguish controlled test inputs. Read-only preflight precedes writable open; the owner repeats the check under its transaction before recovery. The independent actual-old-file probe passed. No migration/reseed. |
| F7: lower cap versus Read behavior/evidence | **Closed as scoped.** Lower maxima affect only explicitly configured private fixtures and their corresponding Read pagination. The default remains 32; no shared Read router, sealed evidence or public profile changes. Removing the old inline-only guard is justified by implemented pages, not new qualification. |

The prior native plan clarifications (exact maximum equality, pre-maintenance admission, fail-closed chain integrity) are all implemented. Missing/overlapping/truncated/wrong-parent chains reject rather than regenerate or truncate. Available full facts and persistent boundaries commit together at `transaction-store.ts:1035`; raw expiry still requires full internal `results` and `next:null`.

## d325 review dispositions

- **Medium 1, expiry before minimum: closed.** `readReceipt` supplies the minimum check as the disclosed-row callback before retention (`transaction-store.ts:1492–1501`). Service tests separately prove timeout when only read-triggered disposal would occur, and 404 when independent maintenance actually forgets the failed receipt. HTTP proves the timeout case. This introduces no retention pin.
- **Medium 2, valid controls swallow decided syntax 400: closed.** `transaction-http.ts:351` preserves parsed status-400 errors before unsupported route gating. Three actual query-syntax cases pass with a supported minimum.
- **Low–Medium 3, point disconnect claim: closed by accurate documentation, not new plumbing.** Progress text now states that HTTP point and receipt disconnects drain the finite remaining wait; source facet AbortSignal support is distinct. The same limit applies to new page waits, which share tracked service lifecycle. No immediate socket cancellation is claimed.
- **Low 4, reserved genesis: closed.** Position normalization maps `genesis` to zero; real HTTP succeeds. Unknown/partial forms remain explicit private gaps.
- **Low 5, polling cost: explicitly deferred with honest limits.** The 5 ms finite polling mechanism remains; documentation disclaims production concurrency/load qualification. This is not repaired or newly qualified by pagination.

## Remaining limits

No additional product ruling is needed for this slice. Full-set pagination, variable authorization/erasure, installed contracts, replication, complete minimum-position semantics and truthful public Transactional discovery remain open. The chain-corruption checks defend the recorded representation; they are not a general hostile-database security claim. No page data or current context is synthesized from later graph state, no schema/catalog/public export changed, and no historical Read evidence was relabeled.
