# Bounded document-patch execution plan, council revision

**Plan only, pending council disposition; no implementation started.** Source baseline is frozen75f279cec24ef8d7eeae0fde56120ddab82ec907 (direct0b34fe7 successor). Earlier plan is preserved at `read-update-document-patch-plan-pre-council.md`; review is `read-update-document-patch-claude-plan-review.txt`. All S1–S11 findings are accepted with the dispositions below. No fixed coverage increase is promised.

## Authority and exact gap

The normative BDP spec governs the reviewer clause map: the row's conformance-table obligation plus the full relevant sections reached by its catalog citations. The canonical catalog determines IDs/order/selection; its selectedText excerpts do not override or exhaust the governing prose. The map is an explicit reviewer artifact in `docs/design/read-update-executable-progress.md`, not a new machine-enforced proof format. Coverage JSON bindings remain the mechanically generated complete ordered action/assertion list; missing clauses and non-crediting breadth are disclosed in its reason. No new inventory schema, catalog requirement, public evidence format or framework.

Full sections already read at0b: `docs/specs/bdp.md:2131–2343` through Explicit Link operations, Property changes/Property-change values2197–2267, Type validation2581–2612, Validation/results and admission2641–2695, Mutation limits2697–2718, full validation-failed definition3442–3474, normative row table6698–6740.75f does not change these laws.

Two actual gaps guide this slice:

- `read-update.singleton.document-patch-merge-replace` is proposed whole but currently has **Bead-only actions**, whereas the normative table expressly requires both Beads and Links. It needs real Link observations or an honest demotion.
- `read-update.singleton.document-patch-validation` is omitted. The current installed cohort has **no properties schemas** (`scripts/installed-reference-observation.mjs:100`, `packages/server/src/installed-reference-configuration.ts:134–144`, `packages/protocol/src/reference-domain.ts:29`). It can observe document-shape refusal but cannot discriminate effective-Type rejection or metadata's exemption from such a schema. This row stays false in this baseline slice.

Current baseline129required/40authored/28proposedwhole/101uncovered is provisional evidence accounting. The merge row may remain whole only after the full reviewer map is supported. Otherwise demote it and update all pins/doc totals; do not preserve a misleading green count.

## Clause map and smallest useful work

| Row | Crediting clause set | Planned row-local observations | Explicit limits |
| --- | --- | --- | --- |
| `document-patch-merge-replace` | Both Beads and Links: selected-document relative pointers, object add replacing only the named member, array insertion/append, root replacement of the complete object, and `{}` clearing the selected object. | Keep the existing Bead arm and mirror its clause set on an unowned Relates Link: object add, array insert/append, null assignment, root replacement, separate properties/metadata clear with omitted other document retained. Static complete document expectations plus independent full GETs. | Existing missing replace/remove checks are supporting/non-crediting breadth for this row; do not add pointer-escape or array replace/remove permutations. Symmetric clause set for both kinds. No Link execution ⇒ demote. |
| `document-patch-atomicity` | One update applies both documents atomically; equal combined JSON values preserve revision and the entire revisionProvenance despite different supplied attribution. | Preserve the existing Bead successful combined update, each direction of missing-path failure, independently validated saved context and reordered-key no-op. At most one unowned-Link combined effect and one reordered-key no-op, with returned and GET full records. | Extra Link observations are supporting breadth, not another obligation or coverage award. Drop the proposed duplicate stale-guard probe; expected-revision has its own row. Do not imply concurrent visibility/linearizability from final state. |
| `document-patch-validation` | Complete resulting properties must satisfy every effective Type contract; metadata has no Type schema; non-object/Type-invalid result fails the whole update without changing either document. | New same-ID scenario with about six probes below and independent unchanged-record GETs. Canonical Problem plus expressible diagnostic fields. | Remains whole=false: effective-Type clause, schema-exemption discriminator and message-cause semantics are explicitly unobserved. No cross-row whole credit to update-bead, type-contract or diagnostics rows. |

