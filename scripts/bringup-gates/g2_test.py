#!/usr/bin/env python3
"""The gate's judgement, with no binaries: what fails, what is accepted, what cannot pass."""

import json
import os
import sys
import unittest

sys.dont_write_bytecode = True
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import g2_cli_surface as g2  # noqa: E402

HELP = """Serve things.

Usage:
  bd serve [flags]
  bd serve [command]

Aliases:
  serve, s

Examples:
  bd serve --addr x

Available Commands:
  start       Start it
  stop        Stop it

Custom Group:
  other       Other

Flags:
      --addr string   Where to listen (default "x")
  -h, --help          help for serve
      --quiet-mode    No type here

Global Flags:
      --actor string   Who
"""

LINE = "      --protocol string   What it answers"
EXCEPTIONS = {"accepted": [
    {"id": "protocol-flag", "reason": "accepted",
     "help_lines": {"bd serve": [LINE]}, "flags": {"bd serve": ["--protocol"]},
     "golden_lines": {"a/tree.golden": ["  flag --protocol"]}},
    {"id": "query-order", "reason": "map order", "query_order_commands": ["list --ready", "children *"]},
]}


def walk(**paths):
    return {path.replace("_", " "): {"exit": 0, "stdout": text, "stderr": ""} for path, text in paths.items()}


def record(key, argv=None, **over):
    base = {"key": key, "label": key, "argv": argv or key.split(), "phase": "read", "mode": "json", "exit": 0,
            "stdout": "out\n", "stderr": "", "requests": []}
    base.update(over)
    return base


def exchange(target="/v0/x?a=1&a=2", header="bd", status=200, body="{}"):
    return {"request": {"method": "GET", "target": target, "version": "HTTP/1.1", "headers": [["User-Agent", header]], "body": ""},
            "response": {"status": status, "headers": [["Content-Type", "application/json"]], "body": body}}


class Parsing(unittest.TestCase):
    def test_subcommands_come_from_command_sections_only(self):
        self.assertEqual(g2.parse_subcommands(HELP), ["start", "stop", "other"])

    def test_a_leaf_lists_nothing(self):
        self.assertEqual(g2.parse_subcommands("Long.\n\nUsage:\n  bd show [flags]\n\nFlags:\n  -h, --help   help\n"), [])

    def test_flags_keep_scope_shorthand_type_and_usage(self):
        flags = g2.parse_flags(HELP)
        self.assertEqual(flags["--addr"], {"scope": "Flags", "shorthand": "", "type": "string", "usage": 'Where to listen (default "x")'})
        self.assertEqual(flags["--help"]["shorthand"], "h")
        self.assertEqual(flags["--quiet-mode"]["type"], "")
        self.assertEqual(flags["--actor"]["scope"], "Global Flags")


class Surface(unittest.TestCase):
    def compare(self, base, new):
        return g2.compare_surface(base, new, EXCEPTIONS)

    def ids(self, base, new):
        return [f["id"] for f in self.compare(base, new)[0]]

    def test_identical_is_clean(self):
        found, accepted = self.compare(walk(bd=HELP), walk(bd=HELP))
        self.assertEqual((found, dict(accepted)), ([], {}))

    def test_the_protocol_flag_is_accepted_on_serve(self):
        new = HELP.replace("Global Flags:", LINE + "\n\nGlobal Flags:", 1).replace("\n\n" + LINE, "\n" + LINE, 1)
        found, accepted = self.compare(walk(bd_serve=HELP), walk(bd_serve=new))
        self.assertEqual(found, [])
        self.assertEqual(accepted["flag, bd serve --protocol"], 1)

    def test_the_same_flag_elsewhere_fails(self):
        new = HELP.replace("Flags:\n", "Flags:\n" + LINE + "\n", 1)
        self.assertEqual(self.ids(walk(bd_list=HELP), walk(bd_list=new)),
                         ["surface:help:bd list", "surface:flag:bd list --protocol"])

    def test_the_protocol_flag_with_another_default_fails(self):
        changed = LINE + ' (default "bdp")'
        new = HELP.replace("Flags:\n", "Flags:\n" + changed + "\n", 1)
        self.assertIn("surface:help:bd serve", self.ids(walk(bd_serve=HELP), walk(bd_serve=new)))

    def test_a_new_flag_a_removed_flag_and_a_new_word(self):
        added = HELP.replace("Flags:\n", "Flags:\n      --fresh   New\n", 1)
        self.assertIn("surface:flag:bd x --fresh", self.ids(walk(bd_x=HELP), walk(bd_x=added)))
        self.assertIn("surface:flag:bd x --fresh", self.ids(walk(bd_x=added), walk(bd_x=HELP)))
        reworded = HELP.replace("Where to listen", "Where it listens")
        ids = self.ids(walk(bd_x=HELP), walk(bd_x=reworded))
        self.assertEqual(ids, ["surface:help:bd x", "surface:flag:bd x --addr"])

    def test_commands_added_and_removed(self):
        self.assertEqual(self.ids(walk(bd=HELP), walk(bd=HELP, bd_new=HELP)), ["surface:command:bd new"])
        self.assertEqual(self.ids(walk(bd=HELP, bd_old=HELP), walk(bd=HELP)), ["surface:command:bd old"])

    def test_an_exit_change_is_named(self):
        base, new = walk(bd=HELP), walk(bd=HELP)
        new["bd"]["exit"] = 1
        self.assertEqual(self.compare(base, new)[0][0]["fields"], ["exit: 0 != 1"])


