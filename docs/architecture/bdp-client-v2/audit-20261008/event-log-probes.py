"""Small evidence probes over pinned, illustrative BDP fixtures; not provider conformance."""
import copy
import json
import pathlib
import unittest

ROOT = pathlib.Path(__file__).resolve().parent
FIX = ROOT / "bdp/fixtures/transactional"
def read(name):
    return json.loads((FIX / (name + ".json")).read_text())
BATCH = read("batch")["exchanges"][0]["request"]
FEED = read("changefeed")
GROUP = FEED["exchanges"][0]["response"]["body"]["groups"][0]
SNAP = read("snapshots")["exchanges"][0]["response"]["body"]
SCHEMA = json.loads((ROOT / "bdp/schemas/bdp-v0.schema.json").read_text())["$defs"]

def snapshot_state():
    return {r["id"]: copy.deepcopy(r) for kind in ("beads", "links") for r in SNAP[kind]["items"]}

def apply_fixture_events(state, events):
    """Only this fixture's transitions, with explicit external Type knowledge.

    Not a generic BDP client: the checked-in fixture has one replace and owned creation.
    """
    for e in events:
        d = e["data"]
        if e["type"] == "created":
            record = {"id": e["subject"], "type": e["subjectType"], **copy.deepcopy(d)}
            if e["subjectType"] == "https://work.example/types/decision":
                record["ownedLinks"] = {"https://work.example/types/cites": []}
            state[e["subject"]] = record
        elif e["type"] == "updated":
            record = state[e["subject"]]
            if record["revision"] != d["previousRevision"]:
                raise ValueError("revision gap: reread or resnapshot required")
            if "ownedLink" in d:
                delta = d["ownedLink"]
                assert delta["operation"] == "created"
                link = copy.deepcopy(delta["link"])
                record["ownedLinks"][link["type"]].append(link)
            else:
                for op in d["change"]:
                    assert op["op"] == "replace" and op["path"] == "/status"
                    record["properties"]["status"] = op["value"]
            record["revision"] = d["revision"]
            for member in ("attribution", "changeContext"):
                if member in d:
                    record[member] = copy.deepcopy(d[member])
                else:
                    record.pop(member, None)
        else:
            assert e["type"] == "linked"  # derived fact, not another write
    return state

class Evidence(unittest.TestCase):
    def test_three_batch_operations_generate_six_events_three_final_postimages(self):
        self.assertEqual(len(BATCH["body"]["operations"]), 3)
        self.assertEqual(GROUP["eventCount"], 6)
        self.assertEqual(len(GROUP["events"]), 6)
        self.assertEqual(len(GROUP["changes"]), 3)
        self.assertEqual([e["ordinal"] for e in GROUP["events"]], list(range(6)))

    def test_snapshot_plus_changegroup_equals_seeded_event_reconstruction(self):
        self.assertEqual(SNAP["scopePosition"], GROUP["previousPosition"])
        direct = snapshot_state()
        for change in GROUP["changes"]:
            self.assertEqual(change["operation"], "upsert")
            direct[change["resource"]["id"]] = change["resource"]
        self.assertEqual(apply_fixture_events(snapshot_state(), GROUP["events"]), direct)

    def test_event_reconstruction_requires_previous_revision(self):
        state = snapshot_state()
        state["https://beads.example/acme/beads/task-42"]["revision"] = "wrong-prior"
        with self.assertRaisesRegex(ValueError, "revision gap"):
            apply_fixture_events(state, GROUP["events"])

    def test_owned_creation_event_lacks_authoritative_empty_owned_type_keys(self):
        event = GROUP["events"][0]
        self.assertNotIn("ownedLinks", event["data"])
        self.assertNotIn("ownedLinks", SCHEMA["createdData"]["properties"])
        self.assertFalse(SCHEMA["createdData"]["additionalProperties"])

    def test_erasure_can_exist_without_any_event_or_live_state_change(self):
        group = next(g["body"] for g in FEED["groupExamples"]
                     if g["id"] == "after-ckpt-43-historical-erasure-committed-group")
        self.assertGreater(len(group["erasures"]), 0)
        self.assertEqual(group["events"], [])
        self.assertEqual(group["changes"], [])

    def test_event_envelope_cannot_recover_original_batch_input(self):
        allowed = set(SCHEMA["event"]["properties"])
        self.assertTrue({"idempotencyKey", "operationIndex", "operationName", "expectedRevision", "name"}.isdisjoint(allowed))
        self.assertFalse(SCHEMA["event"]["additionalProperties"])
        self.assertIn("idempotency-key", BATCH["headers"])
        self.assertIn("name", BATCH["body"]["operations"][0])
        self.assertIn("expectedRevision", BATCH["body"]["operations"][2])

    def test_deletion_has_no_before_image_and_context_has_no_actor_attestation(self):
        self.assertEqual(set(SCHEMA["deletedData"]["properties"]), {"revision"})
        self.assertFalse(SCHEMA["deletedData"]["additionalProperties"])
        self.assertIn("changeContext", SCHEMA["createdData"]["properties"])
        self.assertIn("changeContext", SCHEMA["updatedData"]["properties"])

    def test_bead_projection_legitimately_has_nonconsecutive_ordinals(self):
        events = read("events")["exchanges"][0]["response"]["body"]["events"]
        self.assertEqual([e["ordinal"] for e in events], [0, 2, 4])
        self.assertEqual(len({e["transaction"] for e in events}), 1)

if __name__ == "__main__":
    (ROOT / "event-log-worked-example.json").write_text(json.dumps({
        "notice": "Pinned illustrative fixtures, not runtime evidence; see event-log-audit.md",
        "snapshot": SNAP, "batchRequest": BATCH, "changeGroup": GROUP,
    }, indent=2) + "\n")
    unittest.main(verbosity=2)
