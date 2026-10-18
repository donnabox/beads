# Current Type/incident minimum — native code review

## Findings

### P2 — new Type-ID gate executes caller data before the authority capture

**Source:** packages/server/src/transaction-read.ts:30–36 and 198–204. **Affected surface:** direct private minimumFor.perform (and readAtMinimum's forwarding to it); the HTTP path has already parsed a canonical request.

The newly admitted Type arm evaluates `request.id.startsWith(scope)` before calling the authority facet. The existing authority is supposed to reject accessors using descriptor-only capture (`authority-read.ts:297–314`, snapshotRequest at 315–330) and to deliver receiving errors through its observed Promise (`1025–1124`). A Type request with own primitive kind/resource and an own getter for id now invokes that getter first. An object-valued id with a callable startsWith runs that method before the later string validation. An omitted/numeric id can instead throw synchronously at the direct facet call. Parent dfafd rejected this Type branch without reading id and returned a rejected Promise. This is a newly exposed receiving regression, not a finding about legacy request.kind access.

Smallest source-grounded counterexamples (not executed by this reviewer, per the read-only/no-test assignment):

```ts
const request = {
  kind: "resource", resource: "type",
  get id() { calls++; return `${scope}types/task`; },
};
// Gate reads getter; authority subsequently rejects the accessor too late.
const methodRequest = {
  kind: "resource", resource: "type",
  id: { startsWith() { calls++; return true; } },
};
// Gate executes startsWith; authority subsequently rejects non-string id.
// With id: undefined or 1, the direct perform call throws before returning Promise.
```

**Impact:** caller code runs before the authority's data-only request boundary; malformed direct inputs no longer retain Promise delivery. This does not show an HTTP parser bypass, data corruption, or a new public feature.

**Bounded fix:** inspect the newly required Type ID through a descriptor-safe, proxy-safe check and require an own primitive string before its prefix comparison, or reuse a safe capture seam without broadening shared authority API. Do not read an accessor and then type-check its returned value. Keep the full canonical Type validation with the existing authority, preserve known external-Type rejection and ordinary external reads, and return rejected delivery rather than a synchronous exception for these malformed ID cases. Add targeted direct-facet no-getter/no-method and malformed-ID Promise controls plus valid local/external comparators. This can stay in the owned Read implementation/test files; no general kind-dispatch redesign is requested.

### P3 — progress document points to an unavailable local evidence path

**Source:** docs/design/transaction-implementation-progress.md:1668–1669.

The new paragraph says evidence is recorded in `transaction-current-type-incident-minimum-implementation/report.md`, which is an out-of-tree local log location rather than a resolvable repository artifact. The same document now correctly uses a commit-pinned report for the preceding compatible-missing work. Publish the finalized report through root's design-document workflow and replace the bare path with that exact durable reference; do not move, rebind or relabel the frozen evidence. Root already identified this documentation repair separately.

## Review scope and mechanism

Reviewed exact head `99c5a1faa4ad2bbb2dc8d1635ba9e604737473f0`, tree `e491847edc5b17a26d373cf69349bc6ef304d536`, direct parent `dfafd82290fb6f06db1e01522a3684a21d9813a7`, in `/Users/dbox/repos/wt/vickie/vickie-type-incident-minimum`. Working tree clean. Read all six-file diff (785 additions / 28 removals), accepted plan v2/root acceptance and relevant existing authority/request/HTTP/service/consistency/installed seams. The public, Type-lifecycle and retained-minimum boundaries remain excluded.

The production change otherwise delegates to the existing mechanism: one predicate replaces both admission expressions; current Type descriptor, initial Type inventory and initial bead-links are admitted; properties remain bead/link only; collection and incident continuations and alias resolution remain refused. For well-formed primitive Type IDs, canonical trailing-slash Scope plus startsWith excludes known external IDs, and existing canonical Type validation remains later. HTTP early400 and missing/control-refusal ordering are unchanged. Type body constancy does not bypass the owner entry/minimum condition. No service/wait/store/schema/fixture/public-export or claim flags change.

Independently compared dfafd source slices: ordinaryDeliveryOptions, the compatible observed404 callback, and the entire ordinary composed facet are byte-identical. The installed test file is an exact byte prefix of the new file, so both older scenarios/helper are preserved. The one old Type-negative assertion is replaced by a known external Type negative, retaining ordinary ignore-condition and abort controls. Only the three newly supported routes leave the old HTTP refusal table.

## Oracle assessment

- Source tests exercise a real additive in-Scope descriptor while retaining external descriptors and their ordinary reads; exact ordered three-Type summaries; both/inbound/outbound/default complete incident records and self-Link deduplication; unchanged Type bodies with a real later position; known external Type and actual continuation rejection.
- Type inventory and incident tests distinguish above-limit413 from valid-behind rejection, count zero token allocations before eligibility, then allocate after a real commit and follow the retained page ordinarily across a later change. Retained snapshot/current context distinction remains explicit; minimum on that real continuation still refuses.
- Controlled wait/cancel tests use real owner writes between attempts, both foreign errors, finite timeout, pre-aborted actual Request with zero DatabaseSync.prepare calls, and drain/close. They do not promise interruption of synchronous SQLite work. Compatible absent/deleted anchors and empty incident pages are separately checked; incompatible missing cases remain unqualified visibility gaps.
- Actual HTTP tests cover all three surfaces, canonical Host with query preserved, GET/HEAD, eligibility before 304/412, both foreign kinds, timeout/Retry-After, unacceptable Accept, unauthorized handling, malformed/repeated/unsupported controls, compatible404 and incompatible gaps.
- The new installed test independently validates the five-result seed receipt and literal authored fields before reusing opaque Link revisions; owned-source revisions differ from the corresponding initial Bead revision. Its complete incident projections compare against those validated Link images. Descriptor expectations come from the fixed fixture artifact; inventory expectations are literal and ordered. Exact pos-1 and context/no-ETag checks accompany positives and minimum refusals. HEAD bodyOctets, canonical Problem schema/type/status, Retry-After, direction/empty projections, count twenty and the exact seven URL set are asserted. This is a bounded private witness, not full profile qualification.

## Evidence inspected and execution limits

No tests, builds, services, network, credentials, source changes or publication were performed by this reviewer. Read-only git/text/hash/JSON inspection only; this review document is the only written artifact. One metadata inspection command assumed sensitivity.json was an object rather than an array and exited1; corrected inspection read the array successfully. This was a reviewer output-parsing mistake, not a source/test failure.

The author's focused-final.json records 348 passed, zero failed; an earlier focused-first failure in the Link oracle is retained and followed by corrected/final runs. Three separately preserved source mutants fail the intended assertions: ignored Type minimum resolves where rejection is required; skipped direction filtering produces the wrong exact rows; stale context fails the advanced-position comparison. The restored selected run passes ten tests (34 other tests skipped); these are author-executed sensitivity checks, not reviewer-executed probes or installed mutant evidence.

The frozen packed-minimum.json records three passed tests. Summary bindings identify clean exact99c5 source for new Type/incident20 exchanges/one child and separate unchanged collection38/one child plus missing10/one child. All report closed with childExitCode0; remainingInstalledChildren is0; claimEligible remains false. All six summary source hashes match actual frozen files. The new scenario's seven-URL and twenty-request assertions passed. This reviewer has not rerun this evidence and launched no processes requiring cleanup.

Protocol tarball: `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`.
Server tarball: `fa2e9ed46abfc013b3f36d1640d029f360381597f6051de635fccfd5289b39a6`.

Read the final author report and cleanup record after completion. The author explicitly preserves the preliminary dirty emitted run as preparation evidence, distinct from the clean frozen packed run; records the initial missing default metadata oracle failure and subsequent correction; and acknowledges both pending review findings. Cleanup records all three installed child PIDs absent, no path-bound Node processes, removed mutant worktree and clean source. The receiving finding remains actionable despite the existing green cases.

Final inspected artifact hashes:


- report.md: `52f2d21cb466696dc56047977d332c371550d4964fbf9611b1545f1388d2f1be`.
- source.json: `0eed75da546e4f390aced78abd9a3d8b88453acb01d00d98f9a51c58903e5f24`.
- summary.json: `f358f5ffb2bb9f986844b50efc29338e86b6cf7167cc6b2fa157f3cc7a1d3fc8`.
- cleanup.json: `a90d0e18c30976f279b1969ceffd5b0739c908dfcb994c7d5beb7c725ea74887`.
- sensitivity.json: `cc554e34d212f852ebf41b4708e88d3e3e585d8b91ea4a149d1c240791c1884d`.

Root owns disposition and a separately frozen repair successor; no readiness or full qualification is recommended at99c5.
