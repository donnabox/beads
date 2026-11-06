package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/steveyegge/beads/internal/httpapi/spec"
)

// GET <scope>/bdp/call: the OpenAPI document Callables §14 projects from the
// declarations. Parameter and result schemas are v0's own, read from the embedded
// openapi.v0.yaml, so the table says only where each parameter goes.
//
// Nothing here varies by caller or scope: a bearer on this surface admits a client
// to everything, and paths are relative to the server root. It is built once, on
// first read, because parsing the document costs tens of milliseconds that nothing
// else needs to pay.

const (
	callableDocMediaType = "application/vnd.oai.openapi+json;version=3.1.0"
	callableTypeBase     = "https://github.com/gastownhall/bdp/types/callable/"
)

var callableDocument = sync.OnceValues(buildCallableDocument)

func (s *Server) serveCallableList(c *callableCapture, r *http.Request) {
	body, err := callableDocument()
	if err != nil {
		s.event("request_error", "request_id", requestInfo(r.Context()).id, "error", err.Error())
		c.refuse(r, newCallableProblem(bdpUnavailable, ""))
		return
	}
	c.header.Set("Content-Type", callableDocMediaType)
	c.final = true
	c.status = http.StatusOK
	c.body.Write(body)
}

func buildCallableDocument() ([]byte, error) {
	var v0 map[string]any
	if err := yaml.Unmarshal(spec.OpenAPIV0(), &v0); err != nil {
		return nil, fmt.Errorf("parse openapi.v0.yaml: %w", err)
	}
	ops := v0Operations(v0)
	paths := make(map[string]any, len(callableTable))
	seen := make(map[string]string, len(callableTable))
	for i := range callableTable {
		row := &callableTable[i]
		id := callableOperationID(row.name)
		if prior, dup := seen[id]; dup {
			return nil, fmt.Errorf("callables %q and %q give the same operationId %q", prior, row.name, id)
		}
		seen[id] = row.name
		op, ok := ops[row.op]
		if !ok {
			return nil, fmt.Errorf("callable %q routes to %q, which openapi.v0.yaml does not declare", row.name, row.op)
		}
		operation, err := row.openAPIOperation(v0, op)
		if err != nil {
			return nil, err
		}
		paths[callableMount+"/"+row.name] = map[string]any{"post": operation}
	}

	v0Components, _ := v0["components"].(map[string]any)
	schemas := map[string]any{"callableProblem": callableProblemSchema()}
	for name, schema := range mapMember(v0Components, "schemas") {
		schemas[name] = normalizeV0Schema(schema)
	}
	doc := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":   "bd serve Callables",
			"version": "preview-1",
		},
		"paths":      paths,
		"components": map[string]any{"schemas": schemas, "securitySchemes": v0Components["securitySchemes"]},
	}
	if security, ok := v0["security"]; ok {
		doc["security"] = security
	}
	return json.Marshal(doc)
}

// v0Operations indexes the document by operationId.
func v0Operations(v0 map[string]any) map[string]map[string]any {
	ops := map[string]map[string]any{}
	for _, item := range mapMember(v0, "paths") {
		methods, _ := item.(map[string]any)
		for _, raw := range methods {
			if op, ok := raw.(map[string]any); ok {
				if id, ok := op["operationId"].(string); ok {
					ops[id] = op
				}
			}
		}
	}
	return ops
}

func mapMember(node map[string]any, key string) map[string]any {
	member, _ := node[key].(map[string]any)
	return member
}

// deref follows local $refs. Every one in this document is local.
func deref(v0, node map[string]any) map[string]any {
	for {
		ref, ok := node["$ref"].(string)
		if !ok {
			return node
		}
		cur := v0
		for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
			cur = mapMember(cur, part)
		}
		node = cur
	}
}

// callableOperationID is the §14 rule: split at /, -, _ and ., join in lower
// camel case.
func callableOperationID(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool {
		return r == '/' || r == '-' || r == '_' || r == '.'
	})
	for i, word := range words {
		if i > 0 && word != "" {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, "")
}

