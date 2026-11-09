# Working with BDP

BDP is the protocol beads speaks. A copy of it lives in this repo, fixed to one version. You almost never have to think about it. This is the whole of what an engineer needs, in three situations.

## The daily grind

Build and test the way you always do. The copy of BDP is just there, like any other vendored dependency. You don't fetch it, update it, or read the spec.

One thing can happen: a protocol test fails on your change. That test guards what beads promises over the network. A failure means your change altered that promise. Ask yourself one question — did you mean to?

- **No.** It's a side effect. Fix your change until the test passes again.
- **Yes.** You're deliberately changing what beads sends or accepts. That's allowed, but it's never silent. Say so in your PR — what changed and why — so it's a decision someone chose, not a surprise someone finds later.

Two standing rules, and that's all:

- Don't edit the vendored BDP files or the version it's fixed to. That's release management's, not yours.
- If you think the spec itself is wrong, file it. Don't work around it in the code.

## Cutting a release (the endgame)

This is release management's job, not every engineer's.

Before the release, make sure the vendored BDP is at the version you're shipping against. If BDP has moved, bump the copy to its current release and run the checks. The release notes name the BDP version beads matches.

Any deliberate protocol change that's been waiting from the daily grind lands here, with that bump, named.

## Fixes after a release

A fix to already-shipped protocol behavior has two halves, both release management's:

- The spec fix goes on the BDP release branch for that version.
- The matching beads release branch bumps its copy to pick it up. Newer versions only; you never go backward.

If you found the problem, file it. You don't manage the branches or the version yourself.

A fix that doesn't touch protocol behavior is just a normal fix. Nothing special about it.

## For release management: the copy and the release branch

This is the part the three situations above hand off to. Two moving pieces, both yours.

**The vendored copy.** It's fixed to one BDP version. To pick up a newer BDP release, bump it forward: one PR that copies the new files in and runs the checks. Forward only — never roll it back to an older BDP. A beads release names the BDP version it matches, and the "does beads match BDP" checks run right here against the copy, so a bump that breaks them never ships.

**The BDP release branch.** When a shipped beads release needs a protocol fix, the fix is made on BDP's release branch for that version. You bump that beads release branch's copy to pick it up. The main line keeps moving on its own; the release branch only ever takes fixes.

That's the whole of it. Everything else about BDP — writing it, testing it, cutting a version — happens in the BDP repo before a version exists. Someone bumping the copy never has to open it.

## The pin

`pin` (next to this file) is one line: the BDP version this repo is built and tested against. It moves forward only. Today it names a commit in the BDP repo; once the spec, wire schema, and conformance suite are vendored into `spec/` here, it becomes the local version of that copy. Only release management changes it.
