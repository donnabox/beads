package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
)

// Literal expectations below derive from selected batch1 and independent B01,
// B02, B06, B10/B11 relations. They are not an echo of a request/result emitter.
const retainedBatchNoopInput = `{"capabilityVersion":"1","operations":[{"operation":"updateBead","bead":"https://data.test/s/a","properties":{"n":1}},{"operation":"updateBead","bead":"https://data.test/s/a","properties":{"n":0}}]}`
const retainedBatchNoopResult = `{"capabilityVersion":"1","changed":false,"results":[{"operationIndex":0,"operation":"updateBead","id":"https://data.test/s/a"},{"operationIndex":1,"operation":"updateBead","id":"https://data.test/s/a"}],"resources":[{"id":"https://data.test/s/a","resourceKind":"bead","outcome":"unchanged","changed":false,"resource":{"id":"https://data.test/s/a","type":"https://types.test/task","revision":"original-r0","properties":{"n":0}}}]}`
const retainedBatchAllocatedInput = `{"capabilityVersion":"1","operations":[{"operation":"createBead","name":"a","type":"https://types.test/task","properties":{"literal":"@a","large":9007199254740993}},{"operation":"deleteBead","bead":"@a"}]}`
const retainedBatchAllocatedResult = `{"capabilityVersion":"1","changed":true,"transaction":"txn-original","results":[{"operationIndex":0,"operation":"createBead","operationName":"a","id":"https://data.test/s/new"},{"operationIndex":1,"operation":"deleteBead","id":"https://data.test/s/new"}],"resources":[{"id":"https://data.test/s/new","resourceKind":"bead","outcome":"allocated","changed":true}]}`

func TestCanonicalRetainedBatchGrammarAndLiteralPreservation(t *testing.T) {
	valid := []string{
		retainedBatchNoopInput, retainedBatchAllocatedInput,
		`{"capabilityVersion":"1","operations":[{"operation":"createBead","name":"a","type":{"uri":"https://types.test/T","revision":"type-literal"}},{"operation":"createLink","name":"l","type":"https://types.test/L","source":{"uri":"@a","revision":"historical-literal"},"target":"@a","properties":{"notAHandle":"@missing","n":1.234567890123456789e20}},{"operation":"updateLink","link":"@l","propertiesChange":[],"expectedRevision":"baseline-literal"}]}`,
	}
	for _, text := range valid {
		raw := []byte(text)
		before := bytes.Clone(raw)
		if _, err := canonicalRetainedBatchMembers(raw); err != nil {
			t.Fatalf("valid native input refused: %v %s", err, text)
		}
		if !bytes.Equal(raw, before) {
			t.Fatal("native numeric/string/pin bytes rewritten")
		}
	}
	invalid := []string{
		`{"capabilityVersion":"2","operations":[]}`,
		`{"version":"2","operations":[]}`,
		`{"capabilityVersion":"1","operations":[]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"putAlias","path":"a"}]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"deleteBead","bead":"@later"},{"operation":"createBead","name":"later","type":"https://types.test/T"}]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"createLink","name":"a","type":"https://types.test/L","source":"a","target":"b"},{"operation":"deleteBead","bead":"@a"}]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"createBead","name":"a","type":"https://types.test/T"},{"operation":"createBead","name":"a","type":"https://types.test/T"}]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"createBead","type":"@a"}]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"deleteBead","bead":{"uri":"a","revision":"r"}}]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"updateBead","bead":"a","properties":{},"propertiesChange":[]}]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"createBead","type":"https://types.test/T","properties":{"x":1,"x":2}}]}`,
		`{"capabilityVersion":"1","operations":[{"operation":"deleteBead","bead":{"binding":"a"}}]}`,
	}
	for _, text := range invalid {
		if _, err := canonicalRetainedBatchMembers([]byte(text)); err == nil {
			t.Fatalf("invalid native input accepted: %s", text)
		}
	}
}

