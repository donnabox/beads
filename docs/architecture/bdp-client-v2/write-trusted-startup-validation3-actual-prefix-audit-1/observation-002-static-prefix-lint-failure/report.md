# Startup validation3 static-prefix failure audit

Seven stages actually ran: first4 plus typecheck and format-check passed; lint exited1 and was rejected. Typecheck log91d54ffb confirms tsc invocation. Format-check log4d442fb5 reports11 files checked/no fixes. Lint logf8cec426 contains4 hard noControlCharactersInRegex errors at config1087 and loader105,20 warnings and1 info;5 diagnostics were suppressed. Fresh-build, bind-fresh-build and all later stages remain unattempted, and the gate is closed.

All47 sampled PIDs and7 groups are independently absent in the recorded census. Exits show no forced cleanup, lifecycle-recording errors, runtime ownership residue or source drift during execution. Scratch/cache retention remains distinct from runtime ownership. The241 frozen preparation entries remain preserved with the sole active-gate exception.

Actual stages used e577/fa80 source682. After terminal failure root changed only two regex lines in a separate frozen correction. Current682 bytes match new map ee57c9f2; old682 are reconstructed exactly from680 current files and two frozen before copies. This does not claim current source still equals old e577 or accept the correction as validated.

First4 starts used b624d976; typecheck/format/lint starts used released029bef32 with9 approved stages. Observation001 captured closed9 snapshot9115324c, differing solely in released:false. Earlier messages calling it first4 or identical to released9 were incorrect; exact attribution is preserved separately without rewriting original observations. Root diagnosis attributes closure to actual lint failure.

Prior validation2 failure remains preserved. No build/package/consumer/runtime qualification or publication follows from this failed prefix. Root owns any separately frozen correction and fresh validation. Auditor actions were data reads/hash checks, the role guard/process census and new records; no product/helper/Node/compiler/Git/source/gate execution.
