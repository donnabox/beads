# Bounded test council dispositions

Native and actual external direct reviewers inspect the same test-only goal.
Gemini excluded. The initial frozen packet and its failing test evidence are
preserved; repaired-source proof is separate.

| Finding | Disposition |
| --- | --- |
| Medium: empty ContentHash oracle rejects legitimate normalization | Repaired; initial failed run preserved. Final control checks source-cache/counter sentinels do not survive and caller fields stay unchanged, without pinning derived computation timing. |
| Low: unsupported reflected shapes silently skipped | Repaired with explicit scalar arm plus default refusal in this guard. |
| Low–Medium: prose refusal policy need not have behavioral enforcement | Repaired without production refactor: same concrete fixture helper drives prior refusal test and this guard; each classified refusal executes preparation and checks ErrValidation/named field. Missing and reverse-stale fixture coverage fail. |
| Low optional: unexported fields skipped | Deferred outside exported SDK input scope. These cannot be set by an SDK caller; walking standard-library time.Time internals would conflate implementation bookkeeping with source input. No present SDK input defect established. |
| Parent: exact encoding/json omission tag | Use exact tag equality; json:"-," is a serialized field named hyphen, not an ignored field. Current seven tags unchanged. |

Final native confirmation has no actionable finding. Actual initial and
repaired-source reviews completed exit0/empty stderr. The latter suggested one
Low clean/seeded hash comparison, accepted; final tiny confirmation completed
exit0/empty stderr with no actionable finding;
no full source/runtime qualification claimed. Duplicate JSON tag collisions,
custom marshalers and generic round-trip drift remain outside this guard.

The repaired reviewer's Low cache-control finding is addressed by comparing
seeded preparation with an identical clean preparation. This detects dependence
on caller cache values without constraining recomputation timing; timestamps
are excluded from the SDK hash. Exact final source checks remain3pass/0fail/0skip
and pinned lint0issues.

Final actual review is [preserved](close-external-review.txt). Its static limits
remain explicit; non-blocking diagnostic-display suggestions are deferred.
Both requested seats have no remaining actionable finding on this bounded
test-only unit. No third provider or runtime proof.
