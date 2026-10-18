# Formatter-only successor proposal — CLOSED

Observed trigger: this successor's format-check completed raw1/acceptedfalse, log SHA25609a4531ec83f8b2de1428405a8530c0eeca8bcc2f20f8f6e6a38ffd6b2ad6ff4. Typecheck passed on the exact frozen bytes. Lint was not run; do not predict it passes. This plan neither edits source nor runs a formatter. Wait for the active Claude review's terminal result/disposition and root's explicit authoring release.

## Exact correction supported by the log

Only three already-owned test files need formatting. Apply precisely the displayed formatting changes, preserving every expression, literal, identifier, assertion, callback name, request matrix, wait/cleanup/deadline and ordering:

- apps/bdptest/src/transactional-event-missing-context-installed.test.ts:581: expand cleanupTimer=setTimeout into the shown multiline callback and Math.max argument, including the formatter's trailing comma. The same expiresAt-Date.now calculation, timer callback, deadlineExceeded and rejection remain.
- packages/server/src/transaction-events.test.ts:1388: wrap the owner observation console.log/JSON.stringify over multiple lines.
- Same file:1468: wrap the etag/last-modified for-body assertion onto the following line.
- Same file:1477: wrap the successful-page validator assertion after its if condition.
- Same file:1479: wrap the transactional Problem validator assertion's toBe(true) argument.
- Same file:1551: wrap the multiple-failure AggregateError throw after the if condition.
- packages/server/src/transaction-service.test.ts:2640: collapse only the cancelled=f.observe(f.events(...)) call to the formatter's single line.
- Same file:2699–2764: expand the it.each invocation into multiline name/callback arguments, indent the existing callback body one level, and use the displayed trailing callback comma/closing parenthesis. All five parameter rows and their exact expanded names remain unchanged.

These are eight formatting regions in three files, not three semantic defects. The three production files must remain byte-identical. Historical installed prefix and old Event/service callbacks also remain byte-identical; the identified formatting regions are inside newly added callbacks. No general style sweep, cleanup rewrite, helper extraction, timeout change or test weakening. The exact pinned Biome may apply these three files only if root's later release explicitly allows it; otherwise apply the logged textual shape. Inspect the resulting diff for only these changes and stop on anything else.

## Freeze and validation rebind after release

Preserve all current author/review/static evidence, including failed format attempt, with original identity. Freeze a separately named successor: exact before/after hashes, full patch against b76, correction patch against this successor, all679 source hashes, unchanged production and historical prefixes, callback/workload identity. Preserve native/Claude findings and bind their disposition accurately; formatting does not relabel the old code review or typecheck as a new execution.

Prepare a fresh closed validation successor with the same13source/2installed selections,13source stores/15combined stores,91emitted+14installed HTTP and unchanged bounds. Explicitly reuse the successful existing bootstrap/tool record only after proving unchanged lock/dependency inputs/tool paths/bytes and mapping its old source identity to the new formatting-only source. Do not reinstall. Keep the failed run directory immutable; fresh attempt paths only, no retry within it. A root-released fresh typecheck/format/lint prefix can bind the corrected source before the unchanged fresh-build/source/package/consumer chain; no full suite or blanket unrelated rerun. Any additional lint or council finding needs its own bounded disposition rather than being silently included here. All runtime/publication gates remain closed.
