//go:build cgo

package main

import "testing"

func TestGraphPreviewLinkNamespaceRegistration(t *testing.T) {
	for _, name := range []string{"create", "list", "show", "update", "delete", "versions", "compare", "count"} {
		if cmd, _, err := linkCmd.Find([]string{name}); err != nil || cmd == linkCmd || cmd.Name() != name {
			t.Fatalf("bd link %s is not registered: %v", name, err)
		}
	}
	for _, name := range []string{"links", "unlink"} {
		if cmd, _, err := rootCmd.Find([]string{name}); err == nil && cmd != rootCmd && cmd.Name() == name {
			t.Fatalf("preview-only bd %s remains registered", name)
		}
	}
}

func TestGraphPreviewLifecycleLegacyRefusal(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, args := range [][]string{
		{"link", "list", "--incident-to", "beads/plan"},
		{"link", "delete", "links/context", "--force"},
	} {
		work, home := t.TempDir(), t.TempDir()
		before := legacyUpgradeTreeDigest(t, work)
		graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", append(args, "--json")...)
		if after := legacyUpgradeTreeDigest(t, work); after != before {
			t.Fatalf("legacy refusal changed workspace: %v", args)
		}
	}
}
