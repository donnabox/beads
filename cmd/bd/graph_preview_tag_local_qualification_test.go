//go:build cgo

package main

import (
	"testing"

	"github.com/steveyegge/beads/internal/testutil"
)

// This local qualification wrapper is removed after the both-engine run.
func TestGraphPreviewTagLocalQualification(t *testing.T) {
	server := testutil.StartIsolatedDoltContainerHandle(t)
	t.Setenv("BEADS_GRAPH_TEST_SERVER_PORT", server.Port)
	TestGraphPreviewIssueQuickActionsWorkflow(t)
}