// v0Parameters reads a v0 operation's flattened parameters: each one's schema,
// help and whether v0 requires it.
func v0Parameters(v0, op map[string]any) (schemas map[string]any, required map[string]bool) {
	schemas, required = map[string]any{}, map[string]bool{}
	list, _ := op["parameters"].([]any)
	for _, raw := range list {
		node, _ := raw.(map[string]any)
		param := deref(v0, node)
		name, _ := param["name"].(string)
		if param["in"] == "header" || name == "" {
			continue
		}
		schema := map[string]any{}
		for k, v := range mapMember(param, "schema") {
			schema[k] = v
		}
		if help, ok := param["description"]; ok {
			schema["description"] = help
		}
		schemas[name] = schema
		required[name], _ = param["required"].(bool)
	}
	rawBody := mapMember(op, "requestBody")
	if rawBody == nil {
		return schemas, required
	}
	body := deref(v0, rawBody)
	schema := deref(v0, mapMember(mapMember(mapMember(body, "content"), claimContentType), "schema"))
	for name, member := range mapMember(schema, "properties") {
		schemas[name] = member
	}
	names, _ := schema["required"].([]any)
	for _, raw := range names {
		if name, ok := raw.(string); ok {
			required[name] = true
		}
	}
	return schemas, required
}

// adapterParameterSchemas are the arguments no v0 operation declares.
var adapterParameterSchemas = map[string]any{
	"source": map[string]any{"type": "string", "description": "The edge's source; v0's `issue_id`. Exactly one of the flat form and `edges`."},
	"target": map[string]any{"type": "string", "description": "The edge's target; v0's `depends_on_id`."},
	"type":   map[string]any{"type": "string", "description": "The edge type. An omitted type is `blocks`."},
}

// openAPIOperation is one row under the §14 table.
func (row *callableRow) openAPIOperation(v0, op map[string]any) (map[string]any, error) {
	schemas, v0Required := v0Parameters(v0, op)
	properties := map[string]any{}
	required := []any{}
	for _, name := range row.declaredParameters() {
		schema, ok := schemas[name]
		if adapter, isAdapter := adapterParameterSchemas[name]; isAdapter && slices.Contains(row.adapterParams, name) {
			schema, ok = adapter, true
		}
		if !ok {
			return nil, fmt.Errorf("callable %q declares %q, which %s does not", row.name, name, row.op)
		}
		properties[name] = normalizeV0Schema(schema)
		if v0Required[name] && !slices.Contains(row.optional, name) {
			required = append(required, name)
		}
	}

	invocation := map[string]any{
		"arguments": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           properties,
			"required":             required,
		},
	}
	kind := "callable-query"
	var parameters []any
	if row.batch {
		kind = "callable-batch"
		invocation["revisionProvenance"] = map[string]any{"type": "object"}
		parameters = []any{map[string]any{
			"name": "Idempotency-Key", "in": "header", "required": false,
			"description": "Admitted and not retained: this Scope does not advertise idempotency.",
			"schema":      map[string]any{"type": "string", "minLength": 1},
		}}
	}

	result := map[string]any{"result": v0ResultSchema(v0, op)}
	resultRequired := []any{"result"}
	if row.batch {
		result["changed"] = map[string]any{"type": "boolean"}
		resultRequired = append(resultRequired, "changed")
	}

	extension := map[string]any{"type": callableTypeBase + kind}
	if row.idempotent {
		extension["idempotent"] = true
	}
	if row.atomic {
		extension["atomic"] = true
	}
	if row.cli != "" {
		extension["cliMapping"] = json.RawMessage(row.cli)
	}

	errorSchema := map[string]any{"$ref": "#/components/schemas/Problem"}
	if len(row.reasons) > 0 {
		errorSchema = map[string]any{"allOf": []any{errorSchema, map[string]any{
			"properties": map[string]any{"code": map[string]any{"enum": stringsToAny(row.reasons)}},
		}}}
	}
	problem := func(withError bool) map[string]any {
		schema := map[string]any{"$ref": "#/components/schemas/callableProblem"}
		if withError {
			schema = map[string]any{"allOf": []any{schema, map[string]any{
				"properties": map[string]any{"error": errorSchema},
			}}}
		}
		return map[string]any{
			"description": "A Problem",
			"content":     map[string]any{problemMediaType: map[string]any{"schema": schema}},
		}
	}

	operation := map[string]any{
		"operationId": callableOperationID(row.name),
		"requestBody": map[string]any{
			"required": true,
			"content": map[string]any{bdpMediaType: map[string]any{"schema": map[string]any{
				"type": "object", "additionalProperties": false,
				"required": []any{"arguments"}, "properties": invocation,
			}}},
		},
		"responses": map[string]any{
			"200": map[string]any{
				"description": "The operation succeeded",
				"content": map[string]any{bdpMediaType: map[string]any{"schema": map[string]any{
					"type": "object", "additionalProperties": false,
					"required": resultRequired, "properties": result,
				}}},
			},
			"409":     problem(true),
			"default": problem(false),
		},
		"x-bdp-callable": extension,
		"description":    op["description"],
	}
	if summary, ok := op["summary"]; ok {
		operation["summary"] = summary
	}
	if parameters != nil {
		operation["parameters"] = parameters
	}
	return operation, nil
}

