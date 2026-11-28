# Preview 3 import council dispositions

Native Codex and actual Claude reviewed the same pinned implementation snapshot.
Both reported no confirmed implementation code blocker. Gemini was excluded by
Donna. This is source review, not runtime or release qualification.

| Finding | Confidence | Disposition |
| --- | --- | --- |
| Fake cancellation/deadline test | Both reviewers, Medium | Use actual canceled/deadline contexts at final staged writes, inspect independent live-context state, and check cleanup uncertainty. Runtime execution pending. |
| Apply example loses explicit prefix | Both reviewers, Low | Both dry-run and apply retain source-prefix. |
| Prefix config error loses cause | Claude, Low–Medium | Two wrapped errors preserve invalid-store and driver/context identity; pure CLI precedence checks pass. |
| Infrastructure config silently falls back on DB failure | Claude, Low | Import uses existing strict helper and propagates errors; no core helper change. |
| Human output omits adopted prefix | Claude, Low | Summary prints source Issue prefix; memory-only summary does not invent one. Pure guard passes. |
| Duplicate labels conflict with set contract | Claude, Low | Dedup isolated source copy before persistence/comparison; caller unchanged. Pure guard passes. |
| Nonserialized SDK routing/plane/partial state silently ignored | Claude, Low | Refuse before JSON clone. Derived row-version/content-hash remain excluded as documented projections. Pure guard passes. |
| Duplicate owned-Link reads | Claude, Low | Retain independent post-write graph validation within existing resource/read/time ceilings; measure engine cost before widening shared core interfaces. |

The self-contained final repair prompt was supplied to both seats. Native Codex
reports no confirmed introduced blocker. [Actual Claude bounded repair review](repair-claude.txt) and final confirmation
completed. Its confirmed caller-isolation test defect is repaired with a duplicate
before a later distinct label; actual JSON-minus-tag and field-specific validation
checks cover the SDK assumptions. Dry-run output now says it would adopt the
prefix; the strict read error carries its origin. Native final confirmation
reports no introduced blocker.

Claude's second pass requested a one-way cleanup assertion to permit legitimate
driver error variance; its final pass requested guarding the reverse direction
for clean rollback. The returned error chain does not expose a successful rollback
acknowledgment separately, and absence of sql.ErrTxDone cannot prove clean rollback.
We keep ErrTxDone => uncertainty, allow other uncertain cleanup errors, retain
context identity/unchanged-state assertions, and remove the test comment that
claimed to assert the clean branch. The conservative core classification remains
unchanged; no engine-specific inference or new core hook is introduced. The final
summary nit uses an explicit if/else. These dispositions resolve the concrete
introduced defects without claiming runtime evidence.
The genuine completeness fixture and fresh installed two-engine no-skip gate
remain mandatory. Older installed receipts do not qualify these repairs.
