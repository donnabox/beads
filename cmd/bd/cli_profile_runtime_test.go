package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Every invocation gets a fresh fake tree. No production command/global flag,
// config file, provider, engine or installed binary is executed by these tests.
type profileRuntimeFixture struct {
	root, list, configGet, configSet, raw, complete *cobra.Command
	out, diag                                       bytes.Buffer
	sets, args, hooks, runs                         int
	value                                           string
	property                                        []string
}

type profileRuntimeEffectValue struct{ f *profileRuntimeFixture }

func (v *profileRuntimeEffectValue) String() string     { return v.f.value }
func (*profileRuntimeEffectValue) Type() string         { return "string" }
func (v *profileRuntimeEffectValue) Set(s string) error { v.f.sets++; v.f.value = s; return nil }

func newProfileRuntimeFixture() *profileRuntimeFixture {
	f := &profileRuntimeFixture{}
	f.root = &cobra.Command{Use: "bd", SilenceErrors: true, SilenceUsage: true}
	f.root.SetOut(&f.out)
	f.root.SetErr(&f.diag)
	f.root.PersistentFlags().Bool("json", false, "JSON output")
	f.root.PersistentFlags().Bool("quiet", false, "quiet")
	f.root.PersistentPreRunE = func(*cobra.Command, []string) error { f.hooks++; return nil }
	f.list = &cobra.Command{Use: "list", Args: func(*cobra.Command, []string) error { f.args++; return nil },
		RunE: func(cmd *cobra.Command, args []string) error {
			f.runs++
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "listed")
			return err
		}}
	f.list.Flags().IntP("limit", "n", 0, "limit")
	f.list.Flags().StringArrayVar(&f.property, "property-field", nil, "properties")
	f.list.Flags().Var(&profileRuntimeEffectValue{f}, "danger", "compat setter with observable effects")
	f.list.Flags().Bool("status", false, "compat status")
	query := &cobra.Command{Use: "query", RunE: func(*cobra.Command, []string) error { f.runs++; return nil }}
	query.Flags().String("ast", "", "AST")
	link := &cobra.Command{Use: "link", Args: cobra.ExactArgs(2), RunE: func(*cobra.Command, []string) error { f.runs++; return nil }}
	link.AddCommand(&cobra.Command{Use: "list", RunE: func(*cobra.Command, []string) error { f.runs++; return nil }})
	config := &cobra.Command{Use: "config"}
	f.configGet = &cobra.Command{Use: "get", Args: cobra.ExactArgs(1), RunE: func(*cobra.Command, []string) error { f.runs++; return nil }}
	f.configSet = &cobra.Command{Use: "set", Args: cobra.ExactArgs(2), RunE: func(*cobra.Command, []string) error { f.runs++; return nil }}
	config.AddCommand(f.configGet, f.configSet, &cobra.Command{Use: "unset", Args: cobra.ExactArgs(1), RunE: func(*cobra.Command, []string) error { f.runs++; return nil }})
	f.raw = &cobra.Command{Use: "raw", DisableFlagParsing: true, RunE: func(*cobra.Command, []string) error { f.runs++; return nil }}
	f.complete = &cobra.Command{Use: "__complete", DisableFlagParsing: true, Hidden: true, RunE: func(*cobra.Command, []string) error { f.runs++; return nil }}
	f.root.AddCommand(f.list, query, link, config, f.raw, f.complete,
		&cobra.Command{Use: "delete", RunE: func(*cobra.Command, []string) error { f.runs++; return nil }},
		&cobra.Command{Use: "uncovered", Args: func(*cobra.Command, []string) error { f.args++; return nil },
			RunE: func(*cobra.Command, []string) error { f.runs++; return nil }},
		&cobra.Command{Use: "version", RunE: func(*cobra.Command, []string) error { f.runs++; return nil }})
	return f
}

func runtimeProfile(profile cliSurfaceProfile) func(*string) (cliProfileSelection, error) {
	return func(override *string) (cliProfileSelection, error) {
		if override != nil {
			return selectCLIProfile(nil, override)
		}
		return cliProfileSelection{Profile: profile, Source: cliProfileSaved}, nil
	}
}

