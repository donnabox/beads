package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/graphoutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

const graphCurrentDumpByteLimit = 16 << 20

// This diagnostic container deliberately has no restoration contract. Its
// records reuse the current public DTOs; never serialize Inventory.WriterToken.
type graphCurrentDump struct {
	Format     string                   `json:"format"`
	Scope      string                   `json:"scope"`
	Restorable bool                     `json:"restorable"`
	Omissions  []string                 `json:"omissions"`
	Beads      []bdpwire.BeadRecord     `json:"beads"`
	Links      []bdpwire.LinkRecord     `json:"links"`
	Types      []bdpwire.TypeDescriptor `json:"types"`
}

func graphCurrentStateExportSelected(cmd *cobra.Command) bool {
	selected, _ := cmd.Flags().GetBool("current-state")
	return cmd == exportCmd && selected
}

func runGraphPreviewExport(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "current-state", "output"); err != nil {
		return err
	}
	selected, _ := cmd.Flags().GetBool("current-state")
	if !selected || len(args) != 0 {
		return graphFailure("invalid_selector", "graph export requires --current-state and no positional arguments; this interim JSON dump is not restorable", 2)
	}
	path, _ := cmd.Flags().GetString("output")
	if path != "" {
		if err := graphCurrentDumpDestination(path, graphPreviewDir); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	ctx, cancel := context.WithTimeout(getRootContext(), 30*time.Second)
	defer cancel()
	store, err := graphstore.OpenExisting(ctx, graphOptions(graphPreviewConfig))
	if err != nil {
		return graphStorageError(err)
	}
	source := graphCurrentDumpSource{Reader: graphread.New(store), store: store}
	data, err := captureGraphCurrentDump(ctx, source, store.ScopeURL())
	if err != nil {
		return graphStorageErrorWithFallback(err, "route_unavailable")
	}
	publishCtx, publishCancel := context.WithTimeout(getRootContext(), 30*time.Second)
	defer publishCancel()
	if path != "" {
		err = graphoutput.PublishFile(publishCtx, path, data, 0o600)
	} else {
		err = graphoutput.WriteStream(publishCtx, cmd.OutOrStdout(), data)
	}
	if err != nil {
		return graphFailure("output_error", err.Error(), 1)
	}
	return nil
}

// Resolve the parent, not the target: the publisher replaces a target symlink
// rather than following it. A dump must never replace workspace storage files.
func graphCurrentDumpDestination(path, workspace string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return fmt.Errorf("current-state output directory: %w", err)
	}
	root, err := os.Stat(workspace)
	if err != nil {
		return err
	}
	// A destination equal to the storage root may have different spelling on
	// a case-insensitive volume, or itself be a symlink used for discovery.
	// Refuse that root identity; ordinary output symlinks to storage files
	// remain safe because publication replaces them, not their referents.
	target := filepath.Join(parent, filepath.Base(absolute))
	if info, err := os.Stat(target); err == nil && os.SameFile(root, info) {
		return errors.New("current-state output must be outside the .beads storage directory")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	// Compare filesystem identities rather than path spelling. EvalSymlinks
	// alone does not canonicalize case aliases on macOS filesystems.
	for current := parent; ; current = filepath.Dir(current) {
		info, err := os.Stat(current)
		if err != nil {
			return err
		}
		if os.SameFile(root, info) {
			return errors.New("current-state output must be outside the .beads storage directory")
		}
		if filepath.Dir(current) == current {
			break
		}
	}
	return nil
}

type graphCurrentDumpReader interface {
	Inventory(context.Context) (graphread.Inventory, error)
	Close() error
}

type graphCurrentDumpSource struct {
	*graphread.Reader
	store *graphstore.Store
}

func (s graphCurrentDumpSource) Close() error { return s.store.Close() }

// Read and cleanup failures both suppress output, including when the reader
// returns a partial inventory. No read occurs after the checked close.
func captureGraphCurrentDump(ctx context.Context, source graphCurrentDumpReader, scope string) ([]byte, error) {
	closed := false
	defer func() {
		if !closed {
			_ = source.Close() // Panic fallback only; ordinary cleanup is checked.
		}
	}()
	inventory, readErr := source.Inventory(ctx)
	closeErr := source.Close()
	closed = true
	if err := errors.Join(readErr, closeErr, ctx.Err()); err != nil {
		return nil, err
	}
	// Normalize an empty capture to arrays without changing any record fields.
	if inventory.Beads == nil {
		inventory.Beads = []bdpwire.BeadRecord{}
	}
	if inventory.Links == nil {
		inventory.Links = []bdpwire.LinkRecord{}
	}
	if inventory.Types == nil {
		inventory.Types = []bdpwire.TypeDescriptor{}
	}
	dump := graphCurrentDump{
		Format: "beads-current-state-dump-v1", Scope: scope, Restorable: false,
		Omissions: []string{
			"History, deleted Resources and historical endpoint payloads (pinned References remain)",
			"all durable Issue comments (omission counts are not measured)",
			"configuration, issue prefix, event journals, wisps and legacy KV memories",
			"private storage columns, native Type bootstrap bytes and operational authority/custody",
		},
		Beads: inventory.Beads, Links: inventory.Links, Types: inventory.Types,
	}
	data, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: current public DTO encoding: %v", graphstore.ErrInvalidStore, err)
	}
	if len(data) >= graphCurrentDumpByteLimit {
		return nil, fmt.Errorf("%w: current-state JSON exceeds the independent %d-byte encoded-output limit; this interim command cannot split or project the dump; no output was published", graphstore.ErrLimitExceeded, graphCurrentDumpByteLimit)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
