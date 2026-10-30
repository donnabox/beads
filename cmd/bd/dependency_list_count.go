package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	storageissueops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/workapi"
	"github.com/steveyegge/beads/issueops"
)

// Bind the already-initialized command during init rather than referring
// back to listCmd from its RunE dependency graph. Keep exact leaf identity;
// Link List and other commands must not enter the dependency List adapter.
var dependencyListCountLeaf *cobra.Command

func init() {
	dependencyListCountLeaf = listCmd
}

// Admit a changed count flag even when its final scalar value is false: the
// occurrence validator must see true/false conflicts before route selection.
// An explicit false alone still cannot enable dependency Resource output.
func dependencyListCountRequested(cmd *cobra.Command) bool {
	if cmd == nil || cmd != dependencyListCountLeaf || !cmd.Flags().Changed("count-only") || collectionASTSelected(cmd) {
		return false
	}
	return !cmd.Flags().Changed("property-field") && !cmd.Flags().Changed("has-property-key")
}

func dependencyListCountSelected(cmd *cobra.Command) bool {
	return dependencyListCountRequested(cmd) && collectionCountSelected(cmd)
}

func dependencyListCountRenderer(cmd *cobra.Command) (*clioutput.Renderer, error) {
	return dependencyListCountRendererWithArgs(cmd, os.Args[1:])
}

func dependencyListCountRendererWithArgs(cmd *cobra.Command, argv []string) (*clioutput.Renderer, error) {
	failure := canonicalFailureRendererFor(cmd)
	options, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return nil, dependencyListCountFailure(cmd, failure, errors.Join(resourcequery.ErrInvalid, err))
	}
	// Validate Resource masks before discarding their aggregate effect.
	descriptor, err := canonicalCollectionDescriptor(false, nil)
	if err != nil {
		return nil, dependencyListCountFailure(cmd, failure, err)
	}
	if _, err := clioutput.NewCollection(descriptor, options); err != nil {
		return nil, dependencyListCountFailure(cmd, failure, errors.Join(resourcequery.ErrInvalid, err))
	}
	renderer, err := clioutput.New(canonicalCollectionCountDescriptor(), canonicalCollectionCountOptions(options))
	if err != nil {
		return nil, dependencyListCountFailure(cmd, failure, errors.Join(resourcequery.ErrInvalid, err))
	}
	failure = renderer
	// Reuse the canonical equal-repeat/conflict policy, including presence and
	// count scalars. This parses occurrences before consuming last-value flags.
	if err := canonicalScalarOccurrences(cmd, argv, "type", "exact-type", "conforms-to", "paged", "unpaged", "page-size", "limit", "continue", "has-metadata-key", "count-only"); err != nil {
		return nil, dependencyListCountFailure(cmd, failure, errors.Join(resourcequery.ErrInvalid, err))
	}
	if !collectionCountSelected(cmd) {
		return nil, dependencyListCountFailure(cmd, failure, fmt.Errorf("%w: dependency Resource output is not provided by --count-only=false", graphstore.ErrCapabilityUnavailable))
	}
	// No canonical family authority is fabricated from native issue_type text.
	for _, name := range []string{"type", "exact-type", "conforms-to"} {
		if cmd.Flags().Changed(name) {
			return nil, dependencyListCountFailure(cmd, failure, fmt.Errorf("%w: dependency List count lacks canonical Type authority for --%s", graphstore.ErrCapabilityUnavailable, name))
		}
	}
	if _, err := collectionFieldConditions(cmd); err != nil {
		// In particular, ErrLimit is operational refusal, not invalid syntax.
		return nil, canonicalCollectionRequestFailure(cmd, failure, err)
	}
	values, err := collectionFieldArguments(cmd, "metadata-field")
	if err != nil {
		return nil, dependencyListCountFailure(cmd, failure, errors.Join(resourcequery.ErrInvalid, err))
	}
	metadata := map[string]string{}
	for _, value := range values {
		key, text, _ := strings.Cut(value, "=")
		if prior, present := metadata[key]; present && prior != text {
			return nil, dependencyListCountFailure(cmd, failure, fmt.Errorf("%w: dependency List count cannot yet represent conflicting repeated metadata equalities", graphstore.ErrCapabilityUnavailable))
		}
		metadata[key] = text
	}
	if len(cmd.Flags().Args()) != 0 {
		return nil, dependencyListCountFailure(cmd, failure, fmt.Errorf("%w: List count does not accept positional arguments", resourcequery.ErrInvalid))
	}
	if cmd.Flags().Changed("continue") {
		return nil, dependencyListCountFailure(cmd, failure, fmt.Errorf("%w: exact List count cannot resume a delivery continuation", graphstore.ErrCapabilityUnavailable))
	}
	paged, _ := cmd.Flags().GetBool("paged")
	unpaged, _ := cmd.Flags().GetBool("unpaged")
	pageSize, _ := cmd.Flags().GetInt("page-size")
	limit, _ := cmd.Flags().GetInt("limit")
	if pageSize <= 0 {
		return nil, dependencyListCountFailure(cmd, failure, fmt.Errorf("%w: --page-size must be positive", resourcequery.ErrInvalid))
	}
	format := "default"
	if len(options.Formats) > 0 {
		format = options.Formats[0]
	}
	if _, err := clioutput.ResolvePaging(clioutput.PagingOptions{Paged: paged, Unpaged: unpaged, PageSize: pageSize, Limit: limit}, format); err != nil {
		return nil, dependencyListCountFailure(cmd, failure, errors.Join(resourcequery.ErrInvalid, err))
	}
	return renderer, nil
}

