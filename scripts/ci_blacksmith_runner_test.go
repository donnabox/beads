package scripts_test

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// This file is shared between F7a (ci/f7a-blacksmith-fold) and F7c
// (ci/f7c-advisory-workflows). F7a lands first; F7c should rebase onto it and
// delete its own copies of the pieces below rather than keep a second
// definition (two independent `sameRepoBlacksmith4vcpu` consts merge fine
// textually since the literal is identical, but two
// `TestSameRepoBlacksmithExpressionSemantics` functions in the same package
// do not - that was review SF-5/F7c-collision on the F7a review).
//
// Public surface F7c (or any later slice) should reuse:
//   - sameRepoBlacksmith2vcpu / sameRepoBlacksmith4vcpu / sameRepoBlacksmith8vcpu
//   - evalGHExpr(expr string, ctx map[string]string) (any, error)
//   - mustEvalGHRunsOn(t *testing.T, expr string, ctx map[string]string) string
//
// F7c's own TestSameRepoBlacksmithExpressionSemantics (ci_f7c_advisory_test.go)
// and TestBlacksmithAdvisoryJobsReadNoSecrets should be folded into (or
// replaced by calls into) this file's test and
// TestBlacksmithJobsReadNoSecrets (ci_workflow_test.go) respectively, rather
// than kept as independent copies.

// F3: the same "same-repo PR, or merge_group" Blacksmith expression used by
// pr.yml's and pr-risk.yml's bazel-coverage/ci-gate/detect-ci-tier jobs - a
// package-level const so every test that needs it (TestSameRepoBlacksmithRunners,
// TestPRRiskBazelCoverageJob, ...) reads the one literal.
const sameRepoBlacksmith2vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-2vcpu-ubuntu-2404' || 'ubuntu-latest' }}"

// F7a: the same same-repo expression at 4 vCPU and 8 vCPU, for jobs sized
// larger than the 2 vCPU default (check-doc-flags, pr-policy-wrapper,
// pr-risk.yml's test-nix at 4 vCPU; check-release-target-cross-compilation at
// 8 vCPU). F7c's advisory workflows also use the 4 vCPU size.
const sameRepoBlacksmith4vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-4vcpu-ubuntu-2404' || 'ubuntu-latest' }}"
const sameRepoBlacksmith8vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-8vcpu-ubuntu-2404' || 'ubuntu-latest' }}"

// Only native lint needs the declared Linux LLVM host runtime.
const sameRepoBlacksmithNativeLint4vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-4vcpu-ubuntu-2404' || (matrix.target == 'native' && 'ubuntu-24.04' || 'ubuntu-latest') }}"

// F7b: same-repo Blacksmith Windows (blacksmith-*vcpu-windows-2025, public
// beta) and macOS (blacksmith-*vcpu-macos-26) labels, same ternary shape
// as the Linux consts above - forks/Dependabot keep the current GitHub-hosted
// windows-latest/macos-latest label. windows-2025 is GitHub's windows-2025
// image minus the full Visual Studio IDE, EdgeDriver and WinAppDriver; VS
// Build Tools 2022 is present (cgo via MSVC/clang-cl or an installed
// mingw-w64 toolchain), but there is no Linux Docker. Jobs moved onto these
// labels that rely on a specific Windows toolchain detail already document
// that dependency at the call site.
const sameRepoBlacksmithWindows2vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-2vcpu-windows-2025' || 'windows-latest' }}"
const sameRepoBlacksmithWindows4vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-4vcpu-windows-2025' || 'windows-latest' }}"
const sameRepoBlacksmithWindows8vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-8vcpu-windows-2025' || 'windows-latest' }}"
const sameRepoBlacksmithWindows16vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-16vcpu-windows-2025' || 'windows-latest' }}"

// bazel.yml keeps forks on GitHub-hosted runners. Upstream lanes pick
// their runner from the rbe job's decision
// (TestBazelRBEJobDecidesOnce): Blacksmith in modes remote and cache,
// GitHub-hosted otherwise. Mode cache reads rbe-cache (rbe-west's anonymous
// read-only cache), which refuses GitHub-hosted runners, so its lanes (an
// rbe=cache dispatch's, the only way into mode cache) run on Blacksmith.
// bazelRemoteRunsOn builds that ternary for one Blacksmith size so every
// lane's pinned runs-on is the same literal apart from the label.
func bazelRemoteRunsOn(label string) string {
	return "${{ github.repository_owner == 'gastownhall' && (needs.rbe.outputs.mode == 'remote' || needs.rbe.outputs.mode == 'cache') && '" + label + "' || 'ubuntu-latest' }}"
}

