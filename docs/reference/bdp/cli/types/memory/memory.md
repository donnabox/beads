---
title: "Memory CLI: canonical design and product evidence"
description: "Where to find the canonical Memory command design and source-bound Beads implementation evidence."
---

Memory relationships connect useful context, identify replacements and expose
explicit obligations. The [canonical Memory CLI companion](https://github.com/gastownhall/bdp/blob/1979925130a5744add8e45910c8a64931570bc13/docs/cli-design/memory-cli.md)
contains the task walkthrough and command contract beside the
[generic canonical CLI](https://github.com/gastownhall/bdp/blob/1979925130a5744add8e45910c8a64931570bc13/docs/cli-design/bdp-cli.md).

## Canonical command design

The companion is published for review in
[BDP PR302](https://github.com/gastownhall/bdp/pull/302), at exact source revision
`1979925130a5744add8e45910c8a64931570bc13`. It covers ten relationship reads,
five Link creation commands and their shared reference, CRUD and output rules.
The companion names its Memory Type-design dependency and distinguishes generic
CLI rules, product restrictions and deferred mechanics.

This Beads page is a navigation and product-evidence page, not a second
canonical command specification. The companion is a design candidate;
publication does not establish that its commands, output formats or Type model
are implemented or qualified in an installed build.

## Product availability and qualification

Use the [Preview 3 delivery ledger](https://github.com/gastownhall/beads/issues/7656)
and the matching release runbook for the exact build, supported behavior and
validation evidence. Older passing checks qualify only the exact sources and
catalogs they name. The canonical companion changes no runtime dependency,
workspace catalog, documentation release pin or availability switch.

The historical material below includes experimental assessment behavior and
an earlier catalog. Its preservation is not an expansion of the companion's
initial Memory surface. Shared graph machine-output framing and the v0.2
model's deferred pin and lifecycle mechanics retain the explicit limits in
the companion.

## Historical product correspondence

The complete earlier material remains available at immutable sources:

| Material | Status and preserved source |
| --- | --- |
| Earlier generic Memory facade | [Complete page at Beads `83e352d`](https://github.com/donnabox/beads/blob/83e352d834a2801625d29d6f1ba034a9b4cea11b/docs/reference/bdp/cli/types/memory/memory.md), including candidate availability, property editing, retyping, assessment examples and acceptance scope. Its inspected build and Type dependency are historical evidence, not current support claims. |
| Accepted relationship-command proposal | [Memory page in PR206 at `4c352eb`](https://github.com/donnabox/beads/blob/4c352eb1da0aba9de0e6feee5ba712374cfbb0a5/docs/reference/bdp/cli/types/memory/memory.md), preserving the ten-read/five-create design and its historical candidate ledger. The canonical companion reconciles its older syntax and presentation. |
| Earlier implementation and qualification boundaries | [PR206 review and validation record](https://github.com/donnabox/beads/blob/4c352eb1da0aba9de0e6feee5ba712374cfbb0a5/engdocs/design/memory-cli-preview3-review.md). Its graph, implementation-plan and acceptance-case artifacts retain their own review and integration status. |
| Proposal discussion and retained original | [Beads #7758](https://github.com/gastownhall/beads/issues/7758), now pointing to the canonical companion while retaining the earlier proposal. |

The older facade's BDP dependency
`afec9117914bd1c1fe0027c2067e4d89677e40a3` belongs to that source's catalog.
It does not silently qualify or adopt the companion's v0.2 Type-design target.
Follow the canonical companion for the authored command contract and the
source-bound product evidence for implemented behavior.
