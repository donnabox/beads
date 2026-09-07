package graphops_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/graphops"
)

// Every law in laws.go is table-tested here. A table names the sentence it
// pins where the sentence is not obvious from the vector.

func wantValidation(t *testing.T, err error, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: accepted, want ErrValidation", what)
	}
	if !errors.Is(err, graphops.ErrValidation) {
		t.Fatalf("%s: error %v does not wrap ErrValidation", what, err)
	}
}

// --- canonical-ID grammar -------------------------------------------------

func TestValidateCanonicalSegment(t *testing.T) {
	accept := []string{
		"task-42", "Task-42", "a.b_c~d", "!$&'()*+,;=:@", "0", "-",
		"caf%C3%A9",      // non-ASCII is percent-encoded, uppercase
		"a%20b",          // space
		"a%25b",          // a literal percent sign
		"%E2%9C%93",      // U+2713
		"x%3Fy", "x%23y", // "?" and "#" are legal once encoded
		"%7B%7D", "a%5Bb%5D", "a%7Cb", // {} [] | are outside the literal set
		"%C2%80", // U+0080 is not an ASCII control
	}
	for _, seg := range accept {
		if err := graphops.ValidateCanonicalSegment(seg); err != nil {
			t.Errorf("segment %q: refused: %v", seg, err)
		}
	}
	reject := []struct{ seg, why string }{
		{"", "empty"},
		{".", "dot segment"},
		{"..", "dot-dot segment"},
		{"%2E", "encoded dot segment"},
		{"%2E%2E", "encoded dot-dot segment"},
		{"a%2Fb", "encoded separator /"},
		{"a%5Cb", "encoded separator \\"},
		{"a/b", "literal separator"},
		{"a\\b", "literal backslash"},
		{"a%00b", "NUL"},
		{"a%1Fb", "C0 control"},
		{"a%7Fb", "DEL"},
		{"a b", "literal space"},
		{"a?b", "literal ? must be encoded"},
		{"a#b", "literal # must be encoded"},
		{"café", "literal non-ASCII must be encoded"},
		{"caf%c3%a9", "lowercase hex"},
		{"%41", "encoded unreserved character"},
		{"a%2Bb", "encoded literal-set character"},
		{"a%2", "incomplete escape"},
		{"a%zz", "malformed escape"},
		{"%", "bare percent"},
		{"%%41", "bare percent before an escape"},
		{"%C3", "truncated UTF-8"},
		{"%ED%A0%80", "UTF-8-encoded surrogate"},
		{"%C0%80", "overlong UTF-8"},
		{"%FF", "invalid UTF-8"},
		{"\x00", "raw control"},
	}
	for _, tc := range reject {
		wantValidation(t, graphops.ValidateCanonicalSegment(tc.seg), "segment "+tc.why+" "+tc.seg)
	}
}

func TestValidatePathGrammar(t *testing.T) {
	beads := []string{
		"beads/task-42", "beads/projects/alpha/tasks/task-42", "beads/Task-42",
		"beads/caf%C3%A9", "beads/beads", "beads/links/x", "beads/alias/x",
	}
	for _, p := range beads {
		if err := graphops.ValidateBeadPath(p); err != nil {
			t.Errorf("bead path %q refused: %v", p, err)
		}
		if err := graphops.ValidatePath(p, graphops.KindBead); err != nil {
			t.Errorf("ValidatePath(%q, bead) refused: %v", p, err)
		}
		wantValidation(t, graphops.ValidateLinkPath(p), "bead path as link "+p)
	}
	rejectBead := []string{
		"", "beads", "beads/", "/beads/x", "beads//x", "beads/x/", "Beads/x", "BEADS/x",
		"links/x", "alias/x", "beads/.", "beads/..", "beads/x?y", "beads/x#y",
		"beads/x/../y", "bead/x", "beadsx/y", " beads/x", "beads/x ", "beads\\x",
		"https://beads.example/acme/beads/x",
	}
	for _, p := range rejectBead {
		wantValidation(t, graphops.ValidateBeadPath(p), "bead path "+p)
		wantValidation(t, graphops.ValidatePath(p, graphops.KindBead), "ValidatePath bead "+p)
	}
	if err := graphops.ValidateLinkPath("links/assigned-to/81"); err != nil {
		t.Errorf("link path refused: %v", err)
	}
	if err := graphops.ValidatePath("links/assigned-to/81", graphops.KindLink); err != nil {
		t.Errorf("ValidatePath link refused: %v", err)
	}
	wantValidation(t, graphops.ValidateLinkPath("beads/x"), "link path under beads/")
	if err := graphops.ValidateAliasPath("alias/latest"); err != nil {
		t.Errorf("alias path refused: %v", err)
	}
	wantValidation(t, graphops.ValidateAliasPath("beads/x"), "alias path under beads/")
	wantValidation(t, graphops.ValidatePath("beads/x", graphops.ResourceKind("type")), "ValidatePath unknown kind")
}

// Case-differing paths are DISTINCT identities: both are canonical, neither
// is the other, and they order by code unit (uppercase first).
func TestCaseDifferingPathsAreDistinct(t *testing.T) {
	upper, lower := "beads/Task-42", "beads/task-42"
	for _, p := range []string{upper, lower} {
		if err := graphops.ValidateBeadPath(p); err != nil {
			t.Fatalf("%q refused: %v", p, err)
		}
	}
	if upper == lower || graphops.CompareCodeUnits(upper, lower) != -1 {
		t.Fatalf("case-differing paths must be distinct and code-unit ordered")
	}
}

func TestCanonicalSegmentEncoder(t *testing.T) {
	for _, tc := range []struct{ decoded, want string }{
		{"task-42", "task-42"},
		{"café", "caf%C3%A9"},
		{"a b", "a%20b"},
		{"a/b", "a%2Fb"},
		{"a+b", "a+b"},
		{"100%", "100%25"},
		{"[x]", "%5Bx%5D"},
		{"✓", "%E2%9C%93"},
		{"", ""},
	} {
		if got := graphops.CanonicalSegment(tc.decoded); got != tc.want {
			t.Errorf("CanonicalSegment(%q) = %q, want %q", tc.decoded, got, tc.want)
		}
	}
}

