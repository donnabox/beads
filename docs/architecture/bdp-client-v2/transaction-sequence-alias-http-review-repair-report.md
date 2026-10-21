# Alias HTTP review repair

Frozen source `b819ce2a18b558c847db3dfbdca893344822b060`, tree
`1addd90d330d5d9463b8c0a24f42d040b42c500e`, ordinary parent
`82a0a0261ba042be3721a16bdccc37915a0e5fa1`. Clean checkout
`/Users/dbox/repos/wt/vickie/vickie-tx-sequence-alias`, branch
`vickie-tx-sequence-alias`. Commit includes the existing Agent-Signature trailer.

Only three files changed: HTTP tests, installed alias test, progress prose.
Production, package metadata, schema, store/evaluator, conformance artifacts and
Snapshot sources are byte-identical to the parent. `source.json` pins each changed
file. No public admission, merge or readiness claim is made.

## Review dispositions

1. **Accept: independent exact scope positions.** First HTTP sequence creates two
   Beads in separate member commits:pos-2; alias assignment/replay/reopen stay2;
   fresh deletion advances to3. Allocated/pinned sequence creates one Bead and one
   Link:pos-2; another Bead advances to3; repoint/replays stay3. Installed seed is
   one *batch* commit despite two Beads, so its distinct ledger is1; update+Link
   adds2 to3; tail adds1 to4; fresh update adds1 to5. Every installed200 now pins
   this independent effect ledger, including replay and reopen. No effect count
   is inferred from matching response bodies.
2. **Accept envelope/alias witness; reject overstated schema claim.** The allocated
   sequence's alias member is checked with complete literal shape (position,
   outcome, canonical alias, captured creator ID), then immediate307+Location.
   The original first-response schema validation already required target: a
   dropped target could not pass it. This repair adds a direct semantic witness.
3. **Accept two local sensitivity controls.** The same equality helper accepts
   actual GET/member and retained/member comparisons, then rejects copied
   expectations with the old admission target or stale n:0. Both control names
   are reported only after their throwing comparison is checked. Existing live
   409 and retained-miss/fresh-key cases were already negative discriminators;
   this is evidence hardening, not an identified runtime alias defect.
4. **Accept prose/provenance clarification.** Prose describes fixture contents and
   links the external executed ea80 report with its exact SHA/payload. That proof
   exists and remains historical. `declaredSourceCommit` accurately labels the
   installed test's caller-declared source; the supervisor independently checks
   clean actual HEAD before packaging/execution. The captured target binding is
   still checked separately. Earlier report fields are not rewritten.
5. **Accept explicit supersession.** Name all three obsolete statements: alias
   refusal list, alias timing remaining-work entry and alias-marker reservation.
   Preserve their historical text and distinguish current serving behavior.
6. **Accept exact route inventory.** Remove unused delete-alias route (operation
   remains an exercised sequence member), and assert actual observed URL set
   equals the authored ten-route set. This proves neither a dead entry nor an
   unlisted URL occurs in the complete fixture.
7. **Accept context presence.** Every200 requires nonempty epoch/view/position,
   in addition to exact expected position. This does not invent context laws for
   untested error/redirect paths.
8. **Accept bodyOctets validation.** Reject absent, noninteger or negative byte
   observations before recording them.
9. **Accept distinct process witness.** Exactly two clean sessions must have
   different PIDs.
10. **Accept removing vestigial assertion.** Remove unrelated key-a create from
    the set/depth rejection test. Actual same-key set/depth correction assertions
    and the separate deep-sequence no-key-capture case remain intact.
11. **Accept pre-reopen allocated/pinned retry.** Add same-session sequence replay,
    exactpos-3, alias Location and Link full-body afterstate before reopen. This
    adds four requests to that separate loopback HTTP test, not the bounded
    installed fixture. Post-reopen comparison remains.
12. **Defer redundant replay schema calls.** Complete equality against an already
    validated response body preserves its schema shape transitively. No schema
    weakening or additional parser was introduced.

## Exact verification and budgets

Pinned Node24.16.0/pnpm11.20.0; commands, exits and monotonic elapsed times are in
`exits.json`, `checks.py` and `pack.py`. No failed checks in this repair attempt.

- Pre-freeze affected cohort:60tests/2files passed (HTTP59 +installed1),2.696s
  supervisor elapsed. Typecheck, build, boundaries, format, lint and historical
  Read seal passed. Lint retained pre-existing nonfatal warnings; no broad cleanup.
- Fresh frozen build, protocol/server packs and offline script-disabled install
  passed. Fresh installed alias1test passed in1.147s supervisor elapsed.
- Frozen HTTP59tests passed in2.533s supervisor elapsed. These are reruns of the
  same affected cases, not120 distinct tests.
- Installed fixture remains36 exchanges (28+8), two sessions, ten observed URLs,
  and two *local* sensitivity controls. No extra installed HTTP exchange was
  added.5s/exchange,90s cohort,120s outer and60-exchange ceilings unchanged.
- Production/payload did not change, so the root-authorized conditional RU rerun
  was not needed. Prior ea80 targeted RU proof remains at
  `../transaction-sequence-alias-http/packed-summary.json`, not relabeled to b819.
- No full8k suite or redundant425-server/whole-RU cohort repeated.

Fresh target payload remains
`495b3d9c1e6d6a078b91b60f6c41cfb4ab3f8928595467ec895d171d668d53a2`.
Contracts remain
`80c3cc6befaa53f7949b91e00ef00c2658442de1ffc24dfdaf6a16cdc9235992`.
Fresh runtime configuration identity is
`ed7f6ff78a2462ed0fb42b46a205a5c89b5a1ea70a96a79acc07fa7a78f15e63`;
it is identical across this run's reopen. Auth/policy identity changes from the
historical run reflect fresh fixture credentials, not production changes.

`summary.json` pins declared/actual source binding, tarballs, entry, complete
configuration digests,36 response hashes/byte counts/positions, and actual local
control names. PIDs45615 and45618 both exited0; subsequent OS existence checks
found neither. Stores were disposable and removed; no shared database was used.
No child/test supervisor remains. No stopped/inconclusive test is counted as pass.

The original ea80 and82a0 commits, reports and initial preparation failures are
preserved. This repair supplies new source-bound evidence, not retrospective
relabeling or whole-profile qualification.
