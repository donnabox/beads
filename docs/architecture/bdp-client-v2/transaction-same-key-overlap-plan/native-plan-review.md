# Independent native same-key overlap plan review

**No actionable plan blocker found. Accept the bounded proposal with its stated partial-evidence limits.** Authoring and execution remain root-owned gates. This review does not qualify concurrent-join as a whole or release the two-path implementation.

## Exact basis

- Plan SHA256 `84bc96cb8d87692e61e51f5b1ba3c854dbea40906bc268f3f58c925591b14996`.
- Ledger `2cb75de21509029ac9dd06006461fdfc524a04b4356dc9e1ffb78269c1633451`.
- Source binding `f4be4a20691f9a23e0ba61f0ddc77369cde86cc7e412a5b2ec2f6388ec68caeb`.
- Committed source `f5511718371a6935614eae6756e796cef71570c2`, tree `3cd1da66ba3a6cbd0bbdaa2442d10002939d826b`, read with `git show commit:path` in the principal checkout. All 31 listed committed file hashes and byte lengths independently matched. Dirty principal files were not treated as committed inputs.

I authored the earlier six-request recommendation, but not this prepared plan/ledger. This review independently checked its concrete mechanics and source dependencies; it is not presented as an independent invention of the slice. No product imports, tests, builds, installed execution, external review or product changes were performed.

## Correctness and actual overlap claims

The singleton and one-operation batch differ only in delivery envelope, with identical operation kind/record/defaults and no supplied id. Spec 3886–3940 and 3973–3976 support shared semantic identity and one authoritative namespace. HTTP `transaction-http.ts:329–347` routes both to the common submit service; `transaction-submission.ts:121–134,196–213` distinguishes existing receipt from owned execution, and `transaction-store.ts:1057–1094,1112ff` compares prior identity before allocating a new receipt/owner. The existing overlapping source callback at `transaction-submission.test.ts:299–315` remains supporting component evidence, not an installed wire observation.

Requiring both client end calls before first settlement and distinct actual sockets is feasible with Node's asynchronous HTTP event delivery and `agent:false`. It proves two outstanding clients, not overlapping server ownership/admission. The plan explicitly records stronger connect/write milestones without requiring them or inferring server scheduling. No sleeps, pending seam, repeated attempts or hidden retry loop are introduced. This distinction must survive implementation and progress prose: two terminal receipts do not earn pending-window or original-submission wait-bound credit.

Allowing 200/completed or 202/pending while comparing common identity is correct. HTTP 176–187 supplies 202 Location/Retry-After; pending receipts legitimately lack terminal results/positions. The planned at-least-one-terminal prerequisite and both-pending inconclusive stop prevent equality between differently shaped legitimate receipts from becoming a false failure or an unbounded poll. A both-pending result must remain an unsuccessful/inconclusive witness outcome in its output and external adjudication, never a green six-request pass. That is already the plan's stated behavior, not an added protocol restriction.

The issued root validation, schema-before-baseline assertions, full terminal replay comparison, complete one-item current collection and complete one-Event page discriminate shared outcome/allocation from two effects. Actual Event construction at `transaction-batch.ts:692–707,790–798` supports the exact created data; projection at `transaction-events.ts:75–100` supports the explicitly private hashed id. Store 1277–1306 records the committed group, pos-1 and terminal receipt together. This observes one effect and its receipt/Event correlation; it does not inspect execution count, prove crash durability, or exercise a multi-operation atomic rollback.

## Control sensitivity and retained observability

The second-key control is sound for the stated stimulus-sensitivity purpose. Omitted Resource id avoids a same-id creation conflict that would otherwise mask the intended comparison. Distinct valid keys can create two successful Resources; independently checking each actual sent key, result/schema and pos-1/pos-2 before asserting shared receipt id keeps the intended failure reachable. Positive-only shared-key/pos-1 assertions must stay after that boundary as planned. The report must retain both response facts before the named throw; ordinary generic exit 1 is insufficient.

Exact control acceptance requires two successful completed bodies, different receipt/transaction/allocation ids, the observed client overlap, exactly two completed requests and the named `same-key-overlap:shared-receipt-id` assertion. Cleanup/report errors cannot substitute. The fallback six-request maximum on an unexpectedly equal pair is bounded and is never accepted as a successful control. This does not demonstrate that removing a particular internal join check would fail; the plan expressly avoids that claim.

The specified report allowlist contains the facts external adjudication needs: sent-key/route/body fingerprints, status and response fingerprints, safe receipt identities/positions, actual milestone/socket ids, named oracle, performed equality flags, attempt/completion counts and lifecycle outcomes. Full JSON equality is performed in memory and is not reconstructible from report fingerprints; that attribution is explicit. No credential-bearing config, raw request map or arbitrary Error serialization is permitted. Source and archive/consumer provenance remain supervisor-owned, distinct from caller-supplied source labels.

## Bounds and lifecycle

Independently recomputed all three literal POST body hashes/lengths (108,150,150) and both artifact hashes/lengths (170,166). Positive POST bytes 408; expected control 258; combined expected 666; maximum fixed schedule 816. Six positive plus two expected/six maximum control gives 8 expected/12 maximum HTTP, two children/owners/listeners/stores, and at most two concurrent requests. Twelve retained 32768-byte bodies give 393216 bytes; this is honestly a retained-body allowance, not a heap cap. Header/body/target/report bounds and 32 deduplicated pair milestones are finite; the implementation should retain milestone identity per request so A and B are not accidentally collapsed.

The 45 s work, 5 s request, 10 s cleanup, 60 s test and 120 s supervisor bounds are joint deadlines, not additive guaranteed capacity. The launcher uses its existing 10 s bounded ready/close operations; an outer cleanup expiry remains a failure with retained directory/process evidence. The plan correctly distinguishes launch attempted but child unavailable from proven child exit, and does not remove such a directory optimistically. Fresh external-consumer payload checks are mandatory; ordinary workspace-emitted CI mode remains separately attributed. No runtime timeout or public admission change is authorized.

## Scope and residual gates

Two authored paths and one callback, executed positive/control externally and positive-only in ordinary CI, are coherent. No need to add a matrix row, schema/harness API, server scheduling hook, CPU mirror test or repeated full cohort. Principal integration/reanchor, concrete source/selector bindings, draft measurements, actual code council, exact execution results and final CI remain required. The unobserved server-pending branch, broader failure/atomicity/restart/expiry variants, pending product rulings and whole-profile qualification remain explicit gaps. This report requests no semantics decision and does not broaden the work.

Only this report was written under the required unchanged-session guard.
