package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	storageissueops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// versionableIssuesCheckName is the name a person sees for this check in doctor's
// output and in its JSON, so it is pinned here rather than exported.
const versionableIssuesCheckName = "Versionable Issues"

// maxExactNanoseconds is the last whole nanosecond count a version can record
// (2^53-1, about 104.25 days), the bound a gate's timeout meets because it is held
// as nanoseconds.
const maxExactNanoseconds = 1<<53 - 1

// scriptedReader answers List with the rows a test names and records the request,
// so the check can be run over exactly those rows. Ready and Get are not part of
// what the check reads, and the embedded nil Reader makes either one fail loudly if
// the check ever reaches for it.
type scriptedReader struct {
	issueops.Reader
	rows []*types.Issue
	err  error

	calls int
	req   issueops.ListRequest
}

func (r *scriptedReader) List(_ context.Context, req issueops.ListRequest) (issueops.IssuePage, error) {
	r.calls++
	r.req = req
	if r.err != nil {
		return issueops.IssuePage{}, r.err
	}
	page := issueops.IssuePage{Items: []*types.IssueWithCounts{}}
	for _, row := range r.rows {
		page.Items = append(page.Items, &types.IssueWithCounts{Issue: row})
	}
	return page, nil
}

func plainIssue(id string) *types.Issue {
	return &types.Issue{ID: id, Title: id, Status: types.StatusOpen, IssueType: types.TypeTask}
}

func issueWithMetadata(id, metadata string) *types.Issue {
	issue := plainIssue(id)
	issue.Metadata = json.RawMessage(metadata)
	return issue
}

func gateWithTimeout(id string, timeout time.Duration) *types.Issue {
	issue := plainIssue(id)
	issue.IssueType = types.TypeGate
	issue.AwaitType = "timer"
	issue.Timeout = timeout
	return issue
}

