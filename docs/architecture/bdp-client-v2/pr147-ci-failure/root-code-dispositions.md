# Root disposition: PR147 two-test correction

Actual Claude code review completed exit0 in659.843s, no timeout, inputs/packet unchanged, no survivors. Full output SHA605e850a6a478f07c68fe7195608c321b2c1f8733195279a6454e59afbe8b570. Native review found no actionable findings on the same two-file patch. Both councils found no blockers.

1. Accepted: hoist the differing deep JSON parse outside the rejection callback and require the exact primitive-mismatch error. This prevents parse failure from satisfying the negative control. This sole follow-up changes only the new helper test; source/correctness review and all five bounded checks must cover the successor.
2. Retained diagnostics remain directly present before guaranteed throws. Additional matcher-specific diagnostic-shape controls are not needed for the equality contract or this CI repair; no diagnostic removal is proposed.
3. Sparse/extended arrays are outside the finite JSON data domain of these callsites. Length hardening would change pre-existing comparator semantics; defer rather than expand this equivalence repair.
4. Date/Map/exotic objects are likewise outside this JSON domain. Null-prototype JSON records remain supported. No prototype policy change.
5. The replacement asserts successful preparation plus exact branded set operation/Selector; existing deep lifecycle tests cover semantic preservation. Further sibling/pointer assertions are optional, not required for the removed invalid refusal.
6. The frozen helper is unchanged. The claim that inverting !Object.isFrozen would leave its six actual frozen-value sites green is incorrect: inversion would reject them. Separate helper controls are optional hardening, not part of the matcher-hot-path repair.
7. Direct low-level Selector capacity coverage is a pre-existing separate follow-up. The removed expectation never reached the store. No receiving/store branch behavior changes here.

No deadline, depth, production, dependency or profile claim changes. Prior local results stay separately bound to the pre-follow-up patch. Exact successor CI remains mandatory before integration.
