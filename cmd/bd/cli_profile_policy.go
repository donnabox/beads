package main

import "fmt"

// cliSurfaceProfile selects command adapters, independently of graph mode,
// output representation, transport, capabilities and installed packs.
type cliSurfaceProfile string

const (
	cliSurfaceCompat    cliSurfaceProfile = "compat"
	cliSurfaceCanonical cliSurfaceProfile = "canonical"
)

type cliProfileSource string

const (
	cliProfileBuiltin  cliProfileSource = "builtin"
	cliProfileSaved    cliProfileSource = "saved"
	cliProfileArgument cliProfileSource = "argument"
)

type cliProfileSelection struct {
	Profile cliSurfaceProfile
	Source  cliProfileSource
}

// selectCLIProfile resolves already-decoded bootstrap input. A nil pointer is
// absence; an explicitly empty value is not absence. Reading/writing settings
// and decoding argv belong to the bootstrap adapter, not this pure policy.
// No flag name, configuration key, persistence scope or environment override
// is chosen here. A one-call override never writes the saved setting.
func selectCLIProfile(saved, argument *string) (cliProfileSelection, error) {
	value := string(cliSurfaceCompat)
	source := cliProfileBuiltin
	if saved != nil {
		value = *saved
		source = cliProfileSaved
	}
	if argument != nil {
		value = *argument
		source = cliProfileArgument
	}
	switch cliSurfaceProfile(value) {
	case cliSurfaceCompat, cliSurfaceCanonical:
		return cliProfileSelection{Profile: cliSurfaceProfile(value), Source: source}, nil
	default:
		return cliProfileSelection{}, fmt.Errorf("invalid CLI surface profile %q from %s: use canonical or compat", value, source)
	}
}
