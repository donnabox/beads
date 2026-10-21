# Independent fresh-depth source audit

The next serving slice is justified, but the original audit misses one load-bearing deep-value comparison: Snapshot graph validation compares complete inline and first-class owned Link records with Node's `isDeepStrictEqual`. The plan must cover this path before treating guard removal as ready. I found no reason to introduce a protocol JSON nesting limit or to tie this correction to the pending Snapshot minimum ruling.

This is a read-only source review, not a reproduced runtime failure or a stack-depth threshold measurement. No imports, tests, builds, services, source edits or reviewer invocations were performed. I have authored earlier Transactional work; this review is independent of the current fresh-depth audit and proposed implementation.

## Basis

Workspace: `/Users/dbox/repos/wt/vickie/vickie-sequence-depth-retained`, HEAD `df80f8cb86b7763bc7bb005492e43d8840527045`, with the frozen seven-path retained-sequence draft. That draft is an unmerged future assumption, not current main.

Reviewed audit: `fresh-depth-serving-audit.md`, SHA-256 `01dfcca80dc96f190481737e361ff84244d44c1701eff9560c67ad2da549af5f`. All **20 hashes explicitly listed in its inspected-file section** match current bytes. Its prose mentions 29 contexts; this report verifies the 20 listed pins, without inferring nine additional pins. Additional inspected pins:

- `packages/server/src/transaction-runtime-configuration.ts`: `57cd5d38c7199011d05589bc2b8ac1a0e2a4a9dc7585d019f9de5801a0a4d808`.
- `packages/server/src/read-http.ts`: `31934105eff6fff8ded8ce2ca9e3c19028991ca7c3d5f51b848004177e8c9268`.

Line references below are relative to that workspace.

## Required addition: deep owned-Link Snapshot equality

`packages/server/src/transaction-snapshot.ts:323–360` independently reads persisted Resource records. `checkGraph` at361–425 validates each owning Bead's inline Link against the corresponding first-class Link. The comparison at407 passes the **entire Link**, including arbitrary nested properties/metadata, to `isDeepStrictEqual`. Capture calls this validation at699; both public-private owner entry points, preflight806 and capture809, use that path.

This is different from the shallow string-array inventory comparison at399. Iterative freeze and serialization do not establish depth safety of the complete-record comparison. The source audit's output-path conclusion therefore needs this qualification. Source inspection establishes the unbounded nested input to the general comparator; this review has not executed Node internals or established which depth fails on the pinned runtime.

A concrete bounded witness can create endpoint Beads and a Link owned by its source Type, put a deeply nested array in the Link's properties or metadata, then capture an ordinary Snapshot without any minimum tuple. The owning Bead's inline copy and first-class Link are separate parsed values. Simple depth-12,000 arrays occupy about24,001 bytes before their small envelope, so this candidate can fit the existing64 KiB request/properties and256 KiB representation limits. Exact complete bodies must still be measured before execution. This uses an existing source-test depth, not a proposed normative maximum.

Existing `resource-evaluator.test.ts:1570–1618` already checks deep first-class/inline Link equality through create/read/update/delete. It does **not** call Snapshot capture. The earlier depth tests likewise establish evaluator persistence/no-op behavior, not this graph validator.

The smallest implementation plan should preserve graph-integrity checking while using depth-safe structural equality for this pair, and add actual deep owned-Link Snapshot evidence. Preserve object-key-order independence, array order and admitted numeric semantics. Do not drop the comparison or replace it with order-sensitive serialized-text equality. Include a mismatched inline/first-class record rejection check through an established integrity fixture so the repair cannot merely stop detecting divergence. Reusing/extracting the existing iterative equality at `resource-evaluator.ts:383–408` is a candidate, subject to its exact value semantics; this audit does not authorize that edit.

This path is independent of S1: ordinary Snapshot construction already performs it. No Snapshot minimum-policy change or new capability is needed to investigate or fix it.

## Other equality/recursion sites checked

- `transaction-runtime-configuration.ts:359` compares `actual.scopeSnapshot` and `settings.scopeSnapshot`. Defaults89–135 and store640–645 construct only fixed `limits` and `tokenPolicy` fields. Installed Type descriptors/schema bodies are not nested in this pair. The Type closure check369–372 first uses the iterative serializer on individual descriptors, then native JSON.stringify on an array of strings. These are not further arbitrary Resource-value recursion sites.
- Configuration `detached`66–80 has its own explicit administrative depth64/array4096 bounds; `freeze`60–65 traverses captured configuration. Schema artifacts are strings and this is a separate bounded configuration surface. Do not remove these administrative guards as part of supporting deep Resource values.
- Store2079 compares persisted dependency templates. `transaction-dependencies.ts:33–117,118–171` and their internal producers encode deep semantic identity as a **string**; dependencies/aliases are bounded scalar records and reference identities are strings. The ordinary generated comparison does not traverse the deep Resource document. This corroborates the original audit's distinction.
- `resource-evaluator.ts:1124–1166` follows Bead-to-owned-Link materialization recipes; Link recipes do not recursively own further Links. Its recursion is not proportional to JSON value depth.

## Guard-removal plan requirements

Preserve the authoritative admission order at store1133–1193 and sequence1957 onward: retained/dependent comparison, expiry/turn-time reclassification, actual fresh admission controls and ownership insertion. Strict singleton/set preparation and received batch/sequence forms must all be inventoried. Removing depth handling must not swallow other development-capacity errors, move Accept evaluation, consume keys on receiving refusal, or change dependency retraction/committed-prefix behavior.

The shared evaluator already uses iterative copy/equality/freeze/patch operations. Root replacement and shallow array append can carry deep values with a short patch path; Selector depth and patch-path depth do not bound that payload. Spec1043–1070 and2200–2217 do not supply a generic JSON-depth limit. Keep genuine request/properties/representation limits and distinguish postimage failed outcomes from pre-admission refusals. A valid late failure after an earlier staged mutation should continue to demonstrate effects rollback and retained failed receipt, with no leaked Resources or Events.

Retain the proposed fresh create/update/no-op/set/dependent-sequence, replay/conflict, receipt/page/current/Event and reopen evidence, but add the owned-Link Snapshot path above. Measure complete inputs and outputs under existing limits. Finite installed-schema work/frame/diagnostic budgets remain a distinct limitation: `development-contracts.ts:98–118` can turn incomplete evaluation into a private error independently of nesting, even for shallow instances. Neither removing32 nor adding another arbitrary depth constant proves all accepted schemas executable.

The readiness consequence is narrow: the audit supports proceeding to a bounded implementation plan, but **does not yet support guard deletion alone or an all-output depth-safe claim**. No additional human semantic ruling was established by this review. S1/G2 and full public qualification remain separate.
