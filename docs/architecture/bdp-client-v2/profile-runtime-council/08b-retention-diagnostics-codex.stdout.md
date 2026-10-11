# Native review: disposable retention diagnostics

Head `b7eefc9ecad02540f9c26576220841c78f1725cb`; base `486ad62d7c6ecb8e36dad4e596b1597fde425cdd`.

No actionable defect found in this bounded source review. Read all eleven changed-file diffs, the new22-action/67-assertion scenario in full compact JSON, each new catalog citation/coverage binding, new negative cleanup tests, and process helper clock/reopen paths. Exact supplied files/context are in08b-retention-diagnostics-index.json.

The fixture keeps the24h store/member floor, passes a private fixed clock to startup expiry and mutation retention, and reopens the same disposable store/Scope at day−1ms, day and day+1ms. It does not shorten retention or alter stored outcomes. Public configuration and exports do not expose the clock. The control validates one finite nondecreasing integer elapsedMs within seven days; ordinary process controls are refused in retention mode. Children are tracked for cleanup when reopen fails or the runner aborts.

Observations distinguish original retained postimage from newer live state before expiry, successful tombstone expiry and changed-request conflict, and permitted failed-delete forgetting followed by a real successful delete after the subject exists. Subsequent reads/reopen verify preserved state. A guard uses the pre-expiry live revision. The before-expiry retry followed by original-boundary expiry checks that retry does not renew retention.

Recounted combined diagnostic inventory:28 plans,302 actions,974 assertions.57 normative rows have bounded observations;24 remain omitted. Retention-minimum remains omitted because this endpoint advertises no client-known window. Only one creation success and one delete failure are observed; failure forgetting is permitted store behavior rather than a requirement for all providers. Live periodic expiry, other success categories, alias witnesses, expired creator bindings and restore are explicitly unqualified.

Smoke exclusion list includes the new private TX modules; no export/admission widening was identified. No full suite run by this council; parent's480-test/check/smoke/evidence result is separate. No source or evidence edits, no actual24h elapsed observation, and no normative/public-profile qualification claim.
