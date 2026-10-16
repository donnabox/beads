# Runtime prefix through consumer installation — passed, stopped at lock gate

All 13 newly released stages from fresh-build through consumer-install completed once in order, raw exit 0 and supervisor accepted. Historical default and new restart default each selected and passed their intended callback; separate audit stages accepted their evidence. Unselected callback skip is due to exact test-name selection, not missing execution: both callbacks ran separately. All stage source/cleanup checks passed without intervention or survivors.

New default witness actual observation: five HTTP requests (four completed finite responses, one SSE); two sequential owned child PIDs 48457 and 48514, both normal exit code 0/signal null; exactly one expected SSE frame after drain; pending operations 0, failures empty, temporary store removed. The pinned callback establishes clean first exit before reopening the same store. This is development-local evidence, explicitly nonqualifying. Historical observation uses its preserved earlier report shape; no fabricated per-request/child ledger is inferred for it.

Fresh build ran the configured clean then tsc build. Archive verification binds freshly packed protocol/server bytes to actual build and source. Consumer installation resolved eight packages with both BDP packages coming from these local archives and fast-uri pinned at 3.1.5. The new actual lock is available for root acceptance. Installed payload verification and external callback have NOT run; consumer installation alone establishes neither installed provenance nor runtime success.

Stopped before verify-installed-consumer as instructed. No retries, source changes, gate edits, commits or publication. Source remains b762563798d6d72afecfda0a4c8b831de69753df45c82f6050815599a24a1496 on 2dfd2c6cc84a28b1dd336f4184c1a5f910ce8a96. Root must inspect/pin the actual consumer lock, then separately release installed verification and the eventual external witness.

Artifact fingerprints:

```json
{
  "consumer/pnpm-lock.yaml": "a50da9c8894ae56f953779fdeee178ab4091f69f926fbcbd9b95ef0d2d629197",
  "build-artifacts.json": "d0ebd5f678b31915651e676dd365c624e0adf2fa474ec5c17043af3d9293112b",
  "tar-source-provenance.json": "d3edbda9a27219dbe664fdb50c3b30867b9a5d405f3b6a9b566dd4677b457f22",
  "historical-default.accepted-observation.json": "6bc190b85521410f94f1404f7728d0f81e88547b6b6079d62459ca1050020e14",
  "restart-default.accepted-observation.json": "0c90fbab2fa6634b45c3955309c75d6cd8d52154f2992d874a33d903920cc285"
}
```

Stage results:

```json
{
  "fresh-build": {
    "exit": 0,
    "elapsedSeconds": 7.779323833063245,
    "exitRecordSha256": "17785614f91c34c504cccb60f04c15244d1eef35985dfc0cde2f99ae36c241da",
    "logSha256": "232dac5e840e45b3bb06e6e04249b5241f2e1b4f6c797a22c046d916302ac208"
  },
  "bind-fresh-build": {
    "exit": 0,
    "elapsedSeconds": 12.352846125140786,
    "exitRecordSha256": "19f8411960d761c88621afdd45f16f037e2ce93ef25fcde37004edb2db803951",
    "logSha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  },
  "historical-default": {
    "exit": 0,
    "elapsedSeconds": 22.727564917178825,
    "exitRecordSha256": "1fc03b6a4a746ac8d64ce85c74b2e3eaaf68f21cb1a9eea8c3c0eee733d83d47",
    "logSha256": "bf3a5d715052395808b7a4d90e6bb882c87a054e40753f2f1408013e098dee81"
  },
  "review-historical": {
    "exit": 0,
    "elapsedSeconds": 11.794880541972816,
    "exitRecordSha256": "038ee1476fffbb5e47acd67d9ea203697480d3ab3ce06db7858ae70601c8b1ae",
    "logSha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  },
  "restart-default": {
    "exit": 0,
    "elapsedSeconds": 7.3222095421515405,
    "exitRecordSha256": "d5b33abec18047f34918dbaad6e55718b9e4dc69d62a5c162dc171fb904eaf64",
    "logSha256": "4a2561ab5eb037870d52a46b20edf4442565dc2dd1c6ff5046b66fe8a53b4e31"
  },
  "review-restart-default": {
    "exit": 0,
    "elapsedSeconds": 12.267090959008783,
    "exitRecordSha256": "b5abf31e39ffe02e3df29d424e0dd38bf6ce4ff77cbff2570991a01eeb70c64f",
    "logSha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  },
  "node-version": {
    "exit": 0,
    "elapsedSeconds": 5.720920791849494,
    "exitRecordSha256": "6cb6456d91f0e3bdabe53e132344ee388f3239fb49eb92451a2793bc1dfc11ab",
    "logSha256": "a8f8f7ba69d43285857a0ea34017ef8dadd3ba012404bb7dd188fc773cf63ee2"
  },
  "prepare-package-directory": {
    "exit": 0,
    "elapsedSeconds": 11.615816457895562,
    "exitRecordSha256": "a56f03d6f1a7b76bf8c22689ce0a7923451f403bd5840f58d48557ac6bd90944",
    "logSha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  },
  "pack-protocol": {
    "exit": 0,
    "elapsedSeconds": 5.698615999892354,
    "exitRecordSha256": "4c01abdba66b79b465e128890e49c1c05f317c70bde4eb02bd91a24046f17fc4",
    "logSha256": "adfb4ac23367209c93658bc80a3c66945cda7a28e07244992cb1c30a80c72666"
  },
  "pack-server": {
    "exit": 0,
    "elapsedSeconds": 5.6975917080417275,
    "exitRecordSha256": "9dd1a42915b868b3598cdbc8a8b09876f2a0fe437843373b9dc699e0255158a8",
    "logSha256": "a57ca400c760a8c46f9782c412d93f08a893ccd476a0bedb18fa55126e0445e5"
  },
  "verify-archives": {
    "exit": 0,
    "elapsedSeconds": 11.277082666987553,
    "exitRecordSha256": "b118050c9ed582ecf4eb032c8017f364191a3ee97dd8fed2376c286835545cb7",
    "logSha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  },
  "prepare-consumer": {
    "exit": 0,
    "elapsedSeconds": 11.172452833037823,
    "exitRecordSha256": "450d55ac1e791863682cea82c8c64066ba6ee643de8861a49e195f3443dfd123",
    "logSha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  },
  "consumer-install": {
    "exit": 0,
    "elapsedSeconds": 5.920470209093764,
    "exitRecordSha256": "6548c683c93cddb7c797bf92c6511ff64cf6c0e2d01c1cfa496d25d35d156ac9",
    "logSha256": "e051b97f41238269859ad46a3d11cf9c17cf28df2d9be597cbc5d7c0fa921680"
  }
}
```