func TestCanonicalRetainedBatchIndependentFinalResultRelations(t *testing.T) {
	for _, fixture := range [][2]string{{retainedBatchNoopInput, retainedBatchNoopResult}, {retainedBatchAllocatedInput, retainedBatchAllocatedResult}} {
		if err := canonicalRetainedBatchResult([]byte(fixture[1]), []byte(fixture[0]), "https://data.test/s/"); err != nil {
			t.Fatalf("independent result refused: %v", err)
		}
	}
	invalid := []struct{ input, result string }{
		{retainedBatchNoopInput, strings.Replace(retainedBatchNoopResult, `"operationIndex":1`, `"operationIndex":0`, 1)},
		{retainedBatchNoopInput, strings.Replace(retainedBatchNoopResult, `"changed":false`, `"changed":true,"transaction":"fake"`, 1)},
		{retainedBatchNoopInput, strings.Replace(retainedBatchNoopResult, `"changed":false`, `"changed":false,"transaction":"forbidden"`, 1)},
		{retainedBatchNoopInput, strings.Replace(retainedBatchNoopResult, `"outcome":"unchanged"`, `"outcome":"allocated"`, 1)},
		{retainedBatchAllocatedInput, strings.Replace(retainedBatchAllocatedResult, `"operationName":"a"`, `"operationName":"other"`, 1)},
		{retainedBatchAllocatedInput, strings.Replace(retainedBatchAllocatedResult, `"outcome":"allocated"`, `"revision":"must-not-escape","outcome":"allocated"`, 1)},
		{retainedBatchAllocatedInput, strings.Replace(retainedBatchAllocatedResult, `"operation":"deleteBead","id":"https://data.test/s/new"`, `"operation":"deleteBead","id":"https://data.test/s/different"`, 1)},
		{retainedBatchAllocatedInput, `{"outcome":"created","changed":true,"resource":{}}`},
	}
	for _, test := range invalid {
		if err := canonicalRetainedBatchResult([]byte(test.result), []byte(test.input), "https://data.test/s/"); err == nil {
			t.Fatalf("invalid result accepted: %s", test.result)
		}
	}
	// Same-key renamed labels project only mapping labels, not identities or rows.
	renamedInput := strings.ReplaceAll(retainedBatchAllocatedInput, `"a"`, `"renamed"`)
	renamedInput = strings.ReplaceAll(renamedInput, `"@a"`, `"@renamed"`)
	renamedResult := strings.Replace(retainedBatchAllocatedResult, `"operationName":"a"`, `"operationName":"renamed"`, 1)
	if err := canonicalRetainedBatchResult([]byte(renamedResult), []byte(renamedInput), "https://data.test/s/"); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalRetainedBatchFeatureProblemsStayDistinct(t *testing.T) {
	capability := `{"type":"https://github.com/gastownhall/bdp/problems/request","status":422,"code":"capability-not-supported","retry":"after-state-change","capability":"https://github.com/gastownhall/bdp/capabilities/atomic-batch","capabilityVersion":"next","pointer":"/capabilityVersion"}`
	operation := `{"type":"https://github.com/gastownhall/bdp/problems/request","status":422,"code":"operation-not-supported","retry":"after-state-change","operation":"updateBead","pointer":"/operations/0/operation"}`
	for _, raw := range []string{capability, operation} {
		if err := canonicalRetainedBatchProblem([]byte(raw), 422); err != nil {
			t.Fatal(err)
		}
		if err := canonicalRetainedProblem([]byte(raw), 422, canonicalRetainedWriteProblem); err == nil {
			t.Fatal("feature Problem admitted as ordinary singleton")
		}
	}
	for _, raw := range []string{strings.Replace(operation, "/operations/0/operation", "/operations/00/operation", 1), strings.Replace(capability, `"pointer"`, `"extra":true,"pointer"`, 1), strings.Replace(capability, `"status":422`, `"status":400`, 1)} {
		if err := canonicalRetainedBatchProblem([]byte(raw), 422); err == nil {
			t.Fatalf("invalid feature Problem accepted: %s", raw)
		}
	}
	reply := canonicalRetainedReply{Status: 422, Headers: http.Header{"Content-Type": []string{"application/problem+json"}}, NativeBody: []byte(strings.Replace(capability, `"capabilityVersion":"next"`, `"capabilityVersion":"1"`, 1)), Sent: true}
	reply.validateBatch([]byte(retainedBatchNoopInput), "https://data.test/s/")
	if !reply.Checked || reply.OperationErr == nil || reply.ParseErr != nil || !bytes.Equal(reply.Problem, reply.NativeBody) {
		t.Fatal("original feature refusal lost")
	}
}

func TestCanonicalRetainedBatchOriginalCloseOutputCauses(t *testing.T) {
	closeFailure := errors.New("batch response close")
	outputFailure := errors.New("batch output write")
	body := &canonicalRetainedBody{Reader: strings.NewReader(retainedBatchNoopResult), closeErr: closeFailure}
	calls := 0
	options := canonicalRetainedTestBindings(&http.Client{Transport: canonicalRetainedRoundTrip(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{canonicalRetainedMedia}}, Body: body}, nil
	})})
	request, _ := http.NewRequestWithContext(context.Background(), "POST", "https://data.test/s/batch", strings.NewReader(retainedBatchNoopInput))
	reply := options.exchange(request)
	reply.validateBatch([]byte(retainedBatchNoopInput), "https://data.test/s/")
	writer := &canonicalRetainedFailWriter{err: outputFailure}
	output, err := newCanonicalMutationOutput(canonicalMutationOutputRetainedBatch, clioutput.MutationDescriptor{BDPSpec: canonicalRetainedContract, Required: []string{"capabilityVersion", "changed", "results", "resources"}}, clioutput.Options{Formats: []string{"json"}}, writer)
	if err != nil {
		t.Fatal(err)
	}
	result := output.Complete(reply, "", reply.OperationErr, reply.CloseErr)
	if !reply.Checked || result.Disposition != canonicalMutationOutputReturned || result.OperationErr != nil || !errors.Is(result.CloseErr, closeFailure) || !errors.Is(result.OutputErr, outputFailure) || calls != 1 || body.closes != 1 || writer.calls != 1 || !bytes.Equal(reply.Mutation, []byte(retainedBatchNoopResult)) {
		t.Fatalf("batch outcome/Close/output changed: %#v", result)
	}
	output.Complete(reply, "", nil, nil)
	if writer.calls != 1 || calls != 1 {
		t.Fatal("output retried operation")
	}
}

