package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// graphVersionsNotFound is the "there is no such thing" shape, kept apart from
// the other two so a mistyped id never reads as a Resource with no history.
//
// It RE-WRAPS the sentinel rather than calling graphFailure, and that matters.
// graphFailure both prints a diagnostic and returns an exitError, so calling it
// from inside the withGraphStore closure emitted the message and then handed the
// exitError on to graphStorageError, which matched no sentinel, printed
// graph_not_initialized over the top and exited 5 instead of 3. Two diagnostics
// and the wrong code, for one failure. Wrapping keeps a single owner of the
// mapping: graphStorageError turns ErrNotFound into not_found, exit 3, once.
func graphVersionsNotFound(selector string) error {
	return fmt.Errorf("%w: nothing is allocated at %s in this workspace, so it has no version history"+
		" (a deleted Resource still lists its history, so this path was never allocated here)",
		graphstore.ErrNotFound, selector)
}

// graphVersionRowsJSON fixes the wire shape in the CLI layer, where the
// documented contract lives, instead of depending on struct tags in the store
// package. The names are deliberate and documented in
// docs/reference/graph-preview.md.
//
// `ordinal` is NOT called `revision`: graph records already use `revision` for
// the opaque token, and native Issues use it for the row-lock CAS token, which
// is a string. A script that confused the two would silently compare an
// ordering key against an address.
func graphVersionRowsJSON(rows []graphstore.VersionRow) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"ordinal": r.Ordinal,
			"version": r.Version,
			// RFC3339Nano rather than the human column's microseconds: the
			// column is for reading, this is for parsing.
			"change_at":   r.ChangeAt.UTC().Format(time.RFC3339Nano),
			"actor":       r.Actor,
			"attribution": r.Attribution,
		})
	}
	return out
}

// runGraphPreviewVersions lists one graph Resource's versions, newest first.
//
// #5898's third goal is that a history answer has THREE shapes rather than
// two: here it is, there is no such thing, and something prevented a complete
// answer. Two of those three are refusals and belong to graphStorageError,
// which already turns graphstore.ErrNotFound into not_found and
// ErrCapabilityUnavailable into capability_unavailable. The only shape this
// function renders itself is the honest empty: a Resource that exists, on a
// plane that can order, which has recorded nothing yet.
//
// An empty list must never stand in for either refusal. "No versions" is a
// claim about the Resource; "this store cannot answer" is a claim about the
// store, and printing the first when the second is true is a wrong answer
// rather than an empty one. Same principle as cmd/bd/versions.go on the
// native plane.
func runGraphPreviewVersions(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	// Resolved with graphPreviewResourcePath rather than
	// graphPreviewBeadSelector: Links have versions too, so this accepts
	// links/PATH as well as beads/PATH. `compare` resolves the same way for
	// the same reason.
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		kind, rows, err := store.Versions(ctx, path)
		switch {
		case errors.Is(err, graphstore.ErrNotFound):
			// graphStorageError already maps ErrNotFound to not_found, so this
			// branch exists only to say something more useful than the store's
			// own wording. The distinction worth stating is that absence here
			// means NEVER ALLOCATED: a deleted Resource still lists its
			// history, so "no versions" and "no such path" are different
			// answers and a reader should not have to guess which one it got.
			return nil, "", graphVersionsNotFound(args[0])
		case err != nil:
			// Everything else is left whole for graphStorageError, which owns
			// the sentinel-to-code mapping. Interpreting it twice would print
			// two diagnostics for one failure.
			return nil, "", err
		}
		result := map[string]any{
			"resource": path,
			"kind":     string(kind),
			"versions": graphVersionRowsJSON(rows),
		}
		return result, renderGraphVersions(args[0], kind, rows), nil
	})
}

// renderGraphVersions prints the token on its own line under each row.
//
// That is not a layout accident. On this plane the opaque version token is the
// ONLY citable address -- it is what `show --version` and `compare --from/--to`
// accept -- so a listing that elides it cannot be used for the thing the
// listing is for. The ordinal is 1-2 characters and the token is 32, which no
// single terminal row holds alongside a timestamp and an actor, so the token
// gets its own line rather than being truncated into uselessness.
func renderGraphVersions(selector string, kind graphstore.ResourceKind, rows []graphstore.VersionRow) string {
	if len(rows) == 0 {
		// Reachable only when the Resource exists and its plane can order:
		// both refusals returned an error above. Say which of the three
		// answers this is, so it cannot be mistaken for either refusal.
		return fmt.Sprintf(
			"No versions recorded for %s yet.\n\n"+
				"This %s exists and this store can order its versions. Nothing has\n"+
				"changed it since it was created, so there is nothing to list.",
			selector, kind)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Versions of %s (%d)\n\n", selector, len(rows))
	fmt.Fprintf(&b, "  ORD  WHEN                        WHO                  ATTRIB\n")
	for _, r := range rows {
		attrib := r.Attribution
		if attrib == "" {
			// Memory and Link carry no native attribution status. A dash
			// reads as "this plane does not record it"; an empty column reads
			// as missing data.
			attrib = "-"
		}
		fmt.Fprintf(&b, "  %-4d %-27s %-20.20s %s\n",
			r.Ordinal,
			// Microseconds are shown because they are the reason migration
			// 0069 widened change_at: two writes inside one second are
			// ordinary, and a whole-second rendering makes them look
			// simultaneous when they are not.
			r.ChangeAt.UTC().Format("2006-01-02 15:04:05.000000"),
			r.Actor,
			attrib)
		fmt.Fprintf(&b, "       %s\n", r.Version)
	}

	b.WriteString("\nORD orders versions within this store and is NOT a citable address: two\n")
	b.WriteString("clones can hold a different ordinal for the same state. Cite the token\n")
	b.WriteString("under each row instead:\n\n")
	fmt.Fprintf(&b, "  bd show %s --version <token>\n", selector)
	fmt.Fprintf(&b, "  bd compare %s --from <token> --to <token>\n", selector)
	return b.String()
}
