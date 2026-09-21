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
// the allocation projection. Paged incident anchors share these private facts;
// old unpaged incident absence remains errAbsent. Neither
// is a public Reader error or permission to disclose history. These bodies do
// not certify endpoint liveness, installed-type closure, ledger, lease or graph
// validity. Exact Bead/Link reads verify their own installed descriptor.
// Owned and incident collections assume installation was validated for the
// selected state; they do not certify each returned Link descriptor. A future
// facade must establish that prerequisite before interpreting the results.
// Private page bodies require separately validated selected-state physical rows
// to equal the live set; the fixture does not establish that allocation law or
// per-item visibility. Their decoded after-paths are not public cursors. Incident
// pages compose allocation-anchor classification in that same statement. Their
// private entry requires a finite context and refuses oversized canonical path
// operands with errBudget before SQL; Link collection entry is unchanged. The
// protected five-statement budget has no exception.
// Lookahead is validated and charged, so malformed or oversized lookahead refuses
// the entire page. SQL row limits bound transfer, not engine scans or allocation.
// These bodies make no BDP serving, migration or History claim. Only tests call
// them. Fixture schemas and writes remain test-only; there is no production seed.
//
// Beads pages batch raw rows (including charged lookahead), distinct descriptors,
// and complete owned expansions in at most three statements. They require a
// finite context and share conservative private row/byte/group caps. A legal
// singleton may exceed page capacity: installer/effective-capacity reconciliation
// is mandatory before any public traversal guarantee. The decoded afterPath is
// not a public cursor; lookahead descriptor validity is not continuation validity.
//
// The private read-attempt mechanism passively reloads witness facts and owns
// four observations, comparison, body dispatch and cooperative checked cleanup.
// It has no issuer, production resource adapter, Evidence provider or public role.
// Load errors precede SQL; successfully loaded absent/pending facts retain the
// comparator's binding/Scope/witness ordering. No diagnostic admission is implied.
// Work preserves its exact context; cleanup uses a separate finite context.
// Cleanup/cancellation failures withhold results and result-bearing error chains.
// A cooperative deadline and sql.ErrTxDone do not prove physical terminality;
// qualified rollback/connection drain and full claim-era matching remain gates.
// Body Goexit still cleans; Goexit within a cleanup method is outside its
// cooperative contract. Panic(nil) controls qualify the default Go1.26 behavior.
package graphops
