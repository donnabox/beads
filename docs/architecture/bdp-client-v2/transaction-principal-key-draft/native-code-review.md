# Native principal-key draft code review

No actionable findings in the exact frozen two-file draft. This is a read-only code review, not execution proof or execution/publication release.

Reviewed binding SHA256 3c227b20d066ea97684ebdd0f2876734a537d61d53f1c9b3b2c4cdf263f46110, complete patch73432d0cc2cb5e41e8616d19645d1b43c308be712fdd4925d0946ffcd6964256, base18104a03b850553641f24fe83e7318493fa102ca/tree10810a8016a451e241ea974487681dcddd63c2e4, branch codex/vickie-principal-key-isolation. Test SHA e6af6263f164032ea4ae97167740412b9ad2a23af98e2a92729cd85dfa6d9b32; progress document SHA f0171eccfb02d85174abc51dca22f6b456b932fc2286d82319088bf497a5baf1. Actual dirty paths are precisely those two. Complete added-file patch content equals current test; tracked document patch equals Git's actual diff. Hashes were rechecked after review.

The accepted plan783f439e/ledgeraa556158 and root authoring release constrain this review. All seven literal request objects were independently parsed as data and compared exactly to the ledger; body hashes/bytes match, five150B POSTs plus two bodyless GETs. Source calls exchange1 through7 once each, registers one callback, caps attempts7 and uses exactly three allowlisted routes. Positive model7HTTP/one child and negative intended2/max7HTTP/one child remain9expected/14max/two children. These are source-derived counts, not actual observations.

At test399 only B's configured principal changes for the negative; distinct random tokens and all requests/oracles remain. Runtime configuration321–331 maps matched token to principal; HTTP333–343 passes it with the header key; store421–424 and1059 onward read by principal+key. Collapse therefore genuinely reaches conflict on create-b while ordinary A/B can retain separate receipts. exchange452–540 records and validates actual409 Problem/context before completed552's unchanged named expected200 assertion. External adjudication must still require the genuine named failure, exactly two complete responses/status200+409, pos-1/equal current context, and no other error; generic exit1/currentAssertion alone is insufficient.

Full successful receipt and five-field Resource baselines552–608 are schema-validated before use. Rotation/B replay require entire retained receipt equality; A conflict exact Problem; final collection exact a,b+next:null excludes c. Events use eventPage validation and exact source/subjects/type/ordinal/transaction/data with private IDs derived from the Event GET's epoch/view653–669 and schema-validated actual time. Store1235–1255 passes core.transaction to evaluator, batch692–709 records it, events83–101 renders the same private ID. No assumed A/B view inequality, principal-filtered reads or normative ID spelling is introduced.

Corrected cleanup713–730 no longer treats absent returned child as known stop. launchAttempted is set immediately before launch409. Never-attempted launch can remove its store; a returned child must have exitCode/signalCode; attempted pre-ready failure remains unknown-pre-ready, retains its directory and records failure. Existing launcher close is memoized and abort uses the same close. Startup rejection includes launcher bounded cleanup, but the test conservatively does not infer stop from no return. Work abort precedes close; cleanup failures accumulate with primary failure; removal requires known stop. External process/PID proof remains required for pre-ready failures.

Output734–780 is a fixed response-only projection: no whole exchange/request/config/roles or Authorization values. Credential checks reduce to booleans/counts. Binding is validated ready-derived digest/target data, not credentials or child-attested Git. The helper validates all12 operation targets and metadata before dispatch; sourceCommit remains caller attribution. Output/Vitest result paths must be outside removable store. Write errors accumulate; stdout includes updated failureCount. Receipt/Resource/Event bodies are compared in memory, not falsely claimed as retained raw reports.

Transport uses actual exact raw executor617–785, auth-only handles/resolver, fixed host/routes, omitted GET body buffer, completed wire/body digests and bounded response/header buffers.45s work/10s ready+cleanup/60s callback/5s exchange limits remain unchanged. Decoded-body limits are distinguished from framed/wire buffering and peak heap. POST cache/type/context expectations align with actual HTTP builders and POST Vary handling. Schema definitions/required formats were read as data; no Ajv/product module was executed.

The progress addition explicitly describes an unexecuted draft checkpoint and preserves earlier evidence attribution. Final ordinary sequence-main integration, actual static/build/package/consumer/entry/payload proof, both real cohorts/cleanup, actual Claude and exact-head CI remain separate root gates. No whole-profile/public-discovery/view-change/pending-key/crash/principal-read-filtering credit follows.

Supporting source whole-file hashes (relevant functions above):
- scripts/transaction-installed-observation.mjs — 6f772f3776d3c3effa788676d853cc2932a6ee59ad63ac3338ad1b487408914b
- apps/bdptest/test-support/transactional-execution-binding.ts — cde506c72a8710eafe58522900f6db70ed3701cf333560252dc8e7c5e706cf89
- packages/conformance/src/http-executor.ts — 770a8405ae1532484b5a2658619c067832644947888f95e0c6575002376f1b5d
- packages/conformance/test-support/profile-contract-plan.ts — 18c953443e6695cfa48f89677231fecffc349df02bd080f107a5c739db10b728
- packages/server/src/transaction-installed-configuration.ts — e15cb54266848fdd4aa08b19b37f1b2eb080f4599d9414fad8d85b6f33f35d99
- packages/server/src/transaction-runtime-configuration.ts — e594e930f845255cafc7e2f5b5b28149b7e95fe8a09052b187a492c96328f7ca
- packages/server/src/transaction-runtime.ts — 77954f0a85029d76443feac54cd53cda4339ffc126660ace5a9f5af92466c573
- packages/server/src/transaction-http.ts — b51d02801e5924190aba77c377e538db0b8619149d6184e19cb56d440f40c2cf
- packages/server/src/transaction-store.ts — fcd6882b29e538db45867c894c06dfbaa6ca588db0c2e403ef1b7abb14f6c706
- packages/server/src/transaction-batch.ts — 70d941d416979edd54e250d51318cfb7e5d4ddb315ea6131de7ca5f56e98a014
- packages/server/src/transaction-events.ts — c6052433662fcb2d0a202ad4e408a24d4c43f980313fa89659501b4475790db5
- schemas/bdp-v0.schema.json — 0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227

Review completed 2026-10-15T18:15:21.531696+00:00. No tests/builds/services/product imports/product edits/packages/install/reviewer launches. Preparation scaffold was not substituted for product-code review.
