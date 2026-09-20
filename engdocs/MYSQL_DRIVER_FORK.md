# Development MySQL driver dependency

The graph-session candidate uses a public, immutable MySQL driver fork so its
connector can refuse `LOCAL INFILE` before opening a file or invoking a registered
reader. The option is off by default; existing Beads callers retain their driver
behavior. Only the private `internal/storage/graphsession` package opts in.

This is a development dependency. Graphsession has no production caller, and
transport refusal does not establish engine exclusion, server terminality,
production authority, or a public graph-backed BDP service. Its protocol coverage
uses text `COM_QUERY`; the driver's preexisting prepared-query initial-OK/later-
result traversal limitation remains outside that coverage.

## Source and provenance

| Input | Exact source |
| --- | --- |
| Upstream base | `go-sql-driver/mysql` v1.10.0, `a065b60ab6d0c8e15468e7709c7f76acf4431647` |
| Fork | [donnabox/mysql](https://github.com/donnabox/mysql) |
| Holding branch | `codex/janet-disable-local-infile-20260920` |
| Reviewed change | [`ca53f9bcb7277fc1744c631eb4e788356aec80fd`](https://github.com/donnabox/mysql/commit/ca53f9bcb7277fc1744c631eb4e788356aec80fd) |
| Module version | `v1.10.1-0.20260920194038-ca53f9bcb727` |
| Module checksum | `h1:u8EYKR9V5LDNFdp+AYmXpUZQsFWj6DuxEIJyW/kY9GY=` |
| go.mod checksum | `h1:M+cqaI7+xxXGG9swrdeUIoPG3Y3KCkF0pZej+SK+nWk=` |

The fork preserves the upstream parent and history, canonical module declaration,
MPL-2.0 [LICENSE](https://github.com/donnabox/mysql/blob/ca53f9bcb7277fc1744c631eb4e788356aec80fd/LICENSE),
and [AUTHORS](https://github.com/donnabox/mysql/blob/ca53f9bcb7277fc1744c631eb4e788356aec80fd/AUTHORS).
The version was resolved by Go after verifying the upstream tag in the public
fork. A fresh normal-proxy/checksum download matched the exact origin commit,
complete 21-file production Go source set, module files, license and authors.
No local-path replacement, private proxy or checksum exemption is required.

The root module and both example modules explicitly select:

```go
replace github.com/go-sql-driver/mysql v1.10.0 => github.com/donnabox/mysql v1.10.1-0.20260920194038-ca53f9bcb727
```

The example pins deliberately keep their dependency graph aligned. They are not
required by today's example imports, and successful example type-checking does
not exercise graphsession. Go does not inherit replacements from dependency
modules; a downstream consumer needs its own directive if its reachable package
graph uses the new API or it deliberately selects this graph.

## Build and installation boundary

Use a source checkout and the normal `make build` path for this development line.
Local main-module installation remains possible, but Go rejects replace/exclude
directives in a module installed with `go install pkg@version`. Donna's Beads
fork (`donnabox/beads`) also retains the canonical `github.com/steveyegge/beads`
module path. Do not present a Donna branch as a published upstream module version
or change upstream installation examples to point at it.

Main/release promotion and version-suffixed installation compatibility remain
separate holds. The [ICU policy](ICU-POLICY.md) still describes the supported
upstream installation paths; this transport dependency does not replace that
solution.

Normal-module qualification must include exact selected-module assertions,
graphsession hostile-peer and affected/race tests, native and Windows lint,
canonical build, both examples, module verification, tidy stability and the final
combined full test suite. Nix also needs the actual updated `vendorHash` and a
successful direct build at the final source head. A hash-derivation failure is
not a passing Nix gate. These requirements are not claims that every gate has
already completed for this source head.

## Maintenance and removal

Keep the replacement version-specific. If the selected version changes from
v1.10.0, the replacement becomes inactive. Whole-repository tests and lint that
compile graphsession then reject the missing option. Current binary, Nix and release
builds only compile `cmd/bd`, which does not import graphsession, so those builds
alone do not enforce this boundary. Tidy alone does not diagnose an inactive
replacement either.

For an upstream update, ordinarily merge the required release into the fork,
preserving this commit's ancestry; review the combined change and rerun the
protocol/race/default-behavior controls. Publish a new immutable version, update
all three module pins and checksums, and recompute Nix's hash. Keep the published
commit reachable from the holding branch: do not delete that branch or
force-push it. Advance it only through history-preserving changes; never rewrite
the published pin. v1.10.1 is not included in this v1.10.0-based fork merely
because its pseudo-version starts with `v1.10.1-0`.

Track advisories against canonical `github.com/go-sql-driver/mysql` and the actual
upstream base until the fork is removed. A scan of the fork path alone is not
proof of upstream advisory coverage: cached `golang.org/x/vuln` v1.1.4 source
queries the replacement path. Record the actual scanner version and verify its
replacement handling when qualifying a promoted tree; the cached source
observation is not a current vulnerability assessment or a security exception.

Once an upstream release provides an equivalent reviewed contract, remove the
replacement from all three modules, retidy, update Nix and this document, and
repeat the relevant gates before declaring installation compatibility restored.
