# Receipt installed preparation — closed, not executed

This mechanically adapts the reviewed Event installed workflow for the frozen receipt source. It creates no live execution gate and performs no build, archive, consumer installation, product import, service, test or retry. The ordinary clean source commit remains a mandatory root-supplied pending value after actual Claude/control disposition.

Source authority: D/source-binding.json edc5c98c8e5c82f2073e1bf35aa84098efe9552c7dbc74537784de75788cca5a; source diff837bacdc192bcf0401171ecc9a8312ae034beabed7b69620b6a3beaef0cb9de7; base015cac1aa4fa827cc4a2c647dff08454c6c8e6f8. Exactly eight owned paths must match the future clean ordinary direct successor. Any source change requires a reviewed rebind, not relabeling this build.

## Commands and gates

The existing ten-stage order is unchanged: node-version, package-directory, pack-protocol, pack-server, verify-tar, prepare-consumer, consumer-install, verify-installed, installed-positive, installed-review. commands.json contains exact argv/cwd; each is bounded120s and requires all predecessors clean. Root may release only a sequential prefix. Typical root invocation after writing its gate is `python3 supervise.py STAGE`; the supervisor invokes the session guard and pins Node24.16.0 in child PATH. There is no automatic loop or retry.

Each command verifies clean final HEAD/tree, exact direct parent, eight-path hashes/scope, root/council/control/final-source acceptance documents, frozen source lock, all current build inputs and outputs, preparation hashes, prior outcome and cleanup. Final head/tree/acceptance/consumer lock/provenance fields remain pending in execution-gate.example.json. No live execution-gate.json exists. Installed entry overrides are released only after verify-installed and root's actual provenance/consumer-lock pins. No fallback to workspace dist is allowed by the supervisor gate.

## Current build and future package provenance

Read-only build-output-inventory.json binds the actual successful receipt build's original start/exit records and source binding,666 input files including the new test, and312 distinct package-source/output files. This is a dirty-source build record; a later clean commit may change identity only. No Event build is relabeled. Future archives must contain exactly all current package dist/schema members plus manifests/LICENSE, with only the expected server workspace dependency rewrite. LICENSE participates in both archives; the expected archive-member sum is313, which must be established by the actual packs rather than asserted as executed now.

The external consumer uses only the two fresh file archives, offline pnpm11.20, ignore-scripts and prod, with the previously reviewed protocol-tar override and fast-uri3.1.5 override matching the frozen source lock. verify-installed compares all archive/installed/source bytes and exact eight-package dependency closure, resolving every dependency inside the consumer. It independently reproduces the installed payload's sorted path/length/content recursive digest and binds actual entry, lock, manifests, settings and configuration fields. No source attestation is invented: final commit attribution is supervisor verified, while the child's payload digest is independently compared after the real run.

## One bounded installed case

Exact title: `observes genuine installed receipt root and page absence context at current heads`. One case,14 requests maximum/expected, one child/store/directory/listener. Per request5s, cohort45s, test60s, existing startup/close10s each;120s external supervisor. No new source tests, control, timeout, response ceiling or hidden polling is added. Prefix bdp-receipt-missing-context-installed-. The test uses actual response2.id/next and validates root/page relationship before adding exact routes. Its postprocessor binds the actual retained route map and contextual progression pos1→pos2→pos3;401 remains contextless. Dynamic root/page values in the postprocessor come from actual response URLs, while issued id/next and page.receipt equality are established by bound in-process schema/body assertions—not inferred from fingerprints.

The postprocessor requires exactly one passed callback,14 complete response records, actual final source/entry/payload and configuration binding,328 artifact UTF8 bytes, safe header/fingerprint validity, HEAD length/empty-body behavior, zero failures/deadline, normal child exit0, independently absent PID and store directory. Full schemas, bodies and exact request serializer fields are asserted in memory; saved logs contain fingerprints and selected headers only. Actual raw request method is the authored ledger plus returned response mapping, not an invented exactRequest.method field.

## Mechanical delta from Event preparation

- gate.py: receipt parent, eight-path binding, commands digest and scope diagnostic only.
- supervise.py: byte-identical reviewed supervisor; conservative115s TERM/120s KILL, owned process group only, retained logs/snapshots/temp classification and stop-on-failure unchanged.
- packages.py: external consumer name only; archive/member/closure/payload algorithm unchanged.
- review-installed.py: receipt log label/prefix, artifact328,14-record map with actual root/page bindings, pos1/2/3 and anonymous tuple partition, receipt-specific404 HEAD/metadata checks.
- commands.json: receipt paths/title/binding,14/1 workload and output paths. No extra stage.

Every before/after delta is saved as NAME.patch. Scripts were syntax-parsed with Python ast only; none was executed or imported. Original Event files/results remain untouched. Root review and explicit live gates remain necessary.
