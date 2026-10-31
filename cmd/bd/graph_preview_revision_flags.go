package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clirevision"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// A graph revision flag names a local history ordinal or escapes an exact
// version token with '='. Resolve ordinals for this identity before passing a
// guard or historical source to a writer; the writer still checks its current
// predecessor in the same transaction. Legacy-mode flags never call here.
func graphPreviewRevisionFlag(cmd *cobra.Command, name, identity string, privatePath bool) (string, error) {
	raw, _ := cmd.Flags().GetString(name)
	value, err := graphPreviewRevisionValue(name, raw, identity, privatePath)
	if err != nil {
		return "", graphPreviewReferenceFailure(err)
	}
	return value, nil
}

func graphPreviewRevisionValue(name, raw, identity string, privatePath bool) (string, error) {
	selected, err := graphPreviewRevisionSelector(name, raw)
	if err != nil {
		return "", err
	}
	if selected.Token != "" {
		return selected.Token, nil
	}
	if identity == "" {
		return "", fmt.Errorf("--%s requires one Resource identity", name)
	}
	if privatePath {
		return graphPreviewRevisionAtPath(identity, selected.Ordinal)
	}
	resolved, err := graphPreviewResolveParsedReference(graphPreviewReference{Identity: identity, Explicit: true, Ordinal: selected.Ordinal})
	if err != nil {
		return "", err
	}
	return resolved.Revision, nil
}

func graphPreviewRevisionAtPath(path string, ordinal int64) (revision string, err error) {
	defer func() {
		if err != nil {
			err = graphPreviewReferenceResolutionError{err}
		}
	}()
	ctx, cancel := context.WithTimeout(getRootContext(), 30*time.Second)
	defer cancel()
	store, err := openSelectedGraphStore(ctx, graphOptions(graphPreviewConfig))
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, store.Close()) }()
	identity, err := store.AllocatedResourceIdentityForPath(ctx, path)
	if err != nil {
		return "", err
	}
	row, err := store.ResolveLocalRevision(ctx, identity.PrivatePath, ordinal)
	if err != nil {
		return "", err
	}
	return row.Version, nil
}

func graphPreviewParseRevisionFlag(name, raw string) (clirevision.Selector, error) {
	selected, err := graphPreviewRevisionSelector(name, raw)
	if err != nil {
		return clirevision.Selector{}, graphFailure("invalid_selector", err.Error(), 2)
	}
	return selected, nil
}

func graphPreviewRevisionSelector(name, raw string) (clirevision.Selector, error) {
	selected, err := clirevision.Parse(raw, graphstore.PreviewVersionTokenLimit)
	if err != nil {
		if strings.HasPrefix(err.Error(), "use @N") {
			err = fmt.Errorf("use a positive local revision number N or =TOKEN")
		}
		return clirevision.Selector{}, fmt.Errorf("--%s: %w", name, err)
	}
	if selected.Current {
		return clirevision.Selector{}, fmt.Errorf("--%s: use a positive local revision number N or =TOKEN", name)
	}
	return selected, nil
}
