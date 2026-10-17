package issueops

import "encoding/json"

// This file is the replay harness's only entry into this package, and it exists
// on the versioned-history integration branch alone: it is never part of an
// upstream change. It exports the version-history mint's own number gate so the
// harness asks the same question the mint asks, instead of carrying a second
// copy of the rule.
//
// Keep it to forwards. A function here calls an existing function and returns
// its result; it adds no logic and no second implementation. It binds only to
// canonicalDurableState, jcsCanonicalize and refuseUnrepresentableIntegers, and
// every package-level identifier declared here starts with Replay, so nothing
// added to this package elsewhere can collide with it.

// ReplayRefuseUnrepresentableIntegers runs the mint's number gate alone on an
// arbitrary JSON document. It returns nil if every number in doc is admitted and
// an error otherwise; an error that wraps ErrIntegerNotRepresentable is a number
// outside the exact range.
func ReplayRefuseUnrepresentableIntegers(doc []byte) error {
	return refuseUnrepresentableIntegers(doc)
}

// ReplayCanonicalDurableState runs everything the mint does to a snapshot before
// it stores it, the number gate and then RFC 8785 canonicalization, on an
// arbitrary JSON document. The document's number literals reach the gate as
// written, so it answers exactly as the mint would for an issue carrying them.
func ReplayCanonicalDurableState(doc []byte) ([]byte, error) {
	return canonicalDurableState(json.RawMessage(doc))
}
