package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
)

func canonicalRetainedIsolateCredentials(t *testing.T) {
	t.Helper()
	t.Setenv("BEADS_HTTP_TOKEN", "")
	t.Setenv("BEADS_HTTP_TOKEN_COMMAND", "")
	t.Setenv("BEADS_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "absent"))
}
func canonicalRetainedTestBindings(client *http.Client) canonicalRetainedBindings {
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return canonicalRetainedBindings{Client: client, Deadline: time.Now().Add(10 * time.Second)}
}

// This test exercises the real prefix/decorator/Cobra/Args/root-admission caller
// boundary using a fresh initialized tree, never the package-global user tree.
// The separate installed production CLI→SQLite journey is still required.
func TestCanonicalRetainedRootAbsoluteReadCoreOnlyNoMutationState(t *testing.T) {
	isolateBeadsDirForTest(t)
	canonicalRetainedIsolateCredentials(t)
	t.Setenv("BDP_SCOPE_URL", "not-a-scope")
	var server *httptest.Server
	var reads, descriptions, other int
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/resource":
			reads++
			w.Header().Set("Content-Type", canonicalRetainedMedia)
			w.Header().Set("BDP-Resource-Kind", "bead")
			w.Header().Set("Link", "<"+server.URL+"/description>; rel=\"service-desc\"; type=\""+canonicalRetainedMedia+"\"")
			io.WriteString(w, `{"id":"`+server.URL+`/resource","type":"https://types.test/task","revision":"r:exact","properties":{"n":9007199254740993},"metadata":{}}`)
		case "/description":
			descriptions++
			w.Header().Set("Content-Type", canonicalRetainedMedia)
			io.WriteString(w, `{"bdpVersion":"0.1","profile":"core","scope":"`+server.URL+`/scope","beads":"`+server.URL+`/beads","links":"`+server.URL+`/links","events":"`+server.URL+`/events"}`)
		default:
			other++
			http.Error(w, "no mutation directory", 500)
		}
	}))
	defer server.Close()
	options := canonicalRetainedTestBindings(server.Client())
	// Intentionally unusable mutation context: absolute pointRead must not touch it.
	options.JournalRoot = "/unavailable/no-journal"
	options.Owner = ""
	root := &cobra.Command{Use: "bd"}
	var output, diagnostic bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&diagnostic)
	root.SetContext(context.WithValue(context.Background(), canonicalRetainedBindingsKey, options))
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error { _, err := admitCanonicalRetained(cmd); return err }
	show := &cobra.Command{Use: "show", Args: func(*cobra.Command, []string) error { return errors.New("legacy Args must not run") }, RunE: runCanonicalRetained}
	registerCanonicalOutputFlags(show)
	root.AddCommand(show)
	_, err := executeCLIProfileRootWith(root, []string{"show", server.URL + "/resource"}, func(*string) (cliProfileSelection, error) {
		return cliProfileSelection{Profile: cliSurfaceCanonical}, nil
	})
	if err != nil {
		t.Fatalf("absolute Core-only root read: %v %s", err, diagnostic.String())
	}
	if reads != 1 || descriptions != 1 || other != 0 {
		t.Fatalf("unexpected read/navigation: %d/%d/%d", reads, descriptions, other)
	}
	if !strings.Contains(output.String(), "9007199254740993") || strings.Contains(output.String(), "9007199254740992") {
		t.Fatalf("native value changed: %s", output.String())
	}
}