// bazelRRCCacheReaderRunsOn is bazelRemoteRunsOn for the lanes that read the
// remote repo contents cache in mode cache (isBazelRRCCacheReader): their
// GitHub-hosted fallback is ubuntu-24.04, not ubuntu-latest, because the
// cached trees come from a 24.04 host (design R2) and the image must not
// move under them. (Mode cache itself now runs on Blacksmith's
// ubuntu-2404, the family rrc-seed writes from.)
func bazelRRCCacheReaderRunsOn(label string) string {
	return "${{ github.repository_owner == 'gastownhall' && (needs.rbe.outputs.mode == 'remote' || needs.rbe.outputs.mode == 'cache') && '" + label + "' || 'ubuntu-24.04' }}"
}

// bazelBlacksmithModes: the modes in which bazelRemoteRunsOn picks the
// Blacksmith label.
var bazelBlacksmithModes = map[string]bool{"remote": true, "cache": true}

// isBazelRRCCacheReader: a remote repo contents cache reader
// (bazelRRCReadLanes) that also runs in mode cache (not remote-only).
// bazel-dolt-race and bazel-dolt-int run only in mode remote (their own ifs,
// TestBazelDoltRaceArmsExactlyOnePerMode, TestBazelDoltIntArmsExactlyOnePerMode),
// so they keep the lanes' ubuntu-latest fallback like the remote-only jobs:
// the fallback branch of their runs-on ternary is never actually taken.
func isBazelRRCCacheReader(name string) bool {
	return slices.Contains(bazelRRCReadLanes, name) && !bazelRemoteOnlyJobs[name] && name != bazelDoltRaceJobName && name != bazelDoltIntJobName
}

// The default lane size: every action executes on rbe-west, but the
// client's loading and analysis is CPU-bound and gates every action
// (ga-vnycm2.8: 17-38 s on 2 vCPU, 12-26 s on 4), so 4 vCPU.
var bazelLaneRunsOn = bazelRemoteRunsOn("blacksmith-4vcpu-ubuntu-2404")

// rbe-prewarm only dispatches a pool worker: no Bazel client, 2 vCPU.
var bazelPrewarmRunsOn = bazelRemoteRunsOn("blacksmith-2vcpu-ubuntu-2404")

// bazel-test's `bazel test //... --config=ci` spent ~63 of its 81 s in
// client-side loading and analysis (1798 packages, 66k configured targets;
// critical path 7.65 s, every action a remote cache hit), which Skyframe
// parallelizes across the client's cores: 8 vCPU (ga-vnycm2.8 analysis-only:
// 25.6 s on 4 vCPU, 16.3 s on 8).
var bazelTestLaneRunsOn = bazelRemoteRunsOn("blacksmith-8vcpu-ubuntu-2404")

// The package gates: package-npm at 4 vCPU (F3); package-mcp at 8 vCPU with
// pytest -n 16 (bazelMCPPytestWorkers), since its pytest run is dominated by
// per-test `bd init` fixture setup, which is CPU-bound and parallel.
var bazelPackageRunsOn = map[string]string{
	"package-mcp": bazelRemoteRunsOn("blacksmith-8vcpu-ubuntu-2404"),
	"package-npm": bazelRemoteRunsOn("blacksmith-4vcpu-ubuntu-2404"),
}

// package-mcp's BEADS_MCP_PYTEST_WORKERS: 16 xdist workers on the 8 vCPU
// Blacksmith runner (modes remote and cache), package-mcp.sh's default 8 on
// the GitHub-hosted fallback (forks, Dependabot, rbe off/skip).
const bazelMCPPytestWorkers = "${{ github.repository_owner == 'gastownhall' && (needs.rbe.outputs.mode == 'remote' || needs.rbe.outputs.mode == 'cache') && '16' || '8' }}"

// Same-repo Blacksmith macOS (Apple Silicon M4, ARM64). Pinned to macos-26
// rather than blacksmith-*-macos-latest because GitHub's macos-latest (the
// fork/Dependabot fallback) resolves to macos-26-arm64 today; pinning keeps
// both paths on the same OS/arch and keeps the PR legs and main.yml's
// blacksmith-macos-go-build-cache saver on one image regardless of when
// Blacksmith moves its own -latest alias. Bump these (and
// blacksmithMacOSLabel) when GitHub moves macos-latest.
const blacksmithMacOSLabel = "blacksmith-6vcpu-macos-26"
const sameRepoBlacksmithMacOS6vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-6vcpu-macos-26' || 'macos-latest' }}"
const sameRepoBlacksmithMacOS12vcpu = "${{ github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-12vcpu-macos-26' || 'macos-latest' }}"

