# Prospective finite HTTP requests

All tests begin with the configured Scope URL, fixture credential handles, transport host/port and the explicit client-known canonical installed Type/contract ID used by row4. That Type identifier is fixture input, distinct from an operation-target URL oracle; complete discovered Types navigation remains D3a evidence. IPC carries no operation targets, Directory map, receipt URLs or capability. Route selection starts with actual Scope Link, actual returned descriptor, then its actual operations URI. Each credential is resolved inside the executor; observation records and diagnostics never contain tokens, fileDigest or authenticationIdentity.

| Cold ordinal | Method and URL source | Credential and payload | Required afterstate |
| --- | --- | --- | --- |
| 1 | GET configured Scope | writer | successful Scope response and exactly parsed service-desc Link |
| 2 | GET actual Link target | writer | valid profile descriptor, intended canonical Scope, actual Read limits |
| 3 | GET descriptor.operations | writer | actual valid Directory; store target strings from this response only |
| 4 | GET configured installed Type URI | writer | descriptor matches explicit supplied contract; this URI is fixture input, not an operation target oracle |
| 5 | POST Directory.createBead | writer; Idempotency-Key startup-created; JSON id beads/startup, Type from row4, properties n:1 | RU singleton created Resource or TX completed receipt; exactly one created result, issued Resource ID and revision |
| 6 | GET issued Resource ID | writer | exact created Resource body/revision |
| 7 | POST same actual target | same key and exact row5 payload | byte/semantic replay law from existing profile; same Resource/revision; no second create |
| 8 | GET issued Resource ID | reader | same Resource with read-only policy |
| 9 | POST same actual target | reader; distinct key reader-refused; id beads/refused, same Type and n:1 | 403; prior Resource unchanged; verify absent new Resource through later owner readback, without adding an uncounted HTTP request |
| 10 | GET issued Resource ID | anonymous/invalid bearer | 401, no discovery/context leak |
| 11 | HEAD actual descriptor URI | writer | same status/representation headers law, zero body |
| 12 | HEAD actual Directory URI | writer | valid HEAD law, zero body |
| 13 | GET configured absent beads/not-created | writer | actual resource-not-found response; no operation target inference |

Each cold callback has exactly13 planned dispatches. RU uses existing singleton parser; TX uses actual Transactional receipt parser/schema and results[0].resource. No new carrier shape or wire semantics is proposed. Default n:integer schema and writer/reader mapping come from the real deployment file and existing installed captures. All response assertions follow actual observation capture.

Reopen callbacks I03/I04 use generation1 cold navigation rows1-3 plus create/read (5), close and observe exit; generation2 cold1-3 plus read and same-key replay (5). The same canonical Scope, Types, contracts, limits, directory and intended listener are used; only explicit create:false differs. Opaque Scope epoch follows existing profile law, never assumed equal across reopen. Compare retained Resource/revision and actual prior disposition, not opaque token spelling.

RU current-policy I05 uses generation1 cold1-3 plus create (4), then generation2 cold1-3 with new current writer credential, old-token401, new-token Resource read200 and reader mutation403 (6). Store installation/contracts remain equal. Compare actual current authority policy behavior; do not require stored credential or policy identity equality.

Each stop callback I12-I17 uses cold1-3/create/read (5), signal or disconnect, actual exit and listener release, then create:false child cold1-3/read (4). Each peer callback I18/I19 uses cold1-3/create (4), a separately counted non-reading raw CONNECT peer that has actually received refusal bytes, close, then reopened cold1-3/read (4). Record actual peer state/close chronology before shutdown; do not silently substitute a disconnected socket. If the original peer already closed normally after refusal, report already closed and retain only refusal/cleanstop/reopen evidence. Claim shutdown-forced destruction only for observed live owned transport and observed closure; refusal bytes or withheld FIN alone do not establish it. Observe RU and TX independently; no new seam, sleep, retry or transport change is required.

I06-I11 and I20/I21 schedule no HTTP requests. Occupied-port witnesses hold an independently acquired blocker and prove the candidate never successfully acquired that listener; they cannot prove zero listen attempts. Public-refusal witnesses observe the real public executable's refusal and unchanged absent target store, with no private ready message or new public capability. Failure callbacks have no forged readiness or successful-completion counts.
