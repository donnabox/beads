package graphops

import (
	"errors"
	"math"
	"strings"
	"testing"
)

// VerifyLedgerChain refuses a sequence that wraps around. NewLedgerEvent
// never mints a seq of 0 or MaxUint64, so the events that would exercise
// the check are built by hand here, in-package: the law must hold for any
// value a store could hand back, not only for the ones this package minted.
func TestVerifyLedgerChainRefusesWraparoundAndOutOfRangeSeqs(t *testing.T) {
	hashA, hashB := strings.Repeat("a", 64), strings.Repeat("b", 64)
	event := func(seq uint64, prev, hash string) LedgerEvent {
		return LedgerEvent{spec: LedgerEventSpec{Seq: seq, Kind: LedgerPromote, PrevHash: prev, Hash: hash}}
	}
	for name, events := range map[string][]LedgerEvent{
		"MaxUint64 then 0 (the wraparound)": {event(math.MaxUint64, GenesisHash, hashA), event(0, hashA, hashB)},
		"MaxLedgerSeq then 0":               {event(MaxLedgerSeq, GenesisHash, hashA), event(0, hashA, hashB)},
		"seq 0 alone":                       {event(0, GenesisHash, hashA)},
		"seq MaxUint64 alone":               {event(math.MaxUint64, GenesisHash, hashA)},
		"MaxLedgerSeq then MaxUint64":       {event(MaxLedgerSeq, GenesisHash, hashA), event(math.MaxUint64, hashA, hashB)},
	} {
		err := VerifyLedgerChain(GenesisHash, events)
		if !errors.Is(err, ErrValidation) {
			t.Errorf("%s: want ErrValidation, got %v", name, err)
		}
	}
	// The top of the range links normally.
	if err := VerifyLedgerChain(GenesisHash, []LedgerEvent{event(MaxLedgerSeq-1, GenesisHash, hashA), event(MaxLedgerSeq, hashA, hashB)}); err != nil {
		t.Fatalf("a chain ending at MaxLedgerSeq must verify: %v", err)
	}
}

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
