package driver

import (
	"context"
	"time"
)

// Event is something a run tells its Observer. The set is closed: the types
// below are the whole vocabulary, so a consumer can switch on them and know it
// has seen every kind.
type Event interface {
	event()
}

// Observer is told about a run as it goes. It is how anything that measures or
// reports on a run attaches to it without changing the loop. A nil Observer is
// valid and hears nothing. An Observer that returns an error fails the run.
type Observer interface {
	Observe(ctx context.Context, ev Event) error
}

// ObserverFunc is an Observer that is a function.
type ObserverFunc func(ctx context.Context, ev Event) error

// Observe calls f.
func (f ObserverFunc) Observe(ctx context.Context, ev Event) error { return f(ctx, ev) }

// StepIssueResult is emitted once per (step, issue), right after the row for it
// was written, in step order and then issue order.
type StepIssueResult struct {
	// Step is the step's position in the run, from 0.
	Step int
	// Merge and Net are the step's tags: a merge step, and a sampled step that
	// spans several commits.
	Merge bool
	Net   bool
	// WriteLatency is the time spent running the issue's actions; zero when none ran.
	WriteLatency time.Duration
	// Result is the row that was written.
	Result CommitReplayResult
}

// SeedMigrating is emitted by a seeding step after the seed copy is prepared and
// immediately before the first bd command on it. Every bd open of a copy at an
// older schema is the migration, so this is the last moment at which a sample of
// the copy is a sample from before the migration. The loop declares the type and
// never emits it.
type SeedMigrating struct{}

// SeedRecord is what a seeding step records about the copy it made: what it
// removed so the copy cannot reach a remote, the schema versions around the
// migration, what the migration cost, and where the copy stands. It is carried by
// SeedCompleted and reported in the summary.
type SeedRecord struct {
	RemotesStripped     int     `json:"remotes_stripped"`
	BackupsStripped     int     `json:"backups_stripped"`
	IgnoredRowsCleared  int     `json:"ignored_rows_cleared"`
	SchemaBefore        int     `json:"schema_before"`
	SchemaAfter         int     `json:"schema_after"`
	IgnoredSchemaBefore int     `json:"ignored_schema_before"`
	IgnoredSchemaAfter  int     `json:"ignored_schema_after"`
	MigrationSeconds    float64 `json:"migration_seconds"`
	MigrationCommits    int     `json:"migration_commits"`
	SeedHead            string  `json:"seed_head"`
	DoltCLIVersion      string  `json:"dolt_cli_version"`
	LinkedEngine        string  `json:"linked_engine"`
}

// SeedCompleted is emitted by a seeding step after its final assertions, carrying
// the seed record. The loop declares the type and never emits it.
type SeedCompleted struct {
	SeedRecord
}

func (StepIssueResult) event() {}
func (SeedMigrating) event()   {}
func (SeedCompleted) event()   {}
