# Independent alias setup witness review and ordinary merge

Reviewed frozen author `f0f3cb9f4768d9dad4df7986e763eb51e2cb4cd0` against
`21654be3cae51d50e75e817ef8ac57735b8ee4ea` (author commits `6ae9878` and
`f0f3cb9`). No blocking findings. The author checkout was not edited.

## Review findings

The change touches only the development matrix test/helper, diagnostic manifest,
coverage inventory and README. Eleven new real alias GET observations immediately
follow their corresponding repoint/delete/restore/assignment mutations and precede
the retained-replay assertions: six in recorded-targets, three in recorded-misses
and two in pinned-references. Each live alias requires 307, the expected canonical
in-Scope Location and no body; each deletion requires 404/resource-not-found.
The SDK alias resolver uses manual redirects and enforces the bodyless response.
The runner's scope-relative URL normalization first checks canonical serialization,
Scope confinement and canonical path before comparing the target. These witnesses
therefore establish setup state independently of a preceding POST 200.

The helper opts into alias resolution only with `resolveAlias: true`, restricted
to GET and the current fixture Scope's alias root. The added negative control
rejects a non-alias GET, a POST and an off-Scope target before transport. No runtime
or general SDK transport behavior changed. Existing assertion-failure and barrier
abort cleanup negative controls continue passing.

The allocation observation now directly compares the allocated Bead URL against
`beads/reserved`, the Bead identity corresponding to the reserved alias segment.
It separately checks in-Scope/root/one-nonempty-segment/read-back behavior. The
renamed URL-serialization assertion does not claim the independent BDP canonical
segment grammar. The README and inventory explicitly leave deterministic allocator
collision skipping unqualified. Original alias usability and explicit creation
refusal remain observed. No public profile or normative row is newly claimed.

## Composition

Frozen clean merge candidate: `2e2d274d1db60ba9b28a95c553c5bbb86ee5e7b4`.

Parents:
- Published integration `e538e23b6a4a7371b14945f59009323c302e60a3`.
- Reviewed alias author `f0f3cb9f4768d9dad4df7986e763eb51e2cb4cd0`.

Checkout: `/Users/dbox/repos/wt/vickie/bdp-alias-witnesses-review`.
Branch: `codex/vickie-bdp-alias-witnesses-integration`.
Origin: `https://github.com/gastownhall/bdp.git`.

Ordinary merge. Manifest and coverage combined automatically. The only conflict
was the README count sentence: preserved both additive changes as 25 plans and
264 HTTP/lifecycle actions. No action/assertion/runtime resolution edits.
Programmatic comparison verifies each of all 25 merged plans and all 81 inventory
rows equals the applicable reviewed parent version. Relative to the published
parent, only the same four diagnostic files change. Spec, schema, runtime and
checked-in evidence remain unchanged. Both parents' history is preserved.

The combined diagnostic report has 25 passing plans, 264 actions and 827/827
passing assertions; claimEligible remains false. Inventory counts remain 54
bounded-observation rows and 27 omitted rows. This does not enlarge the current
bounded behavioral claims to complete normative coverage.

## Independent validation

Pinned Node 24.16.0 / pnpm 11.20.0; frozen install and build before process tests.

- Exact author head: build and 129 tests across five focused files passed.
- Combined candidate: build and 196 tests across seven focused files passed,
  including SDK transport, development matrix, receiver, disconnect/process
  recovery, sequence executor and development/normative inventory checks.
- Candidate typecheck, format check, dependency boundaries, lint, installed-package
  smoke, development smoke and evidence:verify passed. Lint has only the unchanged
  three warnings and one informational finding.
- Evidence verifier confirmed seal `6142cf73deae687bc3e9668cc014ea408fa02843`,
  run head `4f0ed57906525c36bc55b343b77e0f07edbb7553`, evidence commit
  `72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5`, and 89 target rows.
- git diff --check passed. No broad full-suite run or remote CI run was performed
  for this local test-only integration. No outward writes or profile promotion.

Logs and reports: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/alias-witnesses-review/`.
The candidate diagnostic report is `candidate-report.json`; full test output is
`candidate-tests.log`. The author report/tests are saved separately. All commands
completed with exit zero; no interrupted test or smoke process remains active.