func TestCanonicalRetainedSaveBeforePostAndExplicitOriginalRetry(t *testing.T) {
	canonicalRetainedIsolateCredentials(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	journal := filepath.Join(root, "retained-invocations")
	var posts int
	var firstBody []byte
	var firstKey string
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		entries, err := os.ReadDir(journal)
		if err != nil || len(entries) != 1 {
			t.Error("POST preceded durable custody")
			http.Error(w, "custody missing", 500)
			return
		}
		saved, err := loadCanonicalRetainedInvocation(journal, filepath.Join(journal, entries[0].Name()))
		if err != nil {
			t.Errorf("unreadable custody at POST: %v", err)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(body, saved.record.Request.Body) || r.Header.Get("Idempotency-Key") != saved.record.Request.Key {
			t.Error("POST differs from saved invocation")
		}
		if posts == 1 {
			firstBody = bytes.Clone(body)
			firstKey = r.Header.Get("Idempotency-Key")
		} else if !bytes.Equal(body, firstBody) || r.Header.Get("Idempotency-Key") != firstKey {
			t.Error("retry changed meaning/key")
		}
		w.Header().Set("Content-Type", canonicalRetainedMedia)
		io.WriteString(w, `{"outcome":"created","changed":true,"resource":{"id":"https://data.test/s/a","type":"https://types.test/task","revision":"r1","properties":{}}}`)
	}))
	defer server.Close()
	options := canonicalRetainedTestBindings(server.Client())
	options.Owner = "local-test-owner"
	options.JournalRoot = journal
	binding := canonicalRetainedWriteBinding{options: options, scope: "https://data.test/s/", targets: map[string]string{"createBead": server.URL}}
	request := canonicalRetainedPrepared{Scope: binding.scope, Contract: canonicalRetainedContract, Operation: retainedCreateBead, Method: "POST", Target: server.URL, Body: json.RawMessage(`{"type":"https://types.test/task"}`)}
	reply, path := binding.apply(context.Background(), request)
	if reply.OperationErr != nil || reply.ParseErr != nil || reply.CloseErr != nil || !reply.Checked || path == "" || posts != 1 {
		t.Fatalf("apply: %#v path=%s posts=%d", reply, path, posts)
	}
	again, _ := binding.retry(context.Background(), path, retainedCreateBead)
	if again.OperationErr != nil || again.ParseErr != nil || posts != 2 {
		t.Fatalf("explicit retry: %#v posts=%d", again, posts)
	}
	changed := binding
	changed.options.Owner = "different-local-owner"
	refusal, _ := changed.retry(context.Background(), path, retainedCreateBead)
	if refusal.OperationErr == nil || posts != 2 {
		t.Fatal("recovery silently rebound local owner")
	}
	wrong, _ := binding.retry(context.Background(), path, retainedDeleteBead)
	if wrong.OperationErr == nil || posts != 2 {
		t.Fatal("recovery ignored selected operation")
	}
}

type canonicalRetainedRoundTrip func(*http.Request) (*http.Response, error)

func (f canonicalRetainedRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type canonicalRetainedBody struct {
	io.Reader
	closeErr error
	closes   int
}

func (body *canonicalRetainedBody) Close() error { body.closes++; return body.closeErr }

type canonicalRetainedFailWriter struct {
	calls int
	err   error
}

func (writer *canonicalRetainedFailWriter) Write([]byte) (int, error) {
	writer.calls++
	return 0, writer.err
}

func TestCanonicalRetainedOriginalSuccessCloseAndOutputIndependent(t *testing.T) {
	closeFailure := errors.New("response close failed")
	outputFailure := errors.New("output write failed")
	native := `{"outcome":"updated","changed":false,"resource":{"id":"https://data.test/s/a","type":"https://types.test/task","revision":"unchanged","properties":{}}}`
	body := &canonicalRetainedBody{Reader: strings.NewReader(native), closeErr: closeFailure}
	calls := 0
	options := canonicalRetainedTestBindings(&http.Client{Transport: canonicalRetainedRoundTrip(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{canonicalRetainedMedia}}, Body: body}, nil
	})})
	request, _ := http.NewRequestWithContext(context.Background(), "POST", "https://commands.test/update", strings.NewReader(`{}`))
	reply := options.exchange(context.Background(), request)
	reply.validateMutation(retainedUpdateBead)
	writer := &canonicalRetainedFailWriter{err: outputFailure}
	output, err := newCanonicalMutationOutput(canonicalMutationOutputRetained, clioutput.MutationDescriptor{BDPSpec: canonicalRetainedContract, Required: []string{"outcome", "changed"}}, clioutput.Options{Formats: []string{"json"}}, writer)
	if err != nil {
		t.Fatal(err)
	}
	result := output.Complete(reply, "", reply.OperationErr, reply.CloseErr)
	if result.Disposition != canonicalMutationOutputReturned || result.OperationErr != nil || !errors.Is(result.CloseErr, closeFailure) || !errors.Is(result.OutputErr, outputFailure) || !bytes.Equal(reply.NativeBody, []byte(native)) || !bytes.Equal(reply.Mutation, []byte(native)) || calls != 1 || body.closes != 1 || writer.calls != 1 {
		t.Fatalf("independent outcome/Close/output custody failed: %#v", result)
	}
	again := output.Complete(reply, "", nil, nil)
	if !errors.Is(again.OutputErr, errCanonicalMutationOutputConsumed) || writer.calls != 1 || calls != 1 {
		t.Fatal("completion reentered or retried")
	}
}
func TestCanonicalRetainedBoundedReadPreservesOverflowAndClose(t *testing.T) {
	closeFailure := errors.New("close-overflow")
	body := &canonicalRetainedBody{Reader: strings.NewReader(strings.Repeat("x", int(canonicalRetainedResponseLimit)+2)), closeErr: closeFailure}
	calls := 0
	options := canonicalRetainedTestBindings(&http.Client{Transport: canonicalRetainedRoundTrip(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: body}, nil
	})})
	request, _ := http.NewRequestWithContext(context.Background(), "GET", "https://data.test/a", nil)
	reply := options.exchange(context.Background(), request)
	if reply.OperationErr != nil || reply.ExchangeErr != nil || reply.ReadErr != nil || reply.LimitErr == nil || !errors.Is(reply.CloseErr, closeFailure) || int64(len(reply.NativeBody)) != canonicalRetainedResponseLimit+1 || body.closes != 1 || calls != 1 {
		t.Fatal("overflow lost original bounded bytes/Close or retried")
	}
}

