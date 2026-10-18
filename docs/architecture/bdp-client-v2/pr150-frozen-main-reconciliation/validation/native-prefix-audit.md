# Native completed-prefix audit

Accepted completed prefix only at `d7197e97a29bff8f075028419afad05388e2e655`, tree `630a94d0ca1545f898cb8a8b614d4b7d2dd6d454`. Independently rehashed all701 tracked current bytes against candidate-source.json and reviewed staged source-freeze.json; clean Git status, exact ordered d4/8d parents and unchanged source lock verified. Current preparation index `be10d7342a150dda502bba130eb7e999fb82f6c2850657e4f8c1770adb535c4f` and all indexed files verify. No full-suite completion or runtime/package claim is made; the running full suite and its live outputs were not read as evidence.

All seven prefix stages have raw exit0, accepted true, identical source identity `b9fe0a23fdb4026641227ec273f05cb2144855271094b6429193b10a7412375a`, actual argv/cwd matching commands.json, elapsed below their fixed bounds, automaticRetries0, unchanged source, no forced cleanup/errors/reported survivors/runtime store directories. Raw log SHA256s independently match exit records. All7 sampled prefix PIDs are absent in a fresh process snapshot; terminal per-stage records and actual stage temp trees are consistent with cleanup. This is sampled/group evidence, not an assertion about unobserved detached processes.

Dependencies installed61 packages from the content-addressed store (61 reused,0 downloaded), frozen lock retained, exact declared versions reported. All673 measured workspace tool files and their realpaths/hashes verify, including actual TypeScript7 tsc entry rather than an assumed tsgo name; all base/isolated tools, Python environment and pinned provisioning inputs also verify. The initially guessed tool filename was a stopped root measurement, not a failed product run or a reused tool identity.

Format checked465 files, no fixes, with one warning: the existing3.5MiB read-update matrix exceeds the1.0MiB configured formatter size cap. Lint checked465 files and applied no fixes; actual totals are **1030 warnings,84 infos,1094 diagnostics not shown**. These are not the older58-warning result and no warning-free claim is made. Dependency boundaries passed for11 workspace packages. Typecheck completed with no diagnostics. Build log shows the existing clean-before-tsc sequence; build and bind-build both succeeded.

Fresh build inventory independently reproduced all477 recorded artifact entries, including emitted output, schema and package metadata; head/tree/source identity and all build start/exit/log record hashes match. No product/helper import, tests, rebuild, source edit, gate change or indexed-packet mutation occurred during this audit.

| Completed stage | Seconds | Log SHA256 | Exit record SHA256 |
|---|---:|---|---|
| dependencies | 0.835 | `d35e164b2599e806c08b40f6393d6b04fb816d521ffe035705c866b56632e385` | `5077c7932afd872fdfd599709e538b023878d702f27a8d3404b4624484925684` |
| format-check | 1.275 | `e1133c825c26ef3b81a1020dd291331e090f922efdeab41e0545476f96f7658a` | `b1f32ef2e7b99e7313ed379a3dabc8f1af94f5726c51803ac04adada80b4a44c` |
| lint | 0.737 | `b1de538f5f2eff553a157d258a2fef991a8dc150de7299b49313244f2aba68e3` | `f2285e40156f43fc6d1b3e28dd4d99aa35ef4743e1846adf27cc2c9e1d53227b` |
| check-boundaries | 0.713 | `53f4cd856bde25513486a59799c5a05e0650fbb756bf1580fb07e56c2fdeeef3` | `9e72d3814ef0bd28cfcc991e1ae161a0b55291096383c7b9812c35820575fa33` |
| typecheck | 1.632 | `91d54ffbee73a3ac72294d090e58d77745f8b93edbf94b862807698273a6505a` | `7674d2da011e13f6fc94e4a849a24acd65b1e815143f61ce626da50d90339333` |
| build | 1.136 | `232dac5e840e45b3bb06e6e04249b5241f2e1b4f6c797a22c046d916302ac208` | `c4d0374faaf923c2cde4e47f311a653322edec73423c3abd9ff03e6ee0cd9b75` |
| bind-build | 0.855 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` | `c97f0cd17caac2bd8542aec743f6a7a94bc8540fa9bfe1ca943fbacfec4b050d` |

Audited at 2026-10-17T11:01:57.710427+00:00. Build inventory SHA `91c9c4bb5cfa835bb1f9350605ffe45429dafe6ff742d1bc917802ba7f9b2454`; candidate binding SHA `27b75c94374685505b3503f9a321767831350b25906cae37364ca95bc556fc5f`; workspace tools SHA `87fa39f281cfbad1666fc2a332be09708cd5c6ff20c3b19b4f26d2d4427bf7e2`. Runtime/full-suite, package/consumer and exact-head CI remain separate gates.
