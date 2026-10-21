# Native parser-refusal code review

Verdict: no actionable blocking findings in the frozen three-file draft. Reviewed against plan-v3, root authoring release, native v3 closure and actual-Claude plan dispositions. This is read-only code review, not a compiler, formatter or execution result; exact checks and runtime validation remain required.

## Behavior and scope

The production diff moves only the existing parser400 return ahead of semantic unknown-minimum support classification and updates that guard's comment. It preserves lexical tuple validation, authentication/method/target precedence, earlier alias/receipt/Event/snapshot/changefeed routing, parser404 and valid strict-route guards, and outer HTTP error semantics. It neither adds URL grammar nor changes successful read behavior. Unsupported/repeated query400 is distinguished from reference limit0/query-before-existence parity. Invalid-query foreign context also reaches the same pre-owner parser return, as explicitly dispositioned; otherwise-valid foreign-context policy remains unchanged.

The predecessor collection callback changes exactly its two affected parser500 sentinels to schema-checked400. Strict Type/incident private500, successful unknown-minimum503, materialized400/413 and hidden404 behavior remains tested; its18 requests and8 fake-clock waits remain unchanged. No parser, service, owner, specification, manifest, shared helper or public-admission edits are present.

## Oracles, topology and bounds

The three new source callbacks implement37+13+17=67 actual planned handler calls: three ordinary setup GETs plus64 refusals/controls. Literal opaque is used with real setup context; current and absent controls remain distinct. The local facade delegates actual owner methods, and setup proves an owner entry before counters reset. Every refusal checks no later owner method, wait or token-generation entry. Cleanup closes the service-owned database (or owner if construction failed), removes the owned directory, and aggregates failures. The direct source callbacks create three stores and no listener/child.

Direct paired HEAD checks derive Content-Length from the already-counted GET JSON bytes and require body undefined, preserving the distinction between representation length and transmitted bytes. Problem schema/type/code/retry, private/no-store and context absence are checked; reference metadata is not claimed as a universal protocol rule. Strict Type and parser404 controls remain bodyless private500 without owner entry. No accidental incident-anchor owner-free assertion is introduced.

The separate installed callback has1 setup+12 opaque+4 current/absent+6 noncollection+2 precedence=25 HTTP exchanges, one owned process/store and the existing operator/reader fixture using the operator handle. All24 refusals assert400, no private marker and no projected context; GET checks schema/type/code/retry, HEAD checks zero wire body octets and reference Problem media. It does not infer owner absence from the wire. Existing5s request,90s work,120s callback and80-exchange bounds and cleanup helpers are unchanged. The complete preexisting installed source is an exact byte-identical prefix, including the historical42-exchange callback, which remains a distinct preservation selection.

The frozen selection separates67 new direct calls from33 existing direct calls and91 emitted-runtime requests (18+5+68), across seven stores and three in-process listeners, zero product children. The68-request Type/incident regression count follows1 seed+3*(1 baseline+2*(1 ordinary+4 conditional/timeout+2 context refusals)+2 negotiation/auth)+2 limits+6 queries+8 missing-route checks. The explicit dist import still requires a fresh bound candidate build before collecting this HTTP suite. No specific strict-TypeScript defect was identified by inspection; formatter/typecheck acceptance is deferred to the authorized diagnostic stages rather than inferred.

## Frozen identity

Independently verified HEAD d4f3f4be88c7837b71d6646518ca3589d464ba3d, exactly three dirty files and no untracked files, all679 tracked-source hashes, all before/after snapshots, all context/preparation pins, and full Git binary diff equality.

- source-binding.json: 9c7f735f5ca0a031536749acef3d5293b6fc2eb6742bd4240ef960126eae098e
- draft.patch: 894f249da3251b68a11bab58a2b419ebe18fa7878c45ceb78f15114b850c5d8f
- packages/server/src/transaction-http.ts: 9ee59d7b2f5ecf6bb1ace7bacb70fd5124343dc9320439afeb3c28cf2f174584
- packages/server/src/transaction-http.test.ts: 92581a8ee0f5b43bd843e6541d0b8821c88e07e99edf026be381da7a754d9855
- apps/bdptest/src/transactional-collection-minimum-installed.test.ts: 87cbd8ee973e8d2368cf6c4563ee6fe96849e55c82f9ae3b8b49aad87718f7e7

No tests, product imports, builds, dependency installs, formatter execution, source edits, commits or push were performed. Only this guarded report was written; PR150's checkout was untouched.
