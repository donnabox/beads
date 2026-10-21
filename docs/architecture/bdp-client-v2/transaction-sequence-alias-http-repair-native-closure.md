# Alias HTTP review repair — native closure

## Findings and disposition

**All accepted behavioral/evidence repair findings are closed; no new actionable test/runtime defect found.** One Low documentation correction remains, already identified and owned by root: `docs/design/transaction-implementation-progress.md:1583` uses an absolute local `/Users/dbox/.../report.md` link, unusable to remote repository readers. Replace only its URL with the published fork report `https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/transaction-sequence-alias-http-report.md`. Keep historical ea80 source/payload attribution. This is a docs-only correction and does not require repeating executable checks.

## Exact source and scope

Reviewed frozen `b819ce2a18b558c847db3dfbdca893344822b060`, tree `1addd90d330d5d9463b8c0a24f42d040b42c500e`, against direct parent `82a0a0261ba042be3721a16bdccc37915a0e5fa1`, in `/Users/dbox/repos/wt/vickie/vickie-tx-sequence-alias`. Only three test/doc files differ. Production, schema, package metadata and conformance artifacts are unchanged. The author checkout is clean and was not edited or built by this reviewer.

Read original external HTTP review, root dispositions, exact diff, surrounding fixture flow and the repair's report/source.json/summary.json plus raw machine-readable test outputs. Original ea80 evidence remains historical; no alias-core re-review or Snapshot inference is made.

## Closure map

| Finding | Actual closure |
|---|---|
| F1 exact scope positions | Direct first sequence asserts pos-2 after two independently committed creates, unchanged by alias assignment/replay/reopen and pos-3 after fresh deletion. Allocated/pinned sequence asserts pos-2, then pos-3 after another Bead, unchanged by repoint/live/reopened replay. Installed every200 pins independently derived effect positions: one seed batch→1; update+Link→3; tail→4; fresh update→5. Alias writes correctly do not advance the position. The exact ledger supplements, rather than replaces, existing literal/full-body/alias afterstate discrimination. |
| F2 alias-member and direct307 | The allocated sequence's member1 is compared to a complete literal alias-result shape using the actual allocated creator ID, then an immediate alias GET must return307 with that ID as Location. This observes the assignment before repointing. Original schema validation already required target; the repair does not repeat the external review's overstated missing-target claim. |
| F3 local sensitivity controls | `sameResource` is used for actual GET/member and retained/member equality. Copied expectations independently change target to admission-time Bead A and postimage n to0; both must throw through that same helper before their control names are recorded. These are explicitly local assertion-sensitivity checks, not wire corruption or a claimed runtime mutant. Original real409/miss/fresh-key discriminators remain. |
| F4 fixture versus execution/source identity | Prose describes fixture content conditional on running, cites the original executed ea80 record and exact payload, and does not relabel it. New record field is declaredSourceCommit. External supervisor verifies actual clean HEAD before packing and separately captures target binding. Root's URL-only fix above completes remote usability. |
| F5 supersession | Current prose explicitly names the old refusal list, timing remaining-work entry and marker reservation as historical, rather than only issuing a blanket disclaimer. |
| F6 inventory | Unused direct delete-alias target is removed; deleteAlias remains an executed sequence member. Actual observed URL set must equal the authored ten-route set after successful complete execution. |
| F7–9 envelope/ledger/reopen | Every200 requires nonempty epoch/view/position plus exact position; completed bodyOctets must be a nonnegative safe integer; both successful child sessions must have distinct PIDs and clean exits. |
| F10 vestigial assertion | Only the unrelated key-a create was removed. Actual set/depth same-key correction witnesses and separate sequence depth/no-key case remain. |
| F11 live allocated/pinned retry | Added live sequence replay, unchanged pos-3, current alias307/Location and full Link equality before reopen. Existing reopened replay/afterstate remains. Four extra requests belong to the separate direct HTTP fixture; installed36 is unchanged. |
| F12 repeated schema checks | Intentionally deferred per root: equality against the first schema-validated complete body preserves shape. No schema narrowing or evidence claim is hidden by this disposition. |

## Independent bounded execution

Created a disposable exact git archive under `transaction-sequence-alias-http-repair-native/source`. Copied server emitted bytes from the repair's retained installed package into this archive; reused dependency resolution without install/build or author-output writes. The two changed direct HTTP cases were run with pinned Node24.16.0:

```text
node node_modules/vitest/vitest.mjs run packages/server/src/transaction-http.test.ts -t 'commits a prefix and independent tail|resolves earlier alias writes and mixed allocated' --maxWorkers=1 --reporter=verbose --reporter=json --outputFile.json=/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-sequence-alias-http-repair-native/focused.json
```

**Exit0:2 passed,0 failed,57 unselected/skipped.** Logs are `focused.log` and `focused.json` in that review directory. These are independent actual loopback HTTP observations using the retained installed production modules, not a new full59-test or36-exchange packed run. No review-owned child remains.

## Inspected fresh author evidence

`transaction-sequence-alias-http-review-repair/summary.json` SHA256 `58f07445e1846c37ab0ba3ec3ce1417574d5d7d3fd453000976f96cbf111f5e4` binds this exact clean source/tree/parent. Independently verified all three changed-file byte counts/SHA256s and both actual tarball SHA256s against source/package bytes. Its final HTTP JSON reports59 passed/0 failed; fresh packed alias JSON reports1 passed/0 failed,36 exchanges, two distinct sessions, ten URLs and both completed local sensitivity controls. Actual position values in the completed-response ledger match the independent effect chronology above. PIDs45615/45618 exited0 and were absent at review check.

Positive payload remains `495b3d9c1e6d6a078b91b60f6c41cfb4ab3f8928595467ec895d171d668d53a2`, equal to the original ea80 packed summary. Protocol/server tarballs remain respectively `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c` and `8fae8608a298a2f89822fcbcdbe2c214941fb16c31a8375cc136b1e87cd6f2ea`. Current runtime configuration identity `ed7f6ff78a2462ed0fb42b46a205a5c89b5a1ea70a96a79acc07fa7a78f15e63` is bound across this run's reopen; fresh random credentials explain differences from prior auth/policy/config identity. This supports the authorized decision not to repeat RU for unchanged production; old targeted RU evidence stays pinned to ea80, not b819.

No broad suite, package build, installed rerun, source correction, push or merge was performed by this reviewer. Review scope does not claim public admission, cold discovery, full-profile qualification or new runtime behavior. Root owns the local-link docs successor and final integration.
