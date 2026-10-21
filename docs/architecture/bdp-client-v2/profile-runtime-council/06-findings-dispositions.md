# Packet06: disposition of final-repair reviews

Target `3a8f9d553a2944bc7d1d7812a284f0160dcbe798`, base `59915ea916a210f5ce36611eb6d799f7c3124056`. Both external providers report complete supplied diff. Claude's raw CLI stdout contained only its final text block; complete assistant text was recovered with provenance in `06-final-repairs-claude.recovery.json`. Raw outputs are preserved.

| Finding | Disposition |
|---|---|
| Claude A1 / Gemini count mismatch | Current count is25 plans/264 actions/827 assertions. Historical pending slice had797 assertions:705 prior +62 pending +30 interleaving. Later alias readbacks add29 and allocation adds1. Historical slice reports are not a claim about the newest combined tree. Do not relabel them. |
| Claude A2 matrix budget prose | Low clarity follow-up: the pending slice kept per-action budgets; separate CI repair gave the aggregate test30s. Prose could distinguish those levels. Assertions and inner deadlines remain unchanged. |
| Claude A3 undefined History domain | Existing PartI historicalResolution definition and schema already specify version1. The added table row inherits that earlier definition correctly. |
| Claude A4 no-op tombstone storage | Prose now honestly explains existing growth. Automatic no-op forgetting would alter replay semantics and can apply a previously inert patch to later state. No new retention decision accepted. |
| Claude A5 routing precedence | Intended404 repair for unknown development routes, auth checked first. No supported GET target removed. |
| Claude A6 alias no-body weakness | SDK read-update transport rejects any nonzero response bytes on307. Omitted parsed body does not mean text bytes are ignored. |
| Claude A7 / Gemini alias allocation root | Resource allocation and alias path share the namespace. The meaningful exclusion is beads/reserved; comparing against alias/reserved would be vacuous. |
| Claude A8 missing get / A9 four releases | get is required and SDK methods close over transport, so detaching is safe. Existing four-member disconnect plans use four releases; three-member connected plans use release-response. No current missing-method/hang trigger. |
| Claude A10 slugger duplicate mismatch | Primary github-slugger implementation uses a global collision loop, matching the reviewed index's a,a-1,a-1-1,a-2 progression. No mismatch demonstrated. |
| Claude A11 tab slug mismatch | Existing slug handling already strips tabs; GitHub slugger removes matching punctuation/control characters then replaces spaces. No current tab heading contradiction. |
| Claude A12 repeated parse cost | Unmeasured local validation optimization; prior code also reparsed per citation. No runtime or correctness finding. |
| Claude A13 import style / undefined response / duplicate helper calls | Current lint/types pass and supported manifests call each release once. No reachable defect supplied. |
| Indented or stacked HTML aliases | Current parser's supported ATX/standalone-alias/fenced subset is bounded; no such current corpus target shown. Stacked aliases fail closed rather than permitting an invalid citation. Future general CommonMark support is not a readiness requirement. |
| Gemini CRLF heading concern | Disproved with pinned Node24 regexp probe: heading and HTML alias are extracted correctly from CRLF lines. |

All three native-review bypass reproductions from the earlier held boundary candidate were closed by the corrected follow-up, independently retested by the parent. This packet does not treat that old candidate as integrated.
