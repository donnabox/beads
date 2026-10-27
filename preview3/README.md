# Beads Graph Preview 3

Preview 3 adds disposable graph-workspace import and structured Memory Beads, Links and CLI commands. Donna signed off on this preview on October 27, 2026.

Start with the [runbook](RUNBOOK.md). [OPERATIONS.md](OPERATIONS.md) contains optional HTTP and recovery exercises. The [literal Memory proposal fixture](MEMORY-PROPOSAL.literal-source.json) is for inspection; it is not a supported `bd import` format.

## Build the frozen source

The official source tag is [`graph-preview3`](https://github.com/donnabox/beads/tree/graph-preview3), commit `210f8dad21945af85746a0eac74e6187d54a3735`. It selects the qualified product source; these release documents are maintained on `codex/preview3-release`.

With the repository's Go/build prerequisites installed:

```sh
git clone --branch graph-preview3 --single-branch https://github.com/donnabox/beads.git beads-preview3-source
cd beads-preview3-source
make build BUILD_DIR="$PWD/preview3-build"
export PATH="$PWD/preview3-build:$PATH"
bd version
```

Use a fresh directory outside the source checkout for the runbook exercises. Preview 3 is not production-ready and does not support upgrading an earlier preview workspace. Preserve existing stores and use a copy of exported data for conversion. Local revision tokens and generated Link IDs will differ from the illustrative transcripts.

## Qualification

[Final-head CI](https://github.com/donnabox/beads/actions/runs/40427175960) passed at the exact tagged source: eight effective jobs succeeded and ten conditional jobs skipped. The [build receipt](reference/BUILD-RECEIPT.json), [old-generation refusal and catalog check](reference/LANDED-BOUNDARY.json), and [source equivalence receipt](reference/SOURCE-EQUIVALENCE.json) preserve their original provenance. Their local paths identify the qualification environment; they are not installation paths.

This is a source-only preview. No compiled download package is attached. It does not change the default branch, global installation, package-manager releases or BDP pins.
