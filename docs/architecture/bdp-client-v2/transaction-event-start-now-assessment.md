# SSE start=now plus Last-Event-ID: bounded assessment

## Result and scope correction

The explicit `start-now-last-event-id` development gap belongs to **Scope changefeed SSE (`changes/`)**, not Scope `events/` or Resource `view=events`. `transaction-http.ts:439–440` refuses selected SSE when the parsed changefeed intent is `start: now` and any Last-Event-ID field is present. The application Event parser admits only `after` (plus `view=events` for Resource sources); it does not implement or promise `start=now` (`transaction-events.ts:154–173`). Do not add a start parameter to Event Sources to fix this gap.

The existing law strongly motivates resuming from the header, but does not explicitly settle its precedence over **start=now**: the Scope changefeed section and its canonical catalog row say it overrides **after**. A minimal normative interpretation/clarification is needed before treating this combination as settled implementation work. Recommended choice: **on an otherwise valid Scope changefeed SSE request, one Last-Event-ID checkpoint selects exclusive replay instead of the original start=now intent; without that header, start=now retains its existing first-observation anchoring.** This should not broaden URL grammar or invent a new cursor.

If root confirms that as the authoritative interpretation, the production repair is very small. No store, Event projector, schema, cursor format or pump change is indicated by this inspection.

## Source and limits

Exact frozen checkout `/Users/dbox/repos/wt/vickie/vickie-schema-property-change`, HEAD `61ed530ac6a460598e0e312f5d28ec0e6fa2a33d`, tree `22e448fe0f517c0d6300b732de55494bd70c7938`, matching main `581bad0de6f5eb098b920147f6a7e09809c553ea`. Local source/spec/test inspection only. No tests, builds, servers, network, agents, source/spec edits or outward actions. This report is not execution evidence. Metadata incident-Link membership was not reassessed.

## Existing law and exact implementation

| Source | Established contract / limit |
| --- | --- |
| `docs/specs/bdp.md:5965–5975`, Scope changefeed | Exactly one explicit starting intent: after a checkpoint, retained genesis history, or start=now for future groups. Omitting both is an error. |
| `:6077–6079` | A start=now finite page identifies its captured admission anchor. |
| `:6110–6128` | Scope SSE messages contain complete groups, ID equals group checkpoint. Last-Event-ID overrides original **after**. A client-supplied reconnect checkpoint must be its atomically applied durable checkpoint; remembered transport progress is not sufficient. Stale/unavailable/foreign checkpoints fail explicitly rather than silently advancing. The text does not mention header precedence over start=now. |
| `packages/conformance/catalog/transactional-v1.json:1211–1231` | The required reconnect row selects the same after-specific precedence and durable-checkpoint laws; it adds no start=now precedence rule. |
| `docs/specs/bdp.md:6276–6282` | Application Event Source header is an exclusive replay cursor and overrides original after. This is a different source and token domain, not authorization to add Event start=now. |
| `:6563–6576` | Event IDs and checkpoints share a character profile, but genesis is a Scope checkpoint and never an Event ID. Similar spelling does not make token authorities interchangeable. |
| `transaction-changefeed.ts:86–105` | The URL must already satisfy exact route/query grammar and exactly one after/start. URL after is syntactically parsed even if a header later overrides it. A header alone does not rescue changes/ with no starting query. |
| `transaction-http.ts:429–449` | Parse URL, select media, collect raw header instances, refuse the start/header combination, reject repeated SSE headers, then map one SSE header to `{after: value}`. Finite JSON ignores this header, including malformed/repeated values. Authentication occurs earlier at368. |
| `transaction-service.ts:257–280` | For an after intent, use that checkpoint. For start=now, capture owner head once before minimum retries. Repeated attempts do not re-anchor. Both use existing minimum wait and request signal. |
| `transaction-store.ts:1651ff`, `transaction-changefeed.ts:54–85` | The actual owner checks the supplied checkpoint in its Scope/epoch/view and retained position domain. Reconnect need not be implemented by reading a mutable last-position field. |

The present guard prevents the existing one-header mapping from running for start=now. Removing that guard after a ruling would naturally turn the combination into the existing after-checkpoint path. Re-anchoring at current head instead would discard disconnected history; silently ignoring an incompatible checkpoint also conflicts with the stated explicit-failure rule. Nevertheless, the current after-specific wording alone does not distinguish “header overrides initial start intent” from “this pair is invalid”; choosing a normative400 rejection would also require a ruling, not merely swapping the current500 status.

