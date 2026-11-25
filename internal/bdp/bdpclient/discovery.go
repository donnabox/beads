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

// wireServiceDescriptor is DUAL-FORM (Lydia's ruling 2026-11-25): the ratified
// core.lumen §ServiceDescriptor shape is canonical — version field `version`,
// `revision?` for build provenance, and `links` as the named-field object
// {serviceDesc?, serviceMeta?, serviceDoc?, changes?}. cleanroom-v023 instead
// emits `bdp` for the version and the RFC 8288 {rel,href} ARRAY (the Link-header
// mirror) in the body — its acknowledged shape bug, which Shelly will reshape.
// This decoder accepts BOTH so the client works against the clean-room today and
// a ratified server; the non-ratified branch is dropped once the reshape lands.
type wireServiceDescriptor struct {
	Version  string          `json:"version"` // ratified
	Bdp      string          `json:"bdp"`     // clean-room fallback
	Revision string          `json:"revision,omitempty"`
	Scope    *string         `json:"scope,omitempty"`
	Base     *string         `json:"base,omitempty"`
	Links    json.RawMessage `json:"links"` // object (ratified) or array (clean-room)
}

type wireLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func (w wireServiceDescriptor) toBDP() bdp.ServiceDescriptor {
	version := w.Version
	if version == "" {
		version = w.Bdp
	}
	return bdp.ServiceDescriptor{
		Version:  version,
		Revision: bdp.Rev(w.Revision),
		Scope:    deref(w.Scope),
		Base:     deref(w.Base),
		Links:    foldLinks(w.Links),
	}
}

// foldLinks accepts both the ratified named-field links object and the
// clean-room's RFC 8288 {rel, href} array, yielding the contract ServiceLinks.
func foldLinks(raw json.RawMessage) bdp.ServiceLinks {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return bdp.ServiceLinks{}
	}
	if bytes.HasPrefix(trimmed, []byte("[")) {
		var arr []wireLink
		if json.Unmarshal(raw, &arr) != nil {
			return bdp.ServiceLinks{}
		}
		var sl bdp.ServiceLinks
		for _, l := range arr {
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
	var sl bdp.ServiceLinks
	_ = json.Unmarshal(raw, &sl) // ratified named-field object
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

// wireGuardCaps (patch/delete) is DUAL-FORM: the ratified cap name is
// `ifRevision`; cleanroom-v023 emits the HTTP header name `ifMatch`. Accept both.
type wireGuardCaps struct {
	IfRevision   bool `json:"ifRevision"` // ratified
	IfMatch      bool `json:"ifMatch"`    // clean-room (If-Match header name)
	Precondition bool `json:"precondition"`
}

func (w wireGuardCaps) ifRevision() bool { return w.IfRevision || w.IfMatch }

// wirePutCaps adds the put-only create-if-absent guard — ratified `ifAbsent`,
// clean-room `ifNoneMatch` (If-None-Match). Dual-form like wireGuardCaps.
type wirePutCaps struct {
	IfRevision   bool `json:"ifRevision"`
	IfMatch      bool `json:"ifMatch"`
	IfAbsent     bool `json:"ifAbsent"`
	IfNoneMatch  bool `json:"ifNoneMatch"`
	Precondition bool `json:"precondition"`
}

func (w wirePutCaps) ifRevision() bool { return w.IfRevision || w.IfMatch }
func (w wirePutCaps) ifAbsent() bool   { return w.IfAbsent || w.IfNoneMatch }

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
// the wire guard-key names. The core cap vocabulary is CLOSED to exactly these
// 11 keys (Lydia's ruling 2026-11-25): get, list, pull, create, put, patch,
// delete, revisions, export, import, push. Any OTHER key is deliberately dropped
// and must never be modeled as a core cap: `changes` is a ServiceLinks relation,
// not a core cap (reading /sys/changes IS `pull`); resolve/apply/evaluate are
// non-core; call/cancel advertise via the operations surface, not this record.
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
				c.Put = &bdp.PutCaps{IfRevision: v.ifRevision(), IfAbsent: v.ifAbsent(), Precondition: v.Precondition}
			}
		case "patch":
			var v wireGuardCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Patch = &bdp.PatchCaps{IfRevision: v.ifRevision(), Precondition: v.Precondition}
			}
		case "delete":
			var v wireGuardCaps
			if json.Unmarshal(raw, &v) == nil {
				c.Delete = &bdp.DeleteCaps{IfRevision: v.ifRevision(), Precondition: v.Precondition}
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
