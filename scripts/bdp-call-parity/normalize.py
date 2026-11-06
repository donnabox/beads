"""What two answers may differ in, and nothing else.

Only values that must vary between two independent writes are masked: the
request id, timestamps, revision tokens, and issue ids the server minted. A fixture
id (sw-fix-...) was chosen by the harness and is compared literally.
"""

import json
import re

TIMESTAMP = re.compile(r"\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2}| [+-]\d{4} UTC)?")
REVISION = re.compile(r"-?\b\d{15,}\b")
MINTED_ID = re.compile(r"\bsw-(?!fix-)[a-z0-9]+(?:-[a-z0-9]+)*(?:\.\d+)*\b")
HEX_HASH = re.compile(r"\b[0-9a-v]{32}\b")
UUID = re.compile(r"\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b")
DROPPED_KEYS = {"request_id", "schema_version"}
# Timing and process facts of a CLI run, not of the data.
VARYING_KEYS = {"duration_ms", "elapsed_ms", "took_ms", "query_ms", "total_ms", "pid", "uptime", "uptime_seconds", "arrival_seq"}


class Masker:
    """Numbers minted ids in order of first appearance, so a shape still compares."""

    def __init__(self, extra=None):
        self.minted = {}
        self.extra = list(extra or [])

    def text(self, value):
        for old, new in self.extra:
            value = value.replace(old, new)
        value = TIMESTAMP.sub("<ts>", value)
        value = UUID.sub("<uuid>", value)
        value = REVISION.sub("<rev>", value)
        value = HEX_HASH.sub("<hash>", value)
        return MINTED_ID.sub(self._mint, value)

    def _mint(self, match):
        token = match.group(0)
        if token not in self.minted:
            self.minted[token] = f"<id{len(self.minted) + 1}>"
        return self.minted[token]

    def walk(self, value):
        if isinstance(value, dict):
            return {self.text(k): ("<varies>" if k in VARYING_KEYS else self.walk(v))
                    for k, v in value.items() if k not in DROPPED_KEYS}
        if isinstance(value, list):
            return [self.walk(v) for v in value]
        if isinstance(value, str):
            return self.text(value)
        return value


def unordered(value):
    """The same value with every list sorted, to tell a different order from a different answer."""
    if isinstance(value, dict):
        return {k: unordered(v) for k, v in value.items()}
    if isinstance(value, list):
        return sorted((unordered(v) for v in value), key=lambda v: json.dumps(v, sort_keys=True))
    return value


def strict(value):
    """The comparison for two answers from one server: only request_id may differ."""
    if isinstance(value, dict):
        return {k: strict(v) for k, v in value.items() if k != "request_id"}
    if isinstance(value, list):
        return [strict(v) for v in value]
    return value


def masked(value, extra=None):
    return Masker(extra).walk(value)


def diff(left, right, path="", limit=12):
    """Paths at which two JSON values differ, each with both values, abbreviated."""
    found = []

    def short(value):
        text = json.dumps(value, sort_keys=True)
        return text if len(text) <= 800 else text[:800] + "..."

    def go(a, b, at):
        if len(found) >= limit:
            return
        if isinstance(a, dict) and isinstance(b, dict):
            for key in sorted(set(a) | set(b)):
                if key not in a:
                    found.append(f"{at}/{key}: only on the right: {short(b[key])}")
                elif key not in b:
                    found.append(f"{at}/{key}: only on the left: {short(a[key])}")
                else:
                    go(a[key], b[key], f"{at}/{key}")
        elif isinstance(a, list) and isinstance(b, list):
            if len(a) != len(b):
                found.append(f"{at}: lengths {len(a)} and {len(b)}")
            for i, (x, y) in enumerate(zip(a, b)):
                go(x, y, f"{at}/{i}")
        elif a != b:
            found.append(f"{at}: {short(a)} != {short(b)}")

    go(left, right, path)
    return found
