package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/timeparsing"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
	"github.com/steveyegge/beads/internal/utils"
)

// Returned canonical IDs can be fed back to the CLI. A bare local Bead path
// is CLI shorthand for beads/PATH; mixed commands keep links/PATH explicit.
// No store lookup, alias resolution or remote routing chooses the kind.
func graphPreviewResourcePath(scope, selector string) (string, error) {
	if graph.ValidateBeadPath(selector) == nil || graph.ValidateLinkPath(selector) == nil {
		return selector, nil
	}
	if path, kind, ok := graph.SplitCanonicalURL(scope, selector); ok && (kind == graph.KindBead || kind == graph.KindLink) {
		return path, nil
	}
	return graphPreviewBareBeadPath(selector)
}

func graphPreviewBareBeadPath(selector string) (string, error) {
	for _, root := range []string{"beads/", "links/", "alias/", "types/"} {
		if strings.HasPrefix(selector, root) {
			return "", fmt.Errorf("invalid local selector %q; use a valid beads/PATH or links/PATH", selector)
		}
	}
	if strings.Contains(selector, "://") {
		return "", fmt.Errorf("invalid local selector %q; a URL must be this workspace's Scope URL followed by beads/PATH or links/PATH, and --id accepts only a bare ID or beads/PATH", selector)
	}
	path := "beads/" + selector
	if err := graph.ValidateBeadPath(path); err != nil {
		return "", fmt.Errorf("invalid Bead ID %q: %w", selector, err)
	}
	return path, nil
}

// A Link-only command can use a bare ID without probing the Bead namespace.
// Mixed-resource commands continue to require an explicit links/PATH.
func graphPreviewLinkPath(scope, selector string) (string, error) {
	if graph.ValidateLinkPath(selector) == nil {
		return selector, nil
	}
	if path, kind, ok := graph.SplitCanonicalURL(scope, selector); ok && kind == graph.KindLink {
		return path, nil
	}
	for _, root := range []string{"beads/", "links/", "alias/", "types/"} {
		if strings.HasPrefix(selector, root) {
			return "", fmt.Errorf("invalid Link selector %q; use a valid links/PATH", selector)
		}
	}
	if strings.Contains(selector, "://") {
		return "", fmt.Errorf("invalid Link selector %q; a URL must be this workspace's Scope URL followed by links/PATH", selector)
	}
	path := "links/" + selector
	if err := graph.ValidateLinkPath(path); err != nil {
		return "", fmt.Errorf("invalid Link ID %q: %w", selector, err)
	}
	return path, nil
}

