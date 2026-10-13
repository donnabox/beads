# Native review — bounded W1 follow-up

Frozen `b6e671195ce2897e53f25f0b4f8dbb8c3b8b7c00` against parent
`f2ce17bafd1c3d218cf404b815e12e288983fe05`. Source checkout
`/Users/dbox/repos/wt/vickie/vickie-read-update-w1-followup` remained unchanged.
Independent clone: `/Users/dbox/repos/wt/vickie/bdp-w1-followup-native-review`.
Read the seven findings in `read-update-w1-claude-review.txt` and the frozen author
report. Reviewed the six-file delta; no runtime, protocol, public schema, runner,
admission or sealed-evidence edit is present.

**No actionable finding in this bounded follow-up.** The reviewed changes close
six of the seven lows or accurately narrow their claims. Alias workflow trimming
remains explicitly deferred; it is not silently described as repaired.

## Dispositions

1. **Alias sensitivity — closed.** Both new controls first await the existing actual
   HTTP executor. Fabricating the alias revision changes only assertion input and fails
   exactly `put-alias.body-schema` and `put-alias.alias-result`; fabricating the first
   independent target revision after the put fails exactly `put-alias-a.postimage`.
   Each requires one substitution and a failing row. The target control's method and
   target-binding conditions avoid corrupting the pre-put baseline. Original real
   observations remain captured. The put-creates reason now accurately cites the
   put-repoints workflow's sensitivity rather than inventing a row-local control;
   those two action arrays remain structurally identical.
2. **Omission reasons — closed.** No row-local scenario no longer means no evidence
   anywhere. Discovery aliases and patch validation explicitly identify existing
   cross-row observations without awarding credit. The W1b update-link/owned-source
   rows already have their own reasons at the parent; this follow-up does not replace
   or inflate those assessments.
3. **Opaque-pin equality — closed by downgrade.** `create-link.wholeContract` alone
   changes true to false and is removed from the exact whole-ID assertion. Observed
   storage/echo/no interpretation remain accurately named; equality comparison is
   explicitly unobserved. Independent inventory count is 27 whole of 129 required,
   leaving 102 uncovered, across 36 authored scenarios. This review does not infer
   whole-contract credit merely because all 36 authored scenarios pass.
4. **Metadata non-object — closed.** The added request actually replaces the metadata
   root with numeric 5 through the installed update-bead route, requiring 422,
   `validation-failed`, and the unchanged canonical Problem schema wrapper. The
   immediately following independent GET compares the whole record to the preceding
   captured postimage, whose properties and metadata were independently asserted.
   Independent raw observations show the transmitted 105-byte JSON body, actual 422
   diagnostic `resulting metadata must be an object`, and equal pre/post GET records
   including revision. The row remains incomplete for effective-Type-schema refusal.
5. **Alias workflow duplication — explicitly deferred.** No inventory-contract redesign
   or scenario trimming occurs; there is no correctness regression or claim that this
   maintainability observation was implemented.
6. **Port failure cleanup — closed for the identified leak.** In both test helpers port
   allocation now precedes temp-directory creation. The Read oracle's configuration
   construction and child launch are inside its existing finally-cleanup scope. No new
   production lifecycle behavior was introduced. Static ordering establishes this
   specific port-failure property; no synthetic port-allocation failure was injected.
7. **Fixed client GET routes — closed.** Only known fixed targets referenced by actual
   `installed.exact-request` GET actions enter that route leg; unconditional fixture
   expansion and standalone Directory GET are removed. Native HTTP action routes are
   unchanged, deduplication still runs, and captured canonical Bead reads retain the
   existing Scope/identity gate. Adding an unused Link fixture binding leaves the exact
   configuration digest unchanged in a passing real run. This control proves omission
   from the bound route set; it is not represented as an attempted forbidden wire GET.

## Independent verification

Pinned Node24.16.0 / pnpm11.20.0, frozen install, build before emitted child tests.
External evidence directory: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/w1-followup-native/`.

- PASS **48 tests / 3 files** in 32.54s (`tests.log`, `tests.json`): matrix32,
  installed Read oracles11, cumulative artifact5. Includes actual unfiltered matrix,
  both new alias controls and unused-binding control.
- PASS build, typecheck, dependency boundaries, changed-file format/lint, diff check.
- Captured `report.json`: exact source commit above, sourceDirty:false,
  claimEligible:false, 134 result rows =36pass/98absent harness-error;
  129required/36authored/102uncovered, 36sessions, 337client observations,
  44setup observations, 21native Read observations.
- Exact configuration digest reproduced:
  `f04e98d80971e813b73d1b57d225c0a6bfb0669647d6bbf9b026a908bbf37f26`.
- Catalog `03ca39f24a988a16497d80a52c1fffad7e2c5dbd27a2a521679cc0183672b184`;
  manifest `fcc9f68af721d9ea6725a83e29c24d4cc375fc9dbd653180c2e0b924e608a7b6`;
  fixture `3a60db8714da2490ad67f9bf4f281f50da293393b31c6e1d5bec1ece12fd955c`.
- Independent emitted-install payload identity:
  `d12573ef9a805e8ec5231486d3ddee5fdc23e27cfd2810fa7f566bd44afcbac8`.
  It differs from the author's fresh-packed identity `6503d708…`, as in the earlier
  emitted-versus-packed runs; this review does not relabel one installation as the other.

No failed/interrupted runs, source changes, shared stores or outward writes. Both
source trees are clean and no test-owned process remains. The author's 847-test run,
fresh-packed 53-test run and unchanged 89-row evidence verification were not rerun here;
those are author evidence. Existing W1b assessments were exercised as regressions but
not newly qualified by this delta review. Public profile eligibility remains false.
