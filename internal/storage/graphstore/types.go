// Package graphstore implements the disposable C0 graph preview in the same
// ordinary Dolt database as the existing storage schema. It does not implement
// public BDP authority, migration, restoration, or complete Memory History.
package graphstore

import (
	"encoding/json"
	"errors"

	publicops "github.com/steveyegge/beads/issueops"
)

// SchemaVersion identifies this explicitly experimental storage layout.
const SchemaVersion = 4

// Binding is the exact identity expected by the local workspace metadata.
// WorkspaceID is its canonical filesystem path; C0 does not support moving it.
type Binding struct {
	WorkspaceID   string
	ScopeURL      string
	AuthorityID   string
	SchemaVersion int
}

// Options selects an explicit backend. Existing opens never provision or migrate.
// DataDir is the already-created embedded engine directory, not a second DB.
type Options struct {
	Backend        string
	DataDir        string
	Database       string
	Branch         string
	Binding        Binding
	ServerHost     string
	ServerPort     int
	ServerUser     string
	ServerPassword string
	ServerSocket   string
	ServerTLS      bool
	// IssuePrefix is resolved by normal CLI initialization. It is used only
	// for fresh bootstrap; existing opens read the persisted config value.
	IssuePrefix string
}

// CreateRequest creates a new canonical Memory path; it is not a keyed upsert.
type CreateRequest struct {
	Path  string
	Title string
	Body  string
	Actor string
}

// Properties is the complete payload supported by the preview descriptor.
type Properties struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Record is a complete preview Memory record, including its outgoing owned Links.
// Version identifies retained state; it is not a local ordinal or timestamp.
type Record struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Revision    string            `json:"revision"`
	Version     string            `json:"version"`
	Properties  Properties        `json:"properties"`
	Owned       []json.RawMessage `json:"owned"`
	Attribution Attribution       `json:"attribution"`
}

// Attribution records only what this write supplies. RecordedAt is an observed
// UTC wall-clock time, not an acceptance-order token or an as-of boundary.
type Attribution struct {
	Actor      string `json:"actor"`
	Status     string `json:"status"`
	RecordedAt string `json:"recordedAt"`
}

// IssueRecord projects the authoritative specialized Issue aggregate. It is a
// disposable preview, with mutable issue_type classification under one local
// experimental Type. No public nominal Task/Bug contract is established.
type IssueRecord struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Revision    string            `json:"revision"`
	Version     string            `json:"version"`
	Properties  *publicops.Issue  `json:"properties"`
	Owned       []json.RawMessage `json:"owned"`
	Attribution Attribution       `json:"attribution"`
}

var (
	ErrInvalidStore   = errors.New("graph preview store identity or schema is invalid")
	ErrAlreadyExists  = errors.New("canonical graph path has already been allocated")
	ErrNotFound       = errors.New("graph resource not found")
	ErrConflict       = errors.New("graph transaction conflicted")
	ErrOutcomeUnknown = errors.New("graph transaction outcome is unknown; do not replay automatically")
)

// DependencyRequest asserts one blocking Dependency between live local Issues.
// Path is an optional never-reused Link allocation; empty allocates a fresh path.
type DependencyRequest struct {
	// ExpectedSourceRevision guards generic authoring; empty retains the legacy command contract.
	ExpectedSourceRevision string
	SourcePath             string
	TargetPath             string
	Path                   string
	Actor                  string
}

// LinkRecord projects an authoritative specialized Dependency or generic Link.
// All preview endpoints are unpinned live local Beads.
type LinkRecord struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Revision    string         `json:"revision"`
	Version     string         `json:"version"`
	Source      string         `json:"source"`
	Target      string         `json:"target"`
	Properties  map[string]any `json:"properties"`
	Attribution Attribution    `json:"attribution"`
}

type DependencyResult struct {
	Link    LinkRecord  `json:"link"`
	Source  IssueRecord `json:"source"`
	Changed bool        `json:"changed"`
}

type IssueMutationResult struct {
	Issue   IssueRecord `json:"issue"`
	Changed bool        `json:"changed"`
}

// PreviewOwnedLinkLimit is a disposable descriptor budget, not a production limit.
const PreviewOwnedLinkLimit = 1000

// LinkCreateRequest allocates an independent informational Link. Equal endpoints
// do not deduplicate intent; an explicit allocated Path cannot be reused.
type LinkCreateRequest struct {
	Path, SourcePath, TargetPath, Actor string
	Properties                          map[string]any
	ExpectedSourceRevision              string
	UnconditionalSource                 bool
}

// LinkUpdateRequest replaces properties without changing Type or endpoints.
// A guard is mandatory for the Link and, when owned, its source Memory.
type LinkUpdateRequest struct {
	Path, Actor            string
	Properties             map[string]any
	ExpectedRevision       string
	Unconditional          bool
	ExpectedSourceRevision string
	UnconditionalSource    bool
}

type LinkMutationResult struct {
	Link    LinkRecord `json:"link"`
	Source  any        `json:"source"`
	Changed bool       `json:"changed"`
}