// Every new aggregate failure passes one unrendered error boundary. Preserve
// typed details while using the same canonical collection/query classifiers.
func dependencyListCountFailure(cmd *cobra.Command, failure canonicalFailureRenderer, err error) error {
	var capErr *storageissueops.ErrTooManyRows
	var unsupported *issueops.ErrUnsupported
	var proxy *ProxyCapabilityError
	switch {
	case errors.Is(err, workapi.ErrIncompleteListSelection):
		err = errors.Join(graphstore.ErrInvalidStore, err)
	case errors.As(err, &capErr), errors.As(err, &unsupported), errors.As(err, &proxy):
		err = errors.Join(graphstore.ErrCapabilityUnavailable, err)
	case errors.Is(err, issueops.ErrValidation), errors.Is(err, storage.ErrValidation):
		err = errors.Join(resourcequery.ErrInvalid, err)
	}
	return canonicalReadError(cmd, failure, canonicalCollectionEvaluationError(err), "route_unavailable")
}

func gatherDependencyListCountInput(cmd *cobra.Command, failure canonicalFailureRenderer) (listInput, error) {
	invalid := func(format string, args ...interface{}) error {
		return fmt.Errorf("%w: %s", issueops.ErrValidation, fmt.Sprintf(format, args...))
	}
	in, err := gatherListInputWithFailures(cmd, listInputFailures{
		plain: invalid, respectJSON: invalid,
		hint: func(message, hint string) error { return invalid("%s; hint: %s", message, hint) },
		skipLabels: func(conflicts []string) error {
			return invalid("%s", strings.TrimSpace(formatSkipLabelsConflictError(conflicts)))
		},
	})
	if err != nil {
		return listInput{}, dependencyListCountFailure(cmd, failure, err)
	}
	return in, nil
}

func runDependencyListCount(cmd *cobra.Command) error {
	renderer, err := dependencyListCountRenderer(cmd)
	if err != nil {
		return err
	}
	in, err := gatherDependencyListCountInput(cmd, renderer)
	if err != nil {
		return err
	}
	if in.watchMode || (in.prettyFormat && in.ParentID != "" && !in.ReadyFlag) || in.depsMode != "" {
		return dependencyListCountFailure(cmd, renderer, fmt.Errorf("%w: dependency List count does not yet compose with watch or hierarchical tree/dependency rendering; use a flat List selection", graphstore.ErrCapabilityUnavailable))
	}
	return executeDependencyListCount(cmd, renderer, in.ListRequest, func(ctx context.Context) (issueops.Reader, func() error, error) {
		return openDependencyListCountReader(ctx, in)
	})
}

func openDependencyListCountReader(ctx context.Context, in listInput) (issueops.Reader, func() error, error) {
	if usesProxiedServer() {
		if in.repoOverrideSet {
			return nil, nil, AssertProxyCapability(ProxyModeProxied, ProxyCapRepo)
		}
		reader, err := proxiedIssueReader()
		return reader, nil, err
	}
	activeStore := store
	routedStore, routed, routingRule, err := openRoutedReadStore(ctx, activeStore)
	if err != nil {
		return nil, nil, err
	}
	var cleanup func() error
	if routed {
		activeStore = routedStore
		cleanup = routedStore.Close
		printContributorRoutingNotice(ctx, store, routingRule)
	}
	reader, err := activeStore.IssueReader()
	return reader, cleanup, err
}

// The small acquisition callback is a test seam for routing/accessor failures,
// not a second provider interface. Production still uses the existing accessor.
func executeDependencyListCount(cmd *cobra.Command, renderer *clioutput.Renderer, request issueops.ListRequest,
	acquire func(context.Context) (issueops.Reader, func() error, error)) error {
	ctx := getRootContext()
	reader, cleanup, err := acquire(ctx)
	var page issueops.IssuePage
	if err == nil {
		page, err = workapi.ReadCompleteList(ctx, reader, request)
	}
	if cleanup != nil {
		err = errors.Join(err, cleanup())
	}
	if err != nil {
		return dependencyListCountFailure(cmd, renderer, err)
	}
	return renderCanonicalResult(cmd, renderer, struct {
		Count int `json:"count"`
	}{Count: len(page.Items)})
}
