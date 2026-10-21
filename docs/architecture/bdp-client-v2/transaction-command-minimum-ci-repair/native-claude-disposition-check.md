# Native check — completed Claude delta dispositions

No contract ambiguity requiring a PR123 source change or blocking landing was found in the four findings. Deferring wording polish to the already required cumulative catalog/spec follow-up is reasonable; that follow-up remains outstanding before public qualification.

I read the complete actual Claude output324351ac3c02851b71520065284ee548bfea1737fc3cf4cb727e6819ea482201 and its exit record (exit0, sourceUnchanged=true), and checked against unchanged spec287a0a64bea0c1c3be51b5bb50d528283185ef7886a9c7402b4d4b5a7d5a0941.

1. The shared clause's generic “conditions on reads” could be clearer, but its operative MUST applies only to mutation POSTs. Transactional explicitly preserves minimum controls on reads, including subsequent receipt GET/HEAD. Thus no read-side exception or conflicting instruction is created. Claude expressly identifies no normative loss. Defer clarification; do not infer a weakened receipt minimum requirement.
2. “This profile and cumulative profiles” is less precise than “later cumulative profiles,” but Read exposes no mutation POSTs to which this rule could apply. The clause does not authorize any new target or mutation capability. Transactional inherits Read+Update, so the intended current applicability is intact. Defer the word “later” refinement; no product decision is needed.
3. An ignore obligation requires identifying the exact field names, not understanding their later-profile values. All three are literal names and the clause explicitly ignores presence/completeness/values while retaining ordinary transport constraints. Adding “copied” is optional orientation and must not accidentally condition the rule on how a field arrived. The rule applies equally to independently authored and copied request fields. No forward semantic dependency is required to implement ignoring them.
4. “Shared” versus “inherited” and link-text capitalization are editorial. The link resolves backward to the single operative rule and has no changed force or profile scope. Defer consistency polish.

The moved operative rule still preserves ordinary syntax/transport/header/body bounds, authorization, idempotency, carrier validation and expectedRevision. No schema/catalog/production/guard change is needed for this editorial disposition. The separate missing cumulative catalog mapping remains an acknowledged qualification follow-up, not evidence that this bounded implementation is qualified.

Read-only review and this guarded report write only. No source edit, tests, CI execution, reviewer rerun, push, merge or coordination action. Actual successful Claude review is now established by its saved output/exit; root still owns the final findings disposition and exact landing acceptance.
