# W1b independent native review

Verdict: **bounded pass; no new blocking finding** in frozen `f2ce17bafd1c3d218cf404b815e12e288983fe05`, against direct parent `459d32b67e3d8c2c4e4e667bc7481649ca33eee3`.

Source reviewed read-only in `vickie-read-update-w1b`; independent clone `bdp-read-update-w1b-native-review` remained clean at that exact head. This is source-private executable groundwork, not public profile qualification or a replacement for independent packed-package evidence.

## Scope and contract checks

The six-file diff adds exactly three scenarios and changes exactly their three coverage rows. Structural comparison found no removed or modified preexisting scenarios or other coverage claims. No runtime, server, protocol/schema, public admission, or Read evidence changed.

- `read-update.singleton.update-link` (`packages/conformance/matrices/read-update-v1.json:11496`): separately exercises properties-only and metadata-only updates with prior independently observed Link revisions; checks complete response state, static expected properties/metadata/identity/type/endpoints/provenance, and independent full Link GET equality. Unowned create/update/delete controls preserve both endpoint records and omit the source pair. The three catalog clauses are supported by meaningful observations.
- `read-update.singleton.owned-link-source-revision` (line 16564): owned create, properties update, metadata update, and delete each observe the owner independently, check its static unchanged properties/metadata plus expected fanout provenance and inline Link, and test the current reported revision with a successful guarded no-op and the previous revision with a real 409 revision-mismatch. These controls establish revision usability and advancement rather than merely echoed-string equality. Target records remain identical throughout. Deletion verifies absent ordinary/properties Link GETs and empty owned-link collections. The unowned contrast checks both endpoints and absence of source/sourceRevision. The result schema also prohibits those members on Bead results.
- `read-update.singleton.owned-link-no-op` (line 21022): reorders equal properties and metadata while supplying different actor/agent/message. Before mutation, static expected provenance is checked and full independent Link/owner records are captured. Both direct mutation response and subsequent independent GETs must equal those prior records, retaining opaque revision, committedAt, and provenance. Source revision, inline Link, and target retention are checked separately.

The test controls at `apps/bdptest/src/read-update-matrix.test.ts:678–720` corrupt eight captured assertion inputs only after real HTTP: source revision, inline Link state, Link and owner no-op provenance, no-op source revision, metadata, an illicit source pair on an unowned result, and target state. Each requires exactly one application and the expected failing assertions. These are meaningful oracle sensitivity controls, explicitly not additional wire observations.

## Independent validation

Pinned Node 24.16.0 / pnpm 11.20.0; completed build before tests. Independent focused run: **34 tests passed, 2 files, 23.53 seconds**, maxWorkers 2:

- `apps/bdptest/src/read-update-matrix.test.ts`
- `packages/conformance/src/cumulative-artifact.test.ts`

Logs: `read-update-w1b-native-build.log` / `.exit` and `read-update-w1b-native-tests.log` / `.json` / `.exit` in this directory. Both recorded exits are 0. No source changes or evidence regeneration occurred.

The separate clean actual installed-entry report `read-update-w1b-native-progress.json` records source f2ce17b, sourceDirty false, claimEligible false, 36 planned/completed sessions, 335 client observations, 36 pass plus 98 harness-error results, and 101 uncovered of 129 required contracts. No counterfactual was applied in that clean report. The new rows record, respectively, 45/331, 54/384, and 37/285 recorded actions/assertions (setup excluded).

Exact report bindings:

- catalog: `03ca39f24a988a16497d80a52c1fffad7e2c5dbd27a2a521679cc0183672b184`
- manifest: `61160d1a1acff0f6f875c47d7fae18ca6fd72cc4dee97b0bfe81b1eecd5b5784`
- fixture: `3a60db8714da2490ad67f9bf4f281f50da293393b31c6e1d5bec1ece12fd955c`
- exact configuration: `90961f97ac07547d6e6e59006e9f683dc32fb193618158a16039c51136a8e00c`

## Limits and inherited dispositions

The frozen registry counts 36 bounded and 28 whole-contract rows, leaving 101 uncovered. This review supports the three newly added whole-row claims; it does not reapprove every inherited whole-row claim. The inherited W1 Claude low findings, including the pin-equality coverage caveat, belong to the separately assigned successor and are not introduced by W1b. Preserve that distinction when reporting the aggregate count.

I read the W1b plan/report and inherited W1 Claude review. I independently exercised the emitted installed IPC/factory and actual HTTP route matrix, but did not repeat the author's packed-tarball cohort or broad workspace suite. The eight corruption controls demonstrate assertion sensitivity, not implementation-wide completeness. Public Read admission and existing evidence remain unchanged.
