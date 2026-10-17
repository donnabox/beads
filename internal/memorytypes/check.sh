#!/usr/bin/env bash
# Explicit F0 acceptance entrypoint. Fails if the cached oracle is unavailable;
# never downloads dependencies or hides a missing oracle behind a test skip.
set -euo pipefail
package_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$package_dir/../.." && pwd)"
cd "$repo_root"
./scripts/test.sh "$@" ./internal/memorytypes
uv run --offline --no-project --with-requirements "$package_dir/oracle-requirements.txt" \
    python "$package_dir/check_schema_oracle.py"
