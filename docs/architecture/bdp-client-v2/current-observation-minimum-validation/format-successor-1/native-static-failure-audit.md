# Native actual static-prefix failure audit

Stopped before build/source. Actual typecheck raw0 accepted in4.710189999546856s and format-check raw0 accepted in3.9452561661601067s. Lint raw1/not accepted in3.8490351252257824s; its log SHA6a3d744f38b7bf428e21c3a77be32aa44da3097ee73fcc5a412ba8f267757452 matches the exit record. All three records report unchanged source identity6edbb35f, no forced cleanup/errors/survivors/runtime directories. Independently matched all three raw logs and rehashed all679 current source files to normalized binding97f12116. A fresh ps census found all five sampled prefix PIDs absent:57946,57970,59403,60798,60800.

Actual lint summary:1 error,165 warnings,9 infos;155 diagnostics not shown. The visible capped output does not identify the blocking rule/location. The visible noNonNullAssertion warnings cannot be treated as the cause. No diagnosis of a specific source defect is supported yet. A separately root-authorized read-only lint diagnostic with a sufficient finite max-diagnostics cap can reveal the withheld error while preserving this raw1 attempt; no speculative source edit or automatic retry is approved here.

No fresh-build, focused-source or review-focused-source started record exists. Thus there are no actual97-case results,9-new/28-legacy observations, source lifecycle outcomes or consumer proof to audit. Passing predecessor bootstrap and these two fresh static checks are not source/runtime success. The source-only formatter mutation remains separately attributed; original format failure stays preserved.

This audit performed reads, hashes and ps only; no lint/helper/test/build execution or source/gate mutation. Dependent work remains stopped at the first actual failure.
