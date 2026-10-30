package main

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/pflag"
)

// Proposed receiving spelling. This is not a wire field or an adopted flag.
const cliProfilePrefixFlag = "cli-profile"

// cliBootstrapArity copies only argument-shape metadata, never a Flag, Value,
// Command, callback, backing variable or Changed state from the real tree.
type cliBootstrapArity struct {
	Name        string
	Shorthand   string
	NoOptDefVal string
}

type cliBootstrapOccurrence struct {
	Name  string
	Value string
}

type cliProfilePrefix struct {
	OriginalArgs  []string
	PrefixArgs    []string
	RemainingArgs []string
	ExecutionArgs []string
	Occurrences   []cliBootstrapOccurrence
	ProfileValues []string
	Override      *string
	Terminated    bool
	// Cobra completion wraps the user's argv after an internal command token.
	// That separate preparation seam is unresolved; callers must handle this
	// marker before choosing a profile/catalog or executing the returned argv.
	CompletionPayload bool
	// The root-only parser cannot classify a leaf/local prefix flag. With
	// no profile occurrence yet, ordinary Cobra owns the whole untouched argv.
	// No parsed prefix/Scope authority may be inferred from this result.
	DeferredToCobra bool
}

func snapshotCLIBootstrapArity(sets ...*pflag.FlagSet) ([]cliBootstrapArity, error) {
	byName := make(map[string]cliBootstrapArity)
	var conflict error
	for _, set := range sets {
		if set == nil {
			continue
		}
		set.VisitAll(func(flag *pflag.Flag) {
			shape := cliBootstrapArity{flag.Name, flag.Shorthand, flag.NoOptDefVal}
			if old, exists := byName[shape.Name]; exists && old != shape {
				conflict = fmt.Errorf("conflicting root flag arity for %s", shape.Name)
			}
			byName[shape.Name] = shape
		})
	}
	if conflict != nil {
		return nil, conflict
	}
	result := make([]cliBootstrapArity, 0, len(byName))
	for _, shape := range byName {
		result = append(result, shape)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// Private disposable value: ordinary pflag parses arity through ParseAll, and
// the callback retains raw tokens. The actual tree validates other values once
// during normal parsing; no real flag Value.Set is called by this decoder.
type cliBootstrapValue struct{ text string }

func (v *cliBootstrapValue) Set(value string) error { v.text = value; return nil }
func (v *cliBootstrapValue) String() string         { return v.text }
func (v *cliBootstrapValue) Type() string           { return "string" }

func decodeCLIProfilePrefix(argv []string, shapes []cliBootstrapArity) (cliProfilePrefix, error) {
	var result cliProfilePrefix
	result.OriginalArgs = append([]string(nil), argv...)
	flags := pflag.NewFlagSet("profile-bootstrap", pflag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.SetInterspersed(false)
	seenNames := make(map[string]bool)
	seenShort := make(map[string]bool)
	for _, shape := range shapes {
		if shape.Name == "" || shape.Name == cliProfilePrefixFlag || seenNames[shape.Name] ||
			len(shape.Shorthand) > 1 || shape.Shorthand != "" && seenShort[shape.Shorthand] {
			return result, fmt.Errorf("invalid or colliding bootstrap root shape %q", shape.Name)
		}
		seenNames[shape.Name] = true
		if shape.Shorthand != "" {
			seenShort[shape.Shorthand] = true
		}
		flags.AddFlag(&pflag.Flag{
			Name: shape.Name, Shorthand: shape.Shorthand,
			NoOptDefVal: shape.NoOptDefVal, Value: &cliBootstrapValue{},
		})
	}
	// Cobra normally adds help lazily. Make its arity explicit in this private
	// parser so help does not prematurely terminate prefix decoding.
	if !seenNames["help"] {
		if seenShort["h"] {
			return result, fmt.Errorf("root shorthand h conflicts with help")
		}
		flags.BoolP("help", "h", false, "")
	}
	flags.StringArray(cliProfilePrefixFlag, nil, "")
	var replayPrefix []string
	err := flags.ParseAll(append([]string(nil), argv...), func(flag *pflag.Flag, value string) error {
		result.Occurrences = append(result.Occurrences, cliBootstrapOccurrence{flag.Name, value})
		if flag.Name == cliProfilePrefixFlag {
			result.ProfileValues = append(result.ProfileValues, value)
		} else {
			// Replay root options using an unambiguous equals form, preserving
			// exact value text and occurrence order. This removes only the new
			// prefix selector without a handwritten token/short-cluster scanner.
			replayPrefix = append(replayPrefix, "--"+flag.Name+"="+value)
		}
		return nil
	})
	if err != nil {
		var unknown *pflag.NotExistError
		if errors.As(err, &unknown) {
			if len(result.ProfileValues) == 0 {
				// pflag exposes no safe unparsed-token offset on failure.
				// Do not skip an unknown option or inspect its possible value:
				// defer the ENTIRE original argv to the real command parser.
				return cliProfilePrefix{
					OriginalArgs:    append([]string(nil), argv...),
					RemainingArgs:   append([]string(nil), argv...),
					ExecutionArgs:   append([]string(nil), argv...),
					DeferredToCobra: true,
				}, nil
			}
			return result, fmt.Errorf("after prefix --%s, use only known root options before the command: %w", cliProfilePrefixFlag, err)
		}
		return result, err
	}
	result.RemainingArgs = append([]string(nil), flags.Args()...)
	result.PrefixArgs = append([]string(nil), argv[:len(argv)-len(result.RemainingArgs)]...)
	result.Terminated = flags.ArgsLenAtDash() >= 0
	if result.Terminated {
		replayPrefix = append(replayPrefix, "--")
	}
	result.ExecutionArgs = append(replayPrefix, result.RemainingArgs...)
	if len(result.ProfileValues) == 0 {
		// Preserve every historical token/spelling when the new prefix
		// selector is absent. Replay normalization is needed only to remove
		// an explicitly supplied new selector; raw forwarding stays intact.
		result.ExecutionArgs = append([]string(nil), argv...)
	}
	if len(result.ProfileValues) != 0 {
		// Proposed ordinary scalar-flag policy: last occurrence wins, as with
		// pflag.StringVar. All occurrences remain visible for receiving review.
		value := result.ProfileValues[len(result.ProfileValues)-1]
		if _, err := selectCLIProfile(nil, &value); err != nil {
			return result, err
		}
		result.Override = &value
	}
	if len(result.RemainingArgs) != 0 {
		first := result.RemainingArgs[0]
		result.CompletionPayload = first == "__complete" || first == "__completeNoDesc"
	}
	return result, nil
}

type cliProfilePrefixOnlyValue struct{}

func (*cliProfilePrefixOnlyValue) String() string { return "" }
func (*cliProfilePrefixOnlyValue) Type() string   { return "string" }
func (*cliProfilePrefixOnlyValue) Set(string) error {
	return fmt.Errorf("--%s must precede the command in the root-option prefix", cliProfilePrefixFlag)
}

// Proposed integration guard, not called by production source. Prefix selector
// tokens have already been removed from ExecutionArgs. A normal selected-tree
// parse therefore rejects any actual late selector; a field's literal value
// that happens to spell --cli-profile remains that field's value.
func registerCLIProfilePrefixOnlyGuard(flags *pflag.FlagSet) error {
	if flags.Lookup(cliProfilePrefixFlag) != nil {
		return fmt.Errorf("reserved bootstrap selector --%s already registered", cliProfilePrefixFlag)
	}
	flags.Var(&cliProfilePrefixOnlyValue{}, cliProfilePrefixFlag, "Select CLI profile before the command")
	return nil
}