Validation probes, each starting from independently verified seed state:

1. Bead properties root→null, with an effectful valid metadata proposal.
2. Bead metadata root→array, with an effectful valid properties proposal.
3. Unowned Link properties root→scalar, with an effectful valid metadata proposal.
4. Unowned Link metadata root remove, with an effectful valid properties proposal.
5. Positive arbitrary object metadata on a Bead (a schema-looking invalid value is only object-domain evidence with no installed schema).
6. Positive properties-only Link update proving omitted metadata retained.

Each refusal must be422/validation-failed/validation/never, with canonical nonempty diagnostics, an independent complete Resource GET equal to the previously validated record, and unchanged Link endpoints where applicable. The valid opposite-document patch makes partial persistence detectable. A properties failure may short-circuit before metadata evaluation: this is an **unchanged-state oracle**, not proof both documents were evaluated. The metadata-failure direction exercises the opposite ordering. No cartesian product of shape/kind combinations, duplicate guard matrix, retention claim or artificial schema.

## Diagnostic assertions we can actually express

Use the existing canonical `harnessInstalled_readUpdateProblem` wrapper unchanged: it requires diagnostics on validation-failed, and every entry has nonempty message. For these deterministic single-root failures, assert the required first entry, `instanceLocation == ""`, type/schemaLocation both absent, and metadata failures' document equal`metadata`. For properties, use the current implementation's omission branch and clearly disclose that omission is one legal spelling meaning properties, not a protocol prohibition against explicit`properties`. No request-body pointer relabeling.

**Message names the cause/metadata remains unobserved semantically.** We cannot express that without exact message text or another matcher; minLength is not that proof. Do not invent a matcher, change the schema wrapper or mark this clause proved. Other diagnostics/truncation obligations remain omitted.

## Fixtures, routes and configuration

Reuse already-routed identities in fresh per-scenario stores: `w1.a`/`w1.b` for Beads, `w1b.unowned-source` → `w1b.target` via Relates, and `w1b.unowned-link` (`links/w1b-unowned`). Those are real Task endpoints without ownsOutgoing. Use actual HTTP create/read setup in the row; no cross-row store sharing, direct evaluator or DB mutation. Link afterstate GETs require existing fixed routes; the captured-identity escape hatch is Bead-only and must not be widened.

No new fixture bindings/routes are intended, so both exact-configuration digest pins remain75f's `89d471885903a3e1dda4939f54993a88696e1b8c336dfef287a0c28966c47d0a`. If authored route ordering/set changes unexpectedly, inspect why; do not refresh pins blindly. If a new route is genuinely required, disclose it and recompute both pins as reviewed artifacts. The unused-binding route control remains unchanged. No Type or policy/configuration change in this slice.

Insert the new validation scenario at its canonical position between atomicity and delete-bead. Requirement order comes from the existing canonical selection, never a separate list.

## Sensitivity controls

Controls follow real HTTP capture; original observations remain intact. They test assertion sensitivity, not target behavior or transparent wire forwarding. Require one application and the named failed assertions.

- Link merge: corrupt a retained sibling/array value or cleared object in a schema-valid returned Resource; separately corrupt the independent GET state. The Bead arm passing cannot hide a Link defect.
- Atomicity: corrupt one document only in a refused update's subsequent full GET. For no-op, alter only response revision/provenance while GET remains correct; then the independent GET envelope separately. Reuse existing controls where they exercise exactly these clauses; do not multiply identical probes.
- Validation: substitute a coherent, fully schema-valid alternative Problem tuple, including compatible diagnostics handling; the intended code/status/retry assertions must fail. Inject **both** type and schemaLocation (valid absolute schema keyword URI), so dependentRequired still passes and the generic-shape absence oracle is tested. Change metadata document to a schema-valid wrong selection and require its direct assertion to fail. Corrupt the afterstate record independently.