// sameRepoPlatformsMatrixMarkerRunsOn is the "matrix marker" form (spec
// F7/F7b §2.1) used by mixed-OS matrix jobs (pr-preflight-platforms,
// check-doc-freshness-platforms): Actions does not expand `${{ }}` inside a
// matrix array/include value, so each OS leg's `include` entry instead sets a
// plain string marker field (`runner: same-repo-linux` etc.), and a single
// chained ternary in the job's `runs-on` tests which marker (if any) the
// current leg carries, falling back to `matrix.os` for legs with no marker or
// when the marker's own same-repo condition is false.
//
// The macOS branch resolves like the other two: blacksmith-6vcpu-macos-26
// (sameRepoBlacksmithMacOS6vcpu's label) for a trusted same-repo PR or
// merge_group, GitHub-hosted macos-latest otherwise. F7b review fix S3 kept
// it a plain 'macos-latest' (cost, and no push-to-main Blacksmith-macOS
// saver); the cost is now accepted and main.yml's
// blacksmith-macos-go-build-cache is that saver.
const sameRepoPlatformsMatrixMarkerRunsOn = "${{ matrix.runner == 'same-repo-linux' && (github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-4vcpu-ubuntu-2404' || 'ubuntu-latest') || matrix.runner == 'same-repo-windows' && (github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-4vcpu-windows-2025' || 'windows-latest') || matrix.runner == 'same-repo-macos' && (github.repository_owner == 'gastownhall' && (github.event_name == 'merge_group' || (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository && github.actor != 'dependabot[bot]')) && 'blacksmith-6vcpu-macos-26' || 'macos-latest') || matrix.os }}"

// --- a minimal GitHub Actions expression evaluator -------------------------
//
// Review SF-3 (2026-10-03) on the F7a review: the original
// TestSameRepoBlacksmithExpressionSemantics ran a hand-written Go
// re-implementation of the same-repo Blacksmith formula against its own
// truth table, then string-compared the pinned consts against a template
// built from the same pieces - a tautology that a change to both the
// template and the const (in the same wrong way) would still pass. This
// evaluator actually parses and evaluates the real `${{ ... }}` expression
// string, for the small subset of the GitHub Actions expression language
// this repo's same-repo Blacksmith ternaries use: string literals, dotted
// identifier lookups (resolved from a caller-supplied context map), ==, !=,
// &&, ||, !, parentheses, and the true/false/null literals. Identifiers may
// contain '-' (inputs.fork-farm), and == / != follow GitHub's loose equality
// (ghEquals: strings case-insensitively, mixed types as numbers). It is not a general-purpose GHA
// expression engine: only the argument-free always() status function; no
// other functions, numbers or object/array literals.
//
// Merge queue (ci_merge_queue_test.go): callers simulate a merge_group run
// with github.event_name "merge_group", github.actor
// "github-merge-queue[bot]", github.ref "refs/heads/gh-readonly-queue/...",
// the github.event.merge_group.* fields, and no github.event.pull_request.*
// key at all (null on that event).
//
// && and || use GitHub's own short-circuit-returns-operand semantics (not
// strict booleans: `false && 'x'` is `false`, not `false`'s boolean negation
// of something), which is exactly what lets a `cond && 'labelA' || 'labelB'`
// ternary evaluate to a string label rather than a bool.

type ghToken struct {
	kind string // "ident", "string", "op", "lparen", "rparen", "bang"
	val  string
}

