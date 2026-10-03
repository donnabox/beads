package bdpwire

import (
	"encoding/json"
	"reflect"
	"testing"
)

var invalidHistoryZero = map[string]bool{"changeContext": true, "historyCapability": true, "historyVersionRow": true, "historyVersionsPage": true}

func schemaNullable(t *testing.T, schema map[string]any) bool {
	t.Helper()
	for _, keyword := range []string{"oneOf", "anyOf"} {
		if a, ok := schema[keyword]; ok {
			for _, x := range asSlice(t, a, keyword) {
				if asMap(t, x, keyword)["type"] == "null" {
					return true
				}
			}
		}
	}
	return false
}

// A bounded resolved view shared by parity and zero-value gates. It supports
// exact object refs and the historical Bead's false-property restriction only.
func resolvedObject(t *testing.T, defs map[string]any, name string) map[string]any {
	t.Helper()
	var resolve func(map[string]any, map[string]bool) map[string]any
	resolve = func(s map[string]any, seen map[string]bool) map[string]any {
		if n, ok := refName(t, s); ok {
			if len(s) != 1 || seen[n] {
				t.Fatal("unsupported/cyclic object ref")
			}
			seen[n] = true
			return resolve(asMap(t, defs[n], n), seen)
		}
		if s["type"] == "object" {
			return s
		}
		all := asSlice(t, s["allOf"], "object allOf")
		if len(all) != 2 {
			t.Fatal("unsupported object allOf")
		}
		base := resolve(asMap(t, all[0], "base"), seen)
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		props := map[string]any{}
		for k, v := range asMap(t, base["properties"], "properties") {
			props[k] = v
		}
		restriction := asMap(t, all[1], "restriction")
		if len(restriction) != 2 || restriction["type"] != "object" {
			t.Fatal("unsupported restriction")
		}
		for k, v := range asMap(t, restriction["properties"], "restriction properties") {
			if v != false {
				t.Fatal("unsupported nonfalse restriction")
			}
			if _, ok := props[k]; !ok {
				t.Fatal("restriction lacks base member")
			}
			for _, r := range asSlice(t, base["required"], "required") {
				if r == k {
					t.Fatal("cannot remove required member")
				}
			}
			delete(props, k)
		}
		out["properties"] = props
		return out
	}
	return resolve(asMap(t, defs[name], name), map[string]bool{name: true})
}
func TestProblemForbiddenSetsMatchConditionalSchema(t *testing.T) {
	rows := problemRows(t, loadBundleDefs(t))
	for code, row := range rows {
		want := map[string]bool{}
		for name, s := range row {
			if s == false {
				want[name] = true
			}
		}
		got := problemForbiddenMembers(ReadProblemCode(code))
		if !reflect.DeepEqual(want, got) {
			t.Errorf("%s forbidden got %v want %v", code, got, want)
		}
	}
}
func TestHistorySumSchemaArms(t *testing.T) {
	defs := loadBundleDefs(t)
	for _, x := range []struct {
		name           string
		time, nonempty bool
	}{{"contextTime", true, false}, {"contextString", false, true}, {"contextMessage", false, false}} {
		def := asMap(t, defs[x.name], x.name)
		arms := asSlice(t, def["oneOf"], x.name)
		if len(arms) != 2 {
			t.Fatal("context arm count")
		}
		for i, a := range arms {
			m := asMap(t, a, "arm")
			p := asMap(t, m["properties"], "properties")
			if m["additionalProperties"] != false {
				t.Fatal("open context arm")
			}
			required := []any{"state"}
			if i == 0 {
				required = append(required, "value")
			}
			if !reflect.DeepEqual(m["required"], required) || len(p) != len(required) {
				t.Fatal("context arm members/required drift")
			}
			state := asMap(t, p["state"], "state")
			if i == 0 {
				expected := string(ContextPresent)
				if x.time {
					expected = string(ContextTimePresent)
				}
				if state["const"] != expected {
					t.Fatal("present const")
				}
				value := asMap(t, p["value"], "value")
				if x.time {
					checkMember(t, defs, x.name+"/value", value, reflect.TypeOf(""))
				} else {
					if value["type"] != "string" {
						t.Fatal("context value kind")
					}
					if (value["minLength"] == json.Number("1")) != x.nonempty {
						t.Fatal("context empty-value rule")
					}
				}
			} else {
				want := []string{string(ContextAbsent), string(ContextUndetermined)}
				if x.time {
					want = []string{string(ContextTimeUndetermined)}
				}
				if !reflect.DeepEqual(stringSet(t, asSlice(t, state["enum"], "state enum")), sliceSet(want)) {
					t.Fatal("context state enum drift")
				}
			}
		}
	}
	arms := asSlice(t, asMap(t, defs["historyMissingItem"], "missing")["oneOf"], "missing arms")
	if len(arms) != 3 {
		t.Fatal("missing arm count")
	}
	for i, kind := range []HistoryMissingKind{MissingRecord, MissingProperty, MissingOwnedLinks} {
		m := asMap(t, arms[i], "arm")
		props := asMap(t, m["properties"], "props")
		req := []any{"kind"}
		keys := map[string]bool{"kind": true}
		if kind == MissingProperty {
			req = append(req, "pointer")
			keys["pointer"] = true
			checkMember(t, defs, "historyMissingItem/pointer", asMap(t, props["pointer"], "pointer"), reflect.TypeOf(""))
		}
		if kind == MissingOwnedLinks {
			keys["type"] = true
			checkMember(t, defs, "historyMissingItem/type", asMap(t, props["type"], "type"), reflect.TypeOf(""))
		}
		if m["additionalProperties"] != false || !reflect.DeepEqual(m["required"], req) || len(diff(props, keys)) != 0 || len(diff(keys, props)) != 0 || asMap(t, props["kind"], "kind")["const"] != string(kind) {
			t.Fatal("missing arm drift")
		}
	}
}

