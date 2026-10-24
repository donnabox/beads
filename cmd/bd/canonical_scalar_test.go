package main

import (
	"github.com/spf13/cobra"
	"testing"
)

func TestCanonicalScalarOccurrences(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		bad  bool
	}{
		{"different revisions", []string{"show", "b", "--revision=a", "--revision=b"}, true},
		{"same revisions", []string{"show", "b", "--revision=a", "--revision=a"}, false},
		{"octal differs", []string{"list", "--limit=010", "--limit=10"}, true},
		{"hex same", []string{"list", "--limit=0x10", "--limit=16"}, false},
		{"octal same", []string{"list", "--limit=010", "--limit=8"}, false},
		{"bool same", []string{"list", "--paged=true", "--paged=1"}, false},
		{"bool differs", []string{"list", "--paged=true", "--paged=false"}, true},
		{"after separator", []string{"show", "--revision=a", "--", "--revision=b"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("revision", "", "")
			cmd.Flags().Int("limit", 0, "")
			cmd.Flags().Bool("paged", false, "")
			err := canonicalScalarOccurrences(cmd, test.args, "revision", "limit", "paged")
			if (err != nil) != test.bad {
				t.Fatalf("args=%v err=%v want refusal=%v", test.args, err, test.bad)
			}
		})
	}
}
