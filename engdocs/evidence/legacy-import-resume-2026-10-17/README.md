# Legacy import restart — October 17, 2026

Donna lifted the implementation pause. This artifact records a test-only delta over `c8910bd2b7e5c7631abfa7299ca749cd3d49d32c`; production import behavior is unchanged.

Atomicity and field-family tests now give operations and rollback observations independent bounded contexts rather than sharing initialization’s cumulative deadline. Fault hooks restore on early failure. A narrow SQL COUNT double covers unreadable table, journal sequence and config errors, preserving the named plane, invalid-store classification and original cause. CLI tests cover cancellation/deadline joined with invalid-store and uncertain-outcome precedence.

Canonical focused storage tests passed 17 records and CLI classification passed 9 records, each with zero failures/skips. The runners were explicitly configured without shared-server or Dolt test activation. The CLI runner compiled its subprocess binary but no selected test invoked it. Pinned lint v2.10.1 under Go 1.26.5 reports zero new issues for the affected packages. An initial Go 1.26.3 lint invocation failed to load the module, whose minimum is 1.26.5; its failure log is retained separately.

`manifest.json` binds the exact changed source and counts. Raw runner/JSON/lint logs accompany it. The scoped native successor review reports no findings and closes its unreadable-plane coverage finding. Actual external review is still pending: the restarted frozen predecessor review cannot qualify this later test delta. Final implementation council and exact-source installed embedded/shared-server qualification remain required.

No engine/database/producer store, core/staging/production state or release branch was changed. Runtime qualification waits for Janet’s explicit slot handback and selected current integration base. Full `make test`, cross-platform/full CI and installed tests were not run. This is a draft source-quality increment, not merge/release readiness.

Commit/push hooks that invoke ambient `bd` workspace operations are bypassed to preserve the explicit no-engine/no-core-state boundary while Janet owns the slot. Equivalent formatting, focused compile/test and pinned changed-line lint checks ran directly. The full hook/CI/runtime gates remain outstanding.
