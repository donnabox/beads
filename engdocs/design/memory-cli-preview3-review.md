# Memory CLI Preview 3: review and documentation validation

October 21, 2026. This records review and docs validation for an implementation-pending CLI proposal. It does not qualify an installed command, store adapter, gate, evaluator or release.

## Review coverage and dispositions

The initial native Codex review covered the four-file proposal and its then-current corrections and ended with no substantive findings. Later review and settlement are recorded below. A tentative finding requiring `--view generic` was withdrawn after inspecting PR176 source, help and actual execution receipts: that flag is optional and deprecated in the target candidate. This is not evidence that the baseline already implements the new CLI.

Author review corrected cycle labels: revisiting an ancestor while traversing both directions does not establish a directed cycle. Two authored cases now cover a parallel-edge graph and a DAG with a mixed-direction closed display walk. Ordinary shared references must not be labelled cycles. The draft also explicitly refuses an aggregate answer when any requested root is invalid/unavailable or a query budget fails; a legitimate depth frontier remains distinct.

Actual external review was attempted against the original packet below. Claude Code (`claude -p`, configured model) returned no report before the 660-second timeout (exit 124). Gemini/Antigravity returned no review because a headless command permission was denied, despite process exit 0. Neither initial attempt is approval. The fresh successful Claude review and its settlement below supersede the initial coverage limitation; Gemini coverage is still absent.

## Source identity

The first table identifies the original external-review packet; the second identifies the final sources after native review and corrections. Hashes cover file bytes, not repository commit identity. The initial packet contained 21 read expectations and five authoring expectations; the settled revision below has 22 reads and five authoring expectations. They are authored cases, not executed tests or a new wire representation.

### Original external-review packet

| File | SHA256 |
| --- | --- |
| `docs/reference/bdp/cli/types/memory/memory.md` | `dce76434b646580d945c1d37b1f3ee6f7572437d0a008cd24aa03533dc94d485` |
| `docs/reference/graph-cli-specification-draft.md` | `c662088045f447beb579a01a47fd24eb8332e07aea774a833931516617e9d6c6` |
| `engdocs/design/memory-cli-preview3-cases.json` | `63469ede50616f70cd47828cdedbf8db5590a09fdc5e683a9b660ac4ff5173ad` |
| `engdocs/design/memory-cli-preview3-implementation.md` | `131f1d1b0300e9aeaf55b3017d743e1bf902a9dc2c1e4b37aa50a73269e2d25f` |

### Published PR206 packet and fresh Claude input

| File | SHA256 |
| --- | --- |
| `docs/reference/bdp/cli/types/memory/memory.md` | `abd8d98caefa0045f10fc1ed5db4a190bac61366c452889bf103461e1c61f30e` |
| `docs/reference/graph-cli-specification-draft.md` | `86ec158a46b76595bcfd18d92e826c3bf56242faa841c66e777c4c944a475ce3` |
| `engdocs/design/memory-cli-preview3-cases.json` | `ad3f0387368dad42bfc41a07ccd9141f9fc157a18a642fc52044b10a4d3a9bb7` |
| `engdocs/design/memory-cli-preview3-implementation.md` | `250ec082f6a44f237b7391edba7f6a5e82f14220c8e3fdf90b23bb17be97b70b` |

## Actual documentation checks

- Bazel-backed `make check-docs`: pass, including docsync, doc freshness and pinned CLI flag/reference checks; repeated successfully after the final semantic wording/case corrections.
- `./scripts/generate-cli-docs.sh --check`: pass against the existing v1.2.2 documentation pin; no generated file or pin changed.
- Pinned Mintlify 4.2.687 broken-link check: pass with the bundled supported Node 24.19.0 runtime. The host Node runtime was unsupported; the supported-runtime run is the qualifying check.
- Actual Mintlify browser render: HTTP 200, H1 “Memory CLI specification”, three tables and zero JavaScript errors. A Playwright screenshot was inspected. The owned preview process was stopped afterward.
- Case-reference integrity and `git diff --check`: pass. These structural checks do not execute the graph scenarios.

The graph presentation addition leaves the existing source/qualification ledgers intact. The Memory facade preserves Trish's availability paragraph verbatim from [assessment gate PR205](https://github.com/donnabox/beads/pull/205), head `492e45d1bd5e1786c01e710efd2457e62ee5d908`; its source paragraph SHA256 is `dfb30310dc3e4a8393ba33edc446c8e9c255b0e66987141816958415a7704f26`. Keep one copy when reconciling the overlapping docs hunk. Its target policy does not qualify that runtime candidate.

## Publication and qualification limits

This docs-only branch targets `codex/preview3-integration` at `b6d0493baf7c0a52f68df5a943158290e3b75878`. Native Dependency-mode behavior, runtime sources, generated output and installed contracts are unchanged. The new commands need implementation and exact-source runtime/provider qualification.

The new-branch pre-push hook compares against `origin/main`, which includes inherited integration-branch Go/Bazel changes, rather than this PR's documentation-only base. Publication uses an explicit `--no-verify` push exception to avoid launching a duplicate full runtime suite for inherited work. Normal commit hooks and the affected docs checks run. This exception is not evidence for any inherited runtime capability or a waiver of hosted integration checks.

Raw local receipts and the render screenshot are retained under `/Users/dbox/.local/state/melinda/memory-cli-publication-20261021/`; the durable source hashes and results above preserve the review boundary without claiming those local files are portable artifacts.