func TestCanonicalRetainedCredentialsPerDestinationNoRedirectReplay(t *testing.T) {
	canonicalRetainedIsolateCredentials(t)
	t.Setenv("BEADS_HTTP_TOKEN", "first.example=fixture-only-token")
	first, _ := http.NewRequestWithContext(context.Background(), "GET", "https://first.example/description", nil)
	second, _ := http.NewRequestWithContext(context.Background(), "POST", "https://second.example/operation", nil)
	if _, err := canonicalRetainedAuthorize(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if _, err := canonicalRetainedAuthorize(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	if first.Header.Get("Authorization") == "" || second.Header.Get("Authorization") != "" {
		t.Fatal("credentials crossed destination authority")
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Location", "/followed")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	options, err := newCanonicalRetainedRootBindings(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	reply := options.get(context.Background(), server.URL)
	if reply.Status != 307 || calls != 1 || reply.OperationErr != nil {
		t.Fatalf("redirect followed or exchange replayed: %#v calls=%d", reply, calls)
	}
}
func TestCanonicalRetainedMutationConstructorStrictConfigBeforeJournal(t *testing.T) {
	isolateBeadsDirForTest(t)
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "native"))
	directory := filepath.Join(home, ".config", "bd")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(directory, "config.yaml")
	for _, text := range []string{"scope: {url: null}\n", "scope: {url: [not, scalar]}\n", "scope: [bad-parent]\n", "scope: [\n"} {
		if err := os.WriteFile(configPath, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := newCanonicalRetainedRootBindings(context.Background(), true); err == nil {
			t.Fatalf("invalid mutation config accepted: %s", text)
		}
		if _, err := os.Stat(filepath.Join(directory, "retained-invocations")); !os.IsNotExist(err) {
			t.Fatal("invalid config created a journal")
		}
		if _, err := newCanonicalRetainedRootBindings(context.Background(), false); err != nil {
			t.Fatalf("direct read consulted mutation config: %v", err)
		}
	}
}
func TestCanonicalRetainedCancellationNoAutomaticRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attempts := 0
	options := canonicalRetainedTestBindings(&http.Client{Transport: canonicalRetainedRoundTrip(func(r *http.Request) (*http.Response, error) { attempts++; return nil, r.Context().Err() })})
	request, _ := http.NewRequestWithContext(ctx, "POST", "https://commands.test/mutate", strings.NewReader(`{}`))
	reply := options.exchange(ctx, request)
	if !errors.Is(reply.ExchangeErr, context.Canceled) || reply.OperationErr != nil || reply.ReadErr != nil || attempts != 1 || !reply.Sent {
		t.Fatalf("cancellation was retried or hidden: %#v attempts=%d", reply, attempts)
	}
}

func TestCanonicalRetainedRootExplicitInvalidKeyPrecedesMutationConfig(t *testing.T) {
	isolateBeadsDirForTest(t)
	// A valid key would proceed to missing/default Scope configuration. Invalid
	// explicit keys must stop at the finite root gate before reaching that path.
	for _, key := range []string{"", "not valid", "!", strings.Repeat("a", 257)} {
		root := &cobra.Command{Use: "bd"}
		command := &cobra.Command{Use: "create", Annotations: map[string]string{cliSurfaceProfileAnnotation: string(cliSurfaceCanonical)}}
		command.Flags().String("idempotency-key", "", "")
		root.AddCommand(command)
		argv := []string{"create", "--idempotency-key=" + key}
		command.SetContext(context.WithValue(context.Background(), canonicalRetainedArgvKey, argv))
		if err := command.ParseFlags(argv[1:]); err != nil {
			t.Fatal(err)
		}
		handled, err := admitCanonicalRetained(command)
		if !handled || err == nil {
			t.Fatalf("explicit invalid key was treated as omitted: %q", key)
		}
		// No selected Scope is established by the rejected root invocation.
		if command.Context().Value(canonicalRetainedScopeKey) != nil {
			t.Fatal("key refusal happened after Scope acquisition")
		}
	}
}

func TestCanonicalRetainedScopeHumanLandingDiscoveredWrite(t *testing.T) {
	canonicalRetainedIsolateCredentials(t)
	var server *httptest.Server
	var paths []string
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/scope":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Link", "<"+server.URL+"/machine>; rel=\"service-desc\"; type=\""+canonicalRetainedMedia+"\"")
			io.WriteString(w, "<!doctype html><title>Human Scope landing</title>")
		case "/machine":
			w.Header().Set("Content-Type", canonicalRetainedMedia)
			io.WriteString(w, `{"bdpVersion":"0.1","profile":"write","scope":"`+server.URL+`/scope","beads":"`+server.URL+`/beads","links":"`+server.URL+`/links","events":"`+server.URL+`/events","aliases":"`+server.URL+`/aliases","operations":"`+server.URL+`/directory"}`)
		case "/directory":
			w.Header().Set("Content-Type", canonicalRetainedMedia)
			io.WriteString(w, `{"createBead":"`+server.URL+`/one","updateBead":"`+server.URL+`/two","deleteBead":"`+server.URL+`/three","createLink":"`+server.URL+`/four","updateLink":"`+server.URL+`/five","deleteLink":"`+server.URL+`/six","putAlias":"`+server.URL+`/seven","deleteAlias":"`+server.URL+`/eight"}`)
		default:
			t.Error("unexpected dispatch during navigation")
			http.Error(w, "unexpected", 500)
		}
	}))
	defer server.Close()
	binding, err := discoverCanonicalRetainedWrite(context.Background(), server.URL+"/scope", canonicalRetainedTestBindings(server.Client()))
	if err != nil || binding.targets["updateLink"] != server.URL+"/five" || strings.Join(paths, ",") != "/scope,/machine,/directory" {
		t.Fatalf("permitted human landing refused or navigation guessed: %v %v", err, paths)
	}
	closeFailure := errors.New("landing-close")
	if err := canonicalRetainedScopeLanding(canonicalRetainedReply{Status: 200, NativeBody: []byte("human"), CloseErr: closeFailure}); !errors.Is(err, closeFailure) {
		t.Fatal("landing Close error hidden")
	}
	for _, reply := range []canonicalRetainedReply{{Status: 307}, {Status: 204, NativeBody: []byte("not bodyless")}, {Status: 200, LimitErr: errors.New("overflow")}} {
		if canonicalRetainedScopeLanding(reply) == nil {
			t.Fatal("unsafe landing admitted")
		}
	}
}