// This helper constructs a fresh initialized tree. It exercises actual profile
// decoration, Args, pre-provider admission and batch runner, with injected local
// disposable custody. It does not certify a real SQLite service or installed bd.
func retainedBatchRoot(t *testing.T, options canonicalRetainedBindings) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	root := &cobra.Command{Use: "bd"}
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(io.Discard)
	root.SetContext(context.WithValue(context.Background(), canonicalRetainedBindingsKey, options))
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error { _, err := admitCanonicalRetained(cmd); return err }
	batch := &cobra.Command{Use: "batch", Args: func(*cobra.Command, []string) error { return errors.New("legacy local batch Args ran") }, RunE: runCanonicalRetained}
	for _, name := range []string{"scope", "input-format", "batch-version", "file", "dry-run-mode", "idempotency-key", "retry-file"} {
		batch.Flags().String(name, "", "")
	}
	registerCanonicalOutputFlags(batch)
	root.AddCommand(batch)
	return root, output
}
func retainedBatchExecute(root *cobra.Command, argv []string) error {
	_, err := executeCLIProfileRootWith(root, argv, func(*string) (cliProfileSelection, error) {
		return cliProfileSelection{Profile: cliSurfaceCanonical}, nil
	})
	return err
}
func TestCanonicalRetainedBatchRootSaveOnceAndWithdrawnSupportRecovery(t *testing.T) {
	isolateBeadsDirForTest(t)
	canonicalRetainedIsolateCredentials(t)
	temporary, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	journal := filepath.Join(temporary, "retained-invocations")
	file := filepath.Join(temporary, "input.json")
	if err := os.WriteFile(file, []byte(retainedBatchAllocatedInput), 0600); err != nil {
		t.Fatal(err)
	}
	support := true
	posts, directories := 0, 0
	savedPath := ""
	var originalBody []byte
	originalKey := ""
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/s/":
			w.Header().Set("Link", "<"+server.URL+"/description>; rel=\"service-desc\"; type=\""+canonicalRetainedMedia+"\"")
			w.WriteHeader(204)
		case "/description":
			w.Header().Set("Content-Type", canonicalRetainedMedia)
			capability := ""
			if support {
				capability = `,"capabilities":{"` + canonicalRetainedBatchCapability + `":{"1":{"operation":"` + server.URL + `/s/batch","operations":["createBead","deleteBead"],"maximumOperations":2,"maximumResultBytes":4096}}}`
			}
			io.WriteString(w, `{"bdpVersion":"0.1","profile":"write","scope":"`+server.URL+`/s/","beads":"`+server.URL+`/s/beads","links":"`+server.URL+`/s/links","events":"`+server.URL+`/s/events","aliases":"`+server.URL+`/s/aliases","operations":"`+server.URL+`/s/operations"`+capability+`}`)
		case "/s/operations":
			directories++
			http.Error(w, "batch is not a ninth directory member", 500)
		case "/s/batch":
			posts++
			if r.Method != "POST" || r.Header.Get("Accept") != canonicalRetainedMedia || r.Header.Get("Content-Type") != canonicalRetainedMedia {
				t.Error("wrong native request")
			}
			entries, err := os.ReadDir(journal)
			if err != nil || len(entries) != 1 {
				t.Error("batch POST before private save")
				http.Error(w, "custody", 500)
				return
			}
			savedPath = filepath.Join(journal, entries[0].Name())
			saved, err := loadCanonicalRetainedInvocation(journal, savedPath)
			if err != nil {
				t.Error(err)
			}
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
			}
			if !bytes.Equal(raw, saved.record.Request.Body) || r.Header.Get("Idempotency-Key") != saved.record.Request.Key || saved.record.Request.Operation != retainedAtomicBatch {
				t.Error("batch not exact saved carrier")
			}
			if posts == 1 {
				originalBody = bytes.Clone(raw)
				originalKey = r.Header.Get("Idempotency-Key")
			} else if !bytes.Equal(raw, originalBody) || r.Header.Get("Idempotency-Key") != originalKey {
				t.Error("recovery substituted fresh body/key")
			}
			w.Header().Set("Content-Type", canonicalRetainedMedia)
			io.WriteString(w, retainedBatchAllocatedResult)
		default:
			http.Error(w, "unexpected request", 500)
		}
	}))
	defer server.Close()
	options := canonicalRetainedTestBindings(server.Client())
	options.Owner = "local-batch-test-owner"
	options.JournalRoot = journal
	root, output := retainedBatchRoot(t, options)
	if err := retainedBatchExecute(root, []string{"batch", "--scope", server.URL + "/s/", "--input-format", "json", "--batch-version", "1", "--file", file, "--dry-run-mode", "bdp"}); err != nil {
		t.Fatalf("native preview: %v", err)
	}
	if posts != 0 || directories != 0 || !strings.Contains(output.String(), "capabilityVersion") || strings.Contains(output.String(), "idempotencyKey") {
		t.Fatal("preview submitted/reserved key or changed carrier")
	}
	if _, err := os.Stat(journal); !os.IsNotExist(err) {
		t.Fatal("preview created mutation journal")
	}
	root, output = retainedBatchRoot(t, options)
	if err := retainedBatchExecute(root, []string{"batch", "--scope", server.URL + "/s/", "--input-format", "json", "--batch-version", "1", "--file", file, "--output-format", "json"}); err != nil {
		t.Fatalf("native apply: %v", err)
	}
	if posts != 1 || directories != 0 || savedPath == "" || !strings.Contains(output.String(), "txn-original") {
		t.Fatalf("native route missing: %d/%d %s", posts, directories, output.String())
	}
	support = false
	root, output = retainedBatchRoot(t, options)
	if err := retainedBatchExecute(root, []string{"batch", "--scope", server.URL + "/s/", "--retry-file", savedPath, "--output-format", "json"}); err != nil {
		t.Fatalf("withdrawn support blocked original recovery: %v", err)
	}
	if posts != 2 || directories != 0 || !strings.Contains(output.String(), "txn-original") {
		t.Fatal("withdrawn support recovery did not preserve original result")
	}
	root, _ = retainedBatchRoot(t, options)
	if err := retainedBatchExecute(root, []string{"batch", "--scope", server.URL + "/s/", "--input-format", "json", "--batch-version", "1", "--file", file}); err == nil || posts != 2 {
		t.Fatal("withdrawn support admitted a fresh batch")
	}
	options.Owner = "different-owner"
	root, _ = retainedBatchRoot(t, options)
	if err := retainedBatchExecute(root, []string{"batch", "--scope", server.URL + "/s/", "--retry-file", savedPath}); err == nil || posts != 2 {
		t.Fatal("recovery rebound local owner")
	}
}

