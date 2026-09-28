// Package graphstore implements the disposable C0 graph preview in the same
// ordinary Dolt database as the existing storage schema. It does not implement
// public BDP authority, migration, restoration, or complete Memory History.
package graphstore

import (
	"encoding/json"
	"errors"
)

// SchemaVersion identifies this explicitly experimental storage layout.
const SchemaVersion = 5

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

var (
	ErrInvalidStore   = errors.New("graph preview store identity or schema is invalid")
	ErrAlreadyExists  = errors.New("canonical graph path has already been allocated")
	ErrNotFound       = errors.New("graph resource not found")
	ErrGone           = errors.New("graph resource was deleted")
	ErrConflict       = errors.New("graph transaction conflicted")
	ErrOutcomeUnknown = errors.New("graph transaction outcome is unknown; do not replay automatically")
)

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

type LinkTombstone struct {
	ID              string      `json:"id"`
	Type            string      `json:"type"`
	Revision        string      `json:"revision"`
	Version         string      `json:"version"`
	State           string      `json:"state"`
	PreviousVersion string      `json:"previousVersion"`
	Attribution     Attribution `json:"attribution"`
}

// PreviewOwnedLinkLimit bounds required owned-state acquisition.
const PreviewOwnedLinkLimit = 1000

var (
	ErrCapabilityUnavailable = errors.New("graph preview capability is unavailable")
	ErrLimitExceeded         = errors.New("graph preview result limit exceeded")
)
