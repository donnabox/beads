// Package graphstore implements the disposable C0 graph preview in the same
// ordinary Dolt database as the existing storage schema. It does not implement
// public BDP authority, migration, restoration, or complete Memory History.
package graphstore

import (
	"encoding/json"
	"errors"
)

// SchemaVersion identifies this explicitly experimental storage layout.
const SchemaVersion = 1

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

// Record is a complete preview record. Owned is always a present empty array:
// this descriptor admits no owned Links and the preview has no Link writer.
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
	ErrConflict       = errors.New("graph transaction conflicted")
	ErrOutcomeUnknown = errors.New("graph transaction outcome is unknown; do not replay automatically")
)
