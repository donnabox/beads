# PR #366: actual-main handback to Iris

The [bounded Iris handoff](https://github.com/donnabox/agent-coordination/issues/1#issuecomment-6547097088)
requested ordinary #366 landing and the resulting main identity before Iris
continues #362, then #356. On receipt, [#366](https://github.com/donnabox/beads/pull/366)
was already merged at **2026-10-31 17:58:49 UTC**. This worker verified the
completed landing; it did not perform another merge or change the source.

## Landed source

| Identity | Value |
|---|---|
| Actual main and merge commit | `751d311e4cbf6cca8c604a1ca6f0d4ee0cc05453` |
| Landed tree | `19a18e3731b002af4b7a8b2dfc39a0e098177e08` |
| First parent, prior actual main | `ef907c023b323c52a7153f99502d4e182b8c02ec` |
| Second parent, approved original head | `5a08d8cf3dc8fa9c951b223465a3a69e57eb9b73` |
| Approved and landed BUILD blob | `06f490806873aad578e084e66e4761fe86d2b5a7` |

GitHub's main ref, PR merge metadata and fetched Git objects agree. Comparing
the complete landed tree to its first parent changes only
`internal/httpclient/BUILD.bazel`, +7/-7. The regular test target changes from
two shards to three, with six replacement comment lines. The file's blob
matches the approved head exactly; every other prior-main postimage is
preserved. The tree also matches the prospective tree recorded before landing.
The original source branch still points to the approved head, which is the
merge's second parent. See the [exact landed patch](landed.patch) and
[machine-readable receipt](receipt.json).

## Existing review and qualification record

The [ready disposition](https://github.com/donnabox/beads/pull/366#issuecomment-6546819333)
records Donna's substantive approval of that exact head and delegated landing
by Claudia, who did not author the source. The
[cause-specific baseline receipt](https://github.com/donnabox/beads/pull/366#issuecomment-6547094690)
resolves the selected landing's earlier blanket base-health hold and preserves
the separate remote-worker alert. The merge commit explicitly identifies that
delegated independent landing and those two records. This worker preserves
that recorded disposition; it supplies no new human approval. GitHub returns
no formal reviews and no review threads.

The received head metadata contains 114 successful checks, 11 skipped checks,
and no failed or pending checks. The existing hosted proof and its applicability
limits remain those recorded in the PR: the earlier source/profile, reused
integration action, 394 embedded opt-in HTTP root skips and 53 KNOWN DIVERGENCE
HTTP root skips. No new execution is claimed here.

Actual-main metadata has two successful checks and 20 skipped checks. Skips
receive no execution credit. That metadata does not establish fresh combined
execution or universal main qualification. A post-merge preflight reports
`UNKNOWN` merge state and retains the historical Autofix/Nightly warnings; it
is not a new passing merge gate. The prelanding comment records the applicable
cause-specific receiving. The separate remote-worker environment alert remains
unresolved.

## Receiving step for Iris

Use the actual landed main identity above when composing #362. Preserve its
contribution and bind review, routing witnesses and ordinary gates to the real
resulting source. The saved #362 integration and required-gate failures remain
failures until qualified on that source; #366's old result does not clear them.
Then advance #356 with its bounded export diff and normal acceptance gates.

The [Iris status and compatibility obligation](https://github.com/donnabox/agent-coordination/blob/8a7f5b30544d3364f9d8a2558488b6167ad093d9/context/iris/independent-362-356-20261031/STATUS.md)
still applies: ordinary link-mode export, supported flags, memory defaults,
associations and output/error correspondence require compatibility evidence;
outside link mode, existing behavior must remain unchanged. The interim
`restorable:false` diagnostic is no waiver. No full retained import/History,
wire-contract adoption or A-to-B-to-C transfer qualification follows from #366.

## Validation and scope

Receipt generation asserted the exact ordered parents, main ref, merged
state, complete changed-path set, numstat, tree and blob identities, and
original-branch preservation. It also recorded merge time and head/main checks. The evidence files
have no source, workflow, timeout, assertion, profile or protection changes.
Documentation validation is recorded in [validation.txt](validation.txt).

No Beads database, service, producer, engine, CI run, watcher or retry was started.
This artifact is published on a separate evidence branch; main is unchanged by
this delivery. No resources are held.

Agent-Signature: codex-unknown-model-unknown-reasoning on behalf of Donna Box
