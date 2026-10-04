package driver

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/replay/doltcli"
	"github.com/steveyegge/beads/internal/replay/replaytest"
)

// commitHashes maps each commit message of a dolt database to its hash. The
// fixtures give every commit a distinct message, so a message names a commit.
func commitHashes(t *testing.T, dir string) map[string]string {
	t.Helper()
	_, rows, err := doltcli.Query(context.Background(), dir, "SELECT message, commit_hash FROM dolt_log")
	if err != nil {
		t.Fatalf("reading dolt_log of %s: %v", dir, err)
	}
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		out[r[0].Text] = r[1].Text
	}
	return out
}

// commitAll stages and commits whatever the fixture just changed.
func commitAll(t *testing.T, dir, message string) {
	t.Helper()
	replaytest.RunDolt(t, dir, "add", "-A")
	replaytest.RunDolt(t, dir, "commit", "-m", message)
}

// scratchDAG builds the history the walk has to get right: a base, a mainline
// commit m1, a side branch of two commits f1 and f2 cut from the base, and the
// merge of that branch into the mainline.
//
//	init -- base -- m1 ------- merge
//	           \             /
//	            f1 -- f2 ----
//
// m1 and f1 sit at the same depth, which is the tie that makes an ordering by
// commit_order pair siblings.
func scratchDAG(t *testing.T) (dir string, hashes map[string]string) {
	t.Helper()
	dir = replaytest.NewDoltDB(t, "dag")
	replaytest.RunDolt(t, dir, "sql", "-q", "CREATE TABLE issues (id VARCHAR(64) PRIMARY KEY, title VARCHAR(255))")
	commitAll(t, dir, "bd init")
	replaytest.RunDolt(t, dir, "checkout", "-b", "feat")
	replaytest.RunDolt(t, dir, "sql", "-q", "INSERT INTO issues VALUES ('f-1', 'first on the branch')")
	commitAll(t, dir, "f1")
	replaytest.RunDolt(t, dir, "sql", "-q", "INSERT INTO issues VALUES ('f-2', 'second on the branch')")
	commitAll(t, dir, "f2")
	replaytest.RunDolt(t, dir, "checkout", "main")
	replaytest.RunDolt(t, dir, "sql", "-q", "INSERT INTO issues VALUES ('m-1', 'on the mainline')")
	commitAll(t, dir, "m1")
	replaytest.RunDolt(t, dir, "merge", "feat", "--no-ff", "-m", "merge")
	return dir, commitHashes(t, dir)
}

// B6.WalkDAG: the walk follows first parents from HEAD, so the steps over
// base, m1 || f1, f2, merge are base to m1 and m1 to merge, and the merge step
// is tagged. The side branch is covered by the merge step's net diff and is
// never a step of its own.
func TestB6WalkDAG(t *testing.T) {
	replaytest.Require(t, replaytest.NeedDolt)
	dir, h := scratchDAG(t)

	w, err := ReadWalk(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadWalk: %v", err)
	}
	steps := w.Steps(0)
	if len(steps) != 2 {
		t.Fatalf("got %d steps, want 2 (base to m1, m1 to merge): %+v", len(steps), steps)
	}
	want := []struct {
		from, to string
		merge    bool
	}{
		{h["bd init"], h["m1"], false},
		{h["m1"], h["merge"], true},
	}
	for i, wnt := range want {
		got := steps[i]
		if got.From.Hash != wnt.from || got.To.Hash != wnt.to || got.Merge != wnt.merge || got.Net {
			t.Errorf("step %d = %s to %s merge=%v net=%v, want %s to %s merge=%v net=false",
				i, got.From.Hash, got.To.Hash, got.Merge, got.Net, wnt.from, wnt.to, wnt.merge)
		}
		if got.Index != i {
			t.Errorf("step %d has Index %d", i, got.Index)
		}
	}
	for _, side := range []string{"f1", "f2"} {
		for _, st := range steps {
			if st.From.Hash == h[side] || st.To.Hash == h[side] {
				t.Errorf("side-branch commit %s is part of a step: %+v", side, st)
			}
		}
	}
	// Every commit reachable from HEAD is counted, the side branch included, so
	// the summary can say how much of the history the steps stand for.
	if w.Total != 6 {
		t.Errorf("Total = %d, want 6 (init, base, m1, f1, f2, merge)", w.Total)
	}
	if !reflect.DeepEqual(w.Head().Parents, []string{h["m1"], h["f2"]}) {
		t.Errorf("merge parents = %v, want [m1 f2] with the mainline first", w.Head().Parents)
	}
}

// dagCommits is the scratch DAG's commits, by hand, for the tests that need no
// dolt. The merge lists the mainline first.
func dagCommits() (commits []Commit, head string) {
	commits = []Commit{
		{Hash: "r", Message: "Initialize data repository"},
		{Hash: "b", Parents: []string{"r"}, Message: "bd init"},
		{Hash: "m1", Parents: []string{"b"}, Message: "m1"},
		{Hash: "f1", Parents: []string{"b"}, Message: "f1"},
		{Hash: "f2", Parents: []string{"f1"}, Message: "f2"},
		{Hash: "merge", Parents: []string{"m1", "f2"}, Message: "merge"},
	}
	return commits, "merge"
}

func hashesOf(commits []Commit) []string {
	out := make([]string, len(commits))
	for i, c := range commits {
		out[i] = c.Hash
	}
	return out
}

