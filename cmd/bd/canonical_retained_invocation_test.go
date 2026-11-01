package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Independent literal bodies fix the singleton contract, including the absence
// of experimental member discriminators and exact numeric/pin token values.
func TestCanonicalRetainedNativeSingletonBodies(t *testing.T) {
	cases := []struct {
		op    canonicalRetainedOperation
		args  []string
		flags []string
		want  string
	}{
		{retainedCreateBead, nil, []string{"--type=https://types.test/task", "--properties={\"n\":9007199254740993,\"literal\":\"@a\"}"}, `{"properties":{"n":9007199254740993,"literal":"@a"},"type":"https://types.test/task"}`},
		{retainedUpdateBead, []string{"https://data.test/s/a"}, []string{"--replace-properties={}", "--if-revision==r:<opaque>"}, `{"bead":"https://data.test/s/a","expectedRevision":"r:<opaque>","properties":{}}`},
		{retainedDeleteBead, []string{"https://data.test/s/a"}, []string{"--force"}, `{"bead":"https://data.test/s/a"}`},
		{retainedCreateLink, []string{"[https://data.test/s/a]@=r:one", "https://data.test/s/b"}, []string{"--type=[https://types.test/edge]@=type:one"}, `{"source":{"uri":"https://data.test/s/a","revision":"r:one"},"target":"https://data.test/s/b","type":{"uri":"https://types.test/edge","revision":"type:one"}}`},
		{retainedUpdateLink, []string{"https://data.test/s/l"}, []string{"--to=[https://data.test/s/c]@=r:two"}, `{"link":"https://data.test/s/l","target":{"uri":"https://data.test/s/c","revision":"r:two"}}`},
		{retainedDeleteLink, []string{"https://data.test/s/l"}, []string{"--force", "--if-revision==l:one"}, `{"expectedRevision":"l:one","link":"https://data.test/s/l"}`},
	}
	for _, test := range cases {
		t.Run(string(test.op), func(t *testing.T) {
			command := canonicalRetainedInputProbe()
			if err := command.ParseFlags(test.flags); err != nil {
				t.Fatal(err)
			}
			body, _, err := canonicalRetainedNativeInput(command, test.args, test.op, "https://data.test/s/")
			if err != nil {
				t.Fatal(err)
			}
			var got, want map[string]json.RawMessage
			if json.Unmarshal(body, &got) != nil || json.Unmarshal([]byte(test.want), &want) != nil {
				t.Fatal("invalid test/body JSON")
			}
			if len(got) != len(want) {
				t.Fatalf("singleton body members: %s", body)
			}
			for name, expected := range want {
				var left, right bytes.Buffer
				if json.Compact(&left, got[name]) != nil || json.Compact(&right, expected) != nil || left.String() != right.String() {
					t.Fatalf("member %s: %s want %s", name, got[name], expected)
				}
			}
			if bytes.Contains(body, []byte("9007199254740992")) {
				t.Fatal("numeric precision was lost")
			}
		})
	}
}
func canonicalRetainedInputProbe() *cobra.Command {
	cmd := &cobra.Command{Use: "probe"}
	for _, name := range []string{"id", "type", "properties", "metadata", "replace-properties", "replace-metadata", "patch-properties", "patch-metadata", "revision-provenance", "if-revision", "from", "to"} {
		cmd.Flags().String(name, "", "")
	}
	for _, name := range []string{"set-property", "unset-property", "set-metadata", "unset-metadata"} {
		cmd.Flags().StringArray(name, nil, "")
	}
	cmd.Flags().Bool("force", false, "")
	return cmd
}
func TestCanonicalRetainedLosslessEditAndDuplicateRefusal(t *testing.T) {
	command := canonicalRetainedInputProbe()
	if err := command.ParseFlags([]string{"--properties={\"large\":9007199254740993,\"fraction\":1.2300e-40}"}); err != nil {
		t.Fatal(err)
	}
	raw, _, err := canonicalRetainedNativeInput(command, []string{"https://data.test/s/a"}, retainedUpdateBead, "https://data.test/s/")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("9007199254740993")) || !bytes.Contains(raw, []byte("1.2300e-40")) {
		t.Fatalf("number tokens changed: %s", raw)
	}
	for _, bad := range []string{`{"a":1,"\u0061":2}`, `{"a":{"x":1,"x":2}}`, `{} {}`, `{"x":NaN}`} {
		if canonicalRetainedStrictJSON([]byte(bad)) == nil {
			t.Fatalf("admitted %s", bad)
		}
	}
}
func TestCanonicalRetainedJournalPreservesOriginalAndRejectsTamper(t *testing.T) {
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(parent, "retained-invocations")
	record := canonicalRetainedJournalRecord{Version: 1, Owner: "local-test-owner", Provider: "test-provider", Request: canonicalRetainedPrepared{Scope: "https://data.test/s/", Contract: canonicalRetainedContract, Operation: retainedCreateBead, Method: "POST", Target: "https://commands.test/create", Key: "Case_Sensitive-1", Body: json.RawMessage(`{"type":"https://types.test/task","properties":{"n":9007199254740993}}`)}}
	saved, err := saveCanonicalRetainedInvocation(root, record)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := loadCanonicalRetainedInvocation(root, saved.path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(loaded.record.Request.Body, record.Request.Body) || loaded.record.Request.Key != record.Request.Key {
		t.Fatal("recovery changed original meaning/key")
	}
	original, err := os.ReadFile(saved.path)
	if err != nil {
		t.Fatal(err)
	}
	again, err := saveCanonicalRetainedInvocation(root, record)
	if err != nil || again.path != saved.path {
		t.Fatalf("same invocation save: %v", err)
	}
	after, err := os.ReadFile(saved.path)
	if err != nil || !bytes.Equal(after, original) {
		t.Fatal("existing custody bytes changed")
	}
	if err := os.WriteFile(saved.path, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCanonicalRetainedInvocation(root, saved.path); err == nil {
		t.Fatal("tampered invocation accepted")
	}
	if _, err := saveCanonicalRetainedInvocation(root, record); err == nil {
		t.Fatal("tampered custody was silently overwritten")
	}
}
func TestCanonicalRetainedJournalRejectsSymlinkAndPublicDirectory(t *testing.T) {
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(parent, "target")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(parent, "symlink")
	if err := os.Symlink(target, symlink); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if canonicalRetainedPrivateDirectory(symlink, false) == nil {
		t.Fatal("symlink journal accepted")
	}
	if err := os.Chmod(target, 0755); err != nil {
		t.Fatal(err)
	}
	if canonicalRetainedPrivateDirectory(target, false) == nil {
		t.Fatal("public journal directory accepted")
	}
}
func TestCanonicalRetainedSubjectGuardAndNoopAdmission(t *testing.T) {
	command := canonicalRetainedInputProbe()
	if err := command.ParseFlags([]string{"--patch-properties=[]"}); err != nil {
		t.Fatal(err)
	}
	body, _, err := canonicalRetainedNativeInput(command, []string{"https://data.test/s/a"}, retainedUpdateBead, "https://data.test/s/")
	if err != nil || !strings.Contains(string(body), `"propertiesChange":[]`) {
		t.Fatalf("explicit empty Change/no-op input: %s %v", body, err)
	}
	if _, _, err := canonicalRetainedNativeInput(command, []string{"[https://data.test/s/a]@=r1"}, retainedUpdateBead, "https://data.test/s/"); err == nil {
		t.Fatal("pinned mutation subject accepted")
	}
}

func TestCanonicalRetainedProblemConditionalShapesAndOriginalBytes(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		valid  bool
	}{
		{"pruned-archive", 410, `{"type":"https://github.com/gastownhall/bdp/problems/gone","code":"resource-pruned","retry":"never","archivedAt":"https://archive.example/r"}`, true},
		{"pruned-pinned-archive", 410, `{"type":"https://github.com/gastownhall/bdp/problems/gone","code":"resource-pruned","retry":"never","archivedAt":{"uri":"https://archive.example/r","revision":"opaque"}}`, true},
		{"unretained-missing", 409, `{"type":"https://github.com/gastownhall/bdp/problems/conflict","code":"revision-unretained","retry":"after-state-change","missing":{"complete":true,"items":[{"kind":"record"},{"kind":"property","pointer":"/properties/title"},{"kind":"owned-links","type":"https://types.example/edge"}]}}`, true},
		{"validation-diagnostics", 422, `{"type":"https://github.com/gastownhall/bdp/problems/validation","code":"validation-failed","retry":"never","diagnostics":[{"message":"Invalid value","type":"https://types.example/task","schemaLocation":"https://types.example/schema#/$defs/title","document":"properties","instanceLocation":"/title"}],"diagnosticsTruncated":false}`, true},
		{"delay-integer-exponent", 429, `{"type":"https://github.com/gastownhall/bdp/problems/rate-limit","code":"rate-limited","retry":"after-delay","retryAfter":1e2,"title":"","detail":"","instance":"urn:request:one","extension":{"opaque":true}}`, true},
		{"validation-missing-diagnostics", 422, `{"type":"https://github.com/gastownhall/bdp/problems/validation","code":"validation-failed","retry":"never"}`, false},
		{"validation-empty-diagnostics", 422, `{"type":"https://github.com/gastownhall/bdp/problems/validation","code":"validation-failed","retry":"never","diagnostics":[]}`, false},
		{"validation-diagnostic-pair", 422, `{"type":"https://github.com/gastownhall/bdp/problems/validation","code":"validation-failed","retry":"never","diagnostics":[{"message":"Invalid","type":"https://types.example/task"}]}`, false},
		{"validation-truncated-type", 422, `{"type":"https://github.com/gastownhall/bdp/problems/validation","code":"validation-failed","retry":"never","diagnostics":[{"message":"Invalid"}],"diagnosticsTruncated":null}`, false},
		{"unretained-missing-required", 409, `{"type":"https://github.com/gastownhall/bdp/problems/conflict","code":"revision-unretained","retry":"after-state-change"}`, false},
		{"unretained-disclosure", 409, `{"type":"https://github.com/gastownhall/bdp/problems/conflict","code":"revision-unretained","retry":"after-state-change","missing":{"complete":false,"items":[]},"resource":null}`, false},
		{"unretained-duplicate-decoded-item", 409, `{"type":"https://github.com/gastownhall/bdp/problems/conflict","code":"revision-unretained","retry":"after-state-change","missing":{"complete":true,"items":[{"kind":"property","pointer":"/properties/a"},{"pointer":"\u002fproperties/a","kind":"property"}]}}`, false},
		{"erased-pointer-disclosure", 410, `{"type":"https://github.com/gastownhall/bdp/problems/gone","code":"resource-erased","retry":"never","pointer":null}`, false},
		{"wrong-archive-diagnosis", 404, `{"type":"https://github.com/gastownhall/bdp/problems/not-found","code":"resource-not-found","retry":"after-state-change","archivedAt":"https://archive.example/r"}`, false},
		{"wrong-diagnostic-diagnosis", 404, `{"type":"https://github.com/gastownhall/bdp/problems/not-found","code":"resource-not-found","retry":"after-state-change","diagnosticsTruncated":false}`, false},
		{"non-delay-retryAfter", 404, `{"type":"https://github.com/gastownhall/bdp/problems/not-found","code":"resource-not-found","retry":"after-state-change","retryAfter":0}`, false},
		{"delay-fraction", 429, `{"type":"https://github.com/gastownhall/bdp/problems/rate-limit","code":"rate-limited","retry":"after-delay","retryAfter":0.5}`, false},
		{"title-type", 404, `{"type":"https://github.com/gastownhall/bdp/problems/not-found","code":"resource-not-found","retry":"after-state-change","title":null}`, false},
		{"instance-relative", 404, `{"type":"https://github.com/gastownhall/bdp/problems/not-found","code":"resource-not-found","retry":"after-state-change","instance":"relative"}`, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			reply := canonicalRetainedReply{Sent: true, Status: test.status, Headers: map[string][]string{"Content-Type": {"application/problem+json"}}, NativeBody: []byte(test.body)}
			reply.validateMutation(retainedUpdateBead)
			if reply.Checked != test.valid || (reply.ParseErr == nil) != test.valid {
				t.Fatalf("Problem admission valid=%v: %#v", test.valid, reply)
			}
			if !bytes.Equal(reply.NativeBody, []byte(test.body)) {
				t.Fatal("validation overwrote original Problem bytes")
			}
			if test.valid && (!bytes.Equal(reply.Problem, reply.NativeBody) || reply.OperationErr == nil) {
				t.Fatal("checked native Problem/disposition was lost")
			}
		})
	}
}

