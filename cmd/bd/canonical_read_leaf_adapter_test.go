package main

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type canonicalReadProbeValue struct {
	calls int
	text  string
}

func (v *canonicalReadProbeValue) String() string         { return v.text }
func (v *canonicalReadProbeValue) Type() string           { return "bool" }
func (v *canonicalReadProbeValue) Set(value string) error { v.calls++; v.text = value; return nil }

type canonicalReadProbeTree struct {
	root, leaf, link        *cobra.Command
	status, debug           *canonicalReadProbeValue
	originalArgs, hook, run int
	output                  bytes.Buffer
}

// Every test invocation gets fresh Flag/Value/bound state; these tests do not
// execute or reset the production global Cobra tree or register initializers.
func newCanonicalReadProbeTree(path []string) *canonicalReadProbeTree {
	probe := &canonicalReadProbeTree{status: &canonicalReadProbeValue{text: "false"}, debug: &canonicalReadProbeValue{text: "false"}}
	probe.root = &cobra.Command{Use: "bd"}
	probe.root.SetOut(&probe.output)
	probe.root.SetErr(&probe.output)
	probe.root.PersistentFlags().VarP(probe.debug, "debug", "D", "compat debug")
	probe.root.PersistentFlags().Lookup("debug").NoOptDefVal = "true"
	probe.root.PersistentFlags().Bool("json", false, "JSON presentation")
	probe.root.PersistentPreRunE = func(*cobra.Command, []string) error { probe.hook++; return nil }
	probe.leaf = &cobra.Command{Use: path[len(path)-1],
		Args: func(*cobra.Command, []string) error { probe.originalArgs++; return nil },
		RunE: func(*cobra.Command, []string) error { probe.run++; return nil },
	}
	probe.leaf.Flags().VarP(probe.status, "status", "s", "compat status")
	probe.leaf.Flags().Lookup("status").NoOptDefVal = "true"
	probe.leaf.Flags().StringArray("property-field", nil, "fixed properties")
	probe.leaf.Flags().StringArray("ast", nil, "AST request")
	if len(path) == 2 {
		probe.link = &cobra.Command{Use: "link"}
		probe.link.AddCommand(probe.leaf)
		probe.root.AddCommand(probe.link)
	} else {
		probe.root.AddCommand(probe.leaf)
	}
	return probe
}

func executeCanonicalReadProbe(t *testing.T, probe *canonicalReadProbeTree, argv []string) (*canonicalReadLeafDecoration, error) {
	t.Helper()
	decoration, err := prepareCanonicalReadLeaf(probe.root, cliSurfaceCanonical, argv)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(decoration.Restore)
	probe.root.SetArgs(argv)
	_, err = probe.root.ExecuteC()
	return decoration, err
}

func TestCanonicalReadAdapterProhibitedSetNeverRuns(t *testing.T) {
	for _, argv := range [][]string{
		{"list", "--status=false"}, {"list", "--status"}, {"list", "-s=false"}, {"list", "-s"},
		{"--debug=false", "list"}, {"list", "--debug=false"}, {"list", "-D=false"},
		{"list", "--status=false", "--status=true"},
		{"link", "list", "--status=false"},
	} {
		t.Run(argv[len(argv)-1], func(t *testing.T) {
			path := []string{"list"}
			if argv[0] == "link" {
				path = []string{"link", "list"}
			}
			probe := newCanonicalReadProbeTree(path)
			decoration, err := executeCanonicalReadProbe(t, probe, argv)
			refusal := decoration.Refusal()
			if err == nil || refusal == nil || refusal.Code != "invalid_selector" || refusal.ExitCode != 2 {
				t.Fatalf("missing typed refusal: %v %#v", err, refusal)
			}
			if probe.status.calls != 0 || probe.debug.calls != 0 || probe.originalArgs != 0 || probe.hook != 0 || probe.run != 0 {
				t.Fatalf("rejected invocation had effects: %#v", probe)
			}
			if probe.output.Len() != 0 {
				t.Fatalf("Cobra printed duplicate parse diagnostic: %q", probe.output.String())
			}
			if refusal.Message != "--status is not in this canonical operation's surface" && refusal.Message != "--debug is not in this canonical operation's surface" {
				t.Fatalf("unexpected/value-bearing policy diagnostic: %q", refusal.Message)
			}
		})
	}
}

func TestCanonicalReadAdapterAllowedValuesKeepSliceInterfaces(t *testing.T) {
	probe := newCanonicalReadProbeTree([]string{"list"})
	field := probe.leaf.Flags().Lookup("property-field")
	original := field.Value
	if _, ok := original.(pflag.SliceValue); !ok {
		t.Fatal("probe is not a SliceValue")
	}
	decoration, err := executeCanonicalReadProbe(t, probe, []string{"list", "--property-field", "a=1", "--property-field=b=2", "--json=false"})
	if err != nil || decoration.Refusal() != nil {
		t.Fatalf("allowed parser rejected: %v", err)
	}
	if field.Value != original {
		t.Fatal("allowed Value identity replaced")
	}
	got, err := probe.leaf.Flags().GetStringArray("property-field")
	if err != nil || !reflect.DeepEqual(got, []string{"a=1", "b=2"}) {
		t.Fatalf("normal repeated SliceValue/getter changed: %v %v", got, err)
	}
	if probe.originalArgs != 1 || probe.hook != 1 || probe.run != 1 {
		t.Fatalf("normal invocation not delegated exactly once: %#v", probe)
	}
}

