### Discovery profile membership

Discovery membership is profile-specific: which members appear in the
discovery document depends on the claimed profile. `bdpVersion`, `profile`,
`scope`, `beads`, `links`, and `types` are required in every profile.
`operations` is required in Read+Update and Transactional and prohibited in
Read. The history, receipt, snapshot, changefeed, and Event members are
required only in Transactional and prohibited in both lower profiles. The
optional `limits` and `maximumEndpointMultiplicity` members may appear in any
profile when their contracts apply. The `aliases` member is optional in
Read, where it appears exactly when the authority serves alias resolution,
and required in Read+Update and Transactional, which offer the alias
targets under [Alias targets](#alias-targets) and therefore serve alias
resolution: an authority without aliases
omits the member, and a client MUST NOT construct alias URLs for an
authority that does not advertise it. The optional
`order` member names the collection order under
[Collection retrieval and selection](#collection-retrieval-and-selection);
omission means the `canonical-uri` baseline. The optional `historicalResolution`
member advertises [Historical resolution](#historical-resolution) in every profile;
it is distinct from the Transactional Scope-history fence members.

| Member | Read | Read+Update | Transactional |
| --- | --- | --- | --- |
| `bdpVersion`, `profile`, `scope` | required | required | required |
| `beads`, `links`, `types` | required | required | required |
| `operations` | prohibited | required | required |
| `scopeEpoch`, `authorizationView`, `headPosition`, `minimumReplayPosition` | prohibited | prohibited | required |
| `receipts`, `snapshot`, `changes`, `events` | prohibited | prohibited | required |
| `limits`, `maximumEndpointMultiplicity` | optional | optional | optional |
| `aliases` | optional | required | required |
| `order` | optional | optional | optional |
| `historicalResolution` | optional | optional | optional |