// offenderIDs lists the ids a refusing check's Detail names, in order. Each entry is
// "<id>: <refusal>", entries are separated by "; ", and a trailing "(+N more)" note
// belongs to the last entry and is not one of its own.
func offenderIDs(detail string) []string {
	var ids []string
	for _, entry := range strings.Split(detail, "; ") {
		if id, _, ok := strings.Cut(entry, ": "); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

type versionableCase struct {
	name    string
	issue   *types.Issue
	refused bool
	field   string // the top-level field the refusal names, when refused
	literal string // the number the refusal names, when refused
}

// versionableCorpus is the whole-issue domain the check and the mint must agree on.
// Metadata covers a number past the range, an ordinary document, null and nothing at
// all. The other field that can reach the range is a gate's timeout, a duration held
// as nanoseconds: 2500h is admitted and 2600h is not, a negative one is refused by the
// same bound, and the boundary itself is exact. There is no duplicate-key case: the
// metadata column collapses duplicate keys on write, so a stored row cannot hold one.
func versionableCorpus() []versionableCase {
	timeoutCase := func(name string, timeout time.Duration, refused bool) versionableCase {
		tc := versionableCase{name: name, issue: gateWithTimeout("gate-1", timeout), refused: refused}
		if refused {
			tc.field = "timeout"
			tc.literal = strconv.FormatInt(int64(timeout), 10)
		}
		return tc
	}
	return []versionableCase{
		{name: "metadata number past the range", issue: issueWithMetadata("md-1", `{"ts":1727000000000000000}`),
			refused: true, field: "metadata", literal: "1727000000000000000"},
		{name: "metadata that is clean", issue: issueWithMetadata("md-2", `{"ok":1,"f":0.1,"s":"x"}`)},
		{name: "metadata that is null", issue: issueWithMetadata("md-3", `null`)},
		{name: "metadata that is empty", issue: issueWithMetadata("md-4", ``)},

		timeoutCase("no timeout", 0, false),
		timeoutCase("timeout 2500h", 2500*time.Hour, false),
		timeoutCase("timeout 2600h", 2600*time.Hour, true),
		timeoutCase("timeout 8760h", 8760*time.Hour, true),
		timeoutCase("timeout -2600h", -2600*time.Hour, true),
		timeoutCase("timeout of exactly 2^53-1 ns", time.Duration(maxExactNanoseconds), false),
		timeoutCase("timeout of exactly 2^53 ns", time.Duration(maxExactNanoseconds+1), true),
	}
}

// The check runs the one function recording runs, over whole issues, so for every
// shape of issue what it reports and what the mint would refuse cannot differ. The
// corpus is first held against CheckIssueVersionable itself, so a wrong expectation
// fails here and not as a disagreement; then the check is run over the issue and must
// name it, with the refusal the shared function gives, the field and the number.
func TestCheckVersionableIssues_AgreesWithTheMintOnWholeIssues(t *testing.T) {
	for _, tc := range versionableCorpus() {
		t.Run(tc.name, func(t *testing.T) {
			shared := storageissueops.CheckIssueVersionable(tc.issue)
			if (shared != nil) != tc.refused {
				t.Fatalf("CheckIssueVersionable = %v, but the corpus says refused=%t", shared, tc.refused)
			}

			check := checkVersionableIssues(t.Context(), &scriptedReader{rows: []*types.Issue{tc.issue}})

			if check.Name != versionableIssuesCheckName {
				t.Errorf("check name = %q, want %q", check.Name, versionableIssuesCheckName)
			}
			if got := check.Status != StatusOK; got != tc.refused {
				t.Fatalf("status %q (%s): reads as refused=%t, the mint's own check says refused=%t", check.Status, check.Message, got, tc.refused)
			}
			if !tc.refused {
				return
			}
			if check.Status != StatusWarning {
				t.Errorf("status = %q, want %q", check.Status, StatusWarning)
			}
			if ids := offenderIDs(check.Detail); len(ids) != 1 || ids[0] != tc.issue.ID {
				t.Errorf("Detail names %v, want exactly [%s]: %q", ids, tc.issue.ID, check.Detail)
			}
			if !strings.Contains(check.Detail, shared.Error()) {
				t.Errorf("Detail %q does not carry the refusal the shared check returns, %q", check.Detail, shared.Error())
			}
			if want := fmt.Sprintf("(field %q)", tc.field); !strings.Contains(check.Detail, want) {
				t.Errorf("Detail %q does not name the field as %s", check.Detail, want)
			}
			if !strings.Contains(check.Detail, tc.literal) {
				t.Errorf("Detail %q does not name the offending number %s", check.Detail, tc.literal)
			}
		})
	}
}

// A store with a gate whose timeout is 2600h must not read clean, whatever state
// versioned history is in: the check never asks, because the rows are what is wrong.
// A check that looked at metadata alone would pass this store and the first write to
// the gate would be refused once history is on.
func TestCheckVersionableIssues_AGateTimeoutOfTwentySixHundredHoursDoesNotReadClean(t *testing.T) {
	reader := &scriptedReader{rows: []*types.Issue{
		issueWithMetadata("md-1", `{"ok":1}`),
		gateWithTimeout("gate-2600h", 2600*time.Hour),
	}}

	check := checkVersionableIssues(t.Context(), reader)

	if check.Status == StatusOK {
		t.Fatalf("a store holding a 2600h gate read clean: %s", check.Message)
	}
	if ids := offenderIDs(check.Detail); len(ids) != 1 || ids[0] != "gate-2600h" {
		t.Fatalf("Detail names %v, want exactly [gate-2600h]: %q", ids, check.Detail)
	}
	if !strings.Contains(check.Detail, `(field "timeout")`) {
		t.Errorf("Detail does not say the gate's timeout is the field: %q", check.Detail)
	}
}

// Ephemeral rows and no-history rows are never versioned, so a value in one can never
// abort a write and is no reason to call a store unfit. The skip is the mint's own
// rule and lives in the scan, so this only holds the check to it: beside a durable
// offender, the rows recording skips are not named, and by themselves they leave the
// store clean.
func TestCheckVersionableIssues_SkipsWhatTheMintNeverVersions(t *testing.T) {
	const oversize = `{"ts":1727000000000000000}`
	wisp := issueWithMetadata("wisp-1", oversize)
	wisp.Ephemeral = true
	noHistory := issueWithMetadata("nohistory-1", oversize)
	noHistory.NoHistory = true
	wispGate := gateWithTimeout("wisp-gate-1", 2600*time.Hour)
	wispGate.Ephemeral = true
	durable := issueWithMetadata("durable-1", oversize)

	alone := checkVersionableIssues(t.Context(), &scriptedReader{rows: []*types.Issue{wisp, noHistory, wispGate}})
	if alone.Status != StatusOK {
		t.Errorf("rows the mint never versions made the store read %q (%s): %q", alone.Status, alone.Message, alone.Detail)
	}

	beside := checkVersionableIssues(t.Context(), &scriptedReader{rows: []*types.Issue{wisp, durable, noHistory, wispGate}})
	if ids := offenderIDs(beside.Detail); len(ids) != 1 || ids[0] != "durable-1" {
		t.Errorf("Detail names %v, want exactly [durable-1]: %q", ids, beside.Detail)
	}
	if !strings.HasPrefix(beside.Message, "1 issue holds") {
		t.Errorf("Message %q does not count one issue", beside.Message)
	}
}

// The report leads with the whole count, names the first twenty in the order the rows
// came, and says how many it left out; a clean row between offenders shifts neither
// the order nor the count. One offender reads as one.
func TestCheckVersionableIssues_NamesTheCountAndTheFirstTwentyInOrder(t *testing.T) {
	rows := []*types.Issue{plainIssue("clean-1")}
	var want []string
	for i := 1; i <= 25; i++ {
		id := fmt.Sprintf("t-%03d", i)
		rows = append(rows, issueWithMetadata(id, `{"ts":1727000000000000000}`))
		if i <= 20 {
			want = append(want, id)
		}
	}

	check := checkVersionableIssues(t.Context(), &scriptedReader{rows: rows})

	if !strings.HasPrefix(check.Message, "25 issues hold") {
		t.Errorf("Message %q does not lead with the whole count of 25", check.Message)
	}
	got := offenderIDs(check.Detail)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Detail names %v, want the first twenty in order %v", got, want)
	}
	if strings.Contains(check.Detail, "t-021") {
		t.Errorf("Detail names an issue past the first twenty: %q", check.Detail)
	}
	if !strings.Contains(check.Detail, "(+5 more)") {
		t.Errorf("Detail does not say five were left out: %q", check.Detail)
	}

	one := checkVersionableIssues(t.Context(), &scriptedReader{rows: rows[:2]})
	if !strings.HasPrefix(one.Message, "1 issue holds") {
		t.Errorf("one offender reads %q, want \"1 issue holds ...\"", one.Message)
	}
	if strings.Contains(one.Detail, "more)") {
		t.Errorf("one offender carries a left-out note: %q", one.Detail)
	}
}

// What fixes a row depends on the field. A number in metadata is replaced or removed
// with bd update, which works with history on or off. A gate's timeout is outside
// bd update's allow-list, so the gate is removed or its column is set with bd sql
// while history is off, and bd sql needs a server-backed store. Each kind of field
// gets its own remedy and a store that has only one kind is not told about the
// other. The remedy never steers at bd doctor --fix: nothing here is repaired by it.
func TestCheckVersionableIssues_TheRemedyIsPerField(t *testing.T) {
	metadataRemedy := []string{
		"bd update <id> --metadata",
		"bd update <id> --unset-metadata <key>",
	}
	timeoutRemedy := []string{
		"9007199254740991",
		"bd update cannot change a gate's timeout",
		"bd delete <id> --force",
		`bd sql "UPDATE issues SET timeout_ns = <nanoseconds> WHERE id = '<id>'"`,
		"while history is off",
		"server-backed",
	}
	metadataRow := issueWithMetadata("md-1", `{"ts":1727000000000000000}`)
	timeoutRow := gateWithTimeout("gate-1", 2600*time.Hour)

	for _, tc := range []struct {
		name    string
		rows    []*types.Issue
		want    []string
		without []string
	}{
		{"metadata only", []*types.Issue{metadataRow}, metadataRemedy, []string{"bd sql", "bd delete", "timeout_ns"}},
		{"timeout only", []*types.Issue{timeoutRow}, timeoutRemedy, []string{"--unset-metadata", "--metadata"}},
		{"both", []*types.Issue{metadataRow, timeoutRow}, append(append([]string{}, metadataRemedy...), timeoutRemedy...), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			check := checkVersionableIssues(t.Context(), &scriptedReader{rows: tc.rows})
			if check.Status == StatusOK {
				t.Fatalf("store with %s offenders read clean", tc.name)
			}
			for _, want := range tc.want {
				if !strings.Contains(check.Fix, want) {
					t.Errorf("Fix does not contain %q:\n%s", want, check.Fix)
				}
			}
			for _, unwanted := range tc.without {
				if strings.Contains(check.Fix, unwanted) {
					t.Errorf("Fix names %q, which does not apply to a store with %s offenders:\n%s", unwanted, tc.name, check.Fix)
				}
			}
			if MentionsFixAdvice(check.Fix) {
				t.Errorf("Fix steers at bd doctor --fix, which repairs nothing here:\n%s", check.Fix)
			}
		})
	}
}

