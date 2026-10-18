# BDP specification change workflow — discussion note

**October 17, 2026 · Nonnormative planning material.** No new merge, rebase, pin change or specification adoption is authorized by this note. It uses the already-inspected next-wave packet; it is not a fresh branch audit.

## Controlling boundary

Donna's direction is explicit: **Vickie must not absorb new specification changes until her current reference Write work is complete and landed in main.** She continues against her recorded `a562` implementation/source checkpoint and `67772742…` specification digest. Those shorthand identifiers come from the handoff; this note does not invent their full SHAs or redefine her checkpoint.

Trish continues Types completion on `codex/trish-types-profile-completion`, based on shared-design head `b12665d64c2aaf6172bd1cdab409cef31ef28f44`. Current public main `3ab9cdf5d45a2ea49a84ab191eac77c9aeaa19d0` includes the bounded PR159 landing. **Main, shared design and Vickie's source/contract pin are three different things.** Advancing one does not authorize moving another.

This overrides suggestions to feed new Types, Memory, provenance, batch-subset or Event semantics into Vickie's current implementation. Review and planning can continue without creating that dependency.

## Recommended arrangement

Keep **one adopted contract baseline and separate reviewed proposal branches**, with the existing shared branch as a draft-composition workspace.

| Surface | Purpose | Rules |
| --- | --- | --- |
| Public main / adopted Core+Write | Stable implementation target and accepted runtime increments | Preserve the adopted protocol until a later explicit adoption. Main may already contain labelled WIP; that does not change the adopted contract. Qualify each runtime increment at its exact candidate. |
| Existing `codex/trish-spec-integration` | Compose related drafts and discover conflicts | It is not Vickie's target and is not an adopted “next version.” Do not automatically merge the entire branch into main or into her work. |
| `codex/trish-types-profile-completion` | Bounded Types draft completion | Finish agreed text, examples, schemas and explicit open decisions against shared `b126`. Keep optional profile and implementation status truthful. |
| Owner proposal branches | Independently reviewable changes such as Memory, provenance, batch or Events | Each names its baseline, depends only on selected prerequisites, and lists exact normative differences. Keep current adoption and implementation separate. |
| Later reconciliation/adoption branch | Integrate a selected approved amendment set after Write completes and lands | Create only with authorization. Start from then-current main; preserve history through ordinary merges; resolve semantic differences individually and validate the resulting exact commit. |

A proposal branch can be reviewed against the shared branch when that makes its diff useful. Record the adoption baseline separately: a PR base is a review convenience, not a normative dependency declaration. Avoid using “based on main” as evidence that a draft matches the adopted contract.

**No new integration branch is necessary now.** The existing shared branch already serves draft composition. Create a later branch only when there is a selected adoption set to reconcile. Otherwise another integration branch would merely add a fourth moving target.

## Authority and status under `specs/`

`specs/` remains the sole location for specification text and associated normative schemas/examples. Proposed specifications also live there, prominently labelled WIP; **location makes them eligible sources of specification work, not automatically adopted requirements**. Do not create a competing future contract in `docs/` or a tracking issue.

Use the existing `specs/README.md` inventory to identify, separately:

- adoption state: current adopted contract, WIP proposal, or superseded;
- exact contract revision and dependencies;
- reference implementation and Beads implementation status/evidence;
- unresolved decisions and owner.

Each draft should say which existing clauses it would replace and what remains unchanged. Its schema IDs/version declarations must not impersonate the current accepted schema bundle. Published/merged WIP is review material; tests of its examples are draft-shape checks, not implementation qualification. Preserve the Memory inventory reconciliation requirement rather than overwriting the table with an older row.

Planning notes, review reports and coordination ledgers can live outside `specs/`, but link to specification text and explicitly carry no competing normative rules. #59 remains the integration/decision ledger, not a second specification.

## Semantic delta ledger

Maintain one compact ledger in #59, linking each proposal's exact revision and its own detailed acceptance cases. For every amendment, record:

| Field | Required content |
| --- | --- |
| Current contract | Adopted clause/schema definition and exact source pin. |
| Proposed delta | Concrete before/after behavior; distinguish settled product direction from unsettled wire choice. |
| Status/owner | Draft, accepted product direction awaiting propagation, adopted protocol, implemented or qualified; named accountable owner. |
| Dependencies | Only necessary amendment IDs. No blanket “wait for all Types” or “wait for all Memory.” |
| Consumers | Request/response DTOs, selectors, CLI mapping, persistence/history, Events, SDKs, importer and discovery as applicable. |
| Checks | Existing tests to preserve; intentional changes to expectations; new positive/negative cases; installed/client behavior and capability evidence required. |
| Transition | Version/capability negotiation, retained-data/event compatibility, old-client refusal and migration policy where applicable. |
| Disposition | Selected adoption batch, explicitly deferred, or superseded with a replacement pointer. |

A passing existing test is not a reason to silently retain old semantics after an accepted amendment; a new fixture is not permission to change current semantics before adoption. Ledger entries make that distinction reviewable.

Start from the next-wave groups already recorded: exact Type affiliation/optionality; endpoint re-pin; deletion/revert; provenance/Inception/Activity; atomic subsets/discovery; lifecycle Events; interchange; Memory query/projection. These are amendment candidates, not one combined release commitment.

## How existing-contract code can still land

Vickie's current Write completion can proceed in bounded source increments against her frozen contract. Code that implements that contract, regression tests and necessary packaging fixes can be qualified and submitted independently of draft semantic adoption. Every candidate must identify what it actually implements and leave unsupported profiles unadvertised.

After authorization for a specific source integration, reconcile only repository/layout and existing-contract changes necessary to land that source. Preserve her normative pin and accepted minimum-read semantics; do not smuggle in new Type mutability, terminal deletion, Activity, batch-subset or Event redesign requirements through conflict resolution. Unexpected semantic differences return to the ledger for explicit disposition.

A source increment landing does **not** satisfy Donna's “current reference Write work complete and landed” gate by implication. Vickie’s measured completion boundary is plan D1–D5: pinned required Read/RU/TX on one installed target/configuration, ordinary cold credentialed discovery/startup, applicable installed law fixtures, fresh cumulative qualification and exact-tree CI. Record that evidence and main landing before feeding new specification adoption into her work. This creates no additional manual completion-approval gate beyond existing main/adoption decisions. The earlier PR159 bounded landing alone does not lift this restriction.

Once that milestone is complete and landed, select the next amendment set, create/reuse an authorized reconciliation branch against then-current main, and qualify the exact combined result. Keep historical evidence pinned to its original source. Downstream consumers change their contract pins deliberately, separately from merely updating implementation binaries.

## Alternatives and tradeoffs

1. **Recommended: owner proposals + existing shared draft branch.** Parallel design, small reviews, no pressure on current Write; requires an explicit ledger and selected adoption rather than a whole-branch merge.
2. **One next-protocol integration branch for everything.** Convenient to view a hypothetical future, but readily becomes a competing authority and binds unrelated work together. The existing shared branch may provide this view only if clearly labelled a draft composition; do not create another or promise wholesale landing.
3. **Publish every WIP directly on main.** Possible in principle with very clear status, but increases accidental consumption and the chance of changing Vickie's apparent target. Not recommended during this constraint. Existing WIP on main stays WIP; do not remove it or reclassify it without review.

## Practical next step

Continue Types completion and independent proposal review. Prepare the owner survey and semantic ledger without changing Vickie's input. Ask Donna to select unresolved product/wire decisions when she returns, one at a time. Ask Vickie only for the completion boundary and evidence through the already-authorized coordination process; do not assign new protocol changes. No merge permission is requested or inferred by this note.

## Existing-contract decisions for the dinner agenda

Vickie identifies two ambiguities in the contract she is already implementing. Keep these separate from future semantic amendments:

- **S1:** evaluate a retained Snapshot/continuation minimum against the current serving context or the immutable retained anchor.
- **G2:** decide whether metadata-only changes to an incident Link belong in the Bead’s Event stream.

Neither is settled by this note. They gate affected full-support claims, not D1 changefeed or D3 discovery preparation. Ask bounded questions after dinner if they remain unresolved; do not infer a ruling from implementation preparation. Vickie reports the three-file private D3 discovery plan ready, with ordinary trusted startup and evidence admission still required.

Vickie mirrored the workflow discussion in her sole plan at `docs/architecture/bdp-client-v2/spec-change-workflow-dinner-20261017.md`; reported source checkpoint `c15dba22` supersedes `4c54afa2`. These are owner-reported references, not a fresh qualification by Trish.