func TestCanonicalReadAdapterArgsGatePrecedesOriginalAndHooks(t *testing.T) {
	for _, argv := range [][]string{{"query"}, {"query", "legacy expression"}, {"list", "extra"}, {"link", "list", "a", "b", "c"}} {
		path := []string{argv[0]}
		if argv[0] == "link" {
			path = []string{"link", "list"}
		}
		probe := newCanonicalReadProbeTree(path)
		decoration, err := executeCanonicalReadProbe(t, probe, argv)
		if err == nil || decoration.Refusal() == nil {
			t.Fatalf("missing Args refusal: %v", err)
		}
		if probe.originalArgs != 0 || probe.hook != 0 || probe.run != 0 {
			t.Fatalf("Args gate ran late: %#v", probe)
		}
	}
	probe := newCanonicalReadProbeTree([]string{"query"})
	decoration, err := executeCanonicalReadProbe(t, probe, []string{"query", "--ast=@not-read-by-policy"})
	if err != nil || decoration.Refusal() != nil || probe.originalArgs != 1 || probe.run != 1 {
		t.Fatalf("AST occurrence not delegated to existing executor: %v", err)
	}
	// Fake RunE deliberately performs no file IO. Admission reads no AST data;
	// actual decoding, duplicates and capabilities remain the checked reader's job.
}

func TestCanonicalReadAdapterFindDoesNotRunCustomArgsAndLeafAliasRefuses(t *testing.T) {
	probe := newCanonicalReadProbeTree([]string{"list"})
	probe.leaf.Aliases = []string{"ls"}
	decoration, err := prepareCanonicalReadLeaf(probe.root, cliSurfaceCanonical, []string{"ls"})
	if err != nil {
		t.Fatal(err)
	}
	defer decoration.Restore()
	if probe.originalArgs != 0 || decoration.Command() != probe.leaf {
		t.Fatal("Find ran original Args or changed leaf identity")
	}
	probe.root.SetArgs([]string{"ls"})
	_, err = probe.root.ExecuteC()
	if err == nil || decoration.Refusal() == nil || probe.originalArgs != 0 || probe.hook != 0 || probe.run != 0 {
		t.Fatalf("alias bypassed Args gate: %v", err)
	}
}

func TestCanonicalReadAdapterDecorationRestoresExactReferences(t *testing.T) {
	probe := newCanonicalReadProbeTree([]string{"link", "list"})
	annotations := map[string]string{"preserve": "yes"}
	probe.leaf.Annotations = annotations
	status := probe.leaf.Flags().Lookup("status")
	status.Hidden = false
	originalValue, originalArgs := status.Value, probe.leaf.Args
	originalNoOpt := status.NoOptDefVal
	decoration, err := prepareCanonicalReadLeaf(probe.root, cliSurfaceCanonical, []string{"link", "list"})
	if err != nil {
		t.Fatal(err)
	}
	if !status.Hidden || status.Value == originalValue || status.NoOptDefVal != originalNoOpt {
		t.Fatal("prohibited decoration malformed")
	}
	if status.Value.String() != originalValue.String() || status.Value.Type() != originalValue.Type() {
		t.Fatal("wrapper changed String/Type")
	}
	if probe.leaf.Annotations[canonicalDefaultAnnotation] != "1" || probe.root.Annotations[canonicalDefaultAnnotation] != "" || probe.link.Annotations[canonicalDefaultAnnotation] != "" {
		t.Fatal("annotation escaped exact leaf")
	}
	if !probe.leaf.SilenceErrors || !probe.leaf.SilenceUsage {
		t.Fatal("Cobra diagnostic suppression missing")
	}
	decoration.Restore()
	decoration.Restore()
	if status.Hidden || status.Value != originalValue || probe.leaf.SilenceErrors || probe.leaf.SilenceUsage {
		t.Fatal("flag or silence state not restored")
	}
	annotations["identity"] = "retained"
	if probe.leaf.Annotations["identity"] != "retained" || probe.leaf.Annotations[canonicalDefaultAnnotation] != "" || canonicalCLIProfileSelected(probe.leaf) {
		t.Fatal("original annotation reference not restored")
	}
	if reflect.ValueOf(probe.leaf.Args).Pointer() != reflect.ValueOf(originalArgs).Pointer() {
		t.Fatal("Args closure not restored")
	}
}

func TestCanonicalReadAdapterCompatUntouchedAndCoverageFailsClosed(t *testing.T) {
	probe := newCanonicalReadProbeTree([]string{"list"})
	original := probe.leaf.Flags().Lookup("status").Value
	decoration, err := prepareCanonicalReadLeaf(probe.root, cliSurfaceCompat, []string{"unknown", "--status=false"})
	if err != nil || decoration.Command() != nil || probe.leaf.Flags().Lookup("status").Value != original {
		t.Fatalf("compat was inspected/decorated: %v", err)
	}
	decoration.Restore()
	probe.root.AddCommand(&cobra.Command{Use: "ready", RunE: func(*cobra.Command, []string) error { return nil }})
	probe.root.AddCommand(&cobra.Command{Use: "link"})
	for _, argv := range [][]string{{"ready"}, {"link"}} {
		_, err := prepareCanonicalReadLeaf(probe.root, cliSurfaceCanonical, argv)
		var refusal *canonicalReadLeafRefusal
		if !errors.As(err, &refusal) || refusal.Code != "capability_unavailable" {
			t.Fatalf("unknown canonical surface fell back: %v", err)
		}
	}
	probe.root.TraverseChildren = true
	if _, err := prepareCanonicalReadLeaf(probe.root, cliSurfaceCanonical, []string{"list"}); err == nil {
		t.Fatal("early parent parser route admitted")
	}
	probe.root.TraverseChildren = false
	probe.leaf.DisableFlagParsing = true
	if _, err := prepareCanonicalReadLeaf(probe.root, cliSurfaceCanonical, []string{"list"}); err == nil {
		t.Fatal("disabled parser route admitted")
	}
}
