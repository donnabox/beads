package main

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
)

func canonicalReadIntentForTest(t *testing.T, selector string) canonicalReadReferenceIntent {
	t.Helper()
	intent, err := parseCanonicalReadReferenceIntent(selector)
	if err != nil {
		t.Fatalf("selector %q: %v", selector, err)
	}
	return intent
}

func TestCanonicalReadReferenceAbsoluteNeedsNoDefaultOrLocalLookup(t *testing.T) {
	for _, tc := range []struct{ name, input, uri, revision string }{
		{"literal-at", "https://example.test/team/notes@4", "https://example.test/team/notes@4", ""},
		{"absolute-current", "[https://example.test/item]@current", "https://example.test/item", ""},
		{"absolute-exact", "[https://example.test/item]@=+ /@=", "https://example.test/item", "+ /@="},
		{"exact-current-token", "[https://example.test/item]@=current", "https://example.test/item", "current"},
		{"exact-numeric-token", "[https://example.test/item]@=4", "https://example.test/item", "4"},
		{"ipv6-wrapper", "[http://[::1]/item]@=r1", "http://[::1]/item", "r1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			intent := canonicalReadIntentForTest(t, tc.input)
			unexpected := canonicalReadReferenceResolvers{
				locateLocal: func(context.Context, string) (string, error) {
					t.Fatal("absolute read invoked local locator")
					return "", nil
				},
				resolveOrdinal: func(context.Context, string, int64) (string, error) {
					t.Fatal("current/exact read invoked ordinal lookup")
					return "", nil
				},
			}
			got, err := resolveCanonicalReadReferenceIntent(context.Background(), intent, unexpected)
			want := bdpwire.Reference{URI: tc.uri, Revision: tc.revision}
			if err != nil || got != want {
				t.Fatalf("absolute selection changed: %+v %v, want%+v", got, err, want)
			}
			// A second resolution without ANY authority callbacks demonstrates
			// that a default Scope/local-store capability was never required.
			got.URI, got.Revision = "changed", "changed"
			again, err := resolveCanonicalReadReferenceIntent(context.Background(), intent, canonicalReadReferenceResolvers{})
			if err != nil || again != want {
				t.Fatalf("returned carrier aliases intent: %+v %v", again, err)
			}
		})
	}
}

func TestCanonicalReadReferenceLocalLocatorAndStableOrdinalAreDistinct(t *testing.T) {
	const uri = "https://selected.example/opaque-resource"
	for _, tc := range []struct {
		input, localID, revision string
		ordinal                  int64
	}{
		{"task", "task", "", 0}, {"task@current", "task", "", 0},
		{"task@=current", "task", "current", 0}, {"task@=4", "task", "4", 0},
		{"task@4", "task", "mapped-full-token", 4},
		{"types/ordinary-type-bead@=saved", "types/ordinary-type-bead", "saved", 0},
		{"plan:v1@=saved", "plan:v1", "saved", 0},
	} {
		t.Run(tc.input, func(t *testing.T) {
			localCalls, ordinalCalls := 0, 0
			got, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, tc.input), canonicalReadReferenceResolvers{
				locateLocal: func(ctx context.Context, id string) (string, error) {
					localCalls++
					if id != tc.localID {
						t.Fatalf("whole local ID rewritten: %q want%q", id, tc.localID)
					}
					return uri, nil // selected authority supplies the whole locator
				},
				resolveOrdinal: func(ctx context.Context, resource string, ordinal int64) (string, error) {
					ordinalCalls++
					if resource != uri || ordinal != tc.ordinal || ordinal == 0 {
						t.Fatalf("ordinal lookup input=%q %d", resource, ordinal)
					}
					return "mapped-full-token", nil
				},
			})
			wantOrdinalCalls := 0
			if tc.ordinal != 0 {
				wantOrdinalCalls = 1
			}
			if err != nil || got != (bdpwire.Reference{URI: uri, Revision: tc.revision}) || localCalls != 1 || ordinalCalls != wantOrdinalCalls {
				t.Fatalf("authority paths conflated: %+v %v local%d ordinal%d", got, err, localCalls, ordinalCalls)
			}
		})
	}
}

