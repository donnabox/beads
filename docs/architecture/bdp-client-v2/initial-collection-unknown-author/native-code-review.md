# Independent native code review — initial collection unknown minima

No actionable findings in the frozen four-file draft. This is static source review, not a compiler, test, timing or installed-provenance pass. Native closure is limited to these exact bytes; actual Claude review, root findings disposition and separately released validation remain required.

## Binding and scope

Inspected actual worktree codex/vickie-initial-collection-unknown at2dfd2c6cc84a28b1dd336f4184c1a5f910ce8a96, complete gitdiff and accepted plan-v2/dispositions. Source-binding SHA256271aa8f79cf3b8bdf1d2d2939a81ce06d793323d658ee739f365a8a56717ac13; actual patch7ab49b685ea70c22d1e87c7ca58c5a7fff4e92a87507798cb82708dc9b432201. All679 tracked source hashes match the binding, exactly four owned dirty files, no untracked files. Before hashes match committed source. No main, conditional or Event branch integration occurred here.

Actual after-source hashes:
- transaction-read.ts: e76595b1da4530c4178a668b72e8bdf91ded7cdea5a6052219a2b10ebffedcb2
- transaction-read.test.ts: 10b690850b9fefd3ec4770e02d1baecbcac410de841e185614d1684cfa9897d0
- transaction-http.test.ts: 78583951db5d5d9f51b8aa03585da38f89082f6f2269306b4d5befa2c75b1958
- transactional-collection-minimum-installed.test.ts: f5924a853e5cf8a2748c2a4067e9425772b26061832968f60a0618d998f153a7

## Implementation semantics

transaction-read.ts86–115 adds exactly the initial noncontinued beads/links classification through existing own-data-descriptor access and proxy guard, with a derived MinimumCondition.kind. Existing capture/data-only/canonical-route receiving remains authoritative, so an accessor or invalid request is not made valid by classification. No duplicated HTTP route predicate was added: unchanged HTTP strict-kind admission naturally accepts the new kind.

The source-condition changes are correctly separated: the resource-not-found bypass is now the explicit original resource/missing allowlist, while successful eligibility is waitable for resource/initial-collection only. Defensive collection refusal semantics therefore stay strict. Materialized invalid-parameter/limit-exceeded returns retain precedence. Recognized minimum handling is unchanged because checkCurrentReadMinimum consults waitable only for uninterpretable positions. Unknown matching-context collection attempts now produce Unproved and flow into the existing fixed-deadline wait/catch-up-timeout path; foreign epoch/view still produces Unsupported. No storage/service/HTTP handler/discovery/profile code changed.

Spec general consistency5841–5843 and timeout5329–5332 support this reference choice; no universal interpretation of the Resource-specific mandate is introduced. Parser-invalid unknown-token500, Type/incident500 and continuationminimum-position-route remain explicitly residual and excluded. Tests do not convert their preserved outcomes into public qualification claims.

## Direct and HTTP evidence design

The new direct callback uses a completed one-position seed with two Beads/two Links and limit1, so premature success would need a retained continuation. Each of six collection/token cases checks literal current context, three wrapped attempts, waits[5,2], tickdelta7 and zero token generation. The raw owner entry inside each wait proves actual owner availability separately from attempt counting. Canonical unknown tokens are opaque,pos-01,pos-9007199254740992. Aggregate18 attempts/12waits and13 raw owner checks after first-wait abort are consistent with the loop and unchanged waiter. Preabort requires zero reads/waits; first-wait abort requires exactly one attempt, then abort before any second attempt. Service close is in finally; existing fixture cleanup remains intact. Read maximum32 fixture alignment matches existing service fixtures and is not a productlimit change.

HTTP callback topology is18 exchanges: seed1, timeouts8, residual500s4, materializedrefusals2, missing/hidden404s3. Eight timeouts produce eight waits under the controlled2ms seam; later residual/refusal loops must preserve that count. Existing Resource assertions remain; both initial collections and both methods are added, including empty Link collection in this HTTP fixture. GETIf-None-Match and HEADnonmatchingIf-Match cannot become304/412 because exact503 is required. The paired body/HEADcontent-length checks use actual GET bytes, schema-validated exactProblem, exactcontext/media/retry, emptyHEAD and absentgapmarker. Direct real HTTP fixture fetch honors the providedmethod. Numeric-looking token variants are checked throughHTTP as well as directly.

## Installed scope, deadlines and privacy

Only the first installed callback gains four opaque requests. Its existing38 exchanges become42; one child/store and eight real timeout requests total. Other installed callbacks, retained-row capture placement, lifecycle helper,80exchange cap,90swork,120scallback and5srequest timeout stay unchanged. New requests occur before retainedcursor capture, so the extra fourseconds cannot expire that later-created cursor. Every added elapsed assertion is>=1000ms and<5000ms, preserving the accepted no-timeout-expansion policy; wall-time success is still unobserved until execution.

Each new installed GET/HEAD checks actual seed/currentcontext, exact503Problem/retry, media, no gapmarker and GET-byte-derivedcontent-length, with emptyHEAD. Logging adds only fixed observation label,collection,method,elapsed andclaimEligible:false. No credentials, request headers, bodies, arbitrary errors or rawconfiguration are added to logs. Existing separately verified package/consumer provenance and cleanup supervisor remain necessary; sourceCommit alone is not elevated into installed proof. The emitted42exchange count still needs external audit as specified, rather than a new sharedhelper/count seam.

## Limits and optional observations

No new proof surface is needed to resolve this review. The old early-gate comment could be expanded for readability, but its current reference to excluded routes remains true and a comment-only extrafile is unnecessary. Source-controlled timing is not installed elapsed evidence; assertions and workload counts above are reviewed expectations only. The fixed waiter may re-scan collections repeatedly, a known accepted cost of this narrow change. Fresh static validation must report its own actual results.

No tests, product imports, services, builds, formatter, source edits, commits or push were performed by this reviewer. Frozen source and patch were rechecked before report write.
