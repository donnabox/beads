// Package canonicalbatch compiles canonical batch authoring into the explicitly
// selected experimental version-2 request. It does no I/O and executes no writes.
// The legacy command-text runner and retained version-1 grammar are untouched.
package canonicalbatch

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/steveyegge/beads/graphops"
)

const Version = "2"
const MaxInputBytes = 4 << 20
const MaxOperations = 1000

var (
	ErrInvalid     = errors.New("invalid canonical batch")
	ErrUnsupported = errors.New("unsupported canonical batch grammar")
	ErrLimit       = errors.New("canonical batch input limit exceeded")
)

// Options explicitly selects the authoring format. Native JSON disallows a
// separate version or arguments. Scope resolves relative authoring References.
// Arguments use NAME=REFERENCE; split occurs only at the first equals sign.
type Options struct {
	Format  string
	Version string
	Args    []string
	Scope   string
}

// Reference is the ordinary URI or exact {uri,revision} protocol sum.
type Reference struct {
	URI      string
	Revision string
}

func (r Reference) MarshalJSON() ([]byte, error) {
	if r.Revision == "" {
		return json.Marshal(r.URI)
	}
	return json.Marshal(struct {
		URI      string `json:"uri"`
		Revision string `json:"revision"`
	}{r.URI, r.Revision})
}

// Operand adds an explicit binding arm at admitted Resource input positions.
// Its fields are mutually exclusive. Empty Operand denotes an absent member.
type Operand struct {
	URI      string
	Revision string
	Binding  string
}

func (o Operand) Present() bool { return o.URI != "" || o.Binding != "" }
func (o Operand) MarshalJSON() ([]byte, error) {
	if o.Binding != "" {
		if o.URI != "" || o.Revision != "" {
			return nil, invalid("binding cannot carry URI or revision")
		}
		return json.Marshal(struct {
			Binding string `json:"binding"`
		}{o.Binding})
	}
	return (Reference{URI: o.URI, Revision: o.Revision}).MarshalJSON()
}

// Member is a detached view of one fully parsed member. Nil RawMessages mean
// omission; explicit empty objects and empty changes remain distinguishable.
// Kind is the statically known result kind; Query remains the canonical shared
// AST envelope. Execution must additionally admit its supported query semantics.
type Member struct {
	Operation          string
	Name               string
	ID                 string
	Type               Reference
	Bead               Operand
	Link               Operand
	Source             Operand
	Target             Operand
	Properties         json.RawMessage
	Metadata           json.RawMessage
	PropertiesChange   json.RawMessage
	MetadataChange     json.RawMessage
	RevisionProvenance json.RawMessage
	ExpectedRevision   string
	Query              json.RawMessage
	Kind               graphops.ResourceKind
	Set                bool
}

// Batch owns all frozen input and returns copies on every accessor.
type Batch struct {
	native  []byte
	members []Member
}

func (b *Batch) Version() string    { return Version }
func (b *Batch) NativeJSON() []byte { return bytes.Clone(b.native) }
func (b *Batch) Members() []Member {
	out := make([]Member, len(b.members))
	for i, m := range b.members {
		out[i] = cloneMember(m)
	}
	return out
}
func cloneMember(m Member) Member {
	m.Properties = bytes.Clone(m.Properties)
	m.Metadata = bytes.Clone(m.Metadata)
	m.PropertiesChange = bytes.Clone(m.PropertiesChange)
	m.MetadataChange = bytes.Clone(m.MetadataChange)
	m.RevisionProvenance = bytes.Clone(m.RevisionProvenance)
	m.Query = bytes.Clone(m.Query)
	return m
}
