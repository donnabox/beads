package clirevision

import (
	"net/url"
	"strings"
)

// HasLeadingLocator reports only whether input starts with a valid URI scheme
// immediately followed by ://. It reuses net/url's scheme grammar, without
// validating or normalizing the remaining locator. Malformed host/path bytes
// must stay on the caller's URL-admission path, not become a local selector.
// A local ID or exact token containing an earlier colon or :// is not enough:
// the scheme must occupy the entire prefix before the chosen delimiter.
func HasLeadingLocator(input string) bool {
	delimiter := strings.Index(input, "://")
	if delimiter <= 0 {
		return false
	}
	prefix, err := url.Parse(input[:delimiter+3])
	return err == nil && prefix.Scheme != "" && len(prefix.Scheme) == delimiter
}