func TestCanonicalRetainedDiscoveryExactNamesAndOptionalShapes(t *testing.T) {
	base := `{"bdpVersion":"0.1","profile":"core","scope":"https://data.example/s/","beads":"https://data.example/b","links":"https://data.example/l","events":"https://data.example/e"}`
	for _, name := range []string{"bdpVersion", "profile", "scope", "beads", "links", "events"} {
		wrong := strings.Replace(base, `"`+name+`":`, `"`+strings.ToUpper(name)+`":`, 1)
		if _, err := parseCanonicalRetainedDiscovery([]byte(wrong)); err == nil {
			t.Fatalf("case-insensitive discovery alias admitted: %s", name)
		}
	}
	good := strings.TrimSuffix(base, "}") + `,"order":"canonical-uri","limits":{"page":{"maximumItems":10},"retention":{"maximumSnapshotLifetime":"P1D"}},"maximumEndpointMultiplicity":[{"linkConformsTo":"https://types.example/edge","endpoint":"source","max":1}],"historicalResolution":{"version":1},"capabilities":{"urn:example:future":{"new-version":{"nested":{"literal":"@a"},"n":9007199254740993}},"https://github.com/gastownhall/bdp/capabilities/atomic-batch":{"future":{"opaque":true}}}}`
	desc, err := parseCanonicalRetainedDiscovery([]byte(good))
	if err != nil {
		t.Fatalf("valid optional/opaque capability discovery refused: %v", err)
	}
	if !bytes.Contains(desc.Capabilities, []byte("9007199254740993")) || !bytes.Contains(desc.Capabilities, []byte(`"new-version"`)) {
		t.Fatal("unknown capability/version details were dropped or reinterpreted")
	}
	for _, extra := range []string{
		`,"BDPVERSION":"0.1"`, `,"operations":"https://commands.example/op"`, `,"limits":null`, `,"limits":{"page":{"maximumItems":0}}`, `,"limits":{"patch":{}}`, `,"order":"other"`, `,"historicalResolution":{"VERSION":1}`, `,"maximumEndpointMultiplicity":[{"linkConformsTo":"https://types.example/edge","endpoint":"other","max":1}]`, `,"capabilities":{"urn:future":{}}`, `,"capabilities":{"urn:future":{"1":null}}`, `,"capabilities":{"https://github.com/gastownhall/bdp/capabilities/atomic-batch":{"1":{"operation":"https://commands.example/op","operations":["createBead"]}}}`,
	} {
		if _, err := parseCanonicalRetainedDiscovery([]byte(strings.TrimSuffix(base, "}") + extra + "}")); err == nil {
			t.Fatalf("malformed optional discovery admitted: %s", extra)
		}
	}
}