// A store with nothing the mint would refuse passes, with no remedy to offer, and
// says how many rows it read so that a pass over an empty read is not mistaken for a
// pass over a store.
func TestCheckVersionableIssues_ACleanStorePasses(t *testing.T) {
	for _, tc := range []struct {
		name string
		rows []*types.Issue
		read string
	}{
		{"no rows", nil, "0 issues read"},
		{"clean rows", []*types.Issue{plainIssue("a-1"), issueWithMetadata("a-2", `{"ok":1}`), gateWithTimeout("a-3", 2500*time.Hour)}, "3 issues read"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			check := checkVersionableIssues(t.Context(), &scriptedReader{rows: tc.rows})
			if check.Status != StatusOK {
				t.Fatalf("status %q (%s), want %q", check.Status, check.Message, StatusOK)
			}
			if check.Name != versionableIssuesCheckName {
				t.Errorf("check name = %q, want %q", check.Name, versionableIssuesCheckName)
			}
			if check.Fix != "" {
				t.Errorf("a clean store carries a remedy: %q", check.Fix)
			}
			if !strings.Contains(check.Detail, tc.read) {
				t.Errorf("Detail %q does not say %q", check.Detail, tc.read)
			}
		})
	}
}

// The check reads the way the switch that turns history on reads, because a store the
// check passes must be a store that switch passes: one unbounded read of every status,
// every type (gates are a type a default listing hides, and the timeout is the field
// that case is about) and both planes, in the brief projection with no labels and no
// counts. A page is not a store, so the read has no limit and no row cap.
func TestCheckVersionableIssues_ReadsEveryRowInOneUnboundedRead(t *testing.T) {
	reader := &scriptedReader{rows: []*types.Issue{plainIssue("a-1")}}

	checkVersionableIssues(t.Context(), reader)

	if reader.calls != 1 {
		t.Fatalf("the check read %d times, want one read", reader.calls)
	}
	req := reader.req
	for name, set := range map[string]bool{
		"AllFlag":         req.AllFlag,
		"IncludeAllTypes": req.IncludeAllTypes,
		"Brief":           req.Brief,
		"SkipLabels":      req.SkipLabels,
		"SkipCounts":      req.SkipCounts,
	} {
		if !set {
			t.Errorf("request does not set %s", name)
		}
	}
	if req.Limit == nil || *req.Limit != 0 {
		t.Errorf("request limit = %v, want a pointer to 0 (unlimited)", req.Limit)
	}
	if req.MaxRows != 0 {
		t.Errorf("request caps the read at %d rows, want no cap", req.MaxRows)
	}
}

