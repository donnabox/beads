# Vickie cold-session handoff — 2026-10-18

This is a continuation guide, not a second plan or protocol authority. The sole implementation plan remains `donnabox/beads`, branch `codex/vickie-bdp-client-design`, `docs/architecture/bdp-client-v2/design.md`. The new launch manifest is the entry point. Read this guide before the long historical ledger; only load an old packet when needed.

## Safe starting state

The old role session is `2026-10-08T20:26:35Z-81e5`. Only an operator-authored standalone `resume vickie` may claim a new session. Use compare-and-swap, then verify the remote claim. Old session guard stays pinned to its original claim; never overwrite it to make old timers valid under a new claim.

Automation `vickie-bdp-profiles` is PAUSED for this handoff. It targets old chat `01a11d2c-bcf1-73c3-9045-4f0b291e6d49`. After the new session claims the role, migrate/rebind the automation to the new chat and replace its old expected-session condition before activating it; do not activate a timer that still targets this old chat. No actual next review has launched. All three old-chat workers are idle/completed. Their handles are not transferable; use durable reports and assign fresh workers explicitly. No scoped review/product process was found in the final process inventory. Unrelated app/fleet processes were untouched.

PR167 is READY, not draft, on exact `a56254d682cf4623cab22a552e3d2f695f691795`, tree `144f278c7f5a08516c4ac99ba553d4cc25bcf560`, sole parent PR166 `9f89271e2563a105410789ca8ca1e77dfb4d0b9f`. Local source checkout is clean; all679 current and committed files equal the reviewed/tested source. Exact push CI39082528652 passed all27 steps,9464 tests/186 files/one optional skip, plus7preflight. Native terminal audit2d87ce4d/root8eba253e accepted. Ready state/body/head verified and Trish notified directly in bus comment6243838927. Do not repeat publication, reconciliation or notification.

## First next action: CLOSED changefeed PLAN packet

Local folder:
`/Users/dbox/repos/wt/vickie/bdp-integration-logs/event-admission-next-serving-plan/clean-candidate-rebind-1/claude-plan-review`

Durable mirror: same relative path below `docs/architecture/bdp-client-v2/` in the sole plan branch.

The packet is frozen and **not independently mechanics-reviewed yet**. No review, source authoring, formatter, test, build, install, or service is released. `rootLaunchReleased=false`.

- preparation-index: `ca6c1a1ee86b976b461189711dbe3ca18a079c33c3bcfaf3b3decf9ff9864519`
- binding: `70fe6cd622b7d39e70ecf62d106e49dfbcd1cc28858d0f2954f45a5f1fe01502`
- prompt: `205e1311187873a94f8f02cbfee3e2dc3faf8429816ba74bf156bba749eccc77`
- runner: prefix `fd1a9827`; source reanchor prefix `e6d294a6`
- preparation reports679 source files/24 copied contexts/11 artifacts/34 indexed entries.
- accepted underlying clean plan index `cfa8178e`, plan `99128481`, native `86025dd6`, root `0de5b30e`.
- current design/ledger snapshot is checkpoint `6335248e877924157109b674d0f36628c7491dd3`; copied original CI-pending snapshot remains historical and later rootgreen8eba253e is separately bound.

Read `preparation-status.md`, full `plan.md` in the parent directory, prompt, binding and runner. Have a fresh independent native worker audit mechanics. **Session transition needs an explicit bounded runner/guard successor:** current runner calls old-session `L/guarded-action.py`, which correctly refuses the new claim. Preserve that guard and original packet unchanged. Create a separately named new-session guard and, if required, a separate review packet successor with the exact guard-path/session attribution change and genuine delta/index closure. Do not weaken claim checks or silently rewrite reviewed pins. Complete native/root mechanics inspection before a flag-only launch release.

Then run exactly one bounded actual Claude PLAN attempt (1800s outer/TERM1790, default model, Read/Glob/Grep only, no retries). Preserve all raw/partial/terminal records. Exit0 alone is not approval: inspect findings, independently audit terminal provenance/source invariance/process closure, and explicitly dispose findings. Only then release bounded source authoring on a separate branch from exact-green a562. No Gemini. This is the user's two-provider review requirement, not permission to invent a protocol change.

## What the next changefeed slice does

Six-file plan composes current minimum controls with finite explicit-after/genesis and SSE changefeed reads. Existing finite start-now behavior is already implemented. Preserve old direct strict owner/service calls and old finite-start-now intent; add an explicit current-changefeed intent.

Changefeed ordering is intentionally different from Event ordering: cursor decode/epoch/view/future/retained-row/corruption checks precede semantic minimum evaluation. Retained cursor SQL is legitimate; do not claim zero SQL before minimum. Capture the tuple once before waiter preflight and retain it over retries. Subsequent stream pump keeps minimum undefined. Finite reads ignore Last-Event-ID; SSE retains override/repeated-header ordering. Minimum precedes media/conditional/slot shortcuts.