type canonicalRetainedReadFailure struct{ err error }

func (reader canonicalRetainedReadFailure) Read([]byte) (int, error) { return 0, reader.err }
func TestCanonicalRetainedExchangeReadLimitCloseCompletionChannels(t *testing.T) {
	readFailure := errors.New("body-read-original")
	closeFailure := errors.New("body-close-original")
	body := &canonicalRetainedBody{Reader: canonicalRetainedReadFailure{err: readFailure}, closeErr: closeFailure}
	options := canonicalRetainedTestBindings(&http.Client{Transport: canonicalRetainedRoundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{canonicalRetainedMedia}}, Body: body}, nil
	})})
	request, _ := http.NewRequestWithContext(context.Background(), "POST", "https://commands.test/one", strings.NewReader(`{}`))
	reply := options.exchange(context.Background(), request)
	if reply.OperationErr != nil || reply.ExchangeErr != nil || reply.ReadErr != readFailure || reply.CloseErr != closeFailure || reply.LimitErr != nil || body.closes != 1 {
		t.Fatalf("read failure channel identity changed: %#v", reply)
	}
	reply.validateMutation(retainedCreateBead)
	if reply.Checked || reply.ParseErr != nil {
		t.Fatal("incomplete acquisition interpreted as native result")
	}
	exchangeFailure := errors.New("exchange-original")
	limitFailure := errors.New("limit-original")
	reply.ExchangeErr = exchangeFailure
	reply.LimitErr = limitFailure
	writer := &bytes.Buffer{}
	output, err := newCanonicalMutationOutput(canonicalMutationOutputRetained, clioutput.MutationDescriptor{BDPSpec: canonicalRetainedContract, Required: []string{"outcome", "changed"}}, clioutput.Options{Formats: []string{"json"}}, writer)
	if err != nil {
		t.Fatal(err)
	}
	result := output.Complete(reply, "", reply.OperationErr, reply.CloseErr)
	if result.OperationErr != nil || result.ExchangeErr != exchangeFailure || result.ReadErr != readFailure || result.LimitErr != limitFailure || result.CloseErr != closeFailure || result.Disposition != canonicalMutationOutputUnknown {
		t.Fatalf("completion merged original channels: %#v", result)
	}
	for _, cause := range []error{exchangeFailure, readFailure, limitFailure, closeFailure} {
		if !errors.Is(result.Err(), cause) {
			t.Fatal("final Err lost an independent cause")
		}
	}
}

