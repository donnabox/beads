package httpapi

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/steveyegge/beads/internal/httpapi/typeref"
)

// The Types rule of specs/types/issue/callables.md §2: wherever a Callable carries
// a Bead's Type or a dependency's kind, the member is `type` and its value is a Type
// reference, spelled as Core spells it. typeref holds where the rule reaches, read
// from openapi.v0.yaml; this file runs it both ways on the mount, with Core's own
// mapping (coreProjector) for the values, so the Callables cannot spell a Type
// differently from a Bead record. Arguments go to v0 under v0's names and values;
// v0's result and Problem come back under the Callable's.

// typeRefusal is a value the rule reaches that names no Type here.
type typeRefusal struct {
	at, message string
}

func (e *typeRefusal) Error() string { return e.message }

// callableTypeNames reads a Callable's Type references back to the native names v0
// takes. Nothing is guessed: a short name, an abstract Type, a Type of the other
// kind and a reference with no identity here are each refused in words.
func callableTypeNames(ctx context.Context, p *coreProjector) typeref.Converter {
	return func(kind typeref.Kind, value, at string) (string, error) {
		if !typeref.IsReference(value) {
			example := coreBeadTypeNamespace + "task"
			if kind == typeref.Link {
				example = coreLinkTypeNamespace + "blocks"
			}
			return "", &typeRefusal{at, "`" + value + "` is not a Type reference: a " + kind.String() +
				" Type is spelled as Core spells it, such as `" + example + "`"}
		}
		var name, refusal string
		if kind == typeref.Bead {
			name, refusal = p.beadTypeName(ctx, value)
		} else {
			name, refusal = p.linkTypeName(value)
		}
		if refusal != "" {
			return "", &typeRefusal{at, refusal}
		}
		return name, nil
	}
}

// callableTypeReferences spells v0's native names as Core does. A classification
// with no identity is errCoreUnservable, which Core answers as its temporary
// failure, and so does this mount.
func callableTypeReferences(ctx context.Context, p *coreProjector) typeref.Converter {
	return func(kind typeref.Kind, value, _ string) (string, error) {
		if kind == typeref.Link {
			return p.linkType(value), nil
		}
		return p.beadType(ctx, value)
	}
}

// convertArgumentTypes is the rule on an invocation's arguments, after the service
// defaults and before the adapter: the adapter and v0 see native names under v0's
// member names.
func (row *callableRow) convertArgumentTypes(s *Server, r *http.Request, inv *invocation) *callableProblem {
	plans, err := typeref.Load()
	if err != nil {
		return newCallableProblem(bdpUnavailable, "")
	}
	shape := plans.Arguments(row.op, typeref.ToV0)
	if shape == nil {
		return nil
	}
	changed, err := shape.ApplyMembers(inv.args, callableTypeNames(r.Context(), newCoreProjector(s, r)))
	if err != nil {
		return argumentTypeProblem(err)
	}
	for i, name := range inv.supplied {
		inv.supplied[i] = shape.Rename(name)
	}
	for _, name := range changed {
		if !slices.Contains(inv.supplied, name) {
			inv.rebuild = true
		}
	}
	return nil
}

// argumentTypeProblem is validation-failed, naming the value refused.
func argumentTypeProblem(err error) *callableProblem {
	var refusal *typeRefusal
	var member *typeref.MemberError
	var value *typeref.ValueError
	var at, message string
	switch {
	case errors.As(err, &refusal):
		at, message = refusal.at, refusal.message
	case errors.As(err, &member):
		at, message = member.At, member.Error()
	case errors.As(err, &value):
		at, message = value.At, "a Type reference is a string"
	default:
		return newCallableProblem(bdpMalformedRequest, "the arguments could not be read")
	}
	problem := newCallableProblem(bdpValidationFailed, message)
	problem.Diagnostics = []callableDiagnostic{{Message: message, InstanceLocation: "/arguments" + at}}
	return problem
}

// convertResultTypes is the rule on what v0 answered, inside the route's slot,
// because spelling a Scope-local Type reads the workspace's configuration: a 200 is
// the operation's result, and anything else is v0's Problem.
func (row *callableRow) convertResultTypes(s *Server, r *http.Request, c *callableCapture) {
	if c.final || c.body.Len() == 0 {
		return
	}
	plans, err := typeref.Load()
	if err != nil {
		c.refuse(r, newCallableProblem(bdpUnavailable, ""))
		return
	}
	shape := plans.Problem(typeref.ToCallable)
	if c.status == 0 || c.status == http.StatusOK {
		shape = plans.Result(row.op, typeref.ToCallable)
	}
	out, changed, err := shape.Apply(c.body.Bytes(), callableTypeReferences(r.Context(), newCoreProjector(s, r)))
	var unservable errCoreUnservable
	switch {
	case errors.As(err, &unservable):
		detail := unservable.reason
		if row.batch {
			detail = "the operation ran, and its result cannot be spelled: " + detail
		}
		c.refuse(r, newCallableProblem(bdpUnavailable, detail))
		return
	case err != nil:
		s.event("request_error", "request_id", requestInfo(r.Context()).id, "error", "spelling a result's Types: "+err.Error())
		c.refuse(r, newCallableProblem(bdpUnavailable, ""))
		return
	case !changed:
		return
	}
	c.body.Reset()
	c.body.Write(out)
}

// projectTypeParameter is a v0 parameter's schema under the rule, for the OpenAPI
// document.
func projectTypeParameter(op, name string, schema map[string]any) map[string]any {
	if kind := typeref.ParameterKind(op, name); kind != typeref.None {
		return typeref.ReferenceSchema(kind, schema)
	}
	return schema
}
