# CI headroom proposal: held after successful retry

PR116 unchanged candidate8c CI attempt2 completed all checks successfully and merged as f76bf119f531c6b44d1d3bc43531a4a8da27c4d0 with the exact reviewed tree. No CI configuration change is needed to unblock this delivery. The conditional 15-to-20-minute workflow proposal remains unimplemented.

External Claude plan review completed exit0 and was fully read by root. Its required findings1–3 are accepted for any future revision: include the successful PR115 identical-matrix baseline; distinguish modest added workload from the dominant throughput episode; justify timeout headroom against measured variance, and compare both normal and slow regimes for partitioning. Optional findings4–7 are retained: consistent step/reserve accounting, deliberate repeated preflight, named build-dependent optional skip, and required-check migration if splitting jobs. No claim that the current draft has resolved these findings. The original proposal and full review remain preserved.

Future CI work must first revise and review these grounds; a green retry does not authorize quietly changing workflow limits. No timeout, test, runtime or installed-cohort limit changed here.
