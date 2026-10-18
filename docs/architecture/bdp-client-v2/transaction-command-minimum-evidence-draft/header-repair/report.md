# Narrow header and replay-mode repair

Only the app file changed in two hunks. The hook now compares the authored header prefix exactly (including authored Accept) and rejects every bdp- field in the generated/credential suffix. Existing runner assertions retain full serializer defaults/order/value verification. The replay pair use now explicitly rejects modes other than whole-json with the existing named ledger error. Full Resource/Event equality, schema validation, counts, controls and the 63-second bound remain unchanged.

Previous source is preserved as before.ts. The other four source hashes match the frozen draft. Targeted pinned Biome formatting reported no changes. No tests, build, install, services, commit or push ran. Validation preparation still pins the prior app and requires root-directed rebinding before release.