func TestCLIProfileRuntimeCompatPreservesOriginalRoute(t *testing.T) {
	for _, argv := range [][]string{{"list", "--danger", "x"}, {"--limit=5", "list"}, {"-n5", "list"}, {"config", "get", "cli.profile"}, {"raw", "--unknown", "--cli-profile", "data"}} {
		t.Run(strings.Join(argv, " "), func(t *testing.T) {
			f := newProfileRuntimeFixture()
			_, err := executeCLIProfileRootWith(f.root, argv, runtimeProfile(cliSurfaceCompat))
			if err != nil {
				t.Fatal(err)
			}
			if f.runs != 1 || f.hooks != 1 {
				t.Fatalf("original compat route changed: runs=%d hooks=%d", f.runs, f.hooks)
			}
			if argv[0] == "list" && (f.sets != 1 || f.value != "x" || f.args != 1) {
				t.Fatalf("ordinary parser/Args changed: %+v", f)
			}
			if canonicalOutputDefault(f.list) {
				t.Fatal("compat leaf was decorated")
			}
		})
	}
}

func TestCLIProfileRuntimeCanonicalRejectsBeforeEffects(t *testing.T) {
	// Retained delete is covered. Use a registered, genuinely uncovered leaf
	// to exercise refusal before either execution or help can reach effects.
	if _, covered := lookupCanonicalReadLeaf([]string{"uncovered"}); covered {
		t.Fatal("refusal fixture must stay outside the finite canonical surface")
	}
	for _, argv := range [][]string{{"list", "--danger=x"}, {"uncovered"}, {"uncovered", "--help"}, {"raw", "--help"}, {"query"}, {"list", "--status=false"}} {
		t.Run(strings.Join(argv, " "), func(t *testing.T) {
			f := newProfileRuntimeFixture()
			_, err := executeCLIProfileRootWith(f.root, argv, runtimeProfile(cliSurfaceCanonical))
			if err == nil {
				t.Fatal("expected checked refusal")
			}
			if argv[0] == "uncovered" {
				if code, ok := exitCodeFromError(err); !ok || code != 5 || strings.Count(f.diag.String(), "capability_unavailable:") != 1 {
					t.Fatalf("uncovered leaf lost its capability refusal: err=%v diagnostic=%q", err, f.diag.String())
				}
			}
			if f.sets != 0 || f.args != 0 || f.hooks != 0 || f.runs != 0 {
				t.Fatalf("refused operation had effects: sets=%d args=%d hooks=%d runs=%d", f.sets, f.args, f.hooks, f.runs)
			}
			if strings.Count(f.diag.String(), "invalid_selector:")+strings.Count(f.diag.String(), "capability_unavailable:") != 1 {
				t.Fatalf("diagnostic was missing/duplicated: %q", f.diag.String())
			}
			if strings.Contains(f.diag.String(), "danger=x") {
				t.Fatal("rejected value leaked through pflag wrapper")
			}
		})
	}
}

