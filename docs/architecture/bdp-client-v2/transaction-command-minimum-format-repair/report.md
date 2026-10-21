# Command-minimum formatting-only candidate

Completed outside frozen W. No source, commit, push or gate change. Original draft diff `8f0123c9484a74a5c4995dea84833561056f603c8da64d5f9139107b85d44df8` and all seven source hashes remain exact. Only three copied test files were formatted, using exact copied W biome.json `62beced4e0227b97fe9ade1a1a573bf53a2f8916af8da7ce70f94315fece39dc` and installed Biome2.5.7 under Nodev24.16.0.

- Initial format-check failure retained verbatim as initial-format-check.log (root's source preparation attempt, not relabelled).
- One formatter invocation exit0 in0.072s; one parse-only equality invocation exit0 in0.177s. Each120s maximum, no retry/timeout/forced cleanup.
- Full input/output copies, exact argv/cwd/exits/logs in this directory. correction.patch SHA256 `5d5104f9ada2652d9dd821792d59027bc0ea374d25626db33eeed10e1cace888` contains only formatter output differences for the three authorized test paths; not applied to W.
- @babel/parser8.0.4 parses all three before/after pairs as TypeScript modules without importing/evaluating product modules. Normalized AST strings are exactly equal. Removed only source locations, attached comments and extra metadata (raw spelling/parenthesization); literal values/operators and TemplateElement raw/cooked contents remain. AST hashes/bytes and source hashes are in ast-equality.json and summary.json.
- Formatter/parser parent processes were reaped; independent PID checks confirm absence. No tests/builds/services/package/install execution. Frozen source remains under council.

This candidate is ready for root's bounded formatting application after the source-review hold is released. No claim that unrelated checks or runtime behavior passed follows from parse equality.