type retainedBatchCountingInput struct {
	source io.Reader
	reads  int
}

func (r *retainedBatchCountingInput) Read(p []byte) (int, error) { r.reads++; return r.source.Read(p) }
func TestCanonicalRetainedBatchOwnedNavigationAndPreInputRefusal(t *testing.T) {
	isolateBeadsDirForTest(t)
	canonicalRetainedIsolateCredentials(t)
	for _, target := range []string{"https://other.test/s/batch", "https://data.test/s-other/batch", "https://data.test/s/%2e%2e/batch"} {
		if canonicalRetainedBatchOwnedTarget("https://data.test/s/", target) {
			t.Fatalf("escaped owning Scope: %s", target)
		}
	}
	temporary, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	journal := filepath.Join(temporary, "journal")
	requests, posts := 0, 0
	options := canonicalRetainedTestBindings(&http.Client{Transport: canonicalRetainedRoundTrip(func(r *http.Request) (*http.Response, error) {
		requests++
		headers := http.Header{"Content-Type": []string{canonicalRetainedMedia}}
		status := 200
		body := ""
		switch r.URL.Path {
		case "/s/":
			status = 204
			headers.Set("Link", `<https://data.test/description>; rel="service-desc"; type="`+canonicalRetainedMedia+`"`)
		case "/description":
			body = `{"bdpVersion":"0.1","profile":"write","scope":"https://data.test/s/","beads":"https://data.test/s/beads","links":"https://data.test/s/links","events":"https://data.test/s/events","aliases":"https://data.test/s/aliases","operations":"https://data.test/s/operations","capabilities":{"` + canonicalRetainedBatchCapability + `":{"1":{"operation":"https://data.test/s/batch","operations":["createBead","deleteBead"]}}}}`
		default:
			posts++
			return nil, errors.New("unexpected submission")
		}
		return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}, nil
	})})
	options.Owner = "local-control-owner"
	options.JournalRoot = journal
	previousReporter := graphFailureReporter
	defer func() { graphFailureReporter = previousReporter }()
	code, message := "", ""
	graphFailureReporter = func(c, m string, status int) error { code, message = c, m; return &exitError{Code: status} }
	controls := []struct {
		args    []string
		message string
	}{
		{[]string{"--input-format", "jsonl", "--batch-version", "1", "--file", "-"}, "requires --input-format json"},
		{[]string{"--input-format", "json", "--batch-version", "1", "--file", "-", "--dry-run-mode", "batch"}, "preview requires bdp mode"},
		{[]string{"--retry-file", "saved", "--idempotency-key", "new-key"}, "recovery does not accept fresh --idempotency-key"},
		{[]string{"--input-format", "json", "--batch-version", "1", "--file", "-", "--idempotency-key", ""}, "explicit retained idempotency key"},
	}
	for _, control := range controls {
		code, message = "", ""
		beforeRequests := requests
		root, _ := retainedBatchRoot(t, options)
		input := &retainedBatchCountingInput{source: strings.NewReader(retainedBatchAllocatedInput)}
		root.SetIn(input)
		err := retainedBatchExecute(root, append([]string{"batch", "--scope", "https://data.test/s/"}, control.args...))
		var refusal *exitError
		if err == nil || !errors.As(err, &refusal) || refusal.Code != 2 || code != "invalid_selector" || !strings.Contains(message, control.message) {
			t.Fatalf("wrong refusal: err=%v code=%q message=%q", err, code, message)
		}
		if input.reads != 0 || requests != beforeRequests || posts != 0 {
			t.Fatal("syntax refusal consumed input or performed discovery/submission")
		}
		if _, err := os.Stat(journal); !os.IsNotExist(err) {
			t.Fatal("syntax refusal touched journal")
		}
	}
	// Positive control uses the same valid Scope, bindings and input. Unrelated
	// fixture/config/transport failure cannot satisfy the negative controls.
	code, message = "", ""
	root, _ := retainedBatchRoot(t, options)
	input := &retainedBatchCountingInput{source: strings.NewReader(retainedBatchAllocatedInput)}
	root.SetIn(input)
	if err := retainedBatchExecute(root, []string{"batch", "--scope", "https://data.test/s/", "--input-format", "json", "--batch-version", "1", "--file", "-", "--dry-run-mode", "bdp"}); err != nil || code != "" || input.reads == 0 || requests != 2 || posts != 0 {
		t.Fatalf("positive fixture did not reach preview: %v code=%s reads=%d requests=%d", err, code, input.reads, requests)
	}
	if _, err := os.Stat(journal); !os.IsNotExist(err) {
		t.Fatal("positive preview touched journal")
	}
}

