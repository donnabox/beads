# Snapshot creation minimum — review dispositions

Plan revision only; not implementation approval. Exact revised plan: [v2](transaction-snapshot-creation-minimum-plan-v2.md), SHA256 `2fb0a12ea6061f40033f51a5f7c52a104248e3b65eef296d7bfbd090b42fa9d0`. Original reviewed-v1 remains byte-identical to original `transaction-snapshot-creation-minimum-plan.md`, SHA256 `1f24e59f8fd1e9df2d2ec275353f1e730db27e852199a886d168c244da30ec7b`. Both review outputs remain unchanged.

## External findings

| Finding | Disposition and precise correction |
|---|---|
| 1 HIGH — negotiation versus minimum | Accepted need for explicit choice/discriminator; rejected suggested406-first bypass under root direction. Existing create preflight already observes owner/capacity/integrity before406. v2 keeps that path: compatible minimum→authoritative preflight→406; behind+unacceptable Accept→finite wait→503 catch-up-timeout. Cited norm orders minimum before conditionals, not universally before negotiation. This is bounded implementation ordering, not new protocol law. Contrary normative evidence must be raised before authoring. Corrected review's “full real capture pass”: preflight does not capture. |
| 2 HIGH — behind not converted by HTTP | Accepted. Both service create/preflight minimum paths explicitly run inside existing finite wait. Raw behind cannot reach HTTP. Reject optional suggestion to map unwaited behind immediately to timeout; that would misrepresent waiting and evade the intended repair. Focused HTTP/direct checks cover both paths and exact timeout code/type/context. |
| 3 HIGH — possible second wait | Accepted future-proofing requirement; current second-call premise is unreachable by the inspected candidate/eligible decision inputs. v2 documents those inputs and requires terminal preflight return plus focused one-call/no-capture tests. No shared-deadline redesign or second authorization pass required. |
| 4 MEDIUM-HIGH —503 collision/capacity | Accepted. Exact catch-up-timeout code, unavailable family type, status/retry, actual fixed-head position below minimum, bounded table row-count no-allocation checks, and later successful new handle. Retry count remains scheduler-dependent; no approximate200-attempt pin. |
| 5 MEDIUM-HIGH — deadlines | Accepted. Actual45s aggregate/60s outer/5s wire/10s lifecycle retained. Exact31 positive+2 derivative exchanges,3 new children; separately counted existing larger89-request/2-child and capacity39-request/3-child cohorts. Planned total161/8; new40/3 cap does not misleadingly claim to cover the entire old suite. Implementation base must re-anchor counts. |
| 6 MEDIUM — wait floor | Accepted ambiguity correction; do not adopt requested new900ms bound. v2 says approximately1s per fixed-head request, not hard SLA; unchanged finite harness ceilings and actual duration reporting. Existing deterministic service evidence proves deadline mechanics. |
| 7 MEDIUM — mutable request/control | Accepted. Capture/freeze create/preflight kind and tuple synchronously before run/await, owner resnapshots/rejects retained misuse per attempt; mutate-after-call and pre-abort discriminators added. No change to no-minimum retained paths. |
| 8 MEDIUM — mutant provenance | Accepted. Distinct derivative checkout/payload; preferred actual throwaway commit or explicit parent+patch+tree/build derivative binding. Never report a modified payload as clean parent source; assert differing installed payload digests. Include mutant requests/physical child and expected failing oracle in budget/report. |
| 9 MEDIUM — bodyless context oracle | Accepted. GET/HEAD/304/412 and compatible406 must assert exact independent fixed-head context and minimum eligibility. Foreign failures assert actual context rather than input echoes; HEAD remains bodyless. Private genesis/pos-N interpretation is fixture-specific, not a new public token grammar. |
| 10 LOW-MEDIUM — retained exclusion rationale | Accepted qualification. Only retained position comparison is the unresolved interpretation cited here. Epoch/view refusals may be separate safe work but are excluded from this slice. Original audit is preserved historical, explicitly qualified by v2 rather than silently rewritten. |
| 11 LOW-MEDIUM — nonexistent old controls | Accepted. Existing installed partial create control stays unchanged. Complete-tuple retained-manifest and actual-page gap controls are explicitly NEW. Actual page comes from existing34/34 fixture, not invented URL or small cohort. |
| 12 LOW — source citation drift | Accepted. v2 cites symbols and corrected inspected ranges (observeSnapshot1426–1466; HTTP creation412–430; helper symbols) and makes re-anchoring on root's final reviewed integration base a prerequisite. No false claim current line numbers are future-base anchors. |

## Native findings

| Finding | Disposition |
|---|---|
| P2 budgets/mutant count | Accepted; corrected limits and enumerated every new exchange/session. Larger regression and unchanged capacity runs explicitly outside new small-cohort cap, with separate totals. |
| P2 authorization-view transport | Accepted; C1 safe credential-resolver adapter is expressly owned and reused, including anonymous handle, exact transmitted-value assertion and raw-header omission. No executor firewall weakening. |
| P2 real continuation | Accepted; small2-Bead/1-Link cohort has no continuation claim. Existing34-Bead/34-Link cohort gets1 new actual beads.next minimum-refusal request,88→89 requests, unchanged page size and existing no-minimum/reopen/convergence. |
| P3 progress document | Accepted; only present Snapshot support/cancellation/gap text is added to owned files. No normative/schema/progress sweep. |

## Independent source checks and bounds

Read-only source basis `20d61119b5faf7251a5d285c5c34a1049866ffe5`, tree `3de0c4b5589423d01eb0dcb7df7ad150d184273d`; checkout remained clean. Read actual owner/service/HTTP/minimum helper, Read HTTP semantics, normative consistency/conditional/Snapshot clauses, installed Snapshot lifecycle/transport/tests and C1 view credential adapter. Confirmed readContext position encodes pos-0 for empty head while Snapshot checkpoint is genesis. Confirmed code/type families from protocol readProblem definitions, and the catch-up-timeout override retains unavailable type. No source/build/test/server/network execution or source changes; only external plan copies/revision/dispositions written.

The larger installed ledger was counted from actual source:2 descriptor reads,3 auth/missing/empty observations,4 seed writes,1 capture,2 real continuation reads,2 manifest GET/HEAD,1 wrong-stream request,1 update,3 retained reads after reopen,1 feed and68 point reads =88; new actual-page control makes89. Existing capacity test10+29=39 requests across3 children. New small cohort31+2=33 across3; all named installed cohorts sum161 requests/8 sequential launches. These are plan counts to assert against actual execution, not observations claimed already run. Root acceptance must include this explicitly separated regression accounting; no deadline/capacity/runtime ceiling is raised.

No unresolved fresh-creation product decision was found in the cited norm. Retained-position interpretation, malformed tuple protocol behavior and broader simultaneous-error precedence remain outside. v2 awaits native/root closure before code.
