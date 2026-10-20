package clioutput

import (
	"encoding/json"
	"fmt"
)

// PagingOptions controls delivery, not selection or terminal paging. Zero-valued
// PageSize requests the default of 50; Limit zero means no total cap. Adapters
// distinguish an explicit --page-size=0 and reject it before constructing this.
type PagingOptions struct {
	Paged    bool
	Unpaged  bool
	PageSize int
	Limit    int
}

// Paging is an immutable resolved delivery policy. Providers retain ownership of
// requests, read context and continuation encoding; no query IR is introduced.
type Paging struct {
	paged    bool
	pageSize int
	limit    int
}

// ResolvePaging applies JSONL's unpaged default and ordinary output's page default.
func ResolvePaging(options PagingOptions, format string) (Paging, error) {
	if options.Paged && options.Unpaged {
		return Paging{}, fmt.Errorf("paged and unpaged conflict")
	}
	if options.PageSize < 0 || options.Limit < 0 {
		return Paging{}, fmt.Errorf("negative paging bounds are invalid")
	}
	switch format {
	case "default", "detail", "table", "json", "jsonl":
	default:
		return Paging{}, fmt.Errorf("unsupported paging format %q", format)
	}
	pageSize := options.PageSize
	if pageSize == 0 {
		pageSize = 50
	}
	return Paging{paged: options.Paged || (!options.Unpaged && format != "jsonl"), pageSize: pageSize, limit: options.Limit}, nil
}

// RequestSize returns the maximum next fetch size given delivered items. It can
// reduce a request but cannot promise the provider honors that exact boundary.
func (p Paging) RequestSize(returned int) (int, error) {
	if returned < 0 {
		return 0, fmt.Errorf("returned count cannot be negative")
	}
	size := p.pageSize
	if p.limit != 0 {
		remaining := p.limit - returned
		if remaining < 0 {
			return 0, fmt.Errorf("returned count exceeds limit")
		}
		if remaining < size {
			size = remaining
		}
	}
	return size, nil
}

// PageDelivery describes how much of a provider page can be delivered. A blank
// StopReason means fetch another page. ContinuationUnavailable explicitly marks
// a truncated page whose exact delivered boundary is not resumable. It must not
// be replaced by the provider's next-page token, which would skip unseen items.
type PageDelivery struct {
	Take                    int
	Exhausted               bool
	StopReason              string
	Continuation            json.RawMessage
	ContinuationUnavailable bool
}

// PlanPage never infers exhaustion from page length. next is the provider's
// continuation AFTER its full page; deliveredBoundary, when supplied by the
// provider, resumes after the smaller Take boundary. Nil is unavailable, not a
// special protocol token. The caller must update counts only after output succeeds.
func (p Paging) PlanPage(returned, available int, exhausted bool, next, deliveredBoundary json.RawMessage) (PageDelivery, error) {
	if available < 0 {
		return PageDelivery{}, fmt.Errorf("page item count cannot be negative")
	}
	if _, err := p.RequestSize(returned); err != nil {
		return PageDelivery{}, err
	}
	for _, token := range []json.RawMessage{next, deliveredBoundary} {
		if token != nil && !json.Valid(token) {
			return PageDelivery{}, fmt.Errorf("continuation must be a protocol JSON value")
		}
	}
	take := available
	if p.limit != 0 && take > p.limit-returned {
		take = p.limit - returned
	}
	result := PageDelivery{Take: take}
	if take < available {
		result.StopReason = "limit"
		result.Continuation = append(json.RawMessage(nil), deliveredBoundary...)
		result.ContinuationUnavailable = deliveredBoundary == nil
		return result, nil
	}
	result.Exhausted = exhausted
	result.Continuation = append(json.RawMessage(nil), next...)
	switch {
	case exhausted:
		result.StopReason = "exhausted"
	case p.limit != 0 && returned+take == p.limit:
		result.StopReason = "limit"
	case p.paged:
		result.StopReason = "page"
	}
	result.ContinuationUnavailable = !exhausted && next == nil
	return result, nil
}
