package main

import (
	"strings"
	"testing"
)

func res(exit int, stdout, stderr string) *stepResult {
	return &stepResult{Name: "cur", Exit: exit, Stdout: []byte(stdout), Stderr: []byte(stderr)}
}

func mustAssertion(t *testing.T, js string) *assertion {
	t.Helper()
	a, err := parseAssertion([]byte(js))
	if err != nil {
		t.Fatalf("parseAssertion(%s): %v", js, err)
	}
	return a
}

func TestStderrCode(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   string
	}{
		{"json refusal", `{"code":"revision_conflict","message":"x","retryable":false}` + "\n", "revision_conflict"},
		{"plain text refusal", "gone: graph resource was deleted\n", "gone"},
		{"plain text with digits", "revision_unknown: graph preview retained version is unknown", "revision_unknown"},
		{"first line only", "not_found: graph resource not found\nsecond line: ignored\n", "not_found"},
		{"json without a code", `{"message":"x"}`, ""},
		{"json code that is not a string", `{"code":4}`, ""},
		{"untyped error", "Error: something broke\n", ""},
		{"prose that merely contains a colon", "cannot open: no such file\n", ""}, // "cannot open" has a space: not a code token
		{"empty", "", ""},
	}
	for _, tc := range tests {
		if got := stderrCode([]byte(tc.stderr)); got != tc.want {
			t.Errorf("%s: stderrCode(%q) = %q, want %q", tc.name, tc.stderr, got, tc.want)
		}
	}
}