// Both familiar spellings and the experimental generic Type selector reach
// one domain operation. No alternate Link writer can bypass Issue policy.
func runGraphPreviewAddDependency(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "type", "source", "target", "id", "if-source-revision", "blocked-by", "depends-on", "no-cycle-check"); err != nil {
		return err
	}
	// Keep the ordinary single-edge spellings. Native --blocked-by wins if both
	// aliases are supplied; --no-cycle-check suppresses only a post-add warning
	// for one edge and never bypasses the domain writer's cycle validation.
	blockedBy, _ := cmd.Flags().GetString("blocked-by")
	dependsOn, _ := cmd.Flags().GetString("depends-on")
	// Match dep add's native Args contract even if another graph caller reaches
	// this writer without Cobra's positional validation. An explicitly empty
	// flag is inert; a nonempty flag cannot compete with a positional target.
	if len(args) > 1 && (blockedBy != "" || dependsOn != "") {
		return graphFailure("invalid_selector", "cannot use both positional depends-on-id and --blocked-by/--depends-on flag", 2)
	}
	if len(args) == 1 {
		target := blockedBy
		if target == "" {
			target = dependsOn
		}
		if target != "" {
			args = append(args, target)
		}
	}
	if len(args) != 2 {
		return graphFailure("invalid_selector", "graph Dependency creation requires two Bead IDs or beads/PATH arguments", 2)
	}
	paths := make([]string, len(args))
	for i, selector := range args {
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if err := graph.ValidateBeadPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		paths[i] = path
	}
	var sourceRevision string
	var selectedType string
	raw, _ := cmd.Flags().GetString("type")
	if cmd.Flags().Changed("type") && graphPreviewNominalSelector(raw) {
		if typ, err := graphPreviewTypeURL(graphPreviewConfig.GraphScopeURL, raw); err != nil || typ != graphstore.DependencyTypeURL(graphPreviewConfig.GraphScopeURL) {
			return graphFailure("capability_unavailable", "Dependency creation requires the installed Dependency Type", 5)
		}
		raw = string(types.DepBlocks)
	}
	depType := canonicalDependencyType(types.DependencyType(raw))
	if !depType.IsWellKnown() {
		return graphFailure("capability_unavailable", "Dependency type is not registered in this preview", 5)
	}
	linkPath := ""
	if cmd.Flags().Changed("id") {
		selector, _ := cmd.Flags().GetString("id")
		var err error
		linkPath, err = graphPreviewLinkPath(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	ownedCreate := cmd.Name() == "create" && cmd.Flags().Lookup("source") != nil
	guardedShorthand := cmd.Name() == "link" && cmd.Flags().Changed("if-source-revision")
	if ownedCreate || guardedShorthand {
		sourceRevision, _ = cmd.Flags().GetString("if-source-revision")
		if cmd.Flags().Changed("if-source-revision") && sourceRevision == "" {
			return graphFailure("invalid_selector", "--if-source-revision requires a token", 2)
		}
		selectedType = graphstore.DependencyTypeURL(graphPreviewConfig.GraphScopeURL)
	} else {
		var ok bool
		selectedType, ok = graphPreviewDependencyTypeURL(graphPreviewConfig.GraphScopeURL, depType)
		if !ok {
			return graphFailure("capability_unavailable", "this workspace generation does not support that native Dependency kind", 5)
		}
	}
	return graphPreviewInsertDependency(paths[0], paths[1], selectedType, depType, linkPath, sourceRevision, args[0], args[1])
}

// The store independently checks the installed Type in the write transaction.
func graphPreviewDependencyTypeURL(scope string, native types.DependencyType) (string, bool) {
	if native.IsValid() {
		return graphstore.DependencyTypeURL(scope), true
	}
	return "", false
}

func graphPreviewNativeDependencyType(scope, typeURL string) (types.DependencyType, bool) {
	if typeURL == graphstore.DependencyTypeURL(scope) {
		// Identity admission only; --type supplies the native kind.
		return types.DepBlocks, true
	}
	return "", false
}

// The ordinary dep remove verb addresses an Issue pair, not a Link ID. Pair
// selection and the no-op check happen inside the graph store's one write
// transaction; the actual removal uses the canonical guarded unlink writer.
func runGraphPreviewRemoveDependency(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	if len(args) != 2 {
		return graphFailure("invalid_selector", "dep remove requires two Issue IDs or beads/PATH arguments", 2)
	}
	sourcePath, err := graphPreviewBeadSelector(args[0])
	if err != nil {
		return err
	}
	targetPath, err := graphPreviewBeadSelector(args[1])
	if err != nil {
		return err
	}
	request := graphstore.DependencyRemoveRequest{SourcePath: sourcePath, TargetPath: targetPath, Actor: getActorWithGit()}
	if sourcePath == "beads/"+args[0] {
		request.SourceNativeID = args[0]
	}
	if targetPath == "beads/"+args[1] {
		request.TargetNativeID = args[1]
	}
	return withGraphFamiliarStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		removed, err := store.RemoveDependency(ctx, request)
		if err != nil {
			return nil, "", err
		}
		status, human := "not_found", fmt.Sprintf("No dependency found: %s → %s", removed.Source.Properties.ID, removed.Target.Properties.ID)
		if removed.Removed {
			status, human = "removed", fmt.Sprintf("Removed dependency: %s → %s", removed.Source.Properties.ID, removed.Target.Properties.ID)
		}
		return map[string]any{
			"status": status, "removed": removed.Removed,
			"issue_id": removed.Source.Properties.ID, "depends_on_id": removed.Target.Properties.ID,
		}, human, nil
	})
}

func runGraphPreviewDepBlocks(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "blocks"); err != nil {
		return err
	}
	blocked, _ := cmd.Flags().GetString("blocks")
	if len(args) == 0 && blocked == "" {
		return cmd.Help()
	}
	if len(args) != 1 || blocked == "" {
		return graphFailure("invalid_selector", "dep BLOCKER --blocks BLOCKED requires two Issue IDs", 2)
	}
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	sourcePath, err := graphPreviewBeadSelector(blocked)
	if err != nil {
		return err
	}
	targetPath, err := graphPreviewBeadSelector(args[0])
	if err != nil {
		return err
	}
	return graphPreviewInsertDependency(sourcePath, targetPath, "", types.DepBlocks, "", "", blocked, args[0])
}

