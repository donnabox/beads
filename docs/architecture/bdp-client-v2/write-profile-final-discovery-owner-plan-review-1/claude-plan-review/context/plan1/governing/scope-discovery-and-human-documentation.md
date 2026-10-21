### Scope discovery and human documentation

The optional `historicalResolution` discovery member is defined under
[Historical resolution](#historical-resolution).

Every BDP Scope has one absolute canonical Scope URL ending in `/`. That URI
is the base for resolving local IDs and durable relative references. This
holds even when a request reached the Scope through an alias or redirect. An
ordinary `GET` of the Scope URI MUST return a successful response carrying a
registered
[`service-desc` link relation](https://www.rfc-editor.org/rfc/rfc8631.html)
to the machine-readable JSON discovery document. The `Link` field is the
normative machine discovery mechanism. A client never interprets the Scope
response body as discovery metadata.

The Scope response MAY be `204 No Content`. It MAY instead be `200 OK` with a
useful human-readable representation such as HTML or Markdown. That
representation MAY visibly link to the same service descriptor and MAY
advertise separate human documentation with `service-doc`. But neither a body
nor a repository-style `README.md` is required for BDP conformance.

A minimal Scope response is:

```http
GET /acme/ HTTP/1.1
Host: beads.example

HTTP/1.1 204 No Content
Link: <bdp.json>; rel="service-desc"; type="application/json"
```

If a service supplies an HTML landing page, it SHOULD link visibly to the
discovery document and any human documentation it advertises. A BDP client
follows `service-desc`. It never depends on scraping the human page.

```http
GET /acme/bdp.json HTTP/1.1
Host: beads.example
Accept: application/json
```

The Read discovery document requires `bdpVersion`, `profile`, `scope`,
`beads`, `links`, and `types`. The optional `limits`,
`maximumEndpointMultiplicity`, `order`, and `historicalResolution` members
apply under their contracts. `aliases` appears exactly when the authority
serves alias resolution; a client MUST NOT construct alias URLs otherwise.
An omitted `order` means `canonical-uri`. These are the only members admitted
by the closed `readDiscovery` definition.

A minimum Read discovery representation is:

```json
{
  "bdpVersion": "0",
  "profile": "read",
  "scope": "https://beads.example/acme/",
  "beads": "https://beads.example/acme/beads/",
  "links": "https://beads.example/acme/links/",
  "types": "https://beads.example/acme/types/"
}
```

`scope` is the canonical Scope identity. It is also the base for resolving
references. `bdpVersion` MUST equal `"0"` for a BDP v0 Scope. A client that
does not implement the advertised value stops rather than guessing
compatibility. `profile` is required and is exactly `"read"`,
`"read-update"`, or `"transactional"`. It advertises the Scope's highest
supported cumulative profile. It is a single value rather than an array
because each higher profile claims every lower profile.
BDP v0 fixes one Bead root and one Link root per Scope. The `beads` and
`links` members are absolute HTTP(S) navigation URLs for those roots. In
canonical local IDs, the fixed roots are still exactly `beads/` and `links/`.
The advertised URLs are the collection URLs that correspond to those roots,
and they are the only top-level paths under which this Scope assigns Bead and
Link semantics. The collection URL itself is a Resource. A service MUST NOT
advertise an additional Bead or Link root, mix both Resource kinds beneath
one root, or let the fixed roots of separately described Scopes overlap.
For the canonical Scope URL `S`, `beads` MUST equal the URL produced by
resolving `beads/` against `S`, `links` MUST equal `links/` resolved against
`S`, and `types` MUST equal `types/` resolved against `S`. When present,
`aliases` MUST equal `alias/` resolved against `S`.

A local Bead ID has the form `beads/{id-path}`, and a local Link ID has the
form `links/{id-path}`. In both, `{id-path}` contains one or more nonempty
segments. Those segments are opaque identity: they do not define containment
or child Scopes. Empty, `.`, and `..` segments, controls, backslashes,
queries, fragments, scheme-relative references, and encoded `/` or `\`
separators are invalid. A service decodes percent escapes exactly once and
rejects invalid UTF-8. It emits unreserved characters literally, and emits
all required percent escapes with uppercase hexadecimal digits. It compares
decoded segments exactly, without Unicode normalization.

For example, both `beads/task-42` and
`beads/projects/alpha/tasks/task-42` are valid local Bead IDs. The second
form does not imply that `projects`, `alpha`, or `tasks` is a container or
Scope. A Link ID follows the same rule beneath `links/`, such as
`links/assigned-to/81`.

An input Resource reference may use that canonical local spelling or the
absolute canonical URL. The authority resolves a local reference against
`scope` and canonicalizes it. Before lookup, it verifies that the first
segment is the fixed root for the required Resource kind. A relative endpoint
reference therefore must identify a live Bead in this Scope. An absolute
endpoint reference outside `scope` remains opaque. Resolution never mutates
an endpoint Bead.

Discovery-document members defined by this specification are fixed BDP
vocabulary. Scope, Resource, Type, schema,
discovery navigation, and pagination `next` members are HTTP(S) URLs. BDP
permits arbitrary absolute URIs only for opaque external endpoint
references. BDP schemas assert this distinction with JSON Schema patterns. Schema-aware tooling may additionally
use JSON Schema `format` annotations, but format behavior is not the sole
enforcement mechanism. BDP v0 does not duplicate navigation through
BDP-specific HTTP link relations. `service-desc` is the one required machine
entry relation. Optional `service-doc` and `describedby` uses keep their
registered Web meanings.

