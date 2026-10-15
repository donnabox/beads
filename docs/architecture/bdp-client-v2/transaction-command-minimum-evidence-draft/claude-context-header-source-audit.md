# Context-header source audit

Read-only audit of the actual Claude output at claude-context-header-repair/stdout.txt and the fc4fd747/4b7a84fb classifier precursor. No tests, builds, imports, services or product edits. The original five evidence paths and two production classifier files remain fixed; the subsequently authorized five test/helper corrections are reviewed separately after their freeze.

## Required finding: F4 has concrete helper fallout

Medium. The precursor collection helper at apps/bdptest/src/transactional-collection-minimum-installed.test.ts:255–298 and Snapshot helper at apps/bdptest/src/transactional-snapshot-installed.test.ts:269–311 registered bdp-authorization-view through dynamic viewNames credential handles, supplied it from viewHeaders, and excluded it from authored raw lines. With the correct new snapshotExactHttpMode credential-name exclusion (http-executor.ts:479), an exchange containing this protocol field throws during executor construction, before any HTTP. This is a real existing-caller incompatibility, not a reason to weaken the exclusion.

The narrow repair is auth-only operator/reader handles, an empty anonymous handle, auth-only or empty resolver results, and all caller-supplied headers retained in raw Object.entries order. Preserve assertions, routes, requests, deadlines and cleanup. Root authorized exactly these two helper blocks. Current source contains this repair; its frozen successor receives a separate delta review.

Search scope: repository-wide case-insensitive bdp-authorization-view occurrences and headerNames registrations across tracked/source apps, packages, scripts and docs, with source context at every dynamic view registration. Both viewNames/viewHeaders adapters were confined to these two helpers. Other registrations are ordinary credentials, empty handles, or explicit negative fixtures. No third production classifier was found: executable-manifest.ts:984/1477 and http-executor.ts:479/675 use the two private predicates; the regular expressions have no bypassing direct use. runner.ts:900 and :1551 reuse snapshotExactHttpMode.

Affected collection helper callers (all five existing installed cases author view fields):
- observes genuine installed initial collection minimums, fixed-budget failures, conditionals and retained rows
- observes genuine installed compatible missing minimum404 and preserved behind gap
- observes genuine installed current Type and incident-Link minimums within20requests
- observes genuine installed current alias minimums within18requests
- ignores command read controls while enforcing a later installed receipt minimum

Snapshot direct view callers:
- observes packed-capable canonical Snapshot streams, retained bytes and changefeed-to-point convergence
- observes fresh Snapshot minima with exact current context and retained gaps over packed bytes
- requires a real fixed-head Snapshot wait over packed bytes

The other wire caller, observes nonallocating creation decisions through the actual eight-handle capacity boundary, shares the helper but supplies no authorization-view field. Its credential branches still merit eventual full-suite coverage; it is not a direct old-viewNames failure.

## Exact minimal regression accepted by root

Run after verify-installed against the same freshly installed external consumer, before the three matrix cohorts. No source-stage product children. Environment bindings are BDP_TRANSACTIONAL_INSTALLED_ENTRY and BDP_TRANSACTIONAL_SOURCE_COMMIT; the former requires the latter. Pin actual installed entry/payload and source externally; a test title containing “packed bytes” is not provenance.

1. Collection file above: ^observes genuine installed compatible missing minimum404 and preserved behind gap$ — exactly 10 HTTP (2 POST, 6 GET, 2 HEAD), 9 unique URLs, one child/listener/store-owner lifetime and one temporary directory. Its wrapper checks count <=10 before calling the helper and asserts final count10. Two mutation receipts then seven compatible missing/deleted reads and one behind-position read exercise actual view fields. Existing helper deadline90s/test120s/request5s; response32768B, request head16384B/body65536B.
2. Snapshot file above: ^requires a real fixed-head Snapshot wait over packed bytes$ — exactly 2 sequential GET to the same Snapshot root, one child/listener/store-owner lifetime and one temporary directory; no reopen. It asserts final exchangeCount2, actual initial pos-0, and genuine future minimum 503. Existing helper deadline45s/test60s/request5s; response1MiB, head16384B/body65536B.

