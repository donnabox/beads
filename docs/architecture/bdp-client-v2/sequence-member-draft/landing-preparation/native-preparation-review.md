# Native landing-preparation review

No actionable finding in the frozen preparation packet. This is a read-only review of scripts/examples, not execution, CI observation, landing acceptance or a qualification claim.

Reviewed README63bf17bd, capture-ci.py983e5db9, verify-landing.py dcf7dd15, pins.example.json f095146e and ci-target.example.json c2953802; all five full SHA256 values match preparation-hashes.json SHA540bc3f169395b1a7aa5fef915856de45eff6f7f630754dca263b67e48a3048e. Both live gate files remain absent.

## Verified scope and gates

The verifier's exact four-path set matches the accepted source scope. Example source hashes match final a6eb425d draft; future commit, PR, run/job, tree/synthetic and evidence placeholders remain null. Premerge requires a clean local candidate on the correct branch, direct ordinary ancestry18104a03→candidate, exact file bytes and unchanged workflow. Postmerge requires the tested tree and ordered actual parents[base,candidate], not merely a successful merge API response.

Evidence hashes include final source, native report, actual Claude exit/stdout, root council, source acceptance, installed acceptance, all four installed review-candidate records, CI outcome and root landing acceptance. Claude requires real exit0, no timeout/forced termination, unchanged bound inputs, nonempty stdout matching its retained digest. Root's landing acceptance explicitly binds head/tree/base and the complete exact evidence map; a still-running review cannot satisfy it.

The four installed record fields and command names match the actual review-installed.py producer: positive-fifteen213/23, selected-sequence-delete-control1/2, cross-batch-control204/23, cross-sequence-control205/23. They bind candidate head/tree, closed child totals and owned-root absence, preserve claimEligible:false, and sum623HTTP/71children. The landing verifier deliberately relies on root acceptance for complete underlying archive/payload/provenance, individual oracle and PID evidence; the count summaries alone are not claimed as independent installed proof.

The candidate workflow hash fd0a08da61f74abe19c8db16d879bfff85d8090643bf4da326d3e597ae129be3 matches committed18104a03:.github/workflows/ci.yml. Current pinned Node/pnpm/actions,20-minute job deadline, full suite,7preflight, smoke and Read gate remain intact. Exact27 job step names preserve the preceding PR131 scaffold and require every step successful, including downstream/post steps. No new timeout, selection or CI waiver appears.

capture-ci.py is a single read-only snapshot with at least90seconds between calls and exclusive terminal artifacts. It stores the full terminal raw job log even on failed CI and never retries/cancels. verify-landing.py separately binds successful run/job/checks to candidate, compares raw log hash, extracts checkout's tested SHA, and verifies its tree and ordered parents. Ready/open/clean mergeability and unchanged main are required before merge; no script performs readiness or merge writes.

## Practical boundaries

Root must supply the real CI outcome and inspect actual full log totals, named new/cross callbacks, timing and cleanup before signing rootLandingAcceptance. The script checks the pinned outcome structure and exact successful job, but does not itself parse every test assertion or independently re-adjudicate installed raw reports. README states this boundary; it is consistent with the prior proven scaffold.

Exclusive writes intentionally stop after a partial capture failure rather than overwriting evidence. A source repair, advanced main or new CI attempt needs an explicit successor binding/directory. The strict direct-parent requirement must be revised explicitly if a future ordinary repair commit is introduced; it does not silently permit a changed chain.

No scripts were executed or imported, no outward queries were made for this review, and no product source was mutated. Only local file/Git reads and hash arithmetic were used; the required session guard is used solely for this external report write. Root remains the sole execution, acceptance and publication owner.
