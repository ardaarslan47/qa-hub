---
name: encode-rule
description: "Encode a rule or lesson as a git hook that blocks the commit or push when broken. Use when the user wants something enforced, banned or made mandatory, or when you catch the same correction a second time."
---

# Encode rule

A lesson written as prose is advice; the next agent can miss it. A lesson encoded as a hook rule is a **gate**: the commit or push fails until it is satisfied. This skill turns a lesson into a gate in `.githooks/`.

## 1. Pin the rule

Write the rule as one checkable sentence: what is forbidden or required, and on which input (staged files, commit message, pushed refs). If it needs human judgment and no script can decide it, stop: tell the user it can't be a gate and offer a sharper `CLAUDE.md` line with an example of the failure instead.

Done when you can name an input that must pass and one that must fail.

## 2. Pick the hook

| Input | Hook | What the rule receives |
|---|---|---|
| Staged file content | `pre-commit` | nothing; read `git diff --cached`, file content with `git show ":$path"` (never the working tree: CI's `check-range` swaps only HEAD and the index per commit) |
| Commit message | `commit-msg` | `$1` = message file |
| Branches and tags pushed | `pre-push` | `$1` = remote; stdin lines `<local-ref> <local-sha> <remote-ref> <remote-sha>` |

Prefer the earliest hook that sees the input: a `pre-commit` failure costs less than a `pre-push` one.

## 3. Write the rule

Create `.githooks/rules/<hook>/<kebab-name>`, a bash script, `chmod +x`. `.githooks/run` executes every executable file in that folder. Follow the existing rules (`rules/commit-msg/conventional-commit`, `rules/pre-push/trunk-refs`):

- No comments, no rule/reason header: the script must explain itself. Name the file, variables and functions after what they check (`semver`, `release_branch`, `on_trunk`), so an agent reading the code knows the rule without prose.
- Exit 0 on pass. On violation, write to stderr with a `[<kebab-name>]` prefix: what is wrong, where (file:line, ref, commit), and the exact fix. The reader is an agent that has never seen the rule; the message alone must get it to green.
- Scope `pre-commit` rules to staged content, so existing violations elsewhere don't block unrelated commits. If existing code violates the rule, report the count to the user and offer a separate cleanup commit.

## 4. Prove it

Add `expect ok` and `expect blocked` cases to `.githooks/tests/run` for the inputs from step 1 plus the edge cases you can think of, using real `git commit` / `git push`. Run `.githooks/tests/run`.

Done when every case is green, **and** a copy with your rule replaced by `exit 0` turns every new `blocked` case red (proves the tests exercise the rule), **and** the rule also runs in CI. `.githooks/check-range` covers `commit-msg` and `pre-commit` rules on its own. `pre-push` rules run in CI only where `.github/workflows/rules.yml` calls `.githooks/run pre-push` (today: tag pushes); if your rule must also hold for branch pushes, add that call.

## 5. Retire the prose

If `CLAUDE.md` or a doc states this rule, delete that line: the gate is the source of truth now. Commit the rule, its tests and the deletion together, e.g. `build(hooks): <rule> kuralı`.
