# Principal review repair — frozen, runtime validation not run

Only F1/F7/F9 are changed in the two owned files. The complete before files and integrated dirty diff are retained in before/; original author/review/static artifacts are untouched. Parent is f5511718371a6935614eae6756e796cef71570c2.

F1: only --outputFile and --outputFile= arguments are consumed. Other arguments continue untouched. Missing/empty values and a following option instead of a separated value fail explicitly. Nonempty argv paths pass through path.resolve before existing outsideStore containment; the environment output must still already be absolute. Per-reporter/config-file output discovery is unchanged and remains an acknowledged limitation.

F7: the schema-map callback returns an as-const tuple. F9: remove one blank line and cite full original author18104a03b850553641f24fe83e7318493fa102ca separately from actual integrationf5511718371a6935614eae6756e796cef71570c2. Historical draft-checkpoint evidence is not relabeled.

One targeted pinned Node/Biome format invocation exited0; argv and output are retained. No test, typecheck, build, service, package, commit or push ran. Request bodies/order, counters,7positive/2expected-negative requests,14combined maximum, two children and deadlines are unchanged. Other findings are untouched pending root dispositions.
