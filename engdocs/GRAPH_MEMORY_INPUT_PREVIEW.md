# Explicit Memory body input preview

In a normally initialized graph workspace, `remember` can create a Memory from
positional text, a file, or stdin. Choose exactly one source; `--id` and a
non-whitespace `--title` remain required in this preview.

```sh
bd remember --id beads/plan --title 'Our plan' --body-file plan.md
printf '%s' 'Plan from a pipe' | bd remember --id beads/pipe --title 'Pipe' --stdin
bd remember --id beads/literal --title 'Literal text' -- '--- body text'
bd recall beads/plan
```

The file path is literal: `--body-file -` reads a file named `-`; only `--stdin`
reads standard input. Explicit `--stdin=false` is rejected. Empty files, empty
stdin and empty positional text are valid with the required title. Markdown
frontmatter, CRLF, Unicode and whitespace remain body content. No parsing,
trimming, title derivation or trailing newline is added.

All three graph body sources have a provisional 1 MiB UTF-8 input limit,
advertised as `limits.memoryBodyInputBytes` by `status --graph --json`. This
also bounds positional input that previously had no CLI acquisition limit.
Oversized input returns `capability_unavailable`, without truncation. Missing
files, read errors, invalid UTF-8 and conflicting sources refuse before opening
storage. Readonly and migration-freeze policy is checked before reading input.
Caller-supplied stdin and files must finish reading; this is not a network stream
service and does not impose a wall-clock input deadline.

Creation uses the existing Memory write transaction and retained snapshot;
there is no schema change, second store or alternate writer. A new process can
`show` or `recall` the current Memory or its saved `--version TOKEN`.

The new flags are graph-only and refuse before legacy storage is opened.
Existing legacy `remember` positional/key behavior is preserved. This slice
does not implement generated identities, aliases, derived titles, common
metadata, Inception, derivation fields, rich Memory JSON, ordered History,
restore or public BDP History. It does not make the preview Memory Type a final
contract. The authoring proposal remains [upstream #6703](https://github.com/gastownhall/beads/issues/6703).

## Reproduce the installed-command proof

Install through `make install-force INSTALL_DIR=/absolute/disposable/bin`.
Start an ordinary Dolt 2.1.8 server with a fresh disposable data directory, then:

```sh
python3 scripts/graph-memory-input-smoke.py \
  --bd /absolute/disposable/bin/bd --backend both --server-port PORT \
  --output-dir /absolute/new/evidence-directory
```

The harness initializes each workspace through `bd init`, creates Memories
through the installed CLI, checks complete raw current/exact recall in fresh
processes, and tests invalid source, UTF-8, size and policy refusals. It uses no
schema seeding, SQL fixture or mock. It preserves failed receipts and cleans up
its subprocesses; the caller owns the server. Provision databases serially on
ordinary Dolt 2.1.8. This harness does not qualify concurrent writes, crash
recovery, public HTTP interoperability or the complete delivery plan.
