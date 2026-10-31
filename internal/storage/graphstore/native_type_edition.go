package graphstore

import (
	"bytes"
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/resourceget"
)

// This member belongs ONLY to the private retained-version blob. It is never
// part of a public Resource, user metadata, a legacy11 codec or Type identity.
// Keeping the original domain fields at the root preserves existing backing
// queries of owned/state/previousVersion. Native-domain validators receive the
// unchanged domain bytes after this member has been checked and removed.
const nativeTypeEditionMember = "$nativeTypeEdition"
const nativeTypeEditionCodec = "native-type-edition/1"

type nativeTypeEdition struct {
	Codec string                    `json:"codec"`
	Type  resourceget.TypeReference `json:"type"`
}

func validateNativeTypeEdition(value nativeTypeEdition) error {
	if value.Codec != nativeTypeEditionCodec || graph.ValidateTypeURL(value.Type.URI) != nil || validateVersionToken(value.Type.Revision) != nil {
		return fmt.Errorf("%w: native exact Type edition", ErrInvalidStore)
	}
	return nil
}

// These pure codecs establish representation only. A writer must obtain its
// pin from the checked stored definition, or preserve its exact predecessor's
// recorded pin. Neither codec accepts a missing pin or selects current/latest.
func encodeNativeSnapshotEdition(domain []byte, pin resourceget.TypeReference) ([]byte, error) {
	edition := nativeTypeEdition{Codec: nativeTypeEditionCodec, Type: pin}
	if err := validateNativeTypeEdition(edition); err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(domain, &fields) != nil || fields == nil {
		return nil, fmt.Errorf("%w: native retained record object", ErrInvalidStore)
	}
	if _, present := fields[nativeTypeEditionMember]; present {
		return nil, fmt.Errorf("%w: duplicate native Type edition carrier", ErrInvalidStore)
	}
	var nominal string
	if json.Unmarshal(fields["type"], &nominal) != nil || nominal != pin.URI {
		return nil, fmt.Errorf("%w: native nominal Type differs from exact pin", ErrInvalidStore)
	}
	encoded, err := canonicalJSON(edition)
	if err != nil {
		return nil, err
	}
	fields[nativeTypeEditionMember] = encoded
	return canonicalJSON(fields)
}

func decodeNativeSnapshotEdition(raw []byte) ([]byte, resourceget.TypeReference, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, resourceget.TypeReference{}, fmt.Errorf("%w: native retained record object", ErrInvalidStore)
	}
	carrier, present := fields[nativeTypeEditionMember]
	if !present {
		return nil, resourceget.TypeReference{}, fmt.Errorf("%w: native retained exact Type evidence missing", ErrInvalidStore)
	}
	var edition nativeTypeEdition
	if json.Unmarshal(carrier, &edition) != nil {
		return nil, resourceget.TypeReference{}, fmt.Errorf("%w: native retained Type evidence", ErrInvalidStore)
	}
	if err := validateNativeTypeEdition(edition); err != nil {
		return nil, resourceget.TypeReference{}, err
	}
	canonical, err := canonicalJSON(edition)
	if err != nil || !bytes.Equal(carrier, canonical) {
		return nil, resourceget.TypeReference{}, fmt.Errorf("%w: native retained Type carrier encoding", ErrInvalidStore)
	}
	var nominal string
	if json.Unmarshal(fields["type"], &nominal) != nil || nominal != edition.Type.URI {
		return nil, resourceget.TypeReference{}, fmt.Errorf("%w: native retained Type pin mismatch", ErrInvalidStore)
	}
	delete(fields, nativeTypeEditionMember)
	domain, err := canonicalJSON(fields)
	if err != nil {
		return nil, resourceget.TypeReference{}, err
	}
	// Canonical round-trip also rejects duplicate members and alternate carrier
	// encodings rather than silently projecting one interpretation of a blob.
	canonical, err = encodeNativeSnapshotEdition(domain, edition.Type)
	if err != nil || !bytes.Equal(raw, canonical) {
		return nil, resourceget.TypeReference{}, fmt.Errorf("%w: native retained edition encoding", ErrInvalidStore)
	}
	return domain, edition.Type, nil
}

type nativeIssueTypeEdition struct {
	Codec   string                    `json:"codec"`
	Type    resourceget.TypeReference `json:"type"`
	Context json.RawMessage           `json:"context"`
}

func encodeNativeIssueTypeEdition(value issueRevisionContext, pin resourceget.TypeReference) ([]byte, error) {
	if err := validateNativeTypeEdition(nativeTypeEdition{Codec: nativeTypeEditionCodec, Type: pin}); err != nil {
		return nil, err
	}
	raw, err := encodeIssueRevisionContext(value)
	if err != nil {
		return nil, err
	}
	// Null here records the absence of private native attribution context, not
	// a null schema or a manufactured public properties/metadata value.
	if raw == nil {
		raw = json.RawMessage("null")
	}
	return canonicalJSON(nativeIssueTypeEdition{Codec: nativeTypeEditionCodec, Type: pin, Context: raw})
}

func decodeNativeIssueTypeEdition(raw []byte) (issueRevisionContext, resourceget.TypeReference, error) {
	var value nativeIssueTypeEdition
	if json.Unmarshal(raw, &value) != nil {
		return issueRevisionContext{}, resourceget.TypeReference{}, fmt.Errorf("%w: native Issue exact Type context", ErrInvalidStore)
	}
	if err := validateNativeTypeEdition(nativeTypeEdition{Codec: value.Codec, Type: value.Type}); err != nil {
		return issueRevisionContext{}, resourceget.TypeReference{}, err
	}
	contextBytes := []byte(value.Context)
	if bytes.Equal(contextBytes, []byte("null")) {
		contextBytes = nil
	}
	context, err := decodeIssueRevisionContext(contextBytes)
	if err != nil {
		return issueRevisionContext{}, resourceget.TypeReference{}, err
	}
	canonical, err := encodeNativeIssueTypeEdition(context, value.Type)
	if err != nil || !bytes.Equal(raw, canonical) {
		return issueRevisionContext{}, resourceget.TypeReference{}, fmt.Errorf("%w: native Issue exact context encoding", ErrInvalidStore)
	}
	return context, value.Type, nil
}
