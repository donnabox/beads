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
// Physical absence is private errAbsent, not a public not-found/gone decision:
// the future facade must consult allocation/visibility state. These bodies do
// not certify endpoint liveness, installed-type closure, ledger, lease or graph
// validity. Exact Bead/Link reads verify their own installed descriptor.
// Owned and incident collections assume installation was validated for the
// selected state; they do not certify each returned Link descriptor. A future
// facade must establish that prerequisite before interpreting the results.
// These bodies make no BDP serving, migration or History claim. Only tests call
// them. Fixture schemas and writes remain test-only; there is no production seed.
package graphops