func TestParseAssertionRejects(t *testing.T) {
	tests := []struct {
		name string
		js   string
		want string
	}{
		{"unknown op", `{"op":"exec"}`, "unknown op"},
		{"missing op", `{"equals":0}`, "op"},
		{"unknown field", `{"op":"exit","equals":0,"extra":1}`, "unknown field"},
		{"exit with both forms", `{"op":"exit","equals":0,"nonzero":true}`, "exit"},
		{"exit with no form", `{"op":"exit"}`, "exit"},
		{"exit out of range", `{"op":"exit","equals":300}`, "exit"},
		{"exit nonzero false", `{"op":"exit","nonzero":false}`, "exit"},
		{"stderr_code with both forms", `{"op":"stderr_code","equals":"a","not_equals":"b"}`, "stderr_code"},
		{"stderr_code with empty code", `{"op":"stderr_code","equals":""}`, "stderr_code"},
		{"stdout with no form", `{"op":"stdout"}`, "stdout"},
		{"stdout with two forms", `{"op":"stdout","empty":true,"equals_input":"B1"}`, "stdout"},
		{"contains without marker", `{"op":"contains","in":"stdout"}`, "marker"},
		{"contains with both targets", `{"op":"contains","in":"stdout","path":"$.a","marker":"m"}`, "contains"},
		{"contains with no target", `{"op":"contains","marker":"m"}`, "contains"},
		{"contains on stderr is not in the vocabulary", `{"op":"contains","in":"stderr","marker":"m"}`, "in"},
		{"contains path with wildcard", `{"op":"contains","path":"$.a[*].b","marker":"m"}`, "wildcard"},
		{"contains bad path", `{"op":"contains","path":"$..a","marker":"m"}`, "json path"},
		{"json_path without capture", `{"op":"json_path","path":"$.a"}`, "json_path"},
		{"json_path with wildcard", `{"op":"json_path","path":"$.a[*].b","equals_capture":"V"}`, "wildcard"},
		{"id_set without a declared set", `{"op":"id_set","path":"$.items[*].id"}`, "id_set"},
		{"id_set with duplicate ids", `{"op":"id_set","path":"$.items[*].id","equals":["a","a"]}`, "duplicate"},
		{"compare with a bad relation", `{"op":"compare","left":{"step":"a","field":"stdout"},"right":{"step":"b","field":"stdout"},"relation":"similar"}`, "relation"},
		{"compare with a bad field", `{"op":"compare","left":{"step":"a","field":"stdin"},"right":{"step":"b","field":"stdout"},"relation":"equal"}`, "field"},
		{"compare side with field and path", `{"op":"compare","left":{"step":"a","field":"stdout","path":"$.a"},"right":{"step":"b","field":"stdout"},"relation":"equal"}`, "field"},
		{"compare side with neither", `{"op":"compare","left":{"step":"a"},"right":{"step":"b","field":"stdout"},"relation":"equal"}`, "field"},
		{"stderr_contains without capture", `{"op":"stderr_contains"}`, "capture"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseAssertion([]byte(tc.js))
			if err == nil {
				t.Fatalf("parseAssertion(%s) succeeded, want an error", tc.js)
			}
			if !strings.Contains(strings.ToLower(err.Error()), tc.want) {
				t.Errorf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

type evalCase struct {
	name string
	op   string
	cur  *stepResult
	ok   bool
}

func TestAssertionEval(t *testing.T) {
	env := &evalEnv{
		Steps: map[string]*stepResult{
			"s1": {Name: "s1", Exit: 3, Stdout: []byte("same\n"), Stderr: []byte("gone: deleted\n")},
			"s2": {Name: "s2", Exit: 3, Stdout: []byte("same\n"), Stderr: []byte("not_found: never\n")},
			"s3": {Name: "s3", Exit: 0, Stdout: []byte(`{"result":{"version":"v-one"}}`), Stderr: nil},
			"s4": {Name: "s4", Exit: 0, Stdout: []byte(`{"result":{"version":"v-two"}}`), Stderr: nil},
		},
		Captures: map[string]string{"V": "tok123", "BODY": "first body\n"},
		Inputs:   map[string][]byte{"B1": []byte("first body\n")},
	}
	json := `{"result":{"title":"hello world","n":5,"items":[{"id":"b"},{"id":"a"}],"ids":["y","x"],"version":"tok123"}}`
	cases := []evalCase{
		// 1. exit
		{"exit equals holds", `{"op":"exit","equals":0}`, res(0, "", ""), true},
		{"exit equals fails", `{"op":"exit","equals":0}`, res(1, "", ""), false},
		{"exit nonzero holds", `{"op":"exit","nonzero":true}`, res(3, "", ""), true},
		{"exit nonzero fails on zero", `{"op":"exit","nonzero":true}`, res(0, "", ""), false},

		// 2. stderr_code
		{"stderr_code equals holds (json)", `{"op":"stderr_code","equals":"revision_conflict"}`, res(4, "", `{"code":"revision_conflict","message":"m","retryable":false}`), true},
		{"stderr_code equals holds (plain text)", `{"op":"stderr_code","equals":"gone"}`, res(3, "", "gone: graph resource was deleted\n"), true},
		{"stderr_code equals fails on another code", `{"op":"stderr_code","equals":"gone"}`, res(3, "", "not_found: x\n"), false},
		{"stderr_code equals fails when there is no code", `{"op":"stderr_code","equals":"gone"}`, res(3, "", "Error: boom\n"), false},
		{"stderr_code not_equals holds", `{"op":"stderr_code","not_equals":"gone"}`, res(3, "", "not_found: x\n"), true},
		{"stderr_code not_equals fails", `{"op":"stderr_code","not_equals":"gone"}`, res(3, "", "gone: x\n"), false},

		// 3. stdout: raw bytes, no normalization
		{"stdout equals_input holds", `{"op":"stdout","equals_input":"B1"}`, res(0, "first body\n", ""), true},
		{"stdout equals_input is byte exact", `{"op":"stdout","equals_input":"B1"}`, res(0, "first body", ""), false},
		{"stdout equals_input fails on one byte", `{"op":"stdout","equals_input":"B1"}`, res(0, "first bodx\n", ""), false},
		{"stdout equals_capture holds", `{"op":"stdout","equals_capture":"BODY"}`, res(0, "first body\n", ""), true},
		{"stdout equals_capture fails", `{"op":"stdout","equals_capture":"BODY"}`, res(0, "other\n", ""), false},
		{"stdout empty holds", `{"op":"stdout","empty":true}`, res(3, "", "gone: x"), true},
		{"stdout empty fails on a newline", `{"op":"stdout","empty":true}`, res(0, "\n", ""), false},

		// 4. contains / not_contains, on stdout or a JSON-path string value
		{"contains stdout holds", `{"op":"contains","in":"stdout","marker":"world"}`, res(0, json, ""), true},
		{"contains stdout fails", `{"op":"contains","in":"stdout","marker":"absent"}`, res(0, json, ""), false},
		{"contains path holds", `{"op":"contains","path":"$.result.title","marker":"hello"}`, res(0, json, ""), true},
		{"contains path fails", `{"op":"contains","path":"$.result.title","marker":"absent"}`, res(0, json, ""), false},
		{"contains path fails on a non-string value", `{"op":"contains","path":"$.result.n","marker":"5"}`, res(0, json, ""), false},
		{"contains path fails on a missing path", `{"op":"contains","path":"$.result.zzz","marker":"x"}`, res(0, json, ""), false},
		{"contains path fails when stdout is not json", `{"op":"contains","path":"$.a","marker":"x"}`, res(0, "plain", ""), false},
		{"not_contains stdout holds", `{"op":"not_contains","in":"stdout","marker":"absent"}`, res(0, json, ""), true},
		{"not_contains stdout fails", `{"op":"not_contains","in":"stdout","marker":"world"}`, res(0, json, ""), false},
		{"not_contains path holds", `{"op":"not_contains","path":"$.result.title","marker":"absent"}`, res(0, json, ""), true},
		{"not_contains path fails", `{"op":"not_contains","path":"$.result.title","marker":"hello"}`, res(0, json, ""), false},
		{"not_contains path fails on a missing path", `{"op":"not_contains","path":"$.result.zzz","marker":"x"}`, res(0, json, ""), false},

		// 5. json_path equals a captured value; id_set ignores order
		{"json_path holds", `{"op":"json_path","path":"$.result.version","equals_capture":"V"}`, res(0, json, ""), true},
		{"json_path fails", `{"op":"json_path","path":"$.result.title","equals_capture":"V"}`, res(0, json, ""), false},
		{"json_path fails on an undefined capture", `{"op":"json_path","path":"$.result.version","equals_capture":"NOPE"}`, res(0, json, ""), false},
		{"id_set holds ignoring order", `{"op":"id_set","path":"$.result.items[*].id","equals":["a","b"]}`, res(0, json, ""), true},
		{"id_set holds on a plain array path", `{"op":"id_set","path":"$.result.ids","equals":["x","y"]}`, res(0, json, ""), true},
		{"id_set fails when an id is missing", `{"op":"id_set","path":"$.result.items[*].id","equals":["a","b","c"]}`, res(0, json, ""), false},
		{"id_set fails on an extra id", `{"op":"id_set","path":"$.result.items[*].id","equals":["a"]}`, res(0, json, ""), false},
		{"id_set fails on a duplicate in the actual ids", `{"op":"id_set","path":"$.ids[*]","equals":["a","b"]}`, res(0, `{"ids":["a","a","b"]}`, ""), false},

		// 6. compare two steps' fields
		{"compare stdout equal holds", `{"op":"compare","left":{"step":"s1","field":"stdout"},"right":{"step":"s2","field":"stdout"},"relation":"equal"}`, res(0, "", ""), true},
		{"compare stdout equal fails", `{"op":"compare","left":{"step":"s1","field":"stdout"},"right":{"step":"s3","field":"stdout"},"relation":"equal"}`, res(0, "", ""), false},
		{"compare stderr differ holds", `{"op":"compare","left":{"step":"s1","field":"stderr"},"right":{"step":"s2","field":"stderr"},"relation":"differ"}`, res(0, "", ""), true},
		{"compare stderr differ fails", `{"op":"compare","left":{"step":"s1","field":"stderr"},"right":{"step":"s1","field":"stderr"},"relation":"differ"}`, res(0, "", ""), false},
		{"compare stderr_code differ holds", `{"op":"compare","left":{"step":"s1","field":"stderr_code"},"right":{"step":"s2","field":"stderr_code"},"relation":"differ"}`, res(0, "", ""), true},
		{"compare stderr_code differ fails when both are the same code", `{"op":"compare","left":{"step":"s1","field":"stderr_code"},"right":{"step":"s1","field":"stderr_code"},"relation":"differ"}`, res(0, "", ""), false},
		{"compare stderr_code equal holds", `{"op":"compare","left":{"step":"s1","field":"stderr_code"},"right":{"step":"s1","field":"stderr_code"},"relation":"equal"}`, res(0, "", ""), true},
		{"compare exit equal holds", `{"op":"compare","left":{"step":"s1","field":"exit"},"right":{"step":"s2","field":"exit"},"relation":"equal"}`, res(0, "", ""), true},
		{"compare exit equal fails", `{"op":"compare","left":{"step":"s1","field":"exit"},"right":{"step":"s3","field":"exit"},"relation":"equal"}`, res(0, "", ""), false},
		{"compare json path differ holds", `{"op":"compare","left":{"step":"s3","path":"$.result.version"},"right":{"step":"s4","path":"$.result.version"},"relation":"differ"}`, res(0, "", ""), true},
		{"compare json path equal fails", `{"op":"compare","left":{"step":"s3","path":"$.result.version"},"right":{"step":"s4","path":"$.result.version"},"relation":"equal"}`, res(0, "", ""), false},
		{"compare json path fails on a missing path", `{"op":"compare","left":{"step":"s3","path":"$.nope"},"right":{"step":"s4","path":"$.result.version"},"relation":"differ"}`, res(0, "", ""), false},
		{"compare fails on an unknown step", `{"op":"compare","left":{"step":"ghost","field":"stdout"},"right":{"step":"s1","field":"stdout"},"relation":"equal"}`, res(0, "", ""), false},

		// 7. stderr_contains a captured token
		{"stderr_contains holds", `{"op":"stderr_contains","capture":"V"}`, res(4, "", `{"code":"revision_conflict","message":"current tok123"}`), true},
		{"stderr_contains fails", `{"op":"stderr_contains","capture":"V"}`, res(4, "", `{"code":"revision_conflict","message":"current other"}`), false},
		{"stderr_contains fails on an undefined capture", `{"op":"stderr_contains","capture":"NOPE"}`, res(4, "", "anything"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := mustAssertion(t, tc.op)
			err := a.eval(tc.cur, env)
			if tc.ok && err != nil {
				t.Errorf("assertion %s should hold, got: %v", tc.op, err)
			}
			if !tc.ok && err == nil {
				t.Errorf("assertion %s should fail", tc.op)
			}
		})
	}
}

// The failure message is what an author reads when a scenario goes red, so it
// must say which op failed and against what.
func TestAssertionFailureMessageNamesTheOp(t *testing.T) {
	a := mustAssertion(t, `{"op":"exit","equals":0}`)
	err := a.eval(res(4, "", ""), &evalEnv{})
	if err == nil || !strings.Contains(err.Error(), "exit") || !strings.Contains(err.Error(), "4") {
		t.Errorf("error %v should name the op and the observed exit", err)
	}
}