func TestCanonicalRetainedBatchAdvertisedSubsetAndWholeOperationBound(t *testing.T) {
	members, err := canonicalRetainedBatchMembers([]byte(retainedBatchAllocatedInput))
	if err != nil {
		t.Fatal(err)
	}
	binding := canonicalRetainedBatchBinding{operations: map[canonicalRetainedOperation]bool{retainedCreateBead: true, retainedDeleteBead: true}, maximumOperations: big.NewRat(2, 1)}
	if err := binding.admit(members); err != nil {
		t.Fatal(err)
	}
	binding.maximumOperations = big.NewRat(1, 1)
	if err := binding.admit(members); err == nil {
		t.Fatal("batch exceeded advertised whole-group bound")
	}
	binding.maximumOperations = nil
	delete(binding.operations, retainedDeleteBead)
	if err := binding.admit(members); err == nil {
		t.Fatal("batch ignored disabled member support")
	}
}
func TestCanonicalRetainedBatchDeclaredOwnerRowsUseFinalPublishedRevisions(t *testing.T) {
	request := `{"capabilityVersion":"1","operations":[{"operation":"updateLink","link":"https://data.test/s/l","properties":{"n":2}}]}`
	result := `{"capabilityVersion":"1","changed":true,"transaction":"tx","results":[{"operationIndex":0,"operation":"updateLink","id":"https://data.test/s/l"}],"resources":[{"id":"https://data.test/s/a","resourceKind":"bead","outcome":"unchanged","changed":false,"resource":{"id":"https://data.test/s/a","type":"https://types.test/T","revision":"a-final","properties":{}}},{"id":"https://data.test/s/l","resourceKind":"link","outcome":"updated","changed":true,"resource":{"id":"https://data.test/s/l","type":"https://types.test/L","revision":"l-final","source":{"uri":"https://data.test/s/a","revision":"a-final"},"target":{"uri":"https://data.test/s/b","revision":"b-old"},"properties":{"n":2}},"ownerRevisions":[{"id":"https://data.test/s/a","revision":"a-final"}]}]}`
	if err := canonicalRetainedBatchResult([]byte(result), []byte(request), "https://data.test/s/"); err != nil {
		t.Fatal(err)
	}
	created := strings.Replace(result, `"outcome":"unchanged","changed":false`, `"outcome":"created","changed":true`, 1)
	if err := canonicalRetainedBatchResult([]byte(created), []byte(request), "https://data.test/s/"); err == nil {
		t.Fatal("indirect owner creation without Bead member accepted")
	}
	ownerLive := `"outcome":"unchanged","changed":false,"resource":{"id":"https://data.test/s/a","type":"https://types.test/T","revision":"a-final","properties":{}}`
	ownerDeleted := `"outcome":"deleted","changed":true,"deleted":{"resourceKind":"bead","resource":{"id":"https://data.test/s/a","type":"https://types.test/T","revision":"a-final"}}`
	deleted := strings.Replace(result, ownerLive, ownerDeleted, 1)
	if deleted == result {
		t.Fatal("deleted-owner literal control was not installed")
	}
	if err := canonicalRetainedBatchResult([]byte(deleted), []byte(request), "https://data.test/s/"); err == nil {
		t.Fatal("indirect owner deletion without Bead member accepted")
	}
	// Replace only the owner evidence, leaving the captured source row unchanged.
	invalid := strings.Replace(result, `"ownerRevisions":[{"id":"https://data.test/s/a","revision":"a-final"}]`, `"ownerRevisions":[{"id":"https://data.test/s/a","revision":"staged"}]`, 1)
	if err := canonicalRetainedBatchResult([]byte(invalid), []byte(request), "https://data.test/s/"); err == nil {
		t.Fatal("owner reference substituted an intermediate revision")
	}
	invalid = strings.Replace(result, `"ownerRevisions":[{"id":"https://data.test/s/a","revision":"a-final"}]`, `"ownerRevisions":[{"id":"https://data.test/s/missing","revision":"a-final"}]`, 1)
	if err := canonicalRetainedBatchResult([]byte(invalid), []byte(request), "https://data.test/s/"); err == nil {
		t.Fatal("declared owner lacked corresponding final row")
	}
}