func TestCanonicalURLRoundTrip(t *testing.T) {
	scope := "https://beads.example/acme/"
	if got := graphops.CanonicalURL(scope, "beads/task-42"); got != "https://beads.example/acme/beads/task-42" {
		t.Fatalf("CanonicalURL = %q", got)
	}
	for _, tc := range []struct {
		url  string
		path string
		kind graphops.ResourceKind
		ok   bool
	}{
		{"https://beads.example/acme/beads/task-42", "beads/task-42", graphops.KindBead, true},
		{"https://beads.example/acme/links/l/1", "links/l/1", graphops.KindLink, true},
		{"https://beads.example/acme/alias/x", "", "", false},
		{"https://beads.example/acme/beads/", "", "", false},
		{"https://beads.example/acme/", "", "", false},
		{"https://beads.example/other/beads/x", "", "", false},
		{"https://beads.example/acme/beads/x?y", "", "", false},
		{"https://beads.example/acme/beads/caf%c3%a9", "", "", false},
		{"", "", "", false},
	} {
		path, kind, ok := graphops.SplitCanonicalURL(scope, tc.url)
		if path != tc.path || kind != tc.kind || ok != tc.ok {
			t.Errorf("SplitCanonicalURL(%q) = (%q, %q, %v), want (%q, %q, %v)", tc.url, path, kind, ok, tc.path, tc.kind, tc.ok)
		}
	}
}

// --- code-unit ordering ---------------------------------------------------

func TestCompareCodeUnits(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"a", "b", -1}, {"b", "a", 1}, {"a", "a", 0}, {"a", "ab", -1}, {"ab", "a", 1},
		{"", "", 0}, {"", "a", -1}, {"a", "", 1},
		{"beads/Task", "beads/task", -1},
		{"beads/task-10", "beads/task-9", -1}, // lexicographic, not numeric
		{"z", "\u00e9", -1}, {"\u00e9", "\u20ac", -1},
		{"\U0001F600", "\U0001F601", -1}, {"\U0001F600", "\U0001F600", 0},
		// The cases where code units and bytes disagree: a supplementary
		// character's lead surrogate (D800–DBFF) sorts BELOW U+E000–U+FFFF.
		{"\uFFFF", "\U0001F600", 1},
		{"\uE000", "\U00010000", 1},
		{"\uD7FF", "\U00010000", -1},
		{"x\uFFFF", "x\U0001F600", 1},
	} {
		if got := graphops.CompareCodeUnits(tc.a, tc.b); got != tc.want {
			t.Errorf("CompareCodeUnits(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
	// And the divergence from byte order, stated explicitly.
	if bytes.Compare([]byte("\uFFFF"), []byte("\U0001F600")) != -1 {
		t.Fatal("test premise: UTF-8 bytes order U+FFFF before U+1F600")
	}
	if graphops.CompareCodeUnits("\uFFFF", "\U0001F600") != 1 {
		t.Fatal("code units order U+1F600 before U+FFFF")
	}
}

// --- JSON canonicalization ------------------------------------------------

func TestCanonicalizeJSONNumbers(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"1.0", "1"}, {"-0.0", "0"}, {"-0", "0"}, {"0.0", "0"}, {"0e10", "0"}, {"0", "0"},
		{"1e300", "1e+300"}, {"1E300", "1e+300"},
		{"9007199254740993", "9007199254740993"}, // 2^53+1, exact
		{"9007199254740992", "9007199254740992"},
		{"18446744073709551616", "18446744073709551616"}, // 2^64, exact
		{"1e21", "1e+21"}, {"1e20", "100000000000000000000"},
		{"123456789012345678901", "123456789012345678901"},
		{"1234567890123456789012", "1.234567890123456789012e+21"},
		{"0.000001", "0.000001"}, {"1e-6", "0.000001"},
		{"0.0000001", "1e-7"}, {"1e-7", "1e-7"}, {"-1e-7", "-1e-7"},
		{"123.456e2", "12345.6"}, {"-1.5", "-1.5"}, {"1.50", "1.5"},
		{"100", "100"}, {"1E+2", "100"}, {"1e-1", "0.1"}, {"0.1", "0.1"},
		{"10.0e-1", "1"}, {"25e-1", "2.5"}, {"123e-2", "1.23"}, {"1e0", "1"},
		{"1.5e1", "15"}, {"9.99e2", "999"}, {"1.23e+3", "1230"}, {"0.5", "0.5"}, {"-0.5", "-0.5"},
		{"1000000", "1000000"}, {"5e-324", "5e-324"},
		{"1.7976931348623157e308", "1.7976931348623157e+308"},
		{"-1e400", "-1e+400"}, // beyond double range: exact, where JCS could not serialize
		{"0.1000000000000000055511151231257827021181583404541015625", "0.1000000000000000055511151231257827021181583404541015625"},
		{"333333333.33333329", "333333333.33333329"}, // JCS's appendix rounds this to 333333333.3333333
	} {
		got, err := graphops.CanonicalizeJSON([]byte(tc.in))
		if err != nil {
			t.Errorf("number %q: %v", tc.in, err)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("number %q canonicalized to %q, want %q", tc.in, got, tc.want)
		}
	}
	for _, in := range []string{
		"01", "1.", ".5", "+1", "1e", "1e+", "-", "0x10", "NaN", "Infinity", "1_000", "--1", "1.e5", "-a",
		"1e1000000000000000", // exponent out of range
	} {
		_, err := graphops.CanonicalizeJSON([]byte(in))
		wantValidation(t, err, "number "+in)
	}
}

