# Read+Update guarded overlap — bounded private witness plan

Proposed follow-up only; receipt-context serving remains first. This adds a private observation of ordinary Read+Update guarded requests, without changing the owner, Transactional execution, public admission or cumulative qualification. No implementation, tests, builds or new external review are authorized by this plan.

## Basis and purpose

Inspected clean checkout /Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence at db5a4d6a41814e59704ebef60fda02a2d57e9ad2/tree2e16329e195eb8214836de20340a5887236d78f0, ordinary parent3df82ac480f50c8f486b4e456336366e7acbdf87 atop verified main88c9408e908590ab1a4c64c9b9dee1a8b705999c. The latest successor changes only one test timeout; it is not assumed merged or CI-green. Reanchor to actual accepted main after the higher-priority serving lane before authoring.

Predecessor: read-update-guarded-overlap-feasibility.md. Existing normative basis: docs/specs/bdp.md2091–2114 (singleton body/key/direct-result/replay/cache rules),3498–3499 (stale expectedRevision), and6774–6777 (guarded race, stale guard and idempotent retry obligations). The catalog already contains read-update.singleton.expected-revision-race at packages/conformance/catalog/read-update-v1.json395. Its cumulative test-support coverage entry4134 remains omitted/wholeContract:false. The separate development coverage209–217 and development.race adapter at apps/bdptest/src/development-matrix.test.ts139–144 already exercise a Promise.all development route; that is not an installed canonical-wire or cumulative qualification grant. Preserve all those records unchanged.

Concrete value: demonstrate that two outstanding client requests against the actual ordinary installed RU owner, with distinct keys and one shared current revision, produce exactly one committed update and a retained revision-mismatch refusal. Then verify each key's replay and the complete final Resource. This can supply a future narrowly attributed evidence input; it does not itself retire the cumulative omitted requirement.

Actual owner path: installed-reference-process.ts → installed-reference.ts19–29 → development-runtime.ts457–477 → principalAuthority.mutate → member-executor.ts593/its evaluator call → recovery-store.ts799–801. The synchronous transaction wrapper381–390 surrounds evaluation and commit; resource-evaluator.ts1023–1027 checks expectedRevision against the current Resource. No Transactional service/receipt/sequence replacement or server scheduling hook is proposed.

## Two owned paths, one new callback

1. New apps/bdptest/src/read-update-guarded-overlap-installed.test.ts: one literal callback, small file-local bounded HTTP observer, schema/body/state/replay assertions, safe report and cleanup.
2. docs/design/update-implementation-progress.md: accurately record the partial private witness and its eventual actual source/evidence, preserving cumulative/public gates.

No catalog, coverage, manifest, runner, source helper, launcher, package, workflow, runtime, owner or evaluator edit. The emitted counterfactual lives only in an external disposable dependency-copy directory. No other case is registered. Exact new title:

`preserves one guarded winner and both retained dispositions across overlapping ordinary Read+Update requests`

Exact selection: the new file with anchored literal title above, --maxWorkers=1, verbose plus JSON reporter. All prior BDP environment overrides are cleared except the explicitly pinned test-local entry/output bindings approved for that command. An optional file-local BDP_RU_GUARDED_OVERLAP_ENTRY chooses the actual external consumer entry; default is the existing emitted installed-reference-process.js for ordinary full CI. The test reads actual local Git state and actual entry bytes; it must not accept an arbitrary supplied source-commit label as attestation.

Reuse availablePort, referenceConfiguration and launchInstalledReference from scripts/installed-reference-observation.mjs17–102 unchanged. Do not call its large observeInstalledReference scenario. Use the factory's actual read-update/no-History ready binding and provisioned reference task Type; no Type inventory traversal claim. The temporary availablePort reservation is one additional short-lived TCP listener lifetime per invocation, not HTTP or a second product owner.

## Exactly what overlap is observed

Use two independent Node ClientRequest objects with agent:false and Connection:close for requests A/B. Construct and end A, attach its rejection/fulfillment observation immediately; construct and end B in the same JavaScript turn and attach its observation before awaiting either. Await both through Promise.allSettled, not early-rejecting Promise.all. The request helper must return immediately; it must contain no await/queue before constructing the request.

Record bounded monotonic event ordinals, not wall-clock simultaneity: invoked, end-called, socket-assigned, connected when actually observed, finish, response-start, response-end, error, settled and socket-close. At most one event of each kind per request except explicitly bounded error information; maximum32 events for the pair is sufficient and must be enforced. Assign local socket IDs without dumping socket/header objects. Require A and B invoked/end-called before either settlement callback, separate request identities and separate assigned socket identities. This proves overlapping client-outstanding calls on two requests; it does not imply simultaneous server admission.

