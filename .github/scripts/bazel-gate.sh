#!/usr/bin/env bash
# pr.yml's ci-gate view of its bazel.yml call (the `bazel` job). bazel.yml's
# rbe job decides the execution mode once and exports it as the call's
# rbe-mode / rbe-enabled outputs; this script reads those outputs and never
# re-derives the decision (repo variable, fork, secrets) itself.
#
#   bazel-gate.sh skips      the Bazel gate ids whose job is skipped by design
#                            in this mode, for CI_GATE_SKIPPED_OK:
#                              skip:   every lane and the aggregate (BAZEL)
#                              local:  the remote-only BAZEL_EMBEDDED,
#                                      BAZEL_PROXIED, BAZEL_SERVER_STORAGE,
#                                      BAZEL_INTEGRATION and BAZEL_CMD_DOLT
#                              cache:  BAZEL_EMBEDDED, BAZEL_PROXIED,
#                                      BAZEL_SERVER_STORAGE (integration
#                                      and cmd-dolt run with the read-only
#                                      cache)
#                              remote, fork-ro, fork-rw: none
#                            plus, in any of them, the Dolt-tier lanes
#                            (BAZEL_EMBEDDED, BAZEL_DOLTSERVER, BAZEL_PROXIED,
#                            BAZEL_SERVER_STORAGE, BAZEL_CMD_DOLT) when pr.yml's
#                            scope turned them off (SCOPE_DOLT_TIERS=off: a
#                            merge group) and the call reports it ran without
#                            them (BAZEL_DOLT_TIERS=off). Either half alone
#                            excuses nothing.
#   bazel-gate.sh aggregate  the value to gate on for BAZEL: the call's
#                            aggregate result, or, when the mode is missing
#                            or invalid (the rbe job failed), an unexpected
#                            value that ci-gate.sh rejects.
#   bazel-gate.sh remote     the value to gate on for BAZEL_REMOTE: success
#                            when the lanes executed on rbe-west (modes
#                            remote, fork-ro, fork-rw), else not-run-remotely.
#
# A call that never started (pr.yml's scope skipped it on a push to main,
# where bazel.yml's own push run is every lane, or the scope job failed)
# reports no mode and no result: its aggregate and remote check read skipped,
# which only the scope's own skips (ci-scope.sh) may excuse, and it excuses
# nothing here. Any other skip, a failure, a cancellation, or a lane that
# should run but reported nothing (pr.yml maps an empty output to skipped)
# fails the gate.
#
# Inputs (environment): BAZEL_RBE_MODE, BAZEL_RBE_ENABLED, BAZEL_DOLT_TIERS
# (the call's rbe-mode / rbe-enabled / dolt-tiers outputs), BAZEL_CALL
# (needs.bazel.result), SCOPE_DOLT_TIERS (the scope job's dolt-tiers output).

set -euo pipefail

mode="${BAZEL_RBE_MODE:-}"
enabled="${BAZEL_RBE_ENABLED:-}"
valid=false
case "$mode/$enabled" in
    remote/true | fork-ro/true | fork-rw/true | local/false | cache/false | skip/false) valid=true ;;
esac
# A selection fact supplied only by this workflow's repository-exact decision.
# It is never success evidence. Unrelated local runs retain upstream policy.
local_full=false
case "${BAZEL_LOCAL_FULL_INTENT:-false}" in
    false) ;;
    true)
        if [[ "${GITHUB_REPOSITORY:-}" != donnabox/beads || "$mode/$enabled" != local/false ]]; then
            echo "invalid full-local execution identity" >&2
            exit 1
        fi
        local_full=true
        ;;
    *) echo "invalid full-local selection fact" >&2; exit 1 ;;
esac
not_started=false
if [[ "${BAZEL_CALL:-}" == skipped && -z "$mode" && -z "$enabled" ]]; then
    not_started=true
fi