func TestCanonicalizeJSONStrings(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{`"\u0041"`, `"A"`},
		{`"\u00e9"`, "\"\u00e9\""},
		{`"\u00C9"`, "\"\u00c9\""},
		{`"\ud83d\ude00"`, "\"\U0001F600\""},
		{`"\uD83D\uDE00"`, "\"\U0001F600\""},
		{`"\u2028"`, "\"\u2028\""}, // literal, as JSON.stringify emits it
		{`"\/"`, `"/"`},
		{`"\u001f"`, `"\u001f"`},
		{`"\u000B"`, `"\u000b"`}, // lowercase hex
		{`"\u0000"`, `"\u0000"`},
		{`"\u007f"`, "\"\x7f\""}, // DEL is literal
		{`"\b\f\n\r\t"`, `"\b\f\n\r\t"`},
		{`"\u0008\u000c\u000a\u000d\u0009"`, `"\b\f\n\r\t"`},
		{`"<>&"`, `"<>&"`},
		{`"\""`, `"\""`},
		{`"\\"`, `"\\"`},
		{"\"\u00e9\"", "\"\u00e9\""},
		{`""`, `""`},
	} {
		got, err := graphops.CanonicalizeJSON([]byte(tc.in))
		if err != nil {
			t.Errorf("string %s: %v", tc.in, err)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("string %s canonicalized to %q, want %q", tc.in, got, tc.want)
		}
	}
	for _, in := range []string{
		`"\ud800"`, `"\udc00"`, `"\ud800x"`, `"\ud800\u0041"`, `"\ud800\ud800"`, `"\ud83d\ude0"`,
		"\"a\x01b\"", `"\x"`, `"abc`, "\"\xff\"", "\"\xed\xa0\x80\"", `"\u12"`, `"\u12G4"`, `"\`, `"\u`,
		"\"\xc3\\n\"", // invalid UTF-8 before an escape
	} {
		_, err := graphops.CanonicalizeJSON([]byte(in))
		wantValidation(t, err, "string "+in)
	}
}