class Goldens(unittest.TestCase):
    DIFF = "\n".join(["diff --git a/a/tree.golden b/a/tree.golden", "index 1..2 100644", "--- a/a/tree.golden", "+++ b/a/tree.golden",
                      "@@ -1,0 +2 @@", "+  flag --protocol"])

    def test_a_named_line_is_accepted(self):
        found, accepted = g2.compare_goldens(self.DIFF, EXCEPTIONS)
        self.assertEqual(found, [])
        self.assertEqual(accepted["golden line, a/tree.golden"], 1)

    def test_any_other_line_fails_and_is_named(self):
        found, _ = g2.compare_goldens(self.DIFF + "\n+  flag --hidden-one\n", EXCEPTIONS)
        self.assertEqual(found[0]["id"], "golden:a/tree.golden")
        self.assertEqual(found[0]["fields"], ["+  flag --hidden-one"])

    def test_a_changed_corpus_file_fails(self):
        diff = "diff --git a/corpus/x.json b/corpus/x.json\n@@ -1 +1 @@\n-1\n+2\n"
        self.assertEqual(g2.compare_goldens(diff, EXCEPTIONS)[0][0]["id"], "golden:corpus/x.json")


class Client(unittest.TestCase):
    def judge(self, base, new, results, expected=None):
        return g2.judge_client(base, new, results, EXCEPTIONS, expected)

    def test_same_passes(self):
        one = record("json list")
        self.assertEqual(self.judge([one], [one], [{"key": "json list", "verdict": "same"}])[0], [])

    def test_query_order_only_on_the_commands_the_file_names(self):
        for argv, passes in ((["list", "--ready"], True), (["children", "sw-fix-r4"], True), (["list", "--status", "open"], False), (["ready"], False)):
            key = "json " + " ".join(argv)
            left = record(key, argv, requests=[exchange("/v0/x?a=1&a=2")])
            right = record(key, argv, requests=[exchange("/v0/x?a=2&a=1")])
            found, accepted, problems = self.judge([left], [right], [{"key": key, "verdict": "query-order"}])
            self.assertEqual(problems, [])
            self.assertEqual(not found, passes, key)
            self.assertEqual(bool(accepted), passes, key)

    def test_a_difference_names_the_command_and_field(self):
        left, right = record("json x", requests=[exchange()]), record("json x", requests=[exchange(header="bd2")])
        found, _, _ = self.judge([left], [right], [{"key": "json x", "verdict": "unexplained"}])
        self.assertEqual(found[0]["id"], "client:json x")
        self.assertIn("/requests/0/request/headers/0", found[0]["fields"][0])
        self.assertIn("User-Agent: bd2", found[0]["fields"][0])

    def test_a_seed_failure_on_both_sides_is_not_the_same(self):
        left = record("json x", seed_failed="no")
        _, _, problems = self.judge([left], [left], [{"key": "json x", "verdict": "same"}])
        self.assertIn("seed failed", problems[0])

    def test_a_timeout_on_both_sides_is_not_the_same(self):
        left = record("json x", exit=-1)
        _, _, problems = self.judge([left], [left], [{"key": "json x", "verdict": "same"}])
        self.assertIn("timed out", problems[0])

    def test_a_short_table_or_a_lost_command_cannot_pass(self):
        one, two = record("json x"), record("json y")
        self.assertIn("different commands", self.judge([one, two], [one], [])[2][0])
        self.assertIn("the table has 5", self.judge([one], [one], [], expected=5)[2][0])


