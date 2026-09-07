package graphops

import "testing"

// The URL helpers behind ParseRef are defensive about inputs their one caller
// has already screened (isAbsoluteURI guarantees complete escapes; the
// splitter guarantees a nonempty scheme). Those branches are still part of
// the law's statement, so they are pinned here, in-package, rather than left
// as dead code or removed.

func TestNormalizeEscapesKeepsMalformedInput(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"/a%41b", "/aAb"},   // unreserved: decoded
		{"/a%2fb", "/a%2Fb"}, // reserved: kept, uppercased
		{"/a%", "/a%"},       // incomplete: kept
		{"/a%4", "/a%4"},     // incomplete: kept
		{"/a%zzb", "/a%zzb"}, // malformed: kept
		{"/plain", "/plain"},
		{"/%7e%7E", "/~~"},
	} {
		if got := normalizeEscapes(tc.in); got != tc.want {
			t.Errorf("normalizeEscapes(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRemoveDotSegments(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"/a/../b", "/b"},
		{"/a/./b/", "/a/b/"},
		{"/../a", "/a"},
		{"/a/..", "/"},
		{"/a/.", "/a/"},
		{"/.", "/"},
		{"/..", "/"},
		{"/a/b/../../c", "/c"},
		{"/", "/"},
	} {
		if got := removeDotSegments(tc.in); got != tc.want {
			t.Errorf("removeDotSegments(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsSchemeToken(t *testing.T) {
	for in, want := range map[string]bool{
		"": false, "http": true, "HTTP": true, "h1+.-": true, "1http": false, "ht tp": false, "ht_tp": false, "-x": false,
	} {
		if got := isSchemeToken(in); got != want {
			t.Errorf("isSchemeToken(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestCodeUnitKeyOrdersLikeUTF16(t *testing.T) {
	// U+FFFF (one unit) versus U+10000 (D800 DC00): the lead surrogate sorts
	// first, and two supplementary characters order by their trail units.
	if !(codeUnitKey(0x10000) < codeUnitKey(0xFFFF)) || !(codeUnitKey(0x10000) < codeUnitKey(0x10001)) ||
		!(codeUnitKey('a') < codeUnitKey('b')) || !(codeUnitKey(0xD7FF) < codeUnitKey(0x10000)) {
		t.Fatal("codeUnitKey does not order like UTF-16 code units")
	}
}
