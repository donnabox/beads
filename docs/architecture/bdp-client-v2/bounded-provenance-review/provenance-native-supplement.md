# Native Codex supplement — projection pin and metadata Event repair

Reviewed exact PR65 source `6b64c1cede8e434ae6e09781b372a8c899f7bc91` against frozen review head `d7fccb4390eaf1ad43fda3464f6b407cf393b5f1`, and separately reviewed PR64 prose commit `c1712f2d26f889f723ffb79c01146c87883a65df`. Source checkouts were not modified. No new finding in this bounded supplement.

**The current schema-projection pin finding is closed.** `744fa9e060d5fc33a90fe9b94b242a3e2f484c06` changes the current regression pin to the independently observed `387cd60fc2a253885ae7cedbb9ee491a3dfd832af21e8c093664047f13c41dac`. The schema projection tests now pass on a fresh external archive of `6b64c1c`.

**The inherited owned-Link Event prose finding is closed.** PR64 `c1712f2` now compares both properties and metadata for a no-op; applies each present delta to its own object while preserving omitted values; and describes complete created metadata and properties/metadata updated deltas with the historical-Event omission exception. This implements the already-adopted common metadata contract. It introduces no additional Event, terminal deletion revision, source-provenance feature or new product semantics.

The PR65 merge preserves both sides of the conflict: `link.metadataChange` is applied to Link metadata, and the complete new `revisionProvenance` envelope replaces or is removed from both Link/source records as appropriate. No old `attribution` field was reintroduced in that instruction. The existing consumer predecessor checks and owned-source event behavior remain intact.

Independent validation used Node 24.16.0 and a separate external `git archive` of `6b64c1c`: **208 tests passed across 5 files** covering schema Read projection, general catalog, Read+Update catalog, Transactional catalog, and Transactional wire tests. See `native-supplement-tests.log`. This supplements the earlier 1,007 passing tests; it does not relabel the old frozen-head failures as passes.

The four-file successor delta includes inherited PR64 evidence artifact/constant changes. Those observations qualify the PR64 source only. This supplement does not claim fresh PR65 qualification: the common-envelope cohort remains a separate required actual run and evidence verification, explicitly pending at review time. Full final validation and any later evidence-only successor remain root-owned.
