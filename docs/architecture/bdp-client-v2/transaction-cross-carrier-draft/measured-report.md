# Cross-carrier measurement-only draft freeze

Status: ready for independent preflight; execution and guard changes held. This is an uncommitted ordinary child draft of eff171839fe4f37638284a4c3346428979d26b5d (tree 5c14bcc20ec7a72118bc07d2ed5896703211a660), branch codex/vickie-cross-carrier-draft. Root supplied this verified main after PR120/PR121. All four owned parent files independently match frozen c555. No other tracked source path differs. No commit, push, PR or coordination mutation was made.

Accepted plan: transaction-after-timeout-next-plan-v3.md SHA c021d366eb99b7e0f684a815a29e2ed7068dc12f07e33ea936bf0951fc80e072 and its root acceptance. Scope remains one partial cross-carrier row, twelve actions, two actual-wire controls and the same four files. No runtime, fixture, catalog, schema, public API, configuration mechanism or admission change. No new semantic decision is inferred.

## Measured joint limits

| Form | Actual | Accepted hard stop |
| --- | ---: | ---: |
| Manifest canonical disk bytes | 684206 | 720094 |
| Manifest normal stringify bytes, trailing newline | 695598 | 730676 |
| Manifest nodes | 18107 | 18973 |
| Manifest actions | 199 | exactly 199 |
| Manifest assertions | 2102 | 2172 |
| Coverage disk / normal stringify bytes | 99245 / 99245 | 106043 each |
| Maximum derivative normal stringify bytes | 701716 | 736794 |
| Maximum derivative nodes | 18263 | 19129 |

Canonical growth from parent: disk 64112 bytes, normal stringify 64922 bytes, 1734 nodes, 12 actions, 230 assertions. Coverage growth 11202 bytes. All constraints hold simultaneously. Compact canonical execution serialization is 339575 bytes; it does not replace the normal-form fences. New batch control adds 1341 bytes/37 nodes; new sequence control adds 1354 bytes/36 nodes. Both remain below the old deletion maximum overhead 6118 bytes/156 nodes.

The existing smaller disk/delta/coverage/node guards in transactional-cumulative.test.ts and derivative 650386-byte/17275-node guards in transactional-matrix.test.ts are unchanged. The new derivative guard intentionally copies those existing values. They cannot pass this draft and must be dispositioned by root before execution; this report does not authorize an increase.

## Preservation and authored observation

All 24 old raw objects (12 manifest, 12 coverage) compare byte-for-byte. Removing only the new row reconstructs the complete parent manifest SHA 56318c56ee99bb981336e3f17560e7b079e88f2f85ef463874c1d556a066b2b2 and coverage SHA dafedd035ca20f684559c739ebf2e6155f8446ce8062c3fa60ff86967b259bbd. Existing raw pins remain, with both parent deletion objects newly pinned. Fixture remains unchanged. The canonical row hash is 9edceed3f23528cff25103143dc27aeac69be409420f0ed54c5a4f992c89ba73; the one ordered coverage binding contains all 230 action-qualified assertion IDs.

The row creates one Bead through singleton, reads its actual retained receipt, replays through batch and sequence, updates through batch using propertiesChange, reads that receipt, replays both keys in a two-member sequence, and replays through singleton/batch. Current collection state and two ordered Events bind original and current revisions/transaction links. Sequence presentation uses current index/name while retained records preserve original semantics. Metadata is omitted from input and stored empty metadata is checked. Full-body/Resource comparisons use actual receipt client results; cross-response epoch/view and revision inequality use existing unredacted afterExchange responses. Report redacted values are not comparison sources. No new hook or transport is introduced.

Per-action assertion counts: x-create 18, x-created-facts 20, x-create-batch 21, x-create-sequence 16, x-update 18, x-updated-facts 20, x-two-sequence 19, x-update-singleton 21, x-create-singleton-retry 21, x-create-batch-retry 21, x-current 13, x-events 22. The normal row is preserved separately in authored-row.json. Static weakenings are folded into an existing case: removed retained body, wrong current sequence name and substituted updated record.

## Controls and complete derivative inventory

The batch control changes the actual create title under the retained key. It expects direct 409 idempotency-conflict/never with real serving context and full positive Problem checks; the row stops at x-create-batch after three HTTP exchanges. Exact expected failing IDs: http-status, schema, media, serving-context, disposition, detail, key, no-next, required-position, effect-position, literal-result, retained-root, retained-body. Added positive IDs: control-status, control-schema, control-media, control-type, control-code, control-problem-status, control-retry.

The sequence control changes the actual member key to carrier-create-new while retaining the occupied explicit identity. It expects outer 200 with exactly one flat identity-taken/409/never member at index 0/name first, no resource/outcome; the row stops at x-create-sequence after four exchanges. Exact failing IDs: presentation, created-record. Added positive IDs: control-member, control-no-resource, control-no-outcome. Neither control removes any existing assertion. measurements.json contains exact retained/removed/positive/failing/passing lists and both serialization hashes.

Independent measure.py reproduces all 37 manifest derivative transformations as JSON data; it does not load or run product/test modules. Fifteen other observe invocations retain canonical bytes (metadata5, abort/preabort/duplicate4, deniedall1, omittedroutes2, deniedset2, positive1), totaling52 observe invocations. All individual derivative files and full hashes are retained. These data checks are not executed runner or schema-validator results; schema reference names were checked against existing definitions only.

