package bdpwire

import (
	"bufio"
	"bytes"
	"crypto/sha1" //nolint:gosec // git blob identity is sha1 by definition; it is an identifier here, not a security primitive
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// schema/PIN is the vendoring record and this file is what makes it binding.
// Every vendored file is listed with its sha256 and, for a verbatim upstream
// file, its git blob sha1 — the identity `git hash-object` and the GitHub
// trees API report, so anyone can check the vendored bytes against the pinned
// commit without this package's help. Both digests are recomputed from the
// bytes on disk here; a fixture edited in place, a file added without an
// entry, or an entry with no file all fail. Nothing here touches the network.

type pinEntry struct {
	sha256   string
	local    string
	upstream string
	blob     string
}

type pinFile struct {
	header  map[string]string
	entries []pinEntry
}

func loadPin(t *testing.T) pinFile {
	t.Helper()
	f, err := os.Open(filepath.Join(schemaDir, "PIN"))
	if err != nil {
		t.Fatalf("open PIN: %v", err)
	}
	defer f.Close()

	pin := pinFile{header: map[string]string{}}
	scanner := bufio.NewScanner(f)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) == 2 && strings.HasSuffix(fields[0], ":") {
			pin.header[strings.TrimSuffix(fields[0], ":")] = fields[1]
			continue
		}
		if len(fields) != 4 {
			t.Fatalf("PIN line %d: want `sha256 local upstream blob`, got %q", line, text)
		}
		pin.entries = append(pin.entries, pinEntry{fields[0], fields[1], fields[2], fields[3]})
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read PIN: %v", err)
	}
	return pin
}

func gitBlobSHA1(data []byte) string {
	h := sha1.New() //nolint:gosec // see the import comment
	fmt.Fprintf(h, "blob %d\x00", len(data))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func TestPinHeaderNamesThePinnedCommitAndBundle(t *testing.T) {
	pin := loadPin(t)
	if got := pin.header["commit"]; got != Pin {
		t.Errorf("PIN commit = %q, Pin const = %q: the two must name the same upstream commit", got, Pin)
	}
	if got := pin.header["schema-id"]; got != SchemaID {
		t.Errorf("PIN schema-id = %q, SchemaID const = %q", got, SchemaID)
	}
	if got := pin.header["upstream"]; got != "https://github.com/gastownhall/bdp" {
		t.Errorf("PIN upstream = %q", got)
	}
}

func TestEveryVendoredFileIsPinnedAndUnchanged(t *testing.T) {
	pin := loadPin(t)

	listed := map[string]pinEntry{}
	for _, e := range pin.entries {
		if _, dup := listed[e.local]; dup {
			t.Errorf("PIN lists %s twice", e.local)
		}
		listed[e.local] = e
	}

	onDisk := map[string]bool{}
	err := filepath.WalkDir(schemaDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(schemaDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "PIN" {
			return nil
		}
		onDisk[rel] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", schemaDir, err)
	}

	if unlisted := diff(onDisk, listed); len(unlisted) > 0 {
		t.Errorf("files under schema/ with no PIN entry: %v\nevery vendored file is pinned by sha256 — add a line to schema/PIN with its provenance", unlisted)
	}
	if missing := diff(listed, onDisk); len(missing) > 0 {
		t.Errorf("PIN entries with no file on disk: %v", missing)
	}

	for _, e := range pin.entries {
		if !onDisk[e.local] {
			continue
		}
		data := readSchemaFile(t, e.local)
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != e.sha256 {
			t.Errorf("%s: sha256 %s, PIN says %s\nthe vendored bytes changed; re-vendor from the pinned commit or re-pin deliberately (GENERATOR.md, Re-pinning)", e.local, got, e.sha256)
		}
		if e.blob == "-" {
			if !strings.HasPrefix(e.upstream, "docs/specs/bdp.md#L") {
				t.Errorf("%s: a derived entry must name the spec line it was taken from, got %q", e.local, e.upstream)
			}
			continue
		}
		if got := gitBlobSHA1(data); got != e.blob {
			t.Errorf("%s: git blob sha1 %s, PIN says %s (upstream %s)", e.local, got, e.blob, e.upstream)
		}
	}
}

func TestBundleEntryIsTheNormativeArtifact(t *testing.T) {
	pin := loadPin(t)
	var bundle *pinEntry
	for i := range pin.entries {
		if pin.entries[i].local == "bdp-v0.schema.json" {
			bundle = &pin.entries[i]
		}
	}
	if bundle == nil {
		t.Fatal("PIN has no entry for bdp-v0.schema.json")
	}
	// The spec fixes the artifact's path in the upstream repository; the pin
	// must say it came from there and not from some copy.
	if bundle.upstream != "schemas/bdp-v0.schema.json" {
		t.Errorf("bundle upstream path = %q, want schemas/bdp-v0.schema.json", bundle.upstream)
	}

	onDisk := readSchemaFile(t, "bdp-v0.schema.json")
	embedded := SchemaBundle()
	if !bytes.Equal(onDisk, embedded) {
		t.Fatal("SchemaBundle() differs from schema/bdp-v0.schema.json: the embed directive points somewhere else")
	}

	doc := asMap(t, decodeAny(t, embedded), "bundle")
	if got := doc["$id"]; got != SchemaID {
		t.Errorf("bundle $id = %v, SchemaID = %q", got, SchemaID)
	}
	if got := doc["$schema"]; got != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("bundle $schema = %v, want JSON Schema 2020-12", got)
	}
}

func TestSchemaBundleReturnsACopy(t *testing.T) {
	first := SchemaBundle()
	first[0] = 'x'
	if second := SchemaBundle(); second[0] != '{' {
		t.Fatal("SchemaBundle() handed out the embedded bytes themselves; a caller could corrupt them for everyone")
	}
}
