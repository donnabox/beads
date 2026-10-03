// Package replay_test holds the layout guards for the replay harness: the
// structural rules the harness keeps by construction, checked over the
// library tree and the thin command wrappers that stay under scripts/.
//
//   - TestB1OneRunner: dolt is spawned from exactly one package.
//   - TestB1NoCopiedHelpers: the shared helpers are each defined once.
//   - TestB1Hygiene: no tracker ids, home paths, production ports or incident
//     names appear anywhere in the harness source, comments and fixtures alike.
package replay_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/testutil/bazeltest"
)

// replayTree is the harness library tree; thinMains are the command wrappers
// that stay under scripts/. Every guard scans exactly these.
const replayTree = "internal/replay"

var thinMains = []string{
	"scripts/oracle-query",
	"scripts/mutation-translator",
	"scripts/driver-core",
}

type goFile struct {
	rel    string // slash-separated, relative to the repository root
	isTest bool
}

// harnessFiles lists the Go files under the scan roots. Each root must hold at
// least one file, so a moved or renamed directory fails the guard instead of
// letting it pass over nothing.
func harnessFiles(t *testing.T) (string, []goFile) {
	t.Helper()
	root := bazeltest.RepoRoot(t)
	var files []goFile
	scan := func(rel string, recursive bool) {
		dir := filepath.Join(root, filepath.FromSlash(rel))
		found := 0
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if p != dir && (!recursive || d.Name() == "testdata") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(p, ".go") {
				return nil
			}
			r, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			files = append(files, goFile{rel: filepath.ToSlash(r), isTest: strings.HasSuffix(p, "_test.go")})
			found++
			return nil
		})
		if err != nil {
			t.Fatalf("scanning %s: %v", rel, err)
		}
		if found == 0 {
			t.Fatalf("no Go files under %s: the guard would pass over nothing", rel)
		}
	}
	scan(replayTree, true)
	for _, m := range thinMains {
		scan(m, false)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })
	return root, files
}

func parseFile(t *testing.T, fset *token.FileSet, root, rel string) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(fset, filepath.Join(root, filepath.FromSlash(rel)), nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", rel, err)
	}
	return f
}

// spawnsDolt reports whether f starts, or looks up in order to start, a
// process named "dolt": exec.Command, exec.CommandContext or exec.LookPath
// given the string literal "dolt". A name passed through a variable is not
// seen, so the runner keeps the literal.
func spawnsDolt(f *ast.File) bool {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "exec" {
			return true
		}
		switch sel.Sel.Name {
		case "Command", "CommandContext", "LookPath":
		default:
			return true
		}
		for _, arg := range call.Args {
			if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING && lit.Value == `"dolt"` {
				found = true
			}
		}
		return true
	})
	return found
}

func describe(byKey map[string][]string) string {
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s (%s)", k, strings.Join(byKey[k], ", ")))
	}
	return strings.Join(parts, "; ")
}

// B1.OneRunner: dolt is spawned from exactly one package. Production code
// only; fixtures reach dolt through the same package.
func TestB1OneRunner(t *testing.T) {
	root, files := harnessFiles(t)
	fset := token.NewFileSet()
	spawning := map[string][]string{} // package directory -> files
	for _, f := range files {
		if f.isTest {
			continue
		}
		if spawnsDolt(parseFile(t, fset, root, f.rel)) {
			dir := path.Dir(f.rel)
			spawning[dir] = append(spawning[dir], path.Base(f.rel))
		}
	}
	const runner = "internal/replay/doltcli"
	if len(spawning) != 1 || len(spawning[runner]) == 0 {
		t.Errorf("dolt must be spawned from exactly one package, %s; found %d: %s", runner, len(spawning), describe(spawning))
	}
}

// helperAliases maps each shared helper to the names one definition of it may
// take: the lower-case form used inside a package and the exported form used
// once it is shared.
var helperAliases = map[string][]string{
	"sanitizedEnv": {"sanitizedEnv", "SanitizedEnv"},
	"doltQuery":    {"doltQuery", "DoltQuery", "Query"},
	"rowMap":       {"rowMap", "RowMap"},
	"sqlQuote":     {"sqlQuote", "SQLQuote", "SqlQuote"},
}

// B1.NoCopiedHelpers: each shared helper is defined once across the harness,
// tests included.
func TestB1NoCopiedHelpers(t *testing.T) {
	root, files := harnessFiles(t)
	fset := token.NewFileSet()
	defs := map[string][]string{} // helper -> "file:line" of each definition
	for _, f := range files {
		parsed := parseFile(t, fset, root, f.rel)
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			for helper, aliases := range helperAliases {
				for _, alias := range aliases {
					if fn.Name.Name == alias {
						defs[helper] = append(defs[helper], fmt.Sprintf("%s:%d", f.rel, fset.Position(fn.Pos()).Line))
					}
				}
			}
		}
	}
	names := make([]string, 0, len(helperAliases))
	for helper := range helperAliases {
		names = append(names, helper)
	}
	sort.Strings(names)
	for _, helper := range names {
		if got := defs[helper]; len(got) != 1 {
			t.Errorf("%s must be defined exactly once, found %d: %s", helper, len(got), strings.Join(got, ", "))
		}
	}
}

// hygienePatterns are assembled from fragments so this file cannot match its
// own patterns.
var hygienePatterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{"tracker id", regexp.MustCompile(`\b(be|gm|ga|mc|wy)-[0-9a-z]{4,}(\.[0-9]+)*\b`)},
	{"home path", regexp.MustCompile("/ho" + "me/")},
	{"production port", regexp.MustCompile(`\b28` + `231\b`)},
	{"incident name", regexp.MustCompile("cai" + "rn|beads-test" + "db|production-" + "leak")},
}

// B1.Hygiene: the harness source carries no internal tracker ids, no
// home-directory paths, no production port numbers and no incident names.
// Every line is checked, so string literals in fixtures count as well as
// comments.
func TestB1Hygiene(t *testing.T) {
	root, files := harnessFiles(t)
	var hits []string
	badFiles := map[string]bool{}
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.rel)))
		if err != nil {
			t.Fatalf("reading %s: %v", f.rel, err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			for _, p := range hygienePatterns {
				if p.re.MatchString(line) {
					text := strings.TrimSpace(line)
					if len(text) > 100 {
						text = text[:100] + "..."
					}
					hits = append(hits, fmt.Sprintf("%s:%d: %s: %s", f.rel, i+1, p.name, text))
					badFiles[f.rel] = true
					break
				}
			}
		}
	}
	if len(hits) > 0 {
		t.Errorf("%d matching lines in %d files:\n%s", len(hits), len(badFiles), strings.Join(hits, "\n"))
	}
}