// A read that fails is not a clean store: a check that could not run has not passed.
func TestCheckVersionableIssues_AReadThatFailsDoesNotReadClean(t *testing.T) {
	reader := &scriptedReader{err: errors.New("connection refused")}

	check := checkVersionableIssues(t.Context(), reader)

	if check.Status == StatusOK {
		t.Fatalf("a read that failed read clean: %s", check.Message)
	}
	if check.Status != StatusWarning {
		t.Errorf("status = %q, want %q", check.Status, StatusWarning)
	}
	if !strings.Contains(check.Detail, "connection refused") {
		t.Errorf("Detail does not carry the read's error: %q", check.Detail)
	}
	if !strings.Contains(strings.ToLower(check.Message), "unable") {
		t.Errorf("Message %q does not say the check could not run", check.Message)
	}
}

// With no database there is nothing to check, which every store-reading check here
// answers as a pass; the message says so rather than claiming a store was inspected.
func TestCheckVersionableIssuesWithStore_NoDatabaseYet(t *testing.T) {
	for name, ss := range map[string]*SharedStore{"no shared store": nil, "no store in it": {}} {
		t.Run(name, func(t *testing.T) {
			check := CheckVersionableIssuesWithStore(ss)
			if check.Status != StatusOK {
				t.Fatalf("status %q (%s), want %q", check.Status, check.Message, StatusOK)
			}
			if check.Name != versionableIssuesCheckName {
				t.Errorf("check name = %q, want %q", check.Name, versionableIssuesCheckName)
			}
			if !strings.Contains(strings.ToLower(check.Message), "no database") {
				t.Errorf("Message %q does not say there is no database", check.Message)
			}
		})
	}
}

