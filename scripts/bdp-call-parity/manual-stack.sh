#!/usr/bin/env bash
# A throwaway bd serve for testing by hand, with the BDP reference client pointed at it.
#
#   manual-stack.sh start [--seed] [--token]   Start one and print how to use it
#   manual-stack.sh stop STATE_DIR             Stop it and remove its data
#
# Everything lives in a new temporary directory: its own Dolt, workspace, HOME and client
# state. It never reads or writes a real workspace, and telemetry is off. Stop kills only
# the processes this script recorded, plus any bd daemon whose arguments name its directory.
#
# Required: BD_BIN (the bd to serve), DOLT_BIN (pinned Dolt), BDP_CLIENT (a built checkout of
# the bdp repository) and BDP_NODE (Node 24.22.1 or the version that checkout pins).
set -euo pipefail

die() {
  echo "manual-stack: $*" >&2
  exit 1
}

telemetry_off() {
  export BD_DISABLE_METRICS=1 DO_NOT_TRACK=1 BD_DISABLE_EVENT_FLUSH=1 DOLT_DISABLE_EVENT_FLUSH=1
  export BEADS_DOLT_AUTO_START=0 BD_NON_INTERACTIVE=1
  unset BEADS_DIR BEADS_DB BD_ACTOR BEADS_ACTOR BEADS_DOLT_SERVER_PORT BEADS_DOLT_PASSWORD
}

free_port() {
  python3 -c 'import socket;s=socket.socket();s.bind(("127.0.0.1",0));print(s.getsockname()[1])'
}

wait_for() {
  local url=$1 tries=50
  until curl -s -o /dev/null "$url"; do
    tries=$((tries - 1))
    [[ $tries -gt 0 ]] || die "nothing answered at $url"
    sleep 0.2
  done
}

start() {
  local seed=false token="" self
  self=$(cd "$(dirname "$0")" && pwd -P)/$(basename "$0")
  while [[ $# -gt 0 ]]; do
    case $1 in
      --seed) seed=true ;;
      --token) token=$(python3 -c 'import secrets;print(secrets.token_urlsafe(24))') ;;
      *) die "unknown option: $1" ;;
    esac
    shift
  done
  local name
  for name in BD_BIN DOLT_BIN BDP_NODE; do
    [[ -x ${!name:-} ]] || die "$name must name an executable"
  done
  [[ -f ${BDP_CLIENT:-}/apps/bdp/dist/main.js ]] || die "BDP_CLIENT must be a built bdp checkout"

  telemetry_off
  local root
  root=$(cd "$(mktemp -d "${TMPDIR:-/tmp}/bdp-manual.XXXXXX")" && pwd -P)
  mkdir -p "$root/home" "$root/dolt" "$root/server" "$root/client" "$root/bin"
  export HOME=$root/home XDG_CONFIG_HOME=$root/home XDG_STATE_HOME=$root/home
  "$DOLT_BIN" config --global --add metrics.disabled true >/dev/null
  "$DOLT_BIN" config --global --add user.name manual >/dev/null
  "$DOLT_BIN" config --global --add user.email manual@example.invalid >/dev/null

  local dolt_port serve_port
  dolt_port=$(free_port)
  serve_port=$(free_port)
  (cd "$root/dolt" && exec nohup "$DOLT_BIN" sql-server --host 127.0.0.1 --port "$dolt_port" \
    >"$root/dolt.log" 2>&1) &
  echo $! >"$root/dolt.pid"
  local tries=50
  until (exec 3<>"/dev/tcp/127.0.0.1/$dolt_port") 2>/dev/null; do
    tries=$((tries - 1))
    [[ $tries -gt 0 ]] || die "Dolt did not start; see $root/dolt.log"
    sleep 0.2
  done

  (cd "$root/server" && git init -q && "$BD_BIN" init --server --external \
    --server-host 127.0.0.1 --server-port "$dolt_port" --server-user root \
    --prefix man --quiet --non-interactive) >"$root/init.log" 2>&1 \
    || die "bd init failed; see $root/init.log"
  (
    cd "$root/server"
    if [[ -n $token ]]; then
      (umask 077 && printf '%s\n' "$token" >"$root/token")
      export BEADS_SERVE_TOKEN_FILE=$root/token
    fi
    exec nohup "$BD_BIN" serve --addr "127.0.0.1:$serve_port" >"$root/serve.log" 2>&1
  ) &
  echo $! >"$root/serve.pid"
  local base="http://127.0.0.1:$serve_port"
  wait_for "$base/bdp"

  cat >"$root/bin/bdp" <<WRAP
