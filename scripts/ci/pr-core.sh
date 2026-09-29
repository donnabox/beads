#!/usr/bin/env bash
# Required fast PR Go test contract.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# shellcheck source=../.buildflags
source "$REPO_ROOT/.buildflags"
# shellcheck source=lib/timing.sh
source "$REPO_ROOT/scripts/ci/lib/timing.sh"
# shellcheck source=lib/test-env.sh
source "$REPO_ROOT/scripts/ci/lib/test-env.sh"

cd "$REPO_ROOT"

beads_test_env_enter

GO_TEST_PKG_PARALLEL="${GO_TEST_PKG_PARALLEL:-4}"
GO_TEST_PARALLEL="${GO_TEST_PARALLEL:-4}"

# Reuse the exhaustive sequential graphstore dispatcher. Each group retains
# the same race/short/30m/TestEmbedded policy; all other packages still run once.
# Unique evidence directories preserve failed/repeated local invocations.
mkdir -p "$REPO_ROOT/artifacts/pr-core-go-test"
pr_core_evidence="$(mktemp -d "$REPO_ROOT/artifacts/pr-core-go-test/run.XXXXXX")"
pr_core_args=(--output "$pr_core_evidence/evidence"
    --package-parallel "$GO_TEST_PKG_PARALLEL" --test-parallel "$GO_TEST_PARALLEL")
# nightly's Bazel equivalence consumer gets actual concatenated Go JSON only,
# including partial failed output; dispatcher diagnostics remain on the console.
if [[ -n "${BEADS_PR_CORE_GO_TEST_JSON:-}" ]]; then
    pr_core_args+=(--json-output "$BEADS_PR_CORE_GO_TEST_JSON")
fi

ci_time "pr-core go test" -- \
    python3 "$REPO_ROOT/scripts/ci/macos-go-test.py" "${pr_core_args[@]}"
