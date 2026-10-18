### Operation Directory and singleton targets

The discovered `operations/` Resource is a directory of the generic mutation
targets available under the Scope's profile. `GET operations/` returns names
and relative target URLs as a JSON object, so clients follow the directory
rather than construct paths.

A Read+Update Scope's directory contains exactly eight singleton
targets — the six Resource targets plus `put-alias` and `delete-alias` —
and `sequence`:

```json
{
  "createBead": "create-bead",
  "updateBead": "update-bead",
  "deleteBead": "delete-bead",
  "createLink": "create-link",
  "updateLink": "update-link",
  "deleteLink": "delete-link",
  "putAlias": "put-alias",
  "deleteAlias": "delete-alias",
  "sequence": "sequence"
}
```

Concretely, each Read+Update singleton target accepts `POST` with an
`application/json` body containing its operation record with `operation`
and `name` removed — the bundle defines `createBeadRequest`,
`updateBeadRequest`, `deleteBeadRequest`, `createLinkRequest`,
`updateLinkRequest`, `deleteLinkRequest`, `putAliasRequest`, and
`deleteAliasRequest` — and one required
`Idempotency-Key` field carrying a key under
[Idempotency keys](#idempotency-keys). A singleton never accepts `@name`,
bare or within a Pinned Reference; the bundle's singleton request
definitions reject the spelling, so it is carrier syntax rejected before
execution with `malformed-request`.
A successful singleton returns `200 OK` whose body is the mutation result
defined under [Mutation results](#mutation-results) — or, for an alias
target, the alias result defined under [Alias targets](#alias-targets) —
and carries no `ETag`
and no `Location`: the operation target is not the Resource's URL, and the
result's `resource.id` and `resource.revision` say what those fields would.
A failed singleton
returns the direct problem at its code's HTTP status; a singleton whose
key is retained, in flight, conflicting, or expired answers exactly as the
corresponding sequence member would, as a direct response. Singleton and
sequence forms share one idempotency namespace, one semantic-identity
rule, and one retention rule. Mutation responses carry
`Cache-Control: private, no-store`. A mutation target responds
`405 Method Not Allowed` with `Allow: POST` to every other method, and the
Operation Directory responds `405` with `Allow: GET, HEAD` to every method
but those two; both follow the Read profile's `405` rule — no BDP Problem
body — and its `OPTIONS` rule: when cross-origin access is enabled,
`OPTIONS` is answered according to the CORS rules rather than with `405`
and joins `Allow`; listing it in `Allow` is not the preflight behavior. The
bundle defines the Read+Update discovery document as `readUpdateDiscovery`
and the directory response above as `readUpdateOperationDirectory`.

BDP v0 does not additionally define `POST` on collections or `PUT`, `PATCH`,
or `DELETE` on individual Resource URLs. BDP v0 also does not add a POST-based
read selector fallback. Services enforce bounded GET request-target and
Selector limits and may pre-advertise them through `limits`. A future version
may add another read carrier if implementation evidence requires it.