No code-only Problem substitution that fails schema before reaching the intended semantic discriminator. No isolated type/schemaLocation insertion. No fake write outcome used as qualification evidence.

## Budget and ownership

Nodes are the first structural gate: measure actual JSON nodes against100000, then formatted bytes against4MiB, depth128 and container10000.75f baseline is44,731nodes/1,686,328bytes/depth11/largest-container228. Forecast only a small addition; compute actual values before freeze. Do not increase limits or minify to conceal them. Existing reference limits and the target body policy stay unchanged.

Owned files, in a **new successor** after approval:

- `packages/conformance/matrices/read-update-v1.json`: mirror Link clause arm and new bounded validation row; no TX/retention scenarios.
- `packages/conformance/test-support/read-update-coverage-v1.json`: mechanical bindings and honest reasons/whole flags.
- `apps/bdptest/src/read-update-matrix.test.ts`: bounded controls, authored/uncovered inventory pins and both exact-route digest pins **only if actually changed**.
- `packages/conformance/src/cumulative-artifact.test.ts`: whole-ID literal list/count updates **only if demotion occurs**; newly authored-but-incomplete validation preserves uncovered count.
- `docs/design/read-update-executable-progress.md`: one reviewer clause map, limitations and count consequences.
- Fixture file only if unavoidable; current proposal needs no new binding.

Merge-complete branch: whole-ID list/count unchanged. Demotion branch: remove merge row from the whole-ID literal list,101→102uncovered and28→27whole in matrix/doc pins. New validation scenario is bounded-observation/whole=false with mechanical assertion inventory; no count-driven qualification.

No shared runtime/evaluator, installed configuration, normative/public schema, canonical requirements, artifact loader/planner or public admission files. Root resolves cumulative JSON ownership with Transactional/retention lanes before editing; only one writer at a time. No broad cleanup or new assertion API.

## Later effective-schema cohort, explicitly separate

The existing installed factory supports this through private administrative input; no new runtime feature is needed. Root must approve/configure a new cohort before claiming Type-schema evidence. Nonempty contracts.descriptors **replaces** the shipped set: re-emit all13 Scope-remapped descriptors via `referenceTypeArtifacts(scope)` and the same stringifyJsonValue bytes, then append fresh local Type IDs; host schema resources under`${scope}schemas/`. Never overwrite a contract at an existing ID.

A useful minimal parent/child properties closure could discriminate parent-only failure, child-only failure, required-member removal and whole-result cross-field dependency, with valid metadata supplied alongside failures and the same otherwise-disallowed value accepted in metadata. That is proposed future fixture design, not law. The Type inventory, fixture and actual contract/configuration identity would change; the complete inherited cumulative suite must run against that **one** configuration before whole credit. Separate tests cannot be stitched into current baseline evidence. Keep update-bead/type-contract/diagnostics unqualified until their own complete row-local laws are observed.

## Done predicates and review dispositions

Done when the reviewer clause map agrees with actual action/assertion observations; merge both-kind law is witnessed or demoted; validation remains honestly incomplete; controls fail named semantic assertions; clean→prebuild typecheck→build and affected conformance/app tests pass; fresh packed actual HTTP observations bind clean source/configuration/fixture and expose all uncovered IDs; format, boundaries and unchanged Read evidence pass. Freeze before council, no public qualification or outward action.

S1 ownership/pins: accepted, explicit branch above. S2 existing routes: accepted. S3 reviewer artifact vs mechanical bindings: accepted. S4 expressible diagnostics/message gap: accepted. S5 authority/non-crediting breadth: accepted. S6 six-probe scope and no duplicate stale guard: accepted. S7 symmetric merge arms: accepted. S8 nodes-first budget: accepted (actual measured counts govern; no speculative byte-to-node threshold claim). S9 coherent schema-valid controls: accepted. S10 short-circuit unchanged-state limit: accepted. S11 full remapped closure replacement: accepted. No implementation follows from these plan edits until root/council disposition.