func TestCanonicalRetainedCurrentNestedMetadataAndRetainedOptionalMetadata(t *testing.T) {
	link := `{"id":"https://data.example/link","type":"https://types.example/edge","revision":"l1","source":"https://data.example/a","target":"https://data.example/b","properties":{}}`
	retained := `{"id":"https://data.example/a","type":"https://types.example/task","revision":"r1","properties":{}}`
	if err := validateCanonicalRetainedResource([]byte(retained), ""); err != nil {
		t.Fatalf("retained result incorrectly requires metadata: %v", err)
	}
	if err := validateCanonicalRetainedResource([]byte(link), ""); err != nil {
		t.Fatalf("retained Link incorrectly requires metadata: %v", err)
	}
	for _, nested := range []string{`,"links":{"items":[` + link + `],"next":null}`, `,"ownedLinks":{"https://types.example/edge":[` + link + `]}`} {
		current := strings.TrimSuffix(retained, "}") + `,"metadata":{}` + nested + `}`
		if err := validateCanonicalRetainedCurrentResource([]byte(current), ""); err == nil {
			t.Fatal("nested current Link without metadata admitted")
		}
	}
	completeLink := strings.TrimSuffix(link, "}") + `,"metadata":{}}`
	if err := validateCanonicalRetainedCurrentResource([]byte(completeLink), ""); err != nil {
		t.Fatalf("complete current Link refused: %v", err)
	}
}
