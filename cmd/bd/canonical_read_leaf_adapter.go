package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Semantic profile selection is distinct from output-default decoration.
// Compat can explicitly request canonical Resource formatting without opting
// into the canonical CLI population/default contract.
const cliSurfaceProfileAnnotation = "bd:cli_surface_profile"

func canonicalCLIProfileSelected(command *cobra.Command) bool {
	return command != nil && command.Annotations[cliSurfaceProfileAnnotation] == string(cliSurfaceCanonical)
}

// Decorates one existing initialized leaf for one process invocation. It does
// not own bootstrap/control-plane routing, parse argv, or acquire resources.
// ExecuteC must receive the same argv after any checked outer prefix removal.
// Call Restore after ExecuteC (including error/help paths). Repeated execution
// tests must additionally restore ordinary parser values/custom state or use
// fresh trees; Restore only undoes this adapter's decorations.
type canonicalReadLeafDecoration struct {
	command *cobra.Command
	restore func()
	refusal *canonicalReadLeafRefusal
}

func (d *canonicalReadLeafDecoration) Command() *cobra.Command            { return d.command }
func (d *canonicalReadLeafDecoration) Refusal() *canonicalReadLeafRefusal { return d.refusal }
func (d *canonicalReadLeafDecoration) Restore() {
	if d.restore != nil {
		d.restore()
		d.restore = nil
	}
}

// pflag wraps Value.Set errors (the pinned version also supports Unwrap), but
// its outer diagnostic includes the supplied value. Keep the original name-only
// canonical refusal for the outer renderer; never classify by matching text.
type canonicalReadRejectingValue struct {
	original pflag.Value
	name     string
	reject   func(*canonicalReadLeafRefusal)
}

func (v *canonicalReadRejectingValue) String() string { return v.original.String() }
func (v *canonicalReadRejectingValue) Type() string   { return v.original.Type() }
func (v *canonicalReadRejectingValue) Set(string) error {
	err := &canonicalReadLeafRefusal{Code: "invalid_selector", ExitCode: 2,
		Message: fmt.Sprintf("--%s is not in this canonical operation's surface", v.name)}
	v.reject(err)
	return err // deliberately never call original.Set, even for false/default
}

type canonicalReadFlagDecoration struct {
	flag   *pflag.Flag
	value  pflag.Value
	hidden bool
}

func canonicalReadScopeFlags(command *cobra.Command) []*pflag.Flag {
	seen := make(map[*pflag.Flag]bool)
	var flags []*pflag.Flag
	collect := func(flag *pflag.Flag) {
		if !seen[flag] {
			seen[flag] = true
			flags = append(flags, flag)
		}
	}
	command.Flags().VisitAll(collect)
	command.InheritedFlags().VisitAll(collect)
	command.PersistentFlags().VisitAll(collect)
	return flags
}

func canonicalReadCommandPath(command, root *cobra.Command) []string {
	var reversed []string
	for c := command; c != nil && c != root; c = c.Parent() {
		reversed = append(reversed, c.Name())
	}
	path := make([]string, len(reversed))
	for i := range reversed {
		path[len(reversed)-1-i] = reversed[i]
	}
	return path
}