| Derivative | Normal bytes | Nodes | Byte / node delta |
| --- | ---: | ---: | ---: |
| receipt-binding | 695591 | 18107 | -7 / 0 |
| receipt-method | 695599 | 18107 | 1 / 0 |
| receipt-credential | 695637 | 18108 | 39 / 1 |
| receipt-body-assertion | 694915 | 18086 | -683 / -21 |
| ordinary-capture | 695799 | 18112 | 201 / 5 |
| forbidden-receipt-use | 696308 | 18127 | 710 / 20 |
| set-matched | 695598 | 18107 | 0 / 0 |
| set-order | 695598 | 18107 | 0 / 0 |
| set-index | 695598 | 18107 | 0 / 0 |
| set-poststate | 695605 | 18107 | 7 / 0 |
| cross-wired-receipt | 694876 | 18086 | -722 / -21 |
| effects-changed-poststate | 695605 | 18107 | 7 / 0 |
| effects-rollback-metadata | 695611 | 18107 | 13 / 0 |
| effects-receipt-order | 695598 | 18107 | 0 / 0 |
| effects-event-order | 695598 | 18107 | 0 / 0 |
| effects-zero-results | 695598 | 18107 | 0 / 0 |
| effects-zero-events | 695744 | 18111 | 146 / 4 |
| effects-no-effect | 695597 | 18107 | -1 / 0 |
| pagination-root-next | 695587 | 18107 | -11 / 0 |
| pagination-inline-boundary | 695598 | 18107 | 0 / 0 |
| pagination-page-member | 695598 | 18107 | 0 / 0 |
| pagination-page-parent | 695630 | 18107 | 32 / 0 |
| pagination-historical-title | 695596 | 18107 | -2 / 0 |
| pagination-terminal-next | 695652 | 18107 | 54 / 0 |
| conflict-code | 695595 | 18107 | -3 / 0 |
| conflict-reopened-status | 695598 | 18107 | 0 / 0 |
| conflict-current-record | 695598 | 18107 | 0 / 0 |
| conflict-retained-body | 695585 | 18107 | -13 / 0 |
| conflict-event | 695598 | 18107 | 0 / 0 |
| noop-bead-property | 695606 | 18107 | 8 / 0 |
| noop-link-metadata | 695619 | 18107 | 21 / 0 |
| noop-group-inventory-and-head | 695661 | 18109 | 63 / 2 |
| deletion-incoming-order | 701716 | 18263 | 6118 / 156 |
| deletion-outgoing-order | 701716 | 18263 | 6118 / 156 |
| deletion-committed-prefix-state | 698702 | 18184 | 3104 / 77 |
| cross-batch-semantic-conflict | 696939 | 18144 | 1341 / 37 |
| cross-sequence-fresh-key | 696952 | 18143 | 1354 / 36 |

## Configuration, counts and execution limits

The independent complete ordered 15-route preimage is route-preimage.json, SHA/digest a9a17908b6fa87674d5f573b584f4a1ce1d41d7909fadf48b65e9b05397ddfd4. The prior twelve routes retain their order; create-bead, sequence and update-bead are appended. Both app route inventory and independent preimage are changed together. This binds configuration only, not executable source attestation.

Positive modeled counts: 13 passing partial rows, 20 scenario/reopen sessions plus bootstrap =21 children, 192 HTTP exchanges. New controls:183/184 HTTP and21 children each. Old no-op ceilings181/181/186 and old deletion176/181/177 each use21 children, preserve all affected-row transformations/prefixes and now require every other authored row to pass. Ordinary assertion failure ends its row, not the entire matrix.

Expanded app cases57→59; observe invocations50→52. Existing static cases contain new preflight/weakening checks, rather than registering extra cases. The accepted source-derived full-file observe model remains7342 HTTP/at-most908 children, with unchanged non-observe probes separately accountable. This draft does not claim these as observed totals. The eventual nine selected author cohorts model1641 HTTP/189 children (absolute1728/189), comprising positive+two new controls559/63 and six old affected controls1082/126. No such cohort has run.

All existing time limits remain: positive63000ms, cohort90000ms, request5000ms, outer120000ms, smaller per-case limits and preparation/cleanup bounds. Accepted historical estimate of added recurring wall time is about57s normal/89s slow (rough ranges50–70s/80–110s), not candidate timing. Plan historical 900s CI discussion predates root's separate PR120 headroom integration; this draft makes no CI edit or candidate completion claim. Root must bind actual candidate full-job results and margin independently. Counts, successful positive timing alone, or source configuration hashes cannot establish full profile qualification.

## Execution and limitations

Only source editing, existing formatter source preparation, read-only inspection and independent JSON serialization/counting ran. No tests, builds, type checks, lint, installs, packs, services, CI or application execution ran. Therefore no child or HTTP test traffic exists for this draft, and there is no runtime pass claim. All mutations used guarded-action.py session verification. One Python authoring-command syntax error occurred before its intended writes and was corrected; it was not a product execution. Read-only searches included a missing guessed filename; no product result is inferred.

Partial row and all existing rows remain wholeContract:false; claimEligible remains false. Required250/catalog255 and unplanned237 are explicit coverage accounting, not completion percentages. Singleton negative, sequence origin, all-carrier cross-product, pending/failed/expired/withheld retention, allocation, alias/local binding, epoch/view/principal changes, restart and erasure remain outside this row. Pending operator questions remain unanswered.

Freeze basis: binding.json lists every changed source hash, patch hash and supporting measurement hashes. Source is held for root and independent preflight. No execution or guard amendment is authorized by this freeze.
