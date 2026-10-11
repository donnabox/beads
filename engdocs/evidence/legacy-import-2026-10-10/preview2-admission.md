# Preview 2 import admission preparation

This is an isolated compatibility candidate, not release staging or combined
engine qualification. Donna requested admission only if import does not disrupt
Janet's core work, otherwise Preview 3. Janet's read-only assessment agrees that
the shared edits match the current preview contract; final admission depends on
combined proof after the core source is stable.

## Source isolation

The import implementation is in a new strict parser, store adapter and command
route. No database schema, BDP contract, production workspace or in-place migration
changes are introduced. Admission/capability edits are needed to expose import.
The budget helper keeps all existing commands at 30 seconds and gives import five
minutes. Generic cancellation handling preserves outcome_unknown precedence;
four focused cases cover cancellation/deadline with and without uncertain commit.
The dependency guard refuses cross-prefix/external blocking edges before a native
writer that cannot represent them; it turns a raw SQL failure into validation.

The compatibility merge preserves both parents: independent PR86 a1411241c and
Janet's PR85 b7439a14d3, which already includes PR84 f7e9d7ae9c. The only source
conflicts were the command allowlist and its refusal/help text. Resolution retains
assign/priority/note and adds import. The import installed workflow now explicitly
requests show --format graph-json when it asserts graph records, preserving
Janet's ordinary Issue show --json output change. No other imported behavior is
changed. This branch does not replace either independent PR or mutate staging.

## Qualification completed here

The product source/tree and private canonical binary are pinned in
[preview2-admission.json](preview2-admission.json). Cmd/bd, graphstore and parser
[compile-only checks](preview2-compat-compile.log) passed. The selected parser/help/error tests passed 86
[cases/subtests](preview2-compat-pure.jsonl) with zero failures/skips; this is pure compatibility evidence,
not two-engine admission proof. [make check-docs](preview2-compat-docs.log) passed and the native precommit
lint reported zero issues. No database or server was provisioned by this probe.
The existing 95-case installed two-engine import qualification remains pinned to
its original independent product source and cannot substitute for joined proof.

## Core-first admission route

1. Janet qualifies/stages PR84, then PR85, then the comment slice if retained in
   her core. Iris does not restart, modify or overlap her active runner.
2. Once Janet selects the stable qualified pin, Iris prepares a distinct import
   successor using these mechanical adjustments, preserving all core work. The
   compatibility branch is a merge recipe, not an already qualified release pin.
3. Require green exact-source CI plus installed embedded/shared-server proof on
   that combined source, including import dry-run/apply, retained/current reads,
   ordinary and graph Issue JSON, edits and subsequent creates/dependencies.
   Preserve rollback/repeat/refusal tests, tree/binary provenance and independent
   review. Janet owns qualification scheduling and release integration.
4. If actual overlap, failed qualification or scheduling threatens the core lane
   or misses its admission window, defer import to Preview 3. Core work must not
   wait for this optional feature. Donna owns the final cutline.

Export, graph-format import, full legacy History reconstruction and general
existing-project conversion are outside this stage. The bounded supported subset
and explicit refusals remain in the user guide.
