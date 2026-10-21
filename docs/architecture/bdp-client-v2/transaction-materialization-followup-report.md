# Materialization/owned review followup

Base: `e2f1385aa89bc6ae27e18a802a2a651a8f02e87d`. Branch: `codex/vickie-materialization-followup`. Checkout: `/Users/dbox/repos/wt/vickie/vickie-materialization-followup`. Frozen clean head: `2fc4bb8c87b1651bee81eb7df9c1a57423e81452`.

## Result and scope

The stronger reachable owned workload **exceeded the 1 GiB supervisor gate at M64** during post-reopen verification. That run is inconclusive for its final verification and rejects the earlier claim of comfortable measured margin for M64. Root authorized M32 after notification; the actual emitted M32/E256 tuple completed the stronger workload, replay, receipt pages, finite Link Event pages and durable reopen. No general worst-case heap guarantee is claimed.

Current private counts are M32/E256/O32; byte/diagnostic ceilings are unchanged. The change is installation-bound and refuses prior M64 private stores rather than reinterpreting their receipts. Group bound is 169,345,024 bytes and receipt bound 39,391,232 bytes with K64. The existing finite Event bound derives automatically; only its test's literal changed. No store, service, HTTP, runtime, protocol or public admission edit.

Changed files: transaction-batch.ts, transaction-materialization.test.ts, transaction-owned.test.ts, transaction-set.test.ts, transaction-events.test.ts and docs/design/transaction-implementation-progress.md. Other frozen checkouts untouched.

## Review dispositions

- **F1 accepted:** runtime guard validates all four finite positive integral evaluator fields against the tuple: propertiesBytes, representationBytes, diagnosticBytes and diagnosticCount. TX options require all four; generic RU remains unchanged. Negative matrix includes missing, zero, negative, fractional, infinite and field-specific raised bounds. A thrown integrity refusal releases the claim; retry executes normally rather than retaining an invented protocol failure.
- **F2 accepted with corrected feasibility:** owned source embeds the Link, so two independently exact-R records are impossible for this shape. The final workload reserves only the exact empty source metadata wrapper, grows Link into every remaining available byte, then fills the source residual. Sources are exactly262144 bytes; Links261849–261850. The full65536-byte carrier is envelope plus patch:13061 numeric values, with the remaining token bytes filled using equivalent exponent digits, not outside whitespace. Combined normalized patch287461 bytes. The transient numeric array is removed before final property validation. Actual durable owner execution and normal Read/receipt/Event APIs only; no inspect(), synthetic outcomes or direct store edits. Only a post-close readonly scalar length query measures the selected group body.
- **F3 accepted:** closed equality assertions for owned Resource create/update/delete results and each upsert wrapper reject unexpected siblings such as preimage bodies. Existing exact Event/delta and tombstone assertions remain.
- **F4 accepted within TX tests:** permanent admitted numeric witnesses for1e20, -9e20, 9e9 and -0.0000010000000000000002 check lexical→normalized expansion <=7 and actual stored Event patch values. Exact1e20 output is pinned independently. This is a regression witness for extremal inputs, not an exhaustive proof over all values; no shared protocol change.
- **F5 reconciled:** both historical sections remain. The4096 observation is explicitly dated history, the prior256/257 default Event test is historical, and current32-set-match behavior is separately stated. Current owned test observes128/129 actual Events under a deliberately lowered evaluator-only128Event cap, retaining real rollback/replay/conflict/reopen checks. E256 remains the owner bound; the test does not claim to reach that default boundary under M32.
- **F6 accepted:** self-reported child maxRSS and supervisor sampled gate metric are separate below. A killed child's last self-report is not its peak.
- **F7/F8 deferred:** no current key collision or literal-type consumer defect was shown. Do not broaden evaluator plumbing or exports to address a hypothetical extension.
- **F9 accepted:** receipt proof explicitly charges the first Scope URI through S and the second (`next`) through H, alongside fixed suffixes and bounded failure envelope. Diagnostic/name/Scope allowances must be considered together.
- **F10 disposition:** E>=4M is usability, not a required proof invariant; no module assertion added. The real finite Event reader already validates the derived group byte bound before selecting the body; this successor updates its derived-bound test only.
- **F11 accepted prospectively:** every new command records its actual exit. No historical exit files were fabricated. Prior root full/CI verification remains independent historical evidence, not this run.
- **F12 accepted:** remove the old one-iteration seed loop and index arithmetic; current default count fixture directly seeds its one Bead.

