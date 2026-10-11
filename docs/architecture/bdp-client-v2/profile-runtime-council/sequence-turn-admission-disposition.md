# Sequence turn classification: council disposition

Source: frozen `72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5`.

Claude02b A2 proposes freezing non-owned member projections at admission. The omitted inherited Read+Update clauses resolve this without a new product choice:

- Spec3234–3239: a member whose key could not be claimed because it is retained, expired, or in flight elsewhere is answered in its turn, as of that turn.
- Spec3255–3272: the four key outcomes are decided in the member's turn; unknown means execute and retain disposition.
- Spec4005–4014: Transactional sequence inherits the existing carrier semantics and changes the member disposition to a receipt, admitting all initially unknown keys atomically.
- Spec3993–3999: a forgotten failed receipt leaves an unknown key; a later presentation executes as new.
- Spec4044–4048: a member that finds another owner still pending at its turn returns in-progress without waiting or executing.
- Spec3970–3985: only an explicit direct request already waiting on unresolved comparison may retry admission after retraction, under one finite deadline.
- Spec4074–4075 denies that direct wait/retry machinery to sequences. It does not override first-turn key classification.

## Distinct cases

1. Unknown at admission: `admitSequence` reserves it in the atomic bulk step; `nextSequence` executes the stored claim under its exclusive owner. There is no late admission for this case.
2. Failed receipt retained at admission, forgotten before turn: the first member may change state; when the later member is reached, its old failed key may now be unknown. Turn-time classification admits it and executes as a first successful effect. Existing test511 demonstrates this implementation, but authority comes from the clauses above.
3. Foreign pending at admission, retracted before turn: first-turn classification sees unknown and may admit. Minimal trace: `other = store.admit(p,K,createA)`; `S = store.admitSequence(p,[createA/K])`; `store.retract(other.claim)`; `store.nextSequence(S)` creates A under a new owner. No waiting occurs and no member was previously answered transiently.
4. Foreign pending when actually reached: the member returns in-progress and advances. Later retraction does not cause this sequence to revisit that member. This is the prohibited wait/retry distinction the implementation preserves.

Disposition: reject A2 as a defect. Freezing all non-owned projections at admission would contradict the inherited turn-time rule. No implementation or normative edit required; a short explanatory test comment could be useful but is not necessary for correctness.
