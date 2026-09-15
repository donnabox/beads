package graphsession

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var errIdentity = errors.New("graph session: invalid target")

// endpoint is intentionally private and has no production admission factory.
// These fields cannot prove managed deployment provenance; see package docs.
type endpoint struct{ address, user, password, base, branch string }

func (e endpoint) validate() error {
	host, port, err := net.SplitHostPort(e.address)
	if err != nil {
		return errIdentity
	}
	ip := net.ParseIP(host)
	n, err := strconv.Atoi(port)
	if ip == nil || !ip.IsLoopback() || err != nil || n < 1 || n > 65535 {
		return errIdentity
	}
	if !validName(e.base) || strings.ContainsAny(e.base, "/@") || e.base != strings.ToLower(e.base) || !validName(e.branch) || strings.ContainsAny(e.branch, "@") {
		return errIdentity
	}
	return nil
}

func validName(s string) bool {
	return s != "" && len(s) <= 256 && utf8.ValidString(s) && !strings.ContainsFunc(s, unicode.IsControl)
}

// Mirrors the pinned Dolt provider's lowercase base-map key and first revision
// delimiter. This is a readback consistency check, not physical-root evidence.
func splitDatabase(s string) (string, string) {
	if i := strings.IndexAny(s, "/@"); i >= 0 {
		return strings.ToLower(s[:i]), s[i+1:]
	}
	return strings.ToLower(s), ""
}

func lockName(base string) string {
	sum := sha256.Sum256([]byte("bd.graph.e1\x00" + base))
	return "bd.graph.e1:" + base64.RawURLEncoding.EncodeToString(sum[:])
}

// The pinned Dolt commit spelling is exactly 32 base32 digits [0-9a-v].
// This excludes mutable refs and procedure options, but does not prove that
// evidence selected this commit or that the commit exists.
func validCommitOperand(s string) bool {
	if len(s) != 32 {
		return false
	}
	for i := range s {
		if (s[i] < '0' || s[i] > '9') && (s[i] < 'a' || s[i] > 'v') {
			return false
		}
	}
	return true
}
