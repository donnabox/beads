package graphops

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// The laws: pure functions over strings and bytes, stated once here and
// enforced by every constructor in types.go. Nothing in this file touches a
// store, a clock, or a random source. Each law cites where its text comes
// from — the BDP draft at the pin (docs/specs/bdp.md, commit 0b7d86e7), its
// schema bundle, or the reference implementation that the pinned conformance
// matrix was proven against (packages/protocol/src/read-values.ts) — so a
// reader can check the code against the sentence rather than trust it.

// reason strips the ErrValidation prefix from a validation error's message so
// a law that wraps another law's refusal does not stutter the sentinel.
func reason(err error) string {
	return strings.TrimPrefix(err.Error(), ErrValidation.Error()+": ")
}

// isLowerHex reports whether s is exactly n lowercase hexadecimal digits: the
// shape of every id, revision, fingerprint and hash column.
func isLowerHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// hashHex is SHA-256 as 64 lowercase hex digits: the ledger hash and the
// descriptor fingerprint.
func hashHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ---------------------------------------------------------------------------
// Canonical-ID grammar
//
// BDP, "Scopes and identity" and "Scope discovery": a local Bead ID is
// beads/{id-path}, a local Link ID is links/{id-path}, an alias is
// alias/{alias-path}; {id-path} is one or more nonempty segments that are
// opaque identity. Empty, "." and ".." segments, controls, backslashes,
// queries, fragments, scheme-relative references and encoded "/" or "\"
// separators are invalid. Percent escapes are decoded exactly once and must
// decode to valid UTF-8; unreserved characters are emitted literally and every
// required escape uses uppercase hex; decoded segments compare exactly, with
// no Unicode normalization. A supplied spelling that is not already canonical
// is REJECTED, never normalized: trimming would mint an identity the creator
// did not write.
//
// The literal set is the reference implementation's LITERAL_PATH_CHARACTER
// (read-values.ts): RFC 3986 pchar minus pct-encoded — unreserved, sub-delims,
// ":" and "@". Everything else, non-ASCII included, is percent-encoded, so a
// canonical path is always pure ASCII.
// ---------------------------------------------------------------------------

const (
	beadRoot  = "beads"
	linkRoot  = "links"
	aliasRoot = "alias"
)

const upperHexDigits = "0123456789ABCDEF"
const lowerHexDigits = "0123456789abcdef"

func isLiteralSegmentByte(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	}
	switch c {
	case '.', '_', '~', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '@', '-':
		return true
	}
	return false
}

// isUnreservedByte is RFC 3986 unreserved: a character that is never
// percent-encoded in a canonical URL.
func isUnreservedByte(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		return true
	}
	return c == '-' || c == '.' || c == '_' || c == '~'
}

