# Static successor 3 outcome — stopped on lint failure

- PASS: typecheck, raw exit 0. The actual TS18048 diagnostic is resolved in this source; the full configured tsconfig.check.json check completed successfully.
- PASS: format-check, raw exit 0. One file checked; no fixes applied.
- FAIL: lint, raw exit 1. Three errors: noAssignInExpressions at source lines 432 and 444 (return expressions containing memoized closing assignments), and useIterableCallbackReturn at line 861 (concise forEach callback returning sameFrame). The log also reports 15 noNonNullAssertion warnings, including preserved historical code. No fixes were applied.

Each stage ran once in order under the pinned supervisor and remote session guard. All report unchanged source, no forced cleanup, empty errors/survivor/runtime-directory lists. Source remains 0baf666fe442009a994bb4d7cd324890b4a5081b9c66bfda48c41c1646e55ff1 at HEAD 2dfd2c6cc84a28b1dd336f4184c1a5f910ce8a96, with only the owned test dirty. The entire failed successor-2 packet still matches its predecessor pins.

Stopped after lint failure. No retries, source edits, build or runtime stages occurred. These results do not establish runtime correctness or profile delivery. Root owns any next correction and fresh release.

Release bytes remained unchanged. Both static authorization fields are explicitly true; the unused disposition ordinaryIntegrationAccepted field is false while the gate records true. That metadata difference did not deny static execution in the accepted gate implementation, and is preserved rather than silently rewritten.

Exact results and log/exit hashes:

```json
{
  "typecheck": {
    "exit": 0,
    "accepted": true,
    "elapsedSeconds": 4.337945917155594,
    "logSha256": "91d54ffbee73a3ac72294d090e58d77745f8b93edbf94b862807698273a6505a",
    "exitRecordSha256": "ddbde31a000c2ba6a62d40396fbc9721bb8c5851c302907498d048d94afd4fce",
    "sampledPids": [
      46778,
      46895
    ]
  },
  "format-check": {
    "exit": 0,
    "accepted": true,
    "elapsedSeconds": 4.044380665989593,
    "logSha256": "fd06ed1906864d862af1b302e14f95710bd4ba5d86680bb0976e34e5bf20f03a",
    "exitRecordSha256": "138616128b7ad0c7432a60f4dedf687a56108fb003d15ec381d21c933b1dcffa",
    "sampledPids": [
      48450,
      48452
    ]
  },
  "lint": {
    "exit": 1,
    "accepted": false,
    "elapsedSeconds": 3.7038438748568296,
    "logSha256": "9b0a773c17474a0f692a778730f82849072b73363bde72ebba38939a00adf19f",
    "exitRecordSha256": "873a1c145cc6fbdecbd063bd801e40ee7c978250c89caeeef68203c32d0089cd",
    "sampledPids": [
      50051
    ]
  }
}
```
