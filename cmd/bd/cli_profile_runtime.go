package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/steveyegge/beads/internal/config"
)

// One process, one initialized tree and one checked argv. The existing package
// init config read remains earlier; this gate precedes operation Args/hooks,
// provider/store acquisition and AST input execution. It is not a claim that
// package initialization or routing configuration performs no filesystem IO.
func executeCLIProfileRoot(root *cobra.Command, argv []string) (*cobra.Command, error) {
	return executeCLIProfileRootWith(root, argv, loadCLIProfileForInvocation)
}

type cliProfileRuntimeView struct {
	args     []string
	values   map[string]string
	explicit []string
	help     bool
	version  bool
}

// Inspect only argument shapes with private values; original Value.Set/Args
// are never invoked by this inspection. Raw forwarding is deliberately not
// inspected. Unknown/private partial grammar is not permission to execute.
func inspectCLIProfileInvocation(command *cobra.Command, args []string) (cliProfileRuntimeView, error) {
	view := cliProfileRuntimeView{values: make(map[string]string)}
	command.InitDefaultHelpFlag()
	shapes, err := snapshotCLIBootstrapArity(command.Flags(), command.InheritedFlags())
	if err != nil {
		return view, err
	}
	flags := pflag.NewFlagSet("profile control inspection", pflag.ContinueOnError)
	flags.SetOutput(io.Discard)
	for _, shape := range shapes {
		flags.AddFlag(&pflag.Flag{Name: shape.Name, Shorthand: shape.Shorthand,
			NoOptDefVal: shape.NoOptDefVal, Value: &cliBootstrapValue{}})
	}
	err = flags.ParseAll(append([]string(nil), args...), func(flag *pflag.Flag, value string) error {
		if flag.Name == cliProfilePrefixFlag {
			return (&cliProfilePrefixOnlyValue{}).Set(value)
		}
		view.values[flag.Name] = value
		view.explicit = append(view.explicit, flag.Name)
		return nil
	})
	if err != nil {
		return view, err
	}
	view.args = append([]string(nil), flags.Args()...)
	for name, destination := range map[string]*bool{"help": &view.help, "version": &view.version} {
		if value, set := view.values[name]; set {
			*destination, err = strconv.ParseBool(value)
			if err != nil {
				return view, err
			}
		}
	}
	return view, nil
}

type cliProfileControl struct {
	command    *cobra.Command
	target     *cobra.Command
	view       cliProfileRuntimeView
	name       string
	safeRepair bool
}

func classifyCLIProfileControl(root, command *cobra.Command, args []string) (*cliProfileControl, error) {
	if command == nil || command.DisableFlagParsing {
		return nil, nil
	}
	view, err := inspectCLIProfileInvocation(command, args)
	if err != nil {
		return nil, err // No real flag setter has run.
	}
	path := canonicalReadCommandPath(command, root)
	control := &cliProfileControl{command: command, target: command, view: view, safeRepair: true}
	switch {
	case command == root:
		control.name = "root"
	case len(path) == 1 && path[0] == "help":
		control.name = "help"
		if view.values["docs-root"] != "" {
			control.safeRepair = false // This help option writes files.
		}
	case len(path) == 1 && path[0] == "version":
		control.name = "version"
	case len(path) == 2 && path[0] == "config" &&
		(path[1] == "get" || path[1] == "set" || path[1] == "unset") &&
		len(view.args) != 0 && view.args[0] == config.CLIProfileKey:
		control.name = "profile-config"
	case len(path) == 1 && (path[0] == "link" || path[0] == "config") && len(view.args) == 0:
		control.name = "group-help"
	case view.help:
		control.name = "leaf-help"
	default:
		return nil, nil
	}
	return control, nil
}

func cliProfileRuntimeRefusal(command *cobra.Command, err error) error {
	if isReportedExit(err) {
		return err
	}
	var refusal *canonicalReadLeafRefusal
	if errors.As(err, &refusal) {
		return canonicalFailure(command, canonicalFailureRendererFor(command), refusal.Code, refusal.Message, refusal.ExitCode)
	}
	return canonicalFailure(command, canonicalFailureRendererFor(command), "invalid_selector", err.Error(), 2)
}

func cliProfileUnavailable(command *cobra.Command, message string) error {
	return canonicalFailure(command, nil, "capability_unavailable", message, 5)
}