func TestCanonicalRetainedBatchAssignedCreateIDCorrelation(t *testing.T) {
	request := `{"capabilityVersion":"1","operations":[{"operation":"createBead","id":"assigned","type":"https://types.test/T"}]}`
	result := `{"capabilityVersion":"1","changed":true,"transaction":"tx","results":[{"operationIndex":0,"operation":"createBead","id":"https://data.test/s/assigned"}],"resources":[{"id":"https://data.test/s/assigned","resourceKind":"bead","outcome":"created","changed":true,"resource":{"id":"https://data.test/s/assigned","type":"https://types.test/T","revision":"r","properties":{}}}]}`
	if err := canonicalRetainedBatchResult([]byte(result), []byte(request), "https://data.test/s/"); err != nil {
		t.Fatal(err)
	}
	// Both mapping and row are internally consistent but contradict the authored
	// relative creation ID. Agreement of returned fields alone is insufficient.
	wrong := strings.ReplaceAll(result, "https://data.test/s/assigned", "https://data.test/s/other")
	if err := canonicalRetainedBatchResult([]byte(wrong), []byte(request), "https://data.test/s/"); err == nil {
		t.Fatal("wrong scoped caller-assigned identity accepted")
	}
	absolute := strings.Replace(request, `"id":"assigned"`, `"id":"https://data.test/s/assigned"`, 1)
	if err := canonicalRetainedBatchResult([]byte(result), []byte(absolute), "https://data.test/s/"); err != nil {
		t.Fatal(err)
	}
	normalized := strings.ReplaceAll(retainedBatchNoopInput, "https://data.test/s/a", "https://data.test/s/%61")
	reply := canonicalRetainedReply{Status: 200, Headers: http.Header{"Content-Type": []string{canonicalRetainedMedia}}, NativeBody: []byte(retainedBatchNoopResult), Sent: true}
	reply.validateBatch([]byte(normalized), "https://data.test/s/")
	if !reply.Checked || reply.ParseErr != nil || !bytes.Equal(reply.Mutation, []byte(retainedBatchNoopResult)) {
		t.Fatal("normalized literal Bead Reference result refused")
	}

}

func TestCanonicalRetainedBatchFeatureProblemRequestCorrelation(t *testing.T) {
	capability := `{"type":"https://github.com/gastownhall/bdp/problems/request","status":422,"code":"capability-not-supported","retry":"after-state-change","capability":"https://github.com/gastownhall/bdp/capabilities/atomic-batch","capabilityVersion":"1","pointer":"/capabilityVersion"}`
	operation := `{"type":"https://github.com/gastownhall/bdp/problems/request","status":422,"code":"operation-not-supported","retry":"after-state-change","operation":"updateBead","pointer":"/operations/1/operation"}`
	for _, raw := range []string{capability, operation} {
		reply := canonicalRetainedReply{Status: 422, Headers: http.Header{"Content-Type": []string{"application/problem+json"}}, NativeBody: []byte(raw), Sent: true}
		reply.validateBatch([]byte(retainedBatchNoopInput), "https://data.test/s/")
		if !reply.Checked || reply.ParseErr != nil || reply.OperationErr == nil || !bytes.Equal(reply.Problem, []byte(raw)) {
			t.Fatal("matching feature Problem lost")
		}
	}
	for _, raw := range []string{strings.Replace(capability, `"capabilityVersion":"1"`, `"capabilityVersion":"next"`, 1), strings.Replace(operation, "/operations/1/operation", "/operations/2/operation", 1), strings.Replace(operation, `"operation":"updateBead"`, `"operation":"deleteBead"`, 1)} {
		reply := canonicalRetainedReply{Status: 422, Headers: http.Header{"Content-Type": []string{"application/problem+json"}}, NativeBody: []byte(raw), Sent: true}
		reply.validateBatch([]byte(retainedBatchNoopInput), "https://data.test/s/")
		if reply.Checked || reply.ParseErr == nil || len(reply.Problem) != 0 || !bytes.Equal(reply.NativeBody, []byte(raw)) {
			t.Fatal("unrelated feature Problem marked checked or original bytes lost")
		}
	}
}

