# Internal discovery limits — native code review

**No actionable finding in the bounded four-file change.** Reviewed frozen `fd07bf7ac177d5e53cabcd7666d9f2c77fba6e60`, tree `3b689aa3ef601307b5ee35b5de3c72f64048fcc7`, against direct parent `3f486c245bfd31f41fd42f5acc8e0f13683bad42`. This is source/evidence review only, not approval to expose discovery or claim profile qualification.

## Scope and source checks

Read the complete four-file diff (63 additions/6 removals), accepted projection-v2 plan SHA c94c5528153a5879ba83166be9e8f653d071bd03e884e6b0a12ab9ea7499c348 and root acceptance, plus actual enforcing constants and call sites, unchanged service owner-entry ordering, tests and completed author artifacts. Verified exact clean HEAD and all four per-file source hashes. No build, test, server, network request or repository mutation was run by this reviewer.

- `transaction-discovery.ts` imports frozen `TRANSACTION_RUNTIME_LIMITS` directly from batch and `TRANSACTION_SELECTOR_LIMITS` from set. Diagnostic projection uses actual diagnosticCount/diagnosticBytes, with no duplicate numerical production constants or upper configuration dependency.
- The new check compares exactly captured advertised bytes/depth/nodes to the fixed set triple. Existing Read-versus-enforcement validation remains unchanged; no second Read validator is added. The assembler still runs after the real owner entry releases, and no service/live enforcement path is rewritten.
- Local value-import traversal from batch/set reaches12 existing modules and no discovery/service/HTTP/store backedge. No new source layer, schema root, export, package inventory or public route is introduced.
- Independent literal oracle pins page2/32, Selector4096/32/256 and validation16/524288. It checks full output, hence exactly three limit groups/seven scalars and the existing17 top-level fields. Six direct mismatch cases cover both sides of each Selector bound. Missing/changed validation cases first successfully pass generic protocol parsing, then fail the owner-specific expected-value oracle; generic parser acceptance is not silently narrowed.
- Nested validation immutability is asserted alongside existing page/root immutability and TTL omission. The service callback regression changes only its exact expected validation group; receiver/getter-once, continuation and invalid-callback assertions remain intact.
- Documentation accurately describes mandatory validation disclosure, private projection, shared maximum versus collection default, and remaining serving gates. No public activation, missing-resource behavior, protocol rule, claim flag or coverage change is present.

## Completed evidence inspected, not independently executed

Initial report read was an interim status mentioning a failed cryptographic-signing attempt and pending freeze. The author finalized it during this review. The final report explicitly preserves that failed attempt and explains the Agent-Signature trailer clarification. The final source/summary/JSON outcomes are consistent; no test or packed failure was concealed by the prose update.

Actual author machine-readable outcomes inspected:

- 224 passing tests in6 affected files,0 failed/filtered.
- Typecheck/build/boundaries/format/lint/Read evidence verification exit0.
- Fresh frozen build/pack/offline installation exit0.
- Fresh packed TX fixture:1 passing test,12 completed exchanges/one child; statuses500×6,401×3,200×3, exact declared fd07 source and external installed entry, false claim flags.
- Fresh packed selected RU scenarios:2 passing,200 filtered. These are targeted regressions, not full qualification.

Both retained tarball hashes independently match summary: protocol `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`; server `2b25f7d284f47e7a9f7c7be1105cf8565b067d181cb7bdf19eaa3ee4b87a6d50`. Observed TX payload is `9000444c5d2036133eb8d8d870a760724da49f980d11ca2ec52925336211253a`; contracts `184330942591336c6ad84ef3d32e2704f67006b8ff64882a300488ab4698608a`; TX runtime configuration `eaacd2e9fa9c08cfb61a772b1b0d8ceaca230e15259d1f55f83a6b851e11365c`. Source metadata is explicitly launcher-declared against the frozen package/source binding, not target attestation. Earlier3f/b20 observations remain separate.

Packed TX PID66211 exited0 and was independently checked absent; cleanup artifact reports no matching installed process. This reviewer created no process beyond bounded read/hash commands and wrote only this external review report. Source stays clean and unchanged.

## Limits and remaining gates

No broad suite or installed cohort was rerun. Inherited PR106 export-hash/route-proof repairs and their exact ordinary integration are outside this fd07 delta; their review/CI gates are not waived. Subsequent independent council/current-head CI remain root responsibilities. Discovery activation and missingResource404 semantics remain separate future work.

## Final inspected artifact SHA256s

- `report.md`: `a6e15917942e87cf7b85dd46c5a2cc4588c8dc28ac9cd91c2f3dca21c7f1ec93`.
- `source.json`: `e46b4bfcee73f8f0406aa68a13df57446db392e0d2e801a66c779de6c11ebe92`.
- `summary.json`: `3f2e4a6bc5aaa0f0709e6ca777a9d6db671af8d7fe15652dae8a37997a88cf51`.
- `exits.json`: `f4f627d2902733dadd0e469f400e6a43df5d630f52403b6b4931509a60640ec0`.
- `cleanup.json`: `2138b32cf7556d9f142f719b098351e577ff793bdead36fc0bf7a2d4497a8991`.
