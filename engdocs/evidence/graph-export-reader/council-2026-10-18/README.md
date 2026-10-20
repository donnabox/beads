# Export reader council and test repairs — October 18, 2026

Source `79b2601e39fbae043f3d2558fb79d467ca6d1bec`, tree `1d023cf50124520d1de8706fadcdba141bbdbdff`, parent `b5973def8e395e5acee6707ac02bb3bd654c03f1`. Independent of legacy import; no public format, serializer or CLI is added.

Completed actual full external review exited0 with empty stderr; raw report and exact five-file manifest retained. No production correctness/regression defect found. It independently verified database/sql's Next/Close cause suppression, so the existing cleanup assertions remain correct.

| Finding | Disposition |
|---|---|
| Low1 descriptor identity absent from JSON fingerprint | Fixed independent DeepEqual of fresh/reopened Type descriptors, including order |
| Low2 different deleted-head budget baseline | Fixed matching filler deletion/allocation state; added independent full-source rollback oracles for both isolation blocks |
| Low3 wisp_dependencies independent probe unreachable with intact FK | Document deliberate defense against FK-bypassed corruption; no unnecessary corruption writer introduced |
| Low4 duplicated source-state helper | Deferred shared test refactor; would widen unrelated qualification scope. Existing explicit table sets remain reviewed |
| Low5 use outcome_unknown for feed/count drift | Disputed: reader performed no uncertain mutation; malformed/incomplete driver feed or broken snapshot inventory is not only concurrent persisted corruption. Keep explicit inventory reason and existing ErrInvalidStore until API classification agreement; no sentinel contract change |

Fresh independent native delta review has no actionable finding. Actual external delta review completed exit0 with empty stderr and no actionable finding; report/exit/hash retained. It independently agrees with the Low5 sentinel disposition. One nonblocking duplication nit has the existing comment mitigation and stays within the already deferred helper refactor. Frozen manifest/prompt retained; Gemini excluded.

Canonical no-engine checks use BEADS_TEST_SHARED_SERVER=0 BEADS_TEST_ENV_RUN_DOLT=0 BEADS_TEST_PREBUILT_TEST_BINARY='' scripts/test.sh -json -count=1 -timeout90s ./internal/storage/graphstore with explicit pure selections: context/originalUTF8/closure/projection, rollback/commentRows/badConnection/physicalClose/scanClose (28 records) plus feed/Next cleanup (12 records). Aggregate40 pass records, zero failures/skips, both exit0. Pinned golangci-lint v2.10.1 with GOTOOLCHAIN=go1.26.5 and --new-from-rev=b5973def8e395e5acee6707ac02bb3bd654c03f1 --build-tags=gms_pure_go --timeout=5m on graphstore reports0issues. Gofmt/diff whitespace check passed.

New test assertions are database-backed and UNRUN; compile/pure results do not prove engine behavior. Installed embedded/shared-server qualification, full checks and coordinated format/CLI contracts remain OPEN. Janet owns the engine slot; Iris ran no producer/DB/CLI workflow. Managed hooks bypassed because they invoke ambient bd/broad engine checks; direct evidence is not a waiver. Earlier runtime evidence never qualifies changed source. No release or merge readiness claim.
