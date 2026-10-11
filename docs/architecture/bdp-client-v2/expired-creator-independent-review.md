# Independent expired-creator diagnostics review

Frozen head: `9e11bbbe315bd4896abd02c5859bf3cc7dff658c`, direct child of `b7eefc9ecad02540f9c26576220841c78f1725cb`.
Isolated clean checkout: `/Users/dbox/repos/wt/vickie/bdp-update-expired-creator-review`.
No actionable blocker found. No source edits, integration or outward writes performed.

The four-file diff adds one diagnostic scenario, its three existing normative citations, one bounded coverage mapping and documentation. All previous 28 scenarios are unchanged as parsed JSON; only the expired-creator-binding row changes classification. Runtime, fixture/helper, schema, public exports/configuration and Read seal are untouched.

The initial sequence create supplies no ID and captures its actual response Resource ID. Existing targetBinding and referenceBindings insert that captured identity into later real requests. Startup reopen at the fixture boundary invokes actual store expiry; creator retry must report 410/idempotency-expired with no Resource. A fresh dependent using the renamed @label updates the original Bead, verified by GET of the initially captured ID. A later fresh singleton changes its stage to current. Subsequent renamed sequence retries must return the dependent's older stage while GET still observes current. This distinguishes retained result projection from reexecution and from a replacement target. A further reopen repeats both observations. Successful dependent update also exercises a retained Bead-kind binding; broader wrong-kind and Link-creator cases remain explicitly outside this slice.

Independent validation with Node24.16.0/pnpm11.20.0:
- Frozen install and completed build before process tests: pass.
- 115 tests in six files: pass, 6.98 seconds (development matrix/disconnect, diagnostic inventory, normative Read+Update catalog, profile citation boundaries, sequence executor).
- Saved actual diagnostic run: 29 passing scenarios, 312 actions, 1,031 passing assertions; new plan has 10 actions and 57 passing assertions; claimEligible false. Inventory has 58 bounded rows, 23 omissions.
- Typecheck, boundaries, formatting, diff check and unchanged 89-row Read evidence verifier: pass.
- Installed-package and development process smokes: pass.

Logs, exit codes, structured tests and diagnostic report: expired-creator-independent-* alongside this report. No full suite or public conformance admission claimed. This remains one allocated Bead plus updateBead, selected postimage fields and startup expiry, not full-row qualification. Link creators, aliases, pins, deleted creators, authorization changes, internal tombstone fields and live periodic maintenance are unqualified. The private 24-hour fixture boundary is not a public retention promise; retention-minimum remains omitted.
