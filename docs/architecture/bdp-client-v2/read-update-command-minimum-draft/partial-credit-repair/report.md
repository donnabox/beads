# Partial-credit correction

The command-minimum row now receives partial credit. Its transport/header/body-bound receiving variants remain unobserved, so `wholeContract` is false. The 11 actions, 47 bindings, real-response/control oracles, actual request bytes and parser/time limits are unchanged.

Four files changed from b89554c6004edf99d3278518bee1b150dce3fa98: coverage flag/reason; app exact coverage hash, counts/memberships and four data-size allowances; profile coverage expectations; cumulative whole-ID whitelist. Whole contracts 35→34, uncovered IDs 95→96; authored/required/omitted remain 56/130/74. Coverage grew 102 bytes to 720729; exact data allowances each grow 102, with no runtime cap growth. Other 129 coverage objects and all manifest bytes are preserved. See correction.patch, correction-measurements.json and execution-byte-bridge.json.

Authorized format check passed once in 0.705s; exact CPU selection passed 8, skipped 336, failed 0 in 3.144s. Both supervised exits are 0, source unchanged, 11 sampled PIDs independently absent, no temporary entries remain. No build, service, package, installed rerun, commit or publication occurred. Original installed results remain attributed to b895 and their former 35/95 assessment; only unchanged execution evidence is bridged, not relabeled as successor execution.

Author overlap: this worker authored the original plan/code and this authorized correction. Root owns independent acceptance. The original full, cumulative and fanout empty timeouts remain inconclusive; harness and literal completed reviews are separate. Remaining fanout/credit council is pending.
