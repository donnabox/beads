package schema

// R20's epoch-transition enforcement (gastownhall/beads#5898 revision 9,
// this slice: be-x5jqd.4 / #6136) adds one brand-new table,
// epoch_minted_addresses -- see
// migrations/0071_add_epoch_minted_addresses.up.sql for the full rationale.
// This migration is a plain, unguarded
// CREATE TABLE IF NOT EXISTS with no ALTER and no PREPARE, so it needs none
// of 0067/0068's CLI-hazard-override tests (already covered, for every
// migration including this one, by TestBundleMigrationsWithPreparedALTERAreOverriddenOrJustified
// and TestAllMigrationsSQLUsesDirectDDLForKnownCLIIncompatibilities).

// TestLatestVersionIncludesMigration0071 (pinning LatestVersion() == 71) is
// superseded by TestLatestVersionIncludesMigration0072
// (migration_0072_add_participation_generation_test.go) now that 0072 claims
// the next free slot -- only one such pin lives at a time, matching how this
// test itself already superseded 0070's own version.
//
// This slice originally claimed 0069 too. It was renumbered because
// R7.1's removed_restriction migration -- a PARALLEL SIBLING of this branch,
// not an ancestor -- had also claimed 0069, and the two are now linearized:
// R7.1 goes first (it is published beneath gastownhall/beads#6661), this one
// takes the next contiguous slot. Both then moved up one more slot when
// upstream's 0069_widen_issue_versions_datetime_precision
// (gastownhall/beads#6675) landed first, so R7.1 is 0070 and this is 0071.
// Two files sharing a
// version number panic checkNoDuplicateVersions at store open, before any
// command's RunE, which bricks every bd invocation rather than merely
// reddening a package -- see the same hazard on #6358 vs #6008.
