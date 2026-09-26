# Graph BDP Read HTTP preview

An initialized graph workspace on ordinary shared-server Dolt can now serve
BDP Read through the installed `bd serve` command. It reuses the existing
listener, Host checks, token-file authentication, admission limits and graceful
shutdown. Read routing projects the graph store directly; it does not open a
legacy Issue store or publish legacy mutation routes.

This is a fork preview under [the delivery plan](https://github.com/donnabox/beads/pull/18).
The protocol and independent public client are pinned to gastownhall/bdp
`53bdbd03136875f952af184fce7b3c7af8f74e96`. The September 23, 2026 07:01 PDT
attempt clock is unchanged. Human review and Jim's integration target remain
pending; passing tests do not approve these provisional contracts.

## Try it in a disposable workspace

Start an ordinary Dolt server with its own disposable data directory, then run
these commands in a new temporary Git repository. Choose a fixed free HTTP port
before initialization; the initialized Scope URL remains the canonical identity.
The example assumes Dolt is already listening on port 3307.

```sh
bd init --graph-mode link --scope-url http://127.0.0.1:8765/demo/ \
  --server --external --server-host 127.0.0.1 --server-port 3307 \
  --server-user root --database graph_demo --prefix demo \
  --non-interactive --skip-hooks --skip-agents
bd remember 'Why we chose this design' --id beads/plan --title 'Plan'
bd show beads/plan
bd serve --readonly --addr 127.0.0.1:8765
```

In another terminal:

```sh
curl -i http://127.0.0.1:8765/demo/
curl http://127.0.0.1:8765/demo/bdp.json
curl http://127.0.0.1:8765/demo/beads/plan
curl 'http://127.0.0.1:8765/demo/beads/plan?view=properties'
curl 'http://127.0.0.1:8765/demo/beads/?limit=1'
curl 'http://127.0.0.1:8765/demo/beads/plan?view=links&direction=both'
```

Follow the complete `next` URL from a collection. It is a retained snapshot
continuation: later CLI writes change fresh reads, while an existing cursor
continues the old records and owned Link state. Process restart loses cursors;
start a new collection read afterward. The listener must be reachable through
the persisted Scope URL; an allowed transport Host alias cannot rename Resources.

## Implemented surface and limits

The Scope root returns 204 with a `service-desc` Link. `bdp.json` advertises the
Read profile and limits. GET and HEAD cover canonical Bead, Link and Type reads,
properties, complete Bead/Link/Type inventories, structural and bounded Selector
filters, incident Links, and `include=links` aggregates. An aggregate's Bead,
owned Links and first incident page derive from one storage snapshot.

Known Resources supply revision ETags for exact Resource and properties views.
Accept and conditional handling occur after authentication, authority validation,
query validation, storage access and serialization. Missing Resources and invalid
cursors cannot become 304 responses. HEAD uses the UTF-8 byte length of the GET
representation; unsupported methods return 405. No Last-Modified date is invented.

The preview checks a conservative 16 MiB acquisition budget for the complete
current workspace inside the same read transaction, using SQL byte lengths before
loading payloads into Go. It charges current payloads and retained records,
Issue hydration data, owner copies, all allocation metadata (including tombstones),
and per-row overhead. Orphan informational rows also count before corruption checks. This can refuse an exact
Resource read or a small filtered page because other current data is oversized;
it does not restrict writes or charge non-current History. It is a persisted-byte
budget, not an exact Go heap or SQL-engine memory guarantee. Immutable metadata
also has expected-length checks before decoding. The existing 1,000-live-Resource
inventory cap applies before filtering. Page limits are 100 by default and 1,000 maximum; cursor lifetime is
five minutes. It retains at most 32 snapshots, 4,096 continuation positions and
32 MiB total. Each snapshot is limited to 7 MiB so that envelopes and continuation
URLs fit the advertised 8 MiB representation limit. Request targets are limited
to 32 KiB; Selectors to 16 KiB, depth 256 and 2,048 nodes. Capacity pressure refuses
new snapshots without evicting valid ones. These are reversible preview bounds.

Every accepted token has the same complete-workspace read view. Existing
`--auth-token-file`, `--allowed-host`, non-loopback opt-in and token rotation
apply; token-file reload retains the existing last-good-file policy. Authentication
runs before cursor lookup on every request. Responses use `private, no-store`.
No per-user filtering or revocation of a distinct authorization projection is
claimed. Loopback without a token uses the existing local trust policy.

The server checks the persisted binding and installed contracts on every read,
including retained pages. Out-of-band SQL and hot restore are unsupported:
stop serving before restoring, revalidate the authority, and restart. There is
no TLS or CORS support. The authenticated Scope/discovery read is a readiness
probe; graph mode does not publish legacy `/healthz` or `/v0` routes.

Embedded graph HTTP serving is refused before storage opens or a listener binds.
Embedded CLI initialization and operations remain available. HTTP writes,
History, aliases, mutable Type installation, full Memory conformance and production
migration are unfinished. This preview does not widen the supported Issue CLI
workflow or claim the entire graph delivery plan is complete.

## Reproducible independent-client proof

Install the actual binary with `make install-force INSTALL_DIR=/absolute/bin`.
Check out the exact public BDP pin above, install with pnpm 11.20.0 using
`pnpm install --frozen-lockfile --ignore-scripts`, then build with
`pnpm exec tsc -b packages/client` using Node 24.16.0. Run:

```sh
python3 scripts/graph-bdp-read-smoke.py --bd /absolute/bin/bd \
  --server-port 3307 --bdp-checkout /absolute/bdp-checkout \
  --node /absolute/node --output-dir /absolute/new-receipts
```

The harness initializes a fresh isolated workspace through installed CLI commands,
creates Memory and Issue Beads, mixed Links and a blocking Dependency, and reads
them in new CLI processes. The unchanged public client then discovers and reads
the real HTTP service, traverses all three inventories, applies filters, follows
incident Links and continues an old page across a guarded CLI update.

The pinned client has no aggregate API and consumes a continuation after successful
use. Aggregate and replay checks therefore use real fetch plus the independent
public parsers and are labeled separately. Receipts include actual request URLs,
statuses and hashes, source/client/binary provenance, guarded update output,
process exit codes and active-child counts. No mocked transport, manual schema or
SQL payload is seeded. Failed captures are retained.

The real listener/storage test also covers token rotation, current-authority
failure before a conditional cached response, Host alias identity, error precedence
and aggregate consistency during concurrent writes. The Linux Graph C0 workflow
runs the installed capture alongside the existing embedded/server CLI sequence.
Dolt 2.1.8 fixture provisioning remains serialized as described in the
[projection notes](GRAPH_BDP_READ_PROJECTION.md#shared-server-provisioning-limitation);
this does not claim concurrent unrelated database provisioning is safe.