func executeCLIProfileRootWith(root *cobra.Command, argv []string,
	load func(*string) (cliProfileSelection, error)) (*cobra.Command, error) {
	root.InitDefaultHelpCmd()
	root.InitDefaultHelpFlag()
	shapes, err := snapshotCLIBootstrapArity(root.PersistentFlags(), root.Flags())
	if err != nil {
		return root, cliProfileRuntimeRefusal(root, err)
	}
	prefix, decodeErr := decodeCLIProfilePrefix(argv, shapes)
	if decodeErr != nil {
		if len(prefix.ProfileValues) != 0 {
			return root, cliProfileRuntimeRefusal(root, decodeErr)
		}
		// Preserve ordinary parsing/diagnostics when a partial root parser
		// cannot decide a preexisting invocation. This does not switch profile.
		prefix = cliProfilePrefix{OriginalArgs: append([]string(nil), argv...),
			ExecutionArgs: append([]string(nil), argv...), RemainingArgs: append([]string(nil), argv...),
			DeferredToCobra: true}
	}
	if err := registerCLIProfilePrefixOnlyGuard(root.PersistentFlags()); err != nil {
		return root, cliProfileRuntimeRefusal(root, err)
	}
	// Completion is a wrapped command line, not a normal operation. Recognize
	// an explicit selector in its decodable payload instead of silently using
	// the saved profile. Unknown local arity remains for the real parser/guard.
	if prefix.CompletionPayload {
		payload, payloadErr := decodeCLIProfilePrefix(prefix.RemainingArgs[1:], shapes)
		if len(payload.ProfileValues) != 0 {
			if payloadErr != nil {
				return root, cliProfileRuntimeRefusal(root, payloadErr)
			}
			return root, cliProfileUnavailable(root, "profile selection inside shell-completion payloads is not implemented")
		}
	}
	selection, profileErr := load(prefix.Override)
	if profileErr == nil && selection.Profile != cliSurfaceCompat && selection.Profile != cliSurfaceCanonical {
		profileErr = errors.New("unresolved CLI surface profile")
	}
	if profileErr == nil && selection.Profile == cliSurfaceCompat {
		// Do not inspect/decorate compat leaves or suppress their hooks. The
		// original tree receives the exact historical argv when no new prefix
		// selector was supplied; only an explicit selector is consumed.
		root.SetArgs(append([]string{}, prefix.ExecutionArgs...))
		return root.ExecuteC()
	}
	command, commandArgs, findErr := root.Find(prefix.ExecutionArgs)
	var control *cliProfileControl
	if findErr == nil {
		control, err = classifyCLIProfileControl(root, command, commandArgs)
		if err != nil {
			return command, cliProfileRuntimeRefusal(command, err)
		}
	}
	if profileErr != nil && (control == nil || !control.safeRepair) {
		return root, cliProfileRuntimeRefusal(root, profileErr)
	}
	// A broken saved preference never enables ordinary compat execution.
	// Only the checked help/version/profile-key repair controls remain usable.
	repairOnly := profileErr != nil
	canonical := !repairOnly && selection.Profile == cliSurfaceCanonical
	if prefix.CompletionPayload {
		return root, cliProfileUnavailable(root, "canonical shell completion is not implemented")
	}
	if canonical && findErr != nil {
		return root, cliProfileRuntimeRefusal(root, findErr)
	}
	if (canonical || repairOnly) && control != nil {
		if err := admitCLIProfileControl(root, control); err != nil {
			return control.command, cliProfileRuntimeRefusal(control.command, err)
		}
	}
	if canonical {
		oldErrors, oldUsage := root.SilenceErrors, root.SilenceUsage
		oldCompletion := root.CompletionOptions.DisableDefaultCmd
		root.SilenceErrors, root.SilenceUsage = true, true
		// ExecuteC lazily adds its public completion command after our help
		// projection. Do not advertise that unavailable canonical operation.
		root.CompletionOptions.DisableDefaultCmd = true
		defer func() {
			root.SilenceErrors, root.SilenceUsage = oldErrors, oldUsage
			root.CompletionOptions.DisableDefaultCmd = oldCompletion
		}()
		restore := projectCLIProfileHelp(root)
		defer restore()
	}
	if control != nil {
		// Control handlers execute through the SAME tree and checked argv.
		// Suppress inherited operation setup/epilogues, not their intended
		// help/version/config result or the main cleanup/error-exit path.
		restore := isolateCLIProfileControl(control)
		defer restore()
		root.SetArgs(append([]string{}, prefix.ExecutionArgs...))
		executed, err := root.ExecuteC()
		if canonical && err != nil {
			return executed, cliProfileRuntimeRefusal(control.command, err)
		}
		return executed, err
	}
	savedRootContext := root.Context()
	baseContext := savedRootContext
	if baseContext == nil {
		baseContext = context.Background()
	}
	root.SetContext(context.WithValue(baseContext, canonicalRetainedArgvKey, append([]string(nil), prefix.ExecutionArgs...)))
	defer root.SetContext(savedRootContext)
	decoration, err := prepareCanonicalReadLeaf(root, selection.Profile, prefix.ExecutionArgs)
	if err != nil {
		return root, cliProfileRuntimeRefusal(root, err)
	}
	defer decoration.Restore()
	command = decoration.Command()
	if command == nil {
		return root, cliProfileUnavailable(root, "canonical command admission did not resolve a covered operation")
	}
	savedCommandContext := command.Context()
	commandBaseContext := savedCommandContext
	if commandBaseContext == nil {
		commandBaseContext = root.Context()
	}
	command.SetContext(context.WithValue(commandBaseContext, canonicalRetainedArgvKey, append([]string(nil), prefix.ExecutionArgs...)))
	defer command.SetContext(savedCommandContext)
	argsPassed := false
	originalArgs := command.Args
	command.Args = func(actual *cobra.Command, args []string) error {
		if originalArgs != nil {
			if err := originalArgs(actual, args); err != nil {
				return err
			}
		}
		argsPassed = true
		return nil
	}
	defer func() { command.Args = originalArgs }()
	root.SetArgs(append([]string{}, prefix.ExecutionArgs...))
	executed, err := root.ExecuteC()
	if err == nil || isReportedExit(err) {
		return executed, err
	}
	if refusal := decoration.Refusal(); refusal != nil {
		return command, cliProfileRuntimeRefusal(command, refusal)
	}
	if !argsPassed {
		return command, cliProfileRuntimeRefusal(command, err)
	}
	return command, canonicalReadError(command, canonicalFailureRendererFor(command), err, "capability_unavailable")
}

