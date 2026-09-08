package bdpwire

import (
	"regexp"
	"testing"
)

// THE DOMAIN RUNS AHEAD OF THE PINNED WIRE on the wildcard owned-Link
// declaration (bdp#1 item 5, ruled 2026-09-08): graphops admits the
// ownsOutgoing key "*" (graphops.WildcardOwnedLinkKey), while the bundle at
// Pin keys ownsOutgoing — and a record's ownedLinks — by absoluteHttpUrl, so
// a descriptor carrying "*" is schema-invalid on the wire until the pin moves
// to the bdp change that lands the ruling. This is the tripwire: it asserts
// the pinned facts, so the day the pin admits the key it fails and the
// run-ahead note in graphops retires with it.
//
// The refusal is the BUNDLE's, not the decoder's. The strict decoder holds
// member shape — names, types, nulls, integers — and not key grammar (it
// checks no ownedLinks key either), so it decodes the key; that is asserted
// too, so nobody moves the wire's key rule into the decoder, where the pin
// bump would then have to undo it. Test-only: no production change here.
func TestWildcardOwnedLinkKeyIsNotInThePinnedBundle(t *testing.T) {
	defs := loadBundleDefs(t)
	pattern := regexp.MustCompile(asString(t, asMap(t, defs["absoluteHttpUrl"], "absoluteHttpUrl")["pattern"], "absoluteHttpUrl.pattern"))
	if pattern.MatchString("*") {
		t.Fatalf("absoluteHttpUrl %s admits \"*\"", pattern)
	}
	for _, m := range []struct{ def, member string }{{"typeDescriptor", "ownsOutgoing"}, {"beadRecord", "ownedLinks"}} {
		props := asMap(t, asMap(t, defs[m.def], m.def)["properties"], m.def+".properties")
		names := asMap(t, asMap(t, props[m.member], m.member)["propertyNames"], m.member+".propertyNames")
		if ref := names["$ref"]; ref != "#/$defs/absoluteHttpUrl" {
			t.Fatalf("%s.%s.propertyNames = %v at pin %s: the pinned wire now admits a non-URL key — retire graphops's run-ahead note (WildcardOwnedLinkKey) and this tripwire together", m.def, m.member, names, Pin)
		}
	}
	var d TypeDescriptor
	if err := Unmarshal([]byte(`{"id":"`+typeURL+`","name":"X","describes":"bead","conformsTo":[],"ownsOutgoing":{"*":{"max":1}}}`), &d); err != nil {
		t.Fatalf("the strict decoder holds shape, not key grammar; the wire's key rule lives in the bundle, not here: %v", err)
	}
	if decl, ok := d.OwnsOutgoing["*"]; !ok || decl.Max != 1 {
		t.Fatalf("decoded ownsOutgoing = %+v", d.OwnsOutgoing)
	}
}
