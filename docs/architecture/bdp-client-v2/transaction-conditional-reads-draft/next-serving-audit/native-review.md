# Native Event restart plan review

No blocking semantic or scope defect found. The proposed witness is a useful installed composition of existing source-tested behavior, not a demonstrated serving repair or canonical qualification credit. No implementation or execution is approved by this report.

Reviewed plan SHA256 `061a57d9db704af7ac8c8778ded998e4a64481f861eabaa79ec9a33c229313ec`; source-binding SHA256 `7f736982a2d58f64852aa551e61f225760c7df9efbca0015388d2be83727fe28`. Independently verified all 20 committed source pins at AW `b0213805e4b15bb7bac4e09da5fdddb236cf6ea5`; inspected working copies match those pins. Tree attribution is `0e589d32e707b4cb21350a90618db79e6bdc36d7`. No current-main inspection.

## Source-grounded closure

- Five actual HTTP requests are sufficient: one batch, two finite Scope Event reads, one SSE connection, and one invalid cross-source finite request. The explicit close/reopen uses IPC and does not add HTTP. Two sequential installed children/owners/listeners share one store directory. Neither request 3 nor 4 needs a second mutation.
- The two-create batch supplies one transaction and ordinals 0/1. Store eventProjection/eventPacket (`transaction-store.ts:722–800`) validates actual retained cursor membership and skips ordinals <= the cursor. `transaction-events.ts:95–141` binds Event IDs to epoch/view/source/checkpoint/ordinal. Reopen with unchanged durable store and configuration therefore supports exact E1 identity/time/data comparison to the actual earlier E1; do not synthesize IDs.
- `transaction-http.ts:520–528` passes the single Last-Event-ID ahead of query after. E0 header with E1 query discriminates precedence: the expected E1 would not be returned if the later query cursor won. Require actual 200 SSE headers, complete Event JSON equality, and id/event field equality, then exact recorded complete id-bearing frames after local close/drain. This establishes the observed suffix, not indefinite absence or a finite completed HTTP response.
- The Scope cursor sent to the existing Bead source is a true source mismatch even though its Event subject is that Bead. Source hashing rejects it before membership (`transaction-events.ts:132–141`); HTTP maps the contextual parameter error to 400 invalid-parameter (`transaction-http.ts:639–644`). This needs neither G2 incident-Link membership nor S1 Snapshot minimum policy.
- Spec `bdp.md:6212–6222,6262–6273,6350–6361` supports these exact observations. Existing source reopen/membership tests and same-process SSE reconnect tests explain why this is composition evidence rather than a newly diagnosed implementation gap.

## Concrete authoring requirements within the existing budget

The plan already rejects transport timeout as successful evidence. Implement that explicitly: `packages/server/test-support/raw-sse-recorder.ts` stores a socket failure privately, but `close()` only destroys the socket and awaits closure, and `until()` tests its predicate before checking destroyed state. Do not treat close() resolving as proof of a healthy observation. After receiving E1 and before intentionally closing, require ready/status, a non-aborted work signal, a still-live recorder, and elapsed time below its bound; post-drain check the exact frame list. Preserve observed early destruction/timeouts as failures. No shared helper change or extra request is required for this small fixture.

Bind both launches independently using their actual ready messages, compare served/runtime/payload/artifact/authentication/policy identities, and replace the dial address with the second actual listenUrl. Canonical Host remains reference.invalid. The existing binding helper checks the 12 operation targets and canonical authority mode; physical ports need not match.

Keep separate launch-attempt/known-stop records for both lifetimes. The existing launcher close is memoized and verifies a clean child exit, but a rejected pre-ready launch does not return a child handle; retain the store on unknown stop and fail honestly. The historical installed callback's unconditional rmSync must not be copied. A cleanup deadline is a failure bound, not permission to delete a potentially live store. Outer supervision and independent PID/store checks remain necessary.

Exact callback/title, literal request bytes, response/wire/report ceilings and fixed SSE timeout still need authoring-time measurement as the plan expressly requires. Preserve the historical callback bytes. Full schema and independently expected initial Event facts should precede replay equality so equal wrong responses cannot satisfy the witness. Retain only safe allowlisted report data and actual wire fingerprints; no credentials.

## Limits

No new production fix, public discovery/admission, whole-contract assertion, erasure integration, arbitrary authorization completeness or additional runtime workload is proposed. Five HTTP/two children are prospective, not executed counts. Fresh archives/external consumer/source and entry binding remain prerequisites for installed attribution.

This review used source/spec/helper reads and hashes only. No tests, imports, builds, services, review CLI, network writes, or product edits occurred. I previously authored related private evidence; this is an independent review of this author's plan, not independent authorship of every reused helper.
