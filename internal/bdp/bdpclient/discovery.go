package bdpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/steveyegge/beads/internal/bdp"
)

// This file is the BDP rc2 discovery surface of the client: the GET / service
// descriptor and the GET /sys/capabilities document, plus Discover — the
// side-effect-free conformance handshake `bd connect` runs before it persists
// anything (engdocs/design/bdp-connect.md).
//
// Like wire.go, it is an impedance layer: the clean-room server
// (bdp-cleanroom-v023) diverges from ratified core.lumen §ServiceDescriptor /
// §Capabilities in several documented ways, which the DTOs here absorb so the
// bdp.* contract stays spec-faithful:
//
//   - §ServiceDescriptor's version field is `version` in the spec but `bdp` on
//     the clean-room wire; `links` is the spec's named-field object but an
//     RFC 8288 array of {rel, href} on the wire.
//   - §Capabilities' `core` is a RECORD ({get:…, list:…}) in the spec but the
//     clean-room serializes the trait's Vec as an ARRAY of singleton objects
//     ([{"get":…}, {"list":…}]) — the clean-room source calls this out as a
//     known "SHAPE GAP". decodeCoreOps accepts EITHER form, so the client works
//     against the clean-room today and a ratified record-form server later.
//   - The put/patch/delete guard caps use the HTTP header names `ifMatch` /
//     `ifNoneMatch` on the wire, vs the spec cap names `ifRevision` / `ifAbsent`.

const (
	pathServiceDescriptor = "/"
	pathCapabilities      = "/sys/capabilities"
)

// Discover is the conformance handshake: it fetches the service descriptor,
// negotiates the protocol version, and reads the capabilities document. It is
// read-only — it persists nothing — so bd connect can prove a target speaks rc2
// BDP before attaching a workspace to it.
func (c *RemoteScope) Discover(ctx context.Context) (bdp.ServiceDescriptor, bdp.Capabilities, error) {
	desc, err := c.fetchDescriptor(ctx)
	if err != nil {
		return bdp.ServiceDescriptor{}, bdp.Capabilities{}, err
	}
	if err := negotiateVersion(desc.Version); err != nil {
		return desc, bdp.Capabilities{}, err
	}
	caps, err := c.Capabilities(ctx)
	if err != nil {
		return desc, bdp.Capabilities{}, err
	}
	return desc, caps, nil
}

// fetchDescriptor does GET / and decodes the service descriptor.
func (c *RemoteScope) fetchDescriptor(ctx context.Context) (bdp.ServiceDescriptor, error) {
	resp, body, err := c.do(ctx, http.MethodGet, pathServiceDescriptor, nil, nil)
	if err != nil {
		return bdp.ServiceDescriptor{}, err
	}
	if resp.StatusCode/100 != 2 {
		return bdp.ServiceDescriptor{}, errorFrom(resp.StatusCode, body)
	}
	var w wireServiceDescriptor
	if err := json.Unmarshal(body, &w); err != nil {
		return bdp.ServiceDescriptor{}, fmt.Errorf("bdpclient: bad service descriptor: %w", err)
	}
	return w.toBDP(), nil
}

// fetchCapabilities does GET /sys/capabilities and decodes the document.
func (c *RemoteScope) fetchCapabilities(ctx context.Context) (bdp.Capabilities, error) {
	resp, body, err := c.do(ctx, http.MethodGet, pathCapabilities, nil, nil)
	if err != nil {
		return bdp.Capabilities{}, err
	}
	if resp.StatusCode/100 != 2 {
		return bdp.Capabilities{}, errorFrom(resp.StatusCode, body)
	}
	var w wireCapabilitiesDoc
	if err := json.Unmarshal(body, &w); err != nil {
		return bdp.Capabilities{}, fmt.Errorf("bdpclient: bad capabilities document: %w", err)
	}
	return w.toBDP()
}

// negotiateVersion checks the service descriptor's protocol version is one this
// client speaks (0.2.x). Negotiation reads major.minor (core.lumen
// §ServiceDescriptor).
func negotiateVersion(v string) error {
	if v == "" {
		return fmt.Errorf("bdpclient: service descriptor reported no protocol version")
	}
	major, minor, ok := majorMinor(v)
	if !ok {
		return fmt.Errorf("bdpclient: unparseable protocol version %q", v)
	}
	if major != 0 || minor != 2 {
		return fmt.Errorf("bdpclient: server speaks BDP %s; this client speaks 0.2.x", v)
	}
	return nil
}

func majorMinor(v string) (major, minor int, ok bool) {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return 0, 0, false
	}
	ma, err1 := strconv.Atoi(parts[0])
	mi, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return ma, mi, true
}

// --- ServiceDescriptor wire DTO ---

type wireServiceDescriptor struct {
	// Bdp is the protocol version (the clean-room wire key is "bdp", not the
	// spec's "version").
	Bdp   string     `json:"bdp"`
	Scope *string    `json:"scope,omitempty"`
	Base  *string    `json:"base,omitempty"`
	Links []wireLink `json:"links"`
}

type wireLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func (w wireServiceDescriptor) toBDP() bdp.ServiceDescriptor {
	return bdp.ServiceDescriptor{
		Version: w.Bdp,
		Scope:   deref(w.Scope),
		Base:    deref(w.Base),
		Links:   foldLinks(w.Links),
	}
}

