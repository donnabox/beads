# Publication source synchronization — October 18, 2026

Ordinary merge `4d7a0777d1d652914d35e271c66b98e660f6d496` joins prior publication evidence `5695f486a189fffd5619634a1aee05a60ea8da41` with reader `d684063bfe777cd327e98250c97879d617296201`. Reader source hashes match exactly. Completed full external review and native scoped test repair review are included in the reader evidence; fresh external delta review remains RUNNING. Publication implementation is unchanged. No format/serializer/CLI added.

Exact merged-source canonical focused no-engine selection passes68 test/subtest records (40reader +28publication), zero failures/skips, exit0. Run with BEADS_TEST_SHARED_SERVER=0 BEADS_TEST_ENV_RUN_DOLT=0 BEADS_TEST_PREBUILT_TEST_BINARY='' scripts/test.sh -json, explicit pure roots from prior source-sync-2026-10-17 manifest, -count=1 -timeout90s, graphstore and graphoutput only. Raw JSON and source/parent/hash receipt adjacent. All24 original import-only commits above the shared base, and the new import repair89ca1e312, are absent from export ancestry. No engines/producer/CLI workflow ran.

Managed hooks bypassed to preserve Janet's engine reservation and avoid ambient bd mutation. Full hooks/CI, selected-base adaptation, final reader delta review and exact-source installed embedded/shared-server qualification remain open. No release or merge readiness claim.
