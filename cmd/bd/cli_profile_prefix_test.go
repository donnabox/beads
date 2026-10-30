package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func cliProfilePrefixTestArity() []cliBootstrapArity {
	return []cliBootstrapArity{
		{Name: "directory", Shorthand: "C"},
		{Name: "actor"},
		{Name: "quiet", Shorthand: "q", NoOptDefVal: "true"},
		{Name: "verbose", Shorthand: "v", NoOptDefVal: "true"},
		{Name: "version", Shorthand: "V", NoOptDefVal: "true"},
		{Name: "custom", NoOptDefVal: "auto"},
	}
}

func TestCLIProfilePrefixCases(t *testing.T) {
	for _, tc := range []struct {
		name       string
		argv       []string
		execution  []string
		remaining  []string
		override   string
		values     []string
		terminated bool
		completion bool
		deferred   bool
		wantErr    string
	}{
		{name: "empty"},
		{name: "plain command", argv: []string{"list"}, execution: []string{"list"}, remaining: []string{"list"}},
		{name: "separate selector", argv: []string{"--cli-profile", "canonical", "list"}, execution: []string{"list"}, remaining: []string{"list"}, override: "canonical", values: []string{"canonical"}},
		{name: "equals selector", argv: []string{"--cli-profile=compat", "list"}, execution: []string{"list"}, remaining: []string{"list"}, override: "compat", values: []string{"compat"}},
		{name: "repeated scalar last wins proposal", argv: []string{"--cli-profile", "canonical", "--cli-profile=compat", "list"}, execution: []string{"list"}, remaining: []string{"list"}, override: "compat", values: []string{"canonical", "compat"}},
		{name: "root value resembles selector", argv: []string{"--actor", "--cli-profile", "list"}, execution: []string{"--actor", "--cli-profile", "list"}, remaining: []string{"list"}},
		{name: "repeated root values retain order", argv: []string{"--actor", "first", "--actor=second", "list"}, execution: []string{"--actor", "first", "--actor=second", "list"}, remaining: []string{"list"}},
		{name: "short directory attached", argv: []string{"-Cother", "list"}, execution: []string{"-Cother", "list"}, remaining: []string{"list"}},
		{name: "short directory separate", argv: []string{"-C", "other", "list"}, execution: []string{"-C", "other", "list"}, remaining: []string{"list"}},
		{name: "long directory separate", argv: []string{"--directory", "other", "list"}, execution: []string{"--directory", "other", "list"}, remaining: []string{"list"}},
		{name: "long directory equals", argv: []string{"--directory=other", "list"}, execution: []string{"--directory=other", "list"}, remaining: []string{"list"}},
		{name: "short bool cluster", argv: []string{"-vqCother", "list"}, execution: []string{"-vqCother", "list"}, remaining: []string{"list"}},
		{name: "custom optional arity", argv: []string{"--custom", "list"}, execution: []string{"--custom", "list"}, remaining: []string{"list"}},
		{name: "root boolean explicit false", argv: []string{"--quiet=false", "list"}, execution: []string{"--quiet=false", "list"}, remaining: []string{"list"}},
		{name: "option termination", argv: []string{"--", "--cli-profile", "canonical"}, execution: []string{"--", "--cli-profile", "canonical"}, remaining: []string{"--cli-profile", "canonical"}, terminated: true},
		{name: "termination after selector", argv: []string{"--cli-profile", "canonical", "--", "list"}, execution: []string{"--", "list"}, remaining: []string{"list"}, override: "canonical", values: []string{"canonical"}, terminated: true},
		{name: "suffix directory preserved unresolved", argv: []string{"memory", "recall", "-C", "other"}, execution: []string{"memory", "recall", "-C", "other"}, remaining: []string{"memory", "recall", "-C", "other"}},
		{name: "service value resembles directory", argv: []string{"memory", "recall", "--title", "-C"}, execution: []string{"memory", "recall", "--title", "-C"}, remaining: []string{"memory", "recall", "--title", "-C"}},
		{name: "service value resembles selector", argv: []string{"list", "--title", "--cli-profile"}, execution: []string{"list", "--title", "--cli-profile"}, remaining: []string{"list", "--title", "--cli-profile"}},
		{name: "late selector preserved for normal guard", argv: []string{"list", "--cli-profile", "canonical"}, execution: []string{"list", "--cli-profile", "canonical"}, remaining: []string{"list", "--cli-profile", "canonical"}},
		{name: "help prefix", argv: []string{"-h", "--cli-profile", "canonical"}, execution: []string{"--help=true"}, override: "canonical", values: []string{"canonical"}},
		{name: "help command", argv: []string{"help", "list"}, execution: []string{"help", "list"}, remaining: []string{"help", "list"}},
		{name: "completion needs separate preparation", argv: []string{"__complete", "--cli-profile", "canonical", "list", ""}, execution: []string{"__complete", "--cli-profile", "canonical", "list", ""}, remaining: []string{"__complete", "--cli-profile", "canonical", "list", ""}, completion: true},
		{name: "no description completion", argv: []string{"__completeNoDesc", "list", ""}, execution: []string{"__completeNoDesc", "list", ""}, remaining: []string{"__completeNoDesc", "list", ""}, completion: true},
		{name: "unknown prefix deferred", argv: []string{"--unknown", "list"}, execution: []string{"--unknown", "list"}, remaining: []string{"--unknown", "list"}, deferred: true},
		{name: "long local equals prefix deferred", argv: []string{"--limit=5", "list"}, execution: []string{"--limit=5", "list"}, remaining: []string{"--limit=5", "list"}, deferred: true},
		{name: "long local separate prefix deferred", argv: []string{"--limit", "5", "list"}, execution: []string{"--limit", "5", "list"}, remaining: []string{"--limit", "5", "list"}, deferred: true},
		{name: "short local prefix deferred", argv: []string{"-n5", "list"}, execution: []string{"-n5", "list"}, remaining: []string{"-n5", "list"}, deferred: true},
		{name: "local prefix string value remains data", argv: []string{"--title", "--cli-profile", "list"}, execution: []string{"--title", "--cli-profile", "list"}, remaining: []string{"--title", "--cli-profile", "list"}, deferred: true},
		{name: "explicit opt-in rejects local prefix", argv: []string{"--cli-profile", "canonical", "--limit=5", "list"}, wantErr: "use only known root options"},
		{name: "missing selector value", argv: []string{"--cli-profile"}, wantErr: "needs an argument"},
		{name: "invalid selector value", argv: []string{"--cli-profile=unknown", "list"}, wantErr: "invalid CLI surface profile"},
		{name: "explicit empty selector", argv: []string{"--cli-profile=", "list"}, wantErr: "invalid CLI surface profile"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeCLIProfilePrefix(tc.argv, cliProfilePrefixTestArity())
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("decode error = %v; want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.OriginalArgs, tc.argv) || !reflect.DeepEqual(got.ExecutionArgs, tc.execution) || !reflect.DeepEqual(got.RemainingArgs, tc.remaining) || !reflect.DeepEqual(got.ProfileValues, tc.values) {
				t.Fatalf("unexpected argv projection: %#v", got)
			}
			if got.Terminated != tc.terminated || got.CompletionPayload != tc.completion || got.DeferredToCobra != tc.deferred {
				t.Fatalf("wrong termination/completion markers: %#v", got)
			}
			if tc.override == "" && got.Override != nil || tc.override != "" && (got.Override == nil || *got.Override != tc.override) {
				t.Fatalf("override = %v; want %q", got.Override, tc.override)
			}
		})
	}
}

