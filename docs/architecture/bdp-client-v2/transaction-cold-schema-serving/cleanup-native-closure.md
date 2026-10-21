# Native closure — schema preflight cleanup Low

**Closed. No remaining actionable finding in this exact cleanup correction.** This closes only the Low in native-preflight.md (ec9e0475); it is not full code council, evidence review, compilation success or a public/CI/merge claim.

Read cleanup-only.diff, cleanup-correction.json, updated preflight and root-measured-acceptance.md. Independently compared immutable initial-preflight snapshots with the corrected artifacts rather than relying on the author's now-active mutable checkout. No tests, builds, services, installs, network or source changes were performed.

The only source change is the new adapter test's afterEach hook. It now catches each cleanup callback failure and continues through the remaining callbacks, attempts mock restoration in finally, collects restoration errors too, and throws an AggregateError after all attempts. A failed runtime close therefore no longer skips directory removal or mock restoration. The hook does not catch or replace the test-body failure; its own cleanup error remains separately reportable. No historical service-test hook was refactored.

Independent snapshot verification confirms:

- Initial preflight retains9c92b1ba854586937308834d3b03ee6c6c9ec7fe34cdbf4e93fbcf1a6c5e32ad.
- Updated preflight:b6007591a920b967efb09e35ab5c2a58e24e82973fd2c3f7efa01e1b8ed66bc7.
- Updated bounds:87ac550bbae1df2f907cfce5a51915ee1540d1a4811141d8b76302c2e6e953bc.
- Updated complete diff:f2b2c8366c3b610ab9c6f3a780fc6af0ed6c42f4047e4c4ee8f3a370e851695e.
- Adapter-test bytes change15987→16377 and SHA256fcbf9d9e41505ef987ae17cb6de6d5faa9fef8ec7d080344f92f2e9608143800→9c1b765e72b8be893486c32e4e10860dae422332008af8aa005ce30f0171ecc4. Reconstructed before/after contents from the complete diffs produce exactly cleanup-only.diff.
- The other eight files are unchanged. All five control patches and mutant identities, capture/route/body measurements, source counts, callback expectations and execution budgets are unchanged. Complete source count increases only by16 additions (1089 additions/5 deletions).

Thus source10 cases/38 HTTP/seven listeners plus two owners, derivative expected10/hard32 HTTP/three listeners, and installed14/one plus predecessor expected3/one (17 expected/hard28 across two children) retain their reviewed meanings. No extra cleanup test, request, listener or callback was introduced by this correction.

Root has released that enumerated workload in root-measured-acceptance.md, independently of the parent merge while preserving exact-head CI gates. Actual source/static/control/installed results and any later narrow typing changes remain for the frozen candidate's code/evidence review. Independent installed PID-absence verification is still required separately; logged exit values alone are not that check.
