package workapi

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

type completeListReaderFunc func(context.Context, issueops.ListRequest) (issueops.IssuePage, error)

func (f completeListReaderFunc) List(ctx context.Context, request issueops.ListRequest) (issueops.IssuePage, error) {
	return f(ctx, request)
}

func TestReadCompleteListPreservesSelectionAndClearsDelivery(t *testing.T) {
	limit, priority := 1, 2
	instant := time.Unix(100, 0)
	request := issueops.ListRequest{
		Status: "in_progress", ReadyFlag: true, ParentID: "parent", NoPinnedFlag: true,
		Labels: []string{"release"}, IncludeEphemeral: true, IncludeInfra: true,
		MetadataFields: map[string]string{"n": "1"}, HasMetadataKey: "marker",
		SortBy: "priority", Reverse: true, Limit: &limit, Offset: 3,
		AfterCreatedAt: &instant, AfterID: "earlier", AfterPriority: &priority,
		MaxRows: 100, MaxRowsSource: "--max-rows", IncludeComments: true,
	}
	want := request
	zero := 0
	want.Limit, want.Offset = &zero, 0
	want.AfterCreatedAt, want.AfterID, want.AfterPriority = nil, "", nil
	items := []*types.IssueWithCounts{{Issue: &types.Issue{ID: "a"}}, {Issue: &types.Issue{ID: "b"}}}
	calls := 0
	reader := completeListReaderFunc(func(_ context.Context, got issueops.ListRequest) (issueops.IssuePage, error) {
		calls++
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("request changed selection or cap: got %#v want %#v", got, want)
		}
		return issueops.IssuePage{Items: items}, nil
	})
	page, err := ReadCompleteList(context.Background(), reader, request)
	if err != nil || calls != 1 || !reflect.DeepEqual(page.Items, items) || page.HasMore {
		t.Fatalf("complete selection: %#v calls=%d err=%v", page, calls, err)
	}
	if *request.Limit != 1 || request.Offset != 3 || request.AfterID != "earlier" || request.AfterCreatedAt != &instant {
		t.Fatal("caller request was mutated")
	}
}

func TestReadCompleteListRefusesIncompleteOrMalformedPage(t *testing.T) {
	row := &types.IssueWithCounts{Issue: &types.Issue{ID: "a"}}
	for _, tc := range []struct {
		name string
		page issueops.IssuePage
	}{
		{"truncated", issueops.IssuePage{Items: []*types.IssueWithCounts{row}, HasMore: true}},
		{"nil row", issueops.IssuePage{Items: []*types.IssueWithCounts{nil}}},
		{"nil issue", issueops.IssuePage{Items: []*types.IssueWithCounts{{}}}},
		{"empty ID", issueops.IssuePage{Items: []*types.IssueWithCounts{{Issue: &types.Issue{}}}}},
		{"duplicate ID", issueops.IssuePage{Items: []*types.IssueWithCounts{row, row}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := completeListReaderFunc(func(context.Context, issueops.ListRequest) (issueops.IssuePage, error) { return tc.page, nil })
			page, err := ReadCompleteList(context.Background(), reader, issueops.ListRequest{})
			if !errors.Is(err, ErrIncompleteListSelection) || page.Items != nil || page.HasMore {
				t.Fatalf("partial exact selection escaped: %#v %v", page, err)
			}
		})
	}
}

func TestReadCompleteListPreservesReaderErrorAndNoResult(t *testing.T) {
	failure := errors.New("row cap exceeded")
	reader := completeListReaderFunc(func(context.Context, issueops.ListRequest) (issueops.IssuePage, error) {
		return issueops.IssuePage{Items: []*types.IssueWithCounts{{Issue: &types.Issue{ID: "a"}}}}, failure
	})
	page, err := ReadCompleteList(context.Background(), reader, issueops.ListRequest{})
	if err != failure || page.Items != nil {
		t.Fatalf("reader error was replaced or partial result returned: %#v %v", page, err)
	}
}

func TestReadCompleteListCancellation(t *testing.T) {
	for _, afterRead := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		if !afterRead {
			cancel()
		}
		reader := completeListReaderFunc(func(context.Context, issueops.ListRequest) (issueops.IssuePage, error) {
			calls++
			cancel()
			return issueops.IssuePage{Items: []*types.IssueWithCounts{{Issue: &types.Issue{ID: "a"}}}}, nil
		})
		page, err := ReadCompleteList(ctx, reader, issueops.ListRequest{})
		cancel()
		if !errors.Is(err, context.Canceled) || page.Items != nil || (!afterRead && calls != 0) || (afterRead && calls != 1) {
			t.Fatalf("cancel afterRead=%v: %#v calls=%d err=%v", afterRead, page, calls, err)
		}
	}
}
