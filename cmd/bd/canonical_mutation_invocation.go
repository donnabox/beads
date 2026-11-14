package main

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"
)

type canonicalMutationInvocationKey struct{}

type canonicalMutationInvocation struct {
	command *cobra.Command
	args []string
}

// The outer canonical decorator binds once BEFORE ExecuteC/early admission and
// defers restore immediately. Existing leaf context wins, matching Cobra. A
// nil leaf context requires the actual nonnil execution context supplied by the
// root owner: manufacturing Background here would block later root inheritance.
// The original context, including nil, is restored on every exit, not PostRun.
func bindCanonicalMutationInvocationArgs(cmd *cobra.Command, execution context.Context, argv []string) (func(), error) {
	if cmd == nil {
		return nil, canonicalMutationLeafUnavailable("canonical invocation requires a command")
	}
	original := cmd.Context()
	parent := original
	if parent == nil {
		parent = execution
	}
	if parent == nil {
		return nil, canonicalMutationLeafUnavailable("canonical invocation requires its actual execution context before binding")
	}
	args := append([]string{}, argv...)
	cmd.SetContext(context.WithValue(parent, canonicalMutationInvocationKey{}, canonicalMutationInvocation{command: cmd, args: args}))
	restored := false
	return func() {
		if !restored {
			cmd.SetContext(original)
			restored = true
		}
	}, nil
}

// Compatibility keeps the historical process argv only when genuinely unbound.
// A bound empty invocation is authoritative; a malformed or inherited binding
// cannot grant permission to fall back to unrelated process arguments.
func canonicalMutationInvocationArgs(cmd *cobra.Command) ([]string, error) {
	return canonicalMutationInvocationArgsFrom(cmd, os.Args[1:])
}

func canonicalMutationInvocationArgsFrom(cmd *cobra.Command, processArgs []string) ([]string, error) {
	if cmd == nil {
		return nil, canonicalMutationLeafUnavailable("canonical invocation requires a command")
	}
	if ctx := cmd.Context(); ctx != nil {
		if value := ctx.Value(canonicalMutationInvocationKey{}); value != nil {
			bound, ok := value.(canonicalMutationInvocation)
			if !ok || bound.command != cmd {
				return nil, canonicalMutationLeafUnavailable("canonical invocation arguments belong to a different or invalid command")
			}
			return append([]string{}, bound.args...), nil
		}
	}
	return append([]string{}, processArgs...), nil
}

func canonicalMutationArgumentFailure(err error) error {
	var refusal *canonicalMutationLeafRefusal
	if errors.As(err, &refusal) {
		return graphFailure(refusal.Code, refusal.Message, refusal.ExitCode)
	}
	return graphFailure("invalid_selector", err.Error(), 2)
}
