### Read+Update discovery

Read+Update adds required `operations` and `aliases` navigation members to
Read discovery and advertises `profile: "read-update"`. It inherits the
optional History capability. Its discovery schema is the closed
`readUpdateDiscovery` definition: `bdpVersion`, `profile`, `scope`, `beads`,
`links`, `types`, `operations`, and `aliases` are required; `limits`,
`maximumEndpointMultiplicity`, `order`, and `historicalResolution` are optional
under their inherited contracts. These are the only admitted members.

A minimum Read+Update discovery representation adds its Operation
Directory and its alias root:

```json
{
  "bdpVersion": "0",
  "profile": "read-update",
  "scope": "https://beads.example/acme/",
  "beads": "https://beads.example/acme/beads/",
  "links": "https://beads.example/acme/links/",
  "types": "https://beads.example/acme/types/",
  "operations": "https://beads.example/acme/operations/",
  "aliases": "https://beads.example/acme/alias/"
}
```