func admitCLIProfileControl(root *cobra.Command, control *cliProfileControl) error {
	refuse := func(message string) error {
		return &canonicalReadLeafRefusal{Code: "capability_unavailable", ExitCode: 5, Message: message}
	}
	allowed := map[string]bool{"help": true, "version": true, "json": true, "quiet": true, "no-color": true}
	switch control.name {
	case "root", "version", "profile-config", "group-help":
	case "help":
		allowed["all"], allowed["list"] = true, true
		if len(control.view.args) != 0 {
			target, _, err := root.Find(control.view.args)
			if err != nil {
				return err
			}
			path := canonicalReadCommandPath(target, root)
			if !cliProfileVisiblePath(path) || strings.Join(path, " ") != strings.Join(control.view.args, " ") {
				return refuse("requested help topic is outside the initial canonical surface")
			}
		}
	case "leaf-help":
		descriptor, covered := lookupCanonicalReadLeaf(canonicalReadCommandPath(control.command, root))
		if !covered && !cliProfileVisiblePath(canonicalReadCommandPath(control.command, root)) {
			return refuse("requested command is outside the initial canonical List/Link List/Query coverage")
		}
		for _, name := range descriptor.Flags {
			allowed[name] = true
		}
	default:
		return refuse("control route is not in the canonical surface")
	}
	for _, name := range control.view.explicit {
		if !allowed[name] {
			return &canonicalReadLeafRefusal{Code: "invalid_selector", ExitCode: 2,
				Message: fmt.Sprintf("--%s is not in this canonical control surface", name)}
		}
	}
	return nil
}

func cliProfileVisiblePath(path []string) bool {
	if _, ok := lookupCanonicalReadLeaf(path); ok {
		return true
	}
	switch strings.Join(path, " ") {
	case "", "link", "config", "config get", "config set", "config unset", "help", "version":
		return true
	}
	return false
}