func TestCanonicalRetainedProfilePreservesExplicitLeafContextAndRestores(t *testing.T) {
	root := &cobra.Command{Use: "bd"}
	type marker uint8
	const callerMarker marker = 0
	rootContext := context.WithValue(context.Background(), callerMarker, "root")
	root.SetContext(rootContext)
	deadline := time.Now().Add(time.Second)
	leafBase, deadlineCancel := context.WithDeadline(context.WithValue(context.Background(), callerMarker, "leaf"), deadline)
	defer deadlineCancel()
	leafContext, cancel := context.WithCancel(leafBase)
	cancel()
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	ran := false
	leaf := &cobra.Command{Use: "list", RunE: func(command *cobra.Command, _ []string) error {
		ran = true
		got, ok := command.Context().Deadline()
		if !ok || !got.Equal(deadline) || command.Context().Err() != context.Canceled || command.Context().Value(callerMarker) != "leaf" {
			t.Error("profile runtime replaced explicit caller deadline/cancellation/values")
		}
		argv, _ := command.Context().Value(canonicalRetainedArgvKey).([]string)
		if strings.Join(argv, ",") != "list" {
			t.Error("explicit leaf did not receive actual invocation argv")
		}
		return nil
	}}
	leaf.SetContext(leafContext)
	root.AddCommand(leaf)
	_, err := executeCLIProfileRootWith(root, []string{"list"}, func(*string) (cliProfileSelection, error) {
		return cliProfileSelection{Profile: cliSurfaceCanonical}, nil
	})
	if err != nil || !ran {
		t.Fatalf("fresh profile execution did not reach caller: %v", err)
	}
	if leaf.Context() != leafContext || root.Context() != rootContext {
		t.Fatal("profile invocation failed to restore exact contexts")
	}
}

