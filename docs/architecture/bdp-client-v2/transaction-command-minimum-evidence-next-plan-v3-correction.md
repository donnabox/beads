# Evidence plan v3 — selection field correction

Read the complete v2 native closure and inspected runner.ts:403–420. The result places scenarioSelection and selectedScenarioIds at the top level; declarations contains targetLabel/exactConfigurationDigest. Accepted the sole Low finding.

Preserved v2 unchanged: SHA256 9ea5e2cfba673a38301c981a7d9b15d4e870258bcac468a6167fe8667301feda.
V3 file: transaction-command-minimum-evidence-next-plan-v3.md, SHA256 e5c977762c38746bc5cbf33cefb9652a9c303fe073341520d6c52a370b8d6e50.

The only content delta is `Pin declarations.scenarioSelection and selectedScenarioIds` → `Pin result.scenarioSelection and result.selectedScenarioIds`. To honor the correction-only instruction, the inherited document heading still says v2; the v3 filename and this report identify this successor. No other prose, source, count, selection, route/digest, scope or budget change. Original plans remain preserved. No source edits or execution. The session guard preceded these two artifact writes.
