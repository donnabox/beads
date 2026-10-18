# Independent first installed outcome audit

**Medium — incorrect test oracle blocks installed completion.** The installed test at apps/bdptest/src/transactional-selector-admission-installed.test.ts:324–327 expects `retry: "never"` on the missing `beads/late` GET. Existing Read error law in docs/specs/bdp.md:1901 assigns `resource-not-found` /404 /`after-state-change`; read-request.ts:266 returns the shared `readProblem("resource-not-found")`. The actual response conforms. Correct only that expected retry string; preserve404/code/position/absence and all other oracles. This defect was missed by my earlier source review; the actual run exposed it. It is not evidence of a new runtime defect or permission to rerun.

Actual root run against22777eccc347b9c926de1da92409de9a924ba692 completed five responses and then failed this assertion. The original25-request cohort has not passed. Retained report identifies the verified installed entry, source attribution, failureCount1, child53950 exit0 without signal, childAbsenttrue and directoryAbsenttrue. I independently read the failure JSON/log, confirmed the five observations, found PIDs53941/53945/53950 absent and the owned installed-selector.tmp empty. Supervisor records exit1, no timeout/forced cleanup/errors, no surviving process or runtime directory and unchanged source. No retry was performed by this reviewer.

Package stages remain valid and distinct: all eight preparation/pack/install/verification commands exited0 with unchanged source, clean process/runtime state, and only three identified node-compile-cache directories retained. I independently compared all313 archive members (63protocol,250server) with actual installed bytes and source-file hashes, checked actual member sets, manifest hashes, dependency realpath resolutions, entry hash and consumer-lock hash.312 source-inventory entries count the shared LICENSE once. The fixed eight-package closure is present with local protocol/server tarball resolution; no registry fallback is reported. Package correctness does not turn the failed service cohort into a pass.

Consumer lock SHA `c978c9dfddddb798aa4d6da207031a4d89d66bbd4ed305248f45a484eff4bc49`; installed provenance SHA `d4e65b6c933f3c31ec8a4d43bb545fbda265607fb4a26fc2114f4e4d55e5f746`; payload identity `144f8c506ad961b6be61f2e25cf19ac8a2a862b25df1ae4edbfd5bdff7b7db7f` remains the separately computed package payload, not child-attested Git provenance.

Actual failure artifacts:
- installed-selector.results.json `e4051cc70379499888b532b252845e10f81fab53f096d2b2853a3ad130134258`.
- installed-selector.log `cd087bd70553fc19e07f4ac95cf7d6500f57fa7dd39ee58e62ab6a7ee551ce72`.
- installed-selector.exit.json `f9596b984d58864ca9bc385e79b0ed025259359f70096a124b4a5eb0078f5e26`.

The later test-only correction requires its own exact delta review and new root execution disposition; original source22777 and these failed five-response results remain historical. No product/test/build/install/service command was executed by this reviewer; only read-only inspection plus guarded report write.