// Runs the actual production pre-run refusal path with fresh command objects.
// Invalid captured scalars must return before metrics/defaults/signal setup.
func TestCanonicalRetainedActualRootRefusesBeforeUserConfigEffects(t *testing.T) {
	isolateBeadsDirForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "native"))
	for _, flags := range [][]string{{"--idempotency-key="}, {"--scope=https://one.example/s/", "--scope=https://two.example/s/"}} {
		root := &cobra.Command{Use: "bd"}
		command := &cobra.Command{Use: "create", Annotations: map[string]string{cliSurfaceProfileAnnotation: string(cliSurfaceCanonical)}}
		command.Flags().String("idempotency-key", "", "")
		command.Flags().String("scope", "", "")
		root.AddCommand(command)
		argv := append([]string{"create"}, flags...)
		command.SetContext(context.WithValue(context.Background(), canonicalRetainedArgvKey, argv))
		if err := command.ParseFlags(flags); err != nil {
			t.Fatal(err)
		}
		if err := rootCmd.PersistentPreRunE(command, nil); err == nil {
			t.Fatal("actual root accepted invalid scalar input")
		}
		entries, err := os.ReadDir(home)
		if err != nil || len(entries) != 0 {
			t.Fatalf("invalid input reached user configuration effects: %v %v", entries, err)
		}
	}
}

// Uses the SAME root signal-binding join invoked after setupGracefulShutdown,
// through prefix/Cobra/admission/caller/HTTP, without sending an OS signal.
func TestCanonicalRetainedRootSignalCancelsActualCallerAndPreservesDeadline(t *testing.T) {
	isolateBeadsDirForTest(t)
	canonicalRetainedIsolateCredentials(t)
	signalContext, signalCancel := context.WithCancel(context.Background())
	defer signalCancel()
	callerDeadline := time.Now().Add(time.Second)
	callerContext, callerCancel := context.WithDeadline(context.Background(), callerDeadline)
	defer callerCancel()
	type marker uint8
	callerContext = context.WithValue(callerContext, marker(0), "caller-value")
	reached := false
	canceled := false
	client := &http.Client{Transport: canonicalRetainedRoundTrip(func(request *http.Request) (*http.Response, error) {
		reached = true
		deadline, ok := request.Context().Deadline()
		if !ok || !deadline.Equal(callerDeadline) || request.Context().Value(marker(0)) != "caller-value" {
			t.Error("root signal join replaced inherited deadline/values")
		}
		signalCancel()
		<-request.Context().Done()
		canceled = errors.Is(request.Context().Err(), context.Canceled)
		return nil, request.Context().Err()
	})}
	options := canonicalRetainedTestBindings(client)
	root := &cobra.Command{Use: "bd"}
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	leafContext := context.WithValue(callerContext, canonicalRetainedBindingsKey, options)
	show := &cobra.Command{Use: "show", RunE: runCanonicalRetained}
	show.SetContext(leafContext)
	registerCanonicalOutputFlags(show)
	root.AddCommand(show)
	root.PersistentPreRunE = func(command *cobra.Command, _ []string) error {
		if err := validateCanonicalRetainedSyntax(command); err != nil {
			return err
		}
		bindCanonicalRetainedRootSignal(command, signalContext)
		_, err := admitCanonicalRetained(command)
		return err
	}
	_, err := executeCLIProfileRootWith(root, []string{"show", "https://data.example/current"}, func(*string) (cliProfileSelection, error) {
		return cliProfileSelection{Profile: cliSurfaceCanonical}, nil
	})
	if err == nil || !reached || !canceled {
		t.Fatalf("root signal did not cancel caller HTTP: err=%v reached=%v canceled=%v", err, reached, canceled)
	}
	if show.Context() != leafContext {
		t.Fatal("signal-aware invocation did not restore leaf context")
	}
}