Two distinct seed commits must yield actual CP1/CP2; positive finite and SSE paths preserve whole commit groups. Historical10+4 installed exchanges remain unchanged; new10 are prospective. Source15 callbacks/60HTTP and all derived inventories are prospective, not completion percentages. Accepted N1 correction uses callback-local B/C/F ownership, bounded drain and verified closure before removing owned stores; no unbounded shared global cleanup or deletion of unknown ownership. Preserve default5s source limit, D3s work/4s cleanup-inclusive limit, installed request5s/work45s/cleanup10s/callback60s/outer120s/TERM115, no retry or estimate-based widening.

The original concrete-plan-1 and N1 finding stay frozen; concrete-plan-2 changed exactly four cleanup paragraphs. See `event-admission-next-serving-plan/concrete-plan-2` index8c9c12d4/native95e452c1/rootb1a6f502. The later clean-source rebind changes chronology only.

## Finite remaining delivery — do not confuse PR readiness with completion

The finish line is the installed reference target implementing pinned required Read, Read+Update and Transactional, ordinary cold credentialed discovery/startup and fresh cumulative applicable evidence on the same actual target/configuration. TX is the next usable milestone. Keep distinct nontransactional Update response/retry/atomicity coverage; do not route Update through TX merely to reduce bookkeeping.

1. D1: next changefeed minimum composition above. Event replay/SSE now covered by PR167; earlier current-resource/Type/receipt minimum composition by PR166. Audit remaining required paths against the pinned contract, not old private-gap notes. History/replay/streaming already exist.
2. D2: S1 retained Snapshot/continuation minimum vs immutable anchor, and G2 metadata-only incident-Link membership in Bead Events remain explicit semantic gates. Resolve affected claims, continue unaffected work. Do not silently turn an implementation choice into protocol law.
3. D3: TX twelve-target wire Scope/descriptor/Directory discovery and ordinary installed RU/TX startup/admission. Private RU nine-target cold discovery works; TX assembly exists but HTTP/public admission is incomplete. Internal discovery minimum handling in PR166 is not public TX discovery.
4. D4: applicable installed schema/ownership, authorization/view and erasure/recovery qualification. Existing visible→empty framing is implemented; explicitly justify applicability rather than silently dropping cases.
5. D5: fresh cumulative RU/TX inherited Read evidence, generators/verifiers/admission and combined exact-tree CI. Historical Read seals and private claimEligible:false probes do not meet this finish line.

D3/D5 audit: `profile-delivery-cold-client-audit/audit.md`, SHA prefixd04dcc4e/index2993d2ba. D1 audit: `profile-delivery-consistency-audit/audit.md`, ec4f713f/indexbe0987ba. No need to re-audit all history before the next bounded step.

## Boundaries and ownership

Donna owns product rulings. Trish owns authoritative cross-workstream integration ledger `gastownhall/bdp#59`, Core/Write/Type integration and merges. Vickie owns this implementation plan, code/tests and evidence. Janet owns CLI/staging/playgrounds/shared databases/Preview2. Do not touch their workstreams.

Defer new protocol changes until implementation lands. No new-main chase/merge, main landing, shared-design integration, cleanup/closing old PRs, release, Preview2 promotion or downstream pins without explicit Donna authorization. The old frozen reconciliation was a one-time authorization, already completed as PR159; do not redo it or fold later work into it. Ordinary ancestry on reviewed working branches only. Optional History, YAGNI redesign, Types and unresolved provenance proposals remain separate. Missing-Type minimum-first is an allowed implementation choice, not a fresh product question.

## Frozen checkpoints to preserve

| PR | Exact head | State |
|---|---|---|
|150|d4f3f4be88c7837b71d6646518ca3589d464ba3d|Nominated original frozen candidate|
|159|d7197e97a29bff8f075028419afad05388e2e655|Ready frozen reconciliation; parents150 then frozenmain8d6671ed19b0c0406428ccc2d2161768d01e8b9b; tree630a94d0ca1545f898cb8a8b614d4b7d2dd6d454|
|160|5ceb213b632e17f28714684efc5e35a62b282bf6|Ready alias opaque minimum|
|161|bbfa5b2cac0e381b497384e3a8e29cd937ea2738|Ready finite changefeed start-now minimum|
|162|d9c3eb9e61b1982b9bd56aa77e4cdc114384e886|Ready Scope Event SSE minimum|
|166|9f89271e2563a105410789ca8ca1e77dfb4d0b9f|Ready current-observation minimum|
|167|a56254d682cf4623cab22a552e3d2f695f691795|Ready Event replay/SSE admission minimum; current clean implementation basis|

All individual readiness/validation/raw-log limitations are in the sole design and corresponding packet directories. None authorizes landing main.

