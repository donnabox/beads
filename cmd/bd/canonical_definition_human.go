package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/resourceget"
)

// Human summaries are an adapter over complete checked documents. Explicit
// detail, machine formats and projections retain the canonical renderer.
func definitionHumanView(options clioutput.Options) bool {
	return !options.JSON && len(options.Include) == 0 && len(options.Omit) == 0 &&
		(len(options.Formats) == 0 || len(options.Formats) == 1 && options.Formats[0] == "default")
}

type definitionHumanReference struct {
	URI      string `json:"uri"`
	Revision string `json:"revision"`
}

type definitionHumanRecord struct {
	Type       definitionHumanReference   `json:"type"`
	Properties map[string]json.RawMessage `json:"properties"`
	OwnedLinks map[string][]struct {
		ID       string                   `json:"id"`
		Revision string                   `json:"revision"`
		Type     definitionHumanReference `json:"type"`
		Target   json.RawMessage          `json:"target"`
	} `json:"ownedLinks"`
}

// A concept label requires the actual full checked publication identity and
// edition. Neither a local ID nor a foreign URI basename supplies a label.
func definitionHumanConcept(doc resourceget.DefinitionDocument) string {
	origin, ok := doc.Origin()
	if !ok || doc.Kind() != resourceget.DefinitionBead {
		return ""
	}
	const namespace = "https://github.com/gastownhall/bdp/types/type-system/"
	for _, concept := range []string{"type", "bead-type", "link-type", "refines"} {
		if origin.URI == namespace+concept && (origin.Revision == concept+"-1" || origin.Revision == concept+"-2") {
			return concept
		}
	}
	return ""
}

func definitionHumanText(value string) string {
	var out strings.Builder
	for _, r := range value {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			fmt.Fprintf(&out, "\\u%04X", r)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func definitionHumanReferenceText(raw json.RawMessage) (string, error) {
	var floating string
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &floating); err != nil {
			return "", err
		}
		return definitionHumanText(floating) + " (floating)", nil
	}
	var exact definitionHumanReference
	if err := json.Unmarshal(raw, &exact); err != nil {
		return "", err
	}
	return definitionHumanText(exact.URI) + " @ " + definitionHumanText(exact.Revision), nil
}

func appendDefinitionHuman(out *bytes.Buffer, doc resourceget.DefinitionDocument, descriptor bool) error {
	var record definitionHumanRecord
	if err := json.Unmarshal(doc.NativeJSON(), &record); err != nil {
		return err
	}
	label := definitionHumanConcept(doc)
	if label == "" {
		label = doc.ID()
	}
	fmt.Fprintln(out, definitionHumanText(label))
	if descriptor {
		fmt.Fprintf(out, "  ID: %s\n", definitionHumanText(doc.ID()))
	}
	fmt.Fprintf(out, "  Local ID: %s\n  Revision: %s\n", definitionHumanText(strings.TrimPrefix(doc.ID(), doc.ScopeURL())), definitionHumanText(doc.Revision()))
	fmt.Fprintf(out, "  Metatype: %s @ %s\n", definitionHumanText(record.Type.URI), definitionHumanText(record.Type.Revision))
	if descriptor {
		if origin, ok := doc.Origin(); ok {
			fmt.Fprintf(out, "  Protocol definition: %s @ %s\n", definitionHumanText(origin.URI), definitionHumanText(origin.Revision))
		}
	}
	buckets := make([]string, 0, len(record.OwnedLinks))
	for bucket := range record.OwnedLinks {
		buckets = append(buckets, bucket)
	}
	sort.Strings(buckets)
	for _, bucket := range buckets {
		for _, link := range record.OwnedLinks[bucket] {
			target, err := definitionHumanReferenceText(link.Target)
			if err != nil {
				return err
			}
			name := "Owned Link"
			if bucket == "https://github.com/gastownhall/bdp/types/type-system/refines" &&
				(link.Type.Revision == "refines-1" || link.Type.Revision == "refines-2") {
				name = "Refines"
			}
			fmt.Fprintf(out, "  %s: %s\n", name, target)
			if descriptor {
				fmt.Fprintf(out, "    Link: %s @ %s\n    Type: %s @ %s\n", definitionHumanText(link.ID), definitionHumanText(link.Revision), definitionHumanText(link.Type.URI), definitionHumanText(link.Type.Revision))
			}
		}
	}
	if !descriptor {
		return nil
	}
	fmt.Fprintln(out, "\nDescriptor:")
	if _, present := record.Properties["propertiesSchema"]; !present {
		fmt.Fprintln(out, "  propertiesSchema: absent — no additional local constraints; inherited constraints still apply")
	}
	keys := make([]string, 0, len(record.Properties))
	for key := range record.Properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		raw := record.Properties[key]
		if key == "schemaResources" {
			var resources map[string]json.RawMessage
			if err := json.Unmarshal(raw, &resources); err != nil {
				return err
			}
			fmt.Fprintf(out, "  schemaResources: %d retained resources (complete environment: --format detail or --json)\n", len(resources))
			continue
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, raw, "    ", "  "); err != nil {
			return err
		}
		// JSON escapes C0 controls. Preserve its structural whitespace while
		// also making DEL, C1 and invisible format controls visible in values.
		var prettyText strings.Builder
		for _, r := range pretty.String() {
			if unicode.Is(unicode.Cf, r) || unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
				fmt.Fprintf(&prettyText, "\\u%04X", r)
			} else {
				prettyText.WriteRune(r)
			}
		}
		fmt.Fprintf(out, "  %s:\n    %s\n", definitionHumanText(key), prettyText.String())
	}
	fmt.Fprintln(out, "\nComplete Resource, metadata and owned captures: --format detail or --json")
	return nil
}

func renderDefinitionHuman(doc resourceget.DefinitionDocument) ([]byte, error) {
	var out bytes.Buffer
	fmt.Fprintf(&out, "Scope: %s\n\n", definitionHumanText(doc.ScopeURL()))
	if err := appendDefinitionHuman(&out, doc, true); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func renderDefinitionHumanCollection(scope string, documents []resourceget.DefinitionDocument) ([]byte, error) {
	var out bytes.Buffer
	fmt.Fprintf(&out, "Scope: %s\n\n", definitionHumanText(scope))
	for _, doc := range documents {
		if doc.ScopeURL() != scope {
			return nil, fmt.Errorf("definition display Scope does not match selection")
		}
		if err := appendDefinitionHuman(&out, doc, false); err != nil {
			return nil, err
		}
		fmt.Fprintln(&out)
	}
	fmt.Fprintf(&out, "%d definitions (complete selection)\n", len(documents))
	return out.Bytes(), nil
}
