package main

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// graphPreviewReference is a CLI selector, not a stored BDP Reference. In
// particular, parsing a pin must not make an unpinned graphstore Link look
// pinned: persistence and historical validation are separate writer work.
type graphPreviewReference struct {
	Identity string
	Revision string
	Explicit bool
	Wrapped  bool
}

// graphPreviewParseReference keeps every unwrapped absolute URI literal. This
// includes @ in userinfo and path, and IPv6 authority brackets. A URI with a
// revision uses [URI]@revision; nested IPv6 brackets are balanced here.
func graphPreviewParseReference(input string) (graphPreviewReference, error) {
	if input == "" || !utf8.ValidString(input) {
		return graphPreviewReference{}, fmt.Errorf("reference must be nonempty UTF-8")
	}
	if input[0] == '[' {
		depth, end := 0, -1
		for i, r := range input {
			switch r {
			case '[':
				depth++
			case ']':
				depth--
				if depth < 0 {
					return graphPreviewReference{}, fmt.Errorf("unbalanced URI reference brackets")
				}
				if depth == 0 {
					end = i
					break
				}
			}
			if end >= 0 {
				break
			}
		}
		if end < 0 {
			return graphPreviewReference{}, fmt.Errorf("unbalanced URI reference brackets")
		}
		identity := input[1:end]
		parsed, err := url.Parse(identity)
		if err != nil || parsed.Scheme == "" || graphPreviewURIHasRawPathBrackets(identity) {
			return graphPreviewReference{}, fmt.Errorf("bracketed reference requires an absolute URI")
		}
		suffix := input[end+1:]
		if !strings.HasPrefix(suffix, "@") || len(suffix) == 1 {
			return graphPreviewReference{}, fmt.Errorf("bracketed URI requires @revision")
		}
		if err := graphPreviewReferenceRevision(suffix[1:]); err != nil {
			return graphPreviewReference{}, err
		}
		return graphPreviewReference{Identity: identity, Revision: suffix[1:], Explicit: true, Wrapped: true}, nil
	}
	parsed, err := url.Parse(input)
	if err != nil && strings.Contains(input, "://") {
		return graphPreviewReference{}, fmt.Errorf("invalid absolute URI reference")
	}
	if err == nil && parsed.Scheme != "" {
		return graphPreviewReference{Identity: input}, nil
	}
	if i := strings.IndexByte(input, '@'); i >= 0 {
		if i == 0 || i == len(input)-1 {
			return graphPreviewReference{}, fmt.Errorf("local reference requires ID@revision")
		}
		revision := input[i+1:]
		if err := graphPreviewReferenceRevision(revision); err != nil {
			return graphPreviewReference{}, err
		}
		return graphPreviewReference{Identity: input[:i], Revision: revision, Explicit: true}, nil
	}
	return graphPreviewReference{Identity: input}, nil
}

func graphPreviewReferenceRevision(revision string) error {
	if revision == "" || !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return fmt.Errorf("revision must be nonempty UTF-8 of at most %d bytes", graphstore.PreviewVersionTokenLimit)
	}
	return nil
}

// The flag and suffix are equivalent selectors; the compare-and-swap guards
// --if-revision and --if-source-revision are deliberately not consulted here.
func graphPreviewReferenceWithFlag(input, revision string, flagSet bool) (graphPreviewReference, error) {
	ref, err := graphPreviewParseReference(input)
	if err != nil {
		return ref, err
	}
	if flagSet {
		if ref.Explicit {
			return ref, fmt.Errorf("choose a reference @revision suffix or its revision flag, not both")
		}
		if err := graphPreviewReferenceRevision(revision); err != nil {
			return ref, err
		}
		ref.Revision, ref.Explicit = revision, true
	}
	return ref, nil
}

func graphPreviewWritableIdentity(input string) (string, error) {
	ref, err := graphPreviewParseReference(input)
	if err != nil {
		return "", err
	}
	if ref.Explicit {
		return "", fmt.Errorf("write targets use a bare identity; use --if-revision for a revision guard")
	}
	return ref.Identity, nil
}

// URI square brackets belong to an authority's host literal, not its path,
// query or fragment. Percent-encoded brackets remain ordinary identity bytes.
func graphPreviewURIHasRawPathBrackets(identity string) bool {
	// The caller already parsed an absolute URI. Remove the original scheme
	// by its delimiter: net/url normalizes scheme case, identity must not.
	_, remainder, _ := strings.Cut(identity, ":")
	if strings.HasPrefix(remainder, "//") {
		end := strings.IndexAny(remainder[2:], "/?#")
		if end < 0 {
			return false
		}
		remainder = remainder[2+end:]
	}
	return strings.ContainsAny(remainder, "[]")
}