func ghTokenize(s string) ([]ghToken, error) {
	var toks []ghToken
	i := 0
	isIdentByte := func(b byte) bool {
		return b == '.' || b == '_' || b == '-' ||
			(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
	}
	for i < len(s) {
		c := s[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '(':
			toks = append(toks, ghToken{"lparen", "("})
			i++
		case c == ')':
			toks = append(toks, ghToken{"rparen", ")"})
			i++
		case c == '\'':
			j := i + 1
			for j < len(s) && s[j] != '\'' {
				j++
			}
			if j >= len(s) {
				return nil, fmt.Errorf("unterminated string literal at byte %d in %q", i, s)
			}
			toks = append(toks, ghToken{"string", s[i+1 : j]})
			i = j + 1
		case strings.HasPrefix(s[i:], "=="):
			toks = append(toks, ghToken{"op", "=="})
			i += 2
		case strings.HasPrefix(s[i:], "!="):
			toks = append(toks, ghToken{"op", "!="})
			i += 2
		case strings.HasPrefix(s[i:], "&&"):
			toks = append(toks, ghToken{"op", "&&"})
			i += 2
		case strings.HasPrefix(s[i:], "||"):
			toks = append(toks, ghToken{"op", "||"})
			i += 2
		case c == '!':
			toks = append(toks, ghToken{"bang", "!"})
			i++
		default:
			j := i
			for j < len(s) && isIdentByte(s[j]) {
				j++
			}
			if j == i {
				return nil, fmt.Errorf("unexpected character %q at byte %d in %q", c, i, s)
			}
			toks = append(toks, ghToken{"ident", s[i:j]})
			i = j
		}
	}
	return toks, nil
}

type ghParser struct {
	toks []ghToken
	pos  int
	ctx  map[string]any
}

func (p *ghParser) peek() (ghToken, bool) {
	if p.pos >= len(p.toks) {
		return ghToken{}, false
	}
	return p.toks[p.pos], true
}

func (p *ghParser) next() (ghToken, bool) {
	tok, ok := p.peek()
	if ok {
		p.pos++
	}
	return tok, ok
}

func (p *ghParser) parseExpr() (any, error) { return p.parseOr() }

func (p *ghParser) parseOr() (any, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		tok, ok := p.peek()
		if !ok || tok.kind != "op" || tok.val != "||" {
			return left, nil
		}
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		if !ghTruthy(left) {
			left = right
		}
	}
}

func (p *ghParser) parseAnd() (any, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for {
		tok, ok := p.peek()
		if !ok || tok.kind != "op" || tok.val != "&&" {
			return left, nil
		}
		p.next()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		if ghTruthy(left) {
			left = right
		}
	}
}

func (p *ghParser) parseEquality() (any, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		tok, ok := p.peek()
		if !ok || tok.kind != "op" || (tok.val != "==" && tok.val != "!=") {
			return left, nil
		}
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		eq := ghEquals(left, right)
		if tok.val == "!=" {
			left = !eq
		} else {
			left = eq
		}
	}
}

func (p *ghParser) parseUnary() (any, error) {
	if tok, ok := p.peek(); ok && tok.kind == "bang" {
		p.next()
		v, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return !ghTruthy(v), nil
	}
	return p.parsePrimary()
}

func (p *ghParser) parsePrimary() (any, error) {
	tok, ok := p.next()
	if !ok {
		return nil, fmt.Errorf("unexpected end of expression")
	}
	switch tok.kind {
	case "lparen":
		v, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		closing, ok := p.next()
		if !ok || closing.kind != "rparen" {
			return nil, fmt.Errorf("expected ) in expression")
		}
		return v, nil
	case "string":
		return tok.val, nil
	case "ident":
		// Graph's fan-in deliberately runs after failed dependencies.
		// always() is true regardless of caller status; reject arguments
		// and keep every other function outside this closed evaluator.
		if next, ok := p.peek(); ok && next.kind == "lparen" && tok.val == "always" {
			p.next()
			closing, ok := p.next()
			if !ok || closing.kind != "rparen" {
				return nil, fmt.Errorf("always() takes no arguments")
			}
			return true, nil
		}
		switch tok.val {
		case "true":
			return true, nil
		case "false":
			return false, nil
		case "null":
			return nil, nil
		}
		v, ok := p.ctx[tok.val]
		if !ok {
			// An identifier the caller's context does not mention (for
			// example a deleted fork's head.repo.full_name, or any
			// github.event.pull_request.* field on merge_group) is null in
			// the real evaluator: falsy, == null and == '' (both 0), unequal
			// to every non-empty, non-numeric literal.
			return nil, nil
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected token %+v", tok)
	}
}

func ghTruthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x != ""
	case nil:
		return false
	default:
		return true
	}
}

// ghEquals: GitHub's loose equality for the operand kinds this evaluator
// produces. Two strings compare case-insensitively; two booleans, or two
// nulls, directly; otherwise both sides are coerced to numbers (null and ”
// to 0, true to 1, false to 0, any other string to its numeric value or
// NaN) and NaN equals nothing. So `true == 'true'` is false, as on GitHub
// (the classic workflow_call boolean-input bug), and `false == ”` is true.
func ghEquals(a, b any) bool {
	if sa, ok := a.(string); ok {
		if sb, ok := b.(string); ok {
			return strings.EqualFold(sa, sb)
		}
	}
	if ba, ok := a.(bool); ok {
		if bb, ok := b.(bool); ok {
			return ba == bb
		}
	}
	if a == nil && b == nil {
		return true
	}
	na, nb := ghNumber(a), ghNumber(b)
	return !math.IsNaN(na) && !math.IsNaN(nb) && na == nb
}

