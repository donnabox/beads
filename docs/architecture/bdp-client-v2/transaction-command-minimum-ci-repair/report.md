# PR123 CI boundary failure: one-file editorial repair

Draft only, not tested or committed. Parent `f2ac3f980fc8d35a3b485db04b813ad1b569d94d` / tree `74327f052cef6026b3ea10e186523ace9c87a128` remains unchanged. ONLY `docs/specs/bdp.md` is dirty. Root reported exact CI38645926223 failed two profile-catalog-boundaries tests (9266 passed,1 skipped); no rerun/green claim. This task read the actual spec and boundary rules before editing; it did not execute tests/build/services or modify a guard.

## Cause and repair

The new Read+Update paragraph linked forward to `#transactional-http-consistency`, owned by Part IV/rank3. `assertProfileLinks` rejects a rank2 source link to rank3, and `assertPartLinks` independently rejects any later-Part dependency. Consequently both existing baseline assertions fail:
- `catalog citations respect cumulative profile boundaries resolves internal links without lower-profile forward dependencies`;
- `catalog citations respect cumulative profile boundaries rejects model-to-Read dependencies despite their shared profile rank`.

The repair moves the exact operative MUST/three field names/identity/no-barrier/HTTP syntax and transport-header-body bounds/authorization/idempotency/carrier-validation/expectedRevision text into Read+Update's Mutation command preconditions. The operative rule core is byte-identical; applicability now says this profile and cumulative profiles. It is self-contained and adds no later-profile response/receipt requirement to Read+Update.

Transactional consistency points backward to `#mutation-command-preconditions`, restates the no-barrier/no-identity effect, and retains its receipt GET/HEAD minimum exception. The existing read-minimum paragraph and exact catalog excerpt `It never reports success with an older position.` are unchanged. No semantics, production, test, schema, catalog, heading, guard or allowlist changes. The pending cumulative obligation-mapping follow-up is still separate.

## Why local validation missed it

The accepted finite selection Q ran only `draft Transactional catalog strictly parses and binds every citation to the current specification text` in transactional-catalog.test.ts. That test validates the selected requirement citations/excerpts; it does **not** traverse every Markdown link or enforce profile/Part ordering. The new cross-reference's anchor existed and the old selected excerpt was retained, so Q could pass while the document-wide link boundary failed. `profile-catalog-boundaries.test.ts` was omitted from the targeted local selection, and both plan/code reviews incorrectly treated the resolvable forward link as acceptable. This was a concrete missed validation dependency, not a flaky test or a guard requiring an exception. Actual full CI exposed it; prior local successes remain historically correct but insufficient for this spec edit.

## Exact binding

- Original spec `c2bee88f7a29349f7e420f711d10f3acff1fb0559d012b0d111fd7ac21bd175d`; repaired `287a0a64bea0c1c3be51b5bb50d528283185ef7886a9c7402b4d4b5a7d5a0941` (392735→392774 bytes).
- repair.patch `ececd735849ae0033817273276c673adf4f72bbea9df0865f36b516c688356fd` from the clean PR head; before-bdp.md/after-bdp.md preserve both texts.
- source-binding.json `cc3f9bb7c37e1120ec88d7fb320d8188e227712c1a22be9ac4386992cd00e83e`.
- All other six original owned files still match final-source.json; original packets/bindings and failed CI remain untouched. Existing f2ac landing script/CI acceptance cannot establish readiness of this uncommitted successor; root must rebind after review/commit/newCI.

## Proposed bounded validation, not executed

Root recommends exactly four existing conformance files: profile-catalog-boundaries (21), read-update-catalog (7), transactional-catalog (9), profile-contract-plan (58). Static expansion totals95 cases: the contract-plan file has22 fixed plus parameter lists3/2/3/9/4/4/4/7. These are prospective text counts, not executed enumeration. Existing injected later-profile/later-Part negatives remain intact. CPU/filesystem only,0 HTTP/services/owners/listeners/children; unchanged inner limits and120s supervisor. Exact argv are in validation-proposal.json. No broad workspace suite, runtime/installed rerun, new test or guard weakening is proposed. Native+Claude delta review and root execution release come first. Subsequent exact-head full CI is required before landing. The earlier narrower22-case proposal is preserved in the .initial artifacts and superseded by this four-file recommendation; source patch is unchanged.

No tests, build, install, service, commit, push, PR edit, or remote mutation occurred in this repair task. Only guarded source/report writes and static text/Git inspection. Source is now frozen for the bounded delta review.