func graphPreviewInsertDependency(sourcePath, targetPath, typeURL string, depType types.DependencyType, linkPath, sourceRevision, sourceLabel, targetLabel string) error {
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.AddDependency(ctx, graphstore.DependencyRequest{Path: linkPath, Type: depType, TypeURL: typeURL, SourcePath: sourcePath, TargetPath: targetPath, Actor: getActorWithGit(), ExpectedSourceRevision: sourceRevision})
		if err != nil {
			if strings.Contains(err.Error(), "operation requires a live durable Issue, not generic") {
				kind := "blocking"
				if typeURL != "" && typeURL != graphPreviewBlockingTypeURL(graphPreviewConfig.GraphScopeURL) {
					kind = "native Dependency"
				}
				return nil, "", fmt.Errorf("%s Link Type requires two live Issues; use an informational Type such as types/preview-related for Memory endpoints: %w", kind, err)
			}
			return nil, "", err
		}
		verb := "Created"
		if !result.Changed {
			verb = "Unchanged"
		}
		if depType == types.DepBlocks {
			return result, fmt.Sprintf("%s %s: %s depends on %s\n", verb, result.Link.ID, sourceLabel, targetLabel), nil
		}
		return result, fmt.Sprintf("%s %s: %s → %s\n", verb, result.Link.ID, sourceLabel, targetLabel), nil
	})
}

func runGraphPreviewClose(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if len(args) == 0 {
		lastTouched := GetLastTouchedID()
		if lastTouched == "" {
			return graphFailure("invalid_selector", "no Issue ID provided and no last touched Issue", 2)
		}
		args = []string{lastTouched}
	}
	if err := graphPreviewFlags(cmd, "reason", "resolution", "message", "comment", "reason-file", "force", "session", "suggest-next", "claim-next", "if-revision"); err != nil {
		return err
	}
	force, _ := cmd.Flags().GetBool("force")
	suggestNext, _ := cmd.Flags().GetBool("suggest-next")
	claimNext, _ := cmd.Flags().GetBool("claim-next")
	session, _ := cmd.Flags().GetString("session")
	if session == "" {
		session = os.Getenv("CLAUDE_SESSION_ID")
	}
	reasons, targets, err := resolveCloseReasons(cmd, args)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	if len(targets) == 0 {
		return graphFailure("invalid_selector", "graph close requires at least one Bead ID or beads/PATH", 2)
	}
	revision, _, err := graphPreviewRevisionGuard(cmd, false)
	if err != nil {
		return err
	}
	if revision != "" && len(targets) != 1 {
		return graphFailure("invalid_selector", "--if-revision only works when closing a single Issue", 2)
	}
	if revision != "" && (suggestNext || claimNext) {
		return graphFailure("invalid_selector", "--if-revision does not support --suggest-next or --claim-next", 2)
	}
	if suggestNext && len(targets) != 1 {
		return graphFailure("invalid_selector", "--suggest-next only works when closing a single Issue", 2)
	}
	if err := validateCloseReasons(reasons); err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	paths := make([]string, len(targets))
	for i, selector := range targets {
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if err := graph.ValidateBeadPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		paths[i] = path
	}
	if claimNext {
		return runGraphPreviewCloseClaimNext(paths, targets, reasons, session, force, suggestNext)
	}
	var hadError bool
	var firstSettled string
	err = withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		results := make([]graphstore.IssueMutationResult, 0, len(paths))
		var human strings.Builder
		for i, path := range paths {
			result, err := store.CloseIssueWithOptions(ctx, graphstore.IssueCloseRequest{
				Path: path, Reason: reasonForCloseIndex(reasons, i), Actor: getActorWithGit(), Session: session, Force: force,
				ExpectedRevision: revision,
			})
			if err != nil {
				if len(paths) == 1 {
					return nil, "", err
				}
				_ = graphStorageError(fmt.Errorf("closing %s: %w", targets[i], err))
				hadError = true
				continue
			}
			results = append(results, result)
			if firstSettled == "" {
				firstSettled = path
			}
			if result.OpenChildren > 0 {
				fmt.Fprintf(os.Stderr, "warning: closing %s with %d open child issue(s) still active\n", targets[i], result.OpenChildren)
			}
			verb := "Closed"
			if !result.Changed {
				verb = "Already closed"
			}
			fmt.Fprintf(&human, "%s %s\n", verb, targets[i])
			if suggestNext {
				unblocked, err := store.NewlyUnblockedByClose(ctx, path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not find newly unblocked Issues: %v\n", err)
				} else if len(unblocked) > 0 {
					human.WriteString("\nNewly unblocked:\n")
					for _, issue := range unblocked {
						fmt.Fprintf(&human, "  %s (P%d)\n", issue.ID, issue.Properties.Priority)
					}
					return map[string]any{"closed": []graphstore.IssueRecord{result.Issue}, "unblocked": unblocked}, strings.TrimSuffix(human.String(), "\n"), nil
				}
			}
		}
		if len(paths) == 1 {
			return results[0], human.String(), nil
		}
		return results, strings.TrimSuffix(human.String(), "\n"), nil
	})
	if err != nil {
		return err
	}
	SetLastTouchedID(firstSettled)
	if hadError {
		return SilentExit()
	}
	return nil
}