// ghNumber: GitHub's coercion of an operand to a number.
func ghNumber(v any) float64 {
	switch x := v.(type) {
	case nil:
		return 0
	case bool:
		if x {
			return 1
		}
		return 0
	case string:
		t := strings.TrimSpace(x)
		if t == "" {
			return 0
		}
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return math.NaN()
		}
		return f
	default:
		return math.NaN()
	}
}

// evalGHExpr evaluates a GitHub Actions `${{ ... }}` expression (the wrapper
// is optional) against ctx, a map from dotted identifier (e.g.
// "github.event_name") to its string value. See the package comment above
// for the supported subset.
func evalGHExpr(expr string, ctx map[string]string) (any, error) {
	typed := make(map[string]any, len(ctx))
	for k, v := range ctx {
		typed[k] = v
	}
	return evalGHExprTyped(expr, typed)
}

// evalGHExprTyped: evalGHExpr with typed context values (a bool for a
// boolean input such as inputs.fresh-test-results, nil for null).
func evalGHExprTyped(expr string, ctx map[string]any) (any, error) {
	expr = strings.TrimSpace(expr)
	expr = strings.TrimPrefix(expr, "${{")
	expr = strings.TrimSuffix(expr, "}}")
	expr = strings.TrimSpace(expr)
	toks, err := ghTokenize(expr)
	if err != nil {
		return nil, err
	}
	p := &ghParser{toks: toks, ctx: ctx}
	v, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.pos != len(p.toks) {
		return nil, fmt.Errorf("trailing tokens after expression %q: %v", expr, p.toks[p.pos:])
	}
	return v, nil
}

