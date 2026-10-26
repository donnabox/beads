package legacyimport

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestOrdinaryExportVocabulary(t *testing.T) {
	batch, err := Parse(strings.NewReader(`{"_type":"issue","id":"old-a","title":"Work","priority":0,"dependency_count":0,"dependent_count":0,"comment_count":1,"metadata":{"custom":[1,true]},"comments":[{"id":7,"issue_id":"old-a","author":"a","text":"note","created_at":"2026-01-01T00:00:00Z"}]}
{"_type":"memory","key":"context","value":"remember this"}
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Issues) != 1 || len(batch.Memories) != 1 || batch.Issues[0].Comments[0].ID != "7" || batch.Issues[0].Priority != 0 {
		t.Fatalf("batch=%+v", batch)
	}
}

func TestLegacyInputRefusals(t *testing.T) {
	for _, input := range []string{
		`{"id":"x","id":"y"}`, `{"metadata":{"x":1,"x":2}}`,
		`{"title":"\ud800"}`, `{"\ud800":"x"}`, `{"comments":[{"id":1.5}]}`, `{"comments":[{"id":null}]}`,
		`{"_schema":"future"}`,
		`{"title":"original","Title":"replacement"}`, `{"_type":"memory","key":"x","value":"keep","Value":"drop"}`,
		`{"dependencies":[{"type":"blocks","Type":"parent-child"}]}`, `{"bonded_from":[{"source_id":"x","future":"lost"}]}`,
		`{"bonded_from":[{"source_id":"a","proto_id":"b","bond_type":"parallel"}]}`,
		`{"priority":null}`, `{"no_history":null}`, `{"_type":null,"id":"x"}`, `{"dependent_count":null}`, `{"wisp":null}`,
		`{"_type":"memory","key":"x"}`, `{"_type":"memory","key":"x","value":null}`,
		`null`, `[]`, `{"title":"first"} trailing`, `{"id":"x","title":"x","history":[]}`,
		`{"id":"https://example.invalid/beads/x","type":"types/issue","properties":{"title":"x"}}`,
		`{"_type":"graph","id":"x"}`, `{"_type":"memory","key":"x","value":"v","future":1}`,
		`{"title":"x","wisp":true}`, `{"title":"x","wisp_plane":true}`,
		`{"title":"x","dependency_count":-1}`, `{"title":"x","comments":[{"text":"x","history":[]}]}`,
		`{"_schema":1,"title":"lost"}`, string([]byte{'{', '"', 'x', '"', ':', '"', 255, '"', '}'}),
	} {
		t.Run(input, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(input)); err == nil {
				t.Fatal("accepted unsupported input")
			}
		})
	}
	if _, err := Parse(strings.NewReader(strings.Repeat(" ", MaxBytes+1))); err == nil {
		t.Fatal("accepted oversized stream")
	}
}

func TestHistoricalHeaderAndExactCommentID(t *testing.T) {
	input := `{"_dolt_branch":"main","_dolt_commit":"abc123","_project_id":"p1","_schema":"beads-jsonl/1","_sort":"stable-v1"}
{"id":"old-a","title":"Work","comments":[{"id":9223372036854775807,"text":"note"}]}`
	batch, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if batch.Issues[0].Comments[0].ID != "9223372036854775807" {
		t.Fatal("numeric comment ID changed")
	}
}

func TestHistoricalBondRef(t *testing.T) {
	batch, err := Parse(strings.NewReader(`{"id":"old-a","title":"Work","bonded_from":[{"proto_id":"old-source","bond_type":"parallel"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if batch.Issues[0].BondedFrom[0].SourceID != "old-source" {
		t.Fatal("historical BondRef lost")
	}
}

func TestIgnoreUnknownFieldsOnlyExplicitly(t *testing.T) {
	input := `{"id":"old-a","title":"Keep","future":{"secret":"not a diagnostic"},"metadata":{"properties":{"future":1},"_schema":"user data"},"comments":[{"id":7,"text":"one","future":true},{"id":8,"text":"two","future":false}],"dependencies":[{"issue_id":"old-a","depends_on_id":"old-b","type":"blocks","future":0}]}
{"_type":"memory","key":"notes/plan","value":"Exact body","future":null}`
	if _, err := Parse(strings.NewReader(input)); err == nil {
		t.Fatal("strict Parse discarded fields")
	}
	if _, _, err := ParseWithOptions(strings.NewReader(input), Options{}); err == nil {
		t.Fatal("default options discarded fields")
	}
	batch, report, err := ParseWithOptions(strings.NewReader(input), Options{IgnoreUnknownFields: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Issues) != 1 || len(batch.Memories) != 1 || batch.Issues[0].Title != "Keep" || batch.Memories[0].Value != "Exact body" || batch.Issues[0].Comments[0].ID != "7" || len(batch.Issues[0].Dependencies) != 1 {
		t.Fatalf("known data changed: %+v", batch)
	}
	if got := string(batch.Issues[0].Metadata); got != `{"properties":{"future":1},"_schema":"user data"}` {
		t.Fatalf("freeform metadata was pruned or rewritten: %s", got)
	}
	want := []IgnoredField{
		{Path: `["issue","comments","[]","future"]`, Count: 2},
		{Path: `["issue","dependencies","[]","future"]`, Count: 1},
		{Path: `["issue","future"]`, Count: 1},
		{Path: `["memory","future"]`, Count: 1},
	}
	if !reflect.DeepEqual(report.Fields, want) || report.Total != 5 || report.OmittedDistinct != 0 || report.OmittedOccurrences != 0 {
		t.Fatalf("report=%+v want %+v", report, want)
	}
}

func TestIgnoreUnknownFieldsPreservesKnownRefusals(t *testing.T) {
	for _, input := range []string{
		`{"title":"x","priority":"bad","future":1}`, `{"title":null,"future":1}`,
		`{"title":"x","no_history":null,"future":1}`, `{"title":"x","created_at":"bad","future":1}`,
		`{"title":"x","comments":[{"id":1.5,"future":1}]}`,
		`{"title":"x","comments":[{"id":null,"future":1}]}`,
		`{"title":"x","dependencies":[{"type":null,"future":1}]}`,
		`{"_type":"memory","key":"x","future":1}`, `{"_type":"memory","key":"x","value":null,"future":1}`,
		`{"title":"x","wisp":true,"future":1}`, `{"title":"x","wisp_plane":true,"future":1}`,
		`{"title":"x","dependency_count":-1,"future":1}`,
		`{"title":"x","bonded_from":[{"source_id":"a","proto_id":"b","future":1}]}`,
		`{"title":"x","Title":"wrong case","future":1}`,
		`{"_type":"memory","key":"x","value":"v","Value":"wrong case"}`,
		`{"dependencies":[{"type":"blocks","Type":"wrong case"}]}`,
		`{"title":"x","Wisp":true}`, `{"title":"x","Dependency_count":-1}`,
	} {
		t.Run(input, func(t *testing.T) {
			batch, report, err := ParseWithOptions(strings.NewReader(input), Options{IgnoreUnknownFields: true})
			if err == nil || len(batch.Issues) != 0 || len(batch.Memories) != 0 || report.Total != 0 || len(report.Fields) != 0 {
				t.Fatalf("lost refusal or exposed partial data: batch=%+v report=%+v err=%v", batch, report, err)
			}
		})
	}
	// These known values are admitted by parsing and still reach the existing
	// converter's policy/freshness gate unchanged. The option grants no waiver.
	batch, _, err := ParseWithOptions(strings.NewReader(`{"id":"same","title":"x","ephemeral":true,"no_history":true,"storage_class":"versioned","source_formula":"keep","bonded_from":[{"source_id":"old","bond_type":"parallel","future":1}],"lease_granted_node":"node","dependencies":[{"issue_id":"same","depends_on_id":"peer","type":"blocks","metadata":"{\"key\":1}","thread_id":"keep","future":1}],"future":1}
{"id":"same","title":"collision","future":1}`), Options{IgnoreUnknownFields: true})
	if err != nil {
		t.Fatal(err)
	}
	issue := batch.Issues[0]
	if len(batch.Issues) != 2 || batch.Issues[1].ID != issue.ID || !issue.Ephemeral || !issue.NoHistory || issue.StorageClass != "versioned" || issue.SourceFormula != "keep" || len(issue.BondedFrom) != 1 || issue.LeaseGrantedNode != "node" || issue.Dependencies[0].Metadata != `{"key":1}` || issue.Dependencies[0].ThreadID != "keep" {
		t.Fatalf("known unsupported values/collision were filtered: %+v", issue)
	}
}

func TestIgnoreUnknownFieldsNeverErasesGraphOrHeader(t *testing.T) {
	for _, input := range []string{
		`{"_schema":"beads-jsonl/1","_future":"provenance"}`,
		`{"_schema":"future","future":1}`, `{"_schema":"beads-jsonl/1","_project_id":null}`,
		`{"_schema":"beads-jsonl/1","id":"old-a","title":"not a header"}`,
		`{"_type":"resource","id":"old-a","title":"not an issue"}`, `{"_type":null}`,
		`{"id":"old-a","title":"x","type":"types/memory","properties":{}}`,
		`{"id":"old-a","title":"x","revision":"1"}`, `{"id":"old-a","title":"x","owned":[]}`,
		`{"id":"old-a","title":"x","source":"old-b","target":"old-c"}`,
		`{"id":"old-a","title":"x","attribution":{}}`, `{"id":"old-a","title":"x","history":[]}`,
		`{"id":"old-a","title":"x","resources":[]}`, `{"id":"old-a","title":"x","manifest":{}}`,
		`{"_type":"memory","key":"note","value":"v","_provenance":{}}`,
		`{"comments":[{"id":"c","history":[]}]}`,
	} {
		t.Run(input, func(t *testing.T) {
			if _, _, err := ParseWithOptions(strings.NewReader(input), Options{IgnoreUnknownFields: true}); err == nil {
				t.Fatal("reinterpreted graph data/provenance")
			}
		})
	}
}

func TestIgnoreUnknownFieldsValidatesDiscardedSubtrees(t *testing.T) {
	inputs := []string{
		`{"title":"x","future":{"key":1,"key":2}}`,
		`{"title":"x","future":"\ud800"}`, `{"title":"x","\ud800":"unknown name"}`,
		`{"title":"x","future":invalid}`, `{"title":"x","future":1} trailing`,
		`{"title":"x","future":` + strings.Repeat("[", 130) + `0` + strings.Repeat("]", 130) + `}`,
		string([]byte{'{', '"', 'x', '"', ':', '"', 255, '"', '}'}),
	}
	for i, input := range inputs {
		if _, _, err := ParseWithOptions(strings.NewReader(input), Options{IgnoreUnknownFields: true}); err == nil {
			t.Errorf("input %d discarded malformed JSON", i)
		}
	}
}

func TestIgnoredFieldReportIsBoundedAndSafe(t *testing.T) {
	values := map[string]any{"id": "private-record-id", "title": "private-title"}
	for i := 0; i < 40; i++ {
		values[fmt.Sprintf("field%02d", i)] = "secret-value"
	}
	values["https://example.invalid/field\n\x1b[31m"] = "private-uri-value"
	values[strings.Repeat("é", 100)] = 1
	raw, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	input := string(raw) + "\n" + string(raw)
	_, report, err := ParseWithOptions(strings.NewReader(input), Options{IgnoreUnknownFields: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Fields) != maxReportedPaths || report.Total != 84 || report.Distinct != 42 || report.OmittedDistinct != 10 || report.OmittedOccurrences != 20 {
		t.Fatalf("incorrect bounded counts: %+v", report)
	}
	for _, field := range report.Fields {
		if !json.Valid([]byte(field.Path)) || strings.ContainsAny(field.Path, "\n\x1b") || len(field.Path) > 600 || field.Count != 2 {
			t.Fatalf("unsafe path: %+v", field)
		}
	}
	// Names remain inspectable, even URI-looking ones. Unicode truncation stays
	// valid and explicit; no report ever contains a source value or record ID.
	for _, name := range []string{"https://example.invalid/field\n\x1b[31m", strings.Repeat("é", 100)} {
		raw, _ := json.Marshal(map[string]any{"title": "private-title", name: "secret-value"})
		_, one, err := ParseWithOptions(strings.NewReader(string(raw)), Options{IgnoreUnknownFields: true})
		if err != nil || len(one.Fields) != 1 {
			t.Fatalf("name report: %+v (%v)", one, err)
		}
		field := one.Fields[0]
		if !json.Valid([]byte(field.Path)) || strings.ContainsAny(field.Path, "\n\x1b") || strings.Contains(field.Path, "secret-value") || strings.Contains(field.Path, "private-title") || field.Truncated != (len(name) > maxPathComponentBytes) {
			t.Fatalf("unsafe field: %+v", field)
		}
	}
	_, again, err := ParseWithOptions(strings.NewReader(input), Options{IgnoreUnknownFields: true})
	if err != nil || !reflect.DeepEqual(again, report) {
		t.Fatal("report order or counts are nondeterministic")
	}
}

// Detail/list projections are not complete ordinary exports. Even a false
// omission marker must not be erased to reinterpret that record family.
func TestIgnoreUnknownFieldsRefusesIncompleteProjection(t *testing.T) {
	for _, marker := range []string{"true", "false", "null", `"invalid"`} {
		for _, ignore := range []bool{false, true} {
			input := `{"id":"old-a","title":"projection","comments_omitted":` + marker + `,"future":1}`
			batch, report, err := ParseWithOptions(strings.NewReader(input), Options{IgnoreUnknownFields: ignore})
			if err == nil || len(batch.Issues) != 0 || report.Total != 0 {
				t.Fatalf("projection admitted: marker=%s ignore=%t batch=%+v report=%+v err=%v", marker, ignore, batch, report, err)
			}
		}
	}
}

func TestIgnoreUnknownFieldsRefusesActualGraphEnvelopes(t *testing.T) {
	for _, envelope := range []string{
		`{"schemaVersion":1,"preview":true,"result":{"id":"x"}}`,
		`{"id":"old-a","title":"not an export","schemaVersion":1,"preview":true,"result":{}}`,
		`{"code":"invalid_properties","message":"not an export","retryable":false}`,
	} {
		if _, _, err := ParseWithOptions(strings.NewReader(envelope), Options{IgnoreUnknownFields: true}); err == nil {
			t.Fatal("discarded an actual graph output envelope")
		}
	}
	for _, name := range []string{"schemaVersion", "preview", "result", "code", "retryable"} {
		raw, _ := json.Marshal(map[string]any{"id": "old-a", "title": "not an export", name: nil})
		if _, _, err := ParseWithOptions(strings.NewReader(string(raw)), Options{IgnoreUnknownFields: true}); err == nil {
			t.Fatalf("discarded graph envelope field %s", name)
		}
	}
}

func TestIgnoredFieldPathsAreTerminalSafeJSON(t *testing.T) {
	for _, name := range []string{"c1\u009b2K\u009dtitle", "del\x7f", "bidi\u202eoverride", "é😀", `literal\u009b\quote"`} {
		raw, _ := json.Marshal(map[string]any{"title": "x", name: "private-value"})
		_, report, err := ParseWithOptions(strings.NewReader(string(raw)), Options{IgnoreUnknownFields: true})
		if err != nil || len(report.Fields) != 1 {
			t.Fatalf("name %q report=%+v err=%v", name, report, err)
		}
		path := report.Fields[0].Path
		for _, r := range path {
			if r < 0x20 || r >= 0x7f {
				t.Fatalf("terminal control/non-ASCII rune %U in path %q", r, path)
			}
		}
		var decoded []string
		if err := json.Unmarshal([]byte(path), &decoded); err != nil || !reflect.DeepEqual(decoded, []string{"issue", name}) {
			t.Fatalf("path is double-encoded or changed: %q -> %q (%v)", path, decoded, err)
		}
	}
}

func TestIgnoredFieldTruncatedPathsRemainDistinct(t *testing.T) {
	prefix := strings.Repeat("x", 64)
	raw, _ := json.Marshal(map[string]any{"title": "x", prefix + "one": true, prefix + "two": false})
	_, report, err := ParseWithOptions(strings.NewReader(string(raw)), Options{IgnoreUnknownFields: true})
	if err != nil || report.Total != 2 || report.Distinct != 2 || len(report.Fields) != 2 || report.OmittedDistinct != 0 || report.OmittedOccurrences != 0 {
		t.Fatalf("truncation lost counts: %+v (%v)", report, err)
	}
	one, two := report.Fields[0], report.Fields[1]
	if !one.Truncated || !two.Truncated || one.Path != two.Path || one.PathDigest == two.PathDigest || len(one.PathDigest) != 16 || len(two.PathDigest) != 16 || one.Count != 1 || two.Count != 1 {
		t.Fatalf("truncated original paths indistinguishable: %+v", report)
	}
	_, repeat, err := ParseWithOptions(strings.NewReader(string(raw)), Options{IgnoreUnknownFields: true})
	if err != nil || !reflect.DeepEqual(repeat, report) {
		t.Fatal("full-path digest is not stable")
	}
}

func TestIgnoreUnknownFieldsKnownCaseDiagnostic(t *testing.T) {
	_, _, err := ParseWithOptions(strings.NewReader(`{"id":"old-a","title":"x","Target":"native-event-target"}`), Options{IgnoreUnknownFields: true})
	if err == nil || !strings.Contains(err.Error(), "case-sensitive") || strings.Contains(err.Error(), "graph or provenance") {
		t.Fatalf("known native case alias misdiagnosed: %v", err)
	}
}
