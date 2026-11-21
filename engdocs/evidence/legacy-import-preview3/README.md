# Preview 3 legacy import quality work

Donna deliberately deferred independent draft PR86 to Preview 3. Stage 1 remains
ordinary legacy export JSONL into a fresh graph workspace. Graph-format import
follows the corresponding export later. No production or in-place data is used.
Janet owns the future qualified core integration pin; no Preview 2 admission or
staging is requested.

## Implemented repairs

Source-prefix inference is resolved once across the complete batch. A mixed
numeric/hash batch supports multi-part prefixes independently of source order.
Ambiguous all-letter multi-part IDs need the graph-only `--source-prefix` because
ordinary JSONL omits source configuration. Supplied configuration must match every
Issue ID without rewriting it; results report the adopted `issue_prefix`.

Ordinary Memory keys retain their exact original key and value, including empty
values. Valid direct paths are reserved first; canonical or digest addresses with
collision checks remap other keys deterministically. Imported Issue paths cannot
be overwritten. The complete result maps each original key to its graph URL.

Original string UTF-8 is checked before JSON cloning. Missing prefix configuration
refuses; an existing no-op prefix update remains valid. Memory read-back must match
exact content. Comment verification uses a native-ID index, and the batch has a
separate 1,000-comment cap. Native comment deduplication remains unchanged: data
that would collapse distinct supplied comments refuses the whole transaction.

## Evidence and limits

[source-checks.json](source-checks.json) pins source-file hashes and the parent head.
[pure.jsonl](pure.jsonl) records **105 passing cases/subtests, zero failures/skips**
across parser, pure storage preparation/verification and CLI guard/error/help
checks. All three packages compiled; this run opens no database. [Canonical build provenance](build.json) pins the private installed binary; it is
build evidence only. [Lint](lint.log)
passed all required profiles with zero findings and [docs](docs.log) passed.
[Comment benchmark](comment-benchmark.log) measures verification at 10/100/1,000
comments, roughly 4/7/37 microseconds on this machine. It measures the helper only,
not native database persistence, and imposes no machine-dependent time assertion.

[Earlier installed proof](../legacy-import-2026-10-10/README.md) qualifies the old
product source only. **The current repaired source is not installed-qualified.**
New storage tests cover applied/reopened field families, actual late context
cancellation/deadline, and normalized comment IDs/times and late dedup refusal;
these need both engines. No pending test is claimed to have passed.

The [prepared installed completeness test](completeness-workflow.patch) has now
been activated using Janet's genuine four-Issue/sixteen-Memory producer corpus.
[Activation and frozen evidence](../legacy-import-preview3-preparation-2026-10-18/producer-completeness/README.md)
pin exact source, binary, generator, export and receipt hashes. The patch remains
a historical artifact; do not apply it twice. Production source is unchanged.
Fresh activated-source review and exact-source installed embedded/shared-server
proof with zero required skips remain open; producer success is not import proof.

## Council and open gates

Native Codex and actual Claude reviewed the same frozen implementation prompt;
Gemini is excluded as requested. [Plan findings](plan-claude.txt) led to the prefix,
UTF-8, config, Memory verification and comment-budget repairs. Singleton ambiguous
IDs require explicit configuration rather than an invented prefix; config presence
is checked directly rather than requiring a changed-row count of one.

[Native implementation review](native-review.txt) found the old cancellation test only returned an
error and the prefix example omitted the flag on apply. Both are repaired and
bounded native verification reports no confirmed blocker. [Actual Claude implementation review](implementation-claude.txt) completed with no
confirmed code blocker. The [bounded repair review](repair-claude.txt) and
[final confirmation](final-confirmation-claude.txt) completed. Confirmed test and
wording findings are repaired; [council dispositions](council.md) record the
classification disagreement. Neither seat has a remaining confirmed code blocker. Current source differs from the
snapshot in these repaired tests/docs, error-cause wrapping the benchmark, strict infrastructure-config read, prefix disclosure in human output,
label set normalization and refusal of nonserialized SDK routing/plane state.

The final source repairs preserve underlying config error causes, use the existing
strict infrastructure-config helper, disclose adopted prefixes in human output,
normalize labels as a set on the isolated copy, and refuse unsupported internal
SDK fields before JSON cloning. The real cancellation test additionally checks
uncertainty classification against driver transaction cleanup. The independent
owned-Link post-write read remains intentional; measure its bounded cost in the
engine gate before considering storage-interface changes.

Remaining gates: genuine expanded fixture, applied/reopened preservation and
real cancellation rollback in both installed modes, exact-head hosted checks,
completed two-seat council and cleanup. Related/hierarchy/conditional/waits-for
relationships, custom workflow definitions absent from the dump, and source
history unavailable in ordinary exports remain explicit boundaries in the
[user contract](../../../docs/reference/graph-legacy-import.md). Contract questions
are on the agent bus. This report is progress, not release acceptance.

The prepared completeness patch requests exact retained state with the public
`revision` token and checks that token is nonempty. The CLI deliberately removes
the internal `version` alias; passing an empty alias to `show --version` could
otherwise read current state and give false retained-state evidence. This patch
repair is confirmed by a [bounded native review](prepared-revision-review.txt),
but is not an executed installed test or new runtime qualification. Full
active-source council remains a gate after genuine fixture activation.

