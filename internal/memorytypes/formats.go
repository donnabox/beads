package memorytypes

import (
	"encoding/json"
	"net/netip"
	"regexp"
	"strings"
	"time"
)

// Require an explicit timezone, valid calendar date, and decimal fraction
// (not Go time.Parse's extra comma syntax). No interval semantics are implied.
var timestampShape = regexp.MustCompile(`^([0-9]{4}-[0-9]{2}-[0-9]{2})[Tt]([0-9]{2}:[0-9]{2}:[0-9]{2})(\.[0-9]+)?([Zz]|[+-](?:[01][0-9]|2[0-3]):[0-5][0-9])$`)

func instant(raw json.RawMessage, path string) error {
	text, err := textValue(raw, path, false)
	if err != nil {
		return err
	}
	parts := timestampShape.FindStringSubmatch(text)
	if parts == nil {
		return invalid(path, "an RFC3339 date-time")
	}
	// Only format validation is needed: arbitrary fractional digits are already
	// checked above, so parsing whole seconds does not lose comparison precision.
	seconds, err := time.Parse(time.RFC3339, parts[1]+"T"+parts[2]+strings.ToUpper(parts[4]))
	if err != nil || seconds.Year() == 0 {
		return invalid(path, "an RFC3339 date-time")
	}
	return nil
}

func applicability(raw json.RawMessage) error {
	const path = "/properties/applicability"
	obj, err := object(raw, path)
	if err != nil {
		return err
	}
	if err := closed(obj, path, "description"); err != nil {
		return err
	}
	return textField(obj, "description", path, true, true)
}

func referenceValue(raw json.RawMessage, path string) error {
	var text string
	if len(raw) > 0 && json.Unmarshal(raw, &text) == nil && string(raw) != "null" {
		return absoluteURI(text, path)
	}
	obj, err := object(raw, path)
	if err != nil {
		return invalid(path, "a URI string or {uri, revision} reference")
	}
	if err := closed(obj, path, "uri", "revision"); err != nil {
		return err
	}
	uri, err := textValue(obj["uri"], path+"/uri", false)
	if err != nil {
		return err
	}
	if err := absoluteURI(uri, path+"/uri"); err != nil {
		return err
	}
	revision, err := textValue(obj["revision"], path+"/revision", false)
	if err != nil {
		return err
	}
	if revision == "" {
		return invalid(path+"/revision", "a nonempty revision string")
	}
	return nil
}

var uriScheme = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)
var ipvFutureLiteral = regexp.MustCompile(`^[vV][0-9A-Fa-f]+\.[A-Za-z0-9._~!$&'()*+,;=:-]+$`)

func absoluteURI(text, path string) error {
	if !validURI(text) {
		return invalid(path, "an absolute URI")
	}
	return nil
}

func validURI(text string) bool {
	scheme := uriScheme.FindString(text)
	if scheme == "" {
		return false
	}
	rest := text[len(scheme):]
	// URI, not IRI. Fragment and query allow pchar plus slash and question mark.
	if i := strings.IndexByte(rest, '#'); i >= 0 {
		if !uriChars(rest[i+1:], ":@/?") {
			return false
		}
		rest = rest[:i]
	}
	if i := strings.IndexByte(rest, '?'); i >= 0 {
		if !uriChars(rest[i+1:], ":@/?") {
			return false
		}
		rest = rest[:i]
	}
	if strings.HasPrefix(rest, "//") {
		authority := rest[2:]
		rest = ""
		if i := strings.IndexByte(authority, '/'); i >= 0 {
			authority, rest = authority[:i], authority[i:]
		}
		if !uriAuthority(authority) {
			return false
		}
	}
	return uriChars(rest, ":@/")
}

func uriAuthority(authority string) bool {
	if i := strings.LastIndexByte(authority, '@'); i >= 0 {
		if !uriChars(authority[:i], ":") {
			return false
		}
		authority = authority[i+1:]
	}
	if strings.HasPrefix(authority, "[") {
		i := strings.IndexByte(authority, ']')
		if i < 0 {
			return false
		}
		host := authority[1:i]
		addr, err := netip.ParseAddr(host)
		if !(err == nil && addr.Is6() && addr.Zone() == "") && !ipvFutureLiteral.MatchString(host) {
			return false
		}
		return uriPort(authority[i+1:])
	}
	if i := strings.IndexByte(authority, ':'); i >= 0 {
		return uriChars(authority[:i], "") && uriPort(authority[i:])
	}
	return uriChars(authority, "")
}

func uriPort(s string) bool {
	if s == "" {
		return true
	}
	if s[0] != ':' {
		return false
	}
	for _, r := range s[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// RFC3986 unreserved, sub-delims, and percent-encoded octets, plus the supplied
// component delimiters. This avoids net/url's scheme-specific restrictions and
// its treatment of spaces, escaped host names, and IPvFuture literals.
func uriChars(s, extra string) bool {
	const basic = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~!$&'()*+,;="
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			if i+2 >= len(s) || !hexDigit(s[i+1]) || !hexDigit(s[i+2]) {
				return false
			}
			i += 2
		} else if !strings.ContainsRune(basic+extra, rune(s[i])) {
			return false
		}
	}
	return true
}

func hexDigit(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F'
}
