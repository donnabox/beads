"""Offline candidate-schema oracle; no runtime or installation claim.

Run: uv run --no-project --with 'jsonschema[format]==4.25.1' python \
    internal/memorytypes/check_schema_oracle.py
Only Python dependency acquisition may use the network. Schema references cannot.
"""
import json
from pathlib import Path
from jsonschema import Draft202012Validator, FormatChecker
from referencing import Registry
from referencing.exceptions import NoSuchResource

ROOT = Path(__file__).resolve().parent / "testdata"
NAMES = {"memory": "memory.properties.schema.json", "link": "memory-link.properties.schema.json"}
NAMES.update({name: f"memory-{name}.local-properties.schema.json"
              for name in ("observation", "decision", "procedure", "policy", "lesson")})


def unavailable(uri):
    raise NoSuchResource(ref=uri)


checker = FormatChecker()
# Fail rather than silently disabling an optional format dependency.
for format_name, invalid_value in (("date-time", "not-an-instant"), ("uri", "/relative")):
    assert not checker.conforms(invalid_value, format_name), f"missing {format_name} checker"
validators = {}
for name, filename in NAMES.items():
    schema = json.loads((ROOT / "candidate" / filename).read_text())
    Draft202012Validator.check_schema(schema)
    validators[name] = Draft202012Validator(schema, registry=Registry(retrieve=unavailable),
                                           format_checker=checker)
count = 0
for path in (ROOT / "candidate" / "schema-cases.json", ROOT / "additional-cases.json"):
    for case in json.loads(path.read_text())["cases"]:
        chosen = [validators[case["type"]]]
        if case["type"] not in ("memory", "link"):
            chosen.append(validators["memory"])
        errors = [e.message for validator in chosen for e in validator.iter_errors(case["properties"])]
        expected = case.get("schemaValid", case["valid"])
        assert (not errors) == expected, (case["label"], expected, errors)
        count += 1
print(f"PASS: {count} property cases, Draft 2020-12, explicit formats, offline references")
