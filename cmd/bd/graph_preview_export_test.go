package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/graphoutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

type currentDumpTestSource struct {
	inventory graphread.Inventory
	readErr   error
	closeErr  error
	onClose   func()
	events    []string
}

func (s *currentDumpTestSource) Inventory(context.Context) (graphread.Inventory, error) {
	s.events = append(s.events, "read")
	return s.inventory, s.readErr
}

func (s *currentDumpTestSource) Close() error {
	s.events = append(s.events, "close")
	if s.onClose != nil {
		s.onClose()
	}
	return s.closeErr
}

func TestGraphCurrentDumpPublicDTOsAndScope(t *testing.T) {
	const scope = "https://example.test/work/"
	link := bdpwire.LinkRecord{ID: scope + "edge", Type: scope + "relation", Revision: "=link-v2",
		Source: bdpwire.Reference{URI: scope + "a", Revision: "=old-a"}, Target: bdpwire.Reference{URI: "https://other.test/b"},
		Properties: bdpwire.Properties{"arbitrary": json.RawMessage(`{"雪":[1,true,null]}`)}}
	bead := bdpwire.BeadRecord{ID: scope + "a", Type: scope + "memory", Revision: "=a-v3",
		Properties: bdpwire.Properties{"title": json.RawMessage(`"quote\"\n雪"`)},
		Metadata:   bdpwire.Metadata{"custom": json.RawMessage(`["retained",2]`)},
		OwnedLinks: bdpwire.OwnedLinks{link.Type: {link}}}
	typ := bdpwire.TypeDescriptor{ID: bead.Type, Name: "Memory", Describes: bdpwire.Describes("bead")}
	source := &currentDumpTestSource{inventory: graphread.Inventory{WriterToken: "PRIVATE-CURSOR-DO-NOT-LEAK",
		Beads: []bdpwire.BeadRecord{bead}, Links: []bdpwire.LinkRecord{link}, Types: []bdpwire.TypeDescriptor{typ}}}
	data, err := captureGraphCurrentDump(context.Background(), source, scope)
	if err != nil {
		t.Fatal(err)
	}
	var dump graphCurrentDump
	if err := json.Unmarshal(data, &dump); err != nil {
		t.Fatal(err)
	}
	if dump.Format != "beads-current-state-dump-v1" || dump.Scope != scope || dump.Restorable || len(dump.Omissions) == 0 {
		t.Fatalf("misleading dump declaration: %+v", dump)
	}
	// Compare serialized public DTOs, including pins, arbitrary properties,
	// metadata, owned Links and their intentionally overlapping inventory entry.
	for _, pair := range [][2]any{{dump.Beads[0], bead}, {dump.Links[0], link}, {dump.Types[0], typ}} {
		got, _ := json.Marshal(pair[0])
		want, _ := json.Marshal(pair[1])
		if !bytes.Equal(got, want) {
			t.Fatalf("public DTO changed: got=%s want=%s", got, want)
		}
	}
	if bytes.Contains(data, []byte("PRIVATE-CURSOR")) || bytes.Contains(data, []byte("WriterToken")) || !bytes.HasSuffix(data, []byte("\n")) {
		t.Fatalf("private state leaked or invalid document termination: %s", data)
	}
	if !reflect.DeepEqual(source.events, []string{"read", "close"}) {
		t.Fatalf("capture lifecycle: %v", source.events)
	}
}

func TestGraphCurrentDumpEmptyArrays(t *testing.T) {
	source := &currentDumpTestSource{}
	data, err := captureGraphCurrentDump(context.Background(), source, "https://example.test/work/")
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range []string{"beads", "links", "types"} {
		if !bytes.Contains(data, []byte(`"`+member+`": []`)) {
			t.Fatalf("empty %s is not an array: %s", member, data)
		}
	}
}