// Find uses Cobra's own metadata-based resolution and does not invoke custom
// Args (it may run Cobra's built-in legacyArgs when Args is nil). No second raw
// argv parser is introduced. The existing default TraverseChildren=false and
// normal flag parsing are required; Traverse would parse parent setters early.
//
// Help/version/completion/config repair are separate outer routes. In
// particular Cobra can add default help/version flags later and bypass Args.
// Production OnInitialize callbacks must stay free of operation acquisition;
// Cobra executes them before Args. Current cmd/bd has none outside tests.
// Existing package-init/profile-routing config reads are a separate earlier
// phase; this adapter does not claim zero IO before main or profile selection.
func prepareCanonicalReadLeaf(root *cobra.Command, profile cliSurfaceProfile, argv []string) (*canonicalReadLeafDecoration, error) {
	decoration := &canonicalReadLeafDecoration{}
	if profile == cliSurfaceCompat {
		return decoration, nil
	}
	if profile != cliSurfaceCanonical {
		return nil, &canonicalReadLeafRefusal{Code: "invalid_selector", ExitCode: 2, Message: "unresolved CLI surface profile"}
	}
	if root == nil || root.Root() != root || root.TraverseChildren {
		return nil, &canonicalReadLeafRefusal{Code: "capability_unavailable", ExitCode: 5, Message: "canonical read decoration requires the normal root Find route"}
	}
	command, _, err := root.Find(argv)
	if err != nil {
		return nil, err
	} // retain ordinary Cobra resolution errors
	path := canonicalReadCommandPath(command, root)
	descriptor, covered := lookupCanonicalReadLeaf(path)
	if !covered || !command.Runnable() || command.DisableFlagParsing {
		return nil, &canonicalReadLeafRefusal{Code: "capability_unavailable", ExitCode: 5, Message: "canonical operation is outside the initial runnable List/Link List/Query coverage"}
	}
	decoration.command = command
	allowed := make(map[string]bool, len(descriptor.Flags))
	for _, name := range descriptor.Flags {
		allowed[name] = true
	}
	record := func(err *canonicalReadLeafRefusal) {
		if decoration.refusal == nil {
			decoration.refusal = err
		}
	}
	var flags []canonicalReadFlagDecoration
	for _, flag := range canonicalReadScopeFlags(command) {
		if allowed[flag.Name] {
			continue
		} // preserve Value identity and richer interfaces
		flags = append(flags, canonicalReadFlagDecoration{flag: flag, value: flag.Value, hidden: flag.Hidden})
		flag.Value = &canonicalReadRejectingValue{original: flag.Value, name: flag.Name, reject: record}
		flag.Hidden = true
		// NoOptDefVal, shorthand, defaults, annotations and all other flag
		// metadata are unchanged; pflag retains its own syntax semantics.
	}
	originalArgs := command.Args
	originalSilenceErrors, originalSilenceUsage := command.SilenceErrors, command.SilenceUsage
	command.SilenceErrors, command.SilenceUsage = true, true
	originalAnnotations := command.Annotations
	annotations := make(map[string]string, len(originalAnnotations)+1)
	for key, value := range originalAnnotations {
		annotations[key] = value
	}
	annotations[canonicalDefaultAnnotation] = "1"
	annotations[cliSurfaceProfileAnnotation] = string(cliSurfaceCanonical)
	command.Annotations = annotations // leaf only: never root/link ancestor
	command.Args = func(actual *cobra.Command, args []string) error {
		var explicit []canonicalReadExplicitFlag
		for _, flag := range canonicalReadScopeFlags(actual) {
			if flag.Changed {
				explicit = append(explicit, canonicalReadExplicitFlag{Name: flag.Name, Shorthand: flag.Shorthand, Inherited: actual.InheritedFlags().Lookup(flag.Name) == flag})
			}
		}
		calledAs := actual.CalledAs()
		_, err := admitCanonicalReadLeaf(canonicalReadLeafInvocation{
			Profile: cliSurfaceCanonical, Path: canonicalReadCommandPath(actual, root),
			AliasUsed: calledAs != "" && calledAs != actual.Name(),
			Flags:     explicit, Positionals: len(args),
		})
		if err != nil {
			if refusal, ok := err.(*canonicalReadLeafRefusal); ok {
				record(refusal)
			}
			return err
		}
		if originalArgs != nil {
			return originalArgs(actual, args)
		}
		return nil
	}
	decoration.restore = func() {
		command.Args = originalArgs
		command.SilenceErrors, command.SilenceUsage = originalSilenceErrors, originalSilenceUsage
		command.Annotations = originalAnnotations
		for _, original := range flags {
			original.flag.Value = original.value
			original.flag.Hidden = original.hidden
		}
	}
	return decoration, nil
}