func TestCanonicalizeJSONStructure(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{`{"b":1,"a":2}`, `{"a":2,"b":1}`},
		{" \t\n\r{ \"a\" : [ 1 , 2 ] , \"b\" : { } } \n", `{"a":[1,2],"b":{}}`},
		{`[]`, `[]`}, {`[ ]`, `[]`}, {`{}`, `{}`}, {`{ }`, `{}`},
		{`{"a":{"c":1,"b":2}}`, `{"a":{"b":2,"c":1}}`},
		{`[1,[2,[3]]]`, `[1,[2,[3]]]`},
		{`{"":1}`, `{"":1}`},
		{`true`, `true`}, {` false `, `false`}, {`null`, `null`}, {`"x"`, `"x"`}, {`1`, `1`},
		{`{"10":1,"9":2,"1":3}`, `{"1":3,"10":1,"9":2}`}, // lexicographic keys
		// Keys sort by UTF-16 code unit: the emoji's lead surrogate sorts
		// before U+FFFF, where byte order would put it after.
		{`{"\uffff":1,"\ud83d\ude00":2}`, "{\"\U0001F600\":2,\"\uFFFF\":1}"},
		// RFC 8785 §3.2.3's sorting example.
		{
			`{"\u20ac":"Euro Sign","\r":"Carriage Return","\ufb33":"Hebrew Letter Dalet With Dagesh","1":"One","\ud83d\ude02":"Emoji: Face with Tears of Joy","\u0080":"Control","\u00f6":"Latin Small Letter O With Diaeresis"}`,
			"{\"\\r\":\"Carriage Return\",\"1\":\"One\",\"\u0080\":\"Control\",\"\u00f6\":\"Latin Small Letter O With Diaeresis\",\"\u20ac\":\"Euro Sign\",\"\U0001F602\":\"Emoji: Face with Tears of Joy\",\"\uFB33\":\"Hebrew Letter Dalet With Dagesh\"}",
		},
		// RFC 8785 Appendix's example, under the exact-number rule: identical
		// to the appendix's output except that 333333333.33333329 keeps its
		// digits where JCS rounds it through a double.
		{
			`{"numbers": [333333333.33333329, 1E30, 4.50, 2e-3, 0.000000000000000000000000001], "string": "\u20ac$\u000F\u000aA'\u0042\u0022\u005c\\\"\/", "literals": [null, true, false]}`,
			"{\"literals\":[null,true,false],\"numbers\":[333333333.33333329,1e+30,4.5,0.002,1e-27],\"string\":\"\u20ac$\\u000f\\nA'B\\\"\\\\\\\\\\\"/\"}",
		},
	} {
		got, err := graphops.CanonicalizeJSON([]byte(tc.in))
		if err != nil {
			t.Errorf("%s: %v", tc.in, err)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("%s canonicalized to\n  %q\nwant\n  %q", tc.in, got, tc.want)
		}
	}
	// Canonical form is a fixed point.
	in := `{"z":[1.0,{"b":"\u0041","a":null}],"a":-0.0}`
	once, err := graphops.CanonicalizeJSON([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	twice, err := graphops.CanonicalizeJSON(once)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(once, twice) {
		t.Fatalf("canonical form is not a fixed point: %q then %q", once, twice)
	}
}

func TestCanonicalizeJSONRejects(t *testing.T) {
	for _, in := range []string{
		"", " ", "{}x", "[1,]", `{"a":1,}`, `{a:1}`, "[1 2]", "tru", "nul", "True", "[", "{", `{"a"}`, `{"a":}`,
		"[,1]", "{,}", "1 2", `"a" "b"`, `{"a":1 "b":2}`, "[1;2]", "]", "}", `{"a":1]`, "[1}", "\x00", "{\"a\":1", "{\"a\":1,", "[1", "[1,", "{\"a\":1,\"b\"", "{\"ab", "{\"\x01\":1}",
		`{"a":1,"a":2}`, `{"a":1,"b":2,"a":3}`, `{"x":{"a":1,"a":1}}`,
		`{"a":1,"\u0061":2}`, // duplicate after decoding
		"{\"a\":1}\x00",
	} {
		_, err := graphops.CanonicalizeJSON([]byte(in))
		wantValidation(t, err, "input "+in)
	}
}

func TestCanonicalizeJSONDepth(t *testing.T) {
	ok := strings.Repeat("[", 10000) + strings.Repeat("]", 10000)
	if _, err := graphops.CanonicalizeJSON([]byte(ok)); err != nil {
		t.Fatalf("depth 10000 refused: %v", err)
	}
	tooDeep := strings.Repeat("[", 10001) + strings.Repeat("]", 10001)
	_, err := graphops.CanonicalizeJSON([]byte(tooDeep))
	wantValidation(t, err, "depth 10001 arrays")
	tooDeepObjects := strings.Repeat(`{"a":`, 10001) + "1" + strings.Repeat("}", 10001)
	_, err = graphops.CanonicalizeJSON([]byte(tooDeepObjects))
	wantValidation(t, err, "depth 10001 objects")
}

func TestJSONEqual(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"1", "1.0", true}, {"1", `"1"`, false}, {"1e2", "100", true}, {"0.1", "0.10", true}, {"-0", "0", true},
		{"9007199254740993", "9007199254740992", false}, {"1", "1.00000000000000001", false},
		{`{"a":1,"b":2}`, `{"b":2,"a":1}`, true}, {`{"a":1}`, `{"a":1,"b":2}`, false},
		{"[1,2]", "[2,1]", false}, {"[1,2]", "[1,2]", true}, {"[]", "{}", false},
		{"null", "null", true}, {`{"a":null}`, `{}`, false}, {"true", "false", false},
		{`"\u0041"`, `"A"`, true}, {`"a"`, `"A"`, false},
		{`[1,{"a":[2,3]}]`, ` [ 1 , { "a" : [ 2 , 3 ] } ] `, true},
	} {
		got, err := graphops.JSONEqual([]byte(tc.a), []byte(tc.b))
		if err != nil {
			t.Errorf("JSONEqual(%s, %s): %v", tc.a, tc.b, err)
			continue
		}
		if got != tc.want {
			t.Errorf("JSONEqual(%s, %s) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
	if _, err := graphops.JSONEqual([]byte("{"), []byte("1")); err == nil {
		t.Error("JSONEqual accepted a malformed left operand")
	}
	if _, err := graphops.JSONEqual([]byte("1"), []byte("}")); err == nil {
		t.Error("JSONEqual accepted a malformed right operand")
	}
}

// --- Scope URL and Type URL -----------------------------------------------

func TestValidateScopeURL(t *testing.T) {
	accept := []string{
		"https://beads.example/acme/", "https://beads.example/", "http://localhost:3000/",
		"https://beads.example:8443/acme/", "https://beads.example/caf%C3%A9/", "https://[::1]/",
		"https://[2001:db8::1]:8443/", "https://127.0.0.1/", "https://beads.example/a/b/c/",
		"https://beads.example/local-testing/", "https://beads.example/x/local-test/",
		"https://beads.example/Acme/", "https://my_host/", "http://beads.example:8080/x/",
		"https://beads.example:0/",
	}
	for _, u := range accept {
		if err := graphops.ValidateScopeURL(u); err != nil {
			t.Errorf("Scope URL %q refused: %v", u, err)
		}
	}
	reject := []string{
		"", "https://beads.example/acme", "HTTPS://beads.example/acme/", "Https://beads.example/",
		"https://Beads.Example/acme/", "https://beads.example:443/acme/", "http://beads.example:80/",
		"https://beads.example/acme/?x=1", "https://beads.example/acme/?", "https://beads.example/acme/#f",
		"https://beads.example/acme/#", "https://user:pw@beads.example/acme/", "https://@beads.example/",
		"https://beads.example/a%2Fb/", "https://beads.example/a b/", "https://beads.example/caf%c3%a9/",
		"https://beads.example/%41/", "https://beads.example/./", "https://beads.example/../",
		"https://beads.example/a/./b/", "https://beads.example//x/", "ftp://x/", "https:///acme/",
		"https://[0:0:0:0:0:0:0:1]/", "https://[::ffff:1.2.3.4]/", "https://[::1", "https://[::1]x/",
		"https://[fe80::1%25eth0]/", "https://beads.example/local-test/", "https://beads.example/a\\b/",
		"https://beads.example:/", "https://beads.example:0080/", "https://beads.example:99999/",
		"https://beads.example:abc/", "https://beads.example./", "https://.beads.example/",
		"https://127.1/", "https://01.2.3.4/", "https://1.2.3.4.5/", "https://beads.example/café/",
		"beads.example/acme/", "//beads.example/acme/", "https://beads.example/a?b/",
		"https://beads.example/%/", "https://beads.example/a%2/", "https://beads.example/a[b]/",
		"https://beads.example/\t/", "https://beads example/", "https://beads.example/a|b/",
		"https://beads.example/a{b}/", "https://beads.example/\x7f/", "https:beads.example/",
		"https:/beads.example/", "https://BEADS.example/", "https://beads.example/%E2%9C%93/x%2e/",
		"https://beads.example/a%5Cb/", "https://beads.example/a%00b/",
	}
	for _, u := range reject {
		wantValidation(t, graphops.ValidateScopeURL(u), "Scope URL "+u)
	}
	// Diagnostics never echo the value (a pasted token would land in a log).
	err := graphops.ValidateScopeURL("https://secret-token-value@beads.example/")
	if err == nil || strings.Contains(err.Error(), "secret-token-value") {
		t.Fatalf("Scope URL diagnostic echoed the value: %v", err)
	}
}

func TestNormalizeScopeURL(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"https://EXAMPLE.com:443/scope/", "https://example.com/scope/"},
		{"http://example.com:80/scope/", "http://example.com/scope/"},
		{"https://example.com/scope", "https://example.com/scope/"},
		{"https://example.com", "https://example.com/"},
		{"HTTP://Example.COM", "http://example.com/"},
		{"https://[0:0:0:0:0:0:0:1]:8443/x", "https://[::1]:8443/x/"},
		{"https://example.com:", "https://example.com/"},
		{"https://example.com:8080", "https://example.com:8080/"},
		{"https://example.com/caf%C3%A9", "https://example.com/caf%C3%A9/"},
	} {
		got, err := graphops.NormalizeScopeURL(tc.in)
		if err != nil {
			t.Errorf("NormalizeScopeURL(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("NormalizeScopeURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
		if err := graphops.ValidateScopeURL(got); err != nil {
			t.Errorf("normalized %q does not validate: %v", got, err)
		}
	}
	for _, in := range []string{
		"https://u:p@example.com/", "https://example.com/?q", "https://example.com/#f",
		"https://example.com/a/../b/", "ftp://example.com/", "example.com/", "", "://x/",
		"https://example.com/local-test/", "https://example.com/a b/", "https://example.com/a%2fb/",
		"https://example.com/a[b]", "https:example.com", "https://EXAMPLE.com/%41/",
		"https://[::1", "https://[::1]:x/", "ht tp://example.com/",
	} {
		_, err := graphops.NormalizeScopeURL(in)
		wantValidation(t, err, "NormalizeScopeURL "+in)
	}
}

func TestValidateTypeURL(t *testing.T) {
	accept := []string{
		"https://work.example/types/task", "https://work.example/types/task?v=1", "http://localhost:8080/t",
		"https://work.example/", "https://work.example/a[b]", "https://work.example/a|b",
		"https://work.example/x//y", "https://work.example/schemas/task-properties-v1",
		"https://work.example/t?a=b&c=d", "https://work.example/t?a=%20", "https://work.example/caf%C3%A9",
		"https://[::1]:8443/t", "https://work.example/t?", "https://work.example/t?x=[1]",
		"https://work.example/~user/T.Y_P-E",
	}
	for _, u := range accept {
		if err := graphops.ValidateTypeURL(u); err != nil {
			t.Errorf("Type URL %q refused: %v", u, err)
		}
	}
	reject := []string{
		"", "https://work.example", "HTTPS://work.example/", "https://Work.example/", "https://work.example:443/",
		"https://work.example/#x", "https://work.example/#", "https://u@work.example/", "https://work.example/a b",
		"https://work.example/a%2fb", "https://work.example/%41", "https://work.example/./x",
		"https://work.example/x/../y", "https://work.example/a\\b", "https://work.example/é",
		"https://work.example/a%", "https://work.example/a%4", "https://work.example/?a b",
		"https://work.example/x?a='b'", "https://work.example/x?a=\"b\"", "https://work.example/x?a=<b>",
		"https://work.example/{x}", "https://work.example/x`", "https://work.example/x?q=%zz",
		"work.example/types/task", "beads/x", "https://work.example/\x7f", "https://work.example/a\"b",
		"https://work.example/<x>", "mailto:a@b", "https://work.example:x/", "https://work.example:/",
		"https://:8080/", "https://work.example/x\n", "1http://x/",
	}
	for _, u := range reject {
		wantValidation(t, graphops.ValidateTypeURL(u), "Type URL "+u)
	}
	// Type URL diagnostics DO quote the value: it is public catalog data and
	// the operator needs to see which descriptor is wrong.
	err := graphops.ValidateTypeURL("https://Work.example/x")
	if err == nil || !strings.Contains(err.Error(), "https://Work.example/x") {
		t.Fatalf("Type URL diagnostic should quote the value: %v", err)
	}
}

// --- ledger hashing -------------------------------------------------------

const (
	opID        = "0123456789abcdef0123456789abcdef"
	authorityID = "fedcba9876543210fedcba9876543210"
	scopeURL    = "https://beads.example/acme/"
)

var at = time.Date(2026, 9, 7, 12, 34, 56, 123456789, time.UTC)

func mustEvent(t *testing.T, spec graphops.LedgerEventSpec) graphops.LedgerEvent {
	t.Helper()
	e, err := graphops.NewLedgerEvent(spec)
	if err != nil {
		t.Fatalf("NewLedgerEvent(seq %d %s): %v", spec.Seq, spec.Kind, err)
	}
	return e
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// The byte layout is FROZEN: the P1 migration stores these hashes. A change
// to this test is a change to every stored ledger.
func TestLedgerHashLayoutIsFrozen(t *testing.T) {
	mint := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 1, Kind: graphops.LedgerMint, OpID: opID, ScopeURL: scopeURL,
		AuthorityID: authorityID, Epoch: 1, At: at, PrevHash: graphops.GenesisHash,
	})
	wantMint := `{"at":"2026-09-07T12:34:56.123456Z","authority_id":"fedcba9876543210fedcba9876543210","epoch":1,"kind":"mint","op_id":"0123456789abcdef0123456789abcdef","prev_hash":"0000000000000000000000000000000000000000000000000000000000000000","scope_url":"https://beads.example/acme/","seq":1}`
	if got := string(mint.CanonicalBytes()); got != wantMint {
		t.Fatalf("mint canonical bytes\n got %s\nwant %s", got, wantMint)
	}
	if mint.Hash() != sha256Hex([]byte(wantMint)) {
		t.Fatalf("mint hash %s is not sha256 of the canonical bytes", mint.Hash())
	}
	if mint.At() != at.Truncate(time.Microsecond) {
		t.Fatalf("At() = %v, want microsecond-truncated %v", mint.At(), at.Truncate(time.Microsecond))
	}

	rev, _ := graphops.NewRevision("abc")
	tomb := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 4, Kind: graphops.LedgerTombstone, OpID: opID, Path: "beads/x", ResourceKind: graphops.KindBead,
		Revision: rev, State: graphops.AllocationPruned, AuthorityID: authorityID, Epoch: math.MaxUint64,
		At: time.Date(2026, 9, 7, 14, 34, 56, 0, time.FixedZone("plus2", 2*3600)), PrevHash: mint.Hash(),
	})
	wantTomb := `{"at":"2026-09-07T12:34:56.000000Z","authority_id":"fedcba9876543210fedcba9876543210","epoch":18446744073709551615,"kind":"tombstone","op_id":"0123456789abcdef0123456789abcdef","path":"beads/x","prev_hash":"` + mint.Hash() + `","resource_kind":"bead","revision":"abc","seq":4,"state":"pruned"}`
	if got := string(tomb.CanonicalBytes()); got != wantTomb {
		t.Fatalf("tombstone canonical bytes\n got %s\nwant %s", got, wantTomb)
	}
	if tomb.Hash() != sha256Hex([]byte(wantTomb)) {
		t.Fatal("tombstone hash is not sha256 of the canonical bytes")
	}

	fp := strings.Repeat("ab", 32)
	install := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 2, Kind: graphops.LedgerInstall, OpID: opID, Fingerprint: fp,
		AuthorityID: authorityID, Epoch: 1, At: at, PrevHash: mint.Hash(),
	})
	wantInstall := `{"at":"2026-09-07T12:34:56.123456Z","authority_id":"fedcba9876543210fedcba9876543210","epoch":1,"fingerprint":"` + fp + `","kind":"install","op_id":"0123456789abcdef0123456789abcdef","prev_hash":"` + mint.Hash() + `","seq":2}`
	if got := string(install.CanonicalBytes()); got != wantInstall {
		t.Fatalf("install canonical bytes\n got %s\nwant %s", got, wantInstall)
	}
	// Accessors carry the members through.
	spec := install.Spec()
	if spec.Fingerprint != fp || spec.Hash != install.Hash() || install.Fingerprint() != fp || install.OpID() != opID ||
		install.AuthorityID() != authorityID || install.Epoch() != 1 || install.PrevHash() != mint.Hash() ||
		install.Kind() != graphops.LedgerInstall || install.Seq() != 2 || install.Path() != "" || install.ScopeURL() != "" ||
		install.ResourceKind() != "" || !install.Revision().IsZero() || install.State() != "" || install.IsZero() {
		t.Fatalf("install accessors disagree with the spec: %+v", spec)
	}
	if tomb.Path() != "beads/x" || tomb.ResourceKind() != graphops.KindBead || tomb.State() != graphops.AllocationPruned ||
		!tomb.Revision().Equal(rev) || mint.ScopeURL() != scopeURL {
		t.Fatal("tombstone or mint accessors disagree with the spec")
	}
	if (graphops.LedgerEvent{}).IsZero() == false {
		t.Fatal("zero LedgerEvent must report IsZero")
	}
}

