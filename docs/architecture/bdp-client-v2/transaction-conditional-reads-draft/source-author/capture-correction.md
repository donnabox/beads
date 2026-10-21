# Accepted authoring correction: HTTP captures are URL-only

The reviewed plan proposed two raw POST captures at `/results/0/resource`, then two GET `equalsBinding` outcomes. Current runner.ts1471–1502 unconditionally calls captureValue for each HTTP capture;2468–2495 accepts only a scoped absolute HTTP(S) URL string. Generic JSON value capture is instead the client-action seam1126–1133. The original plan assumption was incorrect. No product execution was needed to establish it.

Root explicitly accepted removal of both object captures and both equalsBinding assertions within the four owned files. The independent app observer still validates each full actual initial/update receipt and stores its Resource; each actual GET is schema-validated and compared by strict full equality to that actual postimage. Canonical manifest schemas and independent literal fields remain. Full dynamic equality is now explicitly app evidence, not a manifest outcome. No runner change, client read, new HTTP, or new product decision.

Assertion N changes200→198 and canonical total2842→2840. Positive14/control4/four-child expectation is unchanged. Before-correction draft bytes are retained under before-capture-correction; these were never executed or an approved frozen source.
