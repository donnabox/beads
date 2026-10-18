# Installed Update integration

Exact candidate: `560c10a75797c2f1f66641d7407e11ba98343a44`, branch `codex/vickie-installed-profiles`, [draft PR88](https://github.com/gastownhall/bdp/pull/88), includes PR86 by ordinary ancestry, targets main for existing CI.

Ordinary merge of `dea6f9e7a95a3e3eba9c667de39360346f25bdf4` into numeric/PR86 integration `fd21cf32c171b305c9d655e7dac720a50abf724c`. The only conflict was the smoke private-export denial list; both parents' entries were retained. Tree verified before merge commit; no source changes afterward. Remote head verified.

Pinned Node24.16.0/pnpm11.20.0: clean build; full serial suite 7,963 passed /124files/one optional skip; typecheck, format, boundaries, lint (existing warnings), installed tarball smoke, development smoke,7preflight tests passed. Existing89-row Read evidence verified unchanged. Logs `installed-profiles-dea-{build,full,typecheck,format,boundaries,lint,smoke,development-smoke,preflight,evidence}.log` in the integration log directory.

Native closes startup/import/owner-limit defects at dea. External4078 review independently passed7,918tests/120files and confirmed three defects already repaired in dea; its exact dea closure is pending. New duplicated-composition Medium and bounded cleanup/identity/marker/closure followups are assigned to a fresh successor. PR88 remains draft. No public write profile, cumulative qualification, main merge or Preview2 claim.
