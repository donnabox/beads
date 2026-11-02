package main

import (
	"github.com/spf13/cobra"
	"testing"
)

func TestCommandOutputFormatSelectionPreservesLocalPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name, root, local, want string
		shadow, changed         bool
	}{
		{name: "absent"},
		{name: "root alias", root: "json", want: "json", changed: true},
		{name: "local alias", shadow: true, local: "JSON", want: "JSON", changed: true},
		{name: "local format wins", shadow: true, root: "json", local: "graph-json", want: "graph-json", changed: true},
		{name: "unchanged shadow retains root alias", shadow: true, root: "json", want: "json", changed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := &cobra.Command{Use: "root"}
			root.PersistentFlags().String("format", "", "")
			child := &cobra.Command{Use: "show"}
			root.AddCommand(child)
			if tc.shadow {
				child.Flags().String("format", "", "")
			}
			if tc.root != "" {
				if err := root.PersistentFlags().Set("format", tc.root); err != nil {
					t.Fatal(err)
				}
			}
			if tc.local != "" {
				if err := child.Flags().Set("format", tc.local); err != nil {
					t.Fatal(err)
				}
			}
			if got := commandOutputFormatChanged(child); got != tc.changed {
				t.Fatalf("explicit=%v,want%v", got, tc.changed)
			}
			if got := commandOutputFormat(child); got != tc.want {
				t.Fatalf("format=%q,want%q", got, tc.want)
			}
		})
	}
}