func TestCanonicalRetainedCurrentReadMetadataContext(t *testing.T) {
	canonicalRetainedIsolateCredentials(t)
	var server *httptest.Server
	withMetadata := false
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", canonicalRetainedMedia)
		if r.URL.Path == "/machine" {
			io.WriteString(w, `{"bdpVersion":"0.1","profile":"core","scope":"`+server.URL+`/scope","beads":"`+server.URL+`/beads","links":"`+server.URL+`/links","events":"`+server.URL+`/events"}`)
			return
		}
		w.Header().Set("BDP-Resource-Kind", "bead")
		w.Header().Set("Link", "<"+server.URL+"/machine>; rel=\"service-desc\"; type=\""+canonicalRetainedMedia+"\"")
		extra := ""
		if withMetadata {
			extra = `,"metadata":{}`
		}
		io.WriteString(w, `{"id":"`+server.URL+`/current","type":"https://types.example/task","revision":"r1","properties":{}`+extra+`}`)
	}))
	defer server.Close()
	binding := canonicalRetainedReadBinding{options: canonicalRetainedTestBindings(server.Client())}
	missing := binding.readResource(context.Background(), server.URL+"/current")
	if missing.Checked || missing.ParseErr == nil || len(missing.Resource) != 0 {
		t.Fatal("current full GET without metadata was checked")
	}
	withMetadata = true
	valid := binding.readResource(context.Background(), server.URL+"/current")
	if !valid.Checked || valid.Err() != nil {
		t.Fatalf("complete current GET refused: %#v", valid)
	}
}

func TestCanonicalRetainedDirectReadSelectsProblemFamily(t *testing.T) {
	canonicalRetainedIsolateCredentials(t)
	cases := []struct {
		name       string
		status     int
		body       string
		readValid  bool
		writeValid bool
	}{
		{"write-only-code", 409, `{"type":"https://github.com/gastownhall/bdp/problems/conflict","code":"idempotency-conflict","retry":"never"}`, false, true},
		{"read-opaque-retryAfter", 404, `{"type":"https://github.com/gastownhall/bdp/problems/not-found","code":"resource-not-found","retry":"after-state-change","retryAfter":0}`, true, false},
		{"read-opaque-truncation", 404, `{"type":"https://github.com/gastownhall/bdp/problems/not-found","code":"resource-not-found","retry":"after-state-change","diagnosticsTruncated":false}`, true, false},
		{"read-opaque-extension-values", 404, `{"type":"https://github.com/gastownhall/bdp/problems/not-found","code":"resource-not-found","retry":"after-state-change","retryAfter":{"opaque":true},"diagnostics":null,"diagnosticsTruncated":"opaque"}`, true, false},
		{"read-common-history-disclosure", 409, `{"type":"https://github.com/gastownhall/bdp/problems/conflict","code":"revision-unretained","retry":"after-state-change","missing":{"complete":false,"items":[]},"resource":null}`, false, false},
		{"read-required-missing", 409, `{"type":"https://github.com/gastownhall/bdp/problems/conflict","code":"revision-unretained","retry":"after-state-change"}`, false, false},
		{"read-archive", 410, `{"type":"https://github.com/gastownhall/bdp/problems/gone","code":"resource-pruned","retry":"never","archivedAt":"https://archive.example/r"}`, true, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: canonicalRetainedRoundTrip(func(*http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: test.status, Header: http.Header{"Content-Type": {"application/problem+json"}}, Body: io.NopCloser(strings.NewReader(test.body))}, nil
			})}
			binding := canonicalRetainedReadBinding{options: canonicalRetainedTestBindings(client)}
			reply := binding.readResource(context.Background(), "https://data.example/current")
			if reply.Checked != test.readValid || calls != 1 {
				t.Fatalf("direct GET family admission readValid=%v calls=%d reply=%#v", test.readValid, calls, reply)
			}
			if !bytes.Equal(reply.NativeBody, []byte(test.body)) {
				t.Fatal("direct read overwrote original Problem bytes")
			}
			if test.readValid && (!bytes.Equal(reply.Problem, reply.NativeBody) || reply.OperationErr == nil || reply.ParseErr != nil) {
				t.Fatal("valid native Read Problem replaced or lost")
			}
			if !test.readValid && (reply.ParseErr == nil || len(reply.Problem) != 0) {
				t.Fatal("invalid Read Problem marked as a checked native diagnosis")
			}
			mutation := canonicalRetainedReply{Sent: true, Status: test.status, Headers: http.Header{"Content-Type": {"application/problem+json"}}, NativeBody: []byte(test.body)}
			mutation.validateMutation(retainedUpdateBead)
			if mutation.Checked != test.writeValid {
				t.Fatalf("mutation family admission writeValid=%v: %#v", test.writeValid, mutation)
			}
		})
	}
}