func chain(t *testing.T) []graphops.LedgerEvent {
	t.Helper()
	mint := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 1, Kind: graphops.LedgerMint, OpID: opID, ScopeURL: scopeURL,
		AuthorityID: authorityID, Epoch: 1, At: at, PrevHash: graphops.GenesisHash,
	})
	install := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 2, Kind: graphops.LedgerInstall, OpID: opID, Fingerprint: strings.Repeat("ab", 32),
		AuthorityID: authorityID, Epoch: 1, At: at, PrevHash: mint.Hash(),
	})
	alloc := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 3, Kind: graphops.LedgerAllocate, OpID: graphops.MintOpaqueToken(), Path: "links/l", ResourceKind: graphops.KindLink,
		Revision: graphops.MintRevision(), AuthorityID: authorityID, Epoch: 1, At: at.Add(time.Second), PrevHash: install.Hash(),
	})
	tomb := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 4, Kind: graphops.LedgerTombstone, OpID: graphops.MintOpaqueToken(), Path: "links/l", ResourceKind: graphops.KindLink,
		State: graphops.AllocationErased, AuthorityID: authorityID, Epoch: 1, At: at.Add(2 * time.Second), PrevHash: alloc.Hash(),
	})
	return []graphops.LedgerEvent{mint, install, alloc, tomb}
}