// mustEvalGHRunsOn evaluates a `runs-on: ${{ ... }}`-style expression and
// returns the resulting runner-label string, failing the test if the
// expression does not evaluate to a string.
func mustEvalGHRunsOn(t *testing.T, expr string, ctx map[string]string) string {
	t.Helper()
	v, err := evalGHExpr(expr, forkUpstreamWorld(ctx))
	if err != nil {
		t.Fatalf("evalGHExpr(%q): %v", expr, err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("evalGHExpr(%q) = %#v (%T), want string", expr, v, v)
	}
	return s
}

// TestSameRepoBlacksmithExpressionSemantics runs the real, pinned
// sameRepoBlacksmith{2,4,8}vcpu expression strings through evalGHExpr (not a
// hand-written mirror of their logic) against every event shape the policy
// cares about: trusted same-repo PRs and merge_group get Blacksmith; forks,
// Dependabot, a deleted fork head, and any other event fall back to
// ubuntu-latest. Because this evaluates the actual expression text, a
// semantic typo in any of the three consts (a dropped `!`, a swapped `&&`/
// `||`, a wrong field name) fails this test even if a hand-rolled Go mirror
// would have been edited to match it.
func TestSameRepoBlacksmithExpressionSemantics(t *testing.T) {
	const ownRepo = "steveyegge/beads"
	type tc struct {
		name       string
		event      string
		headRepo   string // github.event.pull_request.head.repo.full_name; "" = fork or deleted fork head
		actor      string
		wantRunner bool
	}
	cases := []tc{
		{"same-repo PR, human actor", "pull_request", ownRepo, "alice", true},
		{"merge_group always Blacksmith", "merge_group", "", "", true},
		{"fork PR stays ubuntu-latest", "pull_request", "someone-else/beads", "alice", false},
		{"deleted fork head stays ubuntu-latest", "pull_request", "", "alice", false},
		{"same-repo PR, dependabot actor stays ubuntu-latest", "pull_request", ownRepo, "dependabot[bot]", false},
		{"push stays ubuntu-latest", "push", "", "alice", false},
		{"pull_request_target stays ubuntu-latest", "pull_request_target", ownRepo, "alice", false},
		{"schedule stays ubuntu-latest", "schedule", "", "", false},
		{"workflow_dispatch stays ubuntu-latest", "workflow_dispatch", "", "", false},
	}
	type constCase struct {
		label    string // the Blacksmith label the expression resolves to when wantRunner
		fallback string // the GitHub-hosted label it falls back to otherwise
		expr     string
	}
	consts := []constCase{
		{"blacksmith-2vcpu-ubuntu-2404", "ubuntu-latest", sameRepoBlacksmith2vcpu},
		{"blacksmith-4vcpu-ubuntu-2404", "ubuntu-latest", sameRepoBlacksmith4vcpu},
		{"blacksmith-8vcpu-ubuntu-2404", "ubuntu-latest", sameRepoBlacksmith8vcpu},
		{"blacksmith-2vcpu-windows-2025", "windows-latest", sameRepoBlacksmithWindows2vcpu},
		{"blacksmith-4vcpu-windows-2025", "windows-latest", sameRepoBlacksmithWindows4vcpu},
		{"blacksmith-8vcpu-windows-2025", "windows-latest", sameRepoBlacksmithWindows8vcpu},
		{"blacksmith-16vcpu-windows-2025", "windows-latest", sameRepoBlacksmithWindows16vcpu},
		{"blacksmith-6vcpu-macos-26", "macos-latest", sameRepoBlacksmithMacOS6vcpu},
		{"blacksmith-12vcpu-macos-26", "macos-latest", sameRepoBlacksmithMacOS12vcpu},
	}
	for _, c := range cases {
		ctx := map[string]string{
			"github.event_name":                             c.event,
			"github.event.pull_request.head.repo.full_name": c.headRepo,
			"github.repository":                             ownRepo,
			"github.repository_owner":                       "gastownhall",
			"github.actor":                                  c.actor,
		}
		for _, cst := range consts {
			t.Run(c.name+"/"+cst.label, func(t *testing.T) {
				got := mustEvalGHRunsOn(t, cst.expr, ctx)
				want := cst.fallback
				if c.wantRunner {
					want = cst.label
				}
				if got != want {
					t.Errorf("%s: real evaluator on %q = %q, want %q", c.name, cst.expr, got, want)
				}
			})
		}
	}
}

// TestSameRepoPlatformsMatrixMarkerRunsOnExpressionSemantics runs the real,
// pinned sameRepoPlatformsMatrixMarkerRunsOn chained-ternary expression
// (pr-preflight-platforms' and check-doc-freshness-platforms' runs-on, F7b
// §2.1) through evalGHExpr for every (marker, event) combination the policy
// cares about: the linux/windows `runner` markers resolve to that OS's
// Blacksmith label only on a trusted same-repo PR/merge_group, falling back
// to that OS's GitHub-hosted label otherwise (the macos marker included:
// blacksmith-6vcpu-macos-26 vs. macos-latest); and a leg with no marker at
// all (there is none today, but the fallback must still be safe) falls back
// to matrix.os. Every marker is additionally checked against its flat
// same-repo const for every event shape, so the chained ternary's per-OS
// arms cannot drift from the single-OS expressions.
func TestSameRepoPlatformsMatrixMarkerRunsOnExpressionSemantics(t *testing.T) {
	const ownRepo = "steveyegge/beads"
	trustedCtx := map[string]string{
		"github.event_name":                             "pull_request",
		"github.event.pull_request.head.repo.full_name": ownRepo,
		"github.repository":                             ownRepo,
		"github.repository_owner":                       "gastownhall",
		"github.actor":                                  "alice",
	}
	forkCtx := map[string]string{
		"github.event_name":                             "pull_request",
		"github.event.pull_request.head.repo.full_name": "someone-else/beads",
		"github.repository":                             ownRepo,
		"github.repository_owner":                       "gastownhall",
		"github.actor":                                  "alice",
	}
	cases := []struct {
		name       string
		marker     string // matrix.runner
		os         string // matrix.os
		ctx        map[string]string
		wantRunsOn string
	}{
		{"linux marker, trusted", "same-repo-linux", "ubuntu-latest", trustedCtx, "blacksmith-4vcpu-ubuntu-2404"},
		{"linux marker, fork", "same-repo-linux", "ubuntu-latest", forkCtx, "ubuntu-latest"},
		{"windows marker, trusted", "same-repo-windows", "windows-latest", trustedCtx, "blacksmith-4vcpu-windows-2025"},
		{"windows marker, fork", "same-repo-windows", "windows-latest", forkCtx, "windows-latest"},
		{"macos marker, trusted", "same-repo-macos", "macos-latest", trustedCtx, blacksmithMacOSLabel},
		{"macos marker, fork", "same-repo-macos", "macos-latest", forkCtx, "macos-latest"},
		{"no marker falls back to matrix.os", "", "some-other-os", trustedCtx, "some-other-os"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := map[string]string{}
			for k, v := range c.ctx {
				ctx[k] = v
			}
			ctx["matrix.runner"] = c.marker
			ctx["matrix.os"] = c.os
			got := mustEvalGHRunsOn(t, sameRepoPlatformsMatrixMarkerRunsOn, ctx)
			if got != c.wantRunsOn {
				t.Errorf("%s: runs-on = %q, want %q", c.name, got, c.wantRunsOn)
			}
		})
	}

	// Each marker arm must resolve exactly like its flat single-OS const for
	// every event shape: same-repo PR and merge_group get Blacksmith; fork,
	// deleted fork head, Dependabot, push and every other event keep the
	// GitHub-hosted label.
	markerConsts := []struct{ marker, os, flat string }{
		{"same-repo-linux", "ubuntu-latest", sameRepoBlacksmith4vcpu},
		{"same-repo-windows", "windows-latest", sameRepoBlacksmithWindows4vcpu},
		{"same-repo-macos", "macos-latest", sameRepoBlacksmithMacOS6vcpu},
	}
	events := []struct{ name, event, headRepo, actor string }{
		{"same-repo PR", "pull_request", ownRepo, "alice"},
		{"merge_group", "merge_group", "", ""},
		{"fork PR", "pull_request", "someone-else/beads", "alice"},
		{"deleted fork head", "pull_request", "", "alice"},
		{"dependabot", "pull_request", ownRepo, "dependabot[bot]"},
		{"push", "push", "", "alice"},
		{"schedule", "schedule", "", ""},
	}
	for _, m := range markerConsts {
		for _, e := range events {
			t.Run(m.marker+"/"+e.name, func(t *testing.T) {
				ctx := map[string]string{
					"github.event_name":                             e.event,
					"github.event.pull_request.head.repo.full_name": e.headRepo,
					"github.repository":                             ownRepo,
					"github.actor":                                  e.actor,
					"matrix.runner":                                 m.marker,
					"matrix.os":                                     m.os,
				}
				got := mustEvalGHRunsOn(t, sameRepoPlatformsMatrixMarkerRunsOn, ctx)
				want := mustEvalGHRunsOn(t, m.flat, ctx)
				if got != want {
					t.Errorf("marker %s under %s: runs-on = %q, flat const resolves to %q", m.marker, e.name, got, want)
				}
			})
		}
	}
}

