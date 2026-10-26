// Package clirevision handles CLI-only revision selectors. Ordinals are local
// history aliases; callers must resolve them before constructing stored refs.
package clirevision

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Selector describes an adorned reference's suffix, without the @ delimiter.
// Exactly one of Current, Ordinal, or Token is populated after Parse succeeds.
type Selector struct {
	Current bool
	Ordinal int64
	Token   string
}

// Parse accepts current, a positive decimal history ordinal, or =TOKEN. The
// exact-token escape preserves token bytes, including all-numeric tokens and
// the token "current". Bare opaque tokens are not part of this new CLI syntax.
// Explicit token flags remain literal and must not be passed through Parse.
func Parse(suffix string, tokenByteLimit int) (Selector, error) {
	if !utf8.ValidString(suffix) || suffix == "" {
		return Selector{}, fmt.Errorf("revision selector must be nonempty UTF-8")
	}
	if suffix == "current" {
		return Selector{Current: true}, nil
	}
	if token, exact := strings.CutPrefix(suffix, "="); exact {
		if token == "" || len(token) > tokenByteLimit {
			return Selector{}, fmt.Errorf("exact revision token must contain 1 to %d bytes", tokenByteLimit)
		}
		return Selector{Token: token}, nil
	}
	for _, c := range suffix {
		if c < '0' || c > '9' {
			return Selector{}, fmt.Errorf("use @N for a local revision, @current, or @=TOKEN for an exact token")
		}
	}
	n, err := strconv.ParseInt(suffix, 10, 64)
	if err != nil || n <= 0 {
		return Selector{}, fmt.Errorf("local revision must be a positive 64-bit decimal number")
	}
	return Selector{Ordinal: n}, nil
}