func TestLedgerHashIsDeterministicAndTamperEvident(t *testing.T) {
	spec := graphops.LedgerEventSpec{
		Seq: 1, Kind: graphops.LedgerMint, OpID: opID, ScopeURL: scopeURL,
		AuthorityID: authorityID, Epoch: 1, At: at, PrevHash: graphops.GenesisHash,
	}
	a, b := mustEvent(t, spec), mustEvent(t, spec)
	if a.Hash() != b.Hash() || !bytes.Equal(a.CanonicalBytes(), b.CanonicalBytes()) {
		t.Fatal("the same spec must hash the same")
	}
	// Restoring with the stored hash verifies it …
	stored := a.Spec()
	if _, err := graphops.NewLedgerEvent(stored); err != nil {
		t.Fatalf("restoring an untouched event: %v", err)
	}
	// … and every tampered member is caught, including ones beyond the
	// microsecond that canonicalization drops.
	for name, mutate := range map[string]func(*graphops.LedgerEventSpec){
		"at":           func(s *graphops.LedgerEventSpec) { s.At = s.At.Add(time.Microsecond) },
		"epoch":        func(s *graphops.LedgerEventSpec) { s.Epoch++ },
		"seq":          func(s *graphops.LedgerEventSpec) { s.Seq++ },
		"scope_url":    func(s *graphops.LedgerEventSpec) { s.ScopeURL = "https://beads.example/other/" },
		"authority_id": func(s *graphops.LedgerEventSpec) { s.AuthorityID = strings.Repeat("0", 32) },
		"op_id":        func(s *graphops.LedgerEventSpec) { s.OpID = strings.Repeat("1", 32) },
		"prev_hash":    func(s *graphops.LedgerEventSpec) { s.PrevHash = strings.Repeat("1", 64) },
	} {
		tampered := stored
		mutate(&tampered)
		_, err := graphops.NewLedgerEvent(tampered)
		wantValidation(t, err, "tampered "+name)
	}
	// A sub-microsecond change is NOT a change: it never reached the bytes.
	sub := stored
	sub.At = sub.At.Add(500 * time.Nanosecond)
	if _, err := graphops.NewLedgerEvent(sub); err != nil {
		t.Fatalf("sub-microsecond instant should verify: %v", err)
	}
}

func TestVerifyLedgerChain(t *testing.T) {
	events := chain(t)
	if err := graphops.VerifyLedgerChain(graphops.GenesisHash, events); err != nil {
		t.Fatalf("intact chain refused: %v", err)
	}
	if err := graphops.VerifyLedgerChain(events[0].Hash(), events[1:]); err != nil {
		t.Fatalf("suffix chain refused: %v", err)
	}
	if err := graphops.VerifyLedgerChain(graphops.GenesisHash, nil); err != nil {
		t.Fatalf("empty chain refused: %v", err)
	}
	wantValidation(t, graphops.VerifyLedgerChain(graphops.GenesisHash, events[1:]), "chain not starting at genesis")
	wantValidation(t, graphops.VerifyLedgerChain(graphops.GenesisHash, []graphops.LedgerEvent{events[1], events[0]}), "reordered chain")
	wantValidation(t, graphops.VerifyLedgerChain(graphops.GenesisHash, []graphops.LedgerEvent{events[0], events[2]}), "missing link")
	wantValidation(t, graphops.VerifyLedgerChain("nope", events), "bad starting hash")
	wantValidation(t, graphops.VerifyLedgerChain(graphops.GenesisHash, []graphops.LedgerEvent{{}}), "zero event")
	// A gap with intact linkage: seq 5 linked to seq 2's hash.
	skip := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 5, Kind: graphops.LedgerPromote, OpID: opID, AuthorityID: authorityID, Epoch: 2, At: at, PrevHash: events[1].Hash(),
	})
	err := graphops.VerifyLedgerChain(graphops.GenesisHash, []graphops.LedgerEvent{events[0], events[1], skip})
	wantValidation(t, err, "gap")
	if !strings.Contains(err.Error(), "gap") {
		t.Fatalf("gap error should say so: %v", err)
	}
}

