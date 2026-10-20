package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strconv"
)

// Existing scalar flags retain their shipping pflag types. Reparse occurrences
// only for conflict detection; this does not add a selector or query grammar.
func canonicalScalarOccurrences(cmd *cobra.Command, argv []string, names ...string) error {
	probe := pflag.NewFlagSet("canonical-collection-admission", pflag.ContinueOnError)
	probe.SetOutput(&bytes.Buffer{})
	values := make(map[string]*canonicalScalarValues)
	scalar := make(map[string]bool, len(names))
	for _, name := range names {
		scalar[name] = true
	}
	add := func(flag *pflag.Flag) {
		if probe.Lookup(flag.Name) != nil {
			return
		}
		value := &canonicalScalarValues{kind: flag.Value.Type()}
		if scalar[flag.Name] {
			values[flag.Name] = value
		}
		probe.AddFlag(&pflag.Flag{Name: flag.Name, Shorthand: flag.Shorthand, NoOptDefVal: flag.NoOptDefVal, Value: value})
	}
	cmd.Flags().VisitAll(add)
	cmd.InheritedFlags().VisitAll(add)
	if err := probe.Parse(argv); err != nil {
		return err
	}
	for name, value := range values {
		if len(value.values) < 2 {
			continue
		}
		first := value.normalized(value.values[0])
		for _, next := range value.values[1:] {
			if value.normalized(next) != first {
				return fmt.Errorf("conflicting --%s values", name)
			}
		}
	}
	return nil
}

type canonicalScalarValues struct {
	kind   string
	values []string
}

func (v *canonicalScalarValues) Set(value string) error {
	v.values = append(v.values, value)
	return nil
}
func (v *canonicalScalarValues) String() string { return "" }
func (v *canonicalScalarValues) Type() string   { return v.kind }
func (v *canonicalScalarValues) normalized(value string) string {
	if v.kind == "bool" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return strconv.FormatBool(parsed)
		}
	}
	if v.kind == "int" {
		if parsed, err := strconv.ParseInt(value, 0, 64); err == nil {
			return strconv.FormatInt(parsed, 10)
		}
	}
	return value
}