case "${1:-}" in
    skips)
        skips=()
        add() {
            local id
            for id in "$@"; do
                [[ " ${skips[*]-} " == *" $id "* ]] || skips+=("$id")
            done
        }
        if [[ "$local_full" == true ]]; then
            echo ""
            exit 0
        fi
        if [[ "$valid" == true ]]; then
            case "$mode" in
                skip) add BAZEL BAZEL_TEST BAZEL_PURE BAZEL_RELEASE_CROSS BAZEL_EMBEDDED BAZEL_INTEGRATION BAZEL_DOLTSERVER BAZEL_PROXIED BAZEL_SERVER_STORAGE BAZEL_CMD_DOLT ;;
                local) add BAZEL_EMBEDDED BAZEL_INTEGRATION BAZEL_PROXIED BAZEL_SERVER_STORAGE BAZEL_CMD_DOLT ;;
                cache) add BAZEL_EMBEDDED BAZEL_PROXIED BAZEL_SERVER_STORAGE ;;
            esac
            if [[ "${SCOPE_DOLT_TIERS:-}" == off && "${BAZEL_DOLT_TIERS:-}" == off ]]; then
                add BAZEL_EMBEDDED BAZEL_DOLTSERVER BAZEL_PROXIED BAZEL_SERVER_STORAGE BAZEL_CMD_DOLT
            fi
        fi
        echo "${skips[*]-}"
        ;;
    aggregate)
        if [[ "$not_started" == true ]]; then
            echo skipped
        elif [[ "$valid" != true ]]; then
            echo "invalid-rbe-mode:$(printf '%s/%s' "$mode" "$enabled" | tr -cd 'a-z/')"
        else
            echo "${BAZEL_CALL:-}"
        fi
        ;;
    remote)
        if [[ "$not_started" == true ]]; then
            echo skipped
        elif [[ "$valid" == true && "$enabled" == true ]]; then
            echo success
        else
            echo not-run-remotely
        fi
        ;;
    local-full)
        [[ "$local_full" == true && "${BAZEL_DOLT_TIERS:-}" == on ]] || {
            echo "complete local coverage requires exact fork/local identity and all Dolt tiers" >&2; exit 1;
        }
        # Read actual recorded dedicated results, never advisory masked needs.
        for id in BAZEL_TEST BAZEL_PURE BAZEL_RELEASE_CROSS BAZEL_EMBEDDED BAZEL_INTEGRATION BAZEL_DOLTSERVER BAZEL_PROXIED BAZEL_SERVER_STORAGE BAZEL_CMD_DOLT_DIRECT; do
            [[ "${!id:-}" == success ]] || { printf 'full local coverage failed: %s=%s\n' "$id" "${!id:-missing}" >&2; exit 1; }
        done
        [[ -z "${BAZEL_DOLT_RACE:-}" && -z "${BAZEL_DOLT_INT:-}" ]] || {
            echo "unexpected remote union result in full-local execution" >&2; exit 1;
        }
        echo "complete dedicated local Linux coverage passed"
        ;;
    dispatch-cmd)
        case "$mode/$enabled" in
            remote/true) cmd_result="${BAZEL_CMD_DOLT_COMBINED:-}" ;;
            fork-ro/true | fork-rw/true | cache/false) cmd_result="${BAZEL_CMD_DOLT_DIRECT:-}" ;;
            local/false)
                [[ "$local_full" == true && "${BAZEL_LOCAL_FULL:-}" == success ]] || {
                    echo "dispatch local coverage requires complete actual local gate" >&2; exit 1;
                }
                cmd_result="${BAZEL_CMD_DOLT_DIRECT:-}"
                ;;
            *) echo "dispatch command coverage requires a qualified execution mode" >&2; exit 1 ;;
        esac
        if [[ "${BAZEL_INTEGRATION:-}" != success || "$cmd_result" != success ]]; then
            printf 'dispatch coverage failed: integration=%s cmd-dolt=%s mode=%s\n' "${BAZEL_INTEGRATION:-missing}" "${cmd_result:-missing}" "$mode" >&2
            exit 1
        fi
        echo "dispatch integration and cmd/bd coverage passed ($mode)"
        ;;
    *)
        echo "usage: $0 skips|aggregate|remote|local-full|dispatch-cmd" >&2
        exit 2
        ;;
esac
