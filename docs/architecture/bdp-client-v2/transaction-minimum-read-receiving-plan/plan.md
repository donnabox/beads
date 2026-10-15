# Read-minimum receiving: bounded serving fix

Planning only under the sole Autonomous profile completion design. Basis: actual main5d7fa3f049fa0b4db3aa8dae20e20d8e47fb3854/tree057f879132074166f2c0f643a8b45bbd9dd47c6f. W is now at5d7 with another worker's four principal-cumulative files dirty; all this plan's inspected source bytes independently equal committed5d7. No unrelated bytes are copied, edited or treated as clean evidence. Native and actual Claude plan review precede authoring; no discovery/public-profile activation follows.

## Authority and phase distinction

Preserve root's operator-ruling.json, recorded2026-10-15T22:01:43.299160+00:00, direct reply call_4462b86558fb4852992e050db643927e.

Exact question: “If a read request includes any minimum-position control, should it require exactly one valid value for all three fields (minimum position, Scope epoch and authorization view), rejecting missing, repeated or malformed fields with400 before reading state? Recommended closed tuple rule; valid-but-unknown positions separate.”

Exact answer: **“Require all three; return 400 malformed-request (recommended)”**.

The fields are BDP-Minimum-Scope-Position, BDP-Scope-Epoch and BDP-Authorization-View. All absent means no minimum. Any presence requires exactly one nonempty token per field matching `[A-Za-z0-9_-]{1,256}` (spec6602–6625). Invalid receiving returns the existing400/request/never malformed-request Problem after normal transport/method/authentication, before owner observation. GET returns the complete Problem; HEAD returns corresponding metadata without body. No Scope tuple, ETag, Location, discovery Link or development-limit marker is fabricated. Existing non-attesting labeling remains.

This is lexical/cardinality receiving, not identity comparison. `opaque`, `unknown`, `pos-01`, `pos--1` and an overflowing pos-N satisfy the alphabet but need not identify a supported owner position: **do not map them to the new400**. Unknown identities, missing/hidden plus incompatible minima, S1 retained anchors and G2 incident-Link Events remain pending. Mutation POSTs continue ignoring these fields, including malformed/repeated values, within existing global bounds.

Root explicitly accepted the receipt distinction: authenticated malformed400 is uniform for own/foreign/nonexistent receipt and page URLs, before lookup, so it reveals no existence. For lexically complete supported **or unknown-identity** tuples, retain principal/epoch nondisclosure before semantic minimum/wait/expiry. Foreign receipts remain404 there. Add explicit normative wording that the existing receipt/page ordering follows normal HTTP lexical receiving; do not perform an extra lookup or generalize receipt semantics to Resources.

## Two production files, no new runtime seam

1. Add one source-private lexical Headers predicate/classifier in transaction-consistency.ts, sharing its existing token expression. All three null is absent; otherwise any missing/empty/nonmatching value is malformed. Do not call position() or snapshotTransactionMinimum() for this classification. Existing transactionMinimumHeaders(), owner snapshot/comparison/wait and unknown sentinel stay unchanged. No public export/schema/configuration change.
2. Add one gate in transaction-http.ts immediately after the read unauthenticated return389 and before Snapshot/read dispatch: malformed → transactionHttpProblem("malformed-request"). The command branch completes earlier; preserve this placement. Do not duplicate validation in store/service or translate arbitrary thrown faults to400.

Raw cardinality requires no new transport API. runtime145–148 iterates rawHeaders and appends each occurrence into Headers. Repeats, even identical/case-varied, combine with comma; comma is outside the token alphabet. The same lexical predicate therefore rejects repetitions and single comma-list values. Prove the real path using Node raw header arrays, not only Fetch Headers. Ordinary surrounding HTTP OWS is framing; trimmed valid tokens stay valid, whitespace-only becomes empty, internal whitespace fails. CR/LF framing remains the existing transport's responsibility.

## Route inventory and preserved precedence

