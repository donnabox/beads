# Write evidence integration validation

Validated frozen clean source `30351f5249ae41575bf62ed4c924a6a76fe68c4c` on `codex/vickie-write-evidence-integration` in `/Users/dbox/repos/wt/vickie/vickie-write-evidence-integration`. No source changes or outward writes. Remote launch guard was read before work and remained session `2026-10-08T20:26:35Z-81e5`.

## Result

All requested checks passed using Node **24.16.0** and pnpm **11.20.0**:

- Clean build, typecheck, boundaries, format and lint: exit 0.
- Cumulative-artifact tests: **6 passed**, no failures or skips.
- Fresh protocol/server tarballs installed into this report directory's isolated `install/`, then the four-file RU suite: **152 passed**, no failures or skips, **88.66 seconds** supervised command duration. File counts: matrix 117, development-server 20, installed-read-oracles 11, exact-client 4.

The ordinary integration changes only four evidence/test files relative to `0944fbaa19ab2b3399a7547eb6acf74ed9f5c5ff`: RU matrix test, RU manifest, cumulative-artifact test and RU coverage inventory. Server, protocol, client, config and scripts trees are byte-identical by Git tree identity. Full 8k tests were intentionally not repeated; the earlier 0944 CI remains evidence for its own source, not a new 30351 CI result. Reference `579019b676568345f93bb424b2fa6f5774211df2` is absent from reachable ancestry and was not included.

## Actual installed observations

The fresh baseline report records **47 passing authored scenarios**, 47 installed sessions, **903 completed raw HTTP serializer observations**, 44 setup observations, 21 native Read observations and **787 reported exchanges**. These are different observation layers, not additive test counts. The unfiltered cumulative result also reports **87 harness-error scenarios** for unsupported/unplanned execution; they are preserved, not converted into passing evidence. Required rows: **129**; planned rows: **47**; whole rows: **28**; uncovered rows: **101**. `claimEligible` remains **false**.

The separate focused admission report contains **2 passing scenarios**, 2 sessions, 122 raw observations and 345 reported exchanges. It shares the exact source/configuration/catalog/manifest/fixture binding below; it is not a full-profile report. Negative-control tests are part of the 152-test result; the saved positive observation reports have counterfactualApplications 0.

All baseline observations use `raw-http1-serializer`, writeState `complete`; 200 distinct method/URL route pairs are listed in `summary.json`. Canonical Scope is `https://reference.invalid/acme/`, profile `read-update`, History false, fixture `reference-read-update-schema-v1`, capability `installed-read-update-http`.

| Binding | Actual value |
| --- | --- |
| Exact configuration digest | `a8d5687519da4040d2f23989f86f9bc804d30ac6c91a85074397eaf7c92de712` |
| Live configuration identity | `1720b63b49a7be3299d4b2865de4db9fcbe47560f217a32eea6397db7ea54ca5` |
| Installed payload digest | `78a80950acd6c1981d6829fc4d59c05eaafc65e0d68cf20d3d84f61f377acb99` |
| Installed contracts digest | `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe` |
| Policy digest | `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430` |
| Authentication digest | `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e` |
| Cumulative catalog digest | `e26d3fc758a34ebb0b5028be15d28e3e09db87dc7984bba7bad6a00b26329aa0` |
| Manifest digest | `0c6e1c9e9853e82e714e919ba23fa33dbf267c8a36ae36aa9457c1a9d3efc64a` |
| Fixture digest | `64814924334103e9375c043de472f8bf15c8b71a267a176e06cd0443b8249751` |
| Installed RU entry SHA-256 | `3eb9a6bc8df1237d78ba4aae7e95589df6fb7f4781f572e04adda583f359c71d` |
| Protocol tarball SHA-256 | `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c` |
| Server tarball SHA-256 | `aa52a497f921dc00f4b66e94cf31843d0c91613019ddbb9292a9319d1ba8115b` |

The actual RU entry was `install/node_modules/@bdp/server/dist/installed-reference-process.js`. Five constituent catalog hashes and all route counts are retained in the safe whitelist-only `summary.json`. `source.json` also hashes the packaged TX entry as package inventory; no TX run is claimed here. No raw credentials are included in this summary or report.

## Reproduction and cleanup

`run.py` is the exact pinned runner. Stage commands, exit statuses and elapsed times are in `current.json`, `results.json`, individual `.exit.json`/`.log` files and Vitest JSON reports. Each command was process-group supervised with a finite ceiling; all completed naturally. Source cleanliness and exact head were rechecked throughout and at completion (`source-before.json`, `source-after.json`). No owned installed children or test processes remain in the after-run process inventory. An unrelated pre-existing process from another checkout was left untouched.

This validates the requested bounded evidence integration only. It does not enable public write profiles, assert full Update coverage, qualify Reference579, regenerate historical Read evidence, or replace exact-head remote CI.
