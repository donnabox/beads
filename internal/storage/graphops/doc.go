// Package graphops contains private, unwired transaction read bodies for the
// proposed bead graph. It does not implement the public graphops.Reader role.
// A future protected caller must establish authority inside the same transaction
// before calling these bodies; a query handle or Scope string grants no authority.
//
// The public domain import consistently uses alias graph. Revision decoding
// preserves the domain's opaque nonempty UTF-8 law (types.go NewRevision): local
// mint/provenance certification belongs to the future protected facade. Fixture
// writers use MintRevision; opaque endpoint pins are never reformatted. Body
// Scope arguments must satisfy ValidatePersistedScopeURL and come from the same
// snapshot's Scope row in a future facade. An argument alone is no authority.
//
// Fixture helpers live only in _test.go; the wider seed API remains deferred.
// Embedded fixture qualification uses an explicitly recorded module graph; its
// actual build info is emitted. Managed-server qualification is a separate leg.
//
// Exact reads preserve private never-seen/reserved/pruned/erased absence from
// the allocation projection. Incident absence remains private errAbsent. Neither
// is a public Reader error or permission to disclose history. These bodies do
// not certify endpoint liveness, installed-type closure, ledger, lease or graph
// validity. Exact Bead/Link reads verify their own installed descriptor.
// Owned and incident collections assume installation was validated for the
// selected state; they do not certify each returned Link descriptor. A future
// facade must establish that prerequisite before interpreting the results.
// Private page bodies require separately validated selected-state physical rows
// to equal the live set; the fixture does not establish that allocation law or
// per-item visibility. Their decoded after-paths are not public cursors. Incident
// page SQL remains provisional until allocation-anchor classification composes
// into that same statement: the protected five-statement budget has no exception.
// Lookahead is validated and charged, so malformed or oversized lookahead refuses
// the entire page. SQL row limits bound transfer, not engine scans or allocation.
// These bodies make no BDP serving, migration or History claim. Only tests call
// them. Fixture schemas and writes remain test-only; there is no production seed.
package graphops
