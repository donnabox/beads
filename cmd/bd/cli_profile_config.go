package main

import "github.com/steveyegge/beads/internal/config"

// loadCLIProfileForInvocation is a proposed bootstrap seam. The fixed config
// inspect/set/unset repair path must not require this normal-invocation loader.
// Flag spelling and selected-tree integration remain outside this patch.
func loadCLIProfileForInvocation(argument *string) (cliProfileSelection, error) {
	return selectCLIProfileFromReader(argument, func() (string, bool, error) {
		return config.UserYamlValueStrict(config.CLIProfileKey)
	})
}

func selectCLIProfileFromReader(argument *string, read func() (string, bool, error)) (cliProfileSelection, error) {
	if argument != nil {
		return selectCLIProfile(nil, argument)
	}
	value, present, err := read()
	if err != nil {
		return cliProfileSelection{}, err
	}
	if !present {
		return selectCLIProfile(nil, nil)
	}
	return selectCLIProfile(&value, nil)
}