func TestCanonicalReadReferenceAbsoluteOrdinalCannotRetargetOrGuess(t *testing.T) {
	const uri = "https://different.example/not-the-default/task"
	intent := canonicalReadIntentForTest(t, "["+uri+"]@4")
	calls := 0
	got, err := resolveCanonicalReadReferenceIntent(context.Background(), intent, canonicalReadReferenceResolvers{
		locateLocal: func(context.Context, string) (string, error) {
			t.Fatal("absolute ordinal invoked local identity lookup")
			return "", nil
		},
		resolveOrdinal: func(ctx context.Context, resource string, ordinal int64) (string, error) {
			calls++
			if resource != uri || ordinal != 4 {
				t.Fatalf("ordinal authority changed: %q %d", resource, ordinal)
			}
			return "https://token.example/not-a-locator", nil
		},
	})
	if err != nil || calls != 1 || got.URI != uri || got.Revision != "https://token.example/not-a-locator" {
		t.Fatalf("ordinal retargeted identity or replaced token: %+v %v calls%d", got, err, calls)
	}
	missing, err := resolveCanonicalReadReferenceIntent(context.Background(), intent, canonicalReadReferenceResolvers{})
	if missing != (bdpwire.Reference{}) || err == nil || !strings.Contains(err.Error(), "stable ordinal mapping is unavailable") {
		t.Fatalf("unavailable mapping guessed current/nearest: %+v %v", missing, err)
	}
	missing, err = resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, "task"), canonicalReadReferenceResolvers{})
	if missing != (bdpwire.Reference{}) || err == nil || !strings.Contains(err.Error(), "local Resource locator is unavailable") {
		t.Fatalf("relative identity guessed an ambient Scope: %+v %v", missing, err)
	}
}

func TestCanonicalReadReferenceParseAndCallbackFactsFailClosed(t *testing.T) {
	for _, input := range []string{"", "task@0", "task@-1", "task@9223372036854775808", "task@", "task@=", "task@opaque", "a//b", "https://", "ftp://example.test/item", "https://user:secret@example.test/item", "https://example.test/item?revision=x", "https://example.test/item#fragment", "[https://example.test/item]", "[item]@=saved", string([]byte{0xff})} {
		if intent, err := parseCanonicalReadReferenceIntent(input); err == nil || intent.ready {
			t.Fatalf("invalid selector admitted: %q %+v %v", input, intent, err)
		}
	}
	for _, uri := range []string{"", "relative", "ftp://example.test/item", "https://user:secret@example.test/item", "https://example.test/item?revision=x", "https://example.test/item#fragment"} {
		ordinalCalled := false
		got, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, "task@4"), canonicalReadReferenceResolvers{
			locateLocal:    func(context.Context, string) (string, error) { return uri, nil },
			resolveOrdinal: func(context.Context, string, int64) (string, error) { ordinalCalled = true; return "must-not-run", nil },
		})
		if err == nil || got != (bdpwire.Reference{}) || ordinalCalled {
			t.Fatalf("invalid locator used for ordinal lookup: %q %+v %v", uri, got, err)
		}
	}
	for _, token := range []string{"", string([]byte{0xff}), strings.Repeat("x", 4097)} {
		got, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, "[https://example.test/item]@4"), canonicalReadReferenceResolvers{
			resolveOrdinal: func(context.Context, string, int64) (string, error) { return token, nil },
		})
		if err == nil || got != (bdpwire.Reference{}) {
			t.Fatalf("invalid mapped token admitted: length%d %+v %v", len(token), got, err)
		}
	}
}

