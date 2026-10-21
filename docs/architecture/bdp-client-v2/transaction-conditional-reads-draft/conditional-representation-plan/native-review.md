# Native review: conditional representation repair plan

The proposed private parsing context is viable, but the eight-file plan has one concrete caller omission. Add the private profile planner before authoring. The public Read parser/loader must remain strict; no catalog relabel, negative-probe workaround, or public admission change is needed.

Reviewed plan SHA256 `12fb5360ae283d58d35ea3a37175719fe1a5d9b3ba42d6a7046c9b4fc32f1aef`, CW HEAD `1ba65f2f192cc84e44eb09ef89b9693ec6cdaf05`, frozen four-file binding `b630e46d04a67edc5d59b511c8489a680d6275cd14ece3c65aa49a8b48299999`. This is a read-only plan review; no product/parser invocation, test, build, source edit or external review was run.

## Required correction: include the source-private profile planner

`packages/conformance/test-support/profile-contract-plan.ts:245–251` directly calls `parseExecutableScenarioManifest` on the same supplied cumulative manifest. The plan's callsite inventory omitted this test-support directory. Switching only `artifact-bundle.ts` to the new internal parser leaves the planner rejecting the Read row's two successful POSTs and bodies.

Concrete affected calls include `transactional-cumulative.test.ts:190–194` (“copies source bytes before binding…”), and its private-runner firewall test at310. The former explicitly expects all canonical manifest rows to be structurally ready. These cannot pass with the proposed factory-only context selection.

Add this ninth source path to the ownership inventory and pin it. The planner already asserts canonical selection identity, authentic bound configuration, matching profile/history and matching execution configuration before reading the manifest (lines224–244). Reuse the same narrow internal parsing context for the genuine baseline Transactional carrier, preserving:
- Current source copying, source-budget behavior and fixture-oracle validation.
- Exact requirement ID/profile and fixture checks at281–292.
- Unsupported-carrier/History behavior and `claimEligible:false`; planning is not execution.
- Strict public/development/Read+Update behavior. Do not select mutation capability merely because an untrusted catalog string says Transactional.

Add discriminating coverage in the already-owned cumulative test file: the real conditional manifest plans successfully only under the canonical baseline TX configuration; wrong selection/configuration/profile/History and public strict parsing still refuse or retain their existing unsupported-carrier result as applicable. No additional product route, HTTP request or installed child is required.

## Required implementation precision: keep negative-probe rules independent

The current Read method guard contains both mutation restrictions and rejection of `negativeMethodProbe:true` on GET/HEAD/OPTIONS (`executable-manifest.ts:644–654,840–850`). Replacing the entire condition with an effective Transactional profile would silently bypass that marker validation. Threading an execution context is appropriate, but keep marker validation based on the declared requirement/profile and marker semantics.

The plan already says not to reinterpret the marker; make that concrete in both requests/actions tests: ordinary marked Read methods still reject; the existing bodiless POST negative probe remains valid; a Read negative probe carrying an exact body retains its rejection (`executable-manifest.test.ts:1602–1623`). The newly permitted successful setup/update POSTs are unmarked. No new rule asserting a particular response from arbitrary unexecuted input is needed.

## Boundaries and remaining assessment

The remaining design is sound for this bounded repair:

- A module-internal export absent from package root/testing exports and package export map is an appropriate source-private entry. Keep the existing public parser/JSON loader signatures and behavior unchanged; share parsing logic rather than duplicate it. Require version2 and the Transactional carrier on the internal entry, with normal validation errors. Cover both v2 requests and actions; preserve all raw/header/accessor/encoding/capture rules.
- The factory's canonical selection WeakSet check, baseline profile/history test, exact ID→requiredProfile matching, copied digests/deep freeze and cumulative bundle brand remain decisive (`profile-contract-selection.ts:109`; `artifact-bundle.ts:127–143,223`). A parsed object alone grants no bundle or execution authority. Runner checks still reject branded bundles through the public runner and require matching private TX profile and exact Scope/configuration before dispatch (`runner.ts:348–355,2792–2820`).
- Selecting `transactional.receipt.pagination` by stable ID fixes the eight-case unrelated receipt-page fixture without weakening its actions or changing canonical parent bytes. The three binding socket cases use the first scenario only as session identity and need verification, not an inferred source rewrite.
- The two canonical JSON fixtures can stay byte-identical, with Read requirement identity, 14 actions, 198 assertions and partial qualification unchanged. Extra CPU/receipt-authority regression checks must be accounted separately from the 14+4 selected installed HTTP/four-child workload.
- Read provenance analysis matches the actual source: the historical generator records schema, schema-validator, runner and executor provenance; the current seal checks Read schema projection and exact current catalog/manifest/fixture inputs. This parser/bundle repair does not itself rewrite those artifacts or constitute a new current-payload Read observation. Preserve old attribution and run the existing seal after build; do not change seal constants or verifier logic to accommodate an unexpected failure.

No other concrete plan defect found in the inspected parser, bundle, runner, fixture and Read-seal boundaries. The corrected plan should be nine files, with the narrow planner and marker checks above, before successor source/code council and finite validation.
