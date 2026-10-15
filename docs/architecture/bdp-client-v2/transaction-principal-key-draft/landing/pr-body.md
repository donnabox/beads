Adds an installed HTTP regression test for Transactional idempotency keys: two authenticated principals can use the same key independently, while a rotated credential for the same principal replays the original receipt. Reusing that principal’s key for a different request returns a conflict. The test also checks complete final Resources and their linked Events against the retained receipts.

A separate control maps both credentials to one principal and must fail at the second creation with the expected conflict. This adds private test coverage and a progress note; runtime behavior and public profile eligibility are unchanged.

Validation: offline setup, typecheck, build, formatting, lint, boundaries, Read seal and build inventory passed on the reviewed successor. Fresh installed positive/control results, the focused Claude supplement and exact-head CI are pending; update this sentence from their actual outcomes before marking the PR ready.
