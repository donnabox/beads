# Receipt-control test timer isolation

Frozen clean `2fa2d0291d0a4dec95bc8b4e864c8f30c00b74b8`, direct child of fccb951,
branch codex/vickie-transaction-timer-isolation, checkout
/Users/dbox/repos/wt/vickie/vickie-transaction-timer-isolation.

Two files, nine added lines / one removed line. Existing maintenanceIntervalMs is
forwarded through the source-private runtime TEST seam only. Absence preserves the
ordinary one-second default; store validation/lifecycle are unchanged. No installed
configuration or public option/profile is added. The one root-controls test uses a
positive 60-second interval and waits 1.1 real seconds after advancing the authority
clock, crossing the default sweep interval before its HTTP assertions. No broad fake
timers or socket behavior changes.

PASS targeted test (one selected, 37 filtered out), then an emitted-only counterfactual
removed the forwarding expression. It FAILED as expected with 404 instead of500 after
maintenance legitimately forgot the failed receipt. Restored emitted code byte-exactly;
PASS all42 HTTP/runtime/recovery tests across three files, including both actual
hot-journal crash probes. Build preceded emitted tests. Pinned Node24.16.0/pnpm11.20.0.

PASS typecheck, boundaries, formatting and lint (existing three warnings/one info).
No new emitted module/export/dependency; package smokes and unchanged evidence were
already verified on parent fccb and not repeated for this private test-seam-only fix.
No interrupted runs/process cleanup, skipped hooks or outward writes. Source270a's
runtime repair and this test-isolation successor are distinct; the latter removes a
slow-run race in its test, not a protocol/runtime behavior defect.
