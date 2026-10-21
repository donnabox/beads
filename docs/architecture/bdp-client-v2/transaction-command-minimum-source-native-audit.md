# Command minimum actual source results — independent native audit

No actionable findings in the recorded source validation. This closes the bounded source-results audit; controls, fresh installed TX, CI and public qualification are separate.

## Exact basis

- Final source binding: `ff9c63417fbb348bf8ad9863483096ccc0329beb5c05b532333cf75cd0726a38`.
- Actual seven-file working diff and saved draft: `891a3dd1d17d9f754dc81ef6223bdb61de4c330e15bc5ab7760559ca5c563106`.
- Root source summary: `0d5ecc3dbf994af29e1eaab39018f0cac1c7f21f391803c1ef78d14dceb797a2`.
- I independently recomputed all seven actual file hashes and the actual Git diff; all match the final binding. Checkout has exactly the seven recorded modified files and no untracked files in status.

## Recorded execution and selection

The six static commands (typecheck, fresh build, format check, lint, boundaries, historical Read evidence verification) each record raw exit 0 and supervisor exit 0 without timeout, forced cleanup or errors. Format and lint are not warning-free: format reports the unchanged large Read/Update matrix size skip; lint reports 712 warnings and 47 infos. The Read verifier confirms its existing historical artifact, not a new Read runtime cohort.

I read each complete JSON result and independently selected statuses other than skipped/pending, then matched every executed fullName against its actual anchored command pattern. Exact executed counts are HTTP-positive 5, service-exclusions 1, catalog-citations 1 and RU-emitted 1: eight passes, zero failures. Respectively 69, 64, 8 and 14 other cases were skipped. Every result reports success and no failed selected case. Root's pending-versus-skipped filtering correction concerns adjudication of these same saved results; it adds no runtime execution.

The selected titles are:

1. ignores partial and malformed copied read controls on each mutation carrier
2. ignores complete copied tuples on every Transactional mutation target
3. copied read controls preserve key revision and negotiation admission
4. keeps receipt minimum enforcement after commands ignore copied controls
5. preserves authentication on commands carrying copied read controls without reserving keys
6. nonexecuting sequence refusal observation keeps excluded sequence responses unobserved and without refusal context
7. draft Transactional catalog strictly parses and binds every citation to the current specification text
8. ignores copied read minimum fields on Update singleton and sequence commands

The 80 HTTP exchanges for source HTTP A–E remain counts inferred from the passed bound test bodies, not a captured wire transcript. Service exclusions contributes six direct handler calls, not HTTP. RU has an actual six-completed-response record: four POST and two GET, all 200, with child PID 97756, exit 0, childClosed true, directoryAbsent true and claimEligible false. The emitted workspace-child test is not the later fresh packed-consumer TX proof. No raw empty-header claim is added by this audit.

## Cleanup

All ten command exits record empty remaining process groups and empty runtime-root residues, source unchanged, and no sampled PID left alive. I independently queried the exact union of 31 recorded PIDs at 2026-10-14T22:07:46Z; none was present, including RU child 97756. All seven sampled non-cache temporary paths were absent on direct filesystem inspection, including HTTP/service SQLite fixture roots and RU `bdp-installed-irPF19`. Preserved `node-compile-cache` directories are compiler caches, not live service/store leaks. Sampling at 50 ms cannot establish an exhaustive history of short-lived or detached descendants; the recorded helper cleanup, final process checks and direct path checks are the evidence actually available.

## Artifact hashes

- `typecheck.exit.json`: `c425a5c1d34a5027ecbc7a43dad74b4f606b63c31e115fbd23334422cd33b6e3`.
- `build.exit.json`: `bb428285b52816d3860ee772039d258e94fcbb93a2e0a45a6be5014dfaeb9f6b`.
- `format-check.exit.json`: `bff96348c8e10bdbef08f0909721600633520d7aac4102b3f708f2ea3c797d5d`.
- `lint.exit.json`: `ee930a2b261ac5d29a6df21bacff0beb65a4c8bad2dfa713ab3912d95e311013`.
- `check-boundaries.exit.json`: `4fe3aee99e1df30f51c4cc2640b561cd6035172b5a9a2fdf8febdf8a7ac8072d`.
- `evidence-verify.exit.json`: `0eda2eba2c7d846ab8c31b4adc08229bbf9d407b9290d6d762119f3ae1e6bba9`.
- `http-positive.exit.json`: `349bab81713aceec1fad2df1caa533fccb122c3edbd3011009207c3552fa3066`.
- `http-positive.results.json`: `6dae26455840bf19ee25a636b70a3bac76f5f40b2107eb52d39fc662b95bc000`.
- `service-exclusions.exit.json`: `09816d6840b45c22d71f066fd257c3bd5add4f708549520e8d9dd6cf656c153e`.
- `service-exclusions.results.json`: `7f45d517cc88a3a9c711eb4fbea66bcf686fe70ba75b185daee4b3ebe2c123bc`.
- `catalog-citations.exit.json`: `137fffd0440544287a87704100dddbfcde8830bee2485afce675397bedefaf29`.
- `catalog-citations.results.json`: `8ff8336f249ed227f27da87b8934b482511164ff9fec12338983e0adbfdb15bf`.
- `ru-emitted.exit.json`: `f8bde83f1c5ebed9c4cdfa260edf3e6a303114c7f3aec02bebd084012fd083ff`.
- `ru-emitted.results.json`: `40d12fd516ed6e912bc18e989e9d62f42c7ae2f3a09ca7c02a20e9231656f0d8`.

Review was read-only: JSON/log/source hashing, Git diff/status reads and targeted process/path inspection only. No product import, test, build, install, service or control was run by this reviewer. Concurrent isolated control work was not inspected or disturbed. Public Transactional and Update profile delivery remains incomplete.
