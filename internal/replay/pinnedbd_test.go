package replay_test

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"
)

// looksUpBd returns the positions in f where bd is started, or looked up in
// order to start, by name: exec.Command, exec.CommandContext or exec.LookPath
// given the string literal "bd". A name passed through a variable is not seen,
// so production code that starts bd takes its path from the caller instead.
func looksUpBd(fset *token.FileSet, f *ast.File) []string {
	var found []string
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
			if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING && lit.Value == `"bd"` {
				found = append(found, fset.Position(call.Pos()).String())
			}
		}
		return true
	})
	return found
}

// B3.PinnedBd, source guard: nothing in the harness finds bd by searching PATH.
// The code under test is whichever binary the caller names (the driver's own
// build, a test's build from this tree), so a bd that merely happens to be first
// on PATH can never stand in for it. Production code only: a test may put a
// stand-in on PATH precisely to show it is ignored.
func TestB3PinnedBdIsNeverLookedUp(t *testing.T) {
	root, files := harnessFiles(t)
	fset := token.NewFileSet()
	var hits []string
	scanned := 0
	for _, f := range files {
		if f.isTest {
			continue
		}
		scanned++
		hits = append(hits, looksUpBd(fset, parseFile(t, fset, root, f.rel))...)
	}
	if scanned == 0 {
		t.Fatal("no production files scanned: the guard would pass over nothing")
	}
	if len(hits) > 0 {
		t.Errorf("bd is looked up by name in %d place(s); take the path from the caller instead:\n%s", len(hits), strings.Join(hits, "\n"))
	}
}