| Route/seam | New receiving check / preserved boundary |
| --- | --- |
| runtime125–178 | Raw path/Host/method/auth/Origin/bounds unchanged. Schema interceptor precedes handler but declines every minimum-field request without observing; malformed requests reach common gate. No schema-minimum semantics added. |
| HTTP389–450 Snapshot | Malformed400 before query/capture/preflight; no handle/owner entry. Complete fresh minima and complete retained refusal unchanged. Query-only malformed requests retain existing400 invalid-parameter. |
| changes452–477 | Gate before query/reconnect owner work; valid tuple/unknown sentinel and JSON/SSE precedence unchanged. Receiving400 never starts a stream. |
| alias480–482 and Read558–579 | Gate before alias/read routing. Complete tuples preserve alias query404, unknown/visibility behavior and continuations. |
| Events487–521 | Malformed tuple gives same400 on existing/missing source; complete missing Event handling unchanged. |
| receipts/pages523–557 | Uniform malformed400/zero lookup; complete supported/unknown tuples retain service/store nondisclosure first. Use actual returned receipt.id and receipt.next in tests. |
| Read hooks196–233 | Downstream parser sees only lexically valid inputs; no malformed fallback to ordinary success. Unknown tokens still hit existing unsupported path. |
| handler628–634 | Existing finite HTTP semantics once:400 cannot become304/412/406. HEAD bodyless, anonymous401, wrong method405/Allow remain. |

Spec: add closed receiving paragraph under Transactional HTTP consistency after field descriptions; add adjacent receiving-phase clarification before receipt/page ordering4648. Preserve original sentences/catalog selectedText, the POST-ignore law3581–3593, and Read/Update profile boundaries. Catalog schema inventory alone is not a new requirement row; no catalog edit is proposed. Existing citation/boundary static cases must run.

## Exact eleven owned paths

- docs/specs/bdp.md
- packages/server/src/transaction-consistency.ts
- packages/server/src/transaction-http.ts
- packages/server/src/transaction-consistency.test.ts
- packages/server/src/transaction-http.test.ts
- packages/server/src/transaction-snapshot-http.test.ts
- packages/server/src/transaction-schema-http.test.ts
- packages/server/src/transaction-changefeed.test.ts
- apps/bdptest/src/transactional-snapshot-installed.test.ts
- new apps/bdptest/src/transactional-minimum-receiving-installed.test.ts
- docs/design/transaction-implementation-progress.md

No store/service/runtime/shared Read/schema/catalog/runner/helper production changes. Any additional real dependent expectation is reported before expanding scope.

## Test design and existing fallout

New CPU classifier callback covers absent-all; every missing/empty/OWS-only field; invalid internal whitespace/comma/percent/non-ASCII/257-byte tokens; repeat each field; valid maximum-length tokens; alphabet-valid unknown position remaining lexically valid while the existing semantic parser rejects it. Retain existing consistency snapshot/wait cases.

New direct handler callback uses counted **delegating facades over an actual owner/service**, not spies on frozen methods or fabricated success values, to prove zero owner entry for malformed reads across all route classes. Include conditional/unacceptable Accept, anonymous and method checks. Count initial fixture setup separately; always close the actual owner.

New raw HTTP callback proves each duplicate field through runtime rawHeaders, full400 Problem, absent context/marker, HEAD length/body behavior and no SSE on400. Preserve safe actual observations before assertions. A separate one-request source oracle requests beads/ with only Minimum-Scope-Position. An isolated emitted derivative removes only the common new gate: actual original500/minimum-position must fail named `minimum-receiving-malformed-request` after retaining the response. One request/one owner/listener, no child and no hidden baseline GET/write. Build/copy the complete runner/product import graph; do not let aliases select positive source. No installed mutation is needed.

Necessary existing expectation/input repairs (keep full callback coverage, not only edited rows):

