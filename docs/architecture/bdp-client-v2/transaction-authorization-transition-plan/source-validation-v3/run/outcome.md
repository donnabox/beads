# Source-v3 actual partial outcome

Root ran six static stages successfully, then transport stopped with raw/supervisor exit1 in1.386s. Exact selected results: T1 failed, T2 passed; three unrelated cases skipped. Actual HTTP20 (T1 sixteen, T2 four), not planned21. Queued-frame, stale-lease and seven owner cases did not run.

T1 reached the real policy publication and old Snapshot/page/Read-continuation/Changes/Event foreign-view fences. Its16th request returned200, but actual mutationReceipt schema validation rejected `expiresAt` for the withheld branch (`#/oneOf/4/properties/expiresAt/false schema`). The later exact receipt equality and17th replay were not reached. The schema/test remain untouched; production author is diagnosing the projection. Retained response records contain body hashes, not reconstructed raw receipt bodies.

T2 delivered two real Event frames, observed server stream closure without pre-closure client abort and a new-view empty page; that callback passed. This does not make the complete source stage pass or establish full authorization support.

Both runtime close/store-removal reports succeeded. Independent readback finds all sampled PIDs absent and both reported stores absent; transport TMPDIR is empty. No timeout/forced cleanup; source unchanged at all completed stages. Exact hashes, safe observations and sampled PID inventory are in outcome.json. This audit is by the test author, separate from independent code council. All earlier failed/static attempts remain preserved; C1 is closed.