func runGraphPreviewCloseClaimNext(paths, targets, reasons []string, session string, force, suggestNext bool) error {
	var hadError bool
	var claimedID string
	var firstSettled string
	err := withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		perPathReasons := make([]string, len(paths))
		for i := range paths {
			perPathReasons[i] = reasonForCloseIndex(reasons, i)
		}
		result, err := store.CloseIssues(ctx, graphstore.IssueCloseBatchRequest{
			Paths: paths, Reasons: perPathReasons, Actor: getActorWithGit(), Session: session, Force: force, ClaimNext: true,
		})
		if err != nil {
			return nil, "", err
		}
		closed := make([]graphstore.IssueRecord, 0, len(paths))
		var human strings.Builder
		for i, outcome := range result.Outcomes {
			if outcome.Err != nil {
				if len(paths) == 1 {
					return nil, "", outcome.Err
				}
				_ = graphStorageError(fmt.Errorf("closing %s: %w", targets[i], outcome.Err))
				hadError = true
				continue
			}
			closed = append(closed, outcome.Issue)
			if firstSettled == "" {
				firstSettled = paths[i]
			}
			if outcome.OpenChildren > 0 {
				fmt.Fprintf(os.Stderr, "warning: closing %s with %d open child issue(s) still active\n", targets[i], outcome.OpenChildren)
			}
			if outcome.Changed {
				fmt.Fprintf(&human, "Closed %s\n", targets[i])
			} else {
				fmt.Fprintf(&human, "Already closed %s\n", targets[i])
			}
		}
		var output any = closed
		if result.ClaimedNext != nil {
			claimedID = result.ClaimedNext.ID
			output = map[string]any{"closed": closed, "claimed": result.ClaimedNext}
			fmt.Fprintf(&human, "Claimed next ready Issue %s\n", result.ClaimedNext.ID)
		} else if len(closed) > 0 {
			human.WriteString("No ready Issues available to claim.\n")
		}
		if suggestNext && len(closed) > 0 {
			unblocked, err := store.NewlyUnblockedByClose(ctx, paths[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not find newly unblocked Issues: %v\n", err)
			} else if len(unblocked) > 0 {
				fields := map[string]any{"closed": closed, "unblocked": unblocked}
				if result.ClaimedNext != nil {
					fields["claimed"] = result.ClaimedNext
				}
				output = fields
				human.WriteString("Newly unblocked:\n")
				for _, issue := range unblocked {
					fmt.Fprintf(&human, "  %s (P%d)\n", issue.ID, issue.Properties.Priority)
				}
			}
		}
		return output, strings.TrimSuffix(human.String(), "\n"), nil
	})
	if err != nil {
		return err
	}
	if claimedID != "" {
		SetLastTouchedID(claimedID)
	} else {
		SetLastTouchedID(firstSettled)
	}
	if hadError {
		if claimedID != "" {
			_ = graphFailure("partial_close_claim_retained", "--claim-next already claimed "+claimedID+"; it remains assigned despite the sibling refusal", 1)
		}
		return SilentExit()
	}
	return nil
}

