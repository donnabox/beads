# Frozen compiled-contract evidence/disclosure follow-up

Head `7b6f8bcbe7a99776e674b05686080d607c98c775`, direct child of frozen
`cbecfeac9b2279e7424d2499d558e976ff3f9dcf`.
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-contracts-evidence`;
branch `codex/vickie-transaction-contracts-evidence`; GitHub origin.
Clean/frozen; no outward writes. Exactly three files, tests/docs only (88+/15-).

## All seven LOW findings

1. Progress note discloses capture is strongly retained in memory by every shared
   DevelopmentContracts consumer, including RU. Installed-reference also retains its
   hosted UTF8 copies; no deduplication or RU durable-retention implementation implied.
2. External synthetic schema URI is a fixture convention, not enforced: Scope-local
   schema URIs may be installed without a TX schema hosting route. Disclosed explicitly;
   no new refusal/protocol behavior introduced.
3. Existing shared apps/bdptest compiler test now checks detached deep capture freeze
   and literal identity `development-schema5-894c98da1156e66e72d7f5b4321365b9496ec68afea377dc1861bdaaa2cadf55`.
   Observed using actual pre-capture emitted compiler at frozen fccb951, with the
   shared test's exact bundle. Probe asserts predecessor has no capture field.
   Compiler/artifact/schema source diff fccb..parent3e is empty (predecessor-source.diff).
   prior-identity.mjs/.log retain the reproducible observation. cbec's same-expression
   re-derivation is now complemented by a real predecessor literal; no identity change.
4. Note distinguishes explicit capture cap from unbounded-by-TX configuration text.
   Compiled ingestion has aggregate16MiB raw artifact limit plus separate URI bound;
   the review's4096×1MiB multi-GiB estimate overlooks aggregate charging. Serialization
   can expand, configurationIdentity still handles the inventory text, and legacy
   schema-free input lacks ingestion limits. Small fixture length is not a general cap.
5. TX test recreates rolled-back beads/tentative under a new key and asserts completed,
   directly detecting an identity reservation leak after schema failure.
6. Type HEAD against actual emitted HTTP compares200 status and Content-Length to GET
   and actual UTF8 body bytes; HEAD body empty.
7. Failed set request is resubmitted after other graph writes and must return the exact
   original retained failed receipt, including disposition/identity/details.

## Verification

Pinned Node24.16.0/pnpm11.20.0, frozen install, build before emitted/process tests.
- PASS30tests/3files: transaction-installed-contracts, transaction-runtime (including
  three process-loss/restart variants), shared apps/bdptest development-contracts.
- PASS build/typecheck/boundaries/format/lint (inherited3warnings/1info), both installed
  smokes, unchanged89-row Read evidence verification, git diff --check.
- No runtime/production source changes; full690-test parent regression was not repeated.
- No hook bypass, interrupted run or leftover test-owned child. Logs alongside report.

## Owned plan update (no implementation)

../transaction-owned-transitions-plan.md incorporates all external council findings:
absolute canonical changes-map keys vs local evaluator IDs, direct deletion before
snapshot, owned positive fixture separately closed, complete compiled Link descriptor
bundle, field-built nonempty nested delta, and already-integrated visibility closure.
Cardinality warning is stale against required064db8a99629122f130a91ee8b5189432ffe97e9:
root-directed plan preserves that deliberate cardinality-before-cap correction and its
real409vs413 witness.064 external review/root integration remains required before any
owned implementation. No new product decision or bytes/count conflation.