func TestCanonicalReadReferenceCarriesCallerContextAndPreservesErrors(t *testing.T) {
	type contextKey struct{}
	ctx, cancel := context.WithDeadline(context.WithValue(context.Background(), contextKey{}, "selected-context"), time.Now().Add(time.Hour))
	defer cancel()
	deadline, _ := ctx.Deadline()
	sequence := []string{}
	checkContext := func(actual context.Context) {
		if actual != ctx || actual.Value(contextKey{}) != "selected-context" {
			t.Fatal("callback context replaced")
		}
		got, ok := actual.Deadline()
		if !ok || !got.Equal(deadline) {
			t.Fatal("callback deadline replaced")
		}
	}
	_, err := resolveCanonicalReadReferenceIntent(ctx, canonicalReadIntentForTest(t, "task@4"), canonicalReadReferenceResolvers{
		locateLocal: func(actual context.Context, id string) (string, error) {
			checkContext(actual)
			sequence = append(sequence, "locator")
			return "https://example.test/task", nil
		},
		resolveOrdinal: func(actual context.Context, uri string, ordinal int64) (string, error) {
			checkContext(actual)
			sequence = append(sequence, "ordinal")
			return "saved", nil
		},
	})
	if err != nil || !reflect.DeepEqual(sequence, []string{"locator", "ordinal"}) {
		t.Fatalf("context/call order: %v %v", sequence, err)
	}
	fault := errors.New("selected authority unavailable")
	for _, fromOrdinal := range []bool{false, true} {
		got, err := resolveCanonicalReadReferenceIntent(ctx, canonicalReadIntentForTest(t, "task@4"), canonicalReadReferenceResolvers{
			locateLocal: func(context.Context, string) (string, error) {
				if !fromOrdinal {
					return "", fault
				}
				return "https://example.test/task", nil
			},
			resolveOrdinal: func(context.Context, string, int64) (string, error) {
				if !fromOrdinal {
					t.Fatal("ordinal called after locator failure")
				}
				return "", fault
			},
		})
		if err != fault || got != (bdpwire.Reference{}) {
			t.Fatalf("callback error replaced or partial ref returned: %+v %v", got, err)
		}
	}
}

