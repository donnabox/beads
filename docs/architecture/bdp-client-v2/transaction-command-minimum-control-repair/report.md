# Receipt minimum negative-control typing repair

Proposal only; not applied, built, or copied into a runnable checkout. Frozen W and all seven final-draft hashes remain unchanged.

The earlier negative fixture bypassed receiptAtMinimum with service.receipt but left a check for result.kind==unsupported, which is absent from service.receipt's return union. Actual derivative build failed TS2367; its selected test never ran. Original copy/log remain preserved at `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-command-minimum-later-preparation-v3/run/bypass-receipt-build.log` (SHA256 `139d64149d9f8bafe9653a6f590d1b8e199708d1f91ecdb9e7d12501009ffc2e`). This is a negative-fixture typing error, not a production failure or accepted sensitivity result.

The replacement patch is one contiguous hunk: direct receipt dispatch, unchanged refused handling, removal of only the now-unreachable unsupported check and its two-line comment. receiptResponse, page path, minimum parsing and all other code remain unchanged. No casts or product semantics changes.

Before SHA256 `1339f32e0818aee5a44fc01a2034366a73e72c40027d6cde7041f26525aba366`; proposed negative fixture `9da034938dc1a48dd48b961a0ec9fa518dc954d1bddcb2ab38f321a3da13a131`; patch `dca316bf79495061823fb12ed9ffd0c369e51ac2ca32858451c7301b8a691fff`. Exact snapshots and full patch are retained here. Old patch hash `ef665809f6821f8af08b61d00f821dcd5d1f9de1f50cb5cd95213fb120d40b0a` remains preserved.

Intended case E still expects2 HTTP/max6 under unchanged time and cleanup bounds. The mutant must return200 on the first future-minimum receipt GET and fail the named503 oracle with full actual receipt/context in its diagnostic. Generic build/test exit1 is not sufficient. Root/native review and explicit re-release precede any execution. No source/commit/push/gate change.

Preparation note: initial report assembly guessed execution-preparation-v3 instead of the actual later-preparation-v3 log directory and exited1 on a missing log after writing the proposed snapshots/patch. Read-only filename discovery corrected the report binding; patch/source bytes did not change and no product execution occurred.