func TestCanonicalRetainedBatchFailureExceedsSuccessBudgetPreservesCauses(t *testing.T) {
	closeFailure := errors.New("large Problem Close")
	outputFailure := errors.New("large Problem output")
	// Selected ordinary validation-failed allows a native diagnostic exceeding a
	// 1024 success-result budget. It remains below ordinary transport limits.
	native := `{"type":"https://github.com/gastownhall/bdp/problems/validation","status":422,"code":"validation-failed","retry":"never","diagnostics":[{"message":"` + strings.Repeat("x", 1100) + `"}]}`
	body := &canonicalRetainedBody{Reader: strings.NewReader(native), closeErr: closeFailure}
	calls := 0
	options := canonicalRetainedTestBindings(&http.Client{Transport: canonicalRetainedRoundTrip(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 422, Header: http.Header{"Content-Type": []string{"application/problem+json"}}, Body: body}, nil
	})})
	request, _ := http.NewRequestWithContext(context.Background(), "POST", "https://data.test/s/batch", strings.NewReader(retainedBatchNoopInput))
	reply := options.exchange(request)
	reply.validateBatch([]byte(retainedBatchNoopInput), "https://data.test/s/")
	canonicalRetainedBatchResultBound(&reply, big.NewRat(1024, 1))
	writer := &canonicalRetainedFailWriter{err: outputFailure}
	output, err := newCanonicalMutationOutput(canonicalMutationOutputRetainedBatch, clioutput.MutationDescriptor{BDPSpec: canonicalRetainedContract, Required: []string{"capabilityVersion", "changed", "results", "resources"}}, clioutput.Options{Formats: []string{"json"}}, writer)
	if err != nil {
		t.Fatal(err)
	}
	result := output.Complete(reply, "", reply.OperationErr, reply.CloseErr)
	if !reply.Checked || reply.LimitErr != nil || reply.ParseErr != nil || result.Disposition != canonicalMutationOutputFailed || result.OperationErr == nil || !errors.Is(result.CloseErr, closeFailure) || !errors.Is(result.OutputErr, outputFailure) || !bytes.Equal(reply.Problem, []byte(native)) || calls != 1 || body.closes != 1 || writer.calls != 1 {
		t.Fatalf("native failure reclassified or lost independent causes: %#v", result)
	}
	output.Complete(reply, "", nil, nil)
	if writer.calls != 1 || calls != 1 {
		t.Fatal("large Problem retried")
	}
	success := canonicalRetainedReply{Status: 200, Sent: true, Checked: true, NativeBody: []byte(retainedBatchNoopResult), Mutation: []byte(retainedBatchNoopResult)}
	canonicalRetainedBatchResultBound(&success, big.NewRat(1, 1))
	if success.Checked || success.LimitErr == nil || len(success.Mutation) != 0 || !bytes.Equal(success.NativeBody, []byte(retainedBatchNoopResult)) {
		t.Fatal("successful carrier ignored advertised bound or lost original bytes")
	}
}

func TestCanonicalRetainedBatchAbsoluteAliasOriginalExecutionRecovery(t *testing.T) {
	isolateBeadsDirForTest(t)
	canonicalRetainedIsolateCredentials(t)
	temporary, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	journal := filepath.Join(temporary, "journal")
	file := filepath.Join(temporary, "alias-input.json")
	posts, aliasReads := 0, 0
	savedPath := ""
	originalBody := ""
	originalKey := ""
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", canonicalRetainedMedia)
		switch r.URL.Path {
		case "/s/":
			w.Header().Set("Link", "<"+server.URL+"/description>; rel=\"service-desc\"; type=\""+canonicalRetainedMedia+"\"")
			w.WriteHeader(204)
		case "/description":
			io.WriteString(w, `{"bdpVersion":"0.1","profile":"write","scope":"`+server.URL+`/s/","beads":"`+server.URL+`/s/beads","links":"`+server.URL+`/s/links","events":"`+server.URL+`/s/events","aliases":"`+server.URL+`/s/aliases","operations":"`+server.URL+`/s/operations","capabilities":{"`+canonicalRetainedBatchCapability+`":{"1":{"operation":"`+server.URL+`/s/batch","operations":["updateBead"]}}}}`)
		case "/s/batch":
			posts++
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
			}
			entries, err := os.ReadDir(journal)
			if err != nil || len(entries) != 1 {
				t.Error("alias submission lacked custody")
				http.Error(w, "custody", 500)
				return
			}
			savedPath = filepath.Join(journal, entries[0].Name())
			if posts == 1 {
				originalBody = string(raw)
				originalKey = r.Header.Get("Idempotency-Key")
			} else if string(raw) != originalBody || r.Header.Get("Idempotency-Key") != originalKey {
				t.Error("alias recovery changed original request")
			}
			// This is a fixed captured mapping A→B. A later alias state must never be
			// consulted by this client; real service witness retention is separate.
			id := server.URL + "/s/beads/original"
			io.WriteString(w, `{"capabilityVersion":"1","changed":false,"results":[{"operationIndex":0,"operation":"updateBead","id":"`+id+`"}],"resources":[{"id":"`+id+`","resourceKind":"bead","outcome":"unchanged","changed":false,"resource":{"id":"`+id+`","type":"https://types.test/T","revision":"original-r","properties":{}}}]}`)
		default:
			aliasReads++
			http.Error(w, "client must not refresh alias", 500)
		}
	}))
	defer server.Close()
	input := `{"capabilityVersion":"1","operations":[{"operation":"updateBead","bead":"` + server.URL + `/s/aliases/current","properties":{}}]}`
	if err := os.WriteFile(file, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	options := canonicalRetainedTestBindings(server.Client())
	options.Owner = "local-alias-control"
	options.JournalRoot = journal
	root, output := retainedBatchRoot(t, options)
	if err := retainedBatchExecute(root, []string{"batch", "--scope", server.URL + "/s/", "--input-format", "json", "--batch-version", "1", "--file", file, "--output-format", "json"}); err != nil {
		t.Fatalf("absolute alias result refused: %v", err)
	}
	if posts != 1 || aliasReads != 0 || !strings.Contains(output.String(), "original-r") {
		t.Fatal("alias execution reread or lost resolved identity")
	}
	root, output = retainedBatchRoot(t, options)
	if err := retainedBatchExecute(root, []string{"batch", "--scope", server.URL + "/s/", "--retry-file", savedPath, "--output-format", "json"}); err != nil {
		t.Fatalf("original alias recovery refused: %v", err)
	}
	if posts != 2 || aliasReads != 0 || !strings.Contains(output.String(), "original-r") {
		t.Fatal("alias recovery reread or lost original result")
	}
}