## Fresh Claude review and settlement

A fresh actual Claude Code review returned exit 0 after 627.44 seconds on October 21. It reviewed the four-file source packet at PR206 `86e1efcf63ae84736a63975d12ebb01250eb0d50`, matching the preceding published-source hash table. It used the configured default model with no override. Tools were disabled and the complete bounded packet was supplied on stdin; this was source-packet review, not independent inspection or execution of the rest of the repository. No files were modified by Claude.

Claude verified the example/case arithmetic and selection semantics, then returned four High, ten Medium and six Low findings (the last records review-context limits rather than a defect). The author checked the cited contracts and settled them as follows. This is a disposition record, not a claim that Claude re-reviewed the corrected successor or approved it wholesale.

| Finding | Disposition |
| --- | --- |
| 1: root membership lost by global display expansion | Corrected: independent top-level root trees, expansion/emission per root, attributed frontiers and a worked diamond example. Aggregate counts deduplicate identities. |
| 2: reverse/parallel Link placement unspecified | Corrected: choose the expansion tree from admitted directions/stages first; remaining selected Links appear once per root as references, preserving authored arrows. Concrete reverse/parallel example included. Final native review caught and corrected an overly broad incident-recursion draft that would have reparented diamond nodes against the requested direction. |
| 3: documentation gate scope unknown | Verified actual passing docs receipts and `scripts/check-doc-flags.sh`; it does not validate every nested proposal invocation. Plan records that publication precondition and limitation without adding an allowlist or claiming installed support. |
| 4: `decision` selector ambiguity | Clarified existing code/source tests: bare native Issue classification stays native; qualified workspace selector chooses Memory Decision. No new selector policy. |
| 5: direct subjects versus depth | Corrected: direct, not depth-tunable; `--depth` refuses; secondary declarations are out of scope, not a frontier. |
| 6: direction applied to induction | Clarified both normative docs: direction controls reachability only; induced matching Links retain either authored direction. |
| 7: gated retyping path | Clarified gate interaction in retyping section without altering Trish's verbatim availability notice. |
| 8: legacy nominal registrations wording | Clarified inherited selected-catalog policy and cited exact baseline source tests; no removal added by this CLI change. |
| 9: discovery-scope lines | Defined the two human scope lines and their applicability; machine binding remains common-adapter-owned, not an invented field. |
| 10: expected JSON keys mistaken for wire schema | Explicitly labelled expectations logical operation results, not serializer field names. |
| 11: multi-root staged frontier shape | Added explicit frontierByRoot to the staged authored case. |
| 12: mixed-cycle and second-parent test ambiguity | Checked frozen typedef bytes against the preserved hash: Requires-family cycles may mix Defines; only all-Defines cycles are forbidden. Added a mixed-cycle read case and deconfounded the existing second-parent refusal fixture. |
| 13: replacing flat presentation assertions | Plan distinguishes preserved selection/projection tests from intentionally superseded flat-text assertions, preserving attribution. |
| 14: draft availability framing | Adopted draft title, status first and non-imperative lead. Retained the agreed canonical filename/home; no gratuitous relocation or shared navigation change. |
| 15: Requires heading/hash object | Renamed heading Required source specification and specified the hashed file bytes. Exact input revision/hash unchanged. |
| 16: named properties through JSON | Clarified any admitted property may appear in properties JSON when its named flag is absent; duplicate sources still refuse. |
| 17: automatic ID generation | Clarified optional id and reuse of the shared allocator. |
| 18: unrelated subjects meanings/catalog peers | Clarified reverse Subject To command versus Evaluation subjects property; explicitly named Conformance peers. |
| 19: generic examples relying on unselected flags/defaults | Converted the two generic cycle cases to logical operation inputs with explicit depth, not assumed public argv/selector spellings. |
| 20: unsupplied context/repository authority | Recorded review limits; relabelled coordination link as a decision record, not execution evidence. Existing release ledger remains qualification index linking exact source/CI receipts. |

The actual fresh report SHA256 is `68e7614c16ae22af5e8a7f944bceae83f73c15028e5abe0ca3214818d775c0f0`; its local stdout and receipt are retained in the `claude-followup` directory under the local receipt root above. The frozen BDP input was read directly at its existing revision and verified against the unchanged SHA256 before settling finding 12. No Type definition, installed pin, runtime file, navigation or GitHub proposal issue body was changed by this settlement.

Final native follow-up caught one remaining tree-placement contradiction: recursively expanding all selected incidence would move a diamond's gamma under delta via a reverse Link. The correction separates the direction/stage-derived expansion tree from remaining selected-Link references, including references to later canonical appearances. This preserves the original directed diamond example and staged supporting paths.

### Corrected successor source hashes

| File | SHA256 |
| --- | --- |
| `docs/reference/bdp/cli/types/memory/memory.md` | `0a8e7a672c050ce1fe0d0c087ba5383c6aa6025845532ba4b2a14bf378196f44` |
| `docs/reference/graph-cli-specification-draft.md` | `73274fe5db0f8550106caade2e094834b9fc39fb02e7f9fc28c09c0676534880` |
| `engdocs/design/memory-cli-preview3-cases.json` | `d1ff432a6e26728e2b1bdcc230c826ed07026f8f57d60389d3203b12b3603d5d` |
| `engdocs/design/memory-cli-preview3-implementation.md` | `bbbc8d9f4d94607e27142c4767eaeb801f4fbdf10e161fb5c6e3aee9c3619335` |