// TestBlacksmithBazelRunnerSizes evaluates every bazel.yml lane's real
// runs-on (and package-mcp's xdist worker count) under each rbe mode: modes
// remote and cache get the lane's Blacksmith size (mode cache reads
// rbe-cache, which refuses GitHub-hosted runners), every other mode the
// GitHub-hosted runner, and package-mcp runs 16 xdist workers exactly when
// it is on its 8 vCPU runner (8 everywhere else, package-mcp.sh's default).
func TestBlacksmithBazelRunnerSizes(t *testing.T) {
	workflow := readCIWorkflow(t, bazelWorkflowName)
	wantRemote := map[string]string{
		bazelJobName:           "blacksmith-8vcpu-ubuntu-2404",
		bazelPackageMCPJobName: "blacksmith-8vcpu-ubuntu-2404",
		bazelPackageNPMJobName: "blacksmith-4vcpu-ubuntu-2404",
		bazelRBEPrewarmJobName: "blacksmith-2vcpu-ubuntu-2404",
	}
	mcpGate := workflow.job(t, bazelPackageMCPJobName).step(t, "Run MCP package gate")
	if got := mcpGate.Env["BEADS_MCP_PYTEST_WORKERS"]; got != bazelMCPPytestWorkers {
		t.Errorf("package-mcp Run MCP package gate BEADS_MCP_PYTEST_WORKERS = %q, want %q", got, bazelMCPPytestWorkers)
	}
	for _, mode := range []string{"remote", "fork-ro", "fork-rw", "local", "cache", "skip"} {
		ctx := map[string]string{"needs.rbe.outputs.mode": mode, "github.repository_owner": "gastownhall"}
		for name, job := range workflow.Jobs {
			if name == bazelRBEJobName || isBazelRRCJob(name) {
				continue // rrc jobs: push/schedule only, never a fork (literal label)
			}
			if name == "bazel-dispatch-coverage" || name == "bazel-local-coverage" {
				if job.RunsOn != "ubuntu-latest" {
					t.Errorf("result-only %s must stay GitHub-hosted, got %q", name, job.RunsOn)
				}
				continue
			}
			want := "ubuntu-latest"
			if isBazelRRCCacheReader(name) || name == bazelServerHostedJobName || name == bazelEmbedJobName {
				want = "ubuntu-24.04"
			}
			if bazelBlacksmithModes[mode] {
				want = "blacksmith-4vcpu-ubuntu-2404"
				if label, ok := wantRemote[name]; ok {
					want = label
				}
			}
			if got := mustEvalGHRunsOn(t, job.RunsOn, ctx); got != want {
				t.Errorf("%s runs-on under mode %s = %q, want %q", name, mode, got, want)
			}
		}
		wantWorkers := "8"
		if mustEvalGHRunsOn(t, workflow.job(t, bazelPackageMCPJobName).RunsOn, ctx) == "blacksmith-8vcpu-ubuntu-2404" {
			wantWorkers = "16"
		}
		if got := mustEvalGHRunsOn(t, mcpGate.Env["BEADS_MCP_PYTEST_WORKERS"], ctx); got != wantWorkers {
			t.Errorf("package-mcp BEADS_MCP_PYTEST_WORKERS under mode %s = %q, want %q", mode, got, wantWorkers)
		}
	}
}