Record the actual milestone order and derive whether both connection or finish events preceded the first response only when both events were observed. If not, record false/unknown as appropriate; do not fail a correct owner merely because the OS/Node schedule serializes the wire. ClientRequest finish means handed to the local transport, not server receipt or admission. No sleep, artificial server barrier, automatic repetition, quota increase or claim of concurrent synchronous owner evaluation is permitted. A stronger requirement for both requests admitted before either response is a different experiment needing separate review; this plan does not promise it.

The winner is nondeterministic. Never pin A or B as the winner. This single schedule does not prove all possible races, distributed writers, crash/reopen durability, or Link/owned-Link behavior.

## Ten-request positive ledger

Scope is https://reference.invalid/acme/. Resolve discovery/directory links from genuine responses and parse them with existing RU parsers. Guard returned targets to the exact canonical Scope and the discovered createBead/updateBead identities, with no userinfo/query/fragment, length and byte limits. Do not generate an operation path and label it discovered.

| # | Request | Key and oracle |
|---|---|---|
|1|GET configured Scope|204, actual service-desc Link, non-attesting marker|
|2|GET that descriptor|200, parseReadUpdateDiscovery; exact Scope/read-update/no History|
|3|GET discovery.operations|200, parseReadUpdateOperationDirectory; retain create/update targets|
|4|POST discovered createBead|key overlap-seed; create beads/overlap of locally provisioned types/task with properties n:0;200, full RU mutation schema, outcome created; retain actual Resource/revision R|
|5|POST discovered updateBead, invocation A|key overlap-a; same Resource and expectedRevision R; replace /n with1|
|6|POST same update target, invocation B|key overlap-b; same Resource and expectedRevision R; replace /n with2|
|7|GET actual seed Resource id|200; full parsed Resource equals winner postimage|
|8|Replay request A exactly|same key/body; status and complete parsed response equal original A|
|9|Replay request B exactly|same key/body; status and complete parsed response equal original B|
|10|GET same Resource|200; complete Resource still equals winner, including revision/provenance|

For5/6 first validate bounded transport completion, acceptable200/409 status class, media/cache/marker and the corresponding full RU mutation-result/Problem schema. Every200 must have outcome updated, the exact requested n and seed id/type/metadata/owned-Link plane, an actual revision different from R, and a valid complete postimage. Do not require the new revisionProvenance to equal the seed's previous-version provenance. Both requests use effectful distinct values, not semantic no-ops.

Then assert named `overlap:exactly-one-winner`: status multiset exactly[200,409]. The loser must equal the current source's four-field direct Problem:
type https://github.com/gastownhall/bdp/problems/conflict, code revision-mismatch, status409, retry after-state-change. resource-evaluator.ts139–160 constructs these fields; member-executor.ts196–201 rejects positional fields in stored dispositions. Do not add sequence operationIndex/name or Transactional receipt/Scope fields. Successful singleton envelope is the complete parsed RU mutation result, with no receipt wrapper. Both mutation dispositions require private,no-store, correct media type, no ETag or Location, and no three Transactional context headers. Replays compare complete original bodies in memory, not regenerated expected opaque tokens. Reads7/10 compare the full winner Resource, not only n.

The seed literal is94UTF8 bytes, SHA-256 b76738409399066a7a7e9a05124d345773db6618130b419a2eb8261e57b8927e:
`{"id":"beads/overlap","type":"https://reference.invalid/acme/types/task","properties":{"n":0}}`.

Each compact update body is135 bytes plus L, the actual UTF8 length of JSON.stringify(R), using the retained canonical subject and property order bead/expectedRevision/propertiesChange. A/B differ only in value1/2. Record actual materialized byte counts and hashes after R is observed; never fabricate a fixed revision-length body hash. Positive total request-body bytes634+4L. Each request remains within the existing65536-byte request ceiling, checked before dispatch; header ceiling16384. This gives a conservative positive body allowance262238bytes including both replays, without assuming an opaque token size. The fixture's actual lengths should be much smaller, but that is not execution evidence.

## Minimal bounded observer and teardown

Do not use requestInstalled's unbounded response-text accumulation. The new file-local node:http reader uses known loopback listener host/port with exact granted URL path, explicit Content-Length, Accept:application/json, no redirects/retries and no Agent reuse. Count dispatch before end(body). Enforce5s absolute exchange deadline from invocation (destroy request/response on expiry),16384-byte response-header bound and32768-byte streaming body bound before appending. Attach error/aborted/premature-close handling, one-shot settlement, and clear timers in every terminal path. Retain request handles so outer cleanup can destroy both outstanding requests and await their settled promises. Parsing is after complete bounded bytes, with parsing errors rejecting rather than escaping an event callback.

