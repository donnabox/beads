package translate

import (
	"fmt"
	"sort"
)

// TableClass says what replay does with a table that a commit step changed.
type TableClass int

const (
	// TableReplayed tables are replayed: their row diffs become bd actions, and
	// the comparison covers them.
	TableReplayed TableClass = iota + 1
	// TableDerived tables are written by bd or dolt themselves, as a side effect
	// of a replayed write or when a store is initialized, migrated or exported,
	// so they are no input. A change to one is counted, never replayed.
	TableDerived
	// TableUnsupported tables hold real history the translator cannot replay yet.
	// A change to one is a recorded coverage gap, never a silent no-op.
	TableUnsupported
)

func (c TableClass) String() string {
	switch c {
	case TableReplayed:
		return "replayed"
	case TableDerived:
		return "derived"
	case TableUnsupported:
		return "unsupported"
	}
	return fmt.Sprintf("TableClass(%d)", int(c))
}

// TablePolicy is what the translator does with one table, and why.
type TablePolicy struct {
	Class  TableClass
	Reason string
}

// tablePolicy classifies every table a bd store versions, including tables that
// older stores versioned and current ones ignore, and the system tables dolt
// versions alongside them. A table that is not here is a coverage gap flagged
// unknown, and the drift test fails until someone decides which class it is.
var tablePolicy = map[string]TablePolicy{
	"issues":       {TableReplayed, "the issue rows: created, updated, closed and deleted through bd"},
	"dependencies": {TableReplayed, "the edges between issues: added and removed through bd dep, each owned by its issue_id end"},

	"events":               {TableDerived, "audit events bd appends as a side effect of every write; dolt-ignored in current stores"},
	"child_counters":       {TableDerived, "per-parent id counters bd advances as a side effect of creating children"},
	"issue_counter":        {TableDerived, "the id counter bd advances as a side effect of creating issues"},
	"schema_migrations":    {TableDerived, "bd's own migration bookkeeping; the work store is initialized by the integration build, which writes its own rows, so the oracle's migration history is no input"},
	"dolt_ignore":          {TableDerived, "ignore patterns bd writes when it initializes or migrates a store; the work store gets its own"},
	"dolt_nonlocal_tables": {TableDerived, "dolt's list of tables shared across branches, set by bd at initialization; the work store gets its own"},
	"dolt_schemas":         {TableDerived, "views and procedures bd's migrations create; the integration build creates its own"},
	"dirty_issues":         {TableDerived, "export-tracking marks bd sets as a side effect of every issue write"},
	"export_hashes":        {TableDerived, "hashes bd records when it exports issues; no user mutation writes them"},

	"labels":                 {TableUnsupported, "label rows: replayable with bd label once labels and comments are replayed; a recorded gap until then"},
	"comments":               {TableUnsupported, "comment rows: replayable with bd comments once labels and comments are replayed; a recorded gap until then"},
	"config":                 {TableUnsupported, "bd's key-value settings (bd config set): real history with no replay yet"},
	"metadata":               {TableUnsupported, "store-level key-value rows written at creation: real history with no replay yet"},
	"custom_statuses":        {TableUnsupported, "user-declared statuses: real history with no replay yet"},
	"custom_types":           {TableUnsupported, "user-declared issue types: real history with no replay yet"},
	"federation_peers":       {TableUnsupported, "federation peer registrations: real history with no replay yet"},
	"interactions":           {TableUnsupported, "agent interaction log rows: real history with no replay yet"},
	"routes":                 {TableUnsupported, "cross-store routing entries: real history with no replay yet"},
	"repo_mtimes":            {TableUnsupported, "file-mtime cache for multi-repo hydration: real history with no replay yet"},
	"compaction_snapshots":   {TableUnsupported, "snapshots taken by compaction: real history with no replay yet"},
	"issue_snapshots":        {TableUnsupported, "issue snapshots taken by compaction: real history with no replay yet"},
	"provenance_events":      {TableUnsupported, "provenance log rows: real history with no replay yet"},
	"store_epoch":            {TableUnsupported, "versioned-history epoch marker: real history with no replay yet"},
	"epoch_minted_addresses": {TableUnsupported, "versioned-history address registry: real history with no replay yet"},
	"issue_versions":         {TableUnsupported, "versioned-history snapshots: how the candidate side reads them is a later decision, so these rows are real history with no replay yet"},
}

// PolicyFor returns the policy for table, and false when the table is not
// classified.
func PolicyFor(table string) (TablePolicy, bool) {
	p, ok := tablePolicy[table]
	return p, ok
}

// PolicyTables returns the name of every classified table, sorted.
func PolicyTables() []string {
	names := make([]string, 0, len(tablePolicy))
	for name := range tablePolicy {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
