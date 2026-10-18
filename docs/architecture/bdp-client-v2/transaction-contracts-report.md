# Frozen compiled-contract TX slice

Head `cbecfeac9b2279e7424d2499d558e976ff3f9dcf` (clean/frozen).
Parent `3e2b8784daefd5893f990785cfe91f1d872a2ffe`.
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-contracts`.
Branch `codex/vickie-transaction-contracts`; origin https://github.com/gastownhall/bdp.git.
No outward writes. Recovery-low5e and current set repairs remain separate for root ordinary integration.

## Delivered seam

- development-contracts.ts adds a frozen capture on the existing branded object:
  detached ingestion entries (kind, URI, base64 exact bytes/length/hash), artifact
  limits, actual graph/evaluator constant witnesses. No new compiler or installer.
  Existing identity algorithm unchanged, witnessed against the independent previous
  identity calculation. No exact-field validator is applied to the branded object.
- transaction-store.ts accepts branded contracts only. Exact separately supplied
  installed descriptor bytes must agree; compiled Types must be Scope-local and
  every owned descriptor remains refused. Existing Resource evaluation receives that
  registry. Same-owner Read installation is the same accepted descriptor inventory.
- Format8 keeps contractCapture in a separate locked metadata row, bounded at
  142,610,432 serialized UTF8 bytes (derived from existing raw/URI/count ceilings).
  No artifact payload enters configurationIdentity. Existing configuration text adds
  only contract identity and existing runtime properties/diagnostic/body limits.
  Full capture persists for installation lifetime, no artifact GC or lazy refresh.
- At reopen, SQLite can recover a hot journal, then configuration and exact capture
  are checked within existing locked startup before pending/creator cleanup/sweep.
  A claimed digest cannot conceal altered capture. Compiler constants are comparison
  witnesses, not restoration knobs; changed policy/limits fail closed. No migration,
  silent reseed or automatic source refresh. Caller administrative configuration is
  still required; tests reconstruct the compiler offline from saved exact artifacts.
- transaction-runtime.ts accepts plain bounded URI/UTF8 artifact data and compiles
  in its own module instance after requiring explicit stable Scope. No branded object
  crosses JSON/child or source/dist. Redundant types can be omitted with artifacts;
  supplied types still require exact byte agreement. Schema-free input requires types.

Six files,751 added/12 removed lines (529-line new test file); no new module, exports,
package map, HTTP operation, schema, normative spec, AuthorityReadPlane or batch edits.

## Concrete evidence

12 new installed-contract tests, plus one new compiled child-process variant:
- Detached capture arrays/entries/limits frozen despite mutation of original buffers,
  arrays and limits; prior identity unchanged; same identity/types reconstructed offline.
- Inherited transitive schema validates real writes; failed receipt diagnostics pass
  normative mutationReceipt schema. Late batch failure rolls back tentative resources,
  groups and page plans. Equal failed replay stays original after other graph writes.
- Private staged set schema failure retains failed receipt; separate two-Type test
  permits the first selected change then rejects a later effective contract, leaving
  the original resources and group state unchanged.
- Capture remains exact after create/delete, successful/failed receipt expiry,
  maintenance and reopen; same-owner installed Types persist. Configuration identity
  is under4096bytes in fixture, excludes capture payload.
- Tampered capture bytes, missing capture row, changed graph witness or diagnostic
  limit configuration refuse byte-identically before pending cleanup/expiry, leaving
  pending row intact. Changed schema bytes/URI identity and artifact limits refuse;
  compiled schema-free fallback, mismatched descriptors, nonlocal Types, owned Types,
  missing transitive schema and forged/cross-module branded authorities refuse.
- Actual emitted HTTP: Type GET/inventory, Resource GET, unknown Type failure, valid
  singleton and failed batch, independent sequence prefix success/failure/tail success,
  restart with original success and failed receipt replay.
- Actual child process compiles plain artifacts, serves writes/pages, dies SIGKILL,
  restarts on another listener with stable Scope, retains receipt/page identities,
  serves exact Type descriptor and rejects an invalid new write after restart.
- Runtime input refuses missing stable Scope, excessive per-artifact and aggregate
  bytes, excessive array cardinality and accessor-bearing entries before provisioning.
- External predecessor.mjs uses actual frozen fccb format7 emitted writer (not an
  edited marker), leaves durable pending row, then current format8 refuses with file
  bytes identical, pending row retained and no contractCapture inserted.

## Verification

Pinned Node24.16.0/pnpm11.20.0, frozen install; built before emitted/process runs.
- PASS final build (build-final.log).
- PASS final690tests/18files,8.21s (regressions-final.log): all transaction tests,
  AuthorityRead, Read sequence lifecycle, artifact ingestion, schema evaluator and
  existing apps/bdptest compiled-contract authority regressions.
- PASS final typecheck, boundaries, format, lint (only inherited3warnings/1info).
- PASS both installed executable/development smokes (public profile admission and
  deep-import boundaries preserved), unchanged89-row Read evidence verification.
- PASS actual predecessor probe again against final emitted source (predecessor-final.log).
- PASS git diff --check; clean commit, no hook bypass, interrupted runs or leftover
  test-owned children. No full workspace suite; root owns combined/full verification.

Initial targeted failures were fixture/typing corrections: unsupported wildcard
Selector, empty ownsOutgoing, camelCase command URL, sequence-level key, collection
URL without slash, mutable test literal typing and strict plain/compiled descriptor
property-order mismatch. Accepted runtime derive-types seam addressed the last.
No compiler/selector/protocol semantics were relaxed to make those tests pass.

## Council dispositions and limits

Both plan reviews fully dispositioned in ../transaction-after-core-plan.md, including
separate payload row, source/dist branding, scope-before-compile, pinned diagnostic
bounds, identity-vs-capture attribution, policy constant compatibility, explicit empty
aggregate, schema hosting gap, small inline fixture and exact-field guard.

stringifyJsonValue preserves ECMAScript property enumeration; it is NOT a canonical
identity serializer. Store exact installed-byte agreement and compiler identity remain
unchanged. Runtime deriving inventory avoids an unnecessary parent-side compiler.

Still all-visible, empty aggregate and unowned only. No schema document GET/HEAD
hosting, HTTP set activation, optional History, Types lifecycle, public configuration,
complete discovery or conformance claim. RU durable retention is a separate follow-up.
Receipt materialization byte budgeting remains open independently of matched/mutated
COUNT semantics; this slice adds no byte budget or count reinterpretation. Snapshot/
changefeed and unsettled minimum controls remain existing required-profile gaps.
No power-loss/filesystem durability claim follows from process-loss tests.

Root should ordinarily integrate recovery-low5e and reviewed set repairs, then review
this bounded installation slice before any next runtime feature.