func TestGraphCurrentDumpDispatchRefusesBeforeStore(t *testing.T) {
	oldConfig, oldDir, oldOutput := graphPreviewConfig, graphPreviewDir, graphFailureOutput
	graphPreviewConfig = nil // Any attempted storage open would fail this test.
	graphPreviewDir = t.TempDir()
	t.Cleanup(func() { graphPreviewConfig, graphPreviewDir, graphFailureOutput = oldConfig, oldDir, oldOutput })
	for _, test := range []struct {
		name        string
		flags, args []string
		code        int
	}{
		{"missing selection", nil, nil, 2},
		{"false selection", []string{"--current-state=false"}, nil, 2},
		{"positional argument", []string{"--current-state"}, []string{"unexpected"}, 2},
		{"legacy filter", []string{"--current-state", "--all"}, nil, 5},
		{"storage destination", []string{"--current-state", "--output", filepath.Join(graphPreviewDir, "metadata.json")}, nil, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "export"}
			cmd.Flags().Bool("current-state", false, "")
			cmd.Flags().Bool("all", false, "")
			cmd.Flags().String("output", "", "")
			if err := cmd.ParseFlags(test.flags); err != nil {
				t.Fatal(err)
			}
			var out, diagnostic bytes.Buffer
			cmd.SetOut(&out)
			graphFailureOutput = &diagnostic
			err := runGraphPreviewExport(cmd, test.args)
			var failure *exitError
			if !errors.As(err, &failure) || failure.Code != test.code || out.Len() != 0 || diagnostic.Len() == 0 {
				t.Fatalf("pre-store dispatch: err=%v out=%q diagnostic=%q", err, out.String(), diagnostic.String())
			}
		})
	}
}

func TestGraphCurrentDumpRefusalPreservesDestination(t *testing.T) {
	readErr, closeErr := errors.New("capture failed"), errors.New("cleanup failed")
	for _, test := range []struct {
		name   string
		source currentDumpTestSource
		cancel bool
		want   []error
	}{
		{"read", currentDumpTestSource{readErr: readErr}, false, []error{readErr}},
		{"close", currentDumpTestSource{closeErr: closeErr}, false, []error{closeErr}},
		{"both", currentDumpTestSource{readErr: readErr, closeErr: closeErr}, false, []error{readErr, closeErr}},
		{"cancel", currentDumpTestSource{}, true, []error{context.Canceled}},
		{"malformed DTO", currentDumpTestSource{inventory: graphread.Inventory{Beads: []bdpwire.BeadRecord{{Properties: bdpwire.Properties{"bad": json.RawMessage(`{`)}}}}}, false, []error{graphstore.ErrInvalidStore}},
	} {
		t.Run(test.name, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "dump.json")
			if err := os.WriteFile(destination, []byte("previous complete file"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if test.cancel {
				cancel()
			}
			data, err := captureGraphCurrentDump(ctx, &test.source, "https://example.test/work/")
			if err == nil {
				err = graphoutput.PublishFile(ctx, destination, data, 0600)
			}
			for _, want := range test.want {
				if !errors.Is(err, want) {
					t.Fatalf("lost error %v: %v", want, err)
				}
			}
			if len(data) != 0 || !reflect.DeepEqual(test.source.events, []string{"read", "close"}) {
				t.Fatalf("partial output or unchecked cleanup: data=%q events=%v", data, test.source.events)
			}
			actual, err := os.ReadFile(destination)
			if err != nil || string(actual) != "previous complete file" {
				t.Fatalf("destination changed on refusal: %q %v", actual, err)
			}
		})
	}
}

func TestGraphCurrentDumpOutputBound(t *testing.T) {
	source := &currentDumpTestSource{inventory: graphread.Inventory{Beads: []bdpwire.BeadRecord{{Properties: bdpwire.Properties{
		"large": json.RawMessage(`"` + strings.Repeat("x", graphCurrentDumpByteLimit) + `"`),
	}}}}}
	data, err := captureGraphCurrentDump(context.Background(), source, "https://example.test/work/")
	if !errors.Is(err, graphstore.ErrLimitExceeded) || data != nil || !reflect.DeepEqual(source.events, []string{"read", "close"}) {
		t.Fatalf("oversized output leaked or cleanup omitted: bytes=%d events=%v err=%v", len(data), source.events, err)
	}
}

func TestGraphCurrentDumpStorageDestinationRefusal(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, ".beads")
	if err := os.Mkdir(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{workspace, filepath.Join(workspace, "metadata.json"), filepath.Join(workspace, "..", ".beads", "data")} {
		if _, err := graphCurrentDumpDestination(path, workspace); err == nil {
			t.Fatalf("admitted storage destination: %s", path)
		}
	}
	if _, err := graphCurrentDumpDestination(filepath.Join(root, "dump.json"), workspace); err != nil {
		t.Fatal(err)
	}
}

