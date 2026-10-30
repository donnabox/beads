package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	storageissueops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

func TestDependencyListCountNarrowAdmission(t *testing.T) {
	old, oldLeaf := listCmd, dependencyListCountLeaf
	t.Cleanup(func() { listCmd, dependencyListCountLeaf = old, oldLeaf })
	for _, tc := range []struct {
		args                []string
		requested, selected bool
	}{
		{[]string{"--count-only"}, true, true},
		{[]string{"--count-only", "--metadata-field", "n=1"}, true, true},
		{[]string{"--count-only", "--property-field", "title=one"}, false, false},
		{[]string{"--count-only", "--has-property-key", "title"}, false, false},
		{[]string{"--count-only", "--ast", "{}"}, false, false},
		{[]string{"--count-only=false"}, true, false},
		{[]string{"--output-format", "json"}, false, false},
	} {
		cmd, _, _ := canonicalCollectionTestCommand(t, false, tc.args...)
		listCmd, dependencyListCountLeaf = cmd, cmd
		if got := dependencyListCountRequested(cmd); got != tc.requested {
			t.Fatalf("requested %q=%v want %v", tc.args, got, tc.requested)
		}
		if got := dependencyListCountSelected(cmd); got != tc.selected {
			t.Fatalf("selected %q=%v want %v", tc.args, got, tc.selected)
		}
		other, _, _ := canonicalCollectionTestCommand(t, false, tc.args...)
		if dependencyListCountRequested(other) || dependencyListCountSelected(other) {
			t.Fatal("a distinct same-name command entered the registered List adapter")
		}
	}
}

func TestDependencyListCountRejectsUncapturedTypeAuthority(t *testing.T) {
	for _, flag := range []string{"--type", "--exact-type", "--conforms-to"} {
		args := []string{"--count-only", "--output-format", "json", flag, "issue"}
		cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
		_, err := dependencyListCountRendererWithArgs(cmd, args)
		assertDependencyCountFailure(t, err, out.String(), "capability_unavailable", 5, "Type authority")
	}
}

func TestDependencyListCountRendererIgnoresResourceMasksAfterValidation(t *testing.T) {
	args := []string{"--count-only", "--output-format", "json", "--include", "/properties/title", "--omit", "envelope", "--limit", "1", "--page-size", "1"}
	cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
	renderer, err := dependencyListCountRendererWithArgs(cmd, args)
	if err != nil {
		t.Fatal(err)
	}
	if err := renderCanonicalResult(cmd, renderer, struct {
		Count int `json:"count"`
	}{Count: 3}); err != nil {
		t.Fatal(err)
	}
	var value struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(out.Bytes(), &value); err != nil || value.Count != 3 {
		t.Fatalf("aggregate followed Resource mask/page: %s %v", out.String(), err)
	}
}

func TestDependencyListCountRefusesMetadataMapLossBeforeRead(t *testing.T) {
	args := []string{"--count-only", "--output-format", "json", "--metadata-field", "k=a", "--metadata-field", "k=b"}
	cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
	_, err := dependencyListCountRendererWithArgs(cmd, args)
	assertDependencyCountFailure(t, err, out.String(), "capability_unavailable", 5, "conflicting repeated metadata")
}

func TestDependencyListCountScalarOccurrences(t *testing.T) {
	for _, tc := range []struct {
		extra  []string
		bad    bool
		detail string
	}{
		{[]string{"--has-metadata-key", "a", "--has-metadata-key", "b"}, true, "conflicting --has-metadata-key"},
		{[]string{"--has-metadata-key", "a", "--has-metadata-key", "a"}, false, ""},
		{[]string{"--count-only=false"}, true, "conflicting --count-only"},
		{[]string{"--count-only=true"}, false, ""},
		{[]string{"--page-size=1", "--page-size=2"}, true, "conflicting --page-size"},
		{[]string{"--limit=1", "--limit=1"}, false, ""},
	} {
		args := append([]string{"--count-only", "--output-format", "json"}, tc.extra...)
		cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
		_, err := dependencyListCountRendererWithArgs(cmd, args)
		if tc.bad {
			assertDependencyCountFailure(t, err, out.String(), "invalid_selector", 2, tc.detail)
		} else if err != nil || out.Len() != 0 {
			t.Fatalf("identical scalar repetitions changed selection: %q %s %v", tc.extra, out.String(), err)
		} else if cmd.Flags().Changed("has-metadata-key") {
			value, _ := cmd.Flags().GetString("has-metadata-key")
			if value != "a" {
				t.Fatalf("equal presence predicate lost: %q", value)
			}
		}
	}
}

