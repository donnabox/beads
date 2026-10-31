package clirevision

import "testing"

func TestHasLeadingLocatorClassification(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  bool
	}{
		{"https://example.test/item@=saved", true},
		{"https://example.test/item@=arbitrary://token/rev@4", true},
		{"HTTP://[::1]/item", true},
		{"ftp://example.test/item", true},
		{"a+B.c-d://example.test/item", true},
		{"https://", true},
		{"https://%zz/item@=saved", true},
		{"https://[broken/item@=saved", true},
		{"task@=https://token.example/rev", false},
		{"task@=arbitrary://token/rev@4@=next", false},
		{"plan:v1@=https://token.example/rev", false},
		{"namespace:part:task@=arbitrary://token/rev@4", false},
		{"urn:opaque://token", false},
		{"[https://example.test/item]@=saved", false},
		{"1scheme://example.test/item", false},
		{"bad_scheme://example.test/item", false},
		{" https://example.test/item", false},
		{"plan:v1", false}, {"", false}, {"://", false},
	} {
		t.Run(tc.input, func(t *testing.T) {
			if got := HasLeadingLocator(tc.input); got != tc.want {
				t.Fatalf("HasLeadingLocator(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