func TestGraphCurrentDumpLegacyAdmission(t *testing.T) {
	isolateBeadsDirForTest(t)
	t.Chdir(t.TempDir())
	for _, key := range []string{"BEADS_DB", "BD_DB", "BD_GRAPH_MODE"} {
		t.Setenv(key, "")
	}
	workspace := t.TempDir()
	t.Setenv("BEADS_DIR", workspace)
	oldCmd, oldActive, oldCfg, oldDir, oldGeneration := exportCmd, graphPreviewActive, graphPreviewConfig, graphPreviewDir, graphPreviewTypeGeneration
	oldOutput, oldStructured := graphFailureOutput, graphPreviewStructuredErrors
	t.Cleanup(func() {
		exportCmd, graphPreviewActive, graphPreviewConfig, graphPreviewDir, graphPreviewTypeGeneration = oldCmd, oldActive, oldCfg, oldDir, oldGeneration
		graphFailureOutput, graphPreviewStructuredErrors = oldOutput, oldStructured
	})
	for _, setting := range []string{"omitted", "true", "false", "corrupt"} {
		t.Run(setting, func(t *testing.T) {
			cmd := &cobra.Command{Use: "export"}
			parent := &cobra.Command{Use: "bd"}
			parent.AddCommand(cmd)
			cmd.Flags().Bool("current-state", false, "")
			if setting != "omitted" {
				value := setting
				if setting == "corrupt" {
					value = "true"
					if err := os.WriteFile(filepath.Join(workspace, "metadata.json"), []byte("{broken"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := cmd.Flags().Set("current-state", value); err != nil {
					t.Fatal(err)
				}
			}
			exportCmd = cmd
			var diagnostic bytes.Buffer
			graphFailureOutput = &diagnostic
			handled, err := admitGraphPreview(cmd)
			if setting == "omitted" || setting == "false" {
				if handled || err != nil {
					t.Fatalf("legacy export hijacked: handled=%t err=%v", handled, err)
				}
				return
			}
			var failure *exitError
			if !handled || !errors.As(err, &failure) || failure.Code != 5 || !strings.Contains(diagnostic.String(), "graph_not_initialized") || graphPreviewActive {
				t.Fatalf("graph-only export reached legacy storage: handled=%t err=%v diagnostic=%q", handled, err, diagnostic.String())
			}
		})
	}
}

func TestGraphCurrentDumpFilesystemAliases(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, ".beads")
	if err := os.MkdirAll(filepath.Join(workspace, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(workspace, "metadata.json")
	if err := os.WriteFile(metadata, []byte("original metadata"), 0600); err != nil {
		t.Fatal(err)
	}
	// Filesystem identity, including a differently cased root or nested path,
	// must protect the original even when lexical Rel would call it outside.
	original, err := os.Stat(workspace)
	if err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, ".BEADS")
	aliased, aliasErr := os.Stat(alias)
	if aliasErr == nil && os.SameFile(original, aliased) {
		for _, path := range []string{alias, filepath.Join(alias, "metadata.json"), filepath.Join(alias, "nested", "dump.json")} {
			if _, err := graphCurrentDumpDestination(path, workspace); err == nil {
				t.Fatalf("case alias admitted: %s", path)
			}
		}
	} else {
		t.Log("temporary filesystem does not resolve this case alias; ordinary and symlink containment remain exercised")
	}
	if runtime.GOOS != "windows" {
		shortcut := filepath.Join(root, "storage-alias")
		if err := os.Symlink(workspace, shortcut); err != nil {
			t.Fatal(err)
		}
		for _, selectedWorkspace := range []string{workspace, shortcut} {
			if _, err := graphCurrentDumpDestination(shortcut, selectedWorkspace); err == nil {
				t.Fatal("destination replaces a workspace-root symlink")
			}
		}
		if _, err := graphCurrentDumpDestination(filepath.Join(shortcut, "nested", "dump.json"), workspace); err == nil {
			t.Fatal("symlink parent inside storage admitted")
		}
		// The publisher replaces a target symlink, so its stored referent remains
		// untouched. Verify the actual publication instead of only testing admission.
		output := filepath.Join(root, "dump.json")
		if err := os.Symlink(metadata, output); err != nil {
			t.Fatal(err)
		}
		if _, err := graphCurrentDumpDestination(output, workspace); err != nil {
			t.Fatal(err)
		}
		if err := graphoutput.PublishFile(context.Background(), output, []byte("outside output"), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(output)
		if err != nil || string(got) != "outside output" {
			t.Fatalf("output=%q err=%v", got, err)
		}
	}
	got, err := os.ReadFile(metadata)
	if err != nil || string(got) != "original metadata" {
		t.Fatalf("storage metadata modified: %q %v", got, err)
	}
}

func TestGraphCurrentDumpPreparedDestination(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, ".beads")
	safe := filepath.Join(root, ".beads-extra")
	for _, path := range []string{workspace, safe} {
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	// No embedded directory is normal for a server workspace. Similar lexical
	// prefixes must remain valid output destinations.
	prepared, err := graphCurrentDumpDestination(filepath.Join(safe, "dump.json"), workspace)
	if err != nil || !filepath.IsAbs(prepared) {
		t.Fatalf("server destination: %q %v", prepared, err)
	}
	if runtime.GOOS == "windows" {
		t.Log("Unix symlink cases are exercised on Unix hosts")
		return
	}
	metadata := filepath.Join(workspace, "metadata.json")
	if err := os.WriteFile(metadata, []byte("original metadata"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(workspace, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(safe, "link")
	if err := os.Symlink(filepath.Join(workspace, "nested"), link); err != nil {
		t.Fatal(err)
	}
	// Do not use Join here: it would erase the raw /../ before preparation.
	raw := link + string(filepath.Separator) + ".." + string(filepath.Separator) + "metadata.json"
	prepared, err = graphCurrentDumpDestination(raw, workspace)
	if err != nil {
		t.Fatal(err)
	}
	if err := graphoutput.PublishFile(context.Background(), prepared, []byte("outside dump"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(safe, "metadata.json"))
	if err != nil || string(got) != "outside dump" {
		t.Fatalf("normalized destination: %q %v", got, err)
	}

	alias := filepath.Join(root, "output-alias")
	if err := os.Symlink(safe, alias); err != nil {
		t.Fatal(err)
	}
	prepared, err = graphCurrentDumpDestination(filepath.Join(alias, "metadata.json"), workspace)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(workspace, alias); err != nil {
		t.Fatal(err)
	}
	if err := graphoutput.PublishFile(context.Background(), prepared, []byte("original safe destination"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(metadata)
	if err != nil || string(got) != "original metadata" {
		t.Fatalf("prepared destination redirected to storage: %q %v", got, err)
	}
	got, err = os.ReadFile(filepath.Join(safe, "metadata.json"))
	if err != nil || string(got) != "original safe destination" {
		t.Fatalf("checked parent was not retained: %q %v", got, err)
	}
}

func TestGraphCurrentDumpStorageSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	workspace, physical := filepath.Join(root, ".beads"), filepath.Join(root, "physical-data")
	for _, path := range []string{workspace, physical} {
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if runtime.GOOS == "windows" {
		t.Log("Unix storage-symlink cases are exercised on Unix hosts")
		return
	}
	for _, alias := range []string{filepath.Join(workspace, "embeddeddolt"), filepath.Join(workspace, "out"), filepath.Join(root, "data-alias")} {
		if err := os.Symlink(physical, alias); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(physical, "manifest")
	if err := os.WriteFile(file, []byte("original database bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(workspace, "embeddeddolt", "manifest"),
		filepath.Join(workspace, "out", "dump.json"),
		file, filepath.Join(physical, "new.json"), physical,
		filepath.Join(root, "data-alias", "manifest"),
	} {
		if prepared, err := graphCurrentDumpDestination(path, workspace); err == nil || prepared != "" {
			t.Fatalf("storage alias admitted: %q => %q, %v", path, prepared, err)
		}
	}
	got, err := os.ReadFile(file)
	if err != nil || string(got) != "original database bytes" {
		t.Fatalf("storage guard changed source: %q %v", got, err)
	}
}

func TestGraphCurrentDumpPipeline(t *testing.T) {
	readErr, closeErr, openErr, outputErr := errors.New("read"), errors.New("close"), errors.New("open"), errors.New("output")
	for _, test := range []struct {
		name          string
		source        currentDumpTestSource
		openErr       error
		cancelOnClose bool
		outputErr     error
		want          error
		wantEvents    []string
	}{
		{name: "success", wantEvents: []string{"open", "read", "close", "publish"}},
		{name: "open", openErr: openErr, want: openErr, wantEvents: []string{"open"}},
		{name: "read", source: currentDumpTestSource{readErr: readErr}, want: readErr, wantEvents: []string{"open", "read", "close"}},
		{name: "close", source: currentDumpTestSource{closeErr: closeErr}, want: closeErr, wantEvents: []string{"open", "read", "close"}},
		{name: "encoding", source: currentDumpTestSource{inventory: graphread.Inventory{Beads: []bdpwire.BeadRecord{{Properties: bdpwire.Properties{"bad": json.RawMessage(`{`)}}}}}, want: graphstore.ErrInvalidStore, wantEvents: []string{"open", "read", "close"}},
		{name: "output limit", source: currentDumpTestSource{inventory: graphread.Inventory{Beads: []bdpwire.BeadRecord{{Properties: bdpwire.Properties{"large": json.RawMessage(`"` + strings.Repeat("x", graphCurrentDumpByteLimit) + `"`)}}}}}, want: graphstore.ErrLimitExceeded, wantEvents: []string{"open", "read", "close"}},
		{name: "root cancelled on close", cancelOnClose: true, want: context.Canceled, wantEvents: []string{"open", "read", "close"}},
		{name: "publication", outputErr: outputErr, want: outputErr, wantEvents: []string{"open", "read", "close", "publish"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if test.cancelOnClose {
				test.source.onClose = cancel
			}
			var acquisition context.Context
			err := exportGraphCurrentDump(ctx, func(ctx context.Context) (graphCurrentDumpReader, string, error) {
				test.source.events = append(test.source.events, "open")
				acquisition = ctx
				if test.openErr != nil {
					return nil, "", test.openErr
				}
				return &test.source, "https://example.test/work/", nil
			}, func(ctx context.Context, data []byte) error {
				test.source.events = append(test.source.events, "publish")
				if !errors.Is(acquisition.Err(), context.Canceled) || ctx.Err() != nil {
					t.Fatalf("publication inherited acquisition budget: capture=%v publication=%v", acquisition.Err(), ctx.Err())
				}
				var dump graphCurrentDump
				if err := json.Unmarshal(data, &dump); err != nil || dump.Scope != "https://example.test/work/" {
					t.Fatalf("detached scope lost: %+v %v", dump, err)
				}
				cancel()
				if !errors.Is(ctx.Err(), context.Canceled) {
					t.Fatal("publication is detached from caller cancellation")
				}
				return test.outputErr
			})
			if !errors.Is(err, test.want) || !reflect.DeepEqual(test.source.events, test.wantEvents) {
				t.Fatalf("pipeline: error=%v want=%v events=%v want=%v", err, test.want, test.source.events, test.wantEvents)
			}
		})
	}
}

func TestGraphCurrentDumpErrorClassification(t *testing.T) {
	oldOutput, oldStructured := graphFailureOutput, graphPreviewStructuredErrors
	t.Cleanup(func() { graphFailureOutput, graphPreviewStructuredErrors = oldOutput, oldStructured })
	graphPreviewStructuredErrors = true
	for _, test := range []struct {
		stage, code string
		exit        int
	}{
		{"open", "graph_not_initialized", 5},
		{"capture", "route_unavailable", 5},
		{"publish", "output_error", 1},
	} {
		t.Run(test.stage, func(t *testing.T) {
			var diagnostic bytes.Buffer
			graphFailureOutput = &diagnostic
			cause := errors.New("test failure")
			source := &currentDumpTestSource{}
			if test.stage == "capture" {
				source.readErr = cause
			}
			err := exportGraphCurrentDump(context.Background(), func(context.Context) (graphCurrentDumpReader, string, error) {
				if test.stage == "open" {
					return nil, "", graphStorageError(cause)
				}
				return source, "https://example.test/work/", nil
			}, func(context.Context, []byte) error {
				return graphFailure("output_error", cause.Error(), 1)
			})
			err = graphStorageErrorWithFallback(err, "route_unavailable")
			var failure *exitError
			var report struct {
				Code string `json:"code"`
			}
			if !errors.As(err, &failure) || failure.Code != test.exit || json.Unmarshal(diagnostic.Bytes(), &report) != nil || report.Code != test.code {
				t.Fatalf("classification changed: err=%v diagnostic=%q", err, diagnostic.String())
			}
		})
	}
}