| Existing callback | Narrow change |
| --- | --- |
| HTTP360 “fails minimum controls closed before ordinary Read and expired cursor conditional decisions” | Three partial tuple responses400; absent-tuple expired continuation410 remains. |
| HTTP2027 “orders page nondisclosure, minimum waits, expiry and representation using actual owner state” | Partial own/foreign both400/no context. All complete-context assertions retained. |
| HTTP2150 “refuses unsupported root controls after nondisclosure but before read-triggered retention” | Preserve this semantic-unknown test by adding actual saved epoch/view to its unknown position. Keep own500/foreign404 and retained-row inventory; no added requests. |
| HTTP2185 “keeps failed parent rows unchanged for nonexistent page reads with any controls” | Make unknown-identity row complete using actual returned context; retain404/no-retention semantics. New receiving tests independently cover malformed unknown pages400. |
| HTTP3033 “serves conditioned live aliases and preserves bodyless redirects, refusal context and receiving boundaries” | Only partial row becomes400/full Problem; full unknown/visibility gaps unchanged. |
| Snapshot HTTP385 “closed query syntax, unsupported controls and foreign creation context fail before capture or conditional success” | Three partial rows400; full fresh410/full retained500 remain. |
| Schema HTTP232 “preserves minimum refusals before schema conditionals and never observes excluded requests” | Separate malformed400 from complete supported route-gap500; adapter remains undefined/zero observations for all supplied minima. |
| Changefeed “decides actual HTTP minimum refusal before conditional shortcuts with final context” at723–727 | Partial tuple400/no context; valid foreign/satisfied304 retained. |
| Installed Snapshot590 “observes nonallocating creation decisions through the actual eight-handle capacity boundary” | One partial response400; all allocation/reopen/state checks retained.39 HTTP/three sequential children. |
| Installed Snapshot796 “observes fresh Snapshot minima with exact current context and retained gaps over packed bytes” | Rows20/21 become400; row19 full unknown and row22 full retained stay500.31 HTTP/two sequential children. |

Review of other hits found complete tuples/patch merges, command POST cases, or direct internal owner tests. HTTP1912 `unissued`, Event minima, visibility collisions and owner unsupported forms stay unchanged. Draft must extract enclosing describe prefixes/expanded titles exactly. Do not call full-tuple `unissued` malformed merely because an old local variable does.

## Proposed installed21/1 witness

One fixed non-root canonical fixture; writer and another authenticated principal; existing receipt page maxima both2. One three-create batch yields real receipt.next. Read grants do not invent separate views. Cold discovery is not enabled.

1. Batch three creates with malformed copied minimum fields →200 (POST ignore), actual receipt/page/context.
2. Existing Bead GET absent-all →200/full Resource and tuple.
3. Existing Bead GET only minimum →400.
4. HEAD existing Bead missing epoch with If-None-Match:* →400/bodyless.
5–7. Duplicate position, epoch, view respectively, other fields complete →400 each.
8. Complete tuple containing comma token →400.
9–12. Actual receipt then actual next page, each own/other principal, malformed tuple →four400/no-context.
13–14. Other principal actual receipt/page with complete alphabet-valid unknown position →404, existing context rules.
15–16. Owner same receipt/page/complete unknown →existing500/minimum-position, not400.
17. Missing Bead with compatible satisfied tuple →404/current context.
18. Existing Bead with satisfied tuple →200/exact unchanged Resource.
19. Existing Bead with recognized future same-view position →bounded503/catch-up-timeout/current context.
20. Anonymous malformed read →401/no context.
21. Unsupported POST to Resource read target with malformed fields →405/Allow GET, HEAD/no Problem.

No added effect after1, no retries/pagination/reopen in this new case. The partial/malformed outcomes have no owner-observation claim. Literal body/header measurements and complete code selectors freeze before execution.

## Validation and cleanup

Source build/type/format/lint/boundary/Read seal, exact affected source callbacks plus existing citation/profile-boundary static cases; no broad matrix by default. After source/control acceptance and clean candidate, fresh protocol/server archives and external offline consumer: exact source/archive/installed members, eight-package closure and payload/entry/configuration binding. New21/1 plus both changed Snapshot cases39/3 and31/2 gives **91 expected HTTP/six installed child starts**, sequential.39 = first fixture10 + second29 including two reads after reopen;31 is already asserted by its source. Source/control work remains separate and must be fully measured in the draft; do not hide it in91 or call a static expected count a live network quota.

New witness retains five-second request,45-second work,10-second cleanup,60-second test and120-second outer command bounds. Stream under32768-byte response cap, cancel/drain overflow, no redirects, exact dispatch counter. Existing Snapshot helper limits stay unchanged. Preserve primary/transport/assertion/report/cleanup errors; safe allowlisted metadata only, no credentials/raw config. Remove owned store only after proven child stop; attempted launch without returned child is unknown and preserves directory for supervisor adjudication. Independently inspect processes/temp paths; stop first unexpected failure, no replacement start or automatic retry.

Before any execution, root must read the measured draft's full diff/source hashes, exact expanded title selection, **all** callback request/owner/listener counts, bodies and control import proof. This supporting plan authorizes no run. Current source inspection reveals no additional semantic blocker inside the lexical ruling; unknown-token and nondisclosure semantic decisions remain separate.
