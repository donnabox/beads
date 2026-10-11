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
reports no confirmed introduced blocker. Actual Claude bounded repair review is
still active; collect its raw result and disposition before marking council done.
The genuine completeness fixture and fresh installed two-engine no-skip gate
remain mandatory. Older installed receipts do not qualify these repairs.