// There is ONE check over a whole issue, CheckIssueVersionable, and the scan calls it.
// A doctor check that inspected metadata or a timeout itself would be a second one, can
// pass what the mint refuses, and is what the issueops package already forbids beside
// the shared function. So: this package exports nothing named for versionability
// except the check, and the check's own file runs the scan and reads no field of an
// issue to judge it.
func TestTheVersionableIssuesCheckRunsTheSharedScanAndInspectsNoFieldItself(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse doctor package: %v", err)
	}
	allowed := map[string]bool{"CheckVersionableIssuesWithStore": true}
	seen := map[string]bool{}
	var stray []string
	var checkFile *ast.File
	for _, pkg := range pkgs {
		for name, file := range pkg.Files {
			if filepath.Base(name) == "versionable_issues.go" {
				checkFile = file
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil || !fn.Name.IsExported() {
					continue
				}
				if !strings.Contains(strings.ToLower(fn.Name.Name), "versionable") {
					continue
				}
				seen[fn.Name.Name] = true
				if !allowed[fn.Name.Name] {
					stray = append(stray, fmt.Sprintf("%s (%s)", fn.Name.Name, fset.Position(fn.Pos())))
				}
			}
		}
	}
	sort.Strings(stray)
	if len(stray) > 0 {
		t.Errorf("an exported function named for versionability exists beside the check: %s", strings.Join(stray, ", "))
	}
	for name := range allowed {
		if !seen[name] {
			t.Errorf("%s was not found: this guard no longer sees the check it protects", name)
		}
	}

	if checkFile == nil {
		t.Fatal("versionable_issues.go was not found: this guard no longer sees the file it protects")
	}
	var callsScan bool
	var inspected []string
	ast.Inspect(checkFile, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "FindUnversionable" {
				callsScan = true
			}
		case *ast.SelectorExpr:
			if x.Sel.Name == "Metadata" || x.Sel.Name == "Timeout" {
				inspected = append(inspected, fset.Position(x.Pos()).String())
			}
		}
		return true
	})
	if !callsScan {
		t.Error("versionable_issues.go does not call FindUnversionable: the check must run the shared scan, not a copy of it")
	}
	if len(inspected) > 0 {
		t.Errorf("versionable_issues.go reads a field of an issue to judge it, which is a check over part of the issue: %s", strings.Join(inspected, ", "))
	}
}
