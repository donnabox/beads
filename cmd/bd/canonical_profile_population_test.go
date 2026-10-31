package main

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestCanonicalCLIProfileCollectionPopulation(t *testing.T) {
	scope := canonicalCollectionScope
	issueType, memoryType := graphstore.IssueTypeURL(scope), scope+"memory"
	// Closed Issue content and Memory content both belong to canonical List's
	// unfiltered current population. These are detached inventory records;
	// no provider, native store, process or global command tree is executed.
	inventory := graphread.Inventory{
		Beads: []bdpwire.BeadRecord{
			{ID: scope + "i", Type: issueType, Revision: "r1", Properties: bdpwire.Properties{"status": json.RawMessage(`"closed"`)}, Metadata: bdpwire.Metadata{}},
			{ID: scope + "m", Type: memoryType, Revision: "r1", Properties: bdpwire.Properties{"body": json.RawMessage(`"remembered"`)}, Metadata: bdpwire.Metadata{}},
		},
		Types: []bdpwire.TypeDescriptor{
			{ID: issueType, Name: "Issue", Describes: bdpwire.DescribesBead},
			{ID: memoryType, Name: "Memory", Describes: bdpwire.DescribesBead},
		},
	}
	for _, tc := range []struct {
		name                         string
		canonical, outputDefaultOnly bool
		flags                        []string
		want                         []string
	}{
		{name: "compat explicit Resource output keeps Issue default", want: []string{scope + "i"}},
		{name: "output default alone does not select semantic profile", outputDefaultOnly: true, want: []string{scope + "i"}},
		{name: "canonical bare includes Issue and Memory", canonical: true, want: []string{scope + "i", scope + "m"}},
		{name: "canonical explicit Issue narrows", canonical: true, flags: []string{"--type=issue"}, want: []string{scope + "i"}},
		{name: "canonical explicit exact Memory narrows", canonical: true, flags: []string{"--exact-type=" + memoryType}, want: []string{scope + "m"}},
		{name: "canonical any remains all", canonical: true, flags: []string{"--type=any"}, want: []string{scope + "i", scope + "m"}},
		{name: "canonical count uses same all-current selection", canonical: true, flags: []string{"--count-only"}, want: []string{scope + "i", scope + "m"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"--output-format=json", "--unpaged"}, tc.flags...)
			cmd, _, _ := canonicalCollectionTestCommand(t, false, args...)
			// Admission requires an executable leaf; this detached callback is
			// never run by the selection test. Preserve the production guard.
			cmd.RunE = func(*cobra.Command, []string) error { return nil }
			if tc.canonical {
				root := &cobra.Command{Use: "bd"}
				root.AddCommand(cmd)
				decoration, err := prepareCanonicalReadLeaf(root, cliSurfaceCanonical, []string{"list"})
				if err != nil {
					t.Fatal(err)
				}
				defer decoration.Restore()
			} else if tc.outputDefaultOnly {
				cmd.Annotations = map[string]string{canonicalDefaultAnnotation: "1"}
			}
			plan, _ := canonicalCollectionTestPlan(t, cmd, false)
			records, err := plan.selection.Select(inventory)
			if err != nil {
				t.Fatal(err)
			}
			if len(records) != len(tc.want) {
				t.Fatalf("population=%s want IDs=%q", records, tc.want)
			}
			got := map[string]bool{}
			for _, raw := range records {
				var record struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(raw, &record); err != nil {
					t.Fatal(err)
				}
				got[record.ID] = true
			}
			for _, id := range tc.want {
				if !got[id] {
					t.Fatalf("missing selected identity %s from %s", id, records)
				}
			}
		})
	}
}
