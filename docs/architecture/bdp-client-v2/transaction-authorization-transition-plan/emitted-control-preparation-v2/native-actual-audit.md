# Independent actual C1 audit

No actionable discrepancy found. The actual run supports the narrowly stated expected-negative result; it is not installed validation or a complete postpublication stream-revocation proof.

The raw Vitest JSON has exactly one selected failed callback, `narrows current Transactional visibility and fences issued handles after policy publication`, one failure message and four unselected callbacks. The sole failure is `authorization-transition:generation-and-lease` at transaction-runtime.test.ts:655. Actual generation0/all-visible/canMutate=true/unchanged-view contrasts with required generation1/empty/false/changed-view. This is a genuine assertion failure, not a fabricated report. Raw test and wrapped source-supervisor exits remain1, elapsed1.995873917s; no timeout, forced cleanup or supervisor errors. The separate adapter's accepted adjudication is consistent with its read source; adapter0 was reported by root, while the retained raw exit is independently verified here as1.

Actual trace contains six complete finite200 responses (two POST then Resource, collection, Snapshot and Event GETs) plus one open Changes200 stream, in exact indices1–7. Both seed body bytes/digests match the selected source recipe (3728 and132 bytes). Initial pos-1 advances to pos-2, with unchanged epoch/view thereafter. The returned publication object is the coherent old generation0 policy and the stream is open at that observation. The assertion stops before later requests. Therefore this demonstrates sensitivity to suppressing the policy publication as a whole; it does not independently distinguish lease-only revocation failure or execute the remaining ten positive requests. Full finite-body/schema equality was performed by the selected test before the failure; retained fingerprints cannot independently reconstruct those bodies.

Read the complete stdout/stderr, machine results, start/exit/adjudication and adapter. The stderr Python CalledProcessError is the guard surfacing the expected raw exit1, not another test failure. Cleanup keeps the original single failure count, reports runtime close and store removal, and does not claim public eligibility.

Independently rehashed all15 current source files against equal before/after records; all1143 positive/copy source-build members and1932 tool files match their bound inventories. Exactly one positive/copy product file differs: emitted transaction-store.js, original29072fccbd53dedac651670984a63f4daf3706581a05c849eaea72e942ba7823 → derivative8f4a02cbe33ef6c613676f37d83079e6416ef68bc3a8d5163811eff39174ab55. Read the entire one-method patch: replacement returns current authorization through enter, suppressing persistence/view change/revocation together. The test loads the mirror emitted runtime and its relative store dependency. All180 dependency links were checked for exact original/copy link targets and resolved paths, and every copy resolution stays within the mirror. No positive-source mutation is claimed or observed.

Independent current ps snapshot found sampled PIDs6239,6241,6247,6254 and process group6239 absent. The attempt temp directory is empty and the actual recorded bdp-authorization-http-Gs3iHV store is absent. This corroborates the terminal records without claiming exhaustive detection of unobserved short-lived/detached descendants. No rerun, import, service or product mutation was performed during this audit.

Source attribution remains dirty base67a2842e plus final-v4 diffcc03010d7c9390cb3034eaf18eebd9b2ac5dd3522aa72e998e991b5d5e20e189, not a child-attested commit. I reviewed the authorization implementation/repairs but did not author or execute this control. Root retains acceptance.

Exact audited run hashes:
- `adjudication.json`: `0a0eca586c646ee595474fd8d700e752053239518eb46b6e2db48d5c3a681f0f`
- `suppress-policy-publication.exit.json`: `2fada20428149f0701c1ae949be9b53d6ff4cd44f6c4d81a66cce894d65ad696`
- `control.results.json`: `174bf40f2b073d7009b7742cc003e2ad0d163a1e39a324a57fa6c87b30624428`
- `suppress-policy-publication.log`: `9391dcccb5669cb3727b00fde9dba041be9663953a164f829b30978e12e5dffd`
- `suppress-policy-publication.stderr.log`: `09c28ccbe787c7bacead2044ac00c18a1026276ff74ad9676f3b4405273406b1`