// declaredParameters is every argument name a row admits.
func (row *callableRow) declaredParameters() []string {
	names := make([]string, 0, len(row.path)+len(row.query)+len(row.body)+len(row.adapterParams))
	names = append(names, row.path...)
	names = append(names, row.query...)
	names = append(names, row.body...)
	return append(names, row.adapterParams...)
}

func stringsToAny(values []string) []any {
	out := make([]any, len(values))
	for i, value := range values {
		out[i] = value
	}
	return out
}

// v0ResultSchema is the 200 response's schema, by its existing $ref.
func v0ResultSchema(v0, op map[string]any) any {
	response := deref(v0, mapMember(mapMember(op, "responses"), "200"))
	schema := mapMember(mapMember(mapMember(response, "content"), claimContentType), "schema")
	return normalizeV0Schema(schema)
}

// callableProblemSchema is the local form of §13's callableProblem. A declaration's
// $refs stay inside its document (§2.3), so it does not point at the BDP schema.
func callableProblemSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []any{"type", "code", "retry"},
		"properties": map[string]any{
			"type":   map[string]any{"type": "string"},
			"status": map[string]any{"type": "integer"},
			"title":  map[string]any{"type": "string"},
			"detail": map[string]any{"type": "string"},
			"code":   map[string]any{"type": "string"},
			"retry":  map[string]any{"type": "string"},
			"error":  map[string]any{},
		},
	}
}

// normalizeV0Schema returns a copy of a v0 schema in JSON Schema 2020-12 terms:
// OpenAPI 3.0's `nullable` becomes a type array, and the code-generation
// annotations are dropped, as specs/types/issue/callables.md §2 says.
func normalizeV0Schema(node any) any {
	switch value := node.(type) {
	case []any:
		out := make([]any, len(value))
		for i, item := range value {
			out[i] = normalizeV0Schema(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(value))
		for key, member := range value {
			if strings.HasPrefix(key, "x-go-") || key == "x-enum-varnames" || strings.HasPrefix(key, "x-bd-") {
				continue
			}
			if key == "nullable" {
				if _, isBool := member.(bool); isBool {
					continue
				}
			}
			out[key] = normalizeV0Schema(member)
		}
		if nullable, _ := value["nullable"].(bool); nullable {
			if name, ok := out["type"].(string); ok {
				out["type"] = []any{name, "null"}
			}
			if enum, ok := out["enum"].([]any); ok {
				out["enum"] = append(enum, nil)
			}
		}
		return out
	}
	return node
}