func runGraphPreviewDeferral(cmd *cobra.Command, args []string, deferred bool) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	allowed := []string{"if-revision"}
	if deferred {
		allowed = append(allowed, "until", "reason")
	}
	if err := graphPreviewFlags(cmd, allowed...); err != nil {
		return err
	}
	if len(args) == 0 {
		return graphFailure("invalid_selector", "graph deferral requires at least one Issue ID or beads/PATH", 2)
	}
	if len(args) > 1 && cmd.Flags().Changed("if-revision") {
		return graphFailure("invalid_selector", "--if-revision applies to one Issue; use separate commands for guarded updates", 2)
	}
	paths := make([]string, len(args))
	for i, selector := range args {
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if err := graph.ValidateBeadPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		paths[i] = path
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false)
	if err != nil {
		return err
	}
	var until *time.Time
	var reason string
	if deferred {
		raw, _ := cmd.Flags().GetString("until")
		if raw != "" {
			parsed, err := timeparsing.ParseRelativeTime(raw, time.Now())
			if err != nil {
				return graphFailure("invalid_properties", fmt.Sprintf("invalid --until format %q. %s", raw, deferUntilFormatHint), 2)
			}
			until = &parsed
			if parsed.Before(time.Now()) && !jsonOutput {
				fmt.Fprintf(os.Stderr, "%s Defer date %q is in the past. Issue will appear in bd ready immediately.\n",
					ui.RenderWarn("!"), parsed.Local().Format("2006-01-02 15:04"))
				fmt.Fprintln(os.Stderr, "  Did you mean a future date? Use --until=+1h or --until=tomorrow")
			}
		}
		reason, _ = cmd.Flags().GetString("reason")
		reason = strings.TrimSpace(reason)
		if cmd.Flags().Changed("reason") && reason == "" {
			return graphFailure("invalid_properties", "reason cannot be empty", 2)
		}
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		results := make([]graphstore.IssueMutationResult, 0, len(paths))
		var human strings.Builder
		for i, path := range paths {
			result, err := store.SetIssueDeferred(ctx, graphstore.IssueDeferralRequest{
				Path: path, Actor: getActorWithGit(), ExpectedRevision: revision,
				Unconditional: unconditional, Deferred: deferred, Until: until, Reason: reason,
			})
			if err != nil {
				if len(paths) == 1 {
					return nil, "", err
				}
				fmt.Fprintf(os.Stderr, "Error changing deferral for %s: %v\n", args[i], err)
				continue
			}
			results = append(results, result)
			verb := "Undeferred"
			if deferred {
				verb = "Deferred"
			}
			if !result.Changed {
				verb = "Unchanged"
			}
			fmt.Fprintf(&human, "%s %s\n", verb, result.Issue.ID)
		}
		if len(paths) == 1 {
			return results[0], strings.TrimSuffix(human.String(), "\n"), nil
		}
		return results, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

func runGraphPreviewReady(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "limit", "priority", "assignee", "unassigned", "sort", "label", "label-any", "exclude-label", "label-pattern", "label-regex", "type", "exclude-type", "metadata-field", "has-metadata-key", "include-deferred", "claim"); err != nil {
		return err
	}
	if len(args) != 0 {
		return graphFailure("invalid_selector", "graph ready takes no positional arguments", 2)
	}
	if err := graphPreviewLabelFilters(cmd); err != nil {
		return err
	}
	in, err := gatherReadyInput(cmd, nil)
	if err != nil {
		return err
	}
	// The legacy resolver warns and ignores malformed environment values.
	// This deliberately bounded preview refuses unsupported policy instead.
	if raw := os.Getenv(maxRowsEnvVar); raw != "" {
		maxRows, err := strconv.Atoi(raw)
		if err != nil || maxRows < 0 {
			return graphFailure("invalid_properties", "BEADS_MAX_ROWS must be a non-negative integer", 2)
		}
		if maxRows > 0 && !in.claim {
			return graphFailure("capability_unavailable", "graph ready preview does not implement a configured row cap; unset BEADS_MAX_ROWS to request the unfiltered view", 5)
		}
	}
	if in.claim {
		if err := graphPreviewWritePolicy(); err != nil {
			return err
		}
		return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
			claimed, err := store.ClaimReadyIssue(ctx, getActorWithGit(), in.filter)
			if err != nil {
				return nil, "", err
			}
			if claimed == nil {
				return []graphstore.IssueRecord{}, "No ready work to claim", nil
			}
			SetLastTouchedID(claimed.Issue.ID)
			return []graphstore.IssueRecord{claimed.Issue}, "Claimed Issue " + claimed.Issue.ID, nil
		})
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		issues, err := store.ReadyIssuesFiltered(ctx, in.filter)
		if err != nil {
			return nil, "", err
		}
		var human strings.Builder
		for _, issue := range issues {
			fmt.Fprintf(&human, "%s  %s\n", issue.ID, issue.Properties.Title)
		}
		if len(issues) == 0 {
			human.WriteString("No ready Issues.\n")
		}
		return issues, strings.TrimSuffix(human.String(), "\n"), nil
	})
}

// Keep graph admission from turning a supplied blank label filter into an
// unfiltered query. The ordinary fix is tracked separately in upstream #6632;
// this graph route must retain the same distinction until that PR lands.
func graphPreviewLabelFilters(cmd *cobra.Command) error {
	for _, name := range []string{"label", "label-any", "exclude-label"} {
		if !cmd.Flags().Changed(name) {
			continue
		}
		values, err := cmd.Flags().GetStringSlice(name)
		if err != nil {
			return graphFailure("invalid_properties", "invalid --"+name+": "+err.Error(), 2)
		}
		if len(utils.NormalizeLabels(values)) == 0 {
			return graphFailure("invalid_properties", "--"+name+" was supplied but contains no usable label", 2)
		}
	}
	return nil
}
