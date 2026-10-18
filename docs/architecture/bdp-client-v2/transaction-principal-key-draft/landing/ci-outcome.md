# PR133 exact-head CI outcome

**Success**, run38789264246 attempt1/job116609269611. All27 steps succeeded, including full tests, preflight, smoke and Read evidence. Head `09bcaab75800cae89cf738ea88f012168ff733fc`; actual checkout from direct raw log `10a684d8089fa76e9c4a8ccaa3759965ad5bc412`. GitHub's immutable commit record gives ordered parents `[f5511718371a6935614eae6756e796cef71570c2, 09bcaab75800cae89cf738ea88f012168ff733fc]` and tree `fde859381752d15c8caabcbc946d3c2e384b9cf6`, identical to the candidate tree. Root owns final current-main/readiness/merge checks.

Full suite: **9310 passed,1 skipped,179 files**,764.88s. Preflight:7 passed/1file,1.44s. The new principal callback passed in1293ms; its retained CI observation records7complete responses, failureCount0, childPID20451 exit0/no signal, closed=true/stopState=stopped and directoryAbsent=true. It reports the actual synthetic checkout as declared source and claimEligible=false. This ordinary CI execution is distinct from the earlier external fresh-package positive/control evidence; neither is relabeled.

Direct raw log: `ci-final/raw-job-direct.log`,3392648 bytes,SHA256 `286c87bebbeb4a989ad235f9fac8d82ef33573414e5a2c70d7d49ce600d8f09b`. Checkout line435, principal observation line5328; exact summary lines and all step conclusions are in ci-final/summary.json. The initial gh-formatted download omitted the full-suite tail; it is preserved separately as raw-job.log and was not used to invent totals. The direct job-log read completed0 and contains the terminal full summaries.

CI's reported child/store cleanup and runner post-job/orphan cleanup are retained. This observer cannot independently inspect remote runner PIDs after completion. No local product process/store/test was started; the read-only observer and log-download commands terminated normally. Existing warnings are not represented as warning-free success.

Observation checks were spaced65seconds or more, with immutable checkout/source reads in between. No retry, cancellation, PR edit, commit, push, readiness change or merge was performed. Earlier source/review/static records remain untouched.
