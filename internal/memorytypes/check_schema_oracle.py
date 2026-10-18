"""Offline candidate-schema oracle; no runtime or installation claim.

Use internal/memorytypes/check.sh for the combined F0 acceptance checks.
See README.md for explicitly acquiring pinned oracle dependencies beforehand.
"""
import json
from importlib.metadata import version
from pathlib import Path
from jsonschema import Draft202012Validator, FormatChecker
from referencing import Registry
from referencing.exceptions import NoSuchResource

ROOT = Path(__file__).resolve().parent / "testdata"
NAMES = {"memory": "memory.properties.schema.json", "link": "memory-link.properties.schema.json"}
NAMES.update({name: f"memory-{name}.local-properties.schema.json"
              for name in ("observation", "decision", "procedure", "policy", "lesson", "requirement")})


def unavailable(uri):
    raise NoSuchResource(ref=uri)


for package, expected_version in {"jsonschema": "4.25.1", "rfc3987": "1.3.8", "rfc3339-validator": "0.1.4"}.items():
    assert version(package) == expected_version, f"{package} must be {expected_version}"

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
    cases = json.loads(path.read_text())["cases"]
    assert cases, f"no cases in {path}"
    for case in cases:
        chosen = [validators[case["type"]]]
        if case["type"] not in ("memory", "link"):
            chosen.append(validators["memory"])
        errors = [e.message for validator in chosen for e in validator.iter_errors(case["properties"])]
        expected = case.get("schemaValid", case["valid"])
        assert (not errors) == expected, (case["label"], expected, errors)
        count += 1
print(f"PASS: {count} property cases, Draft 2020-12, explicit formats, offline references")