// foldLinks turns the RFC 8288 {rel, href} array into the spec's named-field
// ServiceLinks object.
func foldLinks(links []wireLink) bdp.ServiceLinks {
	var sl bdp.ServiceLinks
	for _, l := range links {
		switch l.Rel {
		case "service-desc":
			sl.ServiceDesc = l.Href
		case "service-meta":
			sl.ServiceMeta = l.Href
		case "service-doc":
			sl.ServiceDoc = l.Href
		case "tag:bdp.dev,2026:rel/changes", "changes":
			sl.Changes = l.Href
		}
	}
	return sl
}

// --- Capabilities wire DTO ---

type wireCapabilitiesDoc struct {
	Core  json.RawMessage            `json:"core"`
	Scope wireScopeCaps              `json:"scope"`
	Packs map[string]json.RawMessage `json:"packs,omitempty"`
}

type wireScopeCaps struct {
	StorageClasses []string       `json:"storageClasses"`
	RevisionModel  string         `json:"revisionModel"`
	IdModel        string         `json:"idModel"`
	Aliases        *wireAliasCaps `json:"aliases,omitempty"`
	BodyFormSpill  *bool          `json:"bodyFormSpill,omitempty"`
}

type wireAliasCaps struct {
	Indexed bool `json:"indexed"`
}

// wireGuardCaps covers patch/delete, whose wire keys are the HTTP header names.
type wireGuardCaps struct {
	IfMatch      bool `json:"ifMatch"`
	Precondition bool `json:"precondition"`
}

// wirePutCaps is wireGuardCaps plus the put-only ifNoneMatch (ifAbsent) guard.
type wirePutCaps struct {
	IfMatch      bool `json:"ifMatch"`
	IfNoneMatch  bool `json:"ifNoneMatch"`
	Precondition bool `json:"precondition"`
}

func (w wireCapabilitiesDoc) toBDP() (bdp.Capabilities, error) {
	ops, err := decodeCoreOps(w.Core)
	if err != nil {
		return bdp.Capabilities{}, fmt.Errorf("bdpclient: bad core capabilities: %w", err)
	}
	scope, err := w.Scope.toBDP()
	if err != nil {
		return bdp.Capabilities{}, err
	}
	return bdp.Capabilities{Core: foldCore(ops), Scope: scope, Packs: w.Packs}, nil
}

// decodeCoreOps reads the `core` half as a map of op → caps, accepting BOTH the
// clean-room's array-of-singletons ([{"get":…},{"list":…}]) and the ratified
// record object ({get:…, list:…}).
func decodeCoreOps(raw json.RawMessage) (map[string]json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}
	ops := map[string]json.RawMessage{}
	if bytes.HasPrefix(bytes.TrimSpace(raw), []byte("[")) {
		var arr []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, err
		}
		for _, m := range arr {
			for k, v := range m {
				ops[k] = v
			}
		}
		return ops, nil
	}
	if err := json.Unmarshal(raw, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}

// foldCore maps the decoded op→caps into the typed CoreCapabilities, remapping
// the wire guard-key names and dropping ops that are not rc2 Core (resolve,
// apply, evaluate, changes — clean-room/other-profile primitives).
func foldCore(ops map[string]json.RawMessage) bdp.CoreCapabilities {
	var c bdp.CoreCapabilities
	for op, raw := range ops {
		switch op {
		case "get":
			_ = json.Unmarshal(raw, &c.Get)
		case "list":
			var v bdp.ListCaps
			if json.Unmarshal(raw, &v) == nil {
				c.List = &v
			}
		case "pull":
			var v bdp.PullCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Pull = &v
			}
		case "create":
			var v bdp.CreateCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Create = &v
			}
		case "put":
			var v wirePutCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Put = &bdp.PutCaps{IfRevision: v.IfMatch, IfAbsent: v.IfNoneMatch, Precondition: v.Precondition}
			}
		case "patch":
			var v wireGuardCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Patch = &bdp.PatchCaps{IfRevision: v.IfMatch, Precondition: v.Precondition}
			}
		case "delete":
			var v wireGuardCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Delete = &bdp.DeleteCaps{IfRevision: v.IfMatch, Precondition: v.Precondition}
			}
		case "revisions":
			c.Revisions = &bdp.RevisionsCaps{}
		case "export":
			var v bdp.ExportCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Export = &v
			}
		case "import":
			var v bdp.ImportCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Import = &v
			}
		case "push":
			var v bdp.PushCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Push = &v
			}
		}
	}
	return c
}

func (w wireScopeCaps) toBDP() (bdp.ScopeCapabilities, error) {
	// revisionModel and idModel are required, stated-outright guarantee tiers:
	// an unrecognized value must not be guessed.
	rm := bdp.RevisionModel(w.RevisionModel)
	if rm != bdp.RevisionModelVersioned && rm != bdp.RevisionModelUnversioned {
		return bdp.ScopeCapabilities{}, fmt.Errorf("bdpclient: unknown revisionModel %q", w.RevisionModel)
	}
	im := bdp.IdModel(w.IdModel)
	if im != bdp.IdModelPermanent && im != bdp.IdModelReusable {
		return bdp.ScopeCapabilities{}, fmt.Errorf("bdpclient: unknown idModel %q", w.IdModel)
	}
	sc := bdp.ScopeCapabilities{RevisionModel: rm, IdModel: im}
	for _, s := range w.StorageClasses {
		sc.StorageClasses = append(sc.StorageClasses, bdp.StorageClass(s))
	}
	if w.Aliases != nil {
		sc.Aliases = &bdp.AliasCaps{}
	}
	if w.BodyFormSpill != nil {
		sc.BodyFormSpill = *w.BodyFormSpill
	}
	return sc, nil
}
