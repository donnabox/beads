# Native exact-diff CI headroom closure

No actionable findings. Frozen clean commit `40806dd3ac3e74ff94c07330c89ee47dc730080c`, tree `e870145ffa8f49821ea3a1fb59c1c06506b8bf17`, ordinary parent `427cfef6decbb422b393024e36218affc26007fa` exactly implements the externally reviewed and root-accepted one-literal proposal.

Verified `.github/workflows/ci.yml:19` changes only `jobs.verify.timeout-minutes: 15` to `20`. A direct byte substitution reconstructs the complete successor workflow exactly. File mode remains100644. Independently compared all663 recursive tracked tree entries: identical path inventory; all662 other object/type/mode records unchanged.

Parsed predecessor and successor with installed Ruby/Psych (`ruby -ryaml`), without installing dependencies. Both values are Integer scalars15/20. After replacing only the parsed predecessor timeout with20, the complete parsed structures are equal. There remains exactly one job key, verify, with21 steps. Workflow/job/check identity, concurrency, triggers, permissions, runner choice, commands, pinned actions, uncached Go build, required final gates and all inner test/exchange/cohort deadlines and coverage are unchanged.

Read the full external v2 Claude review and root authoring acceptance. The actual patch matches its explicitly accepted literal/scope; separate factual plan corrections do not require more source changes. This closure does not endorse superseded231s arithmetic, explain the historical W1b failure, claim the cancelled run completed, or guarantee20minutes suffices. Root acceptance uses corrected249s setup and921s test allowance with30s reserve; these remain estimates.

Ready for publication and the one root-authorized fresh exact-head CI attempt. Merge still requires successful completion of every existing suite and final gate. Any cancellation must be diagnosed from its cause/annotations before further action; this is not an automatic retry or further-timeout-increase authorization.

Verified SHA-256:

- Predecessor workflow: `b221328fb9b809ab4d64ffeb25448d39cc40b403ffa5ed7900b9fe4d5f64a65e`
- Successor workflow: `fd0a08da61f74abe19c8db16d879bfff85d8090643bf4da326d3e597ae129be3`
- Exact commit diff: `c2eafc74c9605c28adb047555a678703616d2a02c482f35c8ec9930d92446f64`

Read-only Git/tree/hash/YAML inspection only. No tests, builds, installs, services, network or source changes; only this external report was written.
