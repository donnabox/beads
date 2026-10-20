# Canonical show public corpus

`TestCanonicalShowWorkflow` captures these envelopes from real `bd show` and
`bd link show` subprocesses in a disposable graph workspace. It compares their
payloads with the same BDP projector used by the protocol, then checks these
blobs. Both embedded and server storage must produce the same public shape.

Only root and inline owned-Link `revision` tokens become `REVISION`. Identities,
Reference encodings, properties, metadata, attribution, the actual BDP pin and
CLI schema version remain in the corpus. Revision normalization never descends
into user properties or metadata.

Regenerate deliberately, then review the diff:

```sh
./scripts/test.sh ./cmd/bd -run '^TestCanonicalShowWorkflow/embedded$' -count=1 -args -canonical-show-update
```

Normal runs compare without updating. The shipping Issue CLI corpus under
`cmd/bd/protocol/testdata/corpus` remains separate and unchanged.
