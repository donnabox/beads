# Native review: schema PropertyChange evidence

No actionable findings in the reviewed three-file change. This is a bounded source/evidence review, not promotion or public qualification.

Reviewed frozen `c57f5a7a804b1b7ababf2d9195d7fdccbd37dee8`, tree `b7ffb7fb34ea5014ba824be945026a353eb94a70`, against direct parent `627b560c0daf1a76b292615718c86670255ac9af` in `vickie-schema-property-change`. HEAD, parent, tree and clean status were independently checked. Scope: the complete changes to `apps/bdptest/src/read-update-matrix.test.ts`, `packages/conformance/matrices/read-update-reference-continuation-v1.json`, and `packages/conformance/test-support/read-update-coverage-v1.json`, with relevant unchanged runner/client seams and selected numeric catalog clauses. Accepted authority: plan v2 SHA256 `76fa65b2586ab0f5baa09523744e5dd27f4a3c22947d6364e8502aaad9f958ce` and `read-update-after-r-v2-root-acceptance.md`.

## Assertion and control audit

For both Bead and Link, the existing successful create still independently checks identity, child Type, complete properties and metadata, absence quartet, and Link endpoints. Its added capture is the URL-valued `/resource/id`. The converted installed GET independently checks the current record and literal `{n:0,enabled:true,detail:"ok"}` / `{phase:"created"}` before capturing its full `/body`. Runner assertions precede capture (`packages/conformance/src/runner.ts:1096–1134`); this does not introduce unsupported raw body capture.

Each new refused update carries numerically admissible `-1` and the effectful metadata companion `phase:"updated"`. The full configured 422 Problem names the inherited base Type/schema minimum and resulting-document `/n`. The next GET compares the complete record to the independently validated preimage. The correction changes only the numeric leaf to `1` and uses a separate key; both response and subsequent GET independently check full properties/metadata and Link endpoints. Corrected response absence quartet is present at its actual `/resource` and top-level pointers, with 13/15 assertions. No selected clause is silently narrowed, and R remains partial.

The four controls at app lines 5124–5210 operate after actual target execution. The existing client seam checks selected positional action input against executed input, clones only assertion input, and leaves the recorded real output intact (lines 521–575; exact client lines 143–172). Runner failure stops the selected scenario, making its last recorded client observation the target of the control. Tests require one application, exact named failed assertion, real status/write completion, original client value, and completed cleanup. The raw schema-location control changes the declaring schema to the child while keeping the actual 422 exchange. These are assertion-sensitivity controls, not alternate authority behavior.

## Independent preservation and budget audit

Read-only Python JSON parsing and raw object-span comparison against `git show 627b560c:<path>` independently found:

- All four other continuation scenario blocks unchanged byte-for-byte; all 84 untouched old R action blocks unchanged byte-for-byte. Only the four approved old create/read actions change, plus eight additions.
- The compact JSON hash of the parent's 84 unchanged actions is `393f9c3cab0cbfd501e863a75a243a28fa40806c648993eaf2068916cf488c20`, exactly the final source guard. Thus the normalization-to-source hashing repair does not hide parent drift. Initial failed static evidence is retained separately; it is not presented as a pass.
- Manifest 899400 bytes / 23502 JSON nodes; R 96 actions / 687 assertions. Combined R+N 148 / 1060, N unchanged at 52 / 373. Tight manifest headroom is 2495 bytes / 44 nodes.
- Coverage 712792 bytes; only R's reason and bindings change, 619→687. Whole count remains 33, R false and claimEligible false. The two accepted coverage ceilings are 713143 and 712950, leaving 351 and 158 bytes. No other ceiling or runtime/parser/deadline limit changes.
- Derived older 16-action binding partition is 100 / 5546 formatted bytes; new eight-action partition is 66 / 3700. Third-partition exclusion preserves the earlier 48-action guard. The only action-ceiling amendment is the accepted 140→148; assertion ceiling 1300 remains.

## Evidence inspected, not rerun

No tests, builds, servers, credentials or source writes were performed by this reviewer. Actual commands were read-only `git diff`, `git show`, `git rev-parse`, `git status`, `rg`, `sed`, and Python parse/hash/count comparisons. All three source file hashes independently match source-audit.json. The two full written report file hashes independently match summary.json.

Read the completed source-bound `schema-property-change/{summary.json,complete.json,source.json,source-audit.json,static.json,focused.json,installed-ru.json}` and actual shard/selected-row reports. Author evidence records 22 static tests, 20 focused tests and 23 installed tests passing; installed JSON specifically records all four new controls passing. Exact R report is pass with 43 raw exchanges, 53 client actions and 687 passing assertions; N is pass with 21 raw, 31 client and 373 passing assertions. Both reports name c57f5a7, clean source and completed cleanup. Both aggregate shards name this same source; 49 baseline and five continuation authored rows pass. Summary records no remaining installed children, unchanged payload `1b12629f…`, runtime configuration `1720b63b…`, and route digests `3224c31c…` / `272c31e6…`. These are inspected author executions, not independent reruns.

The run remains diagnostic: 54 authored/observed rows, 33 whole, 96 not whole, 75 omitted required rows. Raw corrected-response-to-GET opaque revision linkage remains explicitly unobserved. No automatic R whole credit, public profile claim, or readiness decision is made here.
