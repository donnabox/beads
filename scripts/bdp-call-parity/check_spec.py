#!/usr/bin/env python3
"""Fail when internal/httpapi/callables_table.go drifts from the Issue Callables document.

    ./scripts/bdp-call-parity/check_spec.py --spec PATH_TO_BDP_CHECKOUT

Compares each Callable's name, kind (CallableQuery or CallableBatch) and CLI mapping
against the declarations in specs/types/issue/callables.md, in both directions, as the
reference client's `callables:check` does for its catalog. It also holds the document's
own Names table to its declarations. Parameters are held to openapi.v0.yaml by
TestCallableTableMatchesTheDocument, not here. Exits 1 on any difference and 2 if it
cannot read either side; it reads text only, so it needs no build.
"""
import argparse
import json
import re
import sys
from pathlib import Path

DOCUMENT = "specs/types/issue/callables.md"
TABLE = "internal/httpapi/callables_table.go"
HEADING = re.compile(r"^#{3,4} ([\w./-]+) \((CallableQuery|CallableBatch)\)\s*$", re.M)
SECTION_END = re.compile(r"^#{1,3} ", re.M)
CLI_MAPPING = re.compile(r"^CLI mapping:\s*\n+```json\n(.*?)\n```\s*$", re.M | re.S)
NAMES_ROW = re.compile(r"^\| `([\w./-]+)` \| (CallableQuery|CallableBatch) \|", re.M)


class Unreadable(Exception):
    pass


def declarations(text):
    """name -> (kind, cli mapping or None), from each section of the document."""
    heads = list(HEADING.finditer(text))
    if not heads:
        raise Unreadable(f"{DOCUMENT}: no Callable declarations")
    out = {}
    for i, head in enumerate(heads):
        name, kind = head.group(1), head.group(2)
        end = heads[i + 1].start() if i + 1 < len(heads) else len(text)
        body = text[head.end() : end]
        stop = SECTION_END.search(body)
        if stop:
            body = body[: stop.start()]
        if name in out:
            raise Unreadable(f"{DOCUMENT}: {name} is declared twice")
        out[name] = (kind, cli_mapping(name, body))
    return out


def cli_mapping(name, body):
    if "CLI mapping:" not in body:
        return None
    match = CLI_MAPPING.search(body)
    if match is None:
        raise Unreadable(f"{DOCUMENT}: {name}: unreadable CLI mapping")
    try:
        return json.loads(match.group(1))
    except ValueError as err:
        raise Unreadable(f"{DOCUMENT}: {name}: CLI mapping is not JSON: {err}") from err


def names_table(text):
    return {m.group(1): m.group(2) for m in NAMES_ROW.finditer(text)}


def table_rows(source):
    """name -> (kind, cli mapping or None), from the literal `callableTable`."""
    start = source.find("var callableTable = []callableRow{")
    if start < 0:
        raise Unreadable(f"{TABLE}: no callableTable")
    i = source.index("{", source.index("[]callableRow{", start) + len("[]callableRow"))
    rows, depth, row_start = [], 0, None
    while i < len(source):
        c = source[i]
        if c in '"`':
            i = end_of_string(source, i)
            continue
        if source.startswith("//", i):
            i = source.index("\n", i)
            continue
        if c == "{":
            depth += 1
            if depth == 2:
                row_start = i
        elif c == "}":
            depth -= 1
            if depth == 1:
                rows.append(source[row_start : i + 1])
            elif depth == 0:
                break
        i += 1
    else:
        raise Unreadable(f"{TABLE}: callableTable does not close")
    out = {}
    for row in rows:
        name = re.search(r'\bname:\s*"([^"]+)"', row)
        if name is None:
            raise Unreadable(f"{TABLE}: a row has no name: {row[:80]}")
        if name.group(1) in out:
            raise Unreadable(f"{TABLE}: {name.group(1)} has two rows")
        out[name.group(1)] = (
            "CallableBatch" if re.search(r"\bbatch:\s*true\b", row) else "CallableQuery",
            row_cli(name.group(1), row),
        )
    return out


def end_of_string(source, i):
    quote = source[i]
    j = i + 1
    while source[j] != quote:
        j += 2 if quote == '"' and source[j] == "\\" else 1
    return j + 1


def row_cli(name, row):
    match = re.search(r"\bcli:\s*`([^`]*)`", row)
    if match is None:
        if re.search(r"\bcli:", row):
            raise Unreadable(f"{TABLE}: {name}: cli is not one raw string")
        return None
    try:
        return json.loads(match.group(1))
    except ValueError as err:
        raise Unreadable(f"{TABLE}: {name}: cli is not JSON: {err}") from err


def differences(declared, listed, table):
    out = []
    for name in sorted(declared.keys() - table.keys()):
        out.append(f"{name}: declared in {DOCUMENT}, no row in {TABLE}")
    for name in sorted(table.keys() - declared.keys()):
        out.append(f"{name}: row in {TABLE}, not declared in {DOCUMENT}")
    for name in sorted(declared.keys() & table.keys()):
        spec_kind, spec_cli = declared[name]
        go_kind, go_cli = table[name]
        if spec_kind != go_kind:
            out.append(f"{name}: kind: spec {spec_kind}, table {go_kind}")
        if spec_cli != go_cli:
            out.append(
                f"{name}: cli:\n  spec  {json.dumps(spec_cli, sort_keys=True)}"
                f"\n  table {json.dumps(go_cli, sort_keys=True)}"
            )
    for name in sorted(listed.keys() ^ declared.keys()):
        where = "the Names table" if name in listed else "a declaration"
        out.append(f"{name}: spec is inconsistent: only {where} has it")
    for name in sorted(listed.keys() & declared.keys()):
        if listed[name] != declared[name][0]:
            out.append(
                f"{name}: spec is inconsistent: Names table says {listed[name]}, "
                f"its declaration {declared[name][0]}"
            )
    return out


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--spec", required=True, help="a checkout of gastownhall/bdp")
    parser.add_argument("--repo", default=str(Path(__file__).resolve().parents[2]), help=argparse.SUPPRESS)
    args = parser.parse_args()
    try:
        text = (Path(args.spec) / DOCUMENT).read_text(encoding="utf-8")
        source = (Path(args.repo) / TABLE).read_text(encoding="utf-8")
        declared = declarations(text)
        table = table_rows(source)
    except (OSError, Unreadable) as err:
        print(f"check_spec: {err}", file=sys.stderr)
        return 2
    found = differences(declared, names_table(text), table)
    for line in found:
        print(line)
    if found:
        print(f"check_spec: {len(found)} difference(s)", file=sys.stderr)
        return 1
    print(f"check_spec: {len(table)} Callables match {DOCUMENT}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
