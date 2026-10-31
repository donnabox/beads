package canonicalbatch

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/steveyegge/beads/graphops"
)

// queryShape is the unchanged review-only shared AST shape from
// gastownhall/bdp ba8d18c6355cbd6dd3499befe03641874bb01c73,
// docs/cli-design/query-ir/query.schema.json. It is an authoring artifact,
// not adoption or advertisement of a released Query/Batch capability.
//
//go:embed query.schema.json
var queryShape []byte
var queryOnce sync.Once
var queryValidator *jsonschema.Schema
var queryValidatorError error

type denySchemaLoads struct{}

func (denySchemaLoads) Load(uri string) (any, error) {
	return nil, fmt.Errorf("schema is not retained: %s", uri)
}

func validateQuery(raw []byte) error {
	if len(raw) == 0 || raw[0] != '{' {
		return invalid("query must be an object")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return invalid("query object")
	}
	if err := closed(fields, "ir", "profile", "selection"); err != nil {
		return err
	}
	ir, err := stringField(fields, "ir", true)
	if err != nil {
		return err
	}
	if ir != "bdp-query-design-1" {
		return fmt.Errorf("%w: query IR %q", ErrUnsupported, ir)
	}
	profile, err := stringField(fields, "profile", true)
	if err != nil {
		return err
	}
	if profile != "list" && profile != "query" {
		return invalid("batch query profile must be list or query")
	}
	if _, exists := fields["selection"]; !exists {
		return invalid("missing selection")
	}
	queryOnce.Do(func() {
		compiler := jsonschema.NewCompiler()
		compiler.DefaultDraft(jsonschema.Draft2020)
		compiler.AssertFormat()
		compiler.UseLoader(denySchemaLoads{})
		schema, loadErr := jsonschema.UnmarshalJSON(bytes.NewReader(queryShape))
		if loadErr != nil {
			queryValidatorError = loadErr
			return
		}
		const id = "https://example.invalid/bdp-query-design-1.schema.json"
		if queryValidatorError = compiler.AddResource(id, schema); queryValidatorError != nil {
			return
		}
		queryValidator, queryValidatorError = compiler.Compile(id)
	})
	if queryValidatorError != nil {
		return fmt.Errorf("query shape compiler: %w", queryValidatorError)
	}
	// Validate the shared selection by placing it in a resource-result request.
	// The actual Batch envelope deliberately does not carry result/delivery knobs.
	fields["result"] = json.RawMessage(`{"form":"resources"}`)
	validation, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(validation))
	if err != nil {
		return invalid("query value: %v", err)
	}
	if err = queryValidator.Validate(value); err != nil {
		return invalid("query shape: %v", err)
	}
	return nil
}

// resolveQuery visits only grammar-owned operand positions. A JSON object inside
// a literal is ordinary data even if it resembles a binding or AST node.
func resolveQuery(raw []byte, scope string, author bool, resolve func(*Operand, graphops.ResourceKind, bool) error) (json.RawMessage, graphops.ResourceKind, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, "", invalid("query object")
	}
	selection, kind, err := resolveNode(fields["selection"], scope, author, resolve)
	if err != nil {
		return nil, "", err
	}
	fields["selection"] = selection
	compiled, err := json.Marshal(fields)
	return compiled, kind, err
}
func resolveNode(raw []byte, scope string, author bool, resolve func(*Operand, graphops.ResourceKind, bool) error) (json.RawMessage, graphops.ResourceKind, error) {
	if bytes.Equal(raw, []byte("true")) || bytes.Equal(raw, []byte("false")) {
		return raw, "", nil
	}
	var f map[string]json.RawMessage
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, "", invalid("selection node")
	}
	op, err := stringField(f, "op", true)
	if err != nil {
		return nil, "", err
	}
	var kind graphops.ResourceKind
	switch op {
	case "scan", "refs":
		collection, _ := stringField(f, "collection", true)
		if collection == "beads" {
			kind = graphops.KindBead
		} else {
			kind = graphops.KindLink
		}
		if op == "refs" {
			references := f["references"]
			if len(references) > 0 && references[0] == '{' {
				operand, parseErr := parseOperand(references, scope, author, true)
				if parseErr != nil {
					return nil, "", parseErr
				}
				if err = resolve(&operand, kind, false); err != nil {
					return nil, "", err
				}
				if operand.Binding == "" {
					f["references"], err = json.Marshal([]Operand{operand})
				} else {
					f["references"], err = json.Marshal(operand)
				}
			} else {
				var entries []json.RawMessage
				if err = json.Unmarshal(references, &entries); err != nil {
					return nil, "", invalid("Reference array")
				}
				values := make([]Operand, len(entries))
				for i, entry := range entries {
					values[i], err = parseOperand(entry, scope, author, false)
					if err != nil {
						return nil, "", err
					}
					if values[i].Revision != "" {
						return nil, "", fmt.Errorf("%w: historical batch selection", ErrUnsupported)
					}
					if err = resolve(&values[i], kind, false); err != nil {
						return nil, "", err
					}
				}
				f["references"], err = json.Marshal(values)
			}
			if err != nil {
				return nil, "", err
			}
		}
	case "literal":
		return raw, "", nil
	case "rev":
		return nil, "", fmt.Errorf("%w: staged values have no public revision", ErrUnsupported)
	case "field":
		var path []any
		if err = json.Unmarshal(f["path"], &path); err != nil {
			return nil, "", err
		}
		if len(path) > 0 && path[0] == "revision" {
			return nil, "", fmt.Errorf("%w: staged values have no public revision", ErrUnsupported)
		}
	}
	// These keys are structural in the selected grammar. Literal.value above is
	// deliberately terminal and no arbitrary object members are recursively read.
	childKeys := []string{}
	switch op {
	case "filter", "any", "all":
		childKeys = []string{"input", "predicate"}
	case "walk":
		childKeys = []string{"roots", "through"}
	case "endpoints":
		childKeys = []string{"input"}
	case "endpoint", "field", "id", "typeIs":
		childKeys = []string{"of"}
	case "not":
		childKeys = []string{"predicate"}
	case "compare":
		childKeys = []string{"left", "right"}
	case "present":
		childKeys = []string{"value"}
	case "contains":
		childKeys = []string{"text", "substring"}
	case "member":
		childKeys = []string{"value", "array"}
	}
	for _, key := range childKeys {
		child, exists := f[key]
		if !exists {
			continue
		}
		resolved, childKind, resolveErr := resolveNode(child, scope, author, resolve)
		if resolveErr != nil {
			return nil, "", resolveErr
		}
		f[key] = resolved
		switch {
		case op == "filter" && key == "input":
			kind = childKind
		case op == "walk" && key == "roots":
			if childKind != "" && childKind != graphops.KindBead {
				return nil, "", invalid("walk roots require Beads")
			}
			kind = graphops.KindLink
		case op == "endpoints" && key == "input":
			if childKind != graphops.KindLink {
				return nil, "", invalid("endpoints input requires Links")
			}
			kind = graphops.KindBead
		}
	}
	if op == "and" || op == "or" {
		var entries []json.RawMessage
		if err = json.Unmarshal(f["operands"], &entries); err != nil {
			return nil, "", err
		}
		for i, entry := range entries {
			entries[i], _, err = resolveNode(entry, scope, author, resolve)
			if err != nil {
				return nil, "", err
			}
		}
		f["operands"], err = json.Marshal(entries)
		if err != nil {
			return nil, "", err
		}
	}
	encoded, err := json.Marshal(f)
	return encoded, kind, err
}
