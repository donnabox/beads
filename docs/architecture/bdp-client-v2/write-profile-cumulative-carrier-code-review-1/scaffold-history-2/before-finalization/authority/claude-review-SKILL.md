---
name: claude-review
description: >-
  Get a code or architecture review from Claude Code. Use this skill whenever the user
  wants Claude's opinion, feedback, or review on anything — code files, architecture
  specs, PRs, diffs, documentation, configs, or any text content. Trigger on phrases
  like "claude review", "ask claude", "get claude's take", "what does claude think",
  "claude feedback", "run it by claude", or "get claude's perspective". For
  multi-agent reviews ("review with all agents", "get feedback from everyone"), use the
  review-council skill instead — it orchestrates all three reviewers.
---

# Claude Review

Route review requests to Claude Code. When Claude Code is the current host, use its
native review/delegation path instead of shelling out to the `claude` CLI. When the
current host is not Claude Code, use the external `claude` CLI in review-only mode and
instruct it not to write or modify files.

## Choose the execution path first

- If the current host is Claude Code, run the review natively. Do not spawn a child
  `claude` CLI just to review as Claude.
- If the current host is not Claude Code, verify the CLI exists with `command -v claude`
  and then run it externally.
- If `claude` is unavailable, tell the user Claude review is unavailable in this
  environment instead of silently substituting another reviewer.

## External CLI rules

- Use `claude -p` for headless execution.
- Tell Claude explicitly:
  - `Do not write or modify any files.`
  - `Output findings only in stdout.`
- Do **not** use `--permission-mode plan` by default. It can cause Claude to emit or
  persist plan artifacts instead of returning a plain review in stdout.
- Run Claude from the correct working directory:
  - for repo-context reviews, use the repo root as `cwd`
  - for single-file reviews, use a `cwd` that can read the target file
- If the target is outside the current working directory, do **not** copy it into the
  git worktree just to make Claude see it. Instead:
  - change `cwd`, or
  - pass the file contents in the prompt/stdin, or
  - copy to a temp directory outside the repo
- Use shell-safe prompts:
  - avoid backticks and command substitution inside the prompt string
  - prefer plain text path lists like `server/src/ws`, `src/shared`, `src/api`
- For repo-context reviews on medium/large codebases, expect the review to take
  time. A 5+ minute run can be normal. Do not assume the CLI is hung just
  because stdout is quiet for the first several minutes.
- For long repo-context reviews, do not kill and restart just because no
  partial output has appeared yet. First confirm whether the Claude process is
  still running normally.
- Prefer a direct `claude -p` invocation over extra wrapper layers that buffer
  stdout. If you must use a wrapper or subprocess helper, assume output may not
  appear until the process exits.
- Run one Claude review per target at a time. If a previous attempt was wrong, stop it
  before launching the corrected one.
- Check the actual exit code and actual output before retrying.

## External CLI examples

For file reviews:

```bash
claude -p "Read <file-path> and review this <description>.

Do not write or modify any files. Output findings only in stdout.

Provide a findings-first technical review. For each finding include:
- Severity: Critical, High, Medium, or Low
- The affected file, section, or line
- What is wrong
- Why it matters
- A concrete fix or next step

Focus on problems, risks, regressions, and missing tests. Do not spend time on praise."
```

For diff reviews:

```bash
git diff <base-ref>..HEAD | claude -p "Review this PR diff from stdin.

Do not write or modify any files. Output findings only in stdout.

Flag bugs, security issues, logic errors, behavioral regressions, and missing test
coverage. Return findings first."
```

For repo-context spec or plan reviews:

```bash
claude -p "Review <file-path> for consistency, completeness, and technical accuracy in the context of this repository.

You have full read access to the repository. Start from these relevant files/areas:
- <path-1>
- <path-2>
- <path-3>

Inspect any additional repo files as needed to validate the review target against the real codebase.
Do not write or modify any files. Output findings only in stdout.

Provide a findings-first technical review. For each finding include severity, location, issue, impact, and concrete fix."
```

Do not pass `--model` unless the user explicitly asks for it. The user's preferred
Claude model and auth are already configured.

## Output capture fallback

If the caller's exec/session wrapper is flaky, capture output to files:

1. Use `tmp/` only if it is gitignored, otherwise use an OS temp directory.
2. Redirect stdout/stderr to a file.
3. Write the exit code to a second file.
4. Read the file contents directly after completion.

Example:

```bash
mkdir -p tmp
rm -f tmp/claude-review.txt tmp/claude-review.exit
claude -p "Review ..." > tmp/claude-review.txt 2>&1
printf '%s\n' "$?" > tmp/claude-review.exit
```

If Claude prints a pointer to another artifact instead of the review text, read that
artifact and present the substantive findings. Do not blindly rerun Claude.

## Native Claude-host execution

When Claude Code itself is running the skill, use Claude's native review or delegation
mechanism with the same prompt contract:

- Read the requested files directly when the target is on disk.
- For diffs, provide the diff content or repo context directly.
- Ask for findings first, with severity, location, impact, and fix guidance.
- Keep the review read-only.

## Prompt contract

Use one consistent review prompt shape:

- What is being reviewed: file, diff, spec, config, or design
- Any focus area the user requested
- For repo-context reviews, give Claude a concrete starting set of relevant
  files or subsystems, while still explicitly granting it permission to read
  the broader repo as needed
- A requirement to prioritize concrete findings over praise
- Expected output: numbered findings with severity, location, issue, impact, and fix

If the user gives a specific focus area, keep it and add the findings-first structure.
If they say only "review this", use a broad technical review prompt that covers bugs,
security, scalability, operational risk, and missing tests.

## Reading the output

- Present the result as Claude's review.
- Keep the substantive findings and strip startup noise.
- Preserve file and line references when Claude provides them.
- If Claude says there are no findings, state that explicitly.

## Running alongside other reviewers

- If the user wants Claude plus one other reviewer, launch them in parallel.
- If the user wants Claude, Gemini, and Codex together, switch to `review-council`.