## Minimal decision

**For an otherwise valid `changes/?start=now` SSE request carrying one valid retained Last-Event-ID checkpoint, should the authority resume exclusively after that checkpoint?**

- **Yes — recommended:** the URL specifies initial observation; a later reconnect header carries the client's actual replay position. Changefeed reconnect continues across disconnection without a new head anchor. Clarify the Scope reconnect sentence to cover the initial start=now intent as well as after. Preserve the client obligation to supply only an atomically applied durable checkpoint; the server cannot prove application from the header.
- **No — reject the combination:** explicitly define the refusal and require a client to reconnect with after=durableCheckpoint instead. This is implementable but needs a specified error/status and changes the usable reconnect pattern. Do not choose invalid-parameter by guess or leave a500 as a protocol rule.

Do not select “ignore the header and start anew”: it risks losing intervening groups and makes stale checkpoint handling misleading. No separate decision about metadata Events, Snapshot minimums, absent resources, EventSource browser convenience, or new transport is required here.

## Smallest boundary after recommended ruling

1. `packages/server/src/transaction-http.ts`: remove only the start/header development refusal, retaining URL validation, raw repeated-header rejection, media selection and the existing single-header `{after: ...}` mapping. This file is also owned by the active C1 lane; root must serialize or ordinarily integrate the two narrow deltas. Do not touch C1 predicates/signals or other request precedence.
2. `packages/server/src/transaction-sse.test.ts`: replace only the relevant500 expectation at274 and add actual resume/discrimination cases. Preserve the rest of the raw grammar/auth/media table. Existing start anchoring test at351–379 remains unchanged and is an explicit regression gate.
3. A bounded addition to the existing source-private installed Scope SSE test can observe a real start-now initial connection, durable group capture, disconnect, offline commit and same-URL reconnect. Preserve ordinary installed configuration/payload/source binding and current timeout/cleanup limits. No conformance-row promotion or new runner is required.
4. Separately authorized minimal normative wording/citation alignment if root chooses clarification. No new schema member is necessary. Do not broaden a selected catalog claim merely because this one case becomes observed.

## Concrete discriminators

- **Disconnected history must be replayed:** establish a start=now stream, commit and completely record/apply groupG1, disconnect, commitG2, then reconnect the same start=now URL with G1's genuine checkpoint. ObserveG2 first, not an empty newly anchored stream. CommitG3 after connection and observe it next. Compare complete groups against finite `changes/?after=G1` results and independently expected mutation effects. Assert G1 is not repeated. This fails both ignored-header and head-reanchoring implementations.
- **Current-end reconnect:** reconnect with the latest applied checkpoint, then commit a later group; see only that later group. Reopen the owner and repeat with retained genuine checkpoint, preserving checkpoint bytes and exact group data.
- **Token/authority validation:** under start=now use empty/malformed/repeated raw headers, a real application Event ID, and checkpoints with incompatible epoch/view or beyond the real head. They must reach the existing corresponding invalid/foreign/expired paths, never fresh-start success. Do not derive new error precedence where existing after-header tests already establish it; use that same path as the expected behavior and independently inspect actual Problems/context.
- **Minimum/conditional ordering:** the effective header checkpoint survives minimum waiting; a commit occurring during that wait is included, not skipped by a new start anchor. With GET/HEAD conditional fields, validate checkpoint/minimum before304/412 and retain required current context. HEAD returns no frames and consumes no long-lived stream slot.
- **Unchanged initial intent:** start=now without header still anchors once before the existing minimum wait. Valid after+header behavior stays unchanged. URL after+start together, omitted starting query, unknown/repeated query parameters remain parser refusals even with a header. Finite JSON start=now continues to ignore Last-Event-ID; do not accidentally apply SSE reconnect controls there.
- **Nondisclosure and lifecycle:** unauthenticated requests remain401 without context; actual reconnect/disconnect/shutdown leaves no child or stream slot. Counterfactuals must distinguish ignored header, re-anchoring, inclusive replay and accepting a foreign token; a200 response alone is not sufficient.

The fix, once its narrow precedence ruling is confirmed, reuses the current durable-checkpoint path. It does not implement a new Event Source feature and does not establish whole Transactional qualification.