#!/usr/bin/env bash
exec "$BDP_NODE" "$BDP_CLIENT/apps/bdp/dist/main.js" "\$@"
WRAP
  chmod +x "$root/bin/bdp"
  cat >"$root/env" <<ENV
# source this: the reference client, aimed at this stack. No real bd is on this PATH's front.
export PATH="$root/bin:\$PATH"
export BDP_SCOPE_URL="$base/"
export BDP_SPEC_REVISION="$(git -C "$BDP_CLIENT" rev-parse HEAD)"
export XDG_CONFIG_HOME="$root/client" XDG_STATE_HOME="$root/client"
export DO_NOT_TRACK=1 BD_DISABLE_METRICS=1 BD_DISABLE_EVENT_FLUSH=1 DOLT_DISABLE_EVENT_FLUSH=1
${token:+export BDP_AUTH_TOKEN="$token"}
ENV

  if $seed; then
    (
      # shellcheck disable=SC1091
      source "$root/env"
      export HOME=$root/home
      local_id() { bdp --output-mode json issue create "$@" --actor manual | python3 -c \
        'import json,sys;print(json.load(sys.stdin)["data"]["result"]["id"])'; }
      first=$(local_id "Parent task")
      second=$(local_id "Blocked task" -p 1 -t bug)
      third=$(local_id "Third task" -l demo)
      bdp issue dep add "$second" "$first" --actor manual >/dev/null
      bdp issue comment "$first" "A first comment" --actor manual >/dev/null
      echo "$first $second $third" >"$root/seed-ids"
    ) || die "seeding failed; the stack is still up at $root"
  fi

  cat <<DONE
Stack up.
  bd serve:  $base  (token: $([[ -n $token ]] && echo required || echo none))
  bd build:  $("$BD_BIN" version 2>/dev/null | head -1)
  client:    $(git -C "$BDP_CLIENT" rev-parse --short=8 HEAD)
  seeded:    $(cat "$root/seed-ids" 2>/dev/null || echo "nothing")

In your shell:
  source $root/env

Then, for example:
  bdp list
  bdp issue list
  bdp issue create "Fix login" --actor sam
  bdp call issue/create --arguments '{"title":"Fix login"}' --actor sam
  bdp issue delete --ids ID            # previews
  curl -s $base/bdp | python3 -m json.tool

Logs: $root/serve.log
Stop: $self stop $root
DONE
}

stop() {
  local root=${1:-}
  [[ -n $root && -f $root/serve.pid && -f $root/dolt.pid ]] || die "not a stack directory: $root"
  local pid
  # Only what names this directory, checked before each kill.
  while read -r pid; do
    [[ -n $pid ]] && kill "$pid" 2>/dev/null || true
  done < <(ps -axo pid=,args= | awk -v root="$root" 'index($0, root) && /db-proxy-child/ {print $1}')
  for pid in "$(cat "$root/serve.pid")" "$(cat "$root/dolt.pid")"; do
    if ps -o args= -p "$pid" 2>/dev/null | grep -qE '(bd serve|sql-server)'; then
      kill "$pid" 2>/dev/null || true
    fi
  done
  sleep 1
  chmod -R u+w "$root" && rm -rf "$root"
  echo "Stopped and removed $root"
}

case ${1:-} in
  start) shift && start "$@" ;;
  stop) shift && stop "$@" ;;
  *) sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//' && exit 2 ;;
esac
