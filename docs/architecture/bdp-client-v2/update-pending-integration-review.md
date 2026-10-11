Root publication note: author e073a2b and ordinary integration e538e23 were pushed and remote-verified after this review. PR83 now contains the slice; CI38109989016 is running.

# Independent pending/interleaving qualification review

Frozen author: `e073a2b722d4a4165f6230ce1ffe8c766ee6bb9b` against `59915ea916a210f5ce36611eb6d799f7c3124056`.
Reviewed acceptance: `update-remaining-observations.md`, Recommended next slice / Acceptance.

No blocking findings. Reviewed every action/assertion in the two added plans and all helper/catalog/coverage/doc changes. Pending creator, original dependent and tail keys remain unavailable to overlapping requests; a statically changed pending tail request observes in-progress precedence and no effect. A fresh dependent key's transient failure does not persist: after release it executes, replays, then rejects a changed request without altering live state. Interleaving proves a separate keyed request commits before release, the original later member sees its state, revisions change, and retries return their retained postimages.

The helper awaits the actual original bounded Fetch response. Rejection is consumed immediately. Cleanup aborts/awaits that request, kills/awaits owned children, and then removes fixtures. Existing assertion-failure and abort-at-barrier negative controls passed with the changed helper. No arbitrary target or credential controls, scheduler/runtime/configuration edits, synthetic state, sleeps or expiration changes were introduced.

Coverage remains explicitly partial: exactly six former omissions gained bounded observations; the other 75 rows are unchanged. Counts independently verified: 54 bounded / 27 omitted, 25 plans / 253 actions / 797 assertions, all passing, claimEligible false. Claim atomicity, distributed writers, admission-midpoint faults, policy changes, expiry and non-pending transient variants remain unqualified. Original/replay equality is canonical parsed JSON equality. Creator/source/target GETs each verify the unique creator marker; they do not directly assert cross-binding URI equality. Documentation accurately describes marker observations, not stronger identity qualification.

## Local ordinary merge candidate

Frozen clean candidate: `e538e23b6a4a7371b14945f59009323c302e60a3`.
Parents: `21654be3cae51d50e75e817ef8ac57735b8ee4ea` and `e073a2b722d4a4165f6230ce1ffe8c766ee6bb9b`.
Checkout: `/Users/dbox/repos/wt/vickie/bdp-update-pending-review`.
Branch: `codex/vickie-bdp-pending-qualified-integration`.
No conflict/resolution edits. Tree matches the inspected ordinary merge. Relative to author head, only the previously reviewed test-local 30-second aggregate timeout patch is added. Source author checkout remains untouched. Read evidence commit ancestry and its two-file delta remain preserved; no evidence, normative spec, schema or runtime edits.

## Independent validation

Pinned Node 24.16.0 / pnpm 11.20.0. Build completed before process tests.
- Exact author: build, 97 focused process/HTTP/lifecycle/inventory tests plus 5 normative catalog tests passed.
- Exact combined candidate: build and 102 tests across six files passed in 3.37 seconds. All 797 matrix assertions passed.
- Candidate typecheck, format check, dependency boundaries, installed smoke, development smoke and evidence:verify passed.
- Lint exited zero with unchanged baseline three warnings and one informational finding.
- Evidence verifier confirmed seal 6142cf73deae687bc3e9668cc014ea408fa02843, run head 4f0ed579, evidence commit 72bdaad, 89 target rows.
- No broad full-suite run for this test-only integration; remote CI is still required. No outward writes performed.

Logs/results are adjacent, prefixed `update-pending-review-*` for author checks and `update-pending-integration-*` for candidate checks; explicit exit files cover all candidate commands. Diagnostic report: `update-pending-integration-report.json`. Test JSON: `update-pending-integration-tests.json`.
