# Fresh startup validation5: first nine and build

Accepted actual first9 stages, no findings, observed 2026-10-18T06:35:45.130010+00:00. All nine raw exits0/accepted:true with source unchanged, no forced cleanup/errors/remaining processes/BDP runtime directories. The previous first4 started-gate attribution is preserved separately from the next5 started gate. Actual log hashes and elapsed times are in observation.json.

Typecheck passed. Format checked11 files without fixes. Lint exited0 with20 warnings,1 info and1 diagnostic not shown; no fixes applied. This is not a warning-free lint claim. Fresh build actually ran clean and tsc -b. Bind-fresh-build passed8.165129667s and produced actual inventory `a0594be659e9b54cf43f83bfac18d2c2b0f46f056a267b709c40290dc6b17383`. Independently enumerated and rehashed all484 records:480 dist files and4 package/license/schema inputs, including all8 loader/private-registry emitted entries. No previous build identity is reused.

All682 source hashes remain sourceidentity0c53/fullpatch515e; all11 tool realpaths and byte hashes remain exact. All465 immutable preparation records other than the root-owned active gate match index and docs mirror. Actual84 sampled PIDs and9 groups are independently absent. Retained stage scratch/cache roots remain expected; no bdp-prefixed runtime directory is present below them.

At this observation, root had already released prefix15 and focused-source, its immediate audit, package-directory and protocol-pack started records existed. Their outcomes are outside this first9 observation; neither later success nor later unattempted status is claimed. The active gate snapshot is an epoch record, not the gate attributed to earlier starts.

This review used pure filesystem/hash inspection, one ps census and the mandatory session guard. No helper/product/compiler/runtime/Git/source/gate action. A read-only initial artifact filename error is recorded; the actual build-artifacts.json was then read and verified.
