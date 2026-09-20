# Managed controller with a real engine

The opt-in `TestManagedEnginePersistence` regression couples the private managed
controller to a real Dolt/GMS service. It writes an exact text value over MySQL,
closes and reaps the first child, then starts a second child over the same
storage and reads the value without reseeding. This catches integration defects
that the existing SQL-free process fixtures cannot exercise.

From the repository root, with the supported Go/CGO toolchain:

```sh
go test -tags 'gms_pure_go,graphmanaged_engine' ./internal/doltserver/graphmanaged -run '^TestManagedEnginePersistence$' -count=1 -timeout=4m -v
```

The extra build tag keeps the heavy engine dependency out of ordinary controller
unit tests. Every test run creates its own temporary database, private HOME and
configuration, random test password, and loopback listener on an assigned port.
It does not use an existing Beads workspace or launch an installed Dolt binary.
The child receives the controller's inherited socket; it creates no second SQL
listener. Child PIDs and endpoints are logged. A failed query, close, or reap
fails the test. Temporary test state is removed after the owned children exit.

A run using a local GMS change must supply its explicit `-modfile=...` and retain
the effective build information alongside its test output. A passing run with
the repository's ordinary module graph does not validate a different patched
engine. The September 17 integration run used the separately recorded terminal
session cleanup patch through an isolated module file; repository dependencies
were unchanged.

This test is a private integration fixture. Its profile/configuration/registration
values are synthetic, as in the existing process fixtures. It does **not** qualify
production admission, all session-exclusion laws, engine provenance enforcement,
or a public managed-adapter API. Its small SQL table is not a graph Memory Bead
or Link; no BDP write, transaction, or History capability is implied.