func hexValue(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// CanonicalSegment encodes one DECODED segment the way the authority emits
// it: the literal set as is, every other byte of its UTF-8 form as an
// uppercase percent escape. It is the encoder half of the grammar; an
// authority allocating an id, or a creator spelling one, produces exactly
// this, and ValidateCanonicalSegment accepts exactly this.
func CanonicalSegment(decoded string) string {
	var b strings.Builder
	b.Grow(len(decoded))
	for i := 0; i < len(decoded); i++ {
		c := decoded[i]
		if isLiteralSegmentByte(c) {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(upperHexDigits[c>>4])
		b.WriteByte(upperHexDigits[c&0x0F])
	}
	return b.String()
}

// decodeSegment decodes percent escapes exactly once. It accepts either hex
// case here — the canonical re-encoding comparison is what refuses lowercase.
func decodeSegment(segment string) ([]byte, error) {
	out := make([]byte, 0, len(segment))
	for i := 0; i < len(segment); i++ {
		c := segment[i]
		if c != '%' {
			out = append(out, c)
			continue
		}
		if i+2 >= len(segment) {
			return nil, errors.New("incomplete percent escape")
		}
		hi, ok1 := hexValue(segment[i+1])
		lo, ok2 := hexValue(segment[i+2])
		if !ok1 || !ok2 {
			return nil, errors.New("malformed percent escape")
		}
		out = append(out, hi<<4|lo)
		i += 2
	}
	return out, nil
}

// ValidateCanonicalSegment applies the segment grammar to one segment of a
// local ID, an alias, or a Scope URL path: nonempty; percent escapes complete
// and decoding, once, to valid UTF-8; not "." or ".."; no "/" or "\" and no
// ASCII control (U+0000–U+001F, U+007F) after decoding; and spelled exactly as
// CanonicalSegment would spell the decoded value — which is what refuses a
// lowercase escape, an escaped unreserved character, and a literal non-ASCII
// character.
func ValidateCanonicalSegment(segment string) error {
	if segment == "" {
		return fmt.Errorf("%w: path segment must be nonempty", ErrValidation)
	}
	decoded, err := decodeSegment(segment)
	if err != nil {
		return fmt.Errorf("%w: path segment %q: %v", ErrValidation, segment, err)
	}
	if !utf8.Valid(decoded) {
		return fmt.Errorf("%w: path segment %q does not decode to valid UTF-8", ErrValidation, segment)
	}
	s := string(decoded)
	if s == "." || s == ".." {
		return fmt.Errorf("%w: path segment %q is a dot segment", ErrValidation, segment)
	}
	for _, r := range s {
		switch {
		case r <= 0x1F || r == 0x7F:
			return fmt.Errorf("%w: path segment %q decodes to a control character", ErrValidation, segment)
		case r == '/' || r == '\\':
			return fmt.Errorf("%w: path segment %q decodes to a separator", ErrValidation, segment)
		}
	}
	if CanonicalSegment(s) != segment {
		return fmt.Errorf("%w: path segment %q is not canonically encoded (canonical spelling is %q)", ErrValidation, segment, CanonicalSegment(s))
	}
	return nil
}

func validateRootedPath(path, root, what string) error {
	if strings.ContainsAny(path, "?#") {
		return fmt.Errorf("%w: %s %q must not contain a query or fragment", ErrValidation, what, path)
	}
	segments := strings.Split(path, "/")
	if segments[0] != root || len(segments) < 2 {
		return fmt.Errorf("%w: %s %q must begin with %s/ and contain an ID path", ErrValidation, what, path, root)
	}
	for _, segment := range segments[1:] {
		if err := ValidateCanonicalSegment(segment); err != nil {
			return fmt.Errorf("%w: %s %q: %s", ErrValidation, what, path, reason(err))
		}
	}
	return nil
}

// ValidateBeadPath accepts exactly the canonical local Bead IDs: "beads/"
// followed by one or more canonical segments. Case matters — beads/Task and
// beads/task are two Beads.
func ValidateBeadPath(path string) error { return validateRootedPath(path, beadRoot, "bead path") }

// ValidateLinkPath accepts exactly the canonical local Link IDs under "links/".
func ValidateLinkPath(path string) error { return validateRootedPath(path, linkRoot, "link path") }

// ValidateAliasPath accepts exactly the alias locators under "alias/": the
// same grammar, so whether a spelling names identity or an alias is decidable
// from the spelling alone. Aliases are not served in v0; the law is here so
// the reserved root is refused as identity rather than mistaken for it.
func ValidateAliasPath(path string) error { return validateRootedPath(path, aliasRoot, "alias path") }

// ValidatePath validates a path under the root its kind fixes.
func ValidatePath(path string, kind ResourceKind) error {
	switch kind {
	case KindBead:
		return ValidateBeadPath(path)
	case KindLink:
		return ValidateLinkPath(path)
	}
	return fmt.Errorf("%w: resource kind %q is not bead or link", ErrValidation, kind)
}

// CanonicalURL is the absolute canonical Resource URL: the Scope URL (which
// ends in "/") followed by the Scope-relative path. It is concatenation, by
// design — rows store the path, the URL is computed at the boundary, and a
// Scope URL rotation rewrites nothing. Both inputs are assumed valid.
func CanonicalURL(scopeURL, path string) string { return scopeURL + path }

// SplitCanonicalURL is the inverse: for a URL that is exactly scopeURL
// followed by a canonical Bead or Link path, it returns the path and its kind.
// Anything else — a different Scope, a non-canonical spelling, a path under
// another root — is not a canonical Resource URL of this Scope. scopeURL is
// assumed valid.
func SplitCanonicalURL(scopeURL, url string) (path string, kind ResourceKind, ok bool) {
	if !strings.HasPrefix(url, scopeURL) {
		return "", "", false
	}
	path = url[len(scopeURL):]
	switch {
	case ValidateBeadPath(path) == nil:
		return path, KindBead, true
	case ValidateLinkPath(path) == nil:
		return path, KindLink, true
	}
	return "", "", false
}

// ---------------------------------------------------------------------------
// Code-unit ordering
//
// BDP, "Collection retrieval and selection": the baseline order every
// authority must serve is canonical-uri — ascending lexicographic comparison,
// by Unicode code unit, of each item's absolute canonical id — and the
// reference implementation compares with JavaScript's string operators, which
// order by UTF-16 code unit. RFC 8785 §3.2.3 sorts object keys the same way.
// This is that comparison, without converting either string.
//
// For the ASCII alphabet every canonical path and URL is made of, code-unit
// order IS byte order, which is what the storage leg's binary-collated
// ORDER BY path produces. The two orders differ only for strings mixing
// characters above U+FFFF with characters in U+E000–U+FFFF (a supplementary
// character's lead surrogate sorts below them), which is why the law is stated
// over code units rather than bytes and why JSON keys sort here and not with
// bytes.Compare.
// ---------------------------------------------------------------------------

// codeUnitKey maps a scalar value to an integer whose order is the order of
// its UTF-16 code-unit sequence: a BMP character sorts by its single unit, a
// supplementary character by its lead surrogate and then its trail. Because a
// scalar value is never a surrogate, a BMP unit never ties with a lead
// surrogate, so the packed comparison decides every pair.
func codeUnitKey(r rune) uint32 {
	if r < 0x10000 {
		return uint32(r) << 16
	}
	r -= 0x10000
	return (0xD800+uint32(r>>10))<<16 | (0xDC00 + uint32(r&0x3FF))
}

// CompareCodeUnits orders two strings by their UTF-16 code units, returning
// -1, 0 or +1. Invalid UTF-8 decodes as U+FFFD, as it would in every consumer.
func CompareCodeUnits(a, b string) int {
	for len(a) > 0 && len(b) > 0 {
		ra, na := utf8.DecodeRuneInString(a)
		rb, nb := utf8.DecodeRuneInString(b)
		if ra != rb {
			if codeUnitKey(ra) < codeUnitKey(rb) {
				return -1
			}
			return 1
		}
		a, b = a[na:], b[nb:]
	}
	switch {
	case len(a) == 0 && len(b) == 0:
		return 0
	case len(a) == 0:
		return -1
	}
	return 1
}

// ---------------------------------------------------------------------------
// JSON canonicalization and equality
//
// What is canonicalized: any JSON text (RFC 8259) into ONE byte string, such
// that two texts canonicalize to the same bytes exactly when they are equal
// under RFC 6902 §4.6 — the comparison that decides whether a properties write
// is a no-op. The form is RFC 8785 (JCS) with one stated departure:
//
//   - Whitespace outside strings is removed.
//   - Object members are sorted by the UTF-16 code units of their keys
//     (CompareCodeUnits); a duplicate key is refused, as I-JSON requires.
//   - Strings are serialized as ECMAScript JSON.stringify does: '"' and '\'
//     escaped, U+0008/0009/000A/000C/000D as \b \t \n \f \r, every other
//     control below U+0020 as \u00xx with lowercase hex, and everything else
//     literally in UTF-8 — U+007F, U+2028, non-ASCII, '/' included. An input
//     that is not valid UTF-8, or that escapes a lone surrogate, is refused
//     rather than laundered into U+FFFD.
//   - Literals are true, false, null.
//   - Numbers are serialized with ECMAScript's Number::toString digit
//     placement (RFC 8785 §3.2.2.3): no exponent for 10^-6 ≤ |x| < 10^21,
//     shortest digit string, "e+"/"e-" otherwise — APPLIED TO THE EXACT
//     DECIMAL VALUE OF THE LITERAL, not to its nearest IEEE-754 double. This
//     is the departure. For every literal a double represents exactly under
//     shortest round-trip the output is byte-identical to JCS ("1.0" → "1",
//     "-0.0" → "0", "1e300" → "1e+300", "0.1" → "0.1"); where JCS would round
//     (an integer past 2^53, more than 17 significant digits, an exponent
//     past the double range) the value is preserved exactly instead. The plan
//     forbids float64 laundering of properties, RFC 6902 equality is numeric
//     equality, and the two demands meet only here: 9007199254740993 and
//     9007199254740992 are different values and canonicalize differently.
//
// DECISION: that departure. Alternatives were bit-exact JCS (rounds numbers
// the storage design exists to keep) and source-literal preservation (the
// metadata plane's rule, under which 1 and 1.0 are different values, which
// contradicts the no-op law). Exact-value JCS satisfies both texts.
//
// Nesting is bounded at maxJSONDepth (the depth encoding/json accepts) as a
// stack-safety measure, not a protocol limit; advertised limits are the
// serving surface's.
// ---------------------------------------------------------------------------

const maxJSONDepth = 10000

// CanonicalizeJSON returns the canonical bytes of one JSON text, or
// ErrValidation for anything that is not exactly one well-formed JSON value.
func CanonicalizeJSON(raw []byte) ([]byte, error) {
	s := &jsonScanner{in: raw}
	s.skipSpace()
	out, err := s.value(make([]byte, 0, len(raw)))
	if err != nil {
		return nil, err
	}
	s.skipSpace()
	if s.pos != len(s.in) {
		return nil, s.errorf("unexpected content after the JSON value")
	}
	return out, nil
}

// JSONEqual is RFC 6902 §4.6 equality of two JSON texts: numbers by numeric
// value, strings by code points, arrays element-wise, objects as member sets,
// literals as themselves. Either text failing to parse is an error.
func JSONEqual(a, b []byte) (bool, error) {
	ca, err := CanonicalizeJSON(a)
	if err != nil {
		return false, err
	}
	cb, err := CanonicalizeJSON(b)
	if err != nil {
		return false, err
	}
	return string(ca) == string(cb), nil
}

type jsonScanner struct {
	in    []byte
	pos   int
	depth int
}

func (s *jsonScanner) errorf(format string, args ...any) error {
	return fmt.Errorf("%w: JSON at byte %d: %s", ErrValidation, s.pos, fmt.Sprintf(format, args...))
}

func (s *jsonScanner) skipSpace() {
	for s.pos < len(s.in) {
		switch s.in[s.pos] {
		case ' ', '\t', '\n', '\r':
			s.pos++
		default:
			return
		}
	}
}

func (s *jsonScanner) value(out []byte) ([]byte, error) {
	if s.pos >= len(s.in) {
		return nil, s.errorf("unexpected end of input")
	}
	switch c := s.in[s.pos]; {
	case c == '{':
		return s.object(out)
	case c == '[':
		return s.array(out)
	case c == '"':
		str, err := s.str()
		if err != nil {
			return nil, err
		}
		return appendCanonicalString(out, str), nil
	case c == 't':
		return s.literal(out, "true")
	case c == 'f':
		return s.literal(out, "false")
	case c == 'n':
		return s.literal(out, "null")
	case c == '-' || (c >= '0' && c <= '9'):
		return s.number(out)
	}
	return nil, s.errorf("unexpected byte %q", s.in[s.pos])
}

func (s *jsonScanner) enter() error {
	s.depth++
	if s.depth > maxJSONDepth {
		return s.errorf("nesting deeper than %d", maxJSONDepth)
	}
	return nil
}

type jsonMember struct {
	key   string
	value []byte
}

func (s *jsonScanner) object(out []byte) ([]byte, error) {
	if err := s.enter(); err != nil {
		return nil, err
	}
	s.pos++ // '{'
	s.skipSpace()
	if s.pos < len(s.in) && s.in[s.pos] == '}' {
		s.pos++
		s.depth--
		return append(out, '{', '}'), nil
	}
	var members []jsonMember
	seen := map[string]struct{}{}
	for {
		s.skipSpace()
		if s.pos >= len(s.in) || s.in[s.pos] != '"' {
			return nil, s.errorf("expected an object key")
		}
		key, err := s.str()
		if err != nil {
			return nil, err
		}
		if _, dup := seen[key]; dup {
			return nil, s.errorf("duplicate object key %q", key)
		}
		seen[key] = struct{}{}
		s.skipSpace()
		if s.pos >= len(s.in) || s.in[s.pos] != ':' {
			return nil, s.errorf("expected ':' after object key")
		}
		s.pos++
		s.skipSpace()
		value, err := s.value(nil)
		if err != nil {
			return nil, err
		}
		members = append(members, jsonMember{key: key, value: value})
		s.skipSpace()
		if s.pos >= len(s.in) {
			return nil, s.errorf("unterminated object")
		}
		if s.in[s.pos] == ',' {
			s.pos++
			continue
		}
		if s.in[s.pos] == '}' {
			s.pos++
			break
		}
		return nil, s.errorf("expected ',' or '}' in object")
	}
	sort.SliceStable(members, func(i, j int) bool {
		return CompareCodeUnits(members[i].key, members[j].key) < 0
	})
	out = append(out, '{')
	for i, m := range members {
		if i > 0 {
			out = append(out, ',')
		}
		out = appendCanonicalString(out, m.key)
		out = append(out, ':')
		out = append(out, m.value...)
	}
	s.depth--
	return append(out, '}'), nil
}

func (s *jsonScanner) array(out []byte) ([]byte, error) {
	if err := s.enter(); err != nil {
		return nil, err
	}
	s.pos++ // '['
	s.skipSpace()
	out = append(out, '[')
	if s.pos < len(s.in) && s.in[s.pos] == ']' {
		s.pos++
		s.depth--
		return append(out, ']'), nil
	}
	for i := 0; ; i++ {
		s.skipSpace()
		if i > 0 {
			out = append(out, ',')
		}
		var err error
		if out, err = s.value(out); err != nil {
			return nil, err
		}
		s.skipSpace()
		if s.pos >= len(s.in) {
			return nil, s.errorf("unterminated array")
		}
		if s.in[s.pos] == ',' {
			s.pos++
			continue
		}
		if s.in[s.pos] == ']' {
			s.pos++
			break
		}
		return nil, s.errorf("expected ',' or ']' in array")
	}
	s.depth--
	return append(out, ']'), nil
}

func (s *jsonScanner) literal(out []byte, word string) ([]byte, error) {
	if !strings.HasPrefix(string(s.in[s.pos:]), word) {
		return nil, s.errorf("invalid literal")
	}
	s.pos += len(word)
	return append(out, word...), nil
}

func (s *jsonScanner) hex4() (rune, error) {
	if s.pos+4 > len(s.in) {
		return 0, s.errorf("truncated \\u escape")
	}
	var r rune
	for i := 0; i < 4; i++ {
		v, ok := hexValue(s.in[s.pos+i])
		if !ok {
			return 0, s.errorf("malformed \\u escape")
		}
		r = r<<4 | rune(v)
	}
	s.pos += 4
	return r, nil
}

// str parses a JSON string at the opening quote and returns its decoded
// value. Raw bytes must be valid UTF-8 and free of unescaped controls;
// escapes must be the RFC 8259 set; a \u escape of a surrogate must be half
// of a well-formed pair.
func (s *jsonScanner) str() (string, error) {
	s.pos++ // opening quote
	var buf []byte
	start := s.pos
	for {
		if s.pos >= len(s.in) {
			return "", s.errorf("unterminated string")
		}
		c := s.in[s.pos]
		switch {
		case c == '"':
			segment := s.in[start:s.pos]
			if !utf8.Valid(segment) {
				return "", s.errorf("string is not valid UTF-8")
			}
			s.pos++
			if buf == nil {
				return string(segment), nil
			}
			return string(append(buf, segment...)), nil
		case c == '\\':
			segment := s.in[start:s.pos]
			if !utf8.Valid(segment) {
				return "", s.errorf("string is not valid UTF-8")
			}
			buf = append(buf, segment...)
			s.pos++
			if s.pos >= len(s.in) {
				return "", s.errorf("unterminated string")
			}
			e := s.in[s.pos]
			s.pos++
			switch e {
			case '"', '\\', '/':
				buf = append(buf, e)
			case 'b':
				buf = append(buf, '\b')
			case 'f':
				buf = append(buf, '\f')
			case 'n':
				buf = append(buf, '\n')
			case 'r':
				buf = append(buf, '\r')
			case 't':
				buf = append(buf, '\t')
			case 'u':
				r, err := s.hex4()
				if err != nil {
					return "", err
				}
				switch {
				case r >= 0xD800 && r <= 0xDBFF:
					if s.pos+1 >= len(s.in) || s.in[s.pos] != '\\' || s.in[s.pos+1] != 'u' {
						return "", s.errorf("lone lead surrogate escape")
					}
					s.pos += 2
					trail, err := s.hex4()
					if err != nil {
						return "", err
					}
					if trail < 0xDC00 || trail > 0xDFFF {
						return "", s.errorf("lone lead surrogate escape")
					}
					r = 0x10000 + (r-0xD800)<<10 + (trail - 0xDC00)
				case r >= 0xDC00 && r <= 0xDFFF:
					return "", s.errorf("lone trail surrogate escape")
				}
				buf = utf8.AppendRune(buf, r)
			default:
				return "", s.errorf("invalid escape \\%c", e)
			}
			start = s.pos
		case c < 0x20:
			return "", s.errorf("unescaped control character in string")
		default:
			s.pos++
		}
	}
}

// appendCanonicalString serializes a decoded string as JSON.stringify would.
func appendCanonicalString(out []byte, str string) []byte {
	out = append(out, '"')
	for i := 0; i < len(str); i++ {
		c := str[i]
		switch {
		case c == '"':
			out = append(out, '\\', '"')
		case c == '\\':
			out = append(out, '\\', '\\')
		case c == '\b':
			out = append(out, '\\', 'b')
		case c == '\f':
			out = append(out, '\\', 'f')
		case c == '\n':
			out = append(out, '\\', 'n')
		case c == '\r':
			out = append(out, '\\', 'r')
		case c == '\t':
			out = append(out, '\\', 't')
		case c < 0x20:
			out = append(out, '\\', 'u', '0', '0', lowerHexDigits[c>>4], lowerHexDigits[c&0x0F])
		default:
			out = append(out, c)
		}
	}
	return append(out, '"')
}

// maxExponentMagnitude bounds the exponent of a number's CANONICAL form —
// the e that Number::toString would print — not the exponent the literal
// happened to be spelled with. Beyond it the exact value is still
// representable here, but no consumer could use it, and unbounded
// accumulation would be an integer overflow waiting to happen. Bounding the
// normalized value is what makes canonicalization a fixed point: a literal
// is admitted exactly when its canonical form re-parses, so 10e1000000000000
// (canonically 1e+1000000000001) is refused on the way in rather than stored
// and then unreadable, and 0.1e1000000000001 (canonically 1e+1000000000000)
// is admitted although its spelled exponent exceeds the bound.
const maxExponentMagnitude = 1_000_000_000_000

func (s *jsonScanner) digits() ([]byte, error) {
	start := s.pos
	for s.pos < len(s.in) && s.in[s.pos] >= '0' && s.in[s.pos] <= '9' {
		s.pos++
	}
	if s.pos == start {
		return nil, s.errorf("expected a digit")
	}
	return s.in[start:s.pos], nil
}

// number parses one RFC 8259 number and appends its canonical form.
func (s *jsonScanner) number(out []byte) ([]byte, error) {
	neg := false
	if s.in[s.pos] == '-' {
		neg = true
		s.pos++
	}
	if s.pos >= len(s.in) {
		return nil, s.errorf("expected a digit")
	}
	var intPart []byte
	if s.in[s.pos] == '0' {
		intPart = s.in[s.pos : s.pos+1]
		s.pos++
	} else {
		var err error
		if intPart, err = s.digits(); err != nil {
			return nil, err
		}
	}
	var frac []byte
	if s.pos < len(s.in) && s.in[s.pos] == '.' {
		s.pos++
		var err error
		if frac, err = s.digits(); err != nil {
			return nil, err
		}
	}
	var exp int64
	if s.pos < len(s.in) && (s.in[s.pos] == 'e' || s.in[s.pos] == 'E') {
		s.pos++
		expNeg := false
		if s.pos < len(s.in) && (s.in[s.pos] == '+' || s.in[s.pos] == '-') {
			expNeg = s.in[s.pos] == '-'
			s.pos++
		}
		expDigits, err := s.digits()
		if err != nil {
			return nil, err
		}
		// The canonical exponent differs from the spelled one by fewer
		// places than the literal has digits, so a spelled exponent this far
		// beyond the bound can never normalize back inside it; refusing it
		// here keeps the accumulation from overflowing. The bound itself is
		// applied to the normalized value below.
		limit := maxExponentMagnitude + int64(len(s.in))
		for _, d := range expDigits {
			exp = exp*10 + int64(d-'0')
			if exp > limit {
				return nil, s.errorf("exponent out of range")
			}
		}
		if expNeg {
			exp = -exp
		}
	}
	out, ok := appendCanonicalNumber(out, neg, intPart, frac, exp)
	if !ok {
		return nil, s.errorf("exponent out of range")
	}
	return out, nil
}

// appendCanonicalNumber formats the exact decimal value ±(intPart.frac)×10^exp
// with ECMAScript Number::toString digit placement. It reports false when the
// normalized value's exponent exceeds maxExponentMagnitude, which is the one
// way a syntactically valid number is refused.
func appendCanonicalNumber(out []byte, neg bool, intPart, frac []byte, exp int64) ([]byte, bool) {
	digits := make([]byte, 0, len(intPart)+len(frac))
	digits = append(digits, intPart...)
	digits = append(digits, frac...)
	exp10 := exp - int64(len(frac))
	lead := 0
	for lead < len(digits) && digits[lead] == '0' {
		lead++
	}
	digits = digits[lead:]
	if len(digits) == 0 {
		return append(out, '0'), true // every zero, -0 included, is "0"
	}
	trail := len(digits)
	for digits[trail-1] == '0' {
		trail--
		exp10++
	}
	digits = digits[:trail]
	k := int64(len(digits))
	n := exp10 + k // value = 0.digits × 10^n
	if e := n - 1; e > maxExponentMagnitude || e < -maxExponentMagnitude {
		return out, false
	}
	if neg {
		out = append(out, '-')
	}
	switch {
	case k <= n && n <= 21:
		out = append(out, digits...)
		for i := k; i < n; i++ {
			out = append(out, '0')
		}
	case 0 < n && n <= 21:
		out = append(out, digits[:n]...)
		out = append(out, '.')
		out = append(out, digits[n:]...)
	case -6 < n && n <= 0:
		out = append(out, '0', '.')
		for i := n; i < 0; i++ {
			out = append(out, '0')
		}
		out = append(out, digits...)
	default:
		out = append(out, digits[0])
		if k > 1 {
			out = append(out, '.')
			out = append(out, digits[1:]...)
		}
		out = append(out, 'e')
		e := n - 1
		if e >= 0 {
			out = append(out, '+')
		} else {
			out = append(out, '-')
			e = -e
		}
		out = strconv.AppendInt(out, e, 10)
	}
	return out, true
}

// ---------------------------------------------------------------------------
// Scope URL and Type URL
//
// BDP, "Scope discovery": every Scope has one absolute canonical Scope URL
// ending in "/"; Scope, Resource, Type, schema and navigation members are
// HTTP(S) URLs. The reference implementation (read-values.ts
// parseCanonicalHttpUrl / parseCanonicalScope) makes "canonical" precise: an
// absolute http or https URL with no credentials and no fragment whose WHATWG
// serialization is the input itself — lowercase scheme and host, no default
// port, no dot segments, no character the parser would percent-encode —
// with every percent escape complete, and in the path uppercase and never of
// an unreserved character; a Scope URL additionally has no query, ends in
// "/", and every path segment obeys the ID-segment grammar.
//
// ONE MODEL, THREE USES. The WHATWG parser's host and port rules are stated
// once, in normalizeOrigin (P0 council, 2026-09-07): a host is percent-
// decoded and lowercased, a host that "ends in a number" is an IPv4 address
// in any of the parser's spellings (0x7f000001, 127.1, 0177.0.0.1,
// 2130706433) and serializes as dotted decimal, an IPv6 address serializes
// compressed in lowercase hex with an IPv4-mapped tail in hex pieces, and a
// port is a decimal number without leading zeros that is dropped when it is
// the scheme's default (:0443 is :443 is nothing). validateCanonicalHTTPURL
// accepts a URL exactly when its spelled authority already equals that
// serialization; NormalizeScopeURL rewrites a configured URL's authority to
// it; claimsScope classifies a reference by it, so no spelling of the Scope's
// own origin can slip past the in-Scope endpoint law as "external". An
// external reference — a different origin — is never rewritten: it is
// preserved byte-for-byte, however it is spelled.
//
// The startup contract the design mirrors (docs/design/startup-configuration.md
// at the pin, "Canonical Scope identity") NORMALIZES the spellings that mean
// the same Scope — the origin's spelling and a missing trailing slash — and
// REFUSES the rest, and it reserves the first path segment "local-test" for
// a derived development identity that is never persisted. NormalizeScopeURL
// is that admission rule; ValidateScopeURL is what any Scope URL a client may
// name must satisfy, and ValidatePersistedScopeURL is what this store may
// persist as its own identity.
//
// Scope URL diagnostics never echo the value: the startup contract's reason
// is that a token pasted into the variable by mistake would otherwise be
// copied into a log line by the parse error that rejects it.
// ---------------------------------------------------------------------------

type urlParts struct {
	scheme, host, port, path, query, fragment   string
	hasUserinfo, hasPort, hasQuery, hasFragment bool
}

func isSchemeToken(s string) bool {
	if s == "" || !(s[0] >= 'A' && s[0] <= 'Z' || s[0] >= 'a' && s[0] <= 'z') {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

// splitHTTPURL splits scheme://[userinfo@]host[:port]path[?query][#fragment]
// without normalizing anything.
func splitHTTPURL(s string) (urlParts, error) {
	var p urlParts
	i := strings.IndexByte(s, ':')
	if i <= 0 || !isSchemeToken(s[:i]) {
		return p, errors.New("must be an absolute URL")
	}
	p.scheme = s[:i]
	rest := s[i+1:]
	if !strings.HasPrefix(rest, "//") {
		return p, errors.New("must have an authority")
	}
	rest = rest[2:]
	authority := rest
	if end := strings.IndexAny(rest, "/?#"); end >= 0 {
		authority, rest = rest[:end], rest[end:]
	} else {
		rest = ""
	}
	if j := strings.IndexByte(rest, '#'); j >= 0 {
		p.hasFragment, p.fragment, rest = true, rest[j+1:], rest[:j]
	}
	if j := strings.IndexByte(rest, '?'); j >= 0 {
		p.hasQuery, p.query, rest = true, rest[j+1:], rest[:j]
	}
	p.path = rest
	if j := strings.LastIndexByte(authority, '@'); j >= 0 {
		p.hasUserinfo, authority = true, authority[j+1:]
	}
	if strings.HasPrefix(authority, "[") {
		j := strings.IndexByte(authority, ']')
		if j < 0 {
			return p, errors.New("has an unterminated IPv6 host")
		}
		p.host = authority[:j+1]
		if tail := authority[j+1:]; tail != "" {
			if tail[0] != ':' {
				return p, errors.New("has a malformed authority")
			}
			p.hasPort, p.port = true, tail[1:]
		}
		return p, nil
	}
	if j := strings.LastIndexByte(authority, ':'); j >= 0 {
		p.host, p.hasPort, p.port = authority[:j], true, authority[j+1:]
		return p, nil
	}
	p.host = authority
	return p, nil
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// origin is the normalized (scheme, host, port) of an http(s) URL: the three
// components the WHATWG parser serializes canonically and the tuple RFC 6454
// calls the origin. Two URLs with equal origins and one path prefix name the
// same Scope whatever their spelling; a canonical URL is one whose spelled
// authority already equals its origin.
type origin struct{ scheme, host, port string }

// normalizeOrigin applies the parser's rules to a split URL: the scheme is
// lowercased and must be http or https, the host is normalizeHost's
// serialization, and the port is normalizePort's.
func normalizeOrigin(p urlParts) (origin, error) {
	scheme := strings.ToLower(p.scheme)
	if scheme != "http" && scheme != "https" {
		return origin{}, errors.New("must use the http or https scheme")
	}
	host, err := normalizeHost(p.host)
	if err != nil {
		return origin{}, err
	}
	port, err := normalizePort(p.port, scheme)
	if err != nil {
		return origin{}, err
	}
	return origin{scheme: scheme, host: host, port: port}, nil
}

func defaultPort(scheme string) string {
	if scheme == "https" {
		return "443"
	}
	return "80"
}

// normalizePort is the WHATWG port rule: an absent or empty port is no port,
// a port is decimal digits whose leading zeros carry nothing (:0443 is :443),
// it may not exceed 65535, and the scheme's default port is no port at all.
func normalizePort(port, scheme string) (string, error) {
	if port == "" {
		return "", nil
	}
	if !allDigits(port) {
		return "", errors.New("port must be a decimal number")
	}
	trimmed := strings.TrimLeft(port, "0")
	if trimmed == "" {
		trimmed = "0"
	}
	if len(trimmed) > 5 || (len(trimmed) == 5 && trimmed > "65535") {
		return "", errors.New("port must not exceed 65535")
	}
	if trimmed == defaultPort(scheme) {
		return "", nil
	}
	return trimmed, nil
}

// isForbiddenDomainByte is the WHATWG "forbidden domain code point" set:
// the forbidden host code points (NUL, tab, LF, CR, space, "#", "/", ":",
// "<", ">", "?", "@", "[", "\", "]", "^", "|") plus every C0 control, "%"
// and DEL. A host that contains one after percent-decoding is not a host.
func isForbiddenDomainByte(c byte) bool {
	if c <= 0x20 || c == 0x7F {
		return true
	}
	switch c {
	case '#', '%', '/', ':', '<', '>', '?', '@', '[', '\\', ']', '^', '|':
		return true
	}
	return false
}

// normalizeHost is the WHATWG host parser, restricted to what a canonical
// Scope or Type host may be, returning the host's serialization:
//
//   - a bracketed IPv6 address serializes as the WHATWG IPv6 serializer
//     writes it — lowercase hex, the first longest run of two or more zero
//     pieces compressed, an IPv4-mapped address in hex pieces
//     ([::ffff:102:304], never [::ffff:1.2.3.4]); a zone is refused;
//   - any other host is percent-decoded (a malformed escape is refused), must
//     decode to ASCII with no forbidden domain code point, and is lowercased;
//   - a host whose last label is all digits or a 0x-prefixed hex number "ends
//     in a number" and is parsed as IPv4 — at most four parts, each decimal,
//     octal (leading 0) or hexadecimal (0x), the last filling the remaining
//     bytes — and serializes as dotted decimal: 0x7f000001, 127.1,
//     0177.0.0.1, 2130706433 and 127.0.0.1. are all 127.0.0.1, and beads.123
//     is not a host at all;
//   - otherwise it is a registered name: labels of lowercase letters, digits,
//     "-" and "_" (DNS LDH plus the underscore the tree already accepts), no
//     label empty — except that ONE trailing dot is kept as written, because
//     the parser leaves beads.example. unchanged and treats it as a host
//     distinct from beads.example, and so does this law.
//
// DECISION: IDNA is not applied. A host that decodes to non-ASCII is refused
// rather than mapped, so an internationalized Scope host is configured in its
// A-label (xn--) form, and a U-label spelling of it is classified as an
// external reference rather than resolved. The startup contract already
// refuses non-ASCII input; this closes the percent-encoded route the same way.
func normalizeHost(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("host must be nonempty")
	}
	if raw[0] == '[' {
		// splitHTTPURL only produces a bracketed host with its closing
		// bracket, so the slice below is well-formed by construction.
		return normalizeIPv6Host(raw[1 : len(raw)-1])
	}
	decoded, err := decodeSegment(raw)
	if err != nil {
		return "", fmt.Errorf("host has an %v", err)
	}
	host := make([]byte, 0, len(decoded))
	for _, c := range decoded {
		if c >= 0x80 {
			return "", errors.New("host must be ASCII (IDNA mapping is not applied)")
		}
		if isForbiddenDomainByte(c) {
			return "", errors.New("host contains a character the URL parser forbids")
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		host = append(host, c)
	}
	s := string(host)
	if hostEndsInANumber(s) {
		return parseIPv4Host(s)
	}
	labels := strings.Split(s, ".")
	if n := len(labels); n > 1 && labels[n-1] == "" {
		labels = labels[:n-1] // the one trailing dot: kept in s, not a label
	}
	for _, label := range labels {
		if label == "" {
			return "", errors.New("host must not contain an empty label")
		}
		for i := 0; i < len(label); i++ {
			if c := label[i]; !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
				return "", errors.New("host must be a lowercase registered name or a canonical IP address")
			}
		}
	}
	return s, nil
}

// hostEndsInANumber is the WHATWG "ends in a number" checker over a decoded,
// lowercased, nonempty host: a trailing empty label is dropped first, and
// the last label is a number when it is all digits or parses as a
// 0x-prefixed IPv4 number.
func hostEndsInANumber(host string) bool {
	parts := strings.Split(host, ".")
	if n := len(parts); n > 1 && parts[n-1] == "" {
		parts = parts[:n-1]
	}
	last := parts[len(parts)-1]
	if allDigits(last) {
		return true
	}
	_, ok := parseIPv4Number(last)
	return ok
}

// parseIPv4Number is the WHATWG IPv4 number parser: decimal, octal with a
// leading 0, or hexadecimal with 0x/0X, where an empty digit string after
// the prefix is 0. The value is capped well above the largest meaningful
// part so a long input cannot overflow.
func parseIPv4Number(s string) (uint64, bool) {
	if s == "" {
		return 0, false
	}
	radix := uint64(10)
	switch {
	case len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X'):
		s, radix = s[2:], 16
	case len(s) >= 2 && s[0] == '0':
		s, radix = s[1:], 8
	}
	if s == "" {
		return 0, true
	}
	var n uint64
	for i := 0; i < len(s); i++ {
		d, ok := hexValue(s[i])
		if !ok || uint64(d) >= radix {
			return 0, false
		}
		n = n*radix + uint64(d)
		if n > 1<<40 {
			return 0, false
		}
	}
	return n, true
}

// ipv4LastPartLimit bounds the last part of an IPv4 host with 1, 2, 3 or 4
// parts: it fills the remaining 4, 3, 2 or 1 bytes.
var ipv4LastPartLimit = [...]uint64{1 << 32, 1 << 24, 1 << 16, 1 << 8}

// parseIPv4Host is the WHATWG IPv4 parser over a host that ends in a number,
// returning the dotted-decimal serialization.
func parseIPv4Host(host string) (string, error) {
	parts := strings.Split(host, ".")
	if n := len(parts); n > 1 && parts[n-1] == "" {
		parts = parts[:n-1]
	}
	if len(parts) > 4 {
		return "", errors.New("a numeric host must be an IPv4 address of at most four parts")
	}
	numbers := make([]uint64, len(parts))
	for i, part := range parts {
		n, ok := parseIPv4Number(part)
		if !ok {
			return "", errors.New("a numeric host must be an IPv4 address")
		}
		numbers[i] = n
	}
	for _, n := range numbers[:len(numbers)-1] {
		if n > 255 {
			return "", errors.New("an IPv4 part must not exceed 255")
		}
	}
	last := numbers[len(numbers)-1]
	if last >= ipv4LastPartLimit[len(numbers)-1] {
		return "", errors.New("the last IPv4 part is out of range")
	}
	ipv4 := last
	for i, n := range numbers[:len(numbers)-1] {
		ipv4 += n << (8 * (3 - i))
	}
	return fmt.Sprintf("%d.%d.%d.%d", ipv4>>24&0xFF, ipv4>>16&0xFF, ipv4>>8&0xFF, ipv4&0xFF), nil
}

// normalizeIPv6Host parses the text inside the brackets and returns the
// bracketed WHATWG serialization.
func normalizeIPv6Host(inner string) (string, error) {
	addr, err := netip.ParseAddr(inner)
	if err != nil || !addr.Is6() || addr.Zone() != "" {
		return "", errors.New("IPv6 host must be a bracketed address without a zone")
	}
	return "[" + serializeIPv6(addr) + "]", nil
}

// serializeIPv6 is the WHATWG IPv6 serializer: eight 16-bit pieces in
// lowercase hex without leading zeros, the first longest run of two or more
// zero pieces written as "::", and no dotted-decimal tail for an IPv4-mapped
// address (the parser accepts that spelling; the serializer never emits it).
func serializeIPv6(addr netip.Addr) string {
	b := addr.As16()
	var pieces [8]uint16
	for i := range pieces {
		pieces[i] = uint16(b[2*i])<<8 | uint16(b[2*i+1])
	}
	compress, longest := -1, 1
	for i := 0; i < len(pieces); {
		if pieces[i] != 0 {
			i++
			continue
		}
		j := i
		for j < len(pieces) && pieces[j] == 0 {
			j++
		}
		if j-i > longest {
			compress, longest = i, j-i
		}
		i = j
	}
	var out strings.Builder
	ignoreZeros := false
	for i, piece := range pieces {
		if ignoreZeros && piece == 0 {
			continue
		}
		ignoreZeros = false
		if i == compress {
			if i == 0 {
				out.WriteString("::")
			} else {
				out.WriteByte(':')
			}
			ignoreZeros = true
			continue
		}
		out.WriteString(strconv.FormatUint(uint64(piece), 16))
		if i != len(pieces)-1 {
			out.WriteByte(':')
		}
	}
	return out.String()
}

// completeEscapes reports whether every '%' begins a two-hex-digit escape.
func completeEscapes(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			continue
		}
		if i+2 >= len(s) {
			return false
		}
		_, ok1 := hexValue(s[i+1])
		_, ok2 := hexValue(s[i+2])
		if !ok1 || !ok2 {
			return false
		}
		i += 2
	}
	return true
}

// canonicalPathEscapes checks a path segment's escapes (already complete):
// uppercase hex, and never an unreserved character.
func canonicalPathEscapes(segment string) error {
	for i := 0; i < len(segment); i++ {
		if segment[i] != '%' {
			continue
		}
		hi, lo := segment[i+1], segment[i+2]
		if hi >= 'a' && hi <= 'f' || lo >= 'a' && lo <= 'f' {
			return errors.New("path must use uppercase percent escapes")
		}
		h, _ := hexValue(hi)
		l, _ := hexValue(lo)
		if isUnreservedByte(h<<4 | l) {
			return errors.New("path must not percent-encode an unreserved character")
		}
		i += 2
	}
	return nil
}

// validateCanonicalHTTPURL is the shared half of the Scope and Type URL laws:
// the URL's spelled authority must already be its normalized origin, and its
// path must be one the parser would leave alone. echo says whether
// diagnostics may quote the value.
func validateCanonicalHTTPURL(s, what string, echo bool) (urlParts, error) {
	fail := func(msg string) (urlParts, error) {
		if echo {
			return urlParts{}, fmt.Errorf("%w: %s %q %s", ErrValidation, what, s, msg)
		}
		return urlParts{}, fmt.Errorf("%w: %s %s", ErrValidation, what, msg)
	}
	for i := 0; i < len(s); i++ {
		if c := s[i]; c <= 0x20 || c >= 0x7F || c == '\\' {
			return fail("contains a character that is not a canonical URL character")
		}
	}
	p, err := splitHTTPURL(s)
	if err != nil {
		return fail(err.Error())
	}
	if p.scheme != "http" && p.scheme != "https" {
		return fail("must use the http or https scheme, lowercase")
	}
	if p.hasUserinfo {
		return fail("must not carry credentials")
	}
	if p.hasFragment {
		return fail("must not carry a fragment")
	}
	o, err := normalizeOrigin(p)
	if err != nil {
		return fail(err.Error())
	}
	if o.host != p.host {
		return fail("host must be spelled as the URL parser serializes it (lowercase; IPv4 in dotted decimal; IPv6 compressed in lowercase hex)")
	}
	if p.hasPort {
		switch {
		case p.port == "":
			return fail("an empty port must be omitted")
		case o.port == "":
			return fail("a default port must be omitted")
		case o.port != p.port:
			return fail("port must be a decimal number without leading zeros")
		}
	}
	if p.path == "" {
		return fail("must have a path beginning with /")
	}
	if !completeEscapes(s) {
		return fail("must use complete percent escapes")
	}
	for _, segment := range strings.Split(p.path[1:], "/") {
		if segment == "." || segment == ".." {
			return fail("must not contain dot segments")
		}
		if strings.ContainsAny(segment, "\"<>`{}") {
			return fail("path contains a character the URL parser would encode")
		}
		if err := canonicalPathEscapes(segment); err != nil {
			return fail(err.Error())
		}
	}
	if p.hasQuery && strings.ContainsAny(p.query, "\"<>'") {
		return fail("query contains a character the URL parser would encode")
	}
	return p, nil
}

// ValidateTypeURL accepts exactly the canonical, credential-free HTTP(S) URLs
// a Type ID, a propertiesSchema or a discovery member may be. A query is
// permitted; a fragment is not.
func ValidateTypeURL(s string) error {
	_, err := validateCanonicalHTTPURL(s, "URL", true)
	return err
}

// ValidateScopeURL accepts exactly the canonical Scope URLs: a canonical
// HTTP(S) URL with no query, ending in "/", every path segment under the
// ID-segment grammar. It is the rule for any Scope URL a CLIENT may name —
// ParseRef classifies references against it — and it admits a development
// server's reserved "…/local-test/" Scope, because a client must be able to
// reference one. What this store may persist as its own identity is the
// narrower ValidatePersistedScopeURL.
func ValidateScopeURL(s string) error {
	p, err := validateCanonicalHTTPURL(s, "Scope URL", false)
	if err != nil {
		return err
	}
	if p.hasQuery {
		return fmt.Errorf("%w: Scope URL must not carry a query", ErrValidation)
	}
	if !strings.HasSuffix(p.path, "/") {
		return fmt.Errorf("%w: Scope URL must end in /", ErrValidation)
	}
	segments := strings.Split(p.path, "/")
	for _, segment := range segments[1 : len(segments)-1] {
		if err := ValidateCanonicalSegment(segment); err != nil {
			return fmt.Errorf("%w: Scope URL path: %s", ErrValidation, reason(err))
		}
	}
	return nil
}

// localTestSegment is the first path segment the startup contract reserves
// for a derived development identity.
const localTestSegment = "local-test"

// ValidatePersistedScopeURL accepts exactly the Scope URLs this store may
// PERSIST as its own identity — mint under, rotate to, record in the ledger:
// ValidateScopeURL plus the startup contract's reservation of the first path
// segment "local-test" for a derived development identity that is never
// persisted. This tree has no development mode (engdocs/
// BDP_GRAPH_ARCHITECTURE.md §6, "no dev-mode derivation"), so a persisted
// Scope URL never carries the segment; a client referencing another server's
// local-test Scope goes through ValidateScopeURL and is not affected.
//
// DECISION (P0 council): the refusal lives here, at persisted-identity
// admission, and not in the general law, so that a bdptest development
// server's "…/local-test/" Scope stays referenceable.
func ValidatePersistedScopeURL(s string) error {
	if err := ValidateScopeURL(s); err != nil {
		return err
	}
	p, _ := splitHTTPURL(s)
	if strings.HasPrefix(p.path, "/"+localTestSegment+"/") {
		return fmt.Errorf("%w: a persisted Scope URL path must not begin with the reserved local-test segment", ErrValidation)
	}
	return nil
}

// NormalizeScopeURL is the startup contract's admission rule for a configured
// Scope URL: the authority is rewritten to its normalized origin — scheme
// and host case, the host's IPv4/IPv6 spelling and percent-encoding, a
// default port, a port's leading zeros — a missing trailing slash is added,
// and the result must then satisfy ValidatePersistedScopeURL, so
// credentials, a query, a fragment, a path spelling the URL parser would
// rewrite, and the reserved local-test segment are refused rather than
// repaired. The returned string is the persisted identity.
//
// DECISION: the origin is normalized in full rather than only in case. The
// contract names case, the default port and the slash; the parser makes
// :0443, 0x7f000001 and [::ffff:1.2.3.4] the same origin as :443,
// 127.0.0.1 and [::ffff:102:304], and persisting the spelled form would
// leave the persisted identity unequal to the origin every reference is
// classified against.
func NormalizeScopeURL(s string) (string, error) {
	p, err := splitHTTPURL(s)
	if err != nil {
		return "", fmt.Errorf("%w: Scope URL %s", ErrValidation, err)
	}
	if p.hasUserinfo {
		return "", fmt.Errorf("%w: Scope URL must not carry credentials", ErrValidation)
	}
	o, err := normalizeOrigin(p)
	if err != nil {
		return "", fmt.Errorf("%w: Scope URL %s", ErrValidation, err)
	}
	path := p.path
	if path == "" {
		path = "/"
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	var rebuilt strings.Builder
	rebuilt.WriteString(o.scheme)
	rebuilt.WriteString("://")
	rebuilt.WriteString(o.host)
	if o.port != "" {
		rebuilt.WriteString(":")
		rebuilt.WriteString(o.port)
	}
	rebuilt.WriteString(path)
	if p.hasQuery {
		rebuilt.WriteString("?")
		rebuilt.WriteString(p.query)
	}
	if p.hasFragment {
		rebuilt.WriteString("#")
		rebuilt.WriteString(p.fragment)
	}
	out := rebuilt.String()
	if err := ValidatePersistedScopeURL(out); err != nil {
		return "", err
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Reference classification helpers (the law itself is ParseRef in types.go).
// ---------------------------------------------------------------------------

func isSubDelimByte(c byte) bool {
	switch c {
	case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=':
		return true
	}
	return false
}

// isAbsoluteURI is the schema bundle's absoluteUri: a scheme token followed
// by ':' (pattern ^[A-Za-z][A-Za-z0-9+.-]*:), spelled with the RFC 3986
// character discipline — unreserved, reserved, and complete percent escapes,
// nothing else. A full syntactic parse is not attempted: an external
// reference is opaque and never dereferenced.
func isAbsoluteURI(s string) bool {
	i := strings.IndexByte(s, ':')
	if i <= 0 || !isSchemeToken(s[:i]) {
		return false
	}
	for j := 0; j < len(s); j++ {
		c := s[j]
		switch {
		case isUnreservedByte(c), isSubDelimByte(c):
		case c == ':' || c == '/' || c == '?' || c == '#' || c == '[' || c == ']' || c == '@':
		case c == '%':
			if j+2 >= len(s) {
				return false
			}
			_, ok1 := hexValue(s[j+1])
			_, ok2 := hexValue(s[j+2])
			if !ok1 || !ok2 {
				return false
			}
			j += 2
		default:
			return false
		}
	}
	return true
}

// normalizeEscapes decodes escapes of unreserved characters and uppercases
// the rest (RFC 3986 §6.2.2.2); malformed escapes are left as they are.
func normalizeEscapes(path string) string {
	var b strings.Builder
	b.Grow(len(path))
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c != '%' || i+2 >= len(path) {
			b.WriteByte(c)
			continue
		}
		hi, ok1 := hexValue(path[i+1])
		lo, ok2 := hexValue(path[i+2])
		if !ok1 || !ok2 {
			b.WriteByte(c)
			continue
		}
		if v := hi<<4 | lo; isUnreservedByte(v) {
			b.WriteByte(v)
		} else {
			b.WriteByte('%')
			b.WriteByte(upperHexDigits[hi])
			b.WriteByte(upperHexDigits[lo])
		}
		i += 2
	}
	return b.String()
}

// removeDotSegments is RFC 3986 §5.2.4 over an absolute path.
func removeDotSegments(path string) string {
	segments := strings.Split(path, "/")
	out := make([]string, 0, len(segments))
	for i, segment := range segments {
		last := i == len(segments)-1
		switch segment {
		case ".":
			if last {
				out = append(out, "")
			}
		case "..":
			if len(out) > 1 {
				out = out[:len(out)-1]
			}
			if last {
				out = append(out, "")
			}
		default:
			out = append(out, segment)
		}
	}
	return strings.Join(out, "/")
}

// claimsScope reports whether reference resolves under the canonical Scope
// URL — the test that decides whether a reference CLAIMS an in-Scope Bead
// (and must therefore be that Bead's canonical spelling) or is external. The
// origins are compared normalized (normalizeOrigin: scheme and host case,
// every IPv4 and IPv6 spelling, percent-encoded host characters, a default
// or zero-padded port), and the path after RFC 3986 §6.2.2 escape
// normalization and dot-segment removal. A reference whose authority the
// parser cannot make sense of claims nothing and is external. scopeURL is
// assumed valid.
func claimsScope(scopeURL, reference string) bool {
	p, err := splitHTTPURL(reference)
	if err != nil {
		return false
	}
	ref, err := normalizeOrigin(p)
	if err != nil {
		return false
	}
	sp, _ := splitHTTPURL(scopeURL)
	scope, _ := normalizeOrigin(sp) // scopeURL is valid: its origin is its spelling
	if ref != scope {
		return false
	}
	path := p.path
	if path == "" {
		path = "/"
	}
	return strings.HasPrefix(removeDotSegments(normalizeEscapes(path)), sp.path)
}

// ---------------------------------------------------------------------------
// Ledger hashing
//
// engdocs/BDP_GRAPH_CLI_AND_STORAGE_SPEC.md B2/B4: the ledger is append-only
// and hash-chained; hash = sha256(canonical(event without hash)). This is the
// exact byte layout, FROZEN by the P1 migration that stores it:
//
//	canonical(event) = CanonicalizeJSON of a JSON object with these members
//	and no others, absent optional members OMITTED, keys therefore in this
//	(code-unit) order:
//
//	  "at"            string  RFC 3339 UTC with exactly six fractional digits
//	                          and the "Z" designator: 2006-01-02T15:04:05.000000Z
//	                          (the DATETIME(6) column's precision; the value is
//	                          truncated, not rounded, to the microsecond)
//	  "authority_id"  string  32 lowercase hex digits
//	  "epoch"         number  the authority epoch, as a JSON integer
//	  "fingerprint"   string  64 lowercase hex digits         (optional)
//	  "kind"          string  the event kind
//	  "op_id"         string  32 lowercase hex digits
//	  "path"          string  canonical Scope-relative path    (optional)
//	  "prev_hash"     string  64 lowercase hex digits; GenesisHash for the
//	                          first event of a Scope's history
//	  "resource_kind" string  "bead" | "link"                  (optional)
//	  "revision"      string  the revision token               (optional)
//	  "scope_url"     string  canonical Scope URL              (optional)
//	  "seq"           number  the sequence number, as a JSON integer
//	  "state"         string  "pruned" | "erased"              (optional)
//
//	hash = lowercase hex SHA-256 of those bytes; the next event's prev_hash is
//	this hash. Integers are exact (see the number rule above), so seq and
//	epoch survive the full uint64 range.
//
// DECISION: the layout above — member names as the B4 columns spell them,
// JCS for the framing, microsecond UTC for the instant, all-zero genesis. B2
// gives the formula and the member list; the framing and the formats are
// this file's choice and are pinned by a golden test.
// ---------------------------------------------------------------------------

// GenesisHash is the prev_hash of the first event of a Scope's history: no
// event exists before mint, and the chain starts from sixty-four zeros.
const GenesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

const ledgerAtLayout = "2006-01-02T15:04:05.000000Z07:00"

// ledgerWire is the hashed shape; fields are declared in key order so the
// encoder emits them sorted, and CanonicalizeJSON afterwards makes the
// framing canonical whatever the encoder did.
type ledgerWire struct {
	At           string `json:"at"`
	AuthorityID  string `json:"authority_id"`
	Epoch        uint64 `json:"epoch"`
	Fingerprint  string `json:"fingerprint,omitempty"`
	Kind         string `json:"kind"`
	OpID         string `json:"op_id"`
	Path         string `json:"path,omitempty"`
	PrevHash     string `json:"prev_hash"`
	ResourceKind string `json:"resource_kind,omitempty"`
	Revision     string `json:"revision,omitempty"`
	ScopeURL     string `json:"scope_url,omitempty"`
	Seq          uint64 `json:"seq"`
	State        string `json:"state,omitempty"`
}

// ledgerEventCanonicalBytes renders the hashed bytes of a validated spec.
func ledgerEventCanonicalBytes(spec LedgerEventSpec) []byte {
	w := ledgerWire{
		At:           spec.At.UTC().Truncate(time.Microsecond).Format(ledgerAtLayout),
		AuthorityID:  spec.AuthorityID,
		Epoch:        spec.Epoch,
		Fingerprint:  spec.Fingerprint,
		Kind:         string(spec.Kind),
		OpID:         spec.OpID,
		Path:         spec.Path,
		PrevHash:     spec.PrevHash,
		ResourceKind: string(spec.ResourceKind),
		Revision:     spec.Revision.String(),
		ScopeURL:     spec.ScopeURL,
		Seq:          spec.Seq,
		State:        spec.State,
	}
	raw, _ := json.Marshal(w)             // strings and integers: cannot fail
	canonical, _ := CanonicalizeJSON(raw) // encoder output is one valid value
	return canonical
}

// VerifyLedgerChain checks that events is one contiguous, correctly linked run
// of the ledger continuing from prevHash: each event's prev_hash is the
// preceding hash, sequence numbers lie in [1, MaxLedgerSeq] and ascend by
// exactly one, and every hash was computed by NewLedgerEvent (a zero event
// never links). A gap, a fork, a reordering, or a sequence that wraps around
// (MaxUint64 followed by 0) is ErrValidation naming the first offending
// sequence number.
func VerifyLedgerChain(prevHash string, events []LedgerEvent) error {
	if !isLowerHex(prevHash, 64) {
		return fmt.Errorf("%w: ledger chain must continue from a 64-hex-digit hash", ErrValidation)
	}
	prev := prevHash
	for i, e := range events {
		if e.PrevHash() != prev {
			return fmt.Errorf("%w: ledger event seq %d does not link to the preceding hash", ErrValidation, e.Seq())
		}
		if e.Seq() == 0 || e.Seq() > MaxLedgerSeq {
			return fmt.Errorf("%w: ledger event seq %d is outside 1..MaxLedgerSeq", ErrValidation, e.Seq())
		}
		// Compared without addition, so no sequence number can wrap.
		if i > 0 && (e.Seq() <= events[i-1].Seq() || e.Seq()-events[i-1].Seq() != 1) {
			return fmt.Errorf("%w: ledger gap: seq %d follows seq %d", ErrValidation, e.Seq(), events[i-1].Seq())
		}
		prev = e.Hash()
	}
	return nil
}
