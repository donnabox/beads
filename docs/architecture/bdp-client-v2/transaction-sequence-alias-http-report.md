# Transactional sequence alias HTTP integration

Frozen, clean source: `ea80c77de7cd86244096a5978b65690a7ba637d8`, tree
`c74f80cc11378dddbbddc7baaead44375be9d9cd`, ordinary parent
`70353341c7c51854c490b3c1746c00be13f1d068` (reviewed alias core + merged C1).
Checkout `/Users/dbox/repos/wt/vickie/vickie-tx-sequence-alias`, branch
`vickie-tx-sequence-alias`. Guarded fast-forward push and independent remote
verification completed at this exact head; `remote-verification.json` records the
matching role session and remote before/after. No main merge or readiness claim.

The commit has a Signed-off-by trailer, but no Agent-Signature trailer. Root has
clarified the latter was intended; any attribution-only successor will preserve
this commit and its exact execution evidence, rather than amend or relabel it.
No cryptographic signing configuration was created.

## Change and scope

Four files only (`source.json` contains byte counts and SHA256s):

- `packages/server/src/transaction-http.ts`: remove the unreachable five-line
  `sequence-resource-alias` catch; retain the actual unsupported-composition path.
- `packages/server/src/transaction-http.test.ts`: replace obsolete all-carrier
  refusal/no-prefix-effects assertion with committed prefix, retained member miss
  and independently committed tail; retain set/depth no-key-capture assertions.
  Add allocated creator + earlier alias assignment + opaque pinned Link endpoint
  response/schema/afterstate/repoint/reopen witnesses.
- `apps/bdptest/src/transactional-alias-installed.test.ts`: real raw HTTP over the
  existing installed IPC launcher, closed authored route set, authentic captured
  payload/contracts/auth/policy configuration, no runtime hooks. Sequence and
  receipt envelopes and current Resources use the existing normative schemas.
- `docs/design/transaction-implementation-progress.md`: append current bounded
  serving state while preserving historical slice descriptions.

No store/evaluator, Snapshot, protocol/schema/catalog, public admission, or Type
lifecycle changes. The caller knows private operation targets: this is not a cold
client discovery claim. Earlier core/old-reader evidence remains historical.

## Actual installed ledger

`packed-tx-alias.log/json` observes 36 exchanges, 28 before close +8 after durable
reopen, exactly two owner sessions. The configured canonical Scope is
`https://reference.invalid/acme/`; each session dials its actual loopback port and
sends the canonical Host. The actual binding is equal across reopen.

1. Seed two Beads with one batch and install the alias separately; read its
   initial Location (3 exchanges).
2. Sequence repoints the alias, updates its turn-time Bead, creates a Link with
   bare source and opaque pinned target; read both Beads and the Link (4).
3. Repoint original alias, read Location, replay sequence, prove prefix alias
   write did not rerun, and recover the same update via singleton key (5).
4. Install a previously unseen locator to B, read it, prove same-key equality;
   repoint it to A, read it, prove409 conflict and unchanged A (7).
5. Sequence deletes original alias, records missing-subject failure and commits
   independent tail; read alias404/tail, restore alias/read307, replay the retained
   miss, perform fresh-key update and read its current postimage, then replay old
   successful result (9).
6. Reopen: replay original success and failed sequence, replay singleton success,
   read original alias plus B/A/Link/tail (8).

The negative requests are real target discriminators (unknown alias404, fresh
locator conflict409, retained miss vs fresh-key success); no fabricated response
or derivative payload was used. Complete retained sequence/receipt object equality
is checked after live state has changed. The admitted pinned endpoint retains
`opaque-pin` exactly rather than substituting a live revision.

The test adopts the adjacent installed Set cohort's5s exchange/90s abort/120s
outer budgets and enforces its own60-exchange/two-session ceiling. These are test
budgets, not protocol limits or a general heap/latency proof.

## Verification

Pinned Node24.16.0 and pnpm11.20.0. Full exact commands/exit codes and elapsed
seconds are in `exits.json`; executable supervisors/scripts are adjacent.

- Preparation:59 HTTP tests passed. Typecheck first failed on test-helper
  `unknown` typing, then passed. Two suite-import failures used nonexistent
  `bead/link` schema names; corrected to currentBeadRecord/currentLinkRecord.
  A real installed seed request returned400 because putAlias was placed in batch;
  corrected to the required separate alias command. Those initial logs/JSON are
  preserved, not counted as successful evidence. The failed child exited0.
- Pre-freeze:426 tests/9files passed; typecheck, build, boundaries, format, lint,
  executable smoke, development smoke and historical Read seal passed. Lint exits0
  with inherited warnings; the new test's own non-null warning was corrected.
- Frozen source:425 tests/8files passed across HTTP, alias core, sequence,
  recovery, Read, RU member identity/executor and sequence executor (4.597s wrapper).
- Fresh packed installed TX:1test passed,36actual responses/two clean children
  (1.129s wrapper). Packed protocol/server freshly built from frozen source and
  installed offline with scripts disabled, then loaded from the installed entry.
- Same fresh installed payload:2selected RU tests passed,200unselected skipped
  (2.685s wrapper). These are actual retained retry cross-carrier positives and
  reference/category alias positives, not the entire RU matrix. Reference report
  binds the same frozen source and payload, says cleanupComplete:true and
  claimEligible:false. No unchanged full8k suite was repeated.
- Frozen typecheck/boundaries/format/Read seal passed; no source change occurred
  during or after packing. Final clean audit recorded.

Earlier process supervisors saved logs for every attempt; their shared exit index
was initially replaced per invocation. Current `exits.json` retains final and
frozen checks; preparation failures are independently retained in named log/JSON
files and reported above, not retrofabricated into a complete historical index.

## Bindings and cleanup

See `packed-summary.json` for complete safe binding and all36 response byte/hash
observations. Installed payload:
`495b3d9c1e6d6a078b91b60f6c41cfb4ab3f8928595467ec895d171d668d53a2`.
Installed contracts:
`80c3cc6befaa53f7949b91e00ef00c2658442de1ffc24dfdaf6a16cdc9235992`.
Runtime configuration:
`f3883ac9b60b3aa7f64014bc9b483401c9e13ac2c69ecba6780e8d2611bd441b`.

Packed protocol SHA256:
`2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.
Packed server SHA256:
`8fae8608a298a2f89822fcbcdbe2c214941fb16c31a8375cc136b1e87cd6f2ea`.

TX child PIDs36391/36392 both exited0 and subsequent OS existence checks found
neither. RU harness reported cleanupComplete:true. Process inspection found no
owned installed child/test supervisor remaining. No shared database was used;
disposable stores were removed. Credentials are absent from the safe summaries;
only their actual captured identity digests are recorded. No public profile
advertisement or conformance credit is inferred from these observations.

## Attribution-only successor

After reporting the absence to root, ordinary empty successor `82a0a0261ba042be3721a16bdccc37915a0e5fa1` adds
`Agent-Signature: Vickie; Codex; unknown-model; unknown-reasoning`. It has the exact
same tree `c74f80cc11378dddbbddc7baaead44375be9d9cd` and parent `ea80c77de7cd86244096a5978b65690a7ba637d8`.
Remote role session was checked again and fast-forward push independently verified.
No source bytes changed; no tests were rerun or relabeled. All execution above
remains bound to ea80. See `attribution-successor.json`.