func TestLedgerEventShapePerKind(t *testing.T) {
	rev := graphops.MintRevision()
	base := func(kind graphops.LedgerEventKind) graphops.LedgerEventSpec {
		return graphops.LedgerEventSpec{Seq: 1, Kind: kind, OpID: opID, AuthorityID: authorityID, Epoch: 1, At: at, PrevHash: graphops.GenesisHash}
	}
	with := func(kind graphops.LedgerEventKind, f func(*graphops.LedgerEventSpec)) graphops.LedgerEventSpec {
		s := base(kind)
		f(&s)
		return s
	}
	fp := strings.Repeat("cd", 32)
	for _, tc := range []struct {
		name string
		spec graphops.LedgerEventSpec
		ok   bool
		msg  string
	}{
		{"mint ok", with(graphops.LedgerMint, func(s *graphops.LedgerEventSpec) { s.ScopeURL = scopeURL }), true, ""},
		{"mint without scope_url", base(graphops.LedgerMint), false, "scope_url is required"},
		{"mint with path", with(graphops.LedgerMint, func(s *graphops.LedgerEventSpec) { s.ScopeURL = scopeURL; s.Path = "beads/x" }), false, "path is not a member"},
		{"mint bad scope_url", with(graphops.LedgerMint, func(s *graphops.LedgerEventSpec) { s.ScopeURL = "https://beads.example/acme" }), false, "scope_url:"},
		{"install ok", with(graphops.LedgerInstall, func(s *graphops.LedgerEventSpec) { s.Fingerprint = fp }), true, ""},
		{"install without fingerprint", base(graphops.LedgerInstall), false, "fingerprint is required"},
		{"install bad fingerprint", with(graphops.LedgerInstall, func(s *graphops.LedgerEventSpec) { s.Fingerprint = "xyz" }), false, "fingerprint must be"},
		{"install with revision", with(graphops.LedgerInstall, func(s *graphops.LedgerEventSpec) { s.Fingerprint = fp; s.Revision = rev }), false, "revision is not a member"},
		{"update ok", with(graphops.LedgerUpdate, func(s *graphops.LedgerEventSpec) {
			s.Path = "beads/x"
			s.ResourceKind = graphops.KindBead
			s.Revision = rev
		}), true, ""},
		{"update without revision", with(graphops.LedgerUpdate, func(s *graphops.LedgerEventSpec) { s.Path = "beads/x"; s.ResourceKind = graphops.KindBead }), false, "revision is required"},
		{"update with state", with(graphops.LedgerUpdate, func(s *graphops.LedgerEventSpec) {
			s.Path = "beads/x"
			s.ResourceKind = graphops.KindBead
			s.Revision = rev
			s.State = graphops.AllocationPruned
		}), false, "state is not a member"},
		{"promote ok", base(graphops.LedgerPromote), true, ""},
		{"promote with scope_url", with(graphops.LedgerPromote, func(s *graphops.LedgerEventSpec) { s.ScopeURL = scopeURL }), false, "scope_url is not a member"},
		{"rotate ok", with(graphops.LedgerRotate, func(s *graphops.LedgerEventSpec) { s.ScopeURL = scopeURL }), true, ""},
		{"rotate without scope_url", base(graphops.LedgerRotate), false, "scope_url is required"},
		{"refuse_url ok", with(graphops.LedgerRefuseURL, func(s *graphops.LedgerEventSpec) { s.ScopeURL = scopeURL }), true, ""},
		{"refuse_url with fingerprint", with(graphops.LedgerRefuseURL, func(s *graphops.LedgerEventSpec) { s.ScopeURL = scopeURL; s.Fingerprint = fp }), false, "fingerprint is not a member"},
		{"allocate ok without revision", with(graphops.LedgerAllocate, func(s *graphops.LedgerEventSpec) { s.Path = "links/l"; s.ResourceKind = graphops.KindLink }), true, ""},
		{"allocate ok with revision", with(graphops.LedgerAllocate, func(s *graphops.LedgerEventSpec) {
			s.Path = "links/l"
			s.ResourceKind = graphops.KindLink
			s.Revision = rev
		}), true, ""},
		{"allocate without resource_kind", with(graphops.LedgerAllocate, func(s *graphops.LedgerEventSpec) { s.Path = "links/l" }), false, "resource_kind is required"},
		{"allocate with state", with(graphops.LedgerAllocate, func(s *graphops.LedgerEventSpec) {
			s.Path = "links/l"
			s.ResourceKind = graphops.KindLink
			s.State = graphops.AllocationErased
		}), false, "state is not a member"},
		{"allocate bad kind", with(graphops.LedgerAllocate, func(s *graphops.LedgerEventSpec) { s.Path = "links/l"; s.ResourceKind = "type" }), false, "resource_kind must be"},
		{"allocate path under wrong root", with(graphops.LedgerAllocate, func(s *graphops.LedgerEventSpec) { s.Path = "links/l"; s.ResourceKind = graphops.KindBead }), false, "path:"},
		{"allocate noncanonical path", with(graphops.LedgerAllocate, func(s *graphops.LedgerEventSpec) { s.Path = "beads/caf%c3%a9"; s.ResourceKind = graphops.KindBead }), false, "path:"},
		{"tombstone ok", with(graphops.LedgerTombstone, func(s *graphops.LedgerEventSpec) {
			s.Path = "beads/x"
			s.ResourceKind = graphops.KindBead
			s.State = graphops.AllocationErased
		}), true, ""},
		{"tombstone without state", with(graphops.LedgerTombstone, func(s *graphops.LedgerEventSpec) { s.Path = "beads/x"; s.ResourceKind = graphops.KindBead }), false, "state is required"},
		{"tombstone live", with(graphops.LedgerTombstone, func(s *graphops.LedgerEventSpec) {
			s.Path = "beads/x"
			s.ResourceKind = graphops.KindBead
			s.State = graphops.AllocationLive
		}), false, "state must be pruned or erased"},
		{"tombstone reserved", with(graphops.LedgerTombstone, func(s *graphops.LedgerEventSpec) {
			s.Path = "beads/x"
			s.ResourceKind = graphops.KindBead
			s.State = graphops.AllocationReserved
		}), false, "state must be pruned or erased"},
		{"unknown kind", base(graphops.LedgerEventKind("bogus")), false, "unknown kind"},
		{"bad op_id", with(graphops.LedgerPromote, func(s *graphops.LedgerEventSpec) { s.OpID = "ABC" }), false, "op_id must be"},
		{"uppercase op_id", with(graphops.LedgerPromote, func(s *graphops.LedgerEventSpec) { s.OpID = strings.ToUpper(opID) }), false, "op_id must be"},
		{"bad authority_id", with(graphops.LedgerPromote, func(s *graphops.LedgerEventSpec) { s.AuthorityID = "" }), false, "authority_id must be"},
		{"zero at", with(graphops.LedgerPromote, func(s *graphops.LedgerEventSpec) { s.At = time.Time{} }), false, "at is required"},
		{"bad prev_hash", with(graphops.LedgerPromote, func(s *graphops.LedgerEventSpec) { s.PrevHash = "abc" }), false, "prev_hash must be"},
		{"bad stored hash", with(graphops.LedgerPromote, func(s *graphops.LedgerEventSpec) { s.Hash = strings.Repeat("0", 64) }), false, "does not verify"},
	} {
		_, err := graphops.NewLedgerEvent(tc.spec)
		if tc.ok {
			if err != nil {
				t.Errorf("%s: refused: %v", tc.name, err)
			}
			continue
		}
		wantValidation(t, err, tc.name)
		if !strings.Contains(err.Error(), tc.msg) {
			t.Errorf("%s: error %q does not mention %q", tc.name, err, tc.msg)
		}
	}
	for _, k := range []graphops.LedgerEventKind{graphops.LedgerMint, graphops.LedgerInstall, graphops.LedgerUpdate, graphops.LedgerPromote,
		graphops.LedgerRotate, graphops.LedgerAllocate, graphops.LedgerTombstone, graphops.LedgerRefuseURL} {
		if !k.Valid() {
			t.Errorf("%s should be a valid kind", k)
		}
	}
	if graphops.LedgerEventKind("").Valid() {
		t.Error("empty kind should not be valid")
	}
}

