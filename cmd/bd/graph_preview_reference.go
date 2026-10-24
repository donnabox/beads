package main

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/clirevision"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// graphPreviewReference is a CLI selector, not a stored BDP Reference. In
// particular, parsing a pin must not make an unpinned graphstore Link look
// pinned: persistence and historical validation are separate writer work.
type graphPreviewReference struct {
	Identity        string
	Revision        string
	Explicit        bool
	Wrapped         bool
	Ordinal         int64
	LiteralRevision bool
}

// graphPreviewParseReference uses the existing CLI URL boundary: unwrapped
// locator forms containing :// stay literal, including @ and IPv6 brackets.
// Colon-only spellings remain local shorthand, even when they resemble opaque
// URIs. A canonical Scope URL with a revision uses [URL]@N or [URL]@=TOKEN.
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
					return graphPreviewReference{}, fmt.Errorf("unbalanced URL reference brackets")
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
			return graphPreviewReference{}, fmt.Errorf("unbalanced URL reference brackets")
		}
		identity := input[1:end]
		parsed, err := url.Parse(identity)
		if err != nil || !strings.Contains(identity, "://") || parsed.Host == "" || parsed.Opaque != "" ||
			(!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) ||
			graphPreviewURIHasRawPathBrackets(identity) {
			return graphPreviewReference{}, fmt.Errorf("bracketed reference requires a canonical HTTP(S) Scope URL; local IDs use ID@revision")
		}
		suffix := input[end+1:]
		if !strings.HasPrefix(suffix, "@") || len(suffix) == 1 {
			return graphPreviewReference{}, fmt.Errorf("bracketed URL requires @revision")
		}
		return graphPreviewReferenceSuffix(graphPreviewReference{Identity: identity, Explicit: true, Wrapped: true}, suffix[1:])
	}
	if strings.Contains(input, "://") {
		parsed, err := url.Parse(input)
		if err != nil || parsed.Scheme == "" {
			return graphPreviewReference{}, fmt.Errorf("invalid absolute URL reference")
		}
		return graphPreviewReference{Identity: input}, nil
	}
	if i := strings.IndexByte(input, '@'); i >= 0 {
		if i == 0 || i == len(input)-1 {
			return graphPreviewReference{}, fmt.Errorf("local reference requires ID@revision; address an ID containing @ by its full canonical Scope URL")
		}
		revision := input[i+1:]
		return graphPreviewReferenceSuffix(graphPreviewReference{Identity: input[:i], Explicit: true}, revision)
	}
	return graphPreviewReference{Identity: input}, nil
}

func graphPreviewReferenceSuffix(ref graphPreviewReference, suffix string) (graphPreviewReference, error) {
	selected, err := clirevision.Parse(suffix, graphstore.PreviewVersionTokenLimit)
	if err != nil {
		return graphPreviewReference{}, fmt.Errorf("%w; address an ID containing @ by its full canonical Scope URL", err)
	}
	ref.Ordinal, ref.Revision, ref.LiteralRevision = selected.Ordinal, selected.Token, selected.Token != ""
	if selected.Current {
		ref.Revision = "current"
	}
	return ref, nil
}

func (r graphPreviewReference) currentSelection() bool {
	return !r.LiteralRevision && r.Ordinal == 0 && r.Revision == "current"
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
	return graphPreviewResolveParsedReference(ref)
}

func graphPreviewWritableIdentity(input string) (string, error) {
	ref, err := graphPreviewParseReference(input)
	if err != nil {
		return "", err
	}
	if ref.Explicit {
		return "", fmt.Errorf("write targets use a bare identity; use --if-revision for a revision guard; address an ID containing @ by its full canonical Scope URL")
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
