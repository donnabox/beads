#!/usr/bin/env bash
# One command: ./scripts/bdp-call-parity/run.sh [--out DIR] [--part a|b|both] [--keep]
#
# Builds bd and the Dolt this go.mod pins into a work directory, unless BD_BIN and
# DOLT_BIN name them, then measures. See README.md.
set -euo pipefail

export BD_DISABLE_METRICS=1 DO_NOT_TRACK=1 BD_DISABLE_EVENT_FLUSH=1 DOLT_DISABLE_EVENT_FLUSH=1 PYTHONDONTWRITEBYTECODE=1

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$here/../.." && pwd)"
work="${BDP_PARITY_WORK:-${TMPDIR:-/tmp}/bdp-call-parity-work}"
out="$work/out"
pass=()
while [ $# -gt 0 ]; do
  case "$1" in
    --out) out="$2"; shift 2 ;;
    *) pass+=("$1"); shift ;;
  esac
done
mkdir -p "$work/bin"

bd="${BD_BIN:-$work/bin/bd}"
if [ -z "${BD_BIN:-}" ]; then
  (cd "$repo" && make build BD_BUILD_OUTPUT="$bd" >"$work/build-bd.log" 2>&1)
fi

dolt="${DOLT_BIN:-$work/bin/dolt}"
if [ -z "${DOLT_BIN:-}" ] && [ ! -x "$dolt" ]; then
  build="$work/doltbuild"
  mkdir -p "$build"
  cp "$repo/go.mod" "$repo/go.sum" "$build/"
  # The repo's own replace directives point inside the repo; Dolt needs none of them.
  sed -i.bak '/=> *\.\.*\//d' "$build/go.mod"
  (cd "$build" && GOFLAGS=-mod=mod go build -tags gms_pure_go -o "$dolt" github.com/dolthub/dolt/go/cmd/dolt)
fi

exec python3 "$here/parity.py" --bd "$bd" --dolt "$dolt" --out "$out" "${pass[@]+"${pass[@]}"}"
