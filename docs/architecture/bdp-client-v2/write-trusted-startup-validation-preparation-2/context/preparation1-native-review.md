# Startup validation preparation native review

Revisions required. Preparation1 stays CLOSED. Two narrow helper corrections are needed; no product-source change or new actual reviewer attempt is requested.

1. **N1, medium — actual owned guard closure is not tracked.** In bound-consumer-loader.mjs479/488–493/668–669, a close timeout rejects the witness but is swallowed into workError, then workSettled becomes true. Final cleanup checks only workSettled and server.listening. The latter can already be false before close callback settlement. A failed run may therefore remove its evidence directory and clear its60s fail guard while owned close remains unknown. It does not become a false passing outcome, but violates the explicit preservation/closure rule. Track actual successful close callback settlement; close timeout/error must mark unknown, preserve the directory and retain the fail guard within the unchanged original cleanup/deadline bounds.

2. **N2, medium — missing required failure counters are accepted.** audit.py31–32 makes numFailedTests and numFailedTestSuites optional alongside optional runtime errors. This contradicts the packet's requiredResultKeys and weakens the accepted predecessor. Require both counters as typed integer0; keep numRuntimeErrorTestSuites optional. No new report framework is needed.

Full actual helper/program review otherwise supports the bounded mechanics: exact3-new-file HEAD/index/status protections and682/679/670 source maps;44 prospective callbacks;25 ordered gated stages; fresh tools/build/archive/actual-lock/consumer-proof barriers;264 emitted entries/eight new outputs;32 dynamic plus39 explicit resolver refusals; six real loader calls with2 accepted/4 refused across RU/TX. Full configuration/artifact bodies, public/private/declaration checks and intended-port/store distinctions are retained. Stock smoke's separate consumer provenance limitation is explicit.

Verified all106 indexed candidate artifacts plus index/mirror,101 implementation identities,5 reference pins,15 genuine adaptation/inverse pairs, complete external fragment origins and all682 current source hashes. Static reads/hash comparisons and Python AST parsing only. No helper, product, compiler, test, build, install, runtime, reviewer, git, source or flag action. These are static control-flow findings, not runtime reproductions.

Prepare a separately named minimal successor for N1/N2, then obtain exact native/root correction acceptance before any execution. No deadline widening, retry or infrastructure expansion.