func TestLedgerManifestCovers(t *testing.T) {
	events := chain(t)
	mint := events[0]
	full, err := graphops.NewLedgerManifest(graphops.LedgerManifestSpec{
		ScopeURL: scopeURL, Lineage: mint.Hash(), FirstSeq: 1, LastSeq: 4, PrevHash: graphops.GenesisHash, HeadHash: events[3].Hash(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if full.ScopeURL() != scopeURL || full.Lineage() != mint.Hash() || full.FirstSeq() != 1 || full.LastSeq() != 4 ||
		full.PrevHash() != graphops.GenesisHash || full.HeadHash() != events[3].Hash() || full.IsZero() {
		t.Fatal("manifest accessors disagree with the spec")
	}
	if err := full.Covers(events); err != nil {
		t.Fatalf("full range refused: %v", err)
	}
	wantValidation(t, full.Covers(events[:3]), "too few events")
	wantValidation(t, full.Covers(events[1:]), "wrong first seq")
	wantValidation(t, full.Covers([]graphops.LedgerEvent{events[0], events[1], events[3], events[2]}), "broken chain")

	suffix, err := graphops.NewLedgerManifest(graphops.LedgerManifestSpec{
		ScopeURL: scopeURL, Lineage: mint.Hash(), FirstSeq: 2, LastSeq: 4, PrevHash: mint.Hash(), HeadHash: events[3].Hash(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := suffix.Covers(events[1:]); err != nil {
		t.Fatalf("suffix range refused: %v", err)
	}
	wantValidation(t, suffix.Covers(events[:3]), "suffix with wrong events")

	wrongHead, _ := graphops.NewLedgerManifest(graphops.LedgerManifestSpec{
		ScopeURL: scopeURL, Lineage: mint.Hash(), FirstSeq: 1, LastSeq: 4, PrevHash: graphops.GenesisHash, HeadHash: events[2].Hash(),
	})
	wantValidation(t, wrongHead.Covers(events), "wrong head hash")

	foreign, _ := graphops.NewLedgerManifest(graphops.LedgerManifestSpec{
		ScopeURL: scopeURL, Lineage: strings.Repeat("9", 64), FirstSeq: 1, LastSeq: 4, PrevHash: graphops.GenesisHash, HeadHash: events[3].Hash(),
	})
	wantValidation(t, foreign.Covers(events), "foreign lineage")

	// A range claiming to start at genesis must start with the mint event.
	notMint := mustEvent(t, graphops.LedgerEventSpec{
		Seq: 1, Kind: graphops.LedgerPromote, OpID: opID, AuthorityID: authorityID, Epoch: 1, At: at, PrevHash: graphops.GenesisHash,
	})
	single, _ := graphops.NewLedgerManifest(graphops.LedgerManifestSpec{
		ScopeURL: scopeURL, Lineage: notMint.Hash(), FirstSeq: 1, LastSeq: 1, PrevHash: graphops.GenesisHash, HeadHash: notMint.Hash(),
	})
	wantValidation(t, single.Covers([]graphops.LedgerEvent{notMint}), "genesis range without mint")

	for name, spec := range map[string]graphops.LedgerManifestSpec{
		"bad scope url":  {ScopeURL: "https://beads.example/acme", Lineage: mint.Hash(), FirstSeq: 1, LastSeq: 1, PrevHash: graphops.GenesisHash, HeadHash: mint.Hash()},
		"bad lineage":    {ScopeURL: scopeURL, Lineage: "x", FirstSeq: 1, LastSeq: 1, PrevHash: graphops.GenesisHash, HeadHash: mint.Hash()},
		"bad prev hash":  {ScopeURL: scopeURL, Lineage: mint.Hash(), FirstSeq: 1, LastSeq: 1, PrevHash: "", HeadHash: mint.Hash()},
		"bad head hash":  {ScopeURL: scopeURL, Lineage: mint.Hash(), FirstSeq: 1, LastSeq: 1, PrevHash: graphops.GenesisHash, HeadHash: "ABC"},
		"inverted range": {ScopeURL: scopeURL, Lineage: mint.Hash(), FirstSeq: 4, LastSeq: 1, PrevHash: graphops.GenesisHash, HeadHash: mint.Hash()},
	} {
		_, err := graphops.NewLedgerManifest(spec)
		wantValidation(t, err, "manifest "+name)
	}
	if !(graphops.LedgerManifest{}).IsZero() {
		t.Fatal("zero manifest must report IsZero")
	}
}
