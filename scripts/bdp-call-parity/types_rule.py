"""The Types rule of specs/types/issue/callables.md section 2, written from the spec.

Wherever a Callable carries a Bead's Type or a dependency's kind, the member is `type`
and its value is a Type reference: the shared identities of Issue section 2, or the
Scope-local identity of a configured name. v0's `issue_type` is renamed `type`;
`exclude_type` keeps its name. A name with no identity has no reference, so the
Callable refuses it as validation-failed before v0 runs.

This is deliberately not a port of internal/httpapi/typeref: it reads the spec's table
by the shapes the v0 contract gives those members, so a mistake on either side shows
as a difference. Its case arguments are v0's, and so are the bodies it rewrites.
"""

import copy
import urllib.parse

BEAD_TYPES = "https://github.com/gastownhall/bdp/types/issue/"
LINK_TYPES = "https://github.com/gastownhall/bdp/types/dependency/"
# Issue section 2's tables.
BEAD_NAMES = {"task", "bug", "feature", "epic", "chore", "decision", "message", "molecule", "gate", "spike",
              "story", "milestone", "event"}
LINK_NAMES = {"blocks", "parent-child", "conditional-blocks", "waits-for", "related", "discovered-from",
              "replies-to", "relates-to", "duplicates", "supersedes", "authored-by", "assigned-to", "approved-by",
              "attests", "tracks", "until", "caused-by", "validates", "delegated-from"}
# The section 2 table's parameters.
BEAD_FILTERS = {"issue/ready", "issue/ready-count", "issue/list", "issue/count", "issue/claim-next"}
LINK_FILTERS = {"issue/dep/list", "issue/dep/edges", "issue/dep/count"}
# Configured names, by Scope. The harness configures none.
CONFIGURED = {"bead": set()}


class _Spelling:
    """One invocation's spelling, which records a value that names no Type."""

    def __init__(self, scope):
        self.scope = scope
        self.refused = False

    def bead(self, name):
        if name is None:
            return None
        if name in BEAD_NAMES:
            return BEAD_TYPES + name
        if name == "" or name not in CONFIGURED["bead"]:
            # No identity, so no reference: the Callable has nothing to send but this,
            # and refuses it.
            self.refused = True
        return self.scope + "types/issue/" + urllib.parse.quote(name, safe="") if name else name

    def link(self, name):
        if name is None:
            return None
        if name in LINK_NAMES:
            return LINK_TYPES + name
        if name == "":
            self.refused = True
            return name
        # Every other kind is the Scope's (Issue section 2: configured relationship kinds).
        return self.scope + "types/dependency/" + urllib.parse.quote(name, safe="")

    def each(self, value, spell):
        if isinstance(value, list):
            return [spell(item) for item in value]
        return spell(value)


def _rename_issue_type(obj, spell):
    if isinstance(obj, dict) and spell.member in obj:
        obj[spell.renamed] = spell.bead(obj.pop(spell.member))


def _edges(items, spell):
    for item in items if isinstance(items, list) else []:
        if isinstance(item, dict) and "type" in item:
            item["type"] = spell.link(item["type"])


class _Names:
    """The other way: a Callable's references read back to v0's names."""
    member, renamed = "type", "issue_type"

    def __init__(self, scope):
        self.scope = scope
        self.refused = False

    def _name(self, value, namespace, local):
        if not isinstance(value, str):
            return value
        if value.startswith(namespace):
            return value[len(namespace):]
        if value.startswith(self.scope + local):
            return urllib.parse.unquote(value[len(self.scope + local):])
        self.refused = True
        return value

    def bead(self, value):
        return self._name(value, BEAD_TYPES, "types/issue/")

    def link(self, value):
        return self._name(value, LINK_TYPES, "types/dependency/")

    each = _Spelling.each


_Spelling.member, _Spelling.renamed = "issue_type", "type"


def v0_arguments(name, args, scope):
    """A Callable's arguments as v0 takes them, for replaying them at v0."""
    return _respell(name, args, _Names(scope))


def arguments(name, args, scope):
    """A case's v0 arguments as the Callable takes them, and whether the rule refuses them."""
    return _respell(name, args, _Spelling(scope))


def _respell(name, args, spell):
    out = copy.deepcopy(args)
    if name in BEAD_FILTERS:
        for param in ("type", "exclude_type"):
            if param in out:
                out[param] = spell.each(out[param], spell.bead)
    elif name in LINK_FILTERS:
        if "type" in out:
            out["type"] = spell.each(out["type"], spell.link)
    elif name == "issue/dep/add":
        if "type" in out:
            out["type"] = spell.link(out["type"])
        _edges(out.get("edges"), spell)
    elif name == "issue/create":
        _rename_issue_type(out, spell)
        _edges(out.get("dependencies"), spell)
    elif name == "issue/update":
        _rename_issue_type(out.get("patch"), spell)
    elif name == "issue/create-many":
        for item in out.get("items") or []:
            _rename_issue_type(item, spell)
            _edges(item.get("dependencies") if isinstance(item, dict) else None, spell)
    elif name == "issue/batch":
        for item in out.get("items") or []:
            if not isinstance(item, dict):
                continue
            _rename_issue_type(item.get("create"), spell)
            _rename_issue_type((item.get("update") or {}).get("patch") if isinstance(item.get("update"), dict) else None, spell)
            dep = item.get("dep_add")
            if isinstance(dep, dict) and "type" in dep:
                dep["type"] = spell.link(dep["type"])
    return out, spell.refused


def body(value, scope):
    """A v0 result or Problem body as the Callable answers it."""
    spell = _Spelling(scope)

    def walk(node):
        if isinstance(node, list):
            return [walk(item) for item in node]
        if not isinstance(node, dict):
            return node
        out = {key: walk(item) for key, item in node.items()}
        if "issue_type" in out:
            # Issue section 2: the empty column is the native default, task.
            out["type"] = spell.bead(out.pop("issue_type") or "task")
        if "depends_on_id" in out and "type" in out:
            # Dependency and DependencyEdge.
            out["type"] = spell.link(out["type"])
        if out.get("kind") in ("local", "external") and "type" in out and set(out) <= {"id", "kind", "type"}:
            # Blocker.
            out["type"] = spell.link(out["type"])
        if "op" in out and isinstance(out.get("dep"), dict) and "kind" in out["dep"]:
            # EventRecord.dep: its kind is spelled `type`.
            dep = dict(out["dep"])
            dep["type"] = spell.link(dep.pop("kind"))
            out["dep"] = dep
        return out

    return walk(value)
