Add an installed-server restart test that saves Event cursors, cleanly closes the first process, reopens the same store, and verifies JSON resume, Last-Event-ID precedence for SSE, and rejection of a cursor from a different Event source. The existing Event test remains intact.

Validation: native Codex and Claude reviews dispositioned; typecheck, formatting, lint, clean build, historical/new development tests, strict fresh-package verification, and the external installed test all passed across 19 bounded stages. The external test observed five requests, two normal child exits, a drained stream and a removed store. These local runs used the exact committed file bytes on parent 2dfd with the test delta; new-head CI is separate. Existing lint warnings remain.

Targets only the working sequence/depth branch. This adds bounded evidence, not public write-profile admission or complete profile qualification.
