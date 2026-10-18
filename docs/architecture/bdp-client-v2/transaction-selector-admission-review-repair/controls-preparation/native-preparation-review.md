# Independent emitted-control preparation review

No actionable finding in the exact preparation below. The two controls are suitable for root's separately authorized bounded execution. This report does not authorize or claim execution; the release template is not a root release.

## Derivatives and import binding

Independently compared all 1,104 files in positive-files.json against the current positive workspace and both copies. Every positive file still matches. Exactly one emitted JavaScript file differs in each derivative, and every copied source/test/configuration file remains identical:

- Grammar control: transaction-submission.js changes only the typed Selector grammar refusal to `{ kind: "development-limit", message: error.message }`. Before `9c413ba5154993ad8cbadfe73c33dd26bc916d9503a718dc379b173da809269e`; after `3dd9a9eb66f2c6babc1232eaaf81b68920c1f3dc732b0bdae285322febb3ed11`; emitted patch `3ab5d7b13d1d11114e14268aebbae0e54db3d85451410f02586e66dcf29a613c`.
- Capacity control: transaction-store.js removes only the fresh Selector capacity loop, preserving duplicate/conflict, operation-count and negotiation ordering around it. Before `4bdeb7432d7d19798b21939e8885d933b8e5d1e7cc40b4f1186dff1c5e611916`; after `175ad260ef5182a732699f51a4fcc8ebfb3517e88ccf194435d4997f7cf4034f`; emitted patch `eab498aa2ffdaa60f720d7ec9642b354840ec0d1299ea22a89245c03042dd102`.

I independently traversed literal relative imports from each copied dist/transaction-runtime.js: 48 modules and 119 edges, matching each saved module binding exactly, including hashes. Both mutants are reachable. The unchanged test dynamically imports `new URL("../dist/transaction-runtime.js", import.meta.url).href`, so it loads its copied emitted runtime. The source-based Vitest aliases are rooted in the copied configuration; they do not redirect these relative emitted imports to the positive owner. The copied server's @bdp/protocol symlink resolves to its copied protocol package. Other nonrelative server imports are protocol, fast-uri and Node builtins, not another server owner. Third-party dependencies are shared pinned dependencies, not newly installed products. Original source maps are retained; mapped source stack locations alone must not be mistaken for proof of mutation loading.

## Named oracles, selection and workload

The grammar control selects only `serves Selector syntax and prior-key precedence over capacity`. Its first sequential POST is invalid `$..*` to deleteWhere; the intended mutant produces empty 500 with `x-bdp-development-limit: unsupported-transaction-composition`, versus the normal 400 assertion. The capacity control selects only `serves exact fixed Selector boundaries and leaves refused keys unused`; its first POST is the 4097-byte valid deleteWhere Selector, expected 413 normally. Removing admission capacity lets execution reach the existing evaluator cap, whose error is sanitized to empty 500 without that development-limit marker.

Both immutable test callbacks report actual status, marker and content-length in the existing `Selector receiving status` assertion message. The supervisor requires precisely the selected failed title, no passing title, actual status500 with the correct marker or null and length0, and expected500-to-be400/413 assertion text. A mere nonzero exit, module load failure, or unrelated assertion cannot qualify. Both control paths are therefore discriminating proposals; the actual responses still must be observed.

Each intended failure is at the first request: expected2 HTTP total is an inference from sequential source and named failure, not an instrumented counter. Whole selected-callback maxima remain16+36=52 HTTP, two owners/listeners/directories, zero expected inventory handles and two maximum readonly inventory handles. No installed child, build, pack, new test, hook or product API is added. Original fixture afterEach closes runtime before removing its directory.

## Supervisor and gates

Read complete run-controls.py and its hash-pinned run-positive.py helper, without importing either. Helper import only defines the functions/constants; its main remains guarded. Root release must match the exact supervisor/commands/positive-manifest hashes. Before and after each attempt it verifies the twelve positive source bindings, the full positive copied-file manifest, both complete derivative payloads, patch/binding hashes and local protocol resolution. Commands are pinned, run from their distinct copied workspaces, maxWorkers1, with120-second supervision. Existing run output forbids retry or overwrite. It runs the second control only if the first meets the intended oracle and has no recorded error.

The supervisor starts a separate process group, records progress/process/temp snapshots, requires exit1, stops and rejects any lingering group, retains unexplained fixture/temp residuals, and refuses to accept missing runner JSON or a mismatched oracle. Counters are explicitly labeled inferred/static. Full logs and JSON remain available for root's actual outcome audit; this preparation review does not certify future runtime cleanup or exclude future unrelated runner errors.

## Binding and limitations

- commands.json `cda1c6825cbfd5a7bf14a2ec3f8df92d8197c659e67ec2b2bf94f8e5b8468c7d`.
- run-controls.py `93b13d77dfb2f5fecd1dc618c0de558808ad931194ca2cc499cb47e2a4b7dde5`.
- positive-files.json `0f681290210f871871dcc62c0834377665110ad96f39665d27919bde27d9c523`.
- positive helper `1036f3dfccfebfdd896b47f21054f96cccf2d941a1b20dc730011a57fccf39c6`.
- Grammar module binding `8639e4603d62230fc2354196db28af22f7f9e7054f574c6946d0fee73b78ac8e`.
- Capacity module binding `7689370637ebcdf992846af33bd000a880d62bebdd551aff1607e9d27a70e859`.
- Positive source binding `03a861851a195d5da505f311291bee6d3cd98f63866d4b6df756e2c3ae289afd`.

The preserved preparation-first-failure.txt describes a Python syntax failure before evaluation; it is not a failed product/control run. Reviewer performed only read-only inspection/hash comparison and this guarded report write: no product imports, supervisor invocation, tests, services, installs, builds or source edits. Installed25 and full-profile qualification remain outside this control gate.