func TestCanonicalRetainedBatchLinkSubjectResolvedIdentityCorrelation(t *testing.T) {
	allocated := `{"capabilityVersion":"1","operations":[{"operation":"createLink","type":"https://types.test/L","source":"https://data.test/s/a","target":"https://data.test/s/b"}]}`
	created := `{"capabilityVersion":"1","changed":true,"transaction":"tx","results":[{"operationIndex":0,"operation":"createLink","id":"https://data.test/s/links/generated"}],"resources":[{"id":"https://data.test/s/links/generated","resourceKind":"link","outcome":"created","changed":true,"resource":{"id":"https://data.test/s/links/generated","type":"https://types.test/L","revision":"new","source":"https://data.test/s/a","target":"https://data.test/s/b","properties":{}}}]}`
	if err := canonicalRetainedBatchResult([]byte(created), []byte(allocated), "https://data.test/s/"); err != nil {
		t.Fatalf("omitted Link create ID mistaken for subject: %v", err)
	}
	// Link subjects use canonical local URI comparison, not Bead aliases. Both
	// returned mapping and row can agree with each other while naming wrong Link.
	deletion := `{"capabilityVersion":"1","operations":[{"operation":"deleteLink","link":"links/%6c"}]}`
	result := `{"capabilityVersion":"1","changed":true,"transaction":"tx","results":[{"operationIndex":0,"operation":"deleteLink","id":"https://data.test/s/links/l"}],"resources":[{"id":"https://data.test/s/links/l","resourceKind":"link","outcome":"deleted","changed":true,"deleted":{"resourceKind":"link","resource":{"id":"https://data.test/s/links/l","type":"https://types.test/L","revision":"last-published"}}}]}`
	for _, request := range []string{deletion, strings.Replace(deletion, `"links/%6c"`, `"https://data.test/s/links/%6c"`, 1), strings.Replace(deletion, `"links/%6c"`, `"https://data.test/s/links/l"`, 1)} {
		if err := canonicalRetainedBatchResult([]byte(result), []byte(request), "https://data.test/s/"); err != nil {
			t.Fatalf("normalized Link subject refused: %v", err)
		}
		wrong := strings.ReplaceAll(result, "https://data.test/s/links/l", "https://data.test/s/links/other")
		if err := canonicalRetainedBatchResult([]byte(wrong), []byte(request), "https://data.test/s/"); err == nil {
			t.Fatal("internally consistent wrong Link accepted")
		}
	}
	for _, pair := range [][2]string{{"links/a%2fb", "https://data.test/s/links/a%2Fb"}, {"links/%c3%a9", "https://data.test/s/links/%C3%A9"}, {"links/a:b@c", "https://data.test/s/links/a:b@c"}} {
		got, err := canonicalRetainedBatchLinkIdentity("https://data.test/s/", pair[0])
		if err != nil || got != pair[1] {
			t.Fatalf("once-decoded Link URI changed: %q %v", got, err)
		}
	}
	update := `{"capabilityVersion":"1","operations":[{"operation":"updateLink","link":"https://data.test/s/links/l","properties":{}}]}`
	live := `{"capabilityVersion":"1","changed":false,"results":[{"operationIndex":0,"operation":"updateLink","id":"https://data.test/s/links/l"}],"resources":[{"id":"https://data.test/s/links/l","resourceKind":"link","outcome":"unchanged","changed":false,"resource":{"id":"https://data.test/s/links/l","type":"https://types.test/L","revision":"original","source":"https://data.test/s/a","target":"https://data.test/s/b","properties":{}}}]}`
	if err := canonicalRetainedBatchResult([]byte(live), []byte(update), "https://data.test/s/"); err != nil {
		t.Fatal(err)
	}
	wrong := strings.ReplaceAll(live, "https://data.test/s/links/l", "https://data.test/s/links/wrong")
	if err := canonicalRetainedBatchResult([]byte(wrong), []byte(update), "https://data.test/s/"); err == nil {
		t.Fatal("wrong updated Link result accepted")
	}
}

func TestCanonicalRetainedBatchScopeSlashBoundaryIncludingSavedAdmission(t *testing.T) {
	scope := "https://x.test/tenant/"
	for _, target := range []string{"https://x.test/tenant/", "https://x.test/tenant/commands/batch"} {
		if !canonicalRetainedBatchOwnedTarget(scope, target) {
			t.Fatalf("owning Scope target refused: %s", target)
		}
	}
	for _, target := range []string{"https://x.test/tenant", "https://x.test/tenant-other/batch", "https://x.test/tenant/%2e%2e/batch", "https://other.test/tenant/batch"} {
		if canonicalRetainedBatchOwnedTarget(scope, target) {
			t.Fatalf("target escaped exact slash-bearing Scope: %s", target)
		}
		record := canonicalRetainedJournalRecord{Version: 1, Owner: "local-fixture", Provider: "", Request: canonicalRetainedPrepared{Scope: scope, Contract: canonicalRetainedContract, Operation: retainedAtomicBatch, Method: "POST", Target: target, Key: "original-key", Body: []byte(retainedBatchNoopInput)}}
		if _, err := canonicalRetainedJournalBytes(record); err == nil {
			t.Fatalf("durable invocation admitted escaped batch target: %s", target)
		}
	}
}
