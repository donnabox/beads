#!/usr/bin/env bash
# Run the complete graph qualification matrix in bounded sequential partitions.
# Keep raw output and exit receipts; no test is retried, omitted or synthesized.
set -euo pipefail
if [[ $# -ne 1 || "$1" != /* ]]; then
  echo 'usage: graph-c0-storage-tests.sh ABSOLUTE_EVIDENCE_DIRECTORY' >&2
  exit 2
fi
: "${BEADS_GRAPH_TEST_SERVER_PORT:?ordinary disposable Dolt server port required}"
GRAPH_C0_STORAGE_OUTPUT="$1"
cd "$(dirname "${BASH_SOURCE[0]}")/.."
mkdir -p "$GRAPH_C0_STORAGE_OUTPUT"
for artifact in storage-tests.log graphstore-test-inventory.log graphstore-partitions.json; do
  if [[ -e "$GRAPH_C0_STORAGE_OUTPUT/$artifact" ]]; then
    echo "refusing to overwrite $GRAPH_C0_STORAGE_OUTPUT/$artifact" >&2
    exit 2
  fi
done
export GO_TEST_PKG_PARALLEL=1 TEST_VERBOSE=1 TEST_TIMEOUT=20m
export TEST_RUN=''
./scripts/test.sh -count=1 -list . ./internal/storage/graphstore > "$GRAPH_C0_STORAGE_OUTPUT/graphstore-test-inventory.log" 2>&1
python3 - "$GRAPH_C0_STORAGE_OUTPUT" "$(git rev-parse HEAD)" <<'PY'
import json, pathlib, re, sys
out = pathlib.Path(sys.argv[1])
text = (out / 'graphstore-test-inventory.log').read_text()
package = 'github.com/steveyegge/beads/internal/storage/graphstore'
assert re.search(r'^ok\s+' + re.escape(package) + r'\s+', text, re.M), 'inventory did not complete'
names = re.findall(r'^Test\S+$', text, re.M)
assert names and len(names) == len(set(names)), 'empty or duplicate compiled test inventory'
assert not re.search(r'^(?:Example|Fuzz)\S*$', text, re.M), 'unpartitioned executable root kind'
assert not re.search(r'^(?:FAIL|=== RUN)', text, re.M), 'inventory unexpectedly executed or failed tests'
parts = []
covered = []
for name, pattern in [('a-h', '^Test[A-H]'), ('i-l', '^Test[I-L]'), ('m-z', '^Test[M-Z]')]:
    tests = sorted(test for test in names if re.match(pattern, test))
    assert tests, 'empty partition ' + name
    parts.append(dict(name=name, pattern=pattern, tests=tests))
    covered.extend(tests)
assert sorted(covered) == sorted(names) and len(covered) == len(set(covered)), 'partition coverage differs'
(out / 'graphstore-partitions.json').write_text(json.dumps(dict(
    source=sys.argv[2], package=package, timeout='20m', tests=sorted(names), partitions=parts), indent=2) + '\n')
PY
: > "$GRAPH_C0_STORAGE_OUTPUT/storage-tests.log"
run_partition() {
  local artifact_name="$1"
  shift
  local result
  if ./scripts/test.sh -count=1 "$@" > "$GRAPH_C0_STORAGE_OUTPUT/$artifact_name.log" 2>&1; then
    result=0
  else
    result=$?
  fi
  printf '%s\n' "$result" > "$GRAPH_C0_STORAGE_OUTPUT/$artifact_name.exit"
  cat "$GRAPH_C0_STORAGE_OUTPUT/$artifact_name.log" >> "$GRAPH_C0_STORAGE_OUTPUT/storage-tests.log"
  return "$result"
}
run_partition graphstore-a-h -run '^Test[A-H]' ./internal/storage/graphstore
run_partition graphstore-i-l -run '^Test[I-L]' ./internal/storage/graphstore
run_partition graphstore-m-z -run '^Test[M-Z]' ./internal/storage/graphstore
run_partition storage-other ./internal/configfile ./internal/storage/issueops \
  ./internal/storage/schema ./internal/httpapi/graphread ./internal/httpapi/bdpwire ./internal/httpapi
