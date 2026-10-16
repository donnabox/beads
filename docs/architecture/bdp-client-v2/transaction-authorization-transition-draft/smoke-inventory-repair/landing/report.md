# Exact PR139 landing preparation

Actual head27b4f963, tree25e9d4aa, ordinaryparent87911dce, base6f5e607. All18sourcehashes and ordinary ancestry27b4→879→1647[dae39,6f5], dae39→67a independently checked locally; currentPRhead/base andremote main match. NewCI38843452920/job116788531347 is active. No terminal result exists yet.

Capture helper refuses partial logs, requires exact run/job/head/attempt and extracts actual checkout synthetic from rawlog, then verifies API tree/orderedparents and workflow hash. Root-owned final council acceptance is separately required; original council pointer is attributed only to original feature. Both prior failed CI attempts remain immutable.

Read-only postmerge verifier is adapted fromPR138. Root must first obtain actual mergeSHA, fetch it itself, and supply released root-postmerge-release.json binding successful terminalcapture, finalcouncil and script. Verifier performs GETs/localGitreads only, verifies fullAPI/localtree,18filehashes and ordinarycandidateancestry; no merge/fetch/push/test. No future mergeSHA or CI result is fabricated. Existing1647 installed attribution stays historical; the four-line smoke repair does not alter reviewed product/package bytes.
