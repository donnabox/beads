# P1a installed mutable payload and U2 council repairs

Frozen clean head: `4078ca709b55d363c147a5bdc0d5141ede9ec247`.
Checkout `/Users/dbox/repos/wt/vickie/vickie-update-installed`, branch
`codex/vickie-update-installed`, origin `https://github.com/gastownhall/bdp.git`.
Base U2 `2c5fe9e04250ae036e681f46c60d3592e5f1cdf1` remains frozen in its own checkout.

Two ordinary commits:

1. `46d8809`: private installed mutable reference factory/configuration/process,
   actual packed-process smoke and lifecycle/transport qualification. Root has
   checkpoint-pushed this commit; worker performed no outward writes.
2. `4078ca7`: independently requested U2 council repairs and focused regressions,
   applied to this successor rather than changing frozen U2.

## P1a implementation

The emitted server package includes private `installed-reference-configuration`,
`installed-reference`, and `installed-reference-process` modules. No exports,
bins, ordinary startup flags, public admission changes or evidence reseals.
Ordinary bdptest Read startup remains separate and write profiles fail closed.
The later evidence-gated public path must call this exact factory and capture
fresh evidence for its final payload; IPC is process control, not admission.

Configuration arrives through inherited IPC only. No credentials in argv or
logs. Closed/detached configuration supplies canonical Scope, separate loopback
host/port, absolute directory, explicit create/reopen, trusted credentials and
principal policies, offline Type/schema artifacts and fixed aggregate policy.
The factory compiles before startup. It uses the actual mutable owner for all
reads, writes, alias resolution, Type inventory/descriptor reads and discovery.
Empty descriptor input instantiates the whole shipped reference Type closure
at Scope-local identities; inverse-remapping tests compare every field to the
original descriptors. Schema routes serve exact captured local manifest bytes.
No network fetch, generic plugins or replacement store implementation.

Raw targets are normalized once before all route classification. Valid
absolute-form Read/singleton/sequence URLs use configured canonical origin;
Host and forwarding headers cannot rebind identity. Installed CORS-disabled
mode accepts Origin normally without grant headers and answers OPTIONS with
endpoint-specific bodyless405. Development restrictions are preserved.
Existing Read routing/HTTP helpers are reused without changing their files.

One installed process owns the full lifetime: ready follows compilation/store
validation/principal checks/route construction, second startup is rejected,
SIGTERM or parent-disconnect closes receiving and settles the real owner.
Partial unadmitted bodies close without capturing keys. Listener cleanup failure
cannot skip owner cleanup; an actual injected failure test preserves original
and cleanup causes and proves reopening the same real store succeeds.

Identity metadata is derived, never caller asserted:
- artifact identity includes exact captured closure and compiler input bounds;
- configuration identity includes canonical Scope, artifact identity, fixed
  Read/Update/retention limits, aggregate, credential-to-principal mapping and
  policy snapshots (no raw secrets emitted);
- payload identity hashes actual installed server/protocol JavaScript files.
Listener port/create choice are distinct from persisted Scope identity. These
hashes are observation metadata, not self-certifying admission/evidence tokens.

## U2 council dispositions

- A1 fixed: emit confirmed compiler diagnostic prefix until the receiver stops.
  Then return invalid/incomplete. If compiler truncates first and receiver has
  not stopped, still throw instead of claiming completeness. Actual 1025-allOf
  HTTP create yields422/32 diagnostics/truncated, no Resource, retained replay,
  same-key conflict and corrected fresh-key success. Actual 4000-uniqueItems
  evaluator `limit-work` yields local empty500/no effects; corrected same key
  succeeds, proving no retained key. No evaluator/store changes.
- A2 fixed: shared aggregate capture requires positive max before provisioning.
  max0 refusal leaves no reference.sqlite; corrected same-directory creation
  succeeds.
- A3 fixed: shared snapshot sorts/deduplicates exact aggregate entries before
  installation/configuration hashing and execution. Reorder/exact duplicate
  reopen preserves a real Bead. Different redundant numerical maxima are not
  coalesced; no broader semantic equivalence claim.
