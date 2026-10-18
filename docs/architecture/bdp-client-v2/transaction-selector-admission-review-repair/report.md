# Selector review repair — frozen test-only draft

The authorized repair is frozen in `/Users/dbox/repos/wt/vickie/vickie-selector-review-repair`, branch `codex/vickie-selector-review-repair`, ordinary parent `757aedc8008386745d59d5005498a9256d6ea76e`. It remains uncommitted. Original checkout and packet remain unchanged for ongoing Claude review. This is not an execution release or a compiler-pass claim.

Only three test files differ from the original frozen draft:

- `transaction-submission.test.ts`: capture the real owner in a stable local, retain the actual delegating `vi.fn`, and expose the established plain overload-boundary wrapper. No unknown cast, owner stub, added invocation or changed assertion.
- `read-selector.test.ts`: inside the existing UTF-8 case, independently pin the é+astral literal at25 UTF-8 bytes, then pad with4071/4072 ASCII spaces to4096/4097 and assert grammar analysis accepts/reports both exact byte lengths. No extra registration or owner/HTTP work.
- `transaction-set.test.ts`: inside the existing preparation case, compare raw outer JSONé with literal outer `\u00e9`, assert distinct wire strings and equal decoded Selector, then compare actual normalized identity and complete normalized value. Existing percent-literal, whitespace-identity and empty-pointer oracles remain.

`correction.patch` binds original→repair only (30 insertions/2 deletions); `draft.patch` is the complete12-path parent→repair patch including the new installed file. `source-binding.json` contains all12 original/parent/successor SHA256 values. `preservation.json` verifies the other nine paths, including all production/spec/HTTP/installed source, are byte-identical. Original compiler failure remains at `transaction-selector-admission-draft/typecheck-first.log` and its exit JSON; no rerun occurred here.

Binding SHA256: `03a861851a195d5da505f311291bee6d3cd98f63866d4b6df756e2c3ae289afd`.
Full patch SHA256: `4837959bceec2e16358c46c11a6a2676d113bbaad49ba8b3d184ebd34a366b24`.
Correction SHA256: `56305491aa6d8c9c57f3bf2ec880a508a595d88403fd73f531bd2c81c0326eca`.

`utf8-measurements.json` records independent Python string/UTF-8/JSON arithmetic and specimen hashes. The outer é spelling changes carrier bytes by4 while decoded objects remain equal; only the unexecuted source assertion checks the actual normalization implementation. This arithmetic is not substituted for a passing test.

The existing pinned standalone Biome formatter ran once on exactly the three authorized files, exit0, no fixes needed; argv, inputs/outputs and elapsed time are retained in `format.json`/log. Parse-only Babel extraction confirms exactly unchanged selected titles/counts67+6+13+4=90 and the same installed title. No product modules or test callbacks were loaded. There were no installs, typechecks, builds, tests, controls, services, packages, child processes, commits or pushes in this repair.

`commands.json` rebinds explicit workspace, output, derivative-copy and consumer paths to this successor evidence directory. It otherwise preserves exact argv/selection and all gates/bounds. `selection.json` refreshes actual source lines. `requests.json`, `controls.json` and source illustration patches are byte-identical to the finalized original packet:25 installed requests,19POST/6GET,35,865 total body bytes/20,054 maximum, one child. Source controls remain illustrative **emitted-only** proposals requiring post-build exact emitted-byte approval; do not compile the source illustration (C1 intentionally lacks the required development message and C2 leaves an import). No source-control correctness or execution is claimed.

All workload counts remain prospective:90 source positives/94HTTP/39 owners; two control invocations expected2/max52HTTP and two owners; combined41 owners/21 listeners/22 directories, expected17/max19 readonly handles; installed25/1 separately. No cases, HTTP, owners, limits, requests or production bytes changed. Native delta closure, original Claude review/dispositions and root measured execution acceptance remain pending. Stop here with source frozen.