// B6.WalkDAG, property: the chain depends on the parent pointers and on
// nothing else. dolt_log hands its rows back in no promised order and gives
// siblings the same commit_order, so the rows are shuffled a hundred times and
// the chain never changes.
func TestB6WalkChainIgnoresRowOrder(t *testing.T) {
	commits, head := dagCommits()
	want := []string{"r", "b", "m1", "merge"}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 100; i++ {
		shuffled := append([]Commit(nil), commits...)
		rng.Shuffle(len(shuffled), func(a, b int) { shuffled[a], shuffled[b] = shuffled[b], shuffled[a] })
		chain, err := FirstParentChain(shuffled, head)
		if err != nil {
			t.Fatalf("shuffle %d: FirstParentChain: %v", i, err)
		}
		if got := hashesOf(chain); !reflect.DeepEqual(got, want) {
			t.Fatalf("shuffle %d: chain = %v, want %v", i, got, want)
		}
	}
}

func TestB6FirstParentChainRefusesABrokenLog(t *testing.T) {
	commits, head := dagCommits()
	t.Run("unknown head", func(t *testing.T) {
		if _, err := FirstParentChain(commits, "nope"); err == nil {
			t.Fatal("a head that is not in the log was accepted")
		}
	})
	t.Run("parent missing from the log", func(t *testing.T) {
		broken := []Commit{{Hash: "a", Parents: []string{"gone"}}}
		if _, err := FirstParentChain(broken, "a"); err == nil {
			t.Fatal("a parent that is not in the log was accepted")
		}
	})
	t.Run("cycle", func(t *testing.T) {
		cyc := []Commit{{Hash: "a", Parents: []string{"b"}}, {Hash: "b", Parents: []string{"a"}}}
		if _, err := FirstParentChain(cyc, "a"); err == nil {
			t.Fatal("a cycle was walked without an error")
		}
	})
	t.Run("a root alone", func(t *testing.T) {
		chain, err := FirstParentChain(commits[:1], "r")
		if err != nil || len(chain) != 1 {
			t.Fatalf("chain = %v, err = %v; want the root alone", chain, err)
		}
	})
}

// The bootstrap collapse keeps its rule along the chain: the base is the last
// bootstrap commit on it, and what came before is not replayed.
func TestB6WalkBootstrapCollapse(t *testing.T) {
	chainOf := func(messages ...string) []Commit {
		var out []Commit
		for i, m := range messages {
			c := Commit{Hash: fmt.Sprintf("c%d", i), Message: m}
			if i > 0 {
				c.Parents = []string{fmt.Sprintf("c%d", i-1)}
			}
			out = append(out, c)
		}
		return out
	}

	t.Run("base is the last bootstrap commit", func(t *testing.T) {
		chain := chainOf("Initialize data repository", "bd init", "schema: add a column", "create nonlocal table x", "bd: create", "bd: update")
		w := newWalk(chain, len(chain))
		if got := hashesOf(w.chain); !reflect.DeepEqual(got, []string{"c3", "c4", "c5"}) {
			t.Fatalf("chain after the collapse = %v, want [c3 c4 c5]", got)
		}
		if w.Total != 6 {
			t.Errorf("Total = %d, want 6: the collapsed commits still count", w.Total)
		}
		if got := len(w.Steps(0)); got != 2 {
			t.Errorf("got %d steps, want 2", got)
		}
		if w.Base().Hash != "c3" || w.Head().Hash != "c5" {
			t.Errorf("base %s head %s, want c3 and c5", w.Base().Hash, w.Head().Hash)
		}
	})
	t.Run("no bootstrap commit keeps the whole chain", func(t *testing.T) {
		chain := chainOf("one", "two", "three")
		w := newWalk(chain, len(chain))
		if len(w.chain) != 3 {
			t.Fatalf("chain = %v, want all three commits", hashesOf(w.chain))
		}
	})
}

// A sampled step spans commits of the chain, so it is a net diff; an adjacent
// pair of samples is an ordinary step.
func TestB6WalkSampledSteps(t *testing.T) {
	var chain []Commit
	for i := 0; i < 10; i++ {
		c := Commit{Hash: fmt.Sprintf("c%d", i), Message: "m"}
		if i > 0 {
			c.Parents = []string{fmt.Sprintf("c%d", i-1)}
		}
		chain = append(chain, c)
	}
	w := newWalk(chain, len(chain))

	exhaustive := w.Steps(0)
	if len(exhaustive) != 9 {
		t.Fatalf("exhaustive steps = %d, want 9", len(exhaustive))
	}
	for _, st := range exhaustive {
		if st.Net {
			t.Errorf("an exhaustive step is net: %+v", st)
		}
	}

	sampled := w.Steps(3)
	if len(sampled) != 2 {
		t.Fatalf("sampled steps = %d, want 2", len(sampled))
	}
	for i, st := range sampled {
		if !st.Net {
			t.Errorf("sampled step %d spans several commits and must be net: %+v", i, st)
		}
		if st.Index != i {
			t.Errorf("sampled step %d has Index %d", i, st.Index)
		}
	}
	if sampled[0].From.Hash != "c0" || sampled[1].To.Hash != "c9" {
		t.Errorf("sampled steps run %s to %s, want c0 to c9", sampled[0].From.Hash, sampled[1].To.Hash)
	}

	// Asking for as many samples as there are commits is the exhaustive walk.
	if got := w.Steps(10); len(got) != 9 || got[0].Net {
		t.Errorf("Steps(10) = %d steps, first net=%v; want the exhaustive 9", len(got), got[0].Net)
	}
}
