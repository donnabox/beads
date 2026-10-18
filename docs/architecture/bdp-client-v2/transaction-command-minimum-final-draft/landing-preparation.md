# PR123 landing-check preparation

Prepared, **not executed**. Script `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-command-minimum-final-draft/landing-check.py` SHA256 `8d63c1e06731124d1b6af8238527a1452b2c534cfc042a6fda9bb5d8d93a2bc1`. No product/source/test/build/install/network mutation, readiness transition, merge or PR comment. Only external preparation artifacts were written under the session guard. CI38645926223/job116132639415 was supplied as in progress; no successful outcome is inferred.

Exact pins: PR123, branch codex/vickie-command-minimum, headf2ac3f980fc8d35a3b485db04b813ad1b569d94d, base122800511ee6306e4105e1f73bf3658f35c90a14, tree74327f052cef6026b3ea10e186523ace9c87a128. Twelve current source/review/disposition/acceptance hashes are embedded and copied in landing-review-pins.json. Four exact final source/control/installed evidence hashes are embedded and copied in landing-evidence-pins.json. Pending final actual audit and root landing acceptance are **not fabricated**.

Default mode is read-only readiness. It requires a clean exact local source/single parent/seven hashes, published exact head, actual main and PR base still1228005, every status check successful, exact CI run/job completed successfully and every named install/format/lint/boundary/type/build/full-test/preflight/smoke/Read step successful. It reads the actual complete CI job log, binds its hash to root acceptance, extracts the actual checkout commit and requires its real GitHub tree74327/ordered parents[1228005,f2ac]. Since the head is a direct child of the unchanged base, a normal no-ff merge necessarily preserves its tree; the script does not fetch or write Git objects. Any main/base advance fails closed for explicit separately reviewed integration.

`--verify-merged` does not merge. It reads the already-merged PR and actual remote main, requires the PR merge commit to equal that main, and verifies actual tree74327 with exactly two ordered parents[1228005,f2ac]. Both modes finish with fresh PR/check/main readback. This is observational; root's actual merger must still independently recheck session/base/head immediately and use head matching before mutation, then run actual merge readback. No automatic retry or synthesized merge commit.

Root must create `root-landing-acceptance.json` only after final evidence audit and CI log review. Required fields:
- accepted:true, head/tree/base exact pins, pr:123, ciRun:"38645926223", ciJob:"116132639415", publicQualification:false.
- reviewHashes equal landing-review-pins.json; evidenceHashes equal landing-evidence-pins.json.
- finalActualEvidenceAudit: {path: relative path under L, sha256: actual64hex, accepted:true}. File must exist and hash-match; no placeholder default.
- ciLogSha256: SHA256 of exact text returned by the script's `gh api .../logs` with surrounding whitespace stripped then one newline added (same previous landing-check normalization). Root must bind the genuinely successful full log.

No acceptance file was created. Missing/pending fields are fatal. Outputs, only after all checks pass, are merge-readiness.json or actual-merge-verification.json; releaseAuthorized:false and publicQualification:false remain explicit. The script never runs `gh pr merge`, `gh pr ready`, push, fetch, checkout, tests or service commands. It uses only read-only gh views/API GET, ls-remote and local Git reads, plus local report output.

Proposed future invocations after root review, not executed here:
`python3 L/guarded-action.py python3 /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-command-minimum-final-draft/landing-check.py`
`python3 L/guarded-action.py python3 /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-command-minimum-final-draft/landing-check.py --verify-merged`

The script has been drafted/read as text only in this task; neither it nor its remote checks were executed. Root/native review precedes use. Current prepared scope does not claim completed CI, final actual audit, profile qualification or merge readiness.
