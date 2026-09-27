# Unconditional Memory overwrite disclosure

This is a provisional CLI result shape for
[Memory proposal R14](https://github.com/gastownhall/beads/issues/5877), pending review.
It does not add a BDP wire member, HTTP writes, or a public History capability.

A changed Memory `update --unconditional` returns `replaced` beside `memory`
and `changed`. It identifies the actual predecessor with its canonical `id`,
retained `version` and exact recorded `attribution`. A changed owned informational
Link create, update or unlink using `--unconditional-source` returns the same
projection as `replacedSource`. The Link guard and source guard are independent:
`--unconditional` on a Link does not make its source unconditional.

```json
{
  "changed": true,
  "memory": "<complete new Memory record, abridged here>",
  "replaced": {
    "id": "https://example.invalid/demo/beads/plan",
    "version": "<actual previous opaque token>",
    "attribution": {
      "actor": "prior-author",
      "status": "claimed",
      "recordedAt": "<previous recorded timestamp>"
    }
  }
}
```

The normal CLI JSON envelope still wraps this result. Human output also names
the replaced Memory/version and its recorded attribution. Attribution is carried
data; `unknown` stays unknown and the existing observed timestamp is not a native
commit stamp. Use `bd show beads/plan --version TOKEN` for complete old state,
including owned Links, or `bd recall beads/plan --version TOKEN` for its body.

Guarded updates and no-ops omit the new fields. Unowned Issue informational
Links do not report a replaced Memory. Unconditional Issue text edits also retain
their existing result shape without predecessor disclosure; this capability is
specific to Memory. The target is unchanged. Binding/guard
failure, rollback and uncertain commit acknowledgment return failure, not a
successful disclosure. The prior state is captured in the same checked write
transaction; it is never reconstructed by reading after commit.

`bd status --graph --json` advertises `memoryOverwriteDisclosure`. This narrow
capability does not claim full R14 conflict diagnostics, complete Memory fields,
native History context, lifecycle, adoption or remote writes. The explicit
guard/unconditional default remains a CLI proposal question.

## Verification

`scripts/graph-overwrite-disclosure-smoke.py --bd /absolute/path/to/bd --backend
both --server-port PORT --output-dir DIR` authors disposable workspaces through
normal installed CLI initialization and writes. It compares predecessor results
with captured records and fresh-process exact reads. It saves failures, output
hashes and provenance. Real-store Go tests cover transaction failure and
concurrent ordinary-server writers separately. Serialize unrelated database
provisioning on the released Dolt 2.1.8 server.
