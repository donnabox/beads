package createbatchequiv

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"testing"
)

// This file is fork-owned: it exists only on the versioned-beads fork and
// belongs in no upstream-bound change.
//
// golden/*.json holds upstream's digests. They were recorded from the code
// before the batch-create fast paths existed (see GoldenDirEnv) and are what
// holds every body of this package to what that code stored, so this fork
// never edits, regenerates or skips them. Where this fork's tree stores
// different rows from upstream's by design, the difference is pinned here
// instead (architect Ruling B of be-0a9kgp), per (scenario, table): the digest
// upstream's golden states, the digest this tree stores, the rows that differ
// and the fork mechanism that accounts for them.
//
// checkGolden calls applyForkGoldenDeltas right after it reads a golden:
//
//   - A pin whose upstream digest is not what the checked-in golden states
//     fails the check, naming the entry. An upstream sync re-recorded the
//     golden, so the pin was derived against a golden that is gone and has to
//     be derived again.
//   - A pin whose fork digest equals its upstream digest fails too. The fork
//     stores what upstream stores, so the entry overrides nothing, and a dead
//     entry would hide the next real divergence.
//   - Otherwise the fork digest replaces upstream's for that one table, and the
//     check stays exact: the table has to hash to the fork digest, not merely
//     differ from upstream's.
//
// A pin is derived, never typed in: record the scenarios on this tree and on
// upstream's into scratch directories with GoldenDirEnv set, and take both
// digests and the differing rows from the two records. The fast == per-row
// comparison (compareOutcomes) never goes through any of this.

// forkMechanism names the fork behavior that accounts for a pinned delta. A
// delta no mechanism explains is a divergence nobody has approved.
type forkMechanism string

const (
	// mechanismFenceLegacyUnversioned is the design §16.2b write fence.
	// issueops.RecordVersionInTx skips an update-shaped mint for a row whose
	// participation_generation is NULL, a legacy row that never took part in
	// versioned history. The scenarios seed their issues with versioned history
	// off, so every seeded row is legacy and the fork records fewer
	// issue_versions rows than upstream does: the rows it lacks are the pinned
	// ones.
	mechanismFenceLegacyUnversioned forkMechanism = "fence-legacy-unversioned"

	// mechanismRecorderKeepsUpdatedAt is the recorder's updated_at = updated_at
	// (fork commit d5cbe223e), in both of the mint UPDATEs in
	// issueops/version_history.go. Upstream's recorder lets the column's ON
	// UPDATE clause stamp the wall clock, which moves every row it touches out of
	// the issues_updated_at view (updated_at < '2026-06-01'); the fork leaves
	// those rows in it, so its view holds the pinned rows upstream's lacks.
	mechanismRecorderKeepsUpdatedAt forkMechanism = "recorder-keeps-updated_at"
)

// forkGoldenKey names one table of one scenario's golden.
type forkGoldenKey struct{ Scenario, Table string }

func (k forkGoldenKey) String() string { return k.Scenario + "/" + k.Table }

// forkGoldenDelta is one pinned divergence from upstream's golden.
type forkGoldenDelta struct {
	Mechanism forkMechanism
	// Upstream is the digest golden/<scenario>.json states for the table, as
	// derived; the pin is stale the day that changes.
	Upstream goldenTable
	// Fork is the digest this tree stores for the table.
	Fork goldenTable
	// RowIDs are the rows on which the two differ, sorted: rows upstream's table
	// holds and the fork's lacks ("<issue id>@<revision>", for
	// mechanismFenceLegacyUnversioned), or rows the fork's table holds and
	// upstream's lacks ("<issue id>", for mechanismRecorderKeepsUpdatedAt). They
	// say what the digests stand for; the check itself compares the digests.
	RowIDs []string
}