Total planned12 HTTP, 2 children/listeners/store-owner lifetimes/directories; no retries. Snapshot has static two calls and final count, not a new pre-call hard limiter. Selected credentials are operator only. Do not claim reader/anonymous coverage; root assigns broader existing credential-branch regressions to exact-head CI. In particular Snapshot31 has anonymous but no reader, and collection18 command case is operator-only, so selecting those two would not prove every credential branch either.

Temporary roots: bdp-collection-minimum-installed- and bdp-snapshot-installed-. Both helpers retain original observation/cleanup/deadline errors through finishObservation, close the child and remove the directory, then emit safe JSON. Collection record observation=private-installed-compatible-missing-minimum includes exchanges, completedResponses, closed, childPid, childExitCode. Snapshot observation=private-installed-snapshot includes exchanges, completedResponses and children[{pid,exitCode,signalCode}]. Both include outcome, aborted, declaredSourceCommit, workspaceDirty, entry, entrySha256, binding and claimEligible:false. Check exact expected counts/statuses, completed response wire digests, matching fresh binding, successful exit/closed fields and independent PID/group/owned-temp absence. Reports alone do not establish actual process absence.

## Remaining Claude findings, evidence-based dispositions

F1: the predicted v1 HTTP-action failure is disproven by parseActions source and the actual V2 selected pass. The useful mixed-case issue-specificity improvement is accepted: pin lowercase issue and exclude a secret-header issue, preserving all branches.

F2: the asserted fatal lint gate is false. Actual V2 lint exited0 with warnings, and the repository has existing non-null assertions. Removing four new assertions is accepted warning cleanup, not a repair of a failed gate.

F3: accept a nonvacuous ignored-compatibility-map fixture, with a correction to Claude's proposed oracle. runner.ts:1358 explicitly sets effectiveRequest undefined for exact observations; exact mode must not reread a lossy map. Populate that map with context/auth sentinels, require report request.headers {}, exact Authorization redacted, response auth redacted/context omitted, and full serialized sentinel absence. Do not expect a redacted Authorization entry in the intentionally ignored request map. Existing exact snapshot/semantic assertions remain.

F5: review-input chronology was unexecuted, but actual V2 now passed the exact eight evidence callbacks plus22 classifier regressions (30 total), including bundle construction and two app callbacks. This closes that static evidence gap; installed cohorts remain unexecuted at this audit. No public qualification follows.

F6: accept configuration category and requestWriteState value undefined. HttpTransportError constructor at143–153 uses an optional parameter property; undefined is correct, property nonexistence is not required. Only lowercase exact context-name branch discriminates the new exclusion; the other branches retain existing rejection coverage.

F7: optional selection kind rejected by root; actual selection consumer compares path/fullName and does not use kind. Counts are exact30, not qualification from assertion counts.

F8: optional predicate consolidation/comment/regex refactor deferred by root. Both private implementations are equivalent for accepted ASCII names; keeping the narrow repair avoids an unrelated public/shared seam.

## Supplied-context omissions reconciled

Claude lists six missing-context bullets. Full source confirms no remaining direct secret-regex sites; both runner configuration sites reuse the reviewed snapshot; HttpTransportError exposes the stated fields; the frozen manifest has legitimate protocol-context fields with actual V2 parse/bundle passes; Biome's actual lint outcome is success with warnings; and actual V2 execution supersedes the review input's unexecuted chronology. None permits a blanket full-profile or installed-pass claim.

The actual V2 evidence is retained in transaction-command-minimum-evidence-validation-v2/native-static-outcome.md. The failed predecessor remains retained. This audit identifies the two helper changes needed beyond that green bounded selection; it does not relabel the earlier selection as complete regression coverage.
