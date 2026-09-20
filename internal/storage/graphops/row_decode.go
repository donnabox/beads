package graphops

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
)

var (
	errCorrupt = errors.New("invalid persisted graph row")
	errAbsent  = errors.New("graph row physically absent")
	errBudget  = errors.New("graph read budget exceeded")
)

type resourceRow struct {
	path, typeURL, revision string
	principal, attribution  sql.NullString
	properties              []byte
	propertiesLength        sql.NullInt64
}

type endpointRow struct {
	kind           string
	path, url, pin sql.NullString
}

type linkRow struct {
	resourceRow
	source, target endpointRow
}

// Constructors canonicalize authored input. Persisted bytes must already be
// canonical: never repair stored corruption as a side effect of reading it.
func decodeResource(row resourceRow) (graph.Revision, graph.Attribution, graph.Properties, error) {
	revision, err := graph.NewRevision(row.revision)
	if err != nil {
		return graph.Revision{}, graph.Attribution{}, graph.Properties{}, corrupt(err)
	}
	properties, err := graph.NewProperties(row.properties)
	if err != nil {
		return graph.Revision{}, graph.Attribution{}, graph.Properties{}, corrupt(err)
	}
	if !bytes.Equal(properties.Bytes(), row.properties) {
		return graph.Revision{}, graph.Attribution{}, graph.Properties{}, corrupt(errors.New("noncanonical properties"))
	}
	var attribution graph.Attribution
	if row.principal.Valid != row.attribution.Valid {
		return graph.Revision{}, graph.Attribution{}, graph.Properties{}, corrupt(errors.New("partial attribution"))
	}
	if row.principal.Valid {
		attribution, err = graph.NewAttribution(row.principal.String, graph.AttributionStatus(row.attribution.String))
		if err != nil {
			return graph.Revision{}, graph.Attribution{}, graph.Properties{}, corrupt(err)
		}
	}
	return revision, attribution, properties, nil
}

func decodeBead(row resourceRow) (graph.Bead, error) {
	revision, attribution, properties, err := decodeResource(row)
	if err != nil {
		return graph.Bead{}, err
	}
	bead, err := graph.NewBead(graph.BeadSpec{Path: row.path, TypeURL: row.typeURL, Revision: revision, Attribution: attribution, Properties: properties})
	if err != nil {
		return graph.Bead{}, corrupt(err)
	}
	return bead, nil
}

func decodeEndpoint(scope string, row endpointRow) (graph.Ref, error) {
	if row.pin.Valid && row.pin.String == "" {
		return graph.Ref{}, corrupt(errors.New("empty persisted pin"))
	}
	pin := ""
	if row.pin.Valid {
		pin = row.pin.String
	}
	switch row.kind {
	case "in":
		if !row.path.Valid || row.url.Valid {
			return graph.Ref{}, corrupt(errors.New("invalid in-scope endpoint columns"))
		}
		ref, err := graph.NewInScopeRef(row.path.String, pin)
		if err != nil {
			return graph.Ref{}, corrupt(err)
		}
		return ref, nil
	case "ext":
		if row.path.Valid || !row.url.Valid {
			return graph.Ref{}, corrupt(errors.New("invalid external endpoint columns"))
		}
		ref, err := graph.ParseRef(scope, row.url.String, pin)
		if err != nil {
			return graph.Ref{}, corrupt(err)
		}
		if ref.InScope() {
			return graph.Ref{}, corrupt(errors.New("external endpoint claims the local Scope"))
		}
		return ref, nil
	default:
		return graph.Ref{}, corrupt(fmt.Errorf("unknown endpoint kind %q", row.kind))
	}
}

func decodeLink(scope string, row linkRow) (graph.Link, error) {
	revision, attribution, properties, err := decodeResource(row.resourceRow)
	if err != nil {
		return graph.Link{}, err
	}
	source, err := decodeEndpoint(scope, row.source)
	if err != nil {
		return graph.Link{}, err
	}
	target, err := decodeEndpoint(scope, row.target)
	if err != nil {
		return graph.Link{}, err
	}
	link, err := graph.NewLink(graph.LinkSpec{Path: row.path, TypeURL: row.typeURL, Revision: revision, Source: source, Target: target, Attribution: attribution, Properties: properties})
	if err != nil {
		return graph.Link{}, corrupt(err)
	}
	return link, nil
}

func decodeDescriptor(id string, raw []byte, fingerprint string, kind graph.ResourceKind) (graph.TypeDescriptor, error) {
	descriptor, err := graph.ParseTypeDescriptor(raw)
	if err != nil {
		return graph.TypeDescriptor{}, corrupt(err)
	}
	if descriptor.ID() != id || descriptor.Describes() != kind || descriptor.Fingerprint() != fingerprint || !bytes.Equal(descriptor.CanonicalJSON(), raw) {
		return graph.TypeDescriptor{}, corrupt(errors.New("descriptor identity, kind, fingerprint or canonical bytes mismatch"))
	}
	return descriptor, nil
}

func corrupt(cause error) error { return fmt.Errorf("%w: %v", errCorrupt, cause) }
