# Parser refusal format successor plan

Status: PLAN ONLY; no source mutation or execution authorized by this packet. The actual Claude code review remains bound to the original frozen draft. Root will combine its completed findings with this plan before releasing any authoring.

The formatter reported exactly two test-file sites. `replacements.json` contains the exact old/new text, each requiring one occurrence; `prospective-format-only.patch` is an in-memory projection saved only here. In the installed test, move the existing explanatory comment before its outer loop. In the source test, move its existing explanatory comment before the outer loop and put the innermost loop/await on the formatter-requested single line. No formatter was run to prepare this plan.

## Preserved behavior and evidence

Production `packages/server/src/transaction-http.ts` remains byte-identical. Neither replacement changes executable tokens, loop nesting/order, query strings, controls, assertions, callback names, principal/credential handling, cleanup, or deadlines. New source workload stays 67 calls (37/13/17, including three actual setup GETs); new installed workload stays 25 HTTP calls (one setup and 24 refusals). The historical installed 42-call callback remains untouched. The two corrected parser-400 sentinels and all other prior cases remain untouched. The proposed patch has only the two formatter-reported sites.

The original successful typecheck and failed format check remain evidence for the original draft; neither is relabeled as a successor result. No source/import/test/build/service/Claude/formatter process was started by this preparation. The active review packet and its logs are untouched.

## Successor sequence, after the active review completes

1. Root reads the actual final Claude result, checks unchanged inputs/cleanup, and dispositions findings. Any substantive finding gets an explicit combined correction plan; these prospective hashes apply only if the exact formatting changes alone are authorized.
2. On root authoring release, verify original HEAD and all tracked input hashes, then apply only accepted replacements. Preserve original source packet, review, typecheck and format-failure artifacts. Freeze a new successor packet with actual source hashes, full base diff, correction delta, unchanged production hash, exact dirty scope, and current context/tool bindings. Do not silently overwrite original inputs.
3. Obtain native comparison of the actual correction and root disposition of whether the completed whole-draft Claude review plus the exact nonsemantic delta suffices or requires a focused successor review. Do not infer full council acceptance from formatting, a partial review, or this plan.
4. Prepare a separate closed validation successor bound to the actual corrected bytes and tools. Carry prior successful typecheck only as predecessor evidence; bind fresh static checks for the changed successor as root specifies. Retain the same callbacks, request counts, deadlines, observable supervisor, stop-on-failure/no-retry behavior, and owned cleanup. Root releases execution separately.
5. Retain the explicit fresh bound build prerequisite before HTTP-suite collection because it imports emitted `../dist/transaction-runtime.js`; source aliases alone are insufficient. Installed packaging/provenance and its 25-call observation remain separately bound and released. No integration/publication until exact-head validation and council requirements are met.

No execution gate is released by this plan.
