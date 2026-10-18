# Same-key report overflow: exact native repair review

No actionable finding in the narrow repair. Reviewed working source against HEAD 18f5428686dd84fb139c9dbbcbe39d5958012323, tree 9c78087e1036dd26ca908e1d8feaa91263513717, sole parent 98c600767e8e91a79e109cb741beb91e23b9bebe. Only apps/bdptest/src/transactional-same-key-overlap-installed.test.ts differs; the documentation and all other tracked paths are unchanged.

Exact test SHA256: b13688d6b02c5901e68251620ab6d7e00f9f675af4025319817f9e8cc6c9bca3
Original test SHA256: e6a52fa4d645ffef04e4b6e8ccd9bbcb7495624ae0e3f46536e0e9ac6b707974
Exact repair.patch SHA256: 18651f9db0489fb1ba60ce3f61433e3d7e6ca822d68c1ae45957f6fd0aa0b996

The entire difference is the braces around the second serialization's console fallback plus reportOverflow=true and addFailure(new Error("same-key report ceiling")). I independently reconstructed the successor from the preserved original using only those substitutions and obtained byte-for-byte equality.

The complete callback failure accumulator, catch, cleanup, observation builder, two serialization decisions, exclusive file writer, fallback and final throws were read. addFailure retains the new error object; after finally, one failure is thrown directly and multiple failures as AggregateError. Thus a second serialization overflow can no longer print a failed fallback with failureCount0 while the callback returns successfully. The fallback now truthfully reports at least one failure. Existing primary and cleanup failures remain retained; no exception is thrown from finally by the new lines.

The earlier on-disk observation can remain a prior successful snapshot if only the later serialization exceeds the ceiling. This repair makes the actual callback fail; it does not promise identical snapshots or rewrite an exclusive file. Acceptance must continue requiring the actual test exit/result and complete report, so the earlier file cannot rescue the failed invocation. If both serializations overflow, two distinct ceiling Error objects may be counted; this is a nonpassing reporting failure, never the sole intended negative oracle. No cap, semantic assertion, request, cleanup branch, payload or credentials change.

F3 is a source-established branch bookkeeping defect, not a demonstrated overflow in the current small fixture. No near-limit reproduction or test execution was performed for this review. The focused actual Claude supplement remains unlaunched and separate from this native conclusion. Other original findings remain under their existing source dispositions, not silently reopened or claimed fixed.

## Additional root-executed branch evidence
Root subsequently supplied check-overflow.mjs and overflow-check.json. I read them without executing: the script extracts the actual reporting suffix and injects a small first observation and oversized second observation, with no output file. Retained results show old source returning with zero failures despite failed fallback, repaired source setting overflow and throwing one ceiling error. This is a synthetic extracted-source branch exercise, not an installed run or proof that the ordinary fixture approaches the ceiling. It supports only the exact branch-accounting correction. Result SHA256 a66bf3cfafc99d1f5ed1a98fabcef864e121c9de34a946e4c1b69d059940c66c; script SHA256 d49dda9a63d61d52d7c8ef822a3c4b04185d3d10d68033af93ce0c33afc5ab9f.
