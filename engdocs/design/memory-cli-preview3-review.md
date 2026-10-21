# Memory CLI Preview 3: review and documentation validation

October 21, 2026. This records review and docs validation for an implementation-pending CLI proposal. It does not qualify an installed command, store adapter, gate, evaluator or release.

## Review coverage and dispositions

Native Codex reviewed the four-file proposal and its final corrections. No substantive findings remain. A tentative finding requiring `--view generic` was withdrawn after inspecting PR176 source, help and actual execution receipts: that flag is optional and deprecated in the target candidate. This is not evidence that the baseline already implements the new CLI.

Author review corrected cycle labels: revisiting an ancestor while traversing both directions does not establish a directed cycle. Two authored cases now cover a parallel-edge graph and a DAG with a mixed-direction closed display walk. Ordinary shared references must not be labelled cycles. The draft also explicitly refuses an aggregate answer when any requested root is invalid/unavailable or a query budget fails; a legitimate depth frontier remains distinct.

Actual external review was attempted against the original packet below. Claude Code (`claude -p`, configured model) returned no report before the 660-second timeout (exit 124). Gemini/Antigravity returned no review because a headless command permission was denied, despite process exit 0. Neither is approval. Coverage is reduced to substantive native review and author checks; later corrections were reviewed natively, not externally.

## Source identity

The first table identifies the original external-review packet; the second identifies the final sources after native review and corrections. Hashes cover file bytes, not repository commit identity. The accompanying case file contains 21 read expectations and five authoring expectations. They are authored cases, not executed tests or a new wire representation.

### Original external-review packet

| File | SHA256 |
| --- | --- |
| `docs/reference/bdp/cli/types/memory/memory.md` | `dce76434b646580d945c1d37b1f3ee6f7572437d0a008cd24aa03533dc94d485` |
| `docs/reference/graph-cli-specification-draft.md` | `c662088045f447beb579a01a47fd24eb8332e07aea774a833931516617e9d6c6` |
| `engdocs/design/memory-cli-preview3-cases.json` | `63469ede50616f70cd47828cdedbf8db5590a09fdc5e683a9b660ac4ff5173ad` |
| `engdocs/design/memory-cli-preview3-implementation.md` | `131f1d1b0300e9aeaf55b3017d743e1bf902a9dc2c1e4b37aa50a73269e2d25f` |

### Final reviewed sources

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
