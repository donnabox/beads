# Event witness successor 2 — frozen for independent closure

Root accepted correction plan ea807b194ac198f04a3a025ee3411f2c44353310dfc25150af8386ec44afe8a2 and separately fast-forwarded the worktree to the reviewed working integration2dfd2c6cc84a28b1dd336f4184c1a5f910ce8a96 before this authoring step. The head/tree/parents are captured in source-binding.json; no merge or Git mutation was performed here.

Implemented exactly the seven bounded corrections in the new callback:

1. The local optional pid/listenUrl properties explicitly permit undefined assignments; positive ownership checks remain unchanged.
2. The external-entry guard compares entry against realpathSync(root).
3. witnessSource now declares fingerprintScope as witness-file bytes and tracked diff only, explicitly excluding untracked inputs and installed payload provenance. No new git command, untracked rejection or provenance subsystem.
4. The second local control is named frame-event-identity in both report population and expectation; comparator and objects are unchanged.
5. abortedForCleanup marks only the locally derived finally abort. Work-signal reporting retains preexisting real deadline aborts and omits the synthetic duplicate; cleanup timing/order is unchanged.
6. One local safeFailureMessage function sanitizes JSON messages and messages of fresh thrown Error objects. JSON names are sanitized too. AggregateError no longer retains original errors, stacks, causes or custom assertion fields. Internal phase and sanitized message remain; original failure stacks are intentionally unavailable through the rethrow.
7. Readonly tuples restore inference for the validator entries and all three new header-entry maps. Runtime schema guards and allowlists are unchanged.

The entire module setup and historical callback prefix is byte-identical to successor1. The seed construction is byte-identical, retaining the same244-byte request and key. Five HTTP requests, two sequential children, both local controls and45/10/60/120 budgets are unchanged. No production or shared helper files were edited. Only the owned test file is dirty; no untracked repository files.

The existing pinned Biome native binary formatted only the owned file, exit0, under the session guard; argv/hash and output are retained. The correction diff was then read against the accepted plan. No tests, typecheck, build, product import, installation, service, commit or push was run. This is a frozen authored correction, not an independent review verdict or execution pass. Actual typecheck remains required.

The predecessor source/full patch, new frozen source/full patch and correction patch are retained separately. Completed Claude review, successor1 and original validation preparation remain unchanged. The validation adapter still pins successor1 and its old report/control shape; it must be separately rebound and reviewed before execution. No helper modifications were made here.

Source SHA256: `2e141596766a7f0f9e18d8e09e1a611cad1faf1b793b2ecb1c80a0b823ca7082`

Full patch SHA256: `78523a36550e96b119737fc986569cbeb1fe5eea0ae987888fb8dd7014bedade`

Correction patch SHA256: `1ad1d55fb62018ea20daf2a479eeb187790cd5bbc041ba01d7fb1da5fc627279`
