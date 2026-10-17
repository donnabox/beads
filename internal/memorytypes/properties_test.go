package memorytypes

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type propertyCase struct {
	Label      string          `json:"label"`
	Type       string          `json:"type"`
	Properties json.RawMessage `json:"properties"`
	Valid      bool            `json:"valid"`
	Path       string          `json:"path"`
}

var fixtureKinds = map[string]BeadKind{"memory": Memory, "observation": Observation, "decision": Decision, "procedure": Procedure, "policy": Policy, "lesson": Lesson}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}

func TestCandidateProperties(t *testing.T) {
	var fixture struct {
		Cases []propertyCase `json:"cases"`
	}
	readJSON(t, "testdata/candidate/schema-cases.json", &fixture)
	if len(fixture.Cases) != 64 {
		t.Fatalf("source fixture count = %d, expected 64", len(fixture.Cases))
	}
	for _, c := range fixture.Cases {
		t.Run(c.Label, func(t *testing.T) { checkProperties(t, c) })
	}
}

func TestAdditionalProperties(t *testing.T) {
	var fixture struct {
		Cases []propertyCase `json:"cases"`
	}
	readJSON(t, "testdata/additional-cases.json", &fixture)
	if len(fixture.Cases) == 0 {
		t.Fatal("additional regression cases are empty")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Label, func(t *testing.T) { checkProperties(t, c) })
	}
}

func checkProperties(t *testing.T, c propertyCase) {
	t.Helper()
	before := bytes.Clone(c.Properties)
	var err error
	if c.Type == "link" {
		for _, kind := range []LinkKind{Association, Guidance, AppliesTo, Supports, Challenges, MemorySupersedes} {
			err = ValidateLinkProperties(kind, c.Properties)
			if (err == nil) != c.Valid {
				t.Fatalf("link kind %d: valid=%v, want %v; error=%v", kind, err == nil, c.Valid, err)
			}
		}
	} else {
		kind, ok := fixtureKinds[c.Type]
		if !ok {
			t.Fatal("unknown test kind", c.Type)
		}
		err = ValidateProperties(kind, c.Properties)
	}
	if (err == nil) != c.Valid {
		t.Fatalf("valid=%v, want %v; error=%v", err == nil, c.Valid, err)
	}
	if !bytes.Equal(before, c.Properties) {
		t.Fatal("validation mutated input")
	}
	if err != nil {
		var violation *Violation
		if !errors.As(err, &violation) || violation.Path == "" || violation.Expected == "" {
			t.Fatalf("missing structured diagnostic: %v", err)
		}
		if c.Path != "" && violation.Path != c.Path {
			t.Errorf("path=%s, want %s", violation.Path, c.Path)
		}
	}
}

func TestRejectMalformedJSON(t *testing.T) {
	for _, raw := range []string{"", `{"title":"ok","body":""} false`, `{"title":`, "null", "[]", string([]byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'})} {
		if err := ValidateProperties(Memory, json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
}

func TestCategorySeparation(t *testing.T) {
	data := json.RawMessage(`{"title":"shape","body":"","choice":"choice","steps":["one"]}`)
	for _, kind := range []BeadKind{UnknownBead, Issue, OtherBead, 255} {
		if err := ValidateProperties(kind, data); err == nil {
			t.Errorf("shape established Memory membership for %d", kind)
		}
	}
	for _, kind := range []LinkKind{UnknownLink, 255} {
		if err := ValidateLinkProperties(kind, []byte(`{}`)); err == nil {
			t.Errorf("accepted unknown/Dependency link %d", kind)
		}
	}
	for kind := Memory; kind <= Lesson; kind++ {
		if !kind.ConformsTo(Memory) || !kind.ConformsTo(kind) || kind.ConformsTo(Issue) {
			t.Errorf("incorrect family membership for %d", kind)
		}
	}
	if Issue.ConformsTo(Memory) || Memory.ConformsTo(Procedure) || UnknownBead.ConformsTo(UnknownBead) || OtherBead.ConformsTo(OtherBead) || BeadKind(255).ConformsTo(255) {
		t.Fatal("membership leaked between contracts")
	}
}

func TestCandidateProvenance(t *testing.T) {
	var manifest struct {
		Commit string `json:"commit"`
		Files  []struct {
			File   string `json:"file"`
			SHA256 string `json:"sha256"`
		} `json:"files"`
	}
	readJSON(t, "testdata/candidate/provenance.json", &manifest)
	if manifest.Commit != "50b13ee08f79a00bf8b135495b3b4d432f811992" {
		t.Fatal("unexpected candidate commit", manifest.Commit)
	}
	for _, file := range manifest.Files {
		data, err := os.ReadFile(filepath.Join("testdata/candidate", file.File))
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(data)
		if hex.EncodeToString(digest[:]) != file.SHA256 {
			t.Errorf("changed candidate fixture: %s", file.File)
		}
	}
}

func TestAuthorityReferenceDiagnostic(t *testing.T) {
	for _, value := range []string{"42", "null", "[]"} {
		data := json.RawMessage(`{"title":"policy","body":"","rule":"rule","authority":` + value + `}`)
		err := ValidateProperties(Policy, data)
		var violation *Violation
		if !errors.As(err, &violation) || violation.Expected != "a URI string or {uri, revision} reference" {
			t.Fatalf("authority %s: wrong expectation: %v", value, err)
		}
	}
}