## Measured runs and metric separation

All supervised workloads used the same120s/1GiB limit, sampling child RSS every100ms. `maxRSS` is the child's self-reported process high-water in KiB; sampled peak is what the abort gate actually sees. Both are finite workload observations, not continuous monitoring or a heap theorem.

| Workload | Child self maxRSS KiB | Supervisor peak KiB | Outcome |
| --- | ---: | ---: | --- |
| Historical M256 no-op | 743840 last report before kill | 1154256 | aborted; replay inconclusive |
| Historical128 no-op under original tuple | 802800 | 685744 | completed |
| Historical128 owned/small Links | 975216 | 929808 | completed; insufficient margin |
| Historical M64 no-op | 480000 | 439776 | completed |
| Historical M64 owned/small Links/50KB patch | 557584 | 537328 | completed; not near achievable large-Link workload |
| Stronger M64, source residual1KiB | 958528 last report before kill | 1072128 | aborted8.349s during reopen verification |
| Final M32, tight owned Link/source closure | 562976 | 540304 | completed4.161s |
| Final M32,32 repeated no-op matches | 285472 | 249584 | completed0.435s |

The stronger M64 run produced group70350401 bytes, receipt16705001, sourceR and Links260834–260835. Its initial execution, retained replay and first finite Link Event page completed; final verification did not. Logs and the failed supervisor summary are preserved. It already fails with more source residual than the tighter final M32 fixture.

Final M32 dense run produced group35207784, receipt8385225,33 result entries and one observed Link Event page288170 bytes. The source point state, receipt digest, receipt page and finite Event page compare equal after reopening; the final exact-key retry returns the retained receipt. The no-op run produced receipt8392388 bytes,64 entries and no effect group, with exact replay/page/point checks after reopen. Only one source-filtered finite Event page is qualified by this stress; it is not all-page/global-history or HTTP load qualification.

Preliminary32-workload runs under the old M64 tuple and intermediate source-residual/whitespace variants are retained under explicitly named files. They are not substituted for the final emitted-tuple run above. No arithmetic table cell is described as measured allocation.

## Verification and red evidence

Pinned Node24.16.0/pnpm11.20.0. Build before emitted-process tests.

- Frozen install, build, final typecheck, dependency boundaries, format check, lint, development smoke, packed smoke and unchanged evidence verification: exit0.
- Final tuple regressions:323 passed/9 files, plus103 passed/7 additional files = **426 passed/16 Transactional files**. The command also named a nonexistent transaction-batch.test.ts; Vitest selected nine actual files, and no separate batch-file result is claimed.
- Final affected recheck:117 passed/4 files; after the direct one-Bead seed simplification, that specific default-boundary test passed again.
- Initial red:12 intended diagnostic-bound refusal failures, plus4 numeric-test authoring errors (lossless tokens were wrongly sent directly to stringify without numeric admission). Those four were corrected to use the existing admission step and are not product failures. Initial red18 other tests passed. Focused green104 then passed before the count adjustment.
- Actual prior emitted M64 owner creates a disposable store; current M32 owner refuses it, database bytes remain unchanged, and prior owner reopens/replays exact serialized receipt. PASS. Initial compatibility probe compared JS prototypes and failed on the legitimate lossless/null-prototype versus reopened ordinary-object distinction; corrected probe compares serialized receipt content. Both logs retained.

Evidence files live beside this report: `regressions.*`, `remaining.*`, `final.*`, `last-test.*`, `red.*`, explicit command logs/exits, stress scripts/logs/summaries/exits and compatibility probe. Stress `.exit` records the actual child return (-9 for kill); final separate supervisor exits record its success/failure where invoked. No retroactive status file was invented for older runs. Both smoke scripts and unchanged evidence verification completed successfully. Prior whole-repository CI is not counted as a run of this successor.

## Remaining limits and integration

The collection selector still materializes its candidate collection before count refusal. Large accumulated graphs/history, sparse versus dense objects, principal strings, memory copies and runtime allocation behavior remain outside a universal memory claim. Lowering M trades throughput per set transaction for measured headroom; it does not change normative limit names, cardinality-before-cap ordering, or replay law. Existing old stores require explicit reprovisioning; no migration or receipt reseal is supplied.

Root/G1 and Event repair worker were notified of M32 and the single Event-test derived literal before integration. G1 derives group bounds and Event repair owns its separate reader/HTTP changes. This branch does not edit those implementations. No pushes, PRs, bus/launch edits or public conformance/profile claims.