// forkGoldenDeltas are the nine (scenario, table) entries of architect Ruling B
// of be-0a9kgp: the only tables on which this fork's drivers differ from
// upstream's pre-change golden. Every other table and every verbatim list is
// held to the golden as it stands.
var forkGoldenDeltas = map[forkGoldenKey]forkGoldenDelta{
	{"small", "issue_versions"}: {
		Mechanism: mechanismFenceLegacyUnversioned,
		Upstream:  goldenTable{Rows: 40, SHA256: "521720fb99afd6620884b3c4f309624339caefa7a9d0f93a8926727918834785"},
		Fork:      goldenTable{Rows: 37, SHA256: "e91e4e06302a9697582672591ea7037f75d4dc7dd69f4d7ab095b2b8ce5bb604"},
		RowIDs:    []string{"eq-s1@1", "eq-s2@1", "eq-s4@1"},
	},
	{"depadd", "issue_versions"}: {
		Mechanism: mechanismFenceLegacyUnversioned,
		Upstream:  goldenTable{Rows: 11, SHA256: "0298b7d35b39ee0265143f9f7d9a1adc4ad2848f2e9f19b2452a157f5bdda483"},
		Fork:      goldenTable{Rows: 0, SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		RowIDs: []string{
			"eq-db1@1", "eq-db1@2", "eq-db4@1", "eq-de.1.1@1", "eq-de.2@1", "eq-dk1@1", "eq-dl@1", "eq-dl@2",
			"eq-dl@3", "eq-dr@1", "eq-dx@1",
		},
	},
	{"waitsfor", "issue_versions"}: {
		Mechanism: mechanismFenceLegacyUnversioned,
		Upstream:  goldenTable{Rows: 15, SHA256: "a48e013d03587db247d9c2233588df4dd1062d0bbc1f53a3e119f5ad32ae1007"},
		Fork:      goldenTable{Rows: 0, SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		RowIDs: []string{
			"eq-fic@1", "eq-fic@2", "eq-fs.c1@1", "eq-fs.c1@2", "eq-fs@1", "eq-fw-i-all@1", "eq-fw-i-also@1",
			"eq-fw-i-any@1", "eq-fw-i-anyalso@1", "eq-fw-i-legacy@1", "eq-fw-w-all@1", "eq-fw-w-also@1",
			"eq-fw-w-any@1", "eq-fw-w-anyalso@1", "eq-fw-w-legacy@1",
		},
	},
	{"import458", "issue_versions"}: {
		Mechanism: mechanismFenceLegacyUnversioned,
		Upstream:  goldenTable{Rows: 398, SHA256: "92a48b220cd3ccf4348a0879729e00f2a76f365c8861e527964fd7787043d013"},
		Fork:      goldenTable{Rows: 394, SHA256: "f8f85d333902058116ce0bbd344fe728217e49e1c4557a64d82ebaa1978d7b6d"},
		RowIDs:    []string{"eq-ep@1", "eq-s1@1", "eq-s4@1", "eq-s5@1"},
	},
	{"import458-reject-stale", "issue_versions"}: {
		Mechanism: mechanismFenceLegacyUnversioned,
		Upstream:  goldenTable{Rows: 396, SHA256: "e9bea2d947cd2d2b8129de6aca0ca91ed38f5a82320bb5261fc3aa3b540bd8cd"},
		Fork:      goldenTable{Rows: 394, SHA256: "f8f85d333902058116ce0bbd344fe728217e49e1c4557a64d82ebaa1978d7b6d"},
		RowIDs:    []string{"eq-ep@1", "eq-s1@1"},
	},
	{"small", "issues_updated_at"}: {
		Mechanism: mechanismRecorderKeepsUpdatedAt,
		Upstream:  goldenTable{Rows: 41, SHA256: "2b1590dbb503fb2a1dcc71f78261f593f3c9144c4530c80325428b21d5e2164d"},
		Fork:      goldenTable{Rows: 42, SHA256: "ea303585edc660d2e4ea6e982668fefe4f85efa8b43a6e3d8bcb65850a08b856"},
		RowIDs:    []string{"eq-n2"},
	},
	{"depadd", "issues_updated_at"}: {
		Mechanism: mechanismRecorderKeepsUpdatedAt,
		Upstream:  goldenTable{Rows: 13, SHA256: "7119d0e3827b68a653156835b19fbae4b2b497aa17743ca2abce9ce387870831"},
		Fork:      goldenTable{Rows: 15, SHA256: "bf6295c3ca28217417b2e545df8d756b756f3cea3eb3d0284dadb97b7c8c5377"},
		RowIDs:    []string{"eq-db1", "eq-dl"},
	},
	{"import458", "issues_updated_at"}: {
		Mechanism: mechanismRecorderKeepsUpdatedAt,
		Upstream:  goldenTable{Rows: 401, SHA256: "a256b57e6a67ca0717d1200b4ad867abc1ccff8f9b088bc6c331c18fa6915c7f"},
		Fork:      goldenTable{Rows: 403, SHA256: "de6dbabb1bdd928fe633356cf3199f6124f440168c9ef9d56d76eea9fff0df02"},
		RowIDs:    []string{"eq-b250", "eq-b3"},
	},
	{"import458-reject-stale", "issues_updated_at"}: {
		Mechanism: mechanismRecorderKeepsUpdatedAt,
		Upstream:  goldenTable{Rows: 400, SHA256: "7b83c4335f54068965727a7230e71ad966f75fe6d23170963a142ab1654768cc"},
		Fork:      goldenTable{Rows: 402, SHA256: "80b591ea31a9821b44f006c8dca6d5390923b1e6d55689e0d776fecbb7cdf44b"},
		RowIDs:    []string{"eq-b250", "eq-b3"},
	},
}

// applyForkGoldenDeltas is checkGolden's one hook: it swaps the fork digests
// into want, the golden just read for scenario name, and stops the test on a
// stale or unneeded pin rather than compare against it.
func applyForkGoldenDeltas(t *testing.T, name string, want *golden) {
	t.Helper()
	if err := swapForkGoldenDeltas(name, want, forkGoldenDeltas); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}

// swapForkGoldenDeltas replaces, in want, the digest of each table that
// deltas pins for scenario name with the fork digest. It reports every pin of
// that scenario that is stale (the golden no longer states the digest the pin
// was derived against) or unneeded (the fork digest equals the upstream one),
// and leaves such a table as it is. Pins of other scenarios are not looked at.
func swapForkGoldenDeltas(name string, want *golden, deltas map[forkGoldenKey]forkGoldenDelta) error {
	var errs []error
	for _, key := range sortedForkGoldenKeys(deltas) {
		if key.Scenario != name {
			continue
		}
		d := deltas[key]
		have := want.Tables[key.Table]
		switch {
		case have != d.Upstream:
			errs = append(errs, fmt.Errorf("stale pin %s (%s): golden/%s.json states %s, the pin was derived against %s; "+
				"derive the entry again from a fresh record of this tree and of upstream's",
				key, d.Mechanism, name, digestText(have), digestText(d.Upstream)))
		case d.Fork == d.Upstream:
			errs = append(errs, fmt.Errorf("override %s (%s) is no longer needed: its fork digest equals the upstream digest, %s; delete the entry",
				key, d.Mechanism, digestText(d.Fork)))
		default:
			want.Tables[key.Table] = d.Fork
		}
	}
	return errors.Join(errs...)
}

// digestText renders a digest the way checkGolden's own failure does.
func digestText(g goldenTable) string { return fmt.Sprintf("%d rows %s", g.Rows, g.SHA256) }

// sortedForkGoldenKeys returns the keys of deltas by scenario, then table.
func sortedForkGoldenKeys(deltas map[forkGoldenKey]forkGoldenDelta) []forkGoldenKey {
	keys := make([]forkGoldenKey, 0, len(deltas))
	for k := range deltas {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b forkGoldenKey) int {
		return cmp.Or(cmp.Compare(a.Scenario, b.Scenario), cmp.Compare(a.Table, b.Table))
	})
	return keys
}
