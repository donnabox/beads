# Packet08a dispositions

Frozen b7eefc9 vs486ad62. Native and Claude reviews complete. Claude read all supplied199KB source/diff without truncation. Gemini reviewed nothing: individual quota exhaustion/429, exit3. No retry planned.

**No new confirmed correctness defect after checking omitted upstream context.** Exact external source checked natively is listed below; do not imply Claude saw it.

| Claude item | Disposition |
|---|---|
| A1 stored Resource kind not cross-checked | Optional defensive invariant assertion, not a proven wrong-kind commit. Both template and fact derive from the same carrier-validated creating member and its pinned attempt/slot. Divergence needs invalid input already rejected upstream, corruption or an internal implementation defect not demonstrated here. |
| A2 named update used as creator | Unreachable input. `read-update-values.ts:309` invokes schema validation before name handling; sequenceUpdateBead at schema1609 permits only update members, operation and idempotencyKey, with unevaluatedProperties:false. It has no name. Wrong-kind/forward/unknown bindings are then checked at328–342. Root independently confirmed. |
| A3 references apply only to operation0 | Current callers intentionally provide references only to one-operation sequence-member carriers. An assertion could document this internal helper precondition; no live multi-operation misuse found. No refactor requested. |
| A4 unresolved ownership CHECK / nullable sentinels | Optional schema-hardening proposal, not current invariant failure. Unresolved templates are owned; execution rejects unresolved state; resolution and terminal commit clear it under exact ownership. Startup removes pending rows before facts. Do not introduce a new private storage-format change merely for representational preference. |
| A5 epoch-stop test uses substituted epoch | True coverage limit: private store lacks restore/epoch mutation and the test substitutes comparisonEpoch. Closed owner maps to temporary unavailability coherently. Public restore/supervisor qualification remains excluded; no new mandatory closed-owner response proposed. |
| A6 missing first creator allocation | Valid creates produce resource identity; retained/expired creator case is exactly one-operation sequence member. Missing allocation is a corrupt invariant and correctly fails closed. No valid receipt lacking creation identity demonstrated. |
| A7 reads outside atomic | Metadata/fact reads are synchronous under the one exclusive owner and enter guard. No competing writer or suspension exists. Comment about atomic ownership applies to admitKey, not every read in the module. |
| A8 note/test precision | Low editorial follow-up: distinguish tests actually run by the author from counts inferable from a diff; a comparison-loop helper delegates comparison logic; one test title says waits although later tests exercise waiting. Existing807 result is external author evidence, not a number the council independently ran. No claim that all807 tests were newly added. |

## Omitted-context questions resolved or bounded

- Q1: named non-creators and @name creation IDs fail closed schema/binding validation; no committed-prefix counterexample.
- Q2: semantic-identity.ts tags every encoded node. Missing alias is an object node; comparison's `encoded[1][0] !== "string"` short-circuits before string operations and returns conflict. The unbound sentinel is a private Symbol, not manufacturable JSON. No malformed-codec value path found.
- Q3: creation evaluator's successful result includes resource.id; expired successful creations retain allocated identity. Internal invariant errors are not ordinary member problems.
- Q4: inherited Idempotency keys at3174–3187 normalize a creator that failed without allocation to the distinguished unbound marker; only the explicit transient class bypasses comparison. idempotency-conflict is permanent, not one of idempotency-in-progress/rate-limited/temporarily-unavailable at2827–2833. No new transient conflict semantics inferred.
- Q5: Read+Update problem table at3385 fixes binding-unavailable as request/400/never; Transactional sequence inherits it.
- Q6: MemberCreatorBinding at member-identity.ts35–38 is exactly bound/unbound/transient; no fourth variant.
- Q7: operation records are validated/frozen/branded upstream. The member carrier literal remains inside the branded PreparedTransaction boundary; no public bypass introduced. Whole-sequence byte and depth bounds remain: independent pinned-Node probe of a66,134-byte one-member sequence is refused with `development sequence byte limit` at both486 andb7. See08a-sequence-byte-bound-repro.{mjs,stdout.txt}. No database used.

## Supplied versus additional context

Claude was supplied all current store/batch/sequence/dependency/comparison files and full changed tests, but not protocol schema/parser, semantic codec, member identity or resource evaluator. Native resolution above inspected those additional unchanged sources explicitly. The same bounded packet went to all providers; this supplementary source inspection is native-only, not retrospectively attributed to Claude or unavailable Gemini.

The packet07 cancellation cause defect is closed atb7 by exact external reproduction (08-sequence-cancel-cause-verification), preserving both immediate/later causes without cycles. No other source edits, public claims or evidence generation occurred.