type cliBootstrapProbeValue struct{ calls int }

func (v *cliBootstrapProbeValue) Set(string) error { v.calls++; return nil }
func (*cliBootstrapProbeValue) String() string     { return "untouched" }
func (*cliBootstrapProbeValue) Type() string       { return "custom" }

func TestCLIProfilePrefixDoesNotMutateSource(t *testing.T) {
	root := pflag.NewFlagSet("source", pflag.ContinueOnError)
	value := &cliBootstrapProbeValue{}
	root.VarP(value, "custom", "x", "")
	root.Lookup("custom").NoOptDefVal = "auto"
	shapes, err := snapshotCLIBootstrapArity(root)
	if err != nil {
		t.Fatal(err)
	}
	argv := []string{"-x", "--cli-profile", "canonical", "list"}
	before := append([]string(nil), argv...)
	got, err := decodeCLIProfilePrefix(argv, shapes)
	if err != nil {
		t.Fatal(err)
	}
	if root.Parsed() || root.Lookup("custom").Changed || value.calls != 0 || !reflect.DeepEqual(argv, before) {
		t.Fatal("bootstrap mutated source flags, values or argv")
	}
	if !reflect.DeepEqual(got.ExecutionArgs, []string{"--custom=auto", "list"}) {
		t.Fatalf("custom arity replay = %#v", got.ExecutionArgs)
	}
	if len(got.Occurrences) != 2 || got.Occurrences[0] != (cliBootstrapOccurrence{Name: "custom", Value: "auto"}) {
		t.Fatalf("custom optional arity was not decoded: %#v", got.Occurrences)
	}
	got.ExecutionArgs[0] = "changed"
	if !reflect.DeepEqual(got.OriginalArgs, before) || !reflect.DeepEqual(argv, before) {
		t.Fatal("returned argv slices share mutable backing storage")
	}
}