Use a30s test-local work deadline covering startup/requests, followed by the existing bounded10s launcher close; callback timeout45s leaves cleanup room. Do not increase the launcher's existing10s ready/close bounds or any product timeout. An external command supervisor remains120s. These are proposed test-local bounds to freeze in the measured draft, not established elapsed times.

Always preserve the primary failure and cleanup failures. Finally destroy/drain active requests, await any outstanding pair, close/reap the child even on schema/assertion failure, then remove the owned bdp-ru-guarded-overlap- directory. Register handles immediately after acquisition and reuse the existing launcher's own failed-start cleanup; do not invoke the launcher again. Emit a safe bounded record after teardown even when assertions fail: exact entry realpath/hash, genuine ready identity fields, actual local source/dirty state, request count/ordered association, status and body fingerprints, named oracle failure, pair event order, child PID/exit/signal and directory presence. No bearer credential, raw configuration, arbitrary exception object, socket object or request-header values are logged. Full schema/replay/state comparisons run in memory; fingerprints alone are not claimed to prove those bodies. External review checks actual selected result and final PID/temp absence.

## One emitted guard control; sixteen total local requests

After the positive succeeds, copy its exact external consumer's server/protocol and complete production dependency closure into one isolated external derivative, retaining source and positive bytes. Bind all copied regular files and symlink resolutions to the accepted consumer; don't install packages inside the copy or run pnpm there. The source test runs from its unchanged checkout with only its actual child entry redirected to the derivative.

At the emitted resource-evaluator.js equivalent of1023–1027, omit only the expectedRevision mismatch guard. Exact before/after bytes and a one-site patch count are required; no TypeScript mutation/rebuild, new runtime hook, scheduler change or source binding override. Verify the child's actual relative module chain reaches that file and bare dependencies resolve through the copied unchanged graph. Hash the derivative's own runtime payload and compare the genuine ready payload to it; never demand that a mutated payload equal the positive one.

Run the same single callback once. The first six requests are identical in structure; both updates should be200 complete valid updated postimages. The unchanged named[200,409] assertion must fail there, before reads/replays7–10. Both pair requests must be settled and recorded before this assertion. Accept only one failed callback with that precise oracle and the two genuine200/schema-valid update results; transport errors, wrong media, generic timeout or cleanup failure are not a successful control. No second mutant or retry is authorized.

Local dynamic ledger: positive10HTTP/1child + negative6HTTP/1child =16HTTP,2callback executions,2ordinary RU owners/store directories,2product listener lifetimes plus2short-lived port-reservation listener lifetimes (4TCP listeners total). At most2HTTP exchanges outstanding per invocation. No in-process server fixture, hidden requests, reopen or polling. Negative body bytes364+2Lnegative, where its independently seeded revision may differ. Combined conservative body allowance393404bytes; total response-body allowance16×32768=524288bytes. Actual measurement must record materialized body counts and milestone completeness separately from these maxima.

Avoid an additional source-built positive just for duplication: after static checks and review, root may freeze an ordinary candidate and run the one fresh packaged positive then the emitted derivative. The new callback will also execute once against emitted workspace bytes in full exact-head CI (10HTTP/1child); that CI execution is separately attributed. If root instead requires a precommit source-built positive, amend the local ledger explicitly to26HTTP/3children before release; it is not implicit permission.

## Installation, review and acceptance boundaries

Before executing: actual merged-parent reanchor, two-path frozen draft, exact selection/body/cleanup packet, source review and actual Claude review, then root release. PinnedNode24.16.0/pnpm11.20.0 and required offline setup/type/build/format/lint/boundary/Read checks remain unchanged. Build before importing emitted launcher/configuration dependencies. No test imports during measurement-only planning.

Fresh protocol/server archives and isolated offline consumer must bind exact clean source, archive contents and installed members with only accepted publication normalization, existing eight-package production closure, entry and recursively hashed payload. Preserve actual configuration/artifact/policy/authentication ready identities. A configuration or caller source label is not child source attestation. The default emitted full-CI run is not called a fresh consumer install.

Done when the exact10-request positive establishes client-call overlap, a nondeterministic single guarded winner, exact direct loser, two retained replays and complete unchanged final state; the one emitted derivative produces precisely the named two-success rejection at request6; both invocations are cleaned up and independently bound to their genuine bytes. Any extra callback/request/child, wrong input shape, insufficient reader bound or ambiguous dependency resolution stops for root disposition.

Keep claimEligible:false and cumulative expected-revision-race omitted/wholeContract:false. This witness can inform a later evidence-mapping decision, but no new conformance vocabulary, profile admission bypass or whole-contract/public qualification is part of this plan. It does not displace the accepted receipt serving fix.
