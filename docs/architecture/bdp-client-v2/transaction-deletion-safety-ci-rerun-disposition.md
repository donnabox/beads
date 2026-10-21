# PR115 first CI timeout and one unchanged-source retry

Root fully read native diagnosis9d2dbdc57668905e63840362bc05b03f6f707ac48d311c51b978de6001f4564e. GitHub explicitly reports job15minute limit. Raw job log shows passing tests ten seconds before cancellation; aggregate gh log truncation had concealed progress. Common test durations median1.877x prior green PR114, Go install188 vs108s, plus real cumulative workload growth. This does not establish a host-level cause and does not prove unfinished tests passed.

Accept one unchanged39c160d0d35294003f400cef8d5ab5c6ec9aec9c failed-job rerun, guarded remote launch check, requested successfully. Attempt2/job115925110459 started2026-10-14T14:57:45Z. Preserve attempt1 cancellation. No source/workflow/deadline/test changes. No merge until exact-head completegreen. If repeatable exhaustion, prepare measured separately reviewed scheduling/workload proposal retaining all tests and every runtime/exchange/cohort bound; no blind repeated retries.