class Server(unittest.TestCase):
    def test_a_different_answer_is_named(self):
        found, problems = g2.judge_server([
            {"label": "probe", "method": "GET", "target": "/x", "status": 200, "same": True},
            {"label": "probe", "method": "GET", "target": "/y", "status": 200, "same": False, "base": {"status": 200}, "new": {"status": 404}},
        ])
        self.assertEqual(problems, [])
        self.assertEqual(found[0]["id"], "server:probe GET /y")
        self.assertIn("/status: 200 != 404", found[0]["fields"][0])

    def test_write_state_is_named(self):
        found, _ = g2.judge_server([{"label": "write x", "same": False, "base": {"status": 200}, "new": {"status": 200}, "state_same": False}])
        self.assertIn("state", found[0]["fields"][0])

    def test_comparing_nothing_cannot_pass(self):
        self.assertEqual(g2.judge_server([])[1], ["the server half compared nothing"])


class Floors(unittest.TestCase):
    def test_a_run_that_touched_nothing_cannot_pass(self):
        failed = [record(f"c{i}", exit=1) for i in range(10)]
        answers = [{"label": "p", "status": 200, "same": True}]
        self.assertTrue(any("exited 0" in p for p in g2.wire_floor_problems(failed, answers)))
        silent = [record(f"c{i}") for i in range(10)]
        self.assertTrue(any("sent a request" in p for p in g2.wire_floor_problems(silent, answers)))
        full = [record(f"c{i}", requests=[exchange()]) for i in range(10)]
        self.assertTrue(any("2xx" in p for p in g2.wire_floor_problems(full, [{"label": "p", "status": 404, "same": True}] * 20)))
        self.assertEqual(g2.wire_floor_problems(full, answers), [])


class Waivers(unittest.TestCase):
    ITEMS = [g2.finding("client:json x", ["a"]), g2.finding("client:json y", ["b"]), g2.finding("surface:flag:bd x --y", ["c"])]

    def test_a_record_waives_only_what_it_names(self):
        waivers = g2.waivers_from({"engdocs/decisions/r.md": "# Why\n\nG2-waives: client:json x\n"})
        standing, waived = g2.apply_waivers(self.ITEMS, waivers)
        self.assertEqual([f["id"] for f in standing], ["client:json y", "surface:flag:bd x --y"])
        self.assertEqual(waived[0]["waived_by"], "engdocs/decisions/r.md")

    def test_a_glob_says_so(self):
        waivers = g2.waivers_from({"r.md": "G2-waives: client:*\n"})
        self.assertEqual(len(g2.apply_waivers(self.ITEMS, waivers)[1]), 2)

    def test_a_record_naming_nothing_waives_nothing(self):
        self.assertEqual(g2.waivers_from({"r.md": "# A decision\n"}), [])

    def test_could_not_compare_outranks_everything(self):
        self.assertEqual(g2.exit_code(["broken"], []), 2)
        self.assertEqual(g2.exit_code(["broken"], self.ITEMS), 2)
        self.assertEqual(g2.exit_code([], self.ITEMS), 1)
        self.assertEqual(g2.exit_code([], []), 0)


class TheFile(unittest.TestCase):
    def test_every_entry_has_an_id_and_a_reason(self):
        with open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "g2_exceptions.json")) as handle:
            data = json.load(handle)
        ids = [e["id"] for e in data["accepted"]]
        self.assertEqual(sorted(ids), sorted(set(ids)))
        self.assertTrue(all(e["reason"].strip() for e in data["accepted"]))
        protocol = next(e for e in data["accepted"] if e["id"] == "protocol-flag")
        # Exact lines, and the flag on the two commands that have it: nothing wider.
        self.assertEqual(sorted(protocol["flags"]), ["bd connect", "bd serve"])
        self.assertEqual(sorted(protocol["help_lines"]), ["bd connect", "bd serve"])


if __name__ == "__main__":
    unittest.main()