## Recent evidence and lessons that must survive

`event-admission-replay-validation/helper-successor-1` is PR167's final successful validation; its source binding deliberately records historical dirty-base9f. Do not rerun that old execution gate after commit a562: root-commit-verification proves byte identity, not rerun attribution. Nineteen fresh main +four fresh required checks passed; only original two install/tool-binding bootstrap stages were reused with17 dependency and11 tool identities checked. Local17 source callbacks/60unselected skips/12new4legacy reports; external historical5/two children and new10/one child are separate. Source679/build476/tar321/consumer870files17links/eight production versions. Payload da80f6d7. Historical entryLiteral/closeEvent absent; new witness has exit+close0/live complete Event frame→drain/pending0/store removal.442 absent sampled PIDs means401fresh-stage+41original-bootstrap, not442fresh. Lint121warnings6infos, not warning-free.

Exact CI raw3886658 bytes SHA7b00eabdbd0527b97a8545f31d280c57c6bfb452bcde33a24678f865b15d0ea7. Twelve source and historical5/new10 CI reports parse; eight other masked reports do not. CI workspace-dist/development-local payload495fa1d4 is distinct from local externalda80. CI packed smoke reportsda80 but claimEligible:false. Remote lifecycle is callback-reported; no independent remote PID/filesystem absence inspection. Historical Read89-row seal retains rune5f7d0c/evidencedbaf63a.

Preserve original failures, never overwrite/relabel: actual format2aea74c1/exit000ef2e8 fixed by five-file syntax-equivalent formatting; historical callback passed then auditor23726468/exitbf930622 failed because it expected newer-only entryLiteral. Corrected strict common resolved-entry/hash/payload audit requires literal only for new callback; earlier preparation review missed this and remains recorded. `format-successor-1` old15-stage passes were not reused; helper-successor-1 reran19+4. No predicate weakening, defaults or deadline widening.

PR166 also preserves original actual review timeout404f6ec3 separately from completed review3c6261d6/final961041fd; bad F2 ordering review/source failure1bbb32e4 was corrected by oracle761e451b restoring Relation,Task canonical HTTP sort. Helper audit numeric0-vs-empty-list failure8a076bfa and malformed unexecuted patch8bd489 stay historical. Do not blindly implement reviewer suggestions. Exact-source preservation and genuine before/after deltas matter; hashes alone do not establish semantics. Source-derived totals are not a wire census or profile completion percentage.

## Locations, tools and operational notes

- Implementation checkout: `/Users/dbox/repos/wt/vickie/vickie-event-admission-replay-minimum`, branch `codex/vickie-event-admission-replay-minimum`, origin gastownhall/bdp.
- Sole plan checkout: `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout`, branch `codex/vickie-bdp-client-design`, origin donnabox/beads.
- Local packet root L: `/Users/dbox/repos/wt/vickie/bdp-integration-logs`. Indexed durable mirrors live below the sole plan's `docs/architecture/bdp-client-v2/`. Scratch generated consumers/build products need not be transferred; reproducible packet inputs and completed logs are pushed. Do not clean old stores/artifacts indiscriminately.
- Node `/Users/dbox/.nvm/versions/node/v24.16.0/bin/node`; pnpm `/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs`; Python `/opt/homebrew/opt/python@3.14/bin/python3.14`. Ambient Node26 is wrong. Build before process/full tests.
- Actual external reviewer `/opt/homebrew/bin/claude`; follow required review skill. No Gemini. Immutable review copies allow live plan checkpoints without changing reviewed context.
- `L/guarded-action.py` is pinned to old session. Preserve it. `L/checkpoint-service-successor.py` is a historical old-session copier and previously expanded the entire ledger into launch; do not blindly reuse it after resume. New session must maintain the compact launch manifest and supporting archives instead.
- Bus `donnabox/agent-coordination#1`; cursor lives in remote launch. Read only incremental comments. Trish ready notice6243838927; own status6118051413. Five old inbox pointers6208400206/6209150016/6209779214/6210373790/6232666628 were read at handoff: they concern prior editorial split/freeze and bounded reconciliation now completed in PR159. They are historical, not five new jobs. Keep no-main rule; do not replay reconciliation.
- Worker names in old chat: event_restart_plan_review (author/CI, completed); conditional_review_recovery (closed packet prep, completed); ci_repair_code_review (native audit, completed). The final next packet has NOT received native mechanics review. Fresh workers need exact bounded assignments; old handles may not resolve. Completed workers require a follow-up task, not a message, to reactivate within a chat.

## Next-session stop conditions

Stop if superseded, operator pauses, the full delivery predicates are verified, or no independent progress remains without a necessary human decision. Do not call a private increment full completion. Recheck role before outward actions; never reclaim automatically. There is no unresolved operator decision needed to inspect/rebind the next review packet. S1/G2 only gate affected claims.