func TestDependencyListCountFieldBudgetClassification(t *testing.T) {
	for _, largeValue := range []bool{false, true} {
		args := []string{"--count-only", "--output-format", "json"}
		if largeValue {
			args = append(args, "--metadata-field", "key="+strings.Repeat("x", 16385))
		} else {
			for i := 0; i < 257; i++ {
				args = append(args, "--metadata-field", fmt.Sprintf("k%d=v", i))
			}
		}
		cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
		_, err := dependencyListCountRendererWithArgs(cmd, args)
		assertDependencyCountFailure(t, err, out.String(), "capability_unavailable", 5, "List field")
	}
}

// Only List is used. Ready/Get make the mock satisfy the existing role without
// creating a replacement provider or allowing a Counter fallback.
type dependencyCountReaderFunc func(context.Context, issueops.ListRequest) (issueops.IssuePage, error)

func (f dependencyCountReaderFunc) List(ctx context.Context, request issueops.ListRequest) (issueops.IssuePage, error) {
	return f(ctx, request)
}
func (f dependencyCountReaderFunc) Ready(context.Context, issueops.ReadyRequest) (issueops.IssuePage, error) {
	return issueops.IssuePage{}, errors.New("unexpected Ready call")
}
func (f dependencyCountReaderFunc) Get(context.Context, issueops.GetRequest) (*issueops.IssueDetails, error) {
	return nil, errors.New("unexpected Get call")
}

func TestDependencyListCountTypedCapHasCanonicalFailureAndNoCount(t *testing.T) {
	args := []string{"--count-only", "--output-format", "json"}
	cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
	renderer, err := dependencyListCountRendererWithArgs(cmd, args)
	if err != nil {
		t.Fatal(err)
	}
	calls, closes := 0, 0
	reader := dependencyCountReaderFunc(func(_ context.Context, req issueops.ListRequest) (issueops.IssuePage, error) {
		calls++
		if req.MaxRows != 2 || req.MaxRowsSource != "--max-rows" || req.Limit == nil || *req.Limit != 0 {
			t.Fatalf("count lost exact cap/request: %#v", req)
		}
		return issueops.IssuePage{Items: []*types.IssueWithCounts{{Issue: &types.Issue{ID: "partial"}}}}, &storageissueops.ErrTooManyRows{Found: 3, Cap: 2, Source: "--max-rows"}
	})
	err = executeDependencyListCount(cmd, renderer, issueops.ListRequest{MaxRows: 2, MaxRowsSource: "--max-rows"}, func(context.Context) (issueops.Reader, func() error, error) {
		return reader, func() error { closes++; return nil }, nil
	})
	assertDependencyCountFailure(t, err, out.String(), "capability_unavailable", 5, "search returned 3 rows, exceeding --max-rows cap of 2")
	if calls != 1 || closes != 1 {
		t.Fatalf("read/cleanup calls = %d/%d", calls, closes)
	}
}

func TestDependencyListCountAcquisitionAndCleanupFailuresRenderOnce(t *testing.T) {
	for _, tc := range []struct {
		stage                string
		acquireErr, closeErr error
	}{
		{"routing", errors.New("routing read failed"), nil},
		{"accessor", errors.New("reader accessor failed"), nil},
		{"close after read", nil, errors.New("routed close failed")},
		{"accessor and close", errors.New("reader accessor failed"), errors.New("routed close failed")},
	} {
		t.Run(tc.stage, func(t *testing.T) {
			args := []string{"--count-only", "--output-format", "json"}
			cmd, out, diagnostics := canonicalCollectionTestCommand(t, false, args...)
			renderer, err := dependencyListCountRendererWithArgs(cmd, args)
			if err != nil {
				t.Fatal(err)
			}
			calls, closes := 0, 0
			reader := dependencyCountReaderFunc(func(context.Context, issueops.ListRequest) (issueops.IssuePage, error) {
				calls++
				return issueops.IssuePage{Items: []*types.IssueWithCounts{{Issue: &types.Issue{ID: "a"}}}}, nil
			})
			err = executeDependencyListCount(cmd, renderer, issueops.ListRequest{}, func(context.Context) (issueops.Reader, func() error, error) {
				return reader, func() error { closes++; return tc.closeErr }, tc.acquireErr
			})
			detail := tc.closeErr
			if tc.acquireErr != nil {
				detail = tc.acquireErr
			}
			assertDependencyCountFailure(t, err, out.String(), "route_unavailable", 5, detail.Error())
			if tc.closeErr != nil && !strings.Contains(out.String(), tc.closeErr.Error()) {
				t.Fatalf("close detail lost: %s", out.String())
			}
			wantCalls := 1
			if tc.acquireErr != nil {
				wantCalls = 0
			}
			if calls != wantCalls || closes != 1 || strings.Count(diagnostics.String(), "route_unavailable:") != 1 {
				t.Fatalf("failure/custody duplicated: reads=%d closes=%d stderr=%s", calls, closes, diagnostics.String())
			}
		})
	}
}

