# RFC 8785 (JCS) test vectors

The `input/` and `expected/` pairs are copied verbatim from
[github.com/gowebpki/jcs](https://github.com/gowebpki/jcs) v1.0.1
(`testdata/input` and `testdata/output` upstream; the latter is named
`expected/` here because this repository's `.gitignore` drops any `output`
tree; Apache License 2.0, Copyright 2021 Bret Jordan & Benedikt Thoma,
Copyright 2006-2019 WebPKI.org), which carries the RFC 8785 JSON
Canonicalization Scheme test suite from
[cyberphone/json-canonicalization](https://github.com/cyberphone/json-canonicalization).
They are the same files as `internal/storage/issueops/testdata/jcs`.

Each `input/<name>.json` must canonicalize to exactly the bytes of
`expected/<name>.json`, with no trailing newline. `TestCanonicalize_RFC8785Vectors`
runs `Canonicalize` over every pair, and `TestCanonicalize_NumbersUseECMA262Formatting`
reads the number cases from `values.json`, so the comparator's canonical form is
pinned to the published vectors rather than to whatever the dependency happens
to emit. Do not edit the pairs; local cases belong in the tests' own tables.
