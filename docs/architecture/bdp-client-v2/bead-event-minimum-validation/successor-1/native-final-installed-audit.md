# Independent native final validation audit

No blocker found. All 21 successor validation stages reached terminal accepted exit 0. No retry, forced cleanup, source change, reported survivor or remaining runtime directory appears in these records. Every log hash was independently compared. Fresh ps census found all 387 sampled PIDs absent; stage-owned temporary directories contain no remaining bdp store. No product/helper/test execution was performed by this reviewer.

The earlier native source-prefix audit remains valid: exact 25 passed source callbacks, 20 raw observations, D's ordered 37 requests in 686 ms, and separately attributed old Scope/historical controls. Current 679 source hashes and the full 476-entry build inventory match the dirty f620 successor binding. Build completed and was bound before source HTTP module collection.

Package verification independently compared 321 archive members with installed bytes and recorded source digests, checked exact member sets and both archive hashes, then rebuilt the complete consumer inventory: 870 files and 17 internal links. All eight production manifest versions/hashes match. The actual consumer lock is `ed327ba0705a91b40e34819726386a8a643bcbbb975208337dabf1994bfab926`. An independent data-only reconstruction of the recursive installed payload (14 package visits, 475 file visits) yields `cc1ccb4180a3090a80f4925bcc28aac1273b5bde1a4111de8b29f31aea537aad`, matching the verified consumer and every child report. Entry, fixture, base commit, dirty-workspace and source-binding identities agree.

The three installed selectors each executed exactly one expected callback and immediately passed their separate audit:

- New Bead: child 51297, exactly four exchanges (seed POST 200, setup GET 200, minimum GET 503, HEAD 503), exact bound URLs/order. Setup position is pos-1; timeout responses retain current context, private/no-store, retry-after 1 and development non-attesting metadata. GET body is 127 bytes; HEAD has zero bytes and matching GET content-length/media.
- Existing Scope: child 54890, three exchanges (GET 200, GET 503, HEAD 503), same paired timeout/header/body invariants. Scope route assertions remain in its pinned callback; the report does not retain URLs.
- Historical missing context: child 58492, 11 exchanges, exact retained method/URL/accept/status matrix: seed and setup, eight missing Bead/Link JSON/SSE GET/HEAD 404s, then unauthenticated 401 with no context leakage.

Each child report is present unchanged in its raw log, has outcome passed, closed=true, exit 0, null signal, no deadline or failure, a PID sampled by its own supervised stage, and an actually absent stage-owned store. Raw request-body, schema, exact Problem and validator-absence assertions are verified through the unchanged selected callback source and passing result, not invented as retained report fields. The Bead source explicitly validates completed seed/pos-1, real created Event page, Problem schema/exact body, absent ETag/Last-Modified and paired HEAD metadata.

The original failed formatter attempt remains historical. Only its two successful dependency prerequisites were explicitly reused; all 21 successor stages are fresh. This evidence identifies base f620 plus the exact dirty six-file successor, not an already committed candidate. Event, foreign-context, S1/G2 and other excluded scope remains unchanged. This audit grants no whole-profile qualification, CI success, publication or integration authority. Root owns the publication decision and any later commit-to-tested-bytes mapping.

Exact stage, artifact and observation hashes and sampled PIDs are recorded in `native-final-installed-audit.json`.
