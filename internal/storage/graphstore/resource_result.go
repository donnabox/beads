package graphstore

import (
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/resourceget"
)

// ResourceDocument is the immutable SAME-transaction applied-result facet.
// It is populated only by the checked14 native writer before commit, and is
// absent on legacy11 results. It performs no postcommit lookup, selects no new
// head, changes no legacy JSON and grants no authority to a caller-built DTO.
func (r Record) ResourceDocument() (resourceget.ResourceDocument, bool) {
	if r.resourceDocument == nil || r.resourceDocument.Kind() != graph.KindBead || r.resourceDocument.ID() != r.ID || r.resourceDocument.Revision() != r.Revision || r.resourceDocument.Type().URI != r.Type {
		return resourceget.ResourceDocument{}, false
	}
	return *r.resourceDocument, true
}

func (r IssueRecord) ResourceDocument() (resourceget.ResourceDocument, bool) {
	if r.resourceDocument == nil || r.resourceDocument.Kind() != graph.KindBead || r.resourceDocument.ID() != r.ID || r.resourceDocument.Revision() != r.Revision || r.resourceDocument.Type().URI != r.Type {
		return resourceget.ResourceDocument{}, false
	}
	return *r.resourceDocument, true
}

func (r LinkRecord) ResourceDocument() (resourceget.ResourceDocument, bool) {
	if r.resourceDocument == nil || r.resourceDocument.Kind() != graph.KindLink || r.resourceDocument.ID() != r.ID || r.resourceDocument.Revision() != r.Revision || r.resourceDocument.Type().URI != r.Type {
		return resourceget.ResourceDocument{}, false
	}
	return *r.resourceDocument, true
}