func TestCLIProfileLateSelectorGuard(t *testing.T) {
	for _, tc := range []struct {
		argv    []string
		wantErr bool
	}{
		{argv: []string{"--cli-profile", "canonical", "list"}},
		{argv: []string{"list", "--cli-profile=canonical"}, wantErr: true},
		{argv: []string{"--cli-profile", "canonical", "list", "--cli-profile", "compat"}, wantErr: true},
		{argv: []string{"list", "--title", "--cli-profile"}},
		{argv: []string{"list", "--", "--cli-profile", "canonical"}},
	} {
		got, err := decodeCLIProfilePrefix(tc.argv, cliProfilePrefixTestArity())
		if err != nil {
			t.Fatal(err)
		}
		selected := pflag.NewFlagSet("selected", pflag.ContinueOnError)
		selected.String("title", "", "")
		if err := registerCLIProfilePrefixOnlyGuard(selected); err != nil {
			t.Fatal(err)
		}
		err = selected.Parse(got.ExecutionArgs)
		if (err != nil) != tc.wantErr {
			t.Fatalf("selected parse %q = %v; want error %v", tc.argv, err, tc.wantErr)
		}
	}
}

func TestCLIProfilePrefixNoOverrideKeepsExactArgv(t *testing.T) {
	for _, argv := range [][]string{
		nil,
		{"-qv", "-Cother", "list"},
		{"-C", "other", "--directory=third", "list"},
		{"--", "--cli-profile", "canonical"},
		{"--actor", "--cli-profile", "list"},
		{"-qv", "mail", "send", "-C", "other", "--cli-profile", "canonical"},
		{"memory", "recall", "-C", "other", "--title", "--cli-profile"},
		{"", "--cli-profile", "canonical"},
	} {
		original := append([]string(nil), argv...)
		got, err := decodeCLIProfilePrefix(argv, cliProfilePrefixTestArity())
		if err != nil {
			t.Fatal(err)
		}
		if got.Override != nil || !reflect.DeepEqual(got.ExecutionArgs, original) || !reflect.DeepEqual(got.OriginalArgs, original) {
			t.Fatalf("no-selector argv changed: %#v -> %#v", argv, got)
		}
		if len(got.ExecutionArgs) != 0 {
			got.ExecutionArgs[0] = "changed"
			if !reflect.DeepEqual(argv, original) || !reflect.DeepEqual(got.OriginalArgs, original) {
				t.Fatal("execution argv aliases original input or saved original")
			}
		}
	}
}

func TestCLIProfilePrefixDefersLocalArityWithoutScanning(t *testing.T) {
	for _, tc := range []struct {
		argv      []string
		wantLimit int
		wantTitle string
		wantErr   bool
	}{
		{argv: []string{"--limit=5", "list"}, wantLimit: 5},
		{argv: []string{"--limit", "5", "list"}, wantLimit: 5},
		{argv: []string{"-n5", "list"}, wantLimit: 5},
		{argv: []string{"--unknown=5", "list"}, wantErr: true},
		{argv: []string{"--title", "--cli-profile", "list"}, wantTitle: "--cli-profile"},
		// Bootstrap must not scan past the unknown local arity to discover
		// a possible profile selector. The real parser's guard rejects it.
		{argv: []string{"--limit=5", "--cli-profile=canonical", "list"}, wantErr: true},
	} {
		original := append([]string(nil), tc.argv...)
		got, err := decodeCLIProfilePrefix(tc.argv, cliProfilePrefixTestArity())
		if err != nil {
			t.Fatal(err)
		}
		if !got.DeferredToCobra || got.Override != nil || len(got.PrefixArgs) != 0 || len(got.Occurrences) != 0 || !reflect.DeepEqual(got.ExecutionArgs, original) || !reflect.DeepEqual(tc.argv, original) {
			t.Fatalf("unsafe partial-prefix authority or argv change: %#v", got)
		}
		ordinary := pflag.NewFlagSet("ordinary leaf parser", pflag.ContinueOnError)
		ordinary.IntP("limit", "n", 0, "")
		ordinary.String("title", "", "")
		if err := registerCLIProfilePrefixOnlyGuard(ordinary); err != nil {
			t.Fatal(err)
		}
		err = ordinary.Parse(got.ExecutionArgs)
		if (err != nil) != tc.wantErr {
			t.Fatalf("ordinary parse %q = %v; want error %v", tc.argv, err, tc.wantErr)
		}
		if !tc.wantErr {
			limit, _ := ordinary.GetInt("limit")
			title, _ := ordinary.GetString("title")
			if limit != tc.wantLimit || title != tc.wantTitle {
				t.Fatalf("ordinary values changed: limit=%d title=%q", limit, title)
			}
		}
	}
}
