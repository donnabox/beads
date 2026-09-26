# Blocking Dependency unlink preview

In a normally initialized graph workspace, `bd unlink` can remove a blocking
Dependency by canonical Link ID. Read the Link and source Issue first, then
guard both revisions:

```sh
bd links beads/work --direction out --json
bd show beads/work --json
bd unlink links/prerequisite --if-revision LINK_VERSION \
  --if-source-revision SOURCE_VERSION --json
bd show beads/work --json
bd ready --json
```

The exact local Scope URL is also accepted. Explicit `--unconditional` and
`--unconditional-source` can replace their respective observed revisions.
Both guards are required because the source Issue owns the blocking Link.
Missing, conflicting or stale guards refuse before any mutation. An actor is
required by the existing Issue-domain writer; the CLI uses its normal actor
resolution.

Removal delegates to the existing Issue-domain Dependency writer and Jim's
retained-version recorder. It updates the authoritative Dependency table,
derived readiness, retained source Issue version and complete remaining owned
Link set in one checked transaction. There is no schema change, second store,
postcommit repair or automatic retry. Targets and other relationships remain
unchanged. Removing one of several open prerequisites keeps the source blocked;
removing the last makes it eligible for the existing ready query.

The deleted canonical Link ID stays reserved. Current reads and repeated unlink
return `gone`; the old live Link and source versions remain available through
exact `show --version`. The deletion token is a tombstone, not an empty live
Link version. Re-adding the same Issue pair requires a new canonical Link ID.
Internally, deletion releases the private deterministic Dependency-row mapping
while retaining the old public identity and history. This permits the new Link
to use the normal domain writer without reviving the old ID.

Fresh inventories and incident listings omit the deleted Link. The current
BDP Read profile represents a deleted Resource as HTTP404 `resource-not-found`;
the CLI distinguishes `gone`. Previously
issued BDP page continuations retain their original snapshot; the HTTP service
remains Read-only and does not expose public History. Readonly and migration
freeze policies refuse unlink.

This reversible preview admits canonical-ID removal only. Blocking typed-pair
unlink and legacy `bd dep remove` resolution remain unavailable; informational
typed-pair unlink keeps its existing behavior. Generic status edits, notes
safeguards, full Issue compatibility, restoration and complete Memory-required
History remain open. No native atomic commit timestamp or chronology contract
is established here.

The implementation reuses the writer already present at its fixed base. When
integrating [Jim's newer writer stack #6358](https://github.com/gastownhall/beads/pull/6358)
and [#6650](https://github.com/gastownhall/beads/pull/6650), reconcile retained-hook
placement so one removal still records exactly one source version. The existing
graph coordination transaction supplies its conflict control; this is not a
replacement for the broader legacy close/removal race work in
[#6719](https://github.com/gastownhall/beads/pull/6719).

## Installed-command proof

Install through `make install-force INSTALL_DIR=/absolute/disposable/bin` and
start ordinary Dolt 2.1.8 with a fresh disposable data directory. Then run:

```sh
python3 scripts/graph-dependency-unlink-smoke.py \
  --bd /absolute/disposable/bin/bd --backend both --server-port PORT \
  --output-dir /absolute/new/evidence-directory
```

This proof uses normal initialization and installed CLI authoring, including
fresh-process reads, readiness, complete owned state, old versions and same-pair
re-addition. It does not seed SQL or use mock storage. Separate real-store tests
cover rollback, damaged authority and concurrent writers. The public-client
Read proof separately checks live deletion and retained continuations. Provision
databases serially on Dolt 2.1.8; the caller owns the disposable server.