func TestGHScopeAlwaysFunctionIsClosed(t *testing.T) {
	for _, test := range []struct {
		expr string
		want bool
	}{
		{"always()", true}, {"false && always()", false},
		{"always() && false", false}, {"!always()", false},
		{"needs.scope.outputs.linux == 'true' && (always())", true},
	} {
		got, err := evalGHExpr(test.expr, map[string]string{"needs.scope.outputs.linux": "true"})
		if err != nil || ghTruthy(got) != test.want {
			t.Errorf("%q = %v/%v, want%v", test.expr, got, err, test.want)
		}
	}
	for _, expr := range []string{"always(true)", "always(", "unknown()", "success()"} {
		if _, err := evalGHExpr(expr, nil); err == nil {
			t.Errorf("unsupported call accepted: %q", expr)
		}
	}
}

// Read the actual job expression so a lost native-only condition cannot hide
// behind an updated expected literal. Other matrix legs retain their hosts.
func TestNativeLintHostedFallbackMatchesLLVMHost(t *testing.T) {
	expr := readCIWorkflow(t, "pr.yml").job(t, "pr-lint-wrapper").RunsOn
	const ownRepo = "gastownhall/beads"
	cases := []struct {
		name, owner, event, head, actor string
		trusted                         bool
	}{
		{"upstream same repo", "gastownhall", "pull_request", ownRepo, "alice", true},
		{"upstream merge group", "gastownhall", "merge_group", "", "alice", true},
		{"fork head", "gastownhall", "pull_request", "other/beads", "alice", false},
		{"dependabot", "gastownhall", "pull_request", ownRepo, "dependabot[bot]", false},
		{"missing head", "gastownhall", "pull_request", "", "alice", false},
		{"fork owner", "donnabox", "pull_request", "donnabox/beads", "alice", false},
	}
	for _, c := range cases {
		for _, target := range []string{"native", "windows", "darwin"} {
			t.Run(c.name+"/"+target, func(t *testing.T) {
				ctx := map[string]string{
					"github.repository_owner":                       c.owner,
					"github.repository":                             ownRepo,
					"github.event_name":                             c.event,
					"github.event.pull_request.head.repo.full_name": c.head,
					"github.actor":                                  c.actor,
					"matrix.target":                                 target,
				}
				want := "ubuntu-latest"
				if target == "native" {
					want = "ubuntu-24.04"
				}
				if c.trusted {
					want = "blacksmith-4vcpu-ubuntu-2404"
				}
				if got := mustEvalGHRunsOn(t, expr, ctx); got != want {
					t.Errorf("actual lint runner = %q, want %q", got, want)
				}
			})
		}
	}
}

// Evaluate the actual hosted-storage expression under both owners and every
// execution mode; the host ABI pin does not make this local-only job an RRC reader.
func TestHostedStorageFallbackMatchesLLVMHost(t *testing.T) {
	expr := readCIWorkflow(t, bazelWorkflowName).job(t, bazelServerHostedJobName).RunsOn
	if isBazelRRCCacheReader(bazelServerHostedJobName) {
		t.Fatal("hosted storage must not become a trusted RRC reader")
	}
	for _, owner := range []string{"gastownhall", "donnabox"} {
		for _, mode := range []string{"remote", "cache", "local", "fork-ro", "fork-rw", "skip"} {
			t.Run(owner+"/"+mode, func(t *testing.T) {
				ctx := map[string]string{
					"github.repository_owner": owner,
					"needs.rbe.outputs.mode":  mode,
				}
				want := "ubuntu-24.04"
				if owner == "gastownhall" && (mode == "remote" || mode == "cache") {
					want = "blacksmith-4vcpu-ubuntu-2404"
				}
				if got := mustEvalGHRunsOn(t, expr, ctx); got != want {
					t.Errorf("hosted storage runner = %q, want %q", got, want)
				}
			})
		}
	}
}
