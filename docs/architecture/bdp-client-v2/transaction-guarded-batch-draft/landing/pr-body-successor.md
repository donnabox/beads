Adds an installed-HTTP regression for two batches that update the same revision. One batch commits its marker and update; the losing batch rolls back its staged marker and retains a failed receipt that stays identical through replay and receipt GET. Full Resource and Event comparisons check the final state.

The guard-removal control executes both batches successfully and then fails only the expected single-loser assertion. Transport failures retain each original error once; replay checks bind the selected completed request to the immutable retry input.

Validation on `8833c924`: static/build checks, fresh packages and external consumer, seven-response positive and three-response control, and process/store cleanup verified. Native and actual Claude reviews are complete with all findings dispositioned. Exact-head CI is running; the earlier green CI belongs to `9e5f040`.

Scope is the installed test and progress documentation. No public Transactional qualification, release, or Preview2 promotion is claimed.