- A4 fixed: compiled config/policy brands and each credential's policy membership
  checked before provisioning. Missing principal leaves no database and corrected
  create succeeds. Startup discovery must return its actual `discovery` success
  variant, not merely avoid throwing.
- A5 fixed: artifact receipt limits plus fixed graph/evaluator bounds join the
  contract identity. Two successfully compiled byte-identical closures with
  different accepted artifact bounds now have different identities.
- A6 fixed/documented: captured Scope config is frozen; its global policy digest
  is explicitly distinct from per-principal pagination authorization identity.
- A7 bounded limitation retained: read is a finite deny-list fixture, not a
  whole-Scope deny-read capability. No IAM product expansion in this slice.
- A8 fixed: canonical protocol path predicate permits legal nested/dotted/encoded
  Resource identities, observed through actual hidden/read-only HTTP behavior.
- A9 fixed: configured authority exposes only explicit principal facets; default
  unconfigured authority preserves its original facet. Direct test pins absence
  of configured implicit mutate/discover methods.
- A10 fixed: credential principals share one cached facet/route and policy
  membership uses private Sets. Alternate credential for Alice constructs no
  duplicate route and authenticates to the same principal; Bob remains separate.

**Identity migration boundary:** corrected custom contract/aggregate identities
intentionally refuse earlier U2 custom installations. No migration or adoption
of old custom stores is implemented. U2 artifacts/results remain historical at
2c5fe9e and must not be reopened/relabelled as final evidence. The original
unconfigured development installation identity is unchanged. This worker used
only disposable stores; no shared/pinned database touched.

## Verification

Pinned Node24.16.0/pnpm11.20.0; frozen install. Build completed before process tests.

- `final-qualified-tests.log`: **491 passed / 14 files**, maxWorkers2. Includes
  final14 U2 tests,8 installed tests, original development HTTP/discovery/process/
  matrix, Read lifecycle, public Read HTTP/listener, recovery store, compiler and
  artifact ingestion. This is a targeted suite, not the full repository suite.
- `final-build.log`, `final-typecheck.log`, `final-check-boundaries.log`,
  `final-format-check.log`: pass.
- `final-lint.log`: pass, existing3warnings/1info only.
- `final-smoke.log`: actual packed installed entry executed outside workspace,
  all9 targets,13 local Types, raw absolute-form/Origin/Host forwarding, HEAD/
  conditional/method/Accept, duplicates/media failures with corrected-key reuse,
  canonical Scope reopen on different listener and IPC-disconnect reopen.
  Ordinary bdptest/bdpbd Read admission and write refusal remain proven.
- `final-smoke-development.log`: pass.
- `final-evidence-verify.log`: unchanged Read artifact verified; no new observation
  relabelled as that artifact's historical run.
- `git diff --check` clean; no test-owned processes remain.

Final packed payload observation: `b8a74cc5aaf9a6049b61970285ac72710e4623f856b26fd32e401c804e8632db`.
Configuration identity from that disposable run:
`4786177cc61e61a2670b783802bce1d08a07ed3a651dd60947b81a28cf47cfd9`.
Artifact identity:
`development-schema5-e1dff401b5caafb04a4d935775d67eeefeae7bcf306d5d100b979a64eef148e7`.
Configuration identity varies across fresh runs because credentials are fresh;
within a run it is identical across reopen on a different listener port.

## Remaining profile gaps

No public Write profile is activated. Cumulative required-contract runner,
applicability/counterfactuals, independent profile-specific evidence generation
and verification, final public startup/configuration gate and final exact payload
qualification remain. This loopback candidate is not deployment TLS or optional
CORS-enabled/History support. Compiler schema breadth/production work-and-heap
qualification remains separate; true budget refusal is a local500, not a new
protocol Problem or manufactured validation result.

Installed smoke does not establish deterministic between-member interruption,
mid-commit crash or power-loss durability. Existing deterministic development
process tests retain their scope; no IPC scheduling/fault/store hook was added.
No TX owner/store, authority-read/pagination, normative schema, shared evaluator,
Janet/Trish/Preview2 files, shared databases, public exports or old Read artifact
were edited. No product decision identified.