func TestDependencyListCountTypedOperationalAndValidationErrors(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code string
		exit int
	}{
		{resourcequery.ErrLimit, "capability_unavailable", 5},
		{context.Canceled, "capability_unavailable", 5},
		{&issueops.ErrUnsupported{Op: "List", Backend: "test"}, "capability_unavailable", 5},
		{&ProxyCapabilityError{Code: "proxy.test", Message: "proxy unavailable", ExitCode: 1}, "capability_unavailable", 5},
		{issueops.ErrValidation, "invalid_selector", 2},
		{graphstore.ErrInvalidStore, "invalid_store", 5},
	} {
		args := []string{"--count-only", "--output-format", "json"}
		cmd, out, _ := canonicalCollectionTestCommand(t, false, args...)
		renderer, err := dependencyListCountRendererWithArgs(cmd, args)
		if err != nil {
			t.Fatal(err)
		}
		err = dependencyListCountFailure(cmd, renderer, tc.err)
		assertDependencyCountFailure(t, err, out.String(), tc.code, tc.exit, tc.err.Error())
	}
}

func TestDependencyListCountInputValidationRendersExactlyOnce(t *testing.T) {
	t.Setenv(maxRowsEnvVar, "")
	previousJSON := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = previousJSON })
	for _, tc := range []struct {
		flags  []string
		detail string
	}{
		{[]string{"--parent", "a", "--no-parent"}, "--parent and --no-parent are mutually exclusive"},
		{[]string{"--sort", "nonsense"}, "invalid sort field"},
		{[]string{"--max-rows=-1"}, "--max-rows must be non-negative"},
		{[]string{"--created-after", "not-a-time"}, "parsing --created-after"},
	} {
		cmd, out, diagnostics := canonicalCollectionTestCommand(t, false)
		cmd.Flags().String("parent", "", "")
		cmd.Flags().Bool("no-parent", false, "")
		cmd.Flags().String("sort", "", "")
		cmd.Flags().String("created-after", "", "")
		addRoutedMaxRowsFlag(cmd)
		args := append([]string{"--count-only", "--output-format", "json"}, tc.flags...)
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatal(err)
		}
		renderer, err := dependencyListCountRendererWithArgs(cmd, args)
		if err != nil {
			t.Fatal(err)
		}
		_, err = gatherDependencyListCountInput(cmd, renderer)
		assertDependencyCountFailure(t, err, out.String(), "invalid_selector", 2, tc.detail)
		if strings.Count(diagnostics.String(), "invalid_selector:") != 1 || strings.Contains(diagnostics.String(), "Error:") {
			t.Fatalf("legacy and canonical parser failures mixed: %s", diagnostics.String())
		}
	}
}

func assertDependencyCountFailure(t *testing.T, err error, document, code string, exitCode int, detail string) {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(document))
	var failure struct {
		Error struct {
			Problem struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"problem"`
		} `json:"error"`
	}
	if err := decoder.Decode(&failure); err != nil {
		t.Fatalf("not a canonical JSON error: %v %s", err, document)
	}
	var exit *exitError
	if !errors.As(err, &exit) || exit.Code != exitCode || failure.Error.Problem.Code != code || !strings.Contains(failure.Error.Problem.Message, detail) || strings.Contains(document, `"count":`) {
		t.Fatalf("wrong failure or partial count: exit=%v expected=%s/%d detail=%q output=%s", err, code, exitCode, detail, document)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("multiple failure documents or trailing output: %v %s", err, document)
	}
}
