# Native review: S1a Snapshot owner core

Frozen source `b9afb271ef089b0ed204532fe7b1f153813450b3`, direct parent `31aab4db0f61e3d8b1139a7b78ff1ff10f890836`. Reviewed the three-file diff and author report. Independent clean detached clone: `vickie-snapshot-native-review`. Original author source remained unchanged. No outward writes or HTTP/minimum readiness claim.

## Verdict: one required paging repair

### F1 — Medium: current-token framing can commit an unservable next page

Location: `packages/server/src/transaction-snapshot.ts:631–647`; actual next projection at `:502–507`.

The planner reserves an absolute URL containing **this** page's token (`pageToken(..., ordinal)`). A nonterminal page actually returns the token for `ordinal + 1`. At decimal-width transitions such as9→10, that URL is one byte longer. The record inclusion check can consequently accept a page exactly at4MiB using the shorter URL. Only the first page of each stream is materialized by capture's final manifest validation, so an offending later page commits with its promised handle and then throws `Snapshot integrity: serialized page bound` when fetched before expiry.

Independent executable reproduction: `transaction-snapshot-core-native/page-boundary-repro.mjs` and `.log`. It uses the actual emitted owner at the frozen head and a disposable offline complete-record capacity fixture, following the author's capacity-testing pattern. It is **not mutation-admission evidence**.305 Beads:288 small records fill nine count-limited pages, then15 R-sized records plus one261,956-byte record exactly fill the planned4MiB page with173-byte current-token framing and15 commas; one final small record requires a continuation. Capture succeeds, pages1–8 succeed, page9 fails. The aggregate remains well below32MiB and each record is at most R. No retained-data corruption is injected after capture. The core itself accepted/committed this fixture then failed its promised page.

Repair: size framing with the actual successor token/ordinal or a proven conservative maximum successor URL; preserve exact full-page byte validation and progress. Add a digit-transition, near-byte-cap regression that follows every page, checks each full serialized page bound, and refetches after reopen. Do not increase the page bound or weaken its final validation. Root has assigned the author a separate successor; this review remains frozen on b9afb.

## Other reviewed boundaries

- Capture runs in existing guarded owner entry and SQLite atomic boundary. Expiry reclamation commits separately before allocation; capture failure does not undo legitimate prior detail reclamation. Before/after COMMIT error handling uses existing sticky fencing; postcommit delivery fault does not compensate a committed capture. No new key/receipt/group/Resource revision is produced by capture.
- Configuration includes snapshot-9 format, actual Snapshot limit object and HMAC policy before pending retraction/sweep. The existing writable locked-open sequence allows necessary SQLite journal recovery before configuration is readable; it does not add a readonly preflight or automatic migration. Independent prior-format proof is author-supplied, while the permanent focused test independently passed here.
- Count/encoded-byte checks precede the bounded identity list, individual record lengths precede body loading, and complete records use the existing stored-record decoder and canonical comparator. Owned/first-class agreement and local endpoints are checked before issue. These are finite caps, not universal heap or population qualification. The new issuance ledger scan is bounded at65,536 and remains potentially substantial synchronous work.
- Disposition commitment covers identity plus retained epoch/view/anchor/expiry/counts/page counts/seed/plan. Page HMAC binds policy, Snapshot ID, kind and ordinal. Canonical ordinal spelling and43-character tag are checked before timingSafeEqual; expected and candidate lengths agree. Count bounds plus retained seed recognize valid issued tokens after detail purge without scanning all possible ordinals. Unknown/wrong-kind/borrowed tokens remain404; recognized expired tokens become410 before detail-plan access. Tokens remain locators, not authorization credentials.
- Retained page data has range-plan and body commitments, bounded materialization and immutable returned objects. Missing/changed/oversized records and semantic disposition corruptions fail rather than reconstructing from current graph. The tests reseal selected corruptions so they exercise semantic validation beyond checksum rejection.
- Same-entry success context is separate from immutable manifest anchor. No service/auth/HTTP route exists in this commit; nonallocation conditional preflight, expected-failure context pairing and Q1 remain later integration requirements rather than falsely claimed S1a features.

## Evidence precision note

The abrupt-child-exit test proves tentative capture is absent on reopen and graph state survives. It does not inspect the pre-reopen journal/header or prove that dirty database pages reached disk before exit; a small transaction can leave equivalent results without exercising actual hot-journal recovery. Keep that observation described as abrupt-exit/reopen unless a separate recorded journal witness establishes the stronger claim. This is a Low evidence limitation, not a demonstrated recovery correctness defect. Existing explicit uncertain-COMMIT/rollback tests are separate meaningful controls.

## Independent verification

Pinned Node24.16.0/pnpm11.20.0, frozen-lock install and clean build passed. Focused `transaction-snapshot.test.ts` with maxWorkers2: **27 passed,1 file,8.10s** (7.72s test time), no failures/skips. Full logs and JSON: `transaction-snapshot-core-native/{install,build,tests}.log`, `tests.json`. These tests did not cover F1; the external actual-code reproduction does.

No full repository suite, packed Snapshot cohort, HTTP serving, population-wide performance qualification or Q1 semantics were tested. Existing test timeouts were unchanged. The issuance-limit test took3.948s, recorded in JSON; the author's12.38s-under-load/15s budget caveat remains valid despite this faster isolated run. Tests and reproduction removed their disposable stores; no review-owned child remained, and isolated git status was clean. Author source stayed frozen.

## Follow-up closure

The required page-framing defect and abrupt-exit wording follow-up were independently closed at `a0a068c689f5588701c2b7a350bb677dc4f114ae`; see `transaction-snapshot-page-repair-native-review.md`. The findings and evidence above remain pinned to original b9afb.
