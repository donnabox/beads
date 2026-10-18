# Frozen failed installed import qualification — attempt 01

- Exact clean source: `2b1eeaec358713da0d1e1a65b3e8d8cfc891f8de` / tree `69606e3278a5d2b5b53ebfd37d079a6cdce66e65`.
- Janet sole serial runner exit `1`; no overlapping Dolt/test process. Isolated Dolt 2.1.8 servers on ports 64579 and 64769 exited and both ports closed.
- Phase 1 installed workflows: exit 0, zero FAIL/SKIP, both named workflow families observed in embedded and ordinary shared-server subtests; exact-source prebuilt binary retained and hashed in `installed-workflows/result.json`.
- Phase 2 lower graphstore `^TestLegacyImport`: exit 1, seven reported `--- FAIL:` lines, zero SKIP; all 17 top-level import tests were entered. Four leaf failures are `TestLegacyImportAtomicityAndPersistence/{embedded,server}/missing-prefix/dry={false,true}`; parent failures account for the other three lines. The actual error is `ErrInvalidStore` wrapping a stringified `issueops.ReadConfigPrefix` missing-prefix error, while the test also requires `errors.Is(err, sql.ErrNoRows)`. `checkBinding` returns `fmt.Errorf("%w: Issue configuration: %v", ErrInvalidStore, err)` before import-specific `legacyTargetInspectionError` can wrap the `sql.ErrNoRows` cause. This is a source/test error-contract mismatch, not an engine outage. Do not alter this frozen attempt.
- The lower parser and mapper-control phases were not run after this mandatory failure. The runner also expected a prebuilt `bd` binary in every phase; none was produced by the graphstore-only package. That instrumentation assumption is secondary and must be corrected before a repaired-source rerun.
- No source repair, new PASS, PR retarget, protected ref move or **Memory V1** ship advance is inferred. Require Iris source reconciliation, fresh review, then Janet-owned exact-source installed rerun.

## Raw evidence

| File | SHA256 |
| --- | --- |
| `run.py` | `7074d0b7b729972b66edd52546aa772537bb5aea5120c2eb64e39f2c018b8a6b` |
| `run.log` | `1a55fb9bc324bd46b1ee64d5ac73b5384c7fed2ad8927ff5c36641ff5be06d8d` |
| `run.exit` | `4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865` |
| `installed-workflows/result.json` | `90ccf49a7a046a7c7c311e1e25cc39cdb09fc82b898553a3043ad0cd45247bc1` |
| `installed-workflows/test.log` | `fa7fbd2a64f238f8d68ffe3d8fe2ca3c42e696947f6fac07863571e7d1521f85` |
| `installed-workflows/server.log` | `ec817ced5bdc9a0f7771f7ede7300ce03db2a22b5f85f5e5cf9940b0419e92f2` |
| `lower-graphstore/result.json` | `47eb2f72b8966e3df2df84806726e7d7451fd5c7cc947d44d1b3b4f216cb45c1` |
| `lower-graphstore/test.log` | `125b6fbc8903fef93ac7a2408ba18bd46117beece248d4663f977bbfa8290270` |
| `lower-graphstore/server.log` | `60f103b29dee4826ec997d272a67a232c00b396208eb1ce34db2681364cb1180` |
