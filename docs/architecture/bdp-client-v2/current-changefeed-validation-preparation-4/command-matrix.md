# CLOSED concrete command matrix

Every invocation below is prospective and currently blocked by the closed gate. Use argv arrays in commands.json; no shell interpolation or chained commands.

1. node-version

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final node-version

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "--version"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

2. pnpm-version

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final pnpm-version

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "--version"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

3. workspace-install

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final workspace-install

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "install", "--offline", "--frozen-lockfile"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

4. bind-workspace-tools

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final bind-workspace-tools

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/audit.py", "tools"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

5. typecheck

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final typecheck

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "run", "typecheck"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

6. format-check

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final format-check

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum/node_modules/@biomejs/biome/bin/biome", "format", "--config-path=/Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum/biome.json", "packages/server/src/transaction-http.ts", "packages/server/src/transaction-service.ts", "packages/server/src/transaction-store.ts", "packages/server/src/transaction-changefeed.test.ts", "packages/server/src/transaction-sse.test.ts", "apps/bdptest/src/transactional-changefeed-refusal-context-installed.test.ts"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

7. lint

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final lint

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum/node_modules/@biomejs/biome/bin/biome", "lint", "--config-path=/Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum/biome.json", "packages/server/src/transaction-http.ts", "packages/server/src/transaction-service.ts", "packages/server/src/transaction-store.ts", "packages/server/src/transaction-changefeed.test.ts", "packages/server/src/transaction-sse.test.ts", "apps/bdptest/src/transactional-changefeed-refusal-context-installed.test.ts"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

8. fresh-build

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final fresh-build

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "run", "build"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

9. bind-fresh-build

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final bind-fresh-build

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/audit.py", "build"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

10. focused-source

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final focused-source

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum/node_modules/vitest/vitest.mjs", "run", "packages/server/src/transaction-changefeed.test.ts", "packages/server/src/transaction-sse.test.ts", "--maxWorkers=1", "--reporter=verbose", "--reporter=json", "--outputFile.json=/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run/focused-source.results.json", "--testNamePattern", "^(?:preserves\\ changefeed\\ cursor\\ precedence\\ under\\ current\\-admission\\ opaque\\ minima|captures\\ current\\ changefeed\\ admission\\ once\\ across\\ owner\\ entries:\\ start\\-now\\-unknown|captures\\ current\\ changefeed\\ admission\\ once\\ across\\ owner\\ entries:\\ genesis\\-unknown|captures\\ current\\ changefeed\\ admission\\ once\\ across\\ owner\\ entries:\\ retained\\-unknown|captures\\ current\\ changefeed\\ admission\\ once\\ across\\ owner\\ entries:\\ retained\\-pos\\-01|drains\\ current\\ changefeed\\ admission\\ before\\ owner\\ close:\\ start\\-now|drains\\ current\\ changefeed\\ admission\\ before\\ owner\\ close:\\ retained|preserves\\ current\\ changefeed\\ anchor\\ capture\\ before\\ waiter\\ preflight|serves\\ current\\ changefeed\\ opaque\\ minima\\ across\\ finite\\ and\\ SSE\\ admission|keeps\\ start=now's\\ original\\ admission\\ anchor\\ through\\ a\\ minimum\\ wait\\ and\\ drains\\ on\\ close|keeps\\ direct\\ and\\ excluded\\ changefeed\\ minimum\\ paths\\ strict|serves\\ finite\\ start\\-now\\ opaque\\ minima\\ before\\ HTTP\\ shortcuts|decides\\ actual\\ HTTP\\ minimum\\ refusal\\ before\\ conditional\\ shortcuts\\ with\\ final\\ context|disconnect\\ cancels\\ the\\ actual\\ finite\\ minimum\\ wait\\ without\\ consuming\\ its\\ remaining\\ budget|effective\\ reconnect\\ cursor\\ owns\\ epoch/view/retention\\ decisions;\\ overridden\\ query\\ is\\ not\\ looked\\ up)$"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

11. audit-focused-source

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final audit-focused-source

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/audit.py", "source"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

12. prepare-package-directory

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final prepare-package-directory

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/packages.py", "directory"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

13. pack-protocol

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final pack-protocol

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "--filter", "@bdp/protocol", "pack", "--pack-destination", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run/packages"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

14. pack-server

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final pack-server

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "--filter", "@bdp/server", "pack", "--pack-destination", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run/packages"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

15. verify-archives

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final verify-archives

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/packages.py", "tar"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

16. prepare-consumer

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final prepare-consumer

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/packages.py", "consumer"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

17. consumer-install

   Cwd: /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run/consumer

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final consumer-install

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "install", "--offline", "--ignore-scripts", "--no-frozen-lockfile", "--prod"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

18. verify-installed-consumer

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final verify-installed-consumer

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/packages.py", "installed"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

19. historical-context-external

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final historical-context-external

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum/node_modules/vitest/vitest.mjs", "run", "apps/bdptest/src/transactional-changefeed-refusal-context-installed.test.ts", "--maxWorkers=1", "--reporter=verbose", "--reporter=json", "--outputFile.json=/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run/historical-context-external.results.json", "--testNamePattern", "^observes\\ genuine\\ installed\\ current\\ context\\ on\\ changefeed\\ parameter\\ refusals$"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

20. audit-historical-context-external

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final audit-historical-context-external

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/audit.py", "history"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

21. finite-start-now-external

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final finite-start-now-external

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum/node_modules/vitest/vitest.mjs", "run", "apps/bdptest/src/transactional-changefeed-refusal-context-installed.test.ts", "--maxWorkers=1", "--reporter=verbose", "--reporter=json", "--outputFile.json=/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run/finite-start-now-external.results.json", "--testNamePattern", "^observes\\ genuine\\ installed\\ finite\\ start\\-now\\ opaque\\ minimum\\ refusal$"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

22. audit-finite-start-now-external

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final audit-finite-start-now-external

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/audit.py", "finite"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

23. current-changefeed-external

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final current-changefeed-external

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum/node_modules/vitest/vitest.mjs", "run", "apps/bdptest/src/transactional-changefeed-refusal-context-installed.test.ts", "--maxWorkers=1", "--reporter=verbose", "--reporter=json", "--outputFile.json=/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run/current-changefeed-external.results.json", "--testNamePattern", "^observes\\ installed\\ current\\ changefeed\\ minimum\\ admission\\ and\\ retained\\ replay$"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

24. audit-current-changefeed-external

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final audit-current-changefeed-external

   Child argv: ["/opt/homebrew/opt/python@3.14/bin/python3.14", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/audit.py", "current"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

25. required-check-boundaries

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final required-check-boundaries

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "run", "check:boundaries"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

26. required-e2e-ready-preflight

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final required-e2e-ready-preflight

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "run", "e2e:ready:preflight"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

27. required-smoke

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final required-smoke

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "run", "smoke"]

   Outer120s, TERM115s, no retry; raw accepted exits [0].

28. required-evidence-verify

   Cwd: /Users/dbox/repos/wt/vickie/vickie-current-changefeed-minimum

   Supervised invocation:

       /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/current-changefeed-validation-preparation-4/run.py final required-evidence-verify

   Child argv: ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs", "run", "evidence:verify"]

   Outer120s, TERM115s, no retry; raw accepted exits [0, 2].

Root barriers: before first stage require accepted source/CODE/helper/tool binding; after workspace binding pin actual tool record; after build pin actual build; after archives pin actual archive provenance; after consumer install pin actual lock/manifest/settings; after installed verification pin actual proof/inventory; then release three sequential callback/audit pairs. Root final local acceptance and exact-head CI remain separate unresolved actions.
