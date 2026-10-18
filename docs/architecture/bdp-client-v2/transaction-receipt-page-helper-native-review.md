# Native review: private receipt-page observation helper

No actionable findings in this bounded four-file change.

## Target and scope

Reviewed exact `b604320267dd0adcf1ec91854550fb402a1155e7..43e0aff5c5a6142a5f9dc2c50283e437b917d258` at `/Users/dbox/repos/wt/vickie/vickie-receipt-page-observation`. HEAD is clean and its direct parent isb604; tree `5fbf03c5d7196b75fee2c2813e2486a2b5216b2f`. Inspected the entire449-added/15-removed-line diff in transactional-observation.ts/test and runner.ts/test, the enclosing registration/dispatch/validation flow, existing lifecycle tests, accepted receipt plan612beec5 and root acceptance.

This review covers the private helper/runner extension only. Receipt manifest/coverage/app artifacts remain outside source at this head and are held pending the separate size-amendment review. Neither installed nine-row evidence nor receipt-row coverage completion is inferred from these helper tests.

## Capability and validation review

At `packages/conformance/src/transactional-observation.ts:163–249`, each `/id` or `/next` invocation independently checks a canonical mutation POST route, matching response URL, status200, bounded wire, schema, terminal disposition, available detail, canonical root ID and equality of the entire supplied JSON value with the actual wire-decoded body. Completed roots with a next URL validate it even when only the root binding is requested. A page grant requires actual non-null page disclosure. The ledger receives exactly the requested binding; there is no implicit root/page grant and neither capture order depends on its sibling having run. Original `/id` reference shape is preserved.

The page guard at lines81–102 requires the exact canonical URL, same observed parent receipt prefix and one URL-safe token of1–256 characters. It rejects normalization changes, path traversal/extra segments, encoded or dotted tokens, query, fragment, credentials, wrong parent and oversized URLs. `resolve` at lines250–272 additionally requires the binding value to equal the precise recorded target. This does not grant a wildcard page route or follow a later page chain.

The existing ledger limit and duplicate authority rejection apply to both binding kinds. Ordinary captures revoke authority on reused names; arbitrary nonmutation capture sources cannot turn receipt-looking values into authority. A failed source registration cannot add an entry. The runner stops on registration failure; this review does not infer a continue-after-failure permission that bypasses those errors.

At `runner.ts:891–978`, the exact transport configuration contains only the requested GET or HEAD and recorded target. The request uses the source credential reference, current validated session dial route and credential resolver, existing header/body/time bounds, combined operation/session abort signals and the same pending-operation registration. Original source wire digest, principal reference and immutable original session record remain distinct from the later served session. The unchanged lifecycle checks require explicit reopen, increasing generation and stable logical configuration; spontaneous session changes and aborted sessions remain rejected.

At `transactional-observation.ts:273–311`, page GET validation checks actual wire first, then the page schema and parent receipt, then the bounded terminal-page condition before creating its baseline from parsed actual response bytes. Wrong-parent, schema-invalid, nonterminal or wire-inconsistent first responses cannot initialize it. Later GETs compare the entire recorded JSON page. HEAD never assigns or replaces that baseline. Root GET still compares the complete original root JSON, including its original prefix/results/next; removal of the old blanket non-null-next refusal does not permit changed retained root content. Runner schema/identity checking selects root versus page according to the recorded source pointer.

The runner calls its observation callback only after actual status/request/body assertions pass (`runner.ts:1430–1510`), so a failed transport/status observation does not become the first accepted page baseline. Source registration likewise follows exact original request header/body/write validation and successful authored assertions.

## Tests and evidence inspected

Read all added tests and relevant existing regression cases. New actual local HTTP cases exercise both capture orders, root-with-next, HEAD-before-GET, page GET, explicit logical reopen, later-page equality, wrong parent, schema-invalid empty result list, changed page and cancellation. Cancellation checks one attempted page GET, no subsequent page action and an empty pending set; every test server closes connections and awaits server closure in `finally`. The helper cases cover token boundaries, URL rejection classes, duplicate/revoked authority, whole-source-versus-wire mismatch, independently rejected first-page baselines and shared ledger capacity. Existing root session/principal/provenance and source-route tests remain unchanged.

Inspected actual saved `transaction-receipt-page-observation/helper-final-focused.json`:334 passed,0 failed,0 pending (runner252, transactional-cumulative25, transactional-observation57). File SHA256 `f58e5c65335e5cafa3118bff3a143d9debda5dcd65fce91f0e911ac8e9188a38`. `helper-final-stages.json` records exit0 for typecheck, build, boundaries, format, lint and focused checks. The author explicitly preserves earlier typecheck failures and does not claim zero lint warnings.

Reviewer actions were local git/source/report reads and parsing/hashing existing results only. No tests, builds, services, network, credentials, source edits or outward actions were performed. Thus334 is inspected author execution evidence, not an independently rerun suite. Actual source clean state and parent/diff were independently checked.

## Remaining boundary

The helper deliberately supports only the first disclosed terminal continuation page; it rejects a further non-null page pointer. It does not prove installed receipt pagination, default result-window metadata, arbitrary page chains, all operation outcomes, final nine-row fixtures, public grammar or whole-profile qualification. The generated artifact candidate's size overrun remains a separate explicit authoring decision; no fence was changed in this source. Root owns that decision, later evidence integration and readiness.
