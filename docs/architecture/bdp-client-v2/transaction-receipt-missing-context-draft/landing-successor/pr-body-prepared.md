Authenticated receipt lookups discarded the store’s observation context when a receipt root or page was missing or belonged to another principal. Carry the genuine context through service and HTTP404 responses, including HEAD, with identity, minimum-position and expiry checks in their existing order.

Owner, service and HTTP tests cover current context after a later commit, hidden lookups before minimum waits, and contextless401 responses. The installed witness uses a genuinely issued receipt page. One existing minimum-position assertion is updated for the repaired response.

CI exposed an unrelated timing assumption in the existing cancellation witness: its2s overall budget expired before first wire. The test now keeps a bounded10s overall timer and arms an additional100ms deadline after the first actual response. It retains real deadline propagation, child-drain/no-late-publication checks and its20s test timeout; production behavior is unchanged.

Validation: build/static checks and15distinct selected receipt source cases passed. Fresh installed receipt validation passed14responses with package/dependency and cleanup audit. That evidence remains attributed to3968081, whose production and package bytes are unchanged by the cancellation-test repair. The corrected cancellation case and existing pre-aborted case pass, along with typecheck/format. Native and actual Claude reviews accept the repair, with findings dispositioned. Exact successor CI is required before merge.

The emitted omission control produced exactly eight expected tuple failures; its original wrapper postcheck failure is preserved alongside the separately accepted dependency-shim audit. Original CI38739669955 remains failed (9299passed/1failed/1skipped). No profile advertisement or release.

[Review and validation records](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/design.md)