// This fake executor checks routing only, not production delete admission.
func TestCLIProfileRuntimeCanonicalRetainedDeleteRoute(t *testing.T) {
	descriptor, covered := lookupCanonicalReadLeaf([]string{"delete"})
	if !covered || descriptor.Operation != canonicalReadLeafOperation("retained-singleton") || strings.Join(descriptor.Path, " ") != "delete" {
		t.Fatalf("retained delete descriptor unavailable: %#v %v", descriptor, covered)
	}
	f := newProfileRuntimeFixture()
	want, _, err := f.root.Find([]string{"delete"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := executeCLIProfileRootWith(f.root, []string{"delete"}, runtimeProfile(cliSurfaceCanonical))
	if err != nil || got != want || f.runs != 1 || f.hooks != 1 || f.args != 0 || f.sets != 0 || f.diag.Len() != 0 {
		t.Fatalf("retained delete route changed: err=%v same=%v runs=%d hooks=%d args=%d sets=%d diagnostic=%q", err, got == want, f.runs, f.hooks, f.args, f.sets, f.diag.String())
	}
	if canonicalOutputDefault(want) {
		t.Fatal("retained delete decoration was not restored")
	}
}

func TestCLIProfileRuntimeCanonicalAllowedValuesAndOverride(t *testing.T) {
	f := newProfileRuntimeFixture()
	originalValue := f.list.Flags().Lookup("property-field").Value
	loaderCalls := 0
	loader := func(override *string) (cliProfileSelection, error) {
		loaderCalls++
		if override == nil || *override != "canonical" {
			return cliProfileSelection{}, errors.New("missing one-call override")
		}
		return selectCLIProfile(nil, override)
	}
	_, err := executeCLIProfileRootWith(f.root, []string{"--cli-profile=canonical", "list", "--property-field", "x=1", "--property-field=x=2"}, loader)
	if err != nil {
		t.Fatal(err)
	}
	if loaderCalls != 1 || f.runs != 1 || f.hooks != 1 || f.args != 1 {
		t.Fatalf("wrong route counts: loader=%d runs=%d hooks=%d args=%d", loaderCalls, f.runs, f.hooks, f.args)
	}
	if strings.Join(f.property, ",") != "x=1,x=2" {
		t.Fatalf("repeated flags changed: %q", f.property)
	}
	if f.list.Flags().Lookup("property-field").Value != originalValue {
		t.Fatal("allowed Value identity changed")
	}
	if _, ok := originalValue.(pflag.SliceValue); !ok {
		t.Fatal("allowed slice interface lost")
	}
	if canonicalOutputDefault(f.list) {
		t.Fatal("leaf annotation was not restored")
	}
}

func TestCLIProfileRuntimeBrokenPreferenceRepairIsBounded(t *testing.T) {
	broken := func(*string) (cliProfileSelection, error) {
		return cliProfileSelection{}, errors.New("saved profile is invalid")
	}
	for _, argv := range [][]string{{"config", "get", "cli.profile"}, {"config", "set", "cli.profile", "compat"}, {"config", "unset", "cli.profile"}, {"version"}} {
		t.Run(strings.Join(argv, " "), func(t *testing.T) {
			f := newProfileRuntimeFixture()
			_, err := executeCLIProfileRootWith(f.root, argv, broken)
			if err != nil || f.runs != 1 || f.hooks != 0 {
				t.Fatalf("repair unavailable or operation hooks ran: err=%v runs=%d hooks=%d", err, f.runs, f.hooks)
			}
		})
	}
	for _, argv := range [][]string{{"list"}, {"config", "set", "other.key", "compat"}, {"delete"}, {"raw"}} {
		t.Run("refuse "+strings.Join(argv, " "), func(t *testing.T) {
			f := newProfileRuntimeFixture()
			_, err := executeCLIProfileRootWith(f.root, argv, broken)
			if err == nil || f.runs != 0 || f.hooks != 0 {
				t.Fatalf("broken preference enabled ordinary execution: err=%v", err)
			}
		})
	}
}

func TestCLIProfileRuntimeCanonicalHelpAndCompletionBoundaries(t *testing.T) {
	for _, argv := range [][]string{{"--help"}, {"list", "--help"}, {"delete", "--help"}, {"link"}, {"link", "--help"}, {"config", "--help"}, {"config", "get", "--help"}} {
		t.Run(strings.Join(argv, " "), func(t *testing.T) {
			f := newProfileRuntimeFixture()
			_, err := executeCLIProfileRootWith(f.root, argv, runtimeProfile(cliSurfaceCanonical))
			if err != nil || f.runs != 0 || f.hooks != 0 {
				t.Fatalf("help acquired/executed: err=%v runs=%d hooks=%d", err, f.runs, f.hooks)
			}
			if f.out.Len() == 0 || strings.Contains(f.out.String(), "--danger") || strings.Contains(f.out.String(), "--status") || strings.Contains(f.out.String(), "completion") {
				t.Fatalf("canonical help was not projected: %q", f.out.String())
			}
			if f.list.Flags().Lookup("danger").Hidden {
				t.Fatal("help projection not restored")
			}
		})
	}
	f := newProfileRuntimeFixture()
	_, err := executeCLIProfileRootWith(f.root, []string{"__complete", "list", ""}, runtimeProfile(cliSurfaceCanonical))
	if code, ok := exitCodeFromError(err); !ok || code != 5 || f.runs != 0 || f.hooks != 0 {
		t.Fatalf("canonical completion silently used compat: err=%v", err)
	}
}

func TestCLIProfileRuntimeLateSelectorDoesNotChangeProfile(t *testing.T) {
	f := newProfileRuntimeFixture()
	_, err := executeCLIProfileRootWith(f.root, []string{"list", "--cli-profile=canonical"}, runtimeProfile(cliSurfaceCompat))
	if err == nil || f.runs != 0 || f.hooks != 0 || f.args != 0 {
		t.Fatalf("late selector executed: err=%v", err)
	}
}