// Conditional required members are separate from top-level required fields and
// false-property exclusions. A re-pin must not silently move either boundary.
func problemRequiredRows(t *testing.T, defs map[string]any) map[string]map[string]bool {
	t.Helper()
	rows := map[string]map[string]bool{}
	problem := asMap(t, defs["readProblem"], "readProblem")
	for code := range readProblemTable {
		required := map[string]bool{}
		for _, item := range asSlice(t, problem["allOf"], "allOf") {
			clause := asMap(t, item, "clause")
			branch := "else"
			if problemCodeCondition(t, asMap(t, clause["if"], "if"), string(code)) {
				branch = "then"
			}
			if b, ok := clause[branch]; ok {
				if r, ok := asMap(t, b, branch)["required"]; ok {
					for name := range stringSet(t, asSlice(t, r, "required")) {
						required[name] = true
					}
				}
			}
		}
		rows[string(code)] = required
	}
	return rows
}
func TestProblemRequiredSetsMatchConditionalSchema(t *testing.T) {
	for code, required := range problemRequiredRows(t, loadBundleDefs(t)) {
		want := map[string]bool{}
		if ReadProblemCode(code) == CodeRevisionUnretained {
			want["missing"] = true
		}
		if !reflect.DeepEqual(required, want) {
			t.Errorf("%s conditionally required %v, wire guard requires %v", code, required, want)
		}
		// Bind the documented runtime requirement to the actual wire guard too.
		p := NewReadProblem(ReadProblemCode(code))
		if (p.validateWireConditions() != nil) != want["missing"] {
			t.Errorf("%s wire guard required payload drift", code)
		}
	}
}

func TestHistoricalLinkRetainsExactSchemaAndGoShape(t *testing.T) {
	defs := loadBundleDefs(t)
	if !reflect.DeepEqual(defs["historicalLinkRecord"], map[string]any{"$ref": "#/$defs/linkRecord"}) {
		t.Fatal("historical Link is no longer an exact schema alias")
	}
	history, ordinary := reflect.TypeOf(HistoricalLinkRecord{}), reflect.TypeOf(LinkRecord{})
	if history == ordinary || history.NumField() != ordinary.NumField() {
		t.Fatal("historical Link must be distinct with the ordinary Link shape")
	}
	for i := 0; i < ordinary.NumField(); i++ {
		if !reflect.DeepEqual(history.Field(i), ordinary.Field(i)) {
			t.Errorf("historical Link field %d drift", i)
		}
	}
	// Both directions are compile-time checks of the shared underlying type.
	_ = LinkRecord(HistoricalLinkRecord{})
	_ = HistoricalLinkRecord(LinkRecord{})
}