func TestCanonicalReadReferenceCancellationAndZeroIntentBeforeCallbacks(t *testing.T) {
	calls := 0
	resolvers := canonicalReadReferenceResolvers{
		locateLocal:    func(context.Context, string) (string, error) { calls++; return "https://example.test/task", nil },
		resolveOrdinal: func(context.Context, string, int64) (string, error) { calls++; return "saved", nil },
	}
	intent := canonicalReadIntentForTest(t, "task@4")
	if got, err := resolveCanonicalReadReferenceIntent(nil, intent, resolvers); err == nil || got != (bdpwire.Reference{}) || calls != 0 {
		t.Fatalf("nil context acquired: %+v %v calls%d", got, err, calls)
	}
	if got, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadReferenceIntent{}, resolvers); err == nil || got != (bdpwire.Reference{}) || calls != 0 {
		t.Fatalf("zero intent acquired: %+v %v calls%d", got, err, calls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := resolveCanonicalReadReferenceIntent(ctx, intent, resolvers); !errors.Is(err, context.Canceled) || got != (bdpwire.Reference{}) || calls != 0 {
		t.Fatalf("canceled input acquired: %+v %v calls%d", got, err, calls)
	}
	expired, done := context.WithDeadline(context.Background(), time.Unix(1, 0))
	defer done()
	if got, err := resolveCanonicalReadReferenceIntent(expired, intent, resolvers); !errors.Is(err, context.DeadlineExceeded) || got != (bdpwire.Reference{}) || calls != 0 {
		t.Fatalf("expired input acquired: %+v %v calls%d", got, err, calls)
	}
	for _, duringOrdinal := range []bool{false, true} {
		active, stop := context.WithCancel(context.Background())
		ordinalCalls := 0
		got, err := resolveCanonicalReadReferenceIntent(active, intent, canonicalReadReferenceResolvers{
			locateLocal: func(context.Context, string) (string, error) {
				if !duringOrdinal {
					stop()
				}
				return "https://example.test/task", nil
			},
			resolveOrdinal: func(context.Context, string, int64) (string, error) { ordinalCalls++; stop(); return "saved", nil },
		})
		stop()
		if !errors.Is(err, context.Canceled) || got != (bdpwire.Reference{}) || (!duringOrdinal && ordinalCalls != 0) || (duringOrdinal && ordinalCalls != 1) {
			t.Fatalf("post-callback cancellation leaked/fell through: %+v %v ordinal%d", got, err, ordinalCalls)
		}
	}
}

func TestCanonicalReadReferenceWireValueContainsOnlyExactURIToken(t *testing.T) {
	intent := canonicalReadIntentForTest(t, "[https://example.test/rootless]@=+ /@=")
	got, err := resolveCanonicalReadReferenceIntent(context.Background(), intent, canonicalReadReferenceResolvers{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"uri":"https://example.test/rootless","revision":"+ /@="}`; string(raw) != want {
		t.Fatalf("token encoded/reinterpreted or intent leaked onto wire: %s want%s", raw, want)
	}
	current, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, "[https://example.test/rootless]@current"), canonicalReadReferenceResolvers{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(current)
	if err != nil || string(raw) != `"https://example.test/rootless"` {
		t.Fatalf("current gained revision/ordinal: %s %v", raw, err)
	}
}

func TestCanonicalReadReferenceLocatorPolicyPreservesOpaqueURLTokens(t *testing.T) {
	const selectedURI = "https://selected.example/task"
	for _, token := range []string{"https://token.example/rev", "arbitrary://token/rev", "arbitrary://token/rev@4@=next"} {
		input := "task@=" + token
		t.Run(input, func(t *testing.T) {
			// The existing wrapper deliberately retains the old Contains(://)
			// boundary, including its refusal of these URL-shaped local tokens.
			compat, compatErr := graphPreviewParseReference(input)
			explicitCompat, explicitErr := graphPreviewParseReferenceWithLocatorPolicy(input, graphPreviewReferenceCompatLocator)
			if compatErr == nil || explicitErr == nil || compat != explicitCompat || compatErr.Error() != explicitErr.Error() {
				t.Fatalf("compat parser changed: %+v %v / %+v %v", compat, compatErr, explicitCompat, explicitErr)
			}
			localCalls := 0
			got, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, input), canonicalReadReferenceResolvers{
				locateLocal: func(_ context.Context, stringID string) (string, error) {
					localCalls++
					if stringID != "task" {
						t.Fatalf("token retargeted local ID: %q", stringID)
					}
					return selectedURI, nil
				},
				resolveOrdinal: func(context.Context, string, int64) (string, error) {
					t.Fatal("exact token invoked ordinal lookup")
					return "", nil
				},
			})
			if err != nil || got != (bdpwire.Reference{URI: selectedURI, Revision: token}) || localCalls != 1 {
				t.Fatalf("opaque token changed: %+v %v local calls %d", got, err, localCalls)
			}
		})
	}
	// A bracketed locator can carry the same URL-shaped token explicitly.
	bracketed, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, "[https://example.test/item]@=https://token.example/rev@4"), canonicalReadReferenceResolvers{})
	if err != nil || bracketed != (bdpwire.Reference{URI: "https://example.test/item", Revision: "https://token.example/rev@4"}) {
		t.Fatalf("bracketed opaque token changed: %+v %v", bracketed, err)
	}
	for _, input := range []string{"https://example.test/item@=saved", "https://example.test/item@=arbitrary://token/rev@4"} {
		got, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, input), canonicalReadReferenceResolvers{
			locateLocal: func(context.Context, string) (string, error) {
				t.Fatal("absolute URL invoked local locator")
				return "", nil
			},
			resolveOrdinal: func(context.Context, string, int64) (string, error) {
				t.Fatal("absolute URL invoked ordinal lookup")
				return "", nil
			},
		})
		if err != nil || got != (bdpwire.Reference{URI: input}) {
			t.Fatalf("literal absolute @= was split: %+v %v", got, err)
		}
	}
}

func TestCanonicalReadReferenceLocatorPolicyUsesExistingSchemeGrammar(t *testing.T) {
	for _, input := range []string{"https://example.test/item@=saved", "HTTPS://example.test/item@=saved", "http://[::1]/item@=saved", "ftp://example.test/item@=saved", "a+B.c-d://example.test/item@=saved"} {
		compat, compatErr := graphPreviewParseReference(input)
		canonical, canonicalErr := graphPreviewParseReferenceWithLocatorPolicy(input, graphPreviewReferenceCanonicalReadLocator)
		if compatErr != nil || canonicalErr != nil || canonical != compat || canonical.Identity != input || canonical.Explicit {
			t.Fatalf("leading scheme no longer literal: %q %+v %v / %+v %v", input, compat, compatErr, canonical, canonicalErr)
		}
	}
	for _, input := range []string{"https://%zz/item@=saved", "https://[broken/item@=saved"} {
		if ref, err := graphPreviewParseReferenceWithLocatorPolicy(input, graphPreviewReferenceCanonicalReadLocator); err == nil || ref != (graphPreviewReference{}) {
			t.Fatalf("malformed leading locator fell through: %q %+v %v", input, ref, err)
		}
	}
	for _, input := range []string{"ftp://example.test/item@=saved", "a+B.c-d://example.test/item@=saved", "https://%zz/item@=saved", "https://[broken/item@=saved"} {
		if intent, err := parseCanonicalReadReferenceIntent(input); err == nil || intent.ready {
			t.Fatalf("lexical classification bypassed canonical locator admission: %q %+v %v", input, intent, err)
		}
	}
	if ref, err := graphPreviewParseReferenceWithLocatorPolicy("task", graphPreviewReferenceLocatorPolicy(255)); err == nil || ref != (graphPreviewReference{}) {
		t.Fatalf("unknown locator policy admitted: %+v %v", ref, err)
	}
}

func TestCanonicalReadReferenceColonLocalIDKeepsURLShapedToken(t *testing.T) {
	const selectedURI = "https://selected.example/whole-local-id"
	for _, tc := range []struct{ id, token string }{
		{"plan:v1", "https://token.example/rev"},
		{"plan:v1", "arbitrary://token/rev@4@=next"},
		{"namespace:part:task", "https://token.example/rev@=next@4"},
	} {
		input := tc.id + "@=" + tc.token
		t.Run(input, func(t *testing.T) {
			// Compat still treats this entire spelling as the opaque plan:
			// or namespace: URL from its old Contains boundary.
			compat, compatErr := graphPreviewParseReference(input)
			explicitCompat, explicitErr := graphPreviewParseReferenceWithLocatorPolicy(input, graphPreviewReferenceCompatLocator)
			if compatErr != nil || explicitErr != nil || compat != explicitCompat || compat.Identity != input || compat.Explicit {
				t.Fatalf("compat classification changed: %+v %v / %+v %v", compat, compatErr, explicitCompat, explicitErr)
			}
			calls := 0
			got, err := resolveCanonicalReadReferenceIntent(context.Background(), canonicalReadIntentForTest(t, input), canonicalReadReferenceResolvers{
				locateLocal: func(_ context.Context, id string) (string, error) {
					calls++
					if id != tc.id {
						t.Fatalf("colon ID/token was split or retargeted: %q want %q", id, tc.id)
					}
					return selectedURI, nil
				},
				resolveOrdinal: func(context.Context, string, int64) (string, error) {
					t.Fatal("exact URL-shaped token invoked ordinal mapping")
					return "", nil
				},
			})
			if err != nil || got != (bdpwire.Reference{URI: selectedURI, Revision: tc.token}) || calls != 1 {
				t.Fatalf("colon local ID lost opaque token: %+v %v calls %d", got, err, calls)
			}
		})
	}
}