// Presentation projection never grants an operation. Hidden flags still have
// admission wrappers on covered reads; unsupported paths are refused pre-ExecuteC.
func projectCLIProfileHelp(root *cobra.Command) func() {
	var undo []func()
	var walk func(*cobra.Command)
	walk = func(command *cobra.Command) {
		path := canonicalReadCommandPath(command, root)
		oldHidden, oldLong, oldExample, oldAliases := command.Hidden, command.Long, command.Example, command.Aliases
		oldUse, oldShort := command.Use, command.Short
		undo = append(undo, func() {
			command.Hidden, command.Long, command.Example, command.Aliases = oldHidden, oldLong, oldExample, oldAliases
			command.Use, command.Short = oldUse, oldShort
		})
		command.Hidden = !cliProfileVisiblePath(path)
		command.Aliases = nil
		allowed := map[string]bool{"help": true, "version": true, "json": true, "quiet": true, "no-color": true, "directory": true, cliProfilePrefixFlag: true}
		if descriptor, ok := lookupCanonicalReadLeaf(path); ok {
			for _, name := range descriptor.Flags {
				allowed[name] = true
			}
			command.Long, command.Example = "Read Resources using the canonical request and output contract.", ""
			command.Short = "Read canonical Resources"
			if descriptor.RequiresAST {
				command.Use, command.Short = "query --ast JSON", "Evaluate a canonical JSON Query AST"
			}
		}
		if command == root {
			command.Long = "Canonical CLI profile. Current operation coverage: List, Link List and JSON-AST Query, plus help/version/profile configuration controls."
			for _, p := range [][]string{{"list"}, {"link", "list"}, {"query"}} {
				d, _ := lookupCanonicalReadLeaf(p)
				for _, name := range d.Flags {
					allowed[name] = true
				}
			}
		}
		if strings.Join(path, " ") == "config" {
			command.Long, command.Short = "Read or change the user-global CLI profile preference.", "Configure the CLI profile"
		}
		if len(path) == 2 && path[0] == "config" {
			command.Long, command.Example = "This canonical configuration control operates on the user-global CLI profile preference only.", ""
			command.Use = path[1] + " " + config.CLIProfileKey
			if path[1] == "set" {
				command.Use += " canonical|compat"
			}
		}
		if len(path) == 1 && path[0] == "help" {
			allowed["all"], allowed["list"] = true, true
		}
		command.LocalNonPersistentFlags().VisitAll(func(flag *pflag.Flag) {
			old := flag.Hidden
			undo = append(undo, func() { flag.Hidden = old })
			flag.Hidden = !allowed[flag.Name]
		})
		command.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			old := flag.Hidden
			undo = append(undo, func() { flag.Hidden = old })
			flag.Hidden = !allowed[flag.Name]
		})
		for _, child := range command.Commands() {
			walk(child)
		}
	}
	walk(root)
	return func() {
		for i := len(undo) - 1; i >= 0; i-- {
			undo[i]()
		}
	}
}

func isolateCLIProfileControl(control *cliProfileControl) func() {
	var undo []func()
	leaf := control.command
	localPre, localPreE, localPost, localPostE := leaf.PreRun, leaf.PreRunE, leaf.PostRun, leaf.PostRunE
	undo = append(undo, func() {
		leaf.PreRun, leaf.PreRunE, leaf.PostRun, leaf.PostRunE = localPre, localPreE, localPost, localPostE
	})
	leaf.PreRun, leaf.PreRunE, leaf.PostRun, leaf.PostRunE = nil, nil, nil, nil
	for c := control.command; c != nil; c = c.Parent() {
		command := c
		pre, preE, post, postE := c.PersistentPreRun, c.PersistentPreRunE, c.PersistentPostRun, c.PersistentPostRunE
		undo = append(undo, func() {
			command.PersistentPreRun, command.PersistentPreRunE, command.PersistentPostRun, command.PersistentPostRunE = pre, preE, post, postE
		})
		c.PersistentPreRun, c.PersistentPreRunE, c.PersistentPostRun, c.PersistentPostRunE = nil, nil, nil, nil
	}
	if control.name == "group-help" {
		command := control.command
		run, runE, args := command.Run, command.RunE, command.Args
		undo = append(undo, func() { command.Run, command.RunE, command.Args = run, runE, args })
		command.Args = cobra.NoArgs
		command.Run = nil
		command.RunE = func(actual *cobra.Command, args []string) error { return actual.Help() }
	}
	return func() {
		for i := len(undo) - 1; i >= 0; i-- {
			undo[i]()
		}
	}
}