## Fresh destination completeness repair

The freshness guard also refuses unrelated graph payload/Link rows, native
labels and wisp auxiliary rows, and classic config-backed KV Memories including
empty values. Ordinary configuration remains allowed. Refusal returns zero
result and preserves preexisting rows rather than adopting or repairing them.
The authored TestLegacyImportFreshDataPlanes checks apply/dry-run in both modes,
all those data planes, full row/prefix/writer-token state and a configuration
control. Its runtime cases are unrun pending explicit Janet engine handback.

The [initial review manifest](freshness-manifest.json) and
[prompt](freshness-prompt.txt) identify the initial bounded Claude review,
which completed as recorded below. [Native review](freshness-native.txt) found two fixture hazards, both repaired: wisp comment
and label foreign keys require the test-only orphan bypass, and failed FK reset
must discard the borrowed connection. The [repaired source hashes](freshness-repaired-manifest.json)
identify current source. [Pure preparation checks](freshness-pure-repaired.log) passed13cases with0failures/skips,
and [lint allprofiles](freshness-lint-repaired.log) passed; they compile the new
engine-test source without running it. Earlier council/build/runtime evidence
qualifies its recorded source only. The current repair needs bounded council
completion, a new canonical build and fresh two-mode proof.


## Native history freshness repair

The [actual initial Claude review](freshness-initial-claude.txt) completed with
exit0. Its frozen packet predates the two native fixture repairs described
above, so it does not qualify the subsequent source. Its Medium finding was
confirmed: native events, provenance, snapshots and allocation counters could
survive without live Issues and collide with imported IDs. The guard now reuses
both native deletion-table inventories, also checks journal, leases and the
Issue allocator, and refuses a consumed journal sequence (the fresh (0,0) seed
is allowed). Tests now author19 occupied planes in apply/dry-run in both modes;
ordinary config survives an actual apply control. They remain engine-unrun.
The KV prefix uses the central constant with query parameters; refusals name
the table or KV prefix/count. Missing native schema fails closed.

The [final-source manifest](freshness-final-manifest.json),
[prompt](freshness-final-prompt.txt) and [native confirmation](freshness-final-native.txt)
identify the frozen bounded successor review, now completed as recorded below.
[Pure checks](freshness-final-pure.log) pass13 cases with zero failures/skips;
[lint](freshness-final-lint.log) passes all profiles. These checks compile, but
do not execute, the new engine tests. Earlier installed proof remains historical.
The prepared installed workflow now independently asserts the one expected
blocking Link's type/endpoints and absence of owned Links on other Issues;
matching current/retained state alone could otherwise hide a missing edge.
The patch remains inactive pending the genuine corpus.


## Producer fixture publication mechanics

The development generator now verifies structured Build/Commit identity before
initialization, rechecks source/binary stability after export, refuses existing
output, and publishes UTF-8 fixture/receipt files together from a sibling staged
directory. No source/binary mismatch can leave a new dump paired with an old
receipt. It assumes a single publisher; portable directory rename provides no
atomic no-replace guarantee against concurrent creation.

[Mocked mechanics tests](fixture-generator-tests.py) use a deliberately
non-executable fake binary and mock every subprocess. [Child raw results](fixture-generator-final.log)
pass13 tests with no failures/skips. [Initial harness failure](fixture-generator-initial-harness.log)
was a macOS path alias mismatch, corrected in the harness. These tests are not
an ordinary producer corpus or engine qualification. Run from any checkout with
Python; the committed harness resolves this checkout's script. The genuine
completeness corpus still waits for Janet's explicit engine handback.


The [actual frozen history-repair review](freshness-final-claude.txt) completed
exit0 and verified all four immutable hashes at start/end. It confirms the native
history guard, all19 fixture DDLs, seed rules, and absence of false fresh-Init
refusals. Its new Medium G1 identifies a test reliability risk: thousands of
statements shared the initialization90-second budget. Per-case bounded contexts
and independent cleanup are now implemented, preserving full per-mode snapshots.
Low G2 asks isolated assertions for eight already-guarded native/preview planes;
these fixtures are now authored (27cases/26distinctplanes). Low G3 character-count coupling is repaired
with RuneCountInString; Low G4 named unreadable-plane ErrInvalidStore/cause and
fresh initialization guidance is repaired. Native production confirmation finds
no blocker. Repaired test source still needs frozen council and exact engine proof.
No preceding source or mocked generator result qualifies those runtime gates.


[Matrix manifest](matrix-manifest.json) and [native confirmation](matrix-native.txt)
identify the G1-G4 repairs. [Exact canonical pure command](matrix-pure-command.txt)
and [raw result](matrix-pure.jsonl) pass5 roots/13 cases,0failures/0skips; this
compiles27 occupied-plane cases for both modes without executing them. The
bounded actual successor packet is prepared but deliberately **not launched**
at Donna's overnight pause. Resume with that review, genuine corpus generation
only after explicit engine handback, and exact-head installed qualification.
