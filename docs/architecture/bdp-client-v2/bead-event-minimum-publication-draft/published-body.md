Initial finite canonical Bead Event GET/HEAD requests with an opaque minimum and the current epoch/view now wait within the existing bound and return contextual `503 catch-up-timeout` when eligibility remains unproved. Previously they reached the private minimum-position refusal. The service captures the tuple after preflight and preserves it across retries; source visibility is checked before minimum comparison or retained scanning. Authorized deleted Bead identities remain readable Event Sources, while unknown and hidden identities retain contextual 404 responses.

The six-file change adds owner, service, HTTP and installed controls for retained identities, preflight ordering, cancellation/drain, conditional requests and GET/HEAD parity. Historical installed callbacks are preserved. This is stacked on PR #154 and remains separate from the frozen PR #150 checkpoint.

Validation on the exact committed bytes:

- All 21 local validation stages passed: typecheck, format, lint, fresh build, 25 selected source tests, package/consumer verification, and three separately audited installed callbacks.
- The new source HTTP callback completed 37 requests in 686 ms. Installed callbacks completed 4 + 3 + 11 requests against fresh packed artifacts; all three children exited cleanly and stores were removed.
- Independently checked 321 archive members, the complete consumer inventory and payload identity. All 679 committed files match the validated dirty source exactly at `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385`.
- Native and actual Claude code review completed; independent reversal verifies the subsequent two-test formatting correction. The earlier formatting failure remains preserved. [Exact-head CI39001388411](https://github.com/gastownhall/bdp/actions/runs/39001388411) passed all 27 steps: 9,387 tests across 186 files (one skipped), seven preflight checks, package smoke and the historical Read evidence gate.

Foreign-context opaque minima, Link Event minima, after-bearing reads, SSE/reconnect, incompatible media, S1/G2 decisions and public discovery remain outside this slice. This does not advertise complete write-profile support or authorize main integration or release.
